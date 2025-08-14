package entity

import (
	"io"
	"log"
	"strconv"
	"strings"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/parser"
	"github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

// StopNameValidator validates stop names are properly specified
type StopNameValidator struct{}

// NewStopNameValidator creates a new stop name validator
func NewStopNameValidator() *StopNameValidator {
	return &StopNameValidator{}
}

// StopNameInfo represents stop naming information
type StopNameInfo struct {
	StopID        string
	StopName      string
	StopDesc      string
	LocationType  int
	ParentStation string
	RowNumber     int
}

// Validate checks stop names for various issues
func (v *StopNameValidator) Validate(loader *parser.FeedLoader, container *notice.NoticeContainer, config validator.Config) {
	stops := v.loadStops(loader)

	// Build parent station map for context
	parentStations := make(map[string]*StopNameInfo)
	for _, stop := range stops {
		if stop.LocationType == 1 { // Station
			parentStations[stop.StopID] = stop
		}
	}

	for _, stop := range stops {
		v.validateStopName(container, stop, parentStations)
	}
}

// loadStops loads stop information from stops.txt
func (v *StopNameValidator) loadStops(loader *parser.FeedLoader) []*StopNameInfo {
	var stops []*StopNameInfo

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

		stop := v.parseStop(row)
		if stop != nil {
			stops = append(stops, stop)
		}
	}

	return stops
}

// parseStop parses stop information from a row
func (v *StopNameValidator) parseStop(row *parser.CSVRow) *StopNameInfo {
	stopID, hasStopID := row.Values["stop_id"]
	if !hasStopID {
		return nil
	}

	stop := &StopNameInfo{
		StopID:    strings.TrimSpace(stopID),
		RowNumber: row.RowNumber,
	}

	// Parse stop_name
	if stopName, hasStopName := row.Values["stop_name"]; hasStopName {
		stop.StopName = strings.TrimSpace(stopName)
	}

	// Parse stop_desc
	if stopDesc, hasStopDesc := row.Values["stop_desc"]; hasStopDesc {
		stop.StopDesc = strings.TrimSpace(stopDesc)
	}

	// Parse location_type (defaults to 0)
	if locTypeStr, hasLocType := row.Values["location_type"]; hasLocType && strings.TrimSpace(locTypeStr) != "" {
		if locType, err := strconv.Atoi(strings.TrimSpace(locTypeStr)); err == nil {
			stop.LocationType = locType
		}
	}

	// Parse parent_station
	if parentStation, hasParent := row.Values["parent_station"]; hasParent {
		stop.ParentStation = strings.TrimSpace(parentStation)
	}

	return stop
}

// validateStopName validates a single stop's naming
func (v *StopNameValidator) validateStopName(container *notice.NoticeContainer, stop *StopNameInfo, parentStations map[string]*StopNameInfo) {
	// Check if stop_name is required for this location type
	nameRequired := v.isStopNameRequired(stop.LocationType)

	if nameRequired && stop.StopName == "" {
		// Check if this is a child stop that might inherit name from parent
		if stop.ParentStation != "" {
			if parent, exists := parentStations[stop.ParentStation]; exists && parent.StopName != "" {
				// Child can inherit parent name, but should be noted as INFO
				container.AddNotice(notice.NewStopNameMissingButInheritedNotice(
					stop.StopID,
					stop.ParentStation,
					parent.StopName,
					stop.LocationType,
					stop.RowNumber,
				))
			} else {
				// Parent doesn't have a name either
				container.AddNotice(notice.NewMissingRequiredStopNameNotice(
					stop.StopID,
					stop.LocationType,
					stop.RowNumber,
				))
			}
		} else {
			// No parent station, name is definitely required
			container.AddNotice(notice.NewMissingRequiredStopNameNotice(
				stop.StopID,
				stop.LocationType,
				stop.RowNumber,
			))
		}
	}

	// Additional validations only if name exists
	if stop.StopName != "" {
		// Check for generic/placeholder names
		v.checkGenericStopName(container, stop)

		// Check for excessive length
		v.checkStopNameLength(container, stop)

		// Check for problematic characters
		v.checkProblematicCharacters(container, stop)

		// Check if name and description are identical
		v.checkNameDescriptionDuplicate(container, stop)

		// Check for all caps names (poor readability)
		v.checkAllCapsName(container, stop)

		// Check for repeated words
		v.checkRepeatedWords(container, stop)
	}
}

// isStopNameRequired checks if stop_name is required for the location type
func (v *StopNameValidator) isStopNameRequired(locationType int) bool {
	// stop_name is required for:
	// 0 = Stop/Platform
	// 1 = Station
	// 2 = Entrance/Exit
	// stop_name is optional for:
	// 3 = Generic Node
	// 4 = Boarding Area
	return locationType <= 2
}

// checkGenericStopName checks for generic or placeholder stop names
func (v *StopNameValidator) checkGenericStopName(container *notice.NoticeContainer, stop *StopNameInfo) {
	genericNames := []string{
		"stop",
		"station",
		"platform",
		"entrance",
		"exit",
		"node",
		"boarding",
		"test",
		"temp",
		"placeholder",
		"unnamed",
		"unknown",
		"tbd",
		"todo",
		"xxx",
		"???",
	}

	lowerName := strings.ToLower(stop.StopName)
	for _, generic := range genericNames {
		if lowerName == generic || lowerName == generic+" "+generic {
			container.AddNotice(notice.NewGenericStopNameNotice(
				stop.StopID,
				stop.StopName,
				stop.RowNumber,
			))
			break
		}
	}
}

// checkStopNameLength checks for excessively long stop names
func (v *StopNameValidator) checkStopNameLength(container *notice.NoticeContainer, stop *StopNameInfo) {
	const maxRecommendedLength = 100
	const maxAllowedLength = 255

	nameLength := len(stop.StopName)

	if nameLength > maxAllowedLength {
		container.AddNotice(notice.NewStopNameTooLongNotice(
			stop.StopID,
			stop.StopName,
			nameLength,
			maxAllowedLength,
			stop.RowNumber,
			notice.ERROR,
		))
	} else if nameLength > maxRecommendedLength {
		container.AddNotice(notice.NewStopNameTooLongNotice(
			stop.StopID,
			stop.StopName,
			nameLength,
			maxRecommendedLength,
			stop.RowNumber,
			notice.WARNING,
		))
	}
}

// checkProblematicCharacters checks for problematic characters in stop names
func (v *StopNameValidator) checkProblematicCharacters(container *notice.NoticeContainer, stop *StopNameInfo) {
	// Check for control characters
	for i, ch := range stop.StopName {
		if ch < 32 && ch != 9 && ch != 10 && ch != 13 { // Allow tab, newline, carriage return
			container.AddNotice(notice.NewStopNameContainsControlCharacterNotice(
				stop.StopID,
				stop.StopName,
				i,
				int(ch),
				stop.RowNumber,
			))
		}
	}

	// Check for HTML/XML tags
	if strings.Contains(stop.StopName, "<") && strings.Contains(stop.StopName, ">") {
		container.AddNotice(notice.NewStopNameContainsHTMLNotice(
			stop.StopID,
			stop.StopName,
			stop.RowNumber,
		))
	}

	// Check for URL-like content
	if strings.Contains(stop.StopName, "http://") || strings.Contains(stop.StopName, "https://") || strings.Contains(stop.StopName, "www.") {
		container.AddNotice(notice.NewStopNameContainsURLNotice(
			stop.StopID,
			stop.StopName,
			stop.RowNumber,
		))
	}
}

// checkNameDescriptionDuplicate checks if stop_name and stop_desc are identical
func (v *StopNameValidator) checkNameDescriptionDuplicate(container *notice.NoticeContainer, stop *StopNameInfo) {
	if stop.StopDesc != "" && stop.StopName == stop.StopDesc {
		container.AddNotice(notice.NewStopNameDescriptionDuplicateNotice(
			stop.StopID,
			stop.StopName,
			stop.RowNumber,
		))
	}
}

// checkAllCapsName checks for all-caps stop names
func (v *StopNameValidator) checkAllCapsName(container *notice.NoticeContainer, stop *StopNameInfo) {
	// Skip if name is very short (like abbreviations)
	if len(stop.StopName) <= 3 {
		return
	}

	// Check if all letters are uppercase
	hasLowerCase := false
	letterCount := 0
	for _, ch := range stop.StopName {
		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') {
			letterCount++
			if ch >= 'a' && ch <= 'z' {
				hasLowerCase = true
			}
		}
	}

	// If there are letters and none are lowercase, it's all caps
	if letterCount > 0 && !hasLowerCase {
		container.AddNotice(notice.NewStopNameAllCapsNotice(
			stop.StopID,
			stop.StopName,
			stop.RowNumber,
		))
	}
}

// checkRepeatedWords checks for repeated words in stop names
func (v *StopNameValidator) checkRepeatedWords(container *notice.NoticeContainer, stop *StopNameInfo) {
	// Split name into words
	words := strings.Fields(stop.StopName)
	if len(words) < 2 {
		return
	}

	// Check for consecutive repeated words
	for i := 1; i < len(words); i++ {
		if strings.EqualFold(words[i], words[i-1]) && len(words[i]) > 2 {
			container.AddNotice(notice.NewStopNameRepeatedWordNotice(
				stop.StopID,
				stop.StopName,
				words[i],
				stop.RowNumber,
			))
			break
		}
	}
}
