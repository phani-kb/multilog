package multilog

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"
)

func TestNewJsonHandler(t *testing.T) {
	tests := []struct {
		name          string
		opts          CustomHandlerOptions
		expectEnabled bool
	}{
		{
			name: "basic enabled handler",
			opts: CustomHandlerOptions{
				Level:     "info",
				Enabled:   true,
				AddSource: true,
			},
			expectEnabled: true,
		},
		{
			name: "disabled handler",
			opts: CustomHandlerOptions{
				Level:     "info",
				Enabled:   false,
				AddSource: true,
			},
			expectEnabled: false,
		},
		{
			name: "custom level",
			opts: CustomHandlerOptions{
				Level:     "debug",
				Enabled:   true,
				AddSource: false,
			},
			expectEnabled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := os.CreateTemp("", "json_handler_test")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer os.Remove(f.Name())
			defer f.Close()

			tt.opts.File = f.Name()

			handler, err := NewJSONHandler(tt.opts, nil)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if handler == nil {
				t.Fatal("Handler should not be nil")
			}

			if enabled := handler.Enabled(context.Background(), slog.LevelInfo); enabled != tt.expectEnabled {
				t.Errorf("Expected Enabled() to be %v, got %v", tt.expectEnabled, enabled)
			}
		})
	}
}

func TestJsonHandlerHandle(t *testing.T) {
	f, err := os.CreateTemp("", "json_handler_test")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	opts := CustomHandlerOptions{
		Level:               "debug",
		Enabled:             true,
		AddSource:           true,
		File:                f.Name(),
		PatternPlaceholders: []string{"[msg]"},
	}

	handler, err := NewJSONHandler(opts, nil)
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}

	record := slog.Record{
		Time:    time.Now(),
		Message: "test-json-message",
		Level:   slog.LevelInfo,
	}

	record.AddAttrs(slog.String("key1", "value1"))

	err = handler.Handle(context.Background(), record)
	if err != nil {
		t.Fatalf("Handler.Handle failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	content, err := os.ReadFile(f.Name())
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "test-json-message") {
		t.Fatalf("Output doesn't contain expected message. Content: %s", contentStr)
	}

	lines := strings.Split(strings.TrimSpace(contentStr), "\n")
	if len(lines) == 0 {
		t.Fatal("No lines found in output")
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(lines[0]), &data); err != nil {
		t.Fatalf("Output is not valid JSON: %v\nContent: %s", err, lines[0])
	}

	if msg, ok := data["msg"]; !ok ||
		!strings.Contains(fmt.Sprintf("%v", msg), "test-json-message") {
		t.Errorf("Expected JSON to contain message 'test-json-message', got: %v", data)
	}
}

func TestJsonHandlerWithAttrsAndGroups(t *testing.T) {
	f, err := os.CreateTemp("", "json_handler_test")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	opts := CustomHandlerOptions{
		Level:     "debug",
		Enabled:   true,
		AddSource: true,
		File:      f.Name(),
	}

	handler, err := NewJSONHandler(opts, nil)
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}

	attrHandler := handler.WithAttrs([]slog.Attr{
		slog.String("attr1", "value1"),
	})

	if attrHandler == nil {
		t.Fatal("WithAttrs returned nil")
	}

	groupHandler := handler.WithGroup("testgroup")

	if groupHandler == nil {
		t.Fatal("WithGroup returned nil")
	}
}

type MockCustomHandler struct {
	opts    *CustomHandlerOptions
	writer  BufferedWriterInterface
	sb      *strings.Builder
	handler slog.Handler
}

func (h *MockCustomHandler) Enabled(_ context.Context, level slog.Level) bool {
	return h.opts.Enabled && level >= GetSlogLevel(h.opts.Level)
}

func (h *MockCustomHandler) Handle(ctx context.Context, r slog.Record) error {
	return h.handler.Handle(ctx, r)
}

func (h *MockCustomHandler) WithAttrs(_ []slog.Attr) slog.Handler {
	return h
}

func (h *MockCustomHandler) WithGroup(_ string) slog.Handler {
	return h
}

func (h *MockCustomHandler) GetOptions() *CustomHandlerOptions {
	return h.opts
}

func (h *MockCustomHandler) GetStringBuilder() *strings.Builder {
	if h.sb == nil {
		h.sb = &strings.Builder{}
	}
	return h.sb
}

func (h *MockCustomHandler) GetWriter() *bufio.Writer {
	return nil
}

func (h *MockCustomHandler) GetSlogHandler() slog.Handler {
	return h.handler
}

func (h *MockCustomHandler) GetKeyValue(key string, sb *strings.Builder, removeKey bool) string {
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

func (h *MockCustomHandler) CustomWrite(output string) error {
	if h.writer == nil {
		return fmt.Errorf("no writer available")
	}

	_, err := h.writer.WriteString(output)
	if err != nil {
		return fmt.Errorf("failed to write log message: %w", err)
	}

	err = h.writer.Flush()
	if err != nil {
		return fmt.Errorf("failed to flush writer: %w", err)
	}

	return nil
}

func TestJsonHandler_GetKeyValue(t *testing.T) {
	jh := &JSONHandler{}

	tests := []struct {
		name      string
		json      string
		key       string
		removeKey bool
		want      string
		wantJSON  string
	}{
		{
			name:      "key exists - don't remove",
			json:      `{"level":"info","message":"test message","time":"2023-01-01T12:00:00Z"}`,
			key:       "level",
			removeKey: false,
			want:      "info",
			wantJSON:  `{"level":"info","message":"test message","time":"2023-01-01T12:00:00Z"}`,
		},
		{
			name:      "key exists - remove",
			json:      `{"level":"info","message":"test message","time":"2023-01-01T12:00:00Z"}`,
			key:       "level",
			removeKey: true,
			want:      "info",
			wantJSON:  `{"message":"test message","time":"2023-01-01T12:00:00Z"}`,
		},
		{
			name:      "key doesn't exist",
			json:      `{"level":"info","message":"test message"}`,
			key:       "nonexistent",
			removeKey: false,
			want:      "",
			wantJSON:  `{"level":"info","message":"test message"}`,
		},
		{
			name:      "invalid JSON",
			json:      `{invalid_json}`,
			key:       "level",
			removeKey: false,
			want:      "",
			wantJSON:  `{invalid_json}`,
		},
		{
			name:      "numeric value",
			json:      `{"count":42,"message":"test message"}`,
			key:       "count",
			removeKey: true,
			want:      "42",
			wantJSON:  `{"message":"test message"}`,
		},
		{
			name:      "boolean value",
			json:      `{"success":true,"message":"test message"}`,
			key:       "success",
			removeKey: false,
			want:      "true",
			wantJSON:  `{"success":true,"message":"test message"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sb := &strings.Builder{}
			sb.WriteString(tt.json)

			got := jh.GetKeyValue(tt.key, sb, tt.removeKey)

			if got != tt.want {
				t.Errorf("GetKeyValue() = %v, want %v", got, tt.want)
			}

			if tt.name == "invalid JSON" {
				if sb.String() != tt.wantJSON {
					t.Errorf("JSON after extraction = %v, want %v", sb.String(), tt.wantJSON)
				}
				return
			}

			var actualObj map[string]interface{}
			var expectedObj map[string]interface{}

			if err := json.Unmarshal([]byte(sb.String()), &actualObj); err != nil {
				t.Errorf("Could not parse actual JSON: %v", err)
				return
			}

			if err := json.Unmarshal([]byte(tt.wantJSON), &expectedObj); err != nil {
				t.Errorf("Could not parse expected JSON: %v", err)
				return
			}

			if !mapsEqual(actualObj, expectedObj) {
				t.Errorf("JSON after extraction = %v, want %v", sb.String(), tt.wantJSON)
			}
		})
	}
}

func mapsEqual(a, b map[string]interface{}) bool {
	if len(a) != len(b) {
		return false
	}

	for k, v1 := range a {
		v2, ok := b[k]
		if !ok || v1 != v2 {
			return false
		}
	}

	return true
}

func TestJsonHandler_Handle_Errors(t *testing.T) {
	f, err := os.CreateTemp("", "json_handler_test")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	opts := CustomHandlerOptions{
		Level:               "debug",
		Enabled:             true,
		AddSource:           true,
		File:                f.Name(),
		PatternPlaceholders: []string{"[msg]"},
	}

	record := slog.Record{
		Time:    time.Now(),
		Message: "test-json-message",
		Level:   slog.LevelInfo,
	}
	record.AddAttrs(slog.String("key1", "value1"))

	sb := &strings.Builder{}
	mockSlogHandler := &MockSlogHandler{err: fmt.Errorf("mock error")}
	mockHandler := &MockCustomHandler{
		opts:    &opts,
		sb:      sb,
		handler: mockSlogHandler,
		writer:  &MockBufferedWriter{},
	}

	jsonHandler := &JSONHandler{Handler: mockHandler}

	err = jsonHandler.Handle(context.Background(), record)
	if err == nil || !strings.Contains(err.Error(), "failed to handle record") {
		t.Fatalf("Expected error containing 'failed to handle record', got: %v", err)
	}

	validSb := &strings.Builder{}
	validSb.WriteString(`{"level":"INFO","msg":"test-json-message","key1":"value1"}`)

	mockWriteHandler := &MockCustomHandler{
		opts:    &opts,
		sb:      validSb,
		handler: &MockSlogHandler{},
		writer:  &MockBufferedWriter{writeStringErr: fmt.Errorf("mock write error")},
	}

	jsonHandler = &JSONHandler{Handler: mockWriteHandler}

	err = jsonHandler.Handle(context.Background(), record)
	if err == nil || !strings.Contains(err.Error(), "failed to write log message") {
		t.Fatalf("Expected error containing 'failed to write log message', got: %v", err)
	}

	validSb2 := &strings.Builder{}
	validSb2.WriteString(`{"level":"INFO","msg":"test-json-message","key1":"value1"}`)

	mockFlushHandler := &MockCustomHandler{
		opts:    &opts,
		sb:      validSb2,
		handler: &MockSlogHandler{},
		writer:  &MockBufferedWriter{flushErr: fmt.Errorf("mock flush error")},
	}

	jsonHandler = &JSONHandler{Handler: mockFlushHandler}

	err = jsonHandler.Handle(context.Background(), record)
	if err == nil || !strings.Contains(err.Error(), "failed to flush writer") {
		t.Fatalf("Expected error containing 'failed to flush writer', got: %v", err)
	}
}

type BufferedWriterInterface interface {
	WriteString(s string) (int, error)
	Flush() error
	Write(p []byte) (int, error)
}

type MockBufferedWriter struct {
	writeStringErr error
	flushErr       error
}

func (m *MockBufferedWriter) WriteString(s string) (int, error) {
	if m.writeStringErr != nil {
		return 0, m.writeStringErr
	}
	return len(s), nil
}

func (m *MockBufferedWriter) Flush() error {
	return m.flushErr
}

func (m *MockBufferedWriter) Write(p []byte) (int, error) {
	return len(p), nil
}

type MockSlogHandler struct {
	err error
}

func (m *MockSlogHandler) Enabled(_ context.Context, _ slog.Level) bool {
	return true
}

func (m *MockSlogHandler) Handle(_ context.Context, _ slog.Record) error {
	return m.err
}

func (m *MockSlogHandler) WithAttrs(_ []slog.Attr) slog.Handler {
	return m
}

func (m *MockSlogHandler) WithGroup(_ string) slog.Handler {
	return m
}

type MockErrorWriter struct {
	writeErr error
	flushErr error
}

func (m *MockErrorWriter) Write(p []byte) (int, error) {
	if m.writeErr != nil {
		return 0, m.writeErr
	}
	return len(p), nil
}

func (m *MockErrorWriter) Flush() error {
	if m.flushErr != nil {
		return m.flushErr
	}
	return nil
}

func TestJsonHandler_Handle_PerfLevel(t *testing.T) {
	f, err := os.CreateTemp("", "json_handler_perf_test")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	opts := CustomHandlerOptions{
		Level:               "debug",
		Enabled:             true,
		AddSource:           true,
		File:                f.Name(),
		PatternPlaceholders: []string{"[level]", "[msg]"}, // Explicitly exclude [perf]
	}
	handler, err := NewJSONHandler(opts, nil)
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}

	// Create a record with LevelPerf to trigger line 72
	record := slog.Record{
		Time:    time.Now(),
		Message: "performance-test-message",
		Level:   LevelPerf,
	}
	record.AddAttrs(slog.String("metric_id", "performance-test"))

	// Handle the record
	err = handler.Handle(context.Background(), record)
	if err != nil {
		t.Fatalf("Handler.Handle failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond) // Wait for I/O
	content, err := os.ReadFile(f.Name())
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "performance-test-message") {
		t.Fatalf("Output doesn't contain expected message. Content: %s", contentStr)
	}

	// Parse the JSON to verify it contains perf metrics
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(contentStr)), &data); err != nil {
		t.Fatalf("Output is not valid JSON: %v\nContent: %s", err, contentStr)
	}

	// Check if performance metrics were added (line 72)
	perfValue, perfExists := data["perf"]
	if !perfExists {
		t.Error("Performance metrics not added to JSON output")
	} else {
		perfStr := fmt.Sprintf("%v", perfValue)
		if !strings.Contains(perfStr, "goroutines:") || !strings.Contains(perfStr, "heap_alloc:") && !strings.Contains(perfStr, "cpu") {
			t.Errorf("Performance metrics incorrect format: %s", perfStr)
		}
	}
}
