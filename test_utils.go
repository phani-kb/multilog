package multilog

import (
	"context"
	"log/slog"
	"testing"
)

// TestHandler implements slog.Handler for testing purposes
type TestHandler struct {
	t           *testing.T
	called      bool
	lastMessage string
	lastLevel   slog.Level
}

// NewTestHandler creates a new TestHandler for testing purposes.
func NewTestHandler(t *testing.T) *TestHandler {
	return &TestHandler{t: t}
}

// Enabled reports whether the handler is enabled for the given level.
func (h *TestHandler) Enabled(_ context.Context, _ slog.Level) bool {
	return true
}

// Handle processes the log record and sets the called flag.
func (h *TestHandler) Handle(_ context.Context, r slog.Record) error {
	h.called = true
	h.lastMessage = r.Message
	h.lastLevel = r.Level
	return nil
}

// WithAttrs creates a new handler with the given attributes.
func (h *TestHandler) WithAttrs(_ []slog.Attr) slog.Handler {
	return h
}

// WithGroup creates a new handler with the given group name.
func (h *TestHandler) WithGroup(_ string) slog.Handler {
	return h
}

// Called returns whether the handler has been called.
func (h *TestHandler) Called() bool {
	return h.called
}

// Reset resets the handler's state for the next test.
func (h *TestHandler) Reset() {
	h.called = false
	h.lastMessage = ""
	h.lastLevel = 0
}

// NewTestLogger creates a new Logger instance with the given handler.
func NewTestLogger(t *testing.T) (*Logger, *TestHandler) {
	handler := NewTestHandler(t)
	logger := NewLogger(handler)
	return logger, handler
}
