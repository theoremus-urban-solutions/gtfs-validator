package relationship

import (
	"io"
	"log"
	"strconv"
	"strings"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/parser"
	"github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

// TripHeadsignValidator reports trips whose headsign names a stop the vehicle
// only passes through. The headsign is what a waiting passenger reads to
// decide whether to board, so it should name where the vehicle ends up.
type TripHeadsignValidator struct{}

// NewTripHeadsignValidator creates a new trip headsign validator
func NewTripHeadsignValidator() *TripHeadsignValidator {
	return &TripHeadsignValidator{}
}

// headsignCandidate accumulates what one trip's stop times say about its
// headsign: the furthest position reached, and the first position at which a
// stop's name matched. Only these two are kept, so the pass over stop_times
// costs a fixed amount of memory per trip rather than per stop time.
type headsignCandidate struct {
	Headsign  string
	RowNumber int

	LastSequence int
	HasLast      bool

	MatchSequence int
	MatchStopID   string
	HasMatch      bool
}

// Validate reports headsigns matching an intermediate stop of their trip.
func (v *TripHeadsignValidator) Validate(loader *parser.FeedLoader, container *notice.NoticeContainer, config validator.Config) {
	candidates := v.loadHeadsignTrips(loader)
	if len(candidates) == 0 {
		return
	}

	stopNames := v.loadStopNames(loader)
	if len(stopNames) == 0 {
		return
	}

	v.scanStopTimes(loader, candidates, stopNames)

	for tripID, candidate := range candidates {
		// A headsign naming the final stop is the normal case; only a match
		// before the end contradicts what the sign promises.
		if !candidate.HasMatch || !candidate.HasLast || candidate.MatchSequence >= candidate.LastSequence {
			continue
		}
		container.AddNotice(notice.NewTripHeadsignMatchesIntermediateStopNotice(
			tripID,
			candidate.RowNumber,
			candidate.Headsign,
			candidate.MatchStopID,
			candidate.MatchSequence,
		))
	}
}

// loadHeadsignTrips returns the trips that declare a headsign, keyed by trip
// id. Trips without one cannot trip this rule and are not carried further.
func (v *TripHeadsignValidator) loadHeadsignTrips(loader *parser.FeedLoader) map[string]*headsignCandidate {
	candidates := make(map[string]*headsignCandidate)

	reader, err := loader.GetFile("trips.txt")
	if err != nil {
		return candidates
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close reader %v", closeErr)
		}
	}()

	csvFile, err := parser.NewCSVFile(reader, "trips.txt")
	if err != nil {
		return candidates
	}

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		tripID := strings.TrimSpace(row.Values["trip_id"])
		headsign := strings.TrimSpace(row.Values["trip_headsign"])
		if tripID == "" || headsign == "" {
			continue
		}
		candidates[tripID] = &headsignCandidate{
			Headsign:  headsign,
			RowNumber: row.RowNumber,
		}
	}

	return candidates
}

// loadStopNames maps each stop id to its name.
func (v *TripHeadsignValidator) loadStopNames(loader *parser.FeedLoader) map[string]string {
	names := make(map[string]string)

	reader, err := loader.GetFile("stops.txt")
	if err != nil {
		return names
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close reader %v", closeErr)
		}
	}()

	csvFile, err := parser.NewCSVFile(reader, "stops.txt")
	if err != nil {
		return names
	}

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		stopID := strings.TrimSpace(row.Values["stop_id"])
		stopName := strings.TrimSpace(row.Values["stop_name"])
		if stopID == "" || stopName == "" {
			continue
		}
		names[stopID] = stopName
	}

	return names
}

// scanStopTimes makes the single pass over stop_times.txt, recording each
// candidate trip's furthest stop and its first name match.
func (v *TripHeadsignValidator) scanStopTimes(loader *parser.FeedLoader, candidates map[string]*headsignCandidate, stopNames map[string]string) {
	reader, err := loader.GetFile("stop_times.txt")
	if err != nil {
		return
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

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			return
		}
		if err != nil {
			continue
		}

		candidate, wanted := candidates[strings.TrimSpace(row.Values["trip_id"])]
		if !wanted {
			continue
		}
		sequence, err := strconv.Atoi(strings.TrimSpace(row.Values["stop_sequence"]))
		if err != nil {
			continue
		}

		if !candidate.HasLast || sequence > candidate.LastSequence {
			candidate.LastSequence = sequence
			candidate.HasLast = true
		}

		if candidate.HasMatch && sequence >= candidate.MatchSequence {
			continue
		}
		stopID := strings.TrimSpace(row.Values["stop_id"])
		if name, exists := stopNames[stopID]; exists && strings.EqualFold(name, candidate.Headsign) {
			candidate.MatchSequence = sequence
			candidate.MatchStopID = stopID
			candidate.HasMatch = true
		}
	}
}
