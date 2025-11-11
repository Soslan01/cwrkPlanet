package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

type HTTP struct {
	Addr         string        `yaml:"addr"`         // ":8080"
	ReadTimeout  time.Duration `yaml:"readTimeout"`  // "15s"
	WriteTimeout time.Duration `yaml:"writeTimeout"` // "30s"
	IdleTimeout  time.Duration `yaml:"idleTimeout"`  // "60s"
}

type Upstream struct {
	AuthTarget     string `yaml:"authTarget"`
	RoomGRPCTarget string `yaml:"roomGRPCTarget"`
}

type Logging struct {
	Env       string `yaml:"env"`       // dev|stage|prod
	Service   string `yaml:"service"`   // "api-gateway"
	Version   string `yaml:"version"`   // "0.1.0"
	AddSource bool   `yaml:"addSource"` // true/false
	Backend   string `yaml:"backend"`   // "std"|"zap"
	Debug     bool   `yaml:"debug"`     // включает подробные логи
}

type Config struct {
	HTTP     HTTP     `yaml:"http"`
	Logging  Logging  `yaml:"logging"`
	Upstream Upstream `yaml:"upstream"`
}

func Load() (*Config, error) {
	path := os.Getenv("CONFIG_PATH")
	if path == "" {
		path = filepath.Join("internal", "config", "config.yaml")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal yaml: %w", err)
	}

	if cfg.HTTP.Addr == "" {
		cfg.HTTP.Addr = ":8080"
	}
	if cfg.Upstream.RoomGRPCTarget == "" {
		cfg.Upstream.RoomGRPCTarget = "http://localhost:9092"
	}
	if cfg.HTTP.ReadTimeout == 0 {
		cfg.HTTP.ReadTimeout = 15 * time.Second
	}
	if cfg.HTTP.WriteTimeout == 0 {
		cfg.HTTP.WriteTimeout = 30 * time.Second
	}
	if cfg.HTTP.IdleTimeout == 0 {
		cfg.HTTP.IdleTimeout = 60 * time.Second
	}

	return &cfg, nil
}
