package business

import (
	"testing"
	"time"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator/validator"
	"github.com/theoremus-urban-solutions/gtfs-validator/testutil"
)

func TestFeedExpirationDateValidator_Validate(t *testing.T) {
	files := map[string]string{
		"feed_info.txt": "feed_publisher_name,feed_publisher_url,feed_lang,feed_start_date,feed_end_date\nA,http://a,en,20240101,20240115",
	}

	loader := testutil.CreateTestFeedLoader(t, files)
	container := notice.NewNoticeContainer()

	cfg := gtfsvalidator.Config{CurrentDate: time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)}
	v := NewFeedExpirationDateValidator()
	v.Validate(loader, container, cfg)

	codes := map[string]int{}
	for _, n := range container.GetNotices() {
		codes[n.Code()]++
	}

	if codes["feed_expired"] == 0 {
		t.Errorf("expected feed_expired notice when end date is in the past")
	}
}
