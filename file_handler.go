package multilog

import (
	"context"
	"log/slog"
)

// FileHandler is a Handler for file logging.
type FileHandler struct {
	Handler CustomHandlerInterface
}

// NewFileHandler creates a file Handler with the specified options.
func NewFileHandler(opts CustomHandlerOptions) (slog.Handler, error) {
	writer := CreateRotationWriter(opts)

	return &FileHandler{
		Handler: NewCustomHandler(&opts, writer, nil),
	}, nil
}

// Enabled checks if the handler is enabled for the given level.
func (fh *FileHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return fh.Handler.Enabled(ctx, level)
}

// Handle processes the log record and writes it to the file handler.
func (fh *FileHandler) Handle(ctx context.Context, record slog.Record) error {
	return fh.Handler.Handle(ctx, record)
}

// WithAttrs creates a new handler with the given attributes.
func (fh *FileHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return fh.Handler.WithAttrs(attrs)
}

// WithGroup creates a new handler with the given group name.
func (fh *FileHandler) WithGroup(name string) slog.Handler {
	return fh.Handler.WithGroup(name)
}
