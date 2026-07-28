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

	// Naming is checked by entity/route_name_validator.go; this validator only
	// reports what needs the trip index. R2 has no trips.
	if codes["route_without_trips"] == 0 {
		t.Errorf("expected route_without_trips notice for R2")
	}
	for _, code := range []string{"route_both_short_and_long_name_missing", "route_short_name_too_long", "same_name_and_description"} {
		if codes[code] != 0 {
			t.Errorf("%s should come from entity/route_name_validator.go, not this validator", code)
		}
	}
}
