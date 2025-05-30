package multilog

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewConfig_ComplexSuccess(t *testing.T) {
	testFile := filepath.Join(t.TempDir(), "complex_config.yaml")
	configData := []byte(`multilog:
  handlers:
    - type: console
      level: info
      enabled: true
      pattern: "[datetime] [[level]] [msg]"
    - type: file
      subtype: text
      level: error
      enabled: true
      pattern: "[date] - [[time]] [[level]] [[source]] [msg]"
      file: error.log
      max_size: 5 # MB
      max_backups: 7
      max_age: 1 # days
    - type: file
      subtype: json
      level: debug
      enabled: true
      pattern_placeholders: "[datetime], [level], [source], [msg]"
      file: test.json
      max_size: 5 # MB
      max_backups: 7
      max_age: 1 # days`)

	err := os.WriteFile(testFile, configData, 0o644)
	assert.NoError(t, err)

	cfg, err := NewConfig(testFile)
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Len(t, cfg.Multilog.Handlers, 3)

	// Verify console handler
	console := cfg.Multilog.Handlers[0]
	assert.Equal(t, ConsoleHandlerType, console.Type)
	assert.Equal(t, InfoLevel, console.Level)
	assert.True(t, console.Enabled)
	assert.Equal(t, "[datetime] [[level]] [msg]", console.Pattern)
	assert.Equal(t, "", console.PatternPlaceholders)

	// Verify text file handler
	textFile := cfg.Multilog.Handlers[1]
	assert.Equal(t, FileHandlerType, textFile.Type)
	assert.Equal(t, TextHandlerSubType, textFile.SubType)
	assert.Equal(t, ErrorLevel, textFile.Level)
	assert.True(t, textFile.Enabled)
	assert.Equal(t, "error.log", textFile.File)
	assert.Equal(t, "[date] - [[time]] [[level]] [[source]] [msg]", textFile.Pattern)

	// Verify JSON file handler
	jsonFile := cfg.Multilog.Handlers[2]
	assert.Equal(t, FileHandlerType, jsonFile.Type)
	assert.Equal(t, JSONHandlerSubType, jsonFile.SubType)
	assert.Equal(t, DebugLevel, jsonFile.Level)
	assert.True(t, jsonFile.Enabled)
	assert.Equal(t, "test.json", jsonFile.File)
	assert.Equal(t, 5, jsonFile.MaxSize)
	assert.Equal(t, 7, jsonFile.MaxBackups)
	assert.Equal(t, 1, jsonFile.MaxAge)
}

func TestNewConfig_FileError(t *testing.T) {
	_, err := NewConfig("nonexistent_file.yaml")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read config file")
}

func TestNewConfig_UnmarshalError(t *testing.T) {
	testFile := filepath.Join(t.TempDir(), "invalid_config.yaml")
	invalidData := []byte(`:invalid_yaml`)
	err := os.WriteFile(testFile, invalidData, 0o644)
	assert.NoError(t, err)

	_, err = NewConfig(testFile)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal config")
}

func TestValidateHandler_Valid(t *testing.T) {
	handler := &HandlerConfig{
		Type:    ConsoleHandlerType,
		Level:   InfoLevel,
		Enabled: true,
	}
	err := validateHandler(handler)
	assert.NoError(t, err)
}

func TestValidateHandler_InvalidType(t *testing.T) {
	handler := &HandlerConfig{
		Type:  "unknown",
		Level: InfoLevel,
	}
	err := validateHandler(handler)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid handler type")
}

func TestValidateHandler_InvalidLevel(t *testing.T) {
	handler := &HandlerConfig{
		Type:  ConsoleHandlerType,
		Level: "invalid",
	}
	err := validateHandler(handler)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid log level")
}

func TestValidateHandler_FileHandlerNoFile(t *testing.T) {
	handler := &HandlerConfig{
		Type:  FileHandlerType,
		Level: InfoLevel,
	}
	err := validateHandler(handler)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "file handler requires a file")
}

func TestValidateHandler_FileHandlerInvalidSubType(t *testing.T) {
	handler := &HandlerConfig{
		Type:    FileHandlerType,
		SubType: "invalid",
		Level:   InfoLevel,
		File:    "test.log",
	}
	err := validateHandler(handler)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid file handler subtype")
}

func TestGetEnabledHandlers(t *testing.T) {
	config := &Config{
		Multilog: LogConfig{
			Handlers: []HandlerConfig{
				{Type: ConsoleHandlerType, Level: InfoLevel, Enabled: true},
				{Type: FileHandlerType, Level: WarnLevel, Enabled: false},
			},
		},
	}
	enabled := config.GetEnabledHandlers()
	assert.Len(t, enabled, 1)
	assert.Equal(t, ConsoleHandlerType, enabled[0].Type)
}

func TestGetCustomHandlerOptionsForHandler(t *testing.T) {
	cfg := &Config{}
	handlerCfg := HandlerConfig{
		Type:                 FileHandlerType,
		SubType:              TextHandlerSubType,
		Level:                InfoLevel,
		Enabled:              true,
		UseSingleLetterLevel: true,
		Pattern:              "[time] [level]",
		PatternPlaceholders:  " [time], [level] ",
		File:                 "app.log",
	}

	opts, err := cfg.GetCustomHandlerOptionsForHandler(handlerCfg)
	assert.NoError(t, err)
	assert.Equal(t, InfoLevel, opts.Level)
	assert.True(t, opts.Enabled)
	assert.Equal(t, handlerCfg.Pattern, opts.Pattern)
	assert.Equal(t, []string{"[time]", "[level]"}, opts.PatternPlaceholders)
	assert.Equal(t, "app.log", opts.File)
	assert.True(t, opts.AddSource)
	assert.True(t, opts.UseSingleLetterLevel)
}

func TestGetCustomHandlerOptionsForHandler_UnknownType(t *testing.T) {
	cfg := &Config{}
	handlerCfg := HandlerConfig{
		Type:    "unknown",
		Level:   InfoLevel,
		Enabled: true,
	}
	_, err := cfg.GetCustomHandlerOptionsForHandler(handlerCfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown handlerConfig type")
}

func TestCreateHandlers(t *testing.T) {
	cfg := &Config{
		Multilog: LogConfig{
			Handlers: []HandlerConfig{
				{Type: ConsoleHandlerType, Level: InfoLevel, Enabled: true},
				{
					Type:    FileHandlerType,
					SubType: TextHandlerSubType,
					File:    "test.log",
					Level:   DebugLevel,
					Enabled: true,
				},
				{
					Type:    FileHandlerType,
					SubType: JSONHandlerSubType,
					File:    "test.json",
					Level:   DebugLevel,
					Enabled: true,
				},
			},
		},
	}
	handlers, err := CreateHandlers(cfg)
	assert.NoError(t, err)
	assert.Len(t, handlers, 3)
}

func TestCreateHandlers_UnknownType(t *testing.T) {
	cfg := &Config{
		Multilog: LogConfig{
			Handlers: []HandlerConfig{
				{Type: "invalid", Level: InfoLevel, Enabled: true},
			},
		},
	}
	_, err := CreateHandlers(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown handlerConfig type")
}

func TestCreateHandlers_InvalidHandlerType(t *testing.T) {
	cfg := &Config{
		Multilog: LogConfig{
			Handlers: []HandlerConfig{
				{
					Type:    "invalid",
					Level:   InfoLevel,
					Enabled: true,
				},
			},
		},
	}
	handlers, err := CreateHandlers(cfg)
	assert.Error(t, err)
	assert.Nil(t, handlers)
	assert.Contains(t, err.Error(), "unknown handlerConfig type")
}

func TestCreateHandlers_InvalidHandlerSubType(t *testing.T) {
	cfg := &Config{
		Multilog: LogConfig{
			Handlers: []HandlerConfig{
				{
					Type:    FileHandlerType,
					SubType: "invalid",
					Level:   InfoLevel,
					Enabled: true,
				},
			},
		},
	}
	handlers, err := CreateHandlers(cfg)
	assert.Error(t, err)
	assert.Nil(t, handlers)
	assert.Contains(t, err.Error(), "unknown handler subtype")
}

func TestCreateHandler_UnknownType(t *testing.T) {
	options := CustomHandlerOptions{}
	_, err := createHandler("unknown", options)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown handler type")
}

func TestCreateHandler_UnknownSubType(t *testing.T) {
	options := CustomHandlerOptions{
		SubType: "unknown",
	}
	_, err := createHandler("file", options)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown handler subtype")
}

func TestDefaultIfZero(t *testing.T) {
	assert.Equal(t, 10, defaultIfZero(0, 10))
	assert.Equal(t, 5, defaultIfZero(5, 10))
}

func TestRemovePlaceholderChars(t *testing.T) {
	values := map[string]interface{}{
		"[key1]": "value1",
		"key2":   "value2",
	}
	result := RemovePlaceholderChars(values)
	assert.Equal(t, "value1", result["key1"])
	assert.Equal(t, "value2", result["key2"])
	assert.NotContains(t, result, "[key1]")
}

func TestTrimSpaces(t *testing.T) {
	input := []string{" [time] ", "[level] "}
	output := TrimSpaces(input)
	assert.Equal(t, "[time]", output[0])
	assert.Equal(t, "[level]", output[1])
}

func TestValidateHandlers(t *testing.T) {
	handlers := []HandlerConfig{
		{Type: ConsoleHandlerType, Level: InfoLevel, Enabled: true},
		{Type: ConsoleHandlerType, Level: DebugLevel, Enabled: true},
	}
	err := validateHandlers(handlers)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "only one console handler")
}

func TestValidateHandlers_ValidMultiple(t *testing.T) {
	handlers := []HandlerConfig{
		{Type: ConsoleHandlerType, Level: InfoLevel, Enabled: true},
		{Type: FileHandlerType, Level: DebugLevel, File: "test.log", Enabled: true},
	}
	err := validateHandlers(handlers)
	assert.NoError(t, err)
}

func TestGetCustomHandlerOptionsForHandler_DefaultValues(t *testing.T) {
	cfg := &Config{}
	handler := HandlerConfig{
		Type:    FileHandlerType,
		Level:   InfoLevel,
		Enabled: true,
		File:    "test.log",
	}

	options, err := cfg.GetCustomHandlerOptionsForHandler(handler)
	assert.NoError(t, err)
	assert.Equal(t, DefaultLogFileSize, options.MaxSize)
	assert.Equal(t, DefaultLogFileBackups, options.MaxBackups)
	assert.Equal(t, DefaultLogFileAge, options.MaxAge)
	assert.True(t, options.AddSource)
	assert.False(t, options.UseSingleLetterLevel)
}

func TestValidateHandler_ConsoleHandlerWithSubType(t *testing.T) {
	handler := HandlerConfig{
		Type:    ConsoleHandlerType,
		SubType: JSONHandlerSubType,
		Level:   InfoLevel,
		Enabled: true,
	}
	err := validateHandler(&handler)
	assert.NoError(t, err) // Console handler ignores subtype
}

func TestValidateHandlers_MultipleFileHandlers(t *testing.T) {
	handlers := []HandlerConfig{
		{
			Type:    FileHandlerType,
			SubType: TextHandlerSubType,
			Level:   InfoLevel,
			File:    "info.log",
			Enabled: true,
		},
		{
			Type:    FileHandlerType,
			SubType: JSONHandlerSubType,
			Level:   ErrorLevel,
			File:    "error.log",
			Enabled: true,
		},
	}
	err := validateHandlers(handlers)
	assert.NoError(t, err)
}

func TestTrimSpaces_EmptySlice(t *testing.T) {
	result := TrimSpaces([]string{})
	assert.Empty(t, result)
}

func TestTrimSpaces_WithSpaces(t *testing.T) {
	input := []string{" [time] ", "  [level]  ", "[msg] "}
	expected := []string{"[time]", "[level]", "[msg]"}
	result := TrimSpaces(input)
	assert.Equal(t, expected, result)
}

func TestRemovePlaceholderChars_EmptyMap(t *testing.T) {
	result := RemovePlaceholderChars(make(map[string]interface{}))
	assert.Empty(t, result)
}

func TestRemovePlaceholderChars_MixedKeys(t *testing.T) {
	input := map[string]interface{}{
		"[key1]":   "value1",
		"key2":     "value2",
		"[key3]":   "value3",
		"[[key4]]": "value4",
	}
	result := RemovePlaceholderChars(input)
	assert.Equal(t, "value1", result["key1"])
	assert.Equal(t, "value2", result["key2"])
	assert.Equal(t, "value3", result["key3"])
	assert.Equal(t, "value4", result["[key4]"])
}

func TestValidateHandler_EmptyLevel(t *testing.T) {
	handler := HandlerConfig{
		Type:    FileHandlerType,
		SubType: TextHandlerSubType,
		File:    "test.log",
	}
	err := validateHandler(&handler)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid log level")
}

func TestNewConfig_EmptyFile(t *testing.T) {
	testFile := filepath.Join(t.TempDir(), "empty_config.yaml")
	err := os.WriteFile(testFile, []byte(""), 0o644)
	assert.NoError(t, err)

	_, err = NewConfig(testFile)
	assert.NoError(t, err)
}

func TestNewConfig_InvalidConfig(t *testing.T) {
	testFile := filepath.Join(t.TempDir(), "invalid_config.yaml")
	configData := []byte(`multilog:
  handlers:
    - type: unknown
      level: info
      enabled: true
`)
	err := os.WriteFile(testFile, configData, 0o644)
	assert.NoError(t, err)

	_, err = NewConfig(testFile)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid config")
}

func TestNewFileHandler_UnknownType(t *testing.T) {
	options := CustomHandlerOptions{
		SubType: "unknown",
	}
	_, err := newFileHandler(options)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown handler subtype")
}
