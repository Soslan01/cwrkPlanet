package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cwrk-planet/auth-service/internal/config"
	"github.com/cwrk-planet/auth-service/internal/pg"
	"github.com/cwrk-planet/auth-service/internal/repository/postgres"
	"github.com/cwrk-planet/auth-service/internal/security"
	"github.com/cwrk-planet/auth-service/internal/service"
	grpcsrv "github.com/cwrk-planet/auth-service/internal/transport/grpc"
	handler "github.com/cwrk-planet/auth-service/internal/transport/http"
	"github.com/cwrk-planet/logger/pkg/logger"
)

func main() {
	// Config init
	cfg, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		println("failed to load config:", err.Error())
		os.Exit(1)
	}
	if err := cfg.Validate(); err != nil {
		println("failed to validate config:", err.Error())
		os.Exit(1)
	}

	// Logger init
	logger.Init(logger.Config{
		Env:       logger.Env(cfg.Logging.Env),
		Service:   cfg.Logging.Service,
		Version:   cfg.Logging.Version,
		Backend:   logger.Backend(cfg.Logging.Backend),
		AddSource: cfg.Logging.AddSource,
		Debug:     cfg.Logging.Debug,
	})
	slog.Info("starting auth-service", "version", cfg.Logging.Version)

	// PostgreSQL init
	ctx := context.Background()
	pool, err := pg.NewPool(ctx, cfg.Postgres.ToPGConfig())
	if err != nil {
		slog.Error("failed to init postgres", slog.Any("err", err))
		os.Exit(1)
	}
	defer pool.Close()
	slog.Info("connected to postgres")

	// Services init
	usersRepo := postgres.NewUserRepoFromPool(pool)
	sessionsRepo := postgres.NewSessionRepoFromPool(pool)

	passCfg := security.BcryptConfig{
		Cost:      cfg.Security.Password.BcryptCost,
		MinLength: cfg.Security.Password.MinLength,
	}

	private, err := security.LoadRSAPrivateKeyFromPEM(cfg.Security.JWT.PrivateKeyPath)
	if err != nil {
		slog.Error("failed to read private key", slog.Any("err", err))
		os.Exit(1)
	}
	public, err := security.LoadRSAPublicKeyFromPEM(cfg.Security.JWT.PublicKeyPath)
	if err != nil {
		slog.Error("failed to read public key", slog.Any("err", err))
		os.Exit(1)
	}

	jwtSigner := security.NewJWTSigner(
		private,
		public,
		cfg.Security.JWT.Issuer,
		cfg.Security.JWT.Audience,
		cfg.Security.JWT.AccessTTL,
		cfg.Security.JWT.ClockSkew,
	)

	authSvc := service.NewAuthService(
		usersRepo,
		sessionsRepo,
		jwtSigner,
		cfg.Security.JWT.AccessTTL,
		passCfg,
		time.Now,
	)

	// gRPC server init
	grpcServer, err := grpcsrv.New(cfg.Server.GRPCAddr, authSvc)
	if err != nil {
		slog.Error("failed to init grpc server", slog.Any("err", err))
		os.Exit(1)
	}
	if err := grpcServer.Start(ctx); err != nil {
		slog.Error("failed to start grpc", slog.Any("err", err))
		os.Exit(1)
	}

	// HTTP gateway init
	httpServer, err := handler.New(cfg.Server.HTTPAddr, cfg.Server.GRPCAddr)
	if err != nil {
		slog.Error("failed to init http gateway", slog.Any("err", err))
		os.Exit(1)
	}
	if err := httpServer.Start(ctx); err != nil {
		slog.Error("failed to start http gateway", slog.Any("err", err))
		os.Exit(1)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigCh
	slog.Info("shutdown signal received", "signal", sig.String())

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	httpServer.Stop(shutdownCtx)
	grpcServer.Stop(shutdownCtx)
	pool.Close()

	slog.Info("auth-service stopped gracefully")
}
