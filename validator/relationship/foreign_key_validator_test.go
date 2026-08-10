package relationship

import (
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/testutil"
	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

func TestForeignKeyValidator_Validate(t *testing.T) {
	files := map[string]string{
		"agency.txt":          "agency_id,agency_name,agency_url,agency_timezone\nA1,Agency,http://a,UTC",
		"stops.txt":           "stop_id,stop_name\nS1,Stop 1",
		"routes.txt":          "route_id,route_short_name,agency_id\nR1,1,A1\nR2,2,A2",
		"trips.txt":           "route_id,service_id,trip_id\nR1,SVC1,T1\nR2,SVC2,T2",
		"stop_times.txt":      "trip_id,arrival_time,departure_time,stop_id,stop_sequence\nT1,08:00:00,08:00:00,S1,1\nT2,09:00:00,09:00:00,SX,1",
		"calendar.txt":        "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\nSVC1,1,1,1,1,1,0,0,20240101,20241231",
		"calendar_dates.txt":  "service_id,date,exception_type\nSVC2,20240701,1",
		"fare_attributes.txt": "fare_id,price,currency_type\nF1,2.50,USD",
		"fare_rules.txt":      "fare_id,route_id,origin_id,destination_id,contains_id\nF1,R1,,,",
		"pathways.txt":        "pathway_id,from_stop_id,to_stop_id\nP1,S1,S2",
		"levels.txt":          "level_id,level_index,level_name\nL1,0,Ground",
		"frequencies.txt":     "trip_id,start_time,end_time,headway_secs\nT3,08:00:00,09:00:00,600",
	}

	loader := testutil.CreateTestFeedLoader(t, files)
	container := notice.NewNoticeContainer()

	v := NewForeignKeyValidator()
	v.Validate(loader, container, gtfsvalidator.Config{})

	codes := map[string]int{}
	for _, n := range container.GetNotices() {
		codes[n.Code()]++
	}

	if codes["foreign_key_violation"] == 0 {
		t.Fatalf("expected at least one foreign_key_violation notice, got 0: %+v", codes)
	}
}

// TestForeignKeyValidator_ResolvesAgainstTheDefiningFile covers the references
// that used to have validators of their own: a route's agency_id and a fare
// rule's zones. Each is run with the cache both off and on, because the two
// paths build their lookup maps from different sources and a lookup built from
// the referencing file rather than the defining one can never fail.
func TestForeignKeyValidator_ResolvesAgainstTheDefiningFile(t *testing.T) {
	tests := []struct {
		name        string
		files       map[string]string
		expected    int
		description string
	}{
		{
			name: "route naming an agency that agency.txt does not define",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\nA1,Agency,http://a,UTC",
				"routes.txt": "route_id,route_short_name,route_type,agency_id\nR1,1,3,A999",
			},
			expected:    1,
			description: "A999 appears only in routes.txt",
		},
		{
			name: "route naming an agency in the wrong case",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\nAgency1,Agency,http://a,UTC",
				"routes.txt": "route_id,route_short_name,route_type,agency_id\nR1,1,3,agency1",
			},
			expected:    1,
			description: "IDs match exactly or not at all",
		},
		{
			name: "route naming an agency that exists",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\nA1,Agency,http://a,UTC",
				"routes.txt": "route_id,route_short_name,route_type,agency_id\nR1,1,3,A1",
			},
			expected:    0,
			description: "The reference resolves",
		},
		{
			name: "calendar date for a service no trip runs",
			files: map[string]string{
				"calendar.txt": "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\n" +
					"SV1,1,1,1,1,1,0,0,20260101,20261231\n" +
					"SV2,0,0,0,0,0,1,1,20260101,20261231",
				"calendar_dates.txt": "service_id,date,exception_type\nSV2,20260704,1",
				"routes.txt":         "route_id,route_short_name,route_type\nR1,1,3",
				"trips.txt":          "route_id,service_id,trip_id\nR1,SV1,T1",
			},
			expected: 0,
			description: "SV2 is defined in calendar.txt and simply has no trips yet, " +
				"which is a warning about an unused service, not a broken reference",
		},
		{
			name: "trip naming a service the calendars do not define",
			files: map[string]string{
				"calendar.txt": "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\n" +
					"SV1,1,1,1,1,1,0,0,20260101,20261231",
				"routes.txt": "route_id,route_short_name,route_type\nR1,1,3",
				"trips.txt":  "route_id,service_id,trip_id\nR1,SV999,T1",
			},
			expected:    1,
			description: "SV999 appears only in trips.txt",
		},
		{
			name: "trip naming a shape shapes.txt does not define",
			files: map[string]string{
				"calendar.txt": "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\n" +
					"SV1,1,1,1,1,1,0,0,20260101,20261231",
				"routes.txt": "route_id,route_short_name,route_type\nR1,1,3",
				"shapes.txt": "shape_id,shape_pt_lat,shape_pt_lon,shape_pt_sequence\nSH1,0,0,1\nSH1,0,0.01,2",
				"trips.txt":  "route_id,service_id,trip_id,shape_id\nR1,SV1,T1,SH999",
			},
			expected:    1,
			description: "SH999 appears only in trips.txt",
		},
		{
			name: "fare rule naming a zone no stop defines",
			files: map[string]string{
				"stops.txt":           "stop_id,stop_name,stop_lat,stop_lon,zone_id\nS1,Stop 1,0,0,Z1",
				"fare_attributes.txt": "fare_id,price,currency_type\nF1,2.50,USD",
				"fare_rules.txt":      "fare_id,origin_id,destination_id\nF1,Z1,Z999",
			},
			expected:    1,
			description: "Z999 is referenced but never defined",
		},
	}

	for _, tt := range tests {
		for _, caching := range []bool{false, true} {
			t.Run(tt.name, func(t *testing.T) {
				loader := testutil.CreateTestFeedLoader(t, tt.files)
				if caching {
					loader.EnableCaching()
				}
				container := notice.NewNoticeContainer()

				NewForeignKeyValidator().Validate(loader, container, gtfsvalidator.Config{})

				violations := 0
				for _, n := range container.GetNotices() {
					if n.Code() == "foreign_key_violation" {
						violations++
					}
				}

				if violations != tt.expected {
					t.Errorf("caching=%v: expected %d foreign_key_violation notices, got %d (%s)",
						caching, tt.expected, violations, tt.description)
				}
			})
		}
	}
}
