package business

import (
	"io"
	"log"
	"strconv"
	"strings"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/parser"
	"github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

// TransferValidator validates transfer definitions
type TransferValidator struct{}

// NewTransferValidator creates a new transfer validator
func NewTransferValidator() *TransferValidator {
	return &TransferValidator{}
}

// TransferInfo represents transfer information
type TransferInfo struct {
	FromStopID      string
	ToStopID        string
	FromRouteID     string
	ToRouteID       string
	FromTripID      string
	ToTripID        string
	TransferType    int
	MinTransferTime *int
	RowNumber       int
}

// stopReference is what the transfer checks need to know about a stop: whether
// a transfer may name it at all, which only stops/platforms and stations may
// be, and where it is, so the walk between the two ends can be measured.
// Location is nil for a stop with unusable coordinates, which the coordinate
// checks report.
type stopReference struct {
	LocationType int
	Location     *StopLocation
}

// tripReference is the route a trip belongs to, and the stops it calls at, so
// a transfer naming both can be checked for agreement.
type tripReference struct {
	RouteID string
	Stops   map[string]bool
}

// Validate checks transfer definitions
func (v *TransferValidator) Validate(loader *parser.FeedLoader, container *notice.NoticeContainer, config validator.Config) {
	transfers := v.loadTransfers(loader)

	if len(transfers) == 0 {
		return
	}

	// Load stop information for validation
	stops := v.loadStops(loader)
	trips := v.loadTrips(loader, transfers)

	// Validate each transfer
	for _, transfer := range transfers {
		v.validateTransfer(container, transfer, stops)
		v.validateTransferTripReferences(container, transfer, trips)
	}

	// Check for duplicate transfers
	v.validateDuplicateTransfers(container, transfers)
}

// loadTransfers loads transfer information from transfers.txt
func (v *TransferValidator) loadTransfers(loader *parser.FeedLoader) []*TransferInfo {
	var transfers []*TransferInfo

	reader, err := loader.GetFile("transfers.txt")
	if err != nil {
		return transfers
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close reader %v", closeErr)
		}
	}()

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
			break
		}

		transfer := v.parseTransfer(row)
		if transfer != nil {
			transfers = append(transfers, transfer)
		}
	}

	return transfers
}

// parseTransfer parses a transfer record
func (v *TransferValidator) parseTransfer(row *parser.CSVRow) *TransferInfo {
	fromStopID, hasFromStopID := row.Values["from_stop_id"]
	toStopID, hasToStopID := row.Values["to_stop_id"]
	transferTypeStr, hasTransferType := row.Values["transfer_type"]

	if !hasFromStopID || !hasToStopID || !hasTransferType {
		return nil
	}

	transferType, err := strconv.Atoi(strings.TrimSpace(transferTypeStr))
	if err != nil {
		return nil
	}

	transfer := &TransferInfo{
		FromStopID:   strings.TrimSpace(fromStopID),
		ToStopID:     strings.TrimSpace(toStopID),
		FromRouteID:  strings.TrimSpace(row.Values["from_route_id"]),
		ToRouteID:    strings.TrimSpace(row.Values["to_route_id"]),
		FromTripID:   strings.TrimSpace(row.Values["from_trip_id"]),
		ToTripID:     strings.TrimSpace(row.Values["to_trip_id"]),
		TransferType: transferType,
		RowNumber:    row.RowNumber,
	}

	// Parse min_transfer_time if present
	if minTransferTimeStr, hasMinTransferTime := row.Values["min_transfer_time"]; hasMinTransferTime && strings.TrimSpace(minTransferTimeStr) != "" {
		if minTransferTime, err := strconv.Atoi(strings.TrimSpace(minTransferTimeStr)); err == nil {
			transfer.MinTransferTime = &minTransferTime
		}
	}

	return transfer
}

// loadStops loads every stop with its location type.
func (v *TransferValidator) loadStops(loader *parser.FeedLoader) map[string]*stopReference {
	stops := make(map[string]*stopReference)

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
			break
		}

		stopID, hasStopID := row.Values["stop_id"]
		if !hasStopID {
			continue
		}
		stop := &stopReference{}
		if locationType, err := strconv.Atoi(strings.TrimSpace(row.Values["location_type"])); err == nil {
			stop.LocationType = locationType
		}
		lat, latErr := strconv.ParseFloat(strings.TrimSpace(row.Values["stop_lat"]), 64)
		lon, lonErr := strconv.ParseFloat(strings.TrimSpace(row.Values["stop_lon"]), 64)
		if latErr == nil && lonErr == nil {
			stop.Location = &StopLocation{Latitude: lat, Longitude: lon}
		}
		stops[strings.TrimSpace(stopID)] = stop
	}

	return stops
}

// loadTrips loads the route and stop set of only the trips transfers.txt
// names, so a feed without trip transfers pays nothing.
func (v *TransferValidator) loadTrips(loader *parser.FeedLoader, transfers []*TransferInfo) map[string]*tripReference {
	wanted := make(map[string]bool)
	for _, transfer := range transfers {
		for _, tripID := range []string{transfer.FromTripID, transfer.ToTripID} {
			if tripID != "" {
				wanted[tripID] = true
			}
		}
	}
	trips := make(map[string]*tripReference, len(wanted))
	if len(wanted) == 0 {
		return trips
	}

	if reader, err := loader.GetFile("trips.txt"); err == nil {
		defer func() {
			if closeErr := reader.Close(); closeErr != nil {
				log.Printf("Warning: failed to close reader %v", closeErr)
			}
		}()
		if csvFile, err := parser.NewCSVFile(reader, "trips.txt"); err == nil {
			for {
				row, err := csvFile.ReadRow()
				if err == io.EOF {
					break
				}
				if err != nil {
					break
				}
				tripID := strings.TrimSpace(row.Values["trip_id"])
				if !wanted[tripID] {
					continue
				}
				trips[tripID] = &tripReference{
					RouteID: strings.TrimSpace(row.Values["route_id"]),
					Stops:   make(map[string]bool),
				}
			}
		}
	}

	reader, err := loader.GetFile("stop_times.txt")
	if err != nil {
		return trips
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close reader %v", closeErr)
		}
	}()
	csvFile, err := parser.NewCSVFile(reader, "stop_times.txt")
	if err != nil {
		return trips
	}
	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}
		trip, exists := trips[strings.TrimSpace(row.Values["trip_id"])]
		if !exists {
			continue
		}
		if stopID := strings.TrimSpace(row.Values["stop_id"]); stopID != "" {
			trip.Stops[stopID] = true
		}
	}

	return trips
}

// validateTransferTripReferences checks a transfer's trip references against
// the routes and stops those trips actually have.
func (v *TransferValidator) validateTransferTripReferences(container *notice.NoticeContainer, transfer *TransferInfo, trips map[string]*tripReference) {
	ends := []struct {
		tripField, tripID   string
		routeField, routeID string
		stopField, stopID   string
	}{
		{"from_trip_id", transfer.FromTripID, "from_route_id", transfer.FromRouteID, "from_stop_id", transfer.FromStopID},
		{"to_trip_id", transfer.ToTripID, "to_route_id", transfer.ToRouteID, "to_stop_id", transfer.ToStopID},
	}

	for _, end := range ends {
		if end.tripID == "" {
			continue
		}
		trip, exists := trips[end.tripID]
		if !exists {
			container.AddNotice(notice.NewForeignKeyViolationNotice(
				"transfers.txt",
				end.tripField,
				end.tripID,
				transfer.RowNumber,
				"trips.txt",
				"trip_id",
			))
			continue
		}

		if end.routeID != "" && trip.RouteID != end.routeID {
			container.AddNotice(notice.NewTransferWithInvalidTripAndRouteNotice(
				transfer.RowNumber,
				end.tripField, end.tripID,
				end.routeField, end.routeID,
				trip.RouteID,
			))
		}

		if end.stopID != "" && len(trip.Stops) > 0 && !trip.Stops[end.stopID] {
			container.AddNotice(notice.NewTransferWithInvalidTripAndStopNotice(
				transfer.RowNumber,
				end.tripField, end.tripID,
				end.stopField, end.stopID,
			))
		}
	}
}

// validateTransfer validates a single transfer record
func (v *TransferValidator) validateTransfer(container *notice.NoticeContainer, transfer *TransferInfo, stops map[string]*stopReference) {
	// transfer_type's own value is checked by core/field_type_validator.go.

	// Validate stop references. A transfer happens between places a passenger
	// can stand, so only stops/platforms and stations may be named.
	for _, end := range []struct{ field, stopID string }{
		{"from_stop_id", transfer.FromStopID},
		{"to_stop_id", transfer.ToStopID},
	} {
		stop, exists := stops[end.stopID]
		if !exists {
			container.AddNotice(notice.NewForeignKeyViolationNotice(
				"transfers.txt",
				end.field,
				end.stopID,
				transfer.RowNumber,
				"stops.txt",
				"stop_id",
			))
			continue
		}
		if stop.LocationType != 0 && stop.LocationType != 1 {
			container.AddNotice(notice.NewTransferWithInvalidStopLocationTypeNotice(
				transfer.RowNumber,
				end.field,
				end.stopID,
				stop.LocationType,
			))
		}
	}

	// Validate transfer from/to same stop
	if transfer.FromStopID == transfer.ToStopID {
		container.AddNotice(notice.NewTransferToSameStopNotice(
			transfer.FromStopID,
			transfer.RowNumber,
		))
	}

	// Validate min_transfer_time requirements
	v.validateMinTransferTime(container, transfer)

	v.validateTransferDistance(container, transfer, stops)
}

const (
	// maxTransferDistanceMetres is the distance beyond which a transfer stops
	// describing a connection anyone can make and starts describing a typo in
	// one of the two stop_ids.
	maxTransferDistanceMetres = 10000.0

	// noteworthyTransferDistanceMetres is long enough to be worth a look — a
	// sprawling station complex or a deliberate timed connection can reach it
	// — without being wrong on its face.
	noteworthyTransferDistanceMetres = 2000.0
)

// validateTransferDistance measures the walk a transfer asks a passenger to
// make. The two thresholds report the same measurement at different
// confidences, so only the worse of the two fires on any one row.
func (v *TransferValidator) validateTransferDistance(container *notice.NoticeContainer, transfer *TransferInfo, stops map[string]*stopReference) {
	from, hasFrom := stops[transfer.FromStopID]
	to, hasTo := stops[transfer.ToStopID]
	if !hasFrom || !hasTo || from.Location == nil || to.Location == nil {
		return
	}

	metres := haversineMetres(
		from.Location.Latitude, from.Location.Longitude,
		to.Location.Latitude, to.Location.Longitude,
	)

	switch {
	case metres > maxTransferDistanceMetres:
		container.AddNotice(notice.NewTransferDistanceTooLargeNotice(
			transfer.FromStopID, transfer.ToStopID, metres, transfer.RowNumber,
		))
	case metres > noteworthyTransferDistanceMetres:
		container.AddNotice(notice.NewTransferDistanceAbove2KmNotice(
			transfer.FromStopID, transfer.ToStopID, metres, transfer.RowNumber,
		))
	}
}

// validateMinTransferTime validates min_transfer_time field
func (v *TransferValidator) validateMinTransferTime(container *notice.NoticeContainer, transfer *TransferInfo) {
	// min_transfer_time is required for transfer_type = 2
	if transfer.TransferType == 2 && transfer.MinTransferTime == nil {
		container.AddNotice(notice.NewMissingMinTransferTimeNotice(
			transfer.FromStopID,
			transfer.ToStopID,
			transfer.RowNumber,
		))
		return
	}

	// min_transfer_time should not be used for transfer_type = 3
	if transfer.TransferType == 3 && transfer.MinTransferTime != nil {
		container.AddNotice(notice.NewUnnecessaryMinTransferTimeNotice(
			transfer.FromStopID,
			transfer.ToStopID,
			*transfer.MinTransferTime,
			transfer.RowNumber,
		))
	}

	// Validate min_transfer_time value
	if transfer.MinTransferTime != nil {
		if *transfer.MinTransferTime < 0 {
			container.AddNotice(notice.NewNegativeMinTransferTimeNotice(
				transfer.FromStopID,
				transfer.ToStopID,
				*transfer.MinTransferTime,
				transfer.RowNumber,
			))
		}

		// Check for unreasonably long transfer times (more than 1 hour)
		if *transfer.MinTransferTime > 3600 {
			container.AddNotice(notice.NewUnreasonableMinTransferTimeNotice(
				transfer.FromStopID,
				transfer.ToStopID,
				*transfer.MinTransferTime,
				transfer.RowNumber,
			))
		}
	}
}

// validateDuplicateTransfers checks for duplicate transfer definitions
func (v *TransferValidator) validateDuplicateTransfers(container *notice.NoticeContainer, transfers []*TransferInfo) {
	transferMap := make(map[string]*TransferInfo)

	for _, transfer := range transfers {
		key := transfer.FromStopID + "->" + transfer.ToStopID

		if existingTransfer, exists := transferMap[key]; exists {
			container.AddNotice(notice.NewDuplicateTransferNotice(
				transfer.FromStopID,
				transfer.ToStopID,
				transfer.RowNumber,
				existingTransfer.RowNumber,
			))
		} else {
			transferMap[key] = transfer
		}
	}
}
