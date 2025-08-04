package entity

import (
	"io"
	"strconv"
	"strings"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/parser"
	"github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

// BikesAllowanceValidator validates bike allowance information for ferry trips
type BikesAllowanceValidator struct{}

// NewBikesAllowanceValidator creates a new bikes allowance validator
func NewBikesAllowanceValidator() *BikesAllowanceValidator {
	return &BikesAllowanceValidator{}
}

// TripBikeInfo represents trip bike allowance information
type TripBikeInfo struct {
	TripID           string
	RouteID          string
	RouteType        int
	BikesAllowed     *int // pointer to distinguish between unset and 0
	WheelchairAccessible *int
	RowNumber        int
}

// RouteBikeInfo represents route information for bike validation
type RouteBikeInfo struct {
	RouteID   string
	RouteType int
	RowNumber int
}

// Validate checks bike allowance information for ferry trips
func (v *BikesAllowanceValidator) Validate(loader *parser.FeedLoader, container *notice.NoticeContainer, config validator.Config) {
	// Load route information first
	routes := v.loadRoutes(loader)
	
	// Load trip information
	trips := v.loadTrips(loader, routes)
	
	// Validate bike allowance for ferry trips
	for _, trip := range trips {
		v.validateTripBikeAllowance(container, trip)
	}
}

// loadRoutes loads route information from routes.txt
func (v *BikesAllowanceValidator) loadRoutes(loader *parser.FeedLoader) map[string]*RouteBikeInfo {
	routes := make(map[string]*RouteBikeInfo)

	reader, err := loader.GetFile("routes.txt")
	if err != nil {
		return routes
	}
	defer reader.Close()

	csvFile, err := parser.NewCSVFile(reader, "routes.txt")
	if err != nil {
		return routes
	}

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		route := v.parseRoute(row)
		if route != nil {
			routes[route.RouteID] = route
		}
	}

	return routes
}

// parseRoute parses route information
func (v *BikesAllowanceValidator) parseRoute(row *parser.CSVRow) *RouteBikeInfo {
	routeID, hasRouteID := row.Values["route_id"]
	if !hasRouteID {
		return nil
	}

	route := &RouteBikeInfo{
		RouteID:   strings.TrimSpace(routeID),
		RowNumber: row.RowNumber,
	}

	// Parse route_type (required)
	if routeTypeStr, hasRouteType := row.Values["route_type"]; hasRouteType {
		if routeType, err := strconv.Atoi(strings.TrimSpace(routeTypeStr)); err == nil {
			route.RouteType = routeType
		} else {
			return nil // Invalid route_type
		}
	} else {
		return nil // Missing route_type
	}

	return route
}

// loadTrips loads trip information from trips.txt
func (v *BikesAllowanceValidator) loadTrips(loader *parser.FeedLoader, routes map[string]*RouteBikeInfo) []*TripBikeInfo {
	var trips []*TripBikeInfo

	reader, err := loader.GetFile("trips.txt")
	if err != nil {
		return trips
	}
	defer reader.Close()

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

		trip := v.parseTrip(row, routes)
		if trip != nil {
			trips = append(trips, trip)
		}
	}

	return trips
}

// parseTrip parses trip information
func (v *BikesAllowanceValidator) parseTrip(row *parser.CSVRow, routes map[string]*RouteBikeInfo) *TripBikeInfo {
	tripID, hasTripID := row.Values["trip_id"]
	routeID, hasRouteID := row.Values["route_id"]
	
	if !hasTripID || !hasRouteID {
		return nil
	}

	trip := &TripBikeInfo{
		TripID:    strings.TrimSpace(tripID),
		RouteID:   strings.TrimSpace(routeID),
		RowNumber: row.RowNumber,
	}

	// Get route type from routes
	if route, exists := routes[trip.RouteID]; exists {
		trip.RouteType = route.RouteType
	} else {
		return nil // Route not found
	}

	// Parse bikes_allowed field
	if bikesAllowedStr, hasBikesAllowed := row.Values["bikes_allowed"]; hasBikesAllowed && strings.TrimSpace(bikesAllowedStr) != "" {
		if bikesAllowed, err := strconv.Atoi(strings.TrimSpace(bikesAllowedStr)); err == nil {
			trip.BikesAllowed = &bikesAllowed
		}
	}

	// Parse wheelchair_accessible field (for context)
	if wheelchairStr, hasWheelchair := row.Values["wheelchair_accessible"]; hasWheelchair && strings.TrimSpace(wheelchairStr) != "" {
		if wheelchair, err := strconv.Atoi(strings.TrimSpace(wheelchairStr)); err == nil {
			trip.WheelchairAccessible = &wheelchair
		}
	}

	return trip
}

// validateTripBikeAllowance validates bike allowance for a trip
func (v *BikesAllowanceValidator) validateTripBikeAllowance(container *notice.NoticeContainer, trip *TripBikeInfo) {
	// Check if this is a ferry trip (route_type = 4)
	if trip.RouteType == 4 {
		v.validateFerryBikeAllowance(container, trip)
	}

	// Validate bikes_allowed values are valid
	if trip.BikesAllowed != nil {
		v.validateBikesAllowedValue(container, trip)
	}

	// Additional validation for bike-related accessibility
	v.validateBikeAccessibilityConsistency(container, trip)
}

// validateFerryBikeAllowance validates bike allowance for ferry trips
func (v *BikesAllowanceValidator) validateFerryBikeAllowance(container *notice.NoticeContainer, trip *TripBikeInfo) {
	if trip.BikesAllowed == nil {
		// Ferry trips should specify bike allowance
		container.AddNotice(notice.NewMissingBikesAllowedForFerryNotice(
			trip.TripID,
			trip.RouteID,
			trip.RowNumber,
		))
	}
}

// validateBikesAllowedValue validates that bikes_allowed has a valid value
func (v *BikesAllowanceValidator) validateBikesAllowedValue(container *notice.NoticeContainer, trip *TripBikeInfo) {
	if trip.BikesAllowed == nil {
		return
	}

	// Valid values are: 0 (no information), 1 (bikes allowed), 2 (bikes not allowed)
	validValues := []int{0, 1, 2}
	isValid := false
	for _, valid := range validValues {
		if *trip.BikesAllowed == valid {
			isValid = true
			break
		}
	}

	if !isValid {
		container.AddNotice(notice.NewInvalidBikesAllowedValueNotice(
			trip.TripID,
			*trip.BikesAllowed,
			trip.RowNumber,
		))
	}
}

// validateBikeAccessibilityConsistency validates consistency between bike and wheelchair accessibility
func (v *BikesAllowanceValidator) validateBikeAccessibilityConsistency(container *notice.NoticeContainer, trip *TripBikeInfo) {
	// This is an informational check - if bikes are explicitly not allowed but wheelchairs are,
	// it might indicate an accessibility issue worth noting
	if trip.BikesAllowed != nil && trip.WheelchairAccessible != nil {
		if *trip.BikesAllowed == 2 && *trip.WheelchairAccessible == 1 {
			// Bikes not allowed but wheelchairs are - this is fine, just informational
			container.AddNotice(notice.NewBikeWheelchairAccessibilityMismatchNotice(
				trip.TripID,
				trip.RouteID,
				*trip.BikesAllowed,
				*trip.WheelchairAccessible,
				trip.RowNumber,
			))
		}
	}

	// Check for unusual combinations
	if trip.BikesAllowed != nil && trip.RouteType != 4 { // Non-ferry routes
		if *trip.BikesAllowed == 1 {
			// Bikes allowed on non-ferry routes - this is unusual but not invalid
			// Only report for route types where bikes are uncommon
			if v.isBikeUncommonRouteType(trip.RouteType) {
				container.AddNotice(notice.NewUnusualBikeAllowanceNotice(
					trip.TripID,
					trip.RouteID,
					trip.RouteType,
					*trip.BikesAllowed,
					trip.RowNumber,
				))
			}
		}
	}
}

// isBikeUncommonRouteType checks if bikes are uncommon for this route type
func (v *BikesAllowanceValidator) isBikeUncommonRouteType(routeType int) bool {
	// Route types where bikes are typically uncommon:
	// 0 = Tram, 1 = Subway, 2 = Rail, 5 = Cable tram, 6 = Aerial lift, 7 = Funicular
	uncommonTypes := []int{0, 1, 2, 5, 6, 7}
	
	for _, uncommon := range uncommonTypes {
		if routeType == uncommon {
			return true
		}
	}
	
	return false
}