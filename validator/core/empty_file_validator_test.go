package core

import (
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

func TestEmptyFileValidator_Validate(t *testing.T) {
	tests := []struct {
		name                string
		files               map[string]string
		expectedNoticeCodes []string
		description         string
	}{
		{
			name: "all files have data",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,http://metro.example,America/Los_Angeles",
				"stops.txt":  "stop_id,stop_name,stop_lat,stop_lon\n1,Main St,34.05,-118.25",
			},
			expectedNoticeCodes: []string{},
			description:         "All files contain data rows",
		},
		{
			name: "one file is empty (headers only)",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,http://metro.example,America/Los_Angeles",
				"stops.txt":  "stop_id,stop_name,stop_lat,stop_lon", // No data rows
			},
			expectedNoticeCodes: []string{"empty_file"},
			description:         "stops.txt has headers but no data",
		},
		{
			name: "multiple empty files",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone",                                  // Headers only
				"stops.txt":  "stop_id,stop_name,stop_lat,stop_lon",                                               // Headers only
				"routes.txt": "route_id,agency_id,route_short_name,route_long_name,route_type\n1,1,1,Main Line,3", // Has data
			},
			expectedNoticeCodes: []string{"empty_file", "empty_file"},
			description:         "Two files are empty, one has data",
		},
		{
			name: "completely empty file (no headers, no data)",
			files: map[string]string{
				"agency.txt": "", // Completely empty
				"stops.txt":  "stop_id,stop_name,stop_lat,stop_lon\n1,Main St,34.05,-118.25",
			},
			expectedNoticeCodes: []string{},
			description:         "Completely empty files are handled by CSV parser errors, not empty file validator",
		},
		{
			name: "file with only whitespace after headers",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\n   \n  \n", // Whitespace rows
				"stops.txt":  "stop_id,stop_name,stop_lat,stop_lon\n1,Main St,34.05,-118.25",
			},
			expectedNoticeCodes: []string{"empty_file"},
			description:         "File with only whitespace rows after headers is considered empty",
		},
		{
			name: "single file with no data",
			files: map[string]string{
				"feed_info.txt": "feed_publisher_name,feed_publisher_url,feed_lang",
			},
			expectedNoticeCodes: []string{"empty_file"},
			description:         "Optional file with headers but no data",
		},
		{
			name: "file with UTF-8 BOM and no data",
			files: map[string]string{
				"agency.txt": "\ufeffagency_id,agency_name,agency_url,agency_timezone",
			},
			expectedNoticeCodes: []string{"empty_file"},
			description:         "File with UTF-8 BOM and headers but no data rows",
		},
		{
			name: "file with data after empty rows",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\n\n\n1,Metro,http://metro.example,America/Los_Angeles",
			},
			expectedNoticeCodes: []string{},
			description:         "File with empty rows before data should not be flagged as empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test components
			loader := CreateTestFeedLoader(t, tt.files)
			container := notice.NewNoticeContainer()
			validator := NewEmptyFileValidator()
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

func TestEmptyFileValidator_ValidateFileNotEmpty(t *testing.T) {
	tests := []struct {
		name              string
		filename          string
		content           string
		expectEmptyNotice bool
	}{
		{
			name:              "file with data",
			filename:          "agency.txt",
			content:           "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,http://metro.example,America/Los_Angeles",
			expectEmptyNotice: false,
		},
		{
			name:              "file with headers only",
			filename:          "stops.txt",
			content:           "stop_id,stop_name,stop_lat,stop_lon",
			expectEmptyNotice: true,
		},
		{
			name:              "file with headers and empty line",
			filename:          "routes.txt",
			content:           "route_id,agency_id,route_short_name,route_long_name,route_type\n",
			expectEmptyNotice: true,
		},
		{
			name:              "file with headers and whitespace line",
			filename:          "trips.txt",
			content:           "route_id,service_id,trip_id\n   ",
			expectEmptyNotice: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test components
			files := map[string]string{tt.filename: tt.content}
			loader := CreateTestFeedLoader(t, files)
			container := notice.NewNoticeContainer()
			validator := NewEmptyFileValidator()

			// Run validation on specific file
			validator.validateFileNotEmpty(loader, container, tt.filename)

			// Get notices
			notices := container.GetNotices()

			// Check if empty notice was generated as expected
			hasEmptyNotice := false
			for _, notice := range notices {
				if notice.Code() == "empty_file" {
					hasEmptyNotice = true
					// Verify context contains correct filename
					context := notice.Context()
					if filename, ok := context["filename"]; !ok || filename != tt.filename {
						t.Errorf("Expected filename '%s' in notice context, got '%v'", tt.filename, filename)
					}
				}
			}

			if hasEmptyNotice != tt.expectEmptyNotice {
				t.Errorf("Expected empty notice: %v, got empty notice: %v", tt.expectEmptyNotice, hasEmptyNotice)
			}
		})
	}
}

func TestEmptyFileValidator_New(t *testing.T) {
	validator := NewEmptyFileValidator()
	if validator == nil {
		t.Error("NewEmptyFileValidator() returned nil")
	}
}

func TestEmptyFileValidator_FileNotExists(t *testing.T) {
	// Test behavior when file doesn't exist
	loader := CreateTestFeedLoader(t, map[string]string{}) // No files
	container := notice.NewNoticeContainer()
	validator := NewEmptyFileValidator()

	// Try to validate a non-existent file
	validator.validateFileNotEmpty(loader, container, "nonexistent.txt")

	// Should not generate any notices (other validators handle missing files)
	notices := container.GetNotices()
	if len(notices) != 0 {
		t.Errorf("Expected no notices for non-existent file, got %d", len(notices))
	}
}

func TestEmptyFileValidator_MalformedCSV(t *testing.T) {
	// Test behavior with malformed CSV content
	files := map[string]string{
		"malformed.txt": "agency_id,agency_name\n\"unclosed quote", // Malformed CSV
	}
	loader := CreateTestFeedLoader(t, files)
	container := notice.NewNoticeContainer()
	validator := NewEmptyFileValidator()

	validator.validateFileNotEmpty(loader, container, "malformed.txt")

	// Should not generate empty file notices (CSV parsing errors handled elsewhere)
	notices := container.GetNotices()
	for _, notice := range notices {
		if notice.Code() == "empty_file" {
			t.Error("Should not generate empty_file notice for malformed CSV")
		}
	}
}
