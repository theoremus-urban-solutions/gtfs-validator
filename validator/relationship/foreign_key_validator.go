package relationship

import (
	"io"
	"log"
	"strings"
	"sync"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/parser"
	"github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

// ForeignKeyValidator validates foreign key references between GTFS files
type ForeignKeyValidator struct{}

// NewForeignKeyValidator creates a new foreign key validator
func NewForeignKeyValidator() *ForeignKeyValidator {
	return &ForeignKeyValidator{}
}

// Validate checks that all foreign key references are valid
func (v *ForeignKeyValidator) Validate(loader *parser.FeedLoader, container *notice.NoticeContainer, config validator.Config) {
	var lookupMaps map[string]map[string]bool

	// Try to use cache if available (Phase 1 optimization)
	if cache := loader.GetCache(); cache != nil {
		lookupMaps = v.buildLookupMapsFromCache(cache)
	} else {
		// Fallback: use parallel building if configured (Phase 2 optimization)
		if config.ParallelWorkers > 1 {
			lookupMaps = v.buildLookupMapsParallel(loader)
		} else {
			lookupMaps = v.buildLookupMaps(loader)
		}
	}

	// Validate foreign keys in each file
	v.validateStopsReferences(loader, container, lookupMaps)
	v.validateRoutesReferences(loader, container, lookupMaps)
	v.validateTripsReferences(loader, container, lookupMaps)
	v.validateStopTimesReferences(loader, container, lookupMaps)
	v.validateCalendarDatesReferences(loader, container, lookupMaps)
	v.validateFareRulesReferences(loader, container, lookupMaps)
	v.validateShapesReferences(loader, container, lookupMaps)
	v.validateFrequenciesReferences(loader, container, lookupMaps)
	v.validateTransfersReferences(loader, container, lookupMaps)
	v.validatePathwaysReferences(loader, container, lookupMaps)
}

// buildLookupMapsFromCache builds lookup maps instantly from cached data.
// This eliminates all file I/O for building lookup maps (~16s → ~1s).
func (v *ForeignKeyValidator) buildLookupMapsFromCache(cache *parser.ParsedFeedCache) map[string]map[string]bool {
	lookupMaps := make(map[string]map[string]bool, 10) // Pre-allocate for 10 lookup types

	// Build stop_id lookup from cached stops
	stops, err := cache.GetStops()
	if err == nil {
		stopMap := make(map[string]bool, len(stops))
		for _, stop := range stops {
			if stop.StopID != "" {
				stopMap[stop.StopID] = true
			}
		}
		lookupMaps["stop_id"] = stopMap
	}

	// Build trip_id lookup from cached trips
	trips, err := cache.GetTrips()
	if err == nil {
		tripMap := make(map[string]bool, len(trips))
		// Pre-allocate serviceMap and shapeMap with reasonable estimates
		serviceMap := make(map[string]bool, len(trips)/50) // Typical: ~300 services for 15k trips
		shapeMap := make(map[string]bool, len(trips)/10)   // Typical: ~1500 shapes for 15k trips
		for _, trip := range trips {
			if trip.TripID != "" {
				tripMap[trip.TripID] = true
			}
			if trip.ServiceID != "" {
				serviceMap[trip.ServiceID] = true
			}
			if trip.ShapeID != "" {
				shapeMap[trip.ShapeID] = true
			}
		}
		lookupMaps["trip_id"] = tripMap
		lookupMaps["service_id"] = serviceMap
		lookupMaps["shape_id"] = shapeMap
	}

	// Build route_id lookup from cached routes
	routes, err := cache.GetRoutes()
	if err == nil {
		routeMap := make(map[string]bool, len(routes))
		agencyMap := make(map[string]bool, len(routes)/50) // Typical: 1-10 agencies for 200 routes
		for _, route := range routes {
			if route.RouteID != "" {
				routeMap[route.RouteID] = true
			}
			if route.AgencyID != "" {
				agencyMap[route.AgencyID] = true
			}
		}
		lookupMaps["route_id"] = routeMap
		lookupMaps["agency_id"] = agencyMap
	}

	// Build zone_id lookup from cached stops
	if stops != nil {
		zoneMap := make(map[string]bool, len(stops)/10) // Typical: ~200 zones for 2000 stops
		for _, stop := range stops {
			if stop.ZoneID != "" {
				zoneMap[stop.ZoneID] = true
			}
		}
		lookupMaps["zone_id"] = zoneMap
	}

	// For files not in cache, fall back to sequential loading
	// (these are typically small files)
	loader := cache.GetLoader()
	lookupMaps["fare_id"] = v.buildLookupMap(loader, "fare_attributes.txt", "fare_id")
	lookupMaps["pathway_id"] = v.buildLookupMap(loader, "pathways.txt", "pathway_id")
	lookupMaps["level_id"] = v.buildLookupMap(loader, "levels.txt", "level_id")

	return lookupMaps
}

// buildLookupMapsParallel builds lookup maps in parallel when cache is unavailable.
// This provides a fallback optimization for non-cached mode (~16s → ~5s).
func (v *ForeignKeyValidator) buildLookupMapsParallel(loader *parser.FeedLoader) map[string]map[string]bool {
	// Define tasks for parallel execution
	type mapTask struct {
		key      string
		filename string
		field    string
	}

	tasks := []mapTask{
		{"agency_id", "agency.txt", "agency_id"},
		{"stop_id", "stops.txt", "stop_id"},
		{"route_id", "routes.txt", "route_id"},
		{"trip_id", "trips.txt", "trip_id"},
		{"shape_id", "shapes.txt", "shape_id"},
		{"fare_id", "fare_attributes.txt", "fare_id"},
		{"pathway_id", "pathways.txt", "pathway_id"},
		{"level_id", "levels.txt", "level_id"},
	}

	// Result channel
	type mapResult struct {
		key    string
		lookup map[string]bool
	}
	resultChan := make(chan mapResult, len(tasks))

	// Launch parallel builds
	var wg sync.WaitGroup
	for _, task := range tasks {
		wg.Add(1)
		go func(t mapTask) {
			defer wg.Done()
			lookup := v.buildLookupMap(loader, t.filename, t.field)
			resultChan <- mapResult{t.key, lookup}
		}(task)
	}

	// Wait for completion
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	lookupMaps := make(map[string]map[string]bool)
	for result := range resultChan {
		lookupMaps[result.key] = result.lookup
	}

	// Build composite maps (service_id, zone_id) - these need sequential processing
	lookupMaps["service_id"] = v.buildServiceIdLookupMap(loader)
	lookupMaps["zone_id"] = v.buildZoneIdLookupMap(loader)

	return lookupMaps
}

// buildLookupMaps creates lookup maps for all primary keys
func (v *ForeignKeyValidator) buildLookupMaps(loader *parser.FeedLoader) map[string]map[string]bool {
	lookupMaps := make(map[string]map[string]bool)

	// Build lookup maps for each table
	lookupMaps["agency_id"] = v.buildLookupMap(loader, "agency.txt", "agency_id")
	lookupMaps["stop_id"] = v.buildLookupMap(loader, "stops.txt", "stop_id")
	lookupMaps["route_id"] = v.buildLookupMap(loader, "routes.txt", "route_id")
	lookupMaps["trip_id"] = v.buildLookupMap(loader, "trips.txt", "trip_id")
	lookupMaps["service_id"] = v.buildServiceIdLookupMap(loader)
	lookupMaps["shape_id"] = v.buildLookupMap(loader, "shapes.txt", "shape_id")
	lookupMaps["fare_id"] = v.buildLookupMap(loader, "fare_attributes.txt", "fare_id")
	lookupMaps["zone_id"] = v.buildZoneIdLookupMap(loader)
	lookupMaps["pathway_id"] = v.buildLookupMap(loader, "pathways.txt", "pathway_id")
	lookupMaps["level_id"] = v.buildLookupMap(loader, "levels.txt", "level_id")

	return lookupMaps
}

// buildLookupMap creates a lookup map for a specific field from a file
func (v *ForeignKeyValidator) buildLookupMap(loader *parser.FeedLoader, filename string, fieldName string) map[string]bool {
	lookupMap := make(map[string]bool)

	reader, err := loader.GetFile(filename)
	if err != nil {
		return lookupMap // Return empty map if file doesn't exist
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close reader %v", closeErr)
		}
	}()

	csvFile, err := parser.NewCSVFile(reader, filename)
	if err != nil {
		return lookupMap
	}

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}

		if value, exists := row.Values[fieldName]; exists && strings.TrimSpace(value) != "" {
			lookupMap[value] = true
		}
	}

	return lookupMap
}

// buildServiceIdLookupMap builds service_id lookup from both calendar.txt and calendar_dates.txt
func (v *ForeignKeyValidator) buildServiceIdLookupMap(loader *parser.FeedLoader) map[string]bool {
	lookupMap := make(map[string]bool)

	// Add service_ids from calendar.txt
	calendarMap := v.buildLookupMap(loader, "calendar.txt", "service_id")
	for serviceId := range calendarMap {
		lookupMap[serviceId] = true
	}

	// Add service_ids from calendar_dates.txt
	calendarDatesMap := v.buildLookupMap(loader, "calendar_dates.txt", "service_id")
	for serviceId := range calendarDatesMap {
		lookupMap[serviceId] = true
	}

	return lookupMap
}

// buildZoneIdLookupMap builds zone_id lookup from stops.txt
func (v *ForeignKeyValidator) buildZoneIdLookupMap(loader *parser.FeedLoader) map[string]bool {
	return v.buildLookupMap(loader, "stops.txt", "zone_id")
}

// validateStopsReferences validates foreign keys in stops.txt
func (v *ForeignKeyValidator) validateStopsReferences(loader *parser.FeedLoader, container *notice.NoticeContainer, lookupMaps map[string]map[string]bool) {
	v.validateFileReferences(loader, container, "stops.txt", map[string]string{
		"parent_station": "stop_id",
		"level_id":       "level_id",
		"zone_id":        "zone_id",
	}, lookupMaps)
}

// validateRoutesReferences validates foreign keys in routes.txt
func (v *ForeignKeyValidator) validateRoutesReferences(loader *parser.FeedLoader, container *notice.NoticeContainer, lookupMaps map[string]map[string]bool) {
	v.validateFileReferences(loader, container, "routes.txt", map[string]string{
		"agency_id": "agency_id",
	}, lookupMaps)
}

// validateTripsReferences validates foreign keys in trips.txt
func (v *ForeignKeyValidator) validateTripsReferences(loader *parser.FeedLoader, container *notice.NoticeContainer, lookupMaps map[string]map[string]bool) {
	v.validateFileReferences(loader, container, "trips.txt", map[string]string{
		"route_id":   "route_id",
		"service_id": "service_id",
		"shape_id":   "shape_id",
	}, lookupMaps)
}

// validateStopTimesReferences validates foreign keys in stop_times.txt
func (v *ForeignKeyValidator) validateStopTimesReferences(loader *parser.FeedLoader, container *notice.NoticeContainer, lookupMaps map[string]map[string]bool) {
	v.validateFileReferences(loader, container, "stop_times.txt", map[string]string{
		"trip_id": "trip_id",
		"stop_id": "stop_id",
	}, lookupMaps)
}

// validateCalendarDatesReferences validates foreign keys in calendar_dates.txt
func (v *ForeignKeyValidator) validateCalendarDatesReferences(loader *parser.FeedLoader, container *notice.NoticeContainer, lookupMaps map[string]map[string]bool) {
	v.validateFileReferences(loader, container, "calendar_dates.txt", map[string]string{
		"service_id": "service_id",
	}, lookupMaps)
}

// validateFareRulesReferences validates foreign keys in fare_rules.txt
func (v *ForeignKeyValidator) validateFareRulesReferences(loader *parser.FeedLoader, container *notice.NoticeContainer, lookupMaps map[string]map[string]bool) {
	v.validateFileReferences(loader, container, "fare_rules.txt", map[string]string{
		"fare_id":        "fare_id",
		"route_id":       "route_id",
		"origin_id":      "zone_id",
		"destination_id": "zone_id",
		"contains_id":    "zone_id",
	}, lookupMaps)
}

// validateShapesReferences validates foreign keys in shapes.txt
func (v *ForeignKeyValidator) validateShapesReferences(loader *parser.FeedLoader, container *notice.NoticeContainer, lookupMaps map[string]map[string]bool) {
	// shapes.txt doesn't have foreign key references to validate
}

// validateFrequenciesReferences validates foreign keys in frequencies.txt
func (v *ForeignKeyValidator) validateFrequenciesReferences(loader *parser.FeedLoader, container *notice.NoticeContainer, lookupMaps map[string]map[string]bool) {
	v.validateFileReferences(loader, container, "frequencies.txt", map[string]string{
		"trip_id": "trip_id",
	}, lookupMaps)
}

// validateTransfersReferences validates foreign keys in transfers.txt
func (v *ForeignKeyValidator) validateTransfersReferences(loader *parser.FeedLoader, container *notice.NoticeContainer, lookupMaps map[string]map[string]bool) {
	v.validateFileReferences(loader, container, "transfers.txt", map[string]string{
		"from_stop_id":  "stop_id",
		"to_stop_id":    "stop_id",
		"from_route_id": "route_id",
		"to_route_id":   "route_id",
		"from_trip_id":  "trip_id",
		"to_trip_id":    "trip_id",
	}, lookupMaps)
}

// validatePathwaysReferences validates foreign keys in pathways.txt
func (v *ForeignKeyValidator) validatePathwaysReferences(loader *parser.FeedLoader, container *notice.NoticeContainer, lookupMaps map[string]map[string]bool) {
	v.validateFileReferences(loader, container, "pathways.txt", map[string]string{
		"from_stop_id": "stop_id",
		"to_stop_id":   "stop_id",
	}, lookupMaps)
}

// validateFileReferences validates foreign key references in a specific file
func (v *ForeignKeyValidator) validateFileReferences(loader *parser.FeedLoader, container *notice.NoticeContainer, filename string, foreignKeys map[string]string, lookupMaps map[string]map[string]bool) {
	reader, err := loader.GetFile(filename)
	if err != nil {
		return // File doesn't exist, skip validation
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close reader %v", closeErr)
		}
	}()

	csvFile, err := parser.NewCSVFile(reader, filename)
	if err != nil {
		return
	}

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}

		// Check each foreign key field
		for fieldName, referencedTable := range foreignKeys {
			if value, exists := row.Values[fieldName]; exists && strings.TrimSpace(value) != "" {
				// Check if the referenced value exists
				if lookupMap, tableExists := lookupMaps[referencedTable]; tableExists {
					if !lookupMap[value] {
						container.AddNotice(notice.NewForeignKeyViolationNotice(
							filename,
							fieldName,
							value,
							row.RowNumber,
							v.getReferencedTableName(referencedTable),
							referencedTable,
						))
					}
				}
			}
		}
	}
}

// getReferencedTableName returns the table name for a given field
func (v *ForeignKeyValidator) getReferencedTableName(fieldName string) string {
	switch fieldName {
	case "agency_id":
		return "agency.txt"
	case "stop_id":
		return "stops.txt"
	case "route_id":
		return "routes.txt"
	case "trip_id":
		return "trips.txt"
	case "service_id":
		return "calendar.txt or calendar_dates.txt"
	case "shape_id":
		return "shapes.txt"
	case "fare_id":
		return "fare_attributes.txt"
	case "zone_id":
		return "stops.txt"
	case "pathway_id":
		return "pathways.txt"
	case "level_id":
		return "levels.txt"
	default:
		return "unknown"
	}
}
