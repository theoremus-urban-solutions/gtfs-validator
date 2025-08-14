package accessibility

import (
	"io"
	"log"
	"strconv"
	"strings"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/parser"
	"github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

// PathwayValidator validates pathway definitions for accessibility
type PathwayValidator struct{}

// NewPathwayValidator creates a new pathway validator
func NewPathwayValidator() *PathwayValidator {
	return &PathwayValidator{}
}

// validPathwayModes contains valid GTFS pathway modes
var validPathwayModes = map[int]bool{
	1: true, // Walkway
	2: true, // Stairs
	3: true, // Moving sidewalk/travelator
	4: true, // Escalator
	5: true, // Elevator
	6: true, // Fare gate
	7: true, // Exit gate
}

// PathwayInfo represents pathway information
type PathwayInfo struct {
	PathwayID            string
	FromStopID           string
	ToStopID             string
	PathwayMode          int
	IsBidirectional      int
	Length               *float64
	TraversalTime        *int
	StairCount           *int
	MaxSlope             *float64
	MinWidth             *float64
	SignpostedAs         string
	ReversedSignpostedAs string
	RowNumber            int
}

// Validate checks pathway definitions
func (v *PathwayValidator) Validate(loader *parser.FeedLoader, container *notice.NoticeContainer, config validator.Config) {
	pathways := v.loadPathways(loader)
	if len(pathways) == 0 {
		return // No pathways to validate
	}

	// Load stop information for validation
	stops := v.loadStopsForPathways(loader)

	// Validate each pathway
	for _, pathway := range pathways {
		v.validatePathway(container, pathway, stops)
	}

	// Check for duplicate pathways
	v.validateDuplicatePathways(container, pathways)

	// Validate bidirectional consistency
	v.validateBidirectionalConsistency(container, pathways)
}

// loadPathways loads pathway information from pathways.txt
func (v *PathwayValidator) loadPathways(loader *parser.FeedLoader) []*PathwayInfo {
	var pathways []*PathwayInfo

	reader, err := loader.GetFile("pathways.txt")
	if err != nil {
		return pathways
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close %v", closeErr)
		}
	}()

	csvFile, err := parser.NewCSVFile(reader, "pathways.txt")
	if err != nil {
		return pathways
	}

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}

		pathway := v.parsePathway(row)
		if pathway != nil {
			pathways = append(pathways, pathway)
		}
	}

	return pathways
}

// parsePathway parses a pathway record
func (v *PathwayValidator) parsePathway(row *parser.CSVRow) *PathwayInfo {
	pathwayID, hasPathwayID := row.Values["pathway_id"]
	fromStopID, hasFromStopID := row.Values["from_stop_id"]
	toStopID, hasToStopID := row.Values["to_stop_id"]
	pathwayModeStr, hasPathwayMode := row.Values["pathway_mode"]
	isBidirectionalStr, hasIsBidirectional := row.Values["is_bidirectional"]

	if !hasPathwayID || !hasFromStopID || !hasToStopID || !hasPathwayMode || !hasIsBidirectional {
		return nil
	}

	pathwayMode, err := strconv.Atoi(strings.TrimSpace(pathwayModeStr))
	if err != nil {
		return nil
	}

	isBidirectional, err := strconv.Atoi(strings.TrimSpace(isBidirectionalStr))
	if err != nil {
		return nil
	}

	pathway := &PathwayInfo{
		PathwayID:       strings.TrimSpace(pathwayID),
		FromStopID:      strings.TrimSpace(fromStopID),
		ToStopID:        strings.TrimSpace(toStopID),
		PathwayMode:     pathwayMode,
		IsBidirectional: isBidirectional,
		RowNumber:       row.RowNumber,
	}

	// Parse optional fields
	if lengthStr, hasLength := row.Values["length"]; hasLength && strings.TrimSpace(lengthStr) != "" {
		if length, err := strconv.ParseFloat(strings.TrimSpace(lengthStr), 64); err == nil {
			pathway.Length = &length
		}
	}

	if traversalTimeStr, hasTraversalTime := row.Values["traversal_time"]; hasTraversalTime && strings.TrimSpace(traversalTimeStr) != "" {
		if traversalTime, err := strconv.Atoi(strings.TrimSpace(traversalTimeStr)); err == nil {
			pathway.TraversalTime = &traversalTime
		}
	}

	if stairCountStr, hasStairCount := row.Values["stair_count"]; hasStairCount && strings.TrimSpace(stairCountStr) != "" {
		if stairCount, err := strconv.Atoi(strings.TrimSpace(stairCountStr)); err == nil {
			pathway.StairCount = &stairCount
		}
	}

	if maxSlopeStr, hasMaxSlope := row.Values["max_slope"]; hasMaxSlope && strings.TrimSpace(maxSlopeStr) != "" {
		if maxSlope, err := strconv.ParseFloat(strings.TrimSpace(maxSlopeStr), 64); err == nil {
			pathway.MaxSlope = &maxSlope
		}
	}

	if minWidthStr, hasMinWidth := row.Values["min_width"]; hasMinWidth && strings.TrimSpace(minWidthStr) != "" {
		if minWidth, err := strconv.ParseFloat(strings.TrimSpace(minWidthStr), 64); err == nil {
			pathway.MinWidth = &minWidth
		}
	}

	if signpostedAs, hasSignpostedAs := row.Values["signposted_as"]; hasSignpostedAs {
		pathway.SignpostedAs = strings.TrimSpace(signpostedAs)
	}

	if reversedSignpostedAs, hasReversedSignpostedAs := row.Values["reversed_signposted_as"]; hasReversedSignpostedAs {
		pathway.ReversedSignpostedAs = strings.TrimSpace(reversedSignpostedAs)
	}

	return pathway
}

// loadStopsForPathways loads stop information needed for pathway validation
func (v *PathwayValidator) loadStopsForPathways(loader *parser.FeedLoader) map[string]bool {
	stops := make(map[string]bool)

	reader, err := loader.GetFile("stops.txt")
	if err != nil {
		return stops
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close %v", closeErr)
		}
	}()

	csvFile, err := parser.NewCSVFile(reader, "stops.txt")
	if err != nil {
		return stops
	}

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}

		if stopID, hasStopID := row.Values["stop_id"]; hasStopID {
			stops[strings.TrimSpace(stopID)] = true
		}
	}

	return stops
}

// validatePathway validates a single pathway record
func (v *PathwayValidator) validatePathway(container *notice.NoticeContainer, pathway *PathwayInfo, stops map[string]bool) {
	// Validate pathway mode
	if !validPathwayModes[pathway.PathwayMode] {
		container.AddNotice(notice.NewInvalidPathwayModeNotice(
			pathway.PathwayID,
			pathway.PathwayMode,
			pathway.RowNumber,
		))
	}

	// Validate is_bidirectional
	if pathway.IsBidirectional != 0 && pathway.IsBidirectional != 1 {
		container.AddNotice(notice.NewInvalidBidirectionalNotice(
			pathway.PathwayID,
			pathway.IsBidirectional,
			pathway.RowNumber,
		))
	}

	// Validate stop references
	if !stops[pathway.FromStopID] {
		container.AddNotice(notice.NewForeignKeyViolationNotice(
			"pathways.txt",
			"from_stop_id",
			pathway.FromStopID,
			pathway.RowNumber,
			"stops.txt",
			"stop_id",
		))
	}

	if !stops[pathway.ToStopID] {
		container.AddNotice(notice.NewForeignKeyViolationNotice(
			"pathways.txt",
			"to_stop_id",
			pathway.ToStopID,
			pathway.RowNumber,
			"stops.txt",
			"stop_id",
		))
	}

	// Validate pathway from/to same stop
	if pathway.FromStopID == pathway.ToStopID {
		container.AddNotice(notice.NewPathwayToSameStopNotice(
			pathway.PathwayID,
			pathway.FromStopID,
			pathway.RowNumber,
		))
	}

	// Validate pathway-specific requirements
	v.validatePathwaySpecificRequirements(container, pathway)

	// Validate numeric fields
	v.validatePathwayNumericFields(container, pathway)
}

// validatePathwaySpecificRequirements validates requirements specific to pathway modes
func (v *PathwayValidator) validatePathwaySpecificRequirements(container *notice.NoticeContainer, pathway *PathwayInfo) {
	switch pathway.PathwayMode {
	case 2: // Stairs
		// Stairs should have stair_count
		if pathway.StairCount == nil {
			container.AddNotice(notice.NewMissingRecommendedFieldNotice(
				"pathways.txt",
				"stair_count",
				pathway.RowNumber,
			))
		} else if *pathway.StairCount <= 0 {
			container.AddNotice(notice.NewInvalidStairCountNotice(
				pathway.PathwayID,
				*pathway.StairCount,
				pathway.RowNumber,
			))
		}

	case 4: // Escalator
		// Escalators should have stair_count
		if pathway.StairCount == nil {
			container.AddNotice(notice.NewMissingRecommendedFieldNotice(
				"pathways.txt",
				"stair_count",
				pathway.RowNumber,
			))
		}

	case 6, 7: // Fare gate, Exit gate
		// Gates should typically not be bidirectional
		if pathway.IsBidirectional == 1 {
			container.AddNotice(notice.NewUnexpectedBidirectionalGateNotice(
				pathway.PathwayID,
				pathway.PathwayMode,
				pathway.RowNumber,
			))
		}
	}
}

// validatePathwayNumericFields validates numeric field constraints
func (v *PathwayValidator) validatePathwayNumericFields(container *notice.NoticeContainer, pathway *PathwayInfo) {
	// Validate length
	if pathway.Length != nil && *pathway.Length <= 0 {
		container.AddNotice(notice.NewInvalidPathwayLengthNotice(
			pathway.PathwayID,
			*pathway.Length,
			pathway.RowNumber,
		))
	}

	// Validate traversal time
	if pathway.TraversalTime != nil && *pathway.TraversalTime <= 0 {
		container.AddNotice(notice.NewInvalidTraversalTimeNotice(
			pathway.PathwayID,
			*pathway.TraversalTime,
			pathway.RowNumber,
		))
	}

	// Validate max slope (should be reasonable)
	if pathway.MaxSlope != nil {
		if *pathway.MaxSlope < -1.0 || *pathway.MaxSlope > 1.0 {
			container.AddNotice(notice.NewUnreasonableMaxSlopeNotice(
				pathway.PathwayID,
				*pathway.MaxSlope,
				pathway.RowNumber,
			))
		}
	}

	// Validate min width
	if pathway.MinWidth != nil && *pathway.MinWidth <= 0 {
		container.AddNotice(notice.NewInvalidMinWidthNotice(
			pathway.PathwayID,
			*pathway.MinWidth,
			pathway.RowNumber,
		))
	}
}

// validateDuplicatePathways checks for duplicate pathway definitions
func (v *PathwayValidator) validateDuplicatePathways(container *notice.NoticeContainer, pathways []*PathwayInfo) {
	pathwayMap := make(map[string]*PathwayInfo)

	for _, pathway := range pathways {
		key := pathway.FromStopID + "->" + pathway.ToStopID

		if existingPathway, exists := pathwayMap[key]; exists {
			container.AddNotice(notice.NewDuplicatePathwayNotice(
				pathway.PathwayID,
				pathway.FromStopID,
				pathway.ToStopID,
				pathway.RowNumber,
				existingPathway.RowNumber,
			))
		} else {
			pathwayMap[key] = pathway
		}
	}
}

// validateBidirectionalConsistency checks bidirectional pathway consistency
func (v *PathwayValidator) validateBidirectionalConsistency(container *notice.NoticeContainer, pathways []*PathwayInfo) {
	// Create map of pathway directions
	pathwayMap := make(map[string]*PathwayInfo)

	for _, pathway := range pathways {
		forward := pathway.FromStopID + "->" + pathway.ToStopID
		reverse := pathway.ToStopID + "->" + pathway.FromStopID

		pathwayMap[forward] = pathway

		// Check if there's a reverse pathway when this one is not bidirectional
		if pathway.IsBidirectional == 0 {
			if reversePathway, hasReverse := pathwayMap[reverse]; hasReverse {
				// Both directions exist as separate pathways - this is valid
				// But check if they have consistent properties
				if reversePathway.PathwayMode != pathway.PathwayMode {
					container.AddNotice(notice.NewInconsistentBidirectionalPathwayNotice(
						pathway.PathwayID,
						reversePathway.PathwayID,
						pathway.RowNumber,
						reversePathway.RowNumber,
					))
				}
			}
		}
	}
}
