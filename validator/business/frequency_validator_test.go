package business

import (
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/testutil"
	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

func TestFrequencyValidator_Validate(t *testing.T) {
	files := map[string]string{
		"trips.txt": "route_id,service_id,trip_id\nR1,S1,T1",
		// The last window is five minutes long on a ten-minute headway, so it
		// generates no trips; the middle two overlap.
		"frequencies.txt": "trip_id,start_time,end_time,headway_secs,exact_times\nT1,08:00:00,09:00:00,20,0\nT1,08:30:00,08:45:00,300,0\nT1,10:00:00,10:05:00,600,0",
	}

	loader := testutil.CreateTestFeedLoader(t, files)
	container := notice.NewNoticeContainer()

	v := NewFrequencyValidator()
	v.Validate(loader, container, gtfsvalidator.Config{})

	codes := map[string]int{}
	for _, n := range container.GetNotices() {
		codes[n.Code()]++
	}

	if codes["overlapping_frequency"] == 0 {
		t.Errorf("expected overlapping_frequency notice")
	}

	// The field types, the time range and the headway are checked by
	// core/field_type_validator.go.
	if codes["frequency_duration_shorter_than_headway"] == 0 {
		t.Errorf("expected frequency_duration_shorter_than_headway for the 15-minute window on a 600s headway")
	}
}
