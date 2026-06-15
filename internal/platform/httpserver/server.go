// Package httpserver builds the base chi router (shared middleware) and runs the
// HTTP server with graceful shutdown. Route registration is left to the caller
// (cmd/api wires modules onto the two surfaces: /api/app and /api/v1).
package httpserver

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Config is the subset of configuration the HTTP server needs.
type Config struct {
	Port               int
	ReadTimeout        time.Duration
	WriteTimeout       time.Duration
	ShutdownTimeout    time.Duration
	CORSAllowedOrigins []string
}

// Server wraps a chi router and an *http.Server with graceful shutdown.
type Server struct {
	cfg    Config
	log    *slog.Logger
	router chi.Router
	http   *http.Server
}

// New constructs the base router with the standard middleware stack.
func New(cfg Config, log *slog.Logger) *Server {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(recoverer(log))
	r.Use(requestLogger(log))
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(corsMiddleware(cfg.CORSAllowedOrigins))

	return &Server{
		cfg:    cfg,
		log:    log,
		router: r,
		http: &http.Server{
			Addr:         fmt.Sprintf(":%d", cfg.Port),
			Handler:      r,
			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
		},
	}
}

// Router exposes the chi router so the caller can register routes before Run.
func (s *Server) Router() chi.Router { return s.router }

// Run starts the server and blocks until ctx is cancelled, then shuts down
// gracefully within the configured timeout.
func (s *Server) Run(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		s.log.Info("http server started", "addr", s.http.Addr)
		if err := s.http.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return fmt.Errorf("listen and serve: %w", err)
	case <-ctx.Done():
		s.log.Info("shutdown signal received, draining connections")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), s.cfg.ShutdownTimeout)
	defer cancel()
	if err := s.http.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful shutdown: %w", err)
	}
	s.log.Info("http server stopped")
	return nil
}
