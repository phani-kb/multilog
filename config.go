package multilog

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Default time and date formats
const (
	DefaultTimeFormat     = "15:04:05"
	DefaultDateTimeFormat = "2006-01-02 15:04:05"
	DefaultDateFormat     = "2006-01-02"
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
	Type    string `yaml:"type"`
	SubType string `yaml:"subtype,omitempty"`
	Level   string `yaml:"level"`
	Enabled bool   `yaml:"enabled"`
	File    string `yaml:"file,omitempty"`
}

// NewConfig loads the configuration from the specified YAML file.
func NewConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filepath.Clean(filename))
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &config, nil
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

func validateConfig(config *Config) error {
	return validateHandlers(config.Multilog.Handlers)
}

func validateHandlers(handlers []HandlerConfig) error {
	consoleHandlerCount := 0
	for _, handler := range handlers {
		if handler.Type == ConsoleHandlerType {
			consoleHandlerCount++
			if consoleHandlerCount > 1 {
				return fmt.Errorf("only one console handler is allowed")
			}
		}
	}

	return nil
}
