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

// StopLocationValidator validates stop location hierarchies and types
type StopLocationValidator struct{}

// NewStopLocationValidator creates a new stop location validator
func NewStopLocationValidator() *StopLocationValidator {
	return &StopLocationValidator{}
}

// StopInfo represents stop information for validation
type StopInfo struct {
	StopID         string
	StopName       string
	LocationType   int
	ParentStation  string
	ZoneID         string
	StopAccess     string
	RowNumber      int
	HasCoordinates bool
}

// Validate checks stop location consistency and hierarchy
func (v *StopLocationValidator) Validate(loader *parser.FeedLoader, container *notice.NoticeContainer, config validator.Config) {
	stops := v.loadStops(loader)
	hasStation := feedHasStation(stops)

	// Validate each stop
	for _, stop := range stops {
		v.validateStop(container, stop, stops, hasStation)
	}

	// Validate parent-child relationships
	v.validateStopHierarchy(container, stops)

	// Validate that zone-priced routes reach only stops that declare a zone
	v.validateZoneCoverage(loader, container, stops)
}

// loadStops loads stop information from stops.txt
func (v *StopLocationValidator) loadStops(loader *parser.FeedLoader) map[string]*StopInfo {
	stops := make(map[string]*StopInfo)

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
			break
		}

		stopID, hasStopID := row.Values["stop_id"]
		if !hasStopID {
			continue
		}

		stopIDTrimmed := strings.TrimSpace(stopID)

		stop := &StopInfo{
			StopID:    stopIDTrimmed,
			RowNumber: row.RowNumber,
		}

		// Load stop name
		if stopName, hasStopName := row.Values["stop_name"]; hasStopName {
			stop.StopName = strings.TrimSpace(stopName)
		}

		// Load location type (default is 0)
		stop.LocationType = 0
		if locationTypeStr, hasLocationType := row.Values["location_type"]; hasLocationType && strings.TrimSpace(locationTypeStr) != "" {
			if locationType, err := strconv.Atoi(strings.TrimSpace(locationTypeStr)); err == nil {
				stop.LocationType = locationType
			}
		}

		// Load parent station
		if parentStation, hasParentStation := row.Values["parent_station"]; hasParentStation {
			stop.ParentStation = strings.TrimSpace(parentStation)
		}

		stop.ZoneID = strings.TrimSpace(row.Values["zone_id"])
		stop.StopAccess = strings.TrimSpace(row.Values["stop_access"])

		// Check if coordinates are present.
		// A column may exist in the row map but hold an empty value (e.g. a stop
		// row with empty stop_lat/stop_lon), so test for a non-empty value rather
		// than mere column presence.
		latVal, hasLatCol := row.Values["stop_lat"]
		lonVal, hasLonCol := row.Values["stop_lon"]
		stop.HasCoordinates = hasLatCol && hasLonCol &&
			strings.TrimSpace(latVal) != "" && strings.TrimSpace(lonVal) != ""

		stops[stopIDTrimmed] = stop
	}

	return stops
}

// validateStop validates a single stop
func (v *StopLocationValidator) validateStop(container *notice.NoticeContainer, stop *StopInfo, allStops map[string]*StopInfo, hasStation bool) {
	// Validate coordinates requirement
	v.validateCoordinatesRequirement(container, stop)

	// Validate parent station reference
	v.validateParentStationReference(container, stop, allStops)

	// Validate location type specific rules
	v.validateLocationTypeRules(container, stop, hasStation)

	// Validate stop_access, which only a platform inside a station may declare
	v.validateStopAccess(container, stop)
}

// feedHasStation reports whether the feed models station hierarchy at all.
func feedHasStation(stops map[string]*StopInfo) bool {
	for _, stop := range stops {
		if stop.LocationType == 1 {
			return true
		}
	}
	return false
}

// expectedParentLocationType gives the location_type a parent must have for a
// child of the given type, and whether that child may have a parent at all.
// Stations sit at the top of the hierarchy and cannot have one.
func expectedParentLocationType(locationType int) (int, bool) {
	switch locationType {
	case 0, 2, 3: // Stop/platform, entrance/exit, generic node -> station
		return 1, true
	case 4: // Boarding area -> stop/platform
		return 0, true
	default:
		return 0, false
	}
}

// validateCoordinatesRequirement reports locations that must be placed on a
// map but have no coordinates. Generic nodes and boarding areas are exempt:
// they inherit their position from the station around them.
func (v *StopLocationValidator) validateCoordinatesRequirement(container *notice.NoticeContainer, stop *StopInfo) {
	if stop.HasCoordinates {
		return
	}
	switch stop.LocationType {
	case 0, 1, 2:
		container.AddNotice(notice.NewStopWithoutLocationNotice(
			stop.StopID,
			stop.RowNumber,
			stop.LocationType,
		))
	}
}

// validateParentStationReference validates parent_station references
func (v *StopLocationValidator) validateParentStationReference(container *notice.NoticeContainer, stop *StopInfo, allStops map[string]*StopInfo) {
	if stop.ParentStation == "" {
		return // No parent station reference
	}

	parentStop, exists := allStops[stop.ParentStation]
	if !exists {
		container.AddNotice(notice.NewForeignKeyViolationNotice(
			"stops.txt",
			"parent_station",
			stop.ParentStation,
			stop.RowNumber,
			"stops.txt",
			"stop_id",
		))
		return
	}

	expected, mayHaveParent := expectedParentLocationType(stop.LocationType)
	if !mayHaveParent {
		return // A station with a parent is reported by validateLocationTypeRules.
	}
	if parentStop.LocationType != expected {
		container.AddNotice(notice.NewWrongParentLocationTypeNotice(
			stop.StopID,
			stop.RowNumber,
			stop.LocationType,
			stop.ParentStation,
			parentStop.RowNumber,
			parentStop.LocationType,
			expected,
		))
	}
}

// validateLocationTypeRules validates location type specific rules
func (v *StopLocationValidator) validateLocationTypeRules(container *notice.NoticeContainer, stop *StopInfo, hasStation bool) {
	switch stop.LocationType {
	case 0: // Stop/platform
		// A platform outside a station is legal — a lone bus stop is exactly
		// that — so this is an advisory rather than the error raised for the
		// location types that only exist within a station.
		//
		// The advisory is worth making only where the feed models station
		// hierarchy somewhere. A feed that declares no station at all gives a
		// platform nothing to be a child of, so the notice would fire on every
		// stop in the feed and point at no fixable omission.
		if hasStation && stop.ParentStation == "" {
			container.AddNotice(notice.NewPlatformWithoutParentStationNotice(
				stop.StopID,
				stop.RowNumber,
			))
		}

	case 1: // Station
		if stop.ParentStation != "" {
			container.AddNotice(notice.NewStationWithParentStationNotice(
				stop.StopID,
				stop.ParentStation,
				stop.RowNumber,
			))
		}

	case 2, 3, 4: // Entrance/exit, generic node, boarding area
		// These types only have meaning inside a station, so the reference is
		// required rather than optional.
		if stop.ParentStation == "" {
			container.AddNotice(notice.NewLocationWithoutParentStationNotice(
				stop.StopID,
				stop.RowNumber,
				stop.LocationType,
			))
		}
	}
}

// validateStopAccess reports stop_access on a location that cannot carry it.
// The field says how a passenger reaches a platform relative to the station
// around it, so it needs both a platform and a station to be about.
func (v *StopLocationValidator) validateStopAccess(container *notice.NoticeContainer, stop *StopInfo) {
	if stop.StopAccess == "" {
		return
	}
	if stop.LocationType != 0 {
		container.AddNotice(notice.NewStopAccessSpecifiedForIncorrectLocationNotice(
			stop.StopID,
			stop.RowNumber,
			stop.LocationType,
			stop.StopAccess,
		))
		return
	}
	if stop.ParentStation == "" {
		container.AddNotice(notice.NewStopAccessSpecifiedForStopWithNoParentStationNotice(
			stop.StopID,
			stop.RowNumber,
			stop.StopAccess,
		))
	}
}

// validateZoneCoverage reports platforms served by a zone-priced route but
// carrying no zone of their own, which leaves the fare rule inapplicable to
// any journey through them.
//
// The join it needs — fare rules to routes to trips to stop times — is only
// walked when fare_rules.txt actually prices by zone, which most feeds do not.
func (v *StopLocationValidator) validateZoneCoverage(loader *parser.FeedLoader, container *notice.NoticeContainer, stops map[string]*StopInfo) {
	routeIDs, allRoutes := v.loadZoneFareRoutes(loader)
	if !allRoutes && len(routeIDs) == 0 {
		return
	}

	tripIDs := v.loadTripsForRoutes(loader, routeIDs, allRoutes)
	if len(tripIDs) == 0 {
		return
	}

	for _, stop := range v.loadStopsWithoutZone(loader, tripIDs, stops) {
		container.AddNotice(notice.NewStopWithoutZoneIDNotice(
			stop.StopID,
			stop.StopName,
			stop.RowNumber,
		))
	}
}

// loadZoneFareRoutes returns the routes named by a fare rule that prices by
// zone. A qualifying rule with no route_id prices every route, which the
// second return value reports.
func (v *StopLocationValidator) loadZoneFareRoutes(loader *parser.FeedLoader) (map[string]bool, bool) {
	routeIDs := make(map[string]bool)

	reader, err := loader.GetFile("fare_rules.txt")
	if err != nil {
		return routeIDs, false
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close reader %v", closeErr)
		}
	}()

	csvFile, err := parser.NewCSVFile(reader, "fare_rules.txt")
	if err != nil {
		return routeIDs, false
	}

	allRoutes := false
	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}

		zonePriced := false
		for _, field := range []string{"origin_id", "destination_id", "contains_id"} {
			if strings.TrimSpace(row.Values[field]) != "" {
				zonePriced = true
				break
			}
		}
		if !zonePriced {
			continue
		}

		routeID := strings.TrimSpace(row.Values["route_id"])
		if routeID == "" {
			allRoutes = true
			continue
		}
		routeIDs[routeID] = true
	}

	return routeIDs, allRoutes
}

// loadTripsForRoutes returns the trips running on the given routes.
func (v *StopLocationValidator) loadTripsForRoutes(loader *parser.FeedLoader, routeIDs map[string]bool, allRoutes bool) map[string]bool {
	tripIDs := make(map[string]bool)

	reader, err := loader.GetFile("trips.txt")
	if err != nil {
		return tripIDs
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close reader %v", closeErr)
		}
	}()

	csvFile, err := parser.NewCSVFile(reader, "trips.txt")
	if err != nil {
		return tripIDs
	}

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}

		tripID := strings.TrimSpace(row.Values["trip_id"])
		if tripID == "" {
			continue
		}
		if allRoutes || routeIDs[strings.TrimSpace(row.Values["route_id"])] {
			tripIDs[tripID] = true
		}
	}

	return tripIDs
}

// loadStopsWithoutZone streams stop_times.txt and returns the platforms those
// trips call at that declare no zone, in the order they are first served.
func (v *StopLocationValidator) loadStopsWithoutZone(loader *parser.FeedLoader, tripIDs map[string]bool, stops map[string]*StopInfo) []*StopInfo {
	var offenders []*StopInfo

	reader, err := loader.GetFile("stop_times.txt")
	if err != nil {
		return offenders
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close reader %v", closeErr)
		}
	}()

	csvFile, err := parser.NewCSVFile(reader, "stop_times.txt")
	if err != nil {
		return offenders
	}

	reported := make(map[string]bool)
	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}

		if !tripIDs[strings.TrimSpace(row.Values["trip_id"])] {
			continue
		}
		stopID := strings.TrimSpace(row.Values["stop_id"])
		if reported[stopID] {
			continue
		}
		reported[stopID] = true

		stop, exists := stops[stopID]
		if !exists || stop.LocationType != 0 || stop.ZoneID != "" {
			continue
		}
		offenders = append(offenders, stop)
	}

	return offenders
}

// validateStopHierarchy validates the overall stop hierarchy
func (v *StopLocationValidator) validateStopHierarchy(container *notice.NoticeContainer, stops map[string]*StopInfo) {
	// Stations with no children are reported as unused_station by
	// relationship/usage_validator.go.
	v.validateCircularReferences(container, stops)
}

// validateCircularReferences checks for circular parent-child references
func (v *StopLocationValidator) validateCircularReferences(container *notice.NoticeContainer, stops map[string]*StopInfo) {
	for stopID, stop := range stops {
		if stop.ParentStation == "" {
			continue
		}

		// Follow the parent chain to detect cycles
		visited := make(map[string]bool)
		current := stopID

		for current != "" {
			if visited[current] {
				// Circular reference detected
				container.AddNotice(notice.NewCircularStationReferenceNotice(
					stopID,
					stop.RowNumber,
				))
				break
			}

			visited[current] = true

			if currentStop, exists := stops[current]; exists {
				current = currentStop.ParentStation
			} else {
				break
			}
		}
	}
}
