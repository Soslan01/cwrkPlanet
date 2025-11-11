package grpcsrv

import (
	"context"
	"encoding/json"
	"log/slog"
	"net"
	"strings"
	"time"

	authv1 "github.com/cwrk-planet/auth-service/proto/gen/auth/v1"

	"github.com/cwrk-planet/auth-service/internal/service"
	handler "github.com/cwrk-planet/auth-service/internal/transport/http"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

const requestIDKey = "x-request-id"

type Server struct {
	addr string
	gs   *grpc.Server
	ln   net.Listener
}

func New(addr string, svc *service.AuthService) (*Server, error) {
	gs := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			recoveryUnaryInterceptor(),
			requestIDInterceptor(),
			loggingUnaryInterceptor(),
		),
	)

	authv1.RegisterAuthServiceServer(gs, handler.NewAuthHandler(svc))

	return &Server{
		addr: addr,
		gs:   gs,
	}, nil
}

func (s *Server) Start(ctx context.Context) error {
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}

	s.ln = ln
	slog.Info("grpc listening", "addr", s.addr)
	go func() {
		if err := s.gs.Serve(s.ln); err != nil {
			slog.Error("grpc serve stopped", slog.Any("err", err))
		}
	}()

	return nil
}

func (s *Server) Stop(ctx context.Context) {
	done := make(chan struct{})
	go func() {
		s.gs.GracefulStop()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		slog.Error("grpc graceful stop timeout; forcing stop")
		s.gs.Stop()
	}

	if s.ln != nil {
		_ = s.ln.Close()
	}

	slog.Info("grpc stopped")
}

/*
	Interceptors
*/

func recoveryUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp any, err error) {
		defer func() {
			if r := recover(); r != nil {
				slog.Error("panic recovered", "method", info.FullMethod, "panic", r)
				err = status.Error(13 /* codes.Internal */, "internal error")
			}
		}()

		return handler(ctx, req)
	}
}

func loggingUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp any, err error) {
		start := time.Now()

		// request context
		ua, ip := extractUAAndIP(ctx)
		reqJSON := marshalRedacted(req)
		reqID, _ := RequestIDFromContext(ctx)

		slog.Info(
			"grpc request",
			"req_id", reqID,
			"method", info.FullMethod,
			"user_agent", ua,
			"ip", ip,
			"req", clip(reqJSON, 2048),
		)

		resp, err = handler(ctx, req)
		code := status.Code(err)

		// response log
		respJSON := marshalRedacted(resp)
		fields := []any{
			"req_id", reqID,
			"method", info.FullMethod,
			"code", code.String(),
			"duration", time.Since(start),
			"resp", clip(respJSON, 2048),
		}
		if err != nil {
			fields = append(fields, slog.Any("err", err))
			slog.Error("grpc response", fields...)
		} else {
			slog.Info("grpc response", fields...)
		}

		return resp, err
	}
}

func requestIDInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		md, _ := metadata.FromIncomingContext(ctx)
		var reqID string
		if vals := md.Get(requestIDKey); len(vals) > 0 {
			reqID = vals[0]
		}
		if reqID != "" {
			ctx = context.WithValue(ctx, requestIDKey, reqID)
		}

		slog.Debug("incoming gRPC call", "method", info.FullMethod, "req_id", reqID)

		return handler(ctx, req)
	}
}

/*
	Helpers
*/

var redactedKeys = map[string]struct{}{
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

// marshalRedacted JSON с редактированием чувствительных полей
func marshalRedacted(v any) string {
	if v == nil {
		return ""
	}

	if m, ok := v.(proto.Message); ok {
		b, err := protojson.MarshalOptions{
			EmitUnpopulated: false,
			UseEnumNumbers:  false,
			UseProtoNames:   true,
		}.Marshal(m)
		if err == nil {
			return redactJSONBytes(b)
		}
	}

	// Фоллбэк, если unluck - возвращаем стандартный json
	if b, err := json.Marshal(v); err == nil {
		return redactJSONBytes(b)
	}

	return "<unmarshallable>"
}

// redactJSONBytes парсит в generic map/array и рекурсивно затирает значения по ключам
func redactJSONBytes(b []byte) string {
	var data any
	if err := json.Unmarshal(b, &data); err != nil {
		// если не получится распарсить, то лучше показать сырое но обрезанное
		return string(b)
	}
	redactWalk(&data)
	out, err := json.Marshal(data)
	if err != nil {
		return string(b)
	}

	return string(out)
}

func redactWalk(v *any) {
	switch t := (*v).(type) {
	case map[string]any:
		for k, val := range t {
			if _, hit := redactedKeys[strings.ToLower(k)]; hit {
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

func extractUAAndIP(ctx context.Context) (ua string, ip string) {
	// user-agent и прокси-заголовки из метаданных
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if v := firstNonEmpty(md, "user-agent"); v != "" {
			ua = v
		}
		if v := firstNonEmpty(md, "x-forwarded-for"); v != "" {
			ip = firstIPFromXFF(v)
		}
		if ip == "" {
			if v := firstNonEmpty(md, "x-real-ip"); v != "" {
				ip = v
			}
		}
	}

	if ip == "" {
		if p, ok := peer.FromContext(ctx); ok && p != nil && p.Addr != nil {
			host, _, _ := net.SplitHostPort(p.Addr.String())
			if host != "" {
				ip = host
			}
		}
	}

	return
}

func RequestIDFromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(requestIDKey).(string)
	return v, ok
}

func firstNonEmpty(md metadata.MD, key string) string {
	if vals := md.Get(key); len(vals) > 0 {
		s := strings.TrimSpace(vals[0])
		if s != "" {
			return s
		}
	}

	return ""
}

func firstIPFromXFF(xff string) string {
	parts := strings.Split(xff, ",")
	if len(parts) == 0 {
		return ""
	}

	return strings.TrimSpace(parts[0])
}
