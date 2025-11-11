package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

	authv1 "github.com/cwrk-planet/auth-service/proto/gen/auth/v1"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"
)

type Server struct {
	httpAddr string
	grpcAddr string
	srv      *http.Server
	ln       net.Listener
	conn     *grpc.ClientConn
	cancel   context.CancelFunc
}

func New(httpAddr, grpcAddr string) (*Server, error) {
	mux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				EmitUnpopulated: false,
				UseEnumNumbers:  false,
				UseProtoNames:   true,
			},
			UnmarshalOptions: protojson.UnmarshalOptions{
				DiscardUnknown: true,
			},
		}),
	)

	base := context.Background()
	ctx, cancel := context.WithCancel(base)

	// Подключение к gRPC эндпоинту
	clientOps := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()), // TODO: TLS на проде
		grpc.WithDefaultCallOptions(),
	}
	conn, err := grpc.NewClient(grpcAddr, clientOps...)
	if err != nil {
		cancel()
		return nil, err
	}

	if err := authv1.RegisterAuthServiceHandler(ctx, mux, conn); err != nil {
		_ = conn.Close()
		cancel()
		return nil, err
	}

	root := chain(
		mux,
		requestLogger(),
		cors(),
	)

	root = withExtraRoutes(root, func(m *http.ServeMux) {
		m.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
		})
		m.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ready"))
		})
	})

	s := &http.Server{
		Addr:         httpAddr,
		Handler:      root,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &Server{
		httpAddr: httpAddr,
		grpcAddr: grpcAddr,
		srv:      s,
		conn:     conn,
		cancel:   cancel,
	}, nil
}

func (s *Server) Start(ctx context.Context) error {
	ln, err := net.Listen("tcp", s.httpAddr)
	if err != nil {
		return err
	}
	s.ln = ln

	slog.Info("http listening", "addr", s.httpAddr, "grpc_upstream", s.grpcAddr)
	go func() {
		if err := s.srv.Serve(s.ln); !errors.Is(err, http.ErrServerClosed) && err != nil {
			slog.Error("http serve stopped", slog.Any("err", err))
		}
	}()

	return nil
}

// Stop — graceful shutdown с таймаутом 15s.
func (s *Server) Stop(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	if err := s.srv.Shutdown(ctx); err != nil {
		slog.Error("http shutdown error", slog.Any("err", err))
	}
	if s.ln != nil {
		_ = s.ln.Close()
	}

	if s.conn != nil {
		_ = s.conn.Close()
	}
	if s.cancel != nil {
		s.cancel()
	}

	slog.Info("http stopped")
}

// -------------------- Middleware & Helpers --------------------

type middleware func(http.Handler) http.Handler

func chain(h http.Handler, m ...middleware) http.Handler {
	for i := len(m) - 1; i >= 0; i-- {
		h = m[i](h)
	}
	return h
}

// requestLogger — подробное логирование HTTP запросов/ответов, с редактированием чувствительных полей.
func requestLogger() middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ip := clientIP(r)
			ua := r.UserAgent()
			reqDump := safeReadBodyOnce(r)

			slog.Info("http request",
				"method", r.Method,
				"path", r.URL.Path,
				"query", r.URL.RawQuery,
				"ip", ip,
				"user_agent", ua,
				"req", clip(redactJSON([]byte(reqDump)), 2048),
			)

			lrw := &loggingResponseWriter{ResponseWriter: w, status: 200}
			next.ServeHTTP(lrw, r)

			slog.Info("http response",
				"method", r.Method,
				"path", r.URL.Path,
				"status", lrw.status,
				"bytes", lrw.written,
				"duration", time.Since(start),
			)
		})
	}
}

// cors — простая CORS-политика для разработки, а именно - Разрешить всё L:)
func cors() middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Requested-With")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func withExtraRoutes(base http.Handler, mount func(m *http.ServeMux)) http.Handler {
	m := http.NewServeMux()
	mount(m)
	m.Handle("/", base)

	return m
}

type loggingResponseWriter struct {
	http.ResponseWriter
	status  int
	written int64
}

func (w *loggingResponseWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *loggingResponseWriter) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	w.written += int64(n)

	return n, err
}

// --- Redaction / Body reading ---

var redactKeys = map[string]struct{}{
	"password":      {},
	"password_hash": {},
	"refresh":       {},
	"refresh_token": {},
	"access":        {},
	"access_token":  {},
	"token":         {},
	"jwt":           {},
	"authorization": {},
}

func redactJSON(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	var v any
	if err := json.Unmarshal(b, &v); err != nil {
		return string(b)
	}
	redactWalk(&v)
	out, err := json.Marshal(v)
	if err != nil {
		return string(b)
	}

	return string(out)
}

func redactWalk(v *any) {
	switch t := (*v).(type) {
	case map[string]any:
		for k, val := range t {
			if _, ok := redactKeys[strings.ToLower(k)]; ok {
				t[k] = "***REDACTED***"
				continue
			}
			redactWalk(&val)
			t[k] = val
		}
	case []any:
		for i := range t {
			redactWalk(&t[i])
		}
	}
}

func clip(s string, n int) string {
	if n <= 0 || len(s) <= n {
		return s
	}

	return s[:n] + "...(truncated)"
}

func clientIP(r *http.Request) string {
	if xff := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	if xr := strings.TrimSpace(r.Header.Get("X-Real-IP")); xr != "" {
		return xr
	}
	host, _, _ := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if host != "" {
		return host
	}

	return ""
}

func safeReadBodyOnce(r *http.Request) string {
	if r.Body == nil {
		return ""
	}
	ct := strings.ToLower(r.Header.Get("Content-Type"))
	if !strings.Contains(ct, "application/json") {
		return ""
	}
	const max = 1 << 20 // 1 MiB
	limited := http.MaxBytesReader(nil, r.Body, max)
	b, err := io.ReadAll(limited)
	if err != nil && !errors.Is(err, io.EOF) {
		return ""
	}
	r.Body = io.NopCloser(bytes.NewReader(b))

	return string(b)
}
