package meta

import (
	"io"
	"strings"
	"time"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/parser"
	"github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

// FeedInfoValidator validates feed_info.txt content
type FeedInfoValidator struct{}

// NewFeedInfoValidator creates a new feed info validator
func NewFeedInfoValidator() *FeedInfoValidator {
	return &FeedInfoValidator{}
}

// FeedInfo represents feed information
type FeedInfo struct {
	FeedPublisherName string
	FeedPublisherURL  string
	FeedLang          string
	DefaultLang       string
	FeedStartDate     string
	FeedEndDate       string
	FeedVersion       string
	FeedContactEmail  string
	FeedContactURL    string
	RowNumber         int
}

// Validate checks feed info consistency
func (v *FeedInfoValidator) Validate(loader *parser.FeedLoader, container *notice.NoticeContainer, config validator.Config) {
	feedInfos := v.loadFeedInfo(loader)

	if len(feedInfos) == 0 {
		// feed_info.txt is optional
		return
	}

	if len(feedInfos) > 1 {
		// Multiple feed info entries
		container.AddNotice(notice.NewMultipleFeedInfoEntriesNotice(len(feedInfos)))
	}

	// Validate each feed info (should be only one)
	for _, feedInfo := range feedInfos {
		v.validateFeedInfo(container, feedInfo, config)
	}
}

// loadFeedInfo loads feed information from feed_info.txt
func (v *FeedInfoValidator) loadFeedInfo(loader *parser.FeedLoader) []*FeedInfo {
	var feedInfos []*FeedInfo

	reader, err := loader.GetFile("feed_info.txt")
	if err != nil {
		return feedInfos
	}
	defer reader.Close()

	csvFile, err := parser.NewCSVFile(reader, "feed_info.txt")
	if err != nil {
		return feedInfos
	}

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}

		feedInfo := v.parseFeedInfo(row)
		if feedInfo != nil {
			feedInfos = append(feedInfos, feedInfo)
		}
	}

	return feedInfos
}

// parseFeedInfo parses a feed info record
func (v *FeedInfoValidator) parseFeedInfo(row *parser.CSVRow) *FeedInfo {
	feedInfo := &FeedInfo{
		RowNumber: row.RowNumber,
	}

	// Parse required fields
	if feedPublisherName, has := row.Values["feed_publisher_name"]; has {
		feedInfo.FeedPublisherName = strings.TrimSpace(feedPublisherName)
	}
	if feedPublisherURL, has := row.Values["feed_publisher_url"]; has {
		feedInfo.FeedPublisherURL = strings.TrimSpace(feedPublisherURL)
	}
	if feedLang, has := row.Values["feed_lang"]; has {
		feedInfo.FeedLang = strings.TrimSpace(feedLang)
	}

	// Parse optional fields
	if defaultLang, has := row.Values["default_lang"]; has {
		feedInfo.DefaultLang = strings.TrimSpace(defaultLang)
	}
	if feedStartDate, has := row.Values["feed_start_date"]; has {
		feedInfo.FeedStartDate = strings.TrimSpace(feedStartDate)
	}
	if feedEndDate, has := row.Values["feed_end_date"]; has {
		feedInfo.FeedEndDate = strings.TrimSpace(feedEndDate)
	}
	if feedVersion, has := row.Values["feed_version"]; has {
		feedInfo.FeedVersion = strings.TrimSpace(feedVersion)
	}
	if feedContactEmail, has := row.Values["feed_contact_email"]; has {
		feedInfo.FeedContactEmail = strings.TrimSpace(feedContactEmail)
	}
	if feedContactURL, has := row.Values["feed_contact_url"]; has {
		feedInfo.FeedContactURL = strings.TrimSpace(feedContactURL)
	}

	return feedInfo
}

// validateFeedInfo validates a single feed info record
func (v *FeedInfoValidator) validateFeedInfo(container *notice.NoticeContainer, feedInfo *FeedInfo, config validator.Config) {
	// Validate required fields
	if feedInfo.FeedPublisherName == "" {
		container.AddNotice(notice.NewMissingRequiredFieldNotice(
			"feed_info.txt",
			"feed_publisher_name",
			feedInfo.RowNumber,
		))
	}

	if feedInfo.FeedPublisherURL == "" {
		container.AddNotice(notice.NewMissingRequiredFieldNotice(
			"feed_info.txt",
			"feed_publisher_url",
			feedInfo.RowNumber,
		))
	}

	if feedInfo.FeedLang == "" {
		container.AddNotice(notice.NewMissingRequiredFieldNotice(
			"feed_info.txt",
			"feed_lang",
			feedInfo.RowNumber,
		))
	}

	// Validate URLs
	v.validateURLs(container, feedInfo)

	// Validate language codes
	v.validateLanguageCodes(container, feedInfo)

	// Validate date range
	v.validateDateRange(container, feedInfo, config)

	// Validate email format
	v.validateEmail(container, feedInfo)
}

// validateURLs validates URL formats
func (v *FeedInfoValidator) validateURLs(container *notice.NoticeContainer, feedInfo *FeedInfo) {
	// Validate feed_publisher_url
	if feedInfo.FeedPublisherURL != "" {
		if !strings.HasPrefix(feedInfo.FeedPublisherURL, "http://") &&
			!strings.HasPrefix(feedInfo.FeedPublisherURL, "https://") {
			container.AddNotice(notice.NewInvalidURLNotice(
				"feed_info.txt",
				"feed_publisher_url",
				feedInfo.FeedPublisherURL,
				feedInfo.RowNumber,
			))
		}
	}

	// Validate feed_contact_url
	if feedInfo.FeedContactURL != "" {
		if !strings.HasPrefix(feedInfo.FeedContactURL, "http://") &&
			!strings.HasPrefix(feedInfo.FeedContactURL, "https://") {
			container.AddNotice(notice.NewInvalidURLNotice(
				"feed_info.txt",
				"feed_contact_url",
				feedInfo.FeedContactURL,
				feedInfo.RowNumber,
			))
		}
	}
}

// validateLanguageCodes validates language code formats
func (v *FeedInfoValidator) validateLanguageCodes(container *notice.NoticeContainer, feedInfo *FeedInfo) {
	// Validate feed_lang (should be ISO 639-1 two-letter code)
	if feedInfo.FeedLang != "" && len(feedInfo.FeedLang) != 2 {
		container.AddNotice(notice.NewInvalidLanguageCodeNotice(
			"feed_info.txt",
			"feed_lang",
			feedInfo.FeedLang,
			feedInfo.RowNumber,
		))
	}

	// Validate default_lang
	if feedInfo.DefaultLang != "" && len(feedInfo.DefaultLang) != 2 {
		container.AddNotice(notice.NewInvalidLanguageCodeNotice(
			"feed_info.txt",
			"default_lang",
			feedInfo.DefaultLang,
			feedInfo.RowNumber,
		))
	}
}

// validateDateRange validates feed date range
func (v *FeedInfoValidator) validateDateRange(container *notice.NoticeContainer, feedInfo *FeedInfo, config validator.Config) {
	if feedInfo.FeedStartDate == "" || feedInfo.FeedEndDate == "" {
		return
	}

	startDate, startErr := time.Parse("20060102", feedInfo.FeedStartDate)
	endDate, endErr := time.Parse("20060102", feedInfo.FeedEndDate)

	if startErr != nil || endErr != nil {
		return // Format validation handled elsewhere
	}

	// Check if end date is before start date
	if endDate.Before(startDate) {
		container.AddNotice(notice.NewFeedInfoEndDateBeforeStartDateNotice(
			feedInfo.FeedStartDate,
			feedInfo.FeedEndDate,
			feedInfo.RowNumber,
		))
	}

	// Cast CurrentDate to time.Time
	currentDate, ok := config.CurrentDate.(time.Time)
	if !ok {
		currentDate = time.Now()
	}

	// Check if feed has expired
	if endDate.Before(currentDate.AddDate(0, 0, -7)) {
		container.AddNotice(notice.NewExpiredFeedNotice(
			feedInfo.FeedEndDate,
			feedInfo.RowNumber,
		))
	}

	// Check if feed is too far in the future
	if startDate.After(currentDate.AddDate(1, 0, 0)) {
		container.AddNotice(notice.NewFutureFeedStartDateNotice(
			feedInfo.FeedStartDate,
			feedInfo.RowNumber,
		))
	}
}

// validateEmail validates email format
func (v *FeedInfoValidator) validateEmail(container *notice.NoticeContainer, feedInfo *FeedInfo) {
	if feedInfo.FeedContactEmail == "" {
		return
	}

	// Basic email validation
	if !strings.Contains(feedInfo.FeedContactEmail, "@") ||
		!strings.Contains(feedInfo.FeedContactEmail, ".") ||
		strings.HasPrefix(feedInfo.FeedContactEmail, "@") ||
		strings.HasSuffix(feedInfo.FeedContactEmail, "@") {
		container.AddNotice(notice.NewInvalidEmailNotice(
			"feed_info.txt",
			"feed_contact_email",
			feedInfo.FeedContactEmail,
			feedInfo.RowNumber,
		))
	}
}
