// Example of using a multilog with a configuration file to set up multiple handlers.
package main

import (
	"log"
	"log/slog"

	"github.com/phani-kb/multilog"
)

func configExample() {
	configData := []byte(`
multilog:
  handlers:
    - type: console
      level: debug
      enabled: true
      use_single_letter_level: true
      pattern: "[datetime] [[level]] [msg]"
    - type: file
      subtype: text
      level: debug
      enabled: true
      pattern: "[date] - [[time]] [[level]] [[source]] [msg]"
      file: logs/app.log
      max_size: 5 # MB
      max_backups: 7
      max_age: 1 # days
    - type: file
      subtype: json
      level: debug
      enabled: true
      pattern_placeholders: "[datetime], [level], [source], [msg], [perf]"
      file: logs/app.json
      max_size: 5 # MB
      max_backups: 7
      max_age: 1 # days`)

	cfg, err := multilog.NewConfigFromData(configData)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	handlers, err := multilog.CreateHandlers(cfg)
	if err != nil {
		log.Fatalf("Failed to create handlers: %v", err)
	}

	logger := multilog.NewLogger(handlers...)

	slog.SetDefault(logger.Logger)

	LogMessages(logger)
}
