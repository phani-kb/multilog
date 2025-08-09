// Example of using a multilog package for logging with multiple handlers
package main

import (
	"log/slog"

	"github.com/phani-kb/multilog"
)

func basicExample() {
	// Create a console handler
	consoleHandler := multilog.NewConsoleHandler(multilog.CustomHandlerOptions{
		Level:   "perf",
		Enabled: true,
		Pattern: "[time] [level] [msg]",
	})

	// Create a file handler with rotation
	fileHandler, err := multilog.NewFileHandler(multilog.CustomHandlerOptions{
		Level:      "debug",
		Enabled:    true,
		Pattern:    "[datetime] [level] [source] [msg]",
		File:       "logs/app.log",
		MaxSize:    5,
		MaxBackups: 3,
		MaxAge:     7,
	})
	if err != nil {
		slog.Error("Error creating file handler")
		return
	}

	// Create a JSON handler
	jsonHandler, err := multilog.NewJSONHandler(multilog.CustomHandlerOptions{
		Level:   "perf",
		Enabled: true,
		Pattern: "[date] [level] [source] [msg]",
		File:    "logs/app.json",
	}, nil)
	if err != nil {
		slog.Error("Error creating JSON handler")
		return
	}

	// Create a logger with multiple handlers
	logger := multilog.NewLogger(consoleHandler, fileHandler, jsonHandler)

	// Set as default slog logger
	slog.SetDefault(logger.Logger)

	LogMessages(logger)
}
