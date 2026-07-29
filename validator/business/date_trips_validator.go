package business

import (
	"io"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/parser"
	"github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

// DateTripsValidator reports on the span of dates a feed actually runs trips
// on: whether it covers the coming week, whether it has holes in it, and
// whether it agrees with the validity period feed_info.txt declares.
type DateTripsValidator struct{}

// maxServiceSpanDays bounds the day-by-day walk of a single calendar entry, so
// an end_date typo like 29991231 costs a fixed amount of work rather than a
// million iterations.
const maxServiceSpanDays = 3653 // ten years

// maxServiceGapDays is the longest run of days without service that passes
// without comment.
const maxServiceGapDays = 13

// feedValidityGraceDays is how far feed_end_date may run past the last day of
// service before the feed is overclaiming its validity.
const feedValidityGraceDays = 14

// NewDateTripsValidator creates a new date trips validator
func NewDateTripsValidator() *DateTripsValidator {
	return &DateTripsValidator{}
}

// ServiceInfo represents service information
type ServiceInfo struct {
	ServiceID  string
	StartDate  *time.Time
	EndDate    *time.Time
	DaysOfWeek [7]bool // Mon, Tue, Wed, Thu, Fri, Sat, Sun
	RowNumber  int
}

// CalendarException represents calendar_dates.txt exceptions
type CalendarException struct {
	ServiceID     string
	Date          time.Time
	ExceptionType int // 1 = added, 2 = removed
}

// Validate checks the dates the feed runs trips on against the coming week, its
// own continuity, and the validity period feed_info.txt declares
func (v *DateTripsValidator) Validate(loader *parser.FeedLoader, container *notice.NoticeContainer, config validator.Config) {
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

	// Service dates are midnights in UTC, and the day comparisons below are by
	// equality. time.Now() carries a wall-clock time and a local zone, so
	// without this every comparison misses and the feed looks uncovered.
	currentDate = startOfUTCDay(currentDate)

	// Load service information
	services := v.loadServices(loader)
	exceptions := v.loadCalendarExceptions(loader)

	// A feed with no services at all is reported as
	// missing_calendar_and_calendar_date_files by core/missing_files_validator.
	if len(services) == 0 && len(exceptions) == 0 {
		return
	}

	// Counting trips per service in one pass over trips.txt, rather than
	// re-reading the file for each service, keeps this linear in the feed size
	// instead of quadratic.
	serviceDates := v.activeServiceDates(services, exceptions, v.countTripsByService(loader))

	v.validateNext7DaysService(container, serviceDates, currentDate)
	v.validateServiceGaps(container, serviceDates)
	v.validateAgainstFeedPeriod(loader, container, serviceDates)
}

// activeServiceDates returns every date the feed runs at least one trip on, in
// order. A service no trip references contributes nothing however wide its
// window, so those are left out and the result is the feed's real service
// window rather than what its calendars claim.
func (v *DateTripsValidator) activeServiceDates(services map[string]*ServiceInfo, exceptions []CalendarException, tripCounts map[string]int) []time.Time {
	exceptionsByService := make(map[string][]CalendarException)
	for _, exception := range exceptions {
		exceptionsByService[exception.ServiceID] = append(exceptionsByService[exception.ServiceID], exception)
	}

	serviceIDs := make(map[string]bool, len(services)+len(exceptionsByService))
	for serviceID := range services {
		serviceIDs[serviceID] = true
	}
	for serviceID := range exceptionsByService {
		serviceIDs[serviceID] = true
	}

	active := make(map[int64]time.Time)
	for serviceID := range serviceIDs {
		if tripCounts[serviceID] == 0 {
			continue
		}
		for key, date := range v.datesForService(services[serviceID], exceptionsByService[serviceID]) {
			active[key] = date
		}
	}

	dates := make([]time.Time, 0, len(active))
	for _, date := range active {
		dates = append(dates, date)
	}
	sort.Slice(dates, func(i, j int) bool { return dates[i].Before(dates[j]) })

	return dates
}

// datesForService returns the dates one service runs on, keyed by Unix second
// so the caller can union them. A calendar.txt row missing either bound defines
// no window at all — the type layer reports the missing field — but its
// calendar_dates.txt exceptions still stand on their own.
func (v *DateTripsValidator) datesForService(service *ServiceInfo, exceptions []CalendarException) map[int64]time.Time {
	dates := make(map[int64]time.Time)

	if service != nil && service.StartDate != nil && service.EndDate != nil {
		date := *service.StartDate
		for i := 0; !date.After(*service.EndDate) && i < maxServiceSpanDays; i, date = i+1, date.AddDate(0, 0, 1) {
			if service.DaysOfWeek[weekdayIndex(date)] {
				dates[date.Unix()] = date
			}
		}
	}

	for _, exception := range exceptions {
		switch exception.ExceptionType {
		case 1: // Service added
			dates[exception.Date.Unix()] = exception.Date
		case 2: // Service removed
			delete(dates, exception.Date.Unix())
		}
	}

	return dates
}

// weekdayIndex converts a date to the calendar.txt column order, which starts
// at Monday rather than Go's Sunday.
func weekdayIndex(date time.Time) int {
	if date.Weekday() == time.Sunday {
		return 6
	}
	return int(date.Weekday()) - 1
}

// loadServices loads service information from calendar.txt
func (v *DateTripsValidator) loadServices(loader *parser.FeedLoader) map[string]*ServiceInfo {
	services := make(map[string]*ServiceInfo)

	reader, err := loader.GetFile("calendar.txt")
	if err != nil {
		return services
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close reader %v", closeErr)
		}
	}()

	csvFile, err := parser.NewCSVFile(reader, "calendar.txt")
	if err != nil {
		return services
	}

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		service := v.parseService(row)
		if service != nil {
			services[service.ServiceID] = service
		}
	}

	return services
}

// parseService parses service information from calendar.txt
func (v *DateTripsValidator) parseService(row *parser.CSVRow) *ServiceInfo {
	serviceID, hasServiceID := row.Values["service_id"]
	if !hasServiceID {
		return nil
	}

	service := &ServiceInfo{
		ServiceID: strings.TrimSpace(serviceID),
		RowNumber: row.RowNumber,
	}

	// Parse dates
	if startDateStr, hasStart := row.Values["start_date"]; hasStart {
		if startDate := v.parseGTFSDate(strings.TrimSpace(startDateStr)); startDate != nil {
			service.StartDate = startDate
		}
	}
	if endDateStr, hasEnd := row.Values["end_date"]; hasEnd {
		if endDate := v.parseGTFSDate(strings.TrimSpace(endDateStr)); endDate != nil {
			service.EndDate = endDate
		}
	}

	// Parse days of week
	daysFields := []string{"monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday"}
	for i, field := range daysFields {
		if dayValue, hasDay := row.Values[field]; hasDay && strings.TrimSpace(dayValue) == "1" {
			service.DaysOfWeek[i] = true
		}
	}

	return service
}

// loadCalendarExceptions loads exceptions from calendar_dates.txt
func (v *DateTripsValidator) loadCalendarExceptions(loader *parser.FeedLoader) []CalendarException {
	var exceptions []CalendarException

	reader, err := loader.GetFile("calendar_dates.txt")
	if err != nil {
		return exceptions
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close reader %v", closeErr)
		}
	}()

	csvFile, err := parser.NewCSVFile(reader, "calendar_dates.txt")
	if err != nil {
		return exceptions
	}

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		exception := v.parseCalendarException(row)
		if exception != nil {
			exceptions = append(exceptions, *exception)
		}
	}

	return exceptions
}

// parseCalendarException parses calendar exception from calendar_dates.txt
func (v *DateTripsValidator) parseCalendarException(row *parser.CSVRow) *CalendarException {
	serviceID, hasServiceID := row.Values["service_id"]
	dateStr, hasDate := row.Values["date"]
	exceptionTypeStr, hasException := row.Values["exception_type"]

	if !hasServiceID || !hasDate || !hasException {
		return nil
	}

	date := v.parseGTFSDate(strings.TrimSpace(dateStr))
	if date == nil {
		return nil
	}

	exceptionType, err := strconv.Atoi(strings.TrimSpace(exceptionTypeStr))
	if err != nil {
		return nil
	}

	return &CalendarException{
		ServiceID:     strings.TrimSpace(serviceID),
		Date:          *date,
		ExceptionType: exceptionType,
	}
}

// countTripsByService counts the trips of every service in one pass.
func (v *DateTripsValidator) countTripsByService(loader *parser.FeedLoader) map[string]int {
	counts := make(map[string]int)

	reader, err := loader.GetFile("trips.txt")
	if err != nil {
		return counts
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close reader %v", closeErr)
		}
	}()

	csvFile, err := parser.NewCSVFile(reader, "trips.txt")
	if err != nil {
		return counts
	}

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}
		if serviceID, hasServiceID := row.Values["service_id"]; hasServiceID {
			counts[strings.TrimSpace(serviceID)]++
		}
	}

	return counts
}

// parseGTFSDate parses GTFS date format (YYYYMMDD)
func (v *DateTripsValidator) parseGTFSDate(dateStr string) *time.Time {
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

// startOfUTCDay reduces an instant to the midnight-UTC that identifies its
// calendar day, which is how service dates are represented throughout this
// validator.
func startOfUTCDay(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

// formatGTFSDate formats time as GTFS date
func (v *DateTripsValidator) formatGTFSDate(date time.Time) string {
	return date.Format("20060102")
}

// validateNext7DaysService reports a feed whose trips do not cover the coming
// week. Canonical counts the feed as covered while a significant share of its
// trips still run, so a day only counts when trips are scheduled on it.
func (v *DateTripsValidator) validateNext7DaysService(container *notice.NoticeContainer, serviceDates []time.Time, currentDate time.Time) {
	covered := make(map[int64]bool, len(serviceDates))
	for _, date := range serviceDates {
		covered[date.Unix()] = true
	}

	lastCoveredDay := -1
	for i := 0; i < 7; i++ {
		if covered[currentDate.AddDate(0, 0, i).Unix()] {
			lastCoveredDay = i
		}
	}

	if lastCoveredDay == 6 {
		return // Covered through the whole week.
	}

	container.AddNotice(notice.NewTripCoverageNotActiveForNext7DaysNotice(
		v.formatGTFSDate(currentDate),
		v.formatGTFSDate(currentDate.AddDate(0, 0, lastCoveredDay)),
	))
}

// validateServiceGaps reports every run of more than 13 days inside the service
// window on which the feed schedules nothing.
func (v *DateTripsValidator) validateServiceGaps(container *notice.NoticeContainer, serviceDates []time.Time) {
	for i := 1; i < len(serviceDates); i++ {
		gapDays := int(serviceDates[i].Sub(serviceDates[i-1]).Hours() / 24)
		if gapDays <= maxServiceGapDays {
			continue
		}

		container.AddNotice(notice.NewBigGapInServiceNotice(
			v.formatGTFSDate(serviceDates[i-1]),
			v.formatGTFSDate(serviceDates[i]),
			gapDays,
		))
	}
}

// validateAgainstFeedPeriod compares the dates the feed actually runs trips on
// with the validity period it declares in feed_info.txt. The two describe the
// same thing and should agree in both directions: service outside the declared
// period is dropped by consumers that honour it, and a declared period reaching
// far past the last day of service promises coverage that is not there.
func (v *DateTripsValidator) validateAgainstFeedPeriod(loader *parser.FeedLoader, container *notice.NoticeContainer, serviceDates []time.Time) {
	if len(serviceDates) == 0 {
		return
	}

	feedStart, feedEnd, rowNumber := v.loadFeedPeriod(loader)
	if feedStart == nil || feedEnd == nil {
		// A feed_info.txt without dates is reported as missing_feed_info_date.
		return
	}

	windowStart := serviceDates[0]
	windowEnd := serviceDates[len(serviceDates)-1]

	if windowStart.Before(*feedStart) || windowEnd.After(*feedEnd) {
		container.AddNotice(notice.NewServiceWindowOutsideFeedPeriodNotice(
			rowNumber,
			v.formatGTFSDate(*feedStart),
			v.formatGTFSDate(*feedEnd),
			v.formatGTFSDate(windowStart),
			v.formatGTFSDate(windowEnd),
		))
	}

	if feedEnd.After(windowEnd.AddDate(0, 0, feedValidityGraceDays)) {
		container.AddNotice(notice.NewFeedValidBeyondTotalServiceWindowNotice(
			rowNumber,
			v.formatGTFSDate(*feedEnd),
			v.formatGTFSDate(windowEnd),
			int(feedEnd.Sub(windowEnd).Hours()/24),
		))
	}
}

// loadFeedPeriod reads the validity period declared by the first row of
// feed_info.txt. Rows past the first are reported as multiple_feed_info_entries.
func (v *DateTripsValidator) loadFeedPeriod(loader *parser.FeedLoader) (start *time.Time, end *time.Time, rowNumber int) {
	reader, err := loader.GetFile("feed_info.txt")
	if err != nil {
		return nil, nil, 0
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close reader %v", closeErr)
		}
	}()

	csvFile, err := parser.NewCSVFile(reader, "feed_info.txt")
	if err != nil {
		return nil, nil, 0
	}

	row, err := csvFile.ReadRow()
	if err != nil {
		return nil, nil, 0
	}

	return v.parseGTFSDate(strings.TrimSpace(row.Values["feed_start_date"])),
		v.parseGTFSDate(strings.TrimSpace(row.Values["feed_end_date"])),
		row.RowNumber
}
