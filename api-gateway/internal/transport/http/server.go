package http

import (
	"context"
	"net/http"
	"time"

	httphandler "github.com/cwrk-planet/api-gateway/internal/handler/http"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"go.uber.org/zap"
)

type Server struct {
	router     *chi.Mux
	httpServer *http.Server
	logger     *zap.Logger
}

func NewServer(
	address string,
	authHandler *httphandler.AuthHandler,
	logger *zap.Logger,
) *Server {
	router := chi.NewRouter()

	// Middleware
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Recoverer)
	router.Use(httphandler.LoggingMiddleware(logger))
	
	// CORS
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health check
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Register routes
	authHandler.RegisterRoutes(router)

	httpServer := &http.Server{
		Addr:         address,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &Server{
		router:     router,
		httpServer: httpServer,
		logger:     logger,
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	s.logger.Info("Starting HTTP server", zap.String("address", s.httpServer.Addr))
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully shuts down the HTTP server
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down HTTP server")
	return s.httpServer.Shutdown(ctx)
}
