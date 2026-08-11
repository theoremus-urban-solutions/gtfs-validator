package relationship

import (
	"io"
	"log"
	"strconv"
	"strings"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/parser"
	"github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

// StopTimeFieldValidator makes a single streaming pass over stop_times.txt and
// reports the per-row rules that depend on the file's own order.
//
// The optional fields are kept as written rather than parsed into their target
// types, because several of these rules turn on the difference between a field
// given as zero and a field not given at all — missing_timepoint_value is
// exactly that distinction, and an int cannot hold it.
type StopTimeFieldValidator struct{}

// NewStopTimeFieldValidator creates a new stop time field validator
func NewStopTimeFieldValidator() *StopTimeFieldValidator {
	return &StopTimeFieldValidator{}
}

// stopTimeRow is one row of stop_times.txt, with the optional fields kept as
// written so that "absent" stays distinguishable from "zero".
type stopTimeRow struct {
	TripID            string
	StopID            string
	StopSequence      int
	ArrivalTime       string
	DepartureTime     string
	Timepoint         string
	PickupType        string
	DropOffType       string
	ShapeDistTraveled string
	StartPickupWindow string
	EndPickupWindow   string
	LocationGroupID   string
	LocationID        string
	RowNumber         int
}

func (r *stopTimeRow) hasPickupDropOffWindow() bool {
	return r.StartPickupWindow != "" || r.EndPickupWindow != ""
}

// tripBlock is what one trip's rows look like so far: the span they occupy in
// the file, the sequence the last of them carried, and whether anything about
// their order has already broken the rule.
type tripBlock struct {
	StartRow     int
	EndRow       int
	LastSequence int
	Unsorted     bool
}

// Validate checks the per-row stop time rules.
func (v *StopTimeFieldValidator) Validate(loader *parser.FeedLoader, container *notice.NoticeContainer, config validator.Config) {
	routeContinuity := v.loadRouteContinuity(loader)

	reader, err := loader.GetFile("stop_times.txt")
	if err != nil {
		return
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close reader %v", closeErr)
		}
	}()

	csvFile, err := parser.NewCSVFile(reader, "stop_times.txt")
	if err != nil {
		return
	}

	hasTimepointColumn := false
	for _, header := range csvFile.Headers {
		if strings.TrimSpace(header) == "timepoint" {
			hasTimepointColumn = true
			break
		}
	}

	// unsorted_stop_times is about the file as written, so order is tracked as
	// rows arrive rather than after sorting. It is reported once per trip and
	// names the trip's whole span in the file, which is only known once its
	// last row has gone past — hence the per-trip state and the report at EOF.
	blocks := make(map[string]*tripBlock)
	var blockOrder []string
	var previousTripID string

	// A location group or GeoJSON location needs two rows on the same trip:
	// one to enter the area and one to leave it. Count per trip and area.
	areaVisits := make(map[string]*stopTimeRow)
	areaCounts := make(map[string]int)

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		stopTime := v.parseRow(row)
		if stopTime == nil {
			continue
		}

		v.validateTimes(container, stopTime, hasTimepointColumn)
		v.validateWindows(container, stopTime, routeContinuity)
		v.validateShapeDistTraveled(container, stopTime)

		block, seen := blocks[stopTime.TripID]
		if !seen {
			block = &tripBlock{
				StartRow:     stopTime.RowNumber,
				EndRow:       stopTime.RowNumber,
				LastSequence: stopTime.StopSequence,
			}
			blocks[stopTime.TripID] = block
			blockOrder = append(blockOrder, stopTime.TripID)
		} else {
			// Either the sequence went backwards, or the trip's rows resumed
			// after another trip's — the spec requires a trip's rows to be
			// both ordered and contiguous.
			if stopTime.TripID != previousTripID || stopTime.StopSequence <= block.LastSequence {
				block.Unsorted = true
			}
			block.EndRow = stopTime.RowNumber
			block.LastSequence = stopTime.StopSequence
		}
		previousTripID = stopTime.TripID

		for field, value := range map[string]string{
			"location_group_id": stopTime.LocationGroupID,
			"location_id":       stopTime.LocationID,
		} {
			if value == "" {
				continue
			}
			key := stopTime.TripID + "\x00" + field + "\x00" + value
			areaCounts[key]++
			if _, seen := areaVisits[key]; !seen {
				areaVisits[key] = stopTime
			}
		}
	}

	for _, tripID := range blockOrder {
		if block := blocks[tripID]; block.Unsorted {
			container.AddNotice(notice.NewUnsortedStopTimesNotice(
				tripID,
				block.StartRow,
				block.EndRow,
			))
		}
	}

	for key, count := range areaCounts {
		if count >= 2 {
			continue
		}
		first := areaVisits[key]
		parts := strings.Split(key, "\x00")
		container.AddNotice(notice.NewMissingStopTimesRecordNotice(
			first.TripID,
			first.RowNumber,
			parts[1],
			parts[2],
		))
	}
}

// validateTimes checks arrival_time, departure_time and timepoint against each
// other.
func (v *StopTimeFieldValidator) validateTimes(container *notice.NoticeContainer, row *stopTimeRow, hasTimepointColumn bool) {
	hasArrival := row.ArrivalTime != ""
	hasDeparture := row.DepartureTime != ""

	// Times and a pickup/drop-off window describe the service in mutually
	// exclusive ways, so a row carrying both is meaningless.
	if (hasArrival || hasDeparture) && row.hasPickupDropOffWindow() {
		container.AddNotice(notice.NewForbiddenArrivalOrDepartureTimeNotice(
			row.TripID, row.RowNumber, row.StopSequence,
		))
		return
	}

	switch {
	case hasArrival && !hasDeparture:
		container.AddNotice(notice.NewStopTimeWithOnlyArrivalOrDepartureTimeNotice(
			row.TripID, row.RowNumber, row.StopSequence, "arrival_time",
		))
	case !hasArrival && hasDeparture:
		container.AddNotice(notice.NewStopTimeWithOnlyArrivalOrDepartureTimeNotice(
			row.TripID, row.RowNumber, row.StopSequence, "departure_time",
		))
	}

	if row.Timepoint == "" {
		// Only a blank value in a column that exists is a missing timepoint. A
		// file with no timepoint column at all has not omitted a value — the
		// field is optional, and canonical says nothing about it. Reading the
		// field out of a map made the two indistinguishable, so every feed
		// without the column was reported once per timed row.
		if hasTimepointColumn && (hasArrival || hasDeparture) {
			container.AddNotice(notice.NewMissingTimepointValueNotice(
				row.TripID, row.RowNumber, row.StopSequence,
			))
		}
		return
	}

	// A timepoint asserts the vehicle keeps to the published time here, which
	// requires both times to be published.
	if strings.TrimSpace(row.Timepoint) != "1" {
		return
	}
	if !hasArrival {
		container.AddNotice(notice.NewStopTimeTimepointWithoutTimesNotice(
			row.TripID, row.RowNumber, row.StopSequence, "arrival_time",
		))
	}
	if !hasDeparture {
		container.AddNotice(notice.NewStopTimeTimepointWithoutTimesNotice(
			row.TripID, row.RowNumber, row.StopSequence, "departure_time",
		))
	}
}

// validateWindows checks the pickup/drop-off window against the pickup and
// drop-off types on the row and the continuous values on the route.
func (v *StopTimeFieldValidator) validateWindows(container *notice.NoticeContainer, row *stopTimeRow, routeContinuity map[string]*routeContinuousService) {
	if !row.hasPickupDropOffWindow() {
		return
	}

	// Regularly scheduled (0) and coordinate-with-the-driver (3) pickups both
	// name a time, so neither admits a window.
	if pickupType, err := strconv.Atoi(strings.TrimSpace(row.PickupType)); err == nil {
		if pickupType == 0 || pickupType == 3 {
			container.AddNotice(notice.NewForbiddenPickupTypeNotice(
				row.TripID, row.RowNumber, row.StopSequence, pickupType,
			))
		}
	}
	if dropOffType, err := strconv.Atoi(strings.TrimSpace(row.DropOffType)); err == nil {
		if dropOffType == 0 {
			container.AddNotice(notice.NewForbiddenDropOffTypeNotice(
				row.TripID, row.RowNumber, row.StopSequence, dropOffType,
			))
		}
	}

	route, exists := routeContinuity[row.TripID]
	if !exists {
		return
	}
	for _, field := range []struct {
		name  string
		value *int
	}{
		{"continuous_pickup", route.ContinuousPickup},
		{"continuous_drop_off", route.ContinuousDropOff},
	} {
		if field.value == nil {
			continue
		}
		switch *field.value {
		case 0, 2, 3:
			container.AddNotice(notice.NewForbiddenContinuousPickupDropOffNotice(
				row.TripID, route.RouteID, row.RowNumber, row.StopSequence,
				field.name, *field.value,
			))
		}
	}
}

// validateShapeDistTraveled reports a distance along the shape given for
// something that is not a stop.
func (v *StopTimeFieldValidator) validateShapeDistTraveled(container *notice.NoticeContainer, row *stopTimeRow) {
	if row.ShapeDistTraveled == "" || row.StopID != "" {
		return
	}
	container.AddNotice(notice.NewForbiddenShapeDistTraveledNotice(
		row.TripID, row.RowNumber, row.StopSequence, row.ShapeDistTraveled,
	))
}

// routeContinuousService holds the continuous pickup and drop-off declared by
// the route a trip belongs to.
type routeContinuousService struct {
	RouteID           string
	ContinuousPickup  *int
	ContinuousDropOff *int
}

// loadRouteContinuity maps each trip to its route's continuous pickup and
// drop-off values, or returns an empty map when neither field is used.
func (v *StopTimeFieldValidator) loadRouteContinuity(loader *parser.FeedLoader) map[string]*routeContinuousService {
	byTrip := make(map[string]*routeContinuousService)

	routes := make(map[string]*routeContinuousService)
	if reader, err := loader.GetFile("routes.txt"); err == nil {
		defer func() {
			if closeErr := reader.Close(); closeErr != nil {
				log.Printf("Warning: failed to close reader %v", closeErr)
			}
		}()
		if csvFile, err := parser.NewCSVFile(reader, "routes.txt"); err == nil {
			for {
				row, err := csvFile.ReadRow()
				if err == io.EOF {
					break
				}
				if err != nil {
					continue
				}
				routeID := strings.TrimSpace(row.Values["route_id"])
				if routeID == "" {
					continue
				}
				route := &routeContinuousService{RouteID: routeID}
				route.ContinuousPickup = optionalInt(row.Values["continuous_pickup"])
				route.ContinuousDropOff = optionalInt(row.Values["continuous_drop_off"])
				if route.ContinuousPickup != nil || route.ContinuousDropOff != nil {
					routes[routeID] = route
				}
			}
		}
	}

	if len(routes) == 0 {
		return byTrip
	}

	reader, err := loader.GetFile("trips.txt")
	if err != nil {
		return byTrip
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close reader %v", closeErr)
		}
	}()

	csvFile, err := parser.NewCSVFile(reader, "trips.txt")
	if err != nil {
		return byTrip
	}
	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}
		tripID := strings.TrimSpace(row.Values["trip_id"])
		if route, exists := routes[strings.TrimSpace(row.Values["route_id"])]; exists && tripID != "" {
			byTrip[tripID] = route
		}
	}

	return byTrip
}

// optionalInt returns the parsed value, or nil when the field is absent,
// blank, or not a number — the last of which the type layer reports.
func optionalInt(value string) *int {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	parsed, err := strconv.Atoi(trimmed)
	if err != nil {
		return nil
	}
	return &parsed
}

// parseRow reads one stop_times.txt row, returning nil for rows without the
// identifying fields the other validators report on.
func (v *StopTimeFieldValidator) parseRow(row *parser.CSVRow) *stopTimeRow {
	tripID, hasTripID := row.Values["trip_id"]
	stopSequenceStr, hasStopSequence := row.Values["stop_sequence"]
	if !hasTripID || !hasStopSequence {
		return nil
	}

	stopSequence, err := strconv.Atoi(strings.TrimSpace(stopSequenceStr))
	if err != nil {
		return nil
	}

	field := func(name string) string {
		return strings.TrimSpace(row.Values[name])
	}

	return &stopTimeRow{
		TripID:            strings.TrimSpace(tripID),
		StopID:            field("stop_id"),
		StopSequence:      stopSequence,
		ArrivalTime:       field("arrival_time"),
		DepartureTime:     field("departure_time"),
		Timepoint:         field("timepoint"),
		PickupType:        field("pickup_type"),
		DropOffType:       field("drop_off_type"),
		ShapeDistTraveled: field("shape_dist_traveled"),
		StartPickupWindow: field("start_pickup_drop_off_window"),
		EndPickupWindow:   field("end_pickup_drop_off_window"),
		LocationGroupID:   field("location_group_id"),
		LocationID:        field("location_id"),
		RowNumber:         row.RowNumber,
	}
}
