package meta

import (
	"testing"
	"time"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator/validator"
	"github.com/theoremus-urban-solutions/gtfs-validator/testutil"
)

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
	if codes["feed_info_end_date_before_start_date"] == 0 {
		t.Errorf("expected feed_info_end_date_before_start_date notice")
	}
	if codes["invalid_email"] == 0 {
		t.Errorf("expected invalid_email notice")
	}
	if codes["invalid_url"] == 0 {
		t.Errorf("expected invalid_url notice for contact URL without http(s)")
	}
}
