package core

import (
	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/parser"
	"github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

// MissingColumnValidator validates that required columns are present
type MissingColumnValidator struct{}

// NewMissingColumnValidator creates a new missing column validator
func NewMissingColumnValidator() *MissingColumnValidator {
	return &MissingColumnValidator{}
}

// fileRequiredColumns defines required columns for each GTFS file
var fileRequiredColumns = map[string][]string{
	"agency.txt": {
		"agency_name",
		"agency_url",
		"agency_timezone",
	},
	"stops.txt": {
		"stop_id",
	},
	"routes.txt": {
		"route_id",
		"route_type",
	},
	"trips.txt": {
		"route_id",
		"service_id",
		"trip_id",
	},
	"stop_times.txt": {
		"trip_id",
		"stop_id",
		"stop_sequence",
	},
	"calendar.txt": {
		"service_id",
		"monday",
		"tuesday",
		"wednesday",
		"thursday",
		"friday",
		"saturday",
		"sunday",
		"start_date",
		"end_date",
	},
	"calendar_dates.txt": {
		"service_id",
		"date",
		"exception_type",
	},
	"fare_attributes.txt": {
		"fare_id",
		"price",
		"currency_type",
	},
	"fare_rules.txt": {
		"fare_id",
	},
	"shapes.txt": {
		"shape_id",
		"shape_pt_lat",
		"shape_pt_lon",
		"shape_pt_sequence",
	},
	"frequencies.txt": {
		"trip_id",
		"start_time",
		"end_time",
		"headway_secs",
	},
	"transfers.txt": {
		"from_stop_id",
		"to_stop_id",
		"transfer_type",
	},
	"pathways.txt": {
		"pathway_id",
		"from_stop_id",
		"to_stop_id",
		"pathway_mode",
		"is_bidirectional",
	},
	"levels.txt": {
		"level_id",
		"level_index",
	},
	"feed_info.txt": {
		"feed_publisher_name",
		"feed_publisher_url",
		"feed_lang",
	},
}

// Validate checks that required columns are present in GTFS files
func (v *MissingColumnValidator) Validate(loader *parser.FeedLoader, container *notice.NoticeContainer, config validator.Config) {
	files := loader.ListFiles()

	for _, filename := range files {
		v.validateFileColumns(loader, container, filename)
	}
}

// validateFileColumns checks required columns for a single file
func (v *MissingColumnValidator) validateFileColumns(loader *parser.FeedLoader, container *notice.NoticeContainer, filename string) {
	requiredColumns, hasRequiredColumns := fileRequiredColumns[filename]
	if !hasRequiredColumns {
		return // No required columns defined for this file
	}

	reader, err := loader.GetFile(filename)
	if err != nil {
		return // File doesn't exist, other validators handle this
	}
	defer reader.Close()

	csvFile, err := parser.NewCSVFile(reader, filename)
	if err != nil {
		return // File format issues, other validators handle this
	}

	// Create a set of existing headers for fast lookup
	existingHeaders := make(map[string]bool)
	for _, header := range csvFile.Headers {
		existingHeaders[header] = true
	}

	// Check for missing required columns
	for _, requiredColumn := range requiredColumns {
		if !existingHeaders[requiredColumn] {
			container.AddNotice(notice.NewMissingRequiredColumnNotice(
				filename,
				requiredColumn,
			))
		}
	}
}
