package logger

import (
	"os"
	"strings"
)

type Env string

const (
	EnvDev   Env = "dev"
	EnvStage Env = "stage"
	EnvProd  Env = "prod"
)

func DetectEnv() Env {
	raw := strings.ToLower(strings.TrimSpace(os.Getenv("APP_ENV")))

	switch raw {
	case "prod", "production":
		return EnvProd
	case "stage", "staging", " preprod", "pre-production":
		return EnvStage
	default:
		return EnvDev
	}
}
