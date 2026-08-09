package config

import (
	"encoding/json"
	"os"

	"sing-monitor-server/models"
)

type Config struct {
	APIServerPort          int    `json:"api_server_port"`
	SingBoxGrpcAddr        string `json:"sing_box_grpc_addr"`
	CollectIntervalSeconds int    `json:"collect_interval_seconds"`
	DBPath                 string `json:"db_path"`
	DefaultCycleDays       int    `json:"default_cycle_days"`
}

func LoadConfig(path string) (*Config, error) {
	// Default config
	cfg := &Config{
		APIServerPort:          8080,
		SingBoxGrpcAddr:        "127.0.0.1:10000",
		CollectIntervalSeconds: 300,
		DBPath:                 "sing-monitor.db",
		DefaultCycleDays:       models.DefaultCycleDays,
	}

	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			// If config doesn't exist, return default
			return cfg, nil
		}
		return nil, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
