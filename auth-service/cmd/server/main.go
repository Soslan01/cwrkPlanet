package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cwrk-planet/auth-service/internal/config"
	"github.com/cwrk-planet/auth-service/internal/database"
	"github.com/cwrk-planet/auth-service/internal/handler"
	"github.com/cwrk-planet/auth-service/internal/logger"
	"github.com/cwrk-planet/auth-service/internal/repository"
	"github.com/cwrk-planet/auth-service/internal/service"
	"github.com/cwrk-planet/auth-service/proto/auth"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
)

func main() {
	// Load configuration
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config/config.local.yaml"
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "Invalid config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	appLogger, err := logger.New(
		cfg.Logging.Service,
		cfg.Logging.Version,
		cfg.Logging.Env,
		cfg.Logging.Debug,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer appLogger.Sync()

	appLogger.Info("Starting auth-service")

	// Connect to database
	ctx := context.Background()
	db, err := database.New(ctx, cfg.Postgres)
	if err != nil {
		appLogger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Pool.Close()

	appLogger.Info("Connected to database")

	// Run database migrations on startup
	if err := repository.RunMigrations(ctx, db, appLogger); err != nil {
		appLogger.Fatal("Failed to run database migrations", zap.Error(err))
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewSessionRepository(db)

	// Initialize services
	passwordService := service.NewPasswordService(cfg.Security.Password)
	tokenService, err := service.NewTokenService(cfg.Security.JWT)
	if err != nil {
		appLogger.Fatal("Failed to initialize token service", zap.Error(err))
	}

	refreshTTL := 7 * 24 * time.Hour // 7 days default, should be configurable
	authService := service.NewAuthService(
		userRepo,
		sessionRepo,
		passwordService,
		tokenService,
		refreshTTL,
	)

	// Initialize gRPC handler
	authHandler := handler.NewAuthHandler(authService)

	// Create gRPC server with keepalive enforcement
	grpcServer := grpc.NewServer(
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             30 * time.Second, // Minimum time between pings from client
			PermitWithoutStream: true,
		}),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:    30 * time.Second, // Ping if client is idle for this long
			Timeout: 5 * time.Second,  // Wait this long for ping ack before closing
		}),
	)
	auth.RegisterAuthServiceServer(grpcServer, authHandler)

	// Enable reflection for development/testing
	reflection.Register(grpcServer)

	// Start gRPC server
	listener, err := net.Listen("tcp", cfg.Server.GRPCAdress)
	if err != nil {
		appLogger.Fatal("Failed to listen", zap.Error(err), zap.String("address", cfg.Server.GRPCAdress))
	}

	appLogger.Info("Starting gRPC server", zap.String("address", cfg.Server.GRPCAdress))

	// Start server in goroutine
	errChan := make(chan error, 1)
	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			errChan <- err
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-errChan:
		appLogger.Fatal("gRPC server error", zap.Error(err))
	case sig := <-sigChan:
		appLogger.Info("Received signal, shutting down", zap.String("signal", sig.String()))

		// Graceful shutdown
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
		defer cancel()

		stopped := make(chan struct{})
		go func() {
			grpcServer.GracefulStop()
			close(stopped)
		}()

		select {
		case <-stopped:
			appLogger.Info("Server stopped gracefully")
		case <-shutdownCtx.Done():
			appLogger.Warn("Shutdown timeout exceeded, forcing stop")
			grpcServer.Stop()
		}
	}
}
