package tasks

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"apim-deployer/types"
)

func ApplyProfiling(cfg types.Config) error {
	fmt.Println("🎛️  Running profileSetup.sh for enabled profiles...")

	profiles := map[string]struct {
		ProfileName string
		Info        types.NodeProfile
	}{
		"gateway":       {"gateway-worker", cfg.Gateway},
		"tm":            {"traffic-manager", cfg.TrafficManager},
		"control-plane": {"control-plane", cfg.ControlPlane},
	}

	for dirPrefix, profile := range profiles {
		if !profile.Info.Enabled || !profile.Info.EnableProfile {
			continue
		}

		count := profile.Info.Count
		if profile.Info.EnableHA {
			count = 2
		}

		for i := 1; i <= count; i++ {
			binPath := filepath.Join("deployment", fmt.Sprintf("%s-%d", dirPrefix, i), "bin")

			scriptName := "profileSetup.sh"
			if runtime.GOOS == "windows" {
				scriptName = "profileSetup.bat"
			}

			script := filepath.Join(binPath, scriptName)

			_ = os.Chmod(script, 0755) // ensure it's executable

			absPath, err := filepath.Abs(script)
			if err != nil {
				return fmt.Errorf("failed to resolve script path: %v", err)
			}

			cmd := exec.Command(absPath, "-Dprofile="+profile.ProfileName)
			cmd.Dir = binPath
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			fmt.Printf("🔧 Running profileSetup for %s-%d -> %s\n", dirPrefix, i, profile.ProfileName)

			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to apply profileSetup for %s-%d: %v", dirPrefix, i, err)
			}

			fmt.Printf("✅ Profile applied for %s-%d\n", dirPrefix, i)
		}
	}

	return nil
}
