package core

import (
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/testutil"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
)

func TestLeadingTrailingWhitespaceValidator_IsTextField(t *testing.T) {
	validator := NewLeadingTrailingWhitespaceValidator()

	textFields := []string{
		"agency_name", "agency_url", "stop_name", "stop_desc", "route_short_name",
		"route_long_name", "trip_headsign", "stop_headsign", "service_id", "trip_id",
		"stop_id", "route_id", "agency_id", "shape_id", "fare_id", "zone_id",
	}

	numericFields := []string{
		"stop_lat", "stop_lon", "route_type", "direction_id", "location_type",
		"wheelchair_boarding", "wheelchair_accessible", "bikes_allowed", "stop_sequence",
		"pickup_type", "drop_off_type", "timepoint", "monday", "tuesday", "exception_type",
		"payment_method", "transfers", "shape_pt_lat", "shape_pt_lon", "shape_pt_sequence",
		"headway_secs", "exact_times", "transfer_type", "min_transfer_time",
	}

	// Test text fields
	for _, field := range textFields {
		if !validator.isTextField(field) {
			t.Errorf("Expected '%s' to be identified as a text field", field)
		}
	}

	// Test numeric fields
	for _, field := range numericFields {
		if validator.isTextField(field) {
			t.Errorf("Expected '%s' to be identified as a numeric field", field)
		}
	}
}

func TestLeadingTrailingWhitespaceValidator_ShouldValidateField(t *testing.T) {
	validator := NewLeadingTrailingWhitespaceValidator()

	tests := []struct {
		fieldName         string
		significantFields map[string]bool
		expected          bool
		description       string
	}{
		{
			fieldName:         "agency_name",
			significantFields: map[string]bool{"agency_name": true, "agency_id": true},
			expected:          true,
			description:       "Field in significant fields should be validated",
		},
		{
			fieldName:         "agency_phone",
			significantFields: map[string]bool{"agency_name": true, "agency_id": true},
			expected:          false,
			description:       "Field not in significant fields should not be validated",
		},
		{
			fieldName:         "agency_name",
			significantFields: map[string]bool{}, // Empty map - use text field detection
			expected:          true,
			description:       "With empty significant fields, text fields should be validated",
		},
		{
			fieldName:         "stop_lat",
			significantFields: map[string]bool{}, // Empty map - use text field detection
			expected:          false,
			description:       "With empty significant fields, numeric fields should not be validated",
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			result := validator.shouldValidateField(tt.fieldName, tt.significantFields)
			if result != tt.expected {
				t.Errorf("Expected shouldValidateField('%s') to return %v, got %v", tt.fieldName, tt.expected, result)
			}
		})
	}
}

func TestLeadingTrailingWhitespaceValidator_GetSignificantFields(t *testing.T) {
	validator := NewLeadingTrailingWhitespaceValidator()

	tests := []struct {
		filename         string
		expectedFields   []string
		unexpectedFields []string
		description      string
	}{
		{
			filename:         "agency.txt",
			expectedFields:   []string{"agency_id", "agency_name", "agency_url", "agency_timezone"},
			unexpectedFields: []string{"stop_id", "route_id"},
			description:      "Agency file should have agency-specific fields",
		},
		{
			filename:         "stops.txt",
			expectedFields:   []string{"stop_id", "stop_name", "zone_id", "parent_station"},
			unexpectedFields: []string{"agency_id", "route_id"},
			description:      "Stops file should have stop-specific fields",
		},
		{
			filename:         "routes.txt",
			expectedFields:   []string{"route_id", "agency_id", "route_short_name", "route_long_name"},
			unexpectedFields: []string{"stop_id", "trip_id"},
			description:      "Routes file should have route-specific fields",
		},
		{
			filename:         "unknown_file.txt",
			expectedFields:   []string{},
			unexpectedFields: []string{"agency_id", "stop_id"},
			description:      "Unknown files should return empty significant fields map",
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			significantFields := validator.getSignificantFields(tt.filename)

			// Check expected fields are present
			for _, field := range tt.expectedFields {
				if !significantFields[field] {
					t.Errorf("Expected field '%s' to be significant for %s", field, tt.filename)
				}
			}

			// Check unexpected fields are not present
			for _, field := range tt.unexpectedFields {
				if significantFields[field] {
					t.Errorf("Did not expect field '%s' to be significant for %s", field, tt.filename)
				}
			}
		})
	}
}

func TestLeadingTrailingWhitespaceValidator_ValidateFile(t *testing.T) {
	tests := []struct {
		name                string
		filename            string
		content             string
		expectedNoticeCount int
		description         string
	}{
		{
			name:                "file with no whitespace issues",
			filename:            "agency.txt",
			content:             "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,http://metro.example,America/Los_Angeles",
			expectedNoticeCount: 0,
			description:         "Clean file should generate no notices",
		},
		{
			name:                "file with multiple whitespace issues",
			filename:            "agency.txt",
			content:             "agency_id,agency_name,agency_url,agency_timezone\n 1 , Metro , http://metro.example ,America/Los_Angeles",
			expectedNoticeCount: 3, // One per affected value: the first 3 fields
			description:         "File with multiple issues should generate multiple notices",
		},
		{
			name:                "file with CSV parsing error",
			filename:            "agency.txt",
			content:             "agency_id,agency_name,agency_url,agency_timezone\n1,Metro", // Incomplete row
			expectedNoticeCount: 0,
			description:         "CSV parsing errors should not cause crashes",
		},
		{
			name:                "empty file",
			filename:            "agency.txt",
			content:             "agency_id,agency_name,agency_url,agency_timezone", // Headers only
			expectedNoticeCount: 0,
			description:         "Empty file (headers only) should generate no notices",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files := map[string]string{tt.filename: tt.content}
			loader := testutil.CreateTestFeedLoader(t, files)
			container := notice.NewNoticeContainer()
			validator := NewLeadingTrailingWhitespaceValidator()

			validator.validateFile(loader, container, tt.filename)

			notices := container.GetNotices()
			if len(notices) != tt.expectedNoticeCount {
				t.Errorf("Expected %d notices, got %d for %s", tt.expectedNoticeCount, len(notices), tt.description)
			}
		})
	}
}

func TestLeadingTrailingWhitespaceValidator_New(t *testing.T) {
	validator := NewLeadingTrailingWhitespaceValidator()
	if validator == nil {
		t.Error("NewLeadingTrailingWhitespaceValidator() returned nil")
	}
}
