package notice

// Notices defined by the Canonical GTFS Schedule Validator that this repo
// added while closing the gap against
// <https://gtfs-validator.mobilitydata.org/rules.html>. Codes and severities
// match the published rules; the checks are our own implementations.

// UnsortedStopTimesNotice reports stop_times.txt entries for a trip that are
// not in stop_sequence order, or not contiguous in the file. Consumers that
// stream the file rely on both.
type UnsortedStopTimesNotice struct {
	*BaseNotice
}

func NewUnsortedStopTimesNotice(tripID string, rowNumber int, stopSequence int, prevStopSequence int) *UnsortedStopTimesNotice {
	context := map[string]interface{}{
		"tripId":           tripID,
		"csvRowNumber":     rowNumber,
		"stopSequence":     stopSequence,
		"prevStopSequence": prevStopSequence,
	}
	return &UnsortedStopTimesNotice{
		BaseNotice: NewBaseNotice("unsorted_stop_times", INFO, context),
	}
}

// StopTimeWithOnlyArrivalOrDepartureTimeNotice reports a stop time that gives
// one of arrival_time and departure_time but not the other.
type StopTimeWithOnlyArrivalOrDepartureTimeNotice struct {
	*BaseNotice
}

func NewStopTimeWithOnlyArrivalOrDepartureTimeNotice(tripID string, rowNumber int, stopSequence int, specifiedField string) *StopTimeWithOnlyArrivalOrDepartureTimeNotice {
	context := map[string]interface{}{
		"tripId":         tripID,
		"csvRowNumber":   rowNumber,
		"stopSequence":   stopSequence,
		"specifiedField": specifiedField,
	}
	return &StopTimeWithOnlyArrivalOrDepartureTimeNotice{
		BaseNotice: NewBaseNotice("stop_time_with_only_arrival_or_departure_time", ERROR, context),
	}
}

// StopTimeTimepointWithoutTimesNotice reports timepoint=1 without both times.
// A timepoint asserts the vehicle keeps to the published time there, which
// requires a time to keep to.
type StopTimeTimepointWithoutTimesNotice struct {
	*BaseNotice
}

func NewStopTimeTimepointWithoutTimesNotice(tripID string, rowNumber int, stopSequence int, missingField string) *StopTimeTimepointWithoutTimesNotice {
	context := map[string]interface{}{
		"tripId":       tripID,
		"csvRowNumber": rowNumber,
		"stopSequence": stopSequence,
		"fieldName":    missingField,
	}
	return &StopTimeTimepointWithoutTimesNotice{
		BaseNotice: NewBaseNotice("stop_time_timepoint_without_times", ERROR, context),
	}
}

// MissingTimepointValueNotice reports a stop time that gives a time but leaves
// timepoint empty, so consumers cannot tell whether the time is exact.
type MissingTimepointValueNotice struct {
	*BaseNotice
}

func NewMissingTimepointValueNotice(tripID string, rowNumber int, stopSequence int) *MissingTimepointValueNotice {
	context := map[string]interface{}{
		"tripId":       tripID,
		"csvRowNumber": rowNumber,
		"stopSequence": stopSequence,
	}
	return &MissingTimepointValueNotice{
		BaseNotice: NewBaseNotice("missing_timepoint_value", WARNING, context),
	}
}

// ForbiddenArrivalOrDepartureTimeNotice reports a stop time carrying both an
// exact time and a pickup/drop-off window. The two describe the service in
// mutually exclusive ways.
type ForbiddenArrivalOrDepartureTimeNotice struct {
	*BaseNotice
}

func NewForbiddenArrivalOrDepartureTimeNotice(tripID string, rowNumber int, stopSequence int) *ForbiddenArrivalOrDepartureTimeNotice {
	context := map[string]interface{}{
		"tripId":       tripID,
		"csvRowNumber": rowNumber,
		"stopSequence": stopSequence,
	}
	return &ForbiddenArrivalOrDepartureTimeNotice{
		BaseNotice: NewBaseNotice("forbidden_arrival_or_departure_time", ERROR, context),
	}
}

// ForbiddenShapeDistTraveledNotice reports shape_dist_traveled on a stop time
// with no stop_id. Distance along a shape is only meaningful for a stop.
type ForbiddenShapeDistTraveledNotice struct {
	*BaseNotice
}

func NewForbiddenShapeDistTraveledNotice(tripID string, rowNumber int, stopSequence int, shapeDistTraveled string) *ForbiddenShapeDistTraveledNotice {
	context := map[string]interface{}{
		"tripId":            tripID,
		"csvRowNumber":      rowNumber,
		"stopSequence":      stopSequence,
		"shapeDistTraveled": shapeDistTraveled,
	}
	return &ForbiddenShapeDistTraveledNotice{
		BaseNotice: NewBaseNotice("forbidden_shape_dist_traveled", ERROR, context),
	}
}

// ForbiddenPickupTypeNotice reports a pickup/drop-off window on a stop time
// whose pickup_type is regularly scheduled (0) or coordinated with the driver
// (3) — neither of which admits a window.
type ForbiddenPickupTypeNotice struct {
	*BaseNotice
}

func NewForbiddenPickupTypeNotice(tripID string, rowNumber int, stopSequence int, pickupType int) *ForbiddenPickupTypeNotice {
	context := map[string]interface{}{
		"tripId":       tripID,
		"csvRowNumber": rowNumber,
		"stopSequence": stopSequence,
		"pickupType":   pickupType,
	}
	return &ForbiddenPickupTypeNotice{
		BaseNotice: NewBaseNotice("forbidden_pickup_type", ERROR, context),
	}
}

// ForbiddenDropOffTypeNotice is ForbiddenPickupTypeNotice for drop_off_type,
// where only regularly scheduled (0) is forbidden.
type ForbiddenDropOffTypeNotice struct {
	*BaseNotice
}

func NewForbiddenDropOffTypeNotice(tripID string, rowNumber int, stopSequence int, dropOffType int) *ForbiddenDropOffTypeNotice {
	context := map[string]interface{}{
		"tripId":       tripID,
		"csvRowNumber": rowNumber,
		"stopSequence": stopSequence,
		"dropOffType":  dropOffType,
	}
	return &ForbiddenDropOffTypeNotice{
		BaseNotice: NewBaseNotice("forbidden_drop_off_type", ERROR, context),
	}
}

// ForbiddenContinuousPickupDropOffNotice reports a pickup/drop-off window on a
// trip whose route already declares a continuous pickup or drop-off value that
// forbids one.
type ForbiddenContinuousPickupDropOffNotice struct {
	*BaseNotice
}

func NewForbiddenContinuousPickupDropOffNotice(tripID string, routeID string, rowNumber int, stopSequence int, fieldName string, fieldValue int) *ForbiddenContinuousPickupDropOffNotice {
	context := map[string]interface{}{
		"tripId":       tripID,
		"routeId":      routeID,
		"csvRowNumber": rowNumber,
		"stopSequence": stopSequence,
		"fieldName":    fieldName,
		"fieldValue":   fieldValue,
	}
	return &ForbiddenContinuousPickupDropOffNotice{
		BaseNotice: NewBaseNotice("forbidden_continuous_pickup_drop_off", ERROR, context),
	}
}

// MissingStopTimesRecordNotice reports travel to a location group or GeoJSON
// location described by a single stop_times row. Two are required: one to
// enter the area and one to leave it.
type MissingStopTimesRecordNotice struct {
	*BaseNotice
}

func NewMissingStopTimesRecordNotice(tripID string, rowNumber int, fieldName string, fieldValue string) *MissingStopTimesRecordNotice {
	context := map[string]interface{}{
		"tripId":       tripID,
		"csvRowNumber": rowNumber,
		"fieldName":    fieldName,
		"fieldValue":   fieldValue,
	}
	return &MissingStopTimesRecordNotice{
		BaseNotice: NewBaseNotice("missing_stop_times_record", ERROR, context),
	}
}

// StopWithoutStopTimeNotice reports a stop no trip ever calls at. Often a typo
// in stop_times.txt rather than a deliberately unused stop.
type StopWithoutStopTimeNotice struct {
	*BaseNotice
}

func NewStopWithoutStopTimeNotice(stopID string, stopName string, rowNumber int) *StopWithoutStopTimeNotice {
	context := map[string]interface{}{
		"stopId":       stopID,
		"stopName":     stopName,
		"csvRowNumber": rowNumber,
	}
	return &StopWithoutStopTimeNotice{
		BaseNotice: NewBaseNotice("stop_without_stop_time", WARNING, context),
	}
}

// UnusedTripNotice reports a trip with no stop_times rows. It calls nowhere,
// so no consumer can route with it.
type UnusedTripNotice struct {
	*BaseNotice
}

func NewUnusedTripNotice(tripID string, rowNumber int) *UnusedTripNotice {
	context := map[string]interface{}{
		"tripId":       tripID,
		"csvRowNumber": rowNumber,
	}
	return &UnusedTripNotice{
		BaseNotice: NewBaseNotice("unused_trip", WARNING, context),
	}
}

// UnusedStationNotice reports a station with no child locations.
type UnusedStationNotice struct {
	*BaseNotice
}

func NewUnusedStationNotice(stopID string, stopName string, rowNumber int) *UnusedStationNotice {
	context := map[string]interface{}{
		"stopId":       stopID,
		"stopName":     stopName,
		"csvRowNumber": rowNumber,
	}
	return &UnusedStationNotice{
		BaseNotice: NewBaseNotice("unused_station", INFO, context),
	}
}

// MissingRequiredAgencyIDNotice reports an omitted agency_id in a feed with
// more than one agency, where the reference would be ambiguous.
type MissingRequiredAgencyIDNotice struct {
	*BaseNotice
}

func NewMissingRequiredAgencyIDNotice(filename string, fieldName string, rowNumber int) *MissingRequiredAgencyIDNotice {
	context := map[string]interface{}{
		"filename":     filename,
		"fieldName":    fieldName,
		"csvRowNumber": rowNumber,
	}
	return &MissingRequiredAgencyIDNotice{
		BaseNotice: NewBaseNotice("missing_required_agency_id", ERROR, context),
	}
}

// InconsistentAgencyTimezoneNotice reports agencies in one feed declaring
// different timezones. Every time in the feed is read in the agency timezone,
// so two of them make the schedule ambiguous.
type InconsistentAgencyTimezoneNotice struct {
	*BaseNotice
}

func NewInconsistentAgencyTimezoneNotice(expected string, actual string, rowNumber int) *InconsistentAgencyTimezoneNotice {
	context := map[string]interface{}{
		"expected":     expected,
		"actual":       actual,
		"csvRowNumber": rowNumber,
	}
	return &InconsistentAgencyTimezoneNotice{
		BaseNotice: NewBaseNotice("inconsistent_agency_timezone", ERROR, context),
	}
}

// WrongParentLocationTypeNotice reports a parent_station pointing at a
// location whose location_type cannot parent the child's.
type WrongParentLocationTypeNotice struct {
	*BaseNotice
}

func NewWrongParentLocationTypeNotice(stopID string, rowNumber int, locationType int, parentStation string, parentRowNumber int, parentLocationType int, expectedLocationType int) *WrongParentLocationTypeNotice {
	context := map[string]interface{}{
		"stopId":               stopID,
		"csvRowNumber":         rowNumber,
		"locationType":         locationType,
		"parentStation":        parentStation,
		"parentCsvRowNumber":   parentRowNumber,
		"parentLocationType":   parentLocationType,
		"expectedLocationType": expectedLocationType,
	}
	return &WrongParentLocationTypeNotice{
		BaseNotice: NewBaseNotice("wrong_parent_location_type", ERROR, context),
	}
}

// LocationWithoutParentStationNotice reports an entrance, generic node or
// boarding area with no parent_station. Those types only have meaning inside a
// station.
type LocationWithoutParentStationNotice struct {
	*BaseNotice
}

func NewLocationWithoutParentStationNotice(stopID string, rowNumber int, locationType int) *LocationWithoutParentStationNotice {
	context := map[string]interface{}{
		"stopId":       stopID,
		"csvRowNumber": rowNumber,
		"locationType": locationType,
	}
	return &LocationWithoutParentStationNotice{
		BaseNotice: NewBaseNotice("location_without_parent_station", ERROR, context),
	}
}

// StopWithoutLocationNotice reports a stop, station or entrance without
// coordinates, which cannot be placed on a map or matched to a shape.
type StopWithoutLocationNotice struct {
	*BaseNotice
}

func NewStopWithoutLocationNotice(stopID string, rowNumber int, locationType int) *StopWithoutLocationNotice {
	context := map[string]interface{}{
		"stopId":       stopID,
		"csvRowNumber": rowNumber,
		"locationType": locationType,
	}
	return &StopWithoutLocationNotice{
		BaseNotice: NewBaseNotice("stop_without_location", ERROR, context),
	}
}

// TransferWithInvalidStopLocationTypeNotice reports a transfer referencing a
// location that is neither a stop/platform nor a station.
type TransferWithInvalidStopLocationTypeNotice struct {
	*BaseNotice
}

func NewTransferWithInvalidStopLocationTypeNotice(rowNumber int, fieldName string, stopID string, locationType int) *TransferWithInvalidStopLocationTypeNotice {
	context := map[string]interface{}{
		"csvRowNumber": rowNumber,
		"stopIdField":  fieldName,
		"stopId":       stopID,
		"locationType": locationType,
	}
	return &TransferWithInvalidStopLocationTypeNotice{
		BaseNotice: NewBaseNotice("transfer_with_invalid_stop_location_type", ERROR, context),
	}
}

// TransferWithInvalidTripAndRouteNotice reports a transfer whose trip belongs
// to a different route than the one it names.
type TransferWithInvalidTripAndRouteNotice struct {
	*BaseNotice
}

func NewTransferWithInvalidTripAndRouteNotice(rowNumber int, tripFieldName string, tripID string, routeFieldName string, routeID string, expectedRouteID string) *TransferWithInvalidTripAndRouteNotice {
	context := map[string]interface{}{
		"csvRowNumber":     rowNumber,
		"tripIdFieldName":  tripFieldName,
		"tripId":           tripID,
		"routeIdFieldName": routeFieldName,
		"routeId":          routeID,
		"expectedRouteId":  expectedRouteID,
	}
	return &TransferWithInvalidTripAndRouteNotice{
		BaseNotice: NewBaseNotice("transfer_with_invalid_trip_and_route", ERROR, context),
	}
}

// TransferWithInvalidTripAndStopNotice reports a transfer naming a stop the
// referenced trip never calls at.
type TransferWithInvalidTripAndStopNotice struct {
	*BaseNotice
}

func NewTransferWithInvalidTripAndStopNotice(rowNumber int, tripFieldName string, tripID string, stopFieldName string, stopID string) *TransferWithInvalidTripAndStopNotice {
	context := map[string]interface{}{
		"csvRowNumber":    rowNumber,
		"tripIdFieldName": tripFieldName,
		"tripId":          tripID,
		"stopIdFieldName": stopFieldName,
		"stopId":          stopID,
	}
	return &TransferWithInvalidTripAndStopNotice{
		BaseNotice: NewBaseNotice("transfer_with_invalid_trip_and_stop", ERROR, context),
	}
}

// TripCoverageNotActiveForNext7DaysNotice reports a feed whose trips stop
// running within the coming week.
type TripCoverageNotActiveForNext7DaysNotice struct {
	*BaseNotice
}

func NewTripCoverageNotActiveForNext7DaysNotice(currentDate string, lastServiceDate string) *TripCoverageNotActiveForNext7DaysNotice {
	context := map[string]interface{}{
		"currentDate":     currentDate,
		"lastServiceDate": lastServiceDate,
	}
	return &TripCoverageNotActiveForNext7DaysNotice{
		BaseNotice: NewBaseNotice("trip_coverage_not_active_for_next7_days", WARNING, context),
	}
}

// PointNearOriginNotice reports a coordinate at (0, 0), in the Gulf of Guinea.
// A stop there is almost always a field left empty and defaulted to zero.
type PointNearOriginNotice struct {
	*BaseNotice
}

func NewPointNearOriginNotice(filename string, fieldName string, fieldValue string, rowNumber int) *PointNearOriginNotice {
	context := map[string]interface{}{
		"filename":     filename,
		"fieldName":    fieldName,
		"fieldValue":   fieldValue,
		"csvRowNumber": rowNumber,
	}
	return &PointNearOriginNotice{
		BaseNotice: NewBaseNotice("point_near_origin", ERROR, context),
	}
}

// PointNearPoleNotice reports a coordinate at one of the poles, the other
// position a missing value tends to become.
type PointNearPoleNotice struct {
	*BaseNotice
}

func NewPointNearPoleNotice(filename string, fieldName string, fieldValue string, rowNumber int) *PointNearPoleNotice {
	context := map[string]interface{}{
		"filename":     filename,
		"fieldName":    fieldName,
		"fieldValue":   fieldValue,
		"csvRowNumber": rowNumber,
	}
	return &PointNearPoleNotice{
		BaseNotice: NewBaseNotice("point_near_pole", ERROR, context),
	}
}
