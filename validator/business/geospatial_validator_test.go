package business

import (
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator/validator"
	coretest "github.com/theoremus-urban-solutions/gtfs-validator/validator/core"
)

func TestGeospatialValidator_Validate(t *testing.T) {
	files := map[string]string{
		"stops.txt":  "stop_id,stop_name,stop_lat,stop_lon,parent_station\nA,Stop A,0,0,\nB,Stop B,0,0.1,A\nC,Stop C,0.00001,0.00001,",
		"shapes.txt": "shape_id,shape_pt_lat,shape_pt_lon,shape_pt_sequence\ns,0,0,1\ns,0,0.00005,2",
	}

	loader := coretest.CreateTestFeedLoader(t, files)
	container := notice.NewNoticeContainer()

	v := NewGeospatialValidator()
	v.Validate(loader, container, gtfsvalidator.Config{})

	codes := map[string]int{}
	for _, n := range container.GetNotices() {
		codes[n.Code()]++
	}

	// Accept any of several geospatial notices depending on data and bounds
	if codes["child_station_too_far_from_parent"] == 0 && codes["invalid_latitude"] == 0 && codes["invalid_longitude"] == 0 && codes["shape_point_outside_feed_bounds"] == 0 && codes["very_small_feed_coverage"] == 0 {
		t.Errorf("expected at least one geospatial notice to be emitted")
	}
}
