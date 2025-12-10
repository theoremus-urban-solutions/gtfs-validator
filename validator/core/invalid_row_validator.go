package core

import (
	"io"
	"log"
	"strconv"
	"strings"
	"sync"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/parser"
	"github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

// InvalidRowValidator checks for rows with invalid structure or data
type InvalidRowValidator struct{}

// NewInvalidRowValidator creates a new invalid row validator
func NewInvalidRowValidator() *InvalidRowValidator {
	return &InvalidRowValidator{}
}

// Validate checks for invalid rows across GTFS files
func (v *InvalidRowValidator) Validate(loader *parser.FeedLoader, container *notice.NoticeContainer, config validator.Config) {
	// Get list of all GTFS files to check
	gtfsFiles := []string{
		"agency.txt", "stops.txt", "routes.txt", "trips.txt", "stop_times.txt",
		"calendar.txt", "calendar_dates.txt", "fare_attributes.txt",
		"fare_rules.txt", "shapes.txt", "frequencies.txt", "transfers.txt",
		"pathways.txt", "levels.txt", "feed_info.txt", "attributions.txt",
	}

	// Filter to files that exist
	existingFiles := make([]string, 0, len(gtfsFiles))
	for _, filename := range gtfsFiles {
		if loader.HasFile(filename) {
			existingFiles = append(existingFiles, filename)
		}
	}

	// Use parallel validation if configured (Phase 2 optimization)
	workers := config.ParallelWorkers
	if workers > 1 && len(existingFiles) >= 4 {
		v.validateFilesParallel(loader, container, existingFiles, workers)
	} else {
		// Sequential validation for small number of files or single worker
		for _, filename := range existingFiles {
			v.validateFile(loader, container, filename)
		}
	}
}

// validateFilesParallel validates multiple files in parallel using a worker pool.
// This is invoked when config.ParallelWorkers > 1 and there are enough files to benefit from parallelization.
func (v *InvalidRowValidator) validateFilesParallel(
	loader *parser.FeedLoader,
	container *notice.NoticeContainer,
	files []string,
	workers int,
) {
	// Create work channel
	fileChan := make(chan string, len(files))
	var wg sync.WaitGroup

	// Start worker goroutines
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for filename := range fileChan {
				v.validateFile(loader, container, filename)
			}
		}()
	}

	// Distribute work
	for _, filename := range files {
		fileChan <- filename
	}
	close(fileChan)

	// Wait for all workers to complete
	wg.Wait()
}

// validateFile validates a specific GTFS file for invalid rows
func (v *InvalidRowValidator) validateFile(loader *parser.FeedLoader, container *notice.NoticeContainer, filename string) {
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

	headerCount := len(csvFile.Headers)

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			// CSV parsing error - check if it's a wrong number of fields error
			if strings.Contains(err.Error(), "wrong number of fields") {
				// Extract field count information if possible
				// For now, we'll generate a wrong number of fields notice
				container.AddNotice(notice.NewWrongNumberOfFieldsNotice(
					filename,
					headerCount+1, // approximate first data row
					headerCount,
					0, // unknown actual count
				))
			} else {
				// Other CSV parsing errors - treat as generic invalid row
				container.AddNotice(notice.NewInvalidRowNotice(
					filename,
					headerCount+1, // approximate first data row
					"CSV parsing error: "+err.Error(),
				))
			}
			continue
		}

		// Check for rows with wrong number of fields using raw field count if available
		fieldCount := row.RawFieldCount
		if fieldCount != headerCount {
			container.AddNotice(notice.NewWrongNumberOfFieldsNotice(
				filename,
				row.RowNumber,
				headerCount,
				fieldCount,
			))
		}

		// Check for specific invalid patterns based on file type
		v.validateRowContent(container, filename, row)
	}
}

// validateRowContent validates the content of a row based on file-specific rules
func (v *InvalidRowValidator) validateRowContent(container *notice.NoticeContainer, filename string, row *parser.CSVRow) {
	switch filename {
	case "stop_times.txt":
		v.validateStopTimeRow(container, row)
	case StopsFile:
		v.validateStopRow(container, row)
	case RoutesFile:
		v.validateRouteRow(container, row)
	case TripsFile:
		v.validateTripRow(container, row)
	case CalendarFile:
		v.validateCalendarRow(container, row)
	case CalendarDatesFile:
		v.validateCalendarDateRow(container, row)
	case "shapes.txt":
		v.validateShapeRow(container, row)
	case "fare_attributes.txt":
		v.validateFareAttributeRow(container, row)
	case "frequencies.txt":
		v.validateFrequencyRow(container, row)
	case "transfers.txt":
		v.validateTransferRow(container, row)
	}
}

// validateStopTimeRow validates stop_times.txt specific patterns
func (v *InvalidRowValidator) validateStopTimeRow(container *notice.NoticeContainer, row *parser.CSVRow) {
	// Check for negative stop_sequence
	if stopSeqStr, exists := row.Values["stop_sequence"]; exists {
		if stopSeq, err := strconv.Atoi(strings.TrimSpace(stopSeqStr)); err == nil {
			if stopSeq < 0 {
				container.AddNotice(notice.NewNegativeStopSequenceNotice(
					row.Values["trip_id"],
					stopSeq,
					row.RowNumber,
				))
			}
		}
	}

	// Check for negative shape_dist_traveled
	if shapeDistStr, exists := row.Values["shape_dist_traveled"]; exists && strings.TrimSpace(shapeDistStr) != "" {
		if shapeDist, err := strconv.ParseFloat(strings.TrimSpace(shapeDistStr), 64); err == nil {
			if shapeDist < 0 {
				container.AddNotice(notice.NewNegativeShapeDistanceNotice(
					row.Values["trip_id"],
					shapeDist,
					row.RowNumber,
				))
			}
		}
	}
}

// validateStopRow validates stops.txt specific patterns
func (v *InvalidRowValidator) validateStopRow(container *notice.NoticeContainer, row *parser.CSVRow) {
	// Check for invalid location_type values
	if locTypeStr, exists := row.Values["location_type"]; exists && strings.TrimSpace(locTypeStr) != "" {
		if locType, err := strconv.Atoi(strings.TrimSpace(locTypeStr)); err == nil {
			if locType < 0 || locType > 4 {
				container.AddNotice(notice.NewInvalidLocationTypeNotice(
					row.Values["stop_id"],
					locType,
					row.RowNumber,
				))
			}
		}
	}

	// Check for invalid wheelchair_boarding values
	if wheelchairStr, exists := row.Values["wheelchair_boarding"]; exists && strings.TrimSpace(wheelchairStr) != "" {
		if wheelchair, err := strconv.Atoi(strings.TrimSpace(wheelchairStr)); err == nil {
			if wheelchair < 0 || wheelchair > 2 {
				container.AddNotice(notice.NewInvalidWheelchairBoardingNotice(
					row.Values["stop_id"],
					wheelchair,
					row.RowNumber,
				))
			}
		}
	}
}

// validateRouteRow validates routes.txt specific patterns
func (v *InvalidRowValidator) validateRouteRow(container *notice.NoticeContainer, row *parser.CSVRow) {
	// Check for invalid route_type values
	if routeTypeStr, exists := row.Values["route_type"]; exists {
		if routeType, err := strconv.Atoi(strings.TrimSpace(routeTypeStr)); err == nil {
			validTypes := []int{0, 1, 2, 3, 4, 5, 6, 7, 11, 12}
			valid := false
			for _, validType := range validTypes {
				if routeType == validType {
					valid = true
					break
				}
			}
			// Extended route types (100-1700) are also valid
			if routeType >= 100 && routeType <= 1700 {
				valid = true
			}

			if !valid {
				container.AddNotice(notice.NewInvalidRouteTypeNotice(
					row.Values["route_id"],
					routeTypeStr,
					row.RowNumber,
					"Invalid route_type value",
				))
			}
		}
	}
}

// validateTripRow validates trips.txt specific patterns
func (v *InvalidRowValidator) validateTripRow(container *notice.NoticeContainer, row *parser.CSVRow) {
	// Check for invalid direction_id values
	if directionStr, exists := row.Values["direction_id"]; exists && strings.TrimSpace(directionStr) != "" {
		if direction, err := strconv.Atoi(strings.TrimSpace(directionStr)); err == nil {
			if direction < 0 || direction > 1 {
				container.AddNotice(notice.NewInvalidDirectionIdNotice(
					row.Values["trip_id"],
					direction,
					row.RowNumber,
				))
			}
		}
	}

	// Check for invalid wheelchair_accessible values
	if wheelchairStr, exists := row.Values["wheelchair_accessible"]; exists && strings.TrimSpace(wheelchairStr) != "" {
		if wheelchair, err := strconv.Atoi(strings.TrimSpace(wheelchairStr)); err == nil {
			if wheelchair < 0 || wheelchair > 2 {
				container.AddNotice(notice.NewInvalidWheelchairAccessibleNotice(
					row.Values["trip_id"],
					wheelchair,
					row.RowNumber,
				))
			}
		}
	}

	// Check for invalid bikes_allowed values
	if bikesStr, exists := row.Values["bikes_allowed"]; exists && strings.TrimSpace(bikesStr) != "" {
		if bikes, err := strconv.Atoi(strings.TrimSpace(bikesStr)); err == nil {
			if bikes < 0 || bikes > 2 {
				container.AddNotice(notice.NewInvalidBikesAllowedNotice(
					row.Values["trip_id"],
					bikes,
					row.RowNumber,
				))
			}
		}
	}
}

// validateCalendarRow validates calendar.txt specific patterns
func (v *InvalidRowValidator) validateCalendarRow(container *notice.NoticeContainer, row *parser.CSVRow) {
	// Check day fields are 0 or 1
	dayFields := []string{"monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday"}

	for _, field := range dayFields {
		if value, exists := row.Values[field]; exists {
			trimmedValue := strings.TrimSpace(value)
			if trimmedValue != "0" && trimmedValue != "1" && trimmedValue != "" {
				container.AddNotice(notice.NewInvalidDayValueNotice(
					row.Values["service_id"],
					field,
					trimmedValue,
					row.RowNumber,
				))
			}
		}
	}
}

// validateCalendarDateRow validates calendar_dates.txt specific patterns
func (v *InvalidRowValidator) validateCalendarDateRow(container *notice.NoticeContainer, row *parser.CSVRow) {
	// Check for invalid exception_type values
	if exceptionStr, exists := row.Values["exception_type"]; exists {
		if exception, err := strconv.Atoi(strings.TrimSpace(exceptionStr)); err == nil {
			if exception < 1 || exception > 2 {
				container.AddNotice(notice.NewInvalidExceptionTypeNotice(
					row.Values["service_id"],
					row.Values["date"],
					exception,
					row.RowNumber,
				))
			}
		}
	}
}

// validateShapeRow validates shapes.txt specific patterns
func (v *InvalidRowValidator) validateShapeRow(container *notice.NoticeContainer, row *parser.CSVRow) {
	// Check for negative shape_pt_sequence
	if sequenceStr, exists := row.Values["shape_pt_sequence"]; exists {
		if sequence, err := strconv.Atoi(strings.TrimSpace(sequenceStr)); err == nil {
			if sequence < 0 {
				container.AddNotice(notice.NewNegativeShapeSequenceNotice(
					row.Values["shape_id"],
					sequence,
					row.RowNumber,
				))
			}
		}
	}
}

// validateFareAttributeRow validates fare_attributes.txt specific patterns
func (v *InvalidRowValidator) validateFareAttributeRow(container *notice.NoticeContainer, row *parser.CSVRow) {
	// Check for invalid payment_method values
	if paymentStr, exists := row.Values["payment_method"]; exists {
		if payment, err := strconv.Atoi(strings.TrimSpace(paymentStr)); err == nil {
			if payment < 0 || payment > 1 {
				container.AddNotice(notice.NewInvalidPaymentMethodNotice(
					row.Values["fare_id"],
					payment,
					row.RowNumber,
				))
			}
		}
	}

	// Check for invalid transfers values
	if transfersStr, exists := row.Values["transfers"]; exists && strings.TrimSpace(transfersStr) != "" {
		if transfers, err := strconv.Atoi(strings.TrimSpace(transfersStr)); err == nil {
			if transfers < 0 || transfers > 2 {
				container.AddNotice(notice.NewInvalidTransfersNotice(
					row.Values["fare_id"],
					transfers,
					row.RowNumber,
				))
			}
		}
	}
}

// validateFrequencyRow validates frequencies.txt specific patterns
func (v *InvalidRowValidator) validateFrequencyRow(container *notice.NoticeContainer, row *parser.CSVRow) {
	// Check for non-positive headway_secs
	if headwayStr, exists := row.Values["headway_secs"]; exists {
		if headway, err := strconv.Atoi(strings.TrimSpace(headwayStr)); err == nil {
			if headway <= 0 {
				container.AddNotice(notice.NewInvalidHeadwayNotice(
					row.Values["trip_id"],
					headway,
					row.RowNumber,
				))
			}
		}
	}

	// Check for invalid exact_times values
	if exactStr, exists := row.Values["exact_times"]; exists && strings.TrimSpace(exactStr) != "" {
		if exact, err := strconv.Atoi(strings.TrimSpace(exactStr)); err == nil {
			if exact < 0 || exact > 1 {
				container.AddNotice(notice.NewInvalidExactTimesNotice(
					row.Values["trip_id"],
					exact,
					row.RowNumber,
				))
			}
		}
	}
}

// validateTransferRow validates transfers.txt specific patterns
func (v *InvalidRowValidator) validateTransferRow(container *notice.NoticeContainer, row *parser.CSVRow) {
	// Check for invalid transfer_type values
	if transferStr, exists := row.Values["transfer_type"]; exists && strings.TrimSpace(transferStr) != "" {
		if transferType, err := strconv.Atoi(strings.TrimSpace(transferStr)); err == nil {
			if transferType < 0 || transferType > 3 {
				container.AddNotice(notice.NewInvalidTransferTypeNotice(
					row.Values["from_stop_id"],
					row.Values["to_stop_id"],
					transferType,
					row.RowNumber,
				))
			}
		}
	}
}
