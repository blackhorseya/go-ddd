// Package contextx provides an extended context wrapper with structured logging capabilities.
// It wraps the standard context.Context and provides convenient logging methods.
package contextx

import (
	"context"
	"log/slog"
	"runtime"
	"time"

	"go.opentelemetry.io/otel/trace"
)

// Logger defines the interface for structured logging.
type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

// Contextx wraps context.Context with logging capabilities.
type Contextx struct {
	context.Context
}

// context keys for storing values.
type (
	loggerKeyType        struct{}
	fieldsKeyType        struct{}
	requestIDKeyType     struct{}
	traceIDKeyType       struct{}
	userIDKeyType        struct{}
	correlationIDKeyType struct{}
	operationKeyType     struct{}
	serviceKeyType       struct{}
	environmentKeyType   struct{}
)

var (
	loggerKey        = loggerKeyType{}
	fieldsKey        = fieldsKeyType{}
	requestIDKey     = requestIDKeyType{}
	traceIDKey       = traceIDKeyType{}
	userIDKey        = userIDKeyType{}
	correlationIDKey = correlationIDKeyType{}
	operationKey     = operationKeyType{}
	serviceKey       = serviceKeyType{}
	environmentKey   = environmentKeyType{}
)

// defaultLogger is the fallback logger using slog.
var defaultLogger Logger = &slogAdapter{slog.Default()}

// slogAdapter adapts slog.Logger to our Logger interface.
type slogAdapter struct {
	*slog.Logger
}

func (s *slogAdapter) Debug(msg string, args ...any) { s.Logger.Debug(msg, args...) }
func (s *slogAdapter) Info(msg string, args ...any)  { s.Logger.Info(msg, args...) }
func (s *slogAdapter) Warn(msg string, args ...any)  { s.Logger.Warn(msg, args...) }
func (s *slogAdapter) Error(msg string, args ...any) { s.Logger.Error(msg, args...) }

// From creates a Contextx from a standard context.Context.
func From(c context.Context) *Contextx {
	return &Contextx{c}
}

// Background returns a Contextx wrapping context.Background().
func Background() *Contextx {
	return &Contextx{context.Background()}
}

// TODO returns a Contextx wrapping context.TODO().
func TODO() *Contextx {
	return &Contextx{context.TODO()}
}

// WithLogger returns a new context with the given logger attached.
func WithLogger(c context.Context, logger Logger) context.Context {
	return context.WithValue(c, loggerKey, logger)
}

// WithFields returns a new context with additional logging fields.
// These fields will be automatically included in all subsequent log calls.
func WithFields(c context.Context, args ...any) context.Context {
	existing := fieldsFromContext(c)
	newFields := make([]any, 0, len(existing)+len(args))
	newFields = append(newFields, existing...)
	newFields = append(newFields, args...)

	return context.WithValue(c, fieldsKey, newFields)
}

// FromContext extracts the Logger from context, or returns the default logger.
func FromContext(c context.Context) Logger {
	if logger, ok := c.Value(loggerKey).(Logger); ok {
		return logger
	}

	return defaultLogger
}

// fieldsFromContext extracts accumulated fields from context.
func fieldsFromContext(c context.Context) []any {
	if fields, ok := c.Value(fieldsKey).([]any); ok {
		return fields
	}

	return nil
}

// SetDefaultLogger sets the default logger for contexts without an explicit logger.
func SetDefaultLogger(logger Logger) {
	defaultLogger = logger
}

// Debug logs a debug message with optional structured arguments.
func (ctx *Contextx) Debug(msg string, args ...any) {
	ctx.logWithCaller(slog.LevelDebug, msg, args...)
}

// Info logs an info message with optional structured arguments.
func (ctx *Contextx) Info(msg string, args ...any) {
	ctx.logWithCaller(slog.LevelInfo, msg, args...)
}

// Warn logs a warning message with optional structured arguments.
func (ctx *Contextx) Warn(msg string, args ...any) {
	ctx.logWithCaller(slog.LevelWarn, msg, args...)
}

// Error logs an error message with optional structured arguments.
func (ctx *Contextx) Error(msg string, args ...any) {
	ctx.logWithCaller(slog.LevelError, msg, args...)
}

// logWithCaller logs a message with the correct caller location.
// It captures the caller 3 levels up: Callers() -> logWithCaller() -> Info/Debug/etc() -> business code
func (ctx *Contextx) logWithCaller(level slog.Level, msg string, args ...any) {
	// Merge context fields with provided args
	fields := fieldsFromContext(ctx.Context)
	allArgs := append(fields, args...)

	// Check for custom logger: in context or via SetDefaultLogger
	// If a custom logger is set, use it (for testing and custom logger support)
	var customLogger Logger
	if logger, ok := ctx.Value(loggerKey).(Logger); ok {
		customLogger = logger
	} else if _, isSlogAdapter := defaultLogger.(*slogAdapter); !isSlogAdapter {
		// defaultLogger has been replaced with a custom logger via SetDefaultLogger
		customLogger = defaultLogger
	}

	if customLogger != nil {
		switch level {
		case slog.LevelDebug:
			customLogger.Debug(msg, allArgs...)
		case slog.LevelInfo:
			customLogger.Info(msg, allArgs...)
		case slog.LevelWarn:
			customLogger.Warn(msg, allArgs...)
		case slog.LevelError:
			customLogger.Error(msg, allArgs...)
		}
		return
	}

	// For slog default logger, capture caller PC and log directly
	// Skip: Callers, logWithCaller, Info/Debug/etc
	var pcs [1]uintptr
	runtime.Callers(3, pcs[:])

	// Create record with correct caller info and call handler
	r := slog.NewRecord(time.Now(), level, msg, pcs[0])
	r.Add(allArgs...)
	_ = slog.Default().Handler().Handle(ctx.Context, r)
}

// WithLogger returns a new Contextx with the given logger attached.
func (ctx *Contextx) WithLogger(logger Logger) *Contextx {
	return From(WithLogger(ctx.Context, logger))
}

// WithFields returns a new Contextx with additional logging fields.
func (ctx *Contextx) WithFields(args ...any) *Contextx {
	return From(WithFields(ctx.Context, args...))
}

// ============================================================================
// Request ID
// ============================================================================

// WithRequestID returns a new context with the request ID attached.
func WithRequestID(c context.Context, requestID string) context.Context {
	return context.WithValue(c, requestIDKey, requestID)
}

// GetRequestID extracts the request ID from context.
// Returns empty string if not found.
func GetRequestID(c context.Context) string {
	if v, ok := c.Value(requestIDKey).(string); ok {
		return v
	}

	return ""
}

// WithRequestID returns a new Contextx with the request ID attached.
func (ctx *Contextx) WithRequestID(requestID string) *Contextx {
	return From(WithRequestID(ctx.Context, requestID))
}

// RequestID returns the request ID from context.
func (ctx *Contextx) RequestID() string {
	return GetRequestID(ctx.Context)
}

// ============================================================================
// Trace ID (for distributed tracing)
// ============================================================================

// WithTraceID returns a new context with the trace ID attached.
func WithTraceID(c context.Context, traceID string) context.Context {
	return context.WithValue(c, traceIDKey, traceID)
}

// GetTraceID extracts the trace ID from context.
// It first checks for an OpenTelemetry span context, then falls back to context value.
// Returns empty string if not found.
func GetTraceID(c context.Context) string {
	// First, try to get trace ID from OpenTelemetry span context
	span := trace.SpanFromContext(c)
	if span.SpanContext().HasTraceID() {
		return span.SpanContext().TraceID().String()
	}

	// Fallback to context value
	if v, ok := c.Value(traceIDKey).(string); ok {
		return v
	}

	return ""
}

// WithTraceID returns a new Contextx with the trace ID attached.
func (ctx *Contextx) WithTraceID(traceID string) *Contextx {
	return From(WithTraceID(ctx.Context, traceID))
}

// TraceID returns the trace ID from context.
func (ctx *Contextx) TraceID() string {
	return GetTraceID(ctx.Context)
}

// ============================================================================
// User ID
// ============================================================================

// WithUserID returns a new context with the user ID attached.
func WithUserID(c context.Context, userID string) context.Context {
	return context.WithValue(c, userIDKey, userID)
}

// GetUserID extracts the user ID from context.
// Returns empty string if not found.
func GetUserID(c context.Context) string {
	if v, ok := c.Value(userIDKey).(string); ok {
		return v
	}

	return ""
}

// WithUserID returns a new Contextx with the user ID attached.
func (ctx *Contextx) WithUserID(userID string) *Contextx {
	return From(WithUserID(ctx.Context, userID))
}

// UserID returns the user ID from context.
func (ctx *Contextx) UserID() string {
	return GetUserID(ctx.Context)
}

// ============================================================================
// Correlation ID (for cross-service tracing)
// ============================================================================

// WithCorrelationID returns a new context with the correlation ID attached.
func WithCorrelationID(c context.Context, correlationID string) context.Context {
	return context.WithValue(c, correlationIDKey, correlationID)
}

// GetCorrelationID extracts the correlation ID from context.
// Returns empty string if not found.
func GetCorrelationID(c context.Context) string {
	if v, ok := c.Value(correlationIDKey).(string); ok {
		return v
	}

	return ""
}

// WithCorrelationID returns a new Contextx with the correlation ID attached.
func (ctx *Contextx) WithCorrelationID(correlationID string) *Contextx {
	return From(WithCorrelationID(ctx.Context, correlationID))
}

// CorrelationID returns the correlation ID from context.
func (ctx *Contextx) CorrelationID() string {
	return GetCorrelationID(ctx.Context)
}

// ============================================================================
// Operation (current operation/function name)
// ============================================================================

// WithOperation returns a new context with the operation name attached.
func WithOperation(c context.Context, operation string) context.Context {
	return context.WithValue(c, operationKey, operation)
}

// GetOperation extracts the operation name from context.
// Returns empty string if not found.
func GetOperation(c context.Context) string {
	if v, ok := c.Value(operationKey).(string); ok {
		return v
	}

	return ""
}

// WithOperation returns a new Contextx with the operation name attached.
func (ctx *Contextx) WithOperation(operation string) *Contextx {
	return From(WithOperation(ctx.Context, operation))
}

// Operation returns the operation name from context.
func (ctx *Contextx) Operation() string {
	return GetOperation(ctx.Context)
}

// ============================================================================
// Service (service name)
// ============================================================================

// WithService returns a new context with the service name attached.
func WithService(c context.Context, service string) context.Context {
	return context.WithValue(c, serviceKey, service)
}

// GetService extracts the service name from context.
// Returns empty string if not found.
func GetService(c context.Context) string {
	if v, ok := c.Value(serviceKey).(string); ok {
		return v
	}

	return ""
}

// WithService returns a new Contextx with the service name attached.
func (ctx *Contextx) WithService(service string) *Contextx {
	return From(WithService(ctx.Context, service))
}

// Service returns the service name from context.
func (ctx *Contextx) Service() string {
	return GetService(ctx.Context)
}

// ============================================================================
// Environment (dev, staging, prod)
// ============================================================================

// WithEnvironment returns a new context with the environment attached.
func WithEnvironment(c context.Context, env string) context.Context {
	return context.WithValue(c, environmentKey, env)
}

// GetEnvironment extracts the environment from context.
// Returns empty string if not found.
func GetEnvironment(c context.Context) string {
	if v, ok := c.Value(environmentKey).(string); ok {
		return v
	}

	return ""
}

// WithEnvironment returns a new Contextx with the environment attached.
func (ctx *Contextx) WithEnvironment(env string) *Contextx {
	return From(WithEnvironment(ctx.Context, env))
}

// Environment returns the environment from context.
func (ctx *Contextx) Environment() string {
	return GetEnvironment(ctx.Context)
}

// ============================================================================
// Convenience methods
// ============================================================================

// HasRequestID checks if the context has a request ID.
func (ctx *Contextx) HasRequestID() bool {
	return ctx.RequestID() != ""
}

// HasUserID checks if the context has a user ID.
func (ctx *Contextx) HasUserID() bool {
	return ctx.UserID() != ""
}

// HasTraceID checks if the context has a trace ID.
func (ctx *Contextx) HasTraceID() bool {
	return ctx.TraceID() != ""
}

// GetSpanID extracts the span ID from an OpenTelemetry span context.
// Returns empty string if no span is active.
func GetSpanID(c context.Context) string {
	span := trace.SpanFromContext(c)
	if span.SpanContext().HasSpanID() {
		return span.SpanContext().SpanID().String()
	}
	return ""
}

// SpanID returns the span ID from context.
func (ctx *Contextx) SpanID() string {
	return GetSpanID(ctx.Context)
}

// LogFields returns common context values as log fields.
// Useful for automatically including context info in logs.
func (ctx *Contextx) LogFields() []any {
	var fields []any

	if svc := ctx.Service(); svc != "" {
		fields = append(fields, "service", svc)
	}

	if env := ctx.Environment(); env != "" {
		fields = append(fields, "environment", env)
	}

	if op := ctx.Operation(); op != "" {
		fields = append(fields, "operation", op)
	}

	if rid := ctx.RequestID(); rid != "" {
		fields = append(fields, "request_id", rid)
	}

	if tid := ctx.TraceID(); tid != "" {
		fields = append(fields, "trace_id", tid)
	}

	if uid := ctx.UserID(); uid != "" {
		fields = append(fields, "user_id", uid)
	}

	if cid := ctx.CorrelationID(); cid != "" {
		fields = append(fields, "correlation_id", cid)
	}

	return fields
}
