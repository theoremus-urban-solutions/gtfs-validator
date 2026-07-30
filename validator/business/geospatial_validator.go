package business

import (
	"io"
	"log"
	"math"
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
	StopID        string
	StopName      string
	Latitude      float64
	Longitude     float64
	LocationType  int
	ParentStation string
	ZoneID        string
	RowNumber     int
}

// Validate performs comprehensive geospatial validation
func (v *GeospatialValidator) Validate(loader *parser.FeedLoader, container *notice.NoticeContainer, config validator.Config) {
	stops := v.loadGeoStops(loader)
	if len(stops) == 0 {
		return
	}

	v.validateStopSpatialRelationships(container, stops)
}

// loadGeoStops loads stops with geographic information
func (v *GeospatialValidator) loadGeoStops(loader *parser.FeedLoader) map[string]*GeoStop {
	stops := make(map[string]*GeoStop)

	reader, err := loader.GetFile("stops.txt")
	if err != nil {
		return stops
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close reader %v", closeErr)
		}
	}()

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

// A stop_lat or stop_lon outside the globe is reported by
// core/field_type_validator.go as number_out_of_range, which bounds both fields
// from the same table it uses for every other numeric field. Coordinates at
// (0,0) are reported by core/coordinate_validator.go.

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

	// Shape geometry is checked by business/shape_geometry_validator.go, which
	// matches stops against the alignment they are served by rather than
	// against a bounding box derived from those same stops.
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
