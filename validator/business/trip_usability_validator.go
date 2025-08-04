package business

import (
	"io"
	"strings"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/parser"
	"github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

// TripUsabilityValidator validates that trips are usable (have at least 2 stops)
type TripUsabilityValidator struct{}

// NewTripUsabilityValidator creates a new trip usability validator
func NewTripUsabilityValidator() *TripUsabilityValidator {
	return &TripUsabilityValidator{}
}

// Validate checks that all trips have at least 2 stop times
func (v *TripUsabilityValidator) Validate(loader *parser.FeedLoader, container *notice.NoticeContainer, config validator.Config) {
	reader, err := loader.GetFile("stop_times.txt")
	if err != nil {
		return // File doesn't exist, skip validation
	}
	defer reader.Close()

	csvFile, err := parser.NewCSVFile(reader, "stop_times.txt")
	if err != nil {
		return
	}

	// Count stop times per trip
	tripStopCounts := make(map[string]int)
	tripFirstRow := make(map[string]int)

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}

		tripID, hasTripID := row.Values["trip_id"]
		if !hasTripID || strings.TrimSpace(tripID) == "" {
			continue
		}

		tripID = strings.TrimSpace(tripID)

		// Track first row number for each trip (for error reporting)
		if _, exists := tripFirstRow[tripID]; !exists {
			tripFirstRow[tripID] = row.RowNumber
		}

		tripStopCounts[tripID]++
	}

	// Check for trips with fewer than 2 stops
	for tripID, stopCount := range tripStopCounts {
		if stopCount < 2 {
			container.AddNotice(notice.NewTripUsabilityNotice(
				tripID,
				stopCount,
				tripFirstRow[tripID],
			))
		}
	}
}
