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

// minLumaDifference is the smallest gap in perceived brightness that still
// leaves a route name legible against the route's own color.
//
// W3C AERT (http://www.w3.org/TR/2000/WD-AERT-20000426#color-contrast) asks for
// 125, but that figure is written for body text. A route name is rendered the
// way a logo is — large, short, and in a solid patch of color — so a smaller
// gap still reads, and 125 would condemn pairings riders have no trouble with.
const minLumaDifference = 72

// unreadableLumaDifference is where the two colors stop being merely low
// contrast and become the same shade to the eye, leaving the name invisible
// rather than just hard to read.
const unreadableLumaDifference = 10

// RouteColorContrastValidator validates color contrast between route_color and route_text_color
type RouteColorContrastValidator struct{}

// NewRouteColorContrastValidator creates a new route color contrast validator
func NewRouteColorContrastValidator() *RouteColorContrastValidator {
	return &RouteColorContrastValidator{}
}

// ColorInfo represents RGB color information
type ColorInfo struct {
	R, G, B int
	Hex     string
}

// RouteColorInfo represents route color information
type RouteColorInfo struct {
	RouteID        string
	RouteColor     *ColorInfo
	RouteTextColor *ColorInfo
	RowNumber      int
}

// Validate checks color contrast for routes
func (v *RouteColorContrastValidator) Validate(loader *parser.FeedLoader, container *notice.NoticeContainer, config validator.Config) {
	routes := v.loadRouteColors(loader)

	for _, route := range routes {
		v.validateRouteColors(container, route)
	}
}

// loadRouteColors loads route color information from routes.txt
func (v *RouteColorContrastValidator) loadRouteColors(loader *parser.FeedLoader) []RouteColorInfo {
	var routes []RouteColorInfo

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

		route := v.parseRouteColors(row)
		if route != nil {
			routes = append(routes, *route)
		}
	}

	return routes
}

// parseRouteColors parses route color information from a row
func (v *RouteColorContrastValidator) parseRouteColors(row *parser.CSVRow) *RouteColorInfo {
	routeID, hasRouteID := row.Values["route_id"]
	if !hasRouteID {
		return nil
	}

	// A color the agency left out is not a contrast defect. The spec fills the
	// gap with white behind black text, which contrasts by construction, so
	// judging a chosen color against a default nobody picked would either say
	// nothing or report a clash the feed does not contain.
	routeColor := v.parseColor(strings.TrimSpace(row.Values["route_color"]))
	routeTextColor := v.parseColor(strings.TrimSpace(row.Values["route_text_color"]))
	if routeColor == nil || routeTextColor == nil {
		return nil
	}

	return &RouteColorInfo{
		RouteID:        strings.TrimSpace(routeID),
		RouteColor:     routeColor,
		RouteTextColor: routeTextColor,
		RowNumber:      row.RowNumber,
	}
}

// parseColor parses a hex color string into ColorInfo
func (v *RouteColorContrastValidator) parseColor(hexStr string) *ColorInfo {
	// Remove # if present
	hexStr = strings.TrimPrefix(hexStr, "#")

	// Must be exactly 6 characters
	if len(hexStr) != 6 {
		return nil
	}

	// Parse hex components
	r, err1 := strconv.ParseInt(hexStr[0:2], 16, 64)
	g, err2 := strconv.ParseInt(hexStr[2:4], 16, 64)
	b, err3 := strconv.ParseInt(hexStr[4:6], 16, 64)

	if err1 != nil || err2 != nil || err3 != nil {
		return nil
	}

	return &ColorInfo{
		R:   int(r),
		G:   int(g),
		B:   int(b),
		Hex: strings.ToUpper(hexStr),
	}
}

// validateRouteColors reports a route whose name would not stand out against
// the route's own color.
//
// The measure is a gap in perceived brightness, not a WCAG contrast ratio.
// WCAG ratios are calibrated for body text against a 4.5 threshold, which
// condemns pairings riders read without effort: white on a dark green badge
// scores 4.47 and fails, though nobody has trouble with it. Comparing luma
// asks the narrower question a route badge actually poses, which is whether
// large text separates from its background at a glance.
func (v *RouteColorContrastValidator) validateRouteColors(container *notice.NoticeContainer, route RouteColorInfo) {
	lumaDifference := rec601Luma(route.RouteColor) - rec601Luma(route.RouteTextColor)
	if lumaDifference < 0 {
		lumaDifference = -lumaDifference
	}
	if lumaDifference >= minLumaDifference {
		return
	}

	// Low contrast is a legibility complaint, not a broken feed, so it stays a
	// warning until the name is not merely faint but absent.
	severity := notice.WARNING
	if lumaDifference < unreadableLumaDifference {
		severity = notice.ERROR
	}

	container.AddNotice(notice.NewRouteColorContrastNotice(
		route.RouteID,
		route.RouteColor.Hex,
		route.RouteTextColor.Hex,
		float64(lumaDifference),
		float64(minLumaDifference),
		route.RowNumber,
		severity,
	))
}

// rec601Luma returns how bright a color looks, on the same 0-255 scale as its
// components.
//
// The weights are Rec. 601 luma (https://en.wikipedia.org/wiki/Luma_(video)):
// green carries most of the apparent brightness and blue almost none, which is
// why pure blue reads as dark and pure yellow as light. The result is truncated
// to an integer so that the difference reported here is the one the canonical
// validator reports.
func rec601Luma(color *ColorInfo) int {
	return int(0.30*float64(color.R) + 0.59*float64(color.G) + 0.11*float64(color.B))
}
