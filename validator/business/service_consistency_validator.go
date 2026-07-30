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

// ServiceConsistencyValidator validates service consistency across calendar and trips
type ServiceConsistencyValidator struct{}

// NewServiceConsistencyValidator creates a new service consistency validator
func NewServiceConsistencyValidator() *ServiceConsistencyValidator {
	return &ServiceConsistencyValidator{}
}

// ServiceDefinition represents a service definition from calendar.txt
type ServiceDefinition struct {
	ServiceID  string
	StartDate  string
	EndDate    string
	DaysActive []string
	RowNumber  int
}

// ServiceException represents an exception from calendar_dates.txt
type ServiceException struct {
	ServiceID     string
	Date          string
	ExceptionType int
	RowNumber     int
}

// Validate checks service consistency across calendar and calendar_dates
func (v *ServiceConsistencyValidator) Validate(loader *parser.FeedLoader, container *notice.NoticeContainer, config validator.Config) {
	services := v.loadServices(loader)
	exceptions := v.loadServiceExceptions(loader)

	currentDate, ok := config.CurrentDate.(time.Time)
	if !ok {
		currentDate = time.Now()
	}

	v.validateServiceExceptions(container, exceptions)
	v.validateFeedStartsInFuture(container, services, exceptions, currentDate)
}

// loadServices loads service definitions from calendar.txt
func (v *ServiceConsistencyValidator) loadServices(loader *parser.FeedLoader) map[string]*ServiceDefinition {
	services := make(map[string]*ServiceDefinition)

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

		service := v.parseServiceDefinition(row)
		if service != nil {
			services[service.ServiceID] = service
		}
	}

	return services
}

// parseServiceDefinition parses a service definition from calendar.txt
func (v *ServiceConsistencyValidator) parseServiceDefinition(row *parser.CSVRow) *ServiceDefinition {
	serviceID, hasServiceID := row.Values["service_id"]
	if !hasServiceID {
		return nil
	}

	service := &ServiceDefinition{
		ServiceID:  strings.TrimSpace(serviceID),
		RowNumber:  row.RowNumber,
		DaysActive: []string{},
	}

	// Parse date fields
	if startDate, hasStartDate := row.Values["start_date"]; hasStartDate {
		service.StartDate = strings.TrimSpace(startDate)
	}
	if endDate, hasEndDate := row.Values["end_date"]; hasEndDate {
		service.EndDate = strings.TrimSpace(endDate)
	}

	// Parse day fields
	dayFields := []string{"monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday"}
	for _, day := range dayFields {
		if dayValue, hasDay := row.Values[day]; hasDay && strings.TrimSpace(dayValue) == "1" {
			service.DaysActive = append(service.DaysActive, day)
		}
	}

	return service
}

// loadServiceExceptions loads service exceptions from calendar_dates.txt
func (v *ServiceConsistencyValidator) loadServiceExceptions(loader *parser.FeedLoader) map[string][]*ServiceException {
	exceptions := make(map[string][]*ServiceException)

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

		exception := v.parseServiceException(row)
		if exception != nil {
			exceptions[exception.ServiceID] = append(exceptions[exception.ServiceID], exception)
		}
	}

	return exceptions
}

// parseServiceException parses a service exception from calendar_dates.txt
func (v *ServiceConsistencyValidator) parseServiceException(row *parser.CSVRow) *ServiceException {
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

	return &ServiceException{
		ServiceID:     strings.TrimSpace(serviceID),
		Date:          strings.TrimSpace(date),
		ExceptionType: exceptionType,
		RowNumber:     row.RowNumber,
	}
}

// validateServiceExceptions reports a service given two exceptions for the same
// date. Which of the two wins is undefined, so the schedule that day depends on
// the consumer.
func (v *ServiceConsistencyValidator) validateServiceExceptions(container *notice.NoticeContainer, exceptions map[string][]*ServiceException) {
	for serviceID, serviceExceptions := range exceptions {
		// Sort exceptions by date
		sort.Slice(serviceExceptions, func(i, j int) bool {
			return serviceExceptions[i].Date < serviceExceptions[j].Date
		})

		// Check for duplicate dates
		dateMap := make(map[string]*ServiceException)
		for _, exception := range serviceExceptions {
			if existing, exists := dateMap[exception.Date]; exists {
				if existing.ExceptionType != exception.ExceptionType {
					container.AddNotice(notice.NewConflictingCalendarExceptionNotice(
						serviceID,
						exception.Date,
						existing.RowNumber,
						exception.RowNumber,
					))
				} else {
					container.AddNotice(notice.NewDuplicateCalendarExceptionNotice(
						serviceID,
						exception.Date,
						existing.RowNumber,
						exception.RowNumber,
					))
				}
			} else {
				dateMap[exception.Date] = exception
			}
		}
	}
}

// validateFeedStartsInFuture reports a feed in which no service has begun yet.
// The earliest date any service can run is the earliest calendar.txt start_date
// or added calendar_dates.txt date; if even that is still ahead, the feed
// schedules nothing for today.
func (v *ServiceConsistencyValidator) validateFeedStartsInFuture(container *notice.NoticeContainer, services map[string]*ServiceDefinition, exceptions map[string][]*ServiceException, currentDate time.Time) {
	var earliest *time.Time

	consider := func(dateStr string) {
		date, err := time.Parse("20060102", dateStr)
		if err != nil {
			return
		}
		if earliest == nil || date.Before(*earliest) {
			earliest = &date
		}
	}

	for _, service := range services {
		// A service with no active weekday never runs regardless of its window,
		// so it cannot be what covers today.
		if len(service.DaysActive) > 0 {
			consider(service.StartDate)
		}
	}
	for _, serviceExceptions := range exceptions {
		for _, exception := range serviceExceptions {
			if exception.ExceptionType == 1 {
				consider(exception.Date)
			}
		}
	}

	// No parseable service date at all: a feed without calendars is reported as
	// missing_calendar_and_calendar_date_files, and unparseable dates as
	// invalid_date.
	if earliest == nil || !earliest.After(currentDate) {
		return
	}

	container.AddNotice(notice.NewFutureCalendarNotice(
		currentDate.Format("20060102"),
		earliest.Format("20060102"),
	))
}
