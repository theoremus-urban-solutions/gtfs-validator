package business

import (
	"io"
	"sort"
	"strconv"
	"strings"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/parser"
	"github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

// OverlappingFrequencyValidator validates frequency-based trips don't overlap
type OverlappingFrequencyValidator struct{}

// NewOverlappingFrequencyValidator creates a new overlapping frequency validator
func NewOverlappingFrequencyValidator() *OverlappingFrequencyValidator {
	return &OverlappingFrequencyValidator{}
}

// FrequencyEntry represents a frequency entry from frequencies.txt
type FrequencyEntry struct {
	TripID      string
	StartTime   int // seconds from midnight
	EndTime     int // seconds from midnight
	HeadwaySecs int
	ExactTimes  int
	RowNumber   int
}

// TripInfo represents basic trip information
type TripInfo struct {
	TripID    string
	RouteID   string
	ServiceID string
}

// Validate checks for overlapping frequency entries
func (v *OverlappingFrequencyValidator) Validate(loader *parser.FeedLoader, container *notice.NoticeContainer, config validator.Config) {
	frequencies := v.loadFrequencies(loader)
	if len(frequencies) == 0 {
		return // No frequencies to validate
	}

	trips := v.loadTripInfo(loader)
	
	// Group frequencies by trip
	tripFrequencies := make(map[string][]*FrequencyEntry)
	for _, freq := range frequencies {
		tripFrequencies[freq.TripID] = append(tripFrequencies[freq.TripID], freq)
	}

	// Validate each trip's frequencies
	for tripID, freqList := range tripFrequencies {
		v.validateTripFrequencies(container, tripID, freqList, trips[tripID])
	}

	// Also check for overlaps between different trips on same route/service
	v.validateCrossTripOverlaps(container, frequencies, trips)
}

// loadFrequencies loads frequency entries from frequencies.txt
func (v *OverlappingFrequencyValidator) loadFrequencies(loader *parser.FeedLoader) []*FrequencyEntry {
	var frequencies []*FrequencyEntry

	reader, err := loader.GetFile("frequencies.txt")
	if err != nil {
		return frequencies
	}
	defer reader.Close()

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
			continue
		}

		freq := v.parseFrequency(row)
		if freq != nil {
			frequencies = append(frequencies, freq)
		}
	}

	return frequencies
}

// parseFrequency parses a frequency entry
func (v *OverlappingFrequencyValidator) parseFrequency(row *parser.CSVRow) *FrequencyEntry {
	tripID, hasTripID := row.Values["trip_id"]
	startTimeStr, hasStartTime := row.Values["start_time"]
	endTimeStr, hasEndTime := row.Values["end_time"]
	headwayStr, hasHeadway := row.Values["headway_secs"]

	if !hasTripID || !hasStartTime || !hasEndTime || !hasHeadway {
		return nil
	}

	startTime := v.parseGTFSTime(strings.TrimSpace(startTimeStr))
	endTime := v.parseGTFSTime(strings.TrimSpace(endTimeStr))
	headway, err := strconv.Atoi(strings.TrimSpace(headwayStr))

	if startTime == -1 || endTime == -1 || err != nil {
		return nil
	}

	freq := &FrequencyEntry{
		TripID:      strings.TrimSpace(tripID),
		StartTime:   startTime,
		EndTime:     endTime,
		HeadwaySecs: headway,
		RowNumber:   row.RowNumber,
	}

	// Parse optional exact_times
	if exactTimesStr, hasExactTimes := row.Values["exact_times"]; hasExactTimes && strings.TrimSpace(exactTimesStr) != "" {
		if exactTimes, err := strconv.Atoi(strings.TrimSpace(exactTimesStr)); err == nil {
			freq.ExactTimes = exactTimes
		}
	}

	return freq
}

// loadTripInfo loads basic trip information
func (v *OverlappingFrequencyValidator) loadTripInfo(loader *parser.FeedLoader) map[string]*TripInfo {
	trips := make(map[string]*TripInfo)

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
			trips[trip.TripID] = trip
		}
	}

	return trips
}

// parseTrip parses basic trip information
func (v *OverlappingFrequencyValidator) parseTrip(row *parser.CSVRow) *TripInfo {
	tripID, hasTripID := row.Values["trip_id"]
	routeID, hasRouteID := row.Values["route_id"]
	serviceID, hasServiceID := row.Values["service_id"]

	if !hasTripID || !hasRouteID || !hasServiceID {
		return nil
	}

	return &TripInfo{
		TripID:    strings.TrimSpace(tripID),
		RouteID:   strings.TrimSpace(routeID),
		ServiceID: strings.TrimSpace(serviceID),
	}
}

// parseGTFSTime parses GTFS time format (HH:MM:SS) to seconds from midnight
func (v *OverlappingFrequencyValidator) parseGTFSTime(timeStr string) int {
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
func (v *OverlappingFrequencyValidator) formatGTFSTime(totalSeconds int) string {
	hours := totalSeconds / 3600
	minutes := (totalSeconds % 3600) / 60
	seconds := totalSeconds % 60
	return strings.TrimSpace(strings.Join([]string{
		strconv.Itoa(hours),
		strconv.Itoa(minutes),
		strconv.Itoa(seconds),
	}, ":"))
}

// validateTripFrequencies validates frequencies for a single trip
func (v *OverlappingFrequencyValidator) validateTripFrequencies(container *notice.NoticeContainer, tripID string, frequencies []*FrequencyEntry, trip *TripInfo) {
	if len(frequencies) <= 1 {
		return // No overlap possible
	}

	// Sort by start time
	sort.Slice(frequencies, func(i, j int) bool {
		return frequencies[i].StartTime < frequencies[j].StartTime
	})

	// Check for overlaps
	for i := 1; i < len(frequencies); i++ {
		prev := frequencies[i-1]
		curr := frequencies[i]

		if v.doFrequenciesOverlap(prev, curr) {
			container.AddNotice(notice.NewOverlappingFrequencyNotice(
				tripID,
				v.formatGTFSTime(prev.StartTime),
				v.formatGTFSTime(prev.EndTime),
				prev.RowNumber,
				v.formatGTFSTime(curr.StartTime),
				v.formatGTFSTime(curr.EndTime),
				curr.RowNumber,
			))
		}

		// Check for suspicious gaps or very close frequencies
		gap := curr.StartTime - prev.EndTime
		if gap > 0 && gap < 300 { // Gap < 5 minutes might be unintentional
			container.AddNotice(notice.NewSmallFrequencyGapNotice(
				tripID,
				v.formatGTFSTime(prev.EndTime),
				v.formatGTFSTime(curr.StartTime),
				gap,
				curr.RowNumber,
			))
		}
	}

	// Additional validations
	v.validateFrequencyConsistency(container, tripID, frequencies)
}

// doFrequenciesOverlap checks if two frequency entries overlap
func (v *OverlappingFrequencyValidator) doFrequenciesOverlap(freq1, freq2 *FrequencyEntry) bool {
	// Two time intervals overlap if: start1 < end2 AND start2 < end1
	return freq1.StartTime < freq2.EndTime && freq2.StartTime < freq1.EndTime
}

// validateFrequencyConsistency validates consistency within frequency entries
func (v *OverlappingFrequencyValidator) validateFrequencyConsistency(container *notice.NoticeContainer, tripID string, frequencies []*FrequencyEntry) {
	for _, freq := range frequencies {
		// Check for invalid time ranges
		if freq.EndTime <= freq.StartTime {
			container.AddNotice(notice.NewInvalidFrequencyTimeRangeNotice(
				tripID,
				v.formatGTFSTime(freq.StartTime),
				v.formatGTFSTime(freq.EndTime),
				freq.RowNumber,
			))
		}

		// Check for very short frequency periods
		duration := freq.EndTime - freq.StartTime
		if duration < freq.HeadwaySecs {
			container.AddNotice(notice.NewFrequencyDurationShorterThanHeadwayNotice(
				tripID,
				duration,
				freq.HeadwaySecs,
				freq.RowNumber,
			))
		}

		// Check for unreasonably long frequencies (> 24 hours)
		if duration > 86400 { // 24 hours
			container.AddNotice(notice.NewVeryLongFrequencyPeriodNotice(
				tripID,
				v.formatGTFSTime(freq.StartTime),
				v.formatGTFSTime(freq.EndTime),
				duration,
				freq.RowNumber,
			))
		}

		// Check for very short headways (< 1 minute) which might be errors
		if freq.HeadwaySecs < 60 {
			// Need route and service info for this notice, skip for now
			// This validation is already covered by the main FrequencyValidator
		}
	}
}

// validateCrossTripOverlaps validates overlaps between different trips
func (v *OverlappingFrequencyValidator) validateCrossTripOverlaps(container *notice.NoticeContainer, frequencies []*FrequencyEntry, trips map[string]*TripInfo) {
	// Group by route and service for comparison
	routeServiceFreqs := make(map[string][]*FrequencyEntry)

	for _, freq := range frequencies {
		if trip, exists := trips[freq.TripID]; exists {
			key := trip.RouteID + "_" + trip.ServiceID
			routeServiceFreqs[key] = append(routeServiceFreqs[key], freq)
		}
	}

	// Check for overlaps within each route/service group
	for _, freqList := range routeServiceFreqs {
		if len(freqList) <= 1 {
			continue
		}

		// Sort by start time
		sort.Slice(freqList, func(i, j int) bool {
			return freqList[i].StartTime < freqList[j].StartTime
		})

		// Check for overlaps between different trips
		for i := 0; i < len(freqList); i++ {
			for j := i + 1; j < len(freqList); j++ {
				freq1 := freqList[i]
				freq2 := freqList[j]

				if freq1.TripID != freq2.TripID && v.doFrequenciesOverlap(freq1, freq2) {
					trip1 := trips[freq1.TripID]

					container.AddNotice(notice.NewCrossTripFrequencyOverlapNotice(
						freq1.TripID,
						freq2.TripID,
						trip1.RouteID,
						trip1.ServiceID,
						v.formatGTFSTime(freq1.StartTime),
						v.formatGTFSTime(freq1.EndTime),
						v.formatGTFSTime(freq2.StartTime),
						v.formatGTFSTime(freq2.EndTime),
						freq2.RowNumber,
					))
				}
			}
		}
	}
}