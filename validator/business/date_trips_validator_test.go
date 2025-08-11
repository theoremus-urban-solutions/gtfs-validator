package business

import (
	"testing"
	"time"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator/validator"
	coretest "github.com/theoremus-urban-solutions/gtfs-validator/validator/core"
)

func TestDateTripsValidator_Validate(t *testing.T) {
	files := map[string]string{
		"calendar.txt":       "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\nS1,0,0,0,0,0,0,0,20240101,20241231",
		"calendar_dates.txt": "service_id,date,exception_type\nS2,20240101,1",
		"trips.txt":          "route_id,service_id,trip_id\nR1,S3,T1",
	}

	loader := coretest.CreateTestFeedLoader(t, files)
	container := notice.NewNoticeContainer()

	cfg := gtfsvalidator.Config{CurrentDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)}
	v := NewDateTripsValidator()
	v.Validate(loader, container, cfg)

	codes := map[string]int{}
	for _, n := range container.GetNotices() {
		codes[n.Code()]++
	}

	if codes["no_service_next_7_days"] == 0 && codes["insufficient_service_next_7_days"] == 0 {
		t.Errorf("expected either no_service_next_7_days or insufficient_service_next_7_days notice")
	}
}
