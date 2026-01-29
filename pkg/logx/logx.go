package logx

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
)

// Logger wraps slog.Logger and implements the Logger interface.
// It satisfies contextx.Logger through Go's structural typing (duck typing).
type Logger struct {
	*slog.Logger
}

// New creates a new Logger based on the provided configuration.
// Returns an error if the configuration is invalid.
func New(cfg *Config) (*Logger, error) {
	if cfg == nil {
		cfg = defaultConfig()
	}

	level, err := parseLevel(cfg.Level)
	if err != nil {
		return nil, fmt.Errorf("logx: %w", err)
	}

	writer, err := getWriter(cfg.Output)
	if err != nil {
		return nil, fmt.Errorf("logx: %w", err)
	}

	opts := &slog.HandlerOptions{
		Level:       level,
		AddSource:   cfg.AddSource,
		ReplaceAttr: shortenSource,
	}

	handler, err := createHandler(cfg.Format, writer, opts)
	if err != nil {
		return nil, fmt.Errorf("logx: %w", err)
	}

	return &Logger{slog.New(handler)}, nil
}

// MustNew creates a new Logger and panics if configuration is invalid.
// Use this in main() for initialization.
func MustNew(cfg *Config) *Logger {
	l, err := New(cfg)
	if err != nil {
		panic(err)
	}

	return l
}

// Default returns a Logger with default configuration.
func Default() *Logger {
	l, _ := New(nil)
	return l
}

// parseLevel converts level string to slog.Level.
func parseLevel(s string) (slog.Level, error) {
	switch strings.ToLower(s) {
	case "debug":
		return slog.LevelDebug, nil
	case "info", "":
		return slog.LevelInfo, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, fmt.Errorf("unknown log level: %s", s)
	}
}

// getWriter returns the appropriate io.Writer based on output configuration.
func getWriter(output string) (io.Writer, error) {
	switch strings.ToLower(output) {
	case "stdout", "":
		return os.Stdout, nil
	case "stderr":
		return os.Stderr, nil
	default:
		return nil, fmt.Errorf("unsupported output: %s", output)
	}
}

// createHandler creates the appropriate slog.Handler based on format.
func createHandler(format string, w io.Writer, opts *slog.HandlerOptions) (slog.Handler, error) {
	switch strings.ToLower(format) {
	case "json", "":
		return slog.NewJSONHandler(w, opts), nil
	case "text":
		return slog.NewTextHandler(w, opts), nil
	default:
		return nil, fmt.Errorf("unsupported log format: %s", format)
	}
}

// shortenSource shortens the source file path to be relative from project markers.
// It looks for /internal/, /pkg/, or /cmd/ and keeps the relative path from there.
func shortenSource(_ []string, a slog.Attr) slog.Attr {
	if a.Key != slog.SourceKey {
		return a
	}

	source, ok := a.Value.Any().(*slog.Source)
	if !ok {
		return a
	}

	// Find project markers and shorten path
	for _, marker := range []string{"/internal/", "/pkg/", "/cmd/"} {
		if idx := strings.LastIndex(source.File, marker); idx != -1 {
			source.File = source.File[idx+1:] // +1 to skip the leading /
			break
		}
	}

	return a
}

// ============================================================================
// contextx.Logger interface implementation
// ============================================================================

// Debug logs a debug message.
func (l *Logger) Debug(msg string, args ...any) {
	l.Logger.Debug(msg, args...)
}

// Info logs an info message.
func (l *Logger) Info(msg string, args ...any) {
	l.Logger.Info(msg, args...)
}

// Warn logs a warning message.
func (l *Logger) Warn(msg string, args ...any) {
	l.Logger.Warn(msg, args...)
}

// Error logs an error message.
func (l *Logger) Error(msg string, args ...any) {
	l.Logger.Error(msg, args...)
}

// ============================================================================
// Additional methods
// ============================================================================

// With returns a new Logger with the given attributes.
func (l *Logger) With(args ...any) *Logger {
	return &Logger{l.Logger.With(args...)}
}

// WithGroup returns a new Logger with the given group name.
func (l *Logger) WithGroup(name string) *Logger {
	return &Logger{l.Logger.WithGroup(name)}
}

// SetAsDefault sets this logger as the default slog logger.
// This allows contextx to use slog directly with correct caller information.
func (l *Logger) SetAsDefault() {
	slog.SetDefault(l.Logger)
}
