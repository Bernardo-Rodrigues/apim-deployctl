package tasks

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"apim-deployer/types"
)

func CopyPacks(cfg types.Config, source string) error {
	fmt.Println("📦 Copying packs per profile...")

	profiles := map[string]types.NodeProfile{
		"gateway":       cfg.Gateway,
		"tm":            cfg.TrafficManager,
		"control-plane": cfg.ControlPlane,
	}

	for name, profile := range profiles {
		if !profile.Enabled {
			continue
		}
		count := profile.Count
		if profile.EnableHA {
			count = 2
		}
		for i := 1; i <= count; i++ {
			destDir := filepath.Join("deployment", fmt.Sprintf("%s-%d", name, i))
			err := copyDir(source, destDir)
			if err != nil {
				return fmt.Errorf("failed to copy %s-%d: %v", name, i, err)
			}
			fmt.Printf("✅ Copied %s to %s\n", name, destDir)
		}
	}

	return nil
}

func copyDir(src, dest string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath := strings.TrimPrefix(path, src)
		targetPath := filepath.Join(dest, relPath)

		if info.IsDir() {
			return os.MkdirAll(targetPath, info.Mode())
		}

		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		destFile, err := os.Create(targetPath)
		if err != nil {
			return err
		}
		defer destFile.Close()

		_, err = io.Copy(destFile, srcFile)
		return err
	})
}
