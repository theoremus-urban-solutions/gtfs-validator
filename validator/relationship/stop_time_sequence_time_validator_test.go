package relationship

import (
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/testutil"
	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

func TestStopTimeSequenceTimeValidator_Validate(t *testing.T) {
	files := map[string]string{
		"trips.txt":      "route_id,service_id,trip_id\nR1,S1,T1",
		"stop_times.txt": "trip_id,arrival_time,departure_time,stop_id,stop_sequence\nT1,08:10:00,08:05:00,S1,1\nT1,08:04:00,08:04:00,S2,2",
	}

	loader := testutil.CreateTestFeedLoader(t, files)
	container := notice.NewNoticeContainer()

	v := NewStopTimeSequenceTimeValidator()
	v.Validate(loader, container, gtfsvalidator.Config{})

	codes := map[string]int{}
	for _, n := range container.GetNotices() {
		codes[n.Code()]++
	}

	// A row whose arrival is after its own departure is the canonical
	// start_and_end_range_out_of_order, reported by the field type validator
	// which owns every start/end pair. This validator owns only the ordering
	// across stops.
	if codes["stop_time_with_arrival_before_previous_departure_time"] == 0 {
		t.Errorf("expected stop_time_decreasing_time notice across stops")
	}
}
