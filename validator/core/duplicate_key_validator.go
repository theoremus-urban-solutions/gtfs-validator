package core

import (
	"io"
	"log"
	"strings"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/parser"
	"github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

// DuplicateKeyValidator checks for duplicate primary keys in GTFS files
type DuplicateKeyValidator struct{}

// NewDuplicateKeyValidator creates a new duplicate key validator
func NewDuplicateKeyValidator() *DuplicateKeyValidator {
	return &DuplicateKeyValidator{}
}

// FileKeyConfig defines primary key fields for each GTFS file
type FileKeyConfig struct {
	Filename    string
	KeyFields   []string
	IsComposite bool
}

// Validate checks for duplicate primary keys across all GTFS files
func (v *DuplicateKeyValidator) Validate(loader *parser.FeedLoader, container *notice.NoticeContainer, config validator.Config) {
	// Define primary key configurations for each GTFS file
	fileConfigs := []FileKeyConfig{
		{"agency.txt", []string{"agency_id"}, false},
		{"stops.txt", []string{"stop_id"}, false},
		{"routes.txt", []string{"route_id"}, false},
		{"trips.txt", []string{"trip_id"}, false},
		{"stop_times.txt", []string{"trip_id", "stop_sequence"}, true},
		{"calendar.txt", []string{"service_id"}, false},
		{"calendar_dates.txt", []string{"service_id", "date"}, true},
		{"fare_attributes.txt", []string{"fare_id"}, false},
		{"fare_rules.txt", []string{"route_id", "origin_id", "destination_id", "contains_id"}, true},
		{"shapes.txt", []string{"shape_id", "shape_pt_sequence"}, true},
		{"feed_info.txt", []string{}, false}, // No primary key - only one record allowed
		{"frequencies.txt", []string{"trip_id", "start_time"}, true},
		{"transfers.txt", []string{"from_stop_id", "to_stop_id"}, true},
		{"pathways.txt", []string{"pathway_id"}, false},
		{"levels.txt", []string{"level_id"}, false},
		{"attributions.txt", []string{}, false}, // No explicit primary key
	}

	for _, fileConfig := range fileConfigs {
		v.validateFileKeys(loader, container, fileConfig)
	}
}

// validateFileKeys validates primary keys for a specific file
func (v *DuplicateKeyValidator) validateFileKeys(loader *parser.FeedLoader, container *notice.NoticeContainer, config FileKeyConfig) {
	if !loader.HasFile(config.Filename) {
		return // File doesn't exist, skip validation
	}

	reader, err := loader.GetFile(config.Filename)
	if err != nil {
		return
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close reader %v", closeErr)
		}
	}()

	csvFile, err := parser.NewCSVFile(reader, config.Filename)
	if err != nil {
		return
	}

	// Special case for feed_info.txt - only one record allowed
	if config.Filename == "feed_info.txt" {
		v.validateSingleRecordFile(container, csvFile, config.Filename)
		return
	}

	// Skip if no key fields defined
	if len(config.KeyFields) == 0 {
		return
	}

	keyMap := make(map[string]int) // key -> first occurrence row number

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}

		key := v.buildKey(row, config.KeyFields)
		if key == "" {
			continue // Skip rows with missing key components
		}

		if firstRowNumber, exists := keyMap[key]; exists {
			// Duplicate key found
			if config.IsComposite {
				container.AddNotice(notice.NewDuplicateCompositeKeyNotice(
					config.Filename,
					strings.Join(config.KeyFields, "+"),
					key,
					firstRowNumber,
					row.RowNumber,
				))
			} else {
				container.AddNotice(notice.NewDuplicateKeyNotice(
					config.Filename,
					config.KeyFields[0],
					key,
					firstRowNumber,
					row.RowNumber,
				))
			}
		} else {
			keyMap[key] = row.RowNumber
		}
	}
}

// buildKey creates a composite key string from the specified fields
func (v *DuplicateKeyValidator) buildKey(row *parser.CSVRow, keyFields []string) string {
	var keyParts []string

	for _, field := range keyFields {
		if value, exists := row.Values[field]; exists {
			trimmedValue := strings.TrimSpace(value)
			if trimmedValue == "" {
				return "" // Missing key component
			}
			keyParts = append(keyParts, trimmedValue)
		} else {
			return "" // Missing key field
		}
	}

	return strings.Join(keyParts, "|")
}

// validateSingleRecordFile validates files that should contain only one record
func (v *DuplicateKeyValidator) validateSingleRecordFile(container *notice.NoticeContainer, csvFile *parser.CSVFile, filename string) {
	rowCount := 0

	for {
		_, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}
		rowCount++
	}

	if rowCount > 1 {
		container.AddNotice(notice.NewMultipleRecordsInSingleRecordFileNotice(
			filename,
			rowCount,
		))
	}
}
