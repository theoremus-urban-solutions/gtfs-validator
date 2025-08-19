package business

import (
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/testutil"
	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

func TestScheduleConsistencyValidator_Validate(t *testing.T) {
	files := map[string]string{
		"trips.txt":      "route_id,service_id,trip_id\nR1,S1,T1\nR1,S1,T2",
		"stop_times.txt": "trip_id,arrival_time,departure_time,stop_id,stop_sequence,pickup_type,drop_off_type\nT1,08:10:00,08:05:00,A,1,1,0\nT1,08:04:00,08:04:00,B,2,0,1\nT2,09:00:00,09:00:00,A,1,1,1\nT2,13:30:00,13:30:00,B,2,1,1",
	}

	loader := testutil.CreateTestFeedLoader(t, files)
	container := notice.NewNoticeContainer()

	v := NewScheduleConsistencyValidator()
	v.Validate(loader, container, gtfsvalidator.Config{})

	codes := map[string]int{}
	for _, n := range container.GetNotices() {
		codes[n.Code()]++
	}

	if codes["stop_time_arrival_after_departure"] == 0 {
		t.Errorf("expected stop_time_arrival_after_departure notice")
	}
	if codes["stop_time_decreasing_time"] == 0 {
		t.Errorf("expected stop_time_decreasing_time notice")
	}
	if codes["first_stop_no_pickup"] == 0 || codes["last_stop_no_drop_off"] == 0 {
		t.Errorf("expected first_stop_no_pickup and last_stop_no_drop_off notices")
	}
}
