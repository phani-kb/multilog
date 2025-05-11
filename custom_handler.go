package multilog

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"sync"

	"gopkg.in/natefinch/lumberjack.v2"
)

// CustomHandlerOptions contains configuration options for the handler.
type CustomHandlerOptions struct {
	Level                string
	SubType              string
	Enabled              bool
	Pattern              string
	PatternPlaceholders  []string
	AddSource            bool
	UseSingleLetterLevel bool
	ValuePrefixChar      string
	ValueSuffixChar      string
	File                 string
	MaxSize              int
	MaxBackups           int
	MaxAge               int
}

// CustomHandler is a base handler for logging.
type CustomHandler struct {
	Opts    *CustomHandlerOptions
	sb      *strings.Builder
	mu      sync.Mutex
	handler slog.Handler
	writer  *bufio.Writer
}

// CustomHandlerInterface is an interface for the custom handler.
type CustomHandlerInterface interface {
	slog.Handler
	GetOptions() *CustomHandlerOptions
	GetStringBuilder() *strings.Builder
	GetKeyValue(key string, sb *strings.Builder, removeKey bool) string
	GetWriter() *bufio.Writer
	GetSlogHandler() slog.Handler
}

// CustomReplaceAttr is a function type for replacing attributes.
type CustomReplaceAttr func(groups []string, a slog.Attr) slog.Attr

// NewCustomHandler creates a new handler with given configuration.
func NewCustomHandler(
	customOpts *CustomHandlerOptions,
	writer *bufio.Writer,
	replaceAttr CustomReplaceAttr,
) *CustomHandler {
	return nil
}

// Enabled determines if a log message should be logged based on its level.
func (ch *CustomHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return ch.handler.Enabled(ctx, level) && ch.Opts.Enabled
}

// Handle processes the log record and outputs it.
func (ch *CustomHandler) Handle(ctx context.Context, record slog.Record) error {
	if !ch.Enabled(ctx, record.Level) {
		return nil
	}
	ch.mu.Lock()
	defer func() {
		ch.sb.Reset()
		ch.mu.Unlock()
	}()

	if err := ch.handler.Handle(ctx, record); err != nil {
		return fmt.Errorf("failed to handle record: %w", err)
	}

	ch.Opts.Pattern = getPatternForLevel(record.Level, ch.Opts.Pattern)
	placeholders := GetPlaceholders(ch.Opts.Pattern)
	values := GetPlaceholderValues(ch.sb, record, placeholders, ch.GetKeyValue)

	output := buildOutput(ch.Opts.Pattern, values, ch.sb, record.Level)
	if _, err := ch.writer.WriteString(output + "\n"); err != nil {
		return fmt.Errorf("failed to write log message: %w", err)
	}

	if err := ch.writer.Flush(); err != nil {
		return fmt.Errorf("failed to flush writer: %w", err)
	}

	return nil
}

// WithAttrs adds attributes to the handler.
func (ch *CustomHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &CustomHandler{
		Opts:    ch.Opts,
		sb:      ch.sb,
		handler: ch.handler.WithAttrs(attrs),
		writer:  ch.writer,
	}
}

// WithGroup creates a new handler with grouped attributes.
func (ch *CustomHandler) WithGroup(name string) slog.Handler {
	return &CustomHandler{
		Opts:    ch.Opts,
		sb:      ch.sb,
		handler: ch.handler.WithGroup(name),
		writer:  ch.writer,
	}
}

// GetOptions returns the handler options.
func (ch *CustomHandler) GetOptions() *CustomHandlerOptions {
	return ch.Opts
}

// GetStringBuilder returns the handler string builder.
func (ch *CustomHandler) GetStringBuilder() *strings.Builder {
	return ch.sb
}

// GetWriter returns the handler writer.
func (ch *CustomHandler) GetWriter() *bufio.Writer {
	return ch.writer
}

// GetSlogHandler returns the handler slog.Handler.
func (ch *CustomHandler) GetSlogHandler() slog.Handler {
	return ch.handler
}

// GetPlaceholders returns the placeholders from the format.
func GetPlaceholders(format string) []string {
	re := regexp.MustCompile(`\[[a-z]+\]`)
	return re.FindAllString(format, -1)
}

// GetKeyValue returns the value of a key.
func (ch *CustomHandler) GetKeyValue(key string, sb *strings.Builder, removeKey bool) string {
	return ""
}

func getPatternForLevel(level slog.Level, pattern string) string {
	return DefaultFormat
}

func buildOutput(
	pattern string,
	values map[string]string,
	sb *strings.Builder,
	level slog.Level,
) string {
	output := &strings.Builder{}
	return output.String()
}

// GetPlaceholderValues returns the placeholder values.
func GetPlaceholderValues(
	sb *strings.Builder,
	record slog.Record,
	placeholders []string,
	getKeyValue func(string, *strings.Builder, bool) string,
) map[string]string {
	values := make(map[string]string, len(placeholders))
	return values
}

// CreateRotationWriter creates a rotation writer for the given options.
func CreateRotationWriter(opts CustomHandlerOptions) *bufio.Writer {
	logWriter := &lumberjack.Logger{
		Filename:   opts.File,
		MaxSize:    opts.MaxSize,
		MaxBackups: opts.MaxBackups,
		MaxAge:     opts.MaxAge,
		Compress:   false,
	}

	return bufio.NewWriter(logWriter)
}
