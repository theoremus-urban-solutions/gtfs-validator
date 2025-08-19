package business

import (
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator/validator"
	"github.com/theoremus-urban-solutions/gtfs-validator/testutil"
)

func TestTransferTimingValidator_Validate(t *testing.T) {
	files := map[string]string{
		"stops.txt":     "stop_id,stop_name,stop_lat,stop_lon\nA,Stop A,0,0\nB,Stop B,0,0.01",
		"transfers.txt": "from_stop_id,to_stop_id,transfer_type,min_transfer_time\nA,B,2,10\nB,A,3,120\nA,A,0,",
	}

	loader := testutil.CreateTestFeedLoader(t, files)
	container := notice.NewNoticeContainer()

	v := NewTransferTimingValidator()
	v.Validate(loader, container, gtfsvalidator.Config{})

	codes := map[string]int{}
	for _, n := range container.GetNotices() {
		codes[n.Code()]++
	}

	if codes["very_short_transfer_time"] == 0 {
		t.Errorf("expected very_short_transfer_time notice for too short min_transfer_time vs distance")
	}
	if codes["unnecessary_min_transfer_time"] == 0 {
		t.Errorf("expected unnecessary_min_transfer_time for type=3 with min_transfer_time")
	}
	if codes["transfer_to_same_stop"] == 0 {
		t.Errorf("expected transfer_to_same_stop notice")
	}
}
