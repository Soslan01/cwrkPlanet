package tests

import (
	"testing"

	"github.com/cwrk-planet/logger/pkg/logger"
)

func TestDetectEnv(t *testing.T) {
	t.Setenv("APP_ENV", "")
	if got := logger.DetectEnv(); got != logger.EnvDev {
		t.Fatalf("default should be dev, got %q", got)
	}

	t.Setenv("APP_ENV", "stage")
	if got := logger.DetectEnv(); got != logger.EnvStage {
		t.Fatalf("expected stage, got %q", got)
	}

	t.Setenv("APP_ENV", "prod")
	if got := logger.DetectEnv(); got != logger.EnvProd {
		t.Fatalf("expected prod, got %q", got)
	}
}
