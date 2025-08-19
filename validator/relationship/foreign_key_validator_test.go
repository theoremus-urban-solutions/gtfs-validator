package relationship

import (
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/testutil"
	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

func TestForeignKeyValidator_Validate(t *testing.T) {
	files := map[string]string{
		"agency.txt":          "agency_id,agency_name,agency_url,agency_timezone\nA1,Agency,http://a,UTC",
		"stops.txt":           "stop_id,stop_name\nS1,Stop 1",
		"routes.txt":          "route_id,route_short_name,agency_id\nR1,1,A1\nR2,2,A2",
		"trips.txt":           "route_id,service_id,trip_id\nR1,SVC1,T1\nR2,SVC2,T2",
		"stop_times.txt":      "trip_id,arrival_time,departure_time,stop_id,stop_sequence\nT1,08:00:00,08:00:00,S1,1\nT2,09:00:00,09:00:00,SX,1",
		"calendar.txt":        "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\nSVC1,1,1,1,1,1,0,0,20240101,20241231",
		"calendar_dates.txt":  "service_id,date,exception_type\nSVC2,20240701,1",
		"fare_attributes.txt": "fare_id,price,currency_type\nF1,2.50,USD",
		"fare_rules.txt":      "fare_id,route_id,origin_id,destination_id,contains_id\nF1,R1,,,",
		"pathways.txt":        "pathway_id,from_stop_id,to_stop_id\nP1,S1,S2",
		"levels.txt":          "level_id,level_index,level_name\nL1,0,Ground",
		"frequencies.txt":     "trip_id,start_time,end_time,headway_secs\nT3,08:00:00,09:00:00,600",
	}

	loader := testutil.CreateTestFeedLoader(t, files)
	container := notice.NewNoticeContainer()

	v := NewForeignKeyValidator()
	v.Validate(loader, container, gtfsvalidator.Config{})

	codes := map[string]int{}
	for _, n := range container.GetNotices() {
		codes[n.Code()]++
	}

	if codes["foreign_key_violation"] == 0 {
		t.Fatalf("expected at least one foreign_key_violation notice, got 0: %+v", codes)
	}
}
