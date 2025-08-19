package relationship

import (
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator/validator"
	"github.com/theoremus-urban-solutions/gtfs-validator/testutil"
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

	if codes["duplicate_stop_sequence"] == 0 {
		t.Errorf("expected duplicate_stop_sequence notice for duplicate sequence")
	}
}
