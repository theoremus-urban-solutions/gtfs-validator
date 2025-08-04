package business

import (
	"io"
	"math"
	"strconv"
	"strings"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/parser"
	"github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

// TransferTimingValidator validates transfer timing and feasibility
type TransferTimingValidator struct{}

// NewTransferTimingValidator creates a new transfer timing validator
func NewTransferTimingValidator() *TransferTimingValidator {
	return &TransferTimingValidator{}
}

// TransferTimingInfo represents transfer information for timing validation
type TransferTimingInfo struct {
	FromStopID      string
	ToStopID        string
	TransferType    int
	MinTransferTime *int
	RowNumber       int
}

// StopLocationInfo represents stop location for distance calculations
type StopLocationInfo struct {
	StopID    string
	Latitude  float64
	Longitude float64
}

// Validate checks transfer timing and feasibility
func (v *TransferTimingValidator) Validate(loader *parser.FeedLoader, container *notice.NoticeContainer, config validator.Config) {
	transfers := v.loadTransfers(loader)
	if len(transfers) == 0 {
		return
	}

	stopLocations := v.loadStopLocations(loader)

	for _, transfer := range transfers {
		v.validateTransfer(container, transfer, stopLocations)
	}

	// Validate transfer patterns
	v.validateTransferPatterns(container, transfers, stopLocations)
}

// loadTransfers loads transfer information from transfers.txt
func (v *TransferTimingValidator) loadTransfers(loader *parser.FeedLoader) []*TransferTimingInfo {
	var transfers []*TransferTimingInfo

	reader, err := loader.GetFile("transfers.txt")
	if err != nil {
		return transfers
	}
	defer reader.Close()

	csvFile, err := parser.NewCSVFile(reader, "transfers.txt")
	if err != nil {
		return transfers
	}

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		transfer := v.parseTransfer(row)
		if transfer != nil {
			transfers = append(transfers, transfer)
		}
	}

	return transfers
}

// parseTransfer parses transfer information
func (v *TransferTimingValidator) parseTransfer(row *parser.CSVRow) *TransferTimingInfo {
	fromStopID, hasFromStop := row.Values["from_stop_id"]
	toStopID, hasToStop := row.Values["to_stop_id"]

	if !hasFromStop || !hasToStop {
		return nil
	}

	transfer := &TransferTimingInfo{
		FromStopID: strings.TrimSpace(fromStopID),
		ToStopID:   strings.TrimSpace(toStopID),
		RowNumber:  row.RowNumber,
	}

	// Parse transfer_type (optional, defaults to 0)
	if transferTypeStr, hasTransferType := row.Values["transfer_type"]; hasTransferType && strings.TrimSpace(transferTypeStr) != "" {
		if transferType, err := strconv.Atoi(strings.TrimSpace(transferTypeStr)); err == nil {
			transfer.TransferType = transferType
		}
	}

	// Parse min_transfer_time (optional)
	if minTimeStr, hasMinTime := row.Values["min_transfer_time"]; hasMinTime && strings.TrimSpace(minTimeStr) != "" {
		if minTime, err := strconv.Atoi(strings.TrimSpace(minTimeStr)); err == nil {
			transfer.MinTransferTime = &minTime
		}
	}

	return transfer
}

// loadStopLocations loads stop locations for distance calculations
func (v *TransferTimingValidator) loadStopLocations(loader *parser.FeedLoader) map[string]*StopLocationInfo {
	stops := make(map[string]*StopLocationInfo)

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

		stop := v.parseStopLocation(row)
		if stop != nil {
			stops[stop.StopID] = stop
		}
	}

	return stops
}

// parseStopLocation parses stop location information
func (v *TransferTimingValidator) parseStopLocation(row *parser.CSVRow) *StopLocationInfo {
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

	return &StopLocationInfo{
		StopID:    strings.TrimSpace(stopID),
		Latitude:  lat,
		Longitude: lon,
	}
}

// validateTransfer validates individual transfer
func (v *TransferTimingValidator) validateTransfer(container *notice.NoticeContainer, transfer *TransferTimingInfo, stopLocations map[string]*StopLocationInfo) {
	// Validate transfer type
	if !v.isValidTransferType(transfer.TransferType) {
		container.AddNotice(notice.NewInvalidTransferTypeNotice(
			transfer.FromStopID,
			transfer.ToStopID,
			transfer.TransferType,
			transfer.RowNumber,
		))
		return
	}

	// Validate transfer time based on type and distance
	v.validateTransferTime(container, transfer, stopLocations)

	// Check for self-transfers (same stop to same stop)
	if transfer.FromStopID == transfer.ToStopID {
		container.AddNotice(notice.NewTransferToSameStopNotice(
			transfer.FromStopID,
			transfer.RowNumber,
		))
	}
}

// validateTransferTime validates transfer timing
func (v *TransferTimingValidator) validateTransferTime(container *notice.NoticeContainer, transfer *TransferTimingInfo, stopLocations map[string]*StopLocationInfo) {
	fromStop, hasFromStop := stopLocations[transfer.FromStopID]
	toStop, hasToStop := stopLocations[transfer.ToStopID]

	if !hasFromStop || !hasToStop {
		return // Can't validate without coordinates
	}

	distance := v.calculateDistance(fromStop, toStop)

	switch transfer.TransferType {
	case 0: // Recommended transfer point
		v.validateRecommendedTransfer(container, transfer, distance)
	case 1: // Timed transfer point
		v.validateTimedTransfer(container, transfer, distance)
	case 2: // Minimum time required
		v.validateMinimumTimeTransfer(container, transfer, distance)
	case 3: // Transfer not possible
		v.validateNotPossibleTransfer(container, transfer, distance)
	}
}

// validateRecommendedTransfer validates recommended transfer
func (v *TransferTimingValidator) validateRecommendedTransfer(container *notice.NoticeContainer, transfer *TransferTimingInfo, distance float64) {
	if transfer.MinTransferTime != nil {
		// Recommended transfers shouldn't normally specify min_transfer_time
		container.AddNotice(notice.NewUnnecessaryMinTransferTimeNotice(
			transfer.FromStopID,
			transfer.ToStopID,
			*transfer.MinTransferTime,
			transfer.RowNumber,
		))
	}

	// Check if distance is reasonable for recommended transfer
	if distance > 500 { // More than 500 meters might be too far for recommended transfer
		container.AddNotice(notice.NewLongDistanceTransferNotice(
			transfer.FromStopID,
			transfer.ToStopID,
			distance,
			"recommended",
			transfer.RowNumber,
		))
	}
}

// validateTimedTransfer validates timed transfer
func (v *TransferTimingValidator) validateTimedTransfer(container *notice.NoticeContainer, transfer *TransferTimingInfo, distance float64) {
	if transfer.MinTransferTime != nil {
		// Timed transfers shouldn't specify min_transfer_time
		container.AddNotice(notice.NewUnnecessaryMinTransferTimeNotice(
			transfer.FromStopID,
			transfer.ToStopID,
			*transfer.MinTransferTime,
			transfer.RowNumber,
		))
	}

	// Check if distance is reasonable for timed transfer
	if distance > 200 { // More than 200 meters might be too far for timed transfer
		container.AddNotice(notice.NewLongDistanceTransferNotice(
			transfer.FromStopID,
			transfer.ToStopID,
			distance,
			"timed",
			transfer.RowNumber,
		))
	}
}

// validateMinimumTimeTransfer validates minimum time required transfer
func (v *TransferTimingValidator) validateMinimumTimeTransfer(container *notice.NoticeContainer, transfer *TransferTimingInfo, distance float64) {
	if transfer.MinTransferTime == nil {
		// Type 2 transfers must specify min_transfer_time
		container.AddNotice(notice.NewMissingMinTransferTimeNotice(
			transfer.FromStopID,
			transfer.ToStopID,
			transfer.RowNumber,
		))
		return
	}

	minTime := *transfer.MinTransferTime
	
	// Calculate expected walking time (assume 1.4 m/s walking speed)
	expectedWalkTime := int(distance / 1.4)
	
	// Check if min_transfer_time is reasonable
	if minTime < expectedWalkTime {
		container.AddNotice(notice.NewUnrealisticTransferTimeNotice(
			transfer.FromStopID,
			transfer.ToStopID,
			minTime,
			expectedWalkTime,
			distance,
			transfer.RowNumber,
		))
	}

	// Check for very long transfer times (might be error)
	if minTime > 1800 { // More than 30 minutes
		container.AddNotice(notice.NewVeryLongTransferTimeNotice(
			transfer.FromStopID,
			transfer.ToStopID,
			minTime,
			transfer.RowNumber,
		))
	}

	// Check for very short transfer times
	if minTime < 60 { // Less than 1 minute
		container.AddNotice(notice.NewVeryShortTransferTimeNotice(
			transfer.FromStopID,
			transfer.ToStopID,
			minTime,
			transfer.RowNumber,
		))
	}
}

// validateNotPossibleTransfer validates not possible transfer
func (v *TransferTimingValidator) validateNotPossibleTransfer(container *notice.NoticeContainer, transfer *TransferTimingInfo, distance float64) {
	if transfer.MinTransferTime != nil {
		// Not possible transfers shouldn't specify min_transfer_time
		container.AddNotice(notice.NewUnnecessaryMinTransferTimeNotice(
			transfer.FromStopID,
			transfer.ToStopID,
			*transfer.MinTransferTime,
			transfer.RowNumber,
		))
	}

	// Check if very close stops are marked as not possible (might be error)
	if distance < 50 { // Less than 50 meters
		container.AddNotice(notice.NewCloseStopsNotPossibleTransferNotice(
			transfer.FromStopID,
			transfer.ToStopID,
			distance,
			transfer.RowNumber,
		))
	}
}

// validateTransferPatterns validates overall transfer patterns
func (v *TransferTimingValidator) validateTransferPatterns(container *notice.NoticeContainer, transfers []*TransferTimingInfo, stopLocations map[string]*StopLocationInfo) {
	// Check for duplicate transfers
	v.validateDuplicateTransfers(container, transfers)

	// Check for bidirectional transfer consistency
	v.validateBidirectionalTransfers(container, transfers)
}

// validateDuplicateTransfers checks for duplicate transfer definitions
func (v *TransferTimingValidator) validateDuplicateTransfers(container *notice.NoticeContainer, transfers []*TransferTimingInfo) {
	seen := make(map[string][]*TransferTimingInfo)
	
	for _, transfer := range transfers {
		key := transfer.FromStopID + "_" + transfer.ToStopID
		seen[key] = append(seen[key], transfer)
	}

	for _, transferList := range seen {
		if len(transferList) > 1 {
			container.AddNotice(notice.NewDuplicateTransferNotice(
				transferList[0].FromStopID,
				transferList[0].ToStopID,
				transferList[0].RowNumber,
				transferList[1].RowNumber,
			))
		}
	}
}

// validateBidirectionalTransfers checks for bidirectional transfer consistency
func (v *TransferTimingValidator) validateBidirectionalTransfers(container *notice.NoticeContainer, transfers []*TransferTimingInfo) {
	transferMap := make(map[string]*TransferTimingInfo)
	
	for _, transfer := range transfers {
		key := transfer.FromStopID + "_" + transfer.ToStopID
		transferMap[key] = transfer
	}

	for _, transfer := range transfers {
		reverseKey := transfer.ToStopID + "_" + transfer.FromStopID
		
		if reverseTransfer, hasReverse := transferMap[reverseKey]; hasReverse {
			// Check for consistency between bidirectional transfers
			if transfer.TransferType != reverseTransfer.TransferType {
				container.AddNotice(notice.NewInconsistentBidirectionalTransferNotice(
					transfer.FromStopID,
					transfer.ToStopID,
					transfer.TransferType,
					reverseTransfer.TransferType,
					transfer.RowNumber,
				))
			}
		}
	}
}

// isValidTransferType checks if transfer type is valid
func (v *TransferTimingValidator) isValidTransferType(transferType int) bool {
	return transferType >= 0 && transferType <= 3
}

// calculateDistance calculates Haversine distance between two stops in meters
func (v *TransferTimingValidator) calculateDistance(stop1, stop2 *StopLocationInfo) float64 {
	const earthRadius = 6371000 // Earth radius in meters

	lat1Rad := stop1.Latitude * math.Pi / 180
	lat2Rad := stop2.Latitude * math.Pi / 180
	deltaLatRad := (stop2.Latitude - stop1.Latitude) * math.Pi / 180
	deltaLonRad := (stop2.Longitude - stop1.Longitude) * math.Pi / 180

	a := math.Sin(deltaLatRad/2)*math.Sin(deltaLatRad/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLonRad/2)*math.Sin(deltaLonRad/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c
}