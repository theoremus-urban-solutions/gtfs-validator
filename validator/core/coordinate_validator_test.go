package core

import (
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

func TestCoordinateValidator_Validate(t *testing.T) {
	tests := []struct {
		name                string
		files               map[string]string
		expectedNoticeCodes []string
		description         string
	}{
		{
			name: "all coordinates valid",
			files: map[string]string{
				"stops.txt":  "stop_id,stop_name,stop_lat,stop_lon\n1,Main St,34.0522,-118.2437",
				"shapes.txt": "shape_id,shape_pt_lat,shape_pt_lon,shape_pt_sequence\nS1,34.0522,-118.2437,1",
			},
			expectedNoticeCodes: []string{},
			description:         "All coordinates are valid with sufficient precision",
		},
		{
			name: "stops.txt invalid latitude range",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_lat,stop_lon\n1,Main St,91.0000,-118.2437", // Latitude > 90
			},
			expectedNoticeCodes: []string{"invalid_coordinate", "insufficient_coordinate_precision"},
			description:         "Latitude exceeds valid range",
		},
		{
			name: "stops.txt invalid longitude range",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_lat,stop_lon\n1,Main St,34.0522,181.0000", // Longitude > 180
			},
			expectedNoticeCodes: []string{"invalid_coordinate", "insufficient_coordinate_precision"},
			description:         "Longitude exceeds valid range",
		},
		{
			name: "shapes.txt invalid negative latitude range",
			files: map[string]string{
				"shapes.txt": "shape_id,shape_pt_lat,shape_pt_lon,shape_pt_sequence\nS1,-91.0000,-118.2437,1", // Latitude < -90
			},
			expectedNoticeCodes: []string{"invalid_coordinate", "insufficient_coordinate_precision"},
			description:         "Latitude below valid range",
		},
		{
			name: "shapes.txt invalid negative longitude range",
			files: map[string]string{
				"shapes.txt": "shape_id,shape_pt_lat,shape_pt_lon,shape_pt_sequence\nS1,34.0522,-181.0000,1", // Longitude < -180
			},
			expectedNoticeCodes: []string{"invalid_coordinate", "insufficient_coordinate_precision"},
			description:         "Longitude below valid range",
		},
		{
			name: "non-numeric coordinates",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_lat,stop_lon\n1,Main St,invalid,not-a-number",
			},
			expectedNoticeCodes: []string{"invalid_coordinate", "invalid_coordinate"},
			description:         "Non-numeric coordinate values should generate format errors",
		},
		{
			name: "suspicious zero coordinates",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_lat,stop_lon\n1,Main St,0.0000,0.0000",
			},
			expectedNoticeCodes: []string{"suspicious_coordinate", "suspicious_coordinate", "insufficient_coordinate_precision", "insufficient_coordinate_precision"},
			description:         "Zero coordinates are suspicious and may indicate missing data",
		},
		{
			name: "insufficient coordinate precision",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_lat,stop_lon\n1,Main St,34.05,-118.24", // Only 2 decimal places
			},
			expectedNoticeCodes: []string{"insufficient_coordinate_precision", "insufficient_coordinate_precision"},
			description:         "Coordinates with less than 4 decimal places have insufficient precision",
		},
		{
			name: "no decimal places",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_lat,stop_lon\n1,Main St,34,-118", // No decimal places
			},
			expectedNoticeCodes: []string{"insufficient_coordinate_precision", "insufficient_coordinate_precision"},
			description:         "Integer coordinates have very low precision",
		},
		{
			name: "valid boundary coordinates",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_lat,stop_lon\n1,North Pole,90.0000,0.0000\n2,South Pole,-90.0000,0.0000\n3,Date Line,0.0000,180.0000\n4,Anti-Meridian,0.0000,-180.0000",
			},
			expectedNoticeCodes: []string{"suspicious_coordinate", "suspicious_coordinate", "suspicious_coordinate", "suspicious_coordinate", "insufficient_coordinate_precision", "insufficient_coordinate_precision", "insufficient_coordinate_precision", "insufficient_coordinate_precision", "insufficient_coordinate_precision", "insufficient_coordinate_precision", "insufficient_coordinate_precision", "insufficient_coordinate_precision"},
			description:         "Boundary coordinates are valid but zero values are suspicious",
		},
		{
			name: "high precision coordinates",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_lat,stop_lon\n1,Main St,34.052234,-118.243685",
			},
			expectedNoticeCodes: []string{},
			description:         "High precision coordinates should not generate notices",
		},
		{
			name: "mixed valid and invalid coordinates",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_lat,stop_lon\n1,Valid,34.0522,-118.2437\n2,Invalid Lat,91.0000,-118.2437\n3,Invalid Lon,34.0522,181.0000",
			},
			expectedNoticeCodes: []string{"invalid_coordinate", "insufficient_coordinate_precision", "invalid_coordinate", "insufficient_coordinate_precision"},
			description:         "Mix of valid and invalid coordinates",
		},
		{
			name: "empty coordinate fields ignored",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_lat,stop_lon\n1,Main St,,", // Empty coordinates
			},
			expectedNoticeCodes: []string{},
			description:         "Empty coordinate fields should not generate validation errors",
		},
		{
			name: "whitespace-only coordinate fields ignored",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_lat,stop_lon\n1,Main St,   ,   ", // Whitespace coordinates
			},
			expectedNoticeCodes: []string{},
			description:         "Whitespace-only coordinate fields should not generate validation errors",
		},
		{
			name: "coordinate fields with whitespace padding",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_lat,stop_lon\n1,Main St, 34.0522 , -118.2437 ", // Whitespace around coordinates
			},
			expectedNoticeCodes: []string{},
			description:         "Whitespace around coordinate values should be trimmed",
		},
		{
			name: "negative coordinates valid",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_lat,stop_lon\n1,Southern Hemisphere,-34.0522,118.2437",
			},
			expectedNoticeCodes: []string{},
			description:         "Negative coordinates within range should be valid",
		},
		{
			name: "scientific notation coordinates",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_lat,stop_lon\n1,Main St,3.40522e1,-1.182437e2", // Scientific notation
			},
			expectedNoticeCodes: []string{"insufficient_coordinate_precision", "insufficient_coordinate_precision"},
			description:         "Scientific notation coordinates are valid numbers but may have precision issues",
		},
		{
			name: "very high precision coordinates",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_lat,stop_lon\n1,Main St,34.052234567890,-118.243685123456",
			},
			expectedNoticeCodes: []string{},
			description:         "Very high precision coordinates should be valid",
		},
		{
			name: "shapes.txt coordinate validation",
			files: map[string]string{
				"shapes.txt": "shape_id,shape_pt_lat,shape_pt_lon,shape_pt_sequence\nS1,34.0522,-118.2437,1\nS1,91.0000,-118.2437,2\nS1,34.0522,181.0000,3",
			},
			expectedNoticeCodes: []string{"invalid_coordinate", "insufficient_coordinate_precision", "invalid_coordinate", "insufficient_coordinate_precision"},
			description:         "shapes.txt should have same coordinate validation as stops.txt",
		},
		{
			name: "multiple coordinate errors in single row",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_lat,stop_lon\n1,Main St,91.0000,181.0000", // Both coordinates invalid
			},
			expectedNoticeCodes: []string{"invalid_coordinate", "insufficient_coordinate_precision", "invalid_coordinate", "insufficient_coordinate_precision"},
			description:         "Multiple coordinate errors in single row should generate multiple notices",
		},
		{
			name: "files without coordinate fields",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,http://metro.example,America/Los_Angeles",
			},
			expectedNoticeCodes: []string{},
			description:         "Files without coordinate fields should not generate coordinate validation errors",
		},
		{
			name: "missing coordinate files",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,http://metro.example,America/Los_Angeles",
			},
			expectedNoticeCodes: []string{},
			description:         "Missing files with coordinate fields should not cause errors",
		},
		{
			name: "coordinate precision boundary cases",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_lat,stop_lon\n1,Exactly 4 decimals,34.0522,-118.2437\n2,Less than 4,34.052,-118.243\n3,More than 4,34.05223,-118.24376",
			},
			expectedNoticeCodes: []string{"insufficient_coordinate_precision", "insufficient_coordinate_precision"},
			description:         "Only coordinates with less than 4 decimal places should generate precision notices",
		},
		{
			name: "special coordinate values",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_lat,stop_lon\n1,Equator,0.0000,45.0000\n2,Prime Meridian,45.0000,0.0000",
			},
			expectedNoticeCodes: []string{"suspicious_coordinate", "insufficient_coordinate_precision", "suspicious_coordinate", "insufficient_coordinate_precision"},
			description:         "Zero coordinates should generate suspicious notices even when valid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test components
			loader := CreateTestFeedLoader(t, tt.files)
			container := notice.NewNoticeContainer()
			validator := NewCoordinateValidator()
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

func TestCoordinateValidator_ValidateCoordinate(t *testing.T) {
	tests := []struct {
		name            string
		fieldName       string
		coordValue      string
		expectedNotices []string
		description     string
	}{
		{
			name:            "valid latitude",
			fieldName:       "stop_lat",
			coordValue:      "34.0522",
			expectedNotices: []string{},
			description:     "Valid latitude with sufficient precision",
		},
		{
			name:            "valid longitude",
			fieldName:       "stop_lon",
			coordValue:      "-118.2437",
			expectedNotices: []string{},
			description:     "Valid longitude with sufficient precision",
		},
		{
			name:            "invalid latitude too high",
			fieldName:       "stop_lat",
			coordValue:      "91.0000",
			expectedNotices: []string{"invalid_coordinate", "insufficient_coordinate_precision"},
			description:     "Latitude > 90 is invalid",
		},
		{
			name:            "invalid latitude too low",
			fieldName:       "stop_lat",
			coordValue:      "-91.0000",
			expectedNotices: []string{"invalid_coordinate", "insufficient_coordinate_precision"},
			description:     "Latitude < -90 is invalid",
		},
		{
			name:            "invalid longitude too high",
			fieldName:       "stop_lon",
			coordValue:      "181.0000",
			expectedNotices: []string{"invalid_coordinate", "insufficient_coordinate_precision"},
			description:     "Longitude > 180 is invalid",
		},
		{
			name:            "invalid longitude too low",
			fieldName:       "stop_lon",
			coordValue:      "-181.0000",
			expectedNotices: []string{"invalid_coordinate", "insufficient_coordinate_precision"},
			description:     "Longitude < -180 is invalid",
		},
		{
			name:            "suspicious zero latitude",
			fieldName:       "stop_lat",
			coordValue:      "0.0000",
			expectedNotices: []string{"suspicious_coordinate", "insufficient_coordinate_precision"},
			description:     "Zero latitude is suspicious",
		},
		{
			name:            "suspicious zero longitude",
			fieldName:       "stop_lon",
			coordValue:      "0.0000",
			expectedNotices: []string{"suspicious_coordinate", "insufficient_coordinate_precision"},
			description:     "Zero longitude is suspicious",
		},
		{
			name:            "insufficient precision",
			fieldName:       "stop_lat",
			coordValue:      "34.05",
			expectedNotices: []string{"insufficient_coordinate_precision"},
			description:     "Less than 4 decimal places",
		},
		{
			name:            "no decimal places",
			fieldName:       "stop_lat",
			coordValue:      "34",
			expectedNotices: []string{"insufficient_coordinate_precision"},
			description:     "No decimal places at all",
		},
		{
			name:            "non-numeric coordinate",
			fieldName:       "stop_lat",
			coordValue:      "invalid",
			expectedNotices: []string{"invalid_coordinate"},
			description:     "Non-numeric values should generate format errors",
		},
		{
			name:            "boundary latitude positive",
			fieldName:       "stop_lat",
			coordValue:      "90.0000",
			expectedNotices: []string{"insufficient_coordinate_precision"},
			description:     "Latitude 90 is valid but has precision issue",
		},
		{
			name:            "boundary latitude negative",
			fieldName:       "stop_lat",
			coordValue:      "-90.0000",
			expectedNotices: []string{"insufficient_coordinate_precision"},
			description:     "Latitude -90 is valid but has precision issue",
		},
		{
			name:            "boundary longitude positive",
			fieldName:       "stop_lon",
			coordValue:      "180.0000",
			expectedNotices: []string{"insufficient_coordinate_precision"},
			description:     "Longitude 180 is valid but has precision issue",
		},
		{
			name:            "boundary longitude negative",
			fieldName:       "stop_lon",
			coordValue:      "-180.0000",
			expectedNotices: []string{"insufficient_coordinate_precision"},
			description:     "Longitude -180 is valid but has precision issue",
		},
		{
			name:            "high precision coordinate",
			fieldName:       "stop_lat",
			coordValue:      "34.052234567890",
			expectedNotices: []string{},
			description:     "High precision coordinates should be valid",
		},
		{
			name:            "scientific notation",
			fieldName:       "stop_lat",
			coordValue:      "3.40522e1",
			expectedNotices: []string{"insufficient_coordinate_precision"},
			description:     "Scientific notation is valid but may have precision issues",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			container := notice.NewNoticeContainer()
			validator := NewCoordinateValidator()

			validator.validateCoordinate(container, "stops.txt", tt.fieldName, tt.coordValue, 1)

			notices := container.GetNotices()

			// Count notice codes
			expectedCodeCounts := make(map[string]int)
			for _, code := range tt.expectedNotices {
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
					t.Errorf("Expected %d notices with code '%s', got %d for %s", expectedCount, expectedCode, actualCount, tt.description)
				}
			}

			// Check for unexpected notice codes
			for actualCode := range actualCodeCounts {
				if expectedCodeCounts[actualCode] == 0 {
					t.Errorf("Unexpected notice code: %s for %s", actualCode, tt.description)
				}
			}

			// Verify notice context for the first notice if expected
			if len(tt.expectedNotices) > 0 && len(notices) > 0 {
				notice := notices[0]
				context := notice.Context()

				if filename, ok := context["filename"]; !ok || filename != "stops.txt" {
					t.Errorf("Expected filename 'stops.txt' in context, got '%v'", filename)
				}
				if fieldName, ok := context["fieldName"]; !ok || fieldName != tt.fieldName {
					t.Errorf("Expected fieldName '%s' in context, got '%v'", tt.fieldName, fieldName)
				}
				if coordValue, ok := context["fieldValue"]; !ok || coordValue != tt.coordValue {
					t.Errorf("Expected fieldValue '%s' in context, got '%v'", tt.coordValue, coordValue)
				}
				if rowNumber, ok := context["csvRowNumber"]; !ok || rowNumber != 1 {
					t.Errorf("Expected csvRowNumber 1 in context, got '%v'", rowNumber)
				}
			}
		})
	}
}

func TestCoordinateValidator_CoordinateFields(t *testing.T) {
	// Test that coordinateFields map contains expected files and fields
	expectedCoordinateFields := map[string][]string{
		"stops.txt":  {"stop_lat", "stop_lon"},
		"shapes.txt": {"shape_pt_lat", "shape_pt_lon"},
	}

	for filename, expectedFields := range expectedCoordinateFields {
		actualFields, exists := coordinateFields[filename]
		if !exists {
			t.Errorf("Expected file '%s' to be in coordinateFields map", filename)
			continue
		}

		if len(actualFields) != len(expectedFields) {
			t.Errorf("File '%s': expected %d coordinate fields, got %d", filename, len(expectedFields), len(actualFields))
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
				t.Errorf("File '%s': expected coordinate field '%s' not found", filename, expectedField)
			}
		}

		// Check for unexpected fields
		for _, actualField := range actualFields {
			if !expectedMap[actualField] {
				t.Errorf("File '%s': unexpected coordinate field '%s'", filename, actualField)
			}
		}
	}
}

func TestCoordinateValidator_ValidateFileCoordinates(t *testing.T) {
	tests := []struct {
		name                string
		filename            string
		content             string
		coordinateFields    []string
		expectedNoticeCount int
		description         string
	}{
		{
			name:                "valid coordinates",
			filename:            "stops.txt",
			content:             "stop_id,stop_name,stop_lat,stop_lon\n1,Main St,34.0522,-118.2437",
			coordinateFields:    []string{"stop_lat", "stop_lon"},
			expectedNoticeCount: 0,
			description:         "Valid coordinates should not generate notices",
		},
		{
			name:                "invalid coordinates",
			filename:            "stops.txt",
			content:             "stop_id,stop_name,stop_lat,stop_lon\n1,Main St,91.0000,181.0000",
			coordinateFields:    []string{"stop_lat", "stop_lon"},
			expectedNoticeCount: 4, // 2 invalid coordinate + 2 insufficient precision
			description:         "Invalid coordinates should generate notices",
		},
		{
			name:                "missing coordinate fields in data",
			filename:            "stops.txt",
			content:             "stop_id,stop_name\n1,Main St", // Missing coordinate fields
			coordinateFields:    []string{"stop_lat", "stop_lon"},
			expectedNoticeCount: 0,
			description:         "Missing fields should not generate validation errors",
		},
		{
			name:                "empty coordinate fields",
			filename:            "stops.txt",
			content:             "stop_id,stop_name,stop_lat,stop_lon\n1,Main St,,",
			coordinateFields:    []string{"stop_lat", "stop_lon"},
			expectedNoticeCount: 0,
			description:         "Empty coordinate fields should be ignored",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files := map[string]string{tt.filename: tt.content}
			loader := CreateTestFeedLoader(t, files)
			container := notice.NewNoticeContainer()
			validator := NewCoordinateValidator()

			validator.validateFileCoordinates(loader, container, tt.filename, tt.coordinateFields)

			notices := container.GetNotices()
			if len(notices) != tt.expectedNoticeCount {
				t.Errorf("Expected %d notices, got %d for %s", tt.expectedNoticeCount, len(notices), tt.description)
			}
		})
	}
}

func TestCoordinateValidator_New(t *testing.T) {
	validator := NewCoordinateValidator()
	if validator == nil {
		t.Error("NewCoordinateValidator() returned nil")
	}
}
