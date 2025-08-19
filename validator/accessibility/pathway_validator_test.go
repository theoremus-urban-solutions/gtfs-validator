package accessibility

import (
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/testutil"
	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

func TestPathwayValidator_Validate(t *testing.T) {
	files := map[string]string{
		"stops.txt":    "stop_id,stop_name\nA,Stop A\nB,Stop B",
		"pathways.txt": "pathway_id,from_stop_id,to_stop_id,pathway_mode,is_bidirectional,stair_count\nP1,A,B,2,0,\nP2,A,B,2,0,5\nP3,A,A,2,0,1\nP4,A,B,10,0,1\nP5,A,C,2,0,1",
	}
	loader := testutil.CreateTestFeedLoader(t, files)
	container := notice.NewNoticeContainer()
	v := NewPathwayValidator()
	v.Validate(loader, container, gtfsvalidator.Config{})

	codes := map[string]int{}
	for _, n := range container.GetNotices() {
		codes[n.Code()]++
	}
	if codes["missing_recommended_field"] == 0 { // P1 stairs missing stair_count
		t.Errorf("expected missing_recommended_field for stair_count on stairs")
	}
	if codes["pathway_to_same_stop"] == 0 { // P3 A->A
		t.Errorf("expected pathway_to_same_stop for A->A")
	}
	if codes["invalid_pathway_mode"] == 0 { // P4 invalid mode 10
		t.Errorf("expected invalid_pathway_mode for mode 10")
	}
	if codes["foreign_key_violation"] == 0 { // P5 references missing stop C
		t.Errorf("expected foreign_key_violation for missing stop reference")
	}
}
