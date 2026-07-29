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

// DuplicateRouteNameValidator reports routes that another route of the same
// agency and route type is already indistinguishable from.
//
// Two routes collide only when the whole of what a passenger sees — short name
// and long name together — is identical. Sharing just one of the two names is
// ordinary: a family of routes commonly shares a long name while the short
// name separates the branches, and vice versa.
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

// Validate reports each route whose name pair is already taken.
func (v *DuplicateRouteNameValidator) Validate(loader *parser.FeedLoader, container *notice.NoticeContainer, config validator.Config) {
	// The routes are kept in file order so that the route named as the first
	// occurrence is the one a reader would reach first in routes.txt.
	firstByKey := make(map[string]RouteInfo)

	for _, route := range v.loadRoutes(loader) {
		key := v.routeKey(route)
		first, taken := firstByKey[key]
		if !taken {
			firstByKey[key] = route
			continue
		}
		container.AddNotice(notice.NewDuplicateRouteNameCombinationNotice(
			route.RouteID,
			first.RouteLongName,
			first.RouteShortName,
			first.RouteID,
			route.AgencyID,
			route.RouteType,
			route.RowNumber,
		))
	}
}

// routeKey identifies a route by everything that has to match for two routes
// to be confusable: both names, the mode, and the operator.
//
// The names are compared as written. Two routes differing only in case are
// distinguishable on a printed timetable, and feeds legitimately use casing to
// separate an all-caps terminus from a title-case one.
func (v *DuplicateRouteNameValidator) routeKey(route RouteInfo) string {
	return route.RouteLongName + "\x00" + route.RouteShortName + "\x00" +
		strconv.Itoa(route.RouteType) + "\x00" + route.AgencyID
}

// loadRoutes loads route information from routes.txt
func (v *DuplicateRouteNameValidator) loadRoutes(loader *parser.FeedLoader) []RouteInfo {
	var routes []RouteInfo

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
