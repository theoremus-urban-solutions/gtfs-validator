package business

import (
	"fmt"
	"io"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/parser"
	"github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

// TravelSpeedValidator validates that travel speeds between stops are reasonable
type TravelSpeedValidator struct{}

// NewTravelSpeedValidator creates a new travel speed validator
func NewTravelSpeedValidator() *TravelSpeedValidator {
	return &TravelSpeedValidator{}
}

// StopTimeWithLocation represents a stop time with geographic location
type StopTimeWithLocation struct {
	TripID        string
	StopID        string
	StopSequence  int
	ArrivalTime   *int // seconds since midnight
	DepartureTime *int // seconds since midnight
	Latitude      *float64
	Longitude     *float64
	RowNumber     int
}

// RouteTypeSpeedLimits defines speed limits by route type (km/h)
var RouteTypeSpeedLimits = map[int]float64{
	0:  500.0, // Tram, Streetcar, Light rail
	1:  500.0, // Subway, Metro
	2:  500.0, // Rail
	3:  150.0, // Bus
	4:  100.0, // Ferry
	5:  150.0, // Cable tram
	6:  50.0,  // Aerial lift, suspended cable car
	7:  150.0, // Funicular
	11: 150.0, // Trolleybus
	12: 500.0, // Monorail
}

// Validate checks travel speeds between consecutive stops
func (v *TravelSpeedValidator) Validate(loader *parser.FeedLoader, container *notice.NoticeContainer, config validator.Config) {
	// Load stop locations
	stopLocations := v.loadStopLocations(loader)
	if len(stopLocations) == 0 {
		return // No stop location data available
	}

	// Load route types
	routeTypes := v.loadRouteTypes(loader)

	// Process stop times
	v.validateStopTimeSpeeds(loader, container, stopLocations, routeTypes)
}

// loadStopLocations loads stop coordinates from stops.txt
func (v *TravelSpeedValidator) loadStopLocations(loader *parser.FeedLoader) map[string]*StopLocation {
	stopLocations := make(map[string]*StopLocation)

	reader, err := loader.GetFile("stops.txt")
	if err != nil {
		return stopLocations
	}
	defer reader.Close()

	csvFile, err := parser.NewCSVFile(reader, "stops.txt")
	if err != nil {
		return stopLocations
	}

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}

		stopID, hasStopID := row.Values["stop_id"]
		latStr, hasLat := row.Values["stop_lat"]
		lonStr, hasLon := row.Values["stop_lon"]

		if !hasStopID || !hasLat || !hasLon {
			continue
		}

		lat, latErr := strconv.ParseFloat(strings.TrimSpace(latStr), 64)
		lon, lonErr := strconv.ParseFloat(strings.TrimSpace(lonStr), 64)

		if latErr == nil && lonErr == nil {
			stopLocations[strings.TrimSpace(stopID)] = &StopLocation{
				Latitude:  lat,
				Longitude: lon,
			}
		}
	}

	return stopLocations
}

// StopLocation represents a stop's geographic location
type StopLocation struct {
	Latitude  float64
	Longitude float64
}

// loadRouteTypes loads route types from routes.txt and trips.txt
func (v *TravelSpeedValidator) loadRouteTypes(loader *parser.FeedLoader) map[string]int {
	routeTypes := make(map[string]int)
	tripRoutes := make(map[string]string)

	// Load route types from routes.txt
	if reader, err := loader.GetFile("routes.txt"); err == nil {
		defer reader.Close()
		if csvFile, err := parser.NewCSVFile(reader, "routes.txt"); err == nil {
			for {
				row, err := csvFile.ReadRow()
				if err == io.EOF {
					break
				}
				if err != nil {
					break
				}

				routeID, hasRouteID := row.Values["route_id"]
				routeTypeStr, hasRouteType := row.Values["route_type"]

				if hasRouteID && hasRouteType {
					if routeType, err := strconv.Atoi(strings.TrimSpace(routeTypeStr)); err == nil {
						routeTypes[strings.TrimSpace(routeID)] = routeType
					}
				}
			}
		}
	}

	// Load trip-to-route mappings from trips.txt
	if reader, err := loader.GetFile("trips.txt"); err == nil {
		defer reader.Close()
		if csvFile, err := parser.NewCSVFile(reader, "trips.txt"); err == nil {
			for {
				row, err := csvFile.ReadRow()
				if err == io.EOF {
					break
				}
				if err != nil {
					break
				}

				tripID, hasTripID := row.Values["trip_id"]
				routeID, hasRouteID := row.Values["route_id"]

				if hasTripID && hasRouteID {
					tripRoutes[strings.TrimSpace(tripID)] = strings.TrimSpace(routeID)
				}
			}
		}
	}

	// Map trip IDs to route types
	tripRouteTypes := make(map[string]int)
	for tripID, routeID := range tripRoutes {
		if routeType, exists := routeTypes[routeID]; exists {
			tripRouteTypes[tripID] = routeType
		}
	}

	return tripRouteTypes
}

// validateStopTimeSpeeds validates travel speeds in stop_times.txt
func (v *TravelSpeedValidator) validateStopTimeSpeeds(loader *parser.FeedLoader, container *notice.NoticeContainer, stopLocations map[string]*StopLocation, routeTypes map[string]int) {
	reader, err := loader.GetFile("stop_times.txt")
	if err != nil {
		return
	}
	defer reader.Close()

	csvFile, err := parser.NewCSVFile(reader, "stop_times.txt")
	if err != nil {
		return
	}

	// Group stop times by trip_id
	tripStopTimes := make(map[string][]StopTimeWithLocation)

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}

		stopTime := v.parseStopTimeWithLocation(row, stopLocations)
		if stopTime != nil {
			tripStopTimes[stopTime.TripID] = append(tripStopTimes[stopTime.TripID], *stopTime)
		}
	}

	// Validate each trip's travel speeds
	for tripID, stopTimes := range tripStopTimes {
		routeType := routeTypes[tripID]
		v.validateTripTravelSpeeds(container, tripID, stopTimes, routeType)
	}
}

// parseStopTimeWithLocation parses a stop time row with location data
func (v *TravelSpeedValidator) parseStopTimeWithLocation(row *parser.CSVRow, stopLocations map[string]*StopLocation) *StopTimeWithLocation {
	tripID, hasTripID := row.Values["trip_id"]
	stopID, hasStopID := row.Values["stop_id"]
	stopSeqStr, hasStopSeq := row.Values["stop_sequence"]
	arrivalTimeStr, hasArrivalTime := row.Values["arrival_time"]
	departureTimeStr, hasDepartureTime := row.Values["departure_time"]

	if !hasTripID || !hasStopID || !hasStopSeq {
		return nil
	}

	stopSequence, err := strconv.Atoi(strings.TrimSpace(stopSeqStr))
	if err != nil {
		return nil
	}

	stopIDTrimmed := strings.TrimSpace(stopID)
	stopLocation, hasLocation := stopLocations[stopIDTrimmed]
	if !hasLocation {
		return nil // Skip stops without location data
	}

	stopTime := &StopTimeWithLocation{
		TripID:       strings.TrimSpace(tripID),
		StopID:       stopIDTrimmed,
		StopSequence: stopSequence,
		Latitude:     &stopLocation.Latitude,
		Longitude:    &stopLocation.Longitude,
		RowNumber:    row.RowNumber,
	}

	// Parse times (similar to previous validator)
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
func (v *TravelSpeedValidator) parseGTFSTime(timeStr string) (int, error) {
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

// validateTripTravelSpeeds validates travel speeds for a single trip
func (v *TravelSpeedValidator) validateTripTravelSpeeds(container *notice.NoticeContainer, tripID string, stopTimes []StopTimeWithLocation, routeType int) {
	if len(stopTimes) < 2 {
		return
	}

	// Sort by stop_sequence
	sort.Slice(stopTimes, func(i, j int) bool {
		return stopTimes[i].StopSequence < stopTimes[j].StopSequence
	})

	// Get speed limit for this route type (default to bus speed if unknown)
	speedLimit, exists := RouteTypeSpeedLimits[routeType]
	if !exists {
		speedLimit = RouteTypeSpeedLimits[3] // Default to bus speed (150 km/h)
	}

	// Check consecutive stop pairs
	for i := 1; i < len(stopTimes); i++ {
		prev := &stopTimes[i-1]
		curr := &stopTimes[i]

		v.validateStopPairSpeed(container, tripID, prev, curr, speedLimit, routeType)
	}
}

// validateStopPairSpeed validates travel speed between two consecutive stops
func (v *TravelSpeedValidator) validateStopPairSpeed(container *notice.NoticeContainer, tripID string, prev, curr *StopTimeWithLocation, speedLimit float64, routeType int) {
	// Get departure time from previous stop
	var prevTime *int
	if prev.DepartureTime != nil {
		prevTime = prev.DepartureTime
	} else if prev.ArrivalTime != nil {
		prevTime = prev.ArrivalTime
	}

	// Get arrival time at current stop
	var currTime *int
	if curr.ArrivalTime != nil {
		currTime = curr.ArrivalTime
	} else if curr.DepartureTime != nil {
		currTime = curr.DepartureTime
	}

	// Skip if we don't have both times
	if prevTime == nil || currTime == nil {
		return
	}

	// Skip if times are the same or backwards (other validators handle this)
	timeDiffSeconds := *currTime - *prevTime
	if timeDiffSeconds <= 0 {
		return
	}

	// Calculate distance using Haversine formula
	distance := v.haversineDistance(*prev.Latitude, *prev.Longitude, *curr.Latitude, *curr.Longitude)

	// Skip very short distances (< 10 meters) to avoid false positives
	if distance < 0.01 {
		return
	}

	// Calculate speed in km/h
	timeDiffHours := float64(timeDiffSeconds) / 3600.0
	speed := distance / timeDiffHours

	// Check if speed exceeds limit
	if speed > speedLimit {
		container.AddNotice(notice.NewExcessiveTravelSpeedNotice(
			tripID,
			prev.StopID,
			curr.StopID,
			prev.StopSequence,
			curr.StopSequence,
			speed,
			speedLimit,
			distance,
			timeDiffSeconds,
			routeType,
			prev.RowNumber,
			curr.RowNumber,
		))
	}
}

// haversineDistance calculates the great circle distance between two points in kilometers
func (v *TravelSpeedValidator) haversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371 // Earth's radius in kilometers

	// Convert degrees to radians
	lat1Rad := lat1 * math.Pi / 180
	lon1Rad := lon1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	lon2Rad := lon2 * math.Pi / 180

	// Haversine formula
	dlat := lat2Rad - lat1Rad
	dlon := lon2Rad - lon1Rad

	a := math.Sin(dlat/2)*math.Sin(dlat/2) + math.Cos(lat1Rad)*math.Cos(lat2Rad)*math.Sin(dlon/2)*math.Sin(dlon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c
}
