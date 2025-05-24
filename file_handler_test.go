package multilog

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNewFileHandler(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "filehandler_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	filePath := filepath.Join(tempDir, "test.log")

	tests := []struct {
		name          string
		opts          CustomHandlerOptions
		expectEnabled bool
	}{
		{
			name: "basic enabled handler",
			opts: CustomHandlerOptions{
				Level:   "info",
				Enabled: true,
				File:    filePath,
				Pattern: "[time] [level] [message]",
			},
			expectEnabled: true,
		},
		{
			name: "disabled handler",
			opts: CustomHandlerOptions{
				Level:   "info",
				Enabled: false,
				File:    filePath,
			},
			expectEnabled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, err := NewFileHandler(tt.opts)
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

func TestFileHandlerLogging(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "filehandler_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	filePath := filepath.Join(tempDir, "test.log")

	opts := CustomHandlerOptions{
		Level:               "info",
		Enabled:             true,
		File:                filePath,
		Pattern:             "[msg]",
		PatternPlaceholders: []string{"[msg]"},
	}

	handler, err := NewFileHandler(opts)
	if err != nil {
		t.Fatalf("Failed to create file handler: %v", err)
	}

	record := slog.Record{
		Time:    time.Now(),
		Message: "test-file-message",
		Level:   slog.LevelInfo,
	}

	if err := handler.Handle(context.Background(), record); err != nil {
		t.Fatalf("Failed to handle record: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatalf("Log file was not created at %s", filePath)
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	if len(content) == 0 {
		t.Error("Log file is empty")
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "test-file-message") {
		t.Errorf("Log file does not contain expected message. Content: %s", contentStr)
	}
}

func TestFileHandlerWithAttrsAndGroups(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "filehandler_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	filePath := filepath.Join(tempDir, "test.log")

	opts := CustomHandlerOptions{
		Level:   "info",
		Enabled: true,
		File:    filePath,
	}

	handler, err := NewFileHandler(opts)
	if err != nil {
		t.Fatalf("Failed to create file handler: %v", err)
	}

	attrsHandler := handler.WithAttrs([]slog.Attr{
		slog.String("attr1", "value1"),
		slog.Int("attr2", 42),
	})

	if attrsHandler == nil {
		t.Fatal("WithAttrs returned nil")
	}

	groupHandler := handler.WithGroup("test group")

	if groupHandler == nil {
		t.Fatal("WithGroup returned nil")
	}
}
