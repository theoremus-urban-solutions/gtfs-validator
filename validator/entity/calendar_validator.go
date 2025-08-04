package entity

import (
	"io"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/parser"
	"github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

// CalendarValidator validates that at least one of calendar.txt or calendar_dates.txt exists
type CalendarValidator struct{}

// NewCalendarValidator creates a new calendar validator
func NewCalendarValidator() *CalendarValidator {
	return &CalendarValidator{}
}

// Validate checks that at least one calendar file exists
func (v *CalendarValidator) Validate(loader *parser.FeedLoader, container *notice.NoticeContainer, config validator.Config) {
	files := loader.ListFiles()

	hasCalendar := false
	hasCalendarDates := false

	for _, filename := range files {
		if filename == "calendar.txt" {
			hasCalendar = v.fileHasData(loader, filename)
		}
		if filename == "calendar_dates.txt" {
			hasCalendarDates = v.fileHasData(loader, filename)
		}
	}

	if !hasCalendar && !hasCalendarDates {
		container.AddNotice(notice.NewMissingCalendarAndCalendarDateFilesNotice())
	}
}

// fileHasData checks if a file exists and has at least one data row
func (v *CalendarValidator) fileHasData(loader *parser.FeedLoader, filename string) bool {
	reader, err := loader.GetFile(filename)
	if err != nil {
		return false
	}
	defer reader.Close()

	csvFile, err := parser.NewCSVFile(reader, filename)
	if err != nil {
		return false
	}

	// Try to read at least one row
	_, err = csvFile.ReadRow()
	return err != io.EOF && err == nil
}
