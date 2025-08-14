package core

import (
	"io"
	"log"
	"strconv"
	"strings"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/parser"
	"github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

// DateFormatValidator validates GTFS date format (YYYYMMDD)
type DateFormatValidator struct{}

// NewDateFormatValidator creates a new date format validator
func NewDateFormatValidator() *DateFormatValidator {
	return &DateFormatValidator{}
}

// dateFields defines which fields contain date values in each file
var dateFields = map[string][]string{
	"calendar.txt":       {"start_date", "end_date"},
	"calendar_dates.txt": {"date"},
	"feed_info.txt":      {"feed_start_date", "feed_end_date"},
}

// Validate checks date format in GTFS files
func (v *DateFormatValidator) Validate(loader *parser.FeedLoader, container *notice.NoticeContainer, config validator.Config) {
	for filename, fields := range dateFields {
		v.validateFileDateFields(loader, container, filename, fields)
	}
}

// validateFileDateFields validates date fields in a specific file
func (v *DateFormatValidator) validateFileDateFields(loader *parser.FeedLoader, container *notice.NoticeContainer, filename string, dateFieldNames []string) {
	reader, err := loader.GetFile(filename)
	if err != nil {
		return // File doesn't exist, skip validation
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close reader %v", closeErr)
		}
	}()

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

		for _, fieldName := range dateFieldNames {
			if value, exists := row.Values[fieldName]; exists && strings.TrimSpace(value) != "" {
				v.validateDateFormat(container, filename, fieldName, strings.TrimSpace(value), row.RowNumber)
			}
		}
	}
}

// validateDateFormat validates a single date value
func (v *DateFormatValidator) validateDateFormat(container *notice.NoticeContainer, filename string, fieldName string, dateValue string, rowNumber int) {
	if !v.isValidGTFSDate(dateValue) {
		container.AddNotice(notice.NewInvalidDateFormatNotice(
			filename,
			fieldName,
			dateValue,
			rowNumber,
		))
	}
}

// isValidGTFSDate checks if a date string is in valid GTFS format (YYYYMMDD)
func (v *DateFormatValidator) isValidGTFSDate(dateStr string) bool {
	// Must be exactly 8 characters
	if len(dateStr) != 8 {
		return false
	}

	// Must be all digits
	for _, char := range dateStr {
		if char < '0' || char > '9' {
			return false
		}
	}

	// Parse year, month, day
	year, err := strconv.Atoi(dateStr[:4])
	if err != nil || year < 1900 || year > 2200 {
		return false
	}

	month, err := strconv.Atoi(dateStr[4:6])
	if err != nil || month < 1 || month > 12 {
		return false
	}

	day, err := strconv.Atoi(dateStr[6:8])
	if err != nil || day < 1 || day > 31 {
		return false
	}

	// Basic month-day validation
	if month == 2 && day > 29 {
		return false // February max 29 days
	}
	if (month == 4 || month == 6 || month == 9 || month == 11) && day > 30 {
		return false // April, June, September, November max 30 days
	}

	return true
}
