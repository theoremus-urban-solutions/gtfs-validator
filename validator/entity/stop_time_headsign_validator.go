package entity

import (
	"io"
	"strconv"
	"strings"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/parser"
	"github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

// StopTimeHeadsignValidator validates headsign consistency in stop_times.txt
type StopTimeHeadsignValidator struct{}

// NewStopTimeHeadsignValidator creates a new stop time headsign validator
func NewStopTimeHeadsignValidator() *StopTimeHeadsignValidator {
	return &StopTimeHeadsignValidator{}
}

// StopTimeHeadsignInfo represents stop time headsign information
type StopTimeHeadsignInfo struct {
	TripID       string
	StopSequence int
	StopHeadsign string
	RowNumber    int
}

// TripHeadsignInfo represents trip headsign information
type TripHeadsignInfo struct {
	TripID       string
	TripHeadsign string
	RouteID      string
}

// Validate checks headsign consistency within trips and across stops
func (v *StopTimeHeadsignValidator) Validate(loader *parser.FeedLoader, container *notice.NoticeContainer, config validator.Config) {
	stopTimeHeadsigns := v.loadStopTimeHeadsigns(loader)
	if len(stopTimeHeadsigns) == 0 {
		return
	}

	tripHeadsigns := v.loadTripHeadsigns(loader)

	// Group by trip_id
	tripStopHeadsigns := make(map[string][]*StopTimeHeadsignInfo)
	for _, sth := range stopTimeHeadsigns {
		tripStopHeadsigns[sth.TripID] = append(tripStopHeadsigns[sth.TripID], sth)
	}

	// Validate each trip's stop headsigns
	for tripID, headsigns := range tripStopHeadsigns {
		v.validateTripHeadsigns(container, tripID, headsigns, tripHeadsigns[tripID])
	}
}

// loadStopTimeHeadsigns loads stop time headsign information
func (v *StopTimeHeadsignValidator) loadStopTimeHeadsigns(loader *parser.FeedLoader) []*StopTimeHeadsignInfo {
	var headsigns []*StopTimeHeadsignInfo

	reader, err := loader.GetFile("stop_times.txt")
	if err != nil {
		return headsigns
	}
	defer reader.Close()

	csvFile, err := parser.NewCSVFile(reader, "stop_times.txt")
	if err != nil {
		return headsigns
	}

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		headsign := v.parseStopTimeHeadsign(row)
		if headsign != nil {
			headsigns = append(headsigns, headsign)
		}
	}

	return headsigns
}

// parseStopTimeHeadsign parses stop time headsign information
func (v *StopTimeHeadsignValidator) parseStopTimeHeadsign(row *parser.CSVRow) *StopTimeHeadsignInfo {
	tripID, hasTripID := row.Values["trip_id"]
	seqStr, hasSeq := row.Values["stop_sequence"]

	if !hasTripID || !hasSeq {
		return nil
	}

	seq, err := strconv.Atoi(strings.TrimSpace(seqStr))
	if err != nil {
		return nil
	}

	headsign := &StopTimeHeadsignInfo{
		TripID:       strings.TrimSpace(tripID),
		StopSequence: seq,
		RowNumber:    row.RowNumber,
	}

	// Parse optional stop_headsign
	if stopHeadsign, hasHeadsign := row.Values["stop_headsign"]; hasHeadsign {
		headsign.StopHeadsign = strings.TrimSpace(stopHeadsign)
	}

	return headsign
}

// loadTripHeadsigns loads trip headsign information for comparison
func (v *StopTimeHeadsignValidator) loadTripHeadsigns(loader *parser.FeedLoader) map[string]*TripHeadsignInfo {
	trips := make(map[string]*TripHeadsignInfo)

	reader, err := loader.GetFile("trips.txt")
	if err != nil {
		return trips
	}
	defer reader.Close()

	csvFile, err := parser.NewCSVFile(reader, "trips.txt")
	if err != nil {
		return trips
	}

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		trip := v.parseTripHeadsign(row)
		if trip != nil {
			trips[trip.TripID] = trip
		}
	}

	return trips
}

// parseTripHeadsign parses trip headsign information
func (v *StopTimeHeadsignValidator) parseTripHeadsign(row *parser.CSVRow) *TripHeadsignInfo {
	tripID, hasTripID := row.Values["trip_id"]
	routeID, hasRouteID := row.Values["route_id"]

	if !hasTripID || !hasRouteID {
		return nil
	}

	trip := &TripHeadsignInfo{
		TripID:  strings.TrimSpace(tripID),
		RouteID: strings.TrimSpace(routeID),
	}

	// Parse optional trip_headsign
	if tripHeadsign, hasHeadsign := row.Values["trip_headsign"]; hasHeadsign {
		trip.TripHeadsign = strings.TrimSpace(tripHeadsign)
	}

	return trip
}

// validateTripHeadsigns validates headsigns for a single trip
func (v *StopTimeHeadsignValidator) validateTripHeadsigns(container *notice.NoticeContainer, tripID string, stopHeadsigns []*StopTimeHeadsignInfo, tripInfo *TripHeadsignInfo) {
	if len(stopHeadsigns) == 0 {
		return
	}

	// Check for headsign inconsistencies within the trip
	v.validateHeadsignConsistency(container, tripID, stopHeadsigns)

	// Check stop headsigns against trip headsign
	if tripInfo != nil {
		v.validateStopTripHeadsignConsistency(container, tripID, stopHeadsigns, tripInfo)
	}

	// Check for suspicious headsign patterns
	v.validateHeadsignPatterns(container, tripID, stopHeadsigns)
}

// validateHeadsignConsistency checks for inconsistent headsigns within a trip
func (v *StopTimeHeadsignValidator) validateHeadsignConsistency(container *notice.NoticeContainer, tripID string, stopHeadsigns []*StopTimeHeadsignInfo) {
	// Count different headsigns
	headsignCounts := make(map[string]int)
	headsignFirstOccurrence := make(map[string]int)

	for _, sh := range stopHeadsigns {
		if sh.StopHeadsign != "" {
			headsignCounts[sh.StopHeadsign]++
			if _, exists := headsignFirstOccurrence[sh.StopHeadsign]; !exists {
				headsignFirstOccurrence[sh.StopHeadsign] = sh.RowNumber
			}
		}
	}

	// Warning if too many different headsigns in one trip
	if len(headsignCounts) > 5 {
		var headsigns []string
		for headsign := range headsignCounts {
			headsigns = append(headsigns, headsign)
		}

		container.AddNotice(notice.NewTooManyHeadsignsInTripNotice(
			tripID,
			len(headsignCounts),
			headsigns,
		))
	}

	// Check for headsign changes within trip sequence
	v.validateHeadsignSequence(container, tripID, stopHeadsigns)
}

// validateHeadsignSequence checks headsign changes along trip sequence
func (v *StopTimeHeadsignValidator) validateHeadsignSequence(container *notice.NoticeContainer, tripID string, stopHeadsigns []*StopTimeHeadsignInfo) {
	var prevHeadsign string
	var prevSequence int
	headsignChanges := 0

	for _, sh := range stopHeadsigns {
		if strings.TrimSpace(sh.StopHeadsign) != "" {
			if prevHeadsign != "" && sh.StopHeadsign != prevHeadsign {
				headsignChanges++

				// Info notice for headsign change
				container.AddNotice(notice.NewHeadsignChangeWithinTripNotice(
					tripID,
					prevSequence,
					sh.StopSequence,
					prevHeadsign,
					sh.StopHeadsign,
					sh.RowNumber,
				))
			}
			prevHeadsign = sh.StopHeadsign
			prevSequence = sh.StopSequence
		}
	}

	// Warning if too many headsign changes
	if headsignChanges > 3 {
		container.AddNotice(notice.NewFrequentHeadsignChangesNotice(
			tripID,
			headsignChanges,
		))
	}
}

// validateStopTripHeadsignConsistency compares stop headsigns with trip headsign
func (v *StopTimeHeadsignValidator) validateStopTripHeadsignConsistency(container *notice.NoticeContainer, tripID string, stopHeadsigns []*StopTimeHeadsignInfo, tripInfo *TripHeadsignInfo) {
	if tripInfo.TripHeadsign == "" {
		return // No trip headsign to compare against
	}

	// Check if any stop headsigns conflict with trip headsign
	for _, sh := range stopHeadsigns {
		if sh.StopHeadsign != "" && !v.areHeadsignsConsistent(sh.StopHeadsign, tripInfo.TripHeadsign) {
			container.AddNotice(notice.NewStopTripHeadsignMismatchNotice(
				tripID,
				sh.StopSequence,
				sh.StopHeadsign,
				tripInfo.TripHeadsign,
				sh.RowNumber,
			))
		}
	}
}

// areHeadsignsConsistent checks if two headsigns are reasonably consistent
func (v *StopTimeHeadsignValidator) areHeadsignsConsistent(stopHeadsign, tripHeadsign string) bool {
	// Normalize for comparison
	stop := strings.ToLower(strings.TrimSpace(stopHeadsign))
	trip := strings.ToLower(strings.TrimSpace(tripHeadsign))

	if stop == "" || trip == "" {
		return false
	}

	// Direct match
	if stop == trip {
		return true
	}

	// Check if one contains the other (allowing for abbreviations)
	if strings.Contains(stop, trip) || strings.Contains(trip, stop) {
		return true
	}

	// Check for common abbreviations and variations
	if v.areHeadsignVariations(stop, trip) {
		return true
	}

	return false
}

// areHeadsignVariations checks for common headsign variations
func (v *StopTimeHeadsignValidator) areHeadsignVariations(headsign1, headsign2 string) bool {
	// Common abbreviation patterns
	abbreviations := map[string][]string{
		"street":     {"st", "str"},
		"avenue":     {"ave", "av"},
		"boulevard":  {"blvd", "blv"},
		"downtown":   {"dtown", "dt", "dwtn"},
		"center":     {"ctr", "cntr"},
		"station":    {"stn", "sta"},
		"terminal":   {"term", "trml"},
		"university": {"univ", "u"},
		"hospital":   {"hosp", "hsp"},
		"airport":    {"apt", "airpt"},
	}

	for full, abbrevs := range abbreviations {
		for _, abbrev := range abbrevs {
			if (strings.Contains(headsign1, full) && strings.Contains(headsign2, abbrev)) ||
				(strings.Contains(headsign1, abbrev) && strings.Contains(headsign2, full)) {
				return true
			}
		}
	}

	return false
}

// validateHeadsignPatterns checks for suspicious headsign patterns
func (v *StopTimeHeadsignValidator) validateHeadsignPatterns(container *notice.NoticeContainer, tripID string, stopHeadsigns []*StopTimeHeadsignInfo) {
	for _, sh := range stopHeadsigns {
		if sh.StopHeadsign == "" {
			continue // Skip empty headsigns
		}

		// Check for very short headsigns (might be data quality issue)
		if len(strings.TrimSpace(sh.StopHeadsign)) <= 2 {
			container.AddNotice(notice.NewVeryShortHeadsignNotice(
				tripID,
				sh.StopSequence,
				sh.StopHeadsign,
				sh.RowNumber,
			))
		}

		// Check for very long headsigns (might be formatting issue)
		if len(sh.StopHeadsign) > 100 {
			container.AddNotice(notice.NewVeryLongHeadsignNotice(
				tripID,
				sh.StopSequence,
				len(sh.StopHeadsign),
				sh.RowNumber,
			))
		}

		// Check for suspicious characters or patterns
		v.validateHeadsignContent(container, tripID, sh)
	}
}

// validateHeadsignContent validates headsign content quality
func (v *StopTimeHeadsignValidator) validateHeadsignContent(container *notice.NoticeContainer, tripID string, sh *StopTimeHeadsignInfo) {
	headsign := sh.StopHeadsign

	// Check for all caps (might be formatting issue)
	if len(headsign) > 5 && headsign == strings.ToUpper(headsign) && !v.containsLowerCase(headsign) {
		container.AddNotice(notice.NewAllCapsHeadsignNotice(
			tripID,
			sh.StopSequence,
			headsign,
			sh.RowNumber,
		))
	}

	// Check for excessive punctuation
	punctuationCount := 0
	for _, char := range headsign {
		if strings.ContainsRune("!@#$%^&*()[]{}|\\:;\"'<>?", char) {
			punctuationCount++
		}
	}

	if punctuationCount > 5 {
		container.AddNotice(notice.NewExcessivePunctuationHeadsignNotice(
			tripID,
			sh.StopSequence,
			headsign,
			punctuationCount,
			sh.RowNumber,
		))
	}

	// Check for suspicious patterns like "NULL", "N/A", "UNKNOWN"
	suspiciousPatterns := []string{"null", "n/a", "unknown", "none", "tbd", "tba", "test"}
	lowerHeadsign := strings.ToLower(headsign)

	for _, pattern := range suspiciousPatterns {
		if lowerHeadsign == pattern || strings.Contains(lowerHeadsign, pattern) {
			container.AddNotice(notice.NewSuspiciousHeadsignPatternNotice(
				tripID,
				sh.StopSequence,
				headsign,
				pattern,
				sh.RowNumber,
			))
		}
	}
}

// containsLowerCase checks if string contains any lowercase letters
func (v *StopTimeHeadsignValidator) containsLowerCase(s string) bool {
	for _, char := range s {
		if char >= 'a' && char <= 'z' {
			return true
		}
	}
	return false
}
