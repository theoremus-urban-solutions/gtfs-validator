package business

import (
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator/validator"
	"github.com/theoremus-urban-solutions/gtfs-validator/testutil"
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
		if n.Code() == "block_trips_overlap" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected block_trips_overlap notice")
	}
}
