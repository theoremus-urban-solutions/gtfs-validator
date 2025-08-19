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
			name: "valid shape with distances",
			files: map[string]string{
				"shapes.txt": "shape_id,shape_pt_lat,shape_pt_lon,shape_pt_sequence,shape_dist_traveled\n" +
					"shape1,37.7749,-122.4194,0,0.0\n" +
					"shape1,37.7750,-122.4195,1,100.5\n" +
					"shape1,37.7751,-122.4196,2,200.8",
				"trips.txt": "trip_id,route_id,service_id,shape_id\n" +
					"trip1,route1,service1,shape1",
			},
			expectedNoticeCodes: []string{},
			description:         "Valid shape with proper sequences and distances should not generate notices",
		},
		{
			name: "valid shape without distances",
			files: map[string]string{
				"shapes.txt": "shape_id,shape_pt_lat,shape_pt_lon,shape_pt_sequence\n" +
					"shape1,37.7749,-122.4194,0\n" +
					"shape1,37.7750,-122.4195,1\n" +
					"shape1,37.7751,-122.4196,2",
				"trips.txt": "trip_id,route_id,service_id,shape_id\n" +
					"trip1,route1,service1,shape1",
			},
			expectedNoticeCodes: []string{},
			description:         "Valid shape without distances should be acceptable",
		},
		{
			name: "insufficient shape points",
			files: map[string]string{
				"shapes.txt": "shape_id,shape_pt_lat,shape_pt_lon,shape_pt_sequence\n" +
					"shape1,37.7749,-122.4194,0",
				"trips.txt": "trip_id,route_id,service_id,shape_id\n" +
					"trip1,route1,service1,shape1",
			},
			expectedNoticeCodes: []string{"insufficient_shape_points"},
			description:         "Shape with single point should generate error",
		},
		{
			name: "duplicate shape sequence",
			files: map[string]string{
				"shapes.txt": "shape_id,shape_pt_lat,shape_pt_lon,shape_pt_sequence\n" +
					"shape1,37.7749,-122.4194,0\n" +
					"shape1,37.7750,-122.4195,1\n" +
					"shape1,37.7751,-122.4196,1", // Duplicate sequence
				"trips.txt": "trip_id,route_id,service_id,shape_id\n" +
					"trip1,route1,service1,shape1",
			},
			expectedNoticeCodes: []string{"duplicate_shape_sequence", "non_increasing_shape_sequence"},
			description:         "Duplicate sequence numbers should generate error",
		},
		{
			name: "non-increasing shape sequence",
			files: map[string]string{
				"shapes.txt": "shape_id,shape_pt_lat,shape_pt_lon,shape_pt_sequence\n" +
					"shape1,37.7749,-122.4194,0\n" +
					"shape1,37.7750,-122.4195,1\n" +
					"shape1,37.7751,-122.4196,1", // Non-increasing: 1 followed by 1 (equal sequence)
				"trips.txt": "trip_id,route_id,service_id,shape_id\n" +
					"trip1,route1,service1,shape1",
			},
			expectedNoticeCodes: []string{"duplicate_shape_sequence", "non_increasing_shape_sequence"},
			description:         "Equal sequence should generate both duplicate and non-increasing notices",
		},
		{
			name: "inconsistent shape distance",
			files: map[string]string{
				"shapes.txt": "shape_id,shape_pt_lat,shape_pt_lon,shape_pt_sequence,shape_dist_traveled\n" +
					"shape1,37.7749,-122.4194,0,0.0\n" +
					"shape1,37.7750,-122.4195,1,\n" + // Missing distance
					"shape1,37.7751,-122.4196,2,200.8",
				"trips.txt": "trip_id,route_id,service_id,shape_id\n" +
					"trip1,route1,service1,shape1",
			},
			expectedNoticeCodes: []string{"inconsistent_shape_distance"},
			description:         "Missing distance when others have it should generate error",
		},
		{
			name: "decreasing shape distance",
			files: map[string]string{
				"shapes.txt": "shape_id,shape_pt_lat,shape_pt_lon,shape_pt_sequence,shape_dist_traveled\n" +
					"shape1,37.7749,-122.4194,0,100.0\n" +
					"shape1,37.7750,-122.4195,1,50.0\n" + // Decreasing distance
					"shape1,37.7751,-122.4196,2,200.0",
				"trips.txt": "trip_id,route_id,service_id,shape_id\n" +
					"trip1,route1,service1,shape1",
			},
			expectedNoticeCodes: []string{"decreasing_shape_distance"},
			description:         "Decreasing distances should generate error",
		},
		{
			name: "equal shape distance",
			files: map[string]string{
				"shapes.txt": "shape_id,shape_pt_lat,shape_pt_lon,shape_pt_sequence,shape_dist_traveled\n" +
					"shape1,37.7749,-122.4194,0,100.0\n" +
					"shape1,37.7750,-122.4195,1,100.0\n" + // Equal distance
					"shape1,37.7751,-122.4196,2,200.0",
				"trips.txt": "trip_id,route_id,service_id,shape_id\n" +
					"trip1,route1,service1,shape1",
			},
			expectedNoticeCodes: []string{"equal_shape_distance"},
			description:         "Equal consecutive distances should generate warning",
		},
		{
			name: "duplicate shape point coordinates",
			files: map[string]string{
				"shapes.txt": "shape_id,shape_pt_lat,shape_pt_lon,shape_pt_sequence\n" +
					"shape1,37.7749,-122.4194,0\n" +
					"shape1,37.7749,-122.4194,1\n" + // Duplicate coordinates
					"shape1,37.7751,-122.4196,2",
				"trips.txt": "trip_id,route_id,service_id,shape_id\n" +
					"trip1,route1,service1,shape1",
			},
			expectedNoticeCodes: []string{"duplicate_shape_point"},
			description:         "Duplicate consecutive coordinates should generate warning",
		},
		{
			name: "unreasonably long shape segment",
			files: map[string]string{
				"shapes.txt": "shape_id,shape_pt_lat,shape_pt_lon,shape_pt_sequence\n" +
					"shape1,37.7749,-122.4194,0\n" +
					"shape1,38.7749,-121.4194,1\n" + // ~100km+ distance
					"shape1,37.7751,-122.4196,2",
				"trips.txt": "trip_id,route_id,service_id,shape_id\n" +
					"trip1,route1,service1,shape1",
			},
			expectedNoticeCodes: []string{"unreasonably_long_shape_segment", "unreasonably_long_shape_segment"},
			description:         "Very long shape segments should generate warning",
		},
		{
			name: "unused shape",
			files: map[string]string{
				"shapes.txt": "shape_id,shape_pt_lat,shape_pt_lon,shape_pt_sequence\n" +
					"shape1,37.7749,-122.4194,0\n" +
					"shape1,37.7750,-122.4195,1\n" +
					"shape2,37.7751,-122.4196,0\n" +
					"shape2,37.7752,-122.4197,1", // shape2 is unused
				"trips.txt": "trip_id,route_id,service_id,shape_id\n" +
					"trip1,route1,service1,shape1", // Only shape1 is used
			},
			expectedNoticeCodes: []string{"unused_shape"},
			description:         "Unused shapes should generate warning",
		},
		{
			name: "multiple validation issues",
			files: map[string]string{
				"shapes.txt": "shape_id,shape_pt_lat,shape_pt_lon,shape_pt_sequence,shape_dist_traveled\n" +
					"shape1,37.7749,-122.4194,0,0.0\n" +
					"shape1,37.7749,-122.4194,2,\n" + // Duplicate coords + missing distance + no sequence conflict
					"shape1,37.7751,-122.4196,3", // No duplicate sequence
				"trips.txt": "trip_id,route_id,service_id,shape_id\n" +
					"trip1,route1,service1,shape1",
			},
			expectedNoticeCodes: []string{"duplicate_shape_point", "inconsistent_shape_distance"},
			description:         "Multiple validation issues should generate multiple notices",
		},
		{
			name: "shape with negative sequences",
			files: map[string]string{
				"shapes.txt": "shape_id,shape_pt_lat,shape_pt_lon,shape_pt_sequence\n" +
					"shape1,37.7749,-122.4194,-1\n" +
					"shape1,37.7750,-122.4195,0\n" +
					"shape1,37.7751,-122.4196,1",
				"trips.txt": "trip_id,route_id,service_id,shape_id\n" +
					"trip1,route1,service1,shape1",
			},
			expectedNoticeCodes: []string{},
			description:         "Negative sequence numbers should be valid if properly ordered",
		},
		{
			name: "shape with large sequence gaps",
			files: map[string]string{
				"shapes.txt": "shape_id,shape_pt_lat,shape_pt_lon,shape_pt_sequence\n" +
					"shape1,37.7749,-122.4194,0\n" +
					"shape1,37.7750,-122.4195,1000\n" +
					"shape1,37.7751,-122.4196,2000",
				"trips.txt": "trip_id,route_id,service_id,shape_id\n" +
					"trip1,route1,service1,shape1",
			},
			expectedNoticeCodes: []string{},
			description:         "Large gaps in sequence numbers should be acceptable",
		},
		{
			name: "multiple shapes mixed validation",
			files: map[string]string{
				"shapes.txt": "shape_id,shape_pt_lat,shape_pt_lon,shape_pt_sequence\n" +
					"shape1,37.7749,-122.4194,0\n" +
					"shape1,37.7750,-122.4195,1\n" + // Valid shape
					"shape2,37.7751,-122.4196,0", // Single point - invalid
				"trips.txt": "trip_id,route_id,service_id,shape_id\n" +
					"trip1,route1,service1,shape1", // shape2 unused
			},
			expectedNoticeCodes: []string{"insufficient_shape_points", "unused_shape"},
			description:         "Mixed valid and invalid shapes should generate appropriate notices",
		},
		{
			name: "shape with coordinate precision",
			files: map[string]string{
				"shapes.txt": "shape_id,shape_pt_lat,shape_pt_lon,shape_pt_sequence\n" +
					"shape1,37.7749000,-122.4194000,0\n" +
					"shape1,37.7749001,-122.4194001,1\n" + // Very close but not duplicate
					"shape1,37.7751000,-122.4196000,2",
				"trips.txt": "trip_id,route_id,service_id,shape_id\n" +
					"trip1,route1,service1,shape1",
			},
			expectedNoticeCodes: []string{},
			description:         "Very close but distinct coordinates should be valid",
		},
		{
			name: "no shapes file",
			files: map[string]string{
				"trips.txt": "trip_id,route_id,service_id\n" +
					"trip1,route1,service1", // No shape_id
			},
			expectedNoticeCodes: []string{},
			description:         "Missing shapes.txt file should not generate errors",
		},
		{
			name: "shapes with missing required fields",
			files: map[string]string{
				"shapes.txt": "shape_id,shape_pt_lat,shape_pt_lon\n" +
					"shape1,37.7749,-122.4194\n" + // Missing sequence
					"shape_pt_lat,shape_pt_lon,shape_pt_sequence\n" +
					"37.7750,-122.4195,0", // Missing shape_id
				"trips.txt": "trip_id,route_id,service_id,shape_id\n" +
					"trip1,route1,service1,shape1",
			},
			expectedNoticeCodes: []string{},
			description:         "Shapes with missing required fields should be ignored",
		},
		{
			name: "whitespace handling in shapes",
			files: map[string]string{
				"shapes.txt": "shape_id,shape_pt_lat,shape_pt_lon,shape_pt_sequence,shape_dist_traveled\n" +
					" shape1 , 37.7749 , -122.4194 , 0 , 0.0 \n" +
					" shape1 , 37.7750 , -122.4195 , 1 , 100.5 ",
				"trips.txt": "trip_id,route_id,service_id,shape_id\n" +
					"trip1,route1,service1, shape1 ",
			},
			expectedNoticeCodes: []string{},
			description:         "Whitespace should be trimmed properly",
		},
		{
			name: "all shapes used",
			files: map[string]string{
				"shapes.txt": "shape_id,shape_pt_lat,shape_pt_lon,shape_pt_sequence\n" +
					"shape1,37.7749,-122.4194,0\n" +
					"shape1,37.7750,-122.4195,1\n" +
					"shape2,37.7751,-122.4196,0\n" +
					"shape2,37.7752,-122.4197,1",
				"trips.txt": "trip_id,route_id,service_id,shape_id\n" +
					"trip1,route1,service1,shape1\n" +
					"trip2,route2,service2,shape2",
			},
			expectedNoticeCodes: []string{},
			description:         "All shapes being used should not generate unused notices",
		},
		{
			name: "empty shape_id in trips",
			files: map[string]string{
				"shapes.txt": "shape_id,shape_pt_lat,shape_pt_lon,shape_pt_sequence\n" +
					"shape1,37.7749,-122.4194,0\n" +
					"shape1,37.7750,-122.4195,1",
				"trips.txt": "trip_id,route_id,service_id,shape_id\n" +
					"trip1,route1,service1,\n" + // Empty shape_id
					"trip2,route2,service2,shape1",
			},
			expectedNoticeCodes: []string{},
			description:         "Empty shape_id in trips should be ignored for usage tracking",
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

func TestShapeValidator_LoadUsedShapes(t *testing.T) {
	validator := NewShapeValidator()

	tests := []struct {
		name     string
		csvData  string
		expected map[string]bool
	}{
		{
			name: "basic used shapes",
			csvData: "trip_id,route_id,service_id,shape_id\n" +
				"trip1,route1,service1,shape1\n" +
				"trip2,route1,service1,shape2",
			expected: map[string]bool{
				"shape1": true,
				"shape2": true,
			},
		},
		{
			name: "duplicate shape usage",
			csvData: "trip_id,route_id,service_id,shape_id\n" +
				"trip1,route1,service1,shape1\n" +
				"trip2,route1,service1,shape1",
			expected: map[string]bool{
				"shape1": true,
			},
		},
		{
			name: "empty and missing shape_id",
			csvData: "trip_id,route_id,service_id,shape_id\n" +
				"trip1,route1,service1,shape1\n" +
				"trip2,route1,service1,\n" + // Empty shape_id
				"trip_id,route_id,service_id\n" +
				"trip3,route1,service1", // Missing shape_id column
			expected: map[string]bool{
				"shape1": true,
			},
		},
		{
			name: "whitespace trimming",
			csvData: "trip_id,route_id,service_id,shape_id\n" +
				"trip1,route1,service1, shape1 \n" +
				"trip2,route1,service1,  shape2  ",
			expected: map[string]bool{
				"shape1": true,
				"shape2": true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			feedLoader := testutil.CreateTestFeedLoader(t, map[string]string{
				"trips.txt": tt.csvData,
			})

			result := validator.loadUsedShapes(feedLoader)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d used shapes, got %d", len(tt.expected), len(result))
			}

			for shapeID, expected := range tt.expected {
				if actual, exists := result[shapeID]; !exists || actual != expected {
					t.Errorf("Shape %s: expected %v, got %v", shapeID, expected, actual)
				}
			}
		})
	}
}

func TestShapeValidator_HaversineDistance(t *testing.T) {
	validator := NewShapeValidator()

	tests := []struct {
		name                   string
		lat1, lon1, lat2, lon2 float64
		expected               float64
		tolerance              float64
	}{
		{
			name: "same point",
			lat1: 37.7749, lon1: -122.4194,
			lat2: 37.7749, lon2: -122.4194,
			expected: 0.0, tolerance: 1.0,
		},
		{
			name: "short distance",
			lat1: 37.7749, lon1: -122.4194,
			lat2: 37.7750, lon2: -122.4195,
			expected: 14.2, tolerance: 5.0, // Approximately 14.2 meters
		},
		{
			name: "long distance",
			lat1: 37.7749, lon1: -122.4194, // San Francisco
			lat2: 40.7128, lon2: -74.0060, // New York
			expected: 4139000.0, tolerance: 10000.0, // Approximately 4139 km
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.haversineDistance(tt.lat1, tt.lon1, tt.lat2, tt.lon2)

			if result < tt.expected-tt.tolerance || result > tt.expected+tt.tolerance {
				t.Errorf("Expected distance around %.1f (±%.1f), got %.1f",
					tt.expected, tt.tolerance, result)
			}
		})
	}
}

func TestShapeValidator_ApproximatelyEqual(t *testing.T) {
	validator := NewShapeValidator()

	tests := []struct {
		name     string
		a, b     float64
		epsilon  float64
		expected bool
	}{
		{"exactly equal", 1.0, 1.0, 1e-7, true},
		{"within epsilon", 1.0, 1.0000001, 1e-6, true},
		{"outside epsilon", 1.0, 1.001, 1e-6, false},
		{"negative values", -1.0, -1.0000001, 1e-6, true},
		{"zero comparison", 0.0, 0.0000001, 1e-6, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.approximatelyEqual(tt.a, tt.b, tt.epsilon)
			if result != tt.expected {
				t.Errorf("Expected %v for approximatelyEqual(%f, %f, %e), got %v",
					tt.expected, tt.a, tt.b, tt.epsilon, result)
			}
		})
	}
}

func TestShapeValidator_ValidateShapeSequence(t *testing.T) {
	validator := NewShapeValidator()

	tests := []struct {
		name                string
		shape               *ShapeInfo
		expectedNoticeCodes []string
	}{
		{
			name: "valid sequence",
			shape: &ShapeInfo{
				ShapeID: "shape1",
				Points: []*ShapePointDetailed{
					{ShapeID: "shape1", ShapePtSequence: 0, RowNumber: 1},
					{ShapeID: "shape1", ShapePtSequence: 1, RowNumber: 2},
					{ShapeID: "shape1", ShapePtSequence: 2, RowNumber: 3},
				},
			},
			expectedNoticeCodes: []string{},
		},
		{
			name: "duplicate sequence",
			shape: &ShapeInfo{
				ShapeID: "shape1",
				Points: []*ShapePointDetailed{
					{ShapeID: "shape1", ShapePtSequence: 0, RowNumber: 1},
					{ShapeID: "shape1", ShapePtSequence: 1, RowNumber: 2},
					{ShapeID: "shape1", ShapePtSequence: 1, RowNumber: 3}, // Duplicate
				},
			},
			expectedNoticeCodes: []string{"duplicate_shape_sequence", "non_increasing_shape_sequence"},
		},
		{
			name: "non-increasing sequence",
			shape: &ShapeInfo{
				ShapeID: "shape1",
				Points: []*ShapePointDetailed{
					{ShapeID: "shape1", ShapePtSequence: 0, RowNumber: 1},
					{ShapeID: "shape1", ShapePtSequence: 2, RowNumber: 2},
					{ShapeID: "shape1", ShapePtSequence: 1, RowNumber: 3}, // Non-increasing (after sort)
				},
			},
			expectedNoticeCodes: []string{"non_increasing_shape_sequence"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			container := notice.NewNoticeContainer()

			validator.validateShapeSequence(container, tt.shape)

			notices := container.GetNotices()
			var actualNoticeCodes []string
			for _, n := range notices {
				actualNoticeCodes = append(actualNoticeCodes, n.Code())
			}

			expectedSet := make(map[string]bool)
			for _, code := range tt.expectedNoticeCodes {
				expectedSet[code] = true
			}

			for _, code := range actualNoticeCodes {
				if !expectedSet[code] {
					t.Errorf("Unexpected notice code: %s", code)
				}
			}

			for expectedCode := range expectedSet {
				found := false
				for _, actualCode := range actualNoticeCodes {
					if actualCode == expectedCode {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected notice code '%s' not found", expectedCode)
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
