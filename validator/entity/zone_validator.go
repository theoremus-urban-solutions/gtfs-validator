package entity

import (
	"io"
	"log"
	"strings"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/parser"
	"github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

// ZoneValidator validates zone definitions and usage
type ZoneValidator struct{}

// NewZoneValidator creates a new zone validator
func NewZoneValidator() *ZoneValidator {
	return &ZoneValidator{}
}

// ZoneInfo represents zone information from stops
type ZoneInfo struct {
	ZoneID    string
	StopID    string
	RowNumber int
}

// Validate checks zone definitions and usage
func (v *ZoneValidator) Validate(loader *parser.FeedLoader, container *notice.NoticeContainer, config validator.Config) {
	// Load zones from stops.txt
	zones := v.loadZones(loader)

	// Load zone usage from fare_rules.txt
	usedZones := v.loadUsedZones(loader)

	// Validate zones
	v.validateZones(container, zones, usedZones)
}

// loadZones loads zone information from stops.txt
func (v *ZoneValidator) loadZones(loader *parser.FeedLoader) map[string][]*ZoneInfo {
	zones := make(map[string][]*ZoneInfo)

	reader, err := loader.GetFile("stops.txt")
	if err != nil {
		return zones
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close reader %v", closeErr)
		}
	}()

	csvFile, err := parser.NewCSVFile(reader, "stops.txt")
	if err != nil {
		return zones
	}

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}

		// Check for zone_id field
		if zoneID, hasZoneID := row.Values["zone_id"]; hasZoneID {
			zoneIDTrimmed := strings.TrimSpace(zoneID)
			if zoneIDTrimmed != "" {
				stopID := ""
				if sid, hasStopID := row.Values["stop_id"]; hasStopID {
					stopID = strings.TrimSpace(sid)
				}

				zoneInfo := &ZoneInfo{
					ZoneID:    zoneIDTrimmed,
					StopID:    stopID,
					RowNumber: row.RowNumber,
				}

				zones[zoneIDTrimmed] = append(zones[zoneIDTrimmed], zoneInfo)
			}
		}
	}

	return zones
}

// loadUsedZones loads zones used in fare_rules.txt
func (v *ZoneValidator) loadUsedZones(loader *parser.FeedLoader) map[string]bool {
	usedZones := make(map[string]bool)

	reader, err := loader.GetFile("fare_rules.txt")
	if err != nil {
		return usedZones
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close reader %v", closeErr)
		}
	}()

	csvFile, err := parser.NewCSVFile(reader, "fare_rules.txt")
	if err != nil {
		return usedZones
	}

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}

		// Check origin_id
		if originID, hasOriginID := row.Values["origin_id"]; hasOriginID {
			originIDTrimmed := strings.TrimSpace(originID)
			if originIDTrimmed != "" {
				usedZones[originIDTrimmed] = true
			}
		}

		// Check destination_id
		if destinationID, hasDestinationID := row.Values["destination_id"]; hasDestinationID {
			destinationIDTrimmed := strings.TrimSpace(destinationID)
			if destinationIDTrimmed != "" {
				usedZones[destinationIDTrimmed] = true
			}
		}

		// Check contains_id
		if containsID, hasContainsID := row.Values["contains_id"]; hasContainsID {
			containsIDTrimmed := strings.TrimSpace(containsID)
			if containsIDTrimmed != "" {
				usedZones[containsIDTrimmed] = true
			}
		}
	}

	return usedZones
}

// validateZones validates zone consistency
func (v *ZoneValidator) validateZones(container *notice.NoticeContainer, zones map[string][]*ZoneInfo, usedZones map[string]bool) {
	// Check for single stop zones (regardless of usage)
	for zoneID, zoneInfos := range zones {
		if len(zoneInfos) == 1 {
			// Single stop in zone - might be a data quality issue
			container.AddNotice(notice.NewSingleStopZoneNotice(
				zoneID,
				zoneInfos[0].StopID,
				zoneInfos[0].RowNumber,
			))
		}
	}

	// Check for unused zones
	for zoneID, zoneInfos := range zones {
		if !usedZones[zoneID] {
			// Zone defined but not used in fare rules
			container.AddNotice(notice.NewUnusedZoneNotice(
				zoneID,
				zoneInfos[0].RowNumber, // Use first occurrence
			))
		}
	}

	// Check for referenced but undefined zones
	for zoneID := range usedZones {
		if _, exists := zones[zoneID]; !exists {
			container.AddNotice(notice.NewUndefinedZoneNotice(zoneID))
		}
	}

	// Check for zone naming conventions
	v.validateZoneNaming(container, zones)
}

// validateZoneNaming checks zone naming patterns
func (v *ZoneValidator) validateZoneNaming(container *notice.NoticeContainer, zones map[string][]*ZoneInfo) {
	// Check for very long zone IDs
	for zoneID, zoneInfos := range zones {
		if len(zoneID) > 50 {
			container.AddNotice(notice.NewLongZoneIDNotice(
				zoneID,
				len(zoneID),
				zoneInfos[0].RowNumber,
			))
		}

		// Check for zones that look like stop IDs
		if strings.Contains(zoneID, "_") || strings.Contains(zoneID, "-") {
			// This might indicate using stop_id as zone_id
			if len(zoneInfos) == 1 && zoneID == zoneInfos[0].StopID {
				container.AddNotice(notice.NewZoneIDSameAsStopIDNotice(
					zoneID,
					zoneInfos[0].RowNumber,
				))
			}
		}
	}
}
