# WSO2 API Manager 4.x Deployment Tool (Go CLI)

A Go-based CLI tool that automates the deployment of **WSO2 API Manager (APIM)** 4.x environments. It handles ZIP extraction, `deployment.toml` customization, high availability (HA) configuration, and optional database initialization.

> **Only supports WSO2 APIM version 4.x**

---

## Features

- Deploy Gateway, Control Plane, and Traffic Manager profiles
- Generate version-based `deployment.toml` with dynamic config injection
- High Availability (HA) and profiling support per node
- Optional: Run database creation scripts with user-specified DB names
- Cross-platform: Build binaries for Linux, macOS, and Windows

---

## Configuration (`config.json`)

Here’s a sample `config.json`:

```json
{
  "apim_zip_path": "/path/to/wso2am-4.4.0.zip",
  "version": "4.4.0",
  "update_level": "vanilla", //vanilla, latest, 20
  "gateway": {
    "enabled": true,
    "count": 2,
    "enable_ha": true,
    "enable_profiling": true
  },
  "traffic_manager": {
    "enabled": false,
    "count": 1,
    "enable_ha": true,
    "enable_profiling": true
  },
  "control_plane": {
    "enabled": true,
    "count": 1,
    "enable_ha": true,
    "enable_profiling": true
  },
  "database": {
    "host": "localhost",
    "port": 3306,
    "user": "root",
    "password": "password",
    "apim_db_name": "TEST_APIM_DB",
    "shared_db_name": "TEST_SHARED_DB"
  },
  "db_driver_path": "/path/to/mysql-connector-j-8.0.31.jar"
}
```

Explanation of Fields: 

- apim_zip_path: Path to the WSO2 APIM 4.x ZIP file
- version: WSO2 APIM version (must be 4.x)
- update_level: vanilla, latest, or a specific update level
- gateway.count: Number of Gateway nodes to deploy
- enable_ha: Whether to inject HA-related config for this profile
- enable_profiling: Whether to run profiling script or not
- db_driver_path: Path to MySQL JDBC connector JAR
- database block: DB connection and schema config

## Running the Tool

Ensure you’re running the binary from the project root, where folders like scripts/ and templates/ exist.


On macOS Apple
```
./apim-deployer-mac-apple config.json
```

On macOS Intel
```
./apim-deployer-mac-intel config.json
```

On Linux
```
./apim-deployer-linux config.json
```

On Windows
```
./apim-deployer-win.exe config.json
```

## Outputs

After a successful run, you’ll find all generated deployments organized under the /deployment directory:
- A fully extracted WSO2 APIM 4.x runtime
- A tailored deployment.toml based on the role, HA setting, and profiling
- MySQL JDBC driver in the correct repository/components/lib/ path
- Optional: Databases pre-created with your custom names