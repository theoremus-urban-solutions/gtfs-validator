package entity

import (
	"io"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/parser"
	"github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

// PrimaryKeyValidator validates primary key uniqueness
type PrimaryKeyValidator struct{}

// NewPrimaryKeyValidator creates a new primary key validator
func NewPrimaryKeyValidator() *PrimaryKeyValidator {
	return &PrimaryKeyValidator{}
}

// Validate checks primary key uniqueness in all files
func (v *PrimaryKeyValidator) Validate(loader *parser.FeedLoader, container *notice.NoticeContainer, config validator.Config) {
	files := loader.ListFiles()

	for _, filename := range files {
		v.validateFile(loader, container, filename)
	}
}

// validateFile validates primary key uniqueness in a single file
func (v *PrimaryKeyValidator) validateFile(loader *parser.FeedLoader, container *notice.NoticeContainer, filename string) {
	reader, err := loader.GetFile(filename)
	if err != nil {
		return
	}
	defer reader.Close()

	csvFile, err := parser.NewCSVFile(reader, filename)
	if err != nil {
		return
	}

	// Get primary key fields for this file
	primaryKeyFields := v.getPrimaryKeyFields(filename)
	if len(primaryKeyFields) == 0 {
		return // No primary key defined for this file
	}

	// Track seen keys
	seenKeys := make(map[string]int) // key -> first row number

	// Read and validate each row
	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			return
		}

		// Build composite key
		key := v.buildCompositeKey(row, primaryKeyFields)

		// Check for duplicates
		if firstRow, exists := seenKeys[key]; exists {
			// Create appropriate notice based on key type
			if len(primaryKeyFields) == 1 {
				container.AddNotice(notice.NewDuplicateKeyNotice(
					filename,
					primaryKeyFields[0],
					row.Values[primaryKeyFields[0]],
					firstRow,
					row.RowNumber,
				))
			} else {
				// For composite keys, we'll use the first field in the notice
				container.AddNotice(notice.NewDuplicateKeyNotice(
					filename,
					primaryKeyFields[0],
					key, // Use the composite key as the value
					firstRow,
					row.RowNumber,
				))
			}
		} else {
			seenKeys[key] = row.RowNumber
		}
	}
}

// buildCompositeKey builds a composite key from multiple fields
func (v *PrimaryKeyValidator) buildCompositeKey(row *parser.CSVRow, fields []string) string {
	if len(fields) == 1 {
		return row.Values[fields[0]]
	}

	// Join multiple fields with a delimiter
	key := ""
	for i, field := range fields {
		if i > 0 {
			key += "|"
		}
		key += row.Values[field]
	}
	return key
}

// getPrimaryKeyFields returns the primary key fields for a given file
func (v *PrimaryKeyValidator) getPrimaryKeyFields(filename string) []string {
	switch filename {
	case "agency.txt":
		return []string{"agency_id"}
	case "stops.txt":
		return []string{"stop_id"}
	case "routes.txt":
		return []string{"route_id"}
	case "trips.txt":
		return []string{"trip_id"}
	case "stop_times.txt":
		return []string{"trip_id", "stop_sequence"}
	case "calendar.txt":
		return []string{"service_id"}
	case "calendar_dates.txt":
		return []string{"service_id", "date"}
	case "fare_attributes.txt":
		return []string{"fare_id"}
	case "fare_rules.txt":
		// fare_rules has no single primary key, but combinations should be unique
		return []string{"fare_id", "route_id", "origin_id", "destination_id", "contains_id"}
	case "shapes.txt":
		return []string{"shape_id", "shape_pt_sequence"}
	case "frequencies.txt":
		return []string{"trip_id", "start_time"}
	case "transfers.txt":
		return []string{"from_stop_id", "to_stop_id", "from_trip_id", "to_trip_id"}
	case "pathways.txt":
		return []string{"pathway_id"}
	case "levels.txt":
		return []string{"level_id"}
	case "attributions.txt":
		return []string{"attribution_id"}
	default:
		return []string{}
	}
}
