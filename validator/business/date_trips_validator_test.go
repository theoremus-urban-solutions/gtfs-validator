package business

import (
	"testing"
	"time"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/testutil"
	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

func TestDateTripsValidator_Validate(t *testing.T) {
	currentDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name        string
		files       map[string]string
		wantNotice  bool
		description string
	}{
		{
			name: "service runs every day of the week",
			files: map[string]string{
				"calendar.txt": "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\n" +
					"S1,1,1,1,1,1,1,1,20240101,20241231",
				"trips.txt": "route_id,service_id,trip_id\nR1,S1,T1",
			},
			wantNotice:  false,
			description: "Coverage through the whole week is what the rule asks for",
		},
		{
			name: "no day of the week is active",
			files: map[string]string{
				"calendar.txt": "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\n" +
					"S1,0,0,0,0,0,0,0,20240101,20241231",
				"trips.txt": "route_id,service_id,trip_id\nR1,S1,T1",
			},
			wantNotice:  true,
			description: "A service with no active days runs no trips at all",
		},
		{
			name: "service ends mid-week",
			files: map[string]string{
				"calendar.txt": "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\n" +
					"S1,1,1,1,1,1,1,1,20240101,20240103",
				"trips.txt": "route_id,service_id,trip_id\nR1,S1,T1",
			},
			wantNotice:  true,
			description: "Coverage that runs out inside the week is the defect this reports",
		},
		{
			name: "service defined but no trips reference it",
			files: map[string]string{
				"calendar.txt": "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\n" +
					"S1,1,1,1,1,1,1,1,20240101,20241231",
				"trips.txt": "route_id,service_id,trip_id\nR1,S2,T1",
			},
			wantNotice:  true,
			description: "An active calendar with no trips on it covers nothing",
		},
		{
			name: "no calendar at all",
			files: map[string]string{
				"trips.txt": "route_id,service_id,trip_id\nR1,S1,T1",
			},
			wantNotice:  false,
			description: "Reported as missing_calendar_and_calendar_date_files elsewhere",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := testutil.CreateTestFeedLoader(t, tt.files)
			container := notice.NewNoticeContainer()

			v := NewDateTripsValidator()
			v.Validate(loader, container, gtfsvalidator.Config{CurrentDate: currentDate})

			got := 0
			for _, n := range container.GetNotices() {
				if n.Code() == "trip_coverage_not_active_for_next7_days" {
					got++
				}
			}

			if tt.wantNotice && got == 0 {
				t.Errorf("expected trip_coverage_not_active_for_next7_days: %s", tt.description)
			}
			if !tt.wantNotice && got > 0 {
				t.Errorf("unexpected trip_coverage_not_active_for_next7_days: %s", tt.description)
			}
		})
	}
}
