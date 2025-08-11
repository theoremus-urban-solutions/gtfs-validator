package entity

import (
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/parser"
	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

func TestDuplicateRouteNameValidator_Validate(t *testing.T) {
	tests := []struct {
		name                string
		files               map[string]string
		expectedNoticeCodes []string
		description         string
	}{
		{
			name: "no duplicate route names",
			files: map[string]string{
				"routes.txt": "route_id,agency_id,route_short_name,route_long_name,route_type\nR1,A1,Red,Red Line,3\nR2,A1,Blue,Blue Line,3",
			},
			expectedNoticeCodes: []string{},
			description:         "Unique route names should not generate notices",
		},
		{
			name: "duplicate long names same agency and route type",
			files: map[string]string{
				"routes.txt": "route_id,agency_id,route_short_name,route_long_name,route_type\nR1,A1,Red,Metro Line,3\nR2,A1,Blue,Metro Line,3",
			},
			expectedNoticeCodes: []string{"duplicate_route_long_name"},
			description:         "Duplicate long names in same agency/type should generate notice",
		},
		{
			name: "duplicate short names same agency and route type",
			files: map[string]string{
				"routes.txt": "route_id,agency_id,route_short_name,route_long_name,route_type\nR1,A1,Red,Red Line,3\nR2,A1,Red,Blue Line,3",
			},
			expectedNoticeCodes: []string{"duplicate_route_short_name"},
			description:         "Duplicate short names in same agency/type should generate notice",
		},
		{
			name: "duplicate name combination",
			files: map[string]string{
				"routes.txt": "route_id,agency_id,route_short_name,route_long_name,route_type\nR1,A1,Red,Red Line,3\nR2,A1,Red,Red Line,3",
			},
			expectedNoticeCodes: []string{"duplicate_route_short_name", "duplicate_route_long_name", "duplicate_route_name_combination"},
			description:         "Identical name combinations should generate multiple notices",
		},
		{
			name: "same names different agencies - valid",
			files: map[string]string{
				"routes.txt": "route_id,agency_id,route_short_name,route_long_name,route_type\nR1,A1,Red,Red Line,3\nR2,A2,Red,Red Line,3",
			},
			expectedNoticeCodes: []string{},
			description:         "Same names in different agencies should be valid",
		},
		{
			name: "same names different route types - valid",
			files: map[string]string{
				"routes.txt": "route_id,agency_id,route_short_name,route_long_name,route_type\nR1,A1,Red,Red Line,3\nR2,A1,Red,Red Line,1",
			},
			expectedNoticeCodes: []string{},
			description:         "Same names with different route types should be valid",
		},
		{
			name: "case insensitive duplicate detection",
			files: map[string]string{
				"routes.txt": "route_id,agency_id,route_short_name,route_long_name,route_type\nR1,A1,red,Red Line,3\nR2,A1,RED,red line,3",
			},
			expectedNoticeCodes: []string{"duplicate_route_short_name", "duplicate_route_long_name", "duplicate_route_name_combination"},
			description:         "Case differences should still be detected as duplicates",
		},
		{
			name: "empty names ignored",
			files: map[string]string{
				"routes.txt": "route_id,agency_id,route_short_name,route_long_name,route_type\nR1,A1,,Red Line,3\nR2,A1,,Blue Line,3",
			},
			expectedNoticeCodes: []string{},
			description:         "Empty names should not generate duplicate notices",
		},
		{
			name: "whitespace trimmed in comparison",
			files: map[string]string{
				"routes.txt": "route_id,agency_id,route_short_name,route_long_name,route_type\nR1,A1, Red , Red Line ,3\nR2,A1,Red,Red Line,3",
			},
			expectedNoticeCodes: []string{"duplicate_route_short_name", "duplicate_route_long_name", "duplicate_route_name_combination"},
			description:         "Whitespace should be trimmed for comparison",
		},
		{
			name: "missing agency_id defaults to empty",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_long_name,route_type\nR1,Red,Red Line,3\nR2,Red,Blue Line,3",
			},
			expectedNoticeCodes: []string{"duplicate_route_short_name"},
			description:         "Missing agency_id should group routes together",
		},
		{
			name: "invalid route_type ignored",
			files: map[string]string{
				"routes.txt": "route_id,agency_id,route_short_name,route_long_name,route_type\nR1,A1,Red,Red Line,3\nR2,A1,Red,Blue Line,invalid",
			},
			expectedNoticeCodes: []string{},
			description:         "Routes with invalid route_type should be ignored",
		},
		{
			name: "missing route_type ignored",
			files: map[string]string{
				"routes.txt": "route_id,agency_id,route_short_name,route_long_name\nR1,A1,Red,Red Line\nR2,A1,Red,Blue Line",
			},
			expectedNoticeCodes: []string{},
			description:         "Routes without route_type should be ignored",
		},
		{
			name: "missing route_id ignored",
			files: map[string]string{
				"routes.txt": "route_id,agency_id,route_short_name,route_long_name,route_type\nR1,A1,Red,Red Line,3\n,A1,Red,Blue Line,3",
			},
			expectedNoticeCodes: []string{},
			description:         "Routes without route_id should be ignored",
		},
		{
			name: "multiple duplicates same group",
			files: map[string]string{
				"routes.txt": "route_id,agency_id,route_short_name,route_long_name,route_type\nR1,A1,Red,Red Line,3\nR2,A1,Red,Blue Line,3\nR3,A1,Red,Green Line,3",
			},
			expectedNoticeCodes: []string{"duplicate_route_short_name", "duplicate_route_short_name"},
			description:         "Multiple duplicates should generate multiple notices",
		},
		{
			name: "no routes.txt file",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\nA1,Metro,http://metro.example,America/Los_Angeles",
			},
			expectedNoticeCodes: []string{},
			description:         "Missing routes.txt should not cause errors",
		},
		{
			name: "empty routes.txt file",
			files: map[string]string{
				"routes.txt": "route_id,agency_id,route_short_name,route_long_name,route_type",
			},
			expectedNoticeCodes: []string{},
			description:         "Empty routes.txt should not generate notices",
		},
		{
			name: "single route no duplicates possible",
			files: map[string]string{
				"routes.txt": "route_id,agency_id,route_short_name,route_long_name,route_type\nR1,A1,Red,Red Line,3",
			},
			expectedNoticeCodes: []string{},
			description:         "Single route cannot have duplicates",
		},
		{
			name: "mixed valid and invalid routes",
			files: map[string]string{
				"routes.txt": "route_id,agency_id,route_short_name,route_long_name,route_type\nR1,A1,Red,Red Line,3\n,A1,Blue,Blue Line,3\nR3,A1,Red,Green Line,invalid\nR4,A1,Red,Purple Line,3",
			},
			expectedNoticeCodes: []string{"duplicate_route_short_name"},
			description:         "Only valid routes should be checked for duplicates",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := CreateTestFeedLoader(t, tt.files)
			container := notice.NewNoticeContainer()
			validator := NewDuplicateRouteNameValidator()
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

func TestDuplicateRouteNameValidator_ParseRoute(t *testing.T) {
	tests := []struct {
		name        string
		rowData     map[string]string
		expected    *RouteInfo
		description string
	}{
		{
			name: "complete route info",
			rowData: map[string]string{
				"route_id":         "R1",
				"agency_id":        "A1",
				"route_short_name": "Red",
				"route_long_name":  "Red Line",
				"route_type":       "3",
			},
			expected: &RouteInfo{
				RouteID:        "R1",
				AgencyID:       "A1",
				RouteShortName: "Red",
				RouteLongName:  "Red Line",
				RouteType:      3,
				RowNumber:      1,
			},
			description: "Complete route info should be parsed correctly",
		},
		{
			name: "missing agency_id defaults to empty",
			rowData: map[string]string{
				"route_id":         "R1",
				"route_short_name": "Red",
				"route_long_name":  "Red Line",
				"route_type":       "3",
			},
			expected: &RouteInfo{
				RouteID:        "R1",
				AgencyID:       "",
				RouteShortName: "Red",
				RouteLongName:  "Red Line",
				RouteType:      3,
				RowNumber:      1,
			},
			description: "Missing agency_id should default to empty string",
		},
		{
			name: "missing route_id returns nil",
			rowData: map[string]string{
				"agency_id":        "A1",
				"route_short_name": "Red",
				"route_long_name":  "Red Line",
				"route_type":       "3",
			},
			expected:    nil,
			description: "Missing route_id should return nil",
		},
		{
			name: "missing route_type returns nil",
			rowData: map[string]string{
				"route_id":         "R1",
				"agency_id":        "A1",
				"route_short_name": "Red",
				"route_long_name":  "Red Line",
			},
			expected:    nil,
			description: "Missing route_type should return nil",
		},
		{
			name: "invalid route_type returns nil",
			rowData: map[string]string{
				"route_id":         "R1",
				"agency_id":        "A1",
				"route_short_name": "Red",
				"route_long_name":  "Red Line",
				"route_type":       "invalid",
			},
			expected:    nil,
			description: "Invalid route_type should return nil",
		},
		{
			name: "whitespace trimmed",
			rowData: map[string]string{
				"route_id":         " R1 ",
				"agency_id":        " A1 ",
				"route_short_name": " Red ",
				"route_long_name":  " Red Line ",
				"route_type":       " 3 ",
			},
			expected: &RouteInfo{
				RouteID:        "R1",
				AgencyID:       "A1",
				RouteShortName: "Red",
				RouteLongName:  "Red Line",
				RouteType:      3,
				RowNumber:      1,
			},
			description: "Whitespace should be trimmed from all fields",
		},
		{
			name: "empty names preserved",
			rowData: map[string]string{
				"route_id":         "R1",
				"agency_id":        "A1",
				"route_short_name": "",
				"route_long_name":  "",
				"route_type":       "3",
			},
			expected: &RouteInfo{
				RouteID:        "R1",
				AgencyID:       "A1",
				RouteShortName: "",
				RouteLongName:  "",
				RouteType:      3,
				RowNumber:      1,
			},
			description: "Empty names should be preserved as empty strings",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewDuplicateRouteNameValidator()

			// Create mock row
			row := &parser.CSVRow{
				RowNumber: 1,
				Values:    tt.rowData,
			}

			result := validator.parseRoute(row)

			if tt.expected == nil {
				if result != nil {
					t.Errorf("Expected nil result, got %+v", result)
				}
				return
			}

			if result == nil {
				t.Errorf("Expected result %+v, got nil", tt.expected)
				return
			}

			if result.RouteID != tt.expected.RouteID {
				t.Errorf("Expected RouteID '%s', got '%s'", tt.expected.RouteID, result.RouteID)
			}
			if result.AgencyID != tt.expected.AgencyID {
				t.Errorf("Expected AgencyID '%s', got '%s'", tt.expected.AgencyID, result.AgencyID)
			}
			if result.RouteShortName != tt.expected.RouteShortName {
				t.Errorf("Expected RouteShortName '%s', got '%s'", tt.expected.RouteShortName, result.RouteShortName)
			}
			if result.RouteLongName != tt.expected.RouteLongName {
				t.Errorf("Expected RouteLongName '%s', got '%s'", tt.expected.RouteLongName, result.RouteLongName)
			}
			if result.RouteType != tt.expected.RouteType {
				t.Errorf("Expected RouteType %d, got %d", tt.expected.RouteType, result.RouteType)
			}
			if result.RowNumber != tt.expected.RowNumber {
				t.Errorf("Expected RowNumber %d, got %d", tt.expected.RowNumber, result.RowNumber)
			}
		})
	}
}

func TestDuplicateRouteNameValidator_New(t *testing.T) {
	validator := NewDuplicateRouteNameValidator()
	if validator == nil {
		t.Error("NewDuplicateRouteNameValidator() returned nil")
	}
}
