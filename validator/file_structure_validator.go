package validator

import (
	"io"
	"log"
	"strings"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/parser"
)

// FileStructureValidator validates the structure of GTFS files
type FileStructureValidator struct{}

// NewFileStructureValidator creates a new file structure validator
func NewFileStructureValidator() *FileStructureValidator {
	return &FileStructureValidator{}
}

// Validate checks the structure of all GTFS files
func (v *FileStructureValidator) Validate(loader *parser.FeedLoader, container *notice.NoticeContainer, config Config) {
	files := loader.ListFiles()

	for _, filename := range files {
		v.validateFile(loader, container, filename)
	}
}

// validateFile validates a single file's structure
func (v *FileStructureValidator) validateFile(loader *parser.FeedLoader, container *notice.NoticeContainer, filename string) {
	reader, err := loader.GetFile(filename)
	if err != nil {
		// This shouldn't happen as we're iterating over existing files
		return
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close reader %v", closeErr)
		}
	}()

	csvFile, err := parser.NewCSVFile(reader, filename)
	if err != nil {
		if strings.Contains(err.Error(), "empty file") {
			container.AddNotice(notice.NewEmptyFileNotice(filename))
		} else {
			// Add a CSV parsing error notice
			container.AddNotice(notice.NewBaseNotice("csv_parsing_failed", notice.ERROR, map[string]interface{}{
				"filename": filename,
				"error":    err.Error(),
			}))
		}
		return
	}

	// Check for empty file (no data rows)
	err = csvFile.ReadAll()
	if err != nil && err != io.EOF {
		container.AddNotice(notice.NewBaseNotice("csv_parsing_failed", notice.ERROR, map[string]interface{}{
			"filename": filename,
			"error":    err.Error(),
		}))
		return
	}

	if csvFile.IsEmpty() {
		container.AddNotice(notice.NewEmptyFileNotice(filename))
		return
	}

	// Check for unknown columns based on file type
	v.checkUnknownColumns(csvFile, container)
}

// checkUnknownColumns checks for columns not defined in GTFS spec
func (v *FileStructureValidator) checkUnknownColumns(csvFile *parser.CSVFile, container *notice.NoticeContainer) {
	// Define known columns for each file type
	knownColumns := v.getKnownColumns(csvFile.Filename)

	for i, header := range csvFile.Headers {
		if !contains(knownColumns, header) {
			container.AddNotice(notice.NewUnknownColumnNotice(
				csvFile.Filename,
				header,
				i,
			))
		}
	}
}

// getKnownColumns returns the known columns for a given file
func (v *FileStructureValidator) getKnownColumns(filename string) []string {
	// This is a simplified version - in a real implementation,
	// this would be based on the GTFS specification
	switch filename {
	case "agency.txt":
		return []string{
			"agency_id", "agency_name", "agency_url", "agency_timezone",
			"agency_lang", "agency_phone", "agency_fare_url", "agency_email",
		}
	case "stops.txt":
		return []string{
			"stop_id", "stop_code", "stop_name", "stop_desc", "stop_lat", "stop_lon",
			"zone_id", "stop_url", "location_type", "parent_station", "stop_timezone",
			"wheelchair_boarding", "level_id", "platform_code",
		}
	case "routes.txt":
		return []string{
			"route_id", "agency_id", "route_short_name", "route_long_name", "route_desc",
			"route_type", "route_url", "route_color", "route_text_color", "route_sort_order",
			"continuous_pickup", "continuous_drop_off",
		}
	case "trips.txt":
		return []string{
			"route_id", "service_id", "trip_id", "trip_headsign", "trip_short_name",
			"direction_id", "block_id", "shape_id", "wheelchair_accessible", "bikes_allowed",
		}
	case "stop_times.txt":
		return []string{
			"trip_id", "arrival_time", "departure_time", "stop_id", "stop_sequence",
			"stop_headsign", "pickup_type", "drop_off_type", "continuous_pickup",
			"continuous_drop_off", "shape_dist_traveled", "timepoint",
		}
	case "calendar_dates.txt":
		return []string{
			"service_id", "date", "exception_type",
		}
	case "feed_info.txt":
		return []string{
			"feed_publisher_name", "feed_publisher_url", "feed_lang", "default_lang",
			"feed_start_date", "feed_end_date", "feed_version", "feed_contact_email",
			"feed_contact_url",
		}
	case "fare_attributes.txt":
		return []string{
			"fare_id", "price", "currency_type", "payment_method", "transfers",
			"agency_id", "transfer_duration",
		}
	case "shapes.txt":
		return []string{
			"shape_id", "shape_pt_lat", "shape_pt_lon", "shape_pt_sequence",
			"shape_dist_traveled",
		}
	case "levels.txt":
		return []string{
			"level_id", "level_index", "level_name",
		}
	case "pathways.txt":
		return []string{
			"pathway_id", "from_stop_id", "to_stop_id", "pathway_mode",
			"is_bidirectional", "length", "traversal_time", "stair_count",
			"max_slope", "min_width", "signposted_as", "reversed_signposted_as",
		}
	case "transfers.txt":
		return []string{
			"from_stop_id", "to_stop_id", "from_route_id", "to_route_id",
			"from_trip_id", "to_trip_id", "transfer_type", "min_transfer_time",
		}
	case "translations.txt":
		return []string{
			"table_name", "field_name", "language", "translation", "record_id",
			"record_sub_id", "field_value",
		}
	default:
		// Return empty for unknown files - they won't generate unknown column notices
		return []string{}
	}
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
