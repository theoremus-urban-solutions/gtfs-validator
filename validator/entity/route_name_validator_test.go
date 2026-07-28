package entity

import (
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/parser"
)

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
			expectedCodes: []string{"route_both_short_and_long_name_missing"},
		},
		{
			name: "identical names",
			rowData: map[string]string{
				"route_id":         "route1",
				"route_short_name": "Red",
				"route_long_name":  "Red",
				"route_type":       "1",
			},
			expectedCodes: []string{"route_long_name_contains_short_name"},
		},
		{
			name: "long name repeats short name",
			rowData: map[string]string{
				"route_id":         "route1",
				"route_short_name": "14",
				"route_long_name":  "Route 14",
				"route_type":       "3",
			},
			expectedCodes: []string{"route_long_name_contains_short_name"},
		},
		{
			name: "long name only shares a prefix with short name",
			rowData: map[string]string{
				"route_id":         "route1",
				"route_short_name": "1",
				"route_long_name":  "Route 100",
				"route_type":       "3",
			},
			expectedCodes: []string{},
		},
		{
			name: "description duplicates long name",
			rowData: map[string]string{
				"route_id":         "route1",
				"route_short_name": "1",
				"route_long_name":  "Main Line",
				"route_desc":       "main line",
				"route_type":       "3",
			},
			expectedCodes: []string{"same_name_and_description_for_route"},
		},
		{
			name: "description duplicates short name",
			rowData: map[string]string{
				"route_id":         "route1",
				"route_short_name": "N",
				"route_long_name":  "Judah",
				"route_desc":       "N",
				"route_type":       "0",
			},
			expectedCodes: []string{"same_name_and_description_for_route"},
		},
		{
			name: "informative description",
			rowData: map[string]string{
				"route_id":         "route1",
				"route_short_name": "1",
				"route_long_name":  "Main Line",
				"route_desc":       "Serves downtown via Main Street",
				"route_type":       "3",
			},
			expectedCodes: []string{},
		},
		{
			name: "long name in upper case only",
			rowData: map[string]string{
				"route_id":         "route1",
				"route_short_name": "1",
				"route_long_name":  "GALLERIA MALL",
				"route_type":       "3",
			},
			expectedCodes: []string{"mixed_case_recommended_field"},
		},
		{
			name: "long name in lower case only",
			rowData: map[string]string{
				"route_id":         "route1",
				"route_short_name": "1",
				"route_long_name":  "green line",
				"route_type":       "3",
			},
			expectedCodes: []string{"mixed_case_recommended_field"},
		},
		{
			name: "short name in upper case is not judged",
			rowData: map[string]string{
				"route_id":         "route1",
				"route_short_name": "AB",
				"route_long_name":  "Main Line",
				"route_type":       "3",
			},
			expectedCodes: []string{},
		},
		{
			name: "long name in a script without case",
			rowData: map[string]string{
				"route_id":         "route1",
				"route_short_name": "1",
				"route_long_name":  "東京駅行",
				"route_type":       "3",
			},
			expectedCodes: []string{},
		},
		{
			name: "long name too short to carry a case",
			rowData: map[string]string{
				"route_id":         "route1",
				"route_short_name": "1",
				"route_long_name":  "M4",
				"route_type":       "3",
			},
			expectedCodes: []string{},
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
			expectedCodes: []string{"route_both_short_and_long_name_missing"}, // Error overrides recommendations
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

			for actualCode := range actualSet {
				if !expectedSet[actualCode] {
					t.Errorf("Unexpected notice code '%s'. Got: %v", actualCode, actualCodes)
				}
			}
		})
	}
}

func TestRouteNameValidator_ContainsAsWord(t *testing.T) {
	tests := []struct {
		name      string
		haystack  string
		needle    string
		contained bool
	}{
		{name: "exact match", haystack: "604", needle: "604", contained: true},
		{name: "word inside a phrase", haystack: "Route 2: Bellows Falls", needle: "2", contained: true},
		{name: "case insensitive", haystack: "RED LINE", needle: "red", contained: true},
		{name: "longer number sharing a prefix", haystack: "Route 100", needle: "1", contained: false},
		{name: "word sharing a prefix", haystack: "Redline", needle: "Red", contained: false},
		{name: "needle longer than haystack", haystack: "N", needle: "North", contained: false},
		{name: "empty needle", haystack: "Main Line", needle: "", contained: false},
		{name: "punctuation delimits", haystack: "Main/Broadway", needle: "Broadway", contained: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := containsAsWord(tt.haystack, tt.needle); got != tt.contained {
				t.Errorf("containsAsWord(%q, %q) = %v, want %v", tt.haystack, tt.needle, got, tt.contained)
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
