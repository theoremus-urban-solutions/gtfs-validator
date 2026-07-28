package business

import (
	"testing"
	"time"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/testutil"
	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

func TestFeedExpirationDateValidator_Validate(t *testing.T) {
	currentDate := time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)
	const header = "feed_publisher_name,feed_publisher_url,feed_lang,feed_start_date,feed_end_date\n"

	tests := []struct {
		name        string
		files       map[string]string
		wantCodes   []string
		description string
	}{
		{
			name:        "feed valid well beyond 30 days",
			files:       map[string]string{"feed_info.txt": header + "A,http://a,en,20240101,20241231"},
			wantCodes:   nil,
			description: "A feed with a year left to run is what the rule asks for",
		},
		{
			name:        "feed end date already past",
			files:       map[string]string{"feed_info.txt": header + "A,http://a,en,20240101,20240115"},
			wantCodes:   []string{"feed_expiration_date7_days"},
			description: "An expired feed is the extreme case of expiring within 7 days",
		},
		{
			name:        "feed expires inside a week",
			files:       map[string]string{"feed_info.txt": header + "A,http://a,en,20240101,20240205"},
			wantCodes:   []string{"feed_expiration_date7_days"},
			description: "Four days of validity left",
		},
		{
			name:        "feed expires inside a month",
			files:       map[string]string{"feed_info.txt": header + "A,http://a,en,20240101,20240220"},
			wantCodes:   []string{"feed_expiration_date30_days"},
			description: "Nineteen days left trips only the 30-day notice",
		},
		{
			name:        "feed starts in the future",
			files:       map[string]string{"feed_info.txt": header + "A,http://a,en,20240301,20241231"},
			wantCodes:   []string{"future_feed"},
			description: "A feed whose window opens next month covers nothing today",
		},
		{
			name:        "future feed that also expires soon",
			files:       map[string]string{"feed_info.txt": header + "A,http://a,en,20240203,20240206"},
			wantCodes:   []string{"future_feed", "feed_expiration_date7_days"},
			description: "Start and end dates are independent checks",
		},
		{
			name:        "no end date",
			files:       map[string]string{"feed_info.txt": "feed_publisher_name,feed_publisher_url,feed_lang\nA,http://a,en"},
			wantCodes:   nil,
			description: "A missing feed_end_date is reported as missing_feed_info_date elsewhere",
		},
		{
			name:        "unparseable end date",
			files:       map[string]string{"feed_info.txt": header + "A,http://a,en,20240101,not-a-date"},
			wantCodes:   nil,
			description: "Reported as invalid_date by the type layer",
		},
		{
			name:        "no feed_info.txt",
			files:       map[string]string{"calendar.txt": "service_id,monday,start_date,end_date\nS1,1,20240101,20240115"},
			wantCodes:   nil,
			description: "Calendar dates are covered by expired_calendar, not here",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := testutil.CreateTestFeedLoader(t, tt.files)
			container := notice.NewNoticeContainer()

			v := NewFeedExpirationDateValidator()
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
