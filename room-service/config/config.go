package config

import (
	"errors"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type GRPC struct {
	Addr string `yaml:"addr"`
}

type HTTP struct {
	Addr string `yaml:"addr"`
}

type Logging struct {
	Env       string `yaml:"env"`       // dev|prod
	Service   string `yaml:"service"`   // room-service
	Version   string `yaml:"version"`   // v0.1.0
	Backend   string `yaml:"backend"`   // std|zap
	AddSource bool   `yaml:"addSource"` // false|true
	Debug     bool   `yaml:"debug"`     // false|true
}

type Postgres struct {
	DSN string `yaml:"dsn"`
}

type Config struct {
	HTTP     HTTP     `yaml:"http"`
	GRPC     GRPC     `yaml:"grpc"`
	Logging  Logging  `yaml:"logging"`
	Postgres Postgres `yaml:"postgres"`
}

func LoadConfig() (*Config, error) {
	path := os.Getenv("CONFIG_PATH")
	if path == "" {
		path = "./config/config.yaml"
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *Config) validate() error {
	if c.HTTP.Addr == "" {
		return errors.New("http.addr is required")
	}
	if c.GRPC.Addr == "" {
		return errors.New("grpc.addr is required")
	}
	if c.Postgres.DSN == "" {
		return errors.New("postgres.dsn is required")
	}
	// установка дефолтов, если значения не указаны
	if c.Logging.Service == "" {
		c.Logging.Service = "room-service"
	}
	if c.Logging.Env == "" {
		c.Logging.Env = "dev"
	}
	if c.Logging.Version == "" {
		c.Logging.Version = "v0.1.0"
	}
	if c.Logging.Backend == "" {
		c.Logging.Backend = "std"
	}
	return nil
}

// helper для парсинга timeout-ов
func parseDurationOr(def time.Duration, s string) time.Duration {
	if d, err := time.ParseDuration(s); err == nil && d > 0 {
		return d
	}
	return def
}
