package business

import (
	"io"
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
	TripCount  int
	RowNumber  int
}

// ServiceException represents an exception from calendar_dates.txt
type ServiceException struct {
	ServiceID     string
	Date          string
	ExceptionType int
	RowNumber     int
}

// TripService represents a trip's service usage
type TripService struct {
	TripID    string
	ServiceID string
	RouteID   string
	RowNumber int
}

// Validate checks service consistency across calendar, calendar_dates, and trips
func (v *ServiceConsistencyValidator) Validate(loader *parser.FeedLoader, container *notice.NoticeContainer, config validator.Config) {
	// Load services from calendar.txt
	services := v.loadServices(loader)

	// Load exceptions from calendar_dates.txt
	exceptions := v.loadServiceExceptions(loader)

	// Load trip services from trips.txt
	tripServices := v.loadTripServices(loader)

	// Cast CurrentDate to time.Time
	currentDate, ok := config.CurrentDate.(time.Time)
	if !ok {
		currentDate = time.Now()
	}

	// Validate service definitions
	v.validateServiceDefinitions(container, services, currentDate)

	// Validate service exceptions
	v.validateServiceExceptions(container, exceptions, currentDate)

	// Validate service usage consistency
	v.validateServiceUsage(container, services, exceptions, tripServices)

	// Validate service patterns
	v.validateServicePatterns(container, services, tripServices)
}

// loadServices loads service definitions from calendar.txt
func (v *ServiceConsistencyValidator) loadServices(loader *parser.FeedLoader) map[string]*ServiceDefinition {
	services := make(map[string]*ServiceDefinition)

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
	defer reader.Close()

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

// loadTripServices loads trip service assignments from trips.txt
func (v *ServiceConsistencyValidator) loadTripServices(loader *parser.FeedLoader) []*TripService {
	var tripServices []*TripService

	reader, err := loader.GetFile("trips.txt")
	if err != nil {
		return tripServices
	}
	defer reader.Close()

	csvFile, err := parser.NewCSVFile(reader, "trips.txt")
	if err != nil {
		return tripServices
	}

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		tripService := v.parseTripService(row)
		if tripService != nil {
			tripServices = append(tripServices, tripService)
		}
	}

	return tripServices
}

// parseTripService parses a trip service assignment from trips.txt
func (v *ServiceConsistencyValidator) parseTripService(row *parser.CSVRow) *TripService {
	tripID, hasTripID := row.Values["trip_id"]
	serviceID, hasServiceID := row.Values["service_id"]
	routeID, hasRouteID := row.Values["route_id"]

	if !hasTripID || !hasServiceID || !hasRouteID {
		return nil
	}

	return &TripService{
		TripID:    strings.TrimSpace(tripID),
		ServiceID: strings.TrimSpace(serviceID),
		RouteID:   strings.TrimSpace(routeID),
		RowNumber: row.RowNumber,
	}
}

// validateServiceDefinitions validates individual service definitions
func (v *ServiceConsistencyValidator) validateServiceDefinitions(container *notice.NoticeContainer, services map[string]*ServiceDefinition, currentDate time.Time) {
	for _, service := range services {
		// Check if service has active days
		if len(service.DaysActive) == 0 {
			container.AddNotice(notice.NewServiceNeverActiveNotice(
				service.ServiceID,
				service.RowNumber,
			))
		}

		// Check date range validity
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

				// Check if service is too old
				if endDate.Before(currentDate.AddDate(0, 0, -90)) {
					container.AddNotice(notice.NewVeryOldServiceNotice(
						service.ServiceID,
						service.EndDate,
						service.RowNumber,
					))
				}

				// Check if service is too far in the future
				if startDate.After(currentDate.AddDate(2, 0, 0)) {
					container.AddNotice(notice.NewVeryFutureServiceNotice(
						service.ServiceID,
						service.StartDate,
						service.RowNumber,
					))
				}
			}
		}
	}
}

// validateServiceExceptions validates service exceptions
func (v *ServiceConsistencyValidator) validateServiceExceptions(container *notice.NoticeContainer, exceptions map[string][]*ServiceException, currentDate time.Time) {
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

			// Check if exception date is reasonable
			if date, err := time.Parse("20060102", exception.Date); err == nil {
				if date.Before(currentDate.AddDate(-5, 0, 0)) {
					container.AddNotice(notice.NewVeryOldCalendarDateNotice(
						serviceID,
						exception.Date,
						exception.RowNumber,
					))
				}
				if date.After(currentDate.AddDate(5, 0, 0)) {
					container.AddNotice(notice.NewVeryFutureCalendarDateNotice(
						serviceID,
						exception.Date,
						exception.RowNumber,
					))
				}
			}
		}
	}
}

// validateServiceUsage validates service usage consistency
func (v *ServiceConsistencyValidator) validateServiceUsage(container *notice.NoticeContainer, services map[string]*ServiceDefinition, exceptions map[string][]*ServiceException, tripServices []*TripService) {
	// Create service usage map
	serviceUsage := make(map[string]int)
	for _, tripService := range tripServices {
		serviceUsage[tripService.ServiceID]++
	}

	// Update trip counts in service definitions
	for serviceID, count := range serviceUsage {
		if service, exists := services[serviceID]; exists {
			service.TripCount = count
		}
	}

	// Check for services defined but not used
	for serviceID, service := range services {
		if service.TripCount == 0 {
			container.AddNotice(notice.NewUnusedServiceNotice(
				serviceID,
				"calendar.txt",
				service.RowNumber,
			))
		}
	}

	// Check for services used but not defined
	for _, tripService := range tripServices {
		if _, definedInCalendar := services[tripService.ServiceID]; !definedInCalendar {
			if _, definedInExceptions := exceptions[tripService.ServiceID]; !definedInExceptions {
				container.AddNotice(notice.NewUndefinedServiceNotice(tripService.ServiceID))
			}
		}
	}

	// Check for services with very few trips (potential data issues)
	for serviceID, service := range services {
		if service.TripCount > 0 && service.TripCount <= 2 {
			container.AddNotice(notice.NewLowServiceUsageNotice(
				serviceID,
				service.TripCount,
				service.RowNumber,
			))
		}
	}
}

// validateServicePatterns validates service patterns for operational efficiency
func (v *ServiceConsistencyValidator) validateServicePatterns(container *notice.NoticeContainer, services map[string]*ServiceDefinition, tripServices []*TripService) {
	// Group trips by route and service
	routeServiceMap := make(map[string]map[string]int)

	for _, tripService := range tripServices {
		if routeServiceMap[tripService.RouteID] == nil {
			routeServiceMap[tripService.RouteID] = make(map[string]int)
		}
		routeServiceMap[tripService.RouteID][tripService.ServiceID]++
	}

	// Check for routes with too many different services
	for routeID, serviceMap := range routeServiceMap {
		serviceCount := len(serviceMap)
		if serviceCount > 10 {
			container.AddNotice(notice.NewExcessiveServiceVarietyNotice(
				routeID,
				serviceCount,
			))
		}

		// Check for services with very few trips on a route
		for serviceID, tripCount := range serviceMap {
			if tripCount == 1 {
				container.AddNotice(notice.NewSingleTripServiceNotice(
					routeID,
					serviceID,
					tripCount,
				))
			}
		}
	}

	// Analyze service day patterns
	weekdayServices := 0
	weekendServices := 0
	mixedServices := 0

	for _, service := range services {
		if service.TripCount == 0 {
			continue
		}

		hasWeekdays := false
		hasWeekends := false

		for _, day := range service.DaysActive {
			if day == "saturday" || day == "sunday" {
				hasWeekends = true
			} else {
				hasWeekdays = true
			}
		}

		if hasWeekdays && hasWeekends {
			mixedServices++
		} else if hasWeekdays {
			weekdayServices++
		} else if hasWeekends {
			weekendServices++
		}
	}

	// Report service pattern summary
	container.AddNotice(notice.NewServicePatternSummaryNotice(
		weekdayServices,
		weekendServices,
		mixedServices,
		len(services),
	))
}
