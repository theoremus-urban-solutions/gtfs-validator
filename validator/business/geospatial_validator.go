package business

import (
	"io"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/parser"
	"github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

// GeospatialValidator validates geographic consistency and spatial relationships
type GeospatialValidator struct{}

// NewGeospatialValidator creates a new geospatial validator
func NewGeospatialValidator() *GeospatialValidator {
	return &GeospatialValidator{}
}

// GeoStop represents a stop with geographic information
type GeoStop struct {
	StopID       string
	StopName     string
	Latitude     float64
	Longitude    float64
	LocationType int
	ParentStation string
	ZoneID       string
	RowNumber    int
}

// GeoShape represents a shape point with geographic information
type GeoShape struct {
	ShapeID      string
	Latitude     float64
	Longitude    float64
	Sequence     int
	DistTraveled *float64
	RowNumber    int
}

// BoundingBox represents geographic bounds
type BoundingBox struct {
	MinLat float64
	MaxLat float64
	MinLon float64
	MaxLon float64
}

// StopCluster represents spatially clustered stops
type StopCluster struct {
	CenterLat   float64
	CenterLon   float64
	Stops       []*GeoStop
	Radius      float64
	StopCount   int
}

// Validate performs comprehensive geospatial validation
func (v *GeospatialValidator) Validate(loader *parser.FeedLoader, container *notice.NoticeContainer, config validator.Config) {
	// Load geographic data
	stops := v.loadGeoStops(loader)
	shapes := v.loadGeoShapes(loader)

	if len(stops) == 0 {
		return
	}

	// Calculate feed bounding box
	feedBounds := v.calculateFeedBounds(stops, shapes)
	
	// Validate geographic consistency
	v.validateGeographicConsistency(container, stops, feedBounds)
	
	// Validate stop spatial relationships
	v.validateStopSpatialRelationships(container, stops)
	
	// Validate shape geometry if available
	if len(shapes) > 0 {
		v.validateShapeGeometry(container, shapes, feedBounds)
	}
	
	// Analyze stop clustering patterns
	v.analyzeStopClustering(container, stops)
	
	// Validate coordinate precision and accuracy
	v.validateCoordinateQuality(container, stops, shapes)
}

// loadGeoStops loads stops with geographic information
func (v *GeospatialValidator) loadGeoStops(loader *parser.FeedLoader) map[string]*GeoStop {
	stops := make(map[string]*GeoStop)

	reader, err := loader.GetFile("stops.txt")
	if err != nil {
		return stops
	}
	defer reader.Close()

	csvFile, err := parser.NewCSVFile(reader, "stops.txt")
	if err != nil {
		return stops
	}

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		stop := v.parseGeoStop(row)
		if stop != nil {
			stops[stop.StopID] = stop
		}
	}

	return stops
}

// parseGeoStop parses a geographic stop record
func (v *GeospatialValidator) parseGeoStop(row *parser.CSVRow) *GeoStop {
	stopID, hasStopID := row.Values["stop_id"]
	latStr, hasLat := row.Values["stop_lat"]
	lonStr, hasLon := row.Values["stop_lon"]

	if !hasStopID || !hasLat || !hasLon {
		return nil
	}

	lat, err1 := strconv.ParseFloat(strings.TrimSpace(latStr), 64)
	lon, err2 := strconv.ParseFloat(strings.TrimSpace(lonStr), 64)

	if err1 != nil || err2 != nil {
		return nil
	}

	stop := &GeoStop{
		StopID:    strings.TrimSpace(stopID),
		Latitude:  lat,
		Longitude: lon,
		RowNumber: row.RowNumber,
	}

	if stopName, hasName := row.Values["stop_name"]; hasName {
		stop.StopName = strings.TrimSpace(stopName)
	}
	if parentStation, hasParent := row.Values["parent_station"]; hasParent {
		stop.ParentStation = strings.TrimSpace(parentStation)
	}
	if zoneID, hasZone := row.Values["zone_id"]; hasZone {
		stop.ZoneID = strings.TrimSpace(zoneID)
	}
	if locTypeStr, hasLocType := row.Values["location_type"]; hasLocType && strings.TrimSpace(locTypeStr) != "" {
		if locType, err := strconv.Atoi(strings.TrimSpace(locTypeStr)); err == nil {
			stop.LocationType = locType
		}
	}

	return stop
}

// loadGeoShapes loads shape points with geographic information
func (v *GeospatialValidator) loadGeoShapes(loader *parser.FeedLoader) map[string][]*GeoShape {
	shapes := make(map[string][]*GeoShape)

	reader, err := loader.GetFile("shapes.txt")
	if err != nil {
		return shapes
	}
	defer reader.Close()

	csvFile, err := parser.NewCSVFile(reader, "shapes.txt")
	if err != nil {
		return shapes
	}

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		shapePoint := v.parseGeoShape(row)
		if shapePoint != nil {
			shapes[shapePoint.ShapeID] = append(shapes[shapePoint.ShapeID], shapePoint)
		}
	}

	// Sort each shape by sequence
	for shapeID := range shapes {
		sort.Slice(shapes[shapeID], func(i, j int) bool {
			return shapes[shapeID][i].Sequence < shapes[shapeID][j].Sequence
		})
	}

	return shapes
}

// parseGeoShape parses a geographic shape point record
func (v *GeospatialValidator) parseGeoShape(row *parser.CSVRow) *GeoShape {
	shapeID, hasShapeID := row.Values["shape_id"]
	latStr, hasLat := row.Values["shape_pt_lat"]
	lonStr, hasLon := row.Values["shape_pt_lon"]
	seqStr, hasSeq := row.Values["shape_pt_sequence"]

	if !hasShapeID || !hasLat || !hasLon || !hasSeq {
		return nil
	}

	lat, err1 := strconv.ParseFloat(strings.TrimSpace(latStr), 64)
	lon, err2 := strconv.ParseFloat(strings.TrimSpace(lonStr), 64)
	seq, err3 := strconv.Atoi(strings.TrimSpace(seqStr))

	if err1 != nil || err2 != nil || err3 != nil {
		return nil
	}

	shape := &GeoShape{
		ShapeID:   strings.TrimSpace(shapeID),
		Latitude:  lat,
		Longitude: lon,
		Sequence:  seq,
		RowNumber: row.RowNumber,
	}

	if distStr, hasDist := row.Values["shape_dist_traveled"]; hasDist && strings.TrimSpace(distStr) != "" {
		if dist, err := strconv.ParseFloat(strings.TrimSpace(distStr), 64); err == nil {
			shape.DistTraveled = &dist
		}
	}

	return shape
}

// calculateFeedBounds calculates the geographic bounding box of the feed
func (v *GeospatialValidator) calculateFeedBounds(stops map[string]*GeoStop, shapes map[string][]*GeoShape) BoundingBox {
	bounds := BoundingBox{
		MinLat: 90.0,
		MaxLat: -90.0,
		MinLon: 180.0,
		MaxLon: -180.0,
	}

	// Include stops
	for _, stop := range stops {
		if stop.Latitude < bounds.MinLat {
			bounds.MinLat = stop.Latitude
		}
		if stop.Latitude > bounds.MaxLat {
			bounds.MaxLat = stop.Latitude
		}
		if stop.Longitude < bounds.MinLon {
			bounds.MinLon = stop.Longitude
		}
		if stop.Longitude > bounds.MaxLon {
			bounds.MaxLon = stop.Longitude
		}
	}

	// Include shapes
	for _, shapePoints := range shapes {
		for _, point := range shapePoints {
			if point.Latitude < bounds.MinLat {
				bounds.MinLat = point.Latitude
			}
			if point.Latitude > bounds.MaxLat {
				bounds.MaxLat = point.Latitude
			}
			if point.Longitude < bounds.MinLon {
				bounds.MinLon = point.Longitude
			}
			if point.Longitude > bounds.MaxLon {
				bounds.MaxLon = point.Longitude
			}
		}
	}

	return bounds
}

// validateGeographicConsistency validates overall geographic consistency
func (v *GeospatialValidator) validateGeographicConsistency(container *notice.NoticeContainer, stops map[string]*GeoStop, bounds BoundingBox) {
	// Check for coordinates outside reasonable bounds
	for _, stop := range stops {
		// Check for invalid coordinates
		if stop.Latitude < -90 || stop.Latitude > 90 {
			container.AddNotice(notice.NewInvalidLatitudeNotice(
				stop.StopID,
				stop.Latitude,
				stop.RowNumber,
			))
		}
		if stop.Longitude < -180 || stop.Longitude > 180 {
			container.AddNotice(notice.NewInvalidLongitudeNotice(
				stop.StopID,
				stop.Longitude,
				stop.RowNumber,
			))
		}

		// Check for coordinates at (0,0) which are often errors
		if v.approximatelyEqual(stop.Latitude, 0.0, 0.000001) && 
		   v.approximatelyEqual(stop.Longitude, 0.0, 0.000001) {
			container.AddNotice(notice.NewSuspiciousCoordinateNotice(
				"stops.txt",
				"stop_lat",
				strconv.FormatFloat(stop.Latitude, 'f', -1, 64),
				stop.RowNumber,
				"coordinates_at_origin",
			))
		}
	}

	// Calculate feed coverage area
	latSpan := bounds.MaxLat - bounds.MinLat
	lonSpan := bounds.MaxLon - bounds.MinLon
	
	// Very large coverage (> 1000km in any direction) might indicate data errors
	if latSpan > 9.0 || lonSpan > 9.0 { // Roughly 1000km
		container.AddNotice(notice.NewVeryLargeFeedCoverageNotice(
			bounds.MinLat, bounds.MaxLat, bounds.MinLon, bounds.MaxLon,
			latSpan, lonSpan,
		))
	}

	// Very small coverage (< 1km in any direction) might indicate precision issues
	if latSpan < 0.009 && lonSpan < 0.009 { // Roughly 1km
		container.AddNotice(notice.NewVerySmallFeedCoverageNotice(
			bounds.MinLat, bounds.MaxLat, bounds.MinLon, bounds.MaxLon,
			latSpan, lonSpan,
		))
	}
}

// validateStopSpatialRelationships validates spatial relationships between stops
func (v *GeospatialValidator) validateStopSpatialRelationships(container *notice.NoticeContainer, stops map[string]*GeoStop) {
	// Check parent-child station distances
	for _, stop := range stops {
		if stop.ParentStation != "" {
			if parent, exists := stops[stop.ParentStation]; exists {
				distance := v.haversineDistance(
					stop.Latitude, stop.Longitude,
					parent.Latitude, parent.Longitude,
				)

				// Child station too far from parent (> 500m)
				if distance > 500 {
					container.AddNotice(notice.NewChildStationTooFarFromParentNotice(
						stop.StopID,
						stop.ParentStation,
						distance,
						stop.RowNumber,
					))
				}
			}
		}
	}

	// Find stops that are very close to each other (potential duplicates)
	stopList := make([]*GeoStop, 0, len(stops))
	for _, stop := range stops {
		stopList = append(stopList, stop)
	}

	for i := 0; i < len(stopList); i++ {
		for j := i + 1; j < len(stopList); j++ {
			stop1 := stopList[i]
			stop2 := stopList[j]

			distance := v.haversineDistance(
				stop1.Latitude, stop1.Longitude,
				stop2.Latitude, stop2.Longitude,
			)

			// Stops very close together (< 10m) might be duplicates
			if distance < 10 {
				container.AddNotice(notice.NewVeryCloseStopsNotice(
					stop1.StopID,
					stop2.StopID,
					distance,
					stop1.RowNumber,
					stop2.RowNumber,
				))
			}
		}
	}
}

// validateShapeGeometry validates shape geometric properties
func (v *GeospatialValidator) validateShapeGeometry(container *notice.NoticeContainer, shapes map[string][]*GeoShape, bounds BoundingBox) {
	for shapeID, shapePoints := range shapes {
		if len(shapePoints) < 2 {
			continue
		}

		// Check for shape points outside feed bounds (with some tolerance)
		tolerance := 0.01 // ~1km tolerance
		for _, point := range shapePoints {
			if point.Latitude < bounds.MinLat-tolerance || point.Latitude > bounds.MaxLat+tolerance ||
			   point.Longitude < bounds.MinLon-tolerance || point.Longitude > bounds.MaxLon+tolerance {
				container.AddNotice(notice.NewShapePointOutsideFeedBoundsNotice(
					shapeID,
					point.Sequence,
					point.Latitude,
					point.Longitude,
					point.RowNumber,
				))
			}
		}

		// Check for unreasonably long segments
		for i := 1; i < len(shapePoints); i++ {
			curr := shapePoints[i]
			prev := shapePoints[i-1]

			distance := v.haversineDistance(
				prev.Latitude, prev.Longitude,
				curr.Latitude, curr.Longitude,
			)

			// Very long segments (> 50km) might indicate missing points
			if distance > 50000 {
				container.AddNotice(notice.NewUnreasonablyLongShapeSegmentNotice(
					shapeID,
					prev.Sequence,
					curr.Sequence,
					distance,
					curr.RowNumber,
				))
			}
		}

		// Validate shape distance consistency if provided
		v.validateShapeDistanceConsistency(container, shapeID, shapePoints)
	}
}

// validateShapeDistanceConsistency validates shape distance values against geographic distances
func (v *GeospatialValidator) validateShapeDistanceConsistency(container *notice.NoticeContainer, shapeID string, points []*GeoShape) {
	if len(points) < 2 {
		return
	}

	// Check if any points have distance values
	hasAnyDistance := false
	for _, point := range points {
		if point.DistTraveled != nil {
			hasAnyDistance = true
			break
		}
	}

	if !hasAnyDistance {
		return
	}

	// Calculate cumulative geographic distance
	cumulativeGeoDistance := 0.0
	for i := 1; i < len(points); i++ {
		curr := points[i]
		prev := points[i-1]

		segmentDistance := v.haversineDistance(
			prev.Latitude, prev.Longitude,
			curr.Latitude, curr.Longitude,
		)
		cumulativeGeoDistance += segmentDistance

		// Compare with provided distance
		if curr.DistTraveled != nil {
			providedDistance := *curr.DistTraveled
			
			// Allow 20% tolerance for distance differences
			tolerance := cumulativeGeoDistance * 0.2
			if math.Abs(providedDistance-cumulativeGeoDistance) > tolerance && tolerance > 100 {
				container.AddNotice(notice.NewShapeDistanceInconsistentWithGeographyNotice(
					shapeID,
					curr.Sequence,
					providedDistance,
					cumulativeGeoDistance,
					math.Abs(providedDistance-cumulativeGeoDistance),
					curr.RowNumber,
				))
			}
		}
	}
}

// analyzeStopClustering analyzes stop clustering patterns
func (v *GeospatialValidator) analyzeStopClustering(container *notice.NoticeContainer, stops map[string]*GeoStop) {
	// Simple clustering analysis - find areas with high stop density
	clusters := v.findStopClusters(stops, 500) // 500m radius

	// Report clusters with many stops
	for _, cluster := range clusters {
		if cluster.StopCount > 20 {
			container.AddNotice(notice.NewHighStopDensityAreaNotice(
				cluster.CenterLat,
				cluster.CenterLon,
				cluster.Radius,
				cluster.StopCount,
			))
		}
	}

	// Report if there are very few clusters (might indicate spread-out network)
	if len(clusters) < 3 && len(stops) > 100 {
		container.AddNotice(notice.NewLowStopClusteringNotice(
			len(clusters),
			len(stops),
		))
	}
}

// findStopClusters finds clusters of stops within a given radius
func (v *GeospatialValidator) findStopClusters(stops map[string]*GeoStop, radiusMeters float64) []*StopCluster {
	var clusters []*StopCluster
	processed := make(map[string]bool)

	for _, stop := range stops {
		if processed[stop.StopID] {
			continue
		}

		// Find all stops within radius
		clusterStops := []*GeoStop{stop}
		processed[stop.StopID] = true

		for _, otherStop := range stops {
			if processed[otherStop.StopID] {
				continue
			}

			distance := v.haversineDistance(
				stop.Latitude, stop.Longitude,
				otherStop.Latitude, otherStop.Longitude,
			)

			if distance <= radiusMeters {
				clusterStops = append(clusterStops, otherStop)
				processed[otherStop.StopID] = true
			}
		}

		if len(clusterStops) > 1 {
			// Calculate cluster center
			totalLat := 0.0
			totalLon := 0.0
			for _, clusterStop := range clusterStops {
				totalLat += clusterStop.Latitude
				totalLon += clusterStop.Longitude
			}

			clusters = append(clusters, &StopCluster{
				CenterLat: totalLat / float64(len(clusterStops)),
				CenterLon: totalLon / float64(len(clusterStops)),
				Stops:     clusterStops,
				Radius:    radiusMeters,
				StopCount: len(clusterStops),
			})
		}
	}

	return clusters
}

// validateCoordinateQuality validates coordinate precision and accuracy
func (v *GeospatialValidator) validateCoordinateQuality(container *notice.NoticeContainer, stops map[string]*GeoStop, shapes map[string][]*GeoShape) {
	// Check coordinate precision for stops
	for _, stop := range stops {
		precision := v.getCoordinatePrecision(stop.Latitude, stop.Longitude)
		
		// Less than 4 decimal places (~11m precision) might be insufficient
		if precision < 4 {
			container.AddNotice(notice.NewInsufficientCoordinatePrecisionNotice(
				"stops.txt",
				"stop_lat",
				strconv.FormatFloat(stop.Latitude, 'f', -1, 64),
				stop.RowNumber,
				precision,
			))
		}
	}

	// Generate geospatial summary
	container.AddNotice(notice.NewGeospatialSummaryNotice(
		len(stops),
		len(shapes),
		len(v.findStopClusters(stops, 500)),
	))
}

// getCoordinatePrecision estimates decimal precision of a coordinate
func (v *GeospatialValidator) getCoordinatePrecision(lat, lon float64) int {
	latStr := strconv.FormatFloat(lat, 'f', -1, 64)
	lonStr := strconv.FormatFloat(lon, 'f', -1, 64)

	latPrecision := 0
	if dotIndex := strings.Index(latStr, "."); dotIndex != -1 {
		latPrecision = len(latStr) - dotIndex - 1
	}

	lonPrecision := 0
	if dotIndex := strings.Index(lonStr, "."); dotIndex != -1 {
		lonPrecision = len(lonStr) - dotIndex - 1
	}

	// Return minimum precision
	if latPrecision < lonPrecision {
		return latPrecision
	}
	return lonPrecision
}

// haversineDistance calculates distance between two lat/lon points in meters
func (v *GeospatialValidator) haversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadius = 6371000 // Earth radius in meters

	lat1Rad := lat1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	deltaLatRad := (lat2 - lat1) * math.Pi / 180
	deltaLonRad := (lon2 - lon1) * math.Pi / 180

	a := math.Sin(deltaLatRad/2)*math.Sin(deltaLatRad/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLonRad/2)*math.Sin(deltaLonRad/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c
}

// approximatelyEqual checks if two float64 values are approximately equal
func (v *GeospatialValidator) approximatelyEqual(a, b, epsilon float64) bool {
	return math.Abs(a-b) < epsilon
}