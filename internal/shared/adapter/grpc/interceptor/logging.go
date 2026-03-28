package interceptor

import (
	"context"
	"time"

	ggrpc "google.golang.org/grpc"
	"google.golang.org/grpc/status"

	"github.com/blackhorseya/go-ddd/pkg/contextx"
)

// UnaryServerLogging returns a server-side unary interceptor that logs each RPC call.
func UnaryServerLogging() ggrpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *ggrpc.UnaryServerInfo,
		handler ggrpc.UnaryHandler,
	) (any, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		logRPC(ctx, "server", info.FullMethod, time.Since(start), err)
		return resp, err
	}
}

// StreamServerLogging returns a server-side stream interceptor that logs each RPC call.
func StreamServerLogging() ggrpc.StreamServerInterceptor {
	return func(
		srv any,
		ss ggrpc.ServerStream,
		info *ggrpc.StreamServerInfo,
		handler ggrpc.StreamHandler,
	) error {
		start := time.Now()
		err := handler(srv, ss)
		logRPC(ss.Context(), "server", info.FullMethod, time.Since(start), err)
		return err
	}
}

// UnaryClientLogging returns a client-side unary interceptor that logs each RPC call.
func UnaryClientLogging() ggrpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply any,
		cc *ggrpc.ClientConn,
		invoker ggrpc.UnaryInvoker,
		opts ...ggrpc.CallOption,
	) error {
		start := time.Now()
		err := invoker(ctx, method, req, reply, cc, opts...)
		logRPC(ctx, "client", method, time.Since(start), err)
		return err
	}
}

// StreamClientLogging returns a client-side stream interceptor that logs stream lifecycle.
func StreamClientLogging() ggrpc.StreamClientInterceptor {
	return func(
		ctx context.Context,
		desc *ggrpc.StreamDesc,
		cc *ggrpc.ClientConn,
		method string,
		streamer ggrpc.Streamer,
		opts ...ggrpc.CallOption,
	) (ggrpc.ClientStream, error) {
		start := time.Now()
		cs, err := streamer(ctx, desc, cc, method, opts...)
		if err != nil {
			logRPC(ctx, "client", method, time.Since(start), err)
			return nil, err
		}
		return &loggingClientStream{ClientStream: cs, ctx: ctx, method: method, start: start}, nil
	}
}

// loggingClientStream wraps grpc.ClientStream to log when the stream finishes.
type loggingClientStream struct {
	ggrpc.ClientStream
	ctx    context.Context
	method string
	start  time.Time
}

func (s *loggingClientStream) RecvMsg(m any) error {
	err := s.ClientStream.RecvMsg(m)
	if err != nil {
		logRPC(s.ctx, "client", s.method, time.Since(s.start), err)
	}
	return err
}

func logRPC(ctx context.Context, kind, method string, latency time.Duration, err error) {
	st, _ := status.FromError(err)
	logger := contextx.From(ctx).WithFields(
		"trace_id", contextx.GetTraceID(ctx),
		"grpc.kind", kind,
		"grpc.method", method,
		"grpc.code", st.Code().String(),
		"latency", latency.String(),
	)

	if err != nil {
		logger.Error("gRPC call failed", "error", err)
	} else {
		logger.Info("gRPC call completed")
	}
}
