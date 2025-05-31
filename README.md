# multilog
A simple logging library for Go that enables simultaneous logging to multiple destinations with configurable formats and log levels.
## Features

- Multiple handler types:
  - Console handlers for terminal output
  - File handlers with log rotation
  - JSON handlers for structured logging
- Custom log levels including standard levels and performance logging
- Configurable log patterns with placeholders
- Support for structured logging via slog
- Log rotation via lumberjack
- Context-aware logging methods
- Performance metrics collection
- Custom attribute manipulation

## Installation

```bash
go get github.com/phani-kb/multilog
```
## Basic Usage

```go
package main

import (
	"log/slog"

	"github.com/phani-kb/multilog"
)

func main() {
	// Create a console handler
	consoleHandler := multilog.NewConsoleHandler(multilog.CustomHandlerOptions{
		Level:   "perf",
		Enabled: true,
		Pattern: "[time] [level] [msg]",
	})

	// Create a file handler with rotation
	fileHandler, _ := multilog.NewFileHandler(multilog.CustomHandlerOptions{
		Level:      "debug",
		Enabled:    true,
		Pattern:    "[datetime] [level] [source] [msg]",
		File:       "logs/app.log",
		MaxSize:    5,
		MaxBackups: 3,
		MaxAge:     7,
	})

	// Create a JSON handler
	jsonHandler, _ := multilog.NewJSONHandler(multilog.CustomHandlerOptions{
		Level:   "perf",
		Enabled: true,
		Pattern: "[date] [level] [source] [msg]",
		File:    "logs/app.json",
	}, nil)

	// Create a logger with multiple handlers
	logger := multilog.NewLogger(consoleHandler, fileHandler, jsonHandler)

	// Set as default slog logger
	slog.SetDefault(logger.Logger)

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
```

### Example Output

Console output (with default pattern):
```
10:10:09 INFO Info message [user=john action=login]
10:10:09 WARN Warning message [temperature=80]
10:10:09 ERROR Error occurred [err="file not found"]
10:10:09 INFO User john logged in from 192.168.1.1
10:10:09 WARN Temperature is 80 degrees
10:10:09 ERROR Failed to open file: permission denied
10:10:09 PERF Database operation [goroutines:3,alloc:0.419044 MB,sys:6.334976 MB,heap_alloc:0.419044 MB,heap_sys:3.687500 MB,heap_idle:2.625000 MB,heap_inuse:1.062500 MB,stack_sys:0.312500 MB] [operation=query table=users]
```

File output (with source information):
```
2025-05-29 10:10:09 DEBUG unknown Debug message
2025-05-29 10:10:09 INFO unknown Info message [user=john action=login]
2025-05-29 10:10:09 WARN unknown Warning message [temperature=80]
2025-05-29 10:10:09 ERROR unknown Error occurred [err="file not found"]
2025-05-29 10:10:09 DEBUG basic_usage.go:50:main.main Debug message with string formatting
2025-05-29 10:10:09 INFO basic_usage.go:51:main.main User john logged in from 192.168.1.1
2025-05-29 10:10:09 WARN basic_usage.go:52:main.main Temperature is 80 degrees
2025-05-29 10:10:09 ERROR basic_usage.go:53:main.main Failed to open file: permission denied
2025-05-29 10:10:09 PERF basic_usage.go:56:main.main Database operation [goroutines:3,alloc:0.429909 MB,sys:6.334976 MB,heap_alloc:0.429909 MB,heap_sys:3.687500 MB,heap_idle:2.593750 MB,heap_inuse:1.093750 MB,stack_sys:0.312500 MB] [operation=query table=users]
```

JSON output:
```json
{"action":"login","datetime":"2025-05-29 10:10:09","level":"INFO","msg":"Info message","source":"unknown","user":"john"}
{"datetime":"2025-05-29 10:10:09","level":"WARN","msg":"Warning message","source":"unknown","temperature":80}
{"datetime":"2025-05-29 10:10:09","err":"file not found","level":"ERROR","msg":"Error occurred","source":"unknown"}
{"datetime":"2025-05-29 10:10:09","level":"INFO","msg":"User john logged in from 192.168.1.1","source":"basic_usage.go:51:main.main"}
{"datetime":"2025-05-29 10:10:09","level":"WARN","msg":"Temperature is 80 degrees","source":"basic_usage.go:52:main.main"}
{"datetime":"2025-05-29 10:10:09","level":"ERROR","msg":"Failed to open file: permission denied","source":"basic_usage.go:53:main.main"}
{"datetime":"2025-05-29 10:10:09","level":"PERF","msg":"Database operation","operation":"query","perf":"goroutines:3,alloc:0.438530 MB,sys:6.334976 MB,heap_alloc:0.438530 MB,heap_sys:3.687500 MB,heap_idle:2.578125 MB,heap_inuse:1.109375 MB,stack_sys:0.312500 MB","source":"basic_usage.go:56:main.main","table":"users"}
```

## Configuration

Multilog can be configured via code or YAML configuration:

```yaml
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
      file: logs/output.log
      max_size: 5 # MB
      max_backups: 7
      max_age: 1 # days
    - type: file
      subtype: json
      level: debug
      enabled: true
      pattern_placeholders: "[datetime], [level], [source], [msg], [perf]"
      file: logs/output.json
      max_size: 5 # MB
      max_backups: 7
      max_age: 1 # days
```

Load configuration file:

```go
cfg, err := multilog.NewConfig("config.yml")
if err != nil {
    panic(err)
}

handlers, err := multilog.CreateHandlers(cfg)
if err != nil {
    panic(err)
}

logger := multilog.NewLogger(handlers...)
```

## Pattern Placeholders

Customize your log format with these placeholders:

- `[date]` - Date in YYYY-MM-DD format
- `[time]` - Time in HH:MM:SS format
- `[datetime]` - Combined date and time
- `[level]` - Log level (DEBUG, INFO, WARN, ERROR, PERF)
- `[msg]` - Log message
- `[source]` - Source file, line number, and function
- `[perf]` - Performance metrics (goroutines, heap, etc.)

## Log Levels

- `debug` - Detailed debugging information
- `info` - General operational information
- `warn` - Warning events that don't affect operation
- `error` - Error events that might still allow continued operation
- `perf` - Performance metrics (extends slog levels)

## Handler Types

### Console Handler

Outputs logs to stdout with customizable format:

```go
handler := multilog.NewConsoleHandler(multilog.CustomHandlerOptions{
    Level:                "debug",
    Enabled:              true,
    Pattern:              "[time] [level] [msg]",
    UseSingleLetterLevel: true,
})
```

### File Handler

Writes logs to a file with rotation support:

```go
handler, err := multilog.NewFileHandler(multilog.CustomHandlerOptions{
    Level:     "info",
    Enabled:   true,
    Pattern:   "[datetime] [level] [source] [msg]",
    File:      "logs/app.log",
    MaxSize:   5,    // megabytes
    MaxBackups: 3,   // number of backups
    MaxAge:    7,    // days
})
```

### JSON Handler

Structured logging in JSON format:

```go
handler, err := multilog.NewJsonHandler(multilog.CustomHandlerOptions{
    Level:   "debug",
    Enabled: true,
    PatternPlaceholders: []string{
        "[datetime]", "[level]", "[msg]", "[source]",
    },
    File:    "logs/app.json",
})
```

## Custom Handler Options

The `CustomHandlerOptions` struct provides extensive customization for all handlers:

| Option | Type | Description | Default |
|--------|------|-------------|---------|
| `Level` | string | Minimum log level to output | `"info"` |
| `SubType` | string | Handler subtype (e.g., "text", "json") | `"text"` |
| `Enabled` | bool | Whether the handler is active | `true` |
| `Pattern` | string | Log message format pattern | `"[time] [level] [msg]"` |
| `PatternPlaceholders` | []string | Placeholders for JSON handler | `[]string{"[datetime]", "[level]", "[msg]", "[source]"}` |
| `AddSource` | bool | Include source file/line information | `false` |
| `UseSingleLetterLevel` | bool | Use single letter for level (D,I,W,E,P) | `false` |
| `ValuePrefixChar` | string | Character before values | `""` |
| `ValueSuffixChar` | string | Character after values | `""` |
| `File` | string | Log file path | `""` |
| `MaxSize` | int | Max size in MB before rotation | `5` |
| `MaxBackups` | int | Max number of old log files | `1` |
| `MaxAge` | int | Max days to retain old logs | `1` |

### Single Letter Level Example

When `UseSingleLetterLevel` is enabled:

```go
handler := multilog.NewConsoleHandler(multilog.CustomHandlerOptions{
    Level:                "info",
    Enabled:              true,
    Pattern:              "[time] [level] [msg]",
    UseSingleLetterLevel: true,
})
```

Output:
```
10:13:53 I Info message [user=john action=login]
10:13:53 W Warning message [temperature=80]
10:13:53 E Error occurred [err="file not found"]
10:13:53 I User john logged in from 192.168.1.1
10:13:53 W Temperature is 80 degrees
10:13:53 E Failed to open file: permission denied
```

## Performance Metrics

Multilog can automatically collect and log performance metrics:

```go
logger.Perf("API request completed", 
    "endpoint", "/users", 
    "method", "GET", 
    "duration_ms", 42)
```

This will include metrics such as:
- Goroutine count
- Heap allocations
- System memory usage

Example performance log output:
```
10:04:05 PERF API request completed [goroutines:1,alloc:0.250191 MB,sys:6.334976 MB,heap_alloc:0.250191 MB,heap_sys:3.718750 MB,heap_idle:2.906250 MB,heap_inuse:0.812500 MB,stack_sys:0.281250 MB] [endpoint=/users method=GET duration_ms=42]
```

## Advanced Usage

### Context-Aware Logging

```go
ctx := context.Background()

// Add request ID to context
ctx = context.WithValue(ctx, "request_id", "abc-123")

logger.InfoContext(ctx, "Processing request for user %s", "john")
```

### Custom Attribute Replacement

```go
replaceAttr := func(groups []string, a slog.Attr) slog.Attr {
    if a.Key == "password" {
        return slog.String("password", "****")
    }
    return a
}

opts := multilog.CustomHandlerOptions{
    Level:   "info",
    Enabled: true,
}

handler := multilog.NewCustomHandler(&opts, writer, replaceAttr)
```

### Customizing Value Formatting

You can customize how values appear in logs:

```go
handler := multilog.NewConsoleHandler(multilog.CustomHandlerOptions{
    Level:           "debug",
    Enabled:         true,
    Pattern:         "[time] [level] [msg]",
    ValuePrefixChar: "<",
    ValueSuffixChar: ">",
})
```

Output:
```
<10:51:36> <INFO> <Info message>
<10:51:36> <DEBUG> <Debugging information> [user=john action=login]
```

## Implementation Details

### Caller Information Tracking

Multilog captures source location (file, line, function) even through wrapper functions by using `debug.Stack()` to parse the call stack and identify the actual logging call.

> **⚠️ Performance Note**: `debug.Stack()` has overhead. For high-volume production logging, disable source tracking with `AddSource: false`.

Standard Go logging approaches using `runtime.Caller()` with a fixed skip count would only report locations within the logging library itself, not the actual application code that initiated the log call.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgements

- [slog](https://pkg.go.dev/log/slog) - The Go standard library structured logging package
- [lumberjack](https://github.com/natefinch/lumberjack) - For log rotation
