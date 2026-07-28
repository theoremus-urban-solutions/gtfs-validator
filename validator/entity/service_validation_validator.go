package entity

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

// CalendarDateService represents what calendar_dates.txt says about a service.
// LastAddedDate is the latest date the service is added on, which is what
// decides whether the service has expired; removals do not extend a service.
type CalendarDateService struct {
	ServiceID     string
	LastAddedDate *time.Time
	RowNumber     int
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
				ServiceID: serviceIDTrimmed,
				RowNumber: row.RowNumber,
			}
			services[serviceIDTrimmed] = service
		}

		exceptionType, hasExceptionType := row.Values["exception_type"]
		if !hasExceptionType || strings.TrimSpace(exceptionType) != "1" {
			continue
		}

		date, err := v.parseGTFSDate(strings.TrimSpace(row.Values["date"]))
		if err != nil {
			continue
		}
		if service.LastAddedDate == nil || date.After(*service.LastAddedDate) {
			service.LastAddedDate = date
			service.RowNumber = row.RowNumber
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
	report := func(serviceID string, lastActive time.Time, rowNumber int) {
		switch {
		case lastActive.Before(currentDate):
			container.AddNotice(notice.NewExpiredCalendarNotice(
				rowNumber,
				serviceID,
				lastActive.Format("20060102"),
				currentDate.Format("20060102"),
			))
		case lastActive.After(currentDate.AddDate(farFutureServiceYears, 0, 0)):
			container.AddNotice(notice.NewServiceExtendsFarInTheFutureNotice(
				rowNumber,
				serviceID,
				lastActive.Format("20060102"),
				currentDate.Format("20060102"),
			))
		}
	}

	for serviceID, service := range calendarServices {
		lastActive, err := v.parseGTFSDate(service.EndDate)
		if err != nil {
			continue // Invalid or absent date - other validators handle this
		}
		rowNumber := service.RowNumber

		if added, exists := calendarDateServices[serviceID]; exists && added.LastAddedDate != nil && added.LastAddedDate.After(*lastActive) {
			lastActive = added.LastAddedDate
			rowNumber = added.RowNumber
		}

		report(serviceID, *lastActive, rowNumber)
	}

	// Services that exist only in calendar_dates.txt run until their last added
	// date; ones that are only ever removed have no active date at all.
	for serviceID, service := range calendarDateServices {
		if _, inCalendar := calendarServices[serviceID]; inCalendar {
			continue
		}
		if service.LastAddedDate == nil {
			continue
		}

		report(serviceID, *service.LastAddedDate, service.RowNumber)
	}
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
