package tests

import (
	"context"
	"encoding/json"
	"log/slog"
	"testing"

	"github.com/cwrk-planet/logger/pkg/logger"

	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/zap"
)

func TestAttrsFromCtx_PropagatesTraceIDs(t *testing.T) {
	logger.Init(logger.Config{
		Service:          "demo",
		Env:              logger.EnvProd,
		Backend:          logger.BackendZap,
		SampleInitial:    100000,
		SampleThereafter: 100000,
		SampleTick:       1,
	})

	tp := sdktrace.NewTracerProvider()
	defer func() {
		_ = tp.Shutdown(context.Background())
	}()
	otel.SetTracerProvider((tp))
	tr := tp.Tracer("test")

	var outStr string
	func() {
		ctx, span := tr.Start(context.Background(), "op")
		defer span.End()

		outStr = captureStdOut(func() {
			logger.Init(logger.Config{
				Service:          "demo",
				Env:              logger.EnvProd,
				Backend:          logger.BackendZap,
				SampleInitial:    100000,
				SampleThereafter: 100000,
				SampleTick:       1,
			})

			slog.InfoContext(ctx, "with trace", toAttrsFromCtx(ctx)...)
		})
	}()

	// Flush the logger to ensure all logs are written before we parse the output
	if err := zap.L().Sync(); err != nil {
		t.Fatalf("failed to flush logs: %v", err)
	}

	var m map[string]any
	if err := json.Unmarshal([]byte(outStr), &m); err != nil {
		t.Fatalf("expected JSON, got: %s, err=%v", outStr, err)
	}

	if m["trace_id"] == nil || m["span_id"] == nil {
		t.Fatalf("trace_id/span_id missing in log: %v", m)
	}
	if m["msg"] != "with trace" {
		t.Fatalf("msg mismatch: %v", m["msg"])
	}

}
