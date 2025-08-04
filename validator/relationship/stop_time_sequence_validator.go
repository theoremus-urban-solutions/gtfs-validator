package relationship

import (
	"io"
	"sort"
	"strconv"
	"strings"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/parser"
	"github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

// StopTimeSequenceValidator validates stop time sequences and distance ordering
type StopTimeSequenceValidator struct{}

// NewStopTimeSequenceValidator creates a new stop time sequence validator
func NewStopTimeSequenceValidator() *StopTimeSequenceValidator {
	return &StopTimeSequenceValidator{}
}

// StopTime represents a stop time record for validation
type StopTime struct {
	TripID            string
	StopID            string
	StopSequence      int
	ShapeDistTraveled *float64
	RowNumber         int
}

// Validate checks stop time sequences and shape distance ordering
func (v *StopTimeSequenceValidator) Validate(loader *parser.FeedLoader, container *notice.NoticeContainer, config validator.Config) {
	reader, err := loader.GetFile("stop_times.txt")
	if err != nil {
		return // File doesn't exist, skip validation
	}
	defer reader.Close()

	csvFile, err := parser.NewCSVFile(reader, "stop_times.txt")
	if err != nil {
		return
	}

	// Group stop times by trip_id
	tripStopTimes := make(map[string][]StopTime)

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}

		stopTime := v.parseStopTime(row)
		if stopTime != nil {
			tripStopTimes[stopTime.TripID] = append(tripStopTimes[stopTime.TripID], *stopTime)
		}
	}

	// Validate each trip's stop times
	for tripID, stopTimes := range tripStopTimes {
		v.validateTripStopTimes(container, tripID, stopTimes)
	}
}

// parseStopTime parses a stop time row into a StopTime struct
func (v *StopTimeSequenceValidator) parseStopTime(row *parser.CSVRow) *StopTime {
	tripID, hasTripID := row.Values["trip_id"]
	stopID, hasStopID := row.Values["stop_id"]
	stopSeqStr, hasStopSeq := row.Values["stop_sequence"]

	if !hasTripID || !hasStopID || !hasStopSeq {
		return nil
	}

	stopSequence, err := strconv.Atoi(strings.TrimSpace(stopSeqStr))
	if err != nil {
		return nil
	}

	stopTime := &StopTime{
		TripID:       strings.TrimSpace(tripID),
		StopID:       strings.TrimSpace(stopID),
		StopSequence: stopSequence,
		RowNumber:    row.RowNumber,
	}

	// Parse shape_dist_traveled if present
	if shapeDistStr, hasShapeDist := row.Values["shape_dist_traveled"]; hasShapeDist && strings.TrimSpace(shapeDistStr) != "" {
		if shapeDist, err := strconv.ParseFloat(strings.TrimSpace(shapeDistStr), 64); err == nil {
			stopTime.ShapeDistTraveled = &shapeDist
		}
	}

	return stopTime
}

// validateTripStopTimes validates stop times for a single trip
func (v *StopTimeSequenceValidator) validateTripStopTimes(container *notice.NoticeContainer, tripID string, stopTimes []StopTime) {
	if len(stopTimes) < 2 {
		return // Need at least 2 stop times to validate sequence
	}

	// Sort by stop_sequence
	sort.Slice(stopTimes, func(i, j int) bool {
		return stopTimes[i].StopSequence < stopTimes[j].StopSequence
	})

	// Check for duplicate stop sequences
	v.validateDuplicateStopSequences(container, stopTimes)

	// Check for decreasing shape distances
	v.validateShapeDistanceOrder(container, stopTimes)
}

// validateDuplicateStopSequences checks for duplicate stop_sequence values
func (v *StopTimeSequenceValidator) validateDuplicateStopSequences(container *notice.NoticeContainer, stopTimes []StopTime) {
	sequenceMap := make(map[int][]StopTime)

	for _, stopTime := range stopTimes {
		sequenceMap[stopTime.StopSequence] = append(sequenceMap[stopTime.StopSequence], stopTime)
	}

	for sequence, stops := range sequenceMap {
		if len(stops) > 1 {
			for i := 1; i < len(stops); i++ {
				container.AddNotice(notice.NewDuplicateStopSequenceNotice(
					stops[i].TripID,
					sequence,
					stops[i].StopID,
					stops[i].RowNumber,
					stops[0].RowNumber,
				))
			}
		}
	}
}

// validateShapeDistanceOrder checks that shape_dist_traveled values are increasing
func (v *StopTimeSequenceValidator) validateShapeDistanceOrder(container *notice.NoticeContainer, stopTimes []StopTime) {
	var prevStopTime *StopTime

	for i := range stopTimes {
		currentStopTime := &stopTimes[i]

		// Skip if stop doesn't have stop_id (location groups, etc.)
		if strings.TrimSpace(currentStopTime.StopID) == "" {
			continue
		}

		if prevStopTime != nil &&
			prevStopTime.ShapeDistTraveled != nil &&
			currentStopTime.ShapeDistTraveled != nil &&
			*prevStopTime.ShapeDistTraveled >= *currentStopTime.ShapeDistTraveled {

			container.AddNotice(notice.NewDecreasingOrEqualStopTimeDistanceNotice(
				currentStopTime.TripID,
				currentStopTime.StopID,
				currentStopTime.RowNumber,
				*currentStopTime.ShapeDistTraveled,
				currentStopTime.StopSequence,
				prevStopTime.RowNumber,
				*prevStopTime.ShapeDistTraveled,
				prevStopTime.StopSequence,
			))
		}

		prevStopTime = currentStopTime
	}
}
