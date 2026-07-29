package entity

import (
	"io"
	"log"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/parser"
	"github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

// NameComparisonValidator compares the URLs a feed publishes for its agencies,
// routes and stops.
//
// Each of agency_url, route_url and stop_url is meant to lead the rider to a
// different page. When a producer copies one into another the field stops
// carrying information, which no single-file validator can see — the defect
// only exists in the relation between two files, so it is checked here rather
// than in the per-file name validators.
type NameComparisonValidator struct{}

// NewNameComparisonValidator creates a new name comparison validator
func NewNameComparisonValidator() *NameComparisonValidator {
	return &NameComparisonValidator{}
}

// agencyRecord is the part of an agency.txt row these comparisons need.
type agencyRecord struct {
	AgencyID   string
	AgencyName string
	URL        string
	RowNumber  int
}

// routeRecord is the part of a routes.txt row these comparisons need.
type routeRecord struct {
	RouteID   string
	URL       string
	RowNumber int
}

// Validate checks that route and stop URLs are not copies of one another.
func (v *NameComparisonValidator) Validate(loader *parser.FeedLoader, container *notice.NoticeContainer, config validator.Config) {
	agenciesByURL := v.loadAgencies(loader, container)
	routesByURL := v.loadRoutes(loader, container, agenciesByURL)
	v.validateStops(loader, container, agenciesByURL, routesByURL)
}

// loadAgencies indexes agency.txt by URL and reports agency names that are not
// mixed case. The first agency wins a shared URL: reporting a route against
// every agency that shares it would multiply one defect by the size of
// agency.txt.
func (v *NameComparisonValidator) loadAgencies(loader *parser.FeedLoader, container *notice.NoticeContainer) map[string]*agencyRecord {
	agencies := make(map[string]*agencyRecord)

	reader, err := loader.GetFile("agency.txt")
	if err != nil {
		return agencies
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close reader %v", closeErr)
		}
	}()

	csvFile, err := parser.NewCSVFile(reader, "agency.txt")
	if err != nil {
		return agencies
	}

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		agency := &agencyRecord{
			AgencyID:   strings.TrimSpace(row.Values["agency_id"]),
			AgencyName: strings.TrimSpace(row.Values["agency_name"]),
			URL:        strings.TrimSpace(row.Values["agency_url"]),
			RowNumber:  row.RowNumber,
		}

		if needsMixedCase(agency.AgencyName) {
			container.AddNotice(notice.NewMixedCaseRecommendedFieldNotice(
				"agency.txt",
				"agency_name",
				agency.AgencyName,
				agency.RowNumber,
			))
		}

		if key := normalizeURL(agency.URL); key != "" {
			if _, seen := agencies[key]; !seen {
				agencies[key] = agency
			}
		}
	}

	return agencies
}

// loadRoutes indexes routes.txt by URL and reports routes whose URL is an
// agency URL.
func (v *NameComparisonValidator) loadRoutes(loader *parser.FeedLoader, container *notice.NoticeContainer, agenciesByURL map[string]*agencyRecord) map[string]*routeRecord {
	routes := make(map[string]*routeRecord)

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

		route := &routeRecord{
			RouteID:   strings.TrimSpace(row.Values["route_id"]),
			URL:       strings.TrimSpace(row.Values["route_url"]),
			RowNumber: row.RowNumber,
		}

		key := normalizeURL(route.URL)
		if key == "" {
			continue
		}

		if agency, ok := agenciesByURL[key]; ok {
			container.AddNotice(notice.NewSameRouteAndAgencyUrlNotice(
				route.RouteID,
				route.URL,
				route.RowNumber,
				agency.AgencyID,
				agency.AgencyName,
				agency.RowNumber,
			))
		}

		if _, seen := routes[key]; !seen {
			routes[key] = route
		}
	}

	return routes
}

// validateStops reports stop URLs copied from agency.txt or routes.txt.
func (v *NameComparisonValidator) validateStops(loader *parser.FeedLoader, container *notice.NoticeContainer, agenciesByURL map[string]*agencyRecord, routesByURL map[string]*routeRecord) {
	reader, err := loader.GetFile("stops.txt")
	if err != nil {
		return
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close reader %v", closeErr)
		}
	}()

	csvFile, err := parser.NewCSVFile(reader, "stops.txt")
	if err != nil {
		return
	}

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		stopID := strings.TrimSpace(row.Values["stop_id"])
		stopName := strings.TrimSpace(row.Values["stop_name"])
		stopURL := strings.TrimSpace(row.Values["stop_url"])

		key := normalizeURL(stopURL)
		if key == "" {
			continue
		}

		if agency, ok := agenciesByURL[key]; ok {
			container.AddNotice(notice.NewSameStopAndAgencyUrlNotice(
				stopID,
				stopName,
				stopURL,
				row.RowNumber,
				agency.AgencyID,
				agency.AgencyName,
				agency.RowNumber,
			))
		}

		if route, ok := routesByURL[key]; ok {
			container.AddNotice(notice.NewSameStopAndRouteUrlNotice(
				stopID,
				stopName,
				stopURL,
				row.RowNumber,
				route.RouteID,
				route.RowNumber,
			))
		}
	}
}

// normalizeURL folds the differences that do not change which page a URL
// reaches, so that two fields written by different hands still compare equal.
// Scheme and host are case-insensitive, and a trailing slash on a bare host is
// not a different page.
func normalizeURL(url string) string {
	trimmed := strings.TrimSpace(url)
	if trimmed == "" {
		return ""
	}
	return strings.TrimSuffix(strings.ToLower(trimmed), "/")
}

// needsMixedCase reports whether a customer-facing name fails the canonical
// Mixed Case rule. The rule exists because an all-caps name is read out as an
// initialism by screen readers, and no consumer can restore casing once the
// producer has thrown it away.
//
// The check works word by word rather than on the string as a whole, and a
// name is accepted the moment any one of its words carries both cases. That is
// the canonical validator's shape, not the published prose's, and we follow the
// implementation because parity is what producers compare against. Two of its
// consequences read as bugs if you only have the prose:
//
//   - A name of a single word is reported only when that word is entirely
//     lower case. A lone "GALLERIA" passes; "GALLERIA MALL" does not.
//   - A leading run of non-letters counts as a word towards the two-word
//     threshold without ever being able to supply the mixed case. This is what
//     makes "3427 GG 17" reportable — the rule's own bad example, which
//     otherwise has only the one word to its name.
//
// Words too short to carry a case distinction are ignored, and so are words
// written in a script that has none. Chinese, Arabic and Hebrew cannot answer
// this question; canonical still counts them towards the threshold, which
// reports every two-word Hebrew stop name in a feed. We skip them instead —
// the divergence can only ever suppress a notice about text that had no case
// to lose.
func needsMixedCase(value string) bool {
	words, leadingNonLetters := caseWords(value)
	if len(words) == 0 {
		return false
	}

	if len(words) == 1 && !leadingNonLetters {
		return utf8.RuneCountInString(words[0]) > 1 && isAllLowerCase(words[0])
	}

	counted := 0
	if leadingNonLetters {
		counted++
	}

	mixed := false
	for _, word := range words {
		if utf8.RuneCountInString(word) == 1 || !hasCasedLetter(word) {
			continue
		}
		counted++
		if hasUpperCase(word) && hasLowerCase(word) {
			mixed = true
		}
	}

	return counted >= 2 && !mixed
}

// caseWords splits a name into its runs of letters, and reports separately
// whether the name opens with something that is not a letter. Canonical splits
// on non-letters and inherits a leading empty field from Java's split when it
// does; the flag carries that field without putting an empty string in the
// slice for every caller to step over.
func caseWords(value string) (words []string, leadingNonLetters bool) {
	start := -1
	for i, r := range value {
		if unicode.IsLetter(r) {
			if start < 0 {
				start = i
			}
			continue
		}
		if start >= 0 {
			words = append(words, value[start:i])
			start = -1
		} else if i == 0 {
			leadingNonLetters = true
		}
	}
	if start >= 0 {
		words = append(words, value[start:])
	}
	return words, leadingNonLetters
}

func isAllLowerCase(word string) bool {
	for _, r := range word {
		if !unicode.IsLower(r) {
			return false
		}
	}
	return true
}

func hasCasedLetter(word string) bool {
	return hasUpperCase(word) || hasLowerCase(word)
}

func hasUpperCase(word string) bool {
	return strings.ContainsFunc(word, func(r rune) bool {
		return unicode.IsUpper(r) || unicode.IsTitle(r)
	})
}

func hasLowerCase(word string) bool {
	return strings.ContainsFunc(word, unicode.IsLower)
}
