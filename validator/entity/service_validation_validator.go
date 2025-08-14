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

// ServiceInfo represents service information
type ServiceInfo struct {
	ServiceID string
	StartDate string
	EndDate   string
	Days      map[string]bool
	RowNumber int
}

// Validate checks service definitions for consistency
func (v *ServiceValidationValidator) Validate(loader *parser.FeedLoader, container *notice.NoticeContainer, config validator.Config) {
	// Load calendar services
	calendarServices := v.loadCalendarServices(loader)

	// Load calendar date services
	calendarDateServices := v.loadCalendarDateServices(loader)

	// Validate calendar services
	v.validateCalendarServices(container, calendarServices)

	// Validate calendar date services
	v.validateCalendarDateServices(container, calendarDateServices)

	// Check for unused services
	v.validateServiceUsage(loader, container, calendarServices, calendarDateServices)
}

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
func (v *ServiceValidationValidator) loadCalendarDateServices(loader *parser.FeedLoader) map[string]bool {
	services := make(map[string]bool)

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
		if hasServiceID {
			services[strings.TrimSpace(serviceID)] = true
		}
	}

	return services
}

// validateCalendarServices validates calendar.txt services
func (v *ServiceValidationValidator) validateCalendarServices(container *notice.NoticeContainer, services map[string]*ServiceInfo) {
	for _, service := range services {
		v.validateCalendarService(container, service)
	}
}

// validateCalendarService validates a single calendar service
func (v *ServiceValidationValidator) validateCalendarService(container *notice.NoticeContainer, service *ServiceInfo) {
	// Check if service has at least one active day
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

	// Validate date range
	if service.StartDate != "" && service.EndDate != "" {
		v.validateServiceDateRange(container, service)
	}

	// Check if service is expired
	if service.EndDate != "" {
		v.validateServiceExpiration(container, service)
	}
}

// validateServiceDateRange validates that start_date <= end_date
func (v *ServiceValidationValidator) validateServiceDateRange(container *notice.NoticeContainer, service *ServiceInfo) {
	startDate, err1 := v.parseGTFSDate(service.StartDate)
	endDate, err2 := v.parseGTFSDate(service.EndDate)

	if err1 != nil || err2 != nil {
		return // Invalid dates - other validators handle this
	}

	if startDate.After(*endDate) {
		container.AddNotice(notice.NewInvalidServiceDateRangeNotice(
			service.ServiceID,
			service.StartDate,
			service.EndDate,
			service.RowNumber,
		))
	}
}

// validateServiceExpiration checks if service is expired
func (v *ServiceValidationValidator) validateServiceExpiration(container *notice.NoticeContainer, service *ServiceInfo) {
	endDate, err := v.parseGTFSDate(service.EndDate)
	if err != nil {
		return // Invalid date - other validators handle this
	}

	// Check if service ended more than 30 days ago
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	if endDate.Before(thirtyDaysAgo) {
		container.AddNotice(notice.NewExpiredServiceNotice(
			service.ServiceID,
			service.EndDate,
			service.RowNumber,
		))
	}
}

// validateCalendarDateServices validates calendar_dates.txt
func (v *ServiceValidationValidator) validateCalendarDateServices(container *notice.NoticeContainer, services map[string]bool) {
	// Load and validate calendar_dates.txt records
	// This is a placeholder - full implementation would validate exception types, dates, etc.
}

// validateServiceUsage checks if services are actually used by trips
func (v *ServiceValidationValidator) validateServiceUsage(loader *parser.FeedLoader, container *notice.NoticeContainer, calendarServices map[string]*ServiceInfo, calendarDateServices map[string]bool) {
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
	for serviceID := range calendarDateServices {
		if !usedServices[serviceID] {
			container.AddNotice(notice.NewUnusedServiceNotice(
				serviceID,
				"calendar_dates.txt",
				0, // Row number not tracked for calendar_dates
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
