package tests

import (
	"encoding/json"
	"log/slog"
	"testing"

	"github.com/cwrk-planet/logger/pkg/logger"
)

func TestInit_ProdZap_JSONOutPut(t *testing.T) {
	cfg := logger.Config{
		Service:          "demo",
		Version:          "1.2.3",
		Env:              logger.EnvProd,
		Backend:          logger.BackendZap,
		Level:            slog.LevelInfo,
		SampleInitial:    100000,
		SampleThereafter: 100000,
		SampleTick:       1,
	}

	out := captureStdOut(func() {
		logger.Init(cfg)
		slog.Info("booted", slog.String("k", "v"))
	})

	var m map[string]any
	if err := json.Unmarshal([]byte(out), &m); err != nil {
		t.Fatalf("expected JSON line, got %s, err=%v", out, err)
	}

	if m["msg"] != "booted" {
		t.Fatalf("msg mismatch: %v", m["msg"])
	}
	if m["service"] != "demo" || m["env"] != "prod" || m["version"] != "1.2.3" {
		t.Fatalf("attrs missing: service=%v env=%v version=%v", m["service"], m["env"], m["version"])
	}
	if m["level"] != "INFO" {
		t.Fatalf("level mismatch: %v", m["level"])
	}
	if m["k"] != "v" {
		t.Fatalf("custom field missing: %v", m["k"])
	}
}
