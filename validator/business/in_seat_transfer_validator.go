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

// InSeatTransferValidator checks the transfers that let a passenger stay in
// their seat while one trip becomes another (transfer_type 4 and 5). Because
// nobody moves, the two trips have to be the same vehicle continuing on: same
// mode, and joined end to end rather than in the middle of either.
type InSeatTransferValidator struct{}

// NewInSeatTransferValidator creates a new in-seat transfer validator
func NewInSeatTransferValidator() *InSeatTransferValidator {
	return &InSeatTransferValidator{}
}

// inSeatTransfer is one transfers.txt row of type 4 or 5.
type inSeatTransfer struct {
	FromStopID string
	ToStopID   string
	FromTripID string
	ToTripID   string
	RowNumber  int
}

// tripEndpoints is where a trip begins and ends, which is all the in-seat
// check needs from stop_times.txt.
type tripEndpoints struct {
	FirstStopID   string
	FirstSequence int
	LastStopID    string
	LastSequence  int
	Found         bool
}

// Validate checks in-seat transfers against the trips they join.
func (v *InSeatTransferValidator) Validate(loader *parser.FeedLoader, container *notice.NoticeContainer, config validator.Config) {
	transfers := v.loadInSeatTransfers(loader)
	if len(transfers) == 0 {
		return
	}

	// Only the trips these transfers name are of interest, so collect them
	// first and let the stop_times pass ignore everything else.
	wantedTrips := make(map[string]bool)
	for _, transfer := range transfers {
		if transfer.FromTripID != "" {
			wantedTrips[transfer.FromTripID] = true
		}
		if transfer.ToTripID != "" {
			wantedTrips[transfer.ToTripID] = true
		}
	}

	tripRoutes := v.loadTripRoutes(loader, wantedTrips)
	routeTypes := v.loadRouteTypes(loader)
	endpoints := v.loadTripEndpoints(loader, wantedTrips)
	parentStations := v.loadParentStations(loader)

	for _, transfer := range transfers {
		v.validateRouteTypes(container, transfer, tripRoutes, routeTypes)
		v.validateEndpoints(container, transfer, endpoints, parentStations)
	}
}

// validateRouteTypes reports an in-seat transfer between trips whose routes
// run different modes.
func (v *InSeatTransferValidator) validateRouteTypes(container *notice.NoticeContainer, transfer *inSeatTransfer, tripRoutes map[string]string, routeTypes map[string]int) {
	fromRoute, hasFrom := tripRoutes[transfer.FromTripID]
	toRoute, hasTo := tripRoutes[transfer.ToTripID]
	if !hasFrom || !hasTo {
		return
	}
	fromType, hasFromType := routeTypes[fromRoute]
	toType, hasToType := routeTypes[toRoute]
	if !hasFromType || !hasToType || fromType == toType {
		return
	}

	container.AddNotice(notice.NewInconsistentRouteTypeForInSeatTransferNotice(
		transfer.RowNumber,
		transfer.FromTripID,
		fromRoute,
		fromType,
		transfer.ToTripID,
		toRoute,
		toType,
	))
}

// validateEndpoints reports an in-seat transfer that joins a trip anywhere but
// at its end: the passenger must be leaving the arriving trip at its final
// stop and continuing from the departing trip's first.
func (v *InSeatTransferValidator) validateEndpoints(container *notice.NoticeContainer, transfer *inSeatTransfer, endpoints map[string]*tripEndpoints, parentStations map[string]string) {
	if from, exists := endpoints[transfer.FromTripID]; exists && from.Found && transfer.FromStopID != "" {
		if !v.sameLocation(from.LastStopID, transfer.FromStopID, parentStations) {
			container.AddNotice(notice.NewTransferWithSuspiciousMidTripInSeatNotice(
				transfer.RowNumber,
				transfer.FromTripID,
				transfer.FromStopID,
				"from_trip_id",
			))
		}
	}

	if to, exists := endpoints[transfer.ToTripID]; exists && to.Found && transfer.ToStopID != "" {
		if !v.sameLocation(to.FirstStopID, transfer.ToStopID, parentStations) {
			container.AddNotice(notice.NewTransferWithSuspiciousMidTripInSeatNotice(
				transfer.RowNumber,
				transfer.ToTripID,
				transfer.ToStopID,
				"to_trip_id",
			))
		}
	}
}

// sameLocation reports whether a transfer naming transferStopID refers to the
// stop the trip calls at. A transfer may name the station that contains the
// platform, which is the same place to a passenger staying in their seat.
func (v *InSeatTransferValidator) sameLocation(tripStopID string, transferStopID string, parentStations map[string]string) bool {
	return tripStopID == transferStopID || parentStations[tripStopID] == transferStopID
}

// loadInSeatTransfers returns the transfers.txt rows of type 4 and 5.
func (v *InSeatTransferValidator) loadInSeatTransfers(loader *parser.FeedLoader) []*inSeatTransfer {
	var transfers []*inSeatTransfer

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

		transferType, err := strconv.Atoi(strings.TrimSpace(row.Values["transfer_type"]))
		if err != nil || (transferType != 4 && transferType != 5) {
			continue
		}
		transfers = append(transfers, &inSeatTransfer{
			FromStopID: strings.TrimSpace(row.Values["from_stop_id"]),
			ToStopID:   strings.TrimSpace(row.Values["to_stop_id"]),
			FromTripID: strings.TrimSpace(row.Values["from_trip_id"]),
			ToTripID:   strings.TrimSpace(row.Values["to_trip_id"]),
			RowNumber:  row.RowNumber,
		})
	}

	return transfers
}

// loadTripRoutes maps each of the named trips to its route.
func (v *InSeatTransferValidator) loadTripRoutes(loader *parser.FeedLoader, wanted map[string]bool) map[string]string {
	tripRoutes := make(map[string]string, len(wanted))

	reader, err := loader.GetFile("trips.txt")
	if err != nil {
		return tripRoutes
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close reader %v", closeErr)
		}
	}()

	csvFile, err := parser.NewCSVFile(reader, "trips.txt")
	if err != nil {
		return tripRoutes
	}

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
		tripRoutes[tripID] = strings.TrimSpace(row.Values["route_id"])
	}

	return tripRoutes
}

// loadRouteTypes maps each route to its mode.
func (v *InSeatTransferValidator) loadRouteTypes(loader *parser.FeedLoader) map[string]int {
	routeTypes := make(map[string]int)

	reader, err := loader.GetFile("routes.txt")
	if err != nil {
		return routeTypes
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close reader %v", closeErr)
		}
	}()

	csvFile, err := parser.NewCSVFile(reader, "routes.txt")
	if err != nil {
		return routeTypes
	}

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}

		routeID := strings.TrimSpace(row.Values["route_id"])
		routeType, err := strconv.Atoi(strings.TrimSpace(row.Values["route_type"]))
		if routeID == "" || err != nil {
			continue
		}
		routeTypes[routeID] = routeType
	}

	return routeTypes
}

// loadTripEndpoints makes one pass over stop_times.txt, keeping the first and
// last stop of each named trip. Rows for other trips cost a map lookup, and
// the file is never held in memory.
func (v *InSeatTransferValidator) loadTripEndpoints(loader *parser.FeedLoader, wanted map[string]bool) map[string]*tripEndpoints {
	endpoints := make(map[string]*tripEndpoints, len(wanted))

	reader, err := loader.GetFile("stop_times.txt")
	if err != nil {
		return endpoints
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close reader %v", closeErr)
		}
	}()

	csvFile, err := parser.NewCSVFile(reader, "stop_times.txt")
	if err != nil {
		return endpoints
	}

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
		sequence, err := strconv.Atoi(strings.TrimSpace(row.Values["stop_sequence"]))
		if err != nil {
			continue
		}
		stopID := strings.TrimSpace(row.Values["stop_id"])

		trip, exists := endpoints[tripID]
		if !exists {
			endpoints[tripID] = &tripEndpoints{
				FirstStopID:   stopID,
				FirstSequence: sequence,
				LastStopID:    stopID,
				LastSequence:  sequence,
				Found:         true,
			}
			continue
		}
		if sequence < trip.FirstSequence {
			trip.FirstStopID = stopID
			trip.FirstSequence = sequence
		}
		if sequence > trip.LastSequence {
			trip.LastStopID = stopID
			trip.LastSequence = sequence
		}
	}

	return endpoints
}

// loadParentStations maps each stop to the station containing it.
func (v *InSeatTransferValidator) loadParentStations(loader *parser.FeedLoader) map[string]string {
	parents := make(map[string]string)

	reader, err := loader.GetFile("stops.txt")
	if err != nil {
		return parents
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close reader %v", closeErr)
		}
	}()

	csvFile, err := parser.NewCSVFile(reader, "stops.txt")
	if err != nil {
		return parents
	}

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}

		stopID := strings.TrimSpace(row.Values["stop_id"])
		parent := strings.TrimSpace(row.Values["parent_station"])
		if stopID == "" || parent == "" {
			continue
		}
		parents[stopID] = parent
	}

	return parents
}
