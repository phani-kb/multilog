package multilog

import (
	"context"
	"errors"
	"log/slog"
	"testing"
)

type mockHandler struct {
	enabledLevel slog.Level
	handleCalled bool
	returnError  bool
	attrAdded    bool
	groupAdded   bool
}

func (m *mockHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= m.enabledLevel
}

func (m *mockHandler) Handle(_ context.Context, _ slog.Record) error {
	m.handleCalled = true
	if m.returnError {
		return errors.New("mock handler error")
	}
	return nil
}

func (m *mockHandler) WithAttrs(_ []slog.Attr) slog.Handler {
	m.attrAdded = true
	return m
}

func (m *mockHandler) WithGroup(_ string) slog.Handler {
	m.groupAdded = true
	return m
}

func TestNewAggregator(t *testing.T) {
	handler1 := &mockHandler{}
	handler2 := &mockHandler{}
	aggregator := NewAggregator(handler1, handler2)

	if aggregator == nil {
		t.Fatal("expected non-nil aggregator")
	}
}

func TestAggregatorEnabled(t *testing.T) {
	tests := []struct {
		name     string
		handlers []slog.Handler
		level    slog.Level
		want     bool
	}{
		{
			name: "no handler enabled",
			handlers: []slog.Handler{
				&mockHandler{enabledLevel: slog.LevelError},
				&mockHandler{enabledLevel: slog.LevelError},
			},
			level: slog.LevelInfo,
			want:  false,
		},
		{
			name: "one handler enabled",
			handlers: []slog.Handler{
				&mockHandler{enabledLevel: slog.LevelError},
				&mockHandler{enabledLevel: slog.LevelInfo},
			},
			level: slog.LevelInfo,
			want:  true,
		},
		{
			name: "all handlers enabled",
			handlers: []slog.Handler{
				&mockHandler{enabledLevel: slog.LevelDebug},
				&mockHandler{enabledLevel: slog.LevelDebug},
			},
			level: slog.LevelInfo,
			want:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			aggregator := NewAggregator(tt.handlers...)
			got := aggregator.Enabled(context.Background(), tt.level)
			if got != tt.want {
				t.Errorf("Enabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAggregatorHandle(t *testing.T) {
	tests := []struct {
		name          string
		handlers      []slog.Handler
		expectError   bool
		expectedCalls int
	}{
		{
			name: "all handlers process successfully",
			handlers: []slog.Handler{
				&mockHandler{enabledLevel: slog.LevelInfo},
				&mockHandler{enabledLevel: slog.LevelInfo},
			},
			expectError:   false,
			expectedCalls: 2,
		},
		{
			name: "handler returns error",
			handlers: []slog.Handler{
				&mockHandler{enabledLevel: slog.LevelInfo, returnError: true},
				&mockHandler{enabledLevel: slog.LevelInfo},
			},
			expectError:   true,
			expectedCalls: 2,
		},
		{
			name: "some handlers disabled",
			handlers: []slog.Handler{
				&mockHandler{enabledLevel: slog.LevelError},
				&mockHandler{enabledLevel: slog.LevelInfo},
			},
			expectError:   false,
			expectedCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			aggregator := NewAggregator(tt.handlers...)
			record := slog.Record{}
			record.Level = slog.LevelInfo
			record.Message = "test message"

			err := aggregator.Handle(context.Background(), record)
			if (err != nil) != tt.expectError {
				t.Errorf("Handle() error = %v, expectError %v", err, tt.expectError)
			}

			callCount := 0
			for _, h := range tt.handlers {
				if mh, ok := h.(*mockHandler); ok && mh.handleCalled {
					callCount++
				}
			}

			if callCount != tt.expectedCalls {
				t.Errorf("Expected %d handler calls, got %d", tt.expectedCalls, callCount)
			}
		})
	}
}

func TestAggregatorWithAttrs(t *testing.T) {
	handler1 := &mockHandler{}
	handler2 := &mockHandler{}
	aggregator := NewAggregator(handler1, handler2)

	attrs := []slog.Attr{slog.String("key", "value")}
	newAggregator := aggregator.WithAttrs(attrs)

	if newAggregator == nil {
		t.Fatal("expected non-nil aggregator")
	}

	if handler1.attrAdded != true || handler2.attrAdded != true {
		t.Error("WithAttrs was not called on all handlers")
	}
}

func TestAggregatorWithGroup(t *testing.T) {
	handler1 := &mockHandler{}
	handler2 := &mockHandler{}
	aggregator := NewAggregator(handler1, handler2)

	groupName := "testGroup"
	newAggregator := aggregator.WithGroup(groupName)

	if newAggregator == nil {
		t.Fatal("expected non-nil aggregator")
	}

	if handler1.groupAdded != true || handler2.groupAdded != true {
		t.Error("WithGroup was not called on all handlers")
	}
}

func TestGetOtherSourceValue(t *testing.T) {
	tests := []struct {
		name     string
		fn       string
		file     string
		line     int
		expected string
	}{
		{
			name:     "Normal values",
			fn:       "TestFunction",
			file:     "/path/to/file.go",
			line:     42,
			expected: "file.go:42:TestFunction",
		},
		{
			name:     "Empty function name",
			fn:       "",
			file:     "/path/to/file.go",
			line:     42,
			expected: "file.go:42:",
		},
		{
			name:     "Empty file path",
			fn:       "TestFunction",
			file:     "",
			line:     42,
			expected: ".:42:TestFunction",
		},
		{
			name:     "Zero line number",
			fn:       "TestFunction",
			file:     "/path/to/file.go",
			line:     0,
			expected: "file.go:0:TestFunction",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetOtherSourceValue(tt.fn, tt.file, tt.line)
			if result != tt.expected {
				t.Errorf("GetOtherSourceValue() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestGetCallerDetails(t *testing.T) {
	tests := []struct {
		name     string
		fnLine   string
		fileLine string
		wantFn   string
		wantFile string
		wantLine int
	}{
		{
			name:     "Normal case",
			fnLine:   "github.com/testuser/multilog.TestFunction(0x1234)",
			fileLine: "/home/user/go/src/github.com/testuser/multilog/file.go:42 +0x1a3",
			wantFn:   "github.com/testuser/multilog.TestFunction",
			wantFile: "file.go",
			wantLine: 42,
		},
		{
			name:     "No parameters",
			fnLine:   "github.com/testuser/multilog.TestFunction",
			fileLine: "/home/user/go/src/github.com/testuser/multilog/file.go:42 +0x1a3",
			wantFn:   "github.com/testuser/multilog.TestFunction",
			wantFile: "file.go",
			wantLine: 42,
		},
		{
			name:     "With dots in filename",
			fnLine:   "main.main()",
			fileLine: "/home/user/project/main.v2.go:100 +0x1a3",
			wantFn:   "main.main",
			wantFile: "main.v2.go",
			wantLine: 100,
		},
		{
			name:     "Unusual names",
			fnLine:   "github.com/testuser/multilog.(*Type).Method(0x1234)",
			fileLine: "/some-path/with-dashes/file_test.go:999 +0x1a3",
			wantFn:   "github.com/testuser/multilog.(*Type).Method",
			wantFile: "file_test.go",
			wantLine: 999,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFn, gotFile, gotLine := getCallerDetails(tt.fnLine, tt.fileLine)
			if gotFn != tt.wantFn {
				t.Errorf("getCallerDetails() function = %q, want %q", gotFn, tt.wantFn)
			}
			if gotFile != tt.wantFile {
				t.Errorf("getCallerDetails() file = %q, want %q", gotFile, tt.wantFile)
			}
			if gotLine != tt.wantLine {
				t.Errorf("getCallerDetails() line = %d, want %d", gotLine, tt.wantLine)
			}
		})
	}
}

func TestAggregator(t *testing.T) {
	handler1 := NewTestHandler(t)
	handler2 := NewTestHandler(t)

	aggregator := NewAggregator(handler1, handler2)

	ctx := context.Background()
	if !aggregator.Enabled(ctx, slog.LevelInfo) {
		t.Error("Aggregator should be enabled when handlers are enabled")
	}

	record := slog.Record{}
	record.Level = slog.LevelInfo
	record.Message = "test message"

	err := aggregator.Handle(ctx, record)
	if err != nil {
		t.Errorf("Aggregator.Handle should not return error, got: %v", err)
	}

	if !handler1.Called() {
		t.Error("First handler should have been called")
	}

	if !handler2.Called() {
		t.Error("Second handler should have been called")
	}

	attrs := []slog.Attr{slog.String("key", "value")}
	newAgg := aggregator.WithAttrs(attrs)
	if newAgg == nil {
		t.Error("Aggregator.WithAttrs should return non-nil aggregator")
	}

	groupAgg := aggregator.WithGroup("testGroup")
	if groupAgg == nil {
		t.Error("Aggregator.WithGroup should return non-nil aggregator")
	}

	handlerWithError := &ErrorHandler{err: errors.New("test error")}
	errorAgg := NewAggregator(handlerWithError)

	if err := errorAgg.Handle(ctx, record); err == nil {
		t.Error("Aggregator should return error when handler returns error")
	}

	t.Run("handler returns error", func(t *testing.T) {
		counter := &CountingHandler{}

		errorHandler := &ErrorHandler{err: errors.New("test error")}

		agg := NewAggregator(errorHandler, counter)

		err := agg.Handle(ctx, record)
		if err == nil {
			t.Fatal("Expected error from handler, got nil")
		}

		if counter.callCount != 1 {
			t.Errorf("Expected counter to be called once, got %d calls", counter.callCount)
		}
	})

	t.Run("all handlers called on success", func(t *testing.T) {
		counter1 := &CountingHandler{}
		counter2 := &CountingHandler{}

		agg := NewAggregator(counter1, counter2)

		err := agg.Handle(ctx, record)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if counter1.callCount != 1 {
			t.Errorf("Expected counter1 to be called once, got %d", counter1.callCount)
		}

		if counter2.callCount != 1 {
			t.Errorf("Expected counter2 to be called once, got %d", counter2.callCount)
		}
	})
}

type CountingHandler struct {
	callCount int
}

func (h *CountingHandler) Enabled(context.Context, slog.Level) bool {
	return true
}

func (h *CountingHandler) Handle(context.Context, slog.Record) error {
	h.callCount++
	return nil
}

func (h *CountingHandler) WithAttrs([]slog.Attr) slog.Handler {
	return h
}

func (h *CountingHandler) WithGroup(string) slog.Handler {
	return h
}

type ErrorHandler struct {
	err error
}

func (h *ErrorHandler) Enabled(context.Context, slog.Level) bool {
	return true
}

func (h *ErrorHandler) Handle(context.Context, slog.Record) error {
	return h.err
}

func (h *ErrorHandler) WithAttrs([]slog.Attr) slog.Handler {
	return h
}

func (h *ErrorHandler) WithGroup(string) slog.Handler {
	return h
}
