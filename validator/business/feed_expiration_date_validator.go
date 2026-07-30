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

// FeedExpirationDateValidator checks the validity window a feed declares for
// itself in feed_info.txt. Service dates in calendar.txt and calendar_dates.txt
// are a separate question, covered by expired_calendar and future_calendar.
type FeedExpirationDateValidator struct{}

// NewFeedExpirationDateValidator creates a new feed expiration date validator
func NewFeedExpirationDateValidator() *FeedExpirationDateValidator {
	return &FeedExpirationDateValidator{}
}

// Validate checks feed expiration and freshness
func (v *FeedExpirationDateValidator) Validate(loader *parser.FeedLoader, container *notice.NoticeContainer, config validator.Config) {
	currentDate, ok := config.CurrentDate.(time.Time)
	if !ok {
		currentDate = time.Now()
	}

	feedInfo := v.loadFeedInfo(loader)
	if feedInfo == nil {
		// feed_info.txt is optional, and its absence is reported as
		// missing_recommended_file rather than here.
		return
	}

	v.validateFeedStartDate(container, feedInfo, currentDate)
	v.validateFeedEndDate(container, feedInfo, currentDate)
}

// FeedInfo represents feed information
type FeedInfo struct {
	FeedPublisherName string
	FeedPublisherURL  string
	FeedLang          string
	FeedStartDate     *time.Time
	FeedEndDate       *time.Time
	FeedVersion       string
	RowNumber         int
}

// loadFeedInfo loads feed information from feed_info.txt
func (v *FeedExpirationDateValidator) loadFeedInfo(loader *parser.FeedLoader) *FeedInfo {
	reader, err := loader.GetFile("feed_info.txt")
	if err != nil {
		return nil
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close reader %v", closeErr)
		}
	}()

	csvFile, err := parser.NewCSVFile(reader, "feed_info.txt")
	if err != nil {
		return nil
	}

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		feedInfo := &FeedInfo{
			RowNumber: row.RowNumber,
		}

		if name, hasName := row.Values["feed_publisher_name"]; hasName {
			feedInfo.FeedPublisherName = strings.TrimSpace(name)
		}
		if url, hasURL := row.Values["feed_publisher_url"]; hasURL {
			feedInfo.FeedPublisherURL = strings.TrimSpace(url)
		}
		if lang, hasLang := row.Values["feed_lang"]; hasLang {
			feedInfo.FeedLang = strings.TrimSpace(lang)
		}
		if version, hasVersion := row.Values["feed_version"]; hasVersion {
			feedInfo.FeedVersion = strings.TrimSpace(version)
		}

		// Parse dates
		if startDateStr, hasStart := row.Values["feed_start_date"]; hasStart && strings.TrimSpace(startDateStr) != "" {
			if startDate := v.parseGTFSDate(strings.TrimSpace(startDateStr)); startDate != nil {
				feedInfo.FeedStartDate = startDate
			}
		}
		if endDateStr, hasEnd := row.Values["feed_end_date"]; hasEnd && strings.TrimSpace(endDateStr) != "" {
			if endDate := v.parseGTFSDate(strings.TrimSpace(endDateStr)); endDate != nil {
				feedInfo.FeedEndDate = endDate
			}
		}

		// Return first row (feed_info.txt should have only one row)
		return feedInfo
	}

	return nil
}

// parseGTFSDate parses a GTFS date string (YYYYMMDD)
func (v *FeedExpirationDateValidator) parseGTFSDate(dateStr string) *time.Time {
	if len(dateStr) != 8 {
		return nil
	}

	year, err1 := strconv.Atoi(dateStr[0:4])
	month, err2 := strconv.Atoi(dateStr[4:6])
	day, err3 := strconv.Atoi(dateStr[6:8])

	if err1 != nil || err2 != nil || err3 != nil {
		return nil
	}

	if month < 1 || month > 12 || day < 1 || day > 31 {
		return nil
	}

	date := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	return &date
}

// formatGTFSDate formats a time as GTFS date string
func (v *FeedExpirationDateValidator) formatGTFSDate(date time.Time) string {
	return date.Format("20060102")
}

// validateFeedStartDate reports a feed that declares itself as starting later
// than today, and so covers nothing for a consumer reading it now.
func (v *FeedExpirationDateValidator) validateFeedStartDate(container *notice.NoticeContainer, feedInfo *FeedInfo, currentDate time.Time) {
	if feedInfo.FeedStartDate == nil || !feedInfo.FeedStartDate.After(currentDate) {
		return
	}

	container.AddNotice(notice.NewFutureFeedNotice(
		feedInfo.RowNumber,
		v.formatGTFSDate(*feedInfo.FeedStartDate),
		v.formatGTFSDate(currentDate),
	))
}

// validateFeedEndDate reports a feed whose declared validity runs out inside
// the next 7 or 30 days. A feed already past its end date trips the 7-day
// notice, which is the more urgent of the two; only one is emitted.
//
// A missing feed_end_date is reported as missing_feed_info_date, not here.
func (v *FeedExpirationDateValidator) validateFeedEndDate(container *notice.NoticeContainer, feedInfo *FeedInfo, currentDate time.Time) {
	if feedInfo.FeedEndDate == nil {
		return
	}

	endDate := *feedInfo.FeedEndDate
	daysUntilExpiration := int(endDate.Sub(currentDate).Hours() / 24)

	switch {
	case endDate.Before(currentDate.AddDate(0, 0, 7)):
		container.AddNotice(notice.NewFeedExpiresWithin7DaysNotice(
			v.formatGTFSDate(endDate),
			v.formatGTFSDate(currentDate),
			daysUntilExpiration,
		))
	case endDate.Before(currentDate.AddDate(0, 0, 30)):
		container.AddNotice(notice.NewFeedExpiresWithin30DaysNotice(
			v.formatGTFSDate(endDate),
			v.formatGTFSDate(currentDate),
			daysUntilExpiration,
		))
	}
}
