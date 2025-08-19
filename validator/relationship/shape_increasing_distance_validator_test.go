package relationship

import (
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/testutil"
	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

func TestShapeIncreasingDistanceValidator_Validate(t *testing.T) {
	files := map[string]string{
		"shapes.txt": "shape_id,shape_pt_lat,shape_pt_lon,shape_pt_sequence,shape_dist_traveled\ns1,0,0,1,0\ns1,0,0,2,0\ns1,0,0,3,-5",
	}

	loader := testutil.CreateTestFeedLoader(t, files)
	container := notice.NewNoticeContainer()

	v := NewShapeIncreasingDistanceValidator()
	v.Validate(loader, container, gtfsvalidator.Config{})

	if len(container.GetNotices()) == 0 {
		t.Errorf("expected notices for non-increasing distances")
	}
}
