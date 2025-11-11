package main

import (
	"context"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cwrk-planet/logger/pkg/logger"
	"github.com/cwrk-planet/room-service/config"
	"github.com/cwrk-planet/room-service/internal/postgres"
	"github.com/cwrk-planet/room-service/internal/service"
	grpcx "github.com/cwrk-planet/room-service/internal/transport/grpc"
	httpx "github.com/cwrk-planet/room-service/internal/transport/http"
	"github.com/cwrk-planet/room-service/internal/transport/ws"

	"google.golang.org/grpc"
)

func main() {
	// --- config ---
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	logger.Init(logger.Config{
		Env:       logger.Env(cfg.Logging.Env),
		Service:   cfg.Logging.Service,
		Version:   cfg.Logging.Version,
		Backend:   logger.Backend(cfg.Logging.Backend),
		AddSource: cfg.Logging.AddSource,
		Debug:     cfg.Logging.Debug,
	})
	slog.Info("starting room-service",
		"env", cfg.Logging.Env, "version", cfg.Logging.Version)

	// --- postgres ---
	ctx := context.Background()
	db, err := postgres.New(ctx, cfg.Postgres.DSN)
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}
	defer db.Close()

	// --- repos ---
	roomRepo := postgres.NewRoomRepository(db.Pool)
	partRepo := postgres.NewParticipantRepository(db.Pool)
	chatRepo := postgres.NewChatRepository(db.Pool)

	// --- services ---
	roomSvc := service.NewRoomService(roomRepo)
	memberSvc := service.NewMemberService(roomRepo, partRepo)
	chatSvc := service.NewChatService(chatRepo)

	// --- WS Hub & Server ---
	hub := ws.NewHub()
	wsServer := ws.NewServer(hub, memberSvc, chatSvc)

	// --- HTTP ---
	handler := httpx.NewHandler(roomSvc, memberSvc, chatSvc)
	router := httpx.NewRouter(handler, memberSvc, wsServer)
	httpSrv := &http.Server{
		Addr:         cfg.HTTP.Addr,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// --- gRPC ---
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(grpcx.UnaryServerInterceptor()),
		grpc.ChainStreamInterceptor(grpcx.StreamServerInterceptor()),
	)
	grpcSrv := grpcx.NewServer(roomSvc, memberSvc, chatSvc)
	grpcx.Register(grpcServer, grpcSrv)

	// --- run both servers ---
	errCh := make(chan error, 2)

	go func() {
		slog.Info("http listen", "addr", cfg.HTTP.Addr)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	go func() {
		lis, err := net.Listen("tcp", cfg.GRPC.Addr)
		if err != nil {
			errCh <- err
			return
		}
		slog.Info("grpc listen", "addr", cfg.GRPC.Addr)
		if err := grpcServer.Serve(lis); err != nil {
			errCh <- err
		}
	}()

	// --- graceful shutdown ---
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigCh:
		slog.Info("shutdown signal", "sig", sig)
	case err := <-errCh:
		slog.Error("server error", "err", err)
	}

	ctxShutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	grpcServer.GracefulStop()
	_ = httpSrv.Shutdown(ctxShutdown)
	slog.Info("stopped")
}
