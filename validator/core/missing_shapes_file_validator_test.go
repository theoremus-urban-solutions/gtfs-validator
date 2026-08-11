package core

import (
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/testutil"
	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

func TestMissingShapesFileValidator_Validate(t *testing.T) {
	const shapes = "shape_id,shape_pt_lat,shape_pt_lon,shape_pt_sequence\nSH1,34.05,-118.25,1"

	tests := []struct {
		name     string
		files    map[string]string
		expected bool
	}{
		{
			name:     "shapes.txt absent",
			files:    map[string]string{},
			expected: true,
		},
		{
			name:     "shapes.txt present with points",
			files:    map[string]string{"shapes.txt": shapes},
			expected: false,
		},
		{
			name: "shapes.txt present but header-only",
			files: map[string]string{
				"shapes.txt": "shape_id,shape_pt_lat,shape_pt_lon,shape_pt_sequence",
			},
			// A header with no points draws nothing, so it is the same as no
			// shapes at all.
			expected: true,
		},
		{
			name: "zone-based demand-responsive trip",
			files: map[string]string{
				"stop_times.txt": "trip_id,arrival_time,departure_time,stop_id,location_id,stop_sequence\nT1,08:00:00,08:00:00,,Z1,1",
			},
			// The trip serves an area, so there is no path to draw.
			expected: false,
		},
		{
			name: "fixed-stops demand-responsive trip",
			files: map[string]string{
				"location_groups.txt": "location_group_id,location_group_name\nG1,Downtown",
				"stop_times.txt":      "trip_id,arrival_time,departure_time,stop_id,location_group_id,stop_sequence\nT1,08:00:00,08:00:00,,G1,1",
			},
			expected: false,
		},
		{
			name: "location_group_id without the groups themselves",
			files: map[string]string{
				"stop_times.txt": "trip_id,arrival_time,departure_time,stop_id,location_group_id,stop_sequence\nT1,08:00:00,08:00:00,,G1,1",
			},
			// A group id naming no declared group does not establish the
			// feature, so the missing shapes still stand.
			expected: true,
		},
		{
			name: "location_id alongside a stop_id",
			files: map[string]string{
				"stop_times.txt": "trip_id,arrival_time,departure_time,stop_id,location_id,stop_sequence\nT1,08:00:00,08:00:00,S1,Z1,1",
			},
			// The trip stops somewhere definite, so it has a path after all.
			expected: true,
		},
		{
			name: "ordinary stop_times with no flex columns",
			files: map[string]string{
				"stop_times.txt": "trip_id,arrival_time,departure_time,stop_id,stop_sequence\nT1,08:00:00,08:00:00,S1,1",
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := testutil.CreateTestFeedLoader(t, tt.files)
			container := notice.NewNoticeContainer()
			validator := NewMissingShapesFileValidator()

			validator.Validate(loader, container, gtfsvalidator.Config{})

			reported := false
			for _, n := range container.GetNotices() {
				if n.Code() != "missing_recommended_file" {
					t.Errorf("Unexpected notice code: %s", n.Code())
					continue
				}
				if filename, _ := n.Context()["filename"].(string); filename != "shapes.txt" {
					t.Errorf("Expected shapes.txt, got %q", filename)
				}
				reported = true
			}

			if reported != tt.expected {
				t.Errorf("Expected reported=%v, got %v", tt.expected, reported)
			}
		})
	}
}
