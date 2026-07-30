package entity

import (
	"io"
	"log"
	"strings"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/parser"
	"github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

// RouteConsistencyValidator validates route data consistency
type RouteConsistencyValidator struct{}

// NewRouteConsistencyValidator creates a new route consistency validator
func NewRouteConsistencyValidator() *RouteConsistencyValidator {
	return &RouteConsistencyValidator{}
}

// Validate checks route data consistency
func (v *RouteConsistencyValidator) Validate(loader *parser.FeedLoader, container *notice.NoticeContainer, config validator.Config) {
	v.validateRoutes(loader, container)
}

// validateRoutes validates route records
func (v *RouteConsistencyValidator) validateRoutes(loader *parser.FeedLoader, container *notice.NoticeContainer) {
	reader, err := loader.GetFile("routes.txt")
	if err != nil {
		return // No routes file
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close reader %v", closeErr)
		}
	}()

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

		v.validateRouteRecord(container, row)
	}
}

// validateRouteRecord validates a single route record
func (v *RouteConsistencyValidator) validateRouteRecord(container *notice.NoticeContainer, row *parser.CSVRow) {
	routeID, hasRouteID := row.Values["route_id"]
	if !hasRouteID {
		return // Other validators handle missing route_id
	}

	routeIDTrimmed := strings.TrimSpace(routeID)

	// Validate route type

	// Validate route color
	v.validateRouteColor(container, row, routeIDTrimmed)

	// Validate route URL
	v.validateRouteURL(container, row)
}

// validateRouteColor validates route_color and route_text_color fields
func (v *RouteConsistencyValidator) validateRouteColor(container *notice.NoticeContainer, row *parser.CSVRow, routeID string) {
	routeColor, hasRouteColor := row.Values["route_color"]
	routeTextColor, hasRouteTextColor := row.Values["route_text_color"]

	if hasRouteColor && strings.TrimSpace(routeColor) != "" {
		if !v.isValidHexColor(strings.TrimSpace(routeColor)) {
			container.AddNotice(notice.NewInvalidColorNotice(
				routeID,
				"route_color",
				strings.TrimSpace(routeColor),
				row.RowNumber,
			))
		}
	}

	if hasRouteTextColor && strings.TrimSpace(routeTextColor) != "" {
		if !v.isValidHexColor(strings.TrimSpace(routeTextColor)) {
			container.AddNotice(notice.NewInvalidColorNotice(
				routeID,
				"route_text_color",
				strings.TrimSpace(routeTextColor),
				row.RowNumber,
			))
		}
	}
}

// validateRouteURL validates the route_url field
func (v *RouteConsistencyValidator) validateRouteURL(container *notice.NoticeContainer, row *parser.CSVRow) {
	routeURL, hasRouteURL := row.Values["route_url"]
	if !hasRouteURL || strings.TrimSpace(routeURL) == "" {
		return // Route URL is optional
	}

	urlTrimmed := strings.TrimSpace(routeURL)
	if !v.isValidURL(urlTrimmed) {
		container.AddNotice(notice.NewInvalidURLNotice(
			"routes.txt",
			"route_url",
			urlTrimmed,
			row.RowNumber,
		))
	}
}

// isValidHexColor checks if a string is a valid 6-digit hex color
func (v *RouteConsistencyValidator) isValidHexColor(color string) bool {
	if len(color) != 6 {
		return false
	}

	for _, char := range color {
		if (char < '0' || char > '9') &&
			(char < 'A' || char > 'F') &&
			(char < 'a' || char > 'f') {
			return false
		}
	}

	return true
}

// isValidURL performs basic URL validation
func (v *RouteConsistencyValidator) isValidURL(url string) bool {
	// Basic URL validation - must start with http:// or https://
	return strings.HasPrefix(strings.ToLower(url), "http://") ||
		strings.HasPrefix(strings.ToLower(url), "https://")
}
