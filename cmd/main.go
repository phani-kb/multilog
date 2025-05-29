// Package main provides the entry point for the multilog application.
package main

import (
	"flag"
	"fmt"
	"log"
	"log/slog"
	"strings"

	"github.com/phani-kb/multilog"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "config.yml", "Path to configuration file")
	flag.Parse()

	if err := run(*configPath); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func run(configPath string) error {
	handler := slog.Default().Handler()
	if handler != nil {
		// Check if it's not a default handler by checking for the package name
		if h := fmt.Sprintf("%T", handler); strings.Contains(h, "slog.TextHandler") {
			logger := multilog.NewLogger(handler)
			logMessages(logger)
			return nil
		}
	}

	// Parse and validate config using multilog's built-in validation
	cfg, err := multilog.NewConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Create handlers with validated config
	hs, err := multilog.CreateHandlers(cfg)
	if err != nil {
		return fmt.Errorf("failed to create handlers: %w", err)
	}

	// Create logger and log messages
	logger := multilog.NewLogger(hs...)
	slog.SetDefault(logger.Logger)
	logMessages(logger)
	return nil
}

func logMessages(logger *multilog.Logger) {
	// Use the logger parameter consistently instead of the global slog
	logger.Debugf("Debugging information")
	logger.Infof("Starting application %s", "1")
	logger.Warnf("Something went wrong: %s", "warning details")
	logger.Errorf("An error occurred: %s", "permission denied")

	// Use the same logger instance for structured logging too
	logger.Debug("Debugging information with context", "component", "main")
	logger.Info("Starting application", "app_id", "1")
	logger.Warn("Something went wrong", "warning_code", "2")
	logger.Error("An error occurred", "error_code", "3")

	// Performance metrics logging
	logger.Perff("Performance metrics %s %d", "metric_id", 4)
	logger.Perf("Performance metrics", "metric_id", "4", "duration_ms", 250)
}
