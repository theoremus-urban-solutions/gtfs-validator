package core

import (
	"io"
	"log"
	"strings"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/parser"
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

	// Get required fields for this file
	requiredFields := v.getRequiredFields(filename)

	// Read and validate each row
	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			return
		}

		// Check each required field
		for _, field := range requiredFields {
			value, exists := row.Values[field]
			if !exists || strings.TrimSpace(value) == "" {
				// stop_name is conditionally required: a stop, station or
				// entrance must be named, and a generic node or boarding area
				// need not be. For the types that must have one, the omission
				// is reported as missing_stop_name by
				// entity/stop_name_validator.go; for the rest the spec calls
				// the field optional, not recommended, so there is nothing to
				// say. A station's nodes are unnamed by design and a large
				// station complex has dozens of them.
				if filename == StopsFile && field == "stop_name" {
					continue
				}

				container.AddNotice(notice.NewMissingRequiredFieldNotice(
					filename,
					field,
					row.RowNumber,
				))
			}
		}
	}
}

// getRequiredFields returns the required fields for a given file
func (v *RequiredFieldValidator) getRequiredFields(filename string) []string {
	switch filename {
	case AgencyFile:
		return []string{"agency_name", "agency_url", "agency_timezone"}
	case StopsFile:
		return []string{"stop_id", "stop_name"}
	case RoutesFile:
		return []string{"route_id", "route_type"}
	case TripsFile:
		return []string{"route_id", "service_id", "trip_id"}
	case StopTimesFile:
		return []string{"trip_id", "stop_id", "stop_sequence"}
	case CalendarFile:
		return []string{"service_id", "monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday", "start_date", "end_date"}
	case CalendarDatesFile:
		return []string{"service_id", "date", "exception_type"}
	case FareAttributesFile:
		return []string{"fare_id", "price", "currency_type"}
	case ShapesFile:
		return []string{"shape_id", "shape_pt_lat", "shape_pt_lon", "shape_pt_sequence"}
	case FrequenciesFile:
		return []string{"trip_id", "start_time", "end_time", "headway_secs"}
	case TransfersFile:
		return []string{"from_stop_id", "to_stop_id", "transfer_type"}
	case "pathways.txt":
		return []string{"pathway_id", "from_stop_id", "to_stop_id", "pathway_mode", "is_bidirectional"}
	case "levels.txt":
		return []string{"level_id", "level_index"}
	case FeedInfoFile:
		return []string{"feed_publisher_name", "feed_publisher_url", "feed_lang"}
	default:
		return []string{}
	}
}
