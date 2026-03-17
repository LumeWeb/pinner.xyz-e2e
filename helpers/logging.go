package helpers

import (
	"context"
	"fmt"
	"os"
)

// LogLevel represents the severity of a log message
type LogLevel int

const (
	// LevelInfo is for general operational information
	LevelInfo LogLevel = iota
	// LevelDebug is for detailed debugging information
	LevelDebug
	// LevelWarn is for warnings
	LevelWarn
	// LevelError is for errors
	LevelError
)

var debugEnabled = os.Getenv("DEBUG_LOGGING") == "true"

// Log provides structured logging for test scenarios
// It uses godog's context-aware logging when available, falling back to stdout
type Logger struct {
	scenarioName string
}

// NewLogger creates a new logger instance
func NewLogger(scenarioName string) *Logger {
	return &Logger{scenarioName: scenarioName}
}

// logInternal logs a message at the specified level
func (l *Logger) logInternal(ctx context.Context, level LogLevel, format string, args ...interface{}) {
	// Only log debug messages when explicitly enabled
	if level == LevelDebug && !debugEnabled {
		return
	}

	// Use fmt.Printf for both contexts to avoid godog.Logf format string issues
	// godog.Logf expects the format string to be a constant, but we construct it dynamically
	fmt.Printf(format+"\n", args...)
}

// Info logs informational messages about test operations
// Use for significant events like starting cleanup, completing operations
func (l *Logger) Info(ctx context.Context, format string, args ...interface{}) {
	newFormat := "[INFO] %s - " + format
	newArgs := append([]interface{}{l.scenarioName}, args...)
	l.logInternal(ctx, LevelInfo, newFormat, newArgs...)
}

// Debug logs detailed debugging information
// Only emitted when debugging is explicitly enabled
func (l *Logger) Debug(ctx context.Context, format string, args ...interface{}) {
	newFormat := "[DEBUG] %s - " + format
	newArgs := append([]interface{}{l.scenarioName}, args...)
	l.logInternal(ctx, LevelDebug, newFormat, newArgs...)
}

// Warn logs warning messages for non-critical issues
func (l *Logger) Warn(ctx context.Context, format string, args ...interface{}) {
	newFormat := "[WARN] %s - " + format
	newArgs := append([]interface{}{l.scenarioName}, args...)
	l.logInternal(ctx, LevelWarn, newFormat, newArgs...)
}

// Error logs error messages
func (l *Logger) Error(ctx context.Context, format string, args ...interface{}) {
	newFormat := "[ERROR] %s - " + format
	newArgs := append([]interface{}{l.scenarioName}, args...)
	l.logInternal(ctx, LevelError, newFormat, newArgs...)
}
