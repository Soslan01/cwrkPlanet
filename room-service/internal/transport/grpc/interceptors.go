package grpcx

import (
	"context"
	"log/slog"
	"runtime/debug"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Unary logging + recovery + timeout guard (если у вызова нет deadline)
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp any, err error) {
		start := time.Now()
		// deadline guard
		if _, ok := ctx.Deadline(); !ok {
			var cancel context.CancelFunc
			// дефолтный guard — 10 секунд
			ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
			defer cancel()
		}

		defer func() {
			if r := recover(); r != nil {
				slog.Error("grpc unary panic",
					"method", info.FullMethod,
					"panic", r,
					"stack", string(debug.Stack()))
				err = status.Error(codes.Internal, "internal server error")
			}
			slog.Info("grpc unary",
				"method", info.FullMethod,
				"dur_ms", time.Since(start).Milliseconds(),
				"err", errString(err))
		}()

		return handler(ctx, req)
	}
}

func StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(
		srv any,
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) (err error) {
		start := time.Now()

		defer func() {
			if r := recover(); r != nil {
				slog.Error("grpc stream panic",
					"method", info.FullMethod,
					"panic", r,
					"stack", string(debug.Stack()))
				err = status.Error(codes.Internal, "internal server error")
			}
			slog.Info("grpc stream",
				"method", info.FullMethod,
				"dur_ms", time.Since(start).Milliseconds(),
				"err", errString(err))
		}()

		return handler(srv, ss)
	}
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
