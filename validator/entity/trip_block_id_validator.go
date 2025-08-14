package entity

import (
	"io"
	"strconv"
	"strings"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/parser"
	"github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

// TripBlockIdValidator validates block_id assignments in trips
type TripBlockIdValidator struct{}

// NewTripBlockIdValidator creates a new trip block ID validator
func NewTripBlockIdValidator() *TripBlockIdValidator {
	return &TripBlockIdValidator{}
}

// TripBlockInfo represents trip block information
type TripBlockInfo struct {
	TripID    string
	RouteID   string
	ServiceID string
	BlockID   string
	RowNumber int
}

// StopTimeInfo represents stop time information for block validation
type StopTimeInfo struct {
	TripID        string
	StopSequence  int
	ArrivalTime   int // seconds from midnight
	DepartureTime int // seconds from midnight
}

// Validate checks block_id assignments for consistency
func (v *TripBlockIdValidator) Validate(loader *parser.FeedLoader, container *notice.NoticeContainer, config validator.Config) {
	trips := v.loadTrips(loader)
	if len(trips) == 0 {
		return
	}

	// Group trips by block_id
	blockTrips := make(map[string][]*TripBlockInfo)
	for _, trip := range trips {
		if trip.BlockID != "" {
			blockTrips[trip.BlockID] = append(blockTrips[trip.BlockID], trip)
		}
	}

	// Validate each block
	for blockID, blockTripList := range blockTrips {
		v.validateBlock(container, blockID, blockTripList, loader)
	}
}

// loadTrips loads trip information from trips.txt
func (v *TripBlockIdValidator) loadTrips(loader *parser.FeedLoader) []*TripBlockInfo {
	var trips []*TripBlockInfo

	reader, err := loader.GetFile("trips.txt")
	if err != nil {
		return trips
	}
	defer reader.Close()

	csvFile, err := parser.NewCSVFile(reader, "trips.txt")
	if err != nil {
		return trips
	}

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		trip := v.parseTrip(row)
		if trip != nil {
			trips = append(trips, trip)
		}
	}

	return trips
}

// parseTrip parses trip information
func (v *TripBlockIdValidator) parseTrip(row *parser.CSVRow) *TripBlockInfo {
	tripID, hasTripID := row.Values["trip_id"]
	routeID, hasRouteID := row.Values["route_id"]
	serviceID, hasServiceID := row.Values["service_id"]

	if !hasTripID || !hasRouteID || !hasServiceID {
		return nil
	}

	trip := &TripBlockInfo{
		TripID:    strings.TrimSpace(tripID),
		RouteID:   strings.TrimSpace(routeID),
		ServiceID: strings.TrimSpace(serviceID),
		RowNumber: row.RowNumber,
	}

	// Parse optional block_id
	if blockID, hasBlockID := row.Values["block_id"]; hasBlockID {
		trip.BlockID = strings.TrimSpace(blockID)
	}

	return trip
}

// validateBlock validates a block of trips
func (v *TripBlockIdValidator) validateBlock(container *notice.NoticeContainer, blockID string, trips []*TripBlockInfo, loader *parser.FeedLoader) {
	if len(trips) < 2 {
		// Single trip in block - not necessarily a problem, but worth noting
		if len(trips) == 1 {
			container.AddNotice(notice.NewSingleTripBlockNotice(
				blockID,
				trips[0].TripID,
				trips[0].RowNumber,
			))
		}
		return
	}

	// Validate service consistency within block
	v.validateBlockServiceConsistency(container, blockID, trips)

	// Validate route consistency within block
	v.validateBlockRouteConsistency(container, blockID, trips)

	// Load stop times for temporal validation
	stopTimes := v.loadStopTimesForTrips(loader, trips)
	if len(stopTimes) > 0 {
		v.validateBlockTiming(container, blockID, trips, stopTimes)
	}
}

// validateBlockServiceConsistency checks if all trips in block have same service
func (v *TripBlockIdValidator) validateBlockServiceConsistency(container *notice.NoticeContainer, blockID string, trips []*TripBlockInfo) {
	if len(trips) < 2 {
		return
	}

	firstServiceID := trips[0].ServiceID
	for _, trip := range trips[1:] {
		if trip.ServiceID != firstServiceID {
			container.AddNotice(notice.NewBlockServiceMismatchNotice(
				blockID,
				trips[0].TripID,
				firstServiceID,
				trip.TripID,
				trip.ServiceID,
				trip.RowNumber,
			))
		}
	}
}

// validateBlockRouteConsistency checks route patterns within block
func (v *TripBlockIdValidator) validateBlockRouteConsistency(container *notice.NoticeContainer, blockID string, trips []*TripBlockInfo) {
	routeCount := make(map[string]int)
	for _, trip := range trips {
		routeCount[trip.RouteID]++
	}

	// Info notice if block spans multiple routes (common but worth noting)
	if len(routeCount) > 1 {
		var routeIDs []string
		for routeID := range routeCount {
			routeIDs = append(routeIDs, routeID)
		}

		container.AddNotice(notice.NewBlockMultipleRoutesNotice(
			blockID,
			routeIDs,
			len(trips),
		))
	}

	// Warning if too many trips in single block (potential performance issue)
	if len(trips) > 20 {
		container.AddNotice(notice.NewBlockTooManyTripsNotice(
			blockID,
			len(trips),
		))
	}
}

// loadStopTimesForTrips loads stop times for block validation
func (v *TripBlockIdValidator) loadStopTimesForTrips(loader *parser.FeedLoader, trips []*TripBlockInfo) map[string][]*StopTimeInfo {
	stopTimes := make(map[string][]*StopTimeInfo)

	// Create a map of trip IDs for quick lookup
	tripIDs := make(map[string]bool)
	for _, trip := range trips {
		tripIDs[trip.TripID] = true
	}

	reader, err := loader.GetFile("stop_times.txt")
	if err != nil {
		return stopTimes
	}
	defer reader.Close()

	csvFile, err := parser.NewCSVFile(reader, "stop_times.txt")
	if err != nil {
		return stopTimes
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
		if stopTime != nil && tripIDs[stopTime.TripID] {
			stopTimes[stopTime.TripID] = append(stopTimes[stopTime.TripID], stopTime)
		}
	}

	return stopTimes
}

// parseStopTime parses stop time information
func (v *TripBlockIdValidator) parseStopTime(row *parser.CSVRow) *StopTimeInfo {
	tripID, hasTripID := row.Values["trip_id"]
	seqStr, hasSeq := row.Values["stop_sequence"]
	arrivalStr, hasArrival := row.Values["arrival_time"]
	departureStr, hasDeparture := row.Values["departure_time"]

	if !hasTripID || !hasSeq {
		return nil
	}

	seq, err := strconv.Atoi(strings.TrimSpace(seqStr))
	if err != nil {
		return nil
	}

	stopTime := &StopTimeInfo{
		TripID:        strings.TrimSpace(tripID),
		StopSequence:  seq,
		ArrivalTime:   -1,
		DepartureTime: -1,
	}

	// Parse times if available
	if hasArrival && strings.TrimSpace(arrivalStr) != "" {
		stopTime.ArrivalTime = v.parseGTFSTime(strings.TrimSpace(arrivalStr))
	}
	if hasDeparture && strings.TrimSpace(departureStr) != "" {
		stopTime.DepartureTime = v.parseGTFSTime(strings.TrimSpace(departureStr))
	}

	return stopTime
}

// validateBlockTiming validates timing relationships within block
func (v *TripBlockIdValidator) validateBlockTiming(container *notice.NoticeContainer, blockID string, trips []*TripBlockInfo, stopTimes map[string][]*StopTimeInfo) {
	// Find trip end and start times
	tripTimes := make(map[string]struct {
		startTime int
		endTime   int
	})

	for _, trip := range trips {
		times, exists := stopTimes[trip.TripID]
		if !exists || len(times) == 0 {
			continue
		}

		// Find first and last valid times
		firstTime, lastTime := -1, -1

		for _, st := range times {
			if st.DepartureTime >= 0 && (firstTime == -1 || st.DepartureTime < firstTime) {
				firstTime = st.DepartureTime
			}
			if st.ArrivalTime >= 0 && st.ArrivalTime > lastTime {
				lastTime = st.ArrivalTime
			}
		}

		if firstTime >= 0 && lastTime >= 0 {
			tripTimes[trip.TripID] = struct {
				startTime int
				endTime   int
			}{firstTime, lastTime}
		}
	}

	// Check for temporal overlaps within block
	v.validateBlockOverlaps(container, blockID, tripTimes)
}

// validateBlockOverlaps checks for temporal overlaps within block
func (v *TripBlockIdValidator) validateBlockOverlaps(container *notice.NoticeContainer, blockID string, tripTimes map[string]struct {
	startTime int
	endTime   int
}) {
	if len(tripTimes) < 2 {
		return
	}

	// Convert to slice for comparison
	type tripTimeInfo struct {
		tripID    string
		startTime int
		endTime   int
	}

	var trips []tripTimeInfo
	for tripID, times := range tripTimes {
		trips = append(trips, tripTimeInfo{
			tripID:    tripID,
			startTime: times.startTime,
			endTime:   times.endTime,
		})
	}

	// Check for overlaps
	for i := 0; i < len(trips); i++ {
		for j := i + 1; j < len(trips); j++ {
			trip1 := trips[i]
			trip2 := trips[j]

			// Check if trips overlap in time
			if v.doTripsOverlap(trip1.startTime, trip1.endTime, trip2.startTime, trip2.endTime) {
				// Use existing notice signature
				container.AddNotice(notice.NewBlockTripsOverlapNotice(
					blockID,
					trip1.tripID,
					trip2.tripID,
					"", // service1ID - not available in this context
					"", // service2ID - not available in this context
					v.formatGTFSTime(trip1.startTime),
					v.formatGTFSTime(trip1.endTime),
					v.formatGTFSTime(trip2.startTime),
					v.formatGTFSTime(trip2.endTime),
					-1, // trip1RowNumber - not available in this context
					-1, // trip2RowNumber - not available in this context
				))
			}
		}
	}
}

// doTripsOverlap checks if two time ranges overlap
func (v *TripBlockIdValidator) doTripsOverlap(start1, end1, start2, end2 int) bool {
	return start1 < end2 && start2 < end1
}

// parseGTFSTime parses GTFS time format (HH:MM:SS) to seconds from midnight
func (v *TripBlockIdValidator) parseGTFSTime(timeStr string) int {
	parts := strings.Split(timeStr, ":")
	if len(parts) != 3 {
		return -1
	}

	hours, err1 := strconv.Atoi(parts[0])
	minutes, err2 := strconv.Atoi(parts[1])
	seconds, err3 := strconv.Atoi(parts[2])

	if err1 != nil || err2 != nil || err3 != nil {
		return -1
	}

	if minutes < 0 || minutes >= 60 || seconds < 0 || seconds >= 60 {
		return -1
	}

	return hours*3600 + minutes*60 + seconds
}

// formatGTFSTime formats seconds from midnight to GTFS time format
func (v *TripBlockIdValidator) formatGTFSTime(totalSeconds int) string {
	if totalSeconds < 0 {
		return "unknown"
	}

	hours := totalSeconds / 3600
	minutes := (totalSeconds % 3600) / 60
	seconds := totalSeconds % 60

	return strings.TrimSpace(strings.Join([]string{
		strconv.Itoa(hours),
		strconv.Itoa(minutes),
		strconv.Itoa(seconds),
	}, ":"))
}
