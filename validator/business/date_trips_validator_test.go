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
	const calendarHeader = "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\n"
	const feedInfoHeader = "feed_publisher_name,feed_publisher_url,feed_lang,feed_start_date,feed_end_date\n"

	tests := []struct {
		name        string
		files       map[string]string
		wantCodes   []string
		description string
	}{
		{
			name: "service runs every day of the week",
			files: map[string]string{
				"calendar.txt": calendarHeader + "S1,1,1,1,1,1,1,1,20240101,20241231",
				"trips.txt":    "route_id,service_id,trip_id\nR1,S1,T1",
			},
			wantCodes:   nil,
			description: "Coverage through the whole week is what the rule asks for",
		},
		{
			name: "no day of the week is active",
			files: map[string]string{
				"calendar.txt": calendarHeader + "S1,0,0,0,0,0,0,0,20240101,20241231",
				"trips.txt":    "route_id,service_id,trip_id\nR1,S1,T1",
			},
			wantCodes:   []string{"trip_coverage_not_active_for_next7_days"},
			description: "A service with no active days runs no trips at all",
		},
		{
			name: "service ends mid-week",
			files: map[string]string{
				"calendar.txt": calendarHeader + "S1,1,1,1,1,1,1,1,20240101,20240103",
				"trips.txt":    "route_id,service_id,trip_id\nR1,S1,T1",
			},
			wantCodes:   []string{"trip_coverage_not_active_for_next7_days"},
			description: "Coverage that runs out inside the week is the defect this reports",
		},
		{
			name: "service defined but no trips reference it",
			files: map[string]string{
				"calendar.txt": calendarHeader + "S1,1,1,1,1,1,1,1,20240101,20241231",
				"trips.txt":    "route_id,service_id,trip_id\nR1,S2,T1",
			},
			wantCodes:   []string{"trip_coverage_not_active_for_next7_days"},
			description: "An active calendar with no trips on it covers nothing",
		},
		{
			name:        "no calendar at all",
			files:       map[string]string{"trips.txt": "route_id,service_id,trip_id\nR1,S1,T1"},
			wantCodes:   nil,
			description: "Reported as missing_calendar_and_calendar_date_files elsewhere",
		},
		{
			name: "month-long hole between service dates",
			files: map[string]string{
				"calendar_dates.txt": "service_id,date,exception_type\nS1,20240101,1\nS1,20240201,1",
				"trips.txt":          "route_id,service_id,trip_id\nR1,S1,T1",
			},
			wantCodes:   []string{"big_gap_in_service", "trip_coverage_not_active_for_next7_days"},
			description: "Thirty-one days with nothing running is a hole in the feed",
		},
		{
			name: "gap inside the threshold",
			files: map[string]string{
				"calendar_dates.txt": "service_id,date,exception_type\nS1,20240101,1\nS1,20240110,1",
				"trips.txt":          "route_id,service_id,trip_id\nR1,S1,T1",
			},
			wantCodes:   []string{"trip_coverage_not_active_for_next7_days"},
			description: "Nine days apart is ordinary for an occasional service",
		},
		{
			name: "feed valid well past the last day of service",
			files: map[string]string{
				"calendar.txt":  calendarHeader + "S1,1,1,1,1,1,1,1,20240101,20241231",
				"trips.txt":     "route_id,service_id,trip_id\nR1,S1,T1",
				"feed_info.txt": feedInfoHeader + "A,http://a,en,20240101,20250131",
			},
			wantCodes:   []string{"feed_valid_beyond_total_service_window"},
			description: "A month of declared validity with no service behind it",
		},
		{
			name: "service running outside the declared feed period",
			files: map[string]string{
				"calendar.txt":  calendarHeader + "S1,1,1,1,1,1,1,1,20240101,20241231",
				"trips.txt":     "route_id,service_id,trip_id\nR1,S1,T1",
				"feed_info.txt": feedInfoHeader + "A,http://a,en,20240201,20241130",
			},
			wantCodes:   []string{"service_window_outside_feed_period"},
			description: "Service on both sides of the period a consumer would honour",
		},
		{
			name: "feed period matches the service window",
			files: map[string]string{
				"calendar.txt":  calendarHeader + "S1,1,1,1,1,1,1,1,20240101,20241231",
				"trips.txt":     "route_id,service_id,trip_id\nR1,S1,T1",
				"feed_info.txt": feedInfoHeader + "A,http://a,en,20240101,20241231",
			},
			wantCodes:   nil,
			description: "The two agreeing is the point of declaring a period at all",
		},
		{
			name: "feed_info without dates",
			files: map[string]string{
				"calendar.txt":  calendarHeader + "S1,1,1,1,1,1,1,1,20240101,20241231",
				"trips.txt":     "route_id,service_id,trip_id\nR1,S1,T1",
				"feed_info.txt": "feed_publisher_name,feed_publisher_url,feed_lang\nA,http://a,en",
			},
			wantCodes:   nil,
			description: "Reported as missing_feed_info_date elsewhere",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := testutil.CreateTestFeedLoader(t, tt.files)
			container := notice.NewNoticeContainer()

			v := NewDateTripsValidator()
			v.Validate(loader, container, gtfsvalidator.Config{CurrentDate: currentDate})

			got := map[string]int{}
			for _, n := range container.GetNotices() {
				got[n.Code()]++
			}

			for _, code := range tt.wantCodes {
				if got[code] == 0 {
					t.Errorf("expected %s: %s (got %v)", code, tt.description, got)
				}
			}
			if len(got) != len(tt.wantCodes) {
				t.Errorf("expected exactly %v, got %v: %s", tt.wantCodes, got, tt.description)
			}
		})
	}
}
