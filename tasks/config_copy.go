package tasks

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"apim-deployer/types"
)

func CopyConfigurations(cfg types.Config) error {
	fmt.Println("📝 Copying configuration files...")

	profiles := map[string]struct {
		ConfigDir  string
		ProfileCfg types.NodeProfile
	}{
		"gateway":       {"gateway", cfg.Gateway},
		"tm":            {"tm", cfg.TrafficManager},
		"control-plane": {"control-plane", cfg.ControlPlane},
	}

	for dirPrefix, profile := range profiles {
		if !profile.ProfileCfg.Enabled {
			continue
		}

		count := profile.ProfileCfg.Count
		if profile.ProfileCfg.EnableHA {
			count = 2
		}

		for i := 1; i <= count; i++ {
			destPath := filepath.Join("deployment", fmt.Sprintf("%s-%d", dirPrefix, i), "repository", "conf", "deployment.toml")

			var sourcePath string

			if dirPrefix == "gateway" {
				sourcePath = filepath.Join("configs", cfg.Version, "gateway", "deployment.toml")
				if i == 1 {
					if err := CopyFile(sourcePath, destPath); err != nil {
						return fmt.Errorf("❌ Failed to copy gateway-1 config: %v", err)
					}
				} else {
					if err := copyAndAdjustOffset(sourcePath, destPath, i); err != nil {
						return fmt.Errorf("❌ Failed to copy gateway-%d config with offset: %v", i, err)
					}
				}
			} else {
				if profile.ProfileCfg.EnableHA {
					sourcePath = filepath.Join("configs", cfg.Version, profile.ConfigDir, "ha", fmt.Sprintf("deployment%d.toml", i))
				} else {
					sourcePath = filepath.Join("configs", cfg.Version, profile.ConfigDir, "default", "deployment.toml")
				}
				if err := CopyFile(sourcePath, destPath); err != nil {
					return fmt.Errorf("❌ Failed to copy %s-%d config: %v", dirPrefix, i, err)
				}
			}

			fmt.Printf("✅ Copied config for %s-%d\n", dirPrefix, i)
		}
	}

	return nil
}

func copyAndAdjustOffset(src, dst string, increment int) error {
	inFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer inFile.Close()

	scanner := bufio.NewScanner(inFile)
	var lines []string

	offsetRegex := regexp.MustCompile(`(?i)^\s*offset\s*=\s*(\d+)`)

	for scanner.Scan() {
		line := scanner.Text()
		if offsetRegex.MatchString(line) {
			match := offsetRegex.FindStringSubmatch(line)
			if len(match) == 2 {
				base, _ := strconv.Atoi(match[1])
				line = fmt.Sprintf("offset = %d", base+increment)
			}
		}
		lines = append(lines, line)
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	output := strings.Join(lines, "\n")
	os.MkdirAll(filepath.Dir(dst), 0755)
	return os.WriteFile(dst, []byte(output), 0644)
}
