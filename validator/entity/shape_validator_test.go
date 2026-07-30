package entity

import (
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/testutil"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

func TestShapeValidator_Validate(t *testing.T) {
	tests := []struct {
		name                string
		files               map[string]string
		expectedNoticeCodes []string
		description         string
	}{
		{
			name: "increasing distances",
			files: map[string]string{
				"shapes.txt": "shape_id,shape_pt_lat,shape_pt_lon,shape_pt_sequence,shape_dist_traveled\n" +
					"shape1,37.7749,-122.4194,0,0.0\n" +
					"shape1,37.7750,-122.4195,1,100.5\n" +
					"shape1,37.7751,-122.4196,2,200.8",
			},
			expectedNoticeCodes: []string{},
			description:         "shape_dist_traveled increasing along the shape is valid",
		},
		{
			name: "no distances at all",
			files: map[string]string{
				"shapes.txt": "shape_id,shape_pt_lat,shape_pt_lon,shape_pt_sequence\n" +
					"shape1,37.7749,-122.4194,0\n" +
					"shape1,37.7750,-122.4195,1\n" +
					"shape1,37.7751,-122.4196,2",
			},
			expectedNoticeCodes: []string{},
			description:         "shape_dist_traveled is optional",
		},
		{
			name: "single point",
			files: map[string]string{
				"shapes.txt": "shape_id,shape_pt_lat,shape_pt_lon,shape_pt_sequence\n" +
					"shape1,37.7749,-122.4194,0",
			},
			expectedNoticeCodes: []string{"single_shape_point"},
			description:         "A shape needs at least two points to describe a path",
		},
		{
			name: "decreasing distance",
			files: map[string]string{
				"shapes.txt": "shape_id,shape_pt_lat,shape_pt_lon,shape_pt_sequence,shape_dist_traveled\n" +
					"shape1,37.7749,-122.4194,0,100.0\n" +
					"shape1,37.7750,-122.4195,1,50.0\n" +
					"shape1,37.7751,-122.4196,2,200.0",
			},
			expectedNoticeCodes: []string{"decreasing_shape_distance"},
			description:         "shape_dist_traveled must not decrease",
		},
		{
			name: "equal distance, identical coordinates",
			files: map[string]string{
				"shapes.txt": "shape_id,shape_pt_lat,shape_pt_lon,shape_pt_sequence,shape_dist_traveled\n" +
					"shape1,37.7749,-122.4194,0,100.0\n" +
					"shape1,37.7749,-122.4194,1,100.0\n" +
					"shape1,37.7751,-122.4196,2,200.0",
			},
			expectedNoticeCodes: []string{"equal_shape_distance_same_coordinates"},
			description:         "A repeated point is duplicative, not a distance error",
		},
		{
			name: "equal distance, coordinates below the 1.11 m threshold",
			files: map[string]string{
				"shapes.txt": "shape_id,shape_pt_lat,shape_pt_lon,shape_pt_sequence,shape_dist_traveled\n" +
					"shape1,37.774900,-122.419400,0,100.0\n" +
					"shape1,37.774905,-122.419400,1,100.0\n" +
					"shape1,37.775100,-122.419600,2,200.0",
			},
			expectedNoticeCodes: []string{"equal_shape_distance_diff_coordinates_distance_below_threshold"},
			description:         "Sub-metre coordinate differences are rounding, so this is a warning",
		},
		{
			name: "equal distance, coordinates far apart",
			files: map[string]string{
				"shapes.txt": "shape_id,shape_pt_lat,shape_pt_lon,shape_pt_sequence,shape_dist_traveled\n" +
					"shape1,37.7749,-122.4194,0,100.0\n" +
					"shape1,37.7760,-122.4194,1,100.0\n" +
					"shape1,37.7771,-122.4196,2,200.0",
			},
			expectedNoticeCodes: []string{"equal_shape_distance_diff_coordinates"},
			description:         "The shape covers ground the distance does not account for",
		},
		{
			name: "partial distances are not compared",
			files: map[string]string{
				"shapes.txt": "shape_id,shape_pt_lat,shape_pt_lon,shape_pt_sequence,shape_dist_traveled\n" +
					"shape1,37.7749,-122.4194,0,0.0\n" +
					"shape1,37.7750,-122.4195,1,\n" +
					"shape1,37.7751,-122.4196,2,200.8",
			},
			expectedNoticeCodes: []string{},
			description:         "A point without a distance has nothing to compare against",
		},
		{
			name: "negative sequences ordered correctly",
			files: map[string]string{
				"shapes.txt": "shape_id,shape_pt_lat,shape_pt_lon,shape_pt_sequence\n" +
					"shape1,37.7749,-122.4194,-1\n" +
					"shape1,37.7750,-122.4195,0\n" +
					"shape1,37.7751,-122.4196,1",
			},
			expectedNoticeCodes: []string{},
			description:         "Sequence values only need to increase",
		},
		{
			name: "no shapes file",
			files: map[string]string{
				"trips.txt": "trip_id,route_id,service_id\n" +
					"trip1,route1,service1",
			},
			expectedNoticeCodes: []string{},
			description:         "A feed without shapes.txt is valid",
		},
		{
			name: "shape a trip draws",
			files: map[string]string{
				"shapes.txt": "shape_id,shape_pt_lat,shape_pt_lon,shape_pt_sequence\n" +
					"shape1,37.7749,-122.4194,0\n" +
					"shape1,37.7750,-122.4195,1",
				"trips.txt": "trip_id,route_id,service_id,shape_id\n" +
					"trip1,route1,service1,shape1",
			},
			expectedNoticeCodes: []string{},
			description:         "a referenced shape is doing its job",
		},
		{
			name: "shape no trip draws",
			files: map[string]string{
				"shapes.txt": "shape_id,shape_pt_lat,shape_pt_lon,shape_pt_sequence\n" +
					"shape1,37.7749,-122.4194,0\n" +
					"shape1,37.7750,-122.4195,1\n" +
					"shape2,37.7749,-122.4194,0\n" +
					"shape2,37.7750,-122.4195,1",
				"trips.txt": "trip_id,route_id,service_id,shape_id\n" +
					"trip1,route1,service1,shape1",
			},
			expectedNoticeCodes: []string{"unused_shape"},
			description:         "shape2 is drawn by nothing, so it is dead weight or a missing reference",
		},
		{
			name: "trips without shape_id at all",
			files: map[string]string{
				"shapes.txt": "shape_id,shape_pt_lat,shape_pt_lon,shape_pt_sequence\n" +
					"shape1,37.7749,-122.4194,0\n" +
					"shape1,37.7750,-122.4195,1",
				"trips.txt": "trip_id,route_id,service_id\n" +
					"trip1,route1,service1",
			},
			expectedNoticeCodes: []string{"unused_shape"},
			description:         "shape_id is optional on a trip, but a shape nothing names is still unused",
		},
		{
			name: "no trips file to compare against",
			files: map[string]string{
				"shapes.txt": "shape_id,shape_pt_lat,shape_pt_lon,shape_pt_sequence\n" +
					"shape1,37.7749,-122.4194,0\n" +
					"shape1,37.7750,-122.4195,1",
			},
			expectedNoticeCodes: []string{},
			description:         "with no references to check, every shape would look unused; that is the missing-file check's to report",
		},
		{
			name: "whitespace is trimmed",
			files: map[string]string{
				"shapes.txt": "shape_id,shape_pt_lat,shape_pt_lon,shape_pt_sequence,shape_dist_traveled\n" +
					" shape1 , 37.7749 , -122.4194 , 0 , 0.0 \n" +
					" shape1 , 37.7750 , -122.4195 , 1 , 100.5 ",
			},
			expectedNoticeCodes: []string{},
			description:         "Padded values parse the same as unpadded ones",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test feed loader
			feedLoader := testutil.CreateTestFeedLoader(t, tt.files)

			// Create notice container and validator
			container := notice.NewNoticeContainer()
			validator := NewShapeValidator()
			config := gtfsvalidator.Config{}

			// Run validation
			validator.Validate(feedLoader, container, config)

			// Get all notices
			allNotices := container.GetNotices()

			// Extract notice codes
			var actualNoticeCodes []string
			for _, n := range allNotices {
				actualNoticeCodes = append(actualNoticeCodes, n.Code())
			}

			// Check if we got the expected notice codes
			expectedSet := make(map[string]bool)
			for _, code := range tt.expectedNoticeCodes {
				expectedSet[code] = true
			}

			actualSet := make(map[string]bool)
			for _, code := range actualNoticeCodes {
				actualSet[code] = true
			}

			// Verify expected codes are present
			for expectedCode := range expectedSet {
				if !actualSet[expectedCode] {
					t.Errorf("Expected notice code '%s' not found. Got: %v", expectedCode, actualNoticeCodes)
				}
			}

			// If no notices expected, ensure no notices were generated
			if len(tt.expectedNoticeCodes) == 0 && len(actualNoticeCodes) > 0 {
				t.Errorf("Expected no notices, but got: %v", actualNoticeCodes)
			}

			t.Logf("Test '%s': Expected %v, Got %v", tt.name, tt.expectedNoticeCodes, actualNoticeCodes)
		})
	}
}

func TestShapeValidator_LoadShapes(t *testing.T) {
	validator := NewShapeValidator()

	tests := []struct {
		name     string
		csvData  string
		expected map[string]*ShapeInfo
	}{
		{
			name: "basic shape loading",
			csvData: "shape_id,shape_pt_lat,shape_pt_lon,shape_pt_sequence,shape_dist_traveled\n" +
				"shape1,37.7749,-122.4194,0,0.0\n" +
				"shape1,37.7750,-122.4195,1,100.5",
			expected: map[string]*ShapeInfo{
				"shape1": {
					ShapeID: "shape1",
					Points: []*ShapePointDetailed{
						{
							ShapeID:           "shape1",
							ShapePtLat:        37.7749,
							ShapePtLon:        -122.4194,
							ShapePtSequence:   0,
							ShapeDistTraveled: floatPtr(0.0),
							RowNumber:         2,
						},
						{
							ShapeID:           "shape1",
							ShapePtLat:        37.7750,
							ShapePtLon:        -122.4195,
							ShapePtSequence:   1,
							ShapeDistTraveled: floatPtr(100.5),
							RowNumber:         3,
						},
					},
				},
			},
		},
		{
			name: "shape without distances",
			csvData: "shape_id,shape_pt_lat,shape_pt_lon,shape_pt_sequence\n" +
				"shape1,37.7749,-122.4194,0\n" +
				"shape1,37.7750,-122.4195,1",
			expected: map[string]*ShapeInfo{
				"shape1": {
					ShapeID: "shape1",
					Points: []*ShapePointDetailed{
						{
							ShapeID:           "shape1",
							ShapePtLat:        37.7749,
							ShapePtLon:        -122.4194,
							ShapePtSequence:   0,
							ShapeDistTraveled: nil,
							RowNumber:         2,
						},
						{
							ShapeID:           "shape1",
							ShapePtLat:        37.7750,
							ShapePtLon:        -122.4195,
							ShapePtSequence:   1,
							ShapeDistTraveled: nil,
							RowNumber:         3,
						},
					},
				},
			},
		},
		{
			name: "multiple shapes",
			csvData: "shape_id,shape_pt_lat,shape_pt_lon,shape_pt_sequence\n" +
				"shape1,37.7749,-122.4194,0\n" +
				"shape2,37.7751,-122.4196,0\n" +
				"shape1,37.7750,-122.4195,1",
			expected: map[string]*ShapeInfo{
				"shape1": {
					ShapeID: "shape1",
					Points: []*ShapePointDetailed{
						{ShapeID: "shape1", ShapePtLat: 37.7749, ShapePtLon: -122.4194, ShapePtSequence: 0, RowNumber: 2},
						{ShapeID: "shape1", ShapePtLat: 37.7750, ShapePtLon: -122.4195, ShapePtSequence: 1, RowNumber: 4},
					},
				},
				"shape2": {
					ShapeID: "shape2",
					Points: []*ShapePointDetailed{
						{ShapeID: "shape2", ShapePtLat: 37.7751, ShapePtLon: -122.4196, ShapePtSequence: 0, RowNumber: 3},
					},
				},
			},
		},
		{
			name: "out of order sequences get sorted",
			csvData: "shape_id,shape_pt_lat,shape_pt_lon,shape_pt_sequence\n" +
				"shape1,37.7750,-122.4195,2\n" +
				"shape1,37.7749,-122.4194,0\n" +
				"shape1,37.7751,-122.4196,1",
			expected: map[string]*ShapeInfo{
				"shape1": {
					ShapeID: "shape1",
					Points: []*ShapePointDetailed{
						{ShapeID: "shape1", ShapePtLat: 37.7749, ShapePtLon: -122.4194, ShapePtSequence: 0, RowNumber: 3},
						{ShapeID: "shape1", ShapePtLat: 37.7751, ShapePtLon: -122.4196, ShapePtSequence: 1, RowNumber: 4},
						{ShapeID: "shape1", ShapePtLat: 37.7750, ShapePtLon: -122.4195, ShapePtSequence: 2, RowNumber: 2},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			feedLoader := testutil.CreateTestFeedLoader(t, map[string]string{
				"shapes.txt": tt.csvData,
			})

			result := validator.loadShapes(feedLoader)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d shapes, got %d", len(tt.expected), len(result))
			}

			for shapeID, expectedShape := range tt.expected {
				actualShape, exists := result[shapeID]
				if !exists {
					t.Errorf("Expected shape %s not found", shapeID)
					continue
				}

				if actualShape.ShapeID != expectedShape.ShapeID {
					t.Errorf("Shape %s: expected ShapeID %s, got %s", shapeID, expectedShape.ShapeID, actualShape.ShapeID)
				}

				if len(actualShape.Points) != len(expectedShape.Points) {
					t.Errorf("Shape %s: expected %d points, got %d", shapeID, len(expectedShape.Points), len(actualShape.Points))
					continue
				}

				for i, expectedPoint := range expectedShape.Points {
					actualPoint := actualShape.Points[i]

					if actualPoint.ShapeID != expectedPoint.ShapeID {
						t.Errorf("Point %d: expected ShapeID %s, got %s", i, expectedPoint.ShapeID, actualPoint.ShapeID)
					}
					if actualPoint.ShapePtLat != expectedPoint.ShapePtLat {
						t.Errorf("Point %d: expected lat %f, got %f", i, expectedPoint.ShapePtLat, actualPoint.ShapePtLat)
					}
					if actualPoint.ShapePtLon != expectedPoint.ShapePtLon {
						t.Errorf("Point %d: expected lon %f, got %f", i, expectedPoint.ShapePtLon, actualPoint.ShapePtLon)
					}
					if actualPoint.ShapePtSequence != expectedPoint.ShapePtSequence {
						t.Errorf("Point %d: expected sequence %d, got %d", i, expectedPoint.ShapePtSequence, actualPoint.ShapePtSequence)
					}
					if actualPoint.RowNumber != expectedPoint.RowNumber {
						t.Errorf("Point %d: expected row %d, got %d", i, expectedPoint.RowNumber, actualPoint.RowNumber)
					}

					// Check distance traveled
					switch {
					case expectedPoint.ShapeDistTraveled == nil && actualPoint.ShapeDistTraveled != nil:
						t.Errorf("Point %d: expected nil distance, got %f", i, *actualPoint.ShapeDistTraveled)
					case expectedPoint.ShapeDistTraveled != nil && actualPoint.ShapeDistTraveled == nil:
						t.Errorf("Point %d: expected distance %f, got nil", i, *expectedPoint.ShapeDistTraveled)
					case expectedPoint.ShapeDistTraveled != nil && actualPoint.ShapeDistTraveled != nil:
						if *actualPoint.ShapeDistTraveled != *expectedPoint.ShapeDistTraveled {
							t.Errorf("Point %d: expected distance %f, got %f", i, *expectedPoint.ShapeDistTraveled, *actualPoint.ShapeDistTraveled)
						}
					}
				}
			}
		})
	}
}

func TestShapeValidator_New(t *testing.T) {
	validator := NewShapeValidator()
	if validator == nil {
		t.Error("NewShapeValidator() returned nil")
	}
}

// Helper function to create a float64 pointer
func floatPtr(f float64) *float64 {
	return &f
}
