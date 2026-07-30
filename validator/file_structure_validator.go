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
	knownColumns, isSpecFile := gtfsFileColumns[csvFile.Filename]
	if !isSpecFile {
		// Nothing is known about this file's columns, so every one of them
		// would be reported. The file itself is already reported once as
		// unknown_file, which is the finding; repeating it per column is not.
		return
	}

	for i, header := range csvFile.Headers {
		// Surrounding whitespace is a defect of the header itself and is
		// reported as such elsewhere. Matching on the trimmed name stops a
		// merely untidy column from also being read as an unknown one.
		if !contains(knownColumns, strings.TrimSpace(header)) {
			container.AddNotice(notice.NewUnknownColumnNotice(
				csvFile.Filename,
				header,
				i,
			))
		}
	}
}

// gtfsFileColumns lists every column the GTFS Schedule reference defines for
// each file, including the optional and conditionally required ones.
//
// A file absent from this map is not checked at all. The list has to be
// exhaustive to be usable: a single missing field turns an ordinary feed into
// a run of unknown_column notices, one per row of noise for a column that is
// perfectly valid. Silence on a file we do not know is the cheaper error.
var gtfsFileColumns = map[string][]string{
	"agency.txt": {
		"agency_id", "agency_name", "agency_url", "agency_timezone",
		"agency_lang", "agency_phone", "agency_fare_url", "agency_email",
		"cemv_support",
	},
	"stops.txt": {
		"stop_id", "stop_code", "stop_name", "tts_stop_name", "stop_desc",
		"stop_lat", "stop_lon", "zone_id", "stop_url", "location_type",
		"parent_station", "stop_timezone", "wheelchair_boarding",
		"level_id", "platform_code", "stop_access",
	},
	"routes.txt": {
		"route_id", "agency_id", "route_short_name", "route_long_name",
		"route_desc", "route_type", "route_url", "route_color",
		"route_text_color", "route_sort_order", "continuous_pickup",
		"continuous_drop_off", "network_id", "cemv_support",
	},
	"trips.txt": {
		"route_id", "service_id", "trip_id", "trip_headsign",
		"trip_short_name", "direction_id", "block_id", "shape_id",
		"wheelchair_accessible", "bikes_allowed", "cars_allowed",
		"safe_duration_factor", "safe_duration_offset",
	},
	"stop_times.txt": {
		"trip_id", "arrival_time", "departure_time", "stop_id",
		"location_group_id", "location_id", "stop_sequence",
		"stop_headsign", "start_pickup_drop_off_window",
		"end_pickup_drop_off_window", "pickup_type", "drop_off_type",
		"continuous_pickup", "continuous_drop_off", "shape_dist_traveled",
		"timepoint", "pickup_booking_rule_id", "drop_off_booking_rule_id",
	},
	"calendar.txt": {
		"service_id", "monday", "tuesday", "wednesday", "thursday",
		"friday", "saturday", "sunday", "start_date", "end_date",
	},
	"calendar_dates.txt": {
		"service_id", "date", "exception_type",
	},
	"fare_attributes.txt": {
		"fare_id", "price", "currency_type", "payment_method", "transfers",
		"agency_id", "transfer_duration",
	},
	"fare_rules.txt": {
		"fare_id", "route_id", "origin_id", "destination_id", "contains_id",
	},
	"timeframes.txt": {
		"timeframe_group_id", "start_time", "end_time", "service_id",
	},
	"rider_categories.txt": {
		"rider_category_id", "rider_category_name",
		"is_default_fare_category", "eligibility_url",
	},
	"fare_media.txt": {
		"fare_media_id", "fare_media_name", "fare_media_type",
	},
	"fare_products.txt": {
		"fare_product_id", "fare_product_name", "rider_category_id",
		"fare_media_id", "amount", "currency",
	},
	"fare_leg_rules.txt": {
		"leg_group_id", "network_id", "from_area_id", "to_area_id",
		"from_timeframe_group_id", "to_timeframe_group_id",
		"fare_product_id", "rule_priority",
	},
	"fare_leg_join_rules.txt": {
		"from_network_id", "to_network_id", "from_stop_id", "to_stop_id",
	},
	"fare_transfer_rules.txt": {
		"from_leg_group_id", "to_leg_group_id", "transfer_count",
		"duration_limit", "duration_limit_type", "fare_transfer_type",
		"fare_product_id",
	},
	"areas.txt": {
		"area_id", "area_name",
	},
	"stop_areas.txt": {
		"area_id", "stop_id",
	},
	"networks.txt": {
		"network_id", "network_name",
	},
	"route_networks.txt": {
		"network_id", "route_id",
	},
	"shapes.txt": {
		"shape_id", "shape_pt_lat", "shape_pt_lon", "shape_pt_sequence",
		"shape_dist_traveled",
	},
	"frequencies.txt": {
		"trip_id", "start_time", "end_time", "headway_secs", "exact_times",
	},
	"transfers.txt": {
		"from_stop_id", "to_stop_id", "from_route_id", "to_route_id",
		"from_trip_id", "to_trip_id", "transfer_type", "min_transfer_time",
	},
	"pathways.txt": {
		"pathway_id", "from_stop_id", "to_stop_id", "pathway_mode",
		"is_bidirectional", "length", "traversal_time", "stair_count",
		"max_slope", "min_width", "signposted_as", "reversed_signposted_as",
	},
	"levels.txt": {
		"level_id", "level_index", "level_name",
	},
	"location_groups.txt": {
		"location_group_id", "location_group_name",
	},
	"location_group_stops.txt": {
		"location_group_id", "stop_id",
	},
	"booking_rules.txt": {
		"booking_rule_id", "booking_type", "prior_notice_duration_min",
		"prior_notice_duration_max", "prior_notice_last_day",
		"prior_notice_last_time", "prior_notice_start_day",
		"prior_notice_start_time", "prior_notice_service_id", "message",
		"pickup_message", "drop_off_message", "phone_number", "info_url",
		"booking_url",
	},
	"translations.txt": {
		"table_name", "field_name", "language", "translation", "record_id",
		"record_sub_id", "field_value",
	},
	"feed_info.txt": {
		"feed_publisher_name", "feed_publisher_url", "feed_lang",
		"default_lang", "feed_start_date", "feed_end_date", "feed_version",
		"feed_contact_email", "feed_contact_url",
	},
	"attributions.txt": {
		"attribution_id", "agency_id", "route_id", "trip_id",
		"organization_name", "is_producer", "is_operator", "is_authority",
		"attribution_url", "attribution_email", "attribution_phone",
	},
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
