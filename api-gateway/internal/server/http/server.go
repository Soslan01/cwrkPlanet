package http

import (
	"context"
	"net/http"
	"time"
)

type Config struct {
	Addr         string        // ":8080"
	ReadTimeout  time.Duration // 15s
	WriteTimeout time.Duration // 30s
	IdleTimeout  time.Duration // 60s
}

type Server struct {
	cfg Config
	srv *http.Server
}

func New(cfg Config, handler http.Handler) *Server {
	s := &http.Server{
		Addr:         cfg.Addr,
		Handler:      handler,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}
	return &Server{
		cfg: cfg,
		srv: s,
	}
}

// Run запускает HTTP-сервер и блокирует до завершения ctx.
func (s *Server) Run(ctx context.Context) error {
	errCh := make(chan error, 1)

	go func() {
		if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case <-ctx.Done():
		shCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = s.srv.Shutdown(shCtx)
		return nil
	case err := <-errCh:
		return err
	}
}
