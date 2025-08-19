package relationship

import (
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/testutil"
	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

func TestRouteConsistencyValidator_Validate(t *testing.T) {
	files := map[string]string{
		"routes.txt":     "route_id,route_short_name,route_long_name,route_type,agency_id\nR1,,,3,A1\nR2,VeryLongRouteNameExceeds,Route Long,3,A1",
		"trips.txt":      "route_id,service_id,trip_id,direction_id\nR1,S1,T1,0\nR1,S1,T2,1",
		"stop_times.txt": "trip_id,arrival_time,departure_time,stop_id,stop_sequence\nT1,08:00:00,08:00:00,A,1\nT1,09:00:00,09:00:00,B,2\nT2,10:00:00,10:00:00,A,1\nT2,11:00:00,11:00:00,B,2",
	}

	loader := testutil.CreateTestFeedLoader(t, files)
	container := notice.NewNoticeContainer()

	v := NewRouteConsistencyValidator()
	v.Validate(loader, container, gtfsvalidator.Config{})

	codes := map[string]int{}
	for _, n := range container.GetNotices() {
		codes[n.Code()]++
	}

	if codes["missing_route_name"] == 0 {
		t.Errorf("expected missing_route_name notice for route without names")
	}
	if codes["route_short_name_too_long"] == 0 {
		t.Errorf("expected route_short_name_too_long notice")
	}
}
