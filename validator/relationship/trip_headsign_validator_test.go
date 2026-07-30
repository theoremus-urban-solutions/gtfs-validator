package relationship

import (
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/testutil"
	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

func TestTripHeadsignValidator_Validate(t *testing.T) {
	stops := "stop_id,stop_name,stop_lat,stop_lon\nS1,Central,1.0,1.0\nS2,Midtown,2.0,2.0\nS3,Airport,3.0,3.0"

	tests := []struct {
		name      string
		stops     string
		trips     string
		stopTimes string
		expected  int
	}{
		{
			name:      "headsign names the last stop",
			stops:     stops,
			trips:     "route_id,service_id,trip_id,trip_headsign\nR1,SV1,T1,Airport",
			stopTimes: "trip_id,arrival_time,departure_time,stop_id,stop_sequence\nT1,08:00:00,08:00:00,S1,1\nT1,08:10:00,08:10:00,S2,2\nT1,08:20:00,08:20:00,S3,3",
			expected:  0,
		},
		{
			name:      "headsign names an intermediate stop",
			stops:     stops,
			trips:     "route_id,service_id,trip_id,trip_headsign\nR1,SV1,T1,Midtown",
			stopTimes: "trip_id,arrival_time,departure_time,stop_id,stop_sequence\nT1,08:00:00,08:00:00,S1,1\nT1,08:10:00,08:10:00,S2,2\nT1,08:20:00,08:20:00,S3,3",
			expected:  1,
		},
		{
			name:      "match is case insensitive",
			stops:     stops,
			trips:     "route_id,service_id,trip_id,trip_headsign\nR1,SV1,T1,MIDTOWN",
			stopTimes: "trip_id,arrival_time,departure_time,stop_id,stop_sequence\nT1,08:00:00,08:00:00,S1,1\nT1,08:10:00,08:10:00,S2,2\nT1,08:20:00,08:20:00,S3,3",
			expected:  1,
		},
		{
			name:      "rows out of file order still find the last stop",
			stops:     stops,
			trips:     "route_id,service_id,trip_id,trip_headsign\nR1,SV1,T1,Airport",
			stopTimes: "trip_id,arrival_time,departure_time,stop_id,stop_sequence\nT1,08:20:00,08:20:00,S3,3\nT1,08:00:00,08:00:00,S1,1\nT1,08:10:00,08:10:00,S2,2",
			expected:  0,
		},
		{
			name:      "headsign matches no stop",
			stops:     stops,
			trips:     "route_id,service_id,trip_id,trip_headsign\nR1,SV1,T1,Downtown Loop",
			stopTimes: "trip_id,arrival_time,departure_time,stop_id,stop_sequence\nT1,08:00:00,08:00:00,S1,1\nT1,08:10:00,08:10:00,S2,2",
			expected:  0,
		},
		{
			name:      "trip without a headsign",
			stops:     stops,
			trips:     "route_id,service_id,trip_id,trip_headsign\nR1,SV1,T1,",
			stopTimes: "trip_id,arrival_time,departure_time,stop_id,stop_sequence\nT1,08:00:00,08:00:00,S1,1\nT1,08:10:00,08:10:00,S2,2",
			expected:  0,
		},
		{
			// Each call at the named stop misleads the passengers boarding
			// there, so each is worth its own notice.
			name:      "one notice per call at the named stop",
			stops:     stops,
			trips:     "route_id,service_id,trip_id,trip_headsign\nR1,SV1,T1,Midtown",
			stopTimes: "trip_id,arrival_time,departure_time,stop_id,stop_sequence\nT1,08:00:00,08:00:00,S2,1\nT1,08:10:00,08:10:00,S2,2\nT1,08:20:00,08:20:00,S3,3",
			expected:  2,
		},
		{
			// Two distinct stops can share a name, and the headsign is just as
			// ambiguous at each of them.
			name: "two differently identified stops share the headsign name",
			stops: "stop_id,stop_name,stop_lat,stop_lon\n" +
				"S1,Midtown,1.0,1.0\nS2,Midtown,2.0,2.0\nS3,Airport,3.0,3.0",
			trips:     "route_id,service_id,trip_id,trip_headsign\nR1,SV1,T1,Midtown",
			stopTimes: "trip_id,arrival_time,departure_time,stop_id,stop_sequence\nT1,08:00:00,08:00:00,S1,1\nT1,08:10:00,08:10:00,S2,2\nT1,08:20:00,08:20:00,S3,3",
			expected:  2,
		},
		{
			// The final call is what the headsign promises, so only the earlier
			// one contradicts it.
			name:      "trip returning to the named terminus reports only the earlier call",
			stops:     stops,
			trips:     "route_id,service_id,trip_id,trip_headsign\nR1,SV1,T1,Midtown",
			stopTimes: "trip_id,arrival_time,departure_time,stop_id,stop_sequence\nT1,08:00:00,08:00:00,S2,1\nT1,08:10:00,08:10:00,S3,2\nT1,08:20:00,08:20:00,S2,3",
			expected:  1,
		},
		{
			name:      "single stop trip has no intermediate stop",
			stops:     stops,
			trips:     "route_id,service_id,trip_id,trip_headsign\nR1,SV1,T1,Midtown",
			stopTimes: "trip_id,arrival_time,departure_time,stop_id,stop_sequence\nT1,08:00:00,08:00:00,S2,1",
			expected:  0,
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

			v := NewTripHeadsignValidator()
			v.Validate(loader, container, gtfsvalidator.Config{})

			count := 0
			for _, n := range container.GetNotices() {
				if n.Code() == "trip_headsign_matches_intermediate_stop" {
					count++
				}
			}
			if count != tt.expected {
				t.Errorf("trip_headsign_matches_intermediate_stop: got %d, want %d", count, tt.expected)
			}
		})
	}
}
