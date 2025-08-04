package entity

import (
	"io"
	"strconv"
	"strings"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/parser"
	"github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

// RouteNameValidator validates route naming according to GTFS best practices
type RouteNameValidator struct{}

// NewRouteNameValidator creates a new route name validator
func NewRouteNameValidator() *RouteNameValidator {
	return &RouteNameValidator{}
}

// Validate checks route naming best practices
func (v *RouteNameValidator) Validate(loader *parser.FeedLoader, container *notice.NoticeContainer, config validator.Config) {
	reader, err := loader.GetFile("routes.txt")
	if err != nil {
		return // File doesn't exist, skip validation
	}
	defer reader.Close()

	csvFile, err := parser.NewCSVFile(reader, "routes.txt")
	if err != nil {
		return
	}

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}

		v.validateRoute(container, row)
	}
}

// validateRoute validates a single route record
func (v *RouteNameValidator) validateRoute(container *notice.NoticeContainer, row *parser.CSVRow) {
	routeID, hasRouteID := row.Values["route_id"]
	routeShortName, hasRouteShortName := row.Values["route_short_name"]
	routeLongName, hasRouteLongName := row.Values["route_long_name"]
	routeType, hasRouteType := row.Values["route_type"]

	if !hasRouteID {
		return // Other validators handle missing route_id
	}

	// Check if both route_short_name and route_long_name are empty
	shortNameEmpty := !hasRouteShortName || strings.TrimSpace(routeShortName) == ""
	longNameEmpty := !hasRouteLongName || strings.TrimSpace(routeLongName) == ""

	if shortNameEmpty && longNameEmpty {
		container.AddNotice(notice.NewMissingRouteNameNotice(
			strings.TrimSpace(routeID),
			row.RowNumber,
		))
		return
	}

	// Check for identical short and long names
	if !shortNameEmpty && !longNameEmpty {
		shortName := strings.TrimSpace(routeShortName)
		longName := strings.TrimSpace(routeLongName)

		if shortName == longName {
			container.AddNotice(notice.NewSameNameAndDescriptionNotice(
				strings.TrimSpace(routeID),
				"route_short_name",
				"route_long_name",
				shortName,
				row.RowNumber,
			))
		}
	}

	// Validate route type specific naming conventions
	if hasRouteType {
		v.validateRouteTypeNaming(container, row, routeType, routeShortName, routeLongName)
	}
}

// validateRouteTypeNaming validates naming conventions specific to route types
func (v *RouteNameValidator) validateRouteTypeNaming(container *notice.NoticeContainer, row *parser.CSVRow, routeTypeStr string, routeShortName string, routeLongName string) {
	routeType, err := strconv.Atoi(strings.TrimSpace(routeTypeStr))
	if err != nil {
		return // Invalid route type, other validators handle this
	}

	routeID := row.Values["route_id"]

	// For bus routes (type 3), recommend having a short name
	if routeType == 3 && (routeShortName == "" || strings.TrimSpace(routeShortName) == "") {
		container.AddNotice(notice.NewMissingRecommendedFieldNotice(
			"routes.txt",
			"route_short_name",
			row.RowNumber,
		))
	}

	// For rail routes (types 0, 1, 2), recommend having both short and long names
	if routeType == 0 || routeType == 1 || routeType == 2 {
		if routeShortName == "" || strings.TrimSpace(routeShortName) == "" {
			container.AddNotice(notice.NewMissingRecommendedFieldNotice(
				"routes.txt",
				"route_short_name",
				row.RowNumber,
			))
		}
		if routeLongName == "" || strings.TrimSpace(routeLongName) == "" {
			container.AddNotice(notice.NewMissingRecommendedFieldNotice(
				"routes.txt",
				"route_long_name",
				row.RowNumber,
			))
		}
	}

	// Check for route names that are too long
	const maxShortNameLength = 12
	const maxLongNameLength = 100

	if routeShortName != "" && len(strings.TrimSpace(routeShortName)) > maxShortNameLength {
		container.AddNotice(notice.NewRouteShortNameTooLongNotice(
			strings.TrimSpace(routeID),
			strings.TrimSpace(routeShortName),
			len(strings.TrimSpace(routeShortName)),
			maxShortNameLength,
			row.RowNumber,
		))
	}

	if routeLongName != "" && len(strings.TrimSpace(routeLongName)) > maxLongNameLength {
		container.AddNotice(notice.NewRouteLongNameTooLongNotice(
			strings.TrimSpace(routeID),
			strings.TrimSpace(routeLongName),
			len(strings.TrimSpace(routeLongName)),
			maxLongNameLength,
			row.RowNumber,
		))
	}
}
