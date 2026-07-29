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
	// that opens with the short one shows it twice ("14" + "14 Express").
	if !shortNameEmpty && !longNameEmpty && longNameLeadsWithShortName(longName, shortName) {
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

// longNameLeadsWithShortName reports whether longName is shortName followed by
// nothing at all, or by a separator.
//
// The published rule is worded as containment, and its own bad examples include
// "14"/"Route 14", where the short name follows a generic word. The canonical
// validator does not implement that: it tests only whether the long name begins
// with the short name and the next character is a space, "-", "(" or ")". The
// narrowing is deliberate and argued for in its source, so the prose is the
// loose artefact here and the narrower check is the real contract.
//
// We match the implementation rather than the prose, because parity is what
// consumers compare against — a feed that passes upstream and fails here is a
// support ticket, whichever reading is more defensible on paper. The cost is
// one-directional and worth stating plainly: the "Route 14" shape is common,
// and neither validator reports it. See CANONICAL_PARITY.md.
func longNameLeadsWithShortName(longName string, shortName string) bool {
	if shortName == "" || len(longName) < len(shortName) {
		return false
	}
	if !strings.EqualFold(longName[:len(shortName)], shortName) {
		return false
	}

	remainder := longName[len(shortName):]
	if remainder == "" {
		return true
	}

	// The short name has to end where it claims to: "21 Clark Rd" restates
	// route 21, whereas "216 Clark Rd" is a different route that merely opens
	// with the same digits.
	next, _ := utf8.DecodeRuneInString(remainder)
	return unicode.IsSpace(next) || strings.ContainsRune("-()", next)
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
