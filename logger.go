package multilog

import (
	"context"
	"fmt"
	"log/slog"
)

// Log level constants
const (
	UnknownLevel     = "unknown"
	UnknownSource    = "unknown"
	DefaultSLogLevel = slog.LevelInfo
	DefaultLogLevel  = "info"
	LevelPerf        = slog.Level(-1)
)

// PackagePrefix is the prefix used for package-level logging.
const PackagePrefix = "multilog.(*Logger)."

// GenericLogFuncName is the function name used for generic logging.
const GenericLogFuncName = "multilog.(*Logger).log"

// CallIdentifiers contains the identifiers for various logging functions.
var CallIdentifiers = []string{
	"Perf(",
	"Perff(",
	"Infof(",
	"Warnf(",
	"Debugf(",
	"Errorf(",
}

// LevelNamesMap maps slog.Level to string.
var LevelNamesMap = map[slog.Leveler]string{
	slog.LevelDebug: "debug",
	slog.LevelInfo:  "info",
	slog.LevelWarn:  "warn",
	slog.LevelError: "error",
	LevelPerf:       "perf",
}

// LevelMap maps string to slog.Level.
var LevelMap = map[string]slog.Level{
	"debug": slog.LevelDebug,
	"info":  slog.LevelInfo,
	"warn":  slog.LevelWarn,
	"error": slog.LevelError,
	"perf":  LevelPerf,
}

// LoggerInterface defines the logging methods available
type LoggerInterface interface {
	Perff(msg string, args ...any)
	Perf(msg string, args ...any)
	Infof(msg string, args ...any)
	Warnf(msg string, args ...any)
	Errorf(msg string, args ...any)
	Debugf(msg string, args ...any)

	// Debug Structured logging methods
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)

	// WithContext Context-aware logging methods
	WithContext(ctx context.Context) LoggerInterface
	PerfContext(ctx context.Context, msg string, args ...any)
	InfoContext(ctx context.Context, msg string, args ...any)
	WarnContext(ctx context.Context, msg string, args ...any)
	ErrorContext(ctx context.Context, msg string, args ...any)
	DebugContext(ctx context.Context, msg string, args ...any)

	// WithField Structured logging methods
	WithField(key string, value any) LoggerInterface
	WithFields(fields map[string]any) LoggerInterface

	// GetLogger Return the underlying logger for advanced usage
	GetLogger() *slog.Logger
}

// Ensure Logger implements LoggerInterface
var _ LoggerInterface = (*Logger)(nil)

// Logger wraps slog.Logger and allows configuration of handlers.
type Logger struct {
	Logger *slog.Logger
	attrs  []any
}

// NewLogger creates a new logger with the specified handlers.
func NewLogger(handlers ...slog.Handler) *Logger {
	var enabledHandlers []slog.Handler
	for _, handler := range handlers {
		switch h := handler.(type) {
		case *ConsoleHandler:
			if customHandler, ok := h.Handler.(*CustomHandler); ok && customHandler.Opts.Enabled {
				enabledHandlers = append(enabledHandlers, handler)
			}
		case *FileHandler:
			if customHandler, ok := h.Handler.(*CustomHandler); ok && customHandler.Opts.Enabled {
				enabledHandlers = append(enabledHandlers, handler)
			}
		case *JSONHandler:
			if customHandler, ok := h.Handler.(*CustomHandler); ok && customHandler.Opts.Enabled {
				enabledHandlers = append(enabledHandlers, handler)
			}
		default:
			// Accept other handler types directly (useful for testing)
			enabledHandlers = append(enabledHandlers, handler)
		}
	}
	return &Logger{Logger: slog.New(NewAggregator(enabledHandlers...))}
}

// WithLevel returns a new logger with the specified minimum level
func (l *Logger) WithLevel(_ slog.Level) *Logger {
	newLogger := *l
	return &newLogger
}

// WithContext returns a logger with the context attached
func (l *Logger) WithContext(ctx context.Context) LoggerInterface {
	return &ContextLogger{
		Logger: l,
		ctx:    ctx,
	}
}

// WithField returns a logger with the specified field attached to all messages
func (l *Logger) WithField(key string, value any) LoggerInterface {
	newLogger := &Logger{
		Logger: l.Logger.With(key, value),
		attrs:  append(l.attrs, key, value),
	}
	return newLogger
}

// WithFields returns a logger with the specified fields attached to all messages
func (l *Logger) WithFields(fields map[string]any) LoggerInterface {
	// Create a new slice with double the capacity
	args := make([]any, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}

	newLogger := &Logger{
		Logger: l.Logger.With(args...),
		attrs:  append(l.attrs, args...),
	}
	return newLogger
}

// GetLogger returns the underlying slog.Logger
func (l *Logger) GetLogger() *slog.Logger {
	return l.Logger
}

// log logs a message at the given level.
func (l *Logger) log(level slog.Level, msg string, args ...any) {
	// Format the message but don't add any attributes
	l.Logger.Log(context.Background(), level, fmt.Sprintf(msg, args...))
}

// logContext logs a message with context at the given level
func (l *Logger) logContext(ctx context.Context, level slog.Level, msg string, args ...any) {
	// Log without level check
	l.Logger.Log(ctx, level, fmt.Sprintf(msg, args...))
}

// Perff logs performance metrics dynamically.
func (l *Logger) Perff(msg string, args ...any) {
	l.log(LevelPerf, msg, args...)
}

// Perf logs performance metrics statically.
func (l *Logger) Perf(msg string, args ...any) {
	l.Logger.Log(context.Background(), LevelPerf, msg, args...)
}

// Infof logs an informational message.
func (l *Logger) Infof(msg string, args ...any) {
	l.log(slog.LevelInfo, msg, args...)
}

// Warnf logs a warning message.
func (l *Logger) Warnf(msg string, args ...any) {
	l.log(slog.LevelWarn, msg, args...)
}

// Debugf logs a debug message.
func (l *Logger) Debugf(msg string, args ...any) {
	l.log(slog.LevelDebug, msg, args...)
}

// Errorf logs an error message.
func (l *Logger) Errorf(msg string, args ...any) {
	l.log(slog.LevelError, msg, args...)
}

// Debug logs a debug message with structured key-value pairs.
func (l *Logger) Debug(msg string, args ...any) {
	l.Logger.Debug(msg, args...)
}

// Info logs an informational message with structured key-value pairs.
func (l *Logger) Info(msg string, args ...any) {
	l.Logger.Info(msg, args...)
}

// Warn logs a warning message with structured key-value pairs.
func (l *Logger) Warn(msg string, args ...any) {
	l.Logger.Warn(msg, args...)
}

// Error logs an error message with structured key-value pairs.
func (l *Logger) Error(msg string, args ...any) {
	l.Logger.Error(msg, args...)
}

// PerfContext logs performance metrics dynamically.
func (l *Logger) PerfContext(ctx context.Context, msg string, args ...any) {
	l.logContext(ctx, LevelPerf, msg, args...)
}

// InfoContext logs an informational message with structured key-value pairs.
func (l *Logger) InfoContext(ctx context.Context, msg string, args ...any) {
	l.logContext(ctx, slog.LevelInfo, msg, args...)
}

// WarnContext logs a warning message with structured key-value pairs.
func (l *Logger) WarnContext(ctx context.Context, msg string, args ...any) {
	l.logContext(ctx, slog.LevelWarn, msg, args...)
}

// ErrorContext logs an error message with structured key-value pairs.
func (l *Logger) ErrorContext(ctx context.Context, msg string, args ...any) {
	l.logContext(ctx, slog.LevelError, msg, args...)
}

// DebugContext logs a debug message with structured key-value pairs.
func (l *Logger) DebugContext(ctx context.Context, msg string, args ...any) {
	l.logContext(ctx, slog.LevelDebug, msg, args...)
}

// GetSlogLevel returns the slog.Level for the given level string.
func GetSlogLevel(level string) slog.Level {
	if l, ok := LevelMap[level]; ok {
		return l
	}
	return DefaultSLogLevel
}

// GetLevelName returns the string representation of the slog.Level.
func GetLevelName(level slog.Level) string {
	if name, ok := LevelNamesMap[level]; ok {
		return name
	}
	return UnknownLevel
}

// Remove removes the given key from the attributes.
func Remove(key string, a slog.Attr) slog.Attr {
	if a.Key == key {
		return slog.Attr{}
	}
	return a
}

// RemoveKey returns a function that removes the specified attribute key.
func RemoveKey(key string) func(groups []string, a slog.Attr) slog.Attr {
	return func(_ []string, a slog.Attr) slog.Attr {
		return Remove(key, a)
	}
}

// RemoveTimeKey, RemoveLevelKey, RemoveSourceKey, and RemoveMessageKey are functions
var (
	RemoveTimeKey    = RemoveKey(slog.TimeKey)
	RemoveLevelKey   = RemoveKey(slog.LevelKey)
	RemoveSourceKey  = RemoveKey(slog.SourceKey)
	RemoveMessageKey = RemoveKey(slog.MessageKey)
)

// ContainsKey checks if the given key is present in the keys.
func ContainsKey(keys []string, key string) bool {
	for _, k := range keys {
		if k == key {
			return true
		}
	}
	return false
}

// ContextLogger wraps a Logger with a context
type ContextLogger struct {
	*Logger
	ctx context.Context
}

// Perff logs performance metrics dynamically.
func (l *ContextLogger) Perff(msg string, args ...any) {
	l.logContext(l.ctx, LevelPerf, msg, args...)
}

// Perf logs performance metrics statically.
func (l *ContextLogger) Perf(msg string, args ...any) {
	l.Logger.Logger.LogAttrs(l.ctx, LevelPerf, msg, slog.Any("args", args))
}

// Infof logs an informational message with structured key-value pairs.
func (l *ContextLogger) Infof(msg string, args ...any) {
	l.logContext(l.ctx, slog.LevelInfo, msg, args...)
}

// Warnf logs a warning message with structured key-value pairs.
func (l *ContextLogger) Warnf(msg string, args ...any) {
	l.logContext(l.ctx, slog.LevelWarn, msg, args...)
}

// Errorf logs an error message with structured key-value pairs.
func (l *ContextLogger) Errorf(msg string, args ...any) {
	l.logContext(l.ctx, slog.LevelError, msg, args...)
}

// Debugf logs a debug message with structured key-value pairs.
func (l *ContextLogger) Debugf(msg string, args ...any) {
	l.logContext(l.ctx, slog.LevelDebug, msg, args...)
}

// Debug logs a debug message with structured key-value pairs.
func (l *ContextLogger) Debug(msg string, args ...any) {
	l.Logger.Logger.DebugContext(l.ctx, msg, args...)
}

// Info logs an informational message with structured key-value pairs.
func (l *ContextLogger) Info(msg string, args ...any) {
	l.Logger.Logger.InfoContext(l.ctx, msg, args...)
}

// Warn logs a warning message with structured key-value pairs.
func (l *ContextLogger) Warn(msg string, args ...any) {
	l.Logger.Logger.WarnContext(l.ctx, msg, args...)
}

// Error logs an error message with structured key-value pairs.
func (l *ContextLogger) Error(msg string, args ...any) {
	l.Logger.Logger.ErrorContext(l.ctx, msg, args...)
}

// PerfContext logs performance metrics dynamically.
func (l *ContextLogger) PerfContext(ctx context.Context, msg string, args ...any) {
	l.logContext(ctx, LevelPerf, msg, args...)
}

// InfoContext logs an informational message
func (l *ContextLogger) InfoContext(ctx context.Context, msg string, args ...any) {
	l.logContext(ctx, slog.LevelInfo, msg, args...)
}

// WarnContext logs a warning message
func (l *ContextLogger) WarnContext(ctx context.Context, msg string, args ...any) {
	l.logContext(ctx, slog.LevelWarn, msg, args...)
}

// ErrorContext logs an error message
func (l *ContextLogger) ErrorContext(ctx context.Context, msg string, args ...any) {
	l.logContext(ctx, slog.LevelError, msg, args...)
}

// DebugContext logs a debug message
func (l *ContextLogger) DebugContext(ctx context.Context, msg string, args ...any) {
	l.logContext(ctx, slog.LevelDebug, msg, args...)
}

// WithContext for ContextLogger should return a new ContextLogger with the new context
func (l *ContextLogger) WithContext(ctx context.Context) LoggerInterface {
	return &ContextLogger{
		Logger: l.Logger,
		ctx:    ctx,
	}
}

// WithField ensures we maintain the context when adding fields
func (l *ContextLogger) WithField(key string, value any) LoggerInterface {
	newLogger := l.Logger.WithField(key, value).(*Logger)
	return &ContextLogger{
		Logger: newLogger,
		ctx:    l.ctx,
	}
}

// WithFields ensures we maintain the context when adding fields
func (l *ContextLogger) WithFields(fields map[string]any) LoggerInterface {
	newLogger := l.Logger.WithFields(fields).(*Logger)
	return &ContextLogger{
		Logger: newLogger,
		ctx:    l.ctx,
	}
}
