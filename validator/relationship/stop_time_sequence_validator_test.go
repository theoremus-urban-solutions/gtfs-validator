package relationship

import (
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/testutil"
	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

func TestStopTimeSequenceValidator_Validate(t *testing.T) {
	files := map[string]string{
		"trips.txt":      "route_id,service_id,trip_id\nR1,S1,T1",
		"stops.txt":      "stop_id,stop_name\nA,Stop A\nB,Stop B",
		"stop_times.txt": "trip_id,arrival_time,departure_time,stop_id,stop_sequence,shape_dist_traveled\nT1,08:00:00,08:00:00,A,1,0\nT1,08:05:00,08:05:00,B,1,0\nT1,08:10:00,08:10:00,B,2,0",
	}

	loader := testutil.CreateTestFeedLoader(t, files)
	container := notice.NewNoticeContainer()

	v := NewStopTimeSequenceValidator()
	v.Validate(loader, container, gtfsvalidator.Config{})

	codes := map[string]int{}
	for _, n := range container.GetNotices() {
		codes[n.Code()]++
	}

	// The repeated stop_sequence in this feed is reported by
	// core/duplicate_key_validator.go, which keys stop_times.txt on
	// trip_id + stop_sequence. What is left here is the distance ordering.
	if codes["decreasing_or_equal_stop_time_distance"] == 0 {
		t.Errorf("expected decreasing_or_equal_stop_time_distance for stops sharing a distance, got %+v", codes)
	}

	if codes["duplicate_stop_sequence"] != 0 {
		t.Errorf("duplicate_stop_sequence is retired, got %+v", codes)
	}
}
