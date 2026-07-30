package entity

import (
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/testutil"
	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

func TestMixedCaseNameValidator_Validate(t *testing.T) {
	tests := []struct {
		name        string
		files       map[string]string
		expected    int
		description string
	}{
		{
			name: "headsign in upper case only",
			files: map[string]string{
				"trips.txt": "trip_id,route_id,service_id,trip_headsign\nT1,R1,S1,DOWNTOWN LOOP",
			},
			expected:    1,
			description: "trip_headsign is what the vehicle shows the rider",
		},
		{
			name: "headsign in mixed case",
			files: map[string]string{
				"trips.txt": "trip_id,route_id,service_id,trip_headsign\nT1,R1,S1,Downtown Loop",
			},
			expected: 0,
		},
		{
			name: "Cyrillic headsign in upper case only",
			files: map[string]string{
				"trips.txt": "trip_id,route_id,service_id,trip_headsign\nT1,R1,S1,ЦЕНТРАЛНА ГАРА",
			},
			expected:    1,
			description: "Cyrillic has a case distinction to lose",
		},
		{
			name: "Cyrillic headsign in mixed case",
			files: map[string]string{
				"trips.txt": "trip_id,route_id,service_id,trip_headsign\nT1,R1,S1,Централна гара",
			},
			expected: 0,
		},
		{
			name: "headsign in a script without case",
			files: map[string]string{
				"trips.txt": "trip_id,route_id,service_id,trip_headsign\nT1,R1,S1,東京 駅行",
			},
			expected:    0,
			description: "Japanese cannot be written in mixed case",
		},
		{
			name: "single letter headsign",
			files: map[string]string{
				"trips.txt": "trip_id,route_id,service_id,trip_headsign\nT1,R1,S1,A",
			},
			expected: 0,
		},
		{
			name: "both trip fields shout",
			files: map[string]string{
				"trips.txt": "trip_id,route_id,service_id,trip_headsign,trip_short_name\nT1,R1,S1,DOWNTOWN LOOP,MORNING EXPRESS",
			},
			expected:    2,
			description: "Each field is judged on its own",
		},
		{
			name: "headsign column absent",
			files: map[string]string{
				"trips.txt": "trip_id,route_id,service_id\nT1,R1,S1",
			},
			expected: 0,
		},
		{
			name: "empty headsign",
			files: map[string]string{
				"trips.txt": "trip_id,route_id,service_id,trip_headsign\nT1,R1,S1,",
			},
			expected:    0,
			description: "A field with no value is a matter for missing_recommended_field",
		},
		{
			name: "level and pathway signage",
			files: map[string]string{
				"levels.txt":   "level_id,level_index,level_name\nL1,0,CONCOURSE LEVEL",
				"pathways.txt": "pathway_id,from_stop_id,to_stop_id,pathway_mode,is_bidirectional,signposted_as,reversed_signposted_as\nP1,S1,S2,1,1,TO THE TRAINS,TO THE STREET",
			},
			expected:    3,
			description: "Signage is read by riders standing in the station",
		},
		{
			name: "fares and flex text",
			files: map[string]string{
				"networks.txt":        "network_id,network_name\nN1,CITY NETWORK",
				"location_groups.txt": "location_group_id,location_group_name\nG1,AIRPORT ZONE",
				"booking_rules.txt":   "booking_rule_id,booking_type,message\nB1,0,CALL AHEAD TO BOOK",
			},
			expected: 3,
		},
		{
			name:        "no files at all",
			files:       map[string]string{"agency.txt": "agency_id,agency_name,agency_timezone\nA1,METRO TRANSIT,America/Los_Angeles"},
			expected:    0,
			description: "agency.txt is checked where it is already read, not here",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := testutil.CreateTestFeedLoader(t, tt.files)
			container := notice.NewNoticeContainer()
			v := NewMixedCaseNameValidator()

			v.Validate(loader, container, gtfsvalidator.Config{})

			count := 0
			for _, n := range container.GetNotices() {
				if n.Code() != "mixed_case_recommended_field" {
					t.Errorf("Unexpected notice code: %s", n.Code())
					continue
				}
				count++
			}

			if count != tt.expected {
				t.Errorf("Expected %d notices, got %d for case: %s", tt.expected, count, tt.description)
			}
		})
	}
}

func TestMixedCaseNameValidator_New(t *testing.T) {
	if NewMixedCaseNameValidator() == nil {
		t.Error("NewMixedCaseNameValidator() returned nil")
	}
}
