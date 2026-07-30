package core

import (
	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/parser"
	"github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

// MissingFilesValidator validates presence of required and conditional files
type MissingFilesValidator struct{}

// NewMissingFilesValidator creates a new missing files validator
func NewMissingFilesValidator() *MissingFilesValidator {
	return &MissingFilesValidator{}
}

// Validate checks for missing required and conditionally required files
func (v *MissingFilesValidator) Validate(loader *parser.FeedLoader, container *notice.NoticeContainer, config validator.Config) {
	// Check for missing required files
	v.validateRequiredFiles(loader, container)

	// Check for conditional file requirements
	v.validateConditionalFiles(loader, container)

	// Check for files the spec recommends but does not require
	v.validateRecommendedFiles(loader, container)
}

// validateRequiredFiles checks for absolutely required files
func (v *MissingFilesValidator) validateRequiredFiles(loader *parser.FeedLoader, container *notice.NoticeContainer) {
	requiredFiles := []string{
		"agency.txt",
		"stops.txt",
		"routes.txt",
		"trips.txt",
		"stop_times.txt",
	}

	for _, filename := range requiredFiles {
		if !loader.HasFile(filename) {
			container.AddNotice(notice.NewMissingRequiredFileNotice(filename))
		}
	}
}

// validateConditionalFiles checks for conditionally required files
func (v *MissingFilesValidator) validateConditionalFiles(loader *parser.FeedLoader, container *notice.NoticeContainer) {
	// Calendar files: at least one of calendar.txt or calendar_dates.txt must exist
	hasCalendar := loader.HasFile("calendar.txt")
	hasCalendarDates := loader.HasFile("calendar_dates.txt")

	if !hasCalendar && !hasCalendarDates {
		container.AddNotice(notice.NewMissingCalendarAndCalendarDateFilesNotice())
	}

	// translations.txt turns feed_info.txt from recommended into required: a
	// translation is only meaningful relative to the language the feed declares.
	if loader.HasFile("translations.txt") && !loader.HasFile("feed_info.txt") {
		container.AddNotice(notice.NewMissingRequiredFileNotice("feed_info.txt"))
	}

	// Fare rules requires fare attributes: a rule selects a fare it cannot
	// define itself.
	if loader.HasFile("fare_rules.txt") && !loader.HasFile("fare_attributes.txt") {
		container.AddNotice(notice.NewMissingRequiredFileNotice("fare_attributes.txt"))
	}

	// Pathways describe movement between levels, so levels.txt is worth having
	// alongside them, but the spec stops short of requiring it.
	if loader.HasFile("pathways.txt") && !loader.HasFile("levels.txt") {
		container.AddNotice(notice.NewMissingRecommendedFileNotice("levels.txt"))
	}
}

// validateRecommendedFiles checks for files the spec recommends.
//
// feed_info.txt is the only one: it carries the feed's language, version and
// validity range, none of which any other file states. When translations.txt
// makes it outright required, validateConditionalFiles has already said so.
func (v *MissingFilesValidator) validateRecommendedFiles(loader *parser.FeedLoader, container *notice.NoticeContainer) {
	if loader.HasFile("feed_info.txt") || loader.HasFile("translations.txt") {
		return
	}
	container.AddNotice(notice.NewMissingRecommendedFileNotice("feed_info.txt"))
}
