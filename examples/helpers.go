package main

import (
	"log/slog"

	"github.com/phani-kb/multilog"
)

// LogMessages demonstrates the usage of various logging methods
// with the provided logger instance.
func LogMessages(logger *multilog.Logger) {
	// Basic logging
	slog.Debug("Debug message")
	slog.Info("Info message", "user", "john", "action", "login")
	slog.Warn("Warning message", "temperature", 80)
	slog.Error("Error occurred", "err", "file not found")

	// Format-style logging
	logger.Debugf("Debug message with %s formatting", "string")
	logger.Infof("User %s logged in from %s", "john", "192.168.1.1")
	logger.Warnf("Temperature is %d degrees", 80)
	logger.Errorf("Failed to open file: %v", "permission denied")

	// Performance logging
	logger.Perf("Database operation", "operation", "query", "table", "users")
}
