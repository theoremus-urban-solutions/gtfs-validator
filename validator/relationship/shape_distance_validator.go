package relationship

import (
	"io"
	"log"
	"sort"
	"strconv"
	"strings"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/parser"
	"github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

// ShapeDistanceValidator validates that shape_dist_traveled increases along shape points
type ShapeDistanceValidator struct{}

// NewShapeDistanceValidator creates a new shape distance validator
func NewShapeDistanceValidator() *ShapeDistanceValidator {
	return &ShapeDistanceValidator{}
}

// ShapePoint represents a shape point record for validation
type ShapePoint struct {
	ShapeID           string
	ShapePtSequence   int
	ShapeDistTraveled *float64
	RowNumber         int
}

// Validate checks that shape_dist_traveled increases along shape points
func (v *ShapeDistanceValidator) Validate(loader *parser.FeedLoader, container *notice.NoticeContainer, config validator.Config) {
	reader, err := loader.GetFile("shapes.txt")
	if err != nil {
		return // File doesn't exist, skip validation
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close reader %v", closeErr)
		}
	}()

	csvFile, err := parser.NewCSVFile(reader, "shapes.txt")
	if err != nil {
		return
	}

	// Group shape points by shape_id
	shapePoints := make(map[string][]ShapePoint)

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}

		shapePoint := v.parseShapePoint(row)
		if shapePoint != nil {
			shapePoints[shapePoint.ShapeID] = append(shapePoints[shapePoint.ShapeID], *shapePoint)
		}
	}

	// Validate each shape's distance ordering
	for shapeID, points := range shapePoints {
		v.validateShapeDistanceOrder(container, shapeID, points)
	}
}

// parseShapePoint parses a shape point row into a ShapePoint struct
func (v *ShapeDistanceValidator) parseShapePoint(row *parser.CSVRow) *ShapePoint {
	shapeID, hasShapeID := row.Values["shape_id"]
	shapePtSeqStr, hasShapePtSeq := row.Values["shape_pt_sequence"]
	shapeDistStr, hasShapeDist := row.Values["shape_dist_traveled"]

	if !hasShapeID || !hasShapePtSeq {
		return nil
	}

	shapePtSequence, err := strconv.Atoi(strings.TrimSpace(shapePtSeqStr))
	if err != nil {
		return nil
	}

	shapePoint := &ShapePoint{
		ShapeID:         strings.TrimSpace(shapeID),
		ShapePtSequence: shapePtSequence,
		RowNumber:       row.RowNumber,
	}

	// Parse shape_dist_traveled if present
	if hasShapeDist && strings.TrimSpace(shapeDistStr) != "" {
		if shapeDist, err := strconv.ParseFloat(strings.TrimSpace(shapeDistStr), 64); err == nil {
			shapePoint.ShapeDistTraveled = &shapeDist
		}
	}

	return shapePoint
}

// validateShapeDistanceOrder validates that distances increase along a shape
func (v *ShapeDistanceValidator) validateShapeDistanceOrder(container *notice.NoticeContainer, shapeID string, points []ShapePoint) {
	if len(points) < 2 {
		return // Need at least 2 points to validate sequence
	}

	// Sort by shape_pt_sequence
	sort.Slice(points, func(i, j int) bool {
		return points[i].ShapePtSequence < points[j].ShapePtSequence
	})

	// Check for duplicate sequences
	v.validateDuplicateShapeSequences(container, points)

	// Check for decreasing distances
	v.validateIncreasingDistances(container, points)
}

// validateDuplicateShapeSequences checks for duplicate shape_pt_sequence values
func (v *ShapeDistanceValidator) validateDuplicateShapeSequences(container *notice.NoticeContainer, points []ShapePoint) {
	sequenceMap := make(map[int][]ShapePoint)

	for _, point := range points {
		sequenceMap[point.ShapePtSequence] = append(sequenceMap[point.ShapePtSequence], point)
	}

	for sequence, duplicatePoints := range sequenceMap {
		if len(duplicatePoints) > 1 {
			for i := 1; i < len(duplicatePoints); i++ {
				container.AddNotice(notice.NewDuplicateShapeSequenceNotice(
					duplicatePoints[i].ShapeID,
					sequence,
					duplicatePoints[i].RowNumber,
					duplicatePoints[0].RowNumber,
				))
			}
		}
	}
}

// validateIncreasingDistances checks that shape_dist_traveled values increase
func (v *ShapeDistanceValidator) validateIncreasingDistances(container *notice.NoticeContainer, points []ShapePoint) {
	var prevPoint *ShapePoint

	for i := range points {
		currentPoint := &points[i]

		if prevPoint != nil &&
			prevPoint.ShapeDistTraveled != nil &&
			currentPoint.ShapeDistTraveled != nil &&
			*prevPoint.ShapeDistTraveled >= *currentPoint.ShapeDistTraveled {
			container.AddNotice(notice.NewDecreasingOrEqualShapeDistanceNotice(
				currentPoint.ShapeID,
				currentPoint.ShapePtSequence,
				currentPoint.RowNumber,
				*currentPoint.ShapeDistTraveled,
				prevPoint.ShapePtSequence,
				prevPoint.RowNumber,
				*prevPoint.ShapeDistTraveled,
			))
		}

		prevPoint = currentPoint
	}
}
