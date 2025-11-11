package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cwrk-planet/api-gateway/internal/app/auth"
	"github.com/cwrk-planet/api-gateway/internal/app/room"
	"github.com/cwrk-planet/api-gateway/internal/config"
	httpserver "github.com/cwrk-planet/api-gateway/internal/server/http"
	transport "github.com/cwrk-planet/api-gateway/internal/transport/http"

	"github.com/cwrk-planet/logger/pkg/logger"
)

func main() {
	// 1) load config
	cfg, err := config.Load()
	if err != nil {
		println("failed to load config:", err.Error())
		os.Exit(1)
	}

	// 2) init logger (set.Default)
	logger.Init(logger.Config{
		Env:       logger.Env(cfg.Logging.Env),
		Service:   cfg.Logging.Service,
		Version:   cfg.Logging.Version,
		Backend:   logger.Backend(cfg.Logging.Backend),
		AddSource: cfg.Logging.AddSource,
		Debug:     cfg.Logging.Debug,
	})
	slog.Info("starting api-gateway", "version", cfg.Logging.Version)

	// 3) authClient init
	authClient, err := auth.New(auth.Options{
		Target:      cfg.Upstream.AuthTarget,
		Timeout:     5 * time.Second,
		UseInsecure: true, // для dev;
	})
	if err != nil {
		slog.Error("auth client init failed", "err", err)
		os.Exit(1)
	}
	defer func() { _ = authClient.Close() }()

	// 3.1) roomClient init
	roomClient, err := room.New(room.Options{
		Target:      cfg.Upstream.RoomGRPCTarget,
		Timeout:     5 * time.Second,
		UseInsecure: true,
	})
	if err != nil {
		slog.Error("room client init failed", "err", err)
		os.Exit(1)
	}
	defer func() { _ = roomClient.Close() }()

	// 4) router init
	router := transport.NewRouter(transport.Deps{
		AuthClient: authClient,
		RoomClient: roomClient,
	})

	// 5) server init
	srv := httpserver.New(httpserver.Config{
		Addr:         cfg.HTTP.Addr,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
		IdleTimeout:  cfg.HTTP.IdleTimeout,
	}, router)

	// 6) graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := srv.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		slog.Error("server stopped with error", "err", err)
		os.Exit(1)
	}

	slog.Info("api-gateway stopped")
}
