package entity

import (
	"io"
	"log"
	"strings"
	"unicode"

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

// needsMixedCase reports whether a customer-facing name is written in a single
// case. The canonical rule asks these fields for Mixed Case because an
// all-caps name is read out as an initialism by screen readers, and neither
// case can be recovered by a consumer once it is gone.
//
// Only cased letters count. Chinese, Japanese, Arabic and Hebrew have no case
// distinction to make, and a name too short to carry one — a single-letter
// route, say — says nothing about the producer's intent, so both are left
// alone.
func needsMixedCase(value string) bool {
	upper, lower := 0, 0
	for _, r := range value {
		switch {
		case unicode.IsUpper(r), unicode.IsTitle(r):
			upper++
		case unicode.IsLower(r):
			lower++
		}
	}

	if upper+lower < 2 {
		return false
	}

	return upper == 0 || lower == 0
}
