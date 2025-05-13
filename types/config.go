package types

type NodeProfile struct {
	Enabled       bool `json:"enabled"`
	Count         int  `json:"count"`
	EnableHA      bool `json:"enable_ha"`
	EnableProfile bool `json:"enable_profiling"`
}

type DBConfig struct {
	Host         string `json:"host"`
	Port         int    `json:"port"`
	User         string `json:"user"`
	Password     string `json:"password"`
	APIMDBName   string `json:"apim_db_name"`
	SharedDBName string `json:"shared_db_name"`
}

type Config struct {
	APIMZipPath    string      `json:"apim_zip_path"`
	Version        string      `json:"version"`
	UpdateLevel    string      `json:"update_level"`
	Gateway        NodeProfile `json:"gateway"`
	TrafficManager NodeProfile `json:"traffic_manager"`
	ControlPlane   NodeProfile `json:"control_plane"`
	DatabaseConfig DBConfig    `json:"database"`
	DBDriverPath   string      `json:"db_driver_path"`
}
