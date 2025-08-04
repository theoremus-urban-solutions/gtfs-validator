package core

import (
	"io"
	"strconv"
	"strings"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/parser"
	"github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

// TimeFormatValidator validates GTFS time format (HH:MM:SS)
type TimeFormatValidator struct{}

// NewTimeFormatValidator creates a new time format validator
func NewTimeFormatValidator() *TimeFormatValidator {
	return &TimeFormatValidator{}
}

// timeFields defines which fields contain time values in each file
var timeFields = map[string][]string{
	"stop_times.txt":  {"arrival_time", "departure_time"},
	"frequencies.txt": {"start_time", "end_time"},
}

// Validate checks time format in GTFS files
func (v *TimeFormatValidator) Validate(loader *parser.FeedLoader, container *notice.NoticeContainer, config validator.Config) {
	for filename, fields := range timeFields {
		v.validateFileTimeFields(loader, container, filename, fields)
	}
}

// validateFileTimeFields validates time fields in a specific file
func (v *TimeFormatValidator) validateFileTimeFields(loader *parser.FeedLoader, container *notice.NoticeContainer, filename string, timeFieldNames []string) {
	reader, err := loader.GetFile(filename)
	if err != nil {
		return // File doesn't exist, skip validation
	}
	defer reader.Close()

	csvFile, err := parser.NewCSVFile(reader, filename)
	if err != nil {
		return
	}

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}

		for _, fieldName := range timeFieldNames {
			if value, exists := row.Values[fieldName]; exists && strings.TrimSpace(value) != "" {
				v.validateTimeFormat(container, filename, fieldName, strings.TrimSpace(value), row.RowNumber)
			}
		}
	}
}

// validateTimeFormat validates a single time value
func (v *TimeFormatValidator) validateTimeFormat(container *notice.NoticeContainer, filename string, fieldName string, timeValue string, rowNumber int) {
	if !v.isValidGTFSTime(timeValue) {
		container.AddNotice(notice.NewInvalidTimeFormatNotice(
			filename,
			fieldName,
			timeValue,
			rowNumber,
		))
	}
}

// isValidGTFSTime checks if a time string is in valid GTFS format (HH:MM:SS)
func (v *TimeFormatValidator) isValidGTFSTime(timeStr string) bool {
	parts := strings.Split(timeStr, ":")
	if len(parts) != 3 {
		return false
	}

	// Parse hours (can be > 23 for next-day service)
	hours, err := strconv.Atoi(parts[0])
	if err != nil || hours < 0 {
		return false
	}

	// Parse minutes (0-59)
	minutes, err := strconv.Atoi(parts[1])
	if err != nil || minutes < 0 || minutes >= 60 {
		return false
	}

	// Parse seconds (0-59)
	seconds, err := strconv.Atoi(parts[2])
	if err != nil || seconds < 0 || seconds >= 60 {
		return false
	}

	// Check for proper zero-padding
	if len(parts[0]) != 2 || len(parts[1]) != 2 || len(parts[2]) != 2 {
		return false
	}

	return true
}
