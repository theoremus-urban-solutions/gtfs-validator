package business

import (
	"testing"
	"time"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator/validator"
	"github.com/theoremus-urban-solutions/gtfs-validator/testutil"
)

func TestServiceConsistencyValidator_Validate(t *testing.T) {
	files := map[string]string{
		"calendar.txt":       "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\nS1,0,0,0,0,0,0,0,20240101,20231231\nS2,1,0,0,0,0,0,0,20260101,20280101",
		"calendar_dates.txt": "service_id,date,exception_type\nS1,20200101,1\nS1,20200101,2",
		"trips.txt":          "route_id,service_id,trip_id\nR1,S3,T1\nR1,S1,T2",
	}

	loader := testutil.CreateTestFeedLoader(t, files)
	container := notice.NewNoticeContainer()

	cfg := gtfsvalidator.Config{CurrentDate: time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC)}
	v := NewServiceConsistencyValidator()
	v.Validate(loader, container, cfg)

	codes := map[string]int{}
	for _, n := range container.GetNotices() {
		codes[n.Code()]++
	}

	if codes["service_never_active"] == 0 {
		t.Errorf("expected service_never_active notice")
	}
	// Depending on ordering, either conflicting or duplicate exception may be emitted
	if codes["conflicting_calendar_exception"] == 0 && codes["duplicate_calendar_exception"] == 0 {
		t.Errorf("expected one of conflicting_calendar_exception or duplicate_calendar_exception notices")
	}
	if codes["undefined_service"] == 0 {
		t.Errorf("expected undefined_service notice for S3 used in trips only")
	}
}
