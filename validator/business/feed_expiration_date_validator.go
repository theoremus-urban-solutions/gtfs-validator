package business

import (
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/parser"
	"github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

// FeedExpirationDateValidator validates feed expiration and freshness
type FeedExpirationDateValidator struct{}

// NewFeedExpirationDateValidator creates a new feed expiration date validator
func NewFeedExpirationDateValidator() *FeedExpirationDateValidator {
	return &FeedExpirationDateValidator{}
}

// Validate checks feed expiration and freshness
func (v *FeedExpirationDateValidator) Validate(loader *parser.FeedLoader, container *notice.NoticeContainer, config validator.Config) {
	// Type assert CurrentDate from interface{}
	var currentDate time.Time
	if config.CurrentDate != nil {
		if cd, ok := config.CurrentDate.(time.Time); ok {
			currentDate = cd
		} else {
			currentDate = time.Now()
		}
	} else {
		currentDate = time.Now()
	}

	feedInfo := v.loadFeedInfo(loader)
	if feedInfo == nil {
		// No feed_info.txt file, check calendar dates instead
		v.validateCalendarBasedExpiration(loader, container, currentDate)
		return
	}

	// Validate feed expiration based on feed_info.txt
	v.validateFeedInfoExpiration(container, feedInfo, currentDate)

	// Also validate service dates exist for the next period
	v.validateServiceCoverage(loader, container, currentDate)
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
	defer reader.Close()

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

// validateFeedInfoExpiration validates feed expiration based on feed_info.txt
func (v *FeedExpirationDateValidator) validateFeedInfoExpiration(container *notice.NoticeContainer, feedInfo *FeedInfo, currentDate time.Time) {
	if feedInfo.FeedEndDate == nil {
		// No end date specified - this is allowed but not recommended
		container.AddNotice(notice.NewFeedInfoEndDateMissingNotice(feedInfo.RowNumber))
		return
	}

	endDate := *feedInfo.FeedEndDate
	daysUntilExpiration := int(endDate.Sub(currentDate).Hours() / 24)

	// Feed has already expired
	if endDate.Before(currentDate) {
		container.AddNotice(notice.NewFeedExpiredNotice(
			v.formatGTFSDate(endDate),
			v.formatGTFSDate(currentDate),
			-daysUntilExpiration,
		))
		return
	}

	// Feed expires within 7 days (critical)
	if daysUntilExpiration <= 7 {
		container.AddNotice(notice.NewFeedExpiresWithin7DaysNotice(
			v.formatGTFSDate(endDate),
			v.formatGTFSDate(currentDate),
			daysUntilExpiration,
		))
		return
	}

	// Feed expires within 30 days (warning)
	if daysUntilExpiration <= 30 {
		container.AddNotice(notice.NewFeedExpiresWithin30DaysNotice(
			v.formatGTFSDate(endDate),
			v.formatGTFSDate(currentDate),
			daysUntilExpiration,
		))
	}
}

// validateCalendarBasedExpiration validates expiration based on calendar.txt and calendar_dates.txt
func (v *FeedExpirationDateValidator) validateCalendarBasedExpiration(loader *parser.FeedLoader, container *notice.NoticeContainer, currentDate time.Time) {
	// Find latest service date from calendar.txt
	latestCalendarDate := v.findLatestCalendarDate(loader)

	// Find latest service date from calendar_dates.txt
	latestCalendarDatesDate := v.findLatestCalendarDatesDate(loader)

	// Use the later of the two dates
	var latestServiceDate *time.Time
	if latestCalendarDate != nil && latestCalendarDatesDate != nil {
		if latestCalendarDate.After(*latestCalendarDatesDate) {
			latestServiceDate = latestCalendarDate
		} else {
			latestServiceDate = latestCalendarDatesDate
		}
	} else if latestCalendarDate != nil {
		latestServiceDate = latestCalendarDate
	} else if latestCalendarDatesDate != nil {
		latestServiceDate = latestCalendarDatesDate
	}

	if latestServiceDate == nil {
		container.AddNotice(notice.NewNoServiceDateFoundNotice())
		return
	}

	daysUntilExpiration := int(latestServiceDate.Sub(currentDate).Hours() / 24)

	// Service has already ended
	if latestServiceDate.Before(currentDate) {
		container.AddNotice(notice.NewServiceExpiredNotice(
			v.formatGTFSDate(*latestServiceDate),
			v.formatGTFSDate(currentDate),
			-daysUntilExpiration,
		))
		return
	}

	// Service ends within 7 days (critical)
	if daysUntilExpiration <= 7 {
		container.AddNotice(notice.NewServiceExpiresWithin7DaysNotice(
			v.formatGTFSDate(*latestServiceDate),
			v.formatGTFSDate(currentDate),
			daysUntilExpiration,
		))
		return
	}

	// Service ends within 30 days (warning)
	if daysUntilExpiration <= 30 {
		container.AddNotice(notice.NewServiceExpiresWithin30DaysNotice(
			v.formatGTFSDate(*latestServiceDate),
			v.formatGTFSDate(currentDate),
			daysUntilExpiration,
		))
	}
}

// findLatestCalendarDate finds the latest end_date in calendar.txt
func (v *FeedExpirationDateValidator) findLatestCalendarDate(loader *parser.FeedLoader) *time.Time {
	reader, err := loader.GetFile("calendar.txt")
	if err != nil {
		return nil
	}
	defer reader.Close()

	csvFile, err := parser.NewCSVFile(reader, "calendar.txt")
	if err != nil {
		return nil
	}

	var latestDate *time.Time

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		if endDateStr, hasEndDate := row.Values["end_date"]; hasEndDate {
			if endDate := v.parseGTFSDate(strings.TrimSpace(endDateStr)); endDate != nil {
				if latestDate == nil || endDate.After(*latestDate) {
					latestDate = endDate
				}
			}
		}
	}

	return latestDate
}

// findLatestCalendarDatesDate finds the latest date in calendar_dates.txt
func (v *FeedExpirationDateValidator) findLatestCalendarDatesDate(loader *parser.FeedLoader) *time.Time {
	reader, err := loader.GetFile("calendar_dates.txt")
	if err != nil {
		return nil
	}
	defer reader.Close()

	csvFile, err := parser.NewCSVFile(reader, "calendar_dates.txt")
	if err != nil {
		return nil
	}

	var latestDate *time.Time

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		if dateStr, hasDate := row.Values["date"]; hasDate {
			if date := v.parseGTFSDate(strings.TrimSpace(dateStr)); date != nil {
				if latestDate == nil || date.After(*latestDate) {
					latestDate = date
				}
			}
		}
	}

	return latestDate
}

// validateServiceCoverage validates that service exists for the next 7 days
func (v *FeedExpirationDateValidator) validateServiceCoverage(loader *parser.FeedLoader, container *notice.NoticeContainer, currentDate time.Time) {
	// Get all service IDs that are active in the next 7 days
	activeServices := v.getActiveServices(loader, currentDate, currentDate.AddDate(0, 0, 7))

	if len(activeServices) == 0 {
		container.AddNotice(notice.NewNoServiceNext7DaysNotice(
			v.formatGTFSDate(currentDate),
			v.formatGTFSDate(currentDate.AddDate(0, 0, 7)),
		))
		return
	}

	// Check if we have trips for these services
	tripCount := v.countTripsForServices(loader, activeServices)
	if tripCount == 0 {
		container.AddNotice(notice.NewNoTripsNext7DaysNotice(
			v.formatGTFSDate(currentDate),
			v.formatGTFSDate(currentDate.AddDate(0, 0, 7)),
			len(activeServices),
		))
	}
}

// getActiveServices returns service IDs that are active in the given date range
func (v *FeedExpirationDateValidator) getActiveServices(loader *parser.FeedLoader, startDate, endDate time.Time) map[string]bool {
	activeServices := make(map[string]bool)

	// Check calendar.txt for regular services
	v.addCalendarServices(loader, activeServices, startDate, endDate)

	// Check calendar_dates.txt for exceptions
	v.addCalendarDatesServices(loader, activeServices, startDate, endDate)

	return activeServices
}

// addCalendarServices adds services from calendar.txt that are active in the date range
func (v *FeedExpirationDateValidator) addCalendarServices(loader *parser.FeedLoader, activeServices map[string]bool, startDate, endDate time.Time) {
	reader, err := loader.GetFile("calendar.txt")
	if err != nil {
		return
	}
	defer reader.Close()

	csvFile, err := parser.NewCSVFile(reader, "calendar.txt")
	if err != nil {
		return
	}

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		serviceID, hasServiceID := row.Values["service_id"]
		if !hasServiceID {
			continue
		}

		// Parse service period
		serviceStartStr, hasStart := row.Values["start_date"]
		serviceEndStr, hasEnd := row.Values["end_date"]
		if !hasStart || !hasEnd {
			continue
		}

		serviceStart := v.parseGTFSDate(strings.TrimSpace(serviceStartStr))
		serviceEnd := v.parseGTFSDate(strings.TrimSpace(serviceEndStr))
		if serviceStart == nil || serviceEnd == nil {
			continue
		}

		// Check if service period overlaps with our date range
		if serviceEnd.Before(startDate) || serviceStart.After(endDate) {
			continue
		}

		// Check if any day of the week is active
		hasActiveDay := false
		for _, day := range []string{"monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday"} {
			if dayValue, hasDay := row.Values[day]; hasDay && strings.TrimSpace(dayValue) == "1" {
				hasActiveDay = true
				break
			}
		}

		if hasActiveDay {
			activeServices[strings.TrimSpace(serviceID)] = true
		}
	}
}

// addCalendarDatesServices adds services from calendar_dates.txt that are active in the date range
func (v *FeedExpirationDateValidator) addCalendarDatesServices(loader *parser.FeedLoader, activeServices map[string]bool, startDate, endDate time.Time) {
	reader, err := loader.GetFile("calendar_dates.txt")
	if err != nil {
		return
	}
	defer reader.Close()

	csvFile, err := parser.NewCSVFile(reader, "calendar_dates.txt")
	if err != nil {
		return
	}

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		serviceID, hasServiceID := row.Values["service_id"]
		dateStr, hasDate := row.Values["date"]
		exceptionTypeStr, hasException := row.Values["exception_type"]

		if !hasServiceID || !hasDate || !hasException {
			continue
		}

		date := v.parseGTFSDate(strings.TrimSpace(dateStr))
		if date == nil {
			continue
		}

		// Check if date is in our range
		if date.Before(startDate) || date.After(endDate) {
			continue
		}

		exceptionType, err := strconv.Atoi(strings.TrimSpace(exceptionTypeStr))
		if err != nil {
			continue
		}

		// 1 = service added, 2 = service removed
		if exceptionType == 1 {
			activeServices[strings.TrimSpace(serviceID)] = true
		} else if exceptionType == 2 {
			delete(activeServices, strings.TrimSpace(serviceID))
		}
	}
}

// countTripsForServices counts trips that use the given services
func (v *FeedExpirationDateValidator) countTripsForServices(loader *parser.FeedLoader, services map[string]bool) int {
	if len(services) == 0 {
		return 0
	}

	reader, err := loader.GetFile("trips.txt")
	if err != nil {
		return 0
	}
	defer reader.Close()

	csvFile, err := parser.NewCSVFile(reader, "trips.txt")
	if err != nil {
		return 0
	}

	count := 0
	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		if serviceID, hasServiceID := row.Values["service_id"]; hasServiceID {
			if services[strings.TrimSpace(serviceID)] {
				count++
			}
		}
	}

	return count
}
