package business

import (
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/testutil"
	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

func TestTravelSpeedValidator_Validate(t *testing.T) {
	files := map[string]string{
		"stops.txt":      "stop_id,stop_name,stop_lat,stop_lon\nA,Stop A,0,0\nB,Stop B,0,1",
		"routes.txt":     "route_id,route_type\nR1,3",
		"trips.txt":      "route_id,service_id,trip_id\nR1,S1,T1",
		"stop_times.txt": "trip_id,arrival_time,departure_time,stop_id,stop_sequence\nT1,08:00:00,08:00:00,A,1\nT1,08:01:00,08:01:00,B,2",
	}

	loader := testutil.CreateTestFeedLoader(t, files)
	container := notice.NewNoticeContainer()

	v := NewTravelSpeedValidator()
	v.Validate(loader, container, gtfsvalidator.Config{})

	codes := map[string]int{}
	for _, n := range container.GetNotices() {
		codes[n.Code()]++
	}

	if codes["fast_travel_between_consecutive_stops"] == 0 {
		t.Errorf("expected excessive_travel_speed notice (1 deg lon in 60s is huge speed)")
	}
}

// Fixtures use rail (route_type 2), whose 500 km/h consecutive-stop limit is
// well clear of the flat 200 km/h the far-stop rule applies, so each case
// exercises one rule without tripping the other.
func TestTravelSpeedValidator_FastTravelBetweenFarStops(t *testing.T) {
	tests := []struct {
		name         string
		stops        string
		stopTimes    string
		expectNotice bool
		description  string
	}{
		{
			name:  "22 km in five minutes",
			stops: "stop_id,stop_name,stop_lat,stop_lon\nA,Stop A,0,0\nB,Stop B,0,0.2",
			stopTimes: "trip_id,arrival_time,departure_time,stop_id,stop_sequence\n" +
				"T1,08:00:00,08:00:00,A,1\n" +
				"T1,08:05:00,08:05:00,B,2",
			expectNotice: true,
			description:  "267 km/h over a stretch too long for one mistyped time to explain",
		},
		{
			name:  "22 km in twenty minutes",
			stops: "stop_id,stop_name,stop_lat,stop_lon\nA,Stop A,0,0\nB,Stop B,0,0.2",
			stopTimes: "trip_id,arrival_time,departure_time,stop_id,stop_sequence\n" +
				"T1,08:00:00,08:00:00,A,1\n" +
				"T1,08:20:00,08:20:00,B,2",
			expectNotice: false,
			description:  "67 km/h is ordinary regional rail",
		},
		{
			name:  "10 km is not yet far",
			stops: "stop_id,stop_name,stop_lat,stop_lon\nA,Stop A,0,0\nB,Stop B,0,0.08",
			stopTimes: "trip_id,arrival_time,departure_time,stop_id,stop_sequence\n" +
				"T1,08:00:00,08:00:00,A,1\n" +
				"T1,08:01:00,08:01:00,B,2",
			expectNotice: false,
			description:  "8.9 km apart is under the threshold, however fast the hop",
		},
		{
			name:  "distance accumulated over several hops",
			stops: "stop_id,stop_name,stop_lat,stop_lon\nA,Stop A,0,0\nB,Stop B,0,0.054\nC,Stop C,0,0.108",
			stopTimes: "trip_id,arrival_time,departure_time,stop_id,stop_sequence\n" +
				"T1,08:00:00,08:00:00,A,1\n" +
				"T1,08:01:30,08:01:30,B,2\n" +
				"T1,08:03:00,08:03:00,C,3",
			expectNotice: true,
			description:  "neither hop reaches 10 km on its own, but A to C does",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := testutil.CreateTestFeedLoader(t, map[string]string{
				"stops.txt":      tt.stops,
				"routes.txt":     "route_id,route_type\nR1,2",
				"trips.txt":      "route_id,service_id,trip_id\nR1,S1,T1",
				"stop_times.txt": tt.stopTimes,
			})
			container := notice.NewNoticeContainer()

			NewTravelSpeedValidator().Validate(loader, container, gtfsvalidator.Config{})

			var actual []string
			fired := false
			for _, n := range container.GetNotices() {
				actual = append(actual, n.Code())
				if n.Code() == "fast_travel_between_far_stops" {
					fired = true
				}
			}

			if fired != tt.expectNotice {
				t.Errorf("fast_travel_between_far_stops fired = %v, want %v; notices: %v (%s)",
					fired, tt.expectNotice, actual, tt.description)
			}
		})
	}
}
