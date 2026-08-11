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

	tripsWithSpurShape = "route_id,service_id,trip_id,shape_id\n" +
		"R1,S1,T1,SH2"

	// spurShape runs 2.2 km east, turns, and comes back 55 m to the south over
	// 1.7 km of the way it came before heading north. Every stop on that stretch
	// is within the 100 m tolerance of both legs, and closer to the return one,
	// which is the arrangement that used to defeat matching a stop at a time.
	spurShape = "shape_id,shape_pt_lat,shape_pt_lon,shape_pt_sequence\n" +
		"SH2,0.0005,0,1\n" +
		"SH2,0.0005,0.02,2\n" +
		"SH2,0,0.02,3\n" +
		"SH2,0,0.005,4\n" +
		"SH2,0.01,0.005,5"

	// spurStops places A at the start of the spur shape, B and C on the stretch
	// the shape covers twice, and E alone on the leg heading north.
	spurStops = "stop_id,stop_name,stop_lat,stop_lon\n" +
		"A,A,0.0005,0\n" +
		"B,B,0,0.01\n" +
		"C,C,0,0.015\n" +
		"E,E,0.008,0.005"

	tripsWithHairpinShape = "route_id,service_id,trip_id,shape_id\n" +
		"R1,S1,T1,SH3"

	// hairpinStops sits each stop on one leg of the hairpin and nearer the
	// other: P is served on the way out but lies 5.6 m from the return leg and
	// 27.8 m from the leg that serves it, and Q is the mirror of that. Reading
	// either stop on its own puts it on the wrong leg, and reading both that way
	// puts the trip in reverse.
	hairpinStops = "stop_id,stop_name,stop_lat,stop_lon\n" +
		"P,P,0.00025,0.0012\n" +
		"Q,Q,0.00005,0.0008"
)

// hairpinShape runs 222 m east, steps 33 m north, and comes straight back west
// over itself before leaving north. The two legs are close enough that a stop
// beside either is within tolerance of both, and the turn between them is short
// enough that the tolerance never lifts in between. Points are laid every 22 m,
// as a surveyed alignment's are, so the matches along it run continuously into
// one another rather than arriving as a handful of well-separated candidates.
func hairpinShape() string {
	rows := []string{"shape_id,shape_pt_lat,shape_pt_lon,shape_pt_sequence"}
	add := func(lat, lon string) {
		rows = append(rows, "SH3,"+lat+","+lon+","+strconv.Itoa(len(rows)))
	}
	for i := 0; i <= 10; i++ {
		add("0", "0."+leadingZeros(i*2))
	}
	for i := 10; i >= 0; i-- {
		add("0.0003", "0."+leadingZeros(i*2))
	}
	add("0.01", "0")
	return strings.Join(rows, "\n")
}

// leadingZeros renders n ten-thousandths as the fractional digits of a degree,
// so 2 becomes "0002" and 20 becomes "0020".
func leadingZeros(n int) string {
	s := strconv.Itoa(n)
	for len(s) < 4 {
		s = "0" + s
	}
	return s
}

// zigzagShape runs 332 m north and back, over and over, so it crosses (0,0)
// twenty-two times with 664 m of travel between one crossing and the next — far
// enough apart that each counts as a separate pass of the stop, and enough
// passes to carry the stop past the point where the match means anything.
func zigzagShape() string {
	rows := []string{"shape_id,shape_pt_lat,shape_pt_lon,shape_pt_sequence"}
	for i := 0; i < 43; i++ {
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
			name: "shape doubling back over the stops it has already served",
			files: map[string]string{
				"stops.txt":      spurStops,
				"trips.txt":      tripsWithSpurShape,
				"shapes.txt":     spurShape,
				"stop_times.txt": "trip_id,stop_id,stop_sequence\nT1,A,1\nT1,B,2\nT1,C,3\nT1,E,4",
			},
			expectedNoticeCodes: []string{},
			description: "B and C are served on the way out, and the trip reads forwards " +
				"if they are placed there, even though each sits closer to the return leg",
		},
		{
			name: "stops reversed along a shape that doubles back",
			files: map[string]string{
				"stops.txt":      spurStops,
				"trips.txt":      tripsWithSpurShape,
				"shapes.txt":     spurShape,
				"stop_times.txt": "trip_id,stop_id,stop_sequence\nT1,E,1\nT1,C,2\nT1,B,3\nT1,A,4",
			},
			expectedNoticeCodes: []string{"stops_match_shape_out_of_order"},
			description: "E is only on the last leg and A only on the first, so no way of " +
				"reading the doubled-back stretch puts this order forwards",
		},
		{
			name: "stops either side of a hairpin the shape turns inside tolerance",
			files: map[string]string{
				"stops.txt":      hairpinStops,
				"trips.txt":      tripsWithHairpinShape,
				"shapes.txt":     hairpinShape(),
				"stop_times.txt": "trip_id,stop_id,stop_sequence\nT1,P,1\nT1,Q,2",
			},
			expectedNoticeCodes: []string{},
			description: "P outbound then Q on the return reads forwards, and the turn " +
				"between the two legs must not fold them into one place on the shape",
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
			description:         "twenty-two passes leave no way to say which one serves the stop",
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

// TestShapeGeometryValidator_ReportsPatternOnce pins the grouping down: trips
// that share a shape and a stop list have identical geometry, so the finding is
// reported once for the pattern rather than once per trip. Reporting per trip
// multiplied a single misplaced stop by the size of the timetable.
func TestShapeGeometryValidator_ReportsPatternOnce(t *testing.T) {
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
	notices := 0
	for _, n := range container.GetNotices() {
		if n.Code() != "stop_too_far_from_shape" {
			continue
		}
		notices++
		if tripID, ok := n.Context()["tripId"].(string); ok {
			trips[tripID] = true
		}
	}

	// One notice for the pattern, naming the first trip that uses it, however
	// many trips share it.
	if notices != 1 {
		t.Errorf("expected the pattern to be reported once, got %d notices from %v", notices, trips)
	}
	if !trips["T1"] {
		t.Errorf("expected the notice to name T1, got %v", trips)
	}
}
