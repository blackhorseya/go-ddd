package grpc

import (
	ggrpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
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
// The connection is lazy; actual TCP dial happens on the first RPC call.
// Use per-RPC context deadlines to control timeouts.
func NewClientConn(cfg ClientConfig) (*ggrpc.ClientConn, error) {
	var creds ggrpc.DialOption
	if cfg.Insecure {
		creds = ggrpc.WithTransportCredentials(insecure.NewCredentials())
	} else {
		creds = ggrpc.WithTransportCredentials(credentials.NewClientTLSFromCert(nil, ""))
	}

	return ggrpc.NewClient(cfg.Target,
		creds,
		ggrpc.WithDefaultServiceConfig(defaultServiceConfig),
		ggrpc.WithDefaultCallOptions(ggrpc.WaitForReady(true)),
		ggrpc.WithStatsHandler(interceptor.NewClientTracingHandler()),
		ggrpc.WithChainUnaryInterceptor(
			interceptor.UnaryClientLogging(),
		),
		ggrpc.WithChainStreamInterceptor(
			interceptor.StreamClientLogging(),
		),
	)
}
