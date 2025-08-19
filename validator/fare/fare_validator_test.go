package fare

import (
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator/validator"
	"github.com/theoremus-urban-solutions/gtfs-validator/testutil"
)

func TestFareValidator_Validate(t *testing.T) {
	files := map[string]string{
		"fare_attributes.txt": "fare_id,price,currency_type,payment_method,transfers,transfer_duration\nF1,abc,USD,5,3,-10\nF2,1.234567,USD,1,0,60",
		"fare_rules.txt":      "fare_id,origin_id,destination_id,contains_id\nF3,A,A,\nF2,,,C",
	}

	loader := testutil.CreateTestFeedLoader(t, files)
	container := notice.NewNoticeContainer()

	v := NewFareValidator()
	v.Validate(loader, container, gtfsvalidator.Config{})

	codes := map[string]int{}
	for _, n := range container.GetNotices() {
		codes[n.Code()]++
	}

	// From fare_attributes
	if codes["invalid_fare_price"] == 0 {
		t.Errorf("expected invalid_fare_price notice")
	}
	if codes["excessive_price_precision"] == 0 {
		t.Errorf("expected excessive_price_precision notice")
	}
	if codes["invalid_payment_method"] == 0 {
		t.Errorf("expected invalid_payment_method notice")
	}
	if codes["unusual_transfer_value"] == 0 {
		t.Errorf("expected unusual_transfer_value notice for transfers=3")
	}
	if codes["invalid_transfer_duration"] == 0 {
		t.Errorf("expected invalid_transfer_duration notice for negative duration")
	}
	if codes["unnecessary_transfer_duration"] == 0 {
		t.Errorf("expected unnecessary_transfer_duration when transfers=0 but duration given")
	}

	// From fare_rules
	if codes["same_origin_destination"] == 0 {
		t.Errorf("expected same_origin_destination notice")
	}
	// conflicting_fare_rule_fields requires contains_id used with origin/destination; not present here
	// empty_fare_rule requires no rule fields; not present here
	if codes["unused_fare_attribute"] == 0 {
		t.Errorf("expected unused_fare_attribute notice for F1")
	}
}
