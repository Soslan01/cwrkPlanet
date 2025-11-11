package logger

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/cwrk-planet/logger/pkg/logger"
	"github.com/go-chi/chi/v5/middleware"
)

type statusWriter struct {
	http.ResponseWriter
	status int
	bytes  int64
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	n, err := w.ResponseWriter.Write(b)
	w.bytes += int64(n)
	return n, err
}

func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w}

		next.ServeHTTP(sw, r)

		level := slog.LevelInfo
		switch {
		case sw.status >= 500:
			level = slog.LevelError
		case sw.status >= 400:
			level = slog.LevelWarn
		default:
			level = slog.LevelInfo
		}

		logger := L(r.Context())

		logger.LogAttrs(
			r.Context(),
			level,
			"http_request",
			slog.Int("status", sw.status),
			slog.Int("bytes", int(sw.bytes)),
			slog.Duration("duration", time.Since(start)),
			slog.String("remote_ip", r.RemoteAddr),
			slog.String("user_agent", r.UserAgent()),
			slog.String("query", r.URL.RawQuery),
		)
	})
}

type ctxKey int

const loggerKey ctxKey = iota

// WithRequestLoggerCtx кладёт *slog.Logger в контекст
func WithRequestLoggerCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := middleware.GetReqID(r.Context())
		l := logger.L().With(
			slog.String("req_id", reqID),
			slog.String("path", r.URL.Path),
			slog.String("method", r.Method),
		)
		ctx := context.WithValue(r.Context(), loggerKey, l)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// L извлекает логгер из контекста, а если его нет — возвращает глобальный
func L(ctx context.Context) *slog.Logger {
	if v := ctx.Value(loggerKey); v != nil {
		if l, ok := v.(*slog.Logger); ok && l != nil {
			return l
		}
	}
	return logger.L()
}
