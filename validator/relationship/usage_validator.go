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

// UsageValidator reports entities nothing in the feed refers to: stops no trip
// calls at, trips with no stop times, and stations with no children. Each is
// usually a typo in the referring file rather than a deliberate omission.
type UsageValidator struct{}

// NewUsageValidator creates a new usage validator
func NewUsageValidator() *UsageValidator {
	return &UsageValidator{}
}

// stopRecord is the part of a stops.txt row this validator needs.
type stopRecord struct {
	StopID        string
	StopName      string
	LocationType  int
	ParentStation string
	RowNumber     int
}

// Validate reports unreferenced stops, trips and stations.
func (v *UsageValidator) Validate(loader *parser.FeedLoader, container *notice.NoticeContainer, config validator.Config) {
	stops := v.loadStops(loader)
	trips := v.loadTrips(loader)
	if len(stops) == 0 && len(trips) == 0 {
		return
	}

	visitedStops, tripsWithStopTimes := v.scanStopTimes(loader)

	for _, stop := range stops {
		// Only stops and platforms are called at; a station is reached
		// through its children, and entrances and nodes never appear in
		// stop_times.txt at all. A non-stop location that *is* referenced is
		// the other half of the same canonical validator, and the more
		// serious half: the reference does not resolve to anything boardable.
		if stop.LocationType != 0 {
			if row, referenced := visitedStops[stop.StopID]; referenced {
				container.AddNotice(notice.NewLocationWithUnexpectedStopTimeNotice(
					stop.StopID, stop.StopName, stop.RowNumber, row,
				))
			}
			continue
		}
		if _, referenced := visitedStops[stop.StopID]; !referenced {
			container.AddNotice(notice.NewStopWithoutStopTimeNotice(
				stop.StopID, stop.StopName, stop.RowNumber,
			))
		}
	}

	for tripID, rowNumber := range trips {
		if !tripsWithStopTimes[tripID] {
			container.AddNotice(notice.NewUnusedTripNotice(tripID, rowNumber))
		}
	}

	childStations := make(map[string]bool, len(stops))
	for _, stop := range stops {
		if stop.ParentStation != "" {
			childStations[stop.ParentStation] = true
		}
	}
	for _, stop := range stops {
		if stop.LocationType == 1 && !childStations[stop.StopID] {
			container.AddNotice(notice.NewUnusedStationNotice(
				stop.StopID, stop.StopName, stop.RowNumber,
			))
		}
	}
}

// scanStopTimes maps each stop any trip calls at to the first stop_times.txt
// row that references it, and returns the set of trips that have at least one
// stop time. The row number is kept because the unexpected-location notice
// reports the referencing row alongside the stop's own.
func (v *UsageValidator) scanStopTimes(loader *parser.FeedLoader) (map[string]int, map[string]bool) {
	visitedStops := make(map[string]int)
	tripsWithStopTimes := make(map[string]bool)

	reader, err := loader.GetFile("stop_times.txt")
	if err != nil {
		return visitedStops, tripsWithStopTimes
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close reader %v", closeErr)
		}
	}()

	csvFile, err := parser.NewCSVFile(reader, "stop_times.txt")
	if err != nil {
		return visitedStops, tripsWithStopTimes
	}

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}
		if stopID := strings.TrimSpace(row.Values["stop_id"]); stopID != "" {
			if _, seen := visitedStops[stopID]; !seen {
				visitedStops[stopID] = row.RowNumber
			}
		}
		if tripID := strings.TrimSpace(row.Values["trip_id"]); tripID != "" {
			tripsWithStopTimes[tripID] = true
		}
	}

	return visitedStops, tripsWithStopTimes
}

// loadStops reads stops.txt.
func (v *UsageValidator) loadStops(loader *parser.FeedLoader) []*stopRecord {
	var stops []*stopRecord

	reader, err := loader.GetFile("stops.txt")
	if err != nil {
		return stops
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close reader %v", closeErr)
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
			continue
		}

		stopID := strings.TrimSpace(row.Values["stop_id"])
		if stopID == "" {
			continue
		}

		stop := &stopRecord{
			StopID:        stopID,
			StopName:      strings.TrimSpace(row.Values["stop_name"]),
			ParentStation: strings.TrimSpace(row.Values["parent_station"]),
			RowNumber:     row.RowNumber,
		}
		if locationType, err := strconv.Atoi(strings.TrimSpace(row.Values["location_type"])); err == nil {
			stop.LocationType = locationType
		}
		stops = append(stops, stop)
	}

	return stops
}

// loadTrips maps each trip to the row it was declared on.
func (v *UsageValidator) loadTrips(loader *parser.FeedLoader) map[string]int {
	trips := make(map[string]int)

	reader, err := loader.GetFile("trips.txt")
	if err != nil {
		return trips
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close reader %v", closeErr)
		}
	}()

	csvFile, err := parser.NewCSVFile(reader, "trips.txt")
	if err != nil {
		return trips
	}

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}
		if tripID := strings.TrimSpace(row.Values["trip_id"]); tripID != "" {
			trips[tripID] = row.RowNumber
		}
	}

	return trips
}
