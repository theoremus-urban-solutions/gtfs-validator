package entity

import (
	"io"
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
	TripID       string
	StopID       string
	StopSequence int
	ArrivalTime  string
	DepartureTime string
	RowNumber    int
}

// TripPattern represents a unique sequence of stops
type TripPattern struct {
	PatternID string
	StopSequence []string
	Trips     []string
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

	// Analyze trip patterns
	patterns := v.analyzeTripPatterns(tripStopTimes)
	v.validateTripPatterns(container, patterns)
}

// loadStopTimes loads stop times from stop_times.txt
func (v *TripPatternValidator) loadStopTimes(loader *parser.FeedLoader) []*TripStopTime {
	var stopTimes []*TripStopTime

	reader, err := loader.GetFile("stop_times.txt")
	if err != nil {
		return stopTimes
	}
	defer reader.Close()

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
	if len(stopTimes) < 2 {
		container.AddNotice(notice.NewInsufficientStopTimesNotice(
			tripID,
			len(stopTimes),
		))
		return
	}

	// Check for duplicate stop sequences
	seqMap := make(map[int]*TripStopTime)
	for _, stopTime := range stopTimes {
		if existing, exists := seqMap[stopTime.StopSequence]; exists {
			container.AddNotice(notice.NewDuplicateStopSequenceNotice(
				tripID,
				stopTime.StopSequence,
				stopTime.StopID,
				existing.RowNumber,
				stopTime.RowNumber,
			))
		} else {
			seqMap[stopTime.StopSequence] = stopTime
		}
	}

	// Check for non-increasing sequences
	for i := 1; i < len(stopTimes); i++ {
		if stopTimes[i].StopSequence <= stopTimes[i-1].StopSequence {
			container.AddNotice(notice.NewNonIncreasingStopSequenceNotice(
				tripID,
				stopTimes[i].StopSequence,
				stopTimes[i-1].StopSequence,
				stopTimes[i].RowNumber,
			))
		}
	}

	// Check for gaps in sequence (not critical but informational)
	expectedSeq := stopTimes[0].StopSequence
	for _, stopTime := range stopTimes {
		if stopTime.StopSequence != expectedSeq {
			container.AddNotice(notice.NewStopSequenceGapNotice(
				tripID,
				expectedSeq,
				stopTime.StopSequence,
				stopTime.RowNumber,
			))
		}
		expectedSeq = stopTime.StopSequence + 1
	}

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

	// Check for loop trips (first and last stop are the same)
	if len(stopTimes) >= 3 {
		firstStop := stopTimes[0].StopID
		lastStop := stopTimes[len(stopTimes)-1].StopID
		
		if firstStop == lastStop {
			// This is a loop trip - check if it's properly structured
			container.AddNotice(notice.NewLoopRouteNotice(
				tripID,
				firstStop,
				stopTimes[0].RowNumber,
				stopTimes[len(stopTimes)-1].RowNumber,
			))
		}
	}
}

// analyzeTripPatterns analyzes trip patterns to find similar routes
func (v *TripPatternValidator) analyzeTripPatterns(tripStopTimes map[string][]*TripStopTime) map[string]*TripPattern {
	patterns := make(map[string]*TripPattern)
	patternMap := make(map[string]string) // pattern hash -> pattern ID

	patternCounter := 1

	for tripID, stopTimes := range tripStopTimes {
		// Create pattern signature
		var stopSequence []string
		for _, stopTime := range stopTimes {
			stopSequence = append(stopSequence, stopTime.StopID)
		}

		patternHash := strings.Join(stopSequence, "|")
		
		var patternID string
		if existingPatternID, exists := patternMap[patternHash]; exists {
			patternID = existingPatternID
		} else {
			patternID = "pattern_" + strconv.Itoa(patternCounter)
			patternCounter++
			patternMap[patternHash] = patternID
			
			patterns[patternID] = &TripPattern{
				PatternID: patternID,
				StopSequence: stopSequence,
				Trips: []string{},
			}
		}

		patterns[patternID].Trips = append(patterns[patternID].Trips, tripID)
	}

	return patterns
}

// validateTripPatterns validates patterns for efficiency and consistency
func (v *TripPatternValidator) validateTripPatterns(container *notice.NoticeContainer, patterns map[string]*TripPattern) {
	// Check for single-trip patterns (might indicate inefficiency)
	for _, pattern := range patterns {
		if len(pattern.Trips) == 1 {
			container.AddNotice(notice.NewSingleTripPatternNotice(
				pattern.PatternID,
				pattern.Trips[0],
				len(pattern.StopSequence),
			))
		}

		// Check for very short patterns (less than 2 stops)
		if len(pattern.StopSequence) < 2 {
			container.AddNotice(notice.NewShortTripPatternNotice(
				pattern.PatternID,
				len(pattern.StopSequence),
				len(pattern.Trips),
			))
		}

		// Check for very long patterns (might indicate route splitting needs)
		if len(pattern.StopSequence) > 100 {
			container.AddNotice(notice.NewLongTripPatternNotice(
				pattern.PatternID,
				len(pattern.StopSequence),
				len(pattern.Trips),
			))
		}
	}

	// Analysis summary for informational purposes
	totalPatterns := len(patterns)
	totalTrips := 0
	for _, pattern := range patterns {
		totalTrips += len(pattern.Trips)
	}

	if totalTrips > 0 {
		avgTripsPerPattern := float64(totalTrips) / float64(totalPatterns)
		container.AddNotice(notice.NewTripPatternSummaryNotice(
			totalPatterns,
			totalTrips,
			avgTripsPerPattern,
		))
	}
}