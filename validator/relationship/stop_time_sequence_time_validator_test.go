package relationship

import (
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator/validator"
	"github.com/theoremus-urban-solutions/gtfs-validator/testutil"
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

	if codes["stop_time_arrival_after_departure"] == 0 {
		t.Errorf("expected stop_time_arrival_after_departure notice")
	}
	if codes["stop_time_decreasing_time"] == 0 {
		t.Errorf("expected stop_time_decreasing_time notice across stops")
	}
}
