package business

import (
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/testutil"
	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

func TestBlockOverlappingValidator_Validate(t *testing.T) {
	files := map[string]string{
		"trips.txt":      "route_id,service_id,trip_id,block_id\nR1,S1,T1,B1\nR1,S1,T2,B1",
		"stop_times.txt": "trip_id,arrival_time,departure_time,stop_id,stop_sequence\nT1,08:00:00,08:00:00,A,1\nT1,10:00:00,10:00:00,B,2\nT2,09:00:00,09:00:00,C,1\nT2,11:00:00,11:00:00,D,2",
	}

	loader := testutil.CreateTestFeedLoader(t, files)
	container := notice.NewNoticeContainer()

	v := NewBlockOverlappingValidator()
	v.Validate(loader, container, gtfsvalidator.Config{})

	found := false
	for _, n := range container.GetNotices() {
		if n.Code() == "block_trips_with_overlapping_stop_times" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected block_trips_overlap notice")
	}
}

func TestBlockOverlappingValidator_RouteTypeConsistency(t *testing.T) {
	stopTimes := "trip_id,arrival_time,departure_time,stop_id,stop_sequence\n" +
		"T1,08:00:00,08:00:00,A,1\nT1,09:00:00,09:00:00,B,2\n" +
		"T2,10:00:00,10:00:00,C,1\nT2,11:00:00,11:00:00,D,2"

	tests := []struct {
		name     string
		routes   string
		trips    string
		expected int
	}{
		{
			name:     "block on one mode",
			routes:   "route_id,route_short_name,route_type\nR1,1,3\nR2,2,3",
			trips:    "route_id,service_id,trip_id,block_id\nR1,S1,T1,B1\nR2,S1,T2,B1",
			expected: 0,
		},
		{
			name:     "block spanning bus and tram",
			routes:   "route_id,route_short_name,route_type\nR1,1,3\nR2,2,0",
			trips:    "route_id,service_id,trip_id,block_id\nR1,S1,T1,B1\nR2,S1,T2,B1",
			expected: 1,
		},
		{
			name:     "different modes in different blocks",
			routes:   "route_id,route_short_name,route_type\nR1,1,3\nR2,2,0",
			trips:    "route_id,service_id,trip_id,block_id\nR1,S1,T1,B1\nR2,S1,T2,B2",
			expected: 0,
		},
		{
			name:     "route_type missing from routes.txt",
			routes:   "route_id,route_short_name\nR1,1\nR2,2",
			trips:    "route_id,service_id,trip_id,block_id\nR1,S1,T1,B1\nR2,S1,T2,B1",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := testutil.CreateTestFeedLoader(t, map[string]string{
				"routes.txt":     tt.routes,
				"trips.txt":      tt.trips,
				"stop_times.txt": stopTimes,
			})
			container := notice.NewNoticeContainer()

			v := NewBlockOverlappingValidator()
			v.Validate(loader, container, gtfsvalidator.Config{})

			count := 0
			for _, n := range container.GetNotices() {
				if n.Code() == "inconsistent_route_type_for_block_id" {
					count++
				}
			}
			if count != tt.expected {
				t.Errorf("inconsistent_route_type_for_block_id: got %d, want %d", count, tt.expected)
			}
		})
	}
}
