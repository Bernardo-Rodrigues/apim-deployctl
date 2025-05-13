package tasks

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"syscall"

	"golang.org/x/term"
)

func ApplyUpdate(updateLevel string, apimDir string) error {
	if updateLevel == "vanilla" {
		fmt.Println("ℹ️ Skipping update step (vanilla selected)")
		return nil
	}

	fmt.Printf("🔧 Applying update level: %s\n", updateLevel)

	var scriptName string
	switch runtime.GOOS {
	case "darwin":
		scriptName = "wso2update_darwin"
	case "linux":
		scriptName = "wso2update_linux"
	default:
		return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
	scriptPath := filepath.Join(apimDir, "bin", scriptName)

	// Prompt credentials
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("🔐 Enter WSO2 Update Username: ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)

	fmt.Print("🔐 Enter WSO2 Update Password: ")
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		return fmt.Errorf("failed to read password: %v", err)
	}
	password := strings.TrimSpace(string(bytePassword))

	args := buildArgs(updateLevel, username, password)

	// Run first attempt
	err = runUpdateCommand(scriptPath, args)
	if err != nil {
		// Check for auto-update message
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 2 {
			fmt.Println("🔁 Update tool self-updated. Re-running...")
			err = runUpdateCommand(scriptPath, args)
		}
	}

	if err != nil {
		return fmt.Errorf("❌ update script failed: %v", err)
	}

	fmt.Println("✅ Update applied successfully")
	return nil
}

func buildArgs(updateLevel, username, password string) []string {
	var args []string
	if updateLevel != "latest" {
		args = append(args, "-l", updateLevel)
	}
	return append(args, "-u", username, "-p", password)
}

func runUpdateCommand(scriptPath string, args []string) error {
	cmd := exec.Command(scriptPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
