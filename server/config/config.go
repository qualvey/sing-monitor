package config

import (
	"os"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

// Config 对齐原 singbox-monitor 的 config.yaml 结构
type Config struct {
	Log struct {
		Level string `yaml:"level"`
	} `yaml:"log"`

	Singbox struct {
		Address      string `yaml:"address"`
		PollInterval string `yaml:"poll_interval"`
		ResetOnQuery bool   `yaml:"reset_on_query"`
		Pattern      string `yaml:"pattern"`
	} `yaml:"singbox"`

	Postgres struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		DBName   string `yaml:"dbname"`
		SSLMode  string `yaml:"sslmode"`
	} `yaml:"postgres"`

	Control struct {
		Enabled        bool   `yaml:"enabled"`
		ConfigPath     string `yaml:"config_path"`
		TemplatePath   string `yaml:"template_path"`
		CheckCommand   string `yaml:"check_command"`
		ReloadCommand  string `yaml:"reload_command"`
	} `yaml:"control"`

	RealTime struct {
		Enabled            bool `yaml:"enabled"`
		IntervalMS         int  `yaml:"interval_ms"`
		OnlineThresholdSec int  `yaml:"online_threshold_sec"`
	} `yaml:"real_time"`

	Server struct {
		Port int `yaml:"port"`
	} `yaml:"server"`

	Auth struct {
		Password string `yaml:"password"`
		Secret   string `yaml:"secret"`
	} `yaml:"auth"`
}

func LoadConfig(path string) (*Config, error) {
	cfg := &Config{}
	cfg.Log.Level = "info"
	cfg.Singbox.Address = "127.0.0.1:8080"
	cfg.Singbox.PollInterval = "10s"
	cfg.Singbox.ResetOnQuery = true
	cfg.Postgres.Host = "127.0.0.1"
	cfg.Postgres.Port = 5432
	cfg.Postgres.User = "singbox"
	cfg.Postgres.DBName = "singbox"
	cfg.Postgres.SSLMode = "disable"
	cfg.Control.Enabled = true
	cfg.Control.ConfigPath = "/etc/sing-box/config.json"
	cfg.Control.CheckCommand = "sing-box check -c %s"
	cfg.Control.ReloadCommand = "sudo systemctl reload sing-box"
	cfg.RealTime.Enabled = true
	cfg.RealTime.IntervalMS = 1000
	cfg.RealTime.OnlineThresholdSec = 120
	cfg.Server.Port = 8090
	cfg.Auth.Password = "admin"
	cfg.Auth.Secret = "change-me"

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) PollIntervalDuration() time.Duration {
	d, err := time.ParseDuration(c.Singbox.PollInterval)
	if err != nil || d <= 0 {
		return 10 * time.Second
	}
	return d
}

func (c *Config) DSN() string {
	port := c.Postgres.Port
	if port == 0 {
		port = 5432
	}
	return "host=" + c.Postgres.Host +
		" port=" + strconv.Itoa(port) +
		" user=" + c.Postgres.User +
		" password=" + c.Postgres.Password +
		" dbname=" + c.Postgres.DBName +
		" sslmode=" + c.Postgres.SSLMode
}
