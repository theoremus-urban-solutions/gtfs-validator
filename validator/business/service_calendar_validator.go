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

// ServiceCalendarValidator performs advanced calendar and service validation
type ServiceCalendarValidator struct{}

// NewServiceCalendarValidator creates a new service calendar validator
func NewServiceCalendarValidator() *ServiceCalendarValidator {
	return &ServiceCalendarValidator{}
}

// CalendarInfo represents calendar information
type CalendarInfo struct {
	ServiceID string
	Monday    bool
	Tuesday   bool
	Wednesday bool
	Thursday  bool
	Friday    bool
	Saturday  bool
	Sunday    bool
	StartDate time.Time
	EndDate   time.Time
	RowNumber int
}

// CalendarDateInfo represents calendar date information
type CalendarDateInfo struct {
	ServiceID     string
	Date          time.Time
	ExceptionType int
	RowNumber     int
}

// Validate performs advanced calendar and service validation
func (v *ServiceCalendarValidator) Validate(loader *parser.FeedLoader, container *notice.NoticeContainer, config validator.Config) {
	calendars := v.loadCalendars(loader)
	calendarDates := v.loadCalendarDates(loader)

	// Get all service IDs from both sources
	allServiceIDs := v.getAllServiceIDs(calendars, calendarDates)

	// Assert current date type
	currentDate, ok := config.CurrentDate.(time.Time)
	if !ok {
		currentDate = time.Now()
	}

	for serviceID := range allServiceIDs {
		v.validateService(container, serviceID, calendars[serviceID], calendarDates[serviceID], currentDate)
	}

	// Validate calendar patterns
	v.validateCalendarPatterns(container, calendars, calendarDates)
}

// loadCalendars loads calendar information from calendar.txt
func (v *ServiceCalendarValidator) loadCalendars(loader *parser.FeedLoader) map[string]*CalendarInfo {
	calendars := make(map[string]*CalendarInfo)

	reader, err := loader.GetFile("calendar.txt")
	if err != nil {
		return calendars
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close reader %v", closeErr)
		}
	}()

	csvFile, err := parser.NewCSVFile(reader, "calendar.txt")
	if err != nil {
		return calendars
	}

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		calendar := v.parseCalendar(row)
		if calendar != nil {
			calendars[calendar.ServiceID] = calendar
		}
	}

	return calendars
}

// parseCalendar parses calendar information
func (v *ServiceCalendarValidator) parseCalendar(row *parser.CSVRow) *CalendarInfo {
	serviceID, hasServiceID := row.Values["service_id"]
	if !hasServiceID {
		return nil
	}

	calendar := &CalendarInfo{
		ServiceID: strings.TrimSpace(serviceID),
		RowNumber: row.RowNumber,
	}

	// Parse day fields
	if monday, hasMonday := row.Values["monday"]; hasMonday {
		calendar.Monday = strings.TrimSpace(monday) == "1"
	}
	if tuesday, hasTuesday := row.Values["tuesday"]; hasTuesday {
		calendar.Tuesday = strings.TrimSpace(tuesday) == "1"
	}
	if wednesday, hasWednesday := row.Values["wednesday"]; hasWednesday {
		calendar.Wednesday = strings.TrimSpace(wednesday) == "1"
	}
	if thursday, hasThursday := row.Values["thursday"]; hasThursday {
		calendar.Thursday = strings.TrimSpace(thursday) == "1"
	}
	if friday, hasFriday := row.Values["friday"]; hasFriday {
		calendar.Friday = strings.TrimSpace(friday) == "1"
	}
	if saturday, hasSaturday := row.Values["saturday"]; hasSaturday {
		calendar.Saturday = strings.TrimSpace(saturday) == "1"
	}
	if sunday, hasSunday := row.Values["sunday"]; hasSunday {
		calendar.Sunday = strings.TrimSpace(sunday) == "1"
	}

	// Parse dates
	if startDateStr, hasStartDate := row.Values["start_date"]; hasStartDate {
		if startDate, err := v.parseGTFSDate(strings.TrimSpace(startDateStr)); err == nil {
			calendar.StartDate = startDate
		}
	}
	if endDateStr, hasEndDate := row.Values["end_date"]; hasEndDate {
		if endDate, err := v.parseGTFSDate(strings.TrimSpace(endDateStr)); err == nil {
			calendar.EndDate = endDate
		}
	}

	return calendar
}

// loadCalendarDates loads calendar dates information from calendar_dates.txt
func (v *ServiceCalendarValidator) loadCalendarDates(loader *parser.FeedLoader) map[string][]*CalendarDateInfo {
	calendarDates := make(map[string][]*CalendarDateInfo)

	reader, err := loader.GetFile("calendar_dates.txt")
	if err != nil {
		return calendarDates
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close reader %v", closeErr)
		}
	}()

	csvFile, err := parser.NewCSVFile(reader, "calendar_dates.txt")
	if err != nil {
		return calendarDates
	}

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		calendarDate := v.parseCalendarDate(row)
		if calendarDate != nil {
			calendarDates[calendarDate.ServiceID] = append(calendarDates[calendarDate.ServiceID], calendarDate)
		}
	}

	return calendarDates
}

// parseCalendarDate parses calendar date information
func (v *ServiceCalendarValidator) parseCalendarDate(row *parser.CSVRow) *CalendarDateInfo {
	serviceID, hasServiceID := row.Values["service_id"]
	dateStr, hasDate := row.Values["date"]
	exceptionTypeStr, hasExceptionType := row.Values["exception_type"]

	if !hasServiceID || !hasDate || !hasExceptionType {
		return nil
	}

	date, err := v.parseGTFSDate(strings.TrimSpace(dateStr))
	if err != nil {
		return nil
	}

	exceptionType, err := strconv.Atoi(strings.TrimSpace(exceptionTypeStr))
	if err != nil {
		return nil
	}

	return &CalendarDateInfo{
		ServiceID:     strings.TrimSpace(serviceID),
		Date:          date,
		ExceptionType: exceptionType,
		RowNumber:     row.RowNumber,
	}
}

// getAllServiceIDs gets all unique service IDs from both calendar sources
func (v *ServiceCalendarValidator) getAllServiceIDs(calendars map[string]*CalendarInfo, calendarDates map[string][]*CalendarDateInfo) map[string]bool {
	serviceIDs := make(map[string]bool)

	for serviceID := range calendars {
		serviceIDs[serviceID] = true
	}

	for serviceID := range calendarDates {
		serviceIDs[serviceID] = true
	}

	return serviceIDs
}

// validateService validates a single service
func (v *ServiceCalendarValidator) validateService(container *notice.NoticeContainer, serviceID string, calendar *CalendarInfo, calendarDates []*CalendarDateInfo, currentDate time.Time) {
	hasCalendar := calendar != nil
	hasCalendarDates := len(calendarDates) > 0

	// Check for service without any definition
	if !hasCalendar && !hasCalendarDates {
		container.AddNotice(notice.NewServiceWithoutDefinitionNotice(
			serviceID,
		))
		return
	}

	// Validate calendar if present
	if hasCalendar {
		v.validateCalendar(container, calendar, currentDate)
	}

	// Validate calendar dates if present
	if hasCalendarDates {
		v.validateCalendarDates(container, serviceID, calendarDates, currentDate)
	}

	// Check service activity
	v.validateServiceActivity(container, serviceID, calendar, calendarDates, currentDate)
}

// validateCalendar validates calendar information
func (v *ServiceCalendarValidator) validateCalendar(container *notice.NoticeContainer, calendar *CalendarInfo, currentDate time.Time) {
	// Check if no days are selected
	if !v.hasAnyDaySelected(calendar) {
		container.AddNotice(notice.NewCalendarNoDaysSelectedNotice(
			calendar.ServiceID,
			calendar.RowNumber,
		))
	}

	// Check date range validity
	if !calendar.StartDate.IsZero() && !calendar.EndDate.IsZero() {
		if calendar.EndDate.Before(calendar.StartDate) {
			container.AddNotice(notice.NewCalendarEndBeforeStartNotice(
				calendar.ServiceID,
				calendar.StartDate.Format("20060102"),
				calendar.EndDate.Format("20060102"),
				calendar.RowNumber,
			))
		}

		// Check for very long service periods
		duration := calendar.EndDate.Sub(calendar.StartDate)
		if duration > 365*24*time.Hour*2 { // More than 2 years
			container.AddNotice(notice.NewVeryLongServicePeriodNotice(
				calendar.ServiceID,
				calendar.StartDate.Format("20060102"),
				calendar.EndDate.Format("20060102"),
				int(duration.Hours()/24),
				calendar.RowNumber,
			))
		}

		// Check for expired services (ended more than 30 days ago)
		if calendar.EndDate.Before(currentDate.AddDate(0, 0, -30)) {
			container.AddNotice(notice.NewExpiredServiceNotice(
				calendar.ServiceID,
				calendar.EndDate.Format("20060102"),
				calendar.RowNumber,
			))
		}

		// Check for services starting in the far future (more than 1 year)
		if calendar.StartDate.After(currentDate.AddDate(1, 0, 0)) {
			container.AddNotice(notice.NewFutureServiceNotice(
				calendar.ServiceID,
				calendar.StartDate.Format("20060102"),
				calendar.RowNumber,
			))
		}
	}

	// Check for unusual day patterns
	v.validateDayPatterns(container, calendar)
}

// validateCalendarDates validates calendar dates
func (v *ServiceCalendarValidator) validateCalendarDates(container *notice.NoticeContainer, serviceID string, calendarDates []*CalendarDateInfo, currentDate time.Time) {
	datesSeen := make(map[string]*CalendarDateInfo)

	for _, calendarDate := range calendarDates {
		dateStr := calendarDate.Date.Format("20060102")

		// Check for duplicate dates
		if existing, exists := datesSeen[dateStr]; exists {
			container.AddNotice(notice.NewDuplicateCalendarDateNotice(
				serviceID,
				dateStr,
				existing.RowNumber,
				calendarDate.RowNumber,
			))
		}
		datesSeen[dateStr] = calendarDate

		// Validate exception type
		if calendarDate.ExceptionType != 1 && calendarDate.ExceptionType != 2 {
			container.AddNotice(notice.NewInvalidExceptionTypeNotice(
				serviceID,
				dateStr,
				calendarDate.ExceptionType,
				calendarDate.RowNumber,
			))
		}

		// Check for very old dates
		if calendarDate.Date.Before(currentDate.AddDate(-2, 0, 0)) {
			container.AddNotice(notice.NewVeryOldCalendarDateNotice(
				serviceID,
				dateStr,
				calendarDate.RowNumber,
			))
		}

		// Check for dates in far future
		if calendarDate.Date.After(currentDate.AddDate(2, 0, 0)) {
			container.AddNotice(notice.NewVeryFutureCalendarDateNotice(
				serviceID,
				dateStr,
				calendarDate.RowNumber,
			))
		}
	}
}

// validateServiceActivity validates overall service activity
func (v *ServiceCalendarValidator) validateServiceActivity(container *notice.NoticeContainer, serviceID string, calendar *CalendarInfo, calendarDates []*CalendarDateInfo, currentDate time.Time) {
	// Calculate service days in current month
	activeDays := v.calculateActiveDaysInMonth(calendar, calendarDates, currentDate)

	if activeDays == 0 {
		container.AddNotice(notice.NewInactiveServiceCurrentMonthNotice(
			serviceID,
		))
	} else if activeDays < 5 && calendar != nil { // Less than 5 days per month for regular service
		container.AddNotice(notice.NewLowFrequencyServiceNotice(
			serviceID,
			activeDays,
		))
	}
}

// validateDayPatterns validates unusual day patterns
func (v *ServiceCalendarValidator) validateDayPatterns(container *notice.NoticeContainer, calendar *CalendarInfo) {
	// Check for weekend-only service
	if !calendar.Monday && !calendar.Tuesday && !calendar.Wednesday && !calendar.Thursday && !calendar.Friday &&
		(calendar.Saturday || calendar.Sunday) {
		container.AddNotice(notice.NewWeekendOnlyServiceNotice(
			calendar.ServiceID,
			calendar.RowNumber,
		))
	}

	// Check for single day service
	dayCount := 0
	if calendar.Monday {
		dayCount++
	}
	if calendar.Tuesday {
		dayCount++
	}
	if calendar.Wednesday {
		dayCount++
	}
	if calendar.Thursday {
		dayCount++
	}
	if calendar.Friday {
		dayCount++
	}
	if calendar.Saturday {
		dayCount++
	}
	if calendar.Sunday {
		dayCount++
	}

	if dayCount == 1 {
		container.AddNotice(notice.NewSingleDayServiceNotice(
			calendar.ServiceID,
			v.getSingleDayName(calendar),
			calendar.RowNumber,
		))
	}

	// Check for unusual weekday patterns (e.g., Mon-Wed-Fri only)
	if v.hasUnusualWeekdayPattern(calendar) {
		container.AddNotice(notice.NewUnusualServicePatternNotice(
			calendar.ServiceID,
			v.getDayPattern(calendar),
			calendar.RowNumber,
		))
	}
}

// validateCalendarPatterns validates patterns across all calendars
func (v *ServiceCalendarValidator) validateCalendarPatterns(container *notice.NoticeContainer, calendars map[string]*CalendarInfo, calendarDates map[string][]*CalendarDateInfo) {
	// Check for calendar-only vs calendar_dates-only services
	calendarOnlyCount := 0
	calendarDatesOnlyCount := 0
	bothCount := 0

	allServiceIDs := v.getAllServiceIDs(calendars, calendarDates)

	for serviceID := range allServiceIDs {
		hasCalendar := calendars[serviceID] != nil
		hasCalendarDates := len(calendarDates[serviceID]) > 0

		if hasCalendar && hasCalendarDates {
			bothCount++
		} else if hasCalendar {
			calendarOnlyCount++
		} else if hasCalendarDates {
			calendarDatesOnlyCount++
		}
	}

	// Report unusual distribution patterns
	totalServices := len(allServiceIDs)
	if totalServices > 5 {
		if calendarDatesOnlyCount > totalServices/2 {
			container.AddNotice(notice.NewMostlyCalendarDatesServicesNotice(
				calendarDatesOnlyCount,
				totalServices,
			))
		}
	}
}

// Helper functions

// hasAnyDaySelected checks if any day is selected in calendar
func (v *ServiceCalendarValidator) hasAnyDaySelected(calendar *CalendarInfo) bool {
	return calendar.Monday || calendar.Tuesday || calendar.Wednesday || calendar.Thursday ||
		calendar.Friday || calendar.Saturday || calendar.Sunday
}

// calculateActiveDaysInMonth calculates active service days in current month
func (v *ServiceCalendarValidator) calculateActiveDaysInMonth(calendar *CalendarInfo, calendarDates []*CalendarDateInfo, currentDate time.Time) int {
	activeDays := 0

	// Get current month range
	year, month, _ := currentDate.Date()
	monthStart := time.Date(year, month, 1, 0, 0, 0, 0, currentDate.Location())
	monthEnd := monthStart.AddDate(0, 1, -1)

	// Check each day in current month
	for d := monthStart; !d.After(monthEnd); d = d.AddDate(0, 0, 1) {
		isActive := false

		// Check calendar service
		if calendar != nil && !calendar.StartDate.IsZero() && !calendar.EndDate.IsZero() {
			if !d.Before(calendar.StartDate) && !d.After(calendar.EndDate) {
				weekday := d.Weekday()
				switch weekday {
				case time.Monday:
					isActive = calendar.Monday
				case time.Tuesday:
					isActive = calendar.Tuesday
				case time.Wednesday:
					isActive = calendar.Wednesday
				case time.Thursday:
					isActive = calendar.Thursday
				case time.Friday:
					isActive = calendar.Friday
				case time.Saturday:
					isActive = calendar.Saturday
				case time.Sunday:
					isActive = calendar.Sunday
				}
			}
		}

		// Check calendar_dates exceptions
		for _, calendarDate := range calendarDates {
			if calendarDate.Date.Equal(d) {
				switch calendarDate.ExceptionType {
				case 1:
					isActive = true // Service added
				case 2:
					isActive = false // Service removed
				}
				break
			}
		}

		if isActive {
			activeDays++
		}
	}

	return activeDays
}

// getSingleDayName returns the name of the single active day
func (v *ServiceCalendarValidator) getSingleDayName(calendar *CalendarInfo) string {
	if calendar.Monday {
		return "Monday"
	}
	if calendar.Tuesday {
		return "Tuesday"
	}
	if calendar.Wednesday {
		return "Wednesday"
	}
	if calendar.Thursday {
		return "Thursday"
	}
	if calendar.Friday {
		return "Friday"
	}
	if calendar.Saturday {
		return "Saturday"
	}
	if calendar.Sunday {
		return "Sunday"
	}
	return "Unknown"
}

// getDayPattern returns a string describing the day pattern
func (v *ServiceCalendarValidator) getDayPattern(calendar *CalendarInfo) string {
	var days []string
	if calendar.Monday {
		days = append(days, "Mon")
	}
	if calendar.Tuesday {
		days = append(days, "Tue")
	}
	if calendar.Wednesday {
		days = append(days, "Wed")
	}
	if calendar.Thursday {
		days = append(days, "Thu")
	}
	if calendar.Friday {
		days = append(days, "Fri")
	}
	if calendar.Saturday {
		days = append(days, "Sat")
	}
	if calendar.Sunday {
		days = append(days, "Sun")
	}
	return strings.Join(days, ",")
}

// hasUnusualWeekdayPattern checks for unusual weekday patterns
func (v *ServiceCalendarValidator) hasUnusualWeekdayPattern(calendar *CalendarInfo) bool {
	// Pattern: alternating weekdays (Mon-Wed-Fri or Tue-Thu)
	if calendar.Monday && !calendar.Tuesday && calendar.Wednesday && !calendar.Thursday && calendar.Friday {
		return true // Mon-Wed-Fri
	}
	if !calendar.Monday && calendar.Tuesday && !calendar.Wednesday && calendar.Thursday && !calendar.Friday {
		return true // Tue-Thu
	}

	return false
}

// parseGTFSDate parses GTFS date format (YYYYMMDD)
func (v *ServiceCalendarValidator) parseGTFSDate(dateStr string) (time.Time, error) {
	return time.Parse("20060102", dateStr)
}
