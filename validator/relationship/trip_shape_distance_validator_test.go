package relationship

import (
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/testutil"
	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

func TestTripShapeDistanceValidator_Validate(t *testing.T) {
	tests := []struct {
		name      string
		shapes    string
		trips     string
		stopTimes string
		expected  int
	}{
		{
			name:      "shape and stop times both measure distance",
			shapes:    "shape_id,shape_pt_lat,shape_pt_lon,shape_pt_sequence,shape_dist_traveled\nSH1,1.0,1.0,1,0\nSH1,2.0,2.0,2,100",
			trips:     "route_id,service_id,trip_id,shape_id\nR1,SV1,T1,SH1",
			stopTimes: "trip_id,arrival_time,departure_time,stop_id,stop_sequence,shape_dist_traveled\nT1,08:00:00,08:00:00,S1,1,0\nT1,08:10:00,08:10:00,S2,2,100",
			expected:  0,
		},
		{
			name:      "shape has no distances at all",
			shapes:    "shape_id,shape_pt_lat,shape_pt_lon,shape_pt_sequence\nSH1,1.0,1.0,1\nSH1,2.0,2.0,2",
			trips:     "route_id,service_id,trip_id,shape_id\nR1,SV1,T1,SH1",
			stopTimes: "trip_id,arrival_time,departure_time,stop_id,stop_sequence,shape_dist_traveled\nT1,08:00:00,08:00:00,S1,1,0\nT1,08:10:00,08:10:00,S2,2,100",
			expected:  1,
		},
		{
			name:      "shape misses a distance on one point",
			shapes:    "shape_id,shape_pt_lat,shape_pt_lon,shape_pt_sequence,shape_dist_traveled\nSH1,1.0,1.0,1,0\nSH1,2.0,2.0,2,",
			trips:     "route_id,service_id,trip_id,shape_id\nR1,SV1,T1,SH1",
			stopTimes: "trip_id,arrival_time,departure_time,stop_id,stop_sequence,shape_dist_traveled\nT1,08:00:00,08:00:00,S1,1,0",
			expected:  1,
		},
		{
			name:      "stop times supply no distances",
			shapes:    "shape_id,shape_pt_lat,shape_pt_lon,shape_pt_sequence\nSH1,1.0,1.0,1\nSH1,2.0,2.0,2",
			trips:     "route_id,service_id,trip_id,shape_id\nR1,SV1,T1,SH1",
			stopTimes: "trip_id,arrival_time,departure_time,stop_id,stop_sequence\nT1,08:00:00,08:00:00,S1,1",
			expected:  0,
		},
		{
			name:      "trip without a shape",
			shapes:    "shape_id,shape_pt_lat,shape_pt_lon,shape_pt_sequence\nSH1,1.0,1.0,1",
			trips:     "route_id,service_id,trip_id,shape_id\nR1,SV1,T1,",
			stopTimes: "trip_id,arrival_time,departure_time,stop_id,stop_sequence,shape_dist_traveled\nT1,08:00:00,08:00:00,S1,1,0",
			expected:  0,
		},
		{
			name:      "one notice per trip",
			shapes:    "shape_id,shape_pt_lat,shape_pt_lon,shape_pt_sequence\nSH1,1.0,1.0,1\nSH1,2.0,2.0,2",
			trips:     "route_id,service_id,trip_id,shape_id\nR1,SV1,T1,SH1",
			stopTimes: "trip_id,arrival_time,departure_time,stop_id,stop_sequence,shape_dist_traveled\nT1,08:00:00,08:00:00,S1,1,0\nT1,08:10:00,08:10:00,S2,2,100\nT1,08:20:00,08:20:00,S3,3,200",
			expected:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := testutil.CreateTestFeedLoader(t, map[string]string{
				"shapes.txt":     tt.shapes,
				"trips.txt":      tt.trips,
				"stop_times.txt": tt.stopTimes,
			})
			container := notice.NewNoticeContainer()

			v := NewTripShapeDistanceValidator()
			v.Validate(loader, container, gtfsvalidator.Config{})

			count := 0
			for _, n := range container.GetNotices() {
				if n.Code() == "trip_with_shape_dist_traveled_but_no_shape_distances" {
					count++
				}
			}
			if count != tt.expected {
				t.Errorf("trip_with_shape_dist_traveled_but_no_shape_distances: got %d, want %d", count, tt.expected)
			}
		})
	}
}

func TestTripShapeDistanceValidator_NoShapesFile(t *testing.T) {
	loader := testutil.CreateTestFeedLoader(t, map[string]string{
		"trips.txt":      "route_id,service_id,trip_id,shape_id\nR1,SV1,T1,SH1",
		"stop_times.txt": "trip_id,arrival_time,departure_time,stop_id,stop_sequence,shape_dist_traveled\nT1,08:00:00,08:00:00,S1,1,0",
	})
	container := notice.NewNoticeContainer()

	v := NewTripShapeDistanceValidator()
	v.Validate(loader, container, gtfsvalidator.Config{})

	if len(container.GetNotices()) != 0 {
		t.Errorf("expected no notices without shapes.txt, got %d", len(container.GetNotices()))
	}
}
