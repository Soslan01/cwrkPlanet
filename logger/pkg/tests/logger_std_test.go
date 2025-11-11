package tests

import (
	"log/slog"
	"strings"
	"testing"

	"github.com/cwrk-planet/logger/pkg/logger"
)

func TestInit_DevStd_TextOutPut(t *testing.T) {
	cfg := logger.Config{
		Service:   "demo",
		Version:   "v0.0.1",
		Env:       logger.EnvDev,
		Backend:   logger.BackendStd,
		Level:     slog.LevelDebug,
		AddSource: true,
	}

	out := captureStdOut(func() {
		logger.Init(cfg)
		slog.Info("Hello world")
	})

	// Txt handler
	if strings.Contains(out, "{") && strings.Contains(out, "}") {
		t.Fatalf("expected text output in dev/std, got JSON: %s", out)
	}
	if !strings.Contains(out, "Hello world") {
		t.Fatalf("message missing: %s", out)
	}
	if !strings.Contains(out, "service=demo") {
		t.Fatalf("service attr missing: %s", out)
	}
	if !strings.Contains(out, "env=dev") {
		t.Fatalf("env attr missing: %s", out)
	}
}
