package business

import (
	"fmt"
	"io"
	"log"
	"sort"
	"strconv"
	"strings"
	"sync"

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
	// Load all trip schedules with optimized processing
	tripSchedules := v.loadTripSchedulesOptimized(loader)
	if len(tripSchedules) == 0 {
		return
	}

	// Validate ALL trip schedules efficiently using parallel processing
	// This validates ALL the core data: timing consistency, sequence validation,
	// pickup/dropoff rules, trip durations - the most important validations
	v.validateAllTripsParallel(container, tripSchedules)

	// Generate basic statistics summary - no expensive pattern analysis
	v.generateBasicStatistics(container, tripSchedules)
}

// generateBasicStatistics generates basic validation statistics for large datasets
func (v *ScheduleConsistencyValidator) generateBasicStatistics(container *notice.NoticeContainer, tripSchedules map[string]*TripSchedule) {
	totalTrips := len(tripSchedules)
	routeCount := make(map[string]bool)
	serviceCount := make(map[string]bool)

	for _, trip := range tripSchedules {
		routeCount[trip.RouteID] = true
		serviceCount[trip.ServiceID] = true
	}

	// Generate summary notice with basic statistics
	context := map[string]interface{}{
		"totalTrips":    totalTrips,
		"totalRoutes":   len(routeCount),
		"totalServices": len(serviceCount),
		"mode":          "simplified_for_large_dataset",
	}
	container.AddNotice(notice.NewBaseNotice("schedule_validation_summary", notice.INFO, context))
}

// validateAllTripsParallel validates all trips using parallel processing
func (v *ScheduleConsistencyValidator) validateAllTripsParallel(container *notice.NoticeContainer, tripSchedules map[string]*TripSchedule) {
	// Convert map to slice for parallel processing
	trips := make([]*TripSchedule, 0, len(tripSchedules))
	for _, trip := range tripSchedules {
		trips = append(trips, trip)
	}

	// Use reasonable number of workers
	const workers = 8
	tripChan := make(chan *TripSchedule, 100)

	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for trip := range tripChan {
				v.validateTripSchedule(container, trip)
			}
		}()
	}

	// Send work to workers
	for _, trip := range trips {
		tripChan <- trip
	}
	close(tripChan)

	// Wait for completion
	wg.Wait()
}

// loadTripSchedulesOptimized loads ALL trip schedules with optimized processing
func (v *ScheduleConsistencyValidator) loadTripSchedulesOptimized(loader *parser.FeedLoader) map[string]*TripSchedule {
	// First load trip metadata
	tripMetadata := v.loadTripMetadata(loader)
	if len(tripMetadata) == 0 {
		return nil
	}

	// Pre-allocate schedules map with exact size
	schedules := make(map[string]*TripSchedule, len(tripMetadata))

	// Pre-create all trip schedules to avoid map lookups
	for tripID, metadata := range tripMetadata {
		schedules[tripID] = &TripSchedule{
			TripID:    tripID,
			RouteID:   metadata.RouteID,
			ServiceID: metadata.ServiceID,
			StopTimes: make([]*ScheduledStop, 0, 25), // Most trips have ~25 stops
			RowNumber: metadata.RowNumber,
		}
	}

	// Load stop times
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

	// Process ALL stop_times without limits
	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		// Fast parsing without creating intermediate structs
		tripID, hasTripID := row.Values["trip_id"]
		if !hasTripID {
			continue
		}

		tripID = strings.TrimSpace(tripID)
		schedule, exists := schedules[tripID]
		if !exists {
			continue
		}

		// Direct creation of ScheduledStop
		stopID := row.Values["stop_id"]
		stopSeqStr := row.Values["stop_sequence"]
		stopSeq, err := strconv.Atoi(strings.TrimSpace(stopSeqStr))
		if err != nil {
			continue
		}

		scheduledStop := &ScheduledStop{
			StopID:       strings.TrimSpace(stopID),
			StopSequence: stopSeq,
			RowNumber:    row.RowNumber,
		}

		// Parse times only if present
		if arrivalTime, hasArrival := row.Values["arrival_time"]; hasArrival && arrivalTime != "" {
			scheduledStop.ArrivalTime = v.parseGTFSTime(strings.TrimSpace(arrivalTime))
		}
		if departureTime, hasDeparture := row.Values["departure_time"]; hasDeparture && departureTime != "" {
			scheduledStop.DepartureTime = v.parseGTFSTime(strings.TrimSpace(departureTime))
		}

		// Parse optional fields
		if pickupStr, hasPickup := row.Values["pickup_type"]; hasPickup {
			if pickup, err := strconv.Atoi(strings.TrimSpace(pickupStr)); err == nil {
				scheduledStop.PickupType = pickup
			}
		}
		if dropOffStr, hasDropOff := row.Values["drop_off_type"]; hasDropOff {
			if dropOff, err := strconv.Atoi(strings.TrimSpace(dropOffStr)); err == nil {
				scheduledStop.DropOffType = dropOff
			}
		}

		schedule.StopTimes = append(schedule.StopTimes, scheduledStop)
	}

	// Sort all trips' stop times and calculate durations in parallel
	v.parallelSortAndCalculate(schedules)

	return schedules
}

// parallelSortAndCalculate sorts stop times and calculates durations in parallel
func (v *ScheduleConsistencyValidator) parallelSortAndCalculate(schedules map[string]*TripSchedule) {
	// Use goroutines for parallel processing with reasonable concurrency
	const workers = 8
	tripChan := make(chan *TripSchedule, 100)

	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for schedule := range tripChan {
				if len(schedule.StopTimes) > 1 {
					// Sort by stop sequence
					sort.Slice(schedule.StopTimes, func(i, j int) bool {
						return schedule.StopTimes[i].StopSequence < schedule.StopTimes[j].StopSequence
					})
					// Calculate duration
					v.calculateTripDuration(schedule)
				}
			}
		}()
	}

	// Send work to workers
	for _, schedule := range schedules {
		tripChan <- schedule
	}
	close(tripChan)

	// Wait for completion
	wg.Wait()
}

// Note: Replaced with loadTripSchedulesOptimized - old batch processing removed

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

// Note: Removed expensive pattern analysis functions - now focuses on core data validation only
