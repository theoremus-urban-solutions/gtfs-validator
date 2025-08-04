package core

import (
	"io"
	"net/mail"
	"net/url"
	"strings"
	"time"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/parser"
	"github.com/theoremus-urban-solutions/gtfs-validator/types"
	"github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

// FieldFormatValidator validates field formats (URL, email, timezone, etc.)
type FieldFormatValidator struct{}

// NewFieldFormatValidator creates a new field format validator
func NewFieldFormatValidator() *FieldFormatValidator {
	return &FieldFormatValidator{}
}

// Validate checks field formats in all files
func (v *FieldFormatValidator) Validate(loader *parser.FeedLoader, container *notice.NoticeContainer, config validator.Config) {
	files := loader.ListFiles()

	for _, filename := range files {
		v.validateFile(loader, container, filename)
	}
}

// validateFile validates field formats in a single file
func (v *FieldFormatValidator) validateFile(loader *parser.FeedLoader, container *notice.NoticeContainer, filename string) {
	reader, err := loader.GetFile(filename)
	if err != nil {
		return
	}
	defer reader.Close()

	csvFile, err := parser.NewCSVFile(reader, filename)
	if err != nil {
		return
	}

	// Read and validate each row
	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			return
		}

		// Validate fields based on file type
		switch filename {
		case "agency.txt":
			v.validateAgencyFields(row, container, filename)
		case "stops.txt":
			v.validateStopFields(row, container, filename)
		case "routes.txt":
			v.validateRouteFields(row, container, filename)
		case "stop_times.txt":
			v.validateStopTimeFields(row, container, filename)
		case "calendar.txt":
			v.validateCalendarFields(row, container, filename)
		case "calendar_dates.txt":
			v.validateCalendarDateFields(row, container, filename)
		}
	}
}

// validateAgencyFields validates agency.txt fields
func (v *FieldFormatValidator) validateAgencyFields(row *parser.CSVRow, container *notice.NoticeContainer, filename string) {
	// Validate URL
	if url := row.Values["agency_url"]; url != "" {
		if !v.isValidURL(url) {
			container.AddNotice(notice.NewInvalidURLNotice(filename, "agency_url", url, row.RowNumber))
		}
	}

	// Validate email
	if email := row.Values["agency_email"]; email != "" {
		if !v.isValidEmail(email) {
			container.AddNotice(notice.NewInvalidEmailNotice(filename, "agency_email", email, row.RowNumber))
		}
	}

	// Validate timezone
	if tz := row.Values["agency_timezone"]; tz != "" {
		if !v.isValidTimezone(tz) {
			container.AddNotice(notice.NewInvalidTimezoneNotice(filename, "agency_timezone", tz, row.RowNumber))
		}
	}
}

// validateStopFields validates stops.txt fields
func (v *FieldFormatValidator) validateStopFields(row *parser.CSVRow, container *notice.NoticeContainer, filename string) {
	// Validate URL
	if url := row.Values["stop_url"]; url != "" {
		if !v.isValidURL(url) {
			container.AddNotice(notice.NewInvalidURLNotice(filename, "stop_url", url, row.RowNumber))
		}
	}

	// Validate timezone
	if tz := row.Values["stop_timezone"]; tz != "" {
		if !v.isValidTimezone(tz) {
			container.AddNotice(notice.NewInvalidTimezoneNotice(filename, "stop_timezone", tz, row.RowNumber))
		}
	}
}

// validateRouteFields validates routes.txt fields
func (v *FieldFormatValidator) validateRouteFields(row *parser.CSVRow, container *notice.NoticeContainer, filename string) {
	// Validate URL
	if url := row.Values["route_url"]; url != "" {
		if !v.isValidURL(url) {
			container.AddNotice(notice.NewInvalidURLNotice(filename, "route_url", url, row.RowNumber))
		}
	}

	// Validate color
	if color := row.Values["route_color"]; color != "" {
		if _, err := types.ParseGTFSColor(color); err != nil {
			container.AddNotice(notice.NewInvalidFieldFormatNotice(
				filename, "route_color", color, row.RowNumber, "6-digit hexadecimal",
			))
		}
	}

	// Validate text color
	if color := row.Values["route_text_color"]; color != "" {
		if _, err := types.ParseGTFSColor(color); err != nil {
			container.AddNotice(notice.NewInvalidFieldFormatNotice(
				filename, "route_text_color", color, row.RowNumber, "6-digit hexadecimal",
			))
		}
	}
}

// validateStopTimeFields validates stop_times.txt fields
func (v *FieldFormatValidator) validateStopTimeFields(row *parser.CSVRow, container *notice.NoticeContainer, filename string) {
	// Validate arrival time
	if arrTime := row.Values["arrival_time"]; arrTime != "" {
		if _, err := types.ParseGTFSTime(arrTime); err != nil {
			container.AddNotice(notice.NewInvalidFieldFormatNotice(
				filename, "arrival_time", arrTime, row.RowNumber, "HH:MM:SS",
			))
		}
	}

	// Validate departure time
	if depTime := row.Values["departure_time"]; depTime != "" {
		if _, err := types.ParseGTFSTime(depTime); err != nil {
			container.AddNotice(notice.NewInvalidFieldFormatNotice(
				filename, "departure_time", depTime, row.RowNumber, "HH:MM:SS",
			))
		}
	}
}

// validateCalendarFields validates calendar.txt fields
func (v *FieldFormatValidator) validateCalendarFields(row *parser.CSVRow, container *notice.NoticeContainer, filename string) {
	// Validate start date
	if startDate := row.Values["start_date"]; startDate != "" {
		if _, err := types.ParseGTFSDate(startDate); err != nil {
			container.AddNotice(notice.NewInvalidFieldFormatNotice(
				filename, "start_date", startDate, row.RowNumber, "YYYYMMDD",
			))
		}
	}

	// Validate end date
	if endDate := row.Values["end_date"]; endDate != "" {
		if _, err := types.ParseGTFSDate(endDate); err != nil {
			container.AddNotice(notice.NewInvalidFieldFormatNotice(
				filename, "end_date", endDate, row.RowNumber, "YYYYMMDD",
			))
		}
	}
}

// validateCalendarDateFields validates calendar_dates.txt fields
func (v *FieldFormatValidator) validateCalendarDateFields(row *parser.CSVRow, container *notice.NoticeContainer, filename string) {
	// Validate date
	if date := row.Values["date"]; date != "" {
		if _, err := types.ParseGTFSDate(date); err != nil {
			container.AddNotice(notice.NewInvalidFieldFormatNotice(
				filename, "date", date, row.RowNumber, "YYYYMMDD",
			))
		}
	}
}

// isValidURL checks if a string is a valid URL
func (v *FieldFormatValidator) isValidURL(s string) bool {
	u, err := url.Parse(s)
	if err != nil {
		return false
	}
	return u.Scheme == "http" || u.Scheme == "https"
}

// isValidEmail checks if a string is a valid email address
func (v *FieldFormatValidator) isValidEmail(s string) bool {
	_, err := mail.ParseAddress(s)
	return err == nil
}

// isValidTimezone checks if a string is a valid timezone
func (v *FieldFormatValidator) isValidTimezone(s string) bool {
	_, err := time.LoadLocation(s)
	return err == nil
}
