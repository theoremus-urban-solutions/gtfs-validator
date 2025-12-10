package relationship

import (
	"io"
	"log"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/parser"
	"github.com/theoremus-urban-solutions/gtfs-validator/schema"
	"github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

// StopTimeConsistencyValidator validates stop time consistency within trips
type StopTimeConsistencyValidator struct{}

// NewStopTimeConsistencyValidator creates a new stop time consistency validator
func NewStopTimeConsistencyValidator() *StopTimeConsistencyValidator {
	return &StopTimeConsistencyValidator{}
}

// StopTimeInfo represents stop time information for validation
type StopTimeInfo struct {
	TripID            string
	StopID            string
	StopSequence      int
	ArrivalTime       string
	DepartureTime     string
	StopHeadsign      string
	PickupType        *int
	DropOffType       *int
	ShapeDistTraveled *float64
	Timepoint         *int
	RowNumber         int
}

// Validate checks stop time consistency
func (v *StopTimeConsistencyValidator) Validate(loader *parser.FeedLoader, container *notice.NoticeContainer, config validator.Config) {
	var tripStopTimes map[string][]*StopTimeInfo

	// Try to use cache if available (Phase 1 optimization)
	if cache := loader.GetCache(); cache != nil {
		cachedStopTimes, err := cache.GetStopTimesByTrip()
		if err == nil {
			// Convert cached schema.StopTime to StopTimeInfo for validation
			tripStopTimes = v.convertFromCache(cachedStopTimes)
		}
	}

	// Fallback to direct file loading if cache unavailable
	if tripStopTimes == nil {
		stopTimes := v.loadStopTimes(loader)
		tripStopTimes = v.groupByTrip(stopTimes)
	}

	// Use parallel validation if configured (Phase 2 optimization)
	workers := config.ParallelWorkers
	if workers > 1 && len(tripStopTimes) > 10 {
		v.validateTripsParallel(container, tripStopTimes, workers)
	} else {
		// Sequential validation for small datasets or single worker
		for tripID, stopTimes := range tripStopTimes {
			v.validateTripStopTimes(container, tripID, stopTimes)
		}
	}
}

// convertFromCache converts cached schema.StopTime entries to StopTimeInfo for validation.
// The cache groups stop times by trip, so we don't need to re-group them.
func (v *StopTimeConsistencyValidator) convertFromCache(cachedData map[string][]*schema.StopTime) map[string][]*StopTimeInfo {
	result := make(map[string][]*StopTimeInfo, len(cachedData))

	for tripID, stopTimes := range cachedData {
		converted := make([]*StopTimeInfo, 0, len(stopTimes))
		for _, st := range stopTimes {
			info := &StopTimeInfo{
				TripID:        st.TripID,
				StopID:        st.StopID,
				StopSequence:  st.StopSequence,
				ArrivalTime:   st.ArrivalTime,
				DepartureTime: st.DepartureTime,
				StopHeadsign:  st.StopHeadsign,
			}

			// Convert optional fields
			if st.PickupType != "" {
				if pt, err := strconv.Atoi(st.PickupType); err == nil {
					info.PickupType = &pt
				}
			}
			if st.DropOffType != "" {
				if dt, err := strconv.Atoi(st.DropOffType); err == nil {
					info.DropOffType = &dt
				}
			}
			if st.ShapeDistTraveled != "" {
				if sd, err := strconv.ParseFloat(st.ShapeDistTraveled, 64); err == nil {
					info.ShapeDistTraveled = &sd
				}
			}
			if st.Timepoint != 0 {
				tp := st.Timepoint
				info.Timepoint = &tp
			}

			converted = append(converted, info)
		}

		// Sort by stop sequence (same as groupByTrip does)
		sort.Slice(converted, func(i, j int) bool {
			return converted[i].StopSequence < converted[j].StopSequence
		})

		result[tripID] = converted
	}

	return result
}

// validateTripsParallel validates trips in parallel using a worker pool.
// This is invoked when config.ParallelWorkers > 1 and there are enough trips to benefit from parallelization.
func (v *StopTimeConsistencyValidator) validateTripsParallel(
	container *notice.NoticeContainer,
	tripStopTimes map[string][]*StopTimeInfo,
	workers int,
) {
	// Convert map to slice for work distribution
	type tripWork struct {
		tripID    string
		stopTimes []*StopTimeInfo
	}

	work := make([]tripWork, 0, len(tripStopTimes))
	for tripID, stopTimes := range tripStopTimes {
		work = append(work, tripWork{tripID, stopTimes})
	}

	// Create work channel
	workChan := make(chan tripWork, 100)
	var wg sync.WaitGroup

	// Start worker goroutines
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for w := range workChan {
				v.validateTripStopTimes(container, w.tripID, w.stopTimes)
			}
		}()
	}

	// Distribute work
	for _, w := range work {
		workChan <- w
	}
	close(workChan)

	// Wait for all workers to complete
	wg.Wait()
}

// loadStopTimes loads stop time information
func (v *StopTimeConsistencyValidator) loadStopTimes(loader *parser.FeedLoader) []*StopTimeInfo {
	var stopTimes []*StopTimeInfo

	reader, err := loader.GetFile("stop_times.txt")
	if err != nil {
		return stopTimes
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close reader %v", closeErr)
		}
	}()

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
			break
		}

		stopTime := v.parseStopTime(row)
		if stopTime != nil {
			stopTimes = append(stopTimes, stopTime)
		}
	}

	return stopTimes
}

// parseStopTime parses a stop time record
func (v *StopTimeConsistencyValidator) parseStopTime(row *parser.CSVRow) *StopTimeInfo {
	tripID, hasTripID := row.Values["trip_id"]
	stopID, hasStopID := row.Values["stop_id"]
	stopSequenceStr, hasStopSequence := row.Values["stop_sequence"]

	if !hasTripID || !hasStopID || !hasStopSequence {
		return nil
	}

	stopSequence, err := strconv.Atoi(strings.TrimSpace(stopSequenceStr))
	if err != nil {
		return nil
	}

	stopTime := &StopTimeInfo{
		TripID:       strings.TrimSpace(tripID),
		StopID:       strings.TrimSpace(stopID),
		StopSequence: stopSequence,
		RowNumber:    row.RowNumber,
	}

	// Parse optional fields
	if arrivalTime, hasArrival := row.Values["arrival_time"]; hasArrival {
		stopTime.ArrivalTime = strings.TrimSpace(arrivalTime)
	}
	if departureTime, hasDeparture := row.Values["departure_time"]; hasDeparture {
		stopTime.DepartureTime = strings.TrimSpace(departureTime)
	}
	if stopHeadsign, hasHeadsign := row.Values["stop_headsign"]; hasHeadsign {
		stopTime.StopHeadsign = strings.TrimSpace(stopHeadsign)
	}

	// Parse pickup/drop-off types
	if pickupTypeStr, hasPickup := row.Values["pickup_type"]; hasPickup && strings.TrimSpace(pickupTypeStr) != "" {
		if pickupType, err := strconv.Atoi(strings.TrimSpace(pickupTypeStr)); err == nil {
			stopTime.PickupType = &pickupType
		}
	}
	if dropOffTypeStr, hasDropOff := row.Values["drop_off_type"]; hasDropOff && strings.TrimSpace(dropOffTypeStr) != "" {
		if dropOffType, err := strconv.Atoi(strings.TrimSpace(dropOffTypeStr)); err == nil {
			stopTime.DropOffType = &dropOffType
		}
	}

	// Parse shape_dist_traveled
	if shapeDistStr, hasShapeDist := row.Values["shape_dist_traveled"]; hasShapeDist && strings.TrimSpace(shapeDistStr) != "" {
		if shapeDist, err := strconv.ParseFloat(strings.TrimSpace(shapeDistStr), 64); err == nil {
			stopTime.ShapeDistTraveled = &shapeDist
		}
	}

	// Parse timepoint
	if timepointStr, hasTimepoint := row.Values["timepoint"]; hasTimepoint && strings.TrimSpace(timepointStr) != "" {
		if timepoint, err := strconv.Atoi(strings.TrimSpace(timepointStr)); err == nil {
			stopTime.Timepoint = &timepoint
		}
	}

	return stopTime
}

// groupByTrip groups stop times by trip ID
func (v *StopTimeConsistencyValidator) groupByTrip(stopTimes []*StopTimeInfo) map[string][]*StopTimeInfo {
	tripStopTimes := make(map[string][]*StopTimeInfo)

	for _, stopTime := range stopTimes {
		tripStopTimes[stopTime.TripID] = append(tripStopTimes[stopTime.TripID], stopTime)
	}

	// Sort each trip's stop times by sequence
	for _, stopTimes := range tripStopTimes {
		sort.Slice(stopTimes, func(i, j int) bool {
			return stopTimes[i].StopSequence < stopTimes[j].StopSequence
		})
	}

	return tripStopTimes
}

// validateTripStopTimes validates stop times for a single trip
func (v *StopTimeConsistencyValidator) validateTripStopTimes(container *notice.NoticeContainer, tripID string, stopTimes []*StopTimeInfo) {
	if len(stopTimes) == 0 {
		return
	}

	// Check for missing first/last times
	v.validateFirstLastTimes(container, tripID, stopTimes)

	// Check for duplicate stops
	v.validateDuplicateStops(container, tripID, stopTimes)

	// Check for arrival/departure consistency
	v.validateArrivalDepartureConsistency(container, stopTimes)

	// Check for timepoint consistency
	v.validateTimepointConsistency(container, stopTimes)

	// Check for pickup/drop-off consistency
	v.validatePickupDropoffConsistency(container, stopTimes)

	// Check for shape distance consistency
	v.validateShapeDistanceConsistency(container, stopTimes)
}

// validateFirstLastTimes checks that first and last stops have times
func (v *StopTimeConsistencyValidator) validateFirstLastTimes(container *notice.NoticeContainer, tripID string, stopTimes []*StopTimeInfo) {
	if len(stopTimes) == 0 {
		return
	}

	// Check first stop
	first := stopTimes[0]
	if first.ArrivalTime == "" && first.DepartureTime == "" {
		container.AddNotice(notice.NewMissingTripFirstTimeNotice(
			tripID,
			first.StopID,
			first.RowNumber,
		))
	}

	// Check last stop
	last := stopTimes[len(stopTimes)-1]
	if last.ArrivalTime == "" && last.DepartureTime == "" {
		container.AddNotice(notice.NewMissingTripLastTimeNotice(
			tripID,
			last.StopID,
			last.RowNumber,
		))
	}
}

// validateDuplicateStops checks for duplicate stops in a trip
func (v *StopTimeConsistencyValidator) validateDuplicateStops(container *notice.NoticeContainer, tripID string, stopTimes []*StopTimeInfo) {
	stopCount := make(map[string]int)

	for _, stopTime := range stopTimes {
		stopCount[stopTime.StopID]++
	}

	// Report stops that appear multiple times
	for stopID, count := range stopCount {
		if count > 1 {
			// Find the occurrences
			var occurrences []*StopTimeInfo
			for _, stopTime := range stopTimes {
				if stopTime.StopID == stopID {
					occurrences = append(occurrences, stopTime)
				}
			}

			// Check if it's a loop route (same stop at beginning and end)
			if count == 2 && occurrences[0].StopSequence == stopTimes[0].StopSequence &&
				occurrences[1].StopSequence == stopTimes[len(stopTimes)-1].StopSequence {
				// This is likely a loop route - less severe
				container.AddNotice(notice.NewLoopRouteNotice(
					tripID,
					stopID,
					occurrences[0].RowNumber,
					occurrences[1].RowNumber,
				))
			} else {
				// Multiple stops in the middle of the trip
				for i := 1; i < len(occurrences); i++ {
					container.AddNotice(notice.NewDuplicateStopInTripNotice(
						tripID,
						stopID,
						occurrences[i].StopSequence,
						occurrences[i].RowNumber,
					))
				}
			}
		}
	}
}

// validateArrivalDepartureConsistency checks arrival/departure time consistency
func (v *StopTimeConsistencyValidator) validateArrivalDepartureConsistency(container *notice.NoticeContainer, stopTimes []*StopTimeInfo) {
	for _, stopTime := range stopTimes {
		// Skip if both times are empty
		if stopTime.ArrivalTime == "" && stopTime.DepartureTime == "" {
			continue
		}

		// Check if only one time is provided
		if stopTime.ArrivalTime == "" && stopTime.DepartureTime != "" {
			container.AddNotice(notice.NewMissingArrivalTimeNotice(
				stopTime.TripID,
				stopTime.StopID,
				stopTime.StopSequence,
				stopTime.RowNumber,
			))
		} else if stopTime.ArrivalTime != "" && stopTime.DepartureTime == "" {
			container.AddNotice(notice.NewMissingDepartureTimeNotice(
				stopTime.TripID,
				stopTime.StopID,
				stopTime.StopSequence,
				stopTime.RowNumber,
			))
		}
	}
}

// validateTimepointConsistency checks timepoint field consistency
func (v *StopTimeConsistencyValidator) validateTimepointConsistency(container *notice.NoticeContainer, stopTimes []*StopTimeInfo) {
	for _, stopTime := range stopTimes {
		if stopTime.Timepoint == nil {
			continue
		}

		// Validate timepoint value
		if *stopTime.Timepoint != 0 && *stopTime.Timepoint != 1 {
			container.AddNotice(notice.NewInvalidTimepointNotice(
				stopTime.TripID,
				stopTime.StopID,
				*stopTime.Timepoint,
				stopTime.RowNumber,
			))
		}

		// Check if timepoint=0 but times are provided
		if *stopTime.Timepoint == 0 && (stopTime.ArrivalTime != "" || stopTime.DepartureTime != "") {
			// This is allowed but might be confusing
			container.AddNotice(notice.NewTimepointWithoutTimesNotice(
				stopTime.TripID,
				stopTime.StopID,
				stopTime.StopSequence,
				stopTime.RowNumber,
			))
		}
	}
}

// validatePickupDropoffConsistency checks pickup/drop-off type consistency
func (v *StopTimeConsistencyValidator) validatePickupDropoffConsistency(container *notice.NoticeContainer, stopTimes []*StopTimeInfo) {
	if len(stopTimes) < 2 {
		return
	}

	// Check first stop - shouldn't have pickup_type = 1 (no pickup)
	first := stopTimes[0]
	if first.PickupType != nil && *first.PickupType == 1 {
		container.AddNotice(notice.NewFirstStopNoPickupNotice(
			first.TripID,
			first.StopID,
			first.RowNumber,
		))
	}

	// Check last stop - shouldn't have drop_off_type = 1 (no drop-off)
	last := stopTimes[len(stopTimes)-1]
	if last.DropOffType != nil && *last.DropOffType == 1 {
		container.AddNotice(notice.NewLastStopNoDropOffNotice(
			last.TripID,
			last.StopID,
			last.RowNumber,
		))
	}

	// Check if all stops have no pickup or no drop-off
	allNoPickup := true
	allNoDropOff := true
	for _, stopTime := range stopTimes {
		if stopTime.PickupType == nil || *stopTime.PickupType != 1 {
			allNoPickup = false
		}
		if stopTime.DropOffType == nil || *stopTime.DropOffType != 1 {
			allNoDropOff = false
		}
	}

	if allNoPickup {
		container.AddNotice(notice.NewAllStopsNoPickupNotice(stopTimes[0].TripID))
	}
	if allNoDropOff {
		container.AddNotice(notice.NewAllStopsNoDropOffNotice(stopTimes[0].TripID))
	}
}

// validateShapeDistanceConsistency checks shape distance traveled consistency
func (v *StopTimeConsistencyValidator) validateShapeDistanceConsistency(container *notice.NoticeContainer, stopTimes []*StopTimeInfo) {
	hasAnyShapeDist := false
	for _, stopTime := range stopTimes {
		if stopTime.ShapeDistTraveled != nil {
			hasAnyShapeDist = true
			break
		}
	}

	if !hasAnyShapeDist {
		return
	}

	// Check that all stops have shape_dist_traveled if any do
	missingCount := 0
	for _, stopTime := range stopTimes {
		if stopTime.ShapeDistTraveled == nil {
			missingCount++
		}
	}

	if missingCount > 0 && missingCount < len(stopTimes) {
		// Some but not all have shape distances
		container.AddNotice(notice.NewInconsistentStopTimeShapeDistanceNotice(
			stopTimes[0].TripID,
			missingCount,
			len(stopTimes),
		))
	}
}
