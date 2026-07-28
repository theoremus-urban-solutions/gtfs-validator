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

// RouteTypeValidator validates route_type values and consistency
type RouteTypeValidator struct{}

// NewRouteTypeValidator creates a new route type validator
func NewRouteTypeValidator() *RouteTypeValidator {
	return &RouteTypeValidator{}
}

// RouteTypeInfo represents route type information
type RouteTypeInfo struct {
	RouteID        string
	RouteType      int
	RouteShortName string
	RouteLongName  string
	AgencyID       string
	RowNumber      int
}

// Validate checks route_type values for validity and consistency
func (v *RouteTypeValidator) Validate(loader *parser.FeedLoader, container *notice.NoticeContainer, config validator.Config) {
	routes := v.loadRoutes(loader)
	if len(routes) == 0 {
		return
	}

	for _, route := range routes {
		v.validateRouteType(container, route)
	}

	// Check for route type distribution patterns
}

// loadRoutes loads route information from routes.txt
func (v *RouteTypeValidator) loadRoutes(loader *parser.FeedLoader) []*RouteTypeInfo {
	var routes []*RouteTypeInfo

	reader, err := loader.GetFile("routes.txt")
	if err != nil {
		return routes
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close reader %v", closeErr)
		}
	}()

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
			routes = append(routes, route)
		}
	}

	return routes
}

// parseRoute parses route information
func (v *RouteTypeValidator) parseRoute(row *parser.CSVRow) *RouteTypeInfo {
	routeID, hasRouteID := row.Values["route_id"]
	routeTypeStr, hasRouteType := row.Values["route_type"]

	if !hasRouteID || !hasRouteType {
		return nil
	}

	routeType, err := strconv.Atoi(strings.TrimSpace(routeTypeStr))
	if err != nil {
		return nil
	}

	route := &RouteTypeInfo{
		RouteID:   strings.TrimSpace(routeID),
		RouteType: routeType,
		RowNumber: row.RowNumber,
	}

	// Parse optional fields
	if shortName, hasShortName := row.Values["route_short_name"]; hasShortName {
		route.RouteShortName = strings.TrimSpace(shortName)
	}
	if longName, hasLongName := row.Values["route_long_name"]; hasLongName {
		route.RouteLongName = strings.TrimSpace(longName)
	}
	if agencyID, hasAgencyID := row.Values["agency_id"]; hasAgencyID {
		route.AgencyID = strings.TrimSpace(agencyID)
	}

	return route
}

// validateRouteType validates individual route type values
func (v *RouteTypeValidator) validateRouteType(container *notice.NoticeContainer, route *RouteTypeInfo) {
	// Check if route type is valid according to GTFS specification
	if !v.isValidRouteType(route.RouteType) {
		container.AddNotice(notice.NewInvalidRouteTypeNotice(
			route.RouteID,
			strconv.Itoa(route.RouteType),
			route.RowNumber,
			"Invalid route type value",
		))
		return
	}

	// Check for deprecated route types
	if v.isDeprecatedRouteType(route.RouteType) {
		container.AddNotice(notice.NewDeprecatedRouteTypeNotice(
			route.RouteID,
			route.RouteType,
			v.getRecommendedRouteType(route.RouteType),
			route.RowNumber,
		))
	}

	// Check for uncommon route types that might be mistakes
}

// isValidRouteType checks if route type is valid according to GTFS spec
func (v *RouteTypeValidator) isValidRouteType(routeType int) bool {
	// Basic GTFS route types (0-12)
	basicTypes := map[int]bool{
		0:  true, // Tram, Streetcar, Light rail
		1:  true, // Subway, Metro
		2:  true, // Rail
		3:  true, // Bus
		4:  true, // Ferry
		5:  true, // Cable tram
		6:  true, // Aerial lift, suspended cable car
		7:  true, // Funicular
		11: true, // Trolleybus
		12: true, // Monorail
	}

	// Extended route types (100-1700)
	if routeType >= 100 && routeType <= 1700 {
		return v.isValidExtendedRouteType(routeType)
	}

	return basicTypes[routeType]
}

// isValidExtendedRouteType checks extended route types
func (v *RouteTypeValidator) isValidExtendedRouteType(routeType int) bool {
	// Extended GTFS route types based on NeTEx/Transmodel
	switch {
	case routeType >= 100 && routeType <= 117: // Railway Service
		return true
	case routeType >= 200 && routeType <= 209: // Coach Service
		return true
	case routeType >= 300 && routeType <= 399: // Suburban Railway
		return true
	case routeType >= 400 && routeType <= 499: // Urban Railway
		return true
	case routeType >= 500 && routeType <= 599: // Metro Service
		return true
	case routeType >= 600 && routeType <= 699: // Underground Service
		return true
	case routeType >= 700 && routeType <= 799: // Bus Service
		return true
	case routeType >= 800 && routeType <= 899: // Trolleybus Service
		return true
	case routeType >= 900 && routeType <= 999: // Tram Service
		return true
	case routeType >= 1000 && routeType <= 1099: // Water Transport Service
		return true
	case routeType >= 1100 && routeType <= 1199: // Air Service
		return true
	case routeType >= 1200 && routeType <= 1299: // Ferry Service
		return true
	case routeType >= 1300 && routeType <= 1399: // Aerial Lift Service
		return true
	case routeType >= 1400 && routeType <= 1499: // Funicular Service
		return true
	case routeType >= 1500 && routeType <= 1599: // Taxi Service
		return true
	case routeType >= 1700 && routeType <= 1799: // Miscellaneous Service
		return true
	}
	return false
}

// isDeprecatedRouteType checks if route type is deprecated
func (v *RouteTypeValidator) isDeprecatedRouteType(routeType int) bool {
	// Currently no officially deprecated route types in GTFS spec
	// This function can be expanded in future if route types get deprecated
	_ = routeType // Acknowledge parameter until implementation needed
	return false
}

// isUncommonRouteType checks if route type is uncommon
func (v *RouteTypeValidator) isUncommonRouteType(routeType int) bool {
	uncommonTypes := map[int]bool{
		5:  true, // Cable tram
		6:  true, // Aerial lift
		7:  true, // Funicular
		11: true, // Trolleybus
		12: true, // Monorail
	}

	// Extended types are generally less common
	if routeType >= 100 {
		return true
	}

	return uncommonTypes[routeType]
}

// isUncommonAsOnlyRouteType checks if route type is unusual as the only type in feed
func (v *RouteTypeValidator) isUncommonAsOnlyRouteType(routeType int) bool {
	// These types are unusual as the only transit mode in a feed
	return routeType == 4 || routeType == 5 || routeType == 6 ||
		routeType == 7 || routeType == 11 || routeType == 12
}

// getRecommendedRouteType returns recommended replacement for deprecated types
func (v *RouteTypeValidator) getRecommendedRouteType(routeType int) int {
	// Currently no recommendations, but could add mappings for deprecated types
	return routeType
}

// getRouteTypeDescription returns human-readable description
func (v *RouteTypeValidator) getRouteTypeDescription(routeType int) string {
	descriptions := map[int]string{
		0:  "Tram, Streetcar, Light rail",
		1:  "Subway, Metro",
		2:  "Rail",
		3:  "Bus",
		4:  "Ferry",
		5:  "Cable tram",
		6:  "Aerial lift, suspended cable car",
		7:  "Funicular",
		11: "Trolleybus",
		12: "Monorail",
	}

	if desc, exists := descriptions[routeType]; exists {
		return desc
	}

	// Extended route types
	switch {
	case routeType >= 100 && routeType <= 117:
		return "Railway Service"
	case routeType >= 200 && routeType <= 209:
		return "Coach Service"
	case routeType >= 700 && routeType <= 799:
		return "Bus Service"
	case routeType >= 1000 && routeType <= 1099:
		return "Water Transport Service"
	case routeType >= 1300 && routeType <= 1399:
		return "Aerial Lift Service"
	default:
		return "Extended Route Type"
	}
}
