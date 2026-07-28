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

// DateTripsValidator reports feeds whose trips stop running within the coming
// week. A feed that expires mid-week strands every consumer that refreshes on
// a weekly cadence.
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

	// A feed with no services at all is reported as
	// missing_calendar_and_calendar_date_files by core/missing_files_validator.
	if len(services) == 0 && len(exceptions) == 0 {
		return
	}

	v.validateNext7DaysService(container, services, exceptions, currentDate)
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

	// Counting trips per service in one pass over trips.txt, rather than
	// re-reading the file for each service, keeps this linear in the feed
	// size instead of quadratic.
	for serviceID, count := range v.countTripsByService(loader) {
		if service, exists := services[serviceID]; exists {
			service.TripCount = count
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

// formatGTFSDate formats time as GTFS date
func (v *DateTripsValidator) formatGTFSDate(date time.Time) string {
	return date.Format("20060102")
}

// validateNext7DaysService reports a feed whose trips do not cover the coming
// week. Canonical counts the feed as covered while a significant share of its
// trips still run, so a day is only counted when trips are scheduled on it.
func (v *DateTripsValidator) validateNext7DaysService(container *notice.NoticeContainer, services map[string]*ServiceInfo, exceptions []CalendarException, currentDate time.Time) {
	lastCoveredDay := -1

	for i := 0; i < 7; i++ {
		checkDate := currentDate.AddDate(0, 0, i)

		dayTripCount := 0
		for _, serviceID := range v.getActiveServicesForDate(services, exceptions, checkDate) {
			if service, exists := services[serviceID]; exists {
				dayTripCount += service.TripCount
			}
		}
		if dayTripCount > 0 {
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

// getActiveServicesForDate returns service IDs active on a specific date
func (v *DateTripsValidator) getActiveServicesForDate(services map[string]*ServiceInfo, exceptions []CalendarException, date time.Time) []string {
	activeServices := make(map[string]bool)

	// Check regular calendar services
	weekday := int(date.Weekday())
	if weekday == 0 { // Sunday in Go is 0, but we store as index 6
		weekday = 6
	} else {
		weekday-- // Convert to 0-based indexing (Mon=0, Tue=1, etc.)
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
