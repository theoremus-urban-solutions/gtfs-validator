package core

import (
	"io"
	"log"
	"strconv"
	"strings"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/parser"
	"github.com/theoremus-urban-solutions/gtfs-validator/validator"
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
				// Handle special cases for stops.txt
				if filename == "stops.txt" && field == "stop_name" {
					// Check location type - some stop types don't require names
					if v.isStopNameOptionalForLocationType(row.Values) {
						// Create a warning instead of error
						container.AddNotice(notice.NewMissingRecommendedFieldNotice(
							filename,
							field,
							row.RowNumber,
						))
						continue
					}
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
	case "agency.txt":
		return []string{"agency_name", "agency_url", "agency_timezone"}
	case "stops.txt":
		return []string{"stop_id", "stop_name"}
	case "routes.txt":
		return []string{"route_id", "route_type"}
	case "trips.txt":
		return []string{"route_id", "service_id", "trip_id"}
	case "stop_times.txt":
		return []string{"trip_id", "stop_id", "stop_sequence"}
	case "calendar.txt":
		return []string{"service_id", "monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday", "start_date", "end_date"}
	case "calendar_dates.txt":
		return []string{"service_id", "date", "exception_type"}
	case "fare_attributes.txt":
		return []string{"fare_id", "price", "currency_type"}
	case "shapes.txt":
		return []string{"shape_id", "shape_pt_lat", "shape_pt_lon", "shape_pt_sequence"}
	case "frequencies.txt":
		return []string{"trip_id", "start_time", "end_time", "headway_secs"}
	case "transfers.txt":
		return []string{"from_stop_id", "to_stop_id", "transfer_type"}
	case "pathways.txt":
		return []string{"pathway_id", "from_stop_id", "to_stop_id", "pathway_mode", "is_bidirectional"}
	case "levels.txt":
		return []string{"level_id", "level_index"}
	case "feed_info.txt":
		return []string{"feed_publisher_name", "feed_publisher_url", "feed_lang"}
	default:
		return []string{}
	}
}

// isStopNameOptionalForLocationType checks if stop_name is optional for certain location types
func (v *RequiredFieldValidator) isStopNameOptionalForLocationType(rowValues map[string]string) bool {
	locationTypeStr, exists := rowValues["location_type"]
	if !exists {
		return false // Default location type 0 requires stop_name
	}

	locationType, err := strconv.Atoi(locationTypeStr)
	if err != nil {
		return false
	}

	switch locationType {
	case 3: // Generic node - stop_name is optional per GTFS spec
		return true
	case 4: // Boarding area - stop_name is conditional (can inherit from parent)
		// If it has a parent station, name is optional
		parentStation, hasParent := rowValues["parent_station"]
		return hasParent && strings.TrimSpace(parentStation) != ""
	default:
		return false
	}
}
