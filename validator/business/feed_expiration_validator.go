package business

import (
	"io"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/parser"
	"github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

// FeedExpirationValidator validates feed expiration dates
type FeedExpirationValidator struct{}

// NewFeedExpirationValidator creates a new feed expiration validator
func NewFeedExpirationValidator() *FeedExpirationValidator {
	return &FeedExpirationValidator{}
}

// Validate checks feed expiration dates in feed_info.txt
func (v *FeedExpirationValidator) Validate(loader *parser.FeedLoader, container *notice.NoticeContainer, config validator.Config) {
	reader, err := loader.GetFile("feed_info.txt")
	if err != nil {
		return // File doesn't exist, skip validation
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close reader %v", closeErr)
		}
	}()

	csvFile, err := parser.NewCSVFile(reader, "feed_info.txt")
	if err != nil {
		return
	}

	currentDate := time.Now()

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}

		v.validateFeedEndDate(container, row, currentDate)
	}
}

// validateFeedEndDate validates the feed_end_date field
func (v *FeedExpirationValidator) validateFeedEndDate(container *notice.NoticeContainer, row *parser.CSVRow, currentDate time.Time) {
	feedEndDateStr, hasFeedEndDate := row.Values["feed_end_date"]
	if !hasFeedEndDate || strings.TrimSpace(feedEndDateStr) == "" {
		return // No feed_end_date to validate
	}

	feedEndDate, err := v.parseGTFSDate(strings.TrimSpace(feedEndDateStr))
	if err != nil {
		return // Invalid date format, other validators should catch this
	}

	currentDateFormatted := v.formatGTFSDate(currentDate)
	feedEndDateFormatted := v.formatGTFSDate(*feedEndDate)

	// Check if feed expires within 7 days
	sevenDaysFromNow := currentDate.AddDate(0, 0, 7)
	if feedEndDate.Before(sevenDaysFromNow) || feedEndDate.Equal(sevenDaysFromNow) {
		suggestedDate := v.formatGTFSDate(sevenDaysFromNow)
		container.AddNotice(notice.NewFeedExpirationDate7DaysNotice(
			row.RowNumber,
			currentDateFormatted,
			feedEndDateFormatted,
			suggestedDate,
		))
		return
	}

	// Check if feed expires within 30 days
	thirtyDaysFromNow := currentDate.AddDate(0, 0, 30)
	if feedEndDate.Before(thirtyDaysFromNow) || feedEndDate.Equal(thirtyDaysFromNow) {
		suggestedDate := v.formatGTFSDate(thirtyDaysFromNow)
		container.AddNotice(notice.NewFeedExpirationDate30DaysNotice(
			row.RowNumber,
			currentDateFormatted,
			feedEndDateFormatted,
			suggestedDate,
		))
	}
}

// parseGTFSDate parses a GTFS date string (YYYYMMDD) into a time.Time
func (v *FeedExpirationValidator) parseGTFSDate(dateStr string) (*time.Time, error) {
	if len(dateStr) != 8 {
		return nil, &time.ParseError{Layout: "YYYYMMDD", Value: dateStr, LayoutElem: "YYYY", ValueElem: dateStr}
	}

	year, err := strconv.Atoi(dateStr[:4])
	if err != nil {
		return nil, err
	}

	month, err := strconv.Atoi(dateStr[4:6])
	if err != nil {
		return nil, err
	}

	day, err := strconv.Atoi(dateStr[6:8])
	if err != nil {
		return nil, err
	}

	date := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	return &date, nil
}

// formatGTFSDate formats a time.Time into GTFS date format (YYYYMMDD)
func (v *FeedExpirationValidator) formatGTFSDate(date time.Time) string {
	return date.Format("20060102")
}
