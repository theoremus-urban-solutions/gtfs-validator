package entity

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

// ServiceValidationValidator validates service definitions and dates
type ServiceValidationValidator struct{}

// NewServiceValidationValidator creates a new service validation validator
func NewServiceValidationValidator() *ServiceValidationValidator {
	return &ServiceValidationValidator{}
}

// ServiceInfo represents information about a calendar.txt service
type ServiceInfo struct {
	ServiceID string
	StartDate string
	EndDate   string
	Days      map[string]bool
	RowNumber int
}

// CalendarDateService is what calendar_dates.txt says about one service: the
// dates it puts service on and the dates it takes service away. Which of them
// the service last runs on depends on its calendar.txt row as well, so that is
// settled in lastActiveDate rather than here.
//
// RowNumber is the service's first row in the file, which is the row a notice
// about the service as a whole points at.
type CalendarDateService struct {
	ServiceID    string
	AddedDates   []time.Time
	RemovedDates map[int64]bool // keyed by Unix second, as the dates are
	RowNumber    int
}

// Validate checks service definitions for consistency
func (v *ServiceValidationValidator) Validate(loader *parser.FeedLoader, container *notice.NoticeContainer, config validator.Config) {
	currentDate, ok := config.CurrentDate.(time.Time)
	if !ok {
		currentDate = time.Now()
	}

	calendarServices := v.loadCalendarServices(loader)
	calendarDateServices := v.loadCalendarDateServices(loader)

	v.validateActiveDays(container, calendarServices)
	v.validateServiceDates(container, calendarServices, calendarDateServices, currentDate)
	v.validateServiceUsage(loader, container, calendarServices, calendarDateServices)
}

// farFutureServiceYears is how far ahead a service may end before its dates
// stop being a plan and start being a placeholder.
const farFutureServiceYears = 2

// loadCalendarServices loads services from calendar.txt
func (v *ServiceValidationValidator) loadCalendarServices(loader *parser.FeedLoader) map[string]*ServiceInfo {
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
			break
		}

		serviceID, hasServiceID := row.Values["service_id"]
		if !hasServiceID {
			continue
		}

		serviceIDTrimmed := strings.TrimSpace(serviceID)

		service := &ServiceInfo{
			ServiceID: serviceIDTrimmed,
			Days:      make(map[string]bool),
			RowNumber: row.RowNumber,
		}

		// Load start and end dates
		if startDate, hasStartDate := row.Values["start_date"]; hasStartDate {
			service.StartDate = strings.TrimSpace(startDate)
		}
		if endDate, hasEndDate := row.Values["end_date"]; hasEndDate {
			service.EndDate = strings.TrimSpace(endDate)
		}

		// Load service days
		dayFields := []string{"monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday"}
		for _, dayField := range dayFields {
			if dayValue, hasDayValue := row.Values[dayField]; hasDayValue {
				service.Days[dayField] = strings.TrimSpace(dayValue) == "1"
			}
		}

		services[serviceIDTrimmed] = service
	}

	return services
}

// loadCalendarDateServices loads services from calendar_dates.txt
func (v *ServiceValidationValidator) loadCalendarDateServices(loader *parser.FeedLoader) map[string]*CalendarDateService {
	services := make(map[string]*CalendarDateService)

	reader, err := loader.GetFile("calendar_dates.txt")
	if err != nil {
		return services
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close reader %v", closeErr)
		}
	}()

	csvFile, err := parser.NewCSVFile(reader, "calendar_dates.txt")
	if err != nil {
		return services
	}

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}

		serviceID, hasServiceID := row.Values["service_id"]
		if !hasServiceID {
			continue
		}
		serviceIDTrimmed := strings.TrimSpace(serviceID)

		service, exists := services[serviceIDTrimmed]
		if !exists {
			service = &CalendarDateService{
				ServiceID:    serviceIDTrimmed,
				RemovedDates: make(map[int64]bool),
				RowNumber:    row.RowNumber,
			}
			services[serviceIDTrimmed] = service
		}

		date, err := v.parseGTFSDate(strings.TrimSpace(row.Values["date"]))
		if err != nil {
			continue // Reported as invalid_date by the type layer
		}
		switch strings.TrimSpace(row.Values["exception_type"]) {
		case "1":
			service.AddedDates = append(service.AddedDates, *date)
		case "2":
			service.RemovedDates[date.Unix()] = true
		}
	}

	return services
}

// validateActiveDays reports a calendar.txt service that runs on no day of the
// week. Its date window is then irrelevant — the service never runs.
func (v *ServiceValidationValidator) validateActiveDays(container *notice.NoticeContainer, services map[string]*ServiceInfo) {
	for _, service := range services {
		hasActiveDay := false
		for _, isActive := range service.Days {
			if isActive {
				hasActiveDay = true
				break
			}
		}

		if !hasActiveDay {
			container.AddNotice(notice.NewServiceWithoutActiveDaysNotice(
				service.ServiceID,
				service.RowNumber,
			))
		}
	}
}

// validateServiceDates reports every service whose last active date sits
// outside the span worth planning on — already in the past, or so far ahead
// that nobody has checked the schedule that far.
//
// A calendar_dates.txt addition after the calendar.txt end_date keeps a service
// alive, so both files decide the last active date together.
func (v *ServiceValidationValidator) validateServiceDates(container *notice.NoticeContainer, calendarServices map[string]*ServiceInfo, calendarDateServices map[string]*CalendarDateService, currentDate time.Time) {
	type expiredService struct {
		serviceID string
		rowNumber int
	}
	// Expired services with no calendar.txt row behind them. Whether they are
	// worth reporting cannot be decided one at a time; see the flush below.
	var expiredOutsideCalendar []expiredService
	everyServiceExpired := true

	report := func(serviceID string, lastActive time.Time, rowNumber int, inCalendar bool) {
		if lastActive.Before(currentDate) {
			if inCalendar {
				container.AddNotice(notice.NewExpiredCalendarNotice(rowNumber, serviceID))
			} else {
				expiredOutsideCalendar = append(expiredOutsideCalendar, expiredService{serviceID, rowNumber})
			}
			return
		}

		everyServiceExpired = false
		if lastActive.After(currentDate.AddDate(farFutureServiceYears, 0, 0)) {
			container.AddNotice(notice.NewServiceExtendsFarInTheFutureNotice(
				rowNumber,
				serviceID,
				lastActive.Format("20060102"),
				currentDate.Format("20060102"),
			))
		}
	}

	for serviceID, service := range calendarServices {
		lastActive, ok := v.lastActiveDate(service, calendarDateServices[serviceID])
		if !ok {
			// The service never runs at all, which is reported as
			// service_has_no_active_day_of_the_week or, for dates that do not
			// parse, by the type layer.
			continue
		}
		report(serviceID, lastActive, service.RowNumber, true)
	}

	for serviceID, dates := range calendarDateServices {
		if _, inCalendar := calendarServices[serviceID]; inCalendar {
			continue
		}
		lastActive, ok := v.lastActiveDate(nil, dates)
		if !ok {
			// A service only ever removed has no active date to expire.
			continue
		}
		report(serviceID, lastActive, dates.RowNumber, false)
	}

	// The rule is about date ranges that have run out, and a service defined
	// only in calendar_dates.txt has no range — just a set of dates. Such a
	// service is held aside and reported only when calendar.txt is empty AND
	// every service in the feed has expired.
	//
	// The condition exists because a feed with no calendar.txt at all is a
	// different situation from one whose calendars have run out, and only the
	// second is what the rule is about. A feed that keeps its whole schedule in
	// calendar_dates.txt leaves dates behind it as it goes, and reporting each
	// of those would bury the case worth knowing about: a dataset where nothing
	// runs any more. With calendar.txt present, a service it never mentions is
	// a dangling reference, which foreign_key_violation covers.
	//
	// Do not restore per-service reporting here as missing coverage.
	if len(calendarServices) > 0 || !everyServiceExpired {
		return
	}
	// File order, so the notices do not shuffle between runs.
	sort.Slice(expiredOutsideCalendar, func(i, j int) bool {
		return expiredOutsideCalendar[i].rowNumber < expiredOutsideCalendar[j].rowNumber
	})
	for _, expired := range expiredOutsideCalendar {
		container.AddNotice(notice.NewExpiredCalendarNotice(expired.rowNumber, expired.serviceID))
	}
}

// dayFieldsByWeekday names the calendar.txt column that says whether a service
// runs on a given date.
var dayFieldsByWeekday = map[time.Weekday]string{
	time.Monday:    "monday",
	time.Tuesday:   "tuesday",
	time.Wednesday: "wednesday",
	time.Thursday:  "thursday",
	time.Friday:    "friday",
	time.Saturday:  "saturday",
	time.Sunday:    "sunday",
}

// lastActiveDate returns the last date a service really runs on, which is what
// decides whether it has expired.
//
// This is not the calendar.txt end_date. A calendar ending on a Tuesday but
// running only on Sundays last ran the Sunday before, and calendar_dates.txt
// has a say in both directions: an addition pushes the date out past end_date,
// and a removal on the last day pulls it back in.
func (v *ServiceValidationValidator) lastActiveDate(service *ServiceInfo, dates *CalendarDateService) (time.Time, bool) {
	var candidates []time.Time
	var removed map[int64]bool
	if dates != nil {
		candidates = append(candidates, dates.AddedDates...)
		removed = dates.RemovedDates
	}

	// One calendar date per removal, plus one, is enough to outlast them all:
	// every candidate the scan below skips costs a removal of its own.
	candidates = append(candidates, v.lastCalendarDates(service, len(removed)+1)...)

	sort.Slice(candidates, func(i, j int) bool { return candidates[i].After(candidates[j]) })
	for _, date := range candidates {
		if !removed[date.Unix()] {
			return date, true
		}
	}
	return time.Time{}, false
}

// lastCalendarDates returns up to limit of the dates a calendar.txt row runs
// on, latest first. A service active on at least one day of the week reaches
// each of them within a week of walking, so the limit bounds the work even
// when an end_date typo like 29991231 puts the row's span in the millennia.
func (v *ServiceValidationValidator) lastCalendarDates(service *ServiceInfo, limit int) []time.Time {
	if service == nil {
		return nil
	}

	runsOnSomeDay := false
	for _, isActive := range service.Days {
		if isActive {
			runsOnSomeDay = true
			break
		}
	}
	if !runsOnSomeDay {
		// Without this the walk below would cross the whole span to find
		// nothing. validateActiveDays reports the row.
		return nil
	}

	start, startErr := v.parseGTFSDate(service.StartDate)
	end, endErr := v.parseGTFSDate(service.EndDate)
	if startErr != nil || endErr != nil {
		// Invalid or absent dates are reported by the type layer.
		return nil
	}

	// An end_date before the start_date describes no days at all, so the walk
	// below would find nothing and the service would drop out of every
	// date-based rule — silently passing a calendar that has plainly expired.
	// The inverted range is reported on its own as
	// start_and_end_range_out_of_order; for the purpose of "when does this
	// service last run", the date the row nominates as its end is still the
	// answer, and canonical reads it the same way.
	if end.Before(*start) {
		return []time.Time{*end}
	}

	var dates []time.Time
	for date := *end; !date.Before(*start) && len(dates) < limit; date = date.AddDate(0, 0, -1) {
		if service.Days[dayFieldsByWeekday[date.Weekday()]] {
			dates = append(dates, date)
		}
	}
	return dates
}

// validateServiceUsage checks if services are actually used by trips
func (v *ServiceValidationValidator) validateServiceUsage(loader *parser.FeedLoader, container *notice.NoticeContainer, calendarServices map[string]*ServiceInfo, calendarDateServices map[string]*CalendarDateService) {
	// Load services used by trips
	usedServices := v.loadUsedServices(loader)

	// Check for unused calendar services
	for serviceID, service := range calendarServices {
		if !usedServices[serviceID] {
			container.AddNotice(notice.NewUnusedServiceNotice(
				serviceID,
				"calendar.txt",
				service.RowNumber,
			))
		}
	}

	// Check for unused calendar_dates services
	for serviceID, service := range calendarDateServices {
		if _, inCalendar := calendarServices[serviceID]; inCalendar {
			continue
		}
		if !usedServices[serviceID] {
			container.AddNotice(notice.NewUnusedServiceNotice(
				serviceID,
				"calendar_dates.txt",
				service.RowNumber,
			))
		}
	}
}

// loadUsedServices loads service IDs referenced by trips
func (v *ServiceValidationValidator) loadUsedServices(loader *parser.FeedLoader) map[string]bool {
	usedServices := make(map[string]bool)

	reader, err := loader.GetFile("trips.txt")
	if err != nil {
		return usedServices
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close reader %v", closeErr)
		}
	}()

	csvFile, err := parser.NewCSVFile(reader, "trips.txt")
	if err != nil {
		return usedServices
	}

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}

		serviceID, hasServiceID := row.Values["service_id"]
		if hasServiceID {
			usedServices[strings.TrimSpace(serviceID)] = true
		}
	}

	return usedServices
}

// parseGTFSDate parses a GTFS date string (YYYYMMDD) into time.Time
func (v *ServiceValidationValidator) parseGTFSDate(dateStr string) (*time.Time, error) {
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
