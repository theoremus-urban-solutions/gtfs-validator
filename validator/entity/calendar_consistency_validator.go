package entity

import (
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/parser"
	"github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

// CalendarConsistencyValidator validates calendar and calendar_dates consistency
type CalendarConsistencyValidator struct{}

// NewCalendarConsistencyValidator creates a new calendar consistency validator
func NewCalendarConsistencyValidator() *CalendarConsistencyValidator {
	return &CalendarConsistencyValidator{}
}

// CalendarService represents a service defined in calendar.txt
type CalendarService struct {
	ServiceID string
	Monday    bool
	Tuesday   bool
	Wednesday bool
	Thursday  bool
	Friday    bool
	Saturday  bool
	Sunday    bool
	StartDate string
	EndDate   string
	RowNumber int
}

// CalendarDate represents a service exception in calendar_dates.txt
type CalendarDate struct {
	ServiceID     string
	Date          string
	ExceptionType int
	RowNumber     int
}

// Validate checks calendar and calendar_dates consistency
func (v *CalendarConsistencyValidator) Validate(loader *parser.FeedLoader, container *notice.NoticeContainer, config validator.Config) {
	// Load calendar services
	calendarServices := v.loadCalendarServices(loader)

	// Load calendar dates
	calendarDates := v.loadCalendarDates(loader)

	// Cast CurrentDate to time.Time
	currentDate, ok := config.CurrentDate.(time.Time)
	if !ok {
		currentDate = time.Now()
	}

	// Validate calendar services
	for _, service := range calendarServices {
		v.validateCalendarService(container, service, currentDate)
	}

	// Validate calendar dates
	for _, calDate := range calendarDates {
		v.validateCalendarDate(container, calDate, currentDate)
	}

	// Check for services without any definition
	v.validateServiceDefinitions(loader, container, calendarServices, calendarDates)

	// Validate overlapping exceptions
	v.validateCalendarExceptions(container, calendarDates)
}

// loadCalendarServices loads services from calendar.txt
func (v *CalendarConsistencyValidator) loadCalendarServices(loader *parser.FeedLoader) map[string]*CalendarService {
	services := make(map[string]*CalendarService)

	reader, err := loader.GetFile("calendar.txt")
	if err != nil {
		return services
	}
	defer reader.Close()

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

		service := v.parseCalendarService(row)
		if service != nil {
			services[service.ServiceID] = service
		}
	}

	return services
}

// parseCalendarService parses a calendar service record
func (v *CalendarConsistencyValidator) parseCalendarService(row *parser.CSVRow) *CalendarService {
	serviceID, hasServiceID := row.Values["service_id"]
	if !hasServiceID {
		return nil
	}

	service := &CalendarService{
		ServiceID: strings.TrimSpace(serviceID),
		RowNumber: row.RowNumber,
	}

	// Parse day fields
	if monday, hasMonday := row.Values["monday"]; hasMonday {
		service.Monday = strings.TrimSpace(monday) == "1"
	}
	if tuesday, hasTuesday := row.Values["tuesday"]; hasTuesday {
		service.Tuesday = strings.TrimSpace(tuesday) == "1"
	}
	if wednesday, hasWednesday := row.Values["wednesday"]; hasWednesday {
		service.Wednesday = strings.TrimSpace(wednesday) == "1"
	}
	if thursday, hasThursday := row.Values["thursday"]; hasThursday {
		service.Thursday = strings.TrimSpace(thursday) == "1"
	}
	if friday, hasFriday := row.Values["friday"]; hasFriday {
		service.Friday = strings.TrimSpace(friday) == "1"
	}
	if saturday, hasSaturday := row.Values["saturday"]; hasSaturday {
		service.Saturday = strings.TrimSpace(saturday) == "1"
	}
	if sunday, hasSunday := row.Values["sunday"]; hasSunday {
		service.Sunday = strings.TrimSpace(sunday) == "1"
	}

	// Parse date fields
	if startDate, hasStartDate := row.Values["start_date"]; hasStartDate {
		service.StartDate = strings.TrimSpace(startDate)
	}
	if endDate, hasEndDate := row.Values["end_date"]; hasEndDate {
		service.EndDate = strings.TrimSpace(endDate)
	}

	return service
}

// loadCalendarDates loads exceptions from calendar_dates.txt
func (v *CalendarConsistencyValidator) loadCalendarDates(loader *parser.FeedLoader) []*CalendarDate {
	var calendarDates []*CalendarDate

	reader, err := loader.GetFile("calendar_dates.txt")
	if err != nil {
		return calendarDates
	}
	defer reader.Close()

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
			break
		}

		calDate := v.parseCalendarDate(row)
		if calDate != nil {
			calendarDates = append(calendarDates, calDate)
		}
	}

	return calendarDates
}

// parseCalendarDate parses a calendar date record
func (v *CalendarConsistencyValidator) parseCalendarDate(row *parser.CSVRow) *CalendarDate {
	serviceID, hasServiceID := row.Values["service_id"]
	date, hasDate := row.Values["date"]
	exceptionTypeStr, hasExceptionType := row.Values["exception_type"]

	if !hasServiceID || !hasDate || !hasExceptionType {
		return nil
	}

	exceptionType, err := strconv.Atoi(strings.TrimSpace(exceptionTypeStr))
	if err != nil {
		return nil
	}

	return &CalendarDate{
		ServiceID:     strings.TrimSpace(serviceID),
		Date:          strings.TrimSpace(date),
		ExceptionType: exceptionType,
		RowNumber:     row.RowNumber,
	}
}

// validateCalendarService validates a calendar service
func (v *CalendarConsistencyValidator) validateCalendarService(container *notice.NoticeContainer, service *CalendarService, currentDate time.Time) {
	// Check if service runs on any day
	hasAnyDay := service.Monday || service.Tuesday || service.Wednesday ||
		service.Thursday || service.Friday || service.Saturday || service.Sunday

	if !hasAnyDay {
		container.AddNotice(notice.NewServiceNeverActiveNotice(
			service.ServiceID,
			service.RowNumber,
		))
	}

	// Validate date range
	if service.StartDate != "" && service.EndDate != "" {
		startDate, startErr := time.Parse("20060102", service.StartDate)
		endDate, endErr := time.Parse("20060102", service.EndDate)

		if startErr == nil && endErr == nil {
			if endDate.Before(startDate) {
				container.AddNotice(notice.NewInvalidServiceDateRangeNotice(
					service.ServiceID,
					service.StartDate,
					service.EndDate,
					service.RowNumber,
				))
			}

			// Check if service period is in the past
			if endDate.Before(currentDate.AddDate(0, 0, -30)) {
				container.AddNotice(notice.NewExpiredServiceNotice(
					service.ServiceID,
					service.EndDate,
					service.RowNumber,
				))
			}

			// Check if service period is too far in the future
			if startDate.After(currentDate.AddDate(2, 0, 0)) {
				container.AddNotice(notice.NewFutureServiceNotice(
					service.ServiceID,
					service.StartDate,
					service.RowNumber,
				))
			}
		}
	}
}

// validateCalendarDate validates a calendar date exception
func (v *CalendarConsistencyValidator) validateCalendarDate(container *notice.NoticeContainer, calDate *CalendarDate, currentDate time.Time) {
	// Validate exception type
	if calDate.ExceptionType != 1 && calDate.ExceptionType != 2 {
		container.AddNotice(notice.NewInvalidExceptionTypeNotice(
			calDate.ServiceID,
			calDate.Date,
			calDate.ExceptionType,
			calDate.RowNumber,
		))
	}

	// Validate date format and reasonableness
	if date, err := time.Parse("20060102", calDate.Date); err == nil {
		// Check if date is too far in the past
		if date.Before(currentDate.AddDate(-5, 0, 0)) {
			container.AddNotice(notice.NewVeryOldCalendarDateNotice(
				calDate.ServiceID,
				calDate.Date,
				calDate.RowNumber,
			))
		}

		// Check if date is too far in the future
		if date.After(currentDate.AddDate(5, 0, 0)) {
			container.AddNotice(notice.NewVeryFutureCalendarDateNotice(
				calDate.ServiceID,
				calDate.Date,
				calDate.RowNumber,
			))
		}
	}
}

// validateServiceDefinitions checks that all used services are defined
func (v *CalendarConsistencyValidator) validateServiceDefinitions(loader *parser.FeedLoader, container *notice.NoticeContainer, calendarServices map[string]*CalendarService, calendarDates []*CalendarDate) {
	// Collect all service IDs used in trips.txt
	usedServices := v.getUsedServiceIDs(loader)

	// Create set of defined services
	definedServices := make(map[string]bool)
	for serviceID := range calendarServices {
		definedServices[serviceID] = true
	}
	for _, calDate := range calendarDates {
		definedServices[calDate.ServiceID] = true
	}

	// Check for undefined services
	for serviceID := range usedServices {
		if !definedServices[serviceID] {
			container.AddNotice(notice.NewUndefinedServiceNotice(serviceID))
		}
	}

	// Check for unused services
	for serviceID := range definedServices {
		if !usedServices[serviceID] {
			container.AddNotice(notice.NewUnusedServiceNotice(serviceID, "calendar.txt", 0))
		}
	}
}

// getUsedServiceIDs loads service IDs used in trips.txt
func (v *CalendarConsistencyValidator) getUsedServiceIDs(loader *parser.FeedLoader) map[string]bool {
	usedServices := make(map[string]bool)

	reader, err := loader.GetFile("trips.txt")
	if err != nil {
		return usedServices
	}
	defer reader.Close()

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

		if serviceID, hasServiceID := row.Values["service_id"]; hasServiceID {
			usedServices[strings.TrimSpace(serviceID)] = true
		}
	}

	return usedServices
}

// validateCalendarExceptions checks for overlapping calendar exceptions
func (v *CalendarConsistencyValidator) validateCalendarExceptions(container *notice.NoticeContainer, calendarDates []*CalendarDate) {
	// Group exceptions by service and date
	exceptions := make(map[string]map[string][]*CalendarDate)

	for _, calDate := range calendarDates {
		if exceptions[calDate.ServiceID] == nil {
			exceptions[calDate.ServiceID] = make(map[string][]*CalendarDate)
		}
		exceptions[calDate.ServiceID][calDate.Date] = append(exceptions[calDate.ServiceID][calDate.Date], calDate)
	}

	// Check for duplicate exceptions
	for serviceID, serviceExceptions := range exceptions {
		for date, dateExceptions := range serviceExceptions {
			if len(dateExceptions) > 1 {
				// Check if they have different exception types
				firstType := dateExceptions[0].ExceptionType
				for i := 1; i < len(dateExceptions); i++ {
					if dateExceptions[i].ExceptionType != firstType {
						container.AddNotice(notice.NewConflictingCalendarExceptionNotice(
							serviceID,
							date,
							dateExceptions[0].RowNumber,
							dateExceptions[i].RowNumber,
						))
					} else {
						container.AddNotice(notice.NewDuplicateCalendarExceptionNotice(
							serviceID,
							date,
							dateExceptions[0].RowNumber,
							dateExceptions[i].RowNumber,
						))
					}
				}
			}
		}
	}
}
