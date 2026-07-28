package entity

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

// TripPatternValidator validates trip patterns and stop sequences
type TripPatternValidator struct{}

// NewTripPatternValidator creates a new trip pattern validator
func NewTripPatternValidator() *TripPatternValidator {
	return &TripPatternValidator{}
}

// TripStopTime represents a stop time for pattern analysis
type TripStopTime struct {
	TripID        string
	StopID        string
	StopSequence  int
	ArrivalTime   string
	DepartureTime string
	RowNumber     int
}

// TripPattern represents a unique sequence of stops
type TripPattern struct {
	PatternID    string
	StopSequence []string
	Trips        []string
}

// Validate checks trip patterns and stop sequences
func (v *TripPatternValidator) Validate(loader *parser.FeedLoader, container *notice.NoticeContainer, config validator.Config) {
	// Load stop times
	stopTimes := v.loadStopTimes(loader)
	if len(stopTimes) == 0 {
		return
	}

	// Group stop times by trip
	tripStopTimes := v.groupStopTimesByTrip(stopTimes)

	// Validate each trip's stop sequence
	for tripID, tripStops := range tripStopTimes {
		v.validateTripStopSequence(container, tripID, tripStops)
	}

}

// loadStopTimes loads stop times from stop_times.txt
func (v *TripPatternValidator) loadStopTimes(loader *parser.FeedLoader) []*TripStopTime {
	var stopTimes []*TripStopTime

	reader, err := loader.GetFile("stop_times.txt")
	if err != nil {
		return stopTimes
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close reader %v", closeErr)
		}
	}()

	csvFile, err := parser.NewCSVFile(reader, "stop_times.txt")
	if err != nil {
		return stopTimes
	}

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		stopTime := v.parseStopTime(row)
		if stopTime != nil {
			stopTimes = append(stopTimes, stopTime)
		}
	}

	return stopTimes
}

// parseStopTime parses a stop time record
func (v *TripPatternValidator) parseStopTime(row *parser.CSVRow) *TripStopTime {
	tripID, hasTripID := row.Values["trip_id"]
	stopID, hasStopID := row.Values["stop_id"]
	stopSeqStr, hasStopSeq := row.Values["stop_sequence"]

	if !hasTripID || !hasStopID || !hasStopSeq {
		return nil
	}

	stopSeq, err := strconv.Atoi(strings.TrimSpace(stopSeqStr))
	if err != nil {
		return nil
	}

	stopTime := &TripStopTime{
		TripID:       strings.TrimSpace(tripID),
		StopID:       strings.TrimSpace(stopID),
		StopSequence: stopSeq,
		RowNumber:    row.RowNumber,
	}

	if arrivalTime, hasArrival := row.Values["arrival_time"]; hasArrival {
		stopTime.ArrivalTime = strings.TrimSpace(arrivalTime)
	}

	if departureTime, hasDeparture := row.Values["departure_time"]; hasDeparture {
		stopTime.DepartureTime = strings.TrimSpace(departureTime)
	}

	return stopTime
}

// groupStopTimesByTrip groups stop times by trip ID
func (v *TripPatternValidator) groupStopTimesByTrip(stopTimes []*TripStopTime) map[string][]*TripStopTime {
	tripMap := make(map[string][]*TripStopTime)

	for _, stopTime := range stopTimes {
		tripMap[stopTime.TripID] = append(tripMap[stopTime.TripID], stopTime)
	}

	// Sort stop times by stop_sequence for each trip
	for tripID, stops := range tripMap {
		sort.Slice(stops, func(i, j int) bool {
			return stops[i].StopSequence < stops[j].StopSequence
		})
		tripMap[tripID] = stops
	}

	return tripMap
}

// validateTripStopSequence validates the stop sequence for a single trip
func (v *TripPatternValidator) validateTripStopSequence(container *notice.NoticeContainer, tripID string, stopTimes []*TripStopTime) {
	// A trip with fewer than two stops is reported by
	// business/trip_usability_validator.go as unusable_trip.
	if len(stopTimes) < 2 {
		return
	}

	// A stop_sequence repeated within a trip is reported by
	// core/duplicate_key_validator.go, which keys stop_times.txt on
	// trip_id + stop_sequence. A trip whose rows do not ascend is reported by
	// relationship/stop_time_field_validator.go as unsorted_stop_times, which
	// reads them in file order rather than sorted as they are here.

	// Gaps in stop_sequence are legal: the spec requires the values to
	// increase along the trip, not to be contiguous.

	// Check for consecutive duplicate stops
	for i := 1; i < len(stopTimes); i++ {
		if stopTimes[i].StopID == stopTimes[i-1].StopID {
			container.AddNotice(notice.NewConsecutiveDuplicateStopsNotice(
				tripID,
				stopTimes[i].StopID,
				stopTimes[i-1].StopSequence,
				stopTimes[i].StopSequence,
				stopTimes[i].RowNumber,
			))
		}
	}

}
