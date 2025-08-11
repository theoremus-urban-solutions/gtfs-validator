package relationship

import (
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/parser"
)

func TestAttributionValidator_ParseAttribution(t *testing.T) {
	validator := &AttributionValidator{}

	tests := []struct {
		name     string
		row      *parser.CSVRow
		expected *AttributionInfo
	}{
		{
			name: "full attribution record",
			row: &parser.CSVRow{
				Values: map[string]string{
					"attribution_id":    "attr_1",
					"agency_id":         "agency_1",
					"organization_name": "Transit Data Co",
					"is_producer":       "1",
					"is_operator":       "0",
					"is_authority":      "0",
					"attribution_url":   "https://example.com",
					"attribution_email": "contact@example.com",
					"attribution_phone": "+1234567890",
				},
				RowNumber: 2,
			},
			expected: &AttributionInfo{
				AttributionID:    "attr_1",
				AgencyID:         "agency_1",
				OrganizationName: "Transit Data Co",
				IsProducer:       boolPtr(true),
				IsOperator:       boolPtr(false),
				IsAuthority:      boolPtr(false),
				AttributionURL:   "https://example.com",
				AttributionEmail: "contact@example.com",
				AttributionPhone: "+1234567890",
				RowNumber:        2,
			},
		},
		{
			name: "attribution with route reference",
			row: &parser.CSVRow{
				Values: map[string]string{
					"attribution_id":    "attr_1",
					"route_id":          "route_1",
					"organization_name": "Transit Data Co",
					"is_producer":       "1",
				},
				RowNumber: 2,
			},
			expected: &AttributionInfo{
				AttributionID:    "attr_1",
				RouteID:          "route_1",
				OrganizationName: "Transit Data Co",
				IsProducer:       boolPtr(true),
				RowNumber:        2,
			},
		},
		{
			name: "attribution with whitespace",
			row: &parser.CSVRow{
				Values: map[string]string{
					"attribution_id":    "  attr_1  ",
					"agency_id":         "  agency_1  ",
					"organization_name": "  Transit Data Co  ",
					"is_producer":       " 1 ",
				},
				RowNumber: 2,
			},
			expected: &AttributionInfo{
				AttributionID:    "attr_1",
				AgencyID:         "agency_1",
				OrganizationName: "Transit Data Co",
				IsProducer:       boolPtr(true),
				RowNumber:        2,
			},
		},
		{
			name: "attribution with invalid boolean",
			row: &parser.CSVRow{
				Values: map[string]string{
					"attribution_id":    "attr_1",
					"agency_id":         "agency_1",
					"organization_name": "Transit Data Co",
					"is_producer":       "invalid",
				},
				RowNumber: 2,
			},
			expected: &AttributionInfo{
				AttributionID:    "attr_1",
				AgencyID:         "agency_1",
				OrganizationName: "Transit Data Co",
				RowNumber:        2,
			},
		},
		{
			name: "minimal attribution",
			row: &parser.CSVRow{
				Values: map[string]string{
					"organization_name": "Transit Data Co",
				},
				RowNumber: 2,
			},
			expected: &AttributionInfo{
				OrganizationName: "Transit Data Co",
				RowNumber:        2,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.parseAttribution(tt.row)

			if result == nil {
				t.Errorf("Expected %+v, got nil", tt.expected)
				return
			}

			// Compare fields
			if result.AttributionID != tt.expected.AttributionID ||
				result.AgencyID != tt.expected.AgencyID ||
				result.RouteID != tt.expected.RouteID ||
				result.TripID != tt.expected.TripID ||
				result.OrganizationName != tt.expected.OrganizationName ||
				result.AttributionURL != tt.expected.AttributionURL ||
				result.AttributionEmail != tt.expected.AttributionEmail ||
				result.AttributionPhone != tt.expected.AttributionPhone ||
				result.RowNumber != tt.expected.RowNumber {
				t.Errorf("Expected %+v, got %+v", tt.expected, result)
			}

			// Compare boolean pointers
			if !compareBoolPtr(result.IsProducer, tt.expected.IsProducer) ||
				!compareBoolPtr(result.IsOperator, tt.expected.IsOperator) ||
				!compareBoolPtr(result.IsAuthority, tt.expected.IsAuthority) {
				t.Errorf("Boolean values mismatch: Expected %+v, got %+v", tt.expected, result)
			}
		})
	}
}

// Helper functions
func boolPtr(b bool) *bool {
	return &b
}

func compareBoolPtr(a, b *bool) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}
