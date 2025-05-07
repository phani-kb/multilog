package multilog

import (
	"bufio"
	"context"
	"log/slog"
	"os"
)

// ConsoleHandler is a Handler for console logging.
type ConsoleHandler struct {
	Handler CustomHandlerInterface
}

// NewConsoleHandler creates a console Handler with the specified options.
func NewConsoleHandler(opts CustomHandlerOptions) slog.Handler {
	return &ConsoleHandler{
		Handler: NewCustomHandler(&opts, bufio.NewWriter(os.Stdout), nil),
	}
}

// Enabled checks if the handler is enabled for the given level.
func (ch *ConsoleHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return ch.Handler.Enabled(ctx, level)
}

// Handle processes the log record and writes it to the console handler.
func (ch *ConsoleHandler) Handle(ctx context.Context, record slog.Record) error {
	return ch.Handler.Handle(ctx, record)
}

// WithAttrs creates a new handler with the given attributes.
func (ch *ConsoleHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return ch.Handler.WithAttrs(attrs)
}

// WithGroup creates a new handler with the given group name.
func (ch *ConsoleHandler) WithGroup(name string) slog.Handler {
	return ch.Handler.WithGroup(name)
}
