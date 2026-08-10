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

// TestTravelSpeedValidator_MinuteResolutionGrace covers the allowance the
// canonical validator makes for timetables written to the whole minute: the hop
// below is 3.9 km, and reading its one minute literally makes it 236 km/h, well
// over the 150 the mode allows. A minute written as a minute may be anything up
// to nearly two, though, and at two it is an unremarkable 118. Only the reading
// that indicts the feed is certain enough to report, and it is not.
func TestTravelSpeedValidator_MinuteResolutionGrace(t *testing.T) {
	tests := []struct {
		name         string
		stopTimes    string
		expectNotice bool
		description  string
	}{
		{
			name: "both times on the whole minute",
			stopTimes: "trip_id,arrival_time,departure_time,stop_id,stop_sequence\n" +
				"T1,08:00:00,08:00:00,A,1\n" +
				"T1,08:01:00,08:01:00,B,2",
			expectNotice: false,
			description:  "a minute of slack brings 236 km/h down to 118",
		},
		{
			name: "arrival given to the second",
			stopTimes: "trip_id,arrival_time,departure_time,stop_id,stop_sequence\n" +
				"T1,08:00:00,08:00:00,A,1\n" +
				"T1,08:01:01,08:01:01,B,2",
			expectNotice: true,
			description:  "a feed that writes seconds means them, so the minute stands at 232 km/h",
		},
		{
			name: "times that run backwards",
			stopTimes: "trip_id,arrival_time,departure_time,stop_id,stop_sequence\n" +
				"T1,08:05:00,08:05:00,A,1\n" +
				"T1,08:04:00,08:04:00,B,2",
			expectNotice: true,
			description:  "the ground was still covered, and a minute is the shortest it took",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := testutil.CreateTestFeedLoader(t, map[string]string{
				// A and B are 3.9 km apart at the equator.
				"stops.txt":      "stop_id,stop_name,stop_lat,stop_lon\nA,Stop A,0,0\nB,Stop B,0,0.0354",
				"routes.txt":     "route_id,route_type\nR1,3",
				"trips.txt":      "route_id,service_id,trip_id\nR1,S1,T1",
				"stop_times.txt": tt.stopTimes,
			})
			container := notice.NewNoticeContainer()

			NewTravelSpeedValidator().Validate(loader, container, gtfsvalidator.Config{})

			fired := false
			for _, n := range container.GetNotices() {
				if n.Code() == "fast_travel_between_consecutive_stops" {
					fired = true
				}
			}

			if fired != tt.expectNotice {
				t.Errorf("fast_travel_between_consecutive_stops fired = %v, want %v (%s)",
					fired, tt.expectNotice, tt.description)
			}
		})
	}
}

// TestTravelSpeedValidator_LimitsFollowRouteType pins the per-mode thresholds to
// the canonical validator's. The same hop is flagged or not purely by the mode
// flying it: 90 km/h is impossible for a cable tram, brisk for a light rail, and
// unremarkable for a train.
func TestTravelSpeedValidator_LimitsFollowRouteType(t *testing.T) {
	tests := []struct {
		routeType    string
		expectNotice bool
		description  string
	}{
		{routeType: "0", expectNotice: false, description: "light rail is allowed 100 km/h"},
		{routeType: "2", expectNotice: false, description: "rail is allowed 500 km/h"},
		{routeType: "4", expectNotice: true, description: "a ferry is allowed 80 km/h"},
		{routeType: "5", expectNotice: true, description: "a cable tram is allowed 30 km/h"},
		{routeType: "7", expectNotice: true, description: "a funicular is allowed 50 km/h"},
		{routeType: "12", expectNotice: false, description: "a monorail is allowed 150 km/h"},
	}

	for _, tt := range tests {
		t.Run("route_type "+tt.routeType, func(t *testing.T) {
			loader := testutil.CreateTestFeedLoader(t, map[string]string{
				// A and B are 1.5 km apart, covered in a minute given to the
				// second so the whole-minute grace does not muddy the reading.
				"stops.txt":  "stop_id,stop_name,stop_lat,stop_lon\nA,Stop A,0,0\nB,Stop B,0,0.0135",
				"routes.txt": "route_id,route_type\nR1," + tt.routeType,
				"trips.txt":  "route_id,service_id,trip_id\nR1,S1,T1",
				"stop_times.txt": "trip_id,arrival_time,departure_time,stop_id,stop_sequence\n" +
					"T1,08:00:00,08:00:00,A,1\n" +
					"T1,08:01:01,08:01:01,B,2",
			})
			container := notice.NewNoticeContainer()

			NewTravelSpeedValidator().Validate(loader, container, gtfsvalidator.Config{})

			fired := false
			for _, n := range container.GetNotices() {
				if n.Code() == "fast_travel_between_consecutive_stops" {
					fired = true
				}
			}

			if fired != tt.expectNotice {
				t.Errorf("fast_travel_between_consecutive_stops fired = %v, want %v (89 km/h, %s)",
					fired, tt.expectNotice, tt.description)
			}
		})
	}
}

// Fixtures use rail (route_type 2), which both rules hold to the same 500 km/h.
// The far-stop rule is not a stricter limit but a wider view: what separates it
// is the 10 km of ground covered, so the cases below vary the distance and read
// only whether that rule fired, leaving the consecutive-stop rule to its own
// test.
func TestTravelSpeedValidator_FastTravelBetweenFarStops(t *testing.T) {
	tests := []struct {
		name         string
		stops        string
		stopTimes    string
		expectNotice bool
		description  string
	}{
		{
			name:  "22 km in one minute",
			stops: "stop_id,stop_name,stop_lat,stop_lon\nA,Stop A,0,0\nB,Stop B,0,0.2",
			stopTimes: "trip_id,arrival_time,departure_time,stop_id,stop_sequence\n" +
				"T1,08:00:00,08:00:00,A,1\n" +
				"T1,08:01:00,08:01:00,B,2",
			expectNotice: true,
			description:  "667 km/h over a stretch too long for one mistyped time to explain",
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
				"T1,08:00:10,08:00:10,B,2",
			expectNotice: false,
			description:  "8.9 km apart is under the threshold, however fast the hop",
		},
		{
			name:  "distance accumulated over several hops",
			stops: "stop_id,stop_name,stop_lat,stop_lon\nA,Stop A,0,0\nB,Stop B,0,0.054\nC,Stop C,0,0.108",
			stopTimes: "trip_id,arrival_time,departure_time,stop_id,stop_sequence\n" +
				"T1,08:00:00,08:00:00,A,1\n" +
				"T1,08:00:20,08:00:20,B,2\n" +
				"T1,08:00:40,08:00:40,C,3",
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
