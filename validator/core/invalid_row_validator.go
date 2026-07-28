package core

import (
	"io"
	"log"
	"strings"
	"sync"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/parser"
	"github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

// InvalidRowValidator checks a row's shape: whether it has the number of
// fields its header promises. What the fields contain is the field type
// validator's business.
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

	}
}
