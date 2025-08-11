package relationship

import (
	"io"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/parser"
	"github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

// ShapeIncreasingDistanceValidator validates that shape_dist_traveled increases along shapes
type ShapeIncreasingDistanceValidator struct{}

// NewShapeIncreasingDistanceValidator creates a new shape increasing distance validator
func NewShapeIncreasingDistanceValidator() *ShapeIncreasingDistanceValidator {
	return &ShapeIncreasingDistanceValidator{}
}

// ShapePointDistance represents a shape point with distance information
type ShapePointDistance struct {
	ShapeID      string
	Sequence     int
	Latitude     float64
	Longitude    float64
	DistTraveled *float64
	RowNumber    int
}

// Validate checks that shape distances increase monotonically
func (v *ShapeIncreasingDistanceValidator) Validate(loader *parser.FeedLoader, container *notice.NoticeContainer, config validator.Config) {
	shapes := v.loadShapePoints(loader)

	// Group by shape_id and validate each shape
	for shapeID, points := range shapes {
		v.validateShapeDistances(container, shapeID, points)
	}
}

// loadShapePoints loads shape points from shapes.txt
func (v *ShapeIncreasingDistanceValidator) loadShapePoints(loader *parser.FeedLoader) map[string][]*ShapePointDistance {
	shapes := make(map[string][]*ShapePointDistance)

	reader, err := loader.GetFile("shapes.txt")
	if err != nil {
		return shapes
	}
	defer reader.Close()

	csvFile, err := parser.NewCSVFile(reader, "shapes.txt")
	if err != nil {
		return shapes
	}

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		point := v.parseShapePoint(row)
		if point != nil {
			shapes[point.ShapeID] = append(shapes[point.ShapeID], point)
		}
	}

	// Sort each shape by sequence
	for shapeID := range shapes {
		sort.Slice(shapes[shapeID], func(i, j int) bool {
			return shapes[shapeID][i].Sequence < shapes[shapeID][j].Sequence
		})
	}

	return shapes
}

// parseShapePoint parses a shape point from shapes.txt
func (v *ShapeIncreasingDistanceValidator) parseShapePoint(row *parser.CSVRow) *ShapePointDistance {
	shapeID, hasShapeID := row.Values["shape_id"]
	latStr, hasLat := row.Values["shape_pt_lat"]
	lonStr, hasLon := row.Values["shape_pt_lon"]
	seqStr, hasSeq := row.Values["shape_pt_sequence"]

	if !hasShapeID || !hasLat || !hasLon || !hasSeq {
		return nil
	}

	lat, err1 := strconv.ParseFloat(strings.TrimSpace(latStr), 64)
	lon, err2 := strconv.ParseFloat(strings.TrimSpace(lonStr), 64)
	seq, err3 := strconv.Atoi(strings.TrimSpace(seqStr))

	if err1 != nil || err2 != nil || err3 != nil {
		return nil
	}

	point := &ShapePointDistance{
		ShapeID:   strings.TrimSpace(shapeID),
		Sequence:  seq,
		Latitude:  lat,
		Longitude: lon,
		RowNumber: row.RowNumber,
	}

	// Parse optional distance traveled
	if distStr, hasDist := row.Values["shape_dist_traveled"]; hasDist && strings.TrimSpace(distStr) != "" {
		if dist, err := strconv.ParseFloat(strings.TrimSpace(distStr), 64); err == nil {
			point.DistTraveled = &dist
		}
	}

	return point
}

// validateShapeDistances validates distance values for a shape
func (v *ShapeIncreasingDistanceValidator) validateShapeDistances(container *notice.NoticeContainer, shapeID string, points []*ShapePointDistance) {
	if len(points) < 2 {
		return
	}

	// Check if any points have distance values
	hasAnyDistance := false
	for _, point := range points {
		if point.DistTraveled != nil {
			hasAnyDistance = true
			break
		}
	}

	if !hasAnyDistance {
		return // No distance values to validate
	}

	// Validate distance consistency
	v.validateDistanceIncreasing(container, shapeID, points)
	v.validateDistanceRealism(container, shapeID, points)
	v.validateDistanceCompleteness(container, shapeID, points)
}

// validateDistanceIncreasing checks that distances increase along the shape
func (v *ShapeIncreasingDistanceValidator) validateDistanceIncreasing(container *notice.NoticeContainer, shapeID string, points []*ShapePointDistance) {
	var prevDistance *float64
	var prevPoint *ShapePointDistance

	for _, point := range points {
		if point.DistTraveled != nil {
			if prevDistance != nil && prevPoint != nil {
				if *point.DistTraveled < *prevDistance {
					// Distance decreased
					container.AddNotice(notice.NewShapeDistanceDecreasingNotice(
						shapeID,
						prevPoint.Sequence,
						point.Sequence,
						*prevDistance,
						*point.DistTraveled,
						point.RowNumber,
					))
				} else if *point.DistTraveled == *prevDistance && point.Sequence != prevPoint.Sequence {
					// Distance stayed the same between different sequence points
					// This might be acceptable for very short segments, but worth noting
					if v.calculateDistance(prevPoint, point) > 10 { // > 10 meters apart but same distance
						container.AddNotice(notice.NewShapeDistanceNotIncreasingNotice(
							shapeID,
							prevPoint.Sequence,
							point.Sequence,
							*point.DistTraveled,
							point.RowNumber,
						))
					}
				}
			}
			prevDistance = point.DistTraveled
			prevPoint = point
		}
	}
}

// validateDistanceRealism checks that distances are realistic compared to geography
func (v *ShapeIncreasingDistanceValidator) validateDistanceRealism(container *notice.NoticeContainer, shapeID string, points []*ShapePointDistance) {
	var cumulativeGeoDistance float64
	var prevPoint *ShapePointDistance

	for _, point := range points {
		if prevPoint != nil {
			segmentDistance := v.calculateDistance(prevPoint, point)
			cumulativeGeoDistance += segmentDistance

			// If both points have distance values, check consistency
			if point.DistTraveled != nil && prevPoint.DistTraveled != nil {
				providedDistance := *point.DistTraveled - *prevPoint.DistTraveled

				// Check if provided distance is unrealistically different from geographic distance
				if segmentDistance > 0 {
					ratio := providedDistance / segmentDistance

					// Flag if ratio is very unrealistic (< 0.5 or > 3.0)
					if ratio < 0.5 || ratio > 3.0 {
						container.AddNotice(notice.NewUnrealisticShapeDistanceNotice(
							shapeID,
							prevPoint.Sequence,
							point.Sequence,
							providedDistance,
							segmentDistance,
							ratio,
							point.RowNumber,
						))
					}
				}
			}
		}
		prevPoint = point
	}
}

// validateDistanceCompleteness checks for missing distance values in sequences
func (v *ShapeIncreasingDistanceValidator) validateDistanceCompleteness(container *notice.NoticeContainer, shapeID string, points []*ShapePointDistance) {
	hasDistanceValues := 0
	totalPoints := len(points)

	for _, point := range points {
		if point.DistTraveled != nil {
			hasDistanceValues++
		}
	}

	// If some but not all points have distance values, this might be inconsistent
	if hasDistanceValues > 0 && hasDistanceValues < totalPoints {
		missingCount := totalPoints - hasDistanceValues

		// Report if significant portion is missing (> 25%)
		if float64(missingCount)/float64(totalPoints) > 0.25 {
			container.AddNotice(notice.NewIncompleteShapeDistanceNotice(
				shapeID,
				hasDistanceValues,
				totalPoints,
				missingCount,
			))
		}
	}

	// Check for specific problematic patterns
	v.validateDistancePatterns(container, shapeID, points)
}

// validateDistancePatterns checks for specific problematic distance patterns
func (v *ShapeIncreasingDistanceValidator) validateDistancePatterns(container *notice.NoticeContainer, shapeID string, points []*ShapePointDistance) {
	// Check for distance values that start from a non-zero value
	firstPointWithDistance := -1
	for i, point := range points {
		if point.DistTraveled != nil {
			firstPointWithDistance = i
			break
		}
	}

	if firstPointWithDistance > 0 {
		// Distance values don't start from the first point
		firstPoint := points[firstPointWithDistance]
		if *firstPoint.DistTraveled > 100 { // Starting distance > 100 units
			container.AddNotice(notice.NewShapeDistanceNotStartingFromZeroNotice(
				shapeID,
				firstPoint.Sequence,
				*firstPoint.DistTraveled,
				firstPoint.RowNumber,
			))
		}
	}

	// Check for very large jumps in distance
	var prevDistance *float64
	var prevPoint *ShapePointDistance

	for _, point := range points {
		if point.DistTraveled != nil && prevDistance != nil && prevPoint != nil {
			jump := *point.DistTraveled - *prevDistance
			geoDistance := v.calculateDistance(prevPoint, point)

			// Flag jumps that are 10x larger than geographic distance
			if geoDistance > 0 && jump > geoDistance*10 {
				container.AddNotice(notice.NewLargeShapeDistanceJumpNotice(
					shapeID,
					prevPoint.Sequence,
					point.Sequence,
					jump,
					geoDistance,
					point.RowNumber,
				))
			}
		}

		if point.DistTraveled != nil {
			prevDistance = point.DistTraveled
			prevPoint = point
		}
	}
}

// calculateDistance calculates Haversine distance between two points in meters
func (v *ShapeIncreasingDistanceValidator) calculateDistance(point1, point2 *ShapePointDistance) float64 {
	const earthRadius = 6371000 // Earth radius in meters

	lat1Rad := point1.Latitude * math.Pi / 180
	lat2Rad := point2.Latitude * math.Pi / 180
	deltaLatRad := (point2.Latitude - point1.Latitude) * math.Pi / 180
	deltaLonRad := (point2.Longitude - point1.Longitude) * math.Pi / 180

	a := math.Sin(deltaLatRad/2)*math.Sin(deltaLatRad/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLonRad/2)*math.Sin(deltaLonRad/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c
}
