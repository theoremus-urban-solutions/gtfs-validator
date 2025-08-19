package core

import (
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/testutil"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

func TestDateFormatValidator_Validate(t *testing.T) {
	tests := []struct {
		name                string
		files               map[string]string
		expectedNoticeCodes []string
		description         string
	}{
		{
			name: "all date formats valid",
			files: map[string]string{
				CalendarFile:         "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\nS1,1,1,1,1,1,0,0,20250101,20251231",
				"calendar_dates.txt": "service_id,date,exception_type\nS1,20250101,1",
				"feed_info.txt":      "feed_publisher_name,feed_publisher_url,feed_lang,feed_start_date,feed_end_date\nMetro,http://metro.example,en,20250101,20251231",
			},
			expectedNoticeCodes: []string{},
			description:         "All date fields have valid YYYYMMDD format",
		},
		{
			name: "CalendarFile invalid start date",
			files: map[string]string{
				CalendarFile: "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\nS1,1,1,1,1,1,0,0,2025-01-01,20251231", // Hyphenated date
			},
			expectedNoticeCodes: []string{"invalid_date_format"},
			description:         "start_date with hyphens is invalid",
		},
		{
			name: "CalendarFile invalid end date",
			files: map[string]string{
				CalendarFile: "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\nS1,1,1,1,1,1,0,0,20250101,20251301", // Month 13
			},
			expectedNoticeCodes: []string{"invalid_date_format"},
			description:         "end_date with invalid month",
		},
		{
			name: "calendar_dates.txt invalid date",
			files: map[string]string{
				"calendar_dates.txt": "service_id,date,exception_type\nS1,20250230,1", // Feb 30th doesn't exist
			},
			expectedNoticeCodes: []string{"invalid_date_format"},
			description:         "February 30th is invalid",
		},
		{
			name: "feed_info.txt invalid dates",
			files: map[string]string{
				"feed_info.txt": "feed_publisher_name,feed_publisher_url,feed_lang,feed_start_date,feed_end_date\nMetro,http://metro.example,en,2025101,202512311", // Wrong length
			},
			expectedNoticeCodes: []string{"invalid_date_format", "invalid_date_format"},
			description:         "feed dates with wrong length",
		},
		{
			name: "multiple date format errors",
			files: map[string]string{
				CalendarFile: "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\nS1,1,1,1,1,1,0,0,2025-01-01,20251301", // Both dates invalid
			},
			expectedNoticeCodes: []string{"invalid_date_format", "invalid_date_format"},
			description:         "Multiple date format errors in single row",
		},
		{
			name: "valid leap year date",
			files: map[string]string{
				CalendarFile: "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\nS1,1,1,1,1,1,0,0,20240229,20241231", // Feb 29 in leap year
			},
			expectedNoticeCodes: []string{},
			description:         "February 29th should be valid (basic validation allows up to 29)",
		},
		{
			name: "invalid day values",
			files: map[string]string{
				CalendarFile: "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\nS1,1,1,1,1,1,0,0,20250132,20251231", // Day 32
			},
			expectedNoticeCodes: []string{"invalid_date_format"},
			description:         "Day 32 is invalid",
		},
		{
			name: "invalid month values",
			files: map[string]string{
				CalendarFile: "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\nS1,1,1,1,1,1,0,0,20250001,20251231", // Month 00
			},
			expectedNoticeCodes: []string{"invalid_date_format"},
			description:         "Month 00 is invalid",
		},
		{
			name: "invalid year values",
			files: map[string]string{
				CalendarFile: "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\nS1,1,1,1,1,1,0,0,18990101,20251231", // Year 1899
			},
			expectedNoticeCodes: []string{"invalid_date_format"},
			description:         "Year 1899 is too early",
		},
		{
			name: "dates with non-numeric characters",
			files: map[string]string{
				CalendarFile: "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\nS1,1,1,1,1,1,0,0,202A0101,20251231", // Non-numeric character
			},
			expectedNoticeCodes: []string{"invalid_date_format"},
			description:         "Dates must be all numeric",
		},
		{
			name: "dates too short",
			files: map[string]string{
				CalendarFile: "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\nS1,1,1,1,1,1,0,0,2025101,20251231", // 7 characters instead of 8
			},
			expectedNoticeCodes: []string{"invalid_date_format"},
			description:         "Dates must be exactly 8 characters",
		},
		{
			name: "dates too long",
			files: map[string]string{
				CalendarFile: "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\nS1,1,1,1,1,1,0,0,202501011,20251231", // 9 characters
			},
			expectedNoticeCodes: []string{"invalid_date_format"},
			description:         "Dates must be exactly 8 characters, not longer",
		},
		{
			name: "valid dates in 30-day months",
			files: map[string]string{
				CalendarFile: "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\nS1,1,1,1,1,1,0,0,20250430,20250630", // April 30, June 30
			},
			expectedNoticeCodes: []string{},
			description:         "April 30 and June 30 are valid",
		},
		{
			name: "invalid dates in 30-day months",
			files: map[string]string{
				CalendarFile: "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\nS1,1,1,1,1,1,0,0,20250431,20251131", // April 31, November 31
			},
			expectedNoticeCodes: []string{"invalid_date_format", "invalid_date_format"},
			description:         "April 31 and November 31 are invalid",
		},
		{
			name: "empty date fields ignored",
			files: map[string]string{
				CalendarFile: "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\nS1,1,1,1,1,1,0,0,,20251231", // Empty start_date
			},
			expectedNoticeCodes: []string{},
			description:         "Empty date fields should not generate format errors",
		},
		{
			name: "whitespace-only date fields ignored",
			files: map[string]string{
				CalendarFile: "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\nS1,1,1,1,1,1,0,0,   ,20251231", // Whitespace start_date
			},
			expectedNoticeCodes: []string{},
			description:         "Whitespace-only date fields should not generate format errors",
		},
		{
			name: "valid boundary years",
			files: map[string]string{
				CalendarFile: "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\nS1,1,1,1,1,1,0,0,19000101,22001231", // Year 1900 and 2200
			},
			expectedNoticeCodes: []string{},
			description:         "Years 1900 and 2200 should be valid boundaries",
		},
		{
			name: "invalid boundary years",
			files: map[string]string{
				CalendarFile: "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\nS1,1,1,1,1,1,0,0,18991231,22010101", // Year 1899 and 2201
			},
			expectedNoticeCodes: []string{"invalid_date_format", "invalid_date_format"},
			description:         "Years 1899 and 2201 should be invalid",
		},
		{
			name: "multiple rows with date errors",
			files: map[string]string{
				"calendar_dates.txt": "service_id,date,exception_type\nS1,2025-01-01,1\nS2,20250230,1", // Different format errors
			},
			expectedNoticeCodes: []string{"invalid_date_format", "invalid_date_format"},
			description:         "Date format errors across multiple rows",
		},
		{
			name: "files without date fields",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,http://metro.example,America/Los_Angeles",
				"stops.txt":  "stop_id,stop_name,stop_lat,stop_lon\n1,Main St,34.05,-118.25",
			},
			expectedNoticeCodes: []string{},
			description:         "Files without date fields should not generate date format errors",
		},
		{
			name: "missing date files",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,http://metro.example,America/Los_Angeles",
			},
			expectedNoticeCodes: []string{},
			description:         "Missing files with date fields should not cause errors",
		},
		{
			name: "date field with whitespace padding",
			files: map[string]string{
				CalendarFile: "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\nS1,1,1,1,1,1,0,0, 20250101 ,20251231", // Whitespace around date
			},
			expectedNoticeCodes: []string{},
			description:         "Whitespace around date values should be trimmed and validated",
		},
		{
			name: "valid February dates",
			files: map[string]string{
				CalendarFile: "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\nS1,1,1,1,1,1,0,0,20250228,20250229", // Feb 28 and 29
			},
			expectedNoticeCodes: []string{},
			description:         "February 28 and 29 should be valid",
		},
		{
			name: "invalid February dates",
			files: map[string]string{
				CalendarFile: "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\nS1,1,1,1,1,1,0,0,20250230,20250231", // Feb 30 and 31
			},
			expectedNoticeCodes: []string{"invalid_date_format", "invalid_date_format"},
			description:         "February 30 and 31 should be invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test components
			loader := testutil.CreateTestFeedLoader(t, tt.files)
			container := notice.NewNoticeContainer()
			validator := NewDateFormatValidator()
			config := gtfsvalidator.Config{}

			// Run validation
			validator.Validate(loader, container, config)

			// Get notices
			notices := container.GetNotices()

			// Check notice count
			if len(notices) != len(tt.expectedNoticeCodes) {
				t.Errorf("Expected %d notices, got %d for case: %s", len(tt.expectedNoticeCodes), len(notices), tt.description)
			}

			// Count notice codes
			expectedCodeCounts := make(map[string]int)
			for _, code := range tt.expectedNoticeCodes {
				expectedCodeCounts[code]++
			}

			actualCodeCounts := make(map[string]int)
			for _, notice := range notices {
				actualCodeCounts[notice.Code()]++
			}

			// Verify expected codes
			for expectedCode, expectedCount := range expectedCodeCounts {
				actualCount := actualCodeCounts[expectedCode]
				if actualCount != expectedCount {
					t.Errorf("Expected %d notices with code '%s', got %d", expectedCount, expectedCode, actualCount)
				}
			}

			// Check for unexpected notice codes
			for actualCode := range actualCodeCounts {
				if expectedCodeCounts[actualCode] == 0 {
					t.Errorf("Unexpected notice code: %s", actualCode)
				}
			}
		})
	}
}

func TestDateFormatValidator_IsValidGTFSDate(t *testing.T) {
	validator := NewDateFormatValidator()

	tests := []struct {
		dateStr     string
		expected    bool
		description string
	}{
		{"20250101", true, "Standard valid date"},
		{"20251231", true, "End of year"},
		{"20240229", true, "Leap year Feb 29 (basic validation allows)"},
		{"19000101", true, "Minimum year boundary"},
		{"22001231", true, "Maximum year boundary"},

		{"2025-01-01", false, "Date with hyphens"},
		{"2025/01/01", false, "Date with slashes"},
		{"01/01/2025", false, "US format"},
		{"01-01-2025", false, "US format with hyphens"},

		{"2025101", false, "Too short (7 chars)"},
		{"202501011", false, "Too long (9 chars)"},
		{"", false, "Empty string"},

		{"202A0101", false, "Non-numeric character"},
		{"20250A01", false, "Non-numeric month"},
		{"202501A1", false, "Non-numeric day"},

		{"18991231", false, "Year too early"},
		{"22010101", false, "Year too late"},

		{"20250001", false, "Month 00"},
		{"20251301", false, "Month 13"},

		{"20250100", false, "Day 00"},
		{"20250132", false, "Day 32"},

		{"20250230", false, "Feb 30"},
		{"20250231", false, "Feb 31"},
		{"20250431", false, "April 31"},
		{"20250631", false, "June 31"},
		{"20250931", false, "September 31"},
		{"20251131", false, "November 31"},

		{"20250131", true, "Jan 31 (valid)"},
		{"20250331", true, "March 31 (valid)"},
		{"20250531", true, "May 31 (valid)"},
		{"20250731", true, "July 31 (valid)"},
		{"20250831", true, "August 31 (valid)"},
		{"20251031", true, "October 31 (valid)"},
		{"20251231", true, "December 31 (valid)"},

		{"20250430", true, "April 30 (valid)"},
		{"20250630", true, "June 30 (valid)"},
		{"20250930", true, "September 30 (valid)"},
		{"20251130", true, "November 30 (valid)"},

		{"20250228", true, "Feb 28 (valid)"},
		{"20250229", true, "Feb 29 (valid in basic validation)"},
	}

	for _, tt := range tests {
		t.Run(tt.dateStr+"_"+tt.description, func(t *testing.T) {
			result := validator.isValidGTFSDate(tt.dateStr)
			if result != tt.expected {
				t.Errorf("isValidGTFSDate(%q) = %v, expected %v (%s)", tt.dateStr, result, tt.expected, tt.description)
			}
		})
	}
}

func TestDateFormatValidator_ValidateDateFormat(t *testing.T) {
	tests := []struct {
		name         string
		dateValue    string
		expectNotice bool
		description  string
	}{
		{
			name:         "valid date",
			dateValue:    "20250101",
			expectNotice: false,
			description:  "Valid date should not generate notice",
		},
		{
			name:         "invalid date format",
			dateValue:    "2025-01-01",
			expectNotice: true,
			description:  "Invalid format should generate notice",
		},
		{
			name:         "invalid month",
			dateValue:    "20251301",
			expectNotice: true,
			description:  "Invalid month should generate notice",
		},
		{
			name:         "invalid day",
			dateValue:    "20250230",
			expectNotice: true,
			description:  "Invalid day should generate notice",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			container := notice.NewNoticeContainer()
			validator := NewDateFormatValidator()

			validator.validateDateFormat(container, CalendarFile, "start_date", tt.dateValue, 1)

			notices := container.GetNotices()
			hasNotice := len(notices) > 0

			if hasNotice != tt.expectNotice {
				t.Errorf("Expected notice: %v, got notice: %v for %s", tt.expectNotice, hasNotice, tt.description)
			}

			if hasNotice && tt.expectNotice {
				// Verify notice details
				notice := notices[0]
				if notice.Code() != "invalid_date_format" {
					t.Errorf("Expected notice code 'invalid_date_format', got '%s'", notice.Code())
				}

				context := notice.Context()
				if filename, ok := context["filename"]; !ok || filename != CalendarFile {
					t.Errorf("Expected filename '%s' in context, got '%v'", CalendarFile, filename)
				}
				if fieldName, ok := context["fieldName"]; !ok || fieldName != "start_date" {
					t.Errorf("Expected fieldName 'start_date' in context, got '%v'", fieldName)
				}
				if dateValue, ok := context["dateValue"]; !ok || dateValue != tt.dateValue {
					t.Errorf("Expected dateValue '%s' in context, got '%v'", tt.dateValue, dateValue)
				}
				if rowNumber, ok := context["csvRowNumber"]; !ok || rowNumber != 1 {
					t.Errorf("Expected csvRowNumber 1 in context, got '%v'", rowNumber)
				}
			}
		})
	}
}

func TestDateFormatValidator_DateFields(t *testing.T) {
	// Test that dateFields map contains expected files and fields
	expectedDateFields := map[string][]string{
		CalendarFile:         {"start_date", "end_date"},
		"calendar_dates.txt": {"date"},
		"feed_info.txt":      {"feed_start_date", "feed_end_date"},
	}

	for filename, expectedFields := range expectedDateFields {
		actualFields, exists := dateFields[filename]
		if !exists {
			t.Errorf("Expected file '%s' to be in dateFields map", filename)
			continue
		}

		if len(actualFields) != len(expectedFields) {
			t.Errorf("File '%s': expected %d date fields, got %d", filename, len(expectedFields), len(actualFields))
			continue
		}

		// Create maps for easier comparison
		expectedMap := make(map[string]bool)
		for _, field := range expectedFields {
			expectedMap[field] = true
		}

		actualMap := make(map[string]bool)
		for _, field := range actualFields {
			actualMap[field] = true
		}

		// Check all expected fields are present
		for _, expectedField := range expectedFields {
			if !actualMap[expectedField] {
				t.Errorf("File '%s': expected date field '%s' not found", filename, expectedField)
			}
		}

		// Check for unexpected fields
		for _, actualField := range actualFields {
			if !expectedMap[actualField] {
				t.Errorf("File '%s': unexpected date field '%s'", filename, actualField)
			}
		}
	}
}

func TestDateFormatValidator_ValidateFileDateFields(t *testing.T) {
	tests := []struct {
		name            string
		filename        string
		content         string
		dateFields      []string
		expectedNotices int
		description     string
	}{
		{
			name:            "valid date fields",
			filename:        CalendarFile,
			content:         "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\nS1,1,1,1,1,1,0,0,20250101,20251231",
			dateFields:      []string{"start_date", "end_date"},
			expectedNotices: 0,
			description:     "Valid date formats should not generate notices",
		},
		{
			name:            "invalid date fields",
			filename:        CalendarFile,
			content:         "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\nS1,1,1,1,1,1,0,0,2025-01-01,20251301",
			dateFields:      []string{"start_date", "end_date"},
			expectedNotices: 2,
			description:     "Invalid date formats should generate notices",
		},
		{
			name:            "missing date fields in data",
			filename:        CalendarFile,
			content:         "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday\nS1,1,1,1,1,1,0,0", // Missing date fields
			dateFields:      []string{"start_date", "end_date"},
			expectedNotices: 0,
			description:     "Missing fields should not generate format errors",
		},
		{
			name:            "empty date fields",
			filename:        CalendarFile,
			content:         "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\nS1,1,1,1,1,1,0,0,,20251231",
			dateFields:      []string{"start_date", "end_date"},
			expectedNotices: 0,
			description:     "Empty date fields should be ignored",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files := map[string]string{tt.filename: tt.content}
			loader := testutil.CreateTestFeedLoader(t, files)
			container := notice.NewNoticeContainer()
			validator := NewDateFormatValidator()

			validator.validateFileDateFields(loader, container, tt.filename, tt.dateFields)

			notices := container.GetNotices()
			if len(notices) != tt.expectedNotices {
				t.Errorf("Expected %d notices, got %d for %s", tt.expectedNotices, len(notices), tt.description)
			}
		})
	}
}

func TestDateFormatValidator_New(t *testing.T) {
	validator := NewDateFormatValidator()
	if validator == nil {
		t.Error("NewDateFormatValidator() returned nil")
	}
}
