package config

import (
	"errors"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// HTTP / GRPC adresses
type Server struct {
	GRPCAdress      string        `yaml:"grpc_adress"`
	HTTPAdress      string        `yaml:"http_adress"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
}

// Validate conf.Server block
func (s Server) Validate() error {
	if s.GRPCAdress == "" {
		return errors.New("GRPCAdress is required")
	}

	if s.HTTPAdress == "" {
		return errors.New("HTTPAdress is required")
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

// Validate conf.Logger block
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

type Postgres struct {
	DSN               string        `yaml:"dsn"`
	MaxConns          int32         `yaml:"max_conns"`
	MinConns          int32         `yaml:"min_conns"`
	MaxConnLifeTime   time.Duration `yaml:"max_conn_lifetime"`
	MaxConnIdleTime   time.Duration `yaml:"max_conn_idletime"`
	HealthCheckPeriod time.Duration `yaml:"healthcheck_period"`
	ApplicationName   string        `yaml:"application_name"`
}

func (pg Postgres) Validate() error {
	if pg.DSN == "" {
		return errors.New("postgres DSN is required")
	}

	return nil
}

type Password struct {
	MinLength  int `yaml:"min_length"`
	BcryptCost int `yaml:"bcrypt_cost"` // сложность генерации хэша для пароля
}

func (pass Password) Validate() error {
	if pass.MinLength < 6 {
		return errors.New("security.password.min_length must be >= 6")
	}

	if pass.BcryptCost != 0 && (pass.BcryptCost < 4 || pass.BcryptCost > 18) {
		return errors.New("security.password.bcrypt_cost must be in [4...18]")
	}

	return nil
}

type JWT struct {
	Alg            string        `yaml:"alg"`
	PrivateKeyPath string        `yaml:"private_key_path"`
	PublicKeyPath  string        `yaml:"public_key_path"`
	Issuer         string        `yaml:"issuer"`
	Audience       string        `yaml:"audience"`
	AccessTTL      time.Duration `yaml:"access_ttl"`
	ClockSkew      time.Duration `yaml:"clock_skew"`
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
	Logging  Logger   `yaml:"logging"`
}

func (c *Config) Validate() error {
	if err := c.Server.Validate(); err != nil {
		return err
	}
	if err := c.Logging.Validate(); err != nil {
		return err
	}
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
