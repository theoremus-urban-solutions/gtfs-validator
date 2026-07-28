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

			if len(tt.expectedCodes) == 0 && len(actualCodes) > 0 {
				t.Errorf("Expected no notices, but got: %v", actualCodes)
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
