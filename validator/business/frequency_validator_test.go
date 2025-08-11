package business

import (
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator/validator"
	coretest "github.com/theoremus-urban-solutions/gtfs-validator/validator/core"
)

func TestFrequencyValidator_Validate(t *testing.T) {
	files := map[string]string{
		"trips.txt":       "route_id,service_id,trip_id\nR1,S1,T1",
		"frequencies.txt": "trip_id,start_time,end_time,headway_secs,exact_times\nT1,08:00:00,07:00:00,-10,2\nT1,08:00:00,09:00:00,20,0\nT1,08:30:00,08:45:00,600,0",
	}

	loader := coretest.CreateTestFeedLoader(t, files)
	container := notice.NewNoticeContainer()

	v := NewFrequencyValidator()
	v.Validate(loader, container, gtfsvalidator.Config{})

	codes := map[string]int{}
	for _, n := range container.GetNotices() {
		codes[n.Code()]++
	}

	if codes["invalid_frequency_time_range"] == 0 {
		t.Errorf("expected invalid_frequency_time_range notice")
	}
	if codes["invalid_headway"] == 0 {
		t.Errorf("expected invalid_headway notice")
	}
	if codes["unreasonable_headway"] == 0 {
		t.Errorf("expected unreasonable_headway notice for 20 seconds")
	}
	if codes["overlapping_frequency"] == 0 {
		t.Errorf("expected overlapping_frequency notice")
	}
	if codes["invalid_exact_times"] == 0 {
		t.Errorf("expected invalid_exact_times notice")
	}
}
