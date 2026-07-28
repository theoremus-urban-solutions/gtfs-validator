package entity

import (
	"io"
	"log"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

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

	shortName := strings.TrimSpace(routeShortName)
	longName := strings.TrimSpace(routeLongName)

	// Applications routinely render the two names side by side, so a long name
	// that already carries the short one shows it twice ("14 Route 14").
	if !shortNameEmpty && !longNameEmpty && containsAsWord(longName, shortName) {
		container.AddNotice(notice.NewRouteLongNameContainsShortNameNotice(
			strings.TrimSpace(routeID),
			shortName,
			longName,
			row.RowNumber,
		))
	}

	v.validateRouteDescription(container, row, strings.TrimSpace(routeID), shortName, longName)

	// route_long_name is prose read by riders, unlike route_short_name, which
	// is normally a code and legitimately upper case.
	if needsMixedCase(longName) {
		container.AddNotice(notice.NewMixedCaseRecommendedFieldNotice(
			"routes.txt",
			"route_long_name",
			longName,
			row.RowNumber,
		))
	}

	// Validate route type specific naming conventions
	if hasRouteType {
		v.validateRouteTypeNaming(container, row, routeType, routeShortName, routeLongName)
	}
}

// validateRouteDescription checks that route_desc says something the names do
// not. The spec asks it for "useful, quality information" precisely because a
// description that restates the name costs a line of the rider's screen and
// tells them nothing.
func (v *RouteNameValidator) validateRouteDescription(container *notice.NoticeContainer, row *parser.CSVRow, routeID string, shortName string, longName string) {
	routeDesc := strings.TrimSpace(row.Values["route_desc"])
	if routeDesc == "" {
		return
	}

	// Casing alone does not make a description informative, so "RED LINE" as
	// the description of "Red Line" is still a duplicate.
	duplicatedField := ""
	switch {
	case shortName != "" && strings.EqualFold(routeDesc, shortName):
		duplicatedField = "route_short_name"
	case longName != "" && strings.EqualFold(routeDesc, longName):
		duplicatedField = "route_long_name"
	default:
		return
	}

	container.AddNotice(notice.NewSameNameAndDescriptionForRouteNotice(
		routeID,
		routeDesc,
		duplicatedField,
		row.RowNumber,
	))
}

// containsAsWord reports whether needle occurs in haystack bounded by
// something other than a letter or a digit.
//
// A plain substring test would flag route "1" for the long name "Route 100",
// which is not the defect the rule describes: the complaint is that the short
// name is repeated, and "100" is a different name that merely starts with the
// same digit.
func containsAsWord(haystack string, needle string) bool {
	text := []rune(strings.ToLower(haystack))
	word := []rune(strings.ToLower(needle))
	if len(word) == 0 || len(word) > len(text) {
		return false
	}

	for i := 0; i+len(word) <= len(text); i++ {
		if string(text[i:i+len(word)]) != string(word) {
			continue
		}
		if i > 0 && isNameWordRune(text[i-1]) {
			continue
		}
		if end := i + len(word); end < len(text) && isNameWordRune(text[end]) {
			continue
		}
		return true
	}

	return false
}

// isNameWordRune reports whether a rune continues a word rather than
// delimiting one.
func isNameWordRune(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r)
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

	// Check for route names that are too long. The limit is a character count,
	// not a byte count: len() on a Go string counts bytes, which trips the
	// limit at six characters for a Cyrillic or Greek name.
	const maxShortNameLength = 12

	shortName := strings.TrimSpace(routeShortName)
	if nameLength := utf8.RuneCountInString(shortName); nameLength > maxShortNameLength {
		container.AddNotice(notice.NewRouteShortNameTooLongNotice(
			strings.TrimSpace(routeID),
			shortName,
			nameLength,
			maxShortNameLength,
			row.RowNumber,
		))
	}
}
