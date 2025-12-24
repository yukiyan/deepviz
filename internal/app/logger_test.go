package app

import (
	"testing"
)

// TestNewSlogLogger tests SlogLogger creation.
func TestNewSlogLogger(t *testing.T) {
	logger := NewSlogLogger(true, false)
	if logger == nil {
		t.Fatal("expected non-nil logger")
	}

	logger2 := NewSlogLogger(false, false)
	if logger2 == nil {
		t.Fatal("expected non-nil logger")
	}

	logger3 := NewSlogLogger(false, true)
	if logger3 == nil {
		t.Fatal("expected non-nil logger with trace enabled")
	}
}

// TestSlogLogger_Info tests SlogLogger Info method.
func TestSlogLogger_Info(t *testing.T) {
	logger := NewSlogLogger(true, false)
	logger.Info("test info message") // Verify no panic
	logger.Info("test with attrs", "key", "value", "number", 42)
}

// TestSlogLogger_Error tests SlogLogger Error method.
func TestSlogLogger_Error(t *testing.T) {
	logger := NewSlogLogger(false, false)
	logger.Error("test error message") // Verify no panic
	logger.Error("test error with attrs", "error", "something went wrong")
}

// TestSlogLogger_Debug tests SlogLogger Debug method.
func TestSlogLogger_Debug(t *testing.T) {
	logger := NewSlogLogger(true, false)
	logger.Debug("test debug message") // Verify no panic
	logger.Debug("test debug with attrs", "debug_key", "debug_value")
}

// TestSlogLogger_Trace tests SlogLogger Trace method.
func TestSlogLogger_Trace(t *testing.T) {
	logger := NewSlogLogger(false, true)
	logger.Trace("test trace message") // Verify no panic
	logger.Trace("test trace with attrs", "trace_key", "trace_value")
}

// TestNewNullLogger tests NullLogger creation.
func TestNewNullLogger(t *testing.T) {
	logger := NewNullLogger()
	if logger == nil {
		t.Fatal("expected non-nil logger")
	}
}

// TestNullLogger_Info tests NullLogger Info method.
func TestNullLogger_Info(t *testing.T) {
	logger := NewNullLogger()
	logger.Info("test info message") // Verify no panic
	logger.Info("test with attrs", "key", "value")
}

// TestNullLogger_Error tests NullLogger Error method.
func TestNullLogger_Error(t *testing.T) {
	logger := NewNullLogger()
	logger.Error("test error message") // Verify no panic
	logger.Error("test error with attrs", "key", "value")
}

// TestNullLogger_Debug tests NullLogger Debug method.
func TestNullLogger_Debug(t *testing.T) {
	logger := NewNullLogger()
	logger.Debug("test debug message") // Verify no panic
	logger.Debug("test debug with attrs", "key", "value")
}

// TestNullLogger_Trace tests NullLogger Trace method.
func TestNullLogger_Trace(t *testing.T) {
	logger := NewNullLogger()
	logger.Trace("test trace message") // Verify no panic
	logger.Trace("test trace with attrs", "key", "value")
}

// TestMockLogger tests mockLogger basic behavior.
func TestMockLogger(t *testing.T) {
	logger := newMockLogger()
	logger.Info("info1", "key1", "value1")
	logger.Info("info2")
	logger.Error("error1", "error_key", "error_value")
	logger.Debug("debug1")
	logger.Trace("trace1", "trace_key", "trace_value")

	// Verify logs are recorded
	if len(logger.buffer.entries) != 5 {
		t.Errorf("expected 5 log entries, got %d", len(logger.buffer.entries))
	}
}

// TestLoggerInterface_NullLogger tests that NullLogger implements Logger interface.
func TestLoggerInterface_NullLogger(t *testing.T) {
	var logger Logger = NewNullLogger()

	// Call Info/Error/Debug/Trace to ensure coverage
	logger.Info("test info")
	logger.Error("test error")
	logger.Debug("test debug")
	logger.Trace("test trace")

	// Verify no panic (OK if execution completes normally)
}

// TestLoggerInterface_SlogLogger tests that SlogLogger implements Logger interface.
func TestLoggerInterface_SlogLogger(t *testing.T) {
	var logger Logger = NewSlogLogger(true, true)

	// Call Info/Error/Debug/Trace to ensure coverage
	logger.Info("test info")
	logger.Error("test error")
	logger.Debug("test debug")
	logger.Trace("test trace")

	// Verify no panic (OK if execution completes normally)
}

// TestLoggerInterface_MockLogger tests that mockLogger implements Logger interface.
func TestLoggerInterface_MockLogger(t *testing.T) {
	var logger Logger = newMockLogger()

	// Call Info/Error/Debug/Trace to ensure coverage
	logger.Info("test info")
	logger.Error("test error")
	logger.Debug("test debug")
	logger.Trace("test trace")

	// Verify no panic (OK if execution completes normally)
}
