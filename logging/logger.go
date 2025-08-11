// Package logging provides structured logging capabilities for the GTFS validator
package logging

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

// LogLevel represents the severity level of a log message
type LogLevel int

const (
	// DEBUG level for detailed diagnostic information
	DEBUG LogLevel = iota
	// INFO level for general informational messages
	INFO
	// WARN level for warning messages that indicate potential issues
	WARN
	// ERROR level for error messages that indicate failures
	ERROR
)

// String returns the string representation of a log level
func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Logger interface defines the logging contract
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)

	// Formatted logging methods
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})

	// With methods for adding context
	With(fields ...Field) Logger
	WithField(key string, value interface{}) Logger

	// SetLevel sets the minimum log level
	SetLevel(level LogLevel)
}

// Field represents a key-value pair for structured logging
type Field struct {
	Key   string
	Value interface{}
}

// LogEntry represents a single log entry
type LogEntry struct {
	Timestamp time.Time              `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
}

// Formatter interface for different log output formats
type Formatter interface {
	Format(entry *LogEntry) ([]byte, error)
}

// JSONFormatter formats log entries as JSON
type JSONFormatter struct{}

// Format implements the Formatter interface for JSON output
func (f *JSONFormatter) Format(entry *LogEntry) ([]byte, error) {
	data, err := json.Marshal(entry)
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

// TextFormatter formats log entries as human-readable text
type TextFormatter struct {
	// DisableTimestamp disables timestamp in output
	DisableTimestamp bool
	// DisableColors disables colored output
	DisableColors bool
}

// Format implements the Formatter interface for text output
func (f *TextFormatter) Format(entry *LogEntry) ([]byte, error) {
	var buf strings.Builder

	// Add timestamp
	if !f.DisableTimestamp {
		buf.WriteString(entry.Timestamp.Format("2006-01-02 15:04:05"))
		buf.WriteString(" ")
	}

	// Add level with optional colors
	level := entry.Level
	if !f.DisableColors {
		switch entry.Level {
		case "DEBUG":
			level = "\033[36mDEBUG\033[0m" // Cyan
		case "INFO":
			level = "\033[32mINFO\033[0m" // Green
		case "WARN":
			level = "\033[33mWARN\033[0m" // Yellow
		case "ERROR":
			level = "\033[31mERROR\033[0m" // Red
		}
	}

	buf.WriteString("[" + level + "] ")
	buf.WriteString(entry.Message)

	// Add fields
	if len(entry.Fields) > 0 {
		buf.WriteString(" ")
		for k, v := range entry.Fields {
			buf.WriteString(fmt.Sprintf("%s=%v ", k, v))
		}
	}

	buf.WriteString("\n")
	return []byte(buf.String()), nil
}

// standardLogger is the default implementation of Logger
type standardLogger struct {
	mutex     sync.RWMutex
	writer    io.Writer
	formatter Formatter
	level     LogLevel
	fields    map[string]interface{}
}

// NewLogger creates a new logger with default configuration
func NewLogger() Logger {
	return &standardLogger{
		writer:    os.Stdout,
		formatter: &TextFormatter{},
		level:     INFO,
		fields:    make(map[string]interface{}),
	}
}

// NewJSONLogger creates a new logger with JSON formatting
func NewJSONLogger() Logger {
	return &standardLogger{
		writer:    os.Stdout,
		formatter: &JSONFormatter{},
		level:     INFO,
		fields:    make(map[string]interface{}),
	}
}

// NewLoggerWithWriter creates a logger with custom writer
func NewLoggerWithWriter(writer io.Writer) Logger {
	return &standardLogger{
		writer:    writer,
		formatter: &TextFormatter{},
		level:     INFO,
		fields:    make(map[string]interface{}),
	}
}

// Debug logs a debug message
func (l *standardLogger) Debug(msg string, fields ...Field) {
	l.log(DEBUG, msg, fields...)
}

// Info logs an info message
func (l *standardLogger) Info(msg string, fields ...Field) {
	l.log(INFO, msg, fields...)
}

// Warn logs a warning message
func (l *standardLogger) Warn(msg string, fields ...Field) {
	l.log(WARN, msg, fields...)
}

// Error logs an error message
func (l *standardLogger) Error(msg string, fields ...Field) {
	l.log(ERROR, msg, fields...)
}

// Debugf logs a formatted debug message
func (l *standardLogger) Debugf(format string, args ...interface{}) {
	l.log(DEBUG, fmt.Sprintf(format, args...))
}

// Infof logs a formatted info message
func (l *standardLogger) Infof(format string, args ...interface{}) {
	l.log(INFO, fmt.Sprintf(format, args...))
}

// Warnf logs a formatted warning message
func (l *standardLogger) Warnf(format string, args ...interface{}) {
	l.log(WARN, fmt.Sprintf(format, args...))
}

// Errorf logs a formatted error message
func (l *standardLogger) Errorf(format string, args ...interface{}) {
	l.log(ERROR, fmt.Sprintf(format, args...))
}

// With returns a new logger with additional fields
func (l *standardLogger) With(fields ...Field) Logger {
	l.mutex.RLock()
	newFields := make(map[string]interface{})
	for k, v := range l.fields {
		newFields[k] = v
	}
	l.mutex.RUnlock()

	for _, field := range fields {
		newFields[field.Key] = field.Value
	}

	return &standardLogger{
		writer:    l.writer,
		formatter: l.formatter,
		level:     l.level,
		fields:    newFields,
	}
}

// WithField returns a new logger with an additional field
func (l *standardLogger) WithField(key string, value interface{}) Logger {
	return l.With(Field{Key: key, Value: value})
}

// SetLevel sets the minimum log level
func (l *standardLogger) SetLevel(level LogLevel) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.level = level
}

// log writes a log entry
func (l *standardLogger) log(level LogLevel, msg string, fields ...Field) {
	l.mutex.RLock()
	if level < l.level {
		l.mutex.RUnlock()
		return
	}

	// Combine existing fields with new fields
	allFields := make(map[string]interface{})
	for k, v := range l.fields {
		allFields[k] = v
	}
	l.mutex.RUnlock()

	for _, field := range fields {
		allFields[field.Key] = field.Value
	}

	entry := &LogEntry{
		Timestamp: time.Now(),
		Level:     level.String(),
		Message:   msg,
		Fields:    allFields,
	}

	l.mutex.RLock()
	data, err := l.formatter.Format(entry)
	l.mutex.RUnlock()

	if err != nil {
		// Fallback to standard library logger
		log.Printf("Logger formatting error: %v, original message: %s", err, msg)
		return
	}

	l.mutex.RLock()
	_, err = l.writer.Write(data)
	l.mutex.RUnlock()

	if err != nil {
		// Fallback to standard library logger
		log.Printf("Logger write error: %v, original message: %s", err, msg)
	}
}

// Global logger instance
var globalLogger Logger = NewLogger()

// SetGlobalLogger sets the global logger instance
func SetGlobalLogger(logger Logger) {
	globalLogger = logger
}

// GetGlobalLogger returns the global logger instance
func GetGlobalLogger() Logger {
	return globalLogger
}

// Global convenience functions

// Debug logs a debug message using the global logger
func Debug(msg string, fields ...Field) {
	globalLogger.Debug(msg, fields...)
}

// Info logs an info message using the global logger
func Info(msg string, fields ...Field) {
	globalLogger.Info(msg, fields...)
}

// Warn logs a warning message using the global logger
func Warn(msg string, fields ...Field) {
	globalLogger.Warn(msg, fields...)
}

// Error logs an error message using the global logger
func Error(msg string, fields ...Field) {
	globalLogger.Error(msg, fields...)
}

// Debugf logs a formatted debug message using the global logger
func Debugf(format string, args ...interface{}) {
	globalLogger.Debugf(format, args...)
}

// Infof logs a formatted info message using the global logger
func Infof(format string, args ...interface{}) {
	globalLogger.Infof(format, args...)
}

// Warnf logs a formatted warning message using the global logger
func Warnf(format string, args ...interface{}) {
	globalLogger.Warnf(format, args...)
}

// Errorf logs a formatted error message using the global logger
func Errorf(format string, args ...interface{}) {
	globalLogger.Errorf(format, args...)
}

// Helper functions for creating fields

// String creates a string field
func String(key, value string) Field {
	return Field{Key: key, Value: value}
}

// Int creates an int field
func Int(key string, value int) Field {
	return Field{Key: key, Value: value}
}

// Int64 creates an int64 field
func Int64(key string, value int64) Field {
	return Field{Key: key, Value: value}
}

// Float64 creates a float64 field
func Float64(key string, value float64) Field {
	return Field{Key: key, Value: value}
}

// Bool creates a bool field
func Bool(key string, value bool) Field {
	return Field{Key: key, Value: value}
}

// Duration creates a duration field
func Duration(key string, value time.Duration) Field {
	return Field{Key: key, Value: value}
}

// Time creates a time field
func Time(key string, value time.Time) Field {
	return Field{Key: key, Value: value}
}

// Error creates an error field
func ErrorField(key string, err error) Field {
	if err == nil {
		return Field{Key: key, Value: nil}
	}
	return Field{Key: key, Value: err.Error()}
}
