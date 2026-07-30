package notice

// Common validation notices that can occur during GTFS validation

// DuplicateKeyNotice is generated when a duplicate primary key is found
type DuplicateKeyNotice struct {
	*BaseNotice
}

func NewDuplicateKeyNotice(filename string, fieldName string, fieldValue interface{}, rowNumber int, duplicateRow int) *DuplicateKeyNotice {
	context := map[string]interface{}{
		"filename":     filename,
		"fieldName":    fieldName,
		"fieldValue":   fieldValue,
		"csvRowNumber": rowNumber,
		"duplicateRow": duplicateRow,
	}
	return &DuplicateKeyNotice{
		BaseNotice: NewBaseNotice("duplicate_key", ERROR, context),
	}
}

// MissingRequiredFieldNotice is generated when a required field is missing
type MissingRequiredFieldNotice struct {
	*BaseNotice
}

func NewMissingRequiredFieldNotice(filename string, fieldName string, rowNumber int) *MissingRequiredFieldNotice {
	context := map[string]interface{}{
		"filename":     filename,
		"fieldName":    fieldName,
		"csvRowNumber": rowNumber,
	}
	return &MissingRequiredFieldNotice{
		BaseNotice: NewBaseNotice("missing_required_field", ERROR, context),
	}
}

// InvalidFieldFormatNotice is generated when a field has invalid format
type InvalidFieldFormatNotice struct {
	*BaseNotice
}

func NewInvalidFieldFormatNotice(filename string, fieldName string, fieldValue string, rowNumber int, expectedFormat string) *InvalidFieldFormatNotice {
	context := map[string]interface{}{
		"filename":       filename,
		"fieldName":      fieldName,
		"fieldValue":     fieldValue,
		"csvRowNumber":   rowNumber,
		"expectedFormat": expectedFormat,
	}
	return &InvalidFieldFormatNotice{
		BaseNotice: NewBaseNotice("invalid_field_format", WARNING, context),
	}
}

// ForeignKeyViolationNotice is generated when a foreign key reference is invalid
type ForeignKeyViolationNotice struct {
	*BaseNotice
}

func NewForeignKeyViolationNotice(filename string, fieldName string, fieldValue string, rowNumber int, referencedTable string, referencedField string) *ForeignKeyViolationNotice {
	context := map[string]interface{}{
		"filename":        filename,
		"fieldName":       fieldName,
		"fieldValue":      fieldValue,
		"csvRowNumber":    rowNumber,
		"referencedTable": referencedTable,
		"referencedField": referencedField,
	}
	return &ForeignKeyViolationNotice{
		BaseNotice: NewBaseNotice("foreign_key_violation", ERROR, context),
	}
}

// EmptyFileNotice is generated when a file is empty
type EmptyFileNotice struct {
	*BaseNotice
}

func NewEmptyFileNotice(filename string) *EmptyFileNotice {
	context := map[string]interface{}{
		"filename": filename,
	}
	return &EmptyFileNotice{
		BaseNotice: NewBaseNotice("empty_file", ERROR, context),
	}
}

// UnknownColumnNotice is generated when an unknown column is found
type UnknownColumnNotice struct {
	*BaseNotice
}

func NewUnknownColumnNotice(filename string, columnName string, columnIndex int) *UnknownColumnNotice {
	context := map[string]interface{}{
		"filename":    filename,
		"columnName":  columnName,
		"columnIndex": columnIndex,
	}
	return &UnknownColumnNotice{
		BaseNotice: NewBaseNotice("unknown_column", INFO, context),
	}
}

// InvalidURLNotice is generated when a URL field has invalid format
type InvalidURLNotice struct {
	*BaseNotice
}

func NewInvalidURLNotice(filename string, fieldName string, fieldValue string, rowNumber int) *InvalidURLNotice {
	context := map[string]interface{}{
		"filename":     filename,
		"fieldName":    fieldName,
		"fieldValue":   fieldValue,
		"csvRowNumber": rowNumber,
	}
	return &InvalidURLNotice{
		BaseNotice: NewBaseNotice("invalid_url", ERROR, context),
	}
}

// InvalidEmailNotice is generated when an email field has invalid format
type InvalidEmailNotice struct {
	*BaseNotice
}

func NewInvalidEmailNotice(filename string, fieldName string, fieldValue string, rowNumber int) *InvalidEmailNotice {
	context := map[string]interface{}{
		"filename":     filename,
		"fieldName":    fieldName,
		"fieldValue":   fieldValue,
		"csvRowNumber": rowNumber,
	}
	return &InvalidEmailNotice{
		BaseNotice: NewBaseNotice("invalid_email", ERROR, context),
	}
}

// InvalidTimezoneNotice is generated when a timezone is invalid
type InvalidTimezoneNotice struct {
	*BaseNotice
}

func NewInvalidTimezoneNotice(filename string, fieldName string, fieldValue string, rowNumber int) *InvalidTimezoneNotice {
	context := map[string]interface{}{
		"filename":     filename,
		"fieldName":    fieldName,
		"fieldValue":   fieldValue,
		"csvRowNumber": rowNumber,
	}
	return &InvalidTimezoneNotice{
		BaseNotice: NewBaseNotice("invalid_timezone", ERROR, context),
	}
}

// MissingRequiredFileNotice is generated when a required file is missing
type MissingRequiredFileNotice struct {
	*BaseNotice
}

func NewMissingRequiredFileNotice(filename string) *MissingRequiredFileNotice {
	context := map[string]interface{}{
		"filename": filename,
	}
	return &MissingRequiredFileNotice{
		BaseNotice: NewBaseNotice("missing_required_file", ERROR, context),
	}
}

// MissingRecommendedFieldNotice is generated when a recommended field is missing
type MissingRecommendedFieldNotice struct {
	*BaseNotice
}

func NewMissingRecommendedFieldNotice(filename string, fieldName string, rowNumber int) *MissingRecommendedFieldNotice {
	context := map[string]interface{}{
		"filename":     filename,
		"fieldName":    fieldName,
		"csvRowNumber": rowNumber,
	}
	return &MissingRecommendedFieldNotice{
		BaseNotice: NewBaseNotice("missing_recommended_field", WARNING, context),
	}
}

// MissingCalendarAndCalendarDateFilesNotice is generated when both calendar.txt and calendar_dates.txt are missing
type MissingCalendarAndCalendarDateFilesNotice struct {
	*BaseNotice
}

func NewMissingCalendarAndCalendarDateFilesNotice() *MissingCalendarAndCalendarDateFilesNotice {
	context := map[string]interface{}{
		"message": "At least one of calendar.txt or calendar_dates.txt must be provided",
	}
	return &MissingCalendarAndCalendarDateFilesNotice{
		BaseNotice: NewBaseNotice("missing_calendar_and_calendar_date_files", ERROR, context),
	}
}

// DecreasingOrEqualStopTimeDistanceNotice is generated when shape_dist_traveled values are decreasing or equal
type DecreasingOrEqualStopTimeDistanceNotice struct {
	*BaseNotice
}

func NewDecreasingOrEqualStopTimeDistanceNotice(tripID string, stopID string, rowNumber int, shapeDistTraveled float64, stopSequence int, prevRowNumber int, prevShapeDistTraveled float64, prevStopSequence int) *DecreasingOrEqualStopTimeDistanceNotice {
	context := map[string]interface{}{
		"tripId":                tripID,
		"stopId":                stopID,
		"csvRowNumber":          rowNumber,
		"shapeDistTraveled":     shapeDistTraveled,
		"stopSequence":          stopSequence,
		"prevCsvRowNumber":      prevRowNumber,
		"prevShapeDistTraveled": prevShapeDistTraveled,
		"prevStopSequence":      prevStopSequence,
	}
	return &DecreasingOrEqualStopTimeDistanceNotice{
		BaseNotice: NewBaseNotice("decreasing_or_equal_stop_time_distance", ERROR, context),
	}
}

// MissingRouteNameNotice is generated when both route_short_name and route_long_name are missing
type MissingRouteNameNotice struct {
	*BaseNotice
}

func NewMissingRouteNameNotice(routeID string, rowNumber int) *MissingRouteNameNotice {
	context := map[string]interface{}{
		"routeId":      routeID,
		"csvRowNumber": rowNumber,
		"message":      "Either route_short_name or route_long_name must be provided",
	}
	return &MissingRouteNameNotice{
		BaseNotice: NewBaseNotice("route_both_short_and_long_name_missing", ERROR, context),
	}
}

// RouteShortNameTooLongNotice is generated when route_short_name exceeds recommended length
type RouteShortNameTooLongNotice struct {
	*BaseNotice
}

func NewRouteShortNameTooLongNotice(routeID string, routeShortName string, actualLength int, maxLength int, rowNumber int) *RouteShortNameTooLongNotice {
	context := map[string]interface{}{
		"routeId":        routeID,
		"routeShortName": routeShortName,
		"actualLength":   actualLength,
		"maxLength":      maxLength,
		"csvRowNumber":   rowNumber,
	}
	return &RouteShortNameTooLongNotice{
		BaseNotice: NewBaseNotice("route_short_name_too_long", WARNING, context),
	}
}

// TripUsabilityNotice is generated when a trip has fewer than 2 stops
type TripUsabilityNotice struct {
	*BaseNotice
}

func NewTripUsabilityNotice(tripID string, stopCount int, rowNumber int) *TripUsabilityNotice {
	context := map[string]interface{}{
		"tripId":       tripID,
		"stopCount":    stopCount,
		"csvRowNumber": rowNumber,
	}
	return &TripUsabilityNotice{
		BaseNotice: NewBaseNotice("unusable_trip", WARNING, context),
	}
}

// StopTimeArrivalAfterDepartureNotice is generated when arrival time is after departure time
type StopTimeArrivalAfterDepartureNotice struct {
	*BaseNotice
}

func NewStopTimeArrivalAfterDepartureNotice(tripID string, stopSequence int, arrivalTime string, departureTime string, rowNumber int) *StopTimeArrivalAfterDepartureNotice {
	context := map[string]interface{}{
		"tripId":        tripID,
		"stopSequence":  stopSequence,
		"arrivalTime":   arrivalTime,
		"departureTime": departureTime,
		"csvRowNumber":  rowNumber,
	}
	return &StopTimeArrivalAfterDepartureNotice{
		BaseNotice: NewBaseNotice("stop_time_arrival_after_departure", WARNING, context),
	}
}

// StopTimeDecreasingTimeNotice is generated when times decrease along a trip
type StopTimeDecreasingTimeNotice struct {
	*BaseNotice
}

func NewStopTimeDecreasingTimeNotice(tripID string, stopSequence int, arrivalTime string, rowNumber int, prevStopSequence int, prevDepartureTime string, prevRowNumber int) *StopTimeDecreasingTimeNotice {
	context := map[string]interface{}{
		"tripId":            tripID,
		"stopSequence":      stopSequence,
		"arrivalTime":       arrivalTime,
		"csvRowNumber":      rowNumber,
		"prevStopSequence":  prevStopSequence,
		"prevDepartureTime": prevDepartureTime,
		"prevCsvRowNumber":  prevRowNumber,
	}
	return &StopTimeDecreasingTimeNotice{
		BaseNotice: NewBaseNotice("stop_time_with_arrival_before_previous_departure_time", ERROR, context),
	}
}

// ExcessiveTravelSpeedNotice is generated when travel speed between stops is unrealistic
type ExcessiveTravelSpeedNotice struct {
	*BaseNotice
}

func NewExcessiveTravelSpeedNotice(tripID string, fromStopID string, toStopID string, fromStopSequence int, toStopSequence int, speed float64, speedLimit float64, distance float64, timeDiff int, routeType int, fromRowNumber int, toRowNumber int) *ExcessiveTravelSpeedNotice {
	context := map[string]interface{}{
		"tripId":           tripID,
		"fromStopId":       fromStopID,
		"toStopId":         toStopID,
		"fromStopSequence": fromStopSequence,
		"toStopSequence":   toStopSequence,
		"speed":            speed,
		"speedLimit":       speedLimit,
		"distance":         distance,
		"timeDiff":         timeDiff,
		"routeType":        routeType,
		"fromRowNumber":    fromRowNumber,
		"toRowNumber":      toRowNumber,
	}
	return &ExcessiveTravelSpeedNotice{
		BaseNotice: NewBaseNotice("fast_travel_between_consecutive_stops", WARNING, context),
	}
}

// BlockTripsOverlapNotice is generated when trips in the same block overlap in time
type BlockTripsOverlapNotice struct {
	*BaseNotice
}

func NewBlockTripsOverlapNotice(blockID string, trip1ID string, trip2ID string, service1ID string, service2ID string, trip1StartTime string, trip1EndTime string, trip2StartTime string, trip2EndTime string, trip1RowNumber int, trip2RowNumber int) *BlockTripsOverlapNotice {
	context := map[string]interface{}{
		"blockId":        blockID,
		"trip1Id":        trip1ID,
		"trip2Id":        trip2ID,
		"service1Id":     service1ID,
		"service2Id":     service2ID,
		"trip1StartTime": trip1StartTime,
		"trip1EndTime":   trip1EndTime,
		"trip2StartTime": trip2StartTime,
		"trip2EndTime":   trip2EndTime,
		"trip1RowNumber": trip1RowNumber,
		"trip2RowNumber": trip2RowNumber,
	}
	return &BlockTripsOverlapNotice{
		BaseNotice: NewBaseNotice("block_trips_with_overlapping_stop_times", ERROR, context),
	}
}

// UnknownFileNotice is generated when an unknown file is found in the feed
type UnknownFileNotice struct {
	*BaseNotice
}

func NewUnknownFileNotice(filename string) *UnknownFileNotice {
	context := map[string]interface{}{
		"filename": filename,
	}
	return &UnknownFileNotice{
		BaseNotice: NewBaseNotice("unknown_file", INFO, context),
	}
}

// DuplicateHeaderNotice is generated when duplicate column headers are found
type DuplicateHeaderNotice struct {
	*BaseNotice
}

func NewDuplicateHeaderNotice(filename string, headerName string, positions []int) *DuplicateHeaderNotice {
	context := map[string]interface{}{
		"filename":   filename,
		"headerName": headerName,
		"positions":  positions,
	}
	return &DuplicateHeaderNotice{
		BaseNotice: NewBaseNotice("duplicated_column", ERROR, context),
	}
}

// MissingRequiredColumnNotice is generated when a required column is missing
type MissingRequiredColumnNotice struct {
	*BaseNotice
}

func NewMissingRequiredColumnNotice(filename string, columnName string) *MissingRequiredColumnNotice {
	context := map[string]interface{}{
		"filename":   filename,
		"columnName": columnName,
	}
	return &MissingRequiredColumnNotice{
		BaseNotice: NewBaseNotice("missing_required_column", ERROR, context),
	}
}

// InvalidColorNotice is generated when a color field has invalid format
type InvalidColorNotice struct {
	*BaseNotice
}

func NewInvalidColorNotice(routeID string, fieldName string, colorValue string, rowNumber int) *InvalidColorNotice {
	context := map[string]interface{}{
		"routeId":      routeID,
		"fieldName":    fieldName,
		"colorValue":   colorValue,
		"csvRowNumber": rowNumber,
	}
	return &InvalidColorNotice{
		BaseNotice: NewBaseNotice("invalid_color", ERROR, context),
	}
}

// ServiceWithoutActiveDaysNotice is generated when a service has no active days
type ServiceWithoutActiveDaysNotice struct {
	*BaseNotice
}

func NewServiceWithoutActiveDaysNotice(serviceID string, rowNumber int) *ServiceWithoutActiveDaysNotice {
	context := map[string]interface{}{
		"serviceId":    serviceID,
		"csvRowNumber": rowNumber,
	}
	return &ServiceWithoutActiveDaysNotice{
		BaseNotice: NewBaseNotice("service_has_no_active_day_of_the_week", WARNING, context),
	}
}

// UnusedServiceNotice is generated when a service is defined but not used
type UnusedServiceNotice struct {
	*BaseNotice
}

func NewUnusedServiceNotice(serviceID string, filename string, rowNumber int) *UnusedServiceNotice {
	context := map[string]interface{}{
		"serviceId":    serviceID,
		"filename":     filename,
		"csvRowNumber": rowNumber,
	}
	return &UnusedServiceNotice{
		BaseNotice: NewBaseNotice("unused_service", WARNING, context),
	}
}

// StationWithParentStationNotice is generated when a station has a parent station
type StationWithParentStationNotice struct {
	*BaseNotice
}

func NewStationWithParentStationNotice(stopID string, parentStation string, rowNumber int) *StationWithParentStationNotice {
	context := map[string]interface{}{
		"stopId":        stopID,
		"parentStation": parentStation,
		"csvRowNumber":  rowNumber,
	}
	return &StationWithParentStationNotice{
		BaseNotice: NewBaseNotice("station_with_parent_station", ERROR, context),
	}
}

// CircularStationReferenceNotice is generated when there's a circular parent station reference
type CircularStationReferenceNotice struct {
	*BaseNotice
}

func NewCircularStationReferenceNotice(stopID string, rowNumber int) *CircularStationReferenceNotice {
	context := map[string]interface{}{
		"stopId":       stopID,
		"csvRowNumber": rowNumber,
	}
	return &CircularStationReferenceNotice{
		BaseNotice: NewBaseNotice("circular_station_reference", WARNING, context),
	}
}

// OverlappingFrequencyNotice is generated when frequencies overlap for the same trip
type OverlappingFrequencyNotice struct {
	*BaseNotice
}

func NewOverlappingFrequencyNotice(tripID string, startTime1 string, endTime1 string, rowNumber1 int, startTime2 string, endTime2 string, rowNumber2 int) *OverlappingFrequencyNotice {
	context := map[string]interface{}{
		"tripId":     tripID,
		"startTime1": startTime1,
		"endTime1":   endTime1,
		"rowNumber1": rowNumber1,
		"startTime2": startTime2,
		"endTime2":   endTime2,
		"rowNumber2": rowNumber2,
	}
	return &OverlappingFrequencyNotice{
		BaseNotice: NewBaseNotice("overlapping_frequency", ERROR, context),
	}
}

// TransferToSameStopNotice is generated when transfer is from/to the same stop
type TransferToSameStopNotice struct {
	*BaseNotice
}

func NewTransferToSameStopNotice(stopID string, rowNumber int) *TransferToSameStopNotice {
	context := map[string]interface{}{
		"stopId":       stopID,
		"csvRowNumber": rowNumber,
	}
	return &TransferToSameStopNotice{
		BaseNotice: NewBaseNotice("transfer_to_same_stop", WARNING, context),
	}
}

// MissingMinTransferTimeNotice is generated when min_transfer_time is required but missing
type MissingMinTransferTimeNotice struct {
	*BaseNotice
}

func NewMissingMinTransferTimeNotice(fromStopID string, toStopID string, rowNumber int) *MissingMinTransferTimeNotice {
	context := map[string]interface{}{
		"fromStopId":   fromStopID,
		"toStopId":     toStopID,
		"csvRowNumber": rowNumber,
	}
	return &MissingMinTransferTimeNotice{
		BaseNotice: NewBaseNotice("missing_min_transfer_time", WARNING, context),
	}
}

// UnnecessaryMinTransferTimeNotice is generated when min_transfer_time is provided but not needed
type UnnecessaryMinTransferTimeNotice struct {
	*BaseNotice
}

func NewUnnecessaryMinTransferTimeNotice(fromStopID string, toStopID string, minTransferTime int, rowNumber int) *UnnecessaryMinTransferTimeNotice {
	context := map[string]interface{}{
		"fromStopId":      fromStopID,
		"toStopId":        toStopID,
		"minTransferTime": minTransferTime,
		"csvRowNumber":    rowNumber,
	}
	return &UnnecessaryMinTransferTimeNotice{
		BaseNotice: NewBaseNotice("unnecessary_min_transfer_time", WARNING, context),
	}
}

// NegativeMinTransferTimeNotice is generated when min_transfer_time is negative
type NegativeMinTransferTimeNotice struct {
	*BaseNotice
}

func NewNegativeMinTransferTimeNotice(fromStopID string, toStopID string, minTransferTime int, rowNumber int) *NegativeMinTransferTimeNotice {
	context := map[string]interface{}{
		"fromStopId":      fromStopID,
		"toStopId":        toStopID,
		"minTransferTime": minTransferTime,
		"csvRowNumber":    rowNumber,
	}
	return &NegativeMinTransferTimeNotice{
		BaseNotice: NewBaseNotice("negative_min_transfer_time", WARNING, context),
	}
}

// UnreasonableMinTransferTimeNotice is generated when min_transfer_time is unreasonably long
type UnreasonableMinTransferTimeNotice struct {
	*BaseNotice
}

func NewUnreasonableMinTransferTimeNotice(fromStopID string, toStopID string, minTransferTime int, rowNumber int) *UnreasonableMinTransferTimeNotice {
	context := map[string]interface{}{
		"fromStopId":      fromStopID,
		"toStopId":        toStopID,
		"minTransferTime": minTransferTime,
		"csvRowNumber":    rowNumber,
	}
	return &UnreasonableMinTransferTimeNotice{
		BaseNotice: NewBaseNotice("unreasonable_min_transfer_time", WARNING, context),
	}
}

// DuplicateTransferNotice is generated when duplicate transfers are defined
type DuplicateTransferNotice struct {
	*BaseNotice
}

func NewDuplicateTransferNotice(fromStopID string, toStopID string, rowNumber int, duplicateRowNumber int) *DuplicateTransferNotice {
	context := map[string]interface{}{
		"fromStopId":         fromStopID,
		"toStopId":           toStopID,
		"csvRowNumber":       rowNumber,
		"duplicateRowNumber": duplicateRowNumber,
	}
	return &DuplicateTransferNotice{
		BaseNotice: NewBaseNotice("duplicate_transfer", WARNING, context),
	}
}

// PATHWAY VALIDATOR NOTICES

// DuplicatePathwayNotice is generated when duplicate pathways are defined
type DuplicatePathwayNotice struct {
	*BaseNotice
}

func NewDuplicatePathwayNotice(pathwayID string, fromStopID string, toStopID string, rowNumber int, duplicateRowNumber int) *DuplicatePathwayNotice {
	context := map[string]interface{}{
		"pathwayId":          pathwayID,
		"fromStopId":         fromStopID,
		"toStopId":           toStopID,
		"csvRowNumber":       rowNumber,
		"duplicateRowNumber": duplicateRowNumber,
	}
	return &DuplicatePathwayNotice{
		BaseNotice: NewBaseNotice("duplicate_pathway", WARNING, context),
	}
}

// InconsistentBidirectionalPathwayNotice is generated when bidirectional pathways have inconsistent properties
type InconsistentBidirectionalPathwayNotice struct {
	*BaseNotice
}

func NewInconsistentBidirectionalPathwayNotice(pathwayID1 string, pathwayID2 string, rowNumber1 int, rowNumber2 int) *InconsistentBidirectionalPathwayNotice {
	context := map[string]interface{}{
		"pathwayId1":    pathwayID1,
		"pathwayId2":    pathwayID2,
		"csvRowNumber1": rowNumber1,
		"csvRowNumber2": rowNumber2,
	}
	return &InconsistentBidirectionalPathwayNotice{
		BaseNotice: NewBaseNotice("inconsistent_bidirectional_pathway", WARNING, context),
	}
}

// FARE VALIDATOR NOTICES

// UnnecessaryTransferDurationNotice is generated when transfer_duration is provided but transfers is 0
type UnnecessaryTransferDurationNotice struct {
	*BaseNotice
}

func NewUnnecessaryTransferDurationNotice(fareID string, transferDuration int, rowNumber int) *UnnecessaryTransferDurationNotice {
	context := map[string]interface{}{
		"fareId":           fareID,
		"transferDuration": transferDuration,
		"csvRowNumber":     rowNumber,
	}
	return &UnnecessaryTransferDurationNotice{
		BaseNotice: NewBaseNotice("unnecessary_transfer_duration", WARNING, context),
	}
}

// EmptyFareRuleNotice is generated when fare rule has no rule fields
type EmptyFareRuleNotice struct {
	*BaseNotice
}

func NewEmptyFareRuleNotice(fareID string, rowNumber int) *EmptyFareRuleNotice {
	context := map[string]interface{}{
		"fareId":       fareID,
		"csvRowNumber": rowNumber,
	}
	return &EmptyFareRuleNotice{
		BaseNotice: NewBaseNotice("empty_fare_rule", WARNING, context),
	}
}

// ConflictingFareRuleFieldsNotice is generated when conflicting fare rule fields are used
type ConflictingFareRuleFieldsNotice struct {
	*BaseNotice
}

func NewConflictingFareRuleFieldsNotice(fareID string, rowNumber int) *ConflictingFareRuleFieldsNotice {
	context := map[string]interface{}{
		"fareId":       fareID,
		"csvRowNumber": rowNumber,
	}
	return &ConflictingFareRuleFieldsNotice{
		BaseNotice: NewBaseNotice("conflicting_fare_rule_fields", WARNING, context),
	}
}

// UnusedFareAttributeNotice is generated when fare attribute is never used
type UnusedFareAttributeNotice struct {
	*BaseNotice
}

func NewUnusedFareAttributeNotice(fareID string, rowNumber int) *UnusedFareAttributeNotice {
	context := map[string]interface{}{
		"fareId":       fareID,
		"csvRowNumber": rowNumber,
	}
	return &UnusedFareAttributeNotice{
		BaseNotice: NewBaseNotice("unused_fare_attribute", WARNING, context),
	}
}

// LEVEL VALIDATOR NOTICES

// DuplicateLevelIndexNotice is generated when duplicate level indices are found
type DuplicateLevelIndexNotice struct {
	*BaseNotice
}

func NewDuplicateLevelIndexNotice(levelID string, levelIndex float64, rowNumber int, duplicateRowNumber int) *DuplicateLevelIndexNotice {
	context := map[string]interface{}{
		"levelId":            levelID,
		"levelIndex":         levelIndex,
		"csvRowNumber":       rowNumber,
		"duplicateRowNumber": duplicateRowNumber,
	}
	return &DuplicateLevelIndexNotice{
		BaseNotice: NewBaseNotice("duplicate_level_index", WARNING, context),
	}
}

// UnusedLevelNotice is generated when level is never used
type UnusedLevelNotice struct {
	*BaseNotice
}

func NewUnusedLevelNotice(levelID string, rowNumber int) *UnusedLevelNotice {
	context := map[string]interface{}{
		"levelId":      levelID,
		"csvRowNumber": rowNumber,
	}
	return &UnusedLevelNotice{
		BaseNotice: NewBaseNotice("unused_level", WARNING, context),
	}
}

// SHAPE VALIDATOR NOTICES

// InsufficientShapePointsNotice is generated when shape has too few points
type InsufficientShapePointsNotice struct {
	*BaseNotice
}

func NewInsufficientShapePointsNotice(shapeID string, pointCount int) *InsufficientShapePointsNotice {
	context := map[string]interface{}{
		"shapeId":    shapeID,
		"pointCount": pointCount,
	}
	return &InsufficientShapePointsNotice{
		BaseNotice: NewBaseNotice("single_shape_point", WARNING, context),
	}
}

// DecreasingShapeDistanceNotice is generated when shape distances decrease
type DecreasingShapeDistanceNotice struct {
	*BaseNotice
}

func NewDecreasingShapeDistanceNotice(shapeID string, sequence int, currentDistance float64, previousDistance float64, rowNumber int) *DecreasingShapeDistanceNotice {
	context := map[string]interface{}{
		"shapeId":          shapeID,
		"shapePtSequence":  sequence,
		"currentDistance":  currentDistance,
		"previousDistance": previousDistance,
		"csvRowNumber":     rowNumber,
	}
	return &DecreasingShapeDistanceNotice{
		BaseNotice: NewBaseNotice("decreasing_shape_distance", ERROR, context),
	}
}

// EqualShapeDistanceSameCoordinatesNotice reports two consecutive shape points
// with the same shape_dist_traveled and the same coordinates — a duplicated
// shape point rather than a distance error.
type EqualShapeDistanceSameCoordinatesNotice struct {
	*BaseNotice
}

func NewEqualShapeDistanceSameCoordinatesNotice(shapeID string, shapeDistTraveled float64, prevSequence int, sequence int, prevRowNumber int, rowNumber int) *EqualShapeDistanceSameCoordinatesNotice {
	return &EqualShapeDistanceSameCoordinatesNotice{
		BaseNotice: NewBaseNotice("equal_shape_distance_same_coordinates", WARNING, shapeDistanceContext(
			shapeID, shapeDistTraveled, prevSequence, sequence, prevRowNumber, rowNumber, nil,
		)),
	}
}

// EqualShapeDistanceDiffCoordinatesNotice reports two consecutive shape points
// with the same shape_dist_traveled but coordinates more than 1.11 m apart:
// the shape covers ground without the distance advancing.
type EqualShapeDistanceDiffCoordinatesNotice struct {
	*BaseNotice
}

func NewEqualShapeDistanceDiffCoordinatesNotice(shapeID string, shapeDistTraveled float64, prevSequence int, sequence int, prevRowNumber int, rowNumber int, actualDistance float64) *EqualShapeDistanceDiffCoordinatesNotice {
	return &EqualShapeDistanceDiffCoordinatesNotice{
		BaseNotice: NewBaseNotice("equal_shape_distance_diff_coordinates", ERROR, shapeDistanceContext(
			shapeID, shapeDistTraveled, prevSequence, sequence, prevRowNumber, rowNumber, &actualDistance,
		)),
	}
}

// EqualShapeDistanceDiffCoordinatesBelowThresholdNotice reports the same
// defect within 1.11 m, where the coordinate difference is small enough to be
// rounding rather than a real gap.
type EqualShapeDistanceDiffCoordinatesBelowThresholdNotice struct {
	*BaseNotice
}

func NewEqualShapeDistanceDiffCoordinatesBelowThresholdNotice(shapeID string, shapeDistTraveled float64, prevSequence int, sequence int, prevRowNumber int, rowNumber int, actualDistance float64) *EqualShapeDistanceDiffCoordinatesBelowThresholdNotice {
	return &EqualShapeDistanceDiffCoordinatesBelowThresholdNotice{
		BaseNotice: NewBaseNotice("equal_shape_distance_diff_coordinates_distance_below_threshold", WARNING, shapeDistanceContext(
			shapeID, shapeDistTraveled, prevSequence, sequence, prevRowNumber, rowNumber, &actualDistance,
		)),
	}
}

// shapeDistanceContext builds the context the three equal-distance notices
// share, so they stay reportable as one family.
func shapeDistanceContext(shapeID string, shapeDistTraveled float64, prevSequence int, sequence int, prevRowNumber int, rowNumber int, actualDistance *float64) map[string]interface{} {
	context := map[string]interface{}{
		"shapeId":             shapeID,
		"shapeDistTraveled":   shapeDistTraveled,
		"prevShapePtSequence": prevSequence,
		"shapePtSequence":     sequence,
		"prevCsvRowNumber":    prevRowNumber,
		"csvRowNumber":        rowNumber,
	}
	if actualDistance != nil {
		context["actualDistanceBetweenShapePoints"] = *actualDistance
	}
	return context
}

// CALENDAR CONSISTENCY VALIDATOR NOTICES

// ConflictingCalendarExceptionNotice is generated when conflicting exceptions exist
type ConflictingCalendarExceptionNotice struct {
	*BaseNotice
}

func NewConflictingCalendarExceptionNotice(serviceID string, date string, rowNumber1 int, rowNumber2 int) *ConflictingCalendarExceptionNotice {
	context := map[string]interface{}{
		"serviceId":     serviceID,
		"date":          date,
		"csvRowNumber1": rowNumber1,
		"csvRowNumber2": rowNumber2,
	}
	return &ConflictingCalendarExceptionNotice{
		BaseNotice: NewBaseNotice("conflicting_calendar_exception", WARNING, context),
	}
}

// DuplicateCalendarExceptionNotice is generated when duplicate exceptions exist
type DuplicateCalendarExceptionNotice struct {
	*BaseNotice
}

func NewDuplicateCalendarExceptionNotice(serviceID string, date string, rowNumber1 int, rowNumber2 int) *DuplicateCalendarExceptionNotice {
	context := map[string]interface{}{
		"serviceId":     serviceID,
		"date":          date,
		"csvRowNumber1": rowNumber1,
		"csvRowNumber2": rowNumber2,
	}
	return &DuplicateCalendarExceptionNotice{
		BaseNotice: NewBaseNotice("duplicate_calendar_exception", WARNING, context),
	}
}

// ATTRIBUTION VALIDATOR NOTICES

// MultipleAttributionScopesNotice is generated when multiple scopes are specified
type MultipleAttributionScopesNotice struct {
	*BaseNotice
}

func NewMultipleAttributionScopesNotice(attributionID string, rowNumber int) *MultipleAttributionScopesNotice {
	context := map[string]interface{}{
		"attributionId": attributionID,
		"csvRowNumber":  rowNumber,
	}
	return &MultipleAttributionScopesNotice{
		BaseNotice: NewBaseNotice("multiple_attribution_scopes", WARNING, context),
	}
}

// ConflictingAttributionScopeNotice is generated when conflicting scopes are specified
type ConflictingAttributionScopeNotice struct {
	*BaseNotice
}

func NewConflictingAttributionScopeNotice(attributionID string, rowNumber int) *ConflictingAttributionScopeNotice {
	context := map[string]interface{}{
		"attributionId": attributionID,
		"csvRowNumber":  rowNumber,
	}
	return &ConflictingAttributionScopeNotice{
		BaseNotice: NewBaseNotice("conflicting_attribution_scope", WARNING, context),
	}
}

// MissingAttributionContactNotice is generated when no contact information is provided
type MissingAttributionContactNotice struct {
	*BaseNotice
}

func NewMissingAttributionContactNotice(attributionID string, rowNumber int) *MissingAttributionContactNotice {
	context := map[string]interface{}{
		"attributionId": attributionID,
		"csvRowNumber":  rowNumber,
	}
	return &MissingAttributionContactNotice{
		BaseNotice: NewBaseNotice("missing_attribution_contact", WARNING, context),
	}
}

// DuplicateAttributionScopeNotice is generated when duplicate attribution scopes exist
type DuplicateAttributionScopeNotice struct {
	*BaseNotice
}

func NewDuplicateAttributionScopeNotice(attributionID1 string, attributionID2 string, scope string, rowNumber1 int, rowNumber2 int) *DuplicateAttributionScopeNotice {
	context := map[string]interface{}{
		"attributionId1": attributionID1,
		"attributionId2": attributionID2,
		"scope":          scope,
		"csvRowNumber1":  rowNumber1,
		"csvRowNumber2":  rowNumber2,
	}
	return &DuplicateAttributionScopeNotice{
		BaseNotice: NewBaseNotice("duplicate_attribution_scope", WARNING, context),
	}
}

// FEED INFO VALIDATOR NOTICES

// InvalidLanguageCodeNotice is generated when language code is invalid
type InvalidLanguageCodeNotice struct {
	*BaseNotice
}

func NewInvalidLanguageCodeNotice(filename string, fieldName string, languageCode string, rowNumber int) *InvalidLanguageCodeNotice {
	context := map[string]interface{}{
		"filename":     filename,
		"fieldName":    fieldName,
		"languageCode": languageCode,
		"csvRowNumber": rowNumber,
	}
	return &InvalidLanguageCodeNotice{
		BaseNotice: NewBaseNotice("invalid_language_code", ERROR, context),
	}
}

// ZONE VALIDATOR NOTICES

// UnusedZoneNotice is generated when zone is defined but not used
type UnusedZoneNotice struct {
	*BaseNotice
}

func NewUnusedZoneNotice(zoneID string, rowNumber int) *UnusedZoneNotice {
	context := map[string]interface{}{
		"zoneId":       zoneID,
		"csvRowNumber": rowNumber,
	}
	return &UnusedZoneNotice{
		BaseNotice: NewBaseNotice("unused_zone", WARNING, context),
	}
}

// STOP TIME CONSISTENCY VALIDATOR NOTICES

// MissingTripFirstTimeNotice is generated when first stop has no times
type MissingTripFirstTimeNotice struct {
	*BaseNotice
}

func NewMissingTripFirstTimeNotice(tripID string, stopID string, rowNumber int) *MissingTripFirstTimeNotice {
	context := map[string]interface{}{
		"tripId":       tripID,
		"stopId":       stopID,
		"csvRowNumber": rowNumber,
	}
	return &MissingTripFirstTimeNotice{
		BaseNotice: NewBaseNotice("missing_trip_edge", ERROR, context),
	}
}

// MissingTripLastTimeNotice is generated when last stop has no times
type MissingTripLastTimeNotice struct {
	*BaseNotice
}

func NewMissingTripLastTimeNotice(tripID string, stopID string, rowNumber int) *MissingTripLastTimeNotice {
	context := map[string]interface{}{
		"tripId":       tripID,
		"stopId":       stopID,
		"csvRowNumber": rowNumber,
	}
	return &MissingTripLastTimeNotice{
		BaseNotice: NewBaseNotice("missing_trip_edge", ERROR, context),
	}
}

// DuplicateStopInTripNotice is generated when stop appears multiple times in trip
type DuplicateStopInTripNotice struct {
	*BaseNotice
}

func NewDuplicateStopInTripNotice(tripID string, stopID string, stopSequence int, rowNumber int) *DuplicateStopInTripNotice {
	context := map[string]interface{}{
		"tripId":       tripID,
		"stopId":       stopID,
		"stopSequence": stopSequence,
		"csvRowNumber": rowNumber,
	}
	return &DuplicateStopInTripNotice{
		BaseNotice: NewBaseNotice("duplicate_stop_in_trip", WARNING, context),
	}
}

// FirstStopNoPickupNotice is generated when first stop has no pickup
type FirstStopNoPickupNotice struct {
	*BaseNotice
}

func NewFirstStopNoPickupNotice(tripID string, stopID string, rowNumber int) *FirstStopNoPickupNotice {
	context := map[string]interface{}{
		"tripId":       tripID,
		"stopId":       stopID,
		"csvRowNumber": rowNumber,
	}
	return &FirstStopNoPickupNotice{
		BaseNotice: NewBaseNotice("first_stop_no_pickup", WARNING, context),
	}
}

// LastStopNoDropOffNotice is generated when last stop has no drop-off
type LastStopNoDropOffNotice struct {
	*BaseNotice
}

func NewLastStopNoDropOffNotice(tripID string, stopID string, rowNumber int) *LastStopNoDropOffNotice {
	context := map[string]interface{}{
		"tripId":       tripID,
		"stopId":       stopID,
		"csvRowNumber": rowNumber,
	}
	return &LastStopNoDropOffNotice{
		BaseNotice: NewBaseNotice("last_stop_no_drop_off", WARNING, context),
	}
}

// AllStopsNoPickupNotice is generated when all stops have no pickup
type AllStopsNoPickupNotice struct {
	*BaseNotice
}

func NewAllStopsNoPickupNotice(tripID string) *AllStopsNoPickupNotice {
	context := map[string]interface{}{
		"tripId": tripID,
	}
	return &AllStopsNoPickupNotice{
		BaseNotice: NewBaseNotice("all_stops_no_pickup", WARNING, context),
	}
}

// AllStopsNoDropOffNotice is generated when all stops have no drop-off
type AllStopsNoDropOffNotice struct {
	*BaseNotice
}

func NewAllStopsNoDropOffNotice(tripID string) *AllStopsNoDropOffNotice {
	context := map[string]interface{}{
		"tripId": tripID,
	}
	return &AllStopsNoDropOffNotice{
		BaseNotice: NewBaseNotice("all_stops_no_drop_off", WARNING, context),
	}
}

// InconsistentStopTimeShapeDistanceNotice is generated when some stops have shape distances
type InconsistentStopTimeShapeDistanceNotice struct {
	*BaseNotice
}

func NewInconsistentStopTimeShapeDistanceNotice(tripID string, missingCount int, totalCount int) *InconsistentStopTimeShapeDistanceNotice {
	context := map[string]interface{}{
		"tripId":       tripID,
		"missingCount": missingCount,
		"totalCount":   totalCount,
	}
	return &InconsistentStopTimeShapeDistanceNotice{
		BaseNotice: NewBaseNotice("inconsistent_stop_time_shape_distance", WARNING, context),
	}
}

// DuplicateCompositeKeyNotice represents a duplicate composite primary key error
type DuplicateCompositeKeyNotice struct {
	*BaseNotice
}

func NewDuplicateCompositeKeyNotice(filename, keyFields, keyValue string, firstRow, duplicateRow int) *DuplicateCompositeKeyNotice {
	context := map[string]interface{}{
		"filename":     filename,
		"keyFields":    keyFields,
		"keyValue":     keyValue,
		"firstRow":     firstRow,
		"duplicateRow": duplicateRow,
	}
	return &DuplicateCompositeKeyNotice{
		BaseNotice: NewBaseNotice("duplicate_composite_key", WARNING, context),
	}
}

// MultipleRecordsInSingleRecordFileNotice represents multiple records in a file that should have only one
type MultipleRecordsInSingleRecordFileNotice struct {
	*BaseNotice
}

func NewMultipleRecordsInSingleRecordFileNotice(filename string, recordCount int) *MultipleRecordsInSingleRecordFileNotice {
	context := map[string]interface{}{
		"filename":    filename,
		"recordCount": recordCount,
	}
	return &MultipleRecordsInSingleRecordFileNotice{
		BaseNotice: NewBaseNotice("more_than_one_entity", WARNING, context),
	}
}

// WrongNumberOfFieldsNotice represents a row with wrong number of fields
type WrongNumberOfFieldsNotice struct {
	*BaseNotice
}

func NewWrongNumberOfFieldsNotice(filename string, rowNumber, expectedFields, actualFields int) *WrongNumberOfFieldsNotice {
	context := map[string]interface{}{
		"filename":       filename,
		"rowNumber":      rowNumber,
		"expectedFields": expectedFields,
		"actualFields":   actualFields,
	}
	return &WrongNumberOfFieldsNotice{
		BaseNotice: NewBaseNotice("invalid_row_length", ERROR, context),
	}
}

// LeadingWhitespaceNotice represents a field with leading whitespace
type LeadingWhitespaceNotice struct {
	*BaseNotice
}

func NewLeadingWhitespaceNotice(filename, fieldName, fieldValue string, rowNumber int) *LeadingWhitespaceNotice {
	context := map[string]interface{}{
		"filename":   filename,
		"fieldName":  fieldName,
		"fieldValue": fieldValue,
		"rowNumber":  rowNumber,
	}
	return &LeadingWhitespaceNotice{
		BaseNotice: NewBaseNotice("leading_or_trailing_whitespaces", WARNING, context),
	}
}

// TrailingWhitespaceNotice represents a field with trailing whitespace
type TrailingWhitespaceNotice struct {
	*BaseNotice
}

func NewTrailingWhitespaceNotice(filename, fieldName, fieldValue string, rowNumber int) *TrailingWhitespaceNotice {
	context := map[string]interface{}{
		"filename":   filename,
		"fieldName":  fieldName,
		"fieldValue": fieldValue,
		"rowNumber":  rowNumber,
	}
	return &TrailingWhitespaceNotice{
		BaseNotice: NewBaseNotice("leading_or_trailing_whitespaces", WARNING, context),
	}
}

// ConsecutiveDuplicateStopsNotice represents consecutive duplicate stops
type ConsecutiveDuplicateStopsNotice struct {
	*BaseNotice
}

func NewConsecutiveDuplicateStopsNotice(tripID, stopID string, seq1, seq2, rowNumber int) *ConsecutiveDuplicateStopsNotice {
	context := map[string]interface{}{
		"tripId":    tripID,
		"stopId":    stopID,
		"sequence1": seq1,
		"sequence2": seq2,
		"rowNumber": rowNumber,
	}
	return &ConsecutiveDuplicateStopsNotice{
		BaseNotice: NewBaseNotice("consecutive_duplicate_stops", WARNING, context),
	}
}

// StopWithoutServiceNotice represents a stop with neither pickup nor drop-off
type StopWithoutServiceNotice struct {
	*BaseNotice
}

func NewStopWithoutServiceNotice(tripID, stopID string, stopSequence, rowNumber int) *StopWithoutServiceNotice {
	context := map[string]interface{}{
		"tripId":       tripID,
		"stopId":       stopID,
		"stopSequence": stopSequence,
		"rowNumber":    rowNumber,
	}
	return &StopWithoutServiceNotice{
		BaseNotice: NewBaseNotice("stop_without_service", WARNING, context),
	}
}

// RouteWithoutTripsNotice represents a route with no trips
type RouteWithoutTripsNotice struct {
	*BaseNotice
}

func NewRouteWithoutTripsNotice(routeID string, rowNumber int) *RouteWithoutTripsNotice {
	context := map[string]interface{}{
		"routeId":   routeID,
		"rowNumber": rowNumber,
	}
	return &RouteWithoutTripsNotice{
		BaseNotice: NewBaseNotice("route_without_trips", WARNING, context),
	}
}

// ChildStationTooFarFromParentNotice represents child station too far from parent
type ChildStationTooFarFromParentNotice struct {
	*BaseNotice
}

func NewChildStationTooFarFromParentNotice(childID, parentID string, distance float64, rowNumber int) *ChildStationTooFarFromParentNotice {
	context := map[string]interface{}{
		"childId":   childID,
		"parentId":  parentID,
		"distance":  distance,
		"rowNumber": rowNumber,
	}
	return &ChildStationTooFarFromParentNotice{
		BaseNotice: NewBaseNotice("child_station_too_far_from_parent", WARNING, context),
	}
}

// === NETWORK TOPOLOGY NOTICES ===

// === FEED EXPIRATION NOTICES ===

// FeedExpiresWithin7DaysNotice represents feed expiring within 7 days
type FeedExpiresWithin7DaysNotice struct {
	*BaseNotice
}

func NewFeedExpiresWithin7DaysNotice(endDate, currentDate string, daysUntilExpiration int) *FeedExpiresWithin7DaysNotice {
	context := map[string]interface{}{
		"endDate":             endDate,
		"currentDate":         currentDate,
		"daysUntilExpiration": daysUntilExpiration,
	}
	return &FeedExpiresWithin7DaysNotice{
		BaseNotice: NewBaseNotice("feed_expiration_date7_days", WARNING, context),
	}
}

// FeedExpiresWithin30DaysNotice represents feed expiring within 30 days
type FeedExpiresWithin30DaysNotice struct {
	*BaseNotice
}

func NewFeedExpiresWithin30DaysNotice(endDate, currentDate string, daysUntilExpiration int) *FeedExpiresWithin30DaysNotice {
	context := map[string]interface{}{
		"endDate":             endDate,
		"currentDate":         currentDate,
		"daysUntilExpiration": daysUntilExpiration,
	}
	return &FeedExpiresWithin30DaysNotice{
		BaseNotice: NewBaseNotice("feed_expiration_date30_days", WARNING, context),
	}
}

// === DUPLICATE ROUTE NAME NOTICES ===

// DuplicateRouteLongNameNotice represents duplicate route long names
type DuplicateRouteLongNameNotice struct {
	*BaseNotice
}

func NewDuplicateRouteLongNameNotice(routeID, routeLongName, firstRouteID, agencyID string, routeType, rowNumber int) *DuplicateRouteLongNameNotice {
	context := map[string]interface{}{
		"routeId":       routeID,
		"routeLongName": routeLongName,
		"firstRouteId":  firstRouteID,
		"agencyId":      agencyID,
		"routeType":     routeType,
		"csvRowNumber":  rowNumber,
	}
	return &DuplicateRouteLongNameNotice{
		BaseNotice: NewBaseNotice("duplicate_route_name", WARNING, context),
	}
}

// DuplicateRouteShortNameNotice represents duplicate route short names
type DuplicateRouteShortNameNotice struct {
	*BaseNotice
}

func NewDuplicateRouteShortNameNotice(routeID, routeShortName, firstRouteID, agencyID string, routeType, rowNumber int) *DuplicateRouteShortNameNotice {
	context := map[string]interface{}{
		"routeId":        routeID,
		"routeShortName": routeShortName,
		"firstRouteId":   firstRouteID,
		"agencyId":       agencyID,
		"routeType":      routeType,
		"csvRowNumber":   rowNumber,
	}
	return &DuplicateRouteShortNameNotice{
		BaseNotice: NewBaseNotice("duplicate_route_name", WARNING, context),
	}
}

// DuplicateRouteNameCombinationNotice represents duplicate route name combinations
type DuplicateRouteNameCombinationNotice struct {
	*BaseNotice
}

func NewDuplicateRouteNameCombinationNotice(routeID, routeLongName, routeShortName, firstRouteID, agencyID string, routeType, rowNumber int) *DuplicateRouteNameCombinationNotice {
	context := map[string]interface{}{
		"routeId":        routeID,
		"routeLongName":  routeLongName,
		"routeShortName": routeShortName,
		"firstRouteId":   firstRouteID,
		"agencyId":       agencyID,
		"routeType":      routeType,
		"csvRowNumber":   rowNumber,
	}
	return &DuplicateRouteNameCombinationNotice{
		BaseNotice: NewBaseNotice("duplicate_route_name", WARNING, context),
	}
}

// === ROUTE COLOR CONTRAST NOTICES ===

// RouteColorContrastNotice reports a route name that is hard to read against
// its own background.
//
// The measure is the Rec. 601 luma difference between the two colours, not a
// WCAG contrast ratio: WCAG's thresholds are calibrated for body text and are
// too harsh for a large coloured line badge, which is what these two fields
// actually render as.
type RouteColorContrastNotice struct {
	*BaseNotice
}

func NewRouteColorContrastNotice(routeID, routeColor, routeTextColor string, lumaDifference, minimumLumaDifference float64, rowNumber int, severity SeverityLevel) *RouteColorContrastNotice {
	context := map[string]interface{}{
		"routeId":               routeID,
		"routeColor":            routeColor,
		"routeTextColor":        routeTextColor,
		"lumaDifference":        lumaDifference,
		"minimumLumaDifference": minimumLumaDifference,
		"csvRowNumber":          rowNumber,
	}
	return &RouteColorContrastNotice{
		BaseNotice: NewBaseNotice("route_color_contrast", severity, context),
	}
}

// === STOP NAME NOTICES ===

// MissingRequiredStopNameNotice represents missing required stop name
type MissingRequiredStopNameNotice struct {
	*BaseNotice
}

func NewMissingRequiredStopNameNotice(stopID string, locationType, rowNumber int) *MissingRequiredStopNameNotice {
	context := map[string]interface{}{
		"stopId":       stopID,
		"locationType": locationType,
		"csvRowNumber": rowNumber,
	}
	return &MissingRequiredStopNameNotice{
		BaseNotice: NewBaseNotice("missing_stop_name", ERROR, context),
	}
}

// StopNameContainsControlCharacterNotice represents control characters in stop names
type StopNameContainsControlCharacterNotice struct {
	*BaseNotice
}

func NewStopNameContainsControlCharacterNotice(stopID, stopName string, position, charCode, rowNumber int) *StopNameContainsControlCharacterNotice {
	context := map[string]interface{}{
		"stopId":       stopID,
		"stopName":     stopName,
		"position":     position,
		"charCode":     charCode,
		"csvRowNumber": rowNumber,
	}
	return &StopNameContainsControlCharacterNotice{
		BaseNotice: NewBaseNotice("stop_name_contains_control_character", WARNING, context),
	}
}

// === DATE TRIPS NOTICES ===

// === BIKES ALLOWANCE NOTICES ===

// === SHAPE DISTANCE NOTICES ===

// === ADDITIONAL FREQUENCY NOTICES ===

// FrequencyDurationShorterThanHeadwayNotice represents duration shorter than headway
type FrequencyDurationShorterThanHeadwayNotice struct {
	*BaseNotice
}

func NewFrequencyDurationShorterThanHeadwayNotice(tripID string, duration, headway, rowNumber int) *FrequencyDurationShorterThanHeadwayNotice {
	context := map[string]interface{}{
		"tripId":       tripID,
		"duration":     duration,
		"headway":      headway,
		"csvRowNumber": rowNumber,
	}
	return &FrequencyDurationShorterThanHeadwayNotice{
		BaseNotice: NewBaseNotice("frequency_duration_shorter_than_headway", WARNING, context),
	}
}

// CrossTripFrequencyOverlapNotice represents overlapping frequencies between trips
type CrossTripFrequencyOverlapNotice struct {
	*BaseNotice
}

func NewCrossTripFrequencyOverlapNotice(trip1ID, trip2ID, routeID, serviceID, start1, end1, start2, end2 string, rowNumber int) *CrossTripFrequencyOverlapNotice {
	context := map[string]interface{}{
		"trip1Id":      trip1ID,
		"trip2Id":      trip2ID,
		"routeId":      routeID,
		"serviceId":    serviceID,
		"start1":       start1,
		"end1":         end1,
		"start2":       start2,
		"end2":         end2,
		"csvRowNumber": rowNumber,
	}
	return &CrossTripFrequencyOverlapNotice{
		BaseNotice: NewBaseNotice("cross_trip_frequency_overlap", WARNING, context),
	}
}

// === ATTRIBUTION NOTICES ===

// AttributionWithoutRoleNotice represents attribution without any role
type AttributionWithoutRoleNotice struct {
	*BaseNotice
}

func NewAttributionWithoutRoleNotice(attributionID, organizationName string, rowNumber int) *AttributionWithoutRoleNotice {
	context := map[string]interface{}{
		"attributionId":    attributionID,
		"organizationName": organizationName,
		"csvRowNumber":     rowNumber,
	}
	return &AttributionWithoutRoleNotice{
		BaseNotice: NewBaseNotice("attribution_without_role", WARNING, context),
	}
}

// AttributionAllRolesNotice represents attribution with all three roles
type AttributionAllRolesNotice struct {
	*BaseNotice
}

func NewAttributionAllRolesNotice(attributionID, organizationName string, rowNumber int) *AttributionAllRolesNotice {
	context := map[string]interface{}{
		"attributionId":    attributionID,
		"organizationName": organizationName,
		"csvRowNumber":     rowNumber,
	}
	return &AttributionAllRolesNotice{
		BaseNotice: NewBaseNotice("attribution_all_roles", INFO, context),
	}
}

// AttributionRoleNameMismatchNotice represents role-name mismatch
type AttributionRoleNameMismatchNotice struct {
	*BaseNotice
}

func NewAttributionRoleNameMismatchNotice(attributionID, organizationName, expectedRole string, rowNumber int) *AttributionRoleNameMismatchNotice {
	context := map[string]interface{}{
		"attributionId":    attributionID,
		"organizationName": organizationName,
		"expectedRole":     expectedRole,
		"csvRowNumber":     rowNumber,
	}
	return &AttributionRoleNameMismatchNotice{
		BaseNotice: NewBaseNotice("attribution_role_name_mismatch", INFO, context),
	}
}

// === TRIP BLOCK NOTICES ===

// BlockServiceMismatchNotice represents service mismatch within block
type BlockServiceMismatchNotice struct {
	*BaseNotice
}

func NewBlockServiceMismatchNotice(blockID, trip1ID, service1ID, trip2ID, service2ID string, rowNumber int) *BlockServiceMismatchNotice {
	context := map[string]interface{}{
		"blockId":      blockID,
		"trip1Id":      trip1ID,
		"service1Id":   service1ID,
		"trip2Id":      trip2ID,
		"service2Id":   service2ID,
		"csvRowNumber": rowNumber,
	}
	return &BlockServiceMismatchNotice{
		BaseNotice: NewBaseNotice("block_service_mismatch", WARNING, context),
	}
}

// === STOP TIME HEADSIGN NOTICES ===

// === ROUTE TYPE NOTICES ===

// DeprecatedRouteTypeNotice represents deprecated route_type value
type DeprecatedRouteTypeNotice struct {
	*BaseNotice
}

func NewDeprecatedRouteTypeNotice(routeID string, routeType, recommendedType, rowNumber int) *DeprecatedRouteTypeNotice {
	context := map[string]interface{}{
		"routeId":         routeID,
		"routeType":       routeType,
		"recommendedType": recommendedType,
		"csvRowNumber":    rowNumber,
	}
	return &DeprecatedRouteTypeNotice{
		BaseNotice: NewBaseNotice("deprecated_route_type", WARNING, context),
	}
}

// === ADDITIONAL TRANSFER TIMING NOTICES ===

// === SERVICE CALENDAR NOTICES ===

// === VALIDATOR SYSTEM NOTICES ===

// ValidatorErrorNotice reports that one of our validators panicked. It is not
// a feed defect: it means that validator's checks did not run, so the report
// has a hole in it. ERROR is deliberate — at WARNING a crashed validator would
// sit unnoticed among ordinary feed warnings.
type ValidatorErrorNotice struct {
	*BaseNotice
}

func NewValidatorErrorNotice(validatorName string, errorMessage string) *ValidatorErrorNotice {
	context := map[string]interface{}{
		"validatorName": validatorName,
		"errorMessage":  errorMessage,
	}
	return &ValidatorErrorNotice{
		BaseNotice: NewBaseNotice("runtime_exception_in_validator_error", ERROR, context),
	}
}
