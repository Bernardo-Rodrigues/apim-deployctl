package tasks

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

func SetupDatabases(host string, port int, user, password, version string, amdb string, shdb string) error {
	fmt.Println("🛠️ Setting up databases...")

	placeholders := map[string]string{
		"${apim_db_name}":   amdb,
		"${shared_db_name}": shdb,
	}

	apimInit, err := prepareSQLScriptWithPlaceholders("scripts/init/create_apim_db.sql", placeholders)
	if err != nil {
		return fmt.Errorf("prep apim db script: %w", err)
	}
	defer os.Remove(apimInit)

	sharedInit, err := prepareSQLScriptWithPlaceholders("scripts/init/create_shared_db.sql", placeholders)
	if err != nil {
		return fmt.Errorf("prep shared db script: %w", err)
	}
	defer os.Remove(sharedInit)

	// Run init scripts without selecting a DB
	if err := runSQLScript(user, password, host, port, "", apimInit); err != nil {
		return fmt.Errorf("failed to create APIM_DB: %v", err)
	}
	if err := runSQLScript(user, password, host, port, "", sharedInit); err != nil {
		return fmt.Errorf("failed to create SHARED_DB: %v", err)
	}

	// Run version-specific schema with explicit DB name
	schemaDir := filepath.Join("scripts", "schema", version)
	apimSchema := filepath.Join(schemaDir, "apim_tables.sql")
	sharedSchema := filepath.Join(schemaDir, "shared_tables.sql")

	if err := runSQLScript(user, password, host, port, amdb, apimSchema); err != nil {
		return fmt.Errorf("failed to load APIM schema: %v", err)
	}
	if err := runSQLScript(user, password, host, port, shdb, sharedSchema); err != nil {
		return fmt.Errorf("failed to load SHARED schema: %v", err)
	}

	fmt.Println("✅ Database setup complete")
	return nil
}

func runSQLScript(user, password, host string, port int, dbName, scriptPath string) error {
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return fmt.Errorf("script not found: %s", scriptPath)
	}

	// safe to execute in all OS
	args := []string{
		"-u" + user,
		"-p" + password,
		"-h" + host,
		"-P" + strconv.Itoa(port),
	}
	if dbName != "" {
		args = append(args, dbName)
	}
	args = append(args, "-e", "source "+filepath.ToSlash(scriptPath)+";")

	cmd := exec.Command("mysql", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func prepareSQLScriptWithPlaceholders(scriptPath string, placeholders map[string]string) (string, error) {
	data, err := os.ReadFile(scriptPath)
	if err != nil {
		return "", err
	}
	content := string(data)
	for key, value := range placeholders {
		content = strings.ReplaceAll(content, key, value)
	}

	// Write to a temp file
	tmpFile, err := os.CreateTemp("", "wso2-init-*.sql")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	if _, err := tmpFile.WriteString(content); err != nil {
		return "", err
	}

	// Return absolute path
	absPath, err := filepath.Abs(tmpFile.Name())
	if err != nil {
		return "", fmt.Errorf("error in absolute path: %s", err)
	}
	return absPath, nil
}
