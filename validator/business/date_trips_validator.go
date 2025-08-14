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

// DateTripsValidator validates that trips exist for the next 7 days with majority service coverage
type DateTripsValidator struct{}

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
	TripCount  int
	RowNumber  int
}

// CalendarException represents calendar_dates.txt exceptions
type CalendarException struct {
	ServiceID     string
	Date          time.Time
	ExceptionType int // 1 = added, 2 = removed
}

// Validate checks that adequate service exists for the next 7 days
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

	// Load service information
	services := v.loadServices(loader)
	exceptions := v.loadCalendarExceptions(loader)

	// Check if we have any services at all (both calendar.txt and calendar_dates.txt can define services)
	if len(services) == 0 && len(exceptions) == 0 {
		container.AddNotice(notice.NewNoServiceDefinedNotice())
		return
	}

	// Check service coverage for the next 7 days
	v.validateNext7DaysService(container, services, exceptions, currentDate)

	// Check service coverage for the next 30 days (warning level)
	v.validateNext30DaysService(container, services, exceptions, currentDate)
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
			// Count trips for this service
			service.TripCount = v.countTripsForService(loader, service.ServiceID)
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

// countTripsForService counts trips that use a specific service
func (v *DateTripsValidator) countTripsForService(loader *parser.FeedLoader, serviceID string) int {
	reader, err := loader.GetFile("trips.txt")
	if err != nil {
		return 0
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close reader %v", closeErr)
		}
	}()

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

		if tripServiceID, hasServiceID := row.Values["service_id"]; hasServiceID {
			if strings.TrimSpace(tripServiceID) == serviceID {
				count++
			}
		}
	}

	return count
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

// formatGTFSDate formats time as GTFS date
func (v *DateTripsValidator) formatGTFSDate(date time.Time) string {
	return date.Format("20060102")
}

// validateNext7DaysService validates service coverage for next 7 days
func (v *DateTripsValidator) validateNext7DaysService(container *notice.NoticeContainer, services map[string]*ServiceInfo, exceptions []CalendarException, currentDate time.Time) {
	daysWithService := 0
	totalTrips := 0

	// Check each of the next 7 days
	for i := 0; i < 7; i++ {
		checkDate := currentDate.AddDate(0, 0, i)
		activeServices := v.getActiveServicesForDate(services, exceptions, checkDate)

		dayTripCount := 0
		for _, serviceID := range activeServices {
			if service, exists := services[serviceID]; exists {
				dayTripCount += service.TripCount
			}
		}

		if dayTripCount > 0 {
			daysWithService++
			totalTrips += dayTripCount
		}
	}

	// Critical: No service in next 7 days
	if daysWithService == 0 {
		container.AddNotice(notice.NewNoServiceNext7DaysNotice(
			v.formatGTFSDate(currentDate),
			v.formatGTFSDate(currentDate.AddDate(0, 0, 7)),
		))
		return
	}

	// Warning: Less than majority service coverage (< 4 out of 7 days)
	if daysWithService < 4 {
		container.AddNotice(notice.NewInsufficientServiceNext7DaysNotice(
			daysWithService,
			7,
			v.formatGTFSDate(currentDate),
			v.formatGTFSDate(currentDate.AddDate(0, 0, 7)),
		))
	}

	// Warning: Very few trips per day on average
	avgTripsPerDay := float64(totalTrips) / float64(daysWithService)
	if avgTripsPerDay < 10 {
		container.AddNotice(notice.NewLowTripVolumeNext7DaysNotice(
			totalTrips,
			daysWithService,
			avgTripsPerDay,
			v.formatGTFSDate(currentDate),
			v.formatGTFSDate(currentDate.AddDate(0, 0, 7)),
		))
	}
}

// validateNext30DaysService validates service coverage for next 30 days
func (v *DateTripsValidator) validateNext30DaysService(container *notice.NoticeContainer, services map[string]*ServiceInfo, exceptions []CalendarException, currentDate time.Time) {
	daysWithService := 0

	// Check each of the next 30 days
	for i := 0; i < 30; i++ {
		checkDate := currentDate.AddDate(0, 0, i)
		activeServices := v.getActiveServicesForDate(services, exceptions, checkDate)

		if len(activeServices) > 0 {
			daysWithService++
		}
	}

	// Warning: Less than 50% service coverage in next 30 days
	serviceRatio := float64(daysWithService) / 30.0
	if serviceRatio < 0.5 {
		container.AddNotice(notice.NewInsufficientServiceNext30DaysNotice(
			daysWithService,
			30,
			serviceRatio,
			v.formatGTFSDate(currentDate),
			v.formatGTFSDate(currentDate.AddDate(0, 0, 30)),
		))
	}
}

// getActiveServicesForDate returns service IDs active on a specific date
func (v *DateTripsValidator) getActiveServicesForDate(services map[string]*ServiceInfo, exceptions []CalendarException, date time.Time) []string {
	activeServices := make(map[string]bool)

	// Check regular calendar services
	weekday := int(date.Weekday())
	if weekday == 0 { // Sunday in Go is 0, but we store as index 6
		weekday = 6
	} else {
		weekday = weekday - 1 // Convert to 0-based indexing (Mon=0, Tue=1, etc.)
	}

	for serviceID, service := range services {
		// Check if date is within service period
		if service.StartDate != nil && date.Before(*service.StartDate) {
			continue
		}
		if service.EndDate != nil && date.After(*service.EndDate) {
			continue
		}

		// Check if service runs on this day of week
		if service.DaysOfWeek[weekday] {
			activeServices[serviceID] = true
		}
	}

	// Apply calendar exceptions
	for _, exception := range exceptions {
		if exception.Date.Equal(date) {
			switch exception.ExceptionType {
			case 1: // Service added
				activeServices[exception.ServiceID] = true
			case 2: // Service removed
				delete(activeServices, exception.ServiceID)
			}
		}
	}

	// Convert to slice
	result := make([]string, 0, len(activeServices))
	for serviceID := range activeServices {
		result = append(result, serviceID)
	}

	return result
}
