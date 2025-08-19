package core

import (
	"github.com/theoremus-urban-solutions/gtfs-validator/testutil"
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

func TestTimeFormatValidator_Validate(t *testing.T) {
	tests := []struct {
		name                string
		files               map[string]string
		expectedNoticeCodes []string
		description         string
	}{
		{
			name: "all time formats valid",
			files: map[string]string{
				StopTimesFile:     "trip_id,arrival_time,departure_time,stop_id,stop_sequence\nT1,08:00:00,08:01:00,1,1",
				"frequencies.txt": "trip_id,start_time,end_time,headway_secs\nT1,06:00:00,22:00:00,600",
			},
			expectedNoticeCodes: []string{},
			description:         "All time fields have valid HH:MM:SS format",
		},
		{
			name: "StopTimesFile invalid arrival time",
			files: map[string]string{
				StopTimesFile: "trip_id,arrival_time,departure_time,stop_id,stop_sequence\nT1,8:00:00,08:01:00,1,1", // Not zero-padded
			},
			expectedNoticeCodes: []string{"invalid_time_format"},
			description:         "arrival_time missing zero-padding",
		},
		{
			name: "StopTimesFile invalid departure time",
			files: map[string]string{
				StopTimesFile: "trip_id,arrival_time,departure_time,stop_id,stop_sequence\nT1,08:00:00,08:1:00,1,1", // Minute not zero-padded
			},
			expectedNoticeCodes: []string{"invalid_time_format"},
			description:         "departure_time minute not zero-padded",
		},
		{
			name: "frequencies.txt invalid start time",
			files: map[string]string{
				"frequencies.txt": "trip_id,start_time,end_time,headway_secs\nT1,6:00:00,22:00:00,600", // Not zero-padded
			},
			expectedNoticeCodes: []string{"invalid_time_format"},
			description:         "start_time missing zero-padding",
		},
		{
			name: "frequencies.txt invalid end time",
			files: map[string]string{
				"frequencies.txt": "trip_id,start_time,end_time,headway_secs\nT1,06:00:00,22:00:0,600", // Second not zero-padded
			},
			expectedNoticeCodes: []string{"invalid_time_format"},
			description:         "end_time second not zero-padded",
		},
		{
			name: "multiple invalid time formats",
			files: map[string]string{
				StopTimesFile: "trip_id,arrival_time,departure_time,stop_id,stop_sequence\nT1,8:0:0,08:61:00,1,1", // Multiple format errors
			},
			expectedNoticeCodes: []string{"invalid_time_format", "invalid_time_format"},
			description:         "Multiple time format errors in single row",
		},
		{
			name: "valid next-day service times",
			files: map[string]string{
				StopTimesFile: "trip_id,arrival_time,departure_time,stop_id,stop_sequence\nT1,24:00:00,25:30:00,1,1", // Valid next-day times
			},
			expectedNoticeCodes: []string{},
			description:         "Times >= 24:00:00 are valid for next-day service",
		},
		{
			name: "invalid minute values",
			files: map[string]string{
				StopTimesFile: "trip_id,arrival_time,departure_time,stop_id,stop_sequence\nT1,08:60:00,08:00:00,1,1", // 60 minutes invalid
			},
			expectedNoticeCodes: []string{"invalid_time_format"},
			description:         "Minutes must be 0-59",
		},
		{
			name: "invalid second values",
			files: map[string]string{
				StopTimesFile: "trip_id,arrival_time,departure_time,stop_id,stop_sequence\nT1,08:00:60,08:00:00,1,1", // 60 seconds invalid
			},
			expectedNoticeCodes: []string{"invalid_time_format"},
			description:         "Seconds must be 0-59",
		},
		{
			name: "invalid time format - missing colons",
			files: map[string]string{
				StopTimesFile: "trip_id,arrival_time,departure_time,stop_id,stop_sequence\nT1,080000,08:00:00,1,1", // Missing colons
			},
			expectedNoticeCodes: []string{"invalid_time_format"},
			description:         "Time format must include colons",
		},
		{
			name: "invalid time format - too many parts",
			files: map[string]string{
				StopTimesFile: "trip_id,arrival_time,departure_time,stop_id,stop_sequence\nT1,08:00:00:00,08:00:00,1,1", // Too many parts
			},
			expectedNoticeCodes: []string{"invalid_time_format"},
			description:         "Time format must have exactly 3 parts",
		},
		{
			name: "invalid time format - non-numeric",
			files: map[string]string{
				StopTimesFile: "trip_id,arrival_time,departure_time,stop_id,stop_sequence\nT1,AA:BB:CC,08:00:00,1,1", // Non-numeric
			},
			expectedNoticeCodes: []string{"invalid_time_format"},
			description:         "Time parts must be numeric",
		},
		{
			name: "negative time values",
			files: map[string]string{
				StopTimesFile: "trip_id,arrival_time,departure_time,stop_id,stop_sequence\nT1,-1:00:00,08:00:00,1,1", // Negative hour
			},
			expectedNoticeCodes: []string{"invalid_time_format"},
			description:         "Time values cannot be negative",
		},
		{
			name: "empty time fields ignored",
			files: map[string]string{
				StopTimesFile: "trip_id,arrival_time,departure_time,stop_id,stop_sequence\nT1,,08:00:00,1,1", // Empty arrival_time
			},
			expectedNoticeCodes: []string{},
			description:         "Empty time fields should not generate format errors",
		},
		{
			name: "whitespace-only time fields ignored",
			files: map[string]string{
				StopTimesFile: "trip_id,arrival_time,departure_time,stop_id,stop_sequence\nT1,   ,08:00:00,1,1", // Whitespace-only arrival_time
			},
			expectedNoticeCodes: []string{},
			description:         "Whitespace-only time fields should not generate format errors",
		},
		{
			name: "valid edge case times",
			files: map[string]string{
				StopTimesFile: "trip_id,arrival_time,departure_time,stop_id,stop_sequence\nT1,00:00:00,23:59:59,1,1", // Boundary values
			},
			expectedNoticeCodes: []string{},
			description:         "Boundary time values should be valid",
		},
		{
			name: "multiple rows with time errors",
			files: map[string]string{
				StopTimesFile: "trip_id,arrival_time,departure_time,stop_id,stop_sequence\nT1,8:00:00,08:00:00,1,1\nT1,08:00:00,8:00:00,1,2", // Errors in different rows
			},
			expectedNoticeCodes: []string{"invalid_time_format", "invalid_time_format"},
			description:         "Time format errors across multiple rows",
		},
		{
			name: "files without time fields",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,http://metro.example,America/Los_Angeles",
				"stops.txt":  "stop_id,stop_name,stop_lat,stop_lon\n1,Main St,34.05,-118.25",
			},
			expectedNoticeCodes: []string{},
			description:         "Files without time fields should not generate time format errors",
		},
		{
			name: "missing time files",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,http://metro.example,America/Los_Angeles",
			},
			expectedNoticeCodes: []string{},
			description:         "Missing files with time fields should not cause errors",
		},
		{
			name: "valid high hour values",
			files: map[string]string{
				StopTimesFile: "trip_id,arrival_time,departure_time,stop_id,stop_sequence\nT1,25:00:00,26:30:00,1,1\nT1,27:59:59,28:00:00,2,2", // High hour values
			},
			expectedNoticeCodes: []string{},
			description:         "Hour values > 24 are valid for service extending past midnight",
		},
		{
			name: "time field with whitespace padding",
			files: map[string]string{
				StopTimesFile: "trip_id,arrival_time,departure_time,stop_id,stop_sequence\nT1, 08:00:00 ,08:01:00,1,1", // Whitespace around time
			},
			expectedNoticeCodes: []string{},
			description:         "Whitespace around time values should be trimmed and validated",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test components
			loader := testutil.CreateTestFeedLoader(t, tt.files)
			container := notice.NewNoticeContainer()
			validator := NewTimeFormatValidator()
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

func TestTimeFormatValidator_IsValidGTFSTime(t *testing.T) {
	validator := NewTimeFormatValidator()

	tests := []struct {
		timeStr     string
		expected    bool
		description string
	}{
		{"08:00:00", true, "Standard valid time"},
		{"00:00:00", true, "Midnight"},
		{"23:59:59", true, "End of day"},
		{"24:00:00", true, "Next day service start"},
		{"25:30:45", true, "Next day service time"},
		{"99:59:59", true, "Very high hour (valid in GTFS)"},

		{"8:00:00", false, "Hour not zero-padded"},
		{"08:0:00", false, "Minute not zero-padded"},
		{"08:00:0", false, "Second not zero-padded"},
		{"08:60:00", false, "Invalid minute (60)"},
		{"08:00:60", false, "Invalid second (60)"},
		{"08:99:00", false, "Invalid minute (99)"},
		{"08:00:99", false, "Invalid second (99)"},

		{"-1:00:00", false, "Negative hour"},
		{"08:-1:00", false, "Negative minute"},
		{"08:00:-1", false, "Negative second"},

		{"08:00", false, "Missing seconds"},
		{"08", false, "Only hour"},
		{"08:00:00:00", false, "Too many parts"},
		{"", false, "Empty string"},

		{"AA:BB:CC", false, "Non-numeric parts"},
		{"08:AA:00", false, "Non-numeric minute"},
		{"08:00:AA", false, "Non-numeric second"},

		{"08.00.00", false, "Wrong separator"},
		{"08-00-00", false, "Wrong separator"},
		{"080000", false, "No separators"},

		{"8:0:0", false, "All parts not zero-padded"},
		{"008:000:000", false, "Too many digits"},
	}

	for _, tt := range tests {
		t.Run(tt.timeStr+"_"+tt.description, func(t *testing.T) {
			result := validator.isValidGTFSTime(tt.timeStr)
			if result != tt.expected {
				t.Errorf("isValidGTFSTime(%q) = %v, expected %v (%s)", tt.timeStr, result, tt.expected, tt.description)
			}
		})
	}
}

func TestTimeFormatValidator_ValidateTimeFormat(t *testing.T) {
	tests := []struct {
		name         string
		timeValue    string
		expectNotice bool
		description  string
	}{
		{
			name:         "valid time",
			timeValue:    "08:00:00",
			expectNotice: false,
			description:  "Valid time should not generate notice",
		},
		{
			name:         "invalid time format",
			timeValue:    "8:00:00",
			expectNotice: true,
			description:  "Invalid format should generate notice",
		},
		{
			name:         "next day service time",
			timeValue:    "25:30:00",
			expectNotice: false,
			description:  "Next day service times are valid",
		},
		{
			name:         "invalid minute",
			timeValue:    "08:60:00",
			expectNotice: true,
			description:  "Invalid minute should generate notice",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			container := notice.NewNoticeContainer()
			validator := NewTimeFormatValidator()

			validator.validateTimeFormat(container, StopTimesFile, "arrival_time", tt.timeValue, 1)

			notices := container.GetNotices()
			hasNotice := len(notices) > 0

			if hasNotice != tt.expectNotice {
				t.Errorf("Expected notice: %v, got notice: %v for %s", tt.expectNotice, hasNotice, tt.description)
			}

			if hasNotice && tt.expectNotice {
				// Verify notice details
				notice := notices[0]
				if notice.Code() != "invalid_time_format" {
					t.Errorf("Expected notice code 'invalid_time_format', got '%s'", notice.Code())
				}

				context := notice.Context()
				if filename, ok := context["filename"]; !ok || filename != StopTimesFile {
					t.Errorf("Expected filename '%s' in context, got '%v'", StopTimesFile, filename)
				}
				if fieldName, ok := context["fieldName"]; !ok || fieldName != "arrival_time" {
					t.Errorf("Expected fieldName 'arrival_time' in context, got '%v'", fieldName)
				}
				if timeValue, ok := context["timeValue"]; !ok || timeValue != tt.timeValue {
					t.Errorf("Expected timeValue '%s' in context, got '%v'", tt.timeValue, timeValue)
				}
				if rowNumber, ok := context["csvRowNumber"]; !ok || rowNumber != 1 {
					t.Errorf("Expected csvRowNumber 1 in context, got '%v'", rowNumber)
				}
			}
		})
	}
}

func TestTimeFormatValidator_TimeFields(t *testing.T) {
	// Test that timeFields map contains expected files and fields
	expectedTimeFields := map[string][]string{
		StopTimesFile:     {"arrival_time", "departure_time"},
		"frequencies.txt": {"start_time", "end_time"},
	}

	for filename, expectedFields := range expectedTimeFields {
		actualFields, exists := timeFields[filename]
		if !exists {
			t.Errorf("Expected file '%s' to be in timeFields map", filename)
			continue
		}

		if len(actualFields) != len(expectedFields) {
			t.Errorf("File '%s': expected %d time fields, got %d", filename, len(expectedFields), len(actualFields))
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
				t.Errorf("File '%s': expected time field '%s' not found", filename, expectedField)
			}
		}

		// Check for unexpected fields
		for _, actualField := range actualFields {
			if !expectedMap[actualField] {
				t.Errorf("File '%s': unexpected time field '%s'", filename, actualField)
			}
		}
	}
}

func TestTimeFormatValidator_ValidateFileTimeFields(t *testing.T) {
	tests := []struct {
		name            string
		filename        string
		content         string
		timeFields      []string
		expectedNotices int
		description     string
	}{
		{
			name:            "valid time fields",
			filename:        StopTimesFile,
			content:         "trip_id,arrival_time,departure_time,stop_id,stop_sequence\nT1,08:00:00,08:01:00,1,1",
			timeFields:      []string{"arrival_time", "departure_time"},
			expectedNotices: 0,
			description:     "Valid time formats should not generate notices",
		},
		{
			name:            "invalid time fields",
			filename:        StopTimesFile,
			content:         "trip_id,arrival_time,departure_time,stop_id,stop_sequence\nT1,8:00:00,08:1:00,1,1",
			timeFields:      []string{"arrival_time", "departure_time"},
			expectedNotices: 2,
			description:     "Invalid time formats should generate notices",
		},
		{
			name:            "missing time fields in data",
			filename:        StopTimesFile,
			content:         "trip_id,stop_id,stop_sequence\nT1,1,1", // Missing time fields
			timeFields:      []string{"arrival_time", "departure_time"},
			expectedNotices: 0,
			description:     "Missing fields should not generate format errors",
		},
		{
			name:            "empty time fields",
			filename:        StopTimesFile,
			content:         "trip_id,arrival_time,departure_time,stop_id,stop_sequence\nT1,,08:01:00,1,1",
			timeFields:      []string{"arrival_time", "departure_time"},
			expectedNotices: 0,
			description:     "Empty time fields should be ignored",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files := map[string]string{tt.filename: tt.content}
			loader := testutil.CreateTestFeedLoader(t, files)
			container := notice.NewNoticeContainer()
			validator := NewTimeFormatValidator()

			validator.validateFileTimeFields(loader, container, tt.filename, tt.timeFields)

			notices := container.GetNotices()
			if len(notices) != tt.expectedNotices {
				t.Errorf("Expected %d notices, got %d for %s", tt.expectedNotices, len(notices), tt.description)
			}
		})
	}
}

func TestTimeFormatValidator_New(t *testing.T) {
	validator := NewTimeFormatValidator()
	if validator == nil {
		t.Error("NewTimeFormatValidator() returned nil")
	}
}
