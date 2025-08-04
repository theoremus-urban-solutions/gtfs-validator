package entity

import (
	"io"
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
	defer reader.Close()

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

		// Check if coordinates are present
		_, hasLat := row.Values["stop_lat"]
		_, hasLon := row.Values["stop_lon"]
		stop.HasCoordinates = hasLat && hasLon

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

// validateCoordinatesRequirement validates coordinate requirements by location type
func (v *StopLocationValidator) validateCoordinatesRequirement(container *notice.NoticeContainer, stop *StopInfo) {
	// Coordinates are required for certain location types
	requiresCoordinates := false

	switch stop.LocationType {
	case 0, 2, 3, 4: // Stop/platform, entrance/exit, generic node, boarding area
		requiresCoordinates = true
	case 1: // Station - coordinates optional but recommended
		if !stop.HasCoordinates {
			container.AddNotice(notice.NewMissingRecommendedFieldNotice(
				"stops.txt",
				"stop_lat/stop_lon",
				stop.RowNumber,
			))
		}
	}

	if requiresCoordinates && !stop.HasCoordinates {
		container.AddNotice(notice.NewMissingCoordinatesNotice(
			stop.StopID,
			stop.LocationType,
			stop.RowNumber,
		))
	}
}

// validateParentStationReference validates parent_station references
func (v *StopLocationValidator) validateParentStationReference(container *notice.NoticeContainer, stop *StopInfo, allStops map[string]*StopInfo) {
	if stop.ParentStation == "" {
		return // No parent station reference
	}

	// Check if parent station exists
	parentStop, exists := allStops[stop.ParentStation]
	if !exists {
		container.AddNotice(notice.NewInvalidParentStationReferenceNotice(
			stop.StopID,
			stop.ParentStation,
			stop.RowNumber,
		))
		return
	}

	// Validate parent station type based on child location type
	switch stop.LocationType {
	case 0: // Stop/platform - can have station (1) as parent
		if parentStop.LocationType != 1 {
			container.AddNotice(notice.NewInvalidParentStationTypeNotice(
				stop.StopID,
				stop.ParentStation,
				parentStop.LocationType,
				stop.RowNumber,
			))
		}
	case 2: // Entrance/exit - must have station (1) as parent
		if parentStop.LocationType != 1 {
			container.AddNotice(notice.NewInvalidParentStationTypeNotice(
				stop.StopID,
				stop.ParentStation,
				parentStop.LocationType,
				stop.RowNumber,
			))
		}
	case 3: // Generic node - can have station (1) as parent
		if parentStop.LocationType != 1 {
			container.AddNotice(notice.NewInvalidParentStationTypeNotice(
				stop.StopID,
				stop.ParentStation,
				parentStop.LocationType,
				stop.RowNumber,
			))
		}
	case 4: // Boarding area - should have platform (0) as parent
		if parentStop.LocationType != 0 {
			container.AddNotice(notice.NewInvalidParentStationTypeNotice(
				stop.StopID,
				stop.ParentStation,
				parentStop.LocationType,
				stop.RowNumber,
			))
		}
	}
}

// validateLocationTypeRules validates location type specific rules
func (v *StopLocationValidator) validateLocationTypeRules(container *notice.NoticeContainer, stop *StopInfo, allStops map[string]*StopInfo) {
	switch stop.LocationType {
	case 0: // Stop/platform
		// Stops can optionally have a parent station
		break

	case 1: // Station
		// Stations cannot have parent stations
		if stop.ParentStation != "" {
			container.AddNotice(notice.NewStationWithParentStationNotice(
				stop.StopID,
				stop.ParentStation,
				stop.RowNumber,
			))
		}

	case 2: // Entrance/exit
		// Entrances must have a parent station
		if stop.ParentStation == "" {
			container.AddNotice(notice.NewMissingParentStationNotice(
				stop.StopID,
				stop.LocationType,
				stop.RowNumber,
			))
		}

	case 3: // Generic node
		// Generic nodes can optionally have parent stations
		break

	case 4: // Boarding area
		// Boarding areas should have parent stations
		if stop.ParentStation == "" {
			container.AddNotice(notice.NewMissingParentStationNotice(
				stop.StopID,
				stop.LocationType,
				stop.RowNumber,
			))
		}
	}
}

// validateStopHierarchy validates the overall stop hierarchy
func (v *StopLocationValidator) validateStopHierarchy(container *notice.NoticeContainer, stops map[string]*StopInfo) {
	// Check for circular references
	v.validateCircularReferences(container, stops)

	// Check for orphaned stations
	v.validateOrphanedStations(container, stops)
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

// validateOrphanedStations checks for stations with no child stops
func (v *StopLocationValidator) validateOrphanedStations(container *notice.NoticeContainer, stops map[string]*StopInfo) {
	// Count children for each station
	stationChildren := make(map[string]int)

	for _, stop := range stops {
		if stop.LocationType == 1 { // Station
			stationChildren[stop.StopID] = 0
		}
	}

	// Count actual children
	for _, stop := range stops {
		if stop.ParentStation != "" {
			if _, isStation := stationChildren[stop.ParentStation]; isStation {
				stationChildren[stop.ParentStation]++
			}
		}
	}

	// Report stations with no children
	for stationID, childCount := range stationChildren {
		if childCount == 0 {
			if station, exists := stops[stationID]; exists {
				container.AddNotice(notice.NewOrphanedStationNotice(
					stationID,
					station.RowNumber,
				))
			}
		}
	}
}
