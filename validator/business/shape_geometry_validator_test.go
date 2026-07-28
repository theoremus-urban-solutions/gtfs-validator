package business

import (
	"strconv"
	"strings"
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/testutil"
	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

// Fixtures sit on the equator, where 0.001 degrees is about 111 m either way,
// so the 100 m match radius can be crossed by changing one digit.

const (
	tripsWithShape = "route_id,service_id,trip_id,shape_id\n" +
		"R1,S1,T1,SH1"

	straightShape = "shape_id,shape_pt_lat,shape_pt_lon,shape_pt_sequence\n" +
		"SH1,0,0,1\n" +
		"SH1,0,0.005,2\n" +
		"SH1,0,0.01,3"
)

// zigzagShape runs 332 m north and back seven times, so it crosses (0,0) seven
// times with 664 m of travel between one crossing and the next — far enough
// apart that each counts as a separate pass of the stop.
func zigzagShape() string {
	rows := []string{"shape_id,shape_pt_lat,shape_pt_lon,shape_pt_sequence"}
	for i := 0; i < 13; i++ {
		lat := "0"
		if i%2 == 1 {
			lat = "0.003"
		}
		rows = append(rows, "SH1,"+lat+",0,"+strconv.Itoa(i+1))
	}
	return strings.Join(rows, "\n")
}

func TestShapeGeometryValidator_Validate(t *testing.T) {
	tests := []struct {
		name                string
		files               map[string]string
		expectedNoticeCodes []string
		description         string
	}{
		{
			name: "stops sit on the shape",
			files: map[string]string{
				"stops.txt":      "stop_id,stop_name,stop_lat,stop_lon\nA,A,0,0\nB,B,0,0.01",
				"trips.txt":      tripsWithShape,
				"shapes.txt":     straightShape,
				"stop_times.txt": "trip_id,stop_id,stop_sequence\nT1,A,1\nT1,B,2",
			},
			expectedNoticeCodes: []string{},
			description:         "an alignment that passes its stops is what the rules ask for",
		},
		{
			name: "stop nowhere near the shape",
			files: map[string]string{
				"stops.txt":      "stop_id,stop_name,stop_lat,stop_lon\nA,A,0,0\nB,B,0.01,0.01",
				"trips.txt":      tripsWithShape,
				"shapes.txt":     straightShape,
				"stop_times.txt": "trip_id,stop_id,stop_sequence\nT1,A,1\nT1,B,2",
			},
			expectedNoticeCodes: []string{"stop_too_far_from_shape"},
			description:         "B is 1.1 km off the alignment, ten times the tolerance",
		},
		{
			name: "stops match the shape backwards",
			files: map[string]string{
				"stops.txt":      "stop_id,stop_name,stop_lat,stop_lon\nA,A,0,0.008\nB,B,0,0.002",
				"trips.txt":      tripsWithShape,
				"shapes.txt":     straightShape,
				"stop_times.txt": "trip_id,stop_id,stop_sequence\nT1,A,1\nT1,B,2",
			},
			expectedNoticeCodes: []string{"stops_match_shape_out_of_order"},
			description:         "both stops are on the alignment, but in the reverse order",
		},
		{
			name: "stop the shape passes over and over",
			files: map[string]string{
				"stops.txt":      "stop_id,stop_name,stop_lat,stop_lon\nA,A,0,0\nB,B,0.003,0",
				"trips.txt":      tripsWithShape,
				"shapes.txt":     zigzagShape(),
				"stop_times.txt": "trip_id,stop_id,stop_sequence\nT1,A,1\nT1,B,2",
			},
			expectedNoticeCodes: []string{"stop_has_too_many_matches_for_shape"},
			description:         "seven passes leave no way to say which one serves the stop",
		},
		{
			name: "trip travels past the end of its shape",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_lat,stop_lon\nA,A,0,0\nB,B,0,0.01",
				"trips.txt": tripsWithShape,
				"shapes.txt": "shape_id,shape_pt_lat,shape_pt_lon,shape_pt_sequence,shape_dist_traveled\n" +
					"SH1,0,0,1,0\n" +
					"SH1,0,0.01,2,1113",
				"stop_times.txt": "trip_id,stop_id,stop_sequence,shape_dist_traveled\nT1,A,1,0\nT1,B,2,1500",
			},
			expectedNoticeCodes: []string{"trip_distance_exceeds_shape_distance"},
			description:         "the last stop claims 387 units of alignment that does not exist",
		},
		{
			name: "trip overshoots its shape by a rounding error",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_lat,stop_lon\nA,A,0,0\nB,B,0,0.01",
				"trips.txt": tripsWithShape,
				"shapes.txt": "shape_id,shape_pt_lat,shape_pt_lon,shape_pt_sequence,shape_dist_traveled\n" +
					"SH1,0,0,1,0\n" +
					"SH1,0,0.01,2,1113",
				"stop_times.txt": "trip_id,stop_id,stop_sequence,shape_dist_traveled\nT1,A,1,0\nT1,B,2,1118",
			},
			expectedNoticeCodes: []string{"trip_distance_exceeds_shape_distance_below_threshold"},
			description:         "5 units is below the threshold four decimal places can express",
		},
		{
			name: "trip distances agree with the shape",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_lat,stop_lon\nA,A,0,0\nB,B,0,0.01",
				"trips.txt": tripsWithShape,
				"shapes.txt": "shape_id,shape_pt_lat,shape_pt_lon,shape_pt_sequence,shape_dist_traveled\n" +
					"SH1,0,0,1,0\n" +
					"SH1,0,0.01,2,1113",
				"stop_times.txt": "trip_id,stop_id,stop_sequence,shape_dist_traveled\nT1,A,1,0\nT1,B,2,1113",
			},
			expectedNoticeCodes: []string{},
			description:         "distances that line up with the geometry report nothing",
		},
		{
			name: "declared distance points at the wrong part of the shape",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_lat,stop_lon\nA,A,0,0\nB,B,0,0.02",
				"trips.txt": tripsWithShape,
				"shapes.txt": "shape_id,shape_pt_lat,shape_pt_lon,shape_pt_sequence,shape_dist_traveled\n" +
					"SH1,0,0,1,0\n" +
					"SH1,0,0.02,2,2226",
				"stop_times.txt": "trip_id,stop_id,stop_sequence,shape_dist_traveled\nT1,A,1,0\nT1,B,2,0",
			},
			expectedNoticeCodes: []string{"stop_too_far_from_shape_using_user_distance"},
			description:         "B's own distance places it 2.2 km from where B actually is",
		},
		{
			name: "trip without a shape",
			files: map[string]string{
				"stops.txt":      "stop_id,stop_name,stop_lat,stop_lon\nA,A,0,0\nB,B,5,5",
				"trips.txt":      "route_id,service_id,trip_id\nR1,S1,T1",
				"shapes.txt":     straightShape,
				"stop_times.txt": "trip_id,stop_id,stop_sequence\nT1,A,1\nT1,B,2",
			},
			expectedNoticeCodes: []string{},
			description:         "shape_id is optional, and there is nothing to match against without it",
		},
		{
			name: "shape_id naming a shape the feed does not have",
			files: map[string]string{
				"stops.txt":      "stop_id,stop_name,stop_lat,stop_lon\nA,A,0,0\nB,B,5,5",
				"trips.txt":      "route_id,service_id,trip_id,shape_id\nR1,S1,T1,MISSING",
				"shapes.txt":     straightShape,
				"stop_times.txt": "trip_id,stop_id,stop_sequence\nT1,A,1\nT1,B,2",
			},
			expectedNoticeCodes: []string{},
			description:         "the dangling reference is the foreign key check's to report, not this one's",
		},
		{
			name: "no shapes at all",
			files: map[string]string{
				"stops.txt":      "stop_id,stop_name,stop_lat,stop_lon\nA,A,0,0\nB,B,5,5",
				"trips.txt":      "route_id,service_id,trip_id\nR1,S1,T1",
				"stop_times.txt": "trip_id,stop_id,stop_sequence\nT1,A,1\nT1,B,2",
			},
			expectedNoticeCodes: []string{},
			description:         "shapes.txt is optional",
		},
		{
			name: "stop with no coordinates",
			files: map[string]string{
				"stops.txt":      "stop_id,stop_name,stop_lat,stop_lon\nA,A,0,0\nB,B,,",
				"trips.txt":      tripsWithShape,
				"shapes.txt":     straightShape,
				"stop_times.txt": "trip_id,stop_id,stop_sequence\nT1,A,1\nT1,B,2",
			},
			expectedNoticeCodes: []string{},
			description:         "an unplaced stop is the required-field check's to report",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := testutil.CreateTestFeedLoader(t, tt.files)
			container := notice.NewNoticeContainer()

			NewShapeGeometryValidator().Validate(loader, container, gtfsvalidator.Config{})

			var actual []string
			for _, n := range container.GetNotices() {
				actual = append(actual, n.Code())
			}

			found := make(map[string]bool, len(actual))
			for _, code := range actual {
				found[code] = true
			}

			for _, expected := range tt.expectedNoticeCodes {
				if !found[expected] {
					t.Errorf("expected notice %q, got %v (%s)", expected, actual, tt.description)
				}
			}

			if len(tt.expectedNoticeCodes) == 0 && len(actual) > 0 {
				t.Errorf("expected no notices, got %v (%s)", actual, tt.description)
			}
		})
	}
}

// TestShapeGeometryValidator_ReportsEveryTripInAPattern pins the grouping down:
// the geometry is walked once for trips that share a shape and a stop list, but
// each of those trips still has to appear in the report.
func TestShapeGeometryValidator_ReportsEveryTripInAPattern(t *testing.T) {
	loader := testutil.CreateTestFeedLoader(t, map[string]string{
		"stops.txt": "stop_id,stop_name,stop_lat,stop_lon\nA,A,0,0\nB,B,0.01,0.01",
		"trips.txt": "route_id,service_id,trip_id,shape_id\n" +
			"R1,S1,T1,SH1\n" +
			"R1,S1,T2,SH1\n" +
			"R1,S1,T3,SH1",
		"shapes.txt":     straightShape,
		"stop_times.txt": "trip_id,stop_id,stop_sequence\nT1,A,1\nT1,B,2\nT2,A,1\nT2,B,2\nT3,A,1\nT3,B,2",
	})
	container := notice.NewNoticeContainer()

	NewShapeGeometryValidator().Validate(loader, container, gtfsvalidator.Config{})

	trips := make(map[string]bool)
	for _, n := range container.GetNotices() {
		if n.Code() != "stop_too_far_from_shape" {
			continue
		}
		if tripID, ok := n.Context()["tripId"].(string); ok {
			trips[tripID] = true
		}
	}

	for _, tripID := range []string{"T1", "T2", "T3"} {
		if !trips[tripID] {
			t.Errorf("expected %s to be reported, got %v", tripID, trips)
		}
	}
}
