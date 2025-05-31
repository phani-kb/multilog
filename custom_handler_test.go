package multilog

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// testDefaultOptions returns default CustomHandlerOptions for testing
func testDefaultOptions() *CustomHandlerOptions {
	return &CustomHandlerOptions{
		Level:           "info",
		Enabled:         true,
		Pattern:         "[time] [level] [msg]",
		ValuePrefixChar: DefaultValuePrefixChar,
		ValueSuffixChar: DefaultValueSuffixChar,
	}
}

// TestNewCustomHandler tests the initialization of CustomHandler
func TestNewCustomHandler(t *testing.T) {
	// Test with nil options (should use defaults)
	var buf bytes.Buffer
	writer := bufio.NewWriter(&buf)
	handler := NewCustomHandler(nil, writer, nil)

	if handler.Opts.Level != DefaultLogLevel {
		t.Errorf("Expected default level %s, got %s", DefaultLogLevel, handler.Opts.Level)
	}
	if !handler.Opts.Enabled {
		t.Error("Expected Enabled to be true")
	}
	if handler.Opts.Pattern != DefaultFormat {
		t.Errorf("Expected default pattern %s, got %s", DefaultFormat, handler.Opts.Pattern)
	}
	if handler.Opts.AddSource {
		t.Error("Expected AddSource to be false")
	}

	// Test with custom options but nil replaceAttr (should use default replace attr)
	customOpts := &CustomHandlerOptions{
		Level:                "DEBUG",
		SubType:              "custom",
		Enabled:              true,
		Pattern:              "[date] [level] [msg]",
		AddSource:            true,
		UseSingleLetterLevel: true,
	}

	handler = NewCustomHandler(customOpts, writer, nil)
	if handler.Opts != customOpts {
		t.Error("Expected handler options to match provided options")
	}

	// Verify the handler was set up correctly
	if handler.writer != writer {
		t.Error("Expected writer to be set correctly")
	}
	if handler.sb == nil {
		t.Error("Expected string builder to be initialized")
	}
	if handler.handler == nil {
		t.Error("Expected slog handler to be initialized")
	}
}

func TestCustomHandler_Enabled(t *testing.T) {
	opts := &CustomHandlerOptions{Enabled: true}
	handler := NewCustomHandler(opts, bufio.NewWriter(&strings.Builder{}), nil)
	ctx := context.Background()

	assert.True(t, handler.Enabled(ctx, slog.LevelInfo))
}

func TestCustomHandler_Handle(t *testing.T) {
	opts := &CustomHandlerOptions{Enabled: true}
	writer := bufio.NewWriter(&strings.Builder{})
	handler := NewCustomHandler(opts, writer, nil)
	ctx := context.Background()
	record := slog.Record{
		Time:    time.Now(),
		Level:   slog.LevelInfo,
		Message: "test message",
	}

	err := handler.Handle(ctx, record)
	assert.NoError(t, err)
}

func TestCustomHandler_WithAttrs(t *testing.T) {
	opts := &CustomHandlerOptions{}
	handler := NewCustomHandler(opts, bufio.NewWriter(&strings.Builder{}), nil)
	attrs := []slog.Attr{
		{Key: "key1", Value: slog.StringValue("value1")},
	}

	newHandler := handler.WithAttrs(attrs)
	assert.NotNil(t, newHandler)
}

func TestCustomHandler_WithGroup(t *testing.T) {
	opts := &CustomHandlerOptions{}
	handler := NewCustomHandler(opts, bufio.NewWriter(&strings.Builder{}), nil)

	newHandler := handler.WithGroup("group1")
	assert.NotNil(t, newHandler)
}

func TestCustomHandler_GetOptions(t *testing.T) {
	opts := &CustomHandlerOptions{}
	handler := NewCustomHandler(opts, bufio.NewWriter(&strings.Builder{}), nil)

	assert.Equal(t, opts, handler.GetOptions())
}

func TestCustomHandler_GetMutex(t *testing.T) {
	opts := &CustomHandlerOptions{}
	handler := NewCustomHandler(opts, bufio.NewWriter(&strings.Builder{}), nil)

	assert.NotNil(t, &handler.mu)
}

func TestCustomHandler_GetStringBuilder(t *testing.T) {
	opts := &CustomHandlerOptions{}
	handler := NewCustomHandler(opts, bufio.NewWriter(&strings.Builder{}), nil)

	assert.NotNil(t, handler.GetStringBuilder())
}

func TestCustomHandler_GetKeyValue(t *testing.T) {
	opts := &CustomHandlerOptions{}
	handler := NewCustomHandler(opts, bufio.NewWriter(&strings.Builder{}), nil)
	sb := &strings.Builder{}
	sb.WriteString("key1=value1 key2=value2")

	value := handler.GetKeyValue("key1", sb, true)
	assert.Equal(t, "value1", value)
	assert.Equal(t, "key2=value2", sb.String())
}

func TestCustomHandler_GetWriter(t *testing.T) {
	opts := &CustomHandlerOptions{}
	writer := bufio.NewWriter(&strings.Builder{})
	handler := NewCustomHandler(opts, writer, nil)

	assert.Equal(t, writer, handler.GetWriter())
}

func TestCustomHandler_GetSlogHandler(t *testing.T) {
	opts := &CustomHandlerOptions{}
	handler := NewCustomHandler(opts, bufio.NewWriter(&strings.Builder{}), nil)

	assert.NotNil(t, handler.GetSlogHandler())
}

func TestGetSourceValue(t *testing.T) {
	sb := &strings.Builder{}
	sb.WriteString("source=main.go:42")

	getKeyValue := func(key string, sb *strings.Builder, removeKey bool) string {
		parts := strings.Fields(sb.String())
		for i, part := range parts {
			if strings.HasPrefix(part, key+"=") {
				value := strings.TrimPrefix(part, key+"=")
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

	// Test for LevelPerf
	result := GetSourceValue(LevelPerf, sb, getKeyValue)
	assert.Equal(t, "unknown", result)

	// Test for other levels with an empty result
	sb.Reset()
	sb.WriteString("source=main.go:42")
	result = GetSourceValue(slog.LevelInfo, sb, getKeyValue)
	assert.Equal(t, "main.go:42", result)

	// Test for other levels with a non-empty result
	sb.Reset()
	sb.WriteString("source=main.go:42")
	result = GetSourceValue(slog.LevelInfo, sb, getKeyValue)
	assert.Equal(t, "main.go:42", result)
}

// TestGetSourceValueAllBranches tests different branches of the GetSourceValue function
func TestGetSourceValueAllBranches(t *testing.T) {
	// Mock string builder with content
	sb := &strings.Builder{}
	sb.WriteString(
		"time=2023-01-01T12:00:00Z level=INFO msg=\"Test message\" source=test_file.go:42:main.testFunc",
	)

	// Mock getKeyValue function for a regular source
	getKeyValue := func(key string, _ *strings.Builder, _ bool) string {
		if key == slog.SourceKey {
			return "test_file.go:42:main.testFunc"
		}
		return ""
	}

	// Test a normal path (source exists and is valid)
	source := GetSourceValue(slog.LevelInfo, sb, getKeyValue)
	if source != "test_file.go:42:main.testFunc" {
		t.Errorf("Expected source 'test_file.go:42:main.testFunc', got '%s'", source)
	}

	// Mock getKeyValue that returns an empty source
	emptySourceGetKeyValue := func(key string, _ *strings.Builder, _ bool) string {
		if key == slog.SourceKey {
			return ""
		}
		return ""
	}

	// Test an empty source path
	source = GetSourceValue(slog.LevelInfo, sb, emptySourceGetKeyValue)
	// We can't predict the exact output since it depends on the runtime call stack
	if source == "" {
		t.Error("Expected non-empty source value")
	}

	// Test with GenericLogFuncName suffix
	genericFuncGetKeyValue := func(key string, _ *strings.Builder, _ bool) string {
		if key == slog.SourceKey {
			return "file.go:10:" + GenericLogFuncName
		}
		return ""
	}
	source = GetSourceValue(slog.LevelInfo, sb, genericFuncGetKeyValue)
	if source == "" {
		t.Error("Expected non-empty source value")
	}

	source = GetSourceValue(LevelPerf, sb, getKeyValue)
	if source == "" {
		t.Error("Expected non-empty source value for performance level")
	}
}

// TestGetPlaceholderValues tests the placeholder mapping in GetPlaceholderValues
func TestGetPlaceholderValues(t *testing.T) {
	// Set up test data
	sb := &strings.Builder{}
	sb.WriteString("level=INFO source=test_file.go:42:main.testFunc")

	testTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	record := slog.Record{
		Time:    testTime,
		Message: "Test message",
		Level:   slog.LevelInfo,
	}

	// Test all placeholders
	placeholders := []string{
		DatePlaceholder,
		TimePlaceholder,
		DateTimePlaceholder,
		LevelPlaceholder,
		MsgPlaceholder,
		PerfPlaceholder,
		SourcePlaceholder,
	}

	// Mock getKeyValue
	getKeyValue := func(key string, _ *strings.Builder, _ bool) string {
		if key == slog.LevelKey {
			return "INFO"
		}
		if key == slog.SourceKey {
			return "test_file.go:42:main.testFunc"
		}
		return ""
	}

	// Test GetPlaceholderValues
	values := GetPlaceholderValues(sb, record, placeholders, getKeyValue)

	// Verify each placeholder has the correct value
	if values[DatePlaceholder] != testTime.Format(DefaultDateFormat) {
		t.Errorf(
			"Expected date value %s, got %s",
			testTime.Format(DefaultDateFormat),
			values[DatePlaceholder],
		)
	}

	if values[TimePlaceholder] != testTime.Format(DefaultTimeFormat) {
		t.Errorf(
			"Expected time value %s, got %s",
			testTime.Format(DefaultTimeFormat),
			values[TimePlaceholder],
		)
	}

	if values[DateTimePlaceholder] != testTime.Format(DefaultDateTimeFormat) {
		t.Errorf(
			"Expected datetime value %s, got %s",
			testTime.Format(DefaultDateTimeFormat),
			values[DateTimePlaceholder],
		)
	}

	if values[LevelPlaceholder] != "INFO" {
		t.Errorf("Expected level value INFO, got %s", values[LevelPlaceholder])
	}

	if values[MsgPlaceholder] != "Test message" {
		t.Errorf("Expected msg value 'Test message', got %s", values[MsgPlaceholder])
	}

	perfValue, ok := values[PerfPlaceholder].(string)
	if !ok {
		t.Errorf("perfValue is not a string, got: %T", values[PerfPlaceholder])
	} else {
		if !strings.HasPrefix(perfValue, "goroutines:") || !strings.Contains(perfValue, "alloc:") {
			t.Errorf("Performance metrics format incorrect. Got: %s", perfValue)
		}
	}

	if values[SourcePlaceholder] != "test_file.go:42:main.testFunc" {
		t.Errorf(
			"Expected source value 'test_file.go:42:main.testFunc', got %s",
			values[SourcePlaceholder],
		)
	}
}

// TestGetPlaceholderValuesWithNumericType tests that GetPlaceholderValues handles numeric values that may be passed in log messages
func TestGetPlaceholderValuesWithNumericType(t *testing.T) {
	sb := &strings.Builder{}
	sb.WriteString("level=INFO temperature=80")

	testTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	record := slog.Record{
		Time:    testTime,
		Message: "Temperature warning",
		Level:   slog.LevelWarn,
	}

	// Add a placeholder for our custom numeric attribute
	customPlaceholder := "[temperature]"
	placeholders := []string{
		LevelPlaceholder,
		MsgPlaceholder,
		customPlaceholder,
	}

	getKeyValue := func(key string, _ *strings.Builder, _ bool) string {
		if key == "level" {
			return "WARN"
		} else if key == customPlaceholder {
			return "80"
		}
		return ""
	}

	values := GetPlaceholderValues(sb, record, placeholders, getKeyValue)

	temperatureValue := values[customPlaceholder]
	assert.Equal(t, "80", temperatureValue)

	temperature := temperatureValue.(string)
	formattedMsg := fmt.Sprintf("Temperature is %s degrees", temperature)
	assert.Equal(t, "Temperature is 80 degrees", formattedMsg)
}

// TestGetKeyValue tests the parsing of key-value pairs from logs
func TestGetKeyValue(t *testing.T) {
	handler := &CustomHandler{
		Opts: &CustomHandlerOptions{},
		sb:   &strings.Builder{},
	}

	// Test simple key-value extraction
	sb := &strings.Builder{}
	sb.WriteString("time=2023-01-01T12:00:00Z level=INFO msg=\"Test message\"")
	value := handler.GetKeyValue("level", sb, false)
	if value != "INFO" {
		t.Errorf("Expected 'INFO', got '%s'", value)
	}

	// Test quoted value extraction
	sb.Reset()
	sb.WriteString(
		"time=2023-01-01T12:00:00Z level=INFO msg=\"This is a test message with spaces\"",
	)
	value = handler.GetKeyValue("msg", sb, false)
	if value != "This is a test message with spaces" {
		t.Errorf("Expected 'This is a test message with spaces', got '%s'", value)
	}

	// Test removal of a key
	sb.Reset()
	sb.WriteString("time=2023-01-01T12:00:00Z level=INFO msg=\"Test message\"")
	value = handler.GetKeyValue("level", sb, true)
	if value != "INFO" {
		t.Errorf("Expected 'INFO', got '%s'", value)
	}
	if sb.String() != "time=2023-01-01T12:00:00Z msg=\"Test message\"" {
		t.Errorf("Expected key to be removed, got '%s'", sb.String())
	}

	// Test key not found
	sb.Reset()
	sb.WriteString("time=2023-01-01T12:00:00Z level=INFO")
	value = handler.GetKeyValue("notfound", sb, false)
	if value != "" {
		t.Errorf("Expected empty string, got '%s'", value)
	}
}

// TestBuildOutput tests the buildOutput function which formats log messages
func TestBuildOutput(t *testing.T) {
	tests := []struct {
		name      string
		pattern   string
		values    map[string]string
		sbContent string
		level     slog.Level
		expected  string
		checkPerf bool // Whether to check if performance metrics were added
	}{
		{
			name:    "Basic pattern replacement",
			pattern: "[level] [msg]",
			values: map[string]string{
				LevelPlaceholder: "INFO",
				MsgPlaceholder:   "Test message",
			},
			sbContent: "",
			level:     slog.LevelInfo,
			expected:  DefaultValuePrefixChar + "INFO" + DefaultValueSuffixChar + " " + DefaultValuePrefixChar + "Test message" + DefaultValueSuffixChar,
		},
		{
			name:    "With suffix content",
			pattern: "[level] [msg]",
			values: map[string]string{
				LevelPlaceholder: "INFO",
				MsgPlaceholder:   "Test message",
			},
			sbContent: "key1=value1 key2=value2",
			level:     slog.LevelInfo,
			expected: DefaultValuePrefixChar + "INFO" + DefaultValueSuffixChar + " " + DefaultValuePrefixChar + "Test message" + DefaultValueSuffixChar +
				" " + DefaultSuffixStartChar + "key1=value1 key2=value2" + DefaultSuffixEndChar,
		},
		{
			name:    "Performance level with perf placeholder",
			pattern: "[level] [msg] [perf]",
			values: map[string]string{
				LevelPlaceholder: "PERF",
				MsgPlaceholder:   "Performance test",
				PerfPlaceholder:  "100ms",
			},
			sbContent: "",
			level:     LevelPerf,
			expected: DefaultValuePrefixChar + "PERF" + DefaultValueSuffixChar + " " +
				DefaultValuePrefixChar + "Performance test" + DefaultValueSuffixChar + " " +
				DefaultValuePrefixChar + "100ms" + DefaultValueSuffixChar,
			checkPerf: false, // Perf already in a pattern shouldn't add again
		},
		{
			name:    "Performance level without perf placeholder",
			pattern: "[level] [msg]",
			values: map[string]string{
				LevelPlaceholder: "PERF",
				MsgPlaceholder:   "Performance test",
			},
			sbContent: "",
			level:     LevelPerf,
			expected:  "",
			checkPerf: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sb := &strings.Builder{}
			sb.WriteString(tt.sbContent)

			result := buildOutput(tt.pattern, tt.values, sb, tt.level, testDefaultOptions())

			if !tt.checkPerf {
				// Check exact match for non-perf tests
				if result != tt.expected {
					t.Errorf("buildOutput() = %v, want %v", result, tt.expected)
				}
			} else {
				if !strings.Contains(result, DefaultPerfStartChar+"goroutines:") {
					t.Errorf("buildOutput() should contain performance metrics starting with 'goroutines:', got: %v", result)
				}

				// Check the beginning of string matches what we expect
				expectedStart := DefaultValuePrefixChar + "PERF" + DefaultValueSuffixChar + " " +
					DefaultValuePrefixChar + "Performance test" + DefaultValueSuffixChar + " " +
					DefaultPerfStartChar
				if !strings.HasPrefix(result, expectedStart) {
					t.Errorf("buildOutput() should start with %v, got: %v", expectedStart, result)
				}
			}
		})
	}
}

// TestBuildOutputPerfMetricsAddition tests add performance metrics when the level is LevelPerf and perf placeholder isn't in pattern
func TestBuildOutputPerfMetricsAddition(t *testing.T) {
	pattern := "[level] [msg]" // No perf placeholder
	values := map[string]string{
		LevelPlaceholder: "PERF",
		MsgPlaceholder:   "Test performance",
	}
	sb := &strings.Builder{}

	result := buildOutput(pattern, values, sb, LevelPerf, testDefaultOptions())

	if !strings.Contains(result, DefaultPerfStartChar+"goroutines:") {
		t.Errorf(
			"Performance metrics not added to output. Result should contain metrics starting with 'goroutines:', got: %s",
			result,
		)
	}

	// Now test that metrics are NOT added when the pattern already contains perf placeholder
	patternWithPerf := "[level] [msg] [perf]"
	valuesWithPerf := map[string]string{
		LevelPlaceholder: "PERF",
		MsgPlaceholder:   "Test performance",
		PerfPlaceholder:  GetPerformanceMetrics(),
	}

	result = buildOutput(patternWithPerf, valuesWithPerf, sb, LevelPerf, testDefaultOptions())

	// Count occurrences of performance metrics
	count := strings.Count(result, GetPerformanceMetrics())
	if count > 1 {
		t.Errorf("Performance metrics added twice. Expected only once in output: %s", result)
	}
}

// TestBuildOutputWithEmptyValues tests handling of empty placeholder values
func TestBuildOutputWithEmptyValues(t *testing.T) {
	pattern := "[level] [msg] [source]"
	values := map[string]string{
		LevelPlaceholder:  "INFO",
		MsgPlaceholder:    "Test message",
		SourcePlaceholder: "", // Empty value should be skipped
	}
	sb := &strings.Builder{}

	result := buildOutput(pattern, values, sb, slog.LevelInfo, testDefaultOptions())

	// Source placeholder should remain in output since the value is empty
	expected := DefaultValuePrefixChar + "INFO" + DefaultValueSuffixChar + " " +
		DefaultValuePrefixChar + "Test message" + DefaultValueSuffixChar + " [source]"

	if result != expected {
		t.Errorf("buildOutput() = %v, want %v", result, expected)
	}
}

// TestBuildOutputWithSuffixOnly tests when there's only suffix content
func TestBuildOutputWithSuffixOnly(t *testing.T) {
	pattern := ""                 // Empty pattern
	values := map[string]string{} // No values
	sb := &strings.Builder{}
	sb.WriteString("key1=value1 key2=value2")

	result := buildOutput(pattern, values, sb, slog.LevelInfo, testDefaultOptions())

	expected := " " + DefaultSuffixStartChar + "key1=value1 key2=value2" + DefaultSuffixEndChar

	if result != expected {
		t.Errorf("buildOutput() = %v, want %v", result, expected)
	}
}

func TestGenerateDefaultCustomReplaceAttr_SourceHandling(t *testing.T) {
	testCases := []struct {
		name          string
		opts          CustomHandlerOptions
		sourceValue   interface{}
		expected      string
		keysToRemove  []string
		shouldBeEmpty bool
	}{
		{
			name: "source handling with file and function",
			opts: CustomHandlerOptions{
				AddSource: true,
			},
			sourceValue: &slog.Source{
				Function: "github.com/example/pkg/multilog.TestFunc",
				File:     "/home/user/go/src/github.com/example/pkg/file.go",
				Line:     123,
			},
			expected: "file.go:123:multilog.TestFunc",
		},
		{
			name: "source as string",
			opts: CustomHandlerOptions{
				AddSource: true,
			},
			sourceValue: "custom-source-string",
			expected:    "custom-source-string",
		},
		{
			name: "source handling disabled",
			opts: CustomHandlerOptions{
				AddSource: false,
			},
			sourceValue: &slog.Source{
				Function: "github.com/example/pkg/multilog.TestFunc",
				File:     "/home/user/go/src/github.com/example/pkg/file.go",
				Line:     123,
			},
			// Expect no transformation when AddSource is false
			expected: "",
		},
		{
			name: "source key in removal list",
			opts: CustomHandlerOptions{
				AddSource: true,
			},
			sourceValue:   &slog.Source{},
			keysToRemove:  []string{slog.SourceKey},
			shouldBeEmpty: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create the replace attribute function
			replaceAttr := GenerateDefaultCustomReplaceAttr(tc.opts, tc.keysToRemove...)

			// Create a source attribute
			attr := slog.Attr{
				Key:   slog.SourceKey,
				Value: slog.AnyValue(tc.sourceValue),
			}

			// Apply the replacement function
			result := replaceAttr(nil, attr)

			if tc.shouldBeEmpty {
				if result.Key != "" {
					t.Errorf("Expected empty attribute, got key=%s", result.Key)
				}
				return
			}

			if !tc.opts.AddSource {
				// When AddSource is false, the attribute should be unchanged
				if result.Value.String() != attr.Value.String() {
					t.Errorf("Expected unchanged value when AddSource=false, got %v", result.Value)
				}
				return
			}

			// When AddSource is true, verify the value was transformed correctly
			if result.Value.String() != tc.expected && tc.expected != "" {
				t.Errorf(
					"Expected transformed source value '%s', got '%s'",
					tc.expected,
					result.Value,
				)
			}
		})
	}
}

func TestGetPatternForLevel(t *testing.T) {
	tests := []struct {
		name     string
		level    slog.Level
		pattern  string
		expected string
	}{
		{
			name:     "Empty pattern with debug level",
			level:    slog.LevelDebug,
			pattern:  "",
			expected: DefaultDebugFormat,
		},
		{
			name:     "Empty pattern with error level",
			level:    slog.LevelError,
			pattern:  "",
			expected: DefaultErrorFormat,
		},
		{
			name:     "Empty pattern with info level",
			level:    slog.LevelInfo,
			pattern:  "",
			expected: DefaultFormat, // Should use default for non-specific levels
		},
		{
			name:     "Empty pattern with perf level",
			level:    LevelPerf,
			pattern:  "",
			expected: DefaultPerfFormat,
		},
		{
			name:     "Custom pattern with debug level",
			level:    slog.LevelDebug,
			pattern:  "Custom pattern",
			expected: "Custom pattern",
		},
		{
			name:     "Custom pattern with error level",
			level:    slog.LevelError,
			pattern:  "Custom pattern",
			expected: "Custom pattern",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getPatternForLevel(tt.level, tt.pattern)
			if got != tt.expected {
				t.Errorf("getPatternForLevel() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGenerateDefaultCustomReplaceAttr(t *testing.T) {
	tests := []struct {
		name          string
		opts          CustomHandlerOptions
		keysToRemove  []string
		inputAttr     slog.Attr
		groups        []string
		expectedKey   string
		expectedValue string
		expectRemoved bool
	}{
		{
			name: "Level key with SingleLetterLevel true",
			opts: CustomHandlerOptions{
				UseSingleLetterLevel: true,
			},
			keysToRemove:  []string{},
			inputAttr:     slog.String(slog.LevelKey, ""),
			groups:        []string{},
			expectedKey:   slog.LevelKey,
			expectedValue: "D",
			expectRemoved: false,
		},
		{
			name: "Level key with SingleLetterLevel false",
			opts: CustomHandlerOptions{
				UseSingleLetterLevel: false,
			},
			keysToRemove:  []string{},
			inputAttr:     slog.String(slog.LevelKey, ""),
			groups:        []string{},
			expectedKey:   slog.LevelKey,
			expectedValue: "DEBUG",
			expectRemoved: false,
		},
		{
			name: "Source key with AddSource true",
			opts: CustomHandlerOptions{
				AddSource: true,
			},
			keysToRemove: []string{},
			inputAttr: slog.Any(
				slog.SourceKey,
				&slog.Source{File: "/path/to/file.go", Line: 42, Function: "package/name/function"},
			),
			groups:        []string{},
			expectedKey:   slog.SourceKey,
			expectedValue: "file.go:42:function",
			expectRemoved: false,
		},
		{
			name:          "Key in keysToRemove should be removed",
			opts:          CustomHandlerOptions{},
			keysToRemove:  []string{"remove_me"},
			inputAttr:     slog.String("remove_me", "value"),
			groups:        []string{},
			expectedKey:   "",
			expectedValue: "",
			expectRemoved: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			replaceAttr := GenerateDefaultCustomReplaceAttr(tt.opts, tt.keysToRemove...)

			var attr slog.Attr
			if tt.inputAttr.Key == slog.LevelKey {
				// For testing level attributes, we need to create a real level value
				attr = slog.Any(slog.LevelKey, slog.LevelDebug)
			} else {
				attr = tt.inputAttr
			}

			result := replaceAttr(tt.groups, attr)

			if tt.expectRemoved {
				if result.Key != "" {
					t.Errorf("expected attribute to be removed, but got key=%s", result.Key)
				}
			} else {
				if result.Key != tt.expectedKey {
					t.Errorf("expected key %s, got %s", tt.expectedKey, result.Key)
				}

				if value := result.Value.String(); value != tt.expectedValue {
					if tt.inputAttr.Key == slog.LevelKey || tt.inputAttr.Key == slog.SourceKey {
						t.Errorf("expected value %s, got %s", tt.expectedValue, value)
					}
				}
			}
		})
	}
}

func TestBuildOutputWithCustomPrefixSuffix(t *testing.T) {
	customOpts := &CustomHandlerOptions{
		Level:           "info",
		Enabled:         true,
		Pattern:         "[time] [level] [msg]",
		ValuePrefixChar: "<",
		ValueSuffixChar: ">",
	}

	pattern := "[level] [msg]"
	values := map[string]string{
		LevelPlaceholder: "INFO",
		MsgPlaceholder:   "Test message",
	}
	sb := &strings.Builder{}

	result := buildOutput(pattern, values, sb, slog.LevelInfo, customOpts)

	// Should use custom prefix/suffix instead of defaults
	expected := "<INFO> <Test message>"

	if result != expected {
		t.Errorf("buildOutput() with custom prefix/suffix = %v, want %v", result, expected)
	}
}
