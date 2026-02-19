package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	authclient "github.com/cwrk-planet/api-gateway/internal/client/auth"
	"github.com/cwrk-planet/api-gateway/internal/config"
	httphandler "github.com/cwrk-planet/api-gateway/internal/handler/http"
	"github.com/cwrk-planet/api-gateway/internal/logger"
	authservice "github.com/cwrk-planet/api-gateway/internal/service/auth"
	httpserver "github.com/cwrk-planet/api-gateway/internal/transport/http"
	"go.uber.org/zap"
)

func main() {
	// Load configuration
	// CONFIG_PATH set in Docker. When unset, use local config for development.
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config/config.yaml"
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

	appLogger.Info("Starting api-gateway")

	// Initialize gRPC client for auth-service
	authClient, err := authclient.NewClient(
		cfg.Clients.Auth.Address,
		cfg.Clients.Auth.Timeout,
		appLogger.Logger,
	)
	if err != nil {
		appLogger.Fatal("Failed to create auth client", zap.Error(err))
	}
	defer authClient.Close()

	appLogger.Info("Connected to auth-service", zap.String("address", cfg.Clients.Auth.Address))

	// Initialize services
	authService := authservice.NewService(authClient, appLogger.Logger)

	// Initialize handlers
	authHandler := httphandler.NewAuthHandler(authService, appLogger.Logger)

	// Initialize HTTP server
	httpServer := httpserver.NewServer(
		cfg.Server.HTTPAddress,
		authHandler,
		appLogger.Logger,
	)

	// Start server in goroutine
	errChan := make(chan error, 1)
	go func() {
		if err := httpServer.Start(); err != nil && err != context.Canceled {
			errChan <- err
		}
	}()

	appLogger.Info("API Gateway started", zap.String("address", cfg.Server.HTTPAddress))

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-errChan:
		appLogger.Fatal("HTTP server error", zap.Error(err))
	case sig := <-sigChan:
		appLogger.Info("Received signal, shutting down", zap.String("signal", sig.String()))

		// Graceful shutdown
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
		defer cancel()

		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			appLogger.Error("Error during shutdown", zap.Error(err))
		} else {
			appLogger.Info("Server stopped gracefully")
		}
	}
}
