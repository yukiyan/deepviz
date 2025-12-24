package app

import (
	"context"
	"io"
	"log/slog"
	"os"
)

// LevelTrace is a custom log level for detailed tracing.
const LevelTrace = slog.Level(-8)

// Logger is an interface for structured logging.
type Logger interface {
	Info(msg string, args ...any)
	Error(msg string, args ...any)
	Debug(msg string, args ...any)
	Trace(msg string, args ...any)
}

// SlogLogger is a logger that uses slog.
type SlogLogger struct {
	logger *slog.Logger
}

// NewSlogLogger creates a new SlogLogger with JSON output.
func NewSlogLogger(verbose bool, trace bool) *SlogLogger {
	level := slog.LevelInfo
	if verbose {
		level = slog.LevelDebug
	}
	if trace {
		level = LevelTrace
	}

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})

	return &SlogLogger{
		logger: slog.New(handler),
	}
}

// Info outputs an information log.
func (l *SlogLogger) Info(msg string, args ...any) {
	l.logger.Info(msg, args...)
}

// Error outputs an error log.
func (l *SlogLogger) Error(msg string, args ...any) {
	l.logger.Error(msg, args...)
}

// Debug outputs a debug log.
func (l *SlogLogger) Debug(msg string, args ...any) {
	l.logger.Debug(msg, args...)
}

// Trace outputs a trace log.
func (l *SlogLogger) Trace(msg string, args ...any) {
	l.logger.Log(context.Background(), LevelTrace, msg, args...)
}

// NullLogger is a logger that outputs nothing (for testing).
type NullLogger struct{}

// NewNullLogger creates a new NullLogger.
func NewNullLogger() *NullLogger {
	return &NullLogger{}
}

// Info does nothing.
func (l *NullLogger) Info(msg string, args ...any) {}

// Error does nothing.
func (l *NullLogger) Error(msg string, args ...any) {}

// Debug does nothing.
func (l *NullLogger) Debug(msg string, args ...any) {}

// Trace does nothing.
func (l *NullLogger) Trace(msg string, args ...any) {}

// mockLogger is a mock logger for testing.
type mockLogger struct {
	logger *slog.Logger
	buffer *mockLogBuffer
}

// mockLogBuffer captures log output.
type mockLogBuffer struct {
	entries []mockLogEntry
}

type mockLogEntry struct {
	level   slog.Level
	message string
	attrs   map[string]any
}

func (b *mockLogBuffer) Write(p []byte) (n int, err error) {
	return len(p), nil
}

// newMockLogger creates a new mockLogger.
func newMockLogger() *mockLogger {
	buffer := &mockLogBuffer{entries: []mockLogEntry{}}
	handler := &mockLogHandler{buffer: buffer}
	logger := slog.New(handler)

	return &mockLogger{
		logger: logger,
		buffer: buffer,
	}
}

// Info records an information log.
func (m *mockLogger) Info(msg string, args ...any) {
	m.logger.Info(msg, args...)
}

// Error records an error log.
func (m *mockLogger) Error(msg string, args ...any) {
	m.logger.Error(msg, args...)
}

// Debug records a debug log.
func (m *mockLogger) Debug(msg string, args ...any) {
	m.logger.Debug(msg, args...)
}

// Trace records a trace log.
func (m *mockLogger) Trace(msg string, args ...any) {
	m.logger.Log(context.Background(), LevelTrace, msg, args...)
}

// mockLogHandler is a custom slog handler for testing.
type mockLogHandler struct {
	buffer *mockLogBuffer
	attrs  []slog.Attr
	groups []string
}

func (h *mockLogHandler) Enabled(_ context.Context, _ slog.Level) bool {
	return true
}

func (h *mockLogHandler) Handle(_ context.Context, r slog.Record) error {
	attrs := make(map[string]any)
	for _, attr := range h.attrs {
		attrs[attr.Key] = attr.Value.Any()
	}
	r.Attrs(func(a slog.Attr) bool {
		attrs[a.Key] = a.Value.Any()
		return true
	})

	entry := mockLogEntry{
		level:   r.Level,
		message: r.Message,
		attrs:   attrs,
	}
	h.buffer.entries = append(h.buffer.entries, entry)
	return nil
}

func (h *mockLogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := make([]slog.Attr, len(h.attrs)+len(attrs))
	copy(newAttrs, h.attrs)
	copy(newAttrs[len(h.attrs):], attrs)
	return &mockLogHandler{
		buffer: h.buffer,
		attrs:  newAttrs,
		groups: h.groups,
	}
}

func (h *mockLogHandler) WithGroup(name string) slog.Handler {
	newGroups := make([]string, len(h.groups)+1)
	copy(newGroups, h.groups)
	newGroups[len(h.groups)] = name
	return &mockLogHandler{
		buffer: h.buffer,
		attrs:  h.attrs,
		groups: newGroups,
	}
}

var _ io.Writer = (*mockLogBuffer)(nil)
var _ slog.Handler = (*mockLogHandler)(nil)
