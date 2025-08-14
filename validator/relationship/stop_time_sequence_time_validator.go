package relationship

import (
	"fmt"
	"io"
	"log"
	"sort"
	"strconv"
	"strings"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/parser"
	"github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

// StopTimeSequenceTimeValidator validates that arrival/departure times are logical
type StopTimeSequenceTimeValidator struct{}

// NewStopTimeSequenceTimeValidator creates a new stop time sequence time validator
func NewStopTimeSequenceTimeValidator() *StopTimeSequenceTimeValidator {
	return &StopTimeSequenceTimeValidator{}
}

// StopTimeRecord represents a stop time record for time validation
type StopTimeRecord struct {
	TripID        string
	StopSequence  int
	ArrivalTime   *int // seconds since midnight
	DepartureTime *int // seconds since midnight
	RowNumber     int
}

// Validate checks stop time sequences for logical time ordering
func (v *StopTimeSequenceTimeValidator) Validate(loader *parser.FeedLoader, container *notice.NoticeContainer, config validator.Config) {
	reader, err := loader.GetFile("stop_times.txt")
	if err != nil {
		return // File doesn't exist, skip validation
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close reader %v", closeErr)
		}
	}()

	csvFile, err := parser.NewCSVFile(reader, "stop_times.txt")
	if err != nil {
		return
	}

	// Group stop times by trip_id
	tripStopTimes := make(map[string][]StopTimeRecord)

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}

		stopTime := v.parseStopTimeRecord(row)
		if stopTime != nil {
			tripStopTimes[stopTime.TripID] = append(tripStopTimes[stopTime.TripID], *stopTime)
		}
	}

	// Validate each trip's stop times
	for tripID, stopTimes := range tripStopTimes {
		v.validateTripStopTimeTimes(container, tripID, stopTimes)
	}
}

// parseStopTimeRecord parses a stop time row into a StopTimeRecord struct
func (v *StopTimeSequenceTimeValidator) parseStopTimeRecord(row *parser.CSVRow) *StopTimeRecord {
	tripID, hasTripID := row.Values["trip_id"]
	stopSeqStr, hasStopSeq := row.Values["stop_sequence"]
	arrivalTimeStr, hasArrivalTime := row.Values["arrival_time"]
	departureTimeStr, hasDepartureTime := row.Values["departure_time"]

	if !hasTripID || !hasStopSeq {
		return nil
	}

	stopSequence, err := strconv.Atoi(strings.TrimSpace(stopSeqStr))
	if err != nil {
		return nil
	}

	stopTime := &StopTimeRecord{
		TripID:       strings.TrimSpace(tripID),
		StopSequence: stopSequence,
		RowNumber:    row.RowNumber,
	}

	// Parse arrival time if present
	if hasArrivalTime && strings.TrimSpace(arrivalTimeStr) != "" {
		if arrivalSeconds, err := v.parseGTFSTime(strings.TrimSpace(arrivalTimeStr)); err == nil {
			stopTime.ArrivalTime = &arrivalSeconds
		}
	}

	// Parse departure time if present
	if hasDepartureTime && strings.TrimSpace(departureTimeStr) != "" {
		if departureSeconds, err := v.parseGTFSTime(strings.TrimSpace(departureTimeStr)); err == nil {
			stopTime.DepartureTime = &departureSeconds
		}
	}

	return stopTime
}

// parseGTFSTime parses a GTFS time string (HH:MM:SS) into seconds since midnight
func (v *StopTimeSequenceTimeValidator) parseGTFSTime(timeStr string) (int, error) {
	parts := strings.Split(timeStr, ":")
	if len(parts) != 3 {
		return 0, fmt.Errorf("invalid time format: %s", timeStr)
	}

	hours, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, err
	}

	minutes, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, err
	}

	seconds, err := strconv.Atoi(parts[2])
	if err != nil {
		return 0, err
	}

	// GTFS allows hours > 23 for next-day service
	if minutes < 0 || minutes >= 60 || seconds < 0 || seconds >= 60 {
		return 0, fmt.Errorf("invalid time values: %s", timeStr)
	}

	return hours*3600 + minutes*60 + seconds, nil
}

// validateTripStopTimeTimes validates stop times for a single trip
func (v *StopTimeSequenceTimeValidator) validateTripStopTimeTimes(container *notice.NoticeContainer, tripID string, stopTimes []StopTimeRecord) {
	if len(stopTimes) < 2 {
		return // Need at least 2 stop times to validate sequence
	}

	// Sort by stop_sequence
	sort.Slice(stopTimes, func(i, j int) bool {
		return stopTimes[i].StopSequence < stopTimes[j].StopSequence
	})

	// Validate individual stop time consistency (arrival <= departure)
	for _, stopTime := range stopTimes {
		v.validateStopTimeConsistency(container, stopTime)
	}

	// Validate time sequence across stops
	v.validateTimeSequence(container, stopTimes)
}

// validateStopTimeConsistency validates that arrival time <= departure time for a single stop
func (v *StopTimeSequenceTimeValidator) validateStopTimeConsistency(container *notice.NoticeContainer, stopTime StopTimeRecord) {
	if stopTime.ArrivalTime != nil && stopTime.DepartureTime != nil {
		if *stopTime.ArrivalTime > *stopTime.DepartureTime {
			container.AddNotice(notice.NewStopTimeArrivalAfterDepartureNotice(
				stopTime.TripID,
				stopTime.StopSequence,
				v.formatGTFSTime(*stopTime.ArrivalTime),
				v.formatGTFSTime(*stopTime.DepartureTime),
				stopTime.RowNumber,
			))
		}
	}
}

// validateTimeSequence validates that times increase along the trip
func (v *StopTimeSequenceTimeValidator) validateTimeSequence(container *notice.NoticeContainer, stopTimes []StopTimeRecord) {
	var prevDepartureTime *int
	var prevStopTime *StopTimeRecord

	for i := range stopTimes {
		currentStopTime := &stopTimes[i]

		// Get the effective departure time for the current stop
		var currentDepartureTime *int
		if currentStopTime.DepartureTime != nil {
			currentDepartureTime = currentStopTime.DepartureTime
		} else if currentStopTime.ArrivalTime != nil {
			currentDepartureTime = currentStopTime.ArrivalTime
		}

		// Get the effective arrival time for the current stop
		var currentArrivalTime *int
		if currentStopTime.ArrivalTime != nil {
			currentArrivalTime = currentStopTime.ArrivalTime
		} else if currentStopTime.DepartureTime != nil {
			currentArrivalTime = currentStopTime.DepartureTime
		}

		// Check if current arrival time is before previous departure time
		if prevDepartureTime != nil && currentArrivalTime != nil && *currentArrivalTime < *prevDepartureTime {
			container.AddNotice(notice.NewStopTimeDecreasingTimeNotice(
				currentStopTime.TripID,
				currentStopTime.StopSequence,
				v.formatGTFSTime(*currentArrivalTime),
				currentStopTime.RowNumber,
				prevStopTime.StopSequence,
				v.formatGTFSTime(*prevDepartureTime),
				prevStopTime.RowNumber,
			))
		}

		// Update previous values for next iteration
		if currentDepartureTime != nil {
			prevDepartureTime = currentDepartureTime
			prevStopTime = currentStopTime
		}
	}
}

// formatGTFSTime formats seconds since midnight back to HH:MM:SS format
func (v *StopTimeSequenceTimeValidator) formatGTFSTime(seconds int) string {
	hours := seconds / 3600
	minutes := (seconds % 3600) / 60
	secs := seconds % 60
	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, secs)
}
