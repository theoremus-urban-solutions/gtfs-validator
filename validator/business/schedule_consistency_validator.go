package business

import (
	"fmt"
	"io"
	"log"
	"sort"
	"strconv"
	"strings"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/parser"
	"github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

// ScheduleConsistencyValidator validates schedule consistency and timing patterns
type ScheduleConsistencyValidator struct{}

// NewScheduleConsistencyValidator creates a new schedule consistency validator
func NewScheduleConsistencyValidator() *ScheduleConsistencyValidator {
	return &ScheduleConsistencyValidator{}
}

// TripSchedule represents a complete trip schedule
type TripSchedule struct {
	TripID    string
	RouteID   string
	ServiceID string
	StopTimes []*ScheduledStop
	Duration  int // Total trip duration in seconds
	RowNumber int
}

// ScheduledStop represents a scheduled stop with timing
type ScheduledStop struct {
	StopID        string
	StopSequence  int
	ArrivalTime   *TimeOfDay
	DepartureTime *TimeOfDay
	PickupType    int
	DropOffType   int
	RowNumber     int
}

// TimeOfDay represents a GTFS time (can be > 24:00:00)
type TimeOfDay struct {
	Hours   int
	Minutes int
	Seconds int
	Total   int // Total seconds from 00:00:00
}

// RouteSchedulePattern represents scheduling patterns for a route
type RouteSchedulePattern struct {
	RouteID         string
	ServicePatterns map[string]*ServicePattern
	AverageHeadway  float64
	PeakHeadway     float64
	OffPeakHeadway  float64
}

// ServicePattern represents the scheduling pattern for a service
type ServicePattern struct {
	ServiceID      string
	TripCount      int
	AverageHeadway float64
	Headways       []int
	FirstTrip      *TimeOfDay
	LastTrip       *TimeOfDay
}

// Validate performs comprehensive schedule consistency validation
func (v *ScheduleConsistencyValidator) Validate(loader *parser.FeedLoader, container *notice.NoticeContainer, config validator.Config) {
	// Load all trip schedules
	tripSchedules := v.loadTripSchedules(loader)
	if len(tripSchedules) == 0 {
		return
	}

	// Validate individual trip schedules
	for _, trip := range tripSchedules {
		v.validateTripSchedule(container, trip)
	}

	// Analyze route scheduling patterns
	routePatterns := v.analyzeRoutePatterns(tripSchedules)

	// Validate route-level scheduling
	for _, pattern := range routePatterns {
		v.validateRouteScheduling(container, pattern)
	}

	// Validate service-level scheduling
	v.validateServiceScheduling(container, routePatterns)
}

// loadTripSchedules loads all trip schedules from stop_times.txt and trips.txt
func (v *ScheduleConsistencyValidator) loadTripSchedules(loader *parser.FeedLoader) map[string]*TripSchedule {
	schedules := make(map[string]*TripSchedule)

	// First load trip metadata
	tripMetadata := v.loadTripMetadata(loader)

	// Then load stop times and build schedules
	reader, err := loader.GetFile("stop_times.txt")
	if err != nil {
		return schedules
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close reader %v", closeErr)
		}
	}()

	csvFile, err := parser.NewCSVFile(reader, "stop_times.txt")
	if err != nil {
		return schedules
	}

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		stopTime := v.parseStopTime(row)
		if stopTime == nil {
			continue
		}

		tripID := stopTime.TripID
		if schedules[tripID] == nil {
			// Initialize trip schedule
			metadata := tripMetadata[tripID]
			schedules[tripID] = &TripSchedule{
				TripID:    tripID,
				RouteID:   metadata.RouteID,
				ServiceID: metadata.ServiceID,
				StopTimes: []*ScheduledStop{},
				RowNumber: metadata.RowNumber,
			}
		}

		// Convert to ScheduledStop
		scheduledStop := &ScheduledStop{
			StopID:       stopTime.StopID,
			StopSequence: stopTime.StopSequence,
			PickupType:   stopTime.PickupType,
			DropOffType:  stopTime.DropOffType,
			RowNumber:    stopTime.RowNumber,
		}

		// Parse times
		if stopTime.ArrivalTime != "" {
			scheduledStop.ArrivalTime = v.parseGTFSTime(stopTime.ArrivalTime)
		}
		if stopTime.DepartureTime != "" {
			scheduledStop.DepartureTime = v.parseGTFSTime(stopTime.DepartureTime)
		}

		schedules[tripID].StopTimes = append(schedules[tripID].StopTimes, scheduledStop)
	}

	// Sort stop times and calculate durations
	for _, schedule := range schedules {
		sort.Slice(schedule.StopTimes, func(i, j int) bool {
			return schedule.StopTimes[i].StopSequence < schedule.StopTimes[j].StopSequence
		})
		v.calculateTripDuration(schedule)
	}

	return schedules
}

// StopTimeData represents basic stop time data
type StopTimeData struct {
	TripID        string
	StopID        string
	StopSequence  int
	ArrivalTime   string
	DepartureTime string
	PickupType    int
	DropOffType   int
	RowNumber     int
}

// TripMetadata represents basic trip information
type TripMetadata struct {
	RouteID   string
	ServiceID string
	RowNumber int
}

// loadTripMetadata loads trip metadata from trips.txt
func (v *ScheduleConsistencyValidator) loadTripMetadata(loader *parser.FeedLoader) map[string]*TripMetadata {
	metadata := make(map[string]*TripMetadata)

	reader, err := loader.GetFile("trips.txt")
	if err != nil {
		return metadata
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close reader %v", closeErr)
		}
	}()

	csvFile, err := parser.NewCSVFile(reader, "trips.txt")
	if err != nil {
		return metadata
	}

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		tripID, hasTripID := row.Values["trip_id"]
		routeID, hasRouteID := row.Values["route_id"]
		serviceID, hasServiceID := row.Values["service_id"]

		if hasTripID && hasRouteID && hasServiceID {
			metadata[strings.TrimSpace(tripID)] = &TripMetadata{
				RouteID:   strings.TrimSpace(routeID),
				ServiceID: strings.TrimSpace(serviceID),
				RowNumber: row.RowNumber,
			}
		}
	}

	return metadata
}

// parseStopTime parses a stop time record
func (v *ScheduleConsistencyValidator) parseStopTime(row *parser.CSVRow) *StopTimeData {
	tripID, hasTripID := row.Values["trip_id"]
	stopID, hasStopID := row.Values["stop_id"]
	stopSeqStr, hasStopSeq := row.Values["stop_sequence"]

	if !hasTripID || !hasStopID || !hasStopSeq {
		return nil
	}

	stopSeq, err := strconv.Atoi(strings.TrimSpace(stopSeqStr))
	if err != nil {
		return nil
	}

	stopTime := &StopTimeData{
		TripID:       strings.TrimSpace(tripID),
		StopID:       strings.TrimSpace(stopID),
		StopSequence: stopSeq,
		RowNumber:    row.RowNumber,
	}

	// Parse optional fields
	if arrivalTime, hasArrival := row.Values["arrival_time"]; hasArrival {
		stopTime.ArrivalTime = strings.TrimSpace(arrivalTime)
	}
	if departureTime, hasDeparture := row.Values["departure_time"]; hasDeparture {
		stopTime.DepartureTime = strings.TrimSpace(departureTime)
	}
	if pickupStr, hasPickup := row.Values["pickup_type"]; hasPickup && strings.TrimSpace(pickupStr) != "" {
		if pickup, err := strconv.Atoi(strings.TrimSpace(pickupStr)); err == nil {
			stopTime.PickupType = pickup
		}
	}
	if dropOffStr, hasDropOff := row.Values["drop_off_type"]; hasDropOff && strings.TrimSpace(dropOffStr) != "" {
		if dropOff, err := strconv.Atoi(strings.TrimSpace(dropOffStr)); err == nil {
			stopTime.DropOffType = dropOff
		}
	}

	return stopTime
}

// parseGTFSTime parses a GTFS time string (HH:MM:SS, can be > 24:00:00)
func (v *ScheduleConsistencyValidator) parseGTFSTime(timeStr string) *TimeOfDay {
	parts := strings.Split(timeStr, ":")
	if len(parts) != 3 {
		return nil
	}

	hours, err1 := strconv.Atoi(parts[0])
	minutes, err2 := strconv.Atoi(parts[1])
	seconds, err3 := strconv.Atoi(parts[2])

	if err1 != nil || err2 != nil || err3 != nil {
		return nil
	}

	if minutes < 0 || minutes >= 60 || seconds < 0 || seconds >= 60 {
		return nil
	}

	total := hours*3600 + minutes*60 + seconds

	return &TimeOfDay{
		Hours:   hours,
		Minutes: minutes,
		Seconds: seconds,
		Total:   total,
	}
}

// formatTime formats a TimeOfDay back to GTFS time string
func (v *ScheduleConsistencyValidator) formatTime(t *TimeOfDay) string {
	return fmt.Sprintf("%02d:%02d:%02d", t.Hours, t.Minutes, t.Seconds)
}

// calculateTripDuration calculates the total duration of a trip
func (v *ScheduleConsistencyValidator) calculateTripDuration(schedule *TripSchedule) {
	if len(schedule.StopTimes) < 2 {
		return
	}

	firstStop := schedule.StopTimes[0]
	lastStop := schedule.StopTimes[len(schedule.StopTimes)-1]

	var startTime, endTime *TimeOfDay

	// Use departure time for first stop, arrival time for last stop
	if firstStop.DepartureTime != nil {
		startTime = firstStop.DepartureTime
	} else if firstStop.ArrivalTime != nil {
		startTime = firstStop.ArrivalTime
	}

	if lastStop.ArrivalTime != nil {
		endTime = lastStop.ArrivalTime
	} else if lastStop.DepartureTime != nil {
		endTime = lastStop.DepartureTime
	}

	if startTime != nil && endTime != nil {
		schedule.Duration = endTime.Total - startTime.Total
	}
}

// validateTripSchedule validates an individual trip schedule
func (v *ScheduleConsistencyValidator) validateTripSchedule(container *notice.NoticeContainer, schedule *TripSchedule) {
	if len(schedule.StopTimes) < 2 {
		return
	}

	// Check for timing consistency
	var prevTime *TimeOfDay
	for i, stop := range schedule.StopTimes {
		// Check arrival <= departure at same stop
		if stop.ArrivalTime != nil && stop.DepartureTime != nil {
			if stop.ArrivalTime.Total > stop.DepartureTime.Total {
				arrivalStr := v.formatTime(stop.ArrivalTime)
				departureStr := v.formatTime(stop.DepartureTime)
				container.AddNotice(notice.NewStopTimeArrivalAfterDepartureNotice(
					schedule.TripID,
					stop.StopSequence,
					arrivalStr,
					departureStr,
					stop.RowNumber,
				))
			}
		}

		// Check increasing times between stops
		var currentTime *TimeOfDay
		if stop.DepartureTime != nil {
			currentTime = stop.DepartureTime
		} else if stop.ArrivalTime != nil {
			currentTime = stop.ArrivalTime
		}

		if prevTime != nil && currentTime != nil {
			if currentTime.Total < prevTime.Total {
				currentTimeStr := v.formatTime(currentTime)
				prevTimeStr := v.formatTime(prevTime)
				// Find previous stop info
				prevStop := schedule.StopTimes[i-1]
				container.AddNotice(notice.NewStopTimeDecreasingTimeNotice(
					schedule.TripID,
					stop.StopSequence,
					currentTimeStr,
					stop.RowNumber,
					prevStop.StopSequence,
					prevTimeStr,
					prevStop.RowNumber,
				))
			}
		}

		prevTime = currentTime

		// Validate pickup/drop-off rules
		v.validatePickupDropOffRules(container, schedule, stop, i)
	}

	// Check trip duration reasonableness
	if schedule.Duration > 0 {
		// Very short trips (< 2 minutes)
		if schedule.Duration < 120 {
			container.AddNotice(notice.NewVeryShortTripNotice(
				schedule.TripID,
				schedule.Duration,
				len(schedule.StopTimes),
			))
		}

		// Very long trips (> 4 hours)
		if schedule.Duration > 14400 {
			container.AddNotice(notice.NewVeryLongTripNotice(
				schedule.TripID,
				schedule.Duration,
				len(schedule.StopTimes),
			))
		}
	}
}

// validatePickupDropOffRules validates pickup and drop-off rules
func (v *ScheduleConsistencyValidator) validatePickupDropOffRules(container *notice.NoticeContainer, schedule *TripSchedule, stop *ScheduledStop, index int) {
	isFirstStop := index == 0
	isLastStop := index == len(schedule.StopTimes)-1

	// First stop should allow pickup
	if isFirstStop && stop.PickupType == 1 {
		container.AddNotice(notice.NewFirstStopNoPickupNotice(
			schedule.TripID,
			stop.StopID,
			stop.RowNumber,
		))
	}

	// Last stop should allow drop-off
	if isLastStop && stop.DropOffType == 1 {
		container.AddNotice(notice.NewLastStopNoDropOffNotice(
			schedule.TripID,
			stop.StopID,
			stop.RowNumber,
		))
	}

	// Check for stops with neither pickup nor drop-off
	if stop.PickupType == 1 && stop.DropOffType == 1 {
		container.AddNotice(notice.NewStopWithoutServiceNotice(
			schedule.TripID,
			stop.StopID,
			stop.StopSequence,
			stop.RowNumber,
		))
	}
}

// analyzeRoutePatterns analyzes scheduling patterns by route
func (v *ScheduleConsistencyValidator) analyzeRoutePatterns(schedules map[string]*TripSchedule) map[string]*RouteSchedulePattern {
	patterns := make(map[string]*RouteSchedulePattern)

	// Group schedules by route
	routeSchedules := make(map[string][]*TripSchedule)
	for _, schedule := range schedules {
		routeSchedules[schedule.RouteID] = append(routeSchedules[schedule.RouteID], schedule)
	}

	// Analyze each route
	for routeID, routeTrips := range routeSchedules {
		pattern := &RouteSchedulePattern{
			RouteID:         routeID,
			ServicePatterns: make(map[string]*ServicePattern),
		}

		// Group by service
		serviceTrips := make(map[string][]*TripSchedule)
		for _, trip := range routeTrips {
			serviceTrips[trip.ServiceID] = append(serviceTrips[trip.ServiceID], trip)
		}

		// Analyze each service pattern
		for serviceID, trips := range serviceTrips {
			servicePattern := v.analyzeServicePattern(serviceID, trips)
			pattern.ServicePatterns[serviceID] = servicePattern
		}

		patterns[routeID] = pattern
	}

	return patterns
}

// analyzeServicePattern analyzes scheduling patterns for a specific service
func (v *ScheduleConsistencyValidator) analyzeServicePattern(serviceID string, trips []*TripSchedule) *ServicePattern {
	if len(trips) == 0 {
		return nil
	}

	pattern := &ServicePattern{
		ServiceID: serviceID,
		TripCount: len(trips),
		Headways:  []int{},
	}

	// Extract departure times and sort
	var departureTimes []*TimeOfDay
	for _, trip := range trips {
		if len(trip.StopTimes) > 0 && trip.StopTimes[0].DepartureTime != nil {
			departureTimes = append(departureTimes, trip.StopTimes[0].DepartureTime)
		}
	}

	if len(departureTimes) == 0 {
		return pattern
	}

	sort.Slice(departureTimes, func(i, j int) bool {
		return departureTimes[i].Total < departureTimes[j].Total
	})

	pattern.FirstTrip = departureTimes[0]
	pattern.LastTrip = departureTimes[len(departureTimes)-1]

	// Calculate headways
	if len(departureTimes) >= 2 {
		totalHeadway := 0
		for i := 1; i < len(departureTimes); i++ {
			headway := departureTimes[i].Total - departureTimes[i-1].Total
			pattern.Headways = append(pattern.Headways, headway)
			totalHeadway += headway
		}

		if len(pattern.Headways) > 0 {
			pattern.AverageHeadway = float64(totalHeadway) / float64(len(pattern.Headways))
		}
	}

	return pattern
}

// validateRouteScheduling validates route-level scheduling patterns
func (v *ScheduleConsistencyValidator) validateRouteScheduling(container *notice.NoticeContainer, pattern *RouteSchedulePattern) {
	for serviceID, servicePattern := range pattern.ServicePatterns {
		if servicePattern == nil {
			continue
		}

		// Check for very irregular headways
		if len(servicePattern.Headways) >= 3 {
			v.validateHeadwayConsistency(container, pattern.RouteID, serviceID, servicePattern)
		}

		// Check for very short or long service spans
		if servicePattern.FirstTrip != nil && servicePattern.LastTrip != nil {
			serviceSpan := servicePattern.LastTrip.Total - servicePattern.FirstTrip.Total

			// Very short service span (< 1 hour)
			if serviceSpan < 3600 && servicePattern.TripCount > 5 {
				container.AddNotice(notice.NewShortServiceSpanNotice(
					pattern.RouteID,
					serviceID,
					serviceSpan,
					servicePattern.TripCount,
				))
			}

			// Very long service span (> 20 hours)
			if serviceSpan > 72000 {
				container.AddNotice(notice.NewLongServiceSpanNotice(
					pattern.RouteID,
					serviceID,
					serviceSpan,
					servicePattern.TripCount,
				))
			}
		}

		// Check for single trip services
		if servicePattern.TripCount == 1 {
			container.AddNotice(notice.NewSingleTripServiceNotice(
				pattern.RouteID,
				serviceID,
				servicePattern.TripCount,
			))
		}
	}
}

// validateHeadwayConsistency validates headway consistency
func (v *ScheduleConsistencyValidator) validateHeadwayConsistency(container *notice.NoticeContainer, routeID, serviceID string, pattern *ServicePattern) {
	if len(pattern.Headways) < 3 {
		return
	}

	// Calculate coefficient of variation
	mean := pattern.AverageHeadway
	variance := 0.0

	for _, headway := range pattern.Headways {
		diff := float64(headway) - mean
		variance += diff * diff
	}

	variance /= float64(len(pattern.Headways))
	stdDev := variance // Simplified - not taking square root for threshold comparison

	// High variation in headways
	if stdDev > mean*mean*0.25 { // CV > 0.5 (squared)
		container.AddNotice(notice.NewIrregularHeadwayNotice(
			routeID,
			serviceID,
			mean,
			stdDev,
			len(pattern.Headways),
		))
	}

	// Check for very short headways (< 2 minutes)
	for _, headway := range pattern.Headways {
		if headway < 120 {
			container.AddNotice(notice.NewVeryShortHeadwayNotice(
				routeID,
				serviceID,
				headway,
			))
			break
		}
	}

	// Check for very long headways (> 2 hours)
	for _, headway := range pattern.Headways {
		if headway > 7200 {
			container.AddNotice(notice.NewVeryLongHeadwayNotice(
				routeID,
				serviceID,
				headway,
			))
			break
		}
	}
}

// validateServiceScheduling validates service-level scheduling
func (v *ScheduleConsistencyValidator) validateServiceScheduling(container *notice.NoticeContainer, patterns map[string]*RouteSchedulePattern) {
	// Collect statistics
	totalRoutes := len(patterns)
	totalServices := 0
	totalTrips := 0

	for _, pattern := range patterns {
		totalServices += len(pattern.ServicePatterns)
		for _, servicePattern := range pattern.ServicePatterns {
			if servicePattern != nil {
				totalTrips += servicePattern.TripCount
			}
		}
	}

	// Generate summary
	if totalRoutes > 0 {
		avgServicesPerRoute := float64(totalServices) / float64(totalRoutes)
		avgTripsPerService := 0.0
		if totalServices > 0 {
			avgTripsPerService = float64(totalTrips) / float64(totalServices)
		}

		container.AddNotice(notice.NewSchedulingSummaryNotice(
			totalRoutes,
			totalServices,
			totalTrips,
			avgServicesPerRoute,
			avgTripsPerService,
		))
	}
}
