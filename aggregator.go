package multilog

import (
	"context"
	"log/slog"
)

// Aggregator is a handler that forwards logs to multiple handlers.
type Aggregator []slog.Handler

// NewAggregator creates a new aggregator with the given handlers.
func NewAggregator(handlers ...slog.Handler) Aggregator {
	return handlers
}

// Enabled implements slog.Handler.
func (a Aggregator) Enabled(ctx context.Context, level slog.Level) bool {
	for _, h := range a {
		if h.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

// Handle implements slog.Handler.
func (a Aggregator) Handle(ctx context.Context, r slog.Record) error {
	var firstErr error
	for _, h := range a {
		if h.Enabled(ctx, r.Level) {
			if err := h.Handle(ctx, r); err != nil && firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}

// WithAttrs implements slog.Handler.
func (a Aggregator) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make([]slog.Handler, len(a))
	for i, h := range a {
		handlers[i] = h.WithAttrs(attrs)
	}
	return NewAggregator(handlers...)
}

// WithGroup implements slog.Handler.
func (a Aggregator) WithGroup(name string) slog.Handler {
	handlers := make([]slog.Handler, len(a))
	for i, h := range a {
		handlers[i] = h.WithGroup(name)
	}
	return NewAggregator(handlers...)
}

// PerfMetrics represents performance metrics.
type PerfMetrics struct {
	NumGoroutines int
	NumCPUs       int
	MaxThreads    int
	GCHeapAllocs  uint64
	GCHeapFrees   uint64
	HeapFree      uint64
	HeapObjects   uint64
	HeapReleased  uint64
	HeapUnused    uint64
	TotalMemory   uint64
	TotalCPUUsage float64
	UserCPUUsage  float64
}

// SourceInfo represents the source information of a log record.
type SourceInfo struct {
	File string
	Line int
}

// LogRecord represents a log record with additional information.
type LogRecord struct {
	Time        string
	Level       slog.Level
	Message     string
	Source      SourceInfo
	PerfMetrics PerfMetrics
	Format      LogRecordFormat
}

// LogRecordFormat represents the format of a log record.
type LogRecordFormat struct {
	Pattern string
}
