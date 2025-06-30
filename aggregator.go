package multilog

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"runtime/metrics"
	"strconv"
	"strings"
)

// Aggregator is a handler that forwards logs to multiple handlers.
type Aggregator []slog.Handler

// NewAggregator creates a new aggregator with the given handlers.
func NewAggregator(handlers ...slog.Handler) Aggregator {
	return handlers
}

// Enabled implements slog.Handler.
func (a Aggregator) Enabled(ctx context.Context, level slog.Level) bool {
	for _, h := range a {
		if h.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

// Handle implements slog.Handler.
func (a Aggregator) Handle(ctx context.Context, r slog.Record) error {
	var firstErr error
	for _, h := range a {
		if h.Enabled(ctx, r.Level) {
			if err := h.Handle(ctx, r); err != nil && firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}

// WithAttrs implements slog.Handler.
func (a Aggregator) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make([]slog.Handler, len(a))
	for i, h := range a {
		handlers[i] = h.WithAttrs(attrs)
	}
	return NewAggregator(handlers...)
}

// WithGroup implements slog.Handler.
func (a Aggregator) WithGroup(name string) slog.Handler {
	handlers := make([]slog.Handler, len(a))
	for i, h := range a {
		handlers[i] = h.WithGroup(name)
	}
	return NewAggregator(handlers...)
}

// CollectPerfMetrics collects performance metrics.
func CollectPerfMetrics() *PerfMetrics {
	importantMetrics := []string{
		"/gc/heap/allocs:bytes",
		"/gc/heap/frees:bytes",
		"/memory/classes/heap/free:bytes",
		"/memory/classes/heap/objects:bytes",
		"/memory/classes/heap/released:bytes",
		"/memory/classes/heap/unused:bytes",
		"/memory/classes/total:bytes",
		"/sched/goroutines:goroutines",
	}

	descs := metrics.All()
	samples := make([]metrics.Sample, 0, len(importantMetrics))
	for _, desc := range descs {
		for _, impMetric := range importantMetrics {
			if desc.Name == impMetric {
				samples = append(samples, metrics.Sample{Name: desc.Name})
			}
		}
	}
	metrics.Read(samples)
	metricMap := make(map[string]metrics.Sample)
	for _, sample := range samples {
		metricMap[sample.Name] = sample
	}
	cpus := runtime.NumCPU()
	maxThreads := runtime.GOMAXPROCS(0)

	return &PerfMetrics{
		NumGoroutines: runtime.NumGoroutine(),
		NumCPUs:       cpus,
		MaxThreads:    maxThreads,
		GCHeapAllocs:  metricMap["/gc/heap/allocs:bytes"].Value.Uint64() / 1024,
		GCHeapFrees:   metricMap["/gc/heap/frees:bytes"].Value.Uint64() / 1024,
		HeapFree:      metricMap["/memory/classes/heap/free:bytes"].Value.Uint64() / 1024,
		HeapObjects:   metricMap["/memory/classes/heap/objects:bytes"].Value.Uint64() / 1024,
		HeapReleased:  metricMap["/memory/classes/heap/released:bytes"].Value.Uint64() / 1024,
		HeapUnused:    metricMap["/memory/classes/heap/unused:bytes"].Value.Uint64() / 1024,
		TotalMemory:   metricMap["/memory/classes/total:bytes"].Value.Uint64() / 1024,
	}
}

// CollectPerfMetricsWithMemStats collects performance metrics including memory stats.
func CollectPerfMetricsWithMemStats() *PerfMetrics {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	toMB := func(bytes uint64) float64 {
		return float64(bytes) / 1024 / 1024
	}

	return &PerfMetrics{
		NumGoroutines: runtime.NumGoroutine(),
		NumCPUs:       int(m.NumGC),
		MaxThreads:    runtime.GOMAXPROCS(0),
		Alloc:         toMB(m.Alloc),
		TotalAlloc:    toMB(m.TotalAlloc),
		Sys:           toMB(m.Sys),
		HeapAlloc:     toMB(m.HeapAlloc),
		HeapSys:       toMB(m.HeapSys),
		HeapIdle:      toMB(m.HeapIdle),
		HeapInuse:     toMB(m.HeapInuse),
		StackSys:      toMB(m.StackSys),
	}
}

// GetPerformanceMetrics gathers memory and goroutine metrics.
func GetPerformanceMetrics() string {
	return GetPerformanceMetricsWithMemStats()
}

// GetPerformanceMetricsUsingRuntime gathers memory and goroutine metrics.
func GetPerformanceMetricsUsingRuntime() string {
	pm := CollectPerfMetrics()
	return fmt.Sprintf(
		"goroutines:%d,heap_objects:%d,total:%d KB,num_cpu:%d",
		pm.NumGoroutines,
		pm.HeapObjects,
		pm.TotalMemory,
		pm.NumCPUs,
	)
}

// GetPerformanceMetricsWithMemStats gathers memory and goroutine metrics.
func GetPerformanceMetricsWithMemStats() string {
	pm := CollectPerfMetricsWithMemStats()
	return fmt.Sprintf(
		"goroutines:%d,alloc:%f MB,sys:%f MB,heap_alloc:%f MB,heap_sys:%f MB,heap_idle:%f MB,heap_inuse:%f MB,stack_sys:%f MB",
		pm.NumGoroutines,
		pm.Alloc,
		pm.Sys,
		pm.HeapAlloc,
		pm.HeapSys,
		pm.HeapIdle,
		pm.HeapInuse,
		pm.StackSys,
	)
}

// GetCallerInfo retrieves the caller information from the stack trace.
func GetCallerInfo(identifiers ...string) (fn, file string, line int, found bool) {
	stack := debug.Stack()
	lines := strings.Split(string(stack), "\n")

	for i, line := range lines {
		for _, identifier := range identifiers {
			fullIdentifier := PackagePrefix + identifier
			if strings.Contains(line, fullIdentifier) {
				fn, file, lineNum := getCallerDetails(lines[i+2], lines[i+3])
				return fn, file, lineNum, true
			}
		}
	}
	return "unknown", "", 0, false
}

// GetPerfCallerInfo returns the caller information for performance logs.
func GetPerfCallerInfo() (fn, file string, line int, found bool) {
	return GetCallerInfo(CallIdentifiers[0], CallIdentifiers[1])
}

// GetOtherCallerInfo returns the caller information for other logs.
func GetOtherCallerInfo() (fn, file string, line int, found bool) {
	return GetCallerInfo(CallIdentifiers[2:]...)
}

// getCallerDetails returns the caller information.
func getCallerDetails(fnLine, fileLine string) (fn, file string, line int) {
	parts := strings.Split(fileLine, " ")
	fileline := parts[0]
	parts = strings.Split(fileline, ":")
	file = parts[0]
	file = BaseName(file)
	lineNum, _ := strconv.Atoi(parts[1])
	line = lineNum

	fnParts := strings.Split(fnLine, "(")
	if len(fnParts) > 1 {
		fn = strings.Join(fnParts[:len(fnParts)-1], "(")
	} else {
		fn = fnLine
	}

	return fn, file, line
}

// GetOtherSourceValue returns the source value for other logs.
func GetOtherSourceValue(fn, file string, line int) string {
	result := DefaultPerfSourceFormat
	for _, source := range []string{FileSource, LineSource, FuncSource} {
		switch source {
		case FileSource:
			result = strings.ReplaceAll(result, FileSource, BaseName(file))
		case LineSource:
			result = strings.ReplaceAll(result, LineSource, fmt.Sprintf("%d", line))
		case FuncSource:
			result = strings.ReplaceAll(result, FuncSource, fn)
		}
	}
	return result
}

// Contains checks if a string slice contains a specific item.
func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// BaseName returns the base name of a file.
func BaseName(file string) string {
	return filepath.Base(file)
}

// PerfMetrics represents performance metrics.
type PerfMetrics struct {
	NumGoroutines int
	NumCPUs       int
	MaxThreads    int
	GCHeapAllocs  uint64
	GCHeapFrees   uint64
	HeapFree      uint64
	HeapObjects   uint64
	HeapReleased  uint64
	HeapUnused    uint64
	TotalMemory   uint64
	TotalCPUUsage float64
	UserCPUUsage  float64
	Alloc         float64
	TotalAlloc    float64
	Sys           float64
	HeapAlloc     float64
	HeapSys       float64
	HeapIdle      float64
	HeapInuse     float64
	StackSys      float64
}

// SourceInfo represents the source information of a log record.
type SourceInfo struct {
	File string
	Line int
}

// LogRecord represents a log record with additional information.
type LogRecord struct {
	Time        string
	Message     string
	Format      LogRecordFormat
	Source      SourceInfo
	PerfMetrics PerfMetrics
	Level       slog.Level
}

// LogRecordFormat represents the format of a log record.
type LogRecordFormat struct {
	Pattern string
}
