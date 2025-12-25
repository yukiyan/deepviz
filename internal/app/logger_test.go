package app

import (
	"testing"
)

// TestNewSlogLogger tests SlogLogger creation.
func TestNewSlogLogger(t *testing.T) {
	logger := NewSlogLogger(true, "")
	if logger == nil {
		t.Fatal("expected non-nil logger")
	}

	logger2 := NewSlogLogger(false, "")
	if logger2 == nil {
		t.Fatal("expected non-nil logger")
	}

	// Test with log file
	logFile := t.TempDir() + "/test.log"
	logger3 := NewSlogLogger(false, logFile)
	if logger3 == nil {
		t.Fatal("expected non-nil logger with log file")
	}
}

// TestSlogLogger_Info tests SlogLogger Info method.
func TestSlogLogger_Info(t *testing.T) {
	logger := NewSlogLogger(true, "")
	logger.Info("test info message") // Verify no panic
	logger.Info("test with attrs", "key", "value", "number", 42)
}

// TestSlogLogger_Error tests SlogLogger Error method.
func TestSlogLogger_Error(t *testing.T) {
	logger := NewSlogLogger(false, "")
	logger.Error("test error message") // Verify no panic
	logger.Error("test error with attrs", "error", "something went wrong")
}

// TestSlogLogger_Debug tests SlogLogger Debug method.
func TestSlogLogger_Debug(t *testing.T) {
	logger := NewSlogLogger(true, "")
	logger.Debug("test debug message") // Verify no panic
	logger.Debug("test debug with attrs", "debug_key", "debug_value")
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

// TestMockLogger tests mockLogger basic behavior.
func TestMockLogger(t *testing.T) {
	logger := newMockLogger()
	logger.Info("info1", "key1", "value1")
	logger.Info("info2")
	logger.Error("error1", "error_key", "error_value")
	logger.Debug("debug1")

	// Verify logs are recorded
	if len(logger.buffer.entries) != 4 {
		t.Errorf("expected 4 log entries, got %d", len(logger.buffer.entries))
	}
}

// TestLoggerInterface_NullLogger tests that NullLogger implements Logger interface.
func TestLoggerInterface_NullLogger(t *testing.T) {
	var logger Logger = NewNullLogger()

	// Call Info/Error/Debug to ensure coverage
	logger.Info("test info")
	logger.Error("test error")
	logger.Debug("test debug")

	// Verify no panic (OK if execution completes normally)
}

// TestLoggerInterface_SlogLogger tests that SlogLogger implements Logger interface.
func TestLoggerInterface_SlogLogger(t *testing.T) {
	var logger Logger = NewSlogLogger(true, "")

	// Call Info/Error/Debug to ensure coverage
	logger.Info("test info")
	logger.Error("test error")
	logger.Debug("test debug")

	// Verify no panic (OK if execution completes normally)
}

// TestLoggerInterface_MockLogger tests that mockLogger implements Logger interface.
func TestLoggerInterface_MockLogger(t *testing.T) {
	var logger Logger = newMockLogger()

	// Call Info/Error/Debug to ensure coverage
	logger.Info("test info")
	logger.Error("test error")
	logger.Debug("test debug")

	// Verify no panic (OK if execution completes normally)
}
