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

// FrequencyValidator validates frequency-based services
type FrequencyValidator struct{}

// NewFrequencyValidator creates a new frequency validator
func NewFrequencyValidator() *FrequencyValidator {
	return &FrequencyValidator{}
}

// FrequencyInfo represents frequency information
type FrequencyInfo struct {
	TripID      string
	StartTime   int // seconds since midnight
	EndTime     int // seconds since midnight
	HeadwaySecs int
	ExactTimes  int // 0 or 1
	RowNumber   int
}

// Validate checks frequency definitions
func (v *FrequencyValidator) Validate(loader *parser.FeedLoader, container *notice.NoticeContainer, config validator.Config) {
	frequencies := v.loadFrequencies(loader)

	// Validate each frequency record
	for _, frequency := range frequencies {
		v.validateFrequency(container, frequency)
	}

	// Check for overlapping frequencies
	v.validateOverlappingFrequencies(container, frequencies)

	// Trip references, and overlaps between the trips sharing a route/service
	trips := v.loadTripInfo(loader)
	v.validateTripReferences(container, frequencies, trips)
	v.validateCrossTripOverlaps(container, frequencies, trips)
}

// TripInfo is the slice of trips.txt the frequency checks need: which route
// and service a frequency's trip belongs to, so overlaps can be compared
// between the trips that actually run alongside each other.
type TripInfo struct {
	TripID    string
	RouteID   string
	ServiceID string
}

// loadFrequencies loads frequency information from frequencies.txt
func (v *FrequencyValidator) loadFrequencies(loader *parser.FeedLoader) []*FrequencyInfo {
	var frequencies []*FrequencyInfo

	reader, err := loader.GetFile("frequencies.txt")
	if err != nil {
		return frequencies
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close reader %v", closeErr)
		}
	}()

	csvFile, err := parser.NewCSVFile(reader, "frequencies.txt")
	if err != nil {
		return frequencies
	}

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}

		frequency := v.parseFrequency(row)
		if frequency != nil {
			frequencies = append(frequencies, frequency)
		}
	}

	return frequencies
}

// parseFrequency parses a frequency record
func (v *FrequencyValidator) parseFrequency(row *parser.CSVRow) *FrequencyInfo {
	tripID, hasTripID := row.Values["trip_id"]
	startTimeStr, hasStartTime := row.Values["start_time"]
	endTimeStr, hasEndTime := row.Values["end_time"]
	headwaySecsStr, hasHeadwaySecs := row.Values["headway_secs"]

	if !hasTripID || !hasStartTime || !hasEndTime || !hasHeadwaySecs {
		return nil
	}

	frequency := &FrequencyInfo{
		TripID:    strings.TrimSpace(tripID),
		RowNumber: row.RowNumber,
	}

	// Parse start time
	if startTime, err := v.parseGTFSTime(strings.TrimSpace(startTimeStr)); err == nil {
		frequency.StartTime = startTime
	} else {
		return nil
	}

	// Parse end time
	if endTime, err := v.parseGTFSTime(strings.TrimSpace(endTimeStr)); err == nil {
		frequency.EndTime = endTime
	} else {
		return nil
	}

	// Parse headway seconds
	if headwaySecs, err := strconv.Atoi(strings.TrimSpace(headwaySecsStr)); err == nil {
		frequency.HeadwaySecs = headwaySecs
	} else {
		return nil
	}

	// Parse exact_times (optional, defaults to 0)
	if exactTimesStr, hasExactTimes := row.Values["exact_times"]; hasExactTimes && strings.TrimSpace(exactTimesStr) != "" {
		if exactTimes, err := strconv.Atoi(strings.TrimSpace(exactTimesStr)); err == nil {
			frequency.ExactTimes = exactTimes
		}
	}

	return frequency
}

// parseGTFSTime parses a GTFS time string (HH:MM:SS) into seconds since midnight
func (v *FrequencyValidator) parseGTFSTime(timeStr string) (int, error) {
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

// validateFrequency validates a single frequency record
func (v *FrequencyValidator) validateFrequency(container *notice.NoticeContainer, frequency *FrequencyInfo) {
	// Validate time range
	if frequency.StartTime >= frequency.EndTime {
		container.AddNotice(notice.NewInvalidFrequencyTimeRangeNotice(
			frequency.TripID,
			v.formatGTFSTime(frequency.StartTime),
			v.formatGTFSTime(frequency.EndTime),
			frequency.RowNumber,
		))
	}

	// Validate headway seconds
	if frequency.HeadwaySecs <= 0 {
		container.AddNotice(notice.NewInvalidHeadwayNotice(
			frequency.TripID,
			frequency.HeadwaySecs,
			frequency.RowNumber,
		))
	}

	// Validate exact_times field
	if frequency.ExactTimes != 0 && frequency.ExactTimes != 1 {
		container.AddNotice(notice.NewInvalidExactTimesNotice(
			frequency.TripID,
			frequency.ExactTimes,
			frequency.RowNumber,
		))
	}

	// A window shorter than one headway generates no trips at all.
	if duration := frequency.EndTime - frequency.StartTime; duration > 0 &&
		frequency.HeadwaySecs > 0 && duration < frequency.HeadwaySecs {
		container.AddNotice(notice.NewFrequencyDurationShorterThanHeadwayNotice(
			frequency.TripID,
			duration,
			frequency.HeadwaySecs,
			frequency.RowNumber,
		))
	}
}

// validateOverlappingFrequencies checks for overlapping frequency periods for the same trip
func (v *FrequencyValidator) validateOverlappingFrequencies(container *notice.NoticeContainer, frequencies []*FrequencyInfo) {
	// Group frequencies by trip_id
	tripFrequencies := make(map[string][]*FrequencyInfo)

	for _, frequency := range frequencies {
		tripFrequencies[frequency.TripID] = append(tripFrequencies[frequency.TripID], frequency)
	}

	// Check each trip for overlapping frequencies
	for tripID, tripFreqs := range tripFrequencies {
		if len(tripFreqs) < 2 {
			continue
		}

		// Sort by start time
		sort.Slice(tripFreqs, func(i, j int) bool {
			return tripFreqs[i].StartTime < tripFreqs[j].StartTime
		})

		// Check for overlaps
		for i := 0; i < len(tripFreqs)-1; i++ {
			current := tripFreqs[i]
			next := tripFreqs[i+1]

			if current.EndTime > next.StartTime {
				container.AddNotice(notice.NewOverlappingFrequencyNotice(
					tripID,
					v.formatGTFSTime(current.StartTime),
					v.formatGTFSTime(current.EndTime),
					current.RowNumber,
					v.formatGTFSTime(next.StartTime),
					v.formatGTFSTime(next.EndTime),
					next.RowNumber,
				))
			}
		}
	}
}

// validateTripReferences validates that frequency trips exist
func (v *FrequencyValidator) validateTripReferences(container *notice.NoticeContainer, frequencies []*FrequencyInfo, trips map[string]*TripInfo) {
	for _, frequency := range frequencies {
		if _, exists := trips[frequency.TripID]; !exists {
			container.AddNotice(notice.NewForeignKeyViolationNotice(
				"frequencies.txt",
				"trip_id",
				frequency.TripID,
				frequency.RowNumber,
				"trips.txt",
				"trip_id",
			))
		}
	}
}

// validateCrossTripOverlaps reports frequency windows that overlap between
// distinct trips of the same route and service. Overlapping windows on one
// trip are a contradiction; overlapping windows across sibling trips mean the
// route is described twice for the same period.
func (v *FrequencyValidator) validateCrossTripOverlaps(container *notice.NoticeContainer, frequencies []*FrequencyInfo, trips map[string]*TripInfo) {
	routeServiceFreqs := make(map[string][]*FrequencyInfo)
	for _, frequency := range frequencies {
		trip, exists := trips[frequency.TripID]
		if !exists {
			continue
		}
		key := trip.RouteID + "\x00" + trip.ServiceID
		routeServiceFreqs[key] = append(routeServiceFreqs[key], frequency)
	}

	for _, freqList := range routeServiceFreqs {
		if len(freqList) < 2 {
			continue
		}

		sort.Slice(freqList, func(i, j int) bool {
			return freqList[i].StartTime < freqList[j].StartTime
		})

		for i := 0; i < len(freqList); i++ {
			for j := i + 1; j < len(freqList); j++ {
				first, second := freqList[i], freqList[j]
				if first.TripID == second.TripID {
					continue
				}
				// Sorted by start time, so once a later window starts after
				// this one ends nothing further can overlap it.
				if second.StartTime >= first.EndTime {
					break
				}

				trip := trips[first.TripID]
				container.AddNotice(notice.NewCrossTripFrequencyOverlapNotice(
					first.TripID,
					second.TripID,
					trip.RouteID,
					trip.ServiceID,
					v.formatGTFSTime(first.StartTime),
					v.formatGTFSTime(first.EndTime),
					v.formatGTFSTime(second.StartTime),
					v.formatGTFSTime(second.EndTime),
					second.RowNumber,
				))
			}
		}
	}
}

// loadTripInfo loads the route and service of every trip.
func (v *FrequencyValidator) loadTripInfo(loader *parser.FeedLoader) map[string]*TripInfo {
	trips := make(map[string]*TripInfo)

	reader, err := loader.GetFile("trips.txt")
	if err != nil {
		return trips
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close reader %v", closeErr)
		}
	}()

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
			break
		}

		tripID, hasTripID := row.Values["trip_id"]
		if !hasTripID {
			continue
		}
		trips[strings.TrimSpace(tripID)] = &TripInfo{
			TripID:    strings.TrimSpace(tripID),
			RouteID:   strings.TrimSpace(row.Values["route_id"]),
			ServiceID: strings.TrimSpace(row.Values["service_id"]),
		}
	}

	return trips
}

// formatGTFSTime formats seconds since midnight back to HH:MM:SS format
func (v *FrequencyValidator) formatGTFSTime(seconds int) string {
	hours := seconds / 3600
	minutes := (seconds % 3600) / 60
	secs := seconds % 60
	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, secs)
}
