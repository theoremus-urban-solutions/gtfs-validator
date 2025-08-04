package core

import (
	"io"
	"strconv"
	"strings"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/parser"
	"github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

// CoordinateValidator validates latitude and longitude values
type CoordinateValidator struct{}

// NewCoordinateValidator creates a new coordinate validator
func NewCoordinateValidator() *CoordinateValidator {
	return &CoordinateValidator{}
}

// coordinateFields defines which fields contain coordinate values in each file
var coordinateFields = map[string][]string{
	"stops.txt":  {"stop_lat", "stop_lon"},
	"shapes.txt": {"shape_pt_lat", "shape_pt_lon"},
}

// Validate checks coordinate values in GTFS files
func (v *CoordinateValidator) Validate(loader *parser.FeedLoader, container *notice.NoticeContainer, config validator.Config) {
	for filename, fields := range coordinateFields {
		v.validateFileCoordinates(loader, container, filename, fields)
	}
}

// validateFileCoordinates validates coordinate fields in a specific file
func (v *CoordinateValidator) validateFileCoordinates(loader *parser.FeedLoader, container *notice.NoticeContainer, filename string, coordFieldNames []string) {
	reader, err := loader.GetFile(filename)
	if err != nil {
		return // File doesn't exist, skip validation
	}
	defer reader.Close()

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

		for _, fieldName := range coordFieldNames {
			if value, exists := row.Values[fieldName]; exists && strings.TrimSpace(value) != "" {
				v.validateCoordinate(container, filename, fieldName, strings.TrimSpace(value), row.RowNumber)
			}
		}
	}
}

// validateCoordinate validates a single coordinate value
func (v *CoordinateValidator) validateCoordinate(container *notice.NoticeContainer, filename string, fieldName string, coordValue string, rowNumber int) {
	coord, err := strconv.ParseFloat(coordValue, 64)
	if err != nil {
		container.AddNotice(notice.NewInvalidCoordinateNotice(
			filename,
			fieldName,
			coordValue,
			rowNumber,
			"Invalid number format",
		))
		return
	}

	// Validate latitude range
	if strings.Contains(fieldName, "lat") {
		if coord < -90.0 || coord > 90.0 {
			container.AddNotice(notice.NewInvalidCoordinateNotice(
				filename,
				fieldName,
				coordValue,
				rowNumber,
				"Latitude must be between -90 and 90",
			))
		}
		// Check for suspicious latitude values (likely errors)
		if coord == 0.0 {
			container.AddNotice(notice.NewSuspiciousCoordinateNotice(
				filename,
				fieldName,
				coordValue,
				rowNumber,
				"Latitude is exactly 0 (may indicate missing data)",
			))
		}
	}

	// Validate longitude range
	if strings.Contains(fieldName, "lon") {
		if coord < -180.0 || coord > 180.0 {
			container.AddNotice(notice.NewInvalidCoordinateNotice(
				filename,
				fieldName,
				coordValue,
				rowNumber,
				"Longitude must be between -180 and 180",
			))
		}
		// Check for suspicious longitude values (likely errors)
		if coord == 0.0 {
			container.AddNotice(notice.NewSuspiciousCoordinateNotice(
				filename,
				fieldName,
				coordValue,
				rowNumber,
				"Longitude is exactly 0 (may indicate missing data)",
			))
		}
	}

	// Check for insufficient precision (less than 4 decimal places)
	coordStr := strings.TrimSpace(coordValue)
	if dotIndex := strings.Index(coordStr, "."); dotIndex != -1 {
		decimals := len(coordStr) - dotIndex - 1
		if decimals < 4 {
			container.AddNotice(notice.NewInsufficientCoordinatePrecisionNotice(
				filename,
				fieldName,
				coordValue,
				rowNumber,
				decimals,
			))
		}
	} else {
		// No decimal point - very low precision
		container.AddNotice(notice.NewInsufficientCoordinatePrecisionNotice(
			filename,
			fieldName,
			coordValue,
			rowNumber,
			0,
		))
	}
}
