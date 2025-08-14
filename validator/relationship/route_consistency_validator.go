package relationship

import (
	"io"
	"log"
	"sort"
	"strconv"
	"strings"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/parser"
	"github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

// RouteConsistencyValidator validates route consistency and operational patterns
type RouteConsistencyValidator struct{}

// NewRouteConsistencyValidator creates a new route consistency validator
func NewRouteConsistencyValidator() *RouteConsistencyValidator {
	return &RouteConsistencyValidator{}
}

// RouteAnalysis represents comprehensive route analysis
type RouteAnalysis struct {
	RouteID         string
	RouteShortName  string
	RouteLongName   string
	RouteType       int
	AgencyID        string
	TripCount       int
	StopCount       int
	UniqueStopCount int
	ServiceCount    int
	DirectionCount  int
	TotalDistance   float64
	AverageDistance float64
	Trips           []*TripAnalysis
	StopUsage       map[string]int
	ServiceUsage    map[string]int
	RowNumber       int
}

// TripAnalysis represents trip-level analysis
type TripAnalysis struct {
	TripID         string
	ServiceID      string
	DirectionID    int
	StopCount      int
	StopPattern    []string
	PatternHash    string
	HasTimePoints  bool
	HasCoordinates bool
	RowNumber      int
}

// DirectionPattern represents direction-specific patterns
type DirectionPattern struct {
	DirectionID       int
	TripCount         int
	StopPatterns      map[string]int
	MostCommonPattern string
	PatternVariations int
}

// Validate performs comprehensive route consistency validation
func (v *RouteConsistencyValidator) Validate(loader *parser.FeedLoader, container *notice.NoticeContainer, config validator.Config) {
	// Load and analyze routes
	routes := v.loadRoutes(loader)
	if len(routes) == 0 {
		return
	}

	// Enhance route analysis with trip data
	v.enhanceWithTripData(loader, routes)

	// Enhance with stop time data
	v.enhanceWithStopTimeData(loader, routes)

	// Validate each route
	for _, route := range routes {
		v.validateRoute(container, route)
	}

	// Validate route network consistency
	v.validateRouteNetwork(container, routes)
}

// loadRoutes loads basic route information
func (v *RouteConsistencyValidator) loadRoutes(loader *parser.FeedLoader) map[string]*RouteAnalysis {
	routes := make(map[string]*RouteAnalysis)

	reader, err := loader.GetFile("routes.txt")
	if err != nil {
		return routes
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close reader %v", closeErr)
		}
	}()

	csvFile, err := parser.NewCSVFile(reader, "routes.txt")
	if err != nil {
		return routes
	}

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		route := v.parseRoute(row)
		if route != nil {
			routes[route.RouteID] = route
		}
	}

	return routes
}

// parseRoute parses a route record
func (v *RouteConsistencyValidator) parseRoute(row *parser.CSVRow) *RouteAnalysis {
	routeID, hasRouteID := row.Values["route_id"]
	if !hasRouteID {
		return nil
	}

	route := &RouteAnalysis{
		RouteID:      strings.TrimSpace(routeID),
		RowNumber:    row.RowNumber,
		StopUsage:    make(map[string]int),
		ServiceUsage: make(map[string]int),
		Trips:        []*TripAnalysis{},
	}

	if shortName, hasShortName := row.Values["route_short_name"]; hasShortName {
		route.RouteShortName = strings.TrimSpace(shortName)
	}
	if longName, hasLongName := row.Values["route_long_name"]; hasLongName {
		route.RouteLongName = strings.TrimSpace(longName)
	}
	if agencyID, hasAgencyID := row.Values["agency_id"]; hasAgencyID {
		route.AgencyID = strings.TrimSpace(agencyID)
	}
	if routeTypeStr, hasRouteType := row.Values["route_type"]; hasRouteType {
		if routeType, err := strconv.Atoi(strings.TrimSpace(routeTypeStr)); err == nil {
			route.RouteType = routeType
		}
	}

	return route
}

// enhanceWithTripData enhances route analysis with trip information
func (v *RouteConsistencyValidator) enhanceWithTripData(loader *parser.FeedLoader, routes map[string]*RouteAnalysis) {
	reader, err := loader.GetFile("trips.txt")
	if err != nil {
		return
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close reader %v", closeErr)
		}
	}()

	csvFile, err := parser.NewCSVFile(reader, "trips.txt")
	if err != nil {
		return
	}

	directionCounts := make(map[string]map[int]int)

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		trip := v.parseTripForRoute(row)
		if trip == nil {
			continue
		}

		if route, exists := routes[trip.RouteID]; exists {
			// Convert TripForRoute to TripAnalysis
			tripAnalysis := &TripAnalysis{
				TripID:      trip.TripID,
				ServiceID:   trip.ServiceID,
				DirectionID: trip.DirectionID,
				RowNumber:   trip.RowNumber,
			}
			route.Trips = append(route.Trips, tripAnalysis)
			route.TripCount++
			route.ServiceUsage[trip.ServiceID]++

			// Track direction usage
			if directionCounts[trip.RouteID] == nil {
				directionCounts[trip.RouteID] = make(map[int]int)
			}
			directionCounts[trip.RouteID][trip.DirectionID]++
		}
	}

	// Update direction counts
	for routeID, route := range routes {
		if directionMap, exists := directionCounts[routeID]; exists {
			route.DirectionCount = len(directionMap)
			route.ServiceCount = len(route.ServiceUsage)
		}
	}
}

// TripForRoute represents basic trip information for route analysis
type TripForRoute struct {
	RouteID     string
	TripID      string
	ServiceID   string
	DirectionID int
	RowNumber   int
}

// parseTripForRoute parses trip information needed for route analysis
func (v *RouteConsistencyValidator) parseTripForRoute(row *parser.CSVRow) *TripForRoute {
	routeID, hasRouteID := row.Values["route_id"]
	tripID, hasTripID := row.Values["trip_id"]
	serviceID, hasServiceID := row.Values["service_id"]

	if !hasRouteID || !hasTripID || !hasServiceID {
		return nil
	}

	trip := &TripForRoute{
		RouteID:   strings.TrimSpace(routeID),
		TripID:    strings.TrimSpace(tripID),
		ServiceID: strings.TrimSpace(serviceID),
		RowNumber: row.RowNumber,
	}

	if directionStr, hasDirection := row.Values["direction_id"]; hasDirection && strings.TrimSpace(directionStr) != "" {
		if direction, err := strconv.Atoi(strings.TrimSpace(directionStr)); err == nil {
			trip.DirectionID = direction
		}
	}

	return trip
}

// enhanceWithStopTimeData enhances route analysis with stop time patterns
func (v *RouteConsistencyValidator) enhanceWithStopTimeData(loader *parser.FeedLoader, routes map[string]*RouteAnalysis) {
	reader, err := loader.GetFile("stop_times.txt")
	if err != nil {
		return
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close reader %v", closeErr)
		}
	}()

	csvFile, err := parser.NewCSVFile(reader, "stop_times.txt")
	if err != nil {
		return
	}

	// Group stop times by trip
	tripStops := make(map[string][]string)
	tripTimePoints := make(map[string]bool)

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		stopTime := v.parseStopTimeForRoute(row)
		if stopTime == nil {
			continue
		}

		tripStops[stopTime.TripID] = append(tripStops[stopTime.TripID], stopTime.StopID)

		// Check for timepoints
		if stopTime.ArrivalTime != "" || stopTime.DepartureTime != "" {
			tripTimePoints[stopTime.TripID] = true
		}
	}

	// Sort stops by sequence and update trip patterns
	for tripID, stops := range tripStops {
		// Find the route for this trip
		var routeID string
		for _, route := range routes {
			for _, trip := range route.Trips {
				if trip.TripID == tripID {
					routeID = route.RouteID
					trip.StopPattern = stops
					trip.StopCount = len(stops)
					trip.PatternHash = strings.Join(stops, "|")
					trip.HasTimePoints = tripTimePoints[tripID]
					break
				}
			}
			if routeID != "" {
				break
			}
		}

		// Update route-level statistics
		if routeID != "" {
			route := routes[routeID]
			for _, stopID := range stops {
				route.StopUsage[stopID]++
			}
		}
	}

	// Calculate route-level stop statistics
	for _, route := range routes {
		route.StopCount = len(route.StopUsage)
		route.UniqueStopCount = len(route.StopUsage)
	}
}

// StopTimeForRoute represents stop time information for route analysis
type StopTimeForRoute struct {
	TripID        string
	StopID        string
	StopSequence  int
	ArrivalTime   string
	DepartureTime string
}

// parseStopTimeForRoute parses stop time information for route analysis
func (v *RouteConsistencyValidator) parseStopTimeForRoute(row *parser.CSVRow) *StopTimeForRoute {
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

	stopTime := &StopTimeForRoute{
		TripID:       strings.TrimSpace(tripID),
		StopID:       strings.TrimSpace(stopID),
		StopSequence: stopSeq,
	}

	if arrivalTime, hasArrival := row.Values["arrival_time"]; hasArrival {
		stopTime.ArrivalTime = strings.TrimSpace(arrivalTime)
	}
	if departureTime, hasDeparture := row.Values["departure_time"]; hasDeparture {
		stopTime.DepartureTime = strings.TrimSpace(departureTime)
	}

	return stopTime
}

// validateRoute validates an individual route
func (v *RouteConsistencyValidator) validateRoute(container *notice.NoticeContainer, route *RouteAnalysis) {
	// Check route naming
	v.validateRouteNaming(container, route)

	// Check route trip patterns
	v.validateRouteTripPatterns(container, route)

	// Check route service coverage
	v.validateRouteServiceCoverage(container, route)

	// Check route operational efficiency
	v.validateRouteEfficiency(container, route)
}

// validateRouteNaming validates route naming consistency
func (v *RouteConsistencyValidator) validateRouteNaming(container *notice.NoticeContainer, route *RouteAnalysis) {
	// Check for routes without names
	if route.RouteShortName == "" && route.RouteLongName == "" {
		container.AddNotice(notice.NewMissingRouteNameNotice(
			route.RouteID,
			route.RowNumber,
		))
	}

	// Check for very long route names
	if len(route.RouteShortName) > 12 {
		container.AddNotice(notice.NewRouteShortNameTooLongNotice(
			route.RouteID,
			route.RouteShortName,
			len(route.RouteShortName),
			12,
			route.RowNumber,
		))
	}

	if len(route.RouteLongName) > 120 {
		container.AddNotice(notice.NewRouteLongNameTooLongNotice(
			route.RouteID,
			route.RouteLongName,
			len(route.RouteLongName),
			120,
			route.RowNumber,
		))
	}

	// Check for identical short and long names
	if route.RouteShortName != "" && route.RouteLongName != "" {
		if route.RouteShortName == route.RouteLongName {
			container.AddNotice(notice.NewSameNameAndDescriptionNotice(
				route.RouteID,
				"route_short_name",
				"route_long_name",
				route.RouteShortName,
				route.RowNumber,
			))
		}
	}
}

// validateRouteTripPatterns validates trip patterns for the route
func (v *RouteConsistencyValidator) validateRouteTripPatterns(container *notice.NoticeContainer, route *RouteAnalysis) {
	if len(route.Trips) == 0 {
		container.AddNotice(notice.NewRouteWithoutTripsNotice(
			route.RouteID,
			route.RowNumber,
		))
		return
	}

	// Analyze patterns by direction
	directionPatterns := make(map[int]*DirectionPattern)

	for _, trip := range route.Trips {
		if directionPatterns[trip.DirectionID] == nil {
			directionPatterns[trip.DirectionID] = &DirectionPattern{
				DirectionID:  trip.DirectionID,
				StopPatterns: make(map[string]int),
			}
		}

		pattern := directionPatterns[trip.DirectionID]
		pattern.TripCount++
		if trip.PatternHash != "" {
			pattern.StopPatterns[trip.PatternHash]++
		}
	}

	// Validate direction patterns
	for directionID, pattern := range directionPatterns {
		// Find most common pattern
		maxCount := 0
		for patternHash, count := range pattern.StopPatterns {
			if count > maxCount {
				maxCount = count
				pattern.MostCommonPattern = patternHash
			}
		}

		pattern.PatternVariations = len(pattern.StopPatterns)

		// Check for excessive pattern variations
		if pattern.PatternVariations > 5 && pattern.TripCount > 10 {
			container.AddNotice(notice.NewExcessiveRoutePatternVariationsNotice(
				route.RouteID,
				directionID,
				pattern.PatternVariations,
				pattern.TripCount,
			))
		}

		// Check for single trip patterns
		for patternHash, count := range pattern.StopPatterns {
			if count == 1 && pattern.TripCount > 5 {
				container.AddNotice(notice.NewSingleTripPatternNotice(
					"route_"+route.RouteID+"_dir_"+strconv.Itoa(directionID),
					"", // No specific trip ID available here
					len(strings.Split(patternHash, "|")),
				))
				break // Only report once per direction
			}
		}
	}

	// Check direction balance
	if len(directionPatterns) == 2 {
		directions := make([]*DirectionPattern, 0, 2)
		for _, pattern := range directionPatterns {
			directions = append(directions, pattern)
		}

		ratio := float64(directions[0].TripCount) / float64(directions[1].TripCount)
		if ratio > 3.0 || ratio < 0.33 {
			container.AddNotice(notice.NewUnbalancedDirectionTripsNotice(
				route.RouteID,
				directions[0].DirectionID,
				directions[0].TripCount,
				directions[1].DirectionID,
				directions[1].TripCount,
			))
		}
	}
}

// validateRouteServiceCoverage validates service coverage for the route
func (v *RouteConsistencyValidator) validateRouteServiceCoverage(container *notice.NoticeContainer, route *RouteAnalysis) {
	// Check for routes with too many services
	if route.ServiceCount > 15 {
		container.AddNotice(notice.NewExcessiveServiceVarietyNotice(
			route.RouteID,
			route.ServiceCount,
		))
	}

	// Check for routes with very few services
	if route.ServiceCount == 1 && route.TripCount > 20 {
		container.AddNotice(notice.NewLimitedServiceVarietyNotice(
			route.RouteID,
			route.ServiceCount,
			route.TripCount,
		))
	}
}

// validateRouteEfficiency validates operational efficiency
func (v *RouteConsistencyValidator) validateRouteEfficiency(container *notice.NoticeContainer, route *RouteAnalysis) {
	// Check for routes with very few trips
	if route.TripCount < 3 {
		container.AddNotice(notice.NewLowRouteUsageNotice(
			route.RouteID,
			route.TripCount,
			route.ServiceCount,
		))
	}

	// Check for routes with too many stops
	if route.StopCount > 100 {
		container.AddNotice(notice.NewVeryLongRouteNotice(
			route.RouteID,
			route.StopCount,
			route.TripCount,
		))
	}

	// Check for routes with very few stops
	if route.StopCount < 2 {
		container.AddNotice(notice.NewVeryShortRouteNotice(
			route.RouteID,
			route.StopCount,
			route.TripCount,
		))
	}

	// Check timepoint coverage
	tripsWithTimepoints := 0
	for _, trip := range route.Trips {
		if trip.HasTimePoints {
			tripsWithTimepoints++
		}
	}

	if route.TripCount > 0 {
		timepointCoverage := float64(tripsWithTimepoints) / float64(route.TripCount)
		if timepointCoverage < 0.5 {
			container.AddNotice(notice.NewLowTimepointCoverageNotice(
				route.RouteID,
				tripsWithTimepoints,
				route.TripCount,
				timepointCoverage,
			))
		}
	}
}

// validateRouteNetwork validates route network consistency
func (v *RouteConsistencyValidator) validateRouteNetwork(container *notice.NoticeContainer, routes map[string]*RouteAnalysis) {
	// Collect network statistics
	totalRoutes := len(routes)
	totalTrips := 0
	routesByType := make(map[int]int)
	routesByAgency := make(map[string]int)

	for _, route := range routes {
		totalTrips += route.TripCount
		routesByType[route.RouteType]++
		if route.AgencyID != "" {
			routesByAgency[route.AgencyID]++
		}
	}

	// Generate network summary
	if totalRoutes > 0 {
		avgTripsPerRoute := float64(totalTrips) / float64(totalRoutes)

		// Sort route types by frequency
		type routeTypeCount struct {
			routeType int
			count     int
		}
		var sortedTypes []routeTypeCount
		for routeType, count := range routesByType {
			sortedTypes = append(sortedTypes, routeTypeCount{routeType, count})
		}
		sort.Slice(sortedTypes, func(i, j int) bool {
			return sortedTypes[i].count > sortedTypes[j].count
		})

		container.AddNotice(notice.NewRouteNetworkSummaryNotice(
			totalRoutes,
			totalTrips,
			avgTripsPerRoute,
			len(routesByType),
			len(routesByAgency),
		))
	}

	// Check for route type consistency within agencies
	for agencyID, agencyRouteCount := range routesByAgency {
		if agencyRouteCount > 20 {
			// Large agency - check for type diversity
			agencyTypes := make(map[int]int)
			for _, route := range routes {
				if route.AgencyID == agencyID {
					agencyTypes[route.RouteType]++
				}
			}

			if len(agencyTypes) > 5 {
				container.AddNotice(notice.NewHighRouteTypeDiversityNotice(
					agencyID,
					len(agencyTypes),
					agencyRouteCount,
				))
			}
		}
	}
}
