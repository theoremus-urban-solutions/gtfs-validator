package relationship

import (
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator/validator"
	coretest "github.com/theoremus-urban-solutions/gtfs-validator/validator/core"
)

func TestShapeDistanceValidator_Validate(t *testing.T) {
	files := map[string]string{
		"shapes.txt": "shape_id,shape_pt_lat,shape_pt_lon,shape_pt_sequence,shape_dist_traveled\ns1,0,0,1,0\ns1,0,0,1,0\ns1,0,0,2,0",
	}

	loader := coretest.CreateTestFeedLoader(t, files)
	container := notice.NewNoticeContainer()

	v := NewShapeDistanceValidator()
	v.Validate(loader, container, gtfsvalidator.Config{})

	codes := map[string]int{}
	for _, n := range container.GetNotices() {
		codes[n.Code()]++
	}

	if codes["duplicate_shape_sequence"] == 0 {
		t.Errorf("expected duplicate_shape_sequence notice")
	}
	if codes["decreasing_or_equal_shape_distance"] == 0 {
		t.Errorf("expected decreasing_or_equal_shape_distance notice")
	}
}
