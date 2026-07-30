package business

import (
	"math"
	"testing"
)

// Fixtures sit on the equator, where a degree of longitude is 111.32 km and a
// degree of latitude 110.57 km, so expected distances can be read off the
// coordinates.

func shapeDistance(value float64) *float64 {
	return &value
}

func TestShapeIndex_Cumulative(t *testing.T) {
	index := newShapeIndex([]shapePoint{
		{Lat: 0, Lon: 0},
		{Lat: 0, Lon: 0.01},
		{Lat: 0, Lon: 0.02},
	})

	tests := []struct {
		name     string
		at       int
		expected float64
	}{
		{name: "first point is the origin", at: 0, expected: 0},
		{name: "one hop east", at: 1, expected: 1113},
		{name: "two hops east", at: 2, expected: 2226},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := index.cumulative[tt.at]; math.Abs(got-tt.expected) > 5 {
				t.Errorf("cumulative[%d] = %.1f, want about %.1f", tt.at, got, tt.expected)
			}
		})
	}
}

func TestShapeIndex_SegmentDistance(t *testing.T) {
	index := newShapeIndex([]shapePoint{
		{Lat: 0, Lon: 0},
		{Lat: 0, Lon: 0.01},
	})

	tests := []struct {
		name             string
		lat, lon         float64
		expectedMetres   float64
		expectedFraction float64
		description      string
	}{
		{
			name: "on the segment", lat: 0, lon: 0.005,
			expectedMetres: 0, expectedFraction: 0.5,
			description: "the midpoint projects halfway along",
		},
		{
			name: "beside the segment", lat: 0.001, lon: 0.005,
			expectedMetres: 110.6, expectedFraction: 0.5,
			description: "perpendicular offset is the latitude difference",
		},
		{
			name: "beyond the far end", lat: 0, lon: 0.02,
			expectedMetres: 1113, expectedFraction: 1,
			description: "projection clamps to the end of the segment, not past it",
		},
		{
			name: "before the near end", lat: 0, lon: -0.01,
			expectedMetres: 1113, expectedFraction: 0,
			description: "and to the start at the other end",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metres, fraction := index.segmentDistance(tt.lat, tt.lon, 0)
			if math.Abs(metres-tt.expectedMetres) > 5 {
				t.Errorf("distance = %.1f m, want about %.1f m (%s)", metres, tt.expectedMetres, tt.description)
			}
			if math.Abs(fraction-tt.expectedFraction) > 0.01 {
				t.Errorf("fraction = %.3f, want %.3f (%s)", fraction, tt.expectedFraction, tt.description)
			}
		})
	}
}

func TestShapeIndex_Nearby(t *testing.T) {
	// An L: east for 1113 m, then north for 1106 m.
	index := newShapeIndex([]shapePoint{
		{Lat: 0, Lon: 0},
		{Lat: 0, Lon: 0.01},
		{Lat: 0.01, Lon: 0.01},
	})

	tests := []struct {
		name        string
		lat, lon    float64
		radius      float64
		expectMatch bool
		description string
	}{
		{
			name: "on the first leg", lat: 0, lon: 0.005, radius: 100,
			expectMatch: true,
			description: "a point on the polyline matches at zero distance",
		},
		{
			name: "just inside the radius", lat: 0.0008, lon: 0.005, radius: 100,
			expectMatch: true,
			description: "88 m off the leg is within the 100 m match radius",
		},
		{
			name: "outside the radius", lat: 0.002, lon: 0.005, radius: 100,
			expectMatch: false,
			description: "221 m off the leg is not",
		},
		{
			name: "on the second leg", lat: 0.005, lon: 0.01, radius: 100,
			expectMatch: true,
			description: "the grid indexes every segment, not just the first",
		},
		{
			name: "in a different city", lat: 40, lon: -70, radius: 100,
			expectMatch: false,
			description: "cells the shape never occupies return nothing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := index.nearby(tt.lat, tt.lon, tt.radius, nil)
			if got := len(matches) > 0; got != tt.expectMatch {
				t.Errorf("matched = %v, want %v (%s)", got, tt.expectMatch, tt.description)
			}
			for _, match := range matches {
				if match.Metres > tt.radius {
					t.Errorf("match at %.1f m exceeds the %.1f m radius", match.Metres, tt.radius)
				}
			}
		})
	}
}

func TestShapeIndex_NearbyFindsOversizedSegments(t *testing.T) {
	// A single segment spanning two degrees covers far more cells than the
	// index will file it under, so it has to be found through the oversized
	// list instead.
	index := newShapeIndex([]shapePoint{
		{Lat: 0, Lon: 0},
		{Lat: 0, Lon: 2},
	})

	if len(index.oversized) == 0 {
		t.Fatalf("expected the long segment to be held as oversized, cells: %d", len(index.cells))
	}

	if matches := index.nearby(0, 1, 100, nil); len(matches) == 0 {
		t.Errorf("expected a match halfway along the oversized segment")
	}
}

func TestShapeIndex_Nearest(t *testing.T) {
	index := newShapeIndex([]shapePoint{
		{Lat: 0, Lon: 0},
		{Lat: 0, Lon: 0.01},
	})

	nearest := index.nearest(0.01, 0.005)
	if math.Abs(nearest.Metres-1106) > 10 {
		t.Errorf("nearest = %.1f m, want about 1106 m", nearest.Metres)
	}
	if math.Abs(nearest.Along-556) > 10 {
		t.Errorf("nearest along = %.1f m, want about 556 m", nearest.Along)
	}
}

func TestShapeIndex_LocateByUserDistance(t *testing.T) {
	increasing := newShapeIndex([]shapePoint{
		{Lat: 0, Lon: 0, Dist: shapeDistance(0)},
		{Lat: 0, Lon: 0.01, Dist: shapeDistance(1000)},
	})

	tests := []struct {
		name        string
		index       *shapeIndex
		distance    float64
		expectOK    bool
		expectedLon float64
		description string
	}{
		{
			name: "at the start", index: increasing, distance: 0,
			expectOK: true, expectedLon: 0,
			description: "zero distance is the first point",
		},
		{
			name: "halfway", index: increasing, distance: 500,
			expectOK: true, expectedLon: 0.005,
			description: "distances between two points interpolate linearly",
		},
		{
			name: "past the end", index: increasing, distance: 5000,
			expectOK: true, expectedLon: 0.01,
			description: "beyond the shape clamps to its last point",
		},
		{
			name: "before the start", index: increasing, distance: -100,
			expectOK: true, expectedLon: 0,
			description: "and below it to the first",
		},
		{
			name: "decreasing distances", distance: 500, expectOK: false,
			index: newShapeIndex([]shapePoint{
				{Lat: 0, Lon: 0, Dist: shapeDistance(1000)},
				{Lat: 0, Lon: 0.01, Dist: shapeDistance(0)},
			}),
			description: "a shape whose distances run backwards cannot be interpolated",
		},
		{
			name: "no distances at all", distance: 500, expectOK: false,
			index: newShapeIndex([]shapePoint{
				{Lat: 0, Lon: 0},
				{Lat: 0, Lon: 0.01},
			}),
			description: "shape_dist_traveled is optional",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, lon, ok := tt.index.locateByUserDistance(tt.distance)
			if ok != tt.expectOK {
				t.Fatalf("ok = %v, want %v (%s)", ok, tt.expectOK, tt.description)
			}
			if ok && math.Abs(lon-tt.expectedLon) > 1e-6 {
				t.Errorf("lon = %f, want %f (%s)", lon, tt.expectedLon, tt.description)
			}
		})
	}
}

func TestClusterMatches(t *testing.T) {
	tests := []struct {
		name           string
		matches        []shapeMatch
		expectedPasses int
		description    string
	}{
		{
			name:           "no matches",
			matches:        nil,
			expectedPasses: 0,
			description:    "a stop the shape never approaches has no passes",
		},
		{
			name: "neighbouring segments are one pass",
			matches: []shapeMatch{
				{Along: 100, Metres: 40}, {Along: 180, Metres: 10}, {Along: 260, Metres: 55},
			},
			expectedPasses: 1,
			description:    "a straight run past a stop matches many segments at once",
		},
		{
			name: "a loop returning to the stop is two passes",
			matches: []shapeMatch{
				{Along: 100, Metres: 20}, {Along: 2000, Metres: 30},
			},
			expectedPasses: 2,
			description:    "matches further apart than the separation are separate passes",
		},
		{
			name: "unsorted input",
			matches: []shapeMatch{
				{Along: 2000, Metres: 30}, {Along: 100, Metres: 20},
			},
			expectedPasses: 2,
			description:    "clustering orders the matches itself",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			passes := clusterMatches(tt.matches)
			if len(passes) != tt.expectedPasses {
				t.Errorf("got %d passes, want %d (%s)", len(passes), tt.expectedPasses, tt.description)
			}
		})
	}

	t.Run("a pass keeps its closest match", func(t *testing.T) {
		passes := clusterMatches([]shapeMatch{
			{Along: 100, Metres: 40}, {Along: 180, Metres: 10}, {Along: 260, Metres: 55},
		})
		if passes[0].Metres != 10 {
			t.Errorf("pass distance = %.0f, want the closest match of the run (10)", passes[0].Metres)
		}
	})
}

func TestChooseMatch(t *testing.T) {
	tests := []struct {
		name          string
		passes        []shapeMatch
		expectedAlong float64
		description   string
	}{
		{
			name:          "first stop anchors at the earliest pass, not the closest",
			passes:        []shapeMatch{{Along: 900, Metres: 5}, {Along: 100, Metres: 60}},
			expectedAlong: 100,
			description: "a loop's first stop sits near both ends; anchoring at the near " +
				"end would put the trip behind its own start and report every " +
				"following stop as out of order",
		},
		{
			name:          "first stop with a single pass takes it",
			passes:        []shapeMatch{{Along: 400, Metres: 8}},
			expectedAlong: 400,
			description:   "a trip starting mid-shape has only one place it can begin",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if chosen := chooseMatch(tt.passes); chosen.Along != tt.expectedAlong {
				t.Errorf("chose along %.0f, want %.0f (%s)", chosen.Along, tt.expectedAlong, tt.description)
			}
		})
	}
}
