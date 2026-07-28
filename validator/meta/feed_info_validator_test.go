package meta

import (
	"testing"
	"time"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/testutil"
	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

func TestFeedInfoValidator_ValidateRecommendedFields(t *testing.T) {
	const header = "feed_publisher_name,feed_publisher_url,feed_lang,feed_start_date,feed_end_date,feed_contact_email,feed_contact_url\n"

	tests := []struct {
		name                string
		row                 string
		expectedNoticeCodes []string
		description         string
	}{
		{
			name:                "complete feed info",
			row:                 "Metro,http://metro.example,en,20250101,20251231,support@metro.example,http://metro.example/contact",
			expectedNoticeCodes: []string{},
			description:         "A row giving both dates and a contact should generate no notices",
		},
		{
			name:                "neither date given",
			row:                 "Metro,http://metro.example,en,,,support@metro.example,",
			expectedNoticeCodes: []string{},
			description:         "Both dates are optional when neither is given",
		},
		{
			name:                "start date without end date",
			row:                 "Metro,http://metro.example,en,20250101,,support@metro.example,",
			expectedNoticeCodes: []string{"missing_feed_info_date"},
			description:         "feed_start_date without feed_end_date should generate a notice",
		},
		{
			name:                "end date without start date",
			row:                 "Metro,http://metro.example,en,,20251231,support@metro.example,",
			expectedNoticeCodes: []string{"missing_feed_info_date"},
			description:         "feed_end_date without feed_start_date should generate a notice",
		},
		{
			name:                "no contact at all",
			row:                 "Metro,http://metro.example,en,20250101,20251231,,",
			expectedNoticeCodes: []string{"missing_feed_contact_email_and_url"},
			description:         "Neither contact field given should generate a notice",
		},
		{
			name:                "contact url only",
			row:                 "Metro,http://metro.example,en,20250101,20251231,,http://metro.example/contact",
			expectedNoticeCodes: []string{},
			description:         "Either contact field alone is enough",
		},
		{
			name:                "half a date range and no contact",
			row:                 "Metro,http://metro.example,en,,20251231,,",
			expectedNoticeCodes: []string{"missing_feed_info_date", "missing_feed_contact_email_and_url"},
			description:         "Both defects should be reported independently",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := testutil.CreateTestFeedLoader(t, map[string]string{"feed_info.txt": header + tt.row})
			container := notice.NewNoticeContainer()

			cfg := gtfsvalidator.Config{CurrentDate: time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)}
			NewFeedInfoValidator().Validate(loader, container, cfg)

			expectedCounts := make(map[string]int)
			for _, code := range tt.expectedNoticeCodes {
				expectedCounts[code]++
			}
			actualCounts := make(map[string]int)
			for _, n := range container.GetNotices() {
				actualCounts[n.Code()]++
			}

			for code, expected := range expectedCounts {
				if actualCounts[code] != expected {
					t.Errorf("Expected %d notices with code '%s', got %d for %s", expected, code, actualCounts[code], tt.description)
				}
			}
			for code := range actualCounts {
				if expectedCounts[code] == 0 {
					t.Errorf("Unexpected notice code: %s for %s", code, tt.description)
				}
			}
		})
	}
}

func TestFeedInfoValidator_Validate(t *testing.T) {
	files := map[string]string{
		"feed_info.txt": "feed_publisher_name,feed_publisher_url,feed_lang,default_lang,feed_start_date,feed_end_date,feed_version,feed_contact_email,feed_contact_url\n,,en,eng,20260101,20250101,1.0,invalid-email,www.foo",
	}

	loader := testutil.CreateTestFeedLoader(t, files)
	container := notice.NewNoticeContainer()

	cfg := gtfsvalidator.Config{CurrentDate: time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)}
	v := NewFeedInfoValidator()
	v.Validate(loader, container, cfg)

	codes := map[string]int{}
	for _, n := range container.GetNotices() {
		codes[n.Code()]++
	}

	if codes["missing_required_field"] < 2 { // publisher_name and publisher_url
		t.Errorf("expected missing_required_field notices for required fields")
	}
	if codes["invalid_language_code"] == 0 { // default_lang should be 2 letters
		t.Errorf("expected invalid_language_code notice for default_lang")
	}
	// feed_end_date before feed_start_date is reported as
	// start_and_end_range_out_of_order by core/field_type_validator.go, which
	// checks every paired range in the feed.
	if codes["feed_info_end_date_before_start_date"] != 0 {
		t.Errorf("feed_info_end_date_before_start_date should no longer be emitted")
	}
	if codes["invalid_email"] == 0 {
		t.Errorf("expected invalid_email notice")
	}
	if codes["invalid_url"] == 0 {
		t.Errorf("expected invalid_url notice for contact URL without http(s)")
	}
}
