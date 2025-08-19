package entity

import (
	"fmt"
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/testutil"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/parser"
	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

func TestRouteNameValidator_Validate(t *testing.T) {
	tests := []struct {
		name                string
		files               map[string]string
		expectedNoticeCodes []string
		description         string
	}{
		{
			name: "valid route names",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_long_name,route_type\n" +
					"route1,1,Main Line,3\n" +
					"route2,Red,Red Line Express,1",
			},
			expectedNoticeCodes: []string{},
			description:         "Routes with proper short and long names should be valid",
		},
		{
			name: "route with only short name",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_type\n" +
					"route1,42,3",
			},
			expectedNoticeCodes: []string{},
			description:         "Route with only short name should be valid",
		},
		{
			name: "route with only long name",
			files: map[string]string{
				"routes.txt": "route_id,route_long_name,route_type\n" +
					"route1,Central Express Line,3",
			},
			expectedNoticeCodes: []string{"missing_recommended_field"},
			description:         "Bus route with only long name should recommend short name",
		},
		{
			name: "route without both names",
			files: map[string]string{
				"routes.txt": "route_id,route_type\n" +
					"route1,3",
			},
			expectedNoticeCodes: []string{"missing_route_name"},
			description:         "Route without both short and long names should generate error",
		},
		{
			name: "route with empty names",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_long_name,route_type\n" +
					"route1,,,3",
			},
			expectedNoticeCodes: []string{"missing_route_name"},
			description:         "Route with empty name fields should generate error",
		},
		{
			name: "identical short and long names",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_long_name,route_type\n" +
					"route1,Red Line,Red Line,1",
			},
			expectedNoticeCodes: []string{"same_name_and_description"},
			description:         "Route with identical short and long names should generate warning",
		},
		{
			name: "bus route missing short name",
			files: map[string]string{
				"routes.txt": "route_id,route_long_name,route_type\n" +
					"route1,Downtown Express,3",
			},
			expectedNoticeCodes: []string{"missing_recommended_field"},
			description:         "Bus route should have short name for best practices",
		},
		{
			name: "rail route missing short name",
			files: map[string]string{
				"routes.txt": "route_id,route_long_name,route_type\n" +
					"route1,Green Line,0",
			},
			expectedNoticeCodes: []string{"missing_recommended_field"},
			description:         "Rail route should have short name",
		},
		{
			name: "rail route missing long name",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_type\n" +
					"route1,GL,1",
			},
			expectedNoticeCodes: []string{"missing_recommended_field"},
			description:         "Rail route should have long name",
		},
		{
			name: "rail route missing both names",
			files: map[string]string{
				"routes.txt": "route_id,route_type\n" +
					"route1,2",
			},
			expectedNoticeCodes: []string{"missing_route_name"},
			description:         "Rail route without both names should generate error (not just recommendations)",
		},
		{
			name: "route short name too long",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_type\n" +
					"route1,VeryLongShortName123,3",
			},
			expectedNoticeCodes: []string{"route_short_name_too_long"},
			description:         "Route short name longer than 12 characters should generate warning",
		},
		{
			name: "route long name too long",
			files: map[string]string{
				"routes.txt": "route_id,route_long_name,route_type\n" +
					"route1,This is an extremely long route name that exceeds the recommended maximum length for route long names which should be under 100 characters but this one is much longer than that,3",
			},
			expectedNoticeCodes: []string{"route_long_name_too_long"},
			description:         "Route long name longer than 100 characters should generate warning",
		},
		{
			name: "multiple route types with naming issues",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_long_name,route_type\n" +
					"route1,,Downtown Express,3\n" + // Bus missing short name
					"route2,VeryLongNameXY,Green Line,0\n" + // Rail with long short name (14 chars > 12)
					"route3,,Red Line,1\n" + // Rail missing short name
					"route4,Blue,,2", // Rail missing long name
			},
			expectedNoticeCodes: []string{
				"missing_recommended_field",
				"route_short_name_too_long",
				"missing_recommended_field",
				"missing_recommended_field",
			},
			description: "Multiple routes with different naming violations",
		},
		{
			name: "valid edge cases",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_long_name,route_type\n" +
					"route1,123456789012,Very Long Name But Still Under Limit,3\n" + // Exactly 12 chars short name
					"route2,A,B,4", // Very short names are fine
			},
			expectedNoticeCodes: []string{},
			description:         "Routes with names at the character limits should be valid",
		},
		{
			name: "route with whitespace in names",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_long_name,route_type\n" +
					"route1, Red , Blue Line ,1",
			},
			expectedNoticeCodes: []string{},
			description:         "Route names with whitespace should be trimmed and validated correctly",
		},
		{
			name: "ferry and cable car routes",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_long_name,route_type\n" +
					"route1,,Island Ferry,4\n" + // Ferry without short name
					"route2,Cable,Cable Car Line,5", // Cable car with names
			},
			expectedNoticeCodes: []string{},
			description:         "Non-rail/non-bus routes should not require specific naming patterns",
		},
		{
			name: "route with invalid route type",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_long_name,route_type\n" +
					"route1,X,Express Line,invalid",
			},
			expectedNoticeCodes: []string{},
			description:         "Invalid route type should not cause naming validation to fail",
		},
		{
			name: "mixed valid and invalid names",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_long_name,route_type\n" +
					"route1,1,Main Line,3\n" + // Valid bus route
					"route2,Red Line,Red Line,1\n" + // Identical names
					"route3,VeryLongShortName,Green Line,0", // Short name too long
			},
			expectedNoticeCodes: []string{"same_name_and_description", "route_short_name_too_long"},
			description:         "Mix of valid and invalid route names",
		},
		{
			name: "no routes file",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,http://metro.example,America/Los_Angeles",
			},
			expectedNoticeCodes: []string{},
			description:         "Missing routes.txt file should not generate errors",
		},
		{
			name: "routes with missing route_id",
			files: map[string]string{
				"routes.txt": "route_short_name,route_long_name,route_type\n" +
					"1,Main Line,3",
			},
			expectedNoticeCodes: []string{},
			description:         "Routes without route_id should be ignored by name validator",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test feed loader
			feedLoader := testutil.CreateTestFeedLoader(t, tt.files)

			// Create notice container and validator
			container := notice.NewNoticeContainer()
			validator := NewRouteNameValidator()
			config := gtfsvalidator.Config{}

			// Run validation
			validator.Validate(feedLoader, container, config)

			// Get all notices
			allNotices := container.GetNotices()

			// Extract notice codes
			var actualNoticeCodes []string
			for _, n := range allNotices {
				actualNoticeCodes = append(actualNoticeCodes, n.Code())
			}

			// Check if we got the expected notice codes
			expectedSet := make(map[string]bool)
			for _, code := range tt.expectedNoticeCodes {
				expectedSet[code] = true
			}

			actualSet := make(map[string]bool)
			for _, code := range actualNoticeCodes {
				actualSet[code] = true
			}

			// Verify expected codes are present
			for expectedCode := range expectedSet {
				if !actualSet[expectedCode] {
					t.Errorf("Expected notice code '%s' not found. Got: %v", expectedCode, actualNoticeCodes)
				}
			}

			// If no notices expected, ensure no notices were generated
			if len(tt.expectedNoticeCodes) == 0 && len(actualNoticeCodes) > 0 {
				t.Errorf("Expected no notices, but got: %v", actualNoticeCodes)
			}

			t.Logf("Test '%s': Expected %v, Got %v", tt.name, tt.expectedNoticeCodes, actualNoticeCodes)
		})
	}
}

func TestRouteNameValidator_ValidateRoute(t *testing.T) {
	validator := NewRouteNameValidator()

	tests := []struct {
		name          string
		rowData       map[string]string
		expectedCodes []string
	}{
		{
			name: "valid route with both names",
			rowData: map[string]string{
				"route_id":         "route1",
				"route_short_name": "1",
				"route_long_name":  "Main Line",
				"route_type":       "3",
			},
			expectedCodes: []string{},
		},
		{
			name: "missing both names",
			rowData: map[string]string{
				"route_id":   "route1",
				"route_type": "3",
			},
			expectedCodes: []string{"missing_route_name"},
		},
		{
			name: "identical names",
			rowData: map[string]string{
				"route_id":         "route1",
				"route_short_name": "Red",
				"route_long_name":  "Red",
				"route_type":       "1",
			},
			expectedCodes: []string{"same_name_and_description"},
		},
		{
			name: "bus missing short name",
			rowData: map[string]string{
				"route_id":        "route1",
				"route_long_name": "Downtown Express",
				"route_type":      "3",
			},
			expectedCodes: []string{"missing_recommended_field"},
		},
		{
			name: "rail missing both recommendations",
			rowData: map[string]string{
				"route_id":   "route1",
				"route_type": "0",
			},
			expectedCodes: []string{"missing_route_name"}, // Error overrides recommendations
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			container := notice.NewNoticeContainer()

			// Create mock CSV row
			row := &parser.CSVRow{
				RowNumber: 1,
				Values:    tt.rowData,
			}

			validator.validateRoute(container, row)

			notices := container.GetNotices()
			var actualCodes []string
			for _, notice := range notices {
				actualCodes = append(actualCodes, notice.Code())
			}

			// Check expected codes
			expectedSet := make(map[string]bool)
			for _, code := range tt.expectedCodes {
				expectedSet[code] = true
			}

			actualSet := make(map[string]bool)
			for _, code := range actualCodes {
				actualSet[code] = true
			}

			for expectedCode := range expectedSet {
				if !actualSet[expectedCode] {
					t.Errorf("Expected notice code '%s' not found. Got: %v", expectedCode, actualCodes)
				}
			}

			if len(tt.expectedCodes) == 0 && len(actualCodes) > 0 {
				t.Errorf("Expected no notices, but got: %v", actualCodes)
			}
		})
	}
}

func TestRouteNameValidator_NameLengthLimits(t *testing.T) {
	tests := []struct {
		shortName     string
		longName      string
		expectedShort bool // expect short name too long notice
		expectedLong  bool // expect long name too long notice
	}{
		{"123456789012", "Valid Long Name", false, false}, // Exactly 12 chars - valid
		{"1234567890123", "Valid Long Name", true, false}, // 13 chars - too long
		{"Valid", "This is exactly one hundred characters long and should be valid for route long name field testing", false, false},                                                           // Exactly 100 - valid
		{"Valid", "This is more than one hundred characters long and should be invalid for route long name field testing purposes here", false, true},                                          // Over 100 - too long
		{"VeryLongShortName", "This route name is definitely longer than one hundred characters and should trigger the route long name too long notice for comprehensive testing", true, true}, // Both too long
	}

	validator := NewRouteNameValidator()

	for i, tt := range tests {
		t.Run(fmt.Sprintf("test_%d", i), func(t *testing.T) {
			container := notice.NewNoticeContainer()

			row := &parser.CSVRow{
				RowNumber: 1,
				Values: map[string]string{
					"route_id":         "route1",
					"route_short_name": tt.shortName,
					"route_long_name":  tt.longName,
					"route_type":       "3",
				},
			}

			validator.validateRoute(container, row)

			notices := container.GetNotices()
			var codes []string
			for _, notice := range notices {
				codes = append(codes, notice.Code())
			}

			hasShortTooLong := false
			hasLongTooLong := false

			for _, code := range codes {
				if code == "route_short_name_too_long" {
					hasShortTooLong = true
				}
				if code == "route_long_name_too_long" {
					hasLongTooLong = true
				}
			}

			if hasShortTooLong != tt.expectedShort {
				t.Errorf("Short name '%s': expected too long notice %v, got %v", tt.shortName, tt.expectedShort, hasShortTooLong)
			}
			if hasLongTooLong != tt.expectedLong {
				t.Errorf("Long name '%s': expected too long notice %v, got %v", tt.longName, tt.expectedLong, hasLongTooLong)
			}
		})
	}
}

func TestRouteNameValidator_New(t *testing.T) {
	validator := NewRouteNameValidator()
	if validator == nil {
		t.Error("NewRouteNameValidator() returned nil")
	}
}
