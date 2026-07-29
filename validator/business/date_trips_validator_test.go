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
			wantCodes:   nil,
			description: "No date runs a trip, so there is no window to judge: service_has_no_active_day_of_the_week reports this feed",
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
			wantCodes:   nil,
			description: "An active calendar no trip references leaves the feed with no service dates at all",
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
			wantCodes:   []string{"big_gap_in_service"},
			description: "Thirty-one days with nothing running is a hole in the feed, but the window around it still covers the week",
		},
		{
			name: "gap inside the threshold",
			files: map[string]string{
				"calendar_dates.txt": "service_id,date,exception_type\nS1,20240101,1\nS1,20240110,1",
				"trips.txt":          "route_id,service_id,trip_id\nR1,S1,T1",
			},
			wantCodes:   nil,
			description: "Nine days apart is ordinary for an occasional service",
		},
		{
			name: "weekday-only service over the coming week",
			files: map[string]string{
				"calendar.txt": calendarHeader + "S1,1,1,1,1,1,0,0,20240101,20241231",
				"trips.txt":    "route_id,service_id,trip_id\nR1,S1,T1",
			},
			wantCodes:   nil,
			description: "The weekend inside the window is not a coverage defect",
		},
		{
			name: "the busy season starts after the coming week",
			files: map[string]string{
				"calendar.txt": calendarHeader +
					"S1,1,1,1,1,1,1,1,20240601,20241231\n" +
					"S2,1,1,1,1,1,1,1,20240101,20240531",
				"trips.txt": "route_id,service_id,trip_id\n" +
					"R1,S1,T1\nR1,S1,T2\nR1,S1,T3\nR1,S1,T4\nR1,S1,T5\n" +
					"R1,S1,T6\nR1,S1,T7\nR1,S1,T8\nR1,S1,T9\nR1,S1,T10\n" +
					"R1,S2,T11",
			},
			wantCodes:   []string{"trip_coverage_not_active_for_next7_days"},
			description: "One trip a day running now does not stand in for the ten a day the feed is really about",
		},
		{
			name: "the busy season is running now",
			files: map[string]string{
				"calendar.txt": calendarHeader +
					"S1,1,1,1,1,1,1,1,20240101,20240731\n" +
					"S2,1,1,1,1,1,1,1,20240801,20241231",
				"trips.txt": "route_id,service_id,trip_id\n" +
					"R1,S1,T1\nR1,S1,T2\nR1,S1,T3\nR1,S1,T4\nR1,S1,T5\n" +
					"R1,S1,T6\nR1,S1,T7\nR1,S1,T8\nR1,S1,T9\nR1,S1,T10\n" +
					"R1,S2,T11",
			},
			wantCodes:   nil,
			description: "The week ahead sits inside the window the majority of trips run in",
		},
		{
			name: "a long thin calendar around a short busy one",
			files: map[string]string{
				// Under the ratio alone the busy fifty days are lost in the
				// thousand thin ones; the thirty-day limit is what finds them.
				"calendar.txt": calendarHeader +
					"S1,1,1,1,1,1,1,1,20261001,20261120\n" +
					"S2,1,1,1,1,1,1,1,20240101,20260930",
				"trips.txt": "route_id,service_id,trip_id\n" +
					"R1,S1,T1\nR1,S1,T2\nR1,S1,T3\nR1,S1,T4\nR1,S1,T5\n" +
					"R1,S1,T6\nR1,S1,T7\nR1,S1,T8\nR1,S1,T9\nR1,S1,T10\n" +
					"R1,S2,T11",
			},
			wantCodes:   []string{"trip_coverage_not_active_for_next7_days"},
			description: "The window is where the service really is, two years out",
		},
		{
			name: "a frequency-based trip carries the service",
			files: map[string]string{
				"calendar.txt": calendarHeader +
					"S1,1,1,1,1,1,1,1,20240701,20241231\n" +
					"S2,1,1,1,1,1,1,1,20240101,20240630",
				"trips.txt": "route_id,service_id,trip_id\n" +
					"R1,S1,T1\n" +
					"R1,S2,T2\nR1,S2,T3\nR1,S2,T4\nR1,S2,T5\nR1,S2,T6",
				// Six hours every ten minutes: one row, thirty-seven vehicles.
				"frequencies.txt": "trip_id,start_time,end_time,headway_secs\nT1,06:00:00,12:00:00,600",
			},
			wantCodes:   []string{"trip_coverage_not_active_for_next7_days"},
			description: "A frequency-based trip counts for every vehicle it runs, not for its one row",
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

// Service dates are midnight-UTC and the coverage check compares them against
// the current date, so a current date carrying a wall-clock time and a local
// zone used to miss every comparison and report any feed as uncovered. The
// window ending exactly seven days out is where an untruncated time still
// decides it by a matter of hours.
func TestDateTripsValidator_CurrentDateWithClockTime(t *testing.T) {
	const calendarHeader = "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\n"

	feeds := map[string]map[string]string{
		"service all year": {
			"calendar.txt": calendarHeader + "S1,1,1,1,1,1,1,1,20240101,20241231",
			"trips.txt":    "route_id,service_id,trip_id\nR1,S1,T1",
		},
		"service ending exactly seven days out": {
			"calendar.txt": calendarHeader + "S1,1,1,1,1,1,1,1,20240101,20240619",
			"trips.txt":    "route_id,service_id,trip_id\nR1,S1,T1",
		},
	}

	// Mid-afternoon in a zone well east of UTC: the worst case for a
	// comparison that assumes midnight.
	sofia := time.FixedZone("EEST", 3*3600)
	currentDates := []time.Time{
		time.Date(2024, 6, 12, 14, 32, 11, 0, sofia),
		time.Date(2024, 6, 12, 23, 59, 59, 0, sofia),
		time.Date(2024, 6, 12, 0, 0, 0, 0, time.UTC),
	}

	for name, files := range feeds {
		for _, currentDate := range currentDates {
			loader := testutil.CreateTestFeedLoader(t, files)
			container := notice.NewNoticeContainer()

			NewDateTripsValidator().Validate(loader, container, gtfsvalidator.Config{CurrentDate: currentDate})

			for _, n := range container.GetNotices() {
				if n.Code() == "trip_coverage_not_active_for_next7_days" {
					t.Errorf("%s covers the coming week, but %v reported it as uncovered", name, currentDate)
				}
			}
		}
	}
}
