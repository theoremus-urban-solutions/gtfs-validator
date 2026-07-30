package core

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

// CoordinateValidator reports points that parse and are in range but cannot be
// real: the two positions a missing coordinate tends to become.
//
// Whether the values are numbers at all, and whether they are within the
// range of the globe, is checked by field_type_validator.go.
type CoordinateValidator struct{}

// NewCoordinateValidator creates a new coordinate validator
func NewCoordinateValidator() *CoordinateValidator {
	return &CoordinateValidator{}
}

// coordinatePairs names the latitude and longitude fields of each file that
// carries a point.
var coordinatePairs = map[string][2]string{
	StopsFile:    {"stop_lat", "stop_lon"},
	"shapes.txt": {"shape_pt_lat", "shape_pt_lon"},
}

// nearOriginDegrees is how close to (0, 0) counts as the origin. A tenth of a
// degree is roughly 11 km, comfortably inside the Gulf of Guinea and nowhere
// near any land a transit feed would describe.
const nearOriginDegrees = 0.1

// nearPoleDegrees is how close to ±90 counts as the pole.
const nearPoleDegrees = 0.1

// Validate checks the points in every file that carries them.
func (v *CoordinateValidator) Validate(loader *parser.FeedLoader, container *notice.NoticeContainer, config validator.Config) {
	for filename, fields := range coordinatePairs {
		v.validateFileCoordinates(loader, container, filename, fields)
	}
}

// validateFileCoordinates walks one file's points.
func (v *CoordinateValidator) validateFileCoordinates(loader *parser.FeedLoader, container *notice.NoticeContainer, filename string, fields [2]string) {
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

	latField, lonField := fields[0], fields[1]

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}

		lat, latOK := inRangeFloat(row.Values[latField], 90)
		lon, lonOK := inRangeFloat(row.Values[lonField], 180)
		if !latOK || !lonOK {
			continue // Reported as invalid_float or number_out_of_range.
		}

		// (0, 0) is in the Gulf of Guinea. A point there is almost always two
		// fields left empty and defaulted to zero.
		if math.Abs(lat) < nearOriginDegrees && math.Abs(lon) < nearOriginDegrees {
			container.AddNotice(notice.NewPointNearOriginNotice(
				filename, latField, row.Values[latField], row.RowNumber,
			))
			continue
		}

		if math.Abs(lat) > 90-nearPoleDegrees {
			container.AddNotice(notice.NewPointNearPoleNotice(
				filename, latField, row.Values[latField], row.RowNumber,
			))
		}
	}
}

// inRangeFloat parses a coordinate and reports whether it is usable: present,
// numeric and within ±limit.
func inRangeFloat(raw string, limit float64) (float64, bool) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return 0, false
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil || math.Abs(parsed) > limit {
		return 0, false
	}
	return parsed, true
}
