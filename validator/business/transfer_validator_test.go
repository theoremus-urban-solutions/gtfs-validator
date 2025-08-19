package business

import (
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/testutil"
	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

func TestTransferValidator_Validate(t *testing.T) {
	files := map[string]string{
		"stops.txt":     "stop_id,stop_name\nA,Stop A\nB,Stop B",
		"transfers.txt": "from_stop_id,to_stop_id,transfer_type,min_transfer_time\nA,B,4,\nA,A,0,\nA,B,2,",
	}

	loader := testutil.CreateTestFeedLoader(t, files)
	container := notice.NewNoticeContainer()

	v := NewTransferValidator()
	v.Validate(loader, container, gtfsvalidator.Config{})

	codes := map[string]int{}
	for _, n := range container.GetNotices() {
		codes[n.Code()]++
	}

	if codes["invalid_transfer_type"] == 0 {
		t.Errorf("expected invalid_transfer_type notice")
	}
	if codes["transfer_to_same_stop"] == 0 {
		t.Errorf("expected transfer_to_same_stop notice")
	}
	if codes["missing_min_transfer_time"] == 0 {
		t.Errorf("expected missing_min_transfer_time notice for type=2 without time")
	}
}
