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

	// Feed info is required if translations.txt exists
	if loader.HasFile("translations.txt") && !loader.HasFile("feed_info.txt") {
		container.AddNotice(notice.NewMissingFeedInfoNotice())
	}

	// Fare rules requires fare attributes
	if loader.HasFile("fare_rules.txt") && !loader.HasFile("fare_attributes.txt") {
		container.AddNotice(notice.NewMissingFareAttributesNotice())
	}

	// Levels are required if pathways exist
	if loader.HasFile("pathways.txt") && !loader.HasFile("levels.txt") {
		container.AddNotice(notice.NewMissingLevelsNotice())
	}
}
