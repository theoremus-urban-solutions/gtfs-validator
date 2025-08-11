package entity

import (
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

func TestCalendarValidator_Validate(t *testing.T) {
	tests := []struct {
		name                string
		files               map[string]string
		expectedNoticeCodes []string
		description         string
	}{
		{
			name: "calendar.txt with data",
			files: map[string]string{
				"calendar.txt": "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\nS1,1,1,1,1,1,0,0,20250101,20251231",
			},
			expectedNoticeCodes: []string{},
			description:         "calendar.txt with data should be valid",
		},
		{
			name: "calendar_dates.txt with data",
			files: map[string]string{
				"calendar_dates.txt": "service_id,date,exception_type\nS1,20250101,1",
			},
			expectedNoticeCodes: []string{},
			description:         "calendar_dates.txt with data should be valid",
		},
		{
			name: "both calendar files with data",
			files: map[string]string{
				"calendar.txt":       "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\nS1,1,1,1,1,1,0,0,20250101,20251231",
				"calendar_dates.txt": "service_id,date,exception_type\nS1,20250101,1",
			},
			expectedNoticeCodes: []string{},
			description:         "Both calendar files with data should be valid",
		},
		{
			name: "no calendar files",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,http://metro.example,America/Los_Angeles",
			},
			expectedNoticeCodes: []string{"missing_calendar_and_calendar_date_files"},
			description:         "Missing both calendar files should generate notice",
		},
		{
			name: "empty calendar.txt file",
			files: map[string]string{
				"calendar.txt": "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date", // Headers only
			},
			expectedNoticeCodes: []string{"missing_calendar_and_calendar_date_files"},
			description:         "Empty calendar.txt (headers only) should generate notice",
		},
		{
			name: "empty calendar_dates.txt file",
			files: map[string]string{
				"calendar_dates.txt": "service_id,date,exception_type", // Headers only
			},
			expectedNoticeCodes: []string{"missing_calendar_and_calendar_date_files"},
			description:         "Empty calendar_dates.txt (headers only) should generate notice",
		},
		{
			name: "both calendar files empty",
			files: map[string]string{
				"calendar.txt":       "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date",
				"calendar_dates.txt": "service_id,date,exception_type",
			},
			expectedNoticeCodes: []string{"missing_calendar_and_calendar_date_files"},
			description:         "Both calendar files empty should generate notice",
		},
		{
			name: "calendar.txt has data, calendar_dates.txt empty",
			files: map[string]string{
				"calendar.txt":       "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\nS1,1,1,1,1,1,0,0,20250101,20251231",
				"calendar_dates.txt": "service_id,date,exception_type", // Empty
			},
			expectedNoticeCodes: []string{},
			description:         "One file with data is sufficient",
		},
		{
			name: "calendar.txt empty, calendar_dates.txt has data",
			files: map[string]string{
				"calendar.txt":       "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date", // Empty
				"calendar_dates.txt": "service_id,date,exception_type\nS1,20250101,1",
			},
			expectedNoticeCodes: []string{},
			description:         "One file with data is sufficient",
		},
		{
			name: "calendar.txt with multiple rows",
			files: map[string]string{
				"calendar.txt": "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\nS1,1,1,1,1,1,0,0,20250101,20251231\nS2,0,0,0,0,0,1,1,20250101,20251231",
			},
			expectedNoticeCodes: []string{},
			description:         "Multiple rows in calendar.txt should be valid",
		},
		{
			name: "calendar_dates.txt with multiple rows",
			files: map[string]string{
				"calendar_dates.txt": "service_id,date,exception_type\nS1,20250101,1\nS1,20250102,2\nS2,20250103,1",
			},
			expectedNoticeCodes: []string{},
			description:         "Multiple rows in calendar_dates.txt should be valid",
		},
		{
			name: "malformed calendar.txt ignored as no data",
			files: map[string]string{
				"calendar.txt": "invalid_csv_content_without_proper_headers",
			},
			expectedNoticeCodes: []string{"missing_calendar_and_calendar_date_files"},
			description:         "Malformed CSV should be treated as having no data",
		},
		{
			name: "only other GTFS files present",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,http://metro.example,America/Los_Angeles",
				"stops.txt":  "stop_id,stop_name,stop_lat,stop_lon\n1,Main St,34.05,-118.25",
				"routes.txt": "route_id,agency_id,route_short_name,route_type\nR1,1,Red,3",
			},
			expectedNoticeCodes: []string{"missing_calendar_and_calendar_date_files"},
			description:         "Presence of other GTFS files without calendar files should generate notice",
		},
		{
			name: "calendar files with whitespace-only content",
			files: map[string]string{
				"calendar.txt":       "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\n   \n\t\n", // Whitespace rows
				"calendar_dates.txt": "service_id,date,exception_type\n   ",
			},
			expectedNoticeCodes: []string{"missing_calendar_and_calendar_date_files"},
			description:         "Files with only whitespace rows should be treated as empty",
		},
		{
			name: "calendar.txt with valid row after empty rows",
			files: map[string]string{
				"calendar.txt": "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\n\n\nS1,1,1,1,1,1,0,0,20250101,20251231", // Valid row after empty ones
			},
			expectedNoticeCodes: []string{},
			description:         "File with valid data row should be considered as having data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := CreateTestFeedLoader(t, tt.files)
			container := notice.NewNoticeContainer()
			validator := NewCalendarValidator()
			config := gtfsvalidator.Config{}

			validator.Validate(loader, container, config)

			notices := container.GetNotices()

			if len(notices) != len(tt.expectedNoticeCodes) {
				t.Errorf("Expected %d notices, got %d for case: %s", len(tt.expectedNoticeCodes), len(notices), tt.description)
			}

			expectedCodeCounts := make(map[string]int)
			for _, code := range tt.expectedNoticeCodes {
				expectedCodeCounts[code]++
			}

			actualCodeCounts := make(map[string]int)
			for _, notice := range notices {
				actualCodeCounts[notice.Code()]++
			}

			for expectedCode, expectedCount := range expectedCodeCounts {
				actualCount := actualCodeCounts[expectedCode]
				if actualCount != expectedCount {
					t.Errorf("Expected %d notices with code '%s', got %d", expectedCount, expectedCode, actualCount)
				}
			}

			for actualCode := range actualCodeCounts {
				if expectedCodeCounts[actualCode] == 0 {
					t.Errorf("Unexpected notice code: %s", actualCode)
				}
			}
		})
	}
}

func TestCalendarValidator_FileHasData(t *testing.T) {
	tests := []struct {
		name        string
		filename    string
		content     string
		expected    bool
		description string
	}{
		{
			name:        "file with data row",
			filename:    "calendar.txt",
			content:     "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\nS1,1,1,1,1,1,0,0,20250101,20251231",
			expected:    true,
			description: "File with data row should return true",
		},
		{
			name:        "file with headers only",
			filename:    "calendar.txt",
			content:     "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date",
			expected:    false,
			description: "File with headers only should return false",
		},
		{
			name:        "file with empty rows",
			filename:    "calendar.txt",
			content:     "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\n\n\n",
			expected:    false,
			description: "File with empty rows should return false",
		},
		{
			name:        "file with one valid row after empty rows",
			filename:    "calendar.txt",
			content:     "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\n\n\nS1,1,1,1,1,1,0,0,20250101,20251231",
			expected:    true,
			description: "File with valid row after empty rows should return true",
		},
		{
			name:        "malformed CSV file",
			filename:    "calendar.txt",
			content:     "invalid_csv_content",
			expected:    false,
			description: "Malformed CSV should return false",
		},
		{
			name:        "calendar_dates.txt with data",
			filename:    "calendar_dates.txt",
			content:     "service_id,date,exception_type\nS1,20250101,1",
			expected:    true,
			description: "calendar_dates.txt with data should return true",
		},
		{
			name:        "calendar_dates.txt headers only",
			filename:    "calendar_dates.txt",
			content:     "service_id,date,exception_type",
			expected:    false,
			description: "calendar_dates.txt with headers only should return false",
		},
		{
			name:        "file with multiple data rows",
			filename:    "calendar.txt",
			content:     "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\nS1,1,1,1,1,1,0,0,20250101,20251231\nS2,0,0,0,0,0,1,1,20250101,20251231",
			expected:    true,
			description: "File with multiple data rows should return true",
		},
		{
			name:        "completely empty file",
			filename:    "calendar.txt",
			content:     "",
			expected:    false,
			description: "Completely empty file should return false",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files := map[string]string{tt.filename: tt.content}
			loader := CreateTestFeedLoader(t, files)
			validator := NewCalendarValidator()

			result := validator.fileHasData(loader, tt.filename)

			if result != tt.expected {
				t.Errorf("Expected fileHasData('%s') to return %v, got %v for %s", tt.filename, tt.expected, result, tt.description)
			}
		})
	}
}

func TestCalendarValidator_FileHasData_MissingFile(t *testing.T) {
	// Test behavior when file doesn't exist
	loader := CreateTestFeedLoader(t, map[string]string{})
	validator := NewCalendarValidator()

	result := validator.fileHasData(loader, "nonexistent.txt")

	if result != false {
		t.Errorf("Expected fileHasData for missing file to return false, got %v", result)
	}
}

func TestCalendarValidator_New(t *testing.T) {
	validator := NewCalendarValidator()
	if validator == nil {
		t.Error("NewCalendarValidator() returned nil")
	}
}
