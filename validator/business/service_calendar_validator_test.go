package business

import (
	"testing"
	"time"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator/validator"
	"github.com/theoremus-urban-solutions/gtfs-validator/testutil"
)

func TestServiceCalendarValidator_Validate(t *testing.T) {
	files := map[string]string{
		"calendar.txt":       "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\nWKD,1,1,1,1,1,0,0,20240101,20231231\nSAT,0,0,0,0,0,1,0,20240101,20241231\nFUT,0,0,0,0,1,0,0,20270101,20280101",
		"calendar_dates.txt": "service_id,date,exception_type\nWKD,20240115,3\nWKD,20240115,1",
	}

	loader := testutil.CreateTestFeedLoader(t, files)
	container := notice.NewNoticeContainer()

	cfg := gtfsvalidator.Config{CurrentDate: time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC)}
	v := NewServiceCalendarValidator()
	v.Validate(loader, container, cfg)

	codes := map[string]int{}
	for _, n := range container.GetNotices() {
		codes[n.Code()]++
	}

	if codes["calendar_end_before_start"] == 0 {
		t.Errorf("expected calendar_end_before_start for invalid WKD dates")
	}
	if codes["invalid_exception_type"] == 0 {
		t.Errorf("expected invalid_exception_type for exception_type=3")
	}
	if codes["weekend_only_service"] == 0 {
		t.Errorf("expected weekend_only_service for SAT service")
	}
}
