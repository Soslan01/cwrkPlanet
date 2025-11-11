package logger

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/trace"
)

func AttrsFromCtx(ctx context.Context) []slog.Attr {
	span := trace.SpanFromContext(ctx)
	sc := span.SpanContext()

	if !sc.IsValid() {
		return nil
	}

	return []slog.Attr{
		slog.String("trace_id", sc.TraceID().String()),
		slog.String("span_id", sc.SpanID().String()),
	}
}
