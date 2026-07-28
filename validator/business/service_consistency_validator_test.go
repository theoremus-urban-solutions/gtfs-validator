package business

import (
	"testing"
	"time"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/testutil"
	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

func TestServiceConsistencyValidator_Validate(t *testing.T) {
	currentDate := time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC)
	const calendarHeader = "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\n"

	tests := []struct {
		name        string
		files       map[string]string
		wantCodes   []string
		description string
	}{
		{
			name: "service already running",
			files: map[string]string{
				"calendar.txt": calendarHeader + "S1,1,1,1,1,1,0,0,20240101,20241231",
			},
			wantCodes:   nil,
			description: "A service that started before today covers today",
		},
		{
			name: "every service starts in the future",
			files: map[string]string{
				"calendar.txt": calendarHeader +
					"S1,1,1,1,1,1,0,0,20240201,20241231\n" +
					"S2,0,0,0,0,0,1,1,20240301,20241231",
			},
			wantCodes:   []string{"future_calendar"},
			description: "Nothing in the feed runs on the date it is read",
		},
		{
			name: "future calendar rescued by an added exception date",
			files: map[string]string{
				"calendar.txt":       calendarHeader + "S1,1,1,1,1,1,0,0,20240201,20241231",
				"calendar_dates.txt": "service_id,date,exception_type\nS1,20240105,1",
			},
			wantCodes:   nil,
			description: "An added date before today means the service does run now",
		},
		{
			name: "removal exception does not count as a start",
			files: map[string]string{
				"calendar.txt":       calendarHeader + "S1,1,1,1,1,1,0,0,20240201,20241231",
				"calendar_dates.txt": "service_id,date,exception_type\nS1,20240105,2",
			},
			wantCodes:   []string{"future_calendar"},
			description: "Removing service on a past date adds no coverage",
		},
		{
			name: "service with no active weekday is ignored for coverage",
			files: map[string]string{
				"calendar.txt": calendarHeader +
					"S1,0,0,0,0,0,0,0,20240101,20241231\n" +
					"S2,1,1,1,1,1,0,0,20240201,20241231",
			},
			wantCodes:   []string{"future_calendar"},
			description: "A service running on no weekday never covers today, whatever its window",
		},
		{
			name: "conflicting exceptions on the same date",
			files: map[string]string{
				"calendar.txt":       calendarHeader + "S1,1,1,1,1,1,0,0,20240101,20241231",
				"calendar_dates.txt": "service_id,date,exception_type\nS1,20240201,1\nS1,20240201,2",
			},
			wantCodes:   []string{"conflicting_calendar_exception"},
			description: "Adding and removing the same date leaves the schedule undefined",
		},
		{
			name: "duplicate exceptions on the same date",
			files: map[string]string{
				"calendar.txt":       calendarHeader + "S1,1,1,1,1,1,0,0,20240101,20241231",
				"calendar_dates.txt": "service_id,date,exception_type\nS1,20240201,1\nS1,20240201,1",
			},
			wantCodes:   []string{"duplicate_calendar_exception"},
			description: "The same exception twice is a redundant row",
		},
		{
			name:        "no calendar files",
			files:       map[string]string{"trips.txt": "route_id,service_id,trip_id\nR1,S1,T1"},
			wantCodes:   nil,
			description: "Reported as missing_calendar_and_calendar_date_files elsewhere",
		},
		{
			name: "unparseable start dates",
			files: map[string]string{
				"calendar.txt": calendarHeader + "S1,1,1,1,1,1,0,0,not-a-date,20241231",
			},
			wantCodes:   nil,
			description: "Reported as invalid_date by the type layer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := testutil.CreateTestFeedLoader(t, tt.files)
			container := notice.NewNoticeContainer()

			v := NewServiceConsistencyValidator()
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
