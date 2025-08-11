package entity

import (
	"io"
	"strconv"
	"strings"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/parser"
	"github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

// DuplicateRouteNameValidator validates route names are unique within agency/route type
type DuplicateRouteNameValidator struct{}

// NewDuplicateRouteNameValidator creates a new duplicate route name validator
func NewDuplicateRouteNameValidator() *DuplicateRouteNameValidator {
	return &DuplicateRouteNameValidator{}
}

// RouteInfo represents route information for duplication checking
type RouteInfo struct {
	RouteID        string
	RouteLongName  string
	RouteShortName string
	AgencyID       string
	RouteType      int
	RowNumber      int
}

// Validate checks for duplicate route names within same agency and route type
func (v *DuplicateRouteNameValidator) Validate(loader *parser.FeedLoader, container *notice.NoticeContainer, config validator.Config) {
	routes := v.loadRoutes(loader)
	if len(routes) == 0 {
		return
	}

	// Group routes by agency_id and route_type
	routeGroups := make(map[string][]RouteInfo)

	for _, route := range routes {
		// Create key from agency_id and route_type
		key := route.AgencyID + "_" + strconv.Itoa(route.RouteType)
		routeGroups[key] = append(routeGroups[key], route)
	}

	// Check each group for duplicates
	for _, group := range routeGroups {
		v.checkGroupForDuplicates(container, group)
	}
}

// loadRoutes loads route information from routes.txt
func (v *DuplicateRouteNameValidator) loadRoutes(loader *parser.FeedLoader) []RouteInfo {
	var routes []RouteInfo

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
			routes = append(routes, *route)
		}
	}

	return routes
}

// parseRoute parses a route record
func (v *DuplicateRouteNameValidator) parseRoute(row *parser.CSVRow) *RouteInfo {
	routeID, hasRouteID := row.Values["route_id"]
	if !hasRouteID || strings.TrimSpace(routeID) == "" {
		return nil
	}

	route := &RouteInfo{
		RouteID:   strings.TrimSpace(routeID),
		RowNumber: row.RowNumber,
	}

	// Get route names
	if longName, hasLongName := row.Values["route_long_name"]; hasLongName {
		route.RouteLongName = strings.TrimSpace(longName)
	}
	if shortName, hasShortName := row.Values["route_short_name"]; hasShortName {
		route.RouteShortName = strings.TrimSpace(shortName)
	}

	// Get agency_id (optional, defaults to first agency if not specified)
	if agencyID, hasAgencyID := row.Values["agency_id"]; hasAgencyID {
		route.AgencyID = strings.TrimSpace(agencyID)
	} else {
		route.AgencyID = "" // Default agency
	}

	// Get route_type (required)
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

// checkGroupForDuplicates checks a group of routes for duplicate names
func (v *DuplicateRouteNameValidator) checkGroupForDuplicates(container *notice.NoticeContainer, routes []RouteInfo) {
	if len(routes) <= 1 {
		return
	}

	// Check for duplicate long names
	v.checkDuplicateLongNames(container, routes)

	// Check for duplicate short names
	v.checkDuplicateShortNames(container, routes)

	// Check for routes with same long name and short name combination
	v.checkDuplicateNameCombinations(container, routes)
}

// checkDuplicateLongNames checks for duplicate route_long_name within the group
func (v *DuplicateRouteNameValidator) checkDuplicateLongNames(container *notice.NoticeContainer, routes []RouteInfo) {
	longNameMap := make(map[string][]RouteInfo)

	for _, route := range routes {
		// Only check non-empty long names
		if route.RouteLongName != "" {
			// Normalize name for comparison (case-insensitive, trimmed whitespace)
			normalizedName := strings.ToLower(strings.TrimSpace(route.RouteLongName))
			longNameMap[normalizedName] = append(longNameMap[normalizedName], route)
		}
	}

	// Report duplicates
	for _, duplicateRoutes := range longNameMap {
		if len(duplicateRoutes) > 1 {
			// Get the original name from first route
			originalName := duplicateRoutes[0].RouteLongName

			for i := 1; i < len(duplicateRoutes); i++ {
				container.AddNotice(notice.NewDuplicateRouteLongNameNotice(
					duplicateRoutes[i].RouteID,
					originalName,
					duplicateRoutes[0].RouteID,
					duplicateRoutes[i].AgencyID,
					duplicateRoutes[i].RouteType,
					duplicateRoutes[i].RowNumber,
				))
			}
		}
	}
}

// checkDuplicateShortNames checks for duplicate route_short_name within the group
func (v *DuplicateRouteNameValidator) checkDuplicateShortNames(container *notice.NoticeContainer, routes []RouteInfo) {
	shortNameMap := make(map[string][]RouteInfo)

	for _, route := range routes {
		// Only check non-empty short names
		if route.RouteShortName != "" {
			// Normalize name for comparison (case-insensitive, trimmed whitespace)
			normalizedName := strings.ToLower(strings.TrimSpace(route.RouteShortName))
			shortNameMap[normalizedName] = append(shortNameMap[normalizedName], route)
		}
	}

	// Report duplicates
	for _, duplicateRoutes := range shortNameMap {
		if len(duplicateRoutes) > 1 {
			// Get the original name from first route
			originalName := duplicateRoutes[0].RouteShortName

			for i := 1; i < len(duplicateRoutes); i++ {
				container.AddNotice(notice.NewDuplicateRouteShortNameNotice(
					duplicateRoutes[i].RouteID,
					originalName,
					duplicateRoutes[0].RouteID,
					duplicateRoutes[i].AgencyID,
					duplicateRoutes[i].RouteType,
					duplicateRoutes[i].RowNumber,
				))
			}
		}
	}
}

// checkDuplicateNameCombinations checks for routes with identical name combinations
func (v *DuplicateRouteNameValidator) checkDuplicateNameCombinations(container *notice.NoticeContainer, routes []RouteInfo) {
	nameComboMap := make(map[string][]RouteInfo)

	for _, route := range routes {
		// Create combination key from both names (only if both exist)
		if route.RouteLongName != "" && route.RouteShortName != "" {
			// Normalize both names for comparison
			normalizedLong := strings.ToLower(strings.TrimSpace(route.RouteLongName))
			normalizedShort := strings.ToLower(strings.TrimSpace(route.RouteShortName))
			comboKey := normalizedLong + "|" + normalizedShort
			nameComboMap[comboKey] = append(nameComboMap[comboKey], route)
		}
	}

	// Report duplicates
	for _, duplicateRoutes := range nameComboMap {
		if len(duplicateRoutes) > 1 {
			// Get the original names from first route
			firstRoute := duplicateRoutes[0]

			for i := 1; i < len(duplicateRoutes); i++ {
				container.AddNotice(notice.NewDuplicateRouteNameCombinationNotice(
					duplicateRoutes[i].RouteID,
					firstRoute.RouteLongName,
					firstRoute.RouteShortName,
					firstRoute.RouteID,
					duplicateRoutes[i].AgencyID,
					duplicateRoutes[i].RouteType,
					duplicateRoutes[i].RowNumber,
				))
			}
		}
	}
}
