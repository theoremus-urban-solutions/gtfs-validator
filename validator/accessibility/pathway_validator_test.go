package accessibility

import (
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/testutil"
	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

const pathwayStopsHeader = "stop_id,stop_name,location_type,parent_station,stop_access,level_id\n"
const pathwaysHeader = "pathway_id,from_stop_id,to_stop_id,pathway_mode,is_bidirectional,stair_count\n"

// station is the fixture every case starts from: one station with one
// entrance, so that a location's reachability is the case's own doing.
const stationAndEntrance = "ST,Station,1,,,\nE1,Entrance,2,ST,,\n"

// pathwayCodes is every code this validator may emit for a pathway graph.
var pathwayCodes = []string{
	"pathway_loop",
	"pathway_dangling_generic_node",
	"pathway_unreachable_location",
	"pathway_to_wrong_location_type",
	"pathway_to_platform_with_boarding_areas",
	"pathway_to_stop_with_access_outside_of_station_pathways",
	"bidirectional_exit_gate",
	"missing_level_id",
}

func TestPathwayValidator_Validate(t *testing.T) {
	tests := []struct {
		name     string
		stops    string
		pathways string
		expected map[string]int
	}{
		{
			name:     "well formed station",
			stops:    stationAndEntrance + "P1,Platform 1,0,ST,,\n",
			pathways: "PW1,E1,P1,1,1,\n",
			expected: map[string]int{},
		},
		{
			name:     "pathway starting and ending at the same location",
			stops:    stationAndEntrance + "P1,Platform 1,0,ST,,\n",
			pathways: "PW1,E1,P1,1,1,\nPW2,P1,P1,1,1,\n",
			expected: map[string]int{"pathway_loop": 1},
		},
		{
			name:     "exit gate declared bidirectional",
			stops:    stationAndEntrance + "P1,Platform 1,0,ST,,\n",
			pathways: "PW1,E1,P1,7,1,\n",
			expected: map[string]int{"bidirectional_exit_gate": 1},
		},
		{
			name:  "exit gate in one direction only",
			stops: stationAndEntrance + "P1,Platform 1,0,ST,,\n",
			// The gate lets passengers out, so the platform is reachable from
			// the entrance by the other pathway and out through the gate.
			pathways: "PW1,E1,P1,1,0,\nPW2,P1,E1,7,0,\n",
			expected: map[string]int{"bidirectional_exit_gate": 0},
		},
		{
			name:     "pathway ending at a station",
			stops:    stationAndEntrance,
			pathways: "PW1,E1,ST,1,1,\n",
			expected: map[string]int{"pathway_to_wrong_location_type": 1},
		},
		{
			name:     "pathway ending at a platform that has boarding areas",
			stops:    stationAndEntrance + "P1,Platform 1,0,ST,,\nBA1,Boarding area,4,P1,,\n",
			pathways: "PW1,E1,P1,1,1,\n",
			expected: map[string]int{"pathway_to_platform_with_boarding_areas": 1},
		},
		{
			name:     "pathway ending at a stop reached from the street",
			stops:    stationAndEntrance + "P1,Platform 1,0,ST,1,\n",
			pathways: "PW1,E1,P1,1,1,\n",
			expected: map[string]int{"pathway_to_stop_with_access_outside_of_station_pathways": 1},
		},
		{
			name:  "platform that cannot be left",
			stops: stationAndEntrance + "P1,Platform 1,0,ST,,\nP2,Platform 2,0,ST,,\n",
			// P2 is walked to from P1 but nothing leads back out of it.
			pathways: "PW1,E1,P1,1,1,\nPW2,P1,P2,1,0,\n",
			expected: map[string]int{"pathway_unreachable_location": 1},
		},
		{
			name:     "platform with no pathway from any entrance",
			stops:    stationAndEntrance + "P1,Platform 1,0,ST,,\nP2,Platform 2,0,ST,,\n",
			pathways: "PW1,E1,P1,1,1,\nPW2,P2,E1,1,0,\n",
			expected: map[string]int{"pathway_unreachable_location": 1},
		},
		{
			name:     "generic node joined to a single location",
			stops:    stationAndEntrance + "P1,Platform 1,0,ST,,\nN1,Node,3,ST,,\n",
			pathways: "PW1,E1,P1,1,1,\nPW2,P1,N1,1,1,\n",
			expected: map[string]int{"pathway_dangling_generic_node": 1},
		},
		{
			name:     "generic node joining two locations",
			stops:    stationAndEntrance + "P1,Platform 1,0,ST,,\nN1,Node,3,ST,,\n",
			pathways: "PW1,E1,N1,1,1,\nPW2,N1,P1,1,1,\n",
			expected: map[string]int{"pathway_dangling_generic_node": 0},
		},
		{
			name:     "elevator endpoint without a level",
			stops:    stationAndEntrance + "P1,Platform 1,0,ST,,L1\nP2,Platform 2,0,ST,,\n",
			pathways: "PW1,E1,P1,1,1,\nPW2,P1,P2,5,1,\n",
			expected: map[string]int{"missing_level_id": 1},
		},
		{
			name:     "elevator with a level at both ends",
			stops:    stationAndEntrance + "P1,Platform 1,0,ST,,L1\nP2,Platform 2,0,ST,,L2\n",
			pathways: "PW1,E1,P1,1,1,\nPW2,P1,P2,5,1,\n",
			expected: map[string]int{"missing_level_id": 0},
		},
		{
			name:  "pathway mode left blank",
			stops: stationAndEntrance + "P1,Platform 1,0,ST,,\n",
			// The row is still worth checking for its endpoints; the absent
			// mode is the required-field layer's to report.
			pathways: "PW1,E1,P1,,1,\nPW2,P1,P1,,1,\n",
			expected: map[string]int{"pathway_loop": 1, "bidirectional_exit_gate": 0},
		},
		{
			name:     "endpoint that is not in stops.txt",
			stops:    stationAndEntrance,
			pathways: "PW1,E1,P9,1,1,\n",
			expected: map[string]int{"foreign_key_violation": 1, "pathway_to_wrong_location_type": 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := testutil.CreateTestFeedLoader(t, map[string]string{
				"stops.txt":    pathwayStopsHeader + tt.stops,
				"pathways.txt": pathwaysHeader + tt.pathways,
			})
			container := notice.NewNoticeContainer()
			NewPathwayValidator().Validate(loader, container, gtfsvalidator.Config{})

			codes := map[string]int{}
			for _, n := range container.GetNotices() {
				codes[n.Code()]++
			}

			// A case names the codes its fixture is about; every other pathway
			// code must stay silent, so an over-eager rule fails a case that
			// was not written for it.
			for _, code := range pathwayCodes {
				if _, named := tt.expected[code]; !named {
					tt.expected[code] = 0
				}
			}
			for code, want := range tt.expected {
				if codes[code] != want {
					t.Errorf("%s: got %d, want %d", code, codes[code], want)
				}
			}

			// pathway_to_same_stop was our private name for pathway_loop.
			if codes["pathway_to_same_stop"] != 0 {
				t.Errorf("pathway_to_same_stop is retired in favour of pathway_loop")
			}
			// pathway_mode's own value is checked by core/field_type_validator.go.
			if codes["unexpected_enum_value"] != 0 {
				t.Errorf("enum values are the field type validator's business, not this one's")
			}
		})
	}
}

// TestPathwayValidator_UnreachableNamesTheDirection checks that the notice says
// which way the walk fails, since a location one can enter but not leave and
// one nobody can reach are different repairs.
func TestPathwayValidator_UnreachableNamesTheDirection(t *testing.T) {
	loader := testutil.CreateTestFeedLoader(t, map[string]string{
		"stops.txt":    pathwayStopsHeader + stationAndEntrance + "P1,Platform 1,0,ST,,\nP2,Platform 2,0,ST,,\n",
		"pathways.txt": pathwaysHeader + "PW1,E1,P1,1,1,\nPW2,P1,P2,1,0,\n",
	})
	container := notice.NewNoticeContainer()
	NewPathwayValidator().Validate(loader, container, gtfsvalidator.Config{})

	var found bool
	for _, n := range container.GetNotices() {
		if n.Code() != "pathway_unreachable_location" {
			continue
		}
		found = true
		context := n.Context()
		if context["stopId"] != "P2" {
			t.Errorf("stopId: got %v, want P2", context["stopId"])
		}
		if context["hasEntrance"] != true {
			t.Errorf("hasEntrance: got %v, want true", context["hasEntrance"])
		}
		if context["hasExit"] != false {
			t.Errorf("hasExit: got %v, want false", context["hasExit"])
		}
		if context["parentStation"] != "ST" {
			t.Errorf("parentStation: got %v, want ST", context["parentStation"])
		}
	}
	if !found {
		t.Fatal("expected pathway_unreachable_location for P2")
	}
}

// TestPathwayValidator_TraversesEachStationSeparately checks that a defect in
// one station does not follow pathways into another.
func TestPathwayValidator_TraversesEachStationSeparately(t *testing.T) {
	stops := pathwayStopsHeader +
		"ST1,Station 1,1,,,\nE1,Entrance 1,2,ST1,,\nP1,Platform 1,0,ST1,,\n" +
		"ST2,Station 2,1,,,\nP2,Platform 2,0,ST2,,\n"
	// ST2's platform has a pathway but no entrance anywhere in its station.
	pathways := pathwaysHeader + "PW1,E1,P1,1,1,\nPW2,P2,P2,1,1,\n"

	loader := testutil.CreateTestFeedLoader(t, map[string]string{
		"stops.txt":    stops,
		"pathways.txt": pathways,
	})
	container := notice.NewNoticeContainer()
	NewPathwayValidator().Validate(loader, container, gtfsvalidator.Config{})

	unreachable := map[string]bool{}
	for _, n := range container.GetNotices() {
		if n.Code() == "pathway_unreachable_location" {
			unreachable[n.Context()["stopId"].(string)] = true
		}
	}
	if !unreachable["P2"] {
		t.Error("expected P2 to be unreachable: its station has no entrance")
	}
	if unreachable["P1"] {
		t.Error("P1 is reachable from its own station's entrance")
	}
}
