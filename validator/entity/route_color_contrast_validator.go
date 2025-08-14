package entity

import (
	"io"
	"log"
	"math"
	"strconv"
	"strings"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/parser"
	"github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

// RouteColorContrastValidator validates color contrast between route_color and route_text_color
type RouteColorContrastValidator struct{}

// NewRouteColorContrastValidator creates a new route color contrast validator
func NewRouteColorContrastValidator() *RouteColorContrastValidator {
	return &RouteColorContrastValidator{}
}

// ColorInfo represents RGB color information
type ColorInfo struct {
	R, G, B   int
	Hex       string
	IsDefault bool
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

	route := &RouteColorInfo{
		RouteID:   strings.TrimSpace(routeID),
		RowNumber: row.RowNumber,
	}

	// Parse route_color (defaults to white if not specified)
	if routeColorStr, hasRouteColor := row.Values["route_color"]; hasRouteColor && strings.TrimSpace(routeColorStr) != "" {
		route.RouteColor = v.parseColor(strings.TrimSpace(routeColorStr), false)
	} else {
		route.RouteColor = v.parseColor("FFFFFF", true) // Default white
	}

	// Parse route_text_color (defaults to black if not specified)
	if routeTextColorStr, hasRouteTextColor := row.Values["route_text_color"]; hasRouteTextColor && strings.TrimSpace(routeTextColorStr) != "" {
		route.RouteTextColor = v.parseColor(strings.TrimSpace(routeTextColorStr), false)
	} else {
		route.RouteTextColor = v.parseColor("000000", true) // Default black
	}

	// Only return if at least one color is valid
	if route.RouteColor == nil && route.RouteTextColor == nil {
		return nil
	}

	return route
}

// parseColor parses a hex color string into ColorInfo
func (v *RouteColorContrastValidator) parseColor(hexStr string, isDefault bool) *ColorInfo {
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
		R:         int(r),
		G:         int(g),
		B:         int(b),
		Hex:       strings.ToUpper(hexStr),
		IsDefault: isDefault,
	}
}

// validateRouteColors validates color contrast for a route
func (v *RouteColorContrastValidator) validateRouteColors(container *notice.NoticeContainer, route RouteColorInfo) {
	// Skip validation if either color is invalid
	if route.RouteColor == nil || route.RouteTextColor == nil {
		return
	}

	// Calculate contrast ratio
	contrastRatio := v.calculateContrastRatio(route.RouteColor, route.RouteTextColor)

	// WCAG AA standard requires contrast ratio of at least 4.5:1 for normal text
	// WCAG AAA standard requires 7:1, but for transportation we'll use 4.5:1
	minimumContrast := 4.5

	if contrastRatio < minimumContrast {
		// Since Google accepts feeds with poor contrast, use WARNING instead of ERROR
		// Only use ERROR for extremely poor contrast that would be completely unreadable
		var severity notice.SeverityLevel
		if contrastRatio < 1.5 {
			severity = notice.ERROR // Extremely poor contrast (essentially unreadable)
		} else {
			severity = notice.WARNING // Poor but acceptable contrast
		}

		container.AddNotice(notice.NewRouteColorContrastNotice(
			route.RouteID,
			route.RouteColor.Hex,
			route.RouteTextColor.Hex,
			contrastRatio,
			minimumContrast,
			route.RowNumber,
			severity,
		))
	}

	// Additional checks for common problematic combinations
	v.checkProblematicColorCombinations(container, route)
}

// calculateContrastRatio calculates WCAG contrast ratio between two colors
func (v *RouteColorContrastValidator) calculateContrastRatio(color1, color2 *ColorInfo) float64 {
	// Calculate relative luminance for each color
	lum1 := v.calculateRelativeLuminance(color1)
	lum2 := v.calculateRelativeLuminance(color2)

	// Ensure lighter color is numerator
	lighter := math.Max(lum1, lum2)
	darker := math.Min(lum1, lum2)

	// Calculate contrast ratio
	return (lighter + 0.05) / (darker + 0.05)
}

// calculateRelativeLuminance calculates relative luminance according to WCAG formula
func (v *RouteColorContrastValidator) calculateRelativeLuminance(color *ColorInfo) float64 {
	// Convert RGB to linear RGB
	r := v.linearizeColorComponent(float64(color.R) / 255.0)
	g := v.linearizeColorComponent(float64(color.G) / 255.0)
	b := v.linearizeColorComponent(float64(color.B) / 255.0)

	// Calculate luminance using WCAG formula
	return 0.2126*r + 0.7152*g + 0.0722*b
}

// linearizeColorComponent applies gamma correction to color component
func (v *RouteColorContrastValidator) linearizeColorComponent(component float64) float64 {
	if component <= 0.03928 {
		return component / 12.92
	}
	return math.Pow((component+0.055)/1.055, 2.4)
}

// checkProblematicColorCombinations checks for known problematic color combinations
func (v *RouteColorContrastValidator) checkProblematicColorCombinations(container *notice.NoticeContainer, route RouteColorInfo) {
	// Check for white text on light backgrounds
	if v.isLightColor(route.RouteColor) && v.isLightColor(route.RouteTextColor) {
		container.AddNotice(notice.NewLightTextOnLightBackgroundNotice(
			route.RouteID,
			route.RouteColor.Hex,
			route.RouteTextColor.Hex,
			route.RowNumber,
		))
	}

	// Check for dark text on dark backgrounds
	if v.isDarkColor(route.RouteColor) && v.isDarkColor(route.RouteTextColor) {
		container.AddNotice(notice.NewDarkTextOnDarkBackgroundNotice(
			route.RouteID,
			route.RouteColor.Hex,
			route.RouteTextColor.Hex,
			route.RowNumber,
		))
	}

	// Check for identical or nearly identical colors
	if v.colorsAreTooSimilar(route.RouteColor, route.RouteTextColor) {
		container.AddNotice(notice.NewSimilarColorsNotice(
			route.RouteID,
			route.RouteColor.Hex,
			route.RouteTextColor.Hex,
			route.RowNumber,
		))
	}

	// Check for red-green combinations (colorblind accessibility)
	if v.isRedGreenCombination(route.RouteColor, route.RouteTextColor) {
		container.AddNotice(notice.NewRedGreenColorCombinationNotice(
			route.RouteID,
			route.RouteColor.Hex,
			route.RouteTextColor.Hex,
			route.RowNumber,
		))
	}
}

// isLightColor determines if a color is light (luminance > 0.5)
func (v *RouteColorContrastValidator) isLightColor(color *ColorInfo) bool {
	luminance := v.calculateRelativeLuminance(color)
	return luminance > 0.5
}

// isDarkColor determines if a color is dark (luminance < 0.2)
func (v *RouteColorContrastValidator) isDarkColor(color *ColorInfo) bool {
	luminance := v.calculateRelativeLuminance(color)
	return luminance < 0.2
}

// colorsAreTooSimilar checks if two colors are too similar
func (v *RouteColorContrastValidator) colorsAreTooSimilar(color1, color2 *ColorInfo) bool {
	// Calculate Euclidean distance in RGB space
	dr := float64(color1.R - color2.R)
	dg := float64(color1.G - color2.G)
	db := float64(color1.B - color2.B)

	distance := math.Sqrt(dr*dr + dg*dg + db*db)

	// If distance is very small, colors are too similar
	return distance < 50.0 // Threshold for "too similar"
}

// isRedGreenCombination checks for red-green color combinations (problematic for colorblind users)
func (v *RouteColorContrastValidator) isRedGreenCombination(color1, color2 *ColorInfo) bool {
	isRed1 := v.isRedish(color1)
	isGreen1 := v.isGreenish(color1)
	isRed2 := v.isRedish(color2)
	isGreen2 := v.isGreenish(color2)

	return (isRed1 && isGreen2) || (isGreen1 && isRed2)
}

// isRedish determines if a color is reddish
func (v *RouteColorContrastValidator) isRedish(color *ColorInfo) bool {
	return color.R > color.G+30 && color.R > color.B+30 && color.R > 100
}

// isGreenish determines if a color is greenish
func (v *RouteColorContrastValidator) isGreenish(color *ColorInfo) bool {
	return color.G > color.R+30 && color.G > color.B+30 && color.G > 100
}
