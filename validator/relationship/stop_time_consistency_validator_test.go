package relationship

import (
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/testutil"
	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

func TestStopTimeConsistencyValidator_Validate(t *testing.T) {
	const header = "trip_id,arrival_time,departure_time,stop_id,stop_sequence,pickup_type,drop_off_type,shape_dist_traveled\n"

	tests := []struct {
		name      string
		stopTimes string
		expected  map[string]int
	}{
		{
			name: "an ordinary trip",
			stopTimes: header +
				"T1,08:00:00,08:00:00,S1,1,0,0,0\n" +
				"T1,08:10:00,08:10:00,S2,2,0,0,100",
			expected: map[string]int{},
		},
		{
			name: "the first and last stops carry no time",
			stopTimes: header +
				"T1,,,S1,1,0,0,\n" +
				"T1,08:10:00,08:10:00,S2,2,0,0,\n" +
				"T1,,,S3,3,0,0,",
			expected: map[string]int{"missing_trip_edge": 2},
		},
		{
			name: "a stop repeated mid-trip",
			stopTimes: header +
				"T1,08:00:00,08:00:00,S1,1,0,0,\n" +
				"T1,08:10:00,08:10:00,S2,2,0,0,\n" +
				"T1,08:20:00,08:20:00,S2,3,0,0,\n" +
				"T1,08:30:00,08:30:00,S3,4,0,0,",
			expected: map[string]int{"duplicate_stop_in_trip": 1},
		},
		{
			name: "a loop trip returning to where it started",
			stopTimes: header +
				"T1,08:00:00,08:00:00,S1,1,0,0,\n" +
				"T1,08:10:00,08:10:00,S2,2,0,0,\n" +
				"T1,08:20:00,08:20:00,S1,3,0,0,",
			expected: map[string]int{},
		},
		{
			name: "no one may board at the first stop",
			stopTimes: header +
				"T1,08:00:00,08:00:00,S1,1,1,0,\n" +
				"T1,08:10:00,08:10:00,S2,2,0,0,",
			expected: map[string]int{"first_stop_no_pickup": 1},
		},
		{
			name: "no one may alight at the last stop",
			stopTimes: header +
				"T1,08:00:00,08:00:00,S1,1,0,0,\n" +
				"T1,08:10:00,08:10:00,S2,2,0,1,",
			expected: map[string]int{"last_stop_no_drop_off": 1},
		},
		{
			name: "a stop the vehicle passes without serving",
			stopTimes: header +
				"T1,08:00:00,08:00:00,S1,1,0,0,\n" +
				"T1,08:10:00,08:10:00,S2,2,1,1,\n" +
				"T1,08:20:00,08:20:00,S3,3,0,0,",
			expected: map[string]int{"stop_without_service": 1},
		},
		{
			name: "a trip nobody can board",
			stopTimes: header +
				"T1,08:00:00,08:00:00,S1,1,1,0,\n" +
				"T1,08:10:00,08:10:00,S2,2,1,0,",
			expected: map[string]int{"all_stops_no_pickup": 1, "first_stop_no_pickup": 1},
		},
		{
			name: "distances on some stops but not others",
			stopTimes: header +
				"T1,08:00:00,08:00:00,S1,1,0,0,0\n" +
				"T1,08:10:00,08:10:00,S2,2,0,0,\n" +
				"T1,08:20:00,08:20:00,S3,3,0,0,200",
			expected: map[string]int{"inconsistent_stop_time_shape_distance": 1},
		},
		{
			name: "no distances anywhere is not an inconsistency",
			stopTimes: header +
				"T1,08:00:00,08:00:00,S1,1,0,0,\n" +
				"T1,08:10:00,08:10:00,S2,2,0,0,",
			expected: map[string]int{},
		},
		{
			name: "rows out of file order are judged by stop_sequence",
			stopTimes: header +
				"T1,08:20:00,08:20:00,S3,3,0,1,\n" +
				"T1,08:00:00,08:00:00,S1,1,0,0,\n" +
				"T1,08:10:00,08:10:00,S2,2,0,0,",
			expected: map[string]int{"last_stop_no_drop_off": 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := testutil.CreateTestFeedLoader(t, map[string]string{
				"stop_times.txt": tt.stopTimes,
			})
			container := notice.NewNoticeContainer()

			NewStopTimeConsistencyValidator().Validate(loader, container, gtfsvalidator.Config{})

			counts := map[string]int{}
			for _, n := range container.GetNotices() {
				counts[n.Code()]++
			}
			for code, want := range tt.expected {
				if counts[code] != want {
					t.Errorf("%s: got %d, want %d (all: %v)", code, counts[code], want, counts)
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

func TestStopTimeConsistencyValidator_NoStopTimesFile(t *testing.T) {
	loader := testutil.CreateTestFeedLoader(t, map[string]string{
		"trips.txt": "route_id,service_id,trip_id\nR1,SV1,T1",
	})
	container := notice.NewNoticeContainer()

	NewStopTimeConsistencyValidator().Validate(loader, container, gtfsvalidator.Config{})

	if got := len(container.GetNotices()); got != 0 {
		t.Errorf("expected no notices without stop_times.txt, got %d", got)
	}
}

// Every trip is checked whether the work is spread over workers or not, so the
// two paths must agree.
func TestStopTimeConsistencyValidator_ParallelMatchesSequential(t *testing.T) {
	stopTimes := "trip_id,arrival_time,departure_time,stop_id,stop_sequence,pickup_type,drop_off_type\n"
	for _, trip := range []string{"T1", "T2", "T3", "T4", "T5", "T6", "T7", "T8", "T9", "T10", "T11", "T12"} {
		stopTimes += trip + ",08:00:00,08:00:00,S1,1,1,0\n"
		stopTimes += trip + ",08:10:00,08:10:00,S2,2,0,1\n"
	}

	count := func(workers int) map[string]int {
		loader := testutil.CreateTestFeedLoader(t, map[string]string{"stop_times.txt": stopTimes})
		container := notice.NewNoticeContainer()
		NewStopTimeConsistencyValidator().Validate(loader, container, gtfsvalidator.Config{ParallelWorkers: workers})
		counts := map[string]int{}
		for _, n := range container.GetNotices() {
			counts[n.Code()]++
		}
		return counts
	}

	sequential := count(1)
	parallel := count(4)

	if len(sequential) != len(parallel) {
		t.Fatalf("sequential %v, parallel %v", sequential, parallel)
	}
	for code, want := range sequential {
		if parallel[code] != want {
			t.Errorf("%s: sequential %d, parallel %d", code, want, parallel[code])
		}
	}
	if sequential["first_stop_no_pickup"] != 12 || sequential["last_stop_no_drop_off"] != 12 {
		t.Errorf("expected one of each per trip, got %v", sequential)
	}
}
