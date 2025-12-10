package parser

import (
	"fmt"
	"io"
	"log"
	"sync"

	"github.com/theoremus-urban-solutions/gtfs-validator/schema"
)

// ParsedFeedCache provides shared access to parsed GTFS data across validators.
// It loads files once and shares the parsed data, eliminating redundant file I/O and parsing.
// All methods are thread-safe for concurrent validator access.
type ParsedFeedCache struct {
	// Parsed data (loaded once, shared by all validators)
	stopTimes []*schema.StopTime
	trips     []*schema.Trip
	stops     []*schema.Stop
	routes    []*schema.Route

	// Pre-computed indexes (built on-demand)
	tripsByID       map[string]*schema.Trip
	stopsByID       map[string]*schema.Stop
	routesByID      map[string]*schema.Route
	tripsByRoute    map[string][]*schema.Trip
	stopTimesByTrip map[string][]*schema.StopTime

	// Lazy-loading state
	loadedFiles map[string]bool

	// Thread-safety
	mu sync.RWMutex

	// Reference to loader for on-demand file access
	loader *FeedLoader
}

// NewParsedFeedCache creates a new feed cache attached to the given loader.
func NewParsedFeedCache(loader *FeedLoader) *ParsedFeedCache {
	return &ParsedFeedCache{
		loadedFiles: make(map[string]bool),
		loader:      loader,
	}
}

// GetStopTimes returns all stop times from stop_times.txt.
// The data is loaded on first access and cached for subsequent calls.
// This method is thread-safe.
func (c *ParsedFeedCache) GetStopTimes() ([]*schema.StopTime, error) {
	// Fast path: already loaded (read lock)
	c.mu.RLock()
	if c.loadedFiles["stop_times.txt"] {
		result := c.stopTimes
		c.mu.RUnlock()
		return result, nil
	}
	c.mu.RUnlock()

	// Slow path: need to load (write lock)
	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check: another goroutine may have loaded it
	if c.loadedFiles["stop_times.txt"] {
		return c.stopTimes, nil
	}

	// Load the file
	stopTimes, err := c.loadStopTimesInternal()
	if err != nil {
		return nil, err
	}

	c.stopTimes = stopTimes
	c.loadedFiles["stop_times.txt"] = true
	return stopTimes, nil
}

// GetTrips returns all trips from trips.txt.
// The data is loaded on first access and cached for subsequent calls.
// This method is thread-safe.
func (c *ParsedFeedCache) GetTrips() ([]*schema.Trip, error) {
	// Fast path: already loaded (read lock)
	c.mu.RLock()
	if c.loadedFiles["trips.txt"] {
		result := c.trips
		c.mu.RUnlock()
		return result, nil
	}
	c.mu.RUnlock()

	// Slow path: need to load (write lock)
	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check: another goroutine may have loaded it
	if c.loadedFiles["trips.txt"] {
		return c.trips, nil
	}

	// Load the file
	trips, err := c.loadTripsInternal()
	if err != nil {
		return nil, err
	}

	c.trips = trips
	c.loadedFiles["trips.txt"] = true
	return trips, nil
}

// GetStops returns all stops from stops.txt.
// The data is loaded on first access and cached for subsequent calls.
// This method is thread-safe.
func (c *ParsedFeedCache) GetStops() ([]*schema.Stop, error) {
	// Fast path: already loaded (read lock)
	c.mu.RLock()
	if c.loadedFiles["stops.txt"] {
		result := c.stops
		c.mu.RUnlock()
		return result, nil
	}
	c.mu.RUnlock()

	// Slow path: need to load (write lock)
	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check: another goroutine may have loaded it
	if c.loadedFiles["stops.txt"] {
		return c.stops, nil
	}

	// Load the file
	stops, err := c.loadStopsInternal()
	if err != nil {
		return nil, err
	}

	c.stops = stops
	c.loadedFiles["stops.txt"] = true
	return stops, nil
}

// GetRoutes returns all routes from routes.txt.
// The data is loaded on first access and cached for subsequent calls.
// This method is thread-safe.
func (c *ParsedFeedCache) GetRoutes() ([]*schema.Route, error) {
	// Fast path: already loaded (read lock)
	c.mu.RLock()
	if c.loadedFiles["routes.txt"] {
		result := c.routes
		c.mu.RUnlock()
		return result, nil
	}
	c.mu.RUnlock()

	// Slow path: need to load (write lock)
	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check: another goroutine may have loaded it
	if c.loadedFiles["routes.txt"] {
		return c.routes, nil
	}

	// Load the file
	routes, err := c.loadRoutesInternal()
	if err != nil {
		return nil, err
	}

	c.routes = routes
	c.loadedFiles["routes.txt"] = true
	return routes, nil
}

// GetTripByID returns a trip by its ID, building the index on first access.
// This method is thread-safe.
func (c *ParsedFeedCache) GetTripByID(tripID string) (*schema.Trip, bool) {
	c.mu.RLock()
	if c.tripsByID != nil {
		trip, ok := c.tripsByID[tripID]
		c.mu.RUnlock()
		return trip, ok
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check
	if c.tripsByID != nil {
		trip, ok := c.tripsByID[tripID]
		return trip, ok
	}

	// Ensure trips are loaded
	if !c.loadedFiles["trips.txt"] {
		_, err := c.loadTripsInternal()
		if err != nil {
			return nil, false
		}
		c.loadedFiles["trips.txt"] = true
	}

	// Build index
	c.tripsByID = make(map[string]*schema.Trip, len(c.trips))
	for _, trip := range c.trips {
		c.tripsByID[trip.TripID] = trip
	}

	trip, ok := c.tripsByID[tripID]
	return trip, ok
}

// GetStopByID returns a stop by its ID, building the index on first access.
// This method is thread-safe.
func (c *ParsedFeedCache) GetStopByID(stopID string) (*schema.Stop, bool) {
	c.mu.RLock()
	if c.stopsByID != nil {
		stop, ok := c.stopsByID[stopID]
		c.mu.RUnlock()
		return stop, ok
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check
	if c.stopsByID != nil {
		stop, ok := c.stopsByID[stopID]
		return stop, ok
	}

	// Ensure stops are loaded
	if !c.loadedFiles["stops.txt"] {
		_, err := c.loadStopsInternal()
		if err != nil {
			return nil, false
		}
		c.loadedFiles["stops.txt"] = true
	}

	// Build index
	c.stopsByID = make(map[string]*schema.Stop, len(c.stops))
	for _, stop := range c.stops {
		c.stopsByID[stop.StopID] = stop
	}

	stop, ok := c.stopsByID[stopID]
	return stop, ok
}

// GetRouteByID returns a route by its ID, building the index on first access.
// This method is thread-safe.
func (c *ParsedFeedCache) GetRouteByID(routeID string) (*schema.Route, bool) {
	c.mu.RLock()
	if c.routesByID != nil {
		route, ok := c.routesByID[routeID]
		c.mu.RUnlock()
		return route, ok
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check
	if c.routesByID != nil {
		route, ok := c.routesByID[routeID]
		return route, ok
	}

	// Ensure routes are loaded
	if !c.loadedFiles["routes.txt"] {
		_, err := c.loadRoutesInternal()
		if err != nil {
			return nil, false
		}
		c.loadedFiles["routes.txt"] = true
	}

	// Build index
	c.routesByID = make(map[string]*schema.Route, len(c.routes))
	for _, route := range c.routes {
		c.routesByID[route.RouteID] = route
	}

	route, ok := c.routesByID[routeID]
	return route, ok
}

// GetStopTimesByTrip returns all stop times grouped by trip ID.
// This eliminates redundant grouping operations in validators.
// This method is thread-safe.
func (c *ParsedFeedCache) GetStopTimesByTrip() (map[string][]*schema.StopTime, error) {
	c.mu.RLock()
	if c.stopTimesByTrip != nil {
		result := c.stopTimesByTrip
		c.mu.RUnlock()
		return result, nil
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check
	if c.stopTimesByTrip != nil {
		return c.stopTimesByTrip, nil
	}

	// Ensure stop times are loaded
	if !c.loadedFiles["stop_times.txt"] {
		_, err := c.loadStopTimesInternal()
		if err != nil {
			return nil, err
		}
		c.loadedFiles["stop_times.txt"] = true
	}

	// Build index: group stop times by trip
	index := make(map[string][]*schema.StopTime)
	for _, st := range c.stopTimes {
		index[st.TripID] = append(index[st.TripID], st)
	}

	c.stopTimesByTrip = index
	return index, nil
}

// GetTripsByRoute returns all trips grouped by route ID.
// This eliminates O(routes × trips) nested loops in validators.
// This method is thread-safe.
func (c *ParsedFeedCache) GetTripsByRoute() (map[string][]*schema.Trip, error) {
	c.mu.RLock()
	if c.tripsByRoute != nil {
		result := c.tripsByRoute
		c.mu.RUnlock()
		return result, nil
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check
	if c.tripsByRoute != nil {
		return c.tripsByRoute, nil
	}

	// Ensure trips are loaded
	if !c.loadedFiles["trips.txt"] {
		_, err := c.loadTripsInternal()
		if err != nil {
			return nil, err
		}
		c.loadedFiles["trips.txt"] = true
	}

	// Build index: group trips by route
	index := make(map[string][]*schema.Trip)
	for _, trip := range c.trips {
		index[trip.RouteID] = append(index[trip.RouteID], trip)
	}

	c.tripsByRoute = index
	return index, nil
}

// Clear releases all cached data and resets the cache.
// This method is thread-safe.
func (c *ParsedFeedCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.stopTimes = nil
	c.trips = nil
	c.stops = nil
	c.routes = nil
	c.tripsByID = nil
	c.stopsByID = nil
	c.routesByID = nil
	c.tripsByRoute = nil
	c.stopTimesByTrip = nil
	c.loadedFiles = make(map[string]bool)
}

// GetLoader returns the underlying FeedLoader.
// This is useful for validators that need to access files not in the cache.
func (c *ParsedFeedCache) GetLoader() *FeedLoader {
	return c.loader
}

// Internal loading methods (must be called with write lock held)

func (c *ParsedFeedCache) loadStopTimesInternal() ([]*schema.StopTime, error) {
	reader, err := c.loader.GetFile("stop_times.txt")
	if err != nil {
		return nil, fmt.Errorf("failed to open stop_times.txt: %w", err)
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close stop_times.txt reader: %v", closeErr)
		}
	}()

	csvFile, err := NewCSVFile(reader, "stop_times.txt")
	if err != nil {
		return nil, fmt.Errorf("failed to parse stop_times.txt: %w", err)
	}

	var stopTimes []*schema.StopTime
	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		st := &schema.StopTime{}
		if err := parseStopTime(row, st); err == nil {
			stopTimes = append(stopTimes, st)
		}
	}

	return stopTimes, nil
}

func (c *ParsedFeedCache) loadTripsInternal() ([]*schema.Trip, error) {
	reader, err := c.loader.GetFile("trips.txt")
	if err != nil {
		return nil, fmt.Errorf("failed to open trips.txt: %w", err)
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close trips.txt reader: %v", closeErr)
		}
	}()

	csvFile, err := NewCSVFile(reader, "trips.txt")
	if err != nil {
		return nil, fmt.Errorf("failed to parse trips.txt: %w", err)
	}

	var trips []*schema.Trip
	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		trip := &schema.Trip{}
		if err := parseTrip(row, trip); err == nil {
			trips = append(trips, trip)
		}
	}

	return trips, nil
}

func (c *ParsedFeedCache) loadStopsInternal() ([]*schema.Stop, error) {
	reader, err := c.loader.GetFile("stops.txt")
	if err != nil {
		return nil, fmt.Errorf("failed to open stops.txt: %w", err)
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close stops.txt reader: %v", closeErr)
		}
	}()

	csvFile, err := NewCSVFile(reader, "stops.txt")
	if err != nil {
		return nil, fmt.Errorf("failed to parse stops.txt: %w", err)
	}

	var stops []*schema.Stop
	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		stop := &schema.Stop{}
		if err := parseStop(row, stop); err == nil {
			stops = append(stops, stop)
		}
	}

	return stops, nil
}

func (c *ParsedFeedCache) loadRoutesInternal() ([]*schema.Route, error) {
	reader, err := c.loader.GetFile("routes.txt")
	if err != nil {
		return nil, fmt.Errorf("failed to open routes.txt: %w", err)
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close routes.txt reader: %v", closeErr)
		}
	}()

	csvFile, err := NewCSVFile(reader, "routes.txt")
	if err != nil {
		return nil, fmt.Errorf("failed to parse routes.txt: %w", err)
	}

	var routes []*schema.Route
	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		route := &schema.Route{}
		if err := parseRoute(row, route); err == nil {
			routes = append(routes, route)
		}
	}

	return routes, nil
}

// Helper functions to parse CSV rows into schema structs

func parseStopTime(row *CSVRow, st *schema.StopTime) error {
	st.TripID = row.Values["trip_id"]
	st.ArrivalTime = row.Values["arrival_time"]
	st.DepartureTime = row.Values["departure_time"]
	st.StopID = row.Values["stop_id"]
	st.StopHeadsign = row.Values["stop_headsign"]
	st.PickupType = row.Values["pickup_type"]
	st.DropOffType = row.Values["drop_off_type"]
	st.ShapeDistTraveled = row.Values["shape_dist_traveled"]
	st.ContinuousPickup = row.Values["continuous_pickup"]
	st.ContinuousDropOff = row.Values["continuous_drop_off"]

	// Parse stop_sequence
	if seqStr, ok := row.Values["stop_sequence"]; ok && seqStr != "" {
		var seq int
		if _, err := fmt.Sscanf(seqStr, "%d", &seq); err == nil {
			st.StopSequence = seq
		}
	}

	// Parse timepoint
	if tpStr, ok := row.Values["timepoint"]; ok && tpStr != "" {
		var tp int
		if _, err := fmt.Sscanf(tpStr, "%d", &tp); err == nil {
			st.Timepoint = tp
		}
	}

	return nil
}

func parseTrip(row *CSVRow, trip *schema.Trip) error {
	trip.TripID = row.Values["trip_id"]
	trip.RouteID = row.Values["route_id"]
	trip.ServiceID = row.Values["service_id"]
	trip.TripHeadsign = row.Values["trip_headsign"]
	trip.TripShortName = row.Values["trip_short_name"]
	trip.DirectionID = row.Values["direction_id"]
	trip.BlockID = row.Values["block_id"]
	trip.ShapeID = row.Values["shape_id"]

	// Parse wheelchair_accessible
	if wcStr, ok := row.Values["wheelchair_accessible"]; ok && wcStr != "" {
		var wc int
		if _, err := fmt.Sscanf(wcStr, "%d", &wc); err == nil {
			trip.WheelchairAccessible = wc
		}
	}

	// Parse bikes_allowed
	if baStr, ok := row.Values["bikes_allowed"]; ok && baStr != "" {
		var ba int
		if _, err := fmt.Sscanf(baStr, "%d", &ba); err == nil {
			trip.BikesAllowed = ba
		}
	}

	return nil
}

func parseStop(row *CSVRow, stop *schema.Stop) error {
	stop.StopID = row.Values["stop_id"]
	stop.StopCode = row.Values["stop_code"]
	stop.StopName = row.Values["stop_name"]
	stop.StopDesc = row.Values["stop_desc"]
	stop.ParentStation = row.Values["parent_station"]
	stop.StopTimezone = row.Values["stop_timezone"]
	stop.LevelID = row.Values["level_id"]
	stop.StopURL = row.Values["stop_url"]
	stop.PlatformCode = row.Values["platform_code"]
	stop.ZoneID = row.Values["zone_id"]

	// Parse stop_lat
	if latStr, ok := row.Values["stop_lat"]; ok && latStr != "" {
		var lat float64
		if _, err := fmt.Sscanf(latStr, "%f", &lat); err == nil {
			stop.StopLat = lat
		}
	}

	// Parse stop_lon
	if lonStr, ok := row.Values["stop_lon"]; ok && lonStr != "" {
		var lon float64
		if _, err := fmt.Sscanf(lonStr, "%f", &lon); err == nil {
			stop.StopLon = lon
		}
	}

	// Parse location_type
	if ltStr, ok := row.Values["location_type"]; ok && ltStr != "" {
		var lt int
		if _, err := fmt.Sscanf(ltStr, "%d", &lt); err == nil {
			stop.LocationType = lt
		}
	}

	// Parse wheelchair_boarding
	if wbStr, ok := row.Values["wheelchair_boarding"]; ok && wbStr != "" {
		var wb int
		if _, err := fmt.Sscanf(wbStr, "%d", &wb); err == nil {
			stop.WheelchairBoarding = wb
		}
	}

	return nil
}

func parseRoute(row *CSVRow, route *schema.Route) error {
	route.RouteID = row.Values["route_id"]
	route.AgencyID = row.Values["agency_id"]
	route.RouteShortName = row.Values["route_short_name"]
	route.RouteLongName = row.Values["route_long_name"]
	route.RouteDesc = row.Values["route_desc"]
	route.RouteURL = row.Values["route_url"]
	route.RouteColor = row.Values["route_color"]
	route.RouteTextColor = row.Values["route_text_color"]
	route.RouteSortOrder = row.Values["route_sort_order"]
	route.ContinuousPickup = row.Values["continuous_pickup"]
	route.ContinuousDropOff = row.Values["continuous_drop_off"]

	// Parse route_type
	if rtStr, ok := row.Values["route_type"]; ok && rtStr != "" {
		var rt int
		if _, err := fmt.Sscanf(rtStr, "%d", &rt); err == nil {
			route.RouteType = rt
		}
	}

	return nil
}
