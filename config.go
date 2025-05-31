package multilog

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Default time and date formats
const (
	DefaultTimeFormat     = "15:04:05"
	DefaultDateTimeFormat = "2006-01-02 15:04:05"
	DefaultDateFormat     = "2006-01-02"
)

// Default log file settings
const (
	DefaultLogFileSize    = 5 // in megabytes
	DefaultLogFileBackups = 1
	DefaultLogFileAge     = 1 // in days
)

// Log levels
const (
	DebugLevel = "debug"
	InfoLevel  = "info"
	WarnLevel  = "warn"
	ErrorLevel = "error"
	PerfLevel  = "perf"
)

// LogLevels contains all supported log levels.
var LogLevels = []string{DebugLevel, InfoLevel, WarnLevel, ErrorLevel, PerfLevel}

// Handler types
const (
	FileHandlerType    = "file"
	ConsoleHandlerType = "console"
)

// Subtypes for file handlers
const (
	TextHandlerSubType = "text"
	JSONHandlerSubType = "json"
)

// Default value prefix and suffix characters
const (
	DefaultValuePrefixChar = ""
	DefaultValueSuffixChar = ""
)

// Default suffix characters
const (
	DefaultSuffixStartChar = "["
	DefaultSuffixEndChar   = "]"
)

// Default performance metrics characters
const (
	DefaultPerfStartChar = "["
	DefaultPerfEndChar   = "]"
)

// Default placeholder prefix and suffix characters
const (
	DefaultPlaceholderPrefixChar = "["
	DefaultPlaceholderSuffixChar = "]"
)

// Pattern placeholders
const (
	DatePlaceholder     = "[date]"
	TimePlaceholder     = "[time]"
	DateTimePlaceholder = "[datetime]"
	LevelPlaceholder    = "[level]"
	MsgPlaceholder      = "[msg]"
	PerfPlaceholder     = "[perf]"
	SourcePlaceholder   = "[source]"
)

// Source placeholders
const (
	FileSource = "[file]"
	LineSource = "[line]"
	FuncSource = "[func]"
)

// Default log formats
const (
	DefaultFormat           = "[time] [level] [msg]"
	DefaultPerfFormat       = "[time] [level] [perf] [msg]"
	DefaultDebugFormat      = "[time] [level] [perf] [msg] [source]"
	DefaultErrorFormat      = "[time] [level] [msg] [source]"
	DefaultSourceFormat     = "[file]:[line]:[func]"
	DefaultPerfSourceFormat = DefaultSourceFormat
)

// DefaultPatternPlaceholders represents the default pattern placeholders.
var DefaultPatternPlaceholders = []string{
	DateTimePlaceholder,
	LevelPlaceholder,
	MsgPlaceholder,
	SourcePlaceholder,
}

// Config represents the configuration for the application.
type Config struct {
	Multilog LogConfig `yaml:"multilog"`
}

// LogConfig represents the logging configuration.
type LogConfig struct {
	Handlers []HandlerConfig `yaml:"handlers"`
}

// HandlerConfig represents the configuration for a specific handler.
type HandlerConfig struct {
	Type                 string `yaml:"type"`
	SubType              string `yaml:"subtype,omitempty"`
	Level                string `yaml:"level"`
	Enabled              bool   `yaml:"enabled"`
	UseSingleLetterLevel bool   `yaml:"use_single_letter_level,omitempty"`
	Pattern              string `yaml:"pattern,omitempty"`
	PatternPlaceholders  string `yaml:"pattern_placeholders,omitempty"`
	ValuePrefixChar      string `yaml:"value_prefix_char,omitempty"`
	ValueSuffixChar      string `yaml:"value_suffix_char,omitempty"`
	File                 string `yaml:"file,omitempty"`
	MaxSize              int    `yaml:"max_size,omitempty"`
	MaxBackups           int    `yaml:"max_backups,omitempty"`
	MaxAge               int    `yaml:"max_age,omitempty"`
}

// NewConfig loads the configuration from the specified YAML file.
func NewConfig(filename string) (*Config, error) {
	cleanedPath := filepath.Clean(filename)
	if _, err := os.Stat(cleanedPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file does not exist: %s", cleanedPath)
	}
	file, err := os.Open(cleanedPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer func() {
		if cerr := file.Close(); cerr != nil {
			if err == nil {
				err = fmt.Errorf("failed to close config file: %w", cerr)
			}
		}
	}()

	decoder := yaml.NewDecoder(file)
	var config Config
	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to decode config file: %w", err)
	}

	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid config data: %w", err)
	}

	return &config, nil
}

// NewConfigFromData loads the configuration from the provided YAML data.
func NewConfigFromData(data []byte) (*Config, error) {
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config data: %w", err)
	}
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid config data: %w", err)
	}
	return &config, nil
}

func newConsoleHandler(options CustomHandlerOptions) slog.Handler {
	return NewConsoleHandler(options)
}

func newFileHandler(options CustomHandlerOptions) (slog.Handler, error) {
	switch options.SubType {
	case JSONHandlerSubType:
		return NewJSONHandler(options, nil)
	case TextHandlerSubType:
		return NewFileHandler(options)
	default:
		return nil, fmt.Errorf("unknown handler subtype: %s", options.SubType)
	}
}

// GetEnabledHandlers returns the list of enabled handlers from the configuration.
func (c *Config) GetEnabledHandlers() []HandlerConfig {
	var enabledHandlers []HandlerConfig
	for _, handler := range c.Multilog.Handlers {
		if handler.Enabled {
			enabledHandlers = append(enabledHandlers, handler)
		}
	}
	return enabledHandlers
}

// GetCustomHandlerOptionsForHandler returns the custom handler options for the handler.
func (c *Config) GetCustomHandlerOptionsForHandler(
	handlerConfig HandlerConfig,
) (CustomHandlerOptions, error) {
	options := CustomHandlerOptions{
		Level:   handlerConfig.Level,
		SubType: defaultIfEmpty(handlerConfig.SubType, TextHandlerSubType),
		Enabled: handlerConfig.Enabled,
		Pattern: defaultIfEmpty(handlerConfig.Pattern, DefaultFormat),
		PatternPlaceholders: TrimSpaces(
			strings.Split(
				defaultIfEmpty(
					handlerConfig.PatternPlaceholders,
					strings.Join(DefaultPatternPlaceholders, ","),
				),
				",",
			),
		),
		AddSource:            handlerConfig.Type == FileHandlerType,
		UseSingleLetterLevel: handlerConfig.UseSingleLetterLevel,
		ValuePrefixChar:      defaultIfEmpty(handlerConfig.ValuePrefixChar, DefaultValuePrefixChar),
		ValueSuffixChar:      defaultIfEmpty(handlerConfig.ValueSuffixChar, DefaultValueSuffixChar),
		File:                 handlerConfig.File,
		MaxSize:              defaultIfZero(handlerConfig.MaxSize, DefaultLogFileSize),
		MaxBackups:           defaultIfZero(handlerConfig.MaxBackups, DefaultLogFileBackups),
		MaxAge:               defaultIfZero(handlerConfig.MaxAge, DefaultLogFileAge),
	}

	if handlerConfig.Type != ConsoleHandlerType && handlerConfig.Type != FileHandlerType {
		return CustomHandlerOptions{}, fmt.Errorf(
			"unknown handlerConfig type: %s",
			handlerConfig.Type,
		)
	}

	return options, nil
}

// defaultIfEmpty returns the default value if the value is empty.
func defaultIfEmpty(value, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}

// defaultIfZero returns the default value if the value is zero.
func defaultIfZero(value, defaultValue int) int {
	if value == 0 {
		return defaultValue
	}
	return value
}

// validateConfig validates the configuration and provides detailed error messages.
func validateConfig(config *Config) error {
	if err := validateHandlers(config.Multilog.Handlers); err != nil {
		return fmt.Errorf("handler validation failed: %w", err)
	}
	return nil
}

// validateHandlers validates the handlers and provides detailed error messages.
func validateHandlers(handlers []HandlerConfig) error {
	consoleHandlerCount := 0
	for i, handler := range handlers {
		if handler.Type == ConsoleHandlerType {
			consoleHandlerCount++
			if consoleHandlerCount > 1 {
				return fmt.Errorf("handler %d: only one console handler is allowed", i+1)
			}
		}

		if err := validateHandler(&handler); err != nil {
			return fmt.Errorf("handler %d: %w", i+1, err)
		}
	}

	return nil
}

// validateHandler validates the handler.
func validateHandler(handler *HandlerConfig) error {
	if handler.Type != ConsoleHandlerType && handler.Type != FileHandlerType {
		return fmt.Errorf("invalid handler type: %s", handler.Type)
	}

	if !Contains(LogLevels, handler.Level) {
		return fmt.Errorf("invalid log level: %s", handler.Level)
	}

	if handler.Type == FileHandlerType && handler.File == "" {
		return fmt.Errorf("file handler requires a file")
	}

	if handler.SubType != "" {
		if handler.Type == FileHandlerType && handler.SubType != TextHandlerSubType &&
			handler.SubType != JSONHandlerSubType {
			return fmt.Errorf("invalid file handler subtype: %s", handler.SubType)
		}
	}

	return nil
}

// TrimSpaces trims the spaces from the placeholders.
func TrimSpaces(placeholders []string) []string {
	for i, p := range placeholders {
		placeholders[i] = strings.TrimSpace(p)
	}
	return placeholders
}

// RemovePlaceholderChars removes the placeholder characters from the values.
func RemovePlaceholderChars(values map[string]interface{}) map[string]interface{} {
	for k, v := range values {
		if strings.HasPrefix(k, DefaultPlaceholderPrefixChar) &&
			strings.HasSuffix(k, DefaultPlaceholderSuffixChar) {
			delete(values, k)
			k = k[1 : len(k)-1]
			values[k] = v
		}
	}
	return values
}

// CreateHandlers creates the handlers based on the configuration.
func CreateHandlers(config *Config) ([]slog.Handler, error) {
	enabledHandlers := config.GetEnabledHandlers()
	hs := make([]slog.Handler, 0, len(enabledHandlers))

	for _, handlerConfig := range enabledHandlers {
		options, err := config.GetCustomHandlerOptionsForHandler(handlerConfig)
		if err != nil {
			return nil, err
		}

		handler, err := createHandler(handlerConfig.Type, options)
		if err != nil {
			return nil, err
		}

		hs = append(hs, handler)
	}

	return hs, nil
}

func createHandler(handlerType string, options CustomHandlerOptions) (slog.Handler, error) {
	switch handlerType {
	case ConsoleHandlerType:
		return newConsoleHandler(options), nil
	case FileHandlerType:
		handler, err := newFileHandler(options)
		if err != nil {
			return nil, err
		}
		return handler, nil
	default:
		return nil, fmt.Errorf("unknown handler type: %s", handlerType)
	}
}
