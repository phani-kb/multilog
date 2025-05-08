package multilog

// Default log formats
const (
	DefaultFormat           = "[time] [level] [msg]"
	DefaultPerfFormat       = "[time] [level] [perf] [msg]"
	DefaultDebugFormat      = "[time] [level] [perf] [msg] [source]"
	DefaultErrorFormat      = "[time] [level] [msg] [source]"
	DefaultSourceFormat     = "[file]:[line]:[func]"
	DefaultPerfSourceFormat = DefaultSourceFormat
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
	Type                 string `yaml:"type"`
	SubType              string `yaml:"subtype,omitempty"`
	Level                string `yaml:"level"`
	Enabled              bool   `yaml:"enabled"`
	UseSingleLetterLevel bool   `yaml:"use_single_letter_level,omitempty"`
	Pattern              string `yaml:"pattern,omitempty"`
	PatternPlaceholders  string `yaml:"pattern_placeholders,omitempty"`
	File                 string `yaml:"file,omitempty"`
	MaxSize              int    `yaml:"max_size,omitempty"`
	MaxBackups           int    `yaml:"max_backups,omitempty"`
	MaxAge               int    `yaml:"max_age,omitempty"`
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
