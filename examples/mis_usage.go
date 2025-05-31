package main

import (
	"bufio"
	"context"
	"log/slog"
	"os"

	"github.com/phani-kb/multilog"
)

type contextKey string

func misUsage() {
	handler := multilog.NewConsoleHandler(multilog.CustomHandlerOptions{
		Level:                "info",
		Enabled:              true,
		Pattern:              "[time] [level] [msg]",
		UseSingleLetterLevel: true,
	})

	logger := multilog.NewLogger(handler)

	slog.SetDefault(logger.Logger)

	ctx := context.Background()

	// Add request ID to context using custom type key
	const requestIDKey contextKey = "request_id"
	ctx = context.WithValue(ctx, requestIDKey, "abc-123")

	logger.InfoContext(ctx, "Processing request for user %s", "john")

	replaceAttr := func(_ []string, a slog.Attr) slog.Attr {
		if a.Key == "password" {
			return slog.String("password", "****")
		}
		return a
	}

	opts := multilog.CustomHandlerOptions{
		Level:   "info",
		Enabled: true,
	}

	handler = multilog.NewCustomHandler(&opts, bufio.NewWriter(os.Stdout), replaceAttr)

	logger = multilog.NewLogger(handler)

	slog.SetDefault(logger.Logger)

	logger.Info("User login attempt", "user", "john", "password", "secret123")

	handler = multilog.NewConsoleHandler(multilog.CustomHandlerOptions{
		Level:           "debug",
		Enabled:         true,
		Pattern:         "[time] [level] [msg]",
		ValuePrefixChar: "<",
		ValueSuffixChar: ">",
	})

	logger = multilog.NewLogger(handler)

	slog.SetDefault(logger.Logger)

	logger.Info("Info message")
	logger.Debug("Debugging information", "user", "john", "action", "login")
}
