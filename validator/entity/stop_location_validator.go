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

// validLocationTypes contains valid GTFS location types
var validLocationTypes = map[int]bool{
	0: true, // Stop/platform
	1: true, // Station
	2: true, // Entrance/exit
	3: true, // Generic node
	4: true, // Boarding area
}

// StopInfo represents stop information for validation
type StopInfo struct {
	StopID         string
	StopName       string
	LocationType   int
	ParentStation  string
	RowNumber      int
	HasCoordinates bool
}

// Validate checks stop location consistency and hierarchy
func (v *StopLocationValidator) Validate(loader *parser.FeedLoader, container *notice.NoticeContainer, config validator.Config) {
	stops := v.loadStops(loader)

	// Validate each stop
	for _, stop := range stops {
		v.validateStop(container, stop, stops)
	}

	// Validate parent-child relationships
	v.validateStopHierarchy(container, stops)
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
func (v *StopLocationValidator) validateStop(container *notice.NoticeContainer, stop *StopInfo, allStops map[string]*StopInfo) {
	// Validate location type
	v.validateLocationType(container, stop)

	// Validate coordinates requirement
	v.validateCoordinatesRequirement(container, stop)

	// Validate parent station reference
	v.validateParentStationReference(container, stop, allStops)

	// Validate location type specific rules
	v.validateLocationTypeRules(container, stop, allStops)
}

// validateLocationType validates the location_type field
func (v *StopLocationValidator) validateLocationType(container *notice.NoticeContainer, stop *StopInfo) {
	if !validLocationTypes[stop.LocationType] {
		container.AddNotice(notice.NewInvalidLocationTypeNotice(
			stop.StopID,
			stop.LocationType,
			stop.RowNumber,
		))
	}
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
func (v *StopLocationValidator) validateLocationTypeRules(container *notice.NoticeContainer, stop *StopInfo, allStops map[string]*StopInfo) {
	switch stop.LocationType {
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
