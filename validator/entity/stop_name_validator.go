package entity

import (
	"io"
	"log"
	"strconv"
	"strings"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/parser"
	"github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

// StopNameValidator validates stop names are properly specified
type StopNameValidator struct{}

// NewStopNameValidator creates a new stop name validator
func NewStopNameValidator() *StopNameValidator {
	return &StopNameValidator{}
}

// StopNameInfo represents stop naming information
type StopNameInfo struct {
	StopID        string
	StopName      string
	StopDesc      string
	LocationType  int
	ParentStation string
	RowNumber     int
}

// Validate checks stop names for various issues
func (v *StopNameValidator) Validate(loader *parser.FeedLoader, container *notice.NoticeContainer, config validator.Config) {
	stops := v.loadStops(loader)

	// Build parent station map for context
	parentStations := make(map[string]*StopNameInfo)
	for _, stop := range stops {
		if stop.LocationType == 1 { // Station
			parentStations[stop.StopID] = stop
		}
	}

	for _, stop := range stops {
		v.validateStopName(container, stop, parentStations)
	}
}

// loadStops loads stop information from stops.txt
func (v *StopNameValidator) loadStops(loader *parser.FeedLoader) []*StopNameInfo {
	var stops []*StopNameInfo

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

		stop := v.parseStop(row)
		if stop != nil {
			stops = append(stops, stop)
		}
	}

	return stops
}

// parseStop parses stop information from a row
func (v *StopNameValidator) parseStop(row *parser.CSVRow) *StopNameInfo {
	stopID, hasStopID := row.Values["stop_id"]
	if !hasStopID {
		return nil
	}

	stop := &StopNameInfo{
		StopID:    strings.TrimSpace(stopID),
		RowNumber: row.RowNumber,
	}

	// Parse stop_name
	if stopName, hasStopName := row.Values["stop_name"]; hasStopName {
		stop.StopName = strings.TrimSpace(stopName)
	}

	// Parse stop_desc
	if stopDesc, hasStopDesc := row.Values["stop_desc"]; hasStopDesc {
		stop.StopDesc = strings.TrimSpace(stopDesc)
	}

	// Parse location_type (defaults to 0)
	if locTypeStr, hasLocType := row.Values["location_type"]; hasLocType && strings.TrimSpace(locTypeStr) != "" {
		if locType, err := strconv.Atoi(strings.TrimSpace(locTypeStr)); err == nil {
			stop.LocationType = locType
		}
	}

	// Parse parent_station
	if parentStation, hasParent := row.Values["parent_station"]; hasParent {
		stop.ParentStation = strings.TrimSpace(parentStation)
	}

	return stop
}

// validateStopName validates a single stop's naming
func (v *StopNameValidator) validateStopName(container *notice.NoticeContainer, stop *StopNameInfo, parentStations map[string]*StopNameInfo) {
	// Check if stop_name is required for this location type
	nameRequired := v.isStopNameRequired(stop.LocationType)

	// A required stop_name is required whether or not a parent station has
	// one; nothing in GTFS inherits it.
	if nameRequired && stop.StopName == "" {
		container.AddNotice(notice.NewMissingRequiredStopNameNotice(
			stop.StopID,
			stop.LocationType,
			stop.RowNumber,
		))
	}

	// Additional validations only if name exists
	if stop.StopName != "" {
		// Check for problematic characters
		v.checkProblematicCharacters(container, stop)

		// Check if name and description are identical
		v.checkNameDescriptionDuplicate(container, stop)

		// stop_name is announced and displayed to riders, so a single-case
		// name reaches them degraded.
		if needsMixedCase(stop.StopName) {
			container.AddNotice(notice.NewMixedCaseRecommendedFieldNotice(
				"stops.txt",
				"stop_name",
				stop.StopName,
				stop.RowNumber,
			))
		}
	}
}

// isStopNameRequired checks if stop_name is required for the location type
func (v *StopNameValidator) isStopNameRequired(locationType int) bool {
	// stop_name is required for:
	// 0 = Stop/Platform
	// 1 = Station
	// 2 = Entrance/Exit
	// stop_name is optional for:
	// 3 = Generic Node
	// 4 = Boarding Area
	return locationType <= 2
}

// checkProblematicCharacters checks for problematic characters in stop names
func (v *StopNameValidator) checkProblematicCharacters(container *notice.NoticeContainer, stop *StopNameInfo) {
	// Check for control characters
	for i, ch := range stop.StopName {
		if ch < 32 && ch != 9 && ch != 10 && ch != 13 { // Allow tab, newline, carriage return
			container.AddNotice(notice.NewStopNameContainsControlCharacterNotice(
				stop.StopID,
				stop.StopName,
				i,
				int(ch),
				stop.RowNumber,
			))
		}
	}

}

// checkNameDescriptionDuplicate checks if stop_name and stop_desc are identical.
//
// stop_desc is asked for information the name does not already give — where in
// the station the stop is, which entrance to use. A copy of the name adds a
// line to the rider's screen and nothing to what they know. Casing does not
// make it informative, so "MAIN ST" describing "Main St" is still a duplicate.
func (v *StopNameValidator) checkNameDescriptionDuplicate(container *notice.NoticeContainer, stop *StopNameInfo) {
	if stop.StopDesc == "" || !strings.EqualFold(stop.StopDesc, stop.StopName) {
		return
	}

	container.AddNotice(notice.NewSameNameAndDescriptionForStopNotice(
		stop.StopID,
		stop.StopDesc,
		stop.RowNumber,
	))
}
