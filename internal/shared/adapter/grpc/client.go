package grpc

import (
	"time"

	ggrpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/blackhorseya/go-ddd/internal/shared/adapter/grpc/interceptor"
)

const defaultServiceConfig = `{
	"methodConfig": [{
		"name": [{"service": ""}],
		"retryPolicy": {
			"maxAttempts": 3,
			"initialBackoff": "0.1s",
			"maxBackoff": "1s",
			"backoffMultiplier": 2,
			"retryableStatusCodes": ["UNAVAILABLE", "DEADLINE_EXCEEDED"]
		}
	}],
	"healthCheckConfig": {
		"serviceName": ""
	}
}`

// NewClientConn creates a gRPC client connection with tracing, logging interceptors,
// retry policy, and health check support.
func NewClientConn(cfg ClientConfig) (*ggrpc.ClientConn, error) {
	if cfg.Timeout == 0 {
		cfg.Timeout = 5 * time.Second
	}

	opts := []ggrpc.DialOption{
		ggrpc.WithDefaultServiceConfig(defaultServiceConfig),
		ggrpc.WithDefaultCallOptions(ggrpc.WaitForReady(true)),
		ggrpc.WithStatsHandler(interceptor.NewClientTracingHandler()),
		ggrpc.WithChainUnaryInterceptor(
			interceptor.UnaryClientLogging(),
		),
		ggrpc.WithChainStreamInterceptor(
			interceptor.StreamClientLogging(),
		),
	}

	if cfg.Insecure {
		opts = append(opts, ggrpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	return ggrpc.NewClient(cfg.Target, opts...)
}
