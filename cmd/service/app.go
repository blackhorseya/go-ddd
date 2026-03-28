package main

import (
	"context"
	"fmt"
	"sync"

	grpcserver "github.com/blackhorseya/go-ddd/internal/shared/adapter/grpc"
	httpserver "github.com/blackhorseya/go-ddd/internal/shared/adapter/http"
	healthhandler "github.com/blackhorseya/go-ddd/internal/shared/adapter/http/handler"

	identityhandler "github.com/blackhorseya/go-ddd/internal/identity/adapter/http/handler"
)

// App holds the assembled application with all servers and handlers wired up.
type App struct {
	http *httpserver.Server
	grpc *grpcserver.Server
}

// NewApp creates the application, registering all handlers on the HTTP router.
func NewApp(
	http *httpserver.Server,
	grpc *grpcserver.Server,
	health *healthhandler.HealthHandler,
	auth *identityhandler.AuthHandler,
) *App {
	r := http.Router()

	health.Register(r)
	auth.Register(r)

	return &App{http: http, grpc: grpc}
}

// Run starts all servers and blocks until the context is cancelled.
func (a *App) Run(ctx context.Context) error {
	var wg sync.WaitGroup
	errCh := make(chan error, 2)

	childCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	wg.Add(2)

	go func() {
		defer wg.Done()
		if err := a.http.Run(childCtx); err != nil {
			errCh <- fmt.Errorf("http: %w", err)
		}
	}()

	go func() {
		defer wg.Done()
		if err := a.grpc.Run(childCtx); err != nil {
			errCh <- fmt.Errorf("grpc: %w", err)
		}
	}()

	select {
	case err := <-errCh:
		cancel()
		wg.Wait()
		return err
	case <-ctx.Done():
		wg.Wait()
		return nil
	}
}
