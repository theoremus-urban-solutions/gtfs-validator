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

// BikeAllowanceValidator reports ferry trips that do not say whether bikes may
// be carried.
//
// bikes_allowed is optional everywhere in GTFS, but on a ferry the answer
// decides whether a cycling passenger can board at all, so an unspecified value
// there is a gap riders feel.
type BikeAllowanceValidator struct{}

// NewBikeAllowanceValidator creates a new bike allowance validator
func NewBikeAllowanceValidator() *BikeAllowanceValidator {
	return &BikeAllowanceValidator{}
}

// ferryRouteType is the route_type for ferry service.
const ferryRouteType = 4

// Validate checks that every trip on a ferry route declares bikes_allowed.
func (v *BikeAllowanceValidator) Validate(loader *parser.FeedLoader, container *notice.NoticeContainer, config validator.Config) {
	ferryRoutes := v.loadFerryRoutes(loader)
	if len(ferryRoutes) == 0 {
		return
	}

	reader, err := loader.GetFile("trips.txt")
	if err != nil {
		return
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close reader %v", closeErr)
		}
	}()

	csvFile, err := parser.NewCSVFile(reader, "trips.txt")
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

		routeID := strings.TrimSpace(row.Values["route_id"])
		if !ferryRoutes[routeID] {
			continue
		}

		if v.declaresBikeAllowance(row.Values["bikes_allowed"]) {
			continue
		}
		container.AddNotice(notice.NewMissingBikeAllowanceNotice(
			row.RowNumber,
			routeID,
			strings.TrimSpace(row.Values["trip_id"]),
		))
	}
}

// declaresBikeAllowance reports whether bikes_allowed states an answer. Empty
// and 0 both mean "no information"; a value that is neither a number nor an
// answer is reported by the type layer, so it is not counted twice here.
func (v *BikeAllowanceValidator) declaresBikeAllowance(value string) bool {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return false
	}
	parsed, err := strconv.Atoi(trimmed)
	if err != nil {
		return true
	}
	return parsed == 1 || parsed == 2
}

// loadFerryRoutes returns the set of route_ids running ferry service.
func (v *BikeAllowanceValidator) loadFerryRoutes(loader *parser.FeedLoader) map[string]bool {
	ferryRoutes := make(map[string]bool)

	reader, err := loader.GetFile("routes.txt")
	if err != nil {
		return ferryRoutes
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close reader %v", closeErr)
		}
	}()

	csvFile, err := parser.NewCSVFile(reader, "routes.txt")
	if err != nil {
		return ferryRoutes
	}

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		routeType, err := strconv.Atoi(strings.TrimSpace(row.Values["route_type"]))
		if err != nil || routeType != ferryRouteType {
			continue
		}
		routeID := strings.TrimSpace(row.Values["route_id"])
		if routeID != "" {
			ferryRoutes[routeID] = true
		}
	}

	return ferryRoutes
}
