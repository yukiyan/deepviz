package app

import (
	"context"
	"io"
	"log/slog"
	"os"
)

// Logger is an interface for structured logging.
type Logger interface {
	Info(msg string, args ...any)
	Error(msg string, args ...any)
	Debug(msg string, args ...any)
}

// SlogLogger is a logger that uses slog.
type SlogLogger struct {
	logger *slog.Logger
}

// NewSlogLogger creates a new SlogLogger with JSON output.
// Logs to both stdout and file. File output is always at DEBUG level.
func NewSlogLogger(verbose bool, logFilePath string) *SlogLogger {
	stdoutLevel := slog.LevelInfo
	if verbose {
		stdoutLevel = slog.LevelDebug
	}

	// Create stdout handler
	stdoutHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: stdoutLevel,
	})

	// If log file path is provided, create file handler and multi-handler
	if logFilePath != "" {
		logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			// If file creation fails, fall back to stdout only
			return &SlogLogger{
				logger: slog.New(stdoutHandler),
			}
		}

		// File handler always logs at DEBUG level
		fileHandler := slog.NewJSONHandler(logFile, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})

		// Use multi-handler to write to both stdout and file
		multiHandler := &multiHandler{
			handlers: []slog.Handler{stdoutHandler, fileHandler},
		}

		return &SlogLogger{
			logger: slog.New(multiHandler),
		}
	}

	return &SlogLogger{
		logger: slog.New(stdoutHandler),
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

// multiHandler is a slog.Handler that writes to multiple handlers.
type multiHandler struct {
	handlers []slog.Handler
}

func (h *multiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	// Enable if any handler is enabled for this level
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (h *multiHandler) Handle(ctx context.Context, r slog.Record) error {
	// Write to all handlers
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, r.Level) {
			if err := handler.Handle(ctx, r.Clone()); err != nil {
				return err
			}
		}
	}
	return nil
}

func (h *multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newHandlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		newHandlers[i] = handler.WithAttrs(attrs)
	}
	return &multiHandler{handlers: newHandlers}
}

func (h *multiHandler) WithGroup(name string) slog.Handler {
	newHandlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		newHandlers[i] = handler.WithGroup(name)
	}
	return &multiHandler{handlers: newHandlers}
}

var _ slog.Handler = (*multiHandler)(nil)
