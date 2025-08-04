package entity

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

// ShapeValidator validates shape definitions and usage
type ShapeValidator struct{}

// NewShapeValidator creates a new shape validator
func NewShapeValidator() *ShapeValidator {
	return &ShapeValidator{}
}

// ShapePointDetailed represents a single shape point with coordinates
type ShapePointDetailed struct {
	ShapeID           string
	ShapePtLat        float64
	ShapePtLon        float64
	ShapePtSequence   int
	ShapeDistTraveled *float64
	RowNumber         int
}

// ShapeInfo represents aggregated shape information
type ShapeInfo struct {
	ShapeID string
	Points  []*ShapePointDetailed
}

// Validate checks shape definitions and consistency
func (v *ShapeValidator) Validate(loader *parser.FeedLoader, container *notice.NoticeContainer, config validator.Config) {
	shapes := v.loadShapes(loader)
	if len(shapes) == 0 {
		return // No shapes to validate
	}

	// Validate each shape
	for _, shape := range shapes {
		v.validateShape(container, shape)
	}

	// Validate shape usage
	v.validateShapeUsage(loader, container, shapes)
}

// loadShapes loads shape information from shapes.txt
func (v *ShapeValidator) loadShapes(loader *parser.FeedLoader) map[string]*ShapeInfo {
	shapes := make(map[string]*ShapeInfo)

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
			break
		}

		point := v.parseShapePoint(row)
		if point != nil {
			if shapes[point.ShapeID] == nil {
				shapes[point.ShapeID] = &ShapeInfo{
					ShapeID: point.ShapeID,
					Points:  []*ShapePointDetailed{},
				}
			}
			shapes[point.ShapeID].Points = append(shapes[point.ShapeID].Points, point)
		}
	}

	// Sort points by sequence for each shape
	for _, shape := range shapes {
		sort.Slice(shape.Points, func(i, j int) bool {
			return shape.Points[i].ShapePtSequence < shape.Points[j].ShapePtSequence
		})
	}

	return shapes
}

// parseShapePoint parses a shape point record
func (v *ShapeValidator) parseShapePoint(row *parser.CSVRow) *ShapePointDetailed {
	shapeID, hasShapeID := row.Values["shape_id"]
	shapePtLatStr, hasShapePtLat := row.Values["shape_pt_lat"]
	shapePtLonStr, hasShapePtLon := row.Values["shape_pt_lon"]
	shapePtSequenceStr, hasShapePtSequence := row.Values["shape_pt_sequence"]

	if !hasShapeID || !hasShapePtLat || !hasShapePtLon || !hasShapePtSequence {
		return nil
	}

	shapePtLat, err := strconv.ParseFloat(strings.TrimSpace(shapePtLatStr), 64)
	if err != nil {
		return nil
	}

	shapePtLon, err := strconv.ParseFloat(strings.TrimSpace(shapePtLonStr), 64)
	if err != nil {
		return nil
	}

	shapePtSequence, err := strconv.Atoi(strings.TrimSpace(shapePtSequenceStr))
	if err != nil {
		return nil
	}

	point := &ShapePointDetailed{
		ShapeID:         strings.TrimSpace(shapeID),
		ShapePtLat:      shapePtLat,
		ShapePtLon:      shapePtLon,
		ShapePtSequence: shapePtSequence,
		RowNumber:       row.RowNumber,
	}

	// Parse optional shape_dist_traveled
	if shapeDistTraveledStr, hasShapeDistTraveled := row.Values["shape_dist_traveled"]; hasShapeDistTraveled && strings.TrimSpace(shapeDistTraveledStr) != "" {
		if shapeDistTraveled, err := strconv.ParseFloat(strings.TrimSpace(shapeDistTraveledStr), 64); err == nil {
			point.ShapeDistTraveled = &shapeDistTraveled
		}
	}

	return point
}

// validateShape validates a single shape
func (v *ShapeValidator) validateShape(container *notice.NoticeContainer, shape *ShapeInfo) {
	if len(shape.Points) < 2 {
		container.AddNotice(notice.NewInsufficientShapePointsNotice(
			shape.ShapeID,
			len(shape.Points),
		))
		return
	}

	// Validate sequence numbers
	v.validateShapeSequence(container, shape)

	// Validate shape distances
	v.validateShapeDistances(container, shape)

	// Validate shape geometry
	v.validateShapeGeometry(container, shape)
}

// validateShapeSequence validates shape point sequence numbers
func (v *ShapeValidator) validateShapeSequence(container *notice.NoticeContainer, shape *ShapeInfo) {
	sequenceMap := make(map[int]*ShapePointDetailed)

	for _, point := range shape.Points {
		if existingPoint, exists := sequenceMap[point.ShapePtSequence]; exists {
			container.AddNotice(notice.NewDuplicateShapeSequenceNotice(
				shape.ShapeID,
				point.ShapePtSequence,
				point.RowNumber,
				existingPoint.RowNumber,
			))
		} else {
			sequenceMap[point.ShapePtSequence] = point
		}
	}

	// Check for non-increasing sequences
	for i := 1; i < len(shape.Points); i++ {
		if shape.Points[i].ShapePtSequence <= shape.Points[i-1].ShapePtSequence {
			container.AddNotice(notice.NewNonIncreasingShapeSequenceNotice(
				shape.ShapeID,
				shape.Points[i].ShapePtSequence,
				shape.Points[i-1].ShapePtSequence,
				shape.Points[i].RowNumber,
			))
		}
	}
}

// validateShapeDistances validates shape distance values
func (v *ShapeValidator) validateShapeDistances(container *notice.NoticeContainer, shape *ShapeInfo) {
	hasAnyDistance := false
	for _, point := range shape.Points {
		if point.ShapeDistTraveled != nil {
			hasAnyDistance = true
			break
		}
	}

	if !hasAnyDistance {
		return // No distances to validate
	}

	// Check that all points have distances if any do
	for _, point := range shape.Points {
		if point.ShapeDistTraveled == nil {
			container.AddNotice(notice.NewInconsistentShapeDistanceNotice(
				shape.ShapeID,
				point.ShapePtSequence,
				point.RowNumber,
			))
		}
	}

	// Validate distance progression
	for i := 1; i < len(shape.Points); i++ {
		curr := shape.Points[i]
		prev := shape.Points[i-1]

		if curr.ShapeDistTraveled != nil && prev.ShapeDistTraveled != nil {
			if *curr.ShapeDistTraveled < *prev.ShapeDistTraveled {
				container.AddNotice(notice.NewDecreasingShapeDistanceNotice(
					shape.ShapeID,
					curr.ShapePtSequence,
					*curr.ShapeDistTraveled,
					*prev.ShapeDistTraveled,
					curr.RowNumber,
				))
			}

			if *curr.ShapeDistTraveled == *prev.ShapeDistTraveled {
				container.AddNotice(notice.NewEqualShapeDistanceNotice(
					shape.ShapeID,
					curr.ShapePtSequence,
					prev.ShapePtSequence,
					*curr.ShapeDistTraveled,
					curr.RowNumber,
				))
			}
		}
	}
}

// validateShapeGeometry validates shape geometric properties
func (v *ShapeValidator) validateShapeGeometry(container *notice.NoticeContainer, shape *ShapeInfo) {
	// Check for duplicate consecutive points
	for i := 1; i < len(shape.Points); i++ {
		curr := shape.Points[i]
		prev := shape.Points[i-1]

		if v.approximatelyEqual(curr.ShapePtLat, prev.ShapePtLat, 1e-7) &&
			v.approximatelyEqual(curr.ShapePtLon, prev.ShapePtLon, 1e-7) {
			container.AddNotice(notice.NewDuplicateShapePointNotice(
				shape.ShapeID,
				curr.ShapePtSequence,
				prev.ShapePtSequence,
				curr.RowNumber,
			))
		}
	}

	// Check for unreasonably long segments
	for i := 1; i < len(shape.Points); i++ {
		curr := shape.Points[i]
		prev := shape.Points[i-1]

		distance := v.haversineDistance(
			prev.ShapePtLat, prev.ShapePtLon,
			curr.ShapePtLat, curr.ShapePtLon,
		)

		// Flag segments longer than 100km as potentially problematic
		if distance > 100000 {
			container.AddNotice(notice.NewUnreasonablyLongShapeSegmentNotice(
				shape.ShapeID,
				prev.ShapePtSequence,
				curr.ShapePtSequence,
				distance,
				curr.RowNumber,
			))
		}
	}
}

// validateShapeUsage checks if shapes are actually used by trips
func (v *ShapeValidator) validateShapeUsage(loader *parser.FeedLoader, container *notice.NoticeContainer, shapes map[string]*ShapeInfo) {
	// Load used shape IDs from trips.txt
	usedShapes := v.loadUsedShapes(loader)

	// Check for unused shapes
	for shapeID := range shapes {
		if !usedShapes[shapeID] {
			container.AddNotice(notice.NewUnusedShapeNotice(shapeID))
		}
	}
}

// loadUsedShapes loads shape IDs used in trips.txt
func (v *ShapeValidator) loadUsedShapes(loader *parser.FeedLoader) map[string]bool {
	usedShapes := make(map[string]bool)

	reader, err := loader.GetFile("trips.txt")
	if err != nil {
		return usedShapes
	}
	defer reader.Close()

	csvFile, err := parser.NewCSVFile(reader, "trips.txt")
	if err != nil {
		return usedShapes
	}

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}

		if shapeID, hasShapeID := row.Values["shape_id"]; hasShapeID && strings.TrimSpace(shapeID) != "" {
			usedShapes[strings.TrimSpace(shapeID)] = true
		}
	}

	return usedShapes
}

// approximatelyEqual checks if two float64 values are approximately equal
func (v *ShapeValidator) approximatelyEqual(a, b, epsilon float64) bool {
	return math.Abs(a-b) < epsilon
}

// haversineDistance calculates the distance between two lat/lon points in meters
func (v *ShapeValidator) haversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadius = 6371000 // Earth radius in meters

	lat1Rad := lat1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	deltaLatRad := (lat2 - lat1) * math.Pi / 180
	deltaLonRad := (lon2 - lon1) * math.Pi / 180

	a := math.Sin(deltaLatRad/2)*math.Sin(deltaLatRad/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLonRad/2)*math.Sin(deltaLonRad/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c
}
