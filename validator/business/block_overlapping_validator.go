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

// BlockOverlappingValidator validates that trips with the same block_id don't overlap in time
type BlockOverlappingValidator struct{}

// NewBlockOverlappingValidator creates a new block overlapping validator
func NewBlockOverlappingValidator() *BlockOverlappingValidator {
	return &BlockOverlappingValidator{}
}

// TripTimeRange represents a trip's time range for block validation
type TripTimeRange struct {
	TripID    string
	BlockID   string
	ServiceID string
	StartTime int // seconds since midnight
	EndTime   int // seconds since midnight
	RowNumber int
}

// Validate checks that trips with the same block_id don't overlap in time on the same service dates
func (v *BlockOverlappingValidator) Validate(loader *parser.FeedLoader, container *notice.NoticeContainer, config validator.Config) {
	// Load trip block information
	tripBlocks := v.loadTripBlocks(loader)
	if len(tripBlocks) == 0 {
		return // No block information available
	}

	// Check the modes of the routes the block's trips run on. This needs only
	// trips.txt and routes.txt, so it runs before the stop_times work.
	v.validateBlockRouteTypes(loader, container, tripBlocks)

	// Load trip time ranges from stop_times.txt
	tripTimeRanges := v.loadTripTimeRanges(loader, tripBlocks)
	if len(tripTimeRanges) == 0 {
		return // No time range data available
	}

	// Load service date information
	serviceDates := v.loadServiceDates(loader)

	// Validate block overlaps
	v.validateBlockOverlaps(container, tripTimeRanges, serviceDates)
}

// loadTripBlocks loads trip-to-block mappings from trips.txt
func (v *BlockOverlappingValidator) loadTripBlocks(loader *parser.FeedLoader) map[string]*TripBlock {
	tripBlocks := make(map[string]*TripBlock)

	reader, err := loader.GetFile("trips.txt")
	if err != nil {
		return tripBlocks
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close reader %v", closeErr)
		}
	}()

	csvFile, err := parser.NewCSVFile(reader, "trips.txt")
	if err != nil {
		return tripBlocks
	}

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}

		tripID, hasTripID := row.Values["trip_id"]
		blockID, hasBlockID := row.Values["block_id"]
		serviceID, hasServiceID := row.Values["service_id"]

		if hasTripID && hasBlockID && hasServiceID {
			tripIDTrimmed := strings.TrimSpace(tripID)
			blockIDTrimmed := strings.TrimSpace(blockID)
			serviceIDTrimmed := strings.TrimSpace(serviceID)

			if blockIDTrimmed != "" {
				tripBlocks[tripIDTrimmed] = &TripBlock{
					TripID:    tripIDTrimmed,
					BlockID:   blockIDTrimmed,
					ServiceID: serviceIDTrimmed,
					RouteID:   strings.TrimSpace(row.Values["route_id"]),
					RowNumber: row.RowNumber,
				}
			}
		}
	}

	return tripBlocks
}

// TripBlock represents trip block information
type TripBlock struct {
	TripID    string
	BlockID   string
	ServiceID string
	RouteID   string
	RowNumber int
}

// validateBlockRouteTypes reports blocks whose trips run on routes of
// different modes. A block is one vehicle working through the day, and a
// vehicle does not turn from a bus into a tram between trips, so a mixed block
// is either a reused block_id or a mislabelled route.
func (v *BlockOverlappingValidator) validateBlockRouteTypes(loader *parser.FeedLoader, container *notice.NoticeContainer, tripBlocks map[string]*TripBlock) {
	routeTypes := v.loadRouteTypes(loader)
	if len(routeTypes) == 0 {
		return
	}

	// Trips arrive from a map, so they are ordered by their row before the
	// comparison: which trip a block is judged against must not depend on map
	// iteration order.
	trips := make([]*TripBlock, 0, len(tripBlocks))
	for _, trip := range tripBlocks {
		if _, known := routeTypes[trip.RouteID]; known {
			trips = append(trips, trip)
		}
	}
	sort.Slice(trips, func(i, j int) bool {
		return trips[i].RowNumber < trips[j].RowNumber
	})

	// The first trip of a block sets the mode the rest are held to, so each
	// block is reported once per trip that disagrees rather than for every pair.
	firstOfBlock := make(map[string]*TripBlock)
	for _, trip := range trips {
		first, seen := firstOfBlock[trip.BlockID]
		if !seen {
			firstOfBlock[trip.BlockID] = trip
			continue
		}
		if routeTypes[first.RouteID] == routeTypes[trip.RouteID] {
			continue
		}
		container.AddNotice(notice.NewInconsistentRouteTypeForBlockIDNotice(
			trip.BlockID,
			first.TripID,
			first.RouteID,
			routeTypes[first.RouteID],
			trip.TripID,
			trip.RouteID,
			routeTypes[trip.RouteID],
			trip.RowNumber,
		))
	}
}

// loadRouteTypes maps each route to its mode, skipping routes whose type is
// absent or unparseable — the field layer reports those.
func (v *BlockOverlappingValidator) loadRouteTypes(loader *parser.FeedLoader) map[string]int {
	routeTypes := make(map[string]int)

	reader, err := loader.GetFile("routes.txt")
	if err != nil {
		return routeTypes
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close reader %v", closeErr)
		}
	}()

	csvFile, err := parser.NewCSVFile(reader, "routes.txt")
	if err != nil {
		return routeTypes
	}

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}

		routeID := strings.TrimSpace(row.Values["route_id"])
		routeType, err := strconv.Atoi(strings.TrimSpace(row.Values["route_type"]))
		if routeID == "" || err != nil {
			continue
		}
		routeTypes[routeID] = routeType
	}

	return routeTypes
}

// loadTripTimeRanges loads trip time ranges from stop_times.txt
func (v *BlockOverlappingValidator) loadTripTimeRanges(loader *parser.FeedLoader, tripBlocks map[string]*TripBlock) []TripTimeRange {
	var tripTimeRanges []TripTimeRange

	reader, err := loader.GetFile("stop_times.txt")
	if err != nil {
		return tripTimeRanges
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close reader %v", closeErr)
		}
	}()

	csvFile, err := parser.NewCSVFile(reader, "stop_times.txt")
	if err != nil {
		return tripTimeRanges
	}

	// Collect all stop times by trip
	tripStopTimes := make(map[string][]StopTimeForBlock)

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}

		stopTime := v.parseStopTimeForBlock(row)
		if stopTime != nil {
			tripStopTimes[stopTime.TripID] = append(tripStopTimes[stopTime.TripID], *stopTime)
		}
	}

	// Calculate time ranges for each trip
	for tripID, stopTimes := range tripStopTimes {
		tripBlock, hasBlock := tripBlocks[tripID]
		if !hasBlock {
			continue // Skip trips without block information
		}

		timeRange := v.calculateTripTimeRange(tripID, stopTimes, tripBlock)
		if timeRange != nil {
			tripTimeRanges = append(tripTimeRanges, *timeRange)
		}
	}

	return tripTimeRanges
}

// StopTimeForBlock represents a stop time for block validation
type StopTimeForBlock struct {
	TripID        string
	StopSequence  int
	ArrivalTime   *int
	DepartureTime *int
	RowNumber     int
}

// parseStopTimeForBlock parses a stop time row for block validation
func (v *BlockOverlappingValidator) parseStopTimeForBlock(row *parser.CSVRow) *StopTimeForBlock {
	tripID, hasTripID := row.Values["trip_id"]
	stopSeqStr, hasStopSeq := row.Values["stop_sequence"]
	arrivalTimeStr, hasArrivalTime := row.Values["arrival_time"]
	departureTimeStr, hasDepartureTime := row.Values["departure_time"]

	if !hasTripID || !hasStopSeq {
		return nil
	}

	stopSequence, err := strconv.Atoi(strings.TrimSpace(stopSeqStr))
	if err != nil {
		return nil
	}

	stopTime := &StopTimeForBlock{
		TripID:       strings.TrimSpace(tripID),
		StopSequence: stopSequence,
		RowNumber:    row.RowNumber,
	}

	// Parse times
	if hasArrivalTime && strings.TrimSpace(arrivalTimeStr) != "" {
		if arrivalSeconds, err := v.parseGTFSTime(strings.TrimSpace(arrivalTimeStr)); err == nil {
			stopTime.ArrivalTime = &arrivalSeconds
		}
	}

	if hasDepartureTime && strings.TrimSpace(departureTimeStr) != "" {
		if departureSeconds, err := v.parseGTFSTime(strings.TrimSpace(departureTimeStr)); err == nil {
			stopTime.DepartureTime = &departureSeconds
		}
	}

	return stopTime
}

// parseGTFSTime parses a GTFS time string (HH:MM:SS) into seconds since midnight
func (v *BlockOverlappingValidator) parseGTFSTime(timeStr string) (int, error) {
	parts := strings.Split(timeStr, ":")
	if len(parts) != 3 {
		return 0, fmt.Errorf("invalid time format")
	}

	hours, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, err
	}

	minutes, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, err
	}

	seconds, err := strconv.Atoi(parts[2])
	if err != nil {
		return 0, err
	}

	if minutes < 0 || minutes >= 60 || seconds < 0 || seconds >= 60 {
		return 0, fmt.Errorf("invalid time values")
	}

	return hours*3600 + minutes*60 + seconds, nil
}

// calculateTripTimeRange calculates the time range for a trip
func (v *BlockOverlappingValidator) calculateTripTimeRange(tripID string, stopTimes []StopTimeForBlock, tripBlock *TripBlock) *TripTimeRange {
	if len(stopTimes) == 0 {
		return nil
	}

	// Sort by stop sequence
	sort.Slice(stopTimes, func(i, j int) bool {
		return stopTimes[i].StopSequence < stopTimes[j].StopSequence
	})

	// Find earliest time (start of trip)
	var startTime *int
	for _, stopTime := range stopTimes {
		var timeToCheck *int
		if stopTime.ArrivalTime != nil {
			timeToCheck = stopTime.ArrivalTime
		} else if stopTime.DepartureTime != nil {
			timeToCheck = stopTime.DepartureTime
		}

		if timeToCheck != nil && (startTime == nil || *timeToCheck < *startTime) {
			startTime = timeToCheck
		}
	}

	// Find latest time (end of trip)
	var endTime *int
	for i := len(stopTimes) - 1; i >= 0; i-- {
		stopTime := stopTimes[i]
		var timeToCheck *int
		if stopTime.DepartureTime != nil {
			timeToCheck = stopTime.DepartureTime
		} else if stopTime.ArrivalTime != nil {
			timeToCheck = stopTime.ArrivalTime
		}

		if timeToCheck != nil && (endTime == nil || *timeToCheck > *endTime) {
			endTime = timeToCheck
		}
	}

	if startTime == nil || endTime == nil {
		return nil
	}

	return &TripTimeRange{
		TripID:    tripID,
		BlockID:   tripBlock.BlockID,
		ServiceID: tripBlock.ServiceID,
		StartTime: *startTime,
		EndTime:   *endTime,
		RowNumber: tripBlock.RowNumber,
	}
}

// loadServiceDates expands calendar.txt and calendar_dates.txt into the active
// dates for each service. Keeping an entry for services with no active dates
// lets the overlap check distinguish an inactive service from an unknown one.
func (v *BlockOverlappingValidator) loadServiceDates(loader *parser.FeedLoader) map[string]map[int64]struct{} {
	calendarValidator := NewDateTripsValidator()
	services := calendarValidator.loadServices(loader)
	exceptionsByService := groupExceptionsByService(calendarValidator.loadCalendarExceptions(loader))
	serviceDates := make(map[string]map[int64]struct{}, len(services)+len(exceptionsByService))

	for _, serviceID := range serviceIDsOf(services, exceptionsByService) {
		dates := calendarValidator.datesForService(services[serviceID], exceptionsByService[serviceID])
		activeDates := make(map[int64]struct{}, len(dates))
		for date := range dates {
			activeDates[date] = struct{}{}
		}
		serviceDates[serviceID] = activeDates
	}

	return serviceDates
}

// validateBlockOverlaps validates that trips in the same block don't overlap
func (v *BlockOverlappingValidator) validateBlockOverlaps(container *notice.NoticeContainer, tripTimeRanges []TripTimeRange, serviceDates map[string]map[int64]struct{}) {
	// Group trips by block ID
	blockTrips := make(map[string][]TripTimeRange)

	for _, tripRange := range tripTimeRanges {
		blockTrips[tripRange.BlockID] = append(blockTrips[tripRange.BlockID], tripRange)
	}

	// Check each block for overlapping trips
	for blockID, trips := range blockTrips {
		if len(trips) < 2 {
			continue // Need at least 2 trips to have overlaps
		}

		v.validateBlockServiceConsistency(container, blockID, trips)
		v.validateBlockTripOverlaps(container, blockID, trips, serviceDates)
	}
}

// validateBlockServiceConsistency reports blocks whose trips run on different
// services. A block is a vehicle working through the day, so its trips must
// share a calendar; OTP cannot interline them otherwise.
func (v *BlockOverlappingValidator) validateBlockServiceConsistency(container *notice.NoticeContainer, blockID string, trips []TripTimeRange) {
	first := trips[0]
	for _, trip := range trips[1:] {
		if trip.ServiceID == first.ServiceID {
			continue
		}
		container.AddNotice(notice.NewBlockServiceMismatchNotice(
			blockID,
			first.TripID,
			first.ServiceID,
			trip.TripID,
			trip.ServiceID,
			trip.RowNumber,
		))
	}
}

// validateBlockTripOverlaps validates overlaps within a single block
func (v *BlockOverlappingValidator) validateBlockTripOverlaps(container *notice.NoticeContainer, blockID string, trips []TripTimeRange, serviceDates map[string]map[int64]struct{}) {
	// Sort trips by start time for easier comparison
	sort.Slice(trips, func(i, j int) bool {
		return trips[i].StartTime < trips[j].StartTime
	})

	// Check each pair of trips for overlaps
	for i := 0; i < len(trips); i++ {
		for j := i + 1; j < len(trips); j++ {
			trip1 := &trips[i]
			trip2 := &trips[j]

			if v.servicesOverlap(trip1.ServiceID, trip2.ServiceID, serviceDates) && v.tripsOverlap(trip1, trip2) {
				container.AddNotice(notice.NewBlockTripsOverlapNotice(
					blockID,
					trip1.TripID,
					trip2.TripID,
					trip1.ServiceID,
					trip2.ServiceID,
					v.formatGTFSTime(trip1.StartTime),
					v.formatGTFSTime(trip1.EndTime),
					v.formatGTFSTime(trip2.StartTime),
					v.formatGTFSTime(trip2.EndTime),
					trip1.RowNumber,
					trip2.RowNumber,
				))
			}
		}
	}
}

// servicesOverlap reports whether two services are active on at least one
// common date. Unknown services are treated conservatively as overlapping;
// other validators report their missing calendar definitions.
func (v *BlockOverlappingValidator) servicesOverlap(serviceID1, serviceID2 string, serviceDates map[string]map[int64]struct{}) bool {
	dates1, known1 := serviceDates[serviceID1]
	dates2, known2 := serviceDates[serviceID2]
	if !known1 || !known2 {
		return true
	}

	if len(dates1) > len(dates2) {
		dates1, dates2 = dates2, dates1
	}
	for date := range dates1 {
		if _, found := dates2[date]; found {
			return true
		}
	}
	return false
}

// tripsOverlap checks if two trips overlap in time
func (v *BlockOverlappingValidator) tripsOverlap(trip1, trip2 *TripTimeRange) bool {
	// Trips overlap if one starts before the other ends
	return trip1.StartTime < trip2.EndTime && trip2.StartTime < trip1.EndTime
}

// formatGTFSTime formats seconds since midnight back to HH:MM:SS format
func (v *BlockOverlappingValidator) formatGTFSTime(seconds int) string {
	hours := seconds / 3600
	minutes := (seconds % 3600) / 60
	secs := seconds % 60
	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, secs)
}
