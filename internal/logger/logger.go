// Package logger provides structured logging with different log levels
package logger

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// LogLevel represents the severity level of a log message
type LogLevel int

const (
	// DebugLevel is for detailed diagnostic information
	DebugLevel LogLevel = iota
	// InfoLevel is for general information about application progress
	InfoLevel
	// WarnLevel is for potentially harmful situations
	WarnLevel
	// ErrorLevel is for error events that might still allow the application to continue
	ErrorLevel
	// FatalLevel is for very severe error events that will presumably lead the application to abort
	FatalLevel
)

// String returns the string representation of a LogLevel
func (l LogLevel) String() string {
	switch l {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	case FatalLevel:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// Logger represents a structured logger
type Logger struct {
	level  LogLevel
	output io.Writer
	mu     sync.Mutex
}

// New creates a new logger with the specified minimum log level
func New(level LogLevel) *Logger {
	return &Logger{
		level:  level,
		output: os.Stdout,
	}
}

// NewWithOutput creates a new logger with the specified minimum log level and output writer
func NewWithOutput(level LogLevel, output io.Writer) *Logger {
	return &Logger{
		level:  level,
		output: output,
	}
}

// SetLevel sets the minimum log level for the logger
func (l *Logger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// Debug logs a message at debug level
func (l *Logger) Debug(format string, args ...interface{}) {
	if l.level <= DebugLevel {
		l.log(DebugLevel, format, args...)
	}
}

// Info logs a message at info level
func (l *Logger) Info(format string, args ...interface{}) {
	if l.level <= InfoLevel {
		l.log(InfoLevel, format, args...)
	}
}

// Warn logs a message at warn level
func (l *Logger) Warn(format string, args ...interface{}) {
	if l.level <= WarnLevel {
		l.log(WarnLevel, format, args...)
	}
}

// Error logs a message at error level
func (l *Logger) Error(format string, args ...interface{}) {
	if l.level <= ErrorLevel {
		l.log(ErrorLevel, format, args...)
	}
}

// Fatal logs a message at fatal level and exits the program
func (l *Logger) Fatal(format string, args ...interface{}) {
	if l.level <= FatalLevel {
		l.log(FatalLevel, format, args...)
		os.Exit(1)
	}
}

// log writes a formatted log message to the output
func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	timestamp := time.Now().Format("2006-01-02T15:04:05.000Z07:00")
	message := fmt.Sprintf(format, args...)
	logLine := fmt.Sprintf("[%s] %s: %s\n", timestamp, level.String(), message)

	l.output.Write([]byte(logLine))
}