package contextx

import (
	"context"
	"testing"
)

// mockLogger is a test logger that captures log calls.
type mockLogger struct {
	debugCalls []logCall
	infoCalls  []logCall
	warnCalls  []logCall
	errorCalls []logCall
}

type logCall struct {
	msg  string
	args []any
}

func (m *mockLogger) Debug(msg string, args ...any) {
	m.debugCalls = append(m.debugCalls, logCall{msg, args})
}

func (m *mockLogger) Info(msg string, args ...any) {
	m.infoCalls = append(m.infoCalls, logCall{msg, args})
}

func (m *mockLogger) Warn(msg string, args ...any) {
	m.warnCalls = append(m.warnCalls, logCall{msg, args})
}

func (m *mockLogger) Error(msg string, args ...any) {
	m.errorCalls = append(m.errorCalls, logCall{msg, args})
}

func TestFrom(t *testing.T) {
	c := context.Background()
	ctx := From(c)

	if ctx == nil {
		t.Fatal("From returned nil")
	}

	if ctx.Context != c {
		t.Error("From did not wrap the original context")
	}
}

func TestBackground(t *testing.T) {
	ctx := Background()

	if ctx == nil {
		t.Fatal("Background returned nil")
	}
}

func TestTODO(t *testing.T) {
	ctx := TODO()

	if ctx == nil {
		t.Fatal("TODO returned nil")
	}
}

func TestWithLogger(t *testing.T) {
	mock := &mockLogger{}
	c := context.Background()
	c = WithLogger(c, mock)

	ctx := From(c)
	ctx.Info("test message", "key", "value")

	if len(mock.infoCalls) != 1 {
		t.Fatalf("expected 1 info call, got %d", len(mock.infoCalls))
	}

	if mock.infoCalls[0].msg != "test message" {
		t.Errorf("expected message 'test message', got %q", mock.infoCalls[0].msg)
	}
}

func TestWithFields(t *testing.T) {
	mock := &mockLogger{}
	c := context.Background()
	c = WithLogger(c, mock)
	c = WithFields(c, "request_id", "123")

	ctx := From(c)
	ctx.Info("test message", "extra", "data")

	if len(mock.infoCalls) != 1 {
		t.Fatalf("expected 1 info call, got %d", len(mock.infoCalls))
	}

	args := mock.infoCalls[0].args
	if len(args) != 4 {
		t.Fatalf("expected 4 args, got %d", len(args))
	}

	// Check that fields are prepended
	if args[0] != "request_id" || args[1] != "123" {
		t.Errorf("expected prepended fields, got %v", args[:2])
	}

	if args[2] != "extra" || args[3] != "data" {
		t.Errorf("expected appended args, got %v", args[2:])
	}
}

func TestContextxWithLogger(t *testing.T) {
	mock := &mockLogger{}
	ctx := Background().WithLogger(mock)

	ctx.Debug("debug msg")
	ctx.Info("info msg")
	ctx.Warn("warn msg")
	ctx.Error("error msg")

	if len(mock.debugCalls) != 1 {
		t.Errorf("expected 1 debug call, got %d", len(mock.debugCalls))
	}

	if len(mock.infoCalls) != 1 {
		t.Errorf("expected 1 info call, got %d", len(mock.infoCalls))
	}

	if len(mock.warnCalls) != 1 {
		t.Errorf("expected 1 warn call, got %d", len(mock.warnCalls))
	}

	if len(mock.errorCalls) != 1 {
		t.Errorf("expected 1 error call, got %d", len(mock.errorCalls))
	}
}

func TestContextxWithFields(t *testing.T) {
	mock := &mockLogger{}
	ctx := Background().
		WithLogger(mock).
		WithFields("user_id", 42)

	ctx.Info("user action")

	if len(mock.infoCalls) != 1 {
		t.Fatalf("expected 1 info call, got %d", len(mock.infoCalls))
	}

	args := mock.infoCalls[0].args
	if len(args) != 2 {
		t.Fatalf("expected 2 args, got %d", len(args))
	}

	if args[0] != "user_id" || args[1] != 42 {
		t.Errorf("expected fields [user_id, 42], got %v", args)
	}
}

func TestFromContext(t *testing.T) {
	t.Run("with logger", func(t *testing.T) {
		mock := &mockLogger{}
		c := WithLogger(context.Background(), mock)

		logger := FromContext(c)
		if logger != mock {
			t.Error("FromContext did not return the attached logger")
		}
	})

	t.Run("without logger returns default", func(t *testing.T) {
		c := context.Background()
		logger := FromContext(c)

		if logger == nil {
			t.Error("FromContext returned nil for context without logger")
		}
	})
}

func TestSetDefaultLogger(t *testing.T) {
	original := defaultLogger
	defer func() { defaultLogger = original }()

	mock := &mockLogger{}
	SetDefaultLogger(mock)

	ctx := Background()
	ctx.Info("test")

	if len(mock.infoCalls) != 1 {
		t.Errorf("expected 1 info call on new default logger, got %d", len(mock.infoCalls))
	}
}

func TestChainedWithFields(t *testing.T) {
	mock := &mockLogger{}
	ctx := Background().
		WithLogger(mock).
		WithFields("a", 1).
		WithFields("b", 2)

	ctx.Info("message")

	args := mock.infoCalls[0].args
	if len(args) != 4 {
		t.Fatalf("expected 4 args from chained WithFields, got %d", len(args))
	}
}

// ============================================================================
// Request ID Tests
// ============================================================================

func TestRequestID(t *testing.T) {
	t.Run("WithRequestID and GetRequestID", func(t *testing.T) {
		c := context.Background()
		c = WithRequestID(c, "req-123")

		got := GetRequestID(c)
		if got != "req-123" {
			t.Errorf("expected 'req-123', got %q", got)
		}
	})

	t.Run("GetRequestID returns empty for missing", func(t *testing.T) {
		c := context.Background()
		got := GetRequestID(c)

		if got != "" {
			t.Errorf("expected empty string, got %q", got)
		}
	})

	t.Run("Contextx methods", func(t *testing.T) {
		ctx := Background().WithRequestID("req-456")

		if ctx.RequestID() != "req-456" {
			t.Errorf("expected 'req-456', got %q", ctx.RequestID())
		}

		if !ctx.HasRequestID() {
			t.Error("expected HasRequestID to return true")
		}
	})
}

// ============================================================================
// Trace ID Tests
// ============================================================================

func TestTraceID(t *testing.T) {
	t.Run("WithTraceID and GetTraceID", func(t *testing.T) {
		c := context.Background()
		c = WithTraceID(c, "trace-abc")

		got := GetTraceID(c)
		if got != "trace-abc" {
			t.Errorf("expected 'trace-abc', got %q", got)
		}
	})

	t.Run("GetTraceID returns empty for missing", func(t *testing.T) {
		c := context.Background()
		got := GetTraceID(c)

		if got != "" {
			t.Errorf("expected empty string, got %q", got)
		}
	})

	t.Run("Contextx methods", func(t *testing.T) {
		ctx := Background().WithTraceID("trace-xyz")

		if ctx.TraceID() != "trace-xyz" {
			t.Errorf("expected 'trace-xyz', got %q", ctx.TraceID())
		}

		if !ctx.HasTraceID() {
			t.Error("expected HasTraceID to return true")
		}
	})
}

// ============================================================================
// User ID Tests
// ============================================================================

func TestUserID(t *testing.T) {
	t.Run("WithUserID and GetUserID", func(t *testing.T) {
		c := context.Background()
		c = WithUserID(c, "user-001")

		got := GetUserID(c)
		if got != "user-001" {
			t.Errorf("expected 'user-001', got %q", got)
		}
	})

	t.Run("GetUserID returns empty for missing", func(t *testing.T) {
		c := context.Background()
		got := GetUserID(c)

		if got != "" {
			t.Errorf("expected empty string, got %q", got)
		}
	})

	t.Run("Contextx methods", func(t *testing.T) {
		ctx := Background().WithUserID("user-002")

		if ctx.UserID() != "user-002" {
			t.Errorf("expected 'user-002', got %q", ctx.UserID())
		}

		if !ctx.HasUserID() {
			t.Error("expected HasUserID to return true")
		}
	})
}

// ============================================================================
// Correlation ID Tests
// ============================================================================

func TestCorrelationID(t *testing.T) {
	t.Run("WithCorrelationID and GetCorrelationID", func(t *testing.T) {
		c := context.Background()
		c = WithCorrelationID(c, "corr-999")

		got := GetCorrelationID(c)
		if got != "corr-999" {
			t.Errorf("expected 'corr-999', got %q", got)
		}
	})

	t.Run("GetCorrelationID returns empty for missing", func(t *testing.T) {
		c := context.Background()
		got := GetCorrelationID(c)

		if got != "" {
			t.Errorf("expected empty string, got %q", got)
		}
	})

	t.Run("Contextx methods", func(t *testing.T) {
		ctx := Background().WithCorrelationID("corr-888")

		if ctx.CorrelationID() != "corr-888" {
			t.Errorf("expected 'corr-888', got %q", ctx.CorrelationID())
		}
	})
}

// ============================================================================
// LogFields Tests
// ============================================================================

func TestLogFields(t *testing.T) {
	t.Run("returns all set fields", func(t *testing.T) {
		ctx := Background().
			WithRequestID("req-1").
			WithTraceID("trace-1").
			WithUserID("user-1").
			WithCorrelationID("corr-1")

		fields := ctx.LogFields()

		if len(fields) != 8 {
			t.Fatalf("expected 8 fields (4 key-value pairs), got %d", len(fields))
		}

		// Check fields are present
		fieldMap := make(map[string]string)
		for i := 0; i < len(fields); i += 2 {
			key := fields[i].(string)
			value := fields[i+1].(string)
			fieldMap[key] = value
		}

		if fieldMap["request_id"] != "req-1" {
			t.Errorf("expected request_id=req-1, got %s", fieldMap["request_id"])
		}

		if fieldMap["trace_id"] != "trace-1" {
			t.Errorf("expected trace_id=trace-1, got %s", fieldMap["trace_id"])
		}

		if fieldMap["user_id"] != "user-1" {
			t.Errorf("expected user_id=user-1, got %s", fieldMap["user_id"])
		}

		if fieldMap["correlation_id"] != "corr-1" {
			t.Errorf("expected correlation_id=corr-1, got %s", fieldMap["correlation_id"])
		}
	})

	t.Run("returns empty for no fields", func(t *testing.T) {
		ctx := Background()
		fields := ctx.LogFields()

		if len(fields) != 0 {
			t.Errorf("expected 0 fields, got %d", len(fields))
		}
	})

	t.Run("returns partial fields", func(t *testing.T) {
		ctx := Background().WithRequestID("req-only")
		fields := ctx.LogFields()

		if len(fields) != 2 {
			t.Fatalf("expected 2 fields, got %d", len(fields))
		}

		if fields[0] != "request_id" || fields[1] != "req-only" {
			t.Errorf("unexpected fields: %v", fields)
		}
	})
}

// ============================================================================
// Chaining Tests
// ============================================================================

func TestChaining(t *testing.T) {
	mock := &mockLogger{}

	ctx := Background().
		WithLogger(mock).
		WithRequestID("req-chain").
		WithUserID("user-chain").
		WithFields("extra", "value")

	ctx.Info("chained context")

	if ctx.RequestID() != "req-chain" {
		t.Errorf("expected RequestID 'req-chain', got %q", ctx.RequestID())
	}

	if ctx.UserID() != "user-chain" {
		t.Errorf("expected UserID 'user-chain', got %q", ctx.UserID())
	}

	if len(mock.infoCalls) != 1 {
		t.Fatalf("expected 1 info call, got %d", len(mock.infoCalls))
	}
}

// ============================================================================
// Operation Tests
// ============================================================================

func TestOperation(t *testing.T) {
	t.Run("WithOperation and GetOperation", func(t *testing.T) {
		c := context.Background()
		c = WithOperation(c, "CreateOrder")

		got := GetOperation(c)
		if got != "CreateOrder" {
			t.Errorf("expected 'CreateOrder', got %q", got)
		}
	})

	t.Run("GetOperation returns empty for missing", func(t *testing.T) {
		c := context.Background()
		got := GetOperation(c)

		if got != "" {
			t.Errorf("expected empty string, got %q", got)
		}
	})

	t.Run("Contextx methods", func(t *testing.T) {
		ctx := Background().WithOperation("GetUser")

		if ctx.Operation() != "GetUser" {
			t.Errorf("expected 'GetUser', got %q", ctx.Operation())
		}
	})
}

// ============================================================================
// Service Tests
// ============================================================================

func TestService(t *testing.T) {
	t.Run("WithService and GetService", func(t *testing.T) {
		c := context.Background()
		c = WithService(c, "order-service")

		got := GetService(c)
		if got != "order-service" {
			t.Errorf("expected 'order-service', got %q", got)
		}
	})

	t.Run("GetService returns empty for missing", func(t *testing.T) {
		c := context.Background()
		got := GetService(c)

		if got != "" {
			t.Errorf("expected empty string, got %q", got)
		}
	})

	t.Run("Contextx methods", func(t *testing.T) {
		ctx := Background().WithService("user-service")

		if ctx.Service() != "user-service" {
			t.Errorf("expected 'user-service', got %q", ctx.Service())
		}
	})
}

// ============================================================================
// Environment Tests
// ============================================================================

func TestEnvironment(t *testing.T) {
	t.Run("WithEnvironment and GetEnvironment", func(t *testing.T) {
		c := context.Background()
		c = WithEnvironment(c, "production")

		got := GetEnvironment(c)
		if got != "production" {
			t.Errorf("expected 'production', got %q", got)
		}
	})

	t.Run("GetEnvironment returns empty for missing", func(t *testing.T) {
		c := context.Background()
		got := GetEnvironment(c)

		if got != "" {
			t.Errorf("expected empty string, got %q", got)
		}
	})

	t.Run("Contextx methods", func(t *testing.T) {
		ctx := Background().WithEnvironment("staging")

		if ctx.Environment() != "staging" {
			t.Errorf("expected 'staging', got %q", ctx.Environment())
		}
	})
}

// ============================================================================
// Full Integration Test
// ============================================================================

func TestFullContextSetup(t *testing.T) {
	mock := &mockLogger{}

	ctx := Background().
		WithLogger(mock).
		WithService("order-service").
		WithEnvironment("production").
		WithOperation("CreateOrder").
		WithRequestID("req-123").
		WithTraceID("trace-456").
		WithUserID("user-789").
		WithCorrelationID("corr-000")

	ctx.Info("full context test")

	// Verify all values
	if ctx.Service() != "order-service" {
		t.Errorf("Service mismatch")
	}

	if ctx.Environment() != "production" {
		t.Errorf("Environment mismatch")
	}

	if ctx.Operation() != "CreateOrder" {
		t.Errorf("Operation mismatch")
	}

	if ctx.RequestID() != "req-123" {
		t.Errorf("RequestID mismatch")
	}

	if ctx.TraceID() != "trace-456" {
		t.Errorf("TraceID mismatch")
	}

	if ctx.UserID() != "user-789" {
		t.Errorf("UserID mismatch")
	}

	if ctx.CorrelationID() != "corr-000" {
		t.Errorf("CorrelationID mismatch")
	}

	// Verify LogFields includes all
	fields := ctx.LogFields()
	if len(fields) != 14 { // 7 key-value pairs
		t.Errorf("expected 14 fields, got %d", len(fields))
	}
}
