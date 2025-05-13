package tasks

import (
	"fmt"
	"os"
	"path/filepath"
)

func CopyDBDriver(driverPath string, apimDir string) error {
	fmt.Println("📦 Copying database driver...")

	if _, err := os.Stat(driverPath); os.IsNotExist(err) {
		return fmt.Errorf("❌ DB driver not found: %s", driverPath)
	}

	destDir := filepath.Join(apimDir, "repository", "components", "lib")

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("❌ Failed to create lib directory: %v", err)
	}

	destFile := filepath.Join(destDir, filepath.Base(driverPath))
	err := CopyFile(driverPath, destFile)
	if err != nil {
		return fmt.Errorf("❌ Failed to copy DB driver: %v", err)
	}

	fmt.Println("✅ DB driver copied to:", destFile)
	return nil
}
