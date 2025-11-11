package config

import (
	"errors"
	"os"
	"strings"
	"time"

	"github.com/cwrk-planet/auth-service/internal/pg"

	"gopkg.in/yaml.v3"
)

type Server struct {
	GRPCAddr        string        `yaml:"grpcAddr"`
	HTTPAddr        string        `yaml:"httpAddr"`
	ShutdownTimeout time.Duration `yaml:"shutdownTimeout"`
}

type Logging struct {
	Env       string `yaml:"env"`
	Service   string `yaml:"service"`
	Version   string `yaml:"version"`
	Backend   string `yaml:"backend"`
	AddSource bool   `yaml:"addSource"`
	Debug     bool   `yaml:"debug"`
}

type Postgres struct {
	DSN               string        `yaml:"dsn"`
	MaxConns          int32         `yaml:"maxConns"`
	MinConns          int32         `yaml:"minConns"`
	MaxConnLifetime   time.Duration `yaml:"maxConnLifetime"`
	MaxConnIdleTime   time.Duration `yaml:"maxConnIdleTime"`
	HealthCheckPeriod time.Duration `yaml:"healthCheckPeriod"`
	ApplicationName   string        `yaml:"applicationName"`
}

func (p Postgres) Validate() error {
	if p.DSN == "" {
		return errors.New("postgres.DSN is required")
	}

	return nil
}

func (p Postgres) ToPGConfig() pg.Config {
	return pg.Config{
		DSN:               p.DSN,
		MaxConns:          p.MaxConns,
		MinConns:          p.MinConns,
		MaxConnLifetime:   p.MaxConnLifetime,
		MaxConnIdleTime:   p.MaxConnIdleTime,
		HealthCheckPeriod: p.HealthCheckPeriod,
		ApplicationName:   p.ApplicationName,
	}
}

type Password struct {
	MinLength  int `yaml:"minLength"`
	BcryptCost int `yaml:"bcryptCost"`
}

func (p Password) Validate() error {
	if p.MinLength < 6 {
		return errors.New("security.password.minLength must be >= 6")
	}
	if p.BcryptCost != 0 && (p.BcryptCost < 4 || p.BcryptCost > 18) {
		return errors.New("security.password.bcryptCost must be in [4..18]")
	}

	return nil
}

type JWT struct {
	Alg            string        `yaml:"alg"`            // обязательно
	PrivateKeyPath string        `yaml:"privateKeyPath"` // обязательно
	PublicKeyPath  string        `yaml:"publicKeyPath"`  // обязательно
	Issuer         string        `yaml:"issuer"`         // обязательно
	Audience       string        `yaml:"audience"`       // по желанию, но пока особо не проверятся
	AccessTTL      time.Duration `yaml:"accessTTL"`      // напр. 15m
	ClockSkew      time.Duration `yaml:"clockSkew"`      // напр. 30s
}

func (j JWT) Validate() error {
	if j.Alg == "" {
		return errors.New("security.jwt.alg is required")
	}
	if j.PrivateKeyPath == "" {
		return errors.New("security.jwt.privateKeyPath is required")
	}
	if j.PublicKeyPath == "" {
		return errors.New("security.jwt.publicKeyPath is required")
	}
	if j.Issuer == "" {
		return errors.New("security.jwt.issuer is required")
	}
	if j.AccessTTL <= 0 {
		return errors.New("security.jwt.accessTTL must be > 0")
	}
	if j.ClockSkew < 0 || j.ClockSkew > time.Minute {
		return errors.New("security.jwt.clockSkew must be in [0..1m]")
	}

	return nil
}

type Security struct {
	Password Password `yaml:"password"`
	JWT      JWT      `yaml:"jwt"`
}

func (s Security) Validate() error {
	if err := s.Password.Validate(); err != nil {
		return err
	}
	if err := s.JWT.Validate(); err != nil {
		return err
	}

	return nil
}

type Config struct {
	Server   Server   `yaml:"server"`
	Security Security `yaml:"security"`
	Postgres Postgres `yaml:"postgres"`
	Logging  Logging  `yaml:"logging"`
}

func (c *Config) Validate() error {
	if err := c.Security.Validate(); err != nil {
		return err
	}
	if err := c.Postgres.Validate(); err != nil {
		return err
	}

	return nil
}

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
