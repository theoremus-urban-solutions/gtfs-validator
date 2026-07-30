package business

import (
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/testutil"
	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

func TestTripUsabilityValidator_Validate(t *testing.T) {
	tests := []struct {
		name      string
		stopTimes string
		expected  int
	}{
		{
			name:      "a trip calling at two stops",
			stopTimes: "trip_id,arrival_time,departure_time,stop_id,stop_sequence\nT1,08:00:00,08:00:00,S1,1\nT1,08:10:00,08:10:00,S2,2",
			expected:  0,
		},
		{
			name:      "a trip calling at one stop",
			stopTimes: "trip_id,arrival_time,departure_time,stop_id,stop_sequence\nT1,08:00:00,08:00:00,S1,1",
			expected:  1,
		},
		{
			name:      "one notice per unusable trip",
			stopTimes: "trip_id,arrival_time,departure_time,stop_id,stop_sequence\nT1,08:00:00,08:00:00,S1,1\nT2,09:00:00,09:00:00,S1,1\nT3,10:00:00,10:00:00,S1,1\nT3,10:10:00,10:10:00,S2,2",
			expected:  2,
		},
		{
			name:      "rows with no trip_id are not a trip",
			stopTimes: "trip_id,arrival_time,departure_time,stop_id,stop_sequence\n,08:00:00,08:00:00,S1,1\n ,08:10:00,08:10:00,S2,2",
			expected:  0,
		},
		{
			name:      "header only",
			stopTimes: "trip_id,arrival_time,departure_time,stop_id,stop_sequence",
			expected:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := testutil.CreateTestFeedLoader(t, map[string]string{
				"stop_times.txt": tt.stopTimes,
			})
			container := notice.NewNoticeContainer()

			NewTripUsabilityValidator().Validate(loader, container, gtfsvalidator.Config{})

			count := 0
			for _, n := range container.GetNotices() {
				if n.Code() == "unusable_trip" {
					count++
				}
			}
			if count != tt.expected {
				t.Errorf("unusable_trip: got %d, want %d", count, tt.expected)
			}
		})
	}
}

func TestTripUsabilityValidator_NoStopTimesFile(t *testing.T) {
	loader := testutil.CreateTestFeedLoader(t, map[string]string{
		"trips.txt": "route_id,service_id,trip_id\nR1,SV1,T1",
	})
	container := notice.NewNoticeContainer()

	NewTripUsabilityValidator().Validate(loader, container, gtfsvalidator.Config{})

	if got := len(container.GetNotices()); got != 0 {
		t.Errorf("expected no notices without stop_times.txt, got %d", got)
	}
}
