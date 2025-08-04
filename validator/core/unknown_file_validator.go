package core

import (
	"path/filepath"
	"strings"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/parser"
	"github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

// UnknownFileValidator validates that only known GTFS files are present
type UnknownFileValidator struct{}

// NewUnknownFileValidator creates a new unknown file validator
func NewUnknownFileValidator() *UnknownFileValidator {
	return &UnknownFileValidator{}
}

// knownGTFSFiles contains all recognized GTFS file names
var knownGTFSFiles = map[string]bool{
	"agency.txt":               true,
	"stops.txt":                true,
	"routes.txt":               true,
	"trips.txt":                true,
	"stop_times.txt":           true,
	"calendar.txt":             true,
	"calendar_dates.txt":       true,
	"fare_attributes.txt":      true,
	"fare_rules.txt":           true,
	"shapes.txt":               true,
	"frequencies.txt":          true,
	"transfers.txt":            true,
	"pathways.txt":             true,
	"levels.txt":               true,
	"feed_info.txt":            true,
	"translations.txt":         true,
	"attributions.txt":         true,
	"fare_media.txt":           true,
	"fare_products.txt":        true,
	"fare_leg_rules.txt":       true,
	"fare_transfer_rules.txt":  true,
	"areas.txt":                true,
	"stop_areas.txt":           true,
	"networks.txt":             true,
	"route_networks.txt":       true,
	"shapes_geojson.txt":       true,
	"booking_rules.txt":        true,
	"location_groups.txt":      true,
	"location_group_stops.txt": true,
	"locations.geojson":        true,
}

// Validate checks for unknown files in the GTFS feed
func (v *UnknownFileValidator) Validate(loader *parser.FeedLoader, container *notice.NoticeContainer, config validator.Config) {
	files := loader.ListFiles()

	for _, filename := range files {
		v.validateKnownFile(container, filename)
	}
}

// validateKnownFile checks if a file is a recognized GTFS file
func (v *UnknownFileValidator) validateKnownFile(container *notice.NoticeContainer, filename string) {
	// Get the base filename (without directory path)
	baseFilename := filepath.Base(filename)

	// Skip non-text files and hidden files
	if !strings.HasSuffix(baseFilename, ".txt") && !strings.HasSuffix(baseFilename, ".geojson") {
		return
	}

	if strings.HasPrefix(baseFilename, ".") {
		return
	}

	// Check if file is known
	if !knownGTFSFiles[baseFilename] {
		container.AddNotice(notice.NewUnknownFileNotice(baseFilename))
	}
}
