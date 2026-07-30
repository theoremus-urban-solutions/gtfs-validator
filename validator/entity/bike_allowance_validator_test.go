package entity

import (
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/testutil"
	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

func TestBikeAllowanceValidator_Validate(t *testing.T) {
	tests := []struct {
		name                string
		files               map[string]string
		expectedNoticeCodes []string
		description         string
	}{
		{
			name: "ferry trip declaring bikes allowed",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_type\nF1,Ferry,4",
				"trips.txt":  "route_id,service_id,trip_id,bikes_allowed\nF1,S1,T1,1",
			},
			expectedNoticeCodes: []string{},
			description:         "bikes_allowed=1 answers the question",
		},
		{
			name: "ferry trip declaring bikes not allowed",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_type\nF1,Ferry,4",
				"trips.txt":  "route_id,service_id,trip_id,bikes_allowed\nF1,S1,T1,2",
			},
			expectedNoticeCodes: []string{},
			description:         "bikes_allowed=2 also answers the question",
		},
		{
			name: "ferry trip with empty bikes_allowed",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_type\nF1,Ferry,4",
				"trips.txt":  "route_id,service_id,trip_id,bikes_allowed\nF1,S1,T1,",
			},
			expectedNoticeCodes: []string{"missing_bike_allowance"},
			description:         "Empty bikes_allowed on a ferry trip should generate a notice",
		},
		{
			name: "ferry trip without bikes_allowed column",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_type\nF1,Ferry,4",
				"trips.txt":  "route_id,service_id,trip_id\nF1,S1,T1",
			},
			expectedNoticeCodes: []string{"missing_bike_allowance"},
			description:         "Absent bikes_allowed column should generate a notice",
		},
		{
			name: "ferry trip with bikes_allowed=0",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_type\nF1,Ferry,4",
				"trips.txt":  "route_id,service_id,trip_id,bikes_allowed\nF1,S1,T1,0",
			},
			expectedNoticeCodes: []string{"missing_bike_allowance"},
			description:         "bikes_allowed=0 means no information, not an answer",
		},
		{
			name: "non-ferry trip without bikes_allowed",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_type\nB1,Bus,3",
				"trips.txt":  "route_id,service_id,trip_id,bikes_allowed\nB1,S1,T1,0",
			},
			expectedNoticeCodes: []string{},
			description:         "The rule applies to ferry routes only",
		},
		{
			name: "mixed ferry and bus trips",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_type\nF1,Ferry,4\nB1,Bus,3",
				"trips.txt":  "route_id,service_id,trip_id,bikes_allowed\nF1,S1,T1,\nF1,S1,T2,1\nB1,S1,T3,\nF1,S1,T4,0",
			},
			expectedNoticeCodes: []string{"missing_bike_allowance", "missing_bike_allowance"},
			description:         "Only the ferry trips without an answer should be reported",
		},
		{
			name: "non-numeric bikes_allowed left to the type layer",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_type\nF1,Ferry,4",
				"trips.txt":  "route_id,service_id,trip_id,bikes_allowed\nF1,S1,T1,yes",
			},
			expectedNoticeCodes: []string{},
			description:         "An unparseable enum value is reported as such elsewhere, not here",
		},
		{
			name: "no ferry routes",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_type\nB1,Bus,3",
				"trips.txt":  "route_id,service_id,trip_id\nB1,S1,T1",
			},
			expectedNoticeCodes: []string{},
			description:         "A feed with no ferry route should generate no notices",
		},
		{
			name: "no routes.txt",
			files: map[string]string{
				"trips.txt": "route_id,service_id,trip_id\nF1,S1,T1",
			},
			expectedNoticeCodes: []string{},
			description:         "Missing routes.txt is reported elsewhere",
		},
		{
			name: "no trips.txt",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_type\nF1,Ferry,4",
			},
			expectedNoticeCodes: []string{},
			description:         "Missing trips.txt is reported elsewhere",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := testutil.CreateTestFeedLoader(t, tt.files)
			container := notice.NewNoticeContainer()
			v := NewBikeAllowanceValidator()

			v.Validate(loader, container, gtfsvalidator.Config{})

			expectedCounts := make(map[string]int)
			for _, code := range tt.expectedNoticeCodes {
				expectedCounts[code]++
			}
			actualCounts := make(map[string]int)
			for _, n := range container.GetNotices() {
				actualCounts[n.Code()]++
			}

			for code, expected := range expectedCounts {
				if actualCounts[code] != expected {
					t.Errorf("Expected %d notices with code '%s', got %d for %s", expected, code, actualCounts[code], tt.description)
				}
			}
			for code := range actualCounts {
				if expectedCounts[code] == 0 {
					t.Errorf("Unexpected notice code: %s for %s", code, tt.description)
				}
			}
		})
	}
}

func TestBikeAllowanceValidator_New(t *testing.T) {
	if NewBikeAllowanceValidator() == nil {
		t.Error("NewBikeAllowanceValidator() returned nil")
	}
}
