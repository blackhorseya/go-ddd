// Package otelx provides OpenTelemetry tracing and metrics setup and utilities.
package otelx

import (
	"context"
	"errors"
	"fmt"

	"log/slog"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/log/global"
	noopmetric "go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace/noop"
)

// Provider wraps the OpenTelemetry tracer, meter, and logger providers with shutdown capability.
type Provider struct {
	tracerProvider *sdktrace.TracerProvider
	meterProvider  *metric.MeterProvider
	loggerProvider *log.LoggerProvider
}

// Setup initializes OpenTelemetry tracing and metrics based on the provided configuration.
// Returns a Provider that should be shut down when the application exits.
func Setup(ctx context.Context, cfg Config) (*Provider, error) {
	if !cfg.Enabled {
		otel.SetTracerProvider(noop.NewTracerProvider())
		otel.SetMeterProvider(noopmetric.NewMeterProvider())
		return &Provider{}, nil
	}

	// Create resource with service information
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion(cfg.ServiceVersion),
			attribute.String("deployment.environment", cfg.Environment),
		),
		resource.WithHost(),
		resource.WithOS(),
		resource.WithProcess(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Setup tracer provider
	tp, err := setupTracerProvider(ctx, cfg, res)
	if err != nil {
		return nil, fmt.Errorf("failed to setup tracer provider: %w", err)
	}

	// Setup meter provider
	mp, err := setupMeterProvider(ctx, cfg, res)
	if err != nil {
		return nil, fmt.Errorf("failed to setup meter provider: %w", err)
	}

	// Setup logger provider
	lp, err := setupLoggerProvider(ctx, cfg, res)
	if err != nil {
		return nil, fmt.Errorf("failed to setup logger provider: %w", err)
	}

	// Set global propagator
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return &Provider{tracerProvider: tp, meterProvider: mp, loggerProvider: lp}, nil
}

// Shutdown gracefully shuts down all providers.
func (p *Provider) Shutdown(ctx context.Context) error {
	var errs []error
	if p.loggerProvider != nil {
		errs = append(errs, p.loggerProvider.Shutdown(ctx))
	}
	if p.meterProvider != nil {
		errs = append(errs, p.meterProvider.Shutdown(ctx))
	}
	if p.tracerProvider != nil {
		errs = append(errs, p.tracerProvider.Shutdown(ctx))
	}
	return errors.Join(errs...)
}

// SlogHandler returns an slog.Handler that sends logs via OTel.
// Use this to create a fan-out handler alongside the existing stdout handler.
func (p *Provider) SlogHandler() slog.Handler {
	if p.loggerProvider == nil {
		return otelslog.NewHandler("")
	}
	return otelslog.NewHandler("", otelslog.WithLoggerProvider(p.loggerProvider))
}

func setupTracerProvider(ctx context.Context, cfg Config, res *resource.Resource) (*sdktrace.TracerProvider, error) {
	exporter, err := createTraceExporter(ctx, cfg)
	if err != nil {
		return nil, err
	}

	var sampler sdktrace.Sampler
	if cfg.SampleRate >= 1.0 {
		sampler = sdktrace.AlwaysSample()
	} else if cfg.SampleRate <= 0.0 {
		sampler = sdktrace.NeverSample()
	} else {
		sampler = sdktrace.TraceIDRatioBased(cfg.SampleRate)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
	)
	otel.SetTracerProvider(tp)

	return tp, nil
}

func setupLoggerProvider(ctx context.Context, cfg Config, res *resource.Resource) (*log.LoggerProvider, error) {
	if cfg.Exporter != "otlp" {
		return nil, nil
	}

	opts := []otlploghttp.Option{otlploghttp.WithEndpoint(cfg.OTLP.Endpoint)}
	if cfg.OTLP.Insecure {
		opts = append(opts, otlploghttp.WithInsecure())
	}

	exporter, err := otlploghttp.New(ctx, opts...)
	if err != nil {
		return nil, err
	}

	lp := log.NewLoggerProvider(
		log.WithResource(res),
		log.WithProcessor(log.NewBatchProcessor(exporter)),
	)
	global.SetLoggerProvider(lp)

	return lp, nil
}

func setupMeterProvider(ctx context.Context, cfg Config, res *resource.Resource) (*metric.MeterProvider, error) {
	exporter, err := createMetricExporter(ctx, cfg)
	if err != nil {
		return nil, err
	}

	mp := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(metric.NewPeriodicReader(exporter)),
	)
	otel.SetMeterProvider(mp)

	return mp, nil
}

// createTraceExporter creates a span exporter based on configuration.
func createTraceExporter(ctx context.Context, cfg Config) (sdktrace.SpanExporter, error) {
	switch cfg.Exporter {
	case "otlp":
		return createOTLPTraceExporter(ctx, cfg.OTLP)
	case "stdout":
		return stdouttrace.New()
	case "noop", "":
		return stdouttrace.New(stdouttrace.WithWriter(noopWriter{}))
	default:
		return nil, fmt.Errorf("unknown exporter type: %s", cfg.Exporter)
	}
}

func createOTLPTraceExporter(ctx context.Context, cfg OTLPConfig) (sdktrace.SpanExporter, error) {
	switch cfg.Protocol {
	case "grpc":
		opts := []otlptracegrpc.Option{otlptracegrpc.WithEndpoint(cfg.Endpoint)}
		if cfg.Insecure {
			opts = append(opts, otlptracegrpc.WithInsecure())
		}
		return otlptracegrpc.New(ctx, opts...)
	case "http", "":
		opts := []otlptracehttp.Option{otlptracehttp.WithEndpoint(cfg.Endpoint)}
		if cfg.Insecure {
			opts = append(opts, otlptracehttp.WithInsecure())
		}
		return otlptracehttp.New(ctx, opts...)
	default:
		return nil, fmt.Errorf("unknown OTLP protocol: %s", cfg.Protocol)
	}
}

// createMetricExporter creates a metric exporter based on configuration.
func createMetricExporter(ctx context.Context, cfg Config) (metric.Exporter, error) {
	switch cfg.Exporter {
	case "otlp":
		return createOTLPMetricExporter(ctx, cfg.OTLP)
	case "stdout", "noop", "":
		// For non-OTLP modes, use a noop metric exporter
		return createOTLPMetricExporter(ctx, OTLPConfig{
			Endpoint: "localhost:4318",
			Insecure: true,
			Protocol: "http",
		})
	default:
		return nil, fmt.Errorf("unknown exporter type: %s", cfg.Exporter)
	}
}

func createOTLPMetricExporter(ctx context.Context, cfg OTLPConfig) (metric.Exporter, error) {
	switch cfg.Protocol {
	case "grpc":
		opts := []otlpmetricgrpc.Option{otlpmetricgrpc.WithEndpoint(cfg.Endpoint)}
		if cfg.Insecure {
			opts = append(opts, otlpmetricgrpc.WithInsecure())
		}
		return otlpmetricgrpc.New(ctx, opts...)
	case "http", "":
		opts := []otlpmetrichttp.Option{otlpmetrichttp.WithEndpoint(cfg.Endpoint)}
		if cfg.Insecure {
			opts = append(opts, otlpmetrichttp.WithInsecure())
		}
		return otlpmetrichttp.New(ctx, opts...)
	default:
		return nil, fmt.Errorf("unknown OTLP protocol: %s", cfg.Protocol)
	}
}

// noopWriter is a writer that discards all output.
type noopWriter struct{}

func (noopWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}
