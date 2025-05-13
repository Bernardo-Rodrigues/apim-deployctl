package tasks

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"apim-deployer/types"
)

func GenerateConfigurations(cfg types.Config) error {
	fmt.Println("🛠 Generating deployment.toml files dynamically (preserve structure)...")

	profiles := map[string]struct {
		TemplateName string
		Profile      types.NodeProfile
	}{
		"gateway":       {"gateway", cfg.Gateway},
		"tm":            {"tm", cfg.TrafficManager},
		"control-plane": {"control-plane", cfg.ControlPlane},
	}

	for dirPrefix, def := range profiles {
		if !def.Profile.Enabled {
			continue
		}

		count := def.Profile.Count
		if def.Profile.EnableHA {
			count = 2
		}

		for i := 1; i <= count; i++ {
			templatePath := filepath.Join("configs", cfg.Version, def.TemplateName+".toml")
			contentBytes, err := os.ReadFile(templatePath)
			if err != nil {
				return fmt.Errorf("failed to read base template: %v", err)
			}

			contentLines := strings.Split(string(contentBytes), "\n")
			var updatedLines []string
			offsetRegex := regexp.MustCompile(`(?i)^\s*offset\s*=\s*(\d+)`)
			for _, line := range contentLines {
				if offsetRegex.MatchString(line) {
					match := offsetRegex.FindStringSubmatch(line)
					baseOffset, _ := strconv.Atoi(match[1])
					line = fmt.Sprintf("offset = %d", baseOffset+(i-1))
				}
				updatedLines = append(updatedLines, line)
			}

			placeholders := map[string]string{
				"${db_host}":        cfg.DatabaseConfig.Host,
				"${db_port}":        strconv.Itoa(cfg.DatabaseConfig.Port),
				"${db_user}":        cfg.DatabaseConfig.User,
				"${db_password}":    cfg.DatabaseConfig.Password,
				"${apim_db_name}":   cfg.DatabaseConfig.APIMDBName,
				"${shared_db_name}": cfg.DatabaseConfig.SharedDBName,
			}

			for j, line := range updatedLines {
				for key, val := range placeholders {
					line = strings.ReplaceAll(line, key, val)
				}
				updatedLines[j] = line
			}

			appendix := generateHAAppendix(dirPrefix, i, cfg)
			if appendix != "" {
				updatedLines = append(updatedLines, "", appendix)
			}

			destPath := filepath.Join("deployment", fmt.Sprintf("%s-%d", dirPrefix, i), "repository", "conf", "deployment.toml")
			os.MkdirAll(filepath.Dir(destPath), 0755)
			if err := os.WriteFile(destPath, []byte(strings.Join(updatedLines, "\n")), 0644); err != nil {
				return fmt.Errorf("failed to write %s: %v", destPath, err)
			}

			fmt.Printf("✅ Generated config for %s-%d\n", dirPrefix, i)
		}
	}

	return nil
}

func generateHAAppendix(profile string, nodeIndex int, cfg types.Config) string {
	tmEnabled := cfg.TrafficManager.Enabled
	tmHA := cfg.TrafficManager.EnableHA
	cpHA := cfg.ControlPlane.EnableHA
	cpEnabled := cfg.ControlPlane.Enabled

	switch profile {
	case "tm":
		var sb strings.Builder
		if tmHA {
			if nodeIndex == 1 {
				sb.WriteString(`[apim.throttling]
event_duplicate_url = ["tcp://localhost:5675"]
throttle_decision_endpoints = ["tcp://localhost:5674"]\n`)
			} else {
				sb.WriteString(`[apim.throttling]
event_duplicate_url = ["tcp://localhost:5674"]
throttle_decision_endpoints = ["tcp://localhost:5675"]\n`)
			}
		}
		if cpEnabled && !cpHA {
			sb.WriteString(`
[apim.event_hub]
enable = true
username = "$ref{super_admin.username}"
password = "$ref{super_admin.password}"
service_url = "https://localhost:9443/services/"
event_listening_endpoints = ["tcp://localhost:5672"]
`)
		} else if cpEnabled && cpHA {
			sb.WriteString(`
[apim.event_hub]
enable = true
username = "$ref{super_admin.username}"
password = "$ref{super_admin.password}"
service_url = "https://localhost:9443/services/"
event_listening_endpoints = ["tcp://localhost:5672", "tcp://localhost:5673"]
`)
		}
		return sb.String()

	case "control-plane":
		var sb strings.Builder
		if cpHA {
			fmt.Fprintf(&sb, `[apim.event_hub]
enable = true
username = "$ref{super_admin.username}"
password = "$ref{super_admin.password}"
service_url = "https://localhost:9443/services/"
`)
			if nodeIndex == 1 {
				fmt.Fprintf(&sb, `event_listening_endpoints = ["tcp://localhost:5672"]
event_duplicate_url = ["tcp://localhost:5673"]
`)
			} else {
				fmt.Fprintf(&sb, `event_listening_endpoints = ["tcp://localhost:5673"]
event_duplicate_url = ["tcp://localhost:5672"]
`)
			}
			fmt.Fprintf(&sb, `
[[apim.event_hub.publish.url_group]]
urls = ["tcp://localhost:9611"]
auth_urls = ["ssl://localhost:9711"]

[[apim.event_hub.publish.url_group]]
urls = ["tcp://localhost:9612"]
auth_urls = ["ssl://localhost:9712"]
`)
		} else if tmEnabled && tmHA {
			sb.WriteString(`
[apim.throttling]
service_url = "https://localhost:9445/services/"
throttle_decision_endpoints = ["tcp://localhost:5674","tcp://localhost:5675"]

[[apim.throttling.url_group]]
traffic_manager_urls = ["tcp://localhost:9613"]
traffic_manager_auth_urls = ["ssl://localhost:9713"]

[[apim.throttling.url_group]]
traffic_manager_urls = ["tcp://localhost:9614"]
traffic_manager_auth_urls = ["ssl://localhost:9714"]
`)
		} else if tmEnabled && !tmHA {
			sb.WriteString(`
[apim.throttling]
service_url = "https://localhost:9445/services/"
throttle_decision_endpoints = ["tcp://localhost:5674"]

[[apim.throttling.url_group]]
traffic_manager_urls = ["tcp://localhost:9613"]
traffic_manager_auth_urls = ["ssl://localhost:9713"]
`)
		}
		return sb.String()

	case "gateway":
		var sb strings.Builder
		if tmEnabled && tmHA {
			sb.WriteString(`
[apim.throttling]
throttle_decision_endpoints = ["tcp://localhost:5674", "tcp://localhost:5675"]

[[apim.throttling.url_group]]
traffic_manager_urls = ["tcp://localhost:9613"]
traffic_manager_auth_urls = ["ssl://localhost:9713"]

[[apim.throttling.url_group]]
traffic_manager_urls = ["tcp://localhost:9614"]
traffic_manager_auth_urls = ["ssl://localhost:9714"]
`)
		} else if tmEnabled && !tmHA {
			sb.WriteString(`
[apim.throttling]
throttle_decision_endpoints = ["tcp://localhost:5674"]

[[apim.throttling.url_group]]
traffic_manager_urls = ["tcp://localhost:9613"]
traffic_manager_auth_urls = ["ssl://localhost:9713"]
`)
		} else if !tmEnabled && cpHA {
			sb.WriteString(`
[apim.throttling]
service_url = "https://localhost:9443/services/"
throttle_decision_endpoints = ["tcp://localhost:5672","tcp://localhost:5673"]

[[apim.throttling.url_group]]
traffic_manager_urls = ["tcp://localhost:9612"]
traffic_manager_auth_urls = ["ssl://localhost:9712"]

[[apim.throttling.url_group]]
traffic_manager_urls = ["tcp://localhost:9611"]
traffic_manager_auth_urls = ["ssl://localhost:9711"]
`)
		} else if !tmEnabled && cpEnabled && !cpHA {
			sb.WriteString(`
[apim.throttling]
service_url = "https://localhost:9443/services/"
throttle_decision_endpoints = ["tcp://localhost:5672"]

[[apim.throttling.url_group]]
traffic_manager_urls = ["tcp://localhost:9611"]
traffic_manager_auth_urls = ["ssl://localhost:9711"]
`)
		}
		if tmEnabled && cpHA {
			sb.WriteString(`
[apim.event_hub]
enable = true
username = "$ref{super_admin.username}"
password = "$ref{super_admin.password}"
service_url = "https://localhost:9443/services/"
event_listening_endpoints = ["tcp://localhost:5672", "tcp://localhost:5673"]
`)
		} else if tmEnabled && cpEnabled && !cpHA {
			sb.WriteString(`
[apim.event_hub]
enable = true
username = "$ref{super_admin.username}"
password = "$ref{super_admin.password}"
service_url = "https://localhost:9443/services/"
event_listening_endpoints = ["tcp://localhost:5672"]
`)
		}
		return sb.String()
	}

	return ""
}
