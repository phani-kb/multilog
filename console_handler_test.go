package multilog

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"
)

func TestNewConsoleHandler(t *testing.T) {
	tests := []struct {
		name          string
		opts          CustomHandlerOptions
		expectEnabled bool
	}{
		{
			name: "enabled handler",
			opts: CustomHandlerOptions{
				Level:   "info",
				Enabled: true,
				Pattern: "[time] [level] [message]",
			},
			expectEnabled: true,
		},
		{
			name: "disabled handler",
			opts: CustomHandlerOptions{
				Level:   "info",
				Enabled: false,
			},
			expectEnabled: false,
		},
		{
			name: "debug level",
			opts: CustomHandlerOptions{
				Level:   "debug",
				Enabled: true,
			},
			expectEnabled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewConsoleHandler(tt.opts)

			if handler == nil {
				t.Fatal("Handler should not be nil")
			}

			if enabled := handler.Enabled(context.Background(), slog.LevelInfo); enabled != tt.expectEnabled {
				t.Errorf("Expected Enabled() to be %v, got %v", tt.expectEnabled, enabled)
			}
		})
	}
}

func TestConsoleHandlerOutput(t *testing.T) {
	// Save original stdout
	originalStdout := os.Stdout

	// Create a pipe to capture stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}

	// Set stdout to our pipe
	os.Stdout = w

	// Restore stdout when done
	defer func() {
		os.Stdout = originalStdout
	}()

	// Create handler with correct pattern format
	opts := CustomHandlerOptions{
		Level:               "info",
		Enabled:             true,
		Pattern:             "[msg]",
		PatternPlaceholders: []string{"[msg]"},
	}

	handler := NewConsoleHandler(opts)

	// Create a test record with a simple message that won't need escaping
	record := slog.Record{
		Time:    time.Now(),
		Message: "test-console-message",
		Level:   slog.LevelInfo,
	}

	// Handle the record
	err = handler.Handle(context.Background(), record)
	if err != nil {
		t.Fatalf("Failed to handle record: %v", err)
	}

	// Close the writing end of the pipe to make ReadAll complete
	w.Close()

	// Read all output
	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	if err != nil {
		t.Fatalf("Failed to read from pipe: %v", err)
	}

	// Check output - now just checking if the message text appears anywhere in the output
	output := buf.String()
	if !strings.Contains(output, "test-console-message") {
		t.Errorf("Expected output to contain 'test-console-message', got: %s", output)
	}
}

func TestConsoleHandlerWithAttrsAndGroups(t *testing.T) {
	opts := CustomHandlerOptions{
		Level:   "info",
		Enabled: true,
	}

	handler := NewConsoleHandler(opts)

	// Test WithAttrs
	attrsHandler := handler.WithAttrs([]slog.Attr{
		slog.String("attr1", "value1"),
		slog.Int("attr2", 42),
	})

	if attrsHandler == nil {
		t.Fatal("WithAttrs returned nil")
	}

	// Test WithGroup
	groupHandler := handler.WithGroup("testgroup")

	if groupHandler == nil {
		t.Fatal("WithGroup returned nil")
	}
}

func TestConsoleHandlerLevels(t *testing.T) {
	// Test that logging is controlled by level thresholds
	opts := CustomHandlerOptions{
		Level:   "warn", // Only warn and above should be logged
		Enabled: true,
	}

	handler := NewConsoleHandler(opts)

	// Debug should be disabled
	if handler.Enabled(context.Background(), slog.LevelDebug) {
		t.Error("Debug level should be disabled when level is set to warn")
	}

	// Info should be disabled
	if handler.Enabled(context.Background(), slog.LevelInfo) {
		t.Error("Info level should be disabled when level is set to warn")
	}

	// Warn should be enabled
	if !handler.Enabled(context.Background(), slog.LevelWarn) {
		t.Error("Warn level should be enabled when level is set to warn")
	}

	// Error should be enabled
	if !handler.Enabled(context.Background(), slog.LevelError) {
		t.Error("Error level should be enabled when level is set to warn")
	}
}
