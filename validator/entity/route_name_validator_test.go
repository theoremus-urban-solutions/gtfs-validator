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
			// Canonical's own documented bad example, which its implementation
			// does not report because "Route 14" does not begin with "14". We
			// match the implementation; see longNameLeadsWithShortName.
			name: "long name repeats short name after a generic word",
			rowData: map[string]string{
				"route_id":         "route1",
				"route_short_name": "14",
				"route_long_name":  "Route 14",
				"route_type":       "3",
			},
			expectedCodes: []string{},
		},
		{
			name: "long name opens with the short name",
			rowData: map[string]string{
				"route_id":         "route1",
				"route_short_name": "14",
				"route_long_name":  "14 Express",
				"route_type":       "3",
			},
			expectedCodes: []string{"route_long_name_contains_short_name"},
		},
		{
			// Structurally canonical's own bad example "14"/"Route 14", and
			// unreported for the same reason: the check is a prefix test, so a
			// short name after a generic word goes past it. See
			// longNameLeadsWithShortName.
			name: "long name repeats short name after a generic word",
			rowData: map[string]string{
				"route_id":         "route1",
				"route_short_name": "21",
				"route_long_name":  "Линия 21",
				"route_type":       "3",
			},
			expectedCodes: []string{},
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

// TestRouteNameValidator_DocumentedExamples pins the check to the worked
// examples canonical publishes for route_long_name_contains_short_name. They,
// not any one validator run, are what this check is measured against: two of
// the three bad examples name the short name after a generic word, so a rule
// that only looked at the leading position would let them through.
// TestRouteNameValidator_LongNameLeadsWithShortName pins this check to what the
// canonical validator does, which is narrower than what it documents.
//
// Canonical publishes a containment rule with three bad examples, but tests only
// the leading position, so two of its own bad examples go unreported — by it and
// by us. Those two are pinned as false on purpose: they record a known,
// deliberate gap rather than an accident, so anyone widening this check later
// can see exactly what they are changing. See CANONICAL_PARITY.md.
func TestRouteNameValidator_LongNameLeadsWithShortName(t *testing.T) {
	tests := []struct {
		name      string
		shortName string
		longName  string
		leads     bool
	}{
		// Canonical's documented bad examples.
		{name: "bad: identical names", shortName: "604", longName: "604", leads: true},
		{name: "bad, but unreported: generic word then number", shortName: "14", longName: "Route 14", leads: false},
		{name: "bad, but unreported: generic word, number, destination", shortName: "2", longName: "Route 2: Bellows Falls In-Town", leads: false},

		// Canonical's documented good examples.
		{name: "good: letter and a destination", shortName: "N", longName: "Judah", leads: false},
		{name: "good: number and a street", shortName: "6", longName: "ML King Jr Blvd", leads: false},
		{name: "good: number and a boulevard", shortName: "55", longName: "Boulevard Saint Laurent", leads: false},
		{name: "good: number and a pair of suburbs", shortName: "1", longName: "Rangiora/Cashmere", leads: false},

		// The separators canonical accepts after the short name.
		{name: "space separator", shortName: "21", longName: "21 Clark Rd Est", leads: true},
		{name: "hyphen separator", shortName: "21", longName: "21-Clark Rd", leads: true},
		{name: "parenthesis separator", shortName: "21", longName: "21(Clark Rd)", leads: true},
		{name: "nothing after the short name", shortName: "21", longName: "21", leads: true},

		// The short name has to end where it claims to.
		{name: "longer number sharing a prefix", shortName: "21", longName: "216 Clark Rd", leads: false},
		{name: "word sharing a prefix", shortName: "Red", longName: "Redline Express", leads: false},

		// The shape this comparison started from: the same form as "Route 14",
		// and reported by neither validator.
		{name: "generic word in another script", shortName: "21", longName: "\u041b\u0438\u043d\u0438\u044f 21", leads: false},

		{name: "case insensitive", shortName: "red", longName: "RED LINE", leads: true},
		{name: "empty short name", shortName: "", longName: "Main Line", leads: false},
		{name: "short name longer than long name", shortName: "North", longName: "N", leads: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := longNameLeadsWithShortName(tt.longName, tt.shortName); got != tt.leads {
				t.Errorf("longNameLeadsWithShortName(%q, %q) = %v, want %v", tt.longName, tt.shortName, got, tt.leads)
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
