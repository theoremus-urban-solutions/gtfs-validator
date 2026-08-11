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
	"github.com/theoremus-urban-solutions/gtfs-validator/types"
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

// The trip coverage rule is about the window over which the feed runs a
// significant number of its trips, not every date it runs anything at all: one
// summer-only route must not stretch the window across a year in which nothing
// else moves. Canonical settles that with two ratios over the daily trip counts.
//
// The busiest day is a poor yardstick — a single event day would raise the bar
// for every other date — so the yardstick is a high percentile instead: sort the
// daily counts and take the one at maxServiceDateTripCountRatio of the way up.
// On a feed whose window is short but whose calendars run for years, that
// percentile lands in the quiet tail, so the index is pulled to at least
// maxServiceDateTripCountLimit days from the top.
const (
	maxServiceDateTripCountRatio = 0.90
	maxServiceDateTripCountLimit = 30
	majorityTripCountRatio       = 0.75
)

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

	exceptionsByService := groupExceptionsByService(exceptions)

	// Counting trips per service in one pass over trips.txt, rather than
	// re-reading the file for each service, keeps this linear in the feed size
	// instead of quadratic.
	tripCounts := v.countTripsByService(loader)
	calendar := v.buildServiceCalendar(services, exceptionsByService, tripCounts)

	v.validateNext7DaysService(container, calendar, currentDate)
	v.validateServiceGaps(container, calendar.Dates)
	v.validateAgainstFeedPeriod(loader, container, calendar)
}

// groupExceptionsByService collects calendar_dates.txt rows under the service
// they belong to, which is how every check below wants them.
func groupExceptionsByService(exceptions []CalendarException) map[string][]CalendarException {
	byService := make(map[string][]CalendarException)
	for _, exception := range exceptions {
		byService[exception.ServiceID] = append(byService[exception.ServiceID], exception)
	}
	return byService
}

// serviceIDsOf returns every service either file defines, in a stable order so
// the notices come out the same way on every run.
func serviceIDsOf(services map[string]*ServiceInfo, exceptionsByService map[string][]CalendarException) []string {
	seen := make(map[string]bool, len(services)+len(exceptionsByService))
	for serviceID := range services {
		seen[serviceID] = true
	}
	for serviceID := range exceptionsByService {
		seen[serviceID] = true
	}

	ids := make([]string, 0, len(seen))
	for serviceID := range seen {
		ids = append(ids, serviceID)
	}
	sort.Strings(ids)
	return ids
}

// ServiceCalendar is the feed's schedule reduced to what the date rules need:
// every date it runs at least one trip on, in order, how many trips run on each
// of them, and the span each service covers on its own.
type ServiceCalendar struct {
	Dates      []time.Time
	TripCounts map[int64]int // keyed by Unix second, as the dates are
	Windows    map[string]ServiceWindow
}

// ServiceWindow is the first and last date one service is active on.
type ServiceWindow struct {
	Start time.Time
	End   time.Time
}

// buildServiceCalendar spreads each service's trips over the dates that service
// runs on. A service no trip references contributes nothing however wide its
// window, so those are left out and the result describes the feed's real
// service rather than what its calendars claim.
//
// Each service's own window is taken here rather than in the rule that wants
// it, because this is the one place that walks a service's dates: a feed whose
// schedule lives entirely in calendar_dates.txt has hundreds of thousands of
// them, and once is enough.
func (v *DateTripsValidator) buildServiceCalendar(services map[string]*ServiceInfo, exceptionsByService map[string][]CalendarException, tripCounts map[string]int) ServiceCalendar {
	active := make(map[int64]time.Time)
	counts := make(map[int64]int)
	windows := make(map[string]ServiceWindow)
	for _, serviceID := range serviceIDsOf(services, exceptionsByService) {
		tripCount := tripCounts[serviceID]
		if tripCount == 0 {
			continue
		}

		var window ServiceWindow
		for key, date := range v.datesForService(services[serviceID], exceptionsByService[serviceID]) {
			active[key] = date
			counts[key] += tripCount

			if window.Start.IsZero() || date.Before(window.Start) {
				window.Start = date
			}
			if date.After(window.End) {
				window.End = date
			}
		}
		if !window.Start.IsZero() {
			windows[serviceID] = window
		}
	}

	dates := make([]time.Time, 0, len(active))
	for _, date := range active {
		dates = append(dates, date)
	}
	sort.Slice(dates, func(i, j int) bool { return dates[i].Before(dates[j]) })

	return ServiceCalendar{Dates: dates, TripCounts: counts, Windows: windows}
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

// countTripsByService counts the trips of every service in one pass. A
// frequency-based trip stands for as many vehicles as its headway fits into the
// span it covers, and is counted for all of them: one row in trips.txt can be
// most of a day's service.
func (v *DateTripsValidator) countTripsByService(loader *parser.FeedLoader) map[string]int {
	counts := make(map[string]int)
	frequencyCounts := v.countTripsByFrequency(loader)

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
		serviceID, hasServiceID := row.Values["service_id"]
		if !hasServiceID {
			continue
		}
		tripCount, isFrequencyBased := frequencyCounts[strings.TrimSpace(row.Values["trip_id"])]
		if !isFrequencyBased {
			tripCount = 1
		}
		counts[strings.TrimSpace(serviceID)] += tripCount
	}

	return counts
}

// countTripsByFrequency counts how many vehicles each frequency-based trip
// stands for. Trips absent from the result run once, as written.
func (v *DateTripsValidator) countTripsByFrequency(loader *parser.FeedLoader) map[string]int {
	counts := make(map[string]int)

	reader, err := loader.GetFile("frequencies.txt")
	if err != nil {
		return counts
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close reader %v", closeErr)
		}
	}()

	csvFile, err := parser.NewCSVFile(reader, "frequencies.txt")
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

		tripID := strings.TrimSpace(row.Values["trip_id"])
		if tripID == "" {
			continue
		}

		// One vehicle runs whatever the headway says, so the row is worth at
		// least one trip; the rest of the span is worth one per headway.
		counts[tripID]++
		startTime, startErr := types.ParseGTFSTime(strings.TrimSpace(row.Values["start_time"]))
		endTime, endErr := types.ParseGTFSTime(strings.TrimSpace(row.Values["end_time"]))
		headway, headwayErr := strconv.Atoi(strings.TrimSpace(row.Values["headway_secs"]))
		if startErr != nil || endErr != nil || headwayErr != nil || headway <= 0 ||
			endTime.ToSeconds() <= startTime.ToSeconds() {
			// The field-level rules report a malformed row; here it is worth
			// the one trip already counted.
			continue
		}
		counts[tripID] += (endTime.ToSeconds() - startTime.ToSeconds() - 1) / headway
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

// validateNext7DaysService reports a feed whose majority service window does
// not enclose the coming week. The window itself is the subject: individual
// days inside it with nothing running are ordinary — a weekday-only feed has
// two of them every week — and are left to big_gap_in_service.
func (v *DateTripsValidator) validateNext7DaysService(container *notice.NoticeContainer, calendar ServiceCalendar, currentDate time.Time) {
	windowStart, windowEnd, ok := v.majorityServiceWindow(calendar)
	if !ok {
		// No date runs a trip at all. The calendars that lead here are reported
		// as missing_calendar_and_calendar_date_files or
		// service_has_no_active_day_of_the_week.
		return
	}

	if !windowStart.After(currentDate) && !windowEnd.Before(currentDate.AddDate(0, 0, 7)) {
		return
	}

	container.AddNotice(notice.NewTripCoverageNotActiveForNext7DaysNotice(
		v.formatGTFSDate(currentDate),
		v.formatGTFSDate(windowStart),
		v.formatGTFSDate(windowEnd),
	))
}

// majorityServiceWindow returns the first and last date on which the feed runs
// a majority share of the trips it runs on a typical day. See the ratios above
// for how "typical" and "majority" are pinned down.
func (v *DateTripsValidator) majorityServiceWindow(calendar ServiceCalendar) (start time.Time, end time.Time, ok bool) {
	if len(calendar.Dates) == 0 {
		return time.Time{}, time.Time{}, false
	}

	sortedCounts := make([]int, 0, len(calendar.Dates))
	for _, date := range calendar.Dates {
		sortedCounts = append(sortedCounts, calendar.TripCounts[date.Unix()])
	}
	sort.Ints(sortedCounts)

	typicalIndex := max(
		int(float64(len(sortedCounts))*maxServiceDateTripCountRatio),
		len(sortedCounts)-maxServiceDateTripCountLimit,
	)
	if typicalIndex < 0 {
		typicalIndex = 0
	}
	threshold := int(majorityTripCountRatio * float64(sortedCounts[typicalIndex]))

	// The whole span is the fallback, though the day the yardstick came from
	// always clears the threshold, so both loops do find a date.
	start, end = calendar.Dates[0], calendar.Dates[len(calendar.Dates)-1]
	for _, date := range calendar.Dates {
		if calendar.TripCounts[date.Unix()] >= threshold {
			start = date
			break
		}
	}
	for i := len(calendar.Dates) - 1; i >= 0; i-- {
		if calendar.TripCounts[calendar.Dates[i].Unix()] >= threshold {
			end = calendar.Dates[i]
			break
		}
	}

	return start, end, true
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

// validateAgainstFeedPeriod compares the dates the feed runs service on with
// the validity period it declares in feed_info.txt. The two describe the same
// thing and should agree in both directions: service outside the declared
// period is dropped by consumers that honour it, and a declared period reaching
// far past the last day of service promises coverage that is not there.
func (v *DateTripsValidator) validateAgainstFeedPeriod(loader *parser.FeedLoader, container *notice.NoticeContainer, calendar ServiceCalendar) {
	feedStart, feedEnd := v.loadFeedPeriod(loader)
	if feedStart == nil || feedEnd == nil {
		// A feed_info.txt without dates is reported as missing_feed_info_date.
		return
	}

	v.validateServiceWindows(container, calendar.Windows, *feedStart, *feedEnd)

	if len(calendar.Dates) == 0 {
		return
	}
	// Both ends count. A period that starts well before the first day of
	// service overstates coverage exactly as much as one that runs past the
	// last, and testing only the end silently accepted the former. One notice
	// covers the period whichever side is at fault.
	windowStart := calendar.Dates[0]
	windowEnd := calendar.Dates[len(calendar.Dates)-1]
	startsTooEarly := feedStart.Before(windowStart.AddDate(0, 0, -feedValidityGraceDays))
	endsTooLate := feedEnd.After(windowEnd.AddDate(0, 0, feedValidityGraceDays))
	if startsTooEarly || endsTooLate {
		container.AddNotice(notice.NewFeedValidBeyondTotalServiceWindowNotice(
			v.formatGTFSDate(*feedStart),
			v.formatGTFSDate(*feedEnd),
			v.formatGTFSDate(windowStart),
			v.formatGTFSDate(windowEnd),
		))
	}
}

// validateServiceWindows reports each service whose own active dates reach
// outside the declared feed period. Per service rather than once for the feed:
// a single feed-wide span says only that something somewhere is out of period,
// while the service id says which calendar to go and fix.
//
// Only services trips reference are in windows, which is what the rule wants: a
// calendar nothing runs on puts no service outside the period, and its own
// defect is reported as unused_service.
func (v *DateTripsValidator) validateServiceWindows(container *notice.NoticeContainer, windows map[string]ServiceWindow, feedStart time.Time, feedEnd time.Time) {
	serviceIDs := make([]string, 0, len(windows))
	for serviceID := range windows {
		serviceIDs = append(serviceIDs, serviceID)
	}
	sort.Strings(serviceIDs)

	for _, serviceID := range serviceIDs {
		window := windows[serviceID]

		daysBeforeFeedStart, daysAfterFeedEnd := 0, 0
		if window.Start.Before(feedStart) {
			daysBeforeFeedStart = daysBetween(window.Start, feedStart)
		}
		if window.End.After(feedEnd) {
			daysAfterFeedEnd = daysBetween(feedEnd, window.End)
		}
		if daysBeforeFeedStart == 0 && daysAfterFeedEnd == 0 {
			continue
		}

		container.AddNotice(notice.NewServiceWindowOutsideFeedPeriodNotice(
			serviceID,
			v.formatGTFSDate(window.Start),
			v.formatGTFSDate(window.End),
			daysBeforeFeedStart,
			daysAfterFeedEnd,
		))
	}
}

// daysBetween counts whole days from the earlier date to the later one. Both
// are midnight UTC, so no daylight-saving hour can round the division off.
func daysBetween(from time.Time, to time.Time) int {
	return int(to.Sub(from).Hours() / 24)
}

// loadFeedPeriod reads the validity period declared by the first row of
// feed_info.txt. Rows past the first are reported as multiple_feed_info_entries.
func (v *DateTripsValidator) loadFeedPeriod(loader *parser.FeedLoader) (start *time.Time, end *time.Time) {
	reader, err := loader.GetFile("feed_info.txt")
	if err != nil {
		return nil, nil
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close reader %v", closeErr)
		}
	}()

	csvFile, err := parser.NewCSVFile(reader, "feed_info.txt")
	if err != nil {
		return nil, nil
	}

	row, err := csvFile.ReadRow()
	if err != nil {
		return nil, nil
	}

	return v.parseGTFSDate(strings.TrimSpace(row.Values["feed_start_date"])),
		v.parseGTFSDate(strings.TrimSpace(row.Values["feed_end_date"]))
}
