package accessibility

import (
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator/validator"
	"github.com/theoremus-urban-solutions/gtfs-validator/testutil"
)

func TestLevelValidator_Validate(t *testing.T) {
	files := map[string]string{
		"levels.txt": "level_id,level_index,level_name\nL1,0,Ground\nL2,1,\nL3,100,Top", // L2 missing level_name (recommended), L3 unreasonable index
		"stops.txt":  "stop_id,stop_name,level_id\nS1,Stop 1,L1",
	}
	loader := testutil.CreateTestFeedLoader(t, files)
	container := notice.NewNoticeContainer()
	v := NewLevelValidator()
	v.Validate(loader, container, gtfsvalidator.Config{})

	codes := map[string]int{}
	for _, n := range container.GetNotices() {
		codes[n.Code()]++
	}
	if codes["missing_recommended_field"] == 0 {
		t.Errorf("expected missing_recommended_field for level_name")
	}
	if codes["unreasonable_level_index"] == 0 {
		t.Errorf("expected unreasonable_level_index for high index")
	}
}
