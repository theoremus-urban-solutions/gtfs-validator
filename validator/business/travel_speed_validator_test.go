package business

import (
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/testutil"
	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

func TestTravelSpeedValidator_Validate(t *testing.T) {
	files := map[string]string{
		"stops.txt":      "stop_id,stop_name,stop_lat,stop_lon\nA,Stop A,0,0\nB,Stop B,0,1",
		"routes.txt":     "route_id,route_type\nR1,3",
		"trips.txt":      "route_id,service_id,trip_id\nR1,S1,T1",
		"stop_times.txt": "trip_id,arrival_time,departure_time,stop_id,stop_sequence\nT1,08:00:00,08:00:00,A,1\nT1,08:01:00,08:01:00,B,2",
	}

	loader := testutil.CreateTestFeedLoader(t, files)
	container := notice.NewNoticeContainer()

	v := NewTravelSpeedValidator()
	v.Validate(loader, container, gtfsvalidator.Config{})

	codes := map[string]int{}
	for _, n := range container.GetNotices() {
		codes[n.Code()]++
	}

	if codes["excessive_travel_speed"] == 0 {
		t.Errorf("expected excessive_travel_speed notice (1 deg lon in 60s is huge speed)")
	}
}
