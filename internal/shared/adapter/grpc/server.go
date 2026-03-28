package grpc

import (
	"context"
	"fmt"
	"net"

	ggrpc "google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	"github.com/blackhorseya/go-ddd/internal/shared/adapter/grpc/interceptor"
	"github.com/blackhorseya/go-ddd/pkg/contextx"
)

// Server wraps the gRPC server with graceful shutdown support.
type Server struct {
	server *ggrpc.Server
	addr   string
	health *health.Server
}

// NewServer creates a new gRPC server with tracing, logging interceptors,
// health check service, and optional reflection (for development).
func NewServer(cfg ServerConfig, serviceName string, isDev bool) *Server {
	s := ggrpc.NewServer(
		ggrpc.StatsHandler(interceptor.NewServerTracingHandler()),
		ggrpc.ChainUnaryInterceptor(
			interceptor.UnaryServerLogging(),
		),
		ggrpc.ChainStreamInterceptor(
			interceptor.StreamServerLogging(),
		),
	)

	// Register health service.
	hs := health.NewServer()
	healthpb.RegisterHealthServer(s, hs)
	hs.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)

	// Register reflection for development (enables grpcurl, grpcui, etc.).
	if isDev {
		reflection.Register(s)
	}

	return &Server{
		server: s,
		addr:   fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		health: hs,
	}
}

// GRPCServer returns the underlying gRPC server for service registration.
func (s *Server) GRPCServer() *ggrpc.Server {
	return s.server
}

// Run starts the server and blocks until the context is cancelled.
// It handles graceful shutdown when the context is done.
func (s *Server) Run(ctx context.Context) error {
	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("grpc listen: %w", err)
	}

	errCh := make(chan error, 1)

	go func() {
		contextx.From(ctx).Info("starting gRPC server", "addr", s.addr)

		if err := s.server.Serve(lis); err != nil {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return fmt.Errorf("grpc server error: %w", err)
	case <-ctx.Done():
		contextx.From(ctx).Info("shutting down gRPC server")
		s.health.SetServingStatus("", healthpb.HealthCheckResponse_NOT_SERVING)
		s.server.GracefulStop()
		return nil
	}
}

// Addr returns the server address.
func (s *Server) Addr() string {
	return s.addr
}
