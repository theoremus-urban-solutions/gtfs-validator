package core

import (
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/testutil"
	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

func TestCoordinateValidator_Validate(t *testing.T) {
	tests := []struct {
		name                string
		files               map[string]string
		expectedNoticeCodes []string
		description         string
	}{
		{
			name: "real coordinates",
			files: map[string]string{
				StopsFile:    "stop_id,stop_name,stop_lat,stop_lon\n1,Main St,34.0522,-118.2437",
				"shapes.txt": "shape_id,shape_pt_lat,shape_pt_lon,shape_pt_sequence\nS1,34.0522,-118.2437,1",
			},
			expectedNoticeCodes: []string{},
			description:         "A point in Los Angeles is a point",
		},
		{
			name: "point at the origin",
			files: map[string]string{
				StopsFile: "stop_id,stop_name,stop_lat,stop_lon\n1,Main St,0.0,0.0",
			},
			expectedNoticeCodes: []string{"point_near_origin"},
			description:         "Two zeroes are two empty fields, not a stop in the Gulf of Guinea",
		},
		{
			name: "latitude zero on a real longitude",
			files: map[string]string{
				StopsFile: "stop_id,stop_name,stop_lat,stop_lon\n1,Equator,0.0,45.0",
			},
			expectedNoticeCodes: []string{},
			description:         "The equator is a real place; only the pair matters",
		},
		{
			name: "point at the north pole",
			files: map[string]string{
				StopsFile: "stop_id,stop_name,stop_lat,stop_lon\n1,North,90.0,45.0",
			},
			expectedNoticeCodes: []string{"point_near_pole"},
			description:         "The other position a missing value becomes",
		},
		{
			name: "point at the south pole",
			files: map[string]string{
				StopsFile: "stop_id,stop_name,stop_lat,stop_lon\n1,South,-90.0,45.0",
			},
			expectedNoticeCodes: []string{"point_near_pole"},
			description:         "Both poles, not just the north",
		},
		{
			name: "out-of-range latitude is not reported here",
			files: map[string]string{
				StopsFile: "stop_id,stop_name,stop_lat,stop_lon\n1,Main St,91.0,-118.25",
			},
			expectedNoticeCodes: []string{},
			description:         "Reported as number_out_of_range by the field type validator",
		},
		{
			name: "non-numeric coordinate is not reported here",
			files: map[string]string{
				StopsFile: "stop_id,stop_name,stop_lat,stop_lon\n1,Main St,north,-118.25",
			},
			expectedNoticeCodes: []string{},
			description:         "Reported as invalid_float by the field type validator",
		},
		{
			name: "empty coordinates",
			files: map[string]string{
				StopsFile: "stop_id,stop_name,stop_lat,stop_lon\n1,Main St,,",
			},
			expectedNoticeCodes: []string{},
			description:         "A missing coordinate is the required-field validator's business",
		},
		{
			name: "shapes are checked the same way",
			files: map[string]string{
				"shapes.txt": "shape_id,shape_pt_lat,shape_pt_lon,shape_pt_sequence\nS1,0.0,0.0,1",
			},
			expectedNoticeCodes: []string{"point_near_origin"},
			description:         "A shape point at the origin is the same defect as a stop there",
		},
		{
			name: "files without coordinates",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,http://metro.example,America/Los_Angeles",
			},
			expectedNoticeCodes: []string{},
			description:         "Nothing to check",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := testutil.CreateTestFeedLoader(t, tt.files)
			container := notice.NewNoticeContainer()
			validator := NewCoordinateValidator()

			validator.Validate(loader, container, gtfsvalidator.Config{})

			assertNoticeCodes(t, container, tt.expectedNoticeCodes, tt.description)
		})
	}
}

func TestCoordinateValidator_New(t *testing.T) {
	if NewCoordinateValidator() == nil {
		t.Error("NewCoordinateValidator() returned nil")
	}
}
