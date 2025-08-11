package logging

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestTextFormatter(t *testing.T) {
	formatter := &TextFormatter{DisableTimestamp: true, DisableColors: true}

	entry := &LogEntry{
		Timestamp: time.Now(),
		Level:     "INFO",
		Message:   "test message",
		Fields: map[string]interface{}{
			"key1": "value1",
			"key2": 42,
		},
	}

	output, err := formatter.Format(entry)
	if err != nil {
		t.Fatalf("Failed to format entry: %v", err)
	}

	outputStr := string(output)
	t.Logf("Output: %q", outputStr)

	if !strings.Contains(outputStr, "[INFO]") {
		t.Errorf("Output should contain [INFO], got: %q", outputStr)
	}
	if !strings.Contains(outputStr, "test message") {
		t.Error("Output should contain message")
	}
	if !strings.Contains(outputStr, "key1=value1") {
		t.Error("Output should contain fields")
	}
}

func TestJSONFormatter(t *testing.T) {
	formatter := &JSONFormatter{}

	entry := &LogEntry{
		Timestamp: time.Now(),
		Level:     "INFO",
		Message:   "test message",
		Fields: map[string]interface{}{
			"key1": "value1",
			"key2": 42,
		},
	}

	output, err := formatter.Format(entry)
	if err != nil {
		t.Fatalf("Failed to format entry: %v", err)
	}

	// Parse JSON to verify structure
	var parsed map[string]interface{}
	if err := json.Unmarshal(output, &parsed); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	if parsed["level"] != "INFO" {
		t.Error("JSON should contain correct level")
	}
	if parsed["message"] != "test message" {
		t.Error("JSON should contain correct message")
	}
	if fields, ok := parsed["fields"].(map[string]interface{}); ok {
		if fields["key1"] != "value1" {
			t.Error("JSON should contain correct fields")
		}
	} else {
		t.Error("JSON should contain fields object")
	}
}

func TestLoggerLevels(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLoggerWithWriter(&buf)

	// Set level to WARN, should not log DEBUG and INFO
	logger.SetLevel(WARN)

	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warn message")
	logger.Error("error message")

	output := buf.String()

	if strings.Contains(output, "debug message") {
		t.Error("DEBUG message should be filtered out")
	}
	if strings.Contains(output, "info message") {
		t.Error("INFO message should be filtered out")
	}
	if !strings.Contains(output, "warn message") {
		t.Error("WARN message should be included")
	}
	if !strings.Contains(output, "error message") {
		t.Error("ERROR message should be included")
	}
}

func TestLoggerWithFields(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLoggerWithWriter(&buf)

	// Create logger with context fields
	contextLogger := logger.With(
		String("component", "validator"),
		Int("version", 1),
	)

	contextLogger.Info("test message", String("extra", "field"))

	output := buf.String()

	if !strings.Contains(output, "component=validator") {
		t.Error("Output should contain context field")
	}
	if !strings.Contains(output, "version=1") {
		t.Error("Output should contain context field")
	}
	if !strings.Contains(output, "extra=field") {
		t.Error("Output should contain additional field")
	}
}

func TestLoggerWithField(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLoggerWithWriter(&buf)

	logger.WithField("request_id", "12345").Info("processing request")

	output := buf.String()

	if !strings.Contains(output, "request_id=12345") {
		t.Error("Output should contain added field")
	}
}

func TestFormattedLogging(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLoggerWithWriter(&buf)

	logger.Infof("Processing %d records in %.2f seconds", 100, 1.23)

	output := buf.String()

	if !strings.Contains(output, "Processing 100 records in 1.23 seconds") {
		t.Error("Output should contain formatted message")
	}
}

func TestJSONLogger(t *testing.T) {
	var buf bytes.Buffer
	logger := &standardLogger{
		writer:    &buf,
		formatter: &JSONFormatter{},
		level:     INFO,
		fields:    make(map[string]interface{}),
	}

	logger.Info("test message", String("key", "value"))

	var parsed map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	if parsed["message"] != "test message" {
		t.Error("JSON should contain correct message")
	}
}

func TestGlobalLogger(t *testing.T) {
	var buf bytes.Buffer
	originalLogger := GetGlobalLogger()

	// Set custom global logger
	SetGlobalLogger(NewLoggerWithWriter(&buf))

	// Test global functions
	Info("global message")
	Infof("formatted %s", "message")

	// Restore original logger
	SetGlobalLogger(originalLogger)

	output := buf.String()

	if !strings.Contains(output, "global message") {
		t.Error("Global logging should work")
	}
	if !strings.Contains(output, "formatted message") {
		t.Error("Global formatted logging should work")
	}
}

func TestLogLevelString(t *testing.T) {
	tests := []struct {
		level    LogLevel
		expected string
	}{
		{DEBUG, "DEBUG"},
		{INFO, "INFO"},
		{WARN, "WARN"},
		{ERROR, "ERROR"},
		{LogLevel(999), "UNKNOWN"},
	}

	for _, test := range tests {
		if test.level.String() != test.expected {
			t.Errorf("Expected %s, got %s", test.expected, test.level.String())
		}
	}
}

func TestFieldHelpers(t *testing.T) {
	tests := []struct {
		name     string
		field    Field
		expected interface{}
	}{
		{"String", String("key", "value"), "value"},
		{"Int", Int("key", 42), 42},
		{"Int64", Int64("key", int64(42)), int64(42)},
		{"Float64", Float64("key", 3.14), 3.14},
		{"Bool", Bool("key", true), true},
		{"Duration", Duration("key", time.Second), time.Second},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.field.Key != "key" {
				t.Error("Field key should be 'key'")
			}
			if test.field.Value != test.expected {
				t.Errorf("Expected %v, got %v", test.expected, test.field.Value)
			}
		})
	}
}

func TestErrorField(t *testing.T) {
	// Test with nil error
	field := ErrorField("error", nil)
	if field.Value != nil {
		t.Error("Error field with nil error should have nil value")
	}

	// Test with actual error
	err := errors.New("test error")
	field = ErrorField("error", err)
	if field.Value != "test error" {
		t.Error("Error field should contain error string")
	}
}
