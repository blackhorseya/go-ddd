package logx

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"

	"github.com/blackhorseya/go-ddd/pkg/contextx"
)

func TestNew(t *testing.T) {
	t.Run("nil config uses defaults", func(t *testing.T) {
		l, err := New(nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if l == nil {
			t.Fatal("expected logger, got nil")
		}
	})

	t.Run("valid config", func(t *testing.T) {
		cfg := &Config{
			Level:  "debug",
			Format: "json",
			Output: "stdout",
		}

		l, err := New(cfg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if l == nil {
			t.Fatal("expected logger, got nil")
		}
	})

	t.Run("invalid level returns error", func(t *testing.T) {
		cfg := &Config{
			Level: "invalid",
		}

		_, err := New(cfg)
		if err == nil {
			t.Fatal("expected error for invalid level")
		}

		if !strings.Contains(err.Error(), "unknown log level") {
			t.Errorf("expected 'unknown log level' error, got: %v", err)
		}
	})

	t.Run("invalid format returns error", func(t *testing.T) {
		cfg := &Config{
			Format: "xml",
		}

		_, err := New(cfg)
		if err == nil {
			t.Fatal("expected error for invalid format")
		}

		if !strings.Contains(err.Error(), "unsupported log format") {
			t.Errorf("expected 'unsupported log format' error, got: %v", err)
		}
	})

	t.Run("invalid output returns error", func(t *testing.T) {
		cfg := &Config{
			Output: "file",
		}

		_, err := New(cfg)
		if err == nil {
			t.Fatal("expected error for invalid output")
		}

		if !strings.Contains(err.Error(), "unsupported output") {
			t.Errorf("expected 'unsupported output' error, got: %v", err)
		}
	})
}

func TestMustNew(t *testing.T) {
	t.Run("valid config does not panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("unexpected panic: %v", r)
			}
		}()

		l := MustNew(nil)
		if l == nil {
			t.Fatal("expected logger, got nil")
		}
	})

	t.Run("invalid config panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic for invalid config")
			}
		}()

		MustNew(&Config{Level: "invalid"})
	})
}

func TestDefault(t *testing.T) {
	l := Default()
	if l == nil {
		t.Fatal("expected logger, got nil")
	}
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected slog.Level
		wantErr  bool
	}{
		{"debug", slog.LevelDebug, false},
		{"DEBUG", slog.LevelDebug, false},
		{"info", slog.LevelInfo, false},
		{"INFO", slog.LevelInfo, false},
		{"", slog.LevelInfo, false},
		{"warn", slog.LevelWarn, false},
		{"warning", slog.LevelWarn, false},
		{"error", slog.LevelError, false},
		{"invalid", slog.LevelInfo, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			level, err := parseLevel(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error")
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if level != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, level)
			}
		})
	}
}

func TestLoggerImplementsContextxLogger(t *testing.T) {
	l := Default()

	var _ contextx.Logger = l

	// Verify the interface is properly implemented
	l.Debug("debug message")
	l.Info("info message")
	l.Warn("warn message")
	l.Error("error message")
}

func TestLoggerWith(t *testing.T) {
	l := Default()
	l2 := l.With("key", "value")

	if l2 == nil {
		t.Fatal("expected logger, got nil")
	}

	if l2 == l {
		t.Error("With should return a new logger instance")
	}
}

func TestLoggerWithGroup(t *testing.T) {
	l := Default()
	l2 := l.WithGroup("group")

	if l2 == nil {
		t.Fatal("expected logger, got nil")
	}

	if l2 == l {
		t.Error("WithGroup should return a new logger instance")
	}
}

func TestJSONFormat(t *testing.T) {
	var buf bytes.Buffer

	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	l := &Logger{slog.New(handler)}

	l.Info("test message", "key", "value")

	var logEntry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("failed to parse JSON log: %v", err)
	}

	if logEntry["msg"] != "test message" {
		t.Errorf("expected msg='test message', got %v", logEntry["msg"])
	}

	if logEntry["key"] != "value" {
		t.Errorf("expected key='value', got %v", logEntry["key"])
	}
}

func TestTextFormat(t *testing.T) {
	var buf bytes.Buffer

	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	l := &Logger{slog.New(handler)}

	l.Info("test message", "key", "value")

	output := buf.String()
	if !strings.Contains(output, "test message") {
		t.Errorf("expected output to contain 'test message', got: %s", output)
	}

	if !strings.Contains(output, "key=value") {
		t.Errorf("expected output to contain 'key=value', got: %s", output)
	}
}

func TestConfigDefaults(t *testing.T) {
	cfg := defaultConfig()

	if cfg.Level != DefaultLevel {
		t.Errorf("expected level %q, got %q", DefaultLevel, cfg.Level)
	}

	if cfg.Format != DefaultFormat {
		t.Errorf("expected format %q, got %q", DefaultFormat, cfg.Format)
	}

	if cfg.Output != DefaultOutput {
		t.Errorf("expected output %q, got %q", DefaultOutput, cfg.Output)
	}

	if cfg.AddSource {
		t.Error("expected AddSource to be false")
	}
}
