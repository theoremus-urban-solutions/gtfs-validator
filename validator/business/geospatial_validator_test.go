package business

import (
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/testutil"
	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

func TestGeospatialValidator_Validate(t *testing.T) {
	tests := []struct {
		name                string
		stops               string
		expectedNoticeCodes []string
		description         string
	}{
		{
			name:                "child on the same forecourt as its parent",
			stops:               "stop_id,stop_name,stop_lat,stop_lon,parent_station\nA,Stop A,0,0,\nB,Stop B,0,0.001,A",
			expectedNoticeCodes: []string{},
			description:         "111 m apart is an ordinary platform within its station",
		},
		{
			name:                "child a suburb away from its parent",
			stops:               "stop_id,stop_name,stop_lat,stop_lon,parent_station\nA,Stop A,0,0,\nB,Stop B,0,0.1,A",
			expectedNoticeCodes: []string{"child_station_too_far_from_parent"},
			description:         "11 km apart is not one station, whatever parent_station says",
		},
		{
			name:                "latitude off the globe",
			stops:               "stop_id,stop_name,stop_lat,stop_lon\nA,Stop A,95,0",
			expectedNoticeCodes: []string{},
			description:         "there is no 95th parallel, but the bound is the field type validator's to enforce",
		},
		{
			name:                "longitude off the globe",
			stops:               "stop_id,stop_name,stop_lat,stop_lon\nA,Stop A,0,200",
			expectedNoticeCodes: []string{},
			description:         "nor a 200th meridian, and likewise reported as number_out_of_range",
		},
		{
			name:                "parent_station naming a stop the feed does not have",
			stops:               "stop_id,stop_name,stop_lat,stop_lon,parent_station\nA,Stop A,0,0,MISSING",
			expectedNoticeCodes: []string{},
			description:         "the dangling reference is the foreign key check's to report",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := testutil.CreateTestFeedLoader(t, map[string]string{"stops.txt": tt.stops})
			container := notice.NewNoticeContainer()

			NewGeospatialValidator().Validate(loader, container, gtfsvalidator.Config{})

			var actual []string
			found := make(map[string]bool)
			for _, n := range container.GetNotices() {
				actual = append(actual, n.Code())
				found[n.Code()] = true
			}

			for _, expected := range tt.expectedNoticeCodes {
				if !found[expected] {
					t.Errorf("expected notice %q, got %v (%s)", expected, actual, tt.description)
				}
			}

			if len(tt.expectedNoticeCodes) == 0 && len(actual) > 0 {
				t.Errorf("expected no notices, got %v (%s)", actual, tt.description)
			}
		})
	}
}
