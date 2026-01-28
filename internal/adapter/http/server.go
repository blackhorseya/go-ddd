package http

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"

	"github.com/blackhorseya/go-ddd/internal/adapter/http/handler"
	"github.com/blackhorseya/go-ddd/internal/adapter/http/router"
	"github.com/blackhorseya/go-ddd/internal/infrastructure/config"
	"github.com/blackhorseya/go-ddd/pkg/contextx"
	"github.com/blackhorseya/go-ddd/pkg/logx"
	"github.com/gin-gonic/gin"
)

// Server wraps the HTTP server with graceful shutdown support.
type Server struct {
	server *http.Server
	router *gin.Engine
	logger *logx.Logger
}

// NewServer creates a new HTTP server.
func NewServer(cfg config.HTTP, logger *logx.Logger) *Server {
	opts := router.DefaultOptions(logger)
	r := router.New(opts)

	// Register handlers
	handler.NewHealthHandler().Register(r)

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}

	return &Server{
		server: srv,
		router: r,
		logger: logger,
	}
}

// Router returns the underlying Gin engine for additional route registration.
func (s *Server) Router() *gin.Engine {
	return s.router
}

// Run starts the server and blocks until the context is cancelled.
// It handles graceful shutdown when the context is done.
func (s *Server) Run(ctx context.Context) error {
	errCh := make(chan error, 1)

	go func() {
		contextx.From(ctx).Info("starting HTTP server", "addr", s.server.Addr)

		if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return fmt.Errorf("http server error: %w", err)
	case <-ctx.Done():
		contextx.From(ctx).Info("shutting down HTTP server")
		return s.server.Shutdown(context.Background())
	}
}

// Addr returns the server address. Useful for tests.
func (s *Server) Addr() string {
	return s.server.Addr
}

// ListenAndServe starts the server on a random available port.
// Returns the listener for retrieving the actual port. Useful for tests.
func (s *Server) ListenAndServe(ctx context.Context) (net.Listener, error) {
	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		return nil, err
	}

	go func() {
		<-ctx.Done()
		_ = s.server.Shutdown(context.Background())
	}()

	go func() {
		_ = s.server.Serve(ln)
	}()

	return ln, nil
}
