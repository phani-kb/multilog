package main

import (
	"log/slog"

	"github.com/phani-kb/multilog"
)

func singleLetterLevel() {
	handler := multilog.NewConsoleHandler(multilog.CustomHandlerOptions{
		Level:                "info",
		Enabled:              true,
		Pattern:              "[time] [level] [msg]",
		UseSingleLetterLevel: true,
	})

	logger := multilog.NewLogger(handler)

	slog.SetDefault(logger.Logger)

	LogMessages(logger)
}
