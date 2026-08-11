package core

import (
	"io"
	"log"
	"strings"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/parser"
	"github.com/theoremus-urban-solutions/gtfs-validator/schema"
	"github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

// GTFS filename constants
const (
	AgencyFile         = "agency.txt"
	RoutesFile         = "routes.txt"
	TripsFile          = "trips.txt"
	StopTimesFile      = "stop_times.txt"
	StopsFile          = "stops.txt"
	CalendarFile       = "calendar.txt"
	CalendarDatesFile  = "calendar_dates.txt"
	FareAttributesFile = "fare_attributes.txt"
	ShapesFile         = "shapes.txt"
	FrequenciesFile    = "frequencies.txt"
	TransfersFile      = "transfers.txt"
	FeedInfoFile       = "feed_info.txt"
)

// RequiredFieldValidator validates required fields in GTFS files
type RequiredFieldValidator struct{}

// NewRequiredFieldValidator creates a new required field validator
func NewRequiredFieldValidator() *RequiredFieldValidator {
	return &RequiredFieldValidator{}
}

// Validate checks that all required fields are present and non-empty
func (v *RequiredFieldValidator) Validate(loader *parser.FeedLoader, container *notice.NoticeContainer, config validator.Config) {
	files := loader.ListFiles()

	for _, filename := range files {
		v.validateFile(loader, container, filename)
	}
}

// validateFile validates required fields in a single file
func (v *RequiredFieldValidator) validateFile(loader *parser.FeedLoader, container *notice.NoticeContainer, filename string) {
	reader, err := loader.GetFile(filename)
	if err != nil {
		return
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close reader %v", closeErr)
		}
	}()

	csvFile, err := parser.NewCSVFile(reader, filename)
	if err != nil {
		return
	}

	// The spec's own Presence column decides these, rather than a list kept
	// by hand. Conditionally Required fields are deliberately absent: whether
	// they apply depends on the rest of the feed, so each belongs to a rule
	// that knows the condition — stop_name to missing_stop_name, agency_id to
	// missing_required_agency_id, and so on.
	requiredFields, known := schema.FieldsWithPresence(filename, schema.PresenceRequired)
	if !known {
		return
	}
	recommendedFields, _ := schema.FieldsWithPresence(filename, schema.PresenceRecommended)

	// A required field whose column is not in the file at all is one defect,
	// reported once as missing_required_column. Reporting it again per row says
	// nothing new and buries the real finding: removing stop_id from a 177-row
	// stops.txt produced 177 missing_required_field errors on top of the single
	// column error that explains them.
	present := make(map[string]bool, len(csvFile.Headers))
	for _, header := range csvFile.Headers {
		present[strings.TrimSpace(header)] = true
	}

	// Read and validate each row
	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			return
		}

		for _, field := range requiredFields {
			// A few Required fields count a blank as one of their values —
			// fare_attributes.transfers empty means unlimited. Required there
			// means the column must exist, not that every row must fill it.
			if schema.EmptyIsMeaningful(filename, field) {
				continue
			}
			if !present[field] {
				continue
			}
			if isBlank(row.Values, field) {
				container.AddNotice(notice.NewMissingRequiredFieldNotice(
					filename,
					field,
					row.RowNumber,
				))
			}
		}

		for _, field := range recommendedFields {
			if isBlank(row.Values, field) {
				container.AddNotice(notice.NewMissingRecommendedFieldNotice(
					filename,
					field,
					row.RowNumber,
				))
			}
		}
	}
}

// isBlank reports whether a row leaves a field empty, either by omitting the
// column or by carrying nothing in it.
func isBlank(values map[string]string, field string) bool {
	value, present := values[field]
	return !present || strings.TrimSpace(value) == ""
}
