package entity

import (
	"io"
	"log"
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

	v.validateShapesAreUsed(container, shapes, loader)
}

// validateShapesAreUsed reports shapes no trip draws. A feed that cannot be
// read for trips at all is left alone: with no references to compare against,
// every shape would look unused, and the missing-file check owns that failure.
func (v *ShapeValidator) validateShapesAreUsed(container *notice.NoticeContainer, shapes map[string]*ShapeInfo, loader *parser.FeedLoader) {
	referenced, ok := v.loadReferencedShapeIDs(loader)
	if !ok {
		return
	}

	// Sorted so the report does not reshuffle between runs.
	shapeIDs := make([]string, 0, len(shapes))
	for shapeID := range shapes {
		shapeIDs = append(shapeIDs, shapeID)
	}
	sort.Strings(shapeIDs)

	for _, shapeID := range shapeIDs {
		if referenced[shapeID] {
			continue
		}
		container.AddNotice(notice.NewUnusedShapeNotice(
			shapeID,
			shapes[shapeID].Points[0].RowNumber,
		))
	}
}

// loadReferencedShapeIDs collects every shape_id trips.txt names, reporting
// whether the file could be read at all.
func (v *ShapeValidator) loadReferencedShapeIDs(loader *parser.FeedLoader) (map[string]bool, bool) {
	referenced := make(map[string]bool)

	reader, err := loader.GetFile("trips.txt")
	if err != nil {
		return nil, false
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close reader %v", closeErr)
		}
	}()

	csvFile, err := parser.NewCSVFile(reader, "trips.txt")
	if err != nil {
		return nil, false
	}

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}

		if shapeID := strings.TrimSpace(row.Values["shape_id"]); shapeID != "" {
			referenced[shapeID] = true
		}
	}

	return referenced, true
}

// loadShapes loads shape information from shapes.txt
func (v *ShapeValidator) loadShapes(loader *parser.FeedLoader) map[string]*ShapeInfo {
	shapes := make(map[string]*ShapeInfo)

	reader, err := loader.GetFile("shapes.txt")
	if err != nil {
		return shapes
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close reader %v", closeErr)
		}
	}()

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

	v.validateShapeDistances(container, shape)
}

// equalDistanceThresholdMetres is the distance below which two shape points
// sharing a shape_dist_traveled are treated as a rounding artefact rather than
// a real gap. 1.11 m is 1e-5 degrees of latitude — the smallest difference a
// five-decimal coordinate can express.
const equalDistanceThresholdMetres = 1.11

// validateShapeDistances validates the shape_dist_traveled progression along a
// shape. Sorted by shape_pt_sequence, the values must increase: a decrease is
// an error, and equal values mean the shape covers ground the distance does
// not account for.
func (v *ShapeValidator) validateShapeDistances(container *notice.NoticeContainer, shape *ShapeInfo) {
	for i := 1; i < len(shape.Points); i++ {
		curr := shape.Points[i]
		prev := shape.Points[i-1]

		if curr.ShapeDistTraveled == nil || prev.ShapeDistTraveled == nil {
			continue
		}

		switch {
		case *curr.ShapeDistTraveled < *prev.ShapeDistTraveled:
			container.AddNotice(notice.NewDecreasingShapeDistanceNotice(
				shape.ShapeID,
				curr.ShapePtSequence,
				*curr.ShapeDistTraveled,
				*prev.ShapeDistTraveled,
				curr.RowNumber,
			))

		case *curr.ShapeDistTraveled == *prev.ShapeDistTraveled:
			distance := v.haversineDistance(
				prev.ShapePtLat, prev.ShapePtLon,
				curr.ShapePtLat, curr.ShapePtLon,
			)

			switch {
			case distance == 0:
				container.AddNotice(notice.NewEqualShapeDistanceSameCoordinatesNotice(
					shape.ShapeID, *curr.ShapeDistTraveled,
					prev.ShapePtSequence, curr.ShapePtSequence,
					prev.RowNumber, curr.RowNumber,
				))
			case distance < equalDistanceThresholdMetres:
				container.AddNotice(notice.NewEqualShapeDistanceDiffCoordinatesBelowThresholdNotice(
					shape.ShapeID, *curr.ShapeDistTraveled,
					prev.ShapePtSequence, curr.ShapePtSequence,
					prev.RowNumber, curr.RowNumber, distance,
				))
			default:
				container.AddNotice(notice.NewEqualShapeDistanceDiffCoordinatesNotice(
					shape.ShapeID, *curr.ShapeDistTraveled,
					prev.ShapePtSequence, curr.ShapePtSequence,
					prev.RowNumber, curr.RowNumber, distance,
				))
			}
		}
	}
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
