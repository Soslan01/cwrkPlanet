package config

import (
	"errors"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Server configuration
type Server struct {
	HTTPAddress     string        `yaml:"http_address"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
}

func (s Server) Validate() error {
	if s.HTTPAddress == "" {
		return errors.New("http_address is required")
	}
	return nil
}

type Logger struct {
	Env       string `yaml:"env"`
	Service   string `yaml:"service"`
	Version   string `yaml:"version"`
	Backend   string `yaml:"backend"`
	AddSource bool   `yaml:"add_source"`
	Debug     bool   `yaml:"debug"`
}

func (lg Logger) Validate() error {
	if lg.Env == "" {
		return errors.New("env is required")
	}

	if lg.Service == "" {
		return errors.New("service is required")
	}

	if lg.Version == "" {
		return errors.New("version is required")
	}

	// Backend, AddSource, and Debug are optional fields
	return nil
}

// AuthClient configuration for auth-service gRPC client
type AuthClient struct {
	Address string        `yaml:"address"`
	Timeout time.Duration `yaml:"timeout"`
}

func (ac AuthClient) Validate() error {
	if ac.Address == "" {
		return errors.New("auth_client.address is required")
	}
	return nil
}

// Clients configuration
type Clients struct {
	Auth AuthClient `yaml:"auth"`
}

func (c Clients) Validate() error {
	return c.Auth.Validate()
}

// Config is the main configuration structure
type Config struct {
	Server  Server  `yaml:"server"`
	Logging Logger  `yaml:"logging"`
	Clients Clients `yaml:"clients"`
}

func (c *Config) Validate() error {
	if err := c.Server.Validate(); err != nil {
		return err
	}
	if err := c.Logging.Validate(); err != nil {
		return err
	}
	if err := c.Clients.Validate(); err != nil {
		return err
	}
	return nil
}

// LoadConfig loads configuration from YAML file
func LoadConfig(path ...string) (*Config, error) {
	filename := "config/config.yaml"
	if len(path) > 0 && strings.TrimSpace(path[0]) != "" {
		filename = path[0]
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
