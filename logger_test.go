package multilog

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
)

// Define proper context key types for tests
type contextKey string

func TestGetSlogLevel(t *testing.T) {
	tests := []struct {
		name     string
		level    string
		expected slog.Level
	}{
		{"debug", "debug", slog.LevelDebug},
		{"info", "info", slog.LevelInfo},
		{"warn", "warn", slog.LevelWarn},
		{"error", "error", slog.LevelError},
		{"perf", "perf", LevelPerf},
		{"unknown", "unknown", DefaultSLogLevel},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetSlogLevel(tt.level); got != tt.expected {
				t.Errorf("GetSlogLevel(%q) = %v, want %v", tt.level, got, tt.expected)
			}
		})
	}
}

func TestGetLevelName(t *testing.T) {
	tests := []struct {
		name     string
		level    slog.Level
		expected string
	}{
		{"debug", slog.LevelDebug, "debug"},
		{"info", slog.LevelInfo, "info"},
		{"warn", slog.LevelWarn, "warn"},
		{"error", slog.LevelError, "error"},
		{"perf", LevelPerf, "perf"},
		{"unknown", slog.Level(999), UnknownLevel},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetLevelName(tt.level); got != tt.expected {
				t.Errorf("GetLevelName(%v) = %v, want %v", tt.level, got, tt.expected)
			}
		})
	}
}

func TestRemove(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		attr     slog.Attr
		expected slog.Attr
	}{
		{"remove existing", "test", slog.String("test", "value"), slog.Attr{}},
		{"keep different", "test", slog.String("other", "value"), slog.String("other", "value")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Remove(tt.key, tt.attr)
			if got.Key != tt.expected.Key {
				t.Errorf(
					"Remove(%v, %v) key = %v, want %v",
					tt.key,
					tt.attr,
					got.Key,
					tt.expected.Key,
				)
			}
		})
	}
}

func TestRemoveKey(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		attr     slog.Attr
		expected slog.Attr
	}{
		{"remove existing", "test", slog.String("test", "value"), slog.Attr{}},
		{"keep different", "test", slog.String("other", "value"), slog.String("other", "value")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			removeFunc := RemoveKey(tt.key)
			got := removeFunc(nil, tt.attr)
			if got.Key != tt.expected.Key {
				t.Errorf(
					"RemoveKey(%v)(%v) key = %v, want %v",
					tt.key,
					tt.attr,
					got.Key,
					tt.expected.Key,
				)
			}
		})
	}
}

func TestPreDefinedRemoveFunctions(t *testing.T) {
	testAttr := slog.String(slog.TimeKey, "value")
	if got := RemoveTimeKey(nil, testAttr); got.Key != "" {
		t.Errorf("RemoveTimeKey should remove TimeKey, got %v", got.Key)
	}

	testAttr = slog.String(slog.LevelKey, "value")
	if got := RemoveLevelKey(nil, testAttr); got.Key != "" {
		t.Errorf("RemoveLevelKey should remove LevelKey, got %v", got.Key)
	}

	testAttr = slog.String(slog.SourceKey, "value")
	if got := RemoveSourceKey(nil, testAttr); got.Key != "" {
		t.Errorf("RemoveSourceKey should remove SourceKey, got %v", got.Key)
	}

	testAttr = slog.String(slog.MessageKey, "value")
	if got := RemoveMessageKey(nil, testAttr); got.Key != "" {
		t.Errorf("RemoveMessageKey should remove MessageKey, got %v", got.Key)
	}
}

func TestContainsKey(t *testing.T) {
	tests := []struct {
		name     string
		keys     []string
		key      string
		expected bool
	}{
		{"contains", []string{"a", "b", "c"}, "b", true},
		{"not contains", []string{"a", "b", "c"}, "d", false},
		{"empty keys", []string{}, "a", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ContainsKey(tt.keys, tt.key); got != tt.expected {
				t.Errorf("ContainsKey(%v, %v) = %v, want %v", tt.keys, tt.key, got, tt.expected)
			}
		})
	}
}

type MockHandler struct {
	buffer bytes.Buffer
	level  slog.Level
	called bool
}

func (h *MockHandler) Enabled(_ context.Context, l slog.Level) bool {
	return l >= h.level
}

func (h *MockHandler) Handle(_ context.Context, r slog.Record) error {
	h.called = true
	h.buffer.WriteString(r.Message)
	return nil
}

func (h *MockHandler) WithAttrs(_ []slog.Attr) slog.Handler {
	return h
}

func (h *MockHandler) WithGroup(_ string) slog.Handler {
	return h
}

func TestLoggerInterface(t *testing.T) {
	mockHandler := &MockHandler{level: slog.LevelDebug}

	logger := &Logger{Logger: slog.New(mockHandler)}

	tests := []struct {
		name    string
		logFunc func()
		message string
	}{
		{"Debugf", func() { logger.Debugf("debug %s", "message") }, "debug message"},
		{"Infof", func() { logger.Infof("info %s", "message") }, "info message"},
		{"Warnf", func() { logger.Warnf("warn %s", "message") }, "warn message"},
		{"Errorf", func() { logger.Errorf("error %s", "message") }, "error message"},
		{"Perff", func() { logger.Perff("perf %s", "message") }, "perf message"},
		{"Perf", func() { logger.Perf("static perf message") }, "static perf message"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockHandler.buffer.Reset()
			mockHandler.called = false

			tt.logFunc()

			if !mockHandler.called {
				t.Errorf("%s: handler was not called", tt.name)
			}

			if got := mockHandler.buffer.String(); !strings.Contains(got, tt.message) {
				t.Errorf("%s: got message %q, want %q", tt.name, got, tt.message)
			}
		})
	}
}

func TestNewLogger(t *testing.T) {
	// This is a simplified test since we can't easily mock the handler types
	// It just verifies that a logger can be created with no handlers
	logger := NewLogger()
	if logger == nil {
		t.Error("NewLogger() returned nil")
	}

	if logger == nil || logger.Logger == nil {
		t.Error("Logger.Logger is nil")
	}
}

func TestWithContext(t *testing.T) {
	var buf strings.Builder
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	logger := NewLogger(handler)

	ctx := context.WithValue(context.Background(), contextKey("test_key"), "test_value")

	ctxLogger := logger.WithContext(ctx)
	if _, ok := ctxLogger.(*ContextLogger); !ok {
		t.Fatal("WithContext should return a *ContextLogger")
	}
}

func TestContextLoggerMethods(t *testing.T) {
	var buf strings.Builder
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	logger := NewLogger(handler)

	type contextKey string
	ctx := context.WithValue(context.Background(), contextKey("test_key"), "test_value")
	ctxLogger := logger.WithContext(ctx)

	tests := []struct {
		name      string
		logFunc   func()
		checkFunc func() bool
	}{
		{
			name: "Debug method",
			logFunc: func() {
				ctxLogger.Debug("debug message", "key", "value")
			},
			checkFunc: func() bool {
				return strings.Contains(buf.String(), "debug message")
			},
		},
		{
			name: "Info method",
			logFunc: func() {
				ctxLogger.Info("info message", "key", "value")
			},
			checkFunc: func() bool {
				return strings.Contains(buf.String(), "info message")
			},
		},
		{
			name: "Warn method",
			logFunc: func() {
				ctxLogger.Warn("warn message", "key", "value")
			},
			checkFunc: func() bool {
				return strings.Contains(buf.String(), "warn message")
			},
		},
		{
			name: "Error method",
			logFunc: func() {
				ctxLogger.Error("error message", "key", "value")
			},
			checkFunc: func() bool {
				return strings.Contains(buf.String(), "error message")
			},
		},
		{
			name: "Debugf method",
			logFunc: func() {
				ctxLogger.Debugf("debugf %s", "formatted")
			},
			checkFunc: func() bool {
				return strings.Contains(buf.String(), "debugf formatted")
			},
		},
		{
			name: "Infof method",
			logFunc: func() {
				ctxLogger.Infof("infof %s", "formatted")
			},
			checkFunc: func() bool {
				return strings.Contains(buf.String(), "infof formatted")
			},
		},
		{
			name: "Warnf method",
			logFunc: func() {
				ctxLogger.Warnf("warnf %s", "formatted")
			},
			checkFunc: func() bool {
				return strings.Contains(buf.String(), "warnf formatted")
			},
		},
		{
			name: "Errorf method",
			logFunc: func() {
				ctxLogger.Errorf("errorf %s", "formatted")
			},
			checkFunc: func() bool {
				return strings.Contains(buf.String(), "errorf formatted")
			},
		},
		{
			name: "Perf method",
			logFunc: func() {
				ctxLogger.Perf("perf message", "key", "value")
			},
			checkFunc: func() bool {
				return strings.Contains(buf.String(), "perf message")
			},
		},
		{
			name: "Perff method",
			logFunc: func() {
				ctxLogger.Perff("perff %s", "formatted")
			},
			checkFunc: func() bool {
				return strings.Contains(buf.String(), "perff formatted")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			tt.logFunc()
			if !tt.checkFunc() {
				t.Errorf("Expected log message not found in output: %s", buf.String())
			}
		})
	}
}

func TestContextLoggerWithContext(t *testing.T) {
	var buf strings.Builder
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	logger := NewLogger(handler)

	ctx1 := context.WithValue(context.Background(), contextKey("key1"), "value1")
	ctxLogger1 := logger.WithContext(ctx1)

	ctx2 := context.WithValue(context.Background(), contextKey("key2"), "value2")
	ctxLogger2 := ctxLogger1.WithContext(ctx2)

	if _, ok := ctxLogger2.(*ContextLogger); !ok {
		t.Fatal("WithContext on ContextLogger should return another *ContextLogger")
	}

	ctxLogger2.Info("test message")

	if !strings.Contains(buf.String(), "test message") {
		t.Errorf("Expected log message not found in output: %s", buf.String())
	}
}

func TestContextLoggerWithField(t *testing.T) {
	var buf strings.Builder
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	logger := NewLogger(handler)

	ctx := context.WithValue(context.Background(), contextKey("ctx_key"), "ctx_value")
	ctxLogger := logger.WithContext(ctx)

	fieldLogger := ctxLogger.WithField("test_field", "test_value")

	if _, ok := fieldLogger.(*ContextLogger); !ok {
		t.Fatal("WithField on ContextLogger should return another *ContextLogger")
	}

	fieldLogger.Info("field test")

	if !strings.Contains(buf.String(), "test_field") ||
		!strings.Contains(buf.String(), "test_value") {
		t.Errorf("Expected field not found in output: %s", buf.String())
	}
}

func TestContextLoggerWithFields(t *testing.T) {
	var buf strings.Builder
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	logger := NewLogger(handler)

	ctx := context.WithValue(context.Background(), contextKey("ctx_key"), "ctx_value")
	ctxLogger := logger.WithContext(ctx)

	fields := map[string]any{
		"field1": "value1",
		"field2": 42,
	}
	fieldsLogger := ctxLogger.WithFields(fields)

	if _, ok := fieldsLogger.(*ContextLogger); !ok {
		t.Fatal("WithFields on ContextLogger should return another *ContextLogger")
	}

	fieldsLogger.Info("fields test")

	if !strings.Contains(buf.String(), "field1") || !strings.Contains(buf.String(), "value1") ||
		!strings.Contains(buf.String(), "field2") || !strings.Contains(buf.String(), "42") {
		t.Errorf("Expected fields not found in output: %s", buf.String())
	}
}

func TestContextLoggerContextMethods(t *testing.T) {
	var buf strings.Builder
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	logger := NewLogger(handler)

	ctx := context.WithValue(context.Background(), contextKey("test_key"), "test_value")
	ctxLogger := logger.WithContext(ctx)

	ctx2 := context.WithValue(context.Background(), contextKey("test_key2"), "test_value2")

	tests := []struct {
		name      string
		logFunc   func()
		checkFunc func() bool
	}{
		{
			name: "DebugContext method",
			logFunc: func() {
				ctxLogger.DebugContext(ctx2, "debug context message", "key", "value")
			},
			checkFunc: func() bool {
				return strings.Contains(buf.String(), "debug context message")
			},
		},
		{
			name: "InfoContext method",
			logFunc: func() {
				ctxLogger.InfoContext(ctx2, "info context message", "key", "value")
			},
			checkFunc: func() bool {
				return strings.Contains(buf.String(), "info context message")
			},
		},
		{
			name: "WarnContext method",
			logFunc: func() {
				ctxLogger.WarnContext(ctx2, "warn context message", "key", "value")
			},
			checkFunc: func() bool {
				return strings.Contains(buf.String(), "warn context message")
			},
		},
		{
			name: "ErrorContext method",
			logFunc: func() {
				ctxLogger.ErrorContext(ctx2, "error context message", "key", "value")
			},
			checkFunc: func() bool {
				return strings.Contains(buf.String(), "error context message")
			},
		},
		{
			name: "PerfContext method",
			logFunc: func() {
				ctxLogger.PerfContext(ctx2, "perf context message", "key", "value")
			},
			checkFunc: func() bool {
				return strings.Contains(buf.String(), "perf context message")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			tt.logFunc()
			if !tt.checkFunc() {
				t.Errorf("Expected log message not found in output: %s", buf.String())
			}
		})
	}
}
