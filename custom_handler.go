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

// NewCustomHandler creates a new handler with a given configuration.
func NewCustomHandler(
	customOpts *CustomHandlerOptions,
	writer *bufio.Writer,
	replaceAttr CustomReplaceAttr,
) *CustomHandler {
	if customOpts == nil {
		customOpts = &CustomHandlerOptions{
			Level:     DefaultLogLevel,
			Enabled:   true,
			Pattern:   DefaultFormat,
			AddSource: false,
		}
	}

	if replaceAttr == nil {
		replaceAttr = GenerateDefaultCustomReplaceAttr(*customOpts, slog.TimeKey, slog.MessageKey)
	}

	sb := &strings.Builder{}
	return &CustomHandler{
		Opts: customOpts,
		sb:   sb,
		handler: slog.NewTextHandler(sb, &slog.HandlerOptions{
			Level:       GetSlogLevel(customOpts.Level),
			AddSource:   customOpts.AddSource,
			ReplaceAttr: replaceAttr,
		}),
		writer: writer,
	}
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

// GetSourceValue returns the source value.
func GetSourceValue(
	level slog.Level,
	sb *strings.Builder,
	getKeyValue func(string, *strings.Builder, bool) string,
) string {
	result := getKeyValue(slog.SourceKey, sb, true) // TODO: optimize to remove source key only once
	if level == LevelPerf {
		if fn, file, line, ok := GetPerfCallerInfo(); ok {
			return GetOtherSourceValue(fn, file, line)
		}
		return UnknownSource
	} else if result == "" || strings.HasSuffix(result, GenericLogFuncName) {
		if fn, file, line, ok := GetOtherCallerInfo(); ok {
			return GetOtherSourceValue(fn, file, line)
		}
		return UnknownSource
	}

	return result
}

// GetKeyValue returns the value of a key.
func (ch *CustomHandler) GetKeyValue(key string, sb *strings.Builder, removeKey bool) string {
	parts := strings.Fields(sb.String())
	for i, part := range parts {
		if strings.HasPrefix(part, key+"=") {
			value := strings.TrimPrefix(part, key+"=")
			if strings.HasPrefix(value, "\"") {
				value = strings.TrimPrefix(value, "\"")
				for j := i + 1; j < len(parts); j++ {
					value += " " + parts[j]
					if strings.HasSuffix(parts[j], "\"") {
						value = strings.TrimSuffix(value, "\"")
						break
					}
				}
			}
			if removeKey {
				parts = append(parts[:i], parts[i+1:]...)
				sb.Reset()
				sb.WriteString(strings.Join(parts, " "))
			}
			return value
		}
	}
	return ""
}

func getPatternForLevel(level slog.Level, pattern string) string {
	if pattern != "" {
		return pattern
	}

	switch level {
	case LevelPerf:
		return DefaultPerfFormat
	case slog.LevelDebug:
		return DefaultDebugFormat
	case slog.LevelError:
		return DefaultErrorFormat
	default:
		return DefaultFormat
	}
}

func buildOutput(
	pattern string,
	values map[string]string,
	sb *strings.Builder,
	level slog.Level,
) string {
	output := &strings.Builder{}
	result := pattern
	for placeholder, value := range values {
		if value != "" {
			value = AddPrefixSuffix(value)
			result = strings.ReplaceAll(result, placeholder, value)
		}
	}
	output.WriteString(result)

	if level == LevelPerf && !Contains(GetPlaceholders(pattern), PerfPlaceholder) {
		output.WriteString(" ")
		output.WriteString(DefaultPerfStartChar)
		output.WriteString(GetPerformanceMetrics())
		output.WriteString(DefaultPerfEndChar)
	}

	suffix := strings.TrimSuffix(sb.String(), "\n")
	if len(suffix) > 0 {
		output.WriteString(" ")
		output.WriteString(DefaultSuffixStartChar)
		output.WriteString(suffix)
		output.WriteString(DefaultSuffixEndChar)
	}

	return output.String()
}

// AddPrefixSuffix adds the prefix and suffix to the value.
func AddPrefixSuffix(value string) string {
	return DefaultValuePrefixChar + value + DefaultValueSuffixChar
}

// GetPlaceholderValues returns the placeholder values.
func GetPlaceholderValues(
	sb *strings.Builder,
	record slog.Record,
	placeholders []string,
	getKeyValue func(string, *strings.Builder, bool) string,
) map[string]string {
	values := make(map[string]string, len(placeholders))
	for _, placeholder := range placeholders {
		switch placeholder {
		case DatePlaceholder:
			values[DatePlaceholder] = record.Time.Format(DefaultDateFormat)
		case TimePlaceholder:
			values[TimePlaceholder] = record.Time.Format(DefaultTimeFormat)
		case DateTimePlaceholder:
			values[DateTimePlaceholder] = record.Time.Format(DefaultDateTimeFormat)
		case LevelPlaceholder:
			values[LevelPlaceholder] = getKeyValue(slog.LevelKey, sb, true)
		case MsgPlaceholder:
			values[MsgPlaceholder] = record.Message
		case PerfPlaceholder:
			values[PerfPlaceholder] = GetPerformanceMetrics()
		case SourcePlaceholder:
			values[SourcePlaceholder] = GetSourceValue(record.Level, sb, getKeyValue)
		}
	}
	return values
}

// ReplaceAttr is a function type for replacing attributes.
func ReplaceAttr(replaceMap map[string]string) CustomReplaceAttr {
	return func(_ []string, a slog.Attr) slog.Attr {
		if key, ok := replaceMap[a.Key]; ok {
			a.Key = key
		}
		return a
	}
}

// RemoveKeys returns a function that removes the time, level, source, and message keys.
func RemoveKeys() CustomReplaceAttr {
	return func(groups []string, a slog.Attr) slog.Attr {
		a = RemoveTimeKey(groups, a)
		a = RemoveLevelKey(groups, a)
		a = RemoveSourceKey(groups, a)
		a = RemoveMessageKey(groups, a)
		return a
	}
}

// RemoveGivenKeys returns a function that removes the given keys.
func RemoveGivenKeys(keys ...string) CustomReplaceAttr {
	return func(_ []string, a slog.Attr) slog.Attr {
		for _, key := range keys {
			a = Remove(key, a)
		}
		return a
	}
}

// GenerateDefaultCustomReplaceAttr returns a default CustomReplaceAttr function.
func GenerateDefaultCustomReplaceAttr(
	opts CustomHandlerOptions,
	keysToRemove ...string,
) CustomReplaceAttr {
	return func(_ []string, a slog.Attr) slog.Attr {
		if ContainsKey(keysToRemove, a.Key) {
			return slog.Attr{}
		}
		if a.Key == slog.LevelKey {
			level := GetLevelName(a.Value.Any().(slog.Level))
			if opts.UseSingleLetterLevel {
				a.Value = slog.StringValue(strings.ToUpper(level[:1]))
			} else {
				a.Value = slog.StringValue(strings.ToUpper(level))
			}
		}
		if opts.AddSource && a.Key == slog.SourceKey {
			switch source := a.Value.Any().(type) {
			case *slog.Source:
				if source.File != "" {
					f := BaseName(source.File)
					l := source.Line
					fn := source.Function
					fnParts := strings.Split(fn, "/")
					pkgAndFn := fnParts[len(fnParts)-1]
					a.Value = slog.StringValue(fmt.Sprintf("%s:%d:%s", f, l, pkgAndFn))
				}
			case string:
				a.Value = slog.StringValue(source)
			default:
			}
		}
		return a
	}
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
