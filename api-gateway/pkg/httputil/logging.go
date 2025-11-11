package httputil

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

// Логирует метод, путь, статус, длительность, тела запрос/ответ и X-Request-ID.
func MiddlewareLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		var reqBody string
		if strings.Contains(strings.ToLower(r.Header.Get("Content-Type")), "json") && r.Body != nil {
			var buf bytes.Buffer
			tee := io.TeeReader(r.Body, &buf)
			b, _ := io.ReadAll(tee)
			r.Body = io.NopCloser(&buf)
			reqBody = string(b)
		}

		lrw := &logResponseWriter{ResponseWriter: w}
		next.ServeHTTP(lrw, r)

		dur := time.Since(start)
		reqID, _ := FromContext(r.Context())

		slog.Info("http request",
			"req_id", reqID,
			"method", r.Method,
			"path", r.URL.Path,
			"status", lrw.status,
			"bytes", lrw.bytes,
			"duration", dur.String(),
			"req_body", reqBody,
			"resp_body", lrw.body.String(),
		)
	})
}

type logResponseWriter struct {
	http.ResponseWriter
	status int
	bytes  int
	body   bytes.Buffer
}

func (w *logResponseWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *logResponseWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	w.body.Write(b)
	n, err := w.ResponseWriter.Write(b)
	w.bytes += n

	return n, err
}
