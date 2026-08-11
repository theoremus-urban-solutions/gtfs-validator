package core

import (
	"io"
	"strings"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/parser"
	"github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

// MissingShapesFileValidator reports a feed that draws no trip geometry.
//
// shapes.txt is recommended rather than required because a consumer can fall
// back to straight lines between stops, which is wrong on any route that does
// not travel in one. Demand-responsive trips are the exception: a trip that
// serves an area, or a group of stops the rider chooses between, has no single
// path to draw, so a feed built that way is not missing anything.
//
// This is deliberately not part of MissingFilesValidator, which reports the
// other recommended file. Canonical hands its equivalent the whole feed rather
// than one table, so its loader stands the check down unless every table
// parsed — see wholeFeedValidators in implementation.go. MissingFilesValidator
// is a foundation check that runs before that verdict exists, so a shapes
// check living there could not honour it.
type MissingShapesFileValidator struct{}

// NewMissingShapesFileValidator creates a new missing shapes file validator
func NewMissingShapesFileValidator() *MissingShapesFileValidator {
	return &MissingShapesFileValidator{}
}

// Validate reports the absent geometry unless the feed has no path to draw.
func (v *MissingShapesFileValidator) Validate(loader *parser.FeedLoader, container *notice.NoticeContainer, config validator.Config) {
	if hasDataRows(loader, "shapes.txt") || servesDemandResponsiveTrips(loader) {
		return
	}
	container.AddNotice(notice.NewMissingRecommendedFileNotice("shapes.txt"))
}

// hasDataRows reports whether a file is present and holds at least one row
// under its header. A header on its own is the same as an absent file to
// anything looking for content, and that is the line canonical draws here.
func hasDataRows(loader *parser.FeedLoader, filename string) bool {
	if !loader.HasFile(filename) {
		return false
	}

	reader, err := loader.GetFile(filename)
	if err != nil {
		return false
	}
	defer func() { _ = reader.Close() }()

	csvFile, err := parser.NewCSVFile(reader, filename)
	if err != nil {
		return false
	}

	for {
		_, err := csvFile.ReadRow()
		if err == io.EOF {
			return false
		}
		if err != nil {
			// A row the CSV reader rejects is still a row: the file is not
			// header-only. Structural checks report the row itself.
			continue
		}
		return true
	}
}

// servesDemandResponsiveTrips reports whether any trip is served by an area or
// by a group of stops rather than by stops in sequence.
//
// The two columns that express it are GTFS-Flex, which this validator does not
// otherwise consume. Reading the header first means stop_times.txt — the
// largest file in most feeds — is never opened for a feed that does not use
// them, which is nearly all of them.
func servesDemandResponsiveTrips(loader *parser.FeedLoader) bool {
	if !loader.HasFile("stop_times.txt") {
		return false
	}

	reader, err := loader.GetFile("stop_times.txt")
	if err != nil {
		return false
	}
	defer func() { _ = reader.Close() }()

	csvFile, err := parser.NewCSVFile(reader, "stop_times.txt")
	if err != nil {
		return false
	}

	zoneBased, fixedStops := false, false
	for _, header := range csvFile.Headers {
		switch strings.TrimSpace(header) {
		case "location_id":
			zoneBased = true
		case "location_group_id":
			fixedStops = true
		}
	}
	// A group id only means something when the groups themselves are declared.
	fixedStops = fixedStops && hasDataRows(loader, "location_groups.txt")
	if !zoneBased && !fixedStops {
		return false
	}

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			return false
		}
		if err != nil {
			continue
		}
		if strings.TrimSpace(row.Values["trip_id"]) == "" {
			continue
		}
		// A stop id means this row names a place to stop, whatever else it
		// carries, so the trip has a path after all.
		if strings.TrimSpace(row.Values["stop_id"]) != "" {
			continue
		}
		if zoneBased && strings.TrimSpace(row.Values["location_id"]) != "" {
			return true
		}
		if fixedStops && strings.TrimSpace(row.Values["location_group_id"]) != "" {
			return true
		}
	}
}
