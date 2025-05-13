package main

import (
	"apim-deployer/tasks"
	"apim-deployer/types"
	"encoding/json"
	"fmt"
	"log"
	"os"
)

type NodeProfile struct {
	Enabled       bool `json:"enabled"`
	Count         int  `json:"count"`
	EnableHA      bool `json:"enable_ha"`
	EnableProfile bool `json:"enable_profiling"`
}

type DBConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
}

type Config struct {
	APIMZipPath    string      `json:"apim_zip_path"`
	Version        string      `json:"version"`
	UpdateLevel    string      `json:"update_level"`
	Gateway        NodeProfile `json:"gateway"`
	TrafficManager NodeProfile `json:"traffic_manager"`
	DevPortal      NodeProfile `json:"developer_portal"`
	Publisher      NodeProfile `json:"publisher"`
	ControlPlane   NodeProfile `json:"control_plane"`
	DatabaseConfig DBConfig    `json:"database"`
	DBDriverPath   string      `json:"db_driver_path"`
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: apim-deployer <path-to-config.json>")
	}

	configPath := os.Args[1]
	configFile, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}

	var cfg types.Config
	err = json.Unmarshal(configFile, &cfg)
	if err != nil {
		log.Fatalf("Invalid config JSON: %v", err)
	}

	fmt.Printf("Configuration loaded successfully:\n%+v\n", cfg)

	extractDir := "./extracted/"
	subDir, err := tasks.UnzipAPIM(cfg.APIMZipPath, extractDir)
	if err != nil {
		log.Fatalf("Unzip failed: %v", err)
	}

	err = tasks.ApplyUpdate(cfg.UpdateLevel, subDir)
	if err != nil {
		log.Fatalf("Update failed: %v", err)
	}

	err = tasks.CopyDBDriver(cfg.DBDriverPath, subDir)
	if err != nil {
		log.Fatalf("DB Driver copy failed: %v", err)
	}

	err = tasks.SetupDatabases(
		cfg.DatabaseConfig.Host,
		cfg.DatabaseConfig.Port,
		cfg.DatabaseConfig.User,
		cfg.DatabaseConfig.Password,
		cfg.Version,
		cfg.DatabaseConfig.APIMDBName,
		cfg.DatabaseConfig.SharedDBName,
	)
	if err != nil {
		log.Fatalf("❌ Database setup failed: %v", err)
	}

	err = tasks.CopyPacks(cfg, subDir)
	if err != nil {
		log.Fatalf("❌ Failed to copy APIM packs: %v", err)
	}

	err = tasks.ApplyProfiling(cfg)
	if err != nil {
		log.Fatalf("❌ Failed while executing profiling: %v", err)
	}

	err = tasks.GenerateConfigurations(cfg)
	if err != nil {
		log.Fatalf("❌ Config copy failed: %v", err)
	}

	fmt.Println("✅ All steps completed successfully.")
}
