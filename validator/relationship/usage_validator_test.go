package relationship

import (
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/testutil"
	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

func TestUsageValidator_Validate(t *testing.T) {
	tests := []struct {
		name      string
		stops     string
		trips     string
		stopTimes string
		expected  map[string]int
	}{
		{
			name:      "everything is referred to",
			stops:     "stop_id,stop_name,stop_lat,stop_lon,location_type,parent_station\nS1,First,1.0,1.0,0,\nS2,Second,2.0,2.0,0,",
			trips:     "route_id,service_id,trip_id\nR1,SV1,T1",
			stopTimes: "trip_id,arrival_time,departure_time,stop_id,stop_sequence\nT1,08:00:00,08:00:00,S1,1\nT1,08:10:00,08:10:00,S2,2",
			expected:  map[string]int{},
		},
		{
			name:      "a stop no trip calls at",
			stops:     "stop_id,stop_name,stop_lat,stop_lon,location_type,parent_station\nS1,First,1.0,1.0,0,\nS2,Orphan,2.0,2.0,0,",
			trips:     "route_id,service_id,trip_id\nR1,SV1,T1",
			stopTimes: "trip_id,arrival_time,departure_time,stop_id,stop_sequence\nT1,08:00:00,08:00:00,S1,1",
			expected:  map[string]int{"stop_without_stop_time": 1},
		},
		{
			name:      "only stops and platforms are expected in stop_times",
			stops:     "stop_id,stop_name,stop_lat,stop_lon,location_type,parent_station\nS1,First,1.0,1.0,0,\nE1,Entrance,2.0,2.0,2,ST1\nN1,Node,2.0,2.0,3,ST1\nST1,Station,2.0,2.0,1,",
			trips:     "route_id,service_id,trip_id\nR1,SV1,T1",
			stopTimes: "trip_id,arrival_time,departure_time,stop_id,stop_sequence\nT1,08:00:00,08:00:00,S1,1",
			expected:  map[string]int{},
		},
		{
			name:      "a trip with no stop times",
			stops:     "stop_id,stop_name,stop_lat,stop_lon,location_type,parent_station\nS1,First,1.0,1.0,0,",
			trips:     "route_id,service_id,trip_id\nR1,SV1,T1\nR1,SV1,T2",
			stopTimes: "trip_id,arrival_time,departure_time,stop_id,stop_sequence\nT1,08:00:00,08:00:00,S1,1",
			expected:  map[string]int{"unused_trip": 1},
		},
		{
			name:      "a station nothing is a child of",
			stops:     "stop_id,stop_name,stop_lat,stop_lon,location_type,parent_station\nS1,First,1.0,1.0,0,ST1\nST1,Served,1.0,1.0,1,\nST2,Empty,3.0,3.0,1,",
			trips:     "route_id,service_id,trip_id\nR1,SV1,T1",
			stopTimes: "trip_id,arrival_time,departure_time,stop_id,stop_sequence\nT1,08:00:00,08:00:00,S1,1",
			expected:  map[string]int{"unused_station": 1},
		},
		{
			name:      "a stop referred to only by a trip that has no stop times is still unused",
			stops:     "stop_id,stop_name,stop_lat,stop_lon,location_type,parent_station\nS1,Orphan,1.0,1.0,0,",
			trips:     "route_id,service_id,trip_id\nR1,SV1,T1",
			stopTimes: "trip_id,arrival_time,departure_time,stop_id,stop_sequence",
			expected:  map[string]int{"stop_without_stop_time": 1, "unused_trip": 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := testutil.CreateTestFeedLoader(t, map[string]string{
				"stops.txt":      tt.stops,
				"trips.txt":      tt.trips,
				"stop_times.txt": tt.stopTimes,
			})
			container := notice.NewNoticeContainer()

			NewUsageValidator().Validate(loader, container, gtfsvalidator.Config{})

			counts := map[string]int{}
			for _, n := range container.GetNotices() {
				counts[n.Code()]++
			}
			for code, want := range tt.expected {
				if counts[code] != want {
					t.Errorf("%s: got %d, want %d", code, counts[code], want)
				}
			}
			for code, got := range counts {
				if _, wanted := tt.expected[code]; !wanted {
					t.Errorf("unexpected %s: got %d", code, got)
				}
			}
		})
	}
}

func TestUsageValidator_MissingFiles(t *testing.T) {
	loader := testutil.CreateTestFeedLoader(t, map[string]string{
		"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,http://metro.example,Europe/Sofia",
	})
	container := notice.NewNoticeContainer()

	NewUsageValidator().Validate(loader, container, gtfsvalidator.Config{})

	if got := len(container.GetNotices()); got != 0 {
		t.Errorf("expected no notices without stops or trips, got %d", got)
	}
}

func TestUsageValidator_NoStopTimesFile(t *testing.T) {
	loader := testutil.CreateTestFeedLoader(t, map[string]string{
		"stops.txt": "stop_id,stop_name,stop_lat,stop_lon,location_type,parent_station\nS1,First,1.0,1.0,0,",
		"trips.txt": "route_id,service_id,trip_id\nR1,SV1,T1",
	})
	container := notice.NewNoticeContainer()

	NewUsageValidator().Validate(loader, container, gtfsvalidator.Config{})

	counts := map[string]int{}
	for _, n := range container.GetNotices() {
		counts[n.Code()]++
	}
	// Without the file nothing is referred to, so both the stop and the trip
	// are reported rather than silently passed.
	if counts["stop_without_stop_time"] != 1 || counts["unused_trip"] != 1 {
		t.Errorf("got %v, want one stop_without_stop_time and one unused_trip", counts)
	}
}
