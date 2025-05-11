package multilog

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"sync"
)

// JSONHandler is a Handler for JSON logging.
type JSONHandler struct {
	Handler CustomHandlerInterface
}

// Enabled checks if the handler is enabled for the given level.
func (jh *JSONHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return jh.Handler.Enabled(ctx, level)
}

// Handle processes the log record and writes it to the JSON handler.
func (jh *JSONHandler) Handle(ctx context.Context, record slog.Record) error {
	if !jh.Enabled(ctx, record.Level) {
		return nil
	}

	mu := sync.Mutex{}
	mu.Lock()
	defer mu.Unlock()

	sb := jh.Handler.GetStringBuilder()
	defer func() {
		sb.Reset()
	}()

	if err := jh.Handler.GetSlogHandler().Handle(ctx, record); err != nil {
		return fmt.Errorf("failed to handle record: %w", err)
	}

	opts := jh.Handler.GetOptions()
	patternPlaceHolders := opts.PatternPlaceholders
	if len(patternPlaceHolders) == 0 {
		patternPlaceHolders = DefaultPatternPlaceholders
	}
	values := GetPlaceholderValues(sb, record, patternPlaceHolders, jh.GetKeyValue)

	if record.Level == LevelPerf && !ContainsKey(opts.PatternPlaceholders, PerfPlaceholder) {
		values[PerfPlaceholder] = GetPerformanceMetrics()
	}

	values = RemovePlaceholderChars(values)

	err := json.Unmarshal([]byte(sb.String()), &values)
	if err != nil {
		return fmt.Errorf("failed to unmarshal values: %w", err)
	}

	b, err := json.Marshal(values)
	if err != nil {
		return fmt.Errorf("failed to marshal values: %w", err)
	}

	output := fmt.Sprintf("%s\n", string(b))

	// Get the writer from the handler
	writer := jh.Handler.GetWriter()
	if writer != nil {
		// If we have a bufio.Writer, use it
		if _, err := writer.WriteString(output); err != nil {
			return fmt.Errorf("failed to write log message: %w", err)
		}

		if err := writer.Flush(); err != nil {
			return fmt.Errorf("failed to flush writer: %w", err)
		}
	} else {
		customWriter, hasCustomWrite := jh.Handler.(WriterHandler)
		if hasCustomWrite {
			if err := customWriter.CustomWrite(output); err != nil {
				return err
			}
		} else {
			return fmt.Errorf("no writer available")
		}
	}

	return nil
}

// GetKeyValue retrieves the value associated with the given key from the JSON string.
func (jh *JSONHandler) GetKeyValue(key string, sb *strings.Builder, removeKey bool) string {
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(sb.String()), &m); err != nil {
		return ""
	}

	result := ""

	if v, ok := m[key]; ok {
		if removeKey {
			delete(m, key)
		}

		result = fmt.Sprintf("%v", v)

		if b, err := json.Marshal(m); err == nil {
			sb.Reset()
			sb.WriteString(string(b))
		}
	}

	return result
}

// WithAttrs creates a new handler with the given attributes.
func (jh *JSONHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return jh.Handler.WithAttrs(attrs)
}

// WithGroup creates a new handler with the given group name.
func (jh *JSONHandler) WithGroup(name string) slog.Handler {
	return jh.Handler.WithGroup(name)
}

// WriterHandler is an interface for custom write operations.
type WriterHandler interface {
	CustomWrite(output string) error
}
