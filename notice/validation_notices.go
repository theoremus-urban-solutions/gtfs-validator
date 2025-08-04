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
		BaseNotice: NewBaseNotice("invalid_field_format", ERROR, context),
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
		BaseNotice: NewBaseNotice("empty_file", WARNING, context),
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

// DuplicateStopSequenceNotice is generated when duplicate stop_sequence values are found for the same trip
type DuplicateStopSequenceNotice struct {
	*BaseNotice
}

func NewDuplicateStopSequenceNotice(tripID string, stopSequence int, stopID string, rowNumber int, duplicateRowNumber int) *DuplicateStopSequenceNotice {
	context := map[string]interface{}{
		"tripId":               tripID,
		"stopSequence":         stopSequence,
		"stopId":               stopID,
		"csvRowNumber":         rowNumber,
		"duplicateRowNumber":   duplicateRowNumber,
	}
	return &DuplicateStopSequenceNotice{
		BaseNotice: NewBaseNotice("duplicate_stop_sequence", ERROR, context),
	}
}

// DecreasingOrEqualStopTimeDistanceNotice is generated when shape_dist_traveled values are decreasing or equal
type DecreasingOrEqualStopTimeDistanceNotice struct {
	*BaseNotice
}

func NewDecreasingOrEqualStopTimeDistanceNotice(tripID string, stopID string, rowNumber int, shapeDistTraveled float64, stopSequence int, prevRowNumber int, prevShapeDistTraveled float64, prevStopSequence int) *DecreasingOrEqualStopTimeDistanceNotice {
	context := map[string]interface{}{
		"tripId":                 tripID,
		"stopId":                 stopID,
		"csvRowNumber":           rowNumber,
		"shapeDistTraveled":      shapeDistTraveled,
		"stopSequence":           stopSequence,
		"prevCsvRowNumber":       prevRowNumber,
		"prevShapeDistTraveled":  prevShapeDistTraveled,
		"prevStopSequence":       prevStopSequence,
	}
	return &DecreasingOrEqualStopTimeDistanceNotice{
		BaseNotice: NewBaseNotice("decreasing_or_equal_stop_time_distance", ERROR, context),
	}
}

// FeedExpirationDate7DaysNotice is generated when feed expires within 7 days
type FeedExpirationDate7DaysNotice struct {
	*BaseNotice
}

func NewFeedExpirationDate7DaysNotice(rowNumber int, currentDate string, feedEndDate string, suggestedExpirationDate string) *FeedExpirationDate7DaysNotice {
	context := map[string]interface{}{
		"csvRowNumber":            rowNumber,
		"currentDate":             currentDate,
		"feedEndDate":             feedEndDate,
		"suggestedExpirationDate": suggestedExpirationDate,
	}
	return &FeedExpirationDate7DaysNotice{
		BaseNotice: NewBaseNotice("feed_expiration_date_7_days", WARNING, context),
	}
}

// FeedExpirationDate30DaysNotice is generated when feed expires within 30 days
type FeedExpirationDate30DaysNotice struct {
	*BaseNotice
}

func NewFeedExpirationDate30DaysNotice(rowNumber int, currentDate string, feedEndDate string, suggestedExpirationDate string) *FeedExpirationDate30DaysNotice {
	context := map[string]interface{}{
		"csvRowNumber":            rowNumber,
		"currentDate":             currentDate,
		"feedEndDate":             feedEndDate,
		"suggestedExpirationDate": suggestedExpirationDate,
	}
	return &FeedExpirationDate30DaysNotice{
		BaseNotice: NewBaseNotice("feed_expiration_date_30_days", WARNING, context),
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
		BaseNotice: NewBaseNotice("missing_route_name", ERROR, context),
	}
}

// SameNameAndDescriptionNotice is generated when route_short_name and route_long_name are identical
type SameNameAndDescriptionNotice struct {
	*BaseNotice
}

func NewSameNameAndDescriptionNotice(routeID string, fieldName1 string, fieldName2 string, fieldValue string, rowNumber int) *SameNameAndDescriptionNotice {
	context := map[string]interface{}{
		"routeId":      routeID,
		"fieldName1":   fieldName1,
		"fieldName2":   fieldName2,
		"fieldValue":   fieldValue,
		"csvRowNumber": rowNumber,
	}
	return &SameNameAndDescriptionNotice{
		BaseNotice: NewBaseNotice("same_name_and_description", WARNING, context),
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

// RouteLongNameTooLongNotice is generated when route_long_name exceeds recommended length
type RouteLongNameTooLongNotice struct {
	*BaseNotice
}

func NewRouteLongNameTooLongNotice(routeID string, routeLongName string, actualLength int, maxLength int, rowNumber int) *RouteLongNameTooLongNotice {
	context := map[string]interface{}{
		"routeId":       routeID,
		"routeLongName": routeLongName,
		"actualLength":  actualLength,
		"maxLength":     maxLength,
		"csvRowNumber":  rowNumber,
	}
	return &RouteLongNameTooLongNotice{
		BaseNotice: NewBaseNotice("route_long_name_too_long", WARNING, context),
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
		BaseNotice: NewBaseNotice("trip_usability", ERROR, context),
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
		BaseNotice: NewBaseNotice("stop_time_arrival_after_departure", ERROR, context),
	}
}

// StopTimeDecreasingTimeNotice is generated when times decrease along a trip
type StopTimeDecreasingTimeNotice struct {
	*BaseNotice
}

func NewStopTimeDecreasingTimeNotice(tripID string, stopSequence int, arrivalTime string, rowNumber int, prevStopSequence int, prevDepartureTime string, prevRowNumber int) *StopTimeDecreasingTimeNotice {
	context := map[string]interface{}{
		"tripId":             tripID,
		"stopSequence":       stopSequence,
		"arrivalTime":        arrivalTime,
		"csvRowNumber":       rowNumber,
		"prevStopSequence":   prevStopSequence,
		"prevDepartureTime":  prevDepartureTime,
		"prevCsvRowNumber":   prevRowNumber,
	}
	return &StopTimeDecreasingTimeNotice{
		BaseNotice: NewBaseNotice("stop_time_decreasing_time", ERROR, context),
	}
}

// DuplicateShapeSequenceNotice is generated when duplicate shape_pt_sequence values are found
type DuplicateShapeSequenceNotice struct {
	*BaseNotice
}

func NewDuplicateShapeSequenceNotice(shapeID string, shapePtSequence int, rowNumber int, duplicateRowNumber int) *DuplicateShapeSequenceNotice {
	context := map[string]interface{}{
		"shapeId":            shapeID,
		"shapePtSequence":    shapePtSequence,
		"csvRowNumber":       rowNumber,
		"duplicateRowNumber": duplicateRowNumber,
	}
	return &DuplicateShapeSequenceNotice{
		BaseNotice: NewBaseNotice("duplicate_shape_sequence", ERROR, context),
	}
}

// DecreasingOrEqualShapeDistanceNotice is generated when shape_dist_traveled values are decreasing or equal
type DecreasingOrEqualShapeDistanceNotice struct {
	*BaseNotice
}

func NewDecreasingOrEqualShapeDistanceNotice(shapeID string, shapePtSequence int, rowNumber int, shapeDistTraveled float64, prevShapePtSequence int, prevRowNumber int, prevShapeDistTraveled float64) *DecreasingOrEqualShapeDistanceNotice {
	context := map[string]interface{}{
		"shapeId":                shapeID,
		"shapePtSequence":        shapePtSequence,
		"csvRowNumber":           rowNumber,
		"shapeDistTraveled":      shapeDistTraveled,
		"prevShapePtSequence":    prevShapePtSequence,
		"prevCsvRowNumber":       prevRowNumber,
		"prevShapeDistTraveled":  prevShapeDistTraveled,
	}
	return &DecreasingOrEqualShapeDistanceNotice{
		BaseNotice: NewBaseNotice("decreasing_or_equal_shape_distance", ERROR, context),
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
		BaseNotice: NewBaseNotice("excessive_travel_speed", WARNING, context),
	}
}

// BlockTripsOverlapNotice is generated when trips in the same block overlap in time
type BlockTripsOverlapNotice struct {
	*BaseNotice
}

func NewBlockTripsOverlapNotice(blockID string, trip1ID string, trip2ID string, service1ID string, service2ID string, trip1StartTime string, trip1EndTime string, trip2StartTime string, trip2EndTime string, trip1RowNumber int, trip2RowNumber int) *BlockTripsOverlapNotice {
	context := map[string]interface{}{
		"blockId":         blockID,
		"trip1Id":         trip1ID,
		"trip2Id":         trip2ID,
		"service1Id":      service1ID,
		"service2Id":      service2ID,
		"trip1StartTime":  trip1StartTime,
		"trip1EndTime":    trip1EndTime,
		"trip2StartTime":  trip2StartTime,
		"trip2EndTime":    trip2EndTime,
		"trip1RowNumber":  trip1RowNumber,
		"trip2RowNumber":  trip2RowNumber,
	}
	return &BlockTripsOverlapNotice{
		BaseNotice: NewBaseNotice("block_trips_overlap", ERROR, context),
	}
}

// MissingFeedInfoNotice is generated when feed_info.txt is required but missing
type MissingFeedInfoNotice struct {
	*BaseNotice
}

func NewMissingFeedInfoNotice() *MissingFeedInfoNotice {
	context := map[string]interface{}{
		"message": "feed_info.txt is required when translations.txt is present",
	}
	return &MissingFeedInfoNotice{
		BaseNotice: NewBaseNotice("missing_feed_info", ERROR, context),
	}
}

// MissingFareAttributesNotice is generated when fare_attributes.txt is required but missing
type MissingFareAttributesNotice struct {
	*BaseNotice
}

func NewMissingFareAttributesNotice() *MissingFareAttributesNotice {
	context := map[string]interface{}{
		"message": "fare_attributes.txt is required when fare_rules.txt is present",
	}
	return &MissingFareAttributesNotice{
		BaseNotice: NewBaseNotice("missing_fare_attributes", ERROR, context),
	}
}

// MissingLevelsNotice is generated when levels.txt is required but missing
type MissingLevelsNotice struct {
	*BaseNotice
}

func NewMissingLevelsNotice() *MissingLevelsNotice {
	context := map[string]interface{}{
		"message": "levels.txt is required when pathways.txt is present",
	}
	return &MissingLevelsNotice{
		BaseNotice: NewBaseNotice("missing_levels", ERROR, context),
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
		BaseNotice: NewBaseNotice("duplicate_header", ERROR, context),
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

// InvalidTimeFormatNotice is generated when a time field has invalid format
type InvalidTimeFormatNotice struct {
	*BaseNotice
}

func NewInvalidTimeFormatNotice(filename string, fieldName string, timeValue string, rowNumber int) *InvalidTimeFormatNotice {
	context := map[string]interface{}{
		"filename":     filename,
		"fieldName":    fieldName,
		"timeValue":    timeValue,
		"csvRowNumber": rowNumber,
	}
	return &InvalidTimeFormatNotice{
		BaseNotice: NewBaseNotice("invalid_time_format", ERROR, context),
	}
}

// InvalidDateFormatNotice is generated when a date field has invalid format
type InvalidDateFormatNotice struct {
	*BaseNotice
}

func NewInvalidDateFormatNotice(filename string, fieldName string, dateValue string, rowNumber int) *InvalidDateFormatNotice {
	context := map[string]interface{}{
		"filename":     filename,
		"fieldName":    fieldName,
		"dateValue":    dateValue,
		"csvRowNumber": rowNumber,
	}
	return &InvalidDateFormatNotice{
		BaseNotice: NewBaseNotice("invalid_date_format", ERROR, context),
	}
}

// InvalidCoordinateNotice is generated when a coordinate is out of valid range
type InvalidCoordinateNotice struct {
	*BaseNotice
}

func NewInvalidCoordinateNotice(filename string, fieldName string, coordValue string, rowNumber int, reason string) *InvalidCoordinateNotice {
	context := map[string]interface{}{
		"filename":     filename,
		"fieldName":    fieldName,
		"coordValue":   coordValue,
		"csvRowNumber": rowNumber,
		"reason":       reason,
	}
	return &InvalidCoordinateNotice{
		BaseNotice: NewBaseNotice("invalid_coordinate", ERROR, context),
	}
}

// SuspiciousCoordinateNotice is generated when a coordinate looks suspicious
type SuspiciousCoordinateNotice struct {
	*BaseNotice
}

func NewSuspiciousCoordinateNotice(filename string, fieldName string, coordValue string, rowNumber int, reason string) *SuspiciousCoordinateNotice {
	context := map[string]interface{}{
		"filename":     filename,
		"fieldName":    fieldName,
		"coordValue":   coordValue,
		"csvRowNumber": rowNumber,
		"reason":       reason,
	}
	return &SuspiciousCoordinateNotice{
		BaseNotice: NewBaseNotice("suspicious_coordinate", WARNING, context),
	}
}

// InsufficientCoordinatePrecisionNotice is generated when coordinates have insufficient precision
type InsufficientCoordinatePrecisionNotice struct {
	*BaseNotice
}

func NewInsufficientCoordinatePrecisionNotice(filename string, fieldName string, coordValue string, rowNumber int, decimals int) *InsufficientCoordinatePrecisionNotice {
	context := map[string]interface{}{
		"filename":     filename,
		"fieldName":    fieldName,
		"coordValue":   coordValue,
		"csvRowNumber": rowNumber,
		"decimals":     decimals,
	}
	return &InsufficientCoordinatePrecisionNotice{
		BaseNotice: NewBaseNotice("insufficient_coordinate_precision", WARNING, context),
	}
}

// InvalidCurrencyCodeNotice is generated when a currency code is invalid
type InvalidCurrencyCodeNotice struct {
	*BaseNotice
}

func NewInvalidCurrencyCodeNotice(filename string, fieldName string, currencyCode string, rowNumber int, reason string) *InvalidCurrencyCodeNotice {
	context := map[string]interface{}{
		"filename":     filename,
		"fieldName":    fieldName,
		"currencyCode": currencyCode,
		"csvRowNumber": rowNumber,
		"reason":       reason,
	}
	return &InvalidCurrencyCodeNotice{
		BaseNotice: NewBaseNotice("invalid_currency_code", ERROR, context),
	}
}

// MissingAgencyIdNotice is generated when agency_id is required but missing
type MissingAgencyIdNotice struct {
	*BaseNotice
}

func NewMissingAgencyIdNotice(agencyName string, rowNumber int) *MissingAgencyIdNotice {
	context := map[string]interface{}{
		"agencyName":   agencyName,
		"csvRowNumber": rowNumber,
		"message":      "agency_id is required when multiple agencies exist",
	}
	return &MissingAgencyIdNotice{
		BaseNotice: NewBaseNotice("missing_agency_id", ERROR, context),
	}
}

// InvalidAgencyReferenceNotice is generated when a route references an invalid agency
type InvalidAgencyReferenceNotice struct {
	*BaseNotice
}

func NewInvalidAgencyReferenceNotice(routeID string, agencyID string, rowNumber int) *InvalidAgencyReferenceNotice {
	context := map[string]interface{}{
		"routeId":      routeID,
		"agencyId":     agencyID,
		"csvRowNumber": rowNumber,
	}
	return &InvalidAgencyReferenceNotice{
		BaseNotice: NewBaseNotice("invalid_agency_reference", ERROR, context),
	}
}

// MissingRouteAgencyIdNotice is generated when route agency_id is required but missing
type MissingRouteAgencyIdNotice struct {
	*BaseNotice
}

func NewMissingRouteAgencyIdNotice(routeID string, rowNumber int) *MissingRouteAgencyIdNotice {
	context := map[string]interface{}{
		"routeId":      routeID,
		"csvRowNumber": rowNumber,
		"message":      "agency_id is required for routes when multiple agencies exist",
	}
	return &MissingRouteAgencyIdNotice{
		BaseNotice: NewBaseNotice("missing_route_agency_id", ERROR, context),
	}
}

// InvalidRouteTypeNotice is generated when route_type is invalid
type InvalidRouteTypeNotice struct {
	*BaseNotice
}

func NewInvalidRouteTypeNotice(routeID string, routeType string, rowNumber int, reason string) *InvalidRouteTypeNotice {
	context := map[string]interface{}{
		"routeId":      routeID,
		"routeType":    routeType,
		"csvRowNumber": rowNumber,
		"reason":       reason,
	}
	return &InvalidRouteTypeNotice{
		BaseNotice: NewBaseNotice("invalid_route_type", ERROR, context),
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

// PoorColorContrastNotice is generated when route colors have poor contrast
type PoorColorContrastNotice struct {
	*BaseNotice
}

func NewPoorColorContrastNotice(routeID string, routeColor string, routeTextColor string, rowNumber int) *PoorColorContrastNotice {
	context := map[string]interface{}{
		"routeId":        routeID,
		"routeColor":     routeColor,
		"routeTextColor": routeTextColor,
		"csvRowNumber":   rowNumber,
	}
	return &PoorColorContrastNotice{
		BaseNotice: NewBaseNotice("poor_color_contrast", WARNING, context),
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
		BaseNotice: NewBaseNotice("service_without_active_days", ERROR, context),
	}
}

// InvalidServiceDateRangeNotice is generated when start_date > end_date
type InvalidServiceDateRangeNotice struct {
	*BaseNotice
}

func NewInvalidServiceDateRangeNotice(serviceID string, startDate string, endDate string, rowNumber int) *InvalidServiceDateRangeNotice {
	context := map[string]interface{}{
		"serviceId":    serviceID,
		"startDate":    startDate,
		"endDate":      endDate,
		"csvRowNumber": rowNumber,
	}
	return &InvalidServiceDateRangeNotice{
		BaseNotice: NewBaseNotice("invalid_service_date_range", ERROR, context),
	}
}

// ExpiredServiceNotice is generated when a service is expired
type ExpiredServiceNotice struct {
	*BaseNotice
}

func NewExpiredServiceNotice(serviceID string, endDate string, rowNumber int) *ExpiredServiceNotice {
	context := map[string]interface{}{
		"serviceId":    serviceID,
		"endDate":      endDate,
		"csvRowNumber": rowNumber,
	}
	return &ExpiredServiceNotice{
		BaseNotice: NewBaseNotice("expired_service", WARNING, context),
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

// InvalidLocationTypeNotice is generated when location_type is invalid
type InvalidLocationTypeNotice struct {
	*BaseNotice
}

func NewInvalidLocationTypeNotice(stopID string, locationType int, rowNumber int) *InvalidLocationTypeNotice {
	context := map[string]interface{}{
		"stopId":       stopID,
		"locationType": locationType,
		"csvRowNumber": rowNumber,
	}
	return &InvalidLocationTypeNotice{
		BaseNotice: NewBaseNotice("invalid_location_type", ERROR, context),
	}
}

// MissingCoordinatesNotice is generated when coordinates are required but missing
type MissingCoordinatesNotice struct {
	*BaseNotice
}

func NewMissingCoordinatesNotice(stopID string, locationType int, rowNumber int) *MissingCoordinatesNotice {
	context := map[string]interface{}{
		"stopId":       stopID,
		"locationType": locationType,
		"csvRowNumber": rowNumber,
	}
	return &MissingCoordinatesNotice{
		BaseNotice: NewBaseNotice("missing_coordinates", ERROR, context),
	}
}

// InvalidParentStationReferenceNotice is generated when parent_station reference is invalid
type InvalidParentStationReferenceNotice struct {
	*BaseNotice
}

func NewInvalidParentStationReferenceNotice(stopID string, parentStation string, rowNumber int) *InvalidParentStationReferenceNotice {
	context := map[string]interface{}{
		"stopId":        stopID,
		"parentStation": parentStation,
		"csvRowNumber":  rowNumber,
	}
	return &InvalidParentStationReferenceNotice{
		BaseNotice: NewBaseNotice("invalid_parent_station_reference", ERROR, context),
	}
}

// InvalidParentStationTypeNotice is generated when parent station has wrong location type
type InvalidParentStationTypeNotice struct {
	*BaseNotice
}

func NewInvalidParentStationTypeNotice(stopID string, parentStation string, parentLocationType int, rowNumber int) *InvalidParentStationTypeNotice {
	context := map[string]interface{}{
		"stopId":             stopID,
		"parentStation":      parentStation,
		"parentLocationType": parentLocationType,
		"csvRowNumber":       rowNumber,
	}
	return &InvalidParentStationTypeNotice{
		BaseNotice: NewBaseNotice("invalid_parent_station_type", ERROR, context),
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

// MissingParentStationNotice is generated when parent station is required but missing
type MissingParentStationNotice struct {
	*BaseNotice
}

func NewMissingParentStationNotice(stopID string, locationType int, rowNumber int) *MissingParentStationNotice {
	context := map[string]interface{}{
		"stopId":       stopID,
		"locationType": locationType,
		"csvRowNumber": rowNumber,
	}
	return &MissingParentStationNotice{
		BaseNotice: NewBaseNotice("missing_parent_station", ERROR, context),
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
		BaseNotice: NewBaseNotice("circular_station_reference", ERROR, context),
	}
}

// OrphanedStationNotice is generated when a station has no child stops
type OrphanedStationNotice struct {
	*BaseNotice
}

func NewOrphanedStationNotice(stationID string, rowNumber int) *OrphanedStationNotice {
	context := map[string]interface{}{
		"stationId":    stationID,
		"csvRowNumber": rowNumber,
	}
	return &OrphanedStationNotice{
		BaseNotice: NewBaseNotice("orphaned_station", WARNING, context),
	}
}

// InvalidFrequencyTimeRangeNotice is generated when frequency time range is invalid
type InvalidFrequencyTimeRangeNotice struct {
	*BaseNotice
}

func NewInvalidFrequencyTimeRangeNotice(tripID string, startTime string, endTime string, rowNumber int) *InvalidFrequencyTimeRangeNotice {
	context := map[string]interface{}{
		"tripId":       tripID,
		"startTime":    startTime,
		"endTime":      endTime,
		"csvRowNumber": rowNumber,
	}
	return &InvalidFrequencyTimeRangeNotice{
		BaseNotice: NewBaseNotice("invalid_frequency_time_range", ERROR, context),
	}
}

// InvalidHeadwayNotice is generated when headway_secs is invalid
type InvalidHeadwayNotice struct {
	*BaseNotice
}

func NewInvalidHeadwayNotice(tripID string, headwaySecs int, rowNumber int) *InvalidHeadwayNotice {
	context := map[string]interface{}{
		"tripId":       tripID,
		"headwaySecs":  headwaySecs,
		"csvRowNumber": rowNumber,
	}
	return &InvalidHeadwayNotice{
		BaseNotice: NewBaseNotice("invalid_headway", ERROR, context),
	}
}

// UnreasonableHeadwayNotice is generated when headway is unreasonable
type UnreasonableHeadwayNotice struct {
	*BaseNotice
}

func NewUnreasonableHeadwayNotice(tripID string, headwaySecs int, rowNumber int) *UnreasonableHeadwayNotice {
	context := map[string]interface{}{
		"tripId":       tripID,
		"headwaySecs":  headwaySecs,
		"csvRowNumber": rowNumber,
	}
	return &UnreasonableHeadwayNotice{
		BaseNotice: NewBaseNotice("unreasonable_headway", WARNING, context),
	}
}

// InvalidExactTimesNotice is generated when exact_times field is invalid
type InvalidExactTimesNotice struct {
	*BaseNotice
}

func NewInvalidExactTimesNotice(tripID string, exactTimes int, rowNumber int) *InvalidExactTimesNotice {
	context := map[string]interface{}{
		"tripId":       tripID,
		"exactTimes":   exactTimes,
		"csvRowNumber": rowNumber,
	}
	return &InvalidExactTimesNotice{
		BaseNotice: NewBaseNotice("invalid_exact_times", ERROR, context),
	}
}

// OverlappingFrequencyNotice is generated when frequencies overlap for the same trip
type OverlappingFrequencyNotice struct {
	*BaseNotice
}

func NewOverlappingFrequencyNotice(tripID string, startTime1 string, endTime1 string, rowNumber1 int, startTime2 string, endTime2 string, rowNumber2 int) *OverlappingFrequencyNotice {
	context := map[string]interface{}{
		"tripId":       tripID,
		"startTime1":   startTime1,
		"endTime1":     endTime1,
		"rowNumber1":   rowNumber1,
		"startTime2":   startTime2,
		"endTime2":     endTime2,
		"rowNumber2":   rowNumber2,
	}
	return &OverlappingFrequencyNotice{
		BaseNotice: NewBaseNotice("overlapping_frequency", ERROR, context),
	}
}

// InvalidTransferTypeNotice is generated when transfer_type is invalid
type InvalidTransferTypeNotice struct {
	*BaseNotice
}

func NewInvalidTransferTypeNotice(fromStopID string, toStopID string, transferType int, rowNumber int) *InvalidTransferTypeNotice {
	context := map[string]interface{}{
		"fromStopId":   fromStopID,
		"toStopId":     toStopID,
		"transferType": transferType,
		"csvRowNumber": rowNumber,
	}
	return &InvalidTransferTypeNotice{
		BaseNotice: NewBaseNotice("invalid_transfer_type", ERROR, context),
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
		BaseNotice: NewBaseNotice("missing_min_transfer_time", ERROR, context),
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
		BaseNotice: NewBaseNotice("negative_min_transfer_time", ERROR, context),
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
		BaseNotice: NewBaseNotice("duplicate_transfer", ERROR, context),
	}
}

// PATHWAY VALIDATOR NOTICES

// InvalidPathwayModeNotice is generated when pathway_mode has invalid value
type InvalidPathwayModeNotice struct {
	*BaseNotice
}

func NewInvalidPathwayModeNotice(pathwayID string, pathwayMode int, rowNumber int) *InvalidPathwayModeNotice {
	context := map[string]interface{}{
		"pathwayId":    pathwayID,
		"pathwayMode":  pathwayMode,
		"csvRowNumber": rowNumber,
	}
	return &InvalidPathwayModeNotice{
		BaseNotice: NewBaseNotice("invalid_pathway_mode", ERROR, context),
	}
}

// InvalidBidirectionalNotice is generated when is_bidirectional has invalid value
type InvalidBidirectionalNotice struct {
	*BaseNotice
}

func NewInvalidBidirectionalNotice(pathwayID string, isBidirectional int, rowNumber int) *InvalidBidirectionalNotice {
	context := map[string]interface{}{
		"pathwayId":       pathwayID,
		"isBidirectional": isBidirectional,
		"csvRowNumber":    rowNumber,
	}
	return &InvalidBidirectionalNotice{
		BaseNotice: NewBaseNotice("invalid_bidirectional", ERROR, context),
	}
}

// PathwayToSameStopNotice is generated when pathway connects stop to itself
type PathwayToSameStopNotice struct {
	*BaseNotice
}

func NewPathwayToSameStopNotice(pathwayID string, stopID string, rowNumber int) *PathwayToSameStopNotice {
	context := map[string]interface{}{
		"pathwayId":    pathwayID,
		"stopId":       stopID,
		"csvRowNumber": rowNumber,
	}
	return &PathwayToSameStopNotice{
		BaseNotice: NewBaseNotice("pathway_to_same_stop", ERROR, context),
	}
}

// InvalidStairCountNotice is generated when stair_count is invalid
type InvalidStairCountNotice struct {
	*BaseNotice
}

func NewInvalidStairCountNotice(pathwayID string, stairCount int, rowNumber int) *InvalidStairCountNotice {
	context := map[string]interface{}{
		"pathwayId":    pathwayID,
		"stairCount":   stairCount,
		"csvRowNumber": rowNumber,
	}
	return &InvalidStairCountNotice{
		BaseNotice: NewBaseNotice("invalid_stair_count", ERROR, context),
	}
}

// UnexpectedBidirectionalGateNotice is generated when gate is marked as bidirectional
type UnexpectedBidirectionalGateNotice struct {
	*BaseNotice
}

func NewUnexpectedBidirectionalGateNotice(pathwayID string, pathwayMode int, rowNumber int) *UnexpectedBidirectionalGateNotice {
	context := map[string]interface{}{
		"pathwayId":    pathwayID,
		"pathwayMode":  pathwayMode,
		"csvRowNumber": rowNumber,
	}
	return &UnexpectedBidirectionalGateNotice{
		BaseNotice: NewBaseNotice("unexpected_bidirectional_gate", WARNING, context),
	}
}

// InvalidPathwayLengthNotice is generated when pathway length is invalid
type InvalidPathwayLengthNotice struct {
	*BaseNotice
}

func NewInvalidPathwayLengthNotice(pathwayID string, length float64, rowNumber int) *InvalidPathwayLengthNotice {
	context := map[string]interface{}{
		"pathwayId":    pathwayID,
		"length":       length,
		"csvRowNumber": rowNumber,
	}
	return &InvalidPathwayLengthNotice{
		BaseNotice: NewBaseNotice("invalid_pathway_length", ERROR, context),
	}
}

// InvalidTraversalTimeNotice is generated when traversal_time is invalid
type InvalidTraversalTimeNotice struct {
	*BaseNotice
}

func NewInvalidTraversalTimeNotice(pathwayID string, traversalTime int, rowNumber int) *InvalidTraversalTimeNotice {
	context := map[string]interface{}{
		"pathwayId":      pathwayID,
		"traversalTime":  traversalTime,
		"csvRowNumber":   rowNumber,
	}
	return &InvalidTraversalTimeNotice{
		BaseNotice: NewBaseNotice("invalid_traversal_time", ERROR, context),
	}
}

// UnreasonableMaxSlopeNotice is generated when max_slope is unreasonable
type UnreasonableMaxSlopeNotice struct {
	*BaseNotice
}

func NewUnreasonableMaxSlopeNotice(pathwayID string, maxSlope float64, rowNumber int) *UnreasonableMaxSlopeNotice {
	context := map[string]interface{}{
		"pathwayId":    pathwayID,
		"maxSlope":     maxSlope,
		"csvRowNumber": rowNumber,
	}
	return &UnreasonableMaxSlopeNotice{
		BaseNotice: NewBaseNotice("unreasonable_max_slope", WARNING, context),
	}
}

// InvalidMinWidthNotice is generated when min_width is invalid
type InvalidMinWidthNotice struct {
	*BaseNotice
}

func NewInvalidMinWidthNotice(pathwayID string, minWidth float64, rowNumber int) *InvalidMinWidthNotice {
	context := map[string]interface{}{
		"pathwayId":    pathwayID,
		"minWidth":     minWidth,
		"csvRowNumber": rowNumber,
	}
	return &InvalidMinWidthNotice{
		BaseNotice: NewBaseNotice("invalid_min_width", ERROR, context),
	}
}

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

// InvalidPaymentMethodNotice is generated when payment_method has invalid value
type InvalidPaymentMethodNotice struct {
	*BaseNotice
}

func NewInvalidPaymentMethodNotice(fareID string, paymentMethod int, rowNumber int) *InvalidPaymentMethodNotice {
	context := map[string]interface{}{
		"fareId":        fareID,
		"paymentMethod": paymentMethod,
		"csvRowNumber":  rowNumber,
	}
	return &InvalidPaymentMethodNotice{
		BaseNotice: NewBaseNotice("invalid_payment_method", ERROR, context),
	}
}

// InvalidTransfersNotice is generated when transfers field has invalid value
type InvalidTransfersNotice struct {
	*BaseNotice
}

func NewInvalidTransfersNotice(fareID string, transfers int, rowNumber int) *InvalidTransfersNotice {
	context := map[string]interface{}{
		"fareId":       fareID,
		"transfers":    transfers,
		"csvRowNumber": rowNumber,
	}
	return &InvalidTransfersNotice{
		BaseNotice: NewBaseNotice("invalid_transfers", WARNING, context),
	}
}

// UnusualTransferValueNotice is generated when transfers has unusual value
type UnusualTransferValueNotice struct {
	*BaseNotice
}

func NewUnusualTransferValueNotice(fareID string, transfers int, rowNumber int) *UnusualTransferValueNotice {
	context := map[string]interface{}{
		"fareId":       fareID,
		"transfers":    transfers,
		"csvRowNumber": rowNumber,
	}
	return &UnusualTransferValueNotice{
		BaseNotice: NewBaseNotice("unusual_transfer_value", WARNING, context),
	}
}

// InvalidTransferDurationNotice is generated when transfer_duration is invalid
type InvalidTransferDurationNotice struct {
	*BaseNotice
}

func NewInvalidTransferDurationNotice(fareID string, transferDuration int, rowNumber int) *InvalidTransferDurationNotice {
	context := map[string]interface{}{
		"fareId":           fareID,
		"transferDuration": transferDuration,
		"csvRowNumber":     rowNumber,
	}
	return &InvalidTransferDurationNotice{
		BaseNotice: NewBaseNotice("invalid_transfer_duration", ERROR, context),
	}
}

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

// InvalidFarePriceNotice is generated when fare price is invalid
type InvalidFarePriceNotice struct {
	*BaseNotice
}

func NewInvalidFarePriceNotice(fareID string, price string, rowNumber int, reason string) *InvalidFarePriceNotice {
	context := map[string]interface{}{
		"fareId":       fareID,
		"price":        price,
		"csvRowNumber": rowNumber,
		"reason":       reason,
	}
	return &InvalidFarePriceNotice{
		BaseNotice: NewBaseNotice("invalid_fare_price", ERROR, context),
	}
}

// ExcessivePricePrecisionNotice is generated when fare price has too many decimal places
type ExcessivePricePrecisionNotice struct {
	*BaseNotice
}

func NewExcessivePricePrecisionNotice(fareID string, price string, decimals int, rowNumber int) *ExcessivePricePrecisionNotice {
	context := map[string]interface{}{
		"fareId":       fareID,
		"price":        price,
		"decimals":     decimals,
		"csvRowNumber": rowNumber,
	}
	return &ExcessivePricePrecisionNotice{
		BaseNotice: NewBaseNotice("excessive_price_precision", WARNING, context),
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

// SameOriginDestinationNotice is generated when origin and destination are the same
type SameOriginDestinationNotice struct {
	*BaseNotice
}

func NewSameOriginDestinationNotice(fareID string, zoneID string, rowNumber int) *SameOriginDestinationNotice {
	context := map[string]interface{}{
		"fareId":       fareID,
		"zoneId":       zoneID,
		"csvRowNumber": rowNumber,
	}
	return &SameOriginDestinationNotice{
		BaseNotice: NewBaseNotice("same_origin_destination", WARNING, context),
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

// UnreasonableLevelIndexNotice is generated when level_index is unreasonable
type UnreasonableLevelIndexNotice struct {
	*BaseNotice
}

func NewUnreasonableLevelIndexNotice(levelID string, levelIndex float64, rowNumber int) *UnreasonableLevelIndexNotice {
	context := map[string]interface{}{
		"levelId":      levelID,
		"levelIndex":   levelIndex,
		"csvRowNumber": rowNumber,
	}
	return &UnreasonableLevelIndexNotice{
		BaseNotice: NewBaseNotice("unreasonable_level_index", WARNING, context),
	}
}

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
		BaseNotice: NewBaseNotice("duplicate_level_index", ERROR, context),
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
		BaseNotice: NewBaseNotice("insufficient_shape_points", ERROR, context),
	}
}


// NonIncreasingShapeSequenceNotice is generated when shape sequences don't increase
type NonIncreasingShapeSequenceNotice struct {
	*BaseNotice
}

func NewNonIncreasingShapeSequenceNotice(shapeID string, currentSequence int, previousSequence int, rowNumber int) *NonIncreasingShapeSequenceNotice {
	context := map[string]interface{}{
		"shapeId":           shapeID,
		"currentSequence":   currentSequence,
		"previousSequence":  previousSequence,
		"csvRowNumber":      rowNumber,
	}
	return &NonIncreasingShapeSequenceNotice{
		BaseNotice: NewBaseNotice("non_increasing_shape_sequence", ERROR, context),
	}
}

// InconsistentShapeDistanceNotice is generated when shape distances are inconsistent
type InconsistentShapeDistanceNotice struct {
	*BaseNotice
}

func NewInconsistentShapeDistanceNotice(shapeID string, sequence int, rowNumber int) *InconsistentShapeDistanceNotice {
	context := map[string]interface{}{
		"shapeId":         shapeID,
		"shapePtSequence": sequence,
		"csvRowNumber":    rowNumber,
	}
	return &InconsistentShapeDistanceNotice{
		BaseNotice: NewBaseNotice("inconsistent_shape_distance", WARNING, context),
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

// EqualShapeDistanceNotice is generated when consecutive shape points have equal distances
type EqualShapeDistanceNotice struct {
	*BaseNotice
}

func NewEqualShapeDistanceNotice(shapeID string, currentSequence int, previousSequence int, distance float64, rowNumber int) *EqualShapeDistanceNotice {
	context := map[string]interface{}{
		"shapeId":           shapeID,
		"currentSequence":   currentSequence,
		"previousSequence":  previousSequence,
		"distance":          distance,
		"csvRowNumber":      rowNumber,
	}
	return &EqualShapeDistanceNotice{
		BaseNotice: NewBaseNotice("equal_shape_distance", WARNING, context),
	}
}

// DuplicateShapePointNotice is generated when consecutive shape points are identical
type DuplicateShapePointNotice struct {
	*BaseNotice
}

func NewDuplicateShapePointNotice(shapeID string, currentSequence int, previousSequence int, rowNumber int) *DuplicateShapePointNotice {
	context := map[string]interface{}{
		"shapeId":           shapeID,
		"currentSequence":   currentSequence,
		"previousSequence":  previousSequence,
		"csvRowNumber":      rowNumber,
	}
	return &DuplicateShapePointNotice{
		BaseNotice: NewBaseNotice("duplicate_shape_point", WARNING, context),
	}
}

// UnreasonablyLongShapeSegmentNotice is generated when shape segment is unreasonably long
type UnreasonablyLongShapeSegmentNotice struct {
	*BaseNotice
}

func NewUnreasonablyLongShapeSegmentNotice(shapeID string, fromSequence int, toSequence int, distance float64, rowNumber int) *UnreasonablyLongShapeSegmentNotice {
	context := map[string]interface{}{
		"shapeId":      shapeID,
		"fromSequence": fromSequence,
		"toSequence":   toSequence,
		"distance":     distance,
		"csvRowNumber": rowNumber,
	}
	return &UnreasonablyLongShapeSegmentNotice{
		BaseNotice: NewBaseNotice("unreasonably_long_shape_segment", WARNING, context),
	}
}

// UnusedShapeNotice is generated when shape is never used
type UnusedShapeNotice struct {
	*BaseNotice
}

func NewUnusedShapeNotice(shapeID string) *UnusedShapeNotice {
	context := map[string]interface{}{
		"shapeId": shapeID,
	}
	return &UnusedShapeNotice{
		BaseNotice: NewBaseNotice("unused_shape", WARNING, context),
	}
}

// CALENDAR CONSISTENCY VALIDATOR NOTICES

// ServiceNeverActiveNotice is generated when service runs on no days
type ServiceNeverActiveNotice struct {
	*BaseNotice
}

func NewServiceNeverActiveNotice(serviceID string, rowNumber int) *ServiceNeverActiveNotice {
	context := map[string]interface{}{
		"serviceId":    serviceID,
		"csvRowNumber": rowNumber,
	}
	return &ServiceNeverActiveNotice{
		BaseNotice: NewBaseNotice("service_never_active", ERROR, context),
	}
}



// FutureServiceNotice is generated when service period is too far in the future
type FutureServiceNotice struct {
	*BaseNotice
}

func NewFutureServiceNotice(serviceID string, startDate string, rowNumber int) *FutureServiceNotice {
	context := map[string]interface{}{
		"serviceId":    serviceID,
		"startDate":    startDate,
		"csvRowNumber": rowNumber,
	}
	return &FutureServiceNotice{
		BaseNotice: NewBaseNotice("future_service", WARNING, context),
	}
}

// InvalidExceptionTypeNotice is generated when exception_type has invalid value
type InvalidExceptionTypeNotice struct {
	*BaseNotice
}

func NewInvalidExceptionTypeNotice(serviceID string, date string, exceptionType int, rowNumber int) *InvalidExceptionTypeNotice {
	context := map[string]interface{}{
		"serviceId":     serviceID,
		"date":          date,
		"exceptionType": exceptionType,
		"csvRowNumber":  rowNumber,
	}
	return &InvalidExceptionTypeNotice{
		BaseNotice: NewBaseNotice("invalid_exception_type", ERROR, context),
	}
}

// VeryOldCalendarDateNotice is generated when calendar date is very old
type VeryOldCalendarDateNotice struct {
	*BaseNotice
}

func NewVeryOldCalendarDateNotice(serviceID string, date string, rowNumber int) *VeryOldCalendarDateNotice {
	context := map[string]interface{}{
		"serviceId":    serviceID,
		"date":         date,
		"csvRowNumber": rowNumber,
	}
	return &VeryOldCalendarDateNotice{
		BaseNotice: NewBaseNotice("very_old_calendar_date", WARNING, context),
	}
}

// VeryFutureCalendarDateNotice is generated when calendar date is very far in future
type VeryFutureCalendarDateNotice struct {
	*BaseNotice
}

func NewVeryFutureCalendarDateNotice(serviceID string, date string, rowNumber int) *VeryFutureCalendarDateNotice {
	context := map[string]interface{}{
		"serviceId":    serviceID,
		"date":         date,
		"csvRowNumber": rowNumber,
	}
	return &VeryFutureCalendarDateNotice{
		BaseNotice: NewBaseNotice("very_future_calendar_date", WARNING, context),
	}
}

// UndefinedServiceNotice is generated when service is used but not defined
type UndefinedServiceNotice struct {
	*BaseNotice
}

func NewUndefinedServiceNotice(serviceID string) *UndefinedServiceNotice {
	context := map[string]interface{}{
		"serviceId": serviceID,
	}
	return &UndefinedServiceNotice{
		BaseNotice: NewBaseNotice("undefined_service", ERROR, context),
	}
}


// ConflictingCalendarExceptionNotice is generated when conflicting exceptions exist
type ConflictingCalendarExceptionNotice struct {
	*BaseNotice
}

func NewConflictingCalendarExceptionNotice(serviceID string, date string, rowNumber1 int, rowNumber2 int) *ConflictingCalendarExceptionNotice {
	context := map[string]interface{}{
		"serviceId":      serviceID,
		"date":           date,
		"csvRowNumber1":  rowNumber1,
		"csvRowNumber2":  rowNumber2,
	}
	return &ConflictingCalendarExceptionNotice{
		BaseNotice: NewBaseNotice("conflicting_calendar_exception", ERROR, context),
	}
}

// DuplicateCalendarExceptionNotice is generated when duplicate exceptions exist
type DuplicateCalendarExceptionNotice struct {
	*BaseNotice
}

func NewDuplicateCalendarExceptionNotice(serviceID string, date string, rowNumber1 int, rowNumber2 int) *DuplicateCalendarExceptionNotice {
	context := map[string]interface{}{
		"serviceId":      serviceID,
		"date":           date,
		"csvRowNumber1":  rowNumber1,
		"csvRowNumber2":  rowNumber2,
	}
	return &DuplicateCalendarExceptionNotice{
		BaseNotice: NewBaseNotice("duplicate_calendar_exception", WARNING, context),
	}
}

// ATTRIBUTION VALIDATOR NOTICES

// MissingAttributionRoleNotice is generated when no attribution role is specified
type MissingAttributionRoleNotice struct {
	*BaseNotice
}

func NewMissingAttributionRoleNotice(attributionID string, rowNumber int) *MissingAttributionRoleNotice {
	context := map[string]interface{}{
		"attributionId": attributionID,
		"csvRowNumber":  rowNumber,
	}
	return &MissingAttributionRoleNotice{
		BaseNotice: NewBaseNotice("missing_attribution_role", ERROR, context),
	}
}

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
		BaseNotice: NewBaseNotice("conflicting_attribution_scope", ERROR, context),
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

// MultipleFeedInfoEntriesNotice is generated when multiple feed info entries exist
type MultipleFeedInfoEntriesNotice struct {
	*BaseNotice
}

func NewMultipleFeedInfoEntriesNotice(count int) *MultipleFeedInfoEntriesNotice {
	context := map[string]interface{}{
		"count": count,
	}
	return &MultipleFeedInfoEntriesNotice{
		BaseNotice: NewBaseNotice("multiple_feed_info_entries", ERROR, context),
	}
}

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
		BaseNotice: NewBaseNotice("invalid_language_code", WARNING, context),
	}
}

// FeedInfoEndDateBeforeStartDateNotice is generated when feed end date is before start date
type FeedInfoEndDateBeforeStartDateNotice struct {
	*BaseNotice
}

func NewFeedInfoEndDateBeforeStartDateNotice(startDate string, endDate string, rowNumber int) *FeedInfoEndDateBeforeStartDateNotice {
	context := map[string]interface{}{
		"startDate":    startDate,
		"endDate":      endDate,
		"csvRowNumber": rowNumber,
	}
	return &FeedInfoEndDateBeforeStartDateNotice{
		BaseNotice: NewBaseNotice("feed_info_end_date_before_start_date", ERROR, context),
	}
}

// ExpiredFeedNotice is generated when feed has expired
type ExpiredFeedNotice struct {
	*BaseNotice
}

func NewExpiredFeedNotice(endDate string, rowNumber int) *ExpiredFeedNotice {
	context := map[string]interface{}{
		"endDate":      endDate,
		"csvRowNumber": rowNumber,
	}
	return &ExpiredFeedNotice{
		BaseNotice: NewBaseNotice("expired_feed", WARNING, context),
	}
}

// FutureFeedStartDateNotice is generated when feed start date is too far in future
type FutureFeedStartDateNotice struct {
	*BaseNotice
}

func NewFutureFeedStartDateNotice(startDate string, rowNumber int) *FutureFeedStartDateNotice {
	context := map[string]interface{}{
		"startDate":    startDate,
		"csvRowNumber": rowNumber,
	}
	return &FutureFeedStartDateNotice{
		BaseNotice: NewBaseNotice("future_feed_start_date", WARNING, context),
	}
}

// ZONE VALIDATOR NOTICES

// SingleStopZoneNotice is generated when zone has only one stop
type SingleStopZoneNotice struct {
	*BaseNotice
}

func NewSingleStopZoneNotice(zoneID string, stopID string, rowNumber int) *SingleStopZoneNotice {
	context := map[string]interface{}{
		"zoneId":       zoneID,
		"stopId":       stopID,
		"csvRowNumber": rowNumber,
	}
	return &SingleStopZoneNotice{
		BaseNotice: NewBaseNotice("single_stop_zone", WARNING, context),
	}
}

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

// UndefinedZoneNotice is generated when zone is used but not defined
type UndefinedZoneNotice struct {
	*BaseNotice
}

func NewUndefinedZoneNotice(zoneID string) *UndefinedZoneNotice {
	context := map[string]interface{}{
		"zoneId": zoneID,
	}
	return &UndefinedZoneNotice{
		BaseNotice: NewBaseNotice("undefined_zone", ERROR, context),
	}
}

// LongZoneIDNotice is generated when zone ID is very long
type LongZoneIDNotice struct {
	*BaseNotice
}

func NewLongZoneIDNotice(zoneID string, length int, rowNumber int) *LongZoneIDNotice {
	context := map[string]interface{}{
		"zoneId":       zoneID,
		"length":       length,
		"csvRowNumber": rowNumber,
	}
	return &LongZoneIDNotice{
		BaseNotice: NewBaseNotice("long_zone_id", WARNING, context),
	}
}

// ZoneIDSameAsStopIDNotice is generated when zone ID equals stop ID
type ZoneIDSameAsStopIDNotice struct {
	*BaseNotice
}

func NewZoneIDSameAsStopIDNotice(zoneID string, rowNumber int) *ZoneIDSameAsStopIDNotice {
	context := map[string]interface{}{
		"zoneId":       zoneID,
		"csvRowNumber": rowNumber,
	}
	return &ZoneIDSameAsStopIDNotice{
		BaseNotice: NewBaseNotice("zone_id_same_as_stop_id", WARNING, context),
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
		BaseNotice: NewBaseNotice("missing_trip_first_time", ERROR, context),
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
		BaseNotice: NewBaseNotice("missing_trip_last_time", ERROR, context),
	}
}

// LoopRouteNotice is generated when trip starts and ends at same stop
type LoopRouteNotice struct {
	*BaseNotice
}

func NewLoopRouteNotice(tripID string, stopID string, firstRowNumber int, lastRowNumber int) *LoopRouteNotice {
	context := map[string]interface{}{
		"tripId":          tripID,
		"stopId":          stopID,
		"firstRowNumber":  firstRowNumber,
		"lastRowNumber":   lastRowNumber,
	}
	return &LoopRouteNotice{
		BaseNotice: NewBaseNotice("loop_route", INFO, context),
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

// MissingArrivalTimeNotice is generated when arrival time is missing but departure exists
type MissingArrivalTimeNotice struct {
	*BaseNotice
}

func NewMissingArrivalTimeNotice(tripID string, stopID string, stopSequence int, rowNumber int) *MissingArrivalTimeNotice {
	context := map[string]interface{}{
		"tripId":       tripID,
		"stopId":       stopID,
		"stopSequence": stopSequence,
		"csvRowNumber": rowNumber,
	}
	return &MissingArrivalTimeNotice{
		BaseNotice: NewBaseNotice("missing_arrival_time", WARNING, context),
	}
}

// MissingDepartureTimeNotice is generated when departure time is missing but arrival exists
type MissingDepartureTimeNotice struct {
	*BaseNotice
}

func NewMissingDepartureTimeNotice(tripID string, stopID string, stopSequence int, rowNumber int) *MissingDepartureTimeNotice {
	context := map[string]interface{}{
		"tripId":       tripID,
		"stopId":       stopID,
		"stopSequence": stopSequence,
		"csvRowNumber": rowNumber,
	}
	return &MissingDepartureTimeNotice{
		BaseNotice: NewBaseNotice("missing_departure_time", WARNING, context),
	}
}

// InvalidTimepointNotice is generated when timepoint has invalid value
type InvalidTimepointNotice struct {
	*BaseNotice
}

func NewInvalidTimepointNotice(tripID string, stopID string, timepoint int, rowNumber int) *InvalidTimepointNotice {
	context := map[string]interface{}{
		"tripId":       tripID,
		"stopId":       stopID,
		"timepoint":    timepoint,
		"csvRowNumber": rowNumber,
	}
	return &InvalidTimepointNotice{
		BaseNotice: NewBaseNotice("invalid_timepoint", ERROR, context),
	}
}

// TimepointWithoutTimesNotice is generated when timepoint=0 but times exist
type TimepointWithoutTimesNotice struct {
	*BaseNotice
}

func NewTimepointWithoutTimesNotice(tripID string, stopID string, stopSequence int, rowNumber int) *TimepointWithoutTimesNotice {
	context := map[string]interface{}{
		"tripId":       tripID,
		"stopId":       stopID,
		"stopSequence": stopSequence,
		"csvRowNumber": rowNumber,
	}
	return &TimepointWithoutTimesNotice{
		BaseNotice: NewBaseNotice("timepoint_without_times", INFO, context),
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
		BaseNotice: NewBaseNotice("all_stops_no_pickup", ERROR, context),
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
		BaseNotice: NewBaseNotice("all_stops_no_drop_off", ERROR, context),
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
		BaseNotice: NewBaseNotice("duplicate_composite_key", ERROR, context),
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
		BaseNotice: NewBaseNotice("multiple_records_in_single_record_file", ERROR, context),
	}
}

// InvalidRowNotice represents an invalid row structure
type InvalidRowNotice struct {
	*BaseNotice
}

func NewInvalidRowNotice(filename string, rowNumber int, reason string) *InvalidRowNotice {
	context := map[string]interface{}{
		"filename":  filename,
		"rowNumber": rowNumber,
		"reason":    reason,
	}
	return &InvalidRowNotice{
		BaseNotice: NewBaseNotice("invalid_row", ERROR, context),
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
		BaseNotice: NewBaseNotice("wrong_number_of_fields", ERROR, context),
	}
}

// NegativeStopSequenceNotice represents negative stop_sequence
type NegativeStopSequenceNotice struct {
	*BaseNotice
}

func NewNegativeStopSequenceNotice(tripID string, stopSequence, rowNumber int) *NegativeStopSequenceNotice {
	context := map[string]interface{}{
		"tripId":       tripID,
		"stopSequence": stopSequence,
		"rowNumber":    rowNumber,
	}
	return &NegativeStopSequenceNotice{
		BaseNotice: NewBaseNotice("negative_stop_sequence", ERROR, context),
	}
}

// NegativeShapeDistanceNotice represents negative shape_dist_traveled
type NegativeShapeDistanceNotice struct {
	*BaseNotice
}

func NewNegativeShapeDistanceNotice(tripID string, shapeDistance float64, rowNumber int) *NegativeShapeDistanceNotice {
	context := map[string]interface{}{
		"tripId":        tripID,
		"shapeDistance": shapeDistance,
		"rowNumber":     rowNumber,
	}
	return &NegativeShapeDistanceNotice{
		BaseNotice: NewBaseNotice("negative_shape_distance", ERROR, context),
	}
}


// InvalidWheelchairBoardingNotice represents invalid wheelchair_boarding value
type InvalidWheelchairBoardingNotice struct {
	*BaseNotice
}

func NewInvalidWheelchairBoardingNotice(stopID string, wheelchairBoarding, rowNumber int) *InvalidWheelchairBoardingNotice {
	context := map[string]interface{}{
		"stopId":             stopID,
		"wheelchairBoarding": wheelchairBoarding,
		"rowNumber":          rowNumber,
	}
	return &InvalidWheelchairBoardingNotice{
		BaseNotice: NewBaseNotice("invalid_wheelchair_boarding", ERROR, context),
	}
}


// InvalidDirectionIdNotice represents invalid direction_id value
type InvalidDirectionIdNotice struct {
	*BaseNotice
}

func NewInvalidDirectionIdNotice(tripID string, directionId, rowNumber int) *InvalidDirectionIdNotice {
	context := map[string]interface{}{
		"tripId":      tripID,
		"directionId": directionId,
		"rowNumber":   rowNumber,
	}
	return &InvalidDirectionIdNotice{
		BaseNotice: NewBaseNotice("invalid_direction_id", ERROR, context),
	}
}

// InvalidWheelchairAccessibleNotice represents invalid wheelchair_accessible value
type InvalidWheelchairAccessibleNotice struct {
	*BaseNotice
}

func NewInvalidWheelchairAccessibleNotice(tripID string, wheelchairAccessible, rowNumber int) *InvalidWheelchairAccessibleNotice {
	context := map[string]interface{}{
		"tripId":               tripID,
		"wheelchairAccessible": wheelchairAccessible,
		"rowNumber":            rowNumber,
	}
	return &InvalidWheelchairAccessibleNotice{
		BaseNotice: NewBaseNotice("invalid_wheelchair_accessible", ERROR, context),
	}
}

// InvalidBikesAllowedNotice represents invalid bikes_allowed value
type InvalidBikesAllowedNotice struct {
	*BaseNotice
}

func NewInvalidBikesAllowedNotice(tripID string, bikesAllowed, rowNumber int) *InvalidBikesAllowedNotice {
	context := map[string]interface{}{
		"tripId":       tripID,
		"bikesAllowed": bikesAllowed,
		"rowNumber":    rowNumber,
	}
	return &InvalidBikesAllowedNotice{
		BaseNotice: NewBaseNotice("invalid_bikes_allowed", ERROR, context),
	}
}

// InvalidDayValueNotice represents invalid day field value in calendar.txt
type InvalidDayValueNotice struct {
	*BaseNotice
}

func NewInvalidDayValueNotice(serviceID, field, value string, rowNumber int) *InvalidDayValueNotice {
	context := map[string]interface{}{
		"serviceId": serviceID,
		"field":     field,
		"value":     value,
		"rowNumber": rowNumber,
	}
	return &InvalidDayValueNotice{
		BaseNotice: NewBaseNotice("invalid_day_value", ERROR, context),
	}
}

// NegativeShapeSequenceNotice represents negative shape_pt_sequence
type NegativeShapeSequenceNotice struct {
	*BaseNotice
}

func NewNegativeShapeSequenceNotice(shapeID string, sequence, rowNumber int) *NegativeShapeSequenceNotice {
	context := map[string]interface{}{
		"shapeId":   shapeID,
		"sequence":  sequence,
		"rowNumber": rowNumber,
	}
	return &NegativeShapeSequenceNotice{
		BaseNotice: NewBaseNotice("negative_shape_sequence", ERROR, context),
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
		BaseNotice: NewBaseNotice("leading_whitespace", WARNING, context),
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
		BaseNotice: NewBaseNotice("trailing_whitespace", WARNING, context),
	}
}

// WhitespaceOnlyFieldNotice represents a field containing only whitespace
type WhitespaceOnlyFieldNotice struct {
	*BaseNotice
}

func NewWhitespaceOnlyFieldNotice(filename, fieldName string, rowNumber int) *WhitespaceOnlyFieldNotice {
	context := map[string]interface{}{
		"filename":  filename,
		"fieldName": fieldName,
		"rowNumber": rowNumber,
	}
	return &WhitespaceOnlyFieldNotice{
		BaseNotice: NewBaseNotice("whitespace_only_field", WARNING, context),
	}
}

// ExcessiveWhitespaceNotice represents a field with excessive internal whitespace
type ExcessiveWhitespaceNotice struct {
	*BaseNotice
}

func NewExcessiveWhitespaceNotice(filename, fieldName, fieldValue string, rowNumber int) *ExcessiveWhitespaceNotice {
	context := map[string]interface{}{
		"filename":   filename,
		"fieldName":  fieldName,
		"fieldValue": fieldValue,
		"rowNumber":  rowNumber,
	}
	return &ExcessiveWhitespaceNotice{
		BaseNotice: NewBaseNotice("excessive_whitespace", INFO, context),
	}
}

// InsufficientStopTimesNotice represents a trip with insufficient stop times
type InsufficientStopTimesNotice struct {
	*BaseNotice
}

func NewInsufficientStopTimesNotice(tripID string, stopCount int) *InsufficientStopTimesNotice {
	context := map[string]interface{}{
		"tripId":    tripID,
		"stopCount": stopCount,
	}
	return &InsufficientStopTimesNotice{
		BaseNotice: NewBaseNotice("insufficient_stop_times", ERROR, context),
	}
}

// NonIncreasingStopSequenceNotice represents non-increasing stop sequences
type NonIncreasingStopSequenceNotice struct {
	*BaseNotice
}

func NewNonIncreasingStopSequenceNotice(tripID string, currentSeq, previousSeq, rowNumber int) *NonIncreasingStopSequenceNotice {
	context := map[string]interface{}{
		"tripId":      tripID,
		"currentSeq":  currentSeq,
		"previousSeq": previousSeq,
		"rowNumber":   rowNumber,
	}
	return &NonIncreasingStopSequenceNotice{
		BaseNotice: NewBaseNotice("non_increasing_stop_sequence", ERROR, context),
	}
}

// StopSequenceGapNotice represents gaps in stop sequence numbering
type StopSequenceGapNotice struct {
	*BaseNotice
}

func NewStopSequenceGapNotice(tripID string, expectedSeq, actualSeq, rowNumber int) *StopSequenceGapNotice {
	context := map[string]interface{}{
		"tripId":      tripID,
		"expectedSeq": expectedSeq,
		"actualSeq":   actualSeq,
		"rowNumber":   rowNumber,
	}
	return &StopSequenceGapNotice{
		BaseNotice: NewBaseNotice("stop_sequence_gap", INFO, context),
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

// SingleTripPatternNotice represents a pattern used by only one trip
type SingleTripPatternNotice struct {
	*BaseNotice
}

func NewSingleTripPatternNotice(patternID, tripID string, stopCount int) *SingleTripPatternNotice {
	context := map[string]interface{}{
		"patternId": patternID,
		"tripId":    tripID,
		"stopCount": stopCount,
	}
	return &SingleTripPatternNotice{
		BaseNotice: NewBaseNotice("single_trip_pattern", INFO, context),
	}
}

// ShortTripPatternNotice represents a very short trip pattern
type ShortTripPatternNotice struct {
	*BaseNotice
}

func NewShortTripPatternNotice(patternID string, stopCount, tripCount int) *ShortTripPatternNotice {
	context := map[string]interface{}{
		"patternId": patternID,
		"stopCount": stopCount,
		"tripCount": tripCount,
	}
	return &ShortTripPatternNotice{
		BaseNotice: NewBaseNotice("short_trip_pattern", WARNING, context),
	}
}

// LongTripPatternNotice represents a very long trip pattern
type LongTripPatternNotice struct {
	*BaseNotice
}

func NewLongTripPatternNotice(patternID string, stopCount, tripCount int) *LongTripPatternNotice {
	context := map[string]interface{}{
		"patternId": patternID,
		"stopCount": stopCount,
		"tripCount": tripCount,
	}
	return &LongTripPatternNotice{
		BaseNotice: NewBaseNotice("long_trip_pattern", INFO, context),
	}
}

// TripPatternSummaryNotice provides summary information about trip patterns
type TripPatternSummaryNotice struct {
	*BaseNotice
}

func NewTripPatternSummaryNotice(totalPatterns, totalTrips int, avgTripsPerPattern float64) *TripPatternSummaryNotice {
	context := map[string]interface{}{
		"totalPatterns":       totalPatterns,
		"totalTrips":          totalTrips,
		"avgTripsPerPattern": avgTripsPerPattern,
	}
	return &TripPatternSummaryNotice{
		BaseNotice: NewBaseNotice("trip_pattern_summary", INFO, context),
	}
}

// VeryOldServiceNotice represents a service that ended too long ago
type VeryOldServiceNotice struct {
	*BaseNotice
}

func NewVeryOldServiceNotice(serviceID, endDate string, rowNumber int) *VeryOldServiceNotice {
	context := map[string]interface{}{
		"serviceId": serviceID,
		"endDate":   endDate,
		"rowNumber": rowNumber,
	}
	return &VeryOldServiceNotice{
		BaseNotice: NewBaseNotice("very_old_service", WARNING, context),
	}
}

// VeryFutureServiceNotice represents a service that starts too far in the future
type VeryFutureServiceNotice struct {
	*BaseNotice
}

func NewVeryFutureServiceNotice(serviceID, startDate string, rowNumber int) *VeryFutureServiceNotice {
	context := map[string]interface{}{
		"serviceId": serviceID,
		"startDate": startDate,
		"rowNumber": rowNumber,
	}
	return &VeryFutureServiceNotice{
		BaseNotice: NewBaseNotice("very_future_service", WARNING, context),
	}
}

// LowServiceUsageNotice represents a service with very few trips
type LowServiceUsageNotice struct {
	*BaseNotice
}

func NewLowServiceUsageNotice(serviceID string, tripCount, rowNumber int) *LowServiceUsageNotice {
	context := map[string]interface{}{
		"serviceId": serviceID,
		"tripCount": tripCount,
		"rowNumber": rowNumber,
	}
	return &LowServiceUsageNotice{
		BaseNotice: NewBaseNotice("low_service_usage", INFO, context),
	}
}

// ExcessiveServiceVarietyNotice represents a route with too many different services
type ExcessiveServiceVarietyNotice struct {
	*BaseNotice
}

func NewExcessiveServiceVarietyNotice(routeID string, serviceCount int) *ExcessiveServiceVarietyNotice {
	context := map[string]interface{}{
		"routeId":      routeID,
		"serviceCount": serviceCount,
	}
	return &ExcessiveServiceVarietyNotice{
		BaseNotice: NewBaseNotice("excessive_service_variety", WARNING, context),
	}
}

// SingleTripServiceNotice represents a service with only one trip on a route
type SingleTripServiceNotice struct {
	*BaseNotice
}

func NewSingleTripServiceNotice(routeID, serviceID string, tripCount int) *SingleTripServiceNotice {
	context := map[string]interface{}{
		"routeId":   routeID,
		"serviceId": serviceID,
		"tripCount": tripCount,
	}
	return &SingleTripServiceNotice{
		BaseNotice: NewBaseNotice("single_trip_service", INFO, context),
	}
}

// ServicePatternSummaryNotice provides summary of service patterns
type ServicePatternSummaryNotice struct {
	*BaseNotice
}

func NewServicePatternSummaryNotice(weekdayServices, weekendServices, mixedServices, totalServices int) *ServicePatternSummaryNotice {
	context := map[string]interface{}{
		"weekdayServices": weekdayServices,
		"weekendServices": weekendServices,
		"mixedServices":   mixedServices,
		"totalServices":   totalServices,
	}
	return &ServicePatternSummaryNotice{
		BaseNotice: NewBaseNotice("service_pattern_summary", INFO, context),
	}
}

// VeryShortTripNotice represents a trip with very short duration
type VeryShortTripNotice struct {
	*BaseNotice
}

func NewVeryShortTripNotice(tripID string, duration, stopCount int) *VeryShortTripNotice {
	context := map[string]interface{}{
		"tripId":    tripID,
		"duration":  duration,
		"stopCount": stopCount,
	}
	return &VeryShortTripNotice{
		BaseNotice: NewBaseNotice("very_short_trip", WARNING, context),
	}
}

// VeryLongTripNotice represents a trip with very long duration
type VeryLongTripNotice struct {
	*BaseNotice
}

func NewVeryLongTripNotice(tripID string, duration, stopCount int) *VeryLongTripNotice {
	context := map[string]interface{}{
		"tripId":    tripID,
		"duration":  duration,
		"stopCount": stopCount,
	}
	return &VeryLongTripNotice{
		BaseNotice: NewBaseNotice("very_long_trip", WARNING, context),
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
		BaseNotice: NewBaseNotice("stop_without_service", ERROR, context),
	}
}

// ShortServiceSpanNotice represents a service with very short time span
type ShortServiceSpanNotice struct {
	*BaseNotice
}

func NewShortServiceSpanNotice(routeID, serviceID string, serviceSpan, tripCount int) *ShortServiceSpanNotice {
	context := map[string]interface{}{
		"routeId":     routeID,
		"serviceId":   serviceID,
		"serviceSpan": serviceSpan,
		"tripCount":   tripCount,
	}
	return &ShortServiceSpanNotice{
		BaseNotice: NewBaseNotice("short_service_span", INFO, context),
	}
}

// LongServiceSpanNotice represents a service with very long time span
type LongServiceSpanNotice struct {
	*BaseNotice
}

func NewLongServiceSpanNotice(routeID, serviceID string, serviceSpan, tripCount int) *LongServiceSpanNotice {
	context := map[string]interface{}{
		"routeId":     routeID,
		"serviceId":   serviceID,
		"serviceSpan": serviceSpan,
		"tripCount":   tripCount,
	}
	return &LongServiceSpanNotice{
		BaseNotice: NewBaseNotice("long_service_span", WARNING, context),
	}
}

// IrregularHeadwayNotice represents irregular headway patterns
type IrregularHeadwayNotice struct {
	*BaseNotice
}

func NewIrregularHeadwayNotice(routeID, serviceID string, averageHeadway, variance float64, headwayCount int) *IrregularHeadwayNotice {
	context := map[string]interface{}{
		"routeId":        routeID,
		"serviceId":      serviceID,
		"averageHeadway": averageHeadway,
		"variance":       variance,
		"headwayCount":   headwayCount,
	}
	return &IrregularHeadwayNotice{
		BaseNotice: NewBaseNotice("irregular_headway", WARNING, context),
	}
}

// VeryShortHeadwayNotice represents very short headways
type VeryShortHeadwayNotice struct {
	*BaseNotice
}

func NewVeryShortHeadwayNotice(routeID, serviceID string, headway int) *VeryShortHeadwayNotice {
	context := map[string]interface{}{
		"routeId":   routeID,
		"serviceId": serviceID,
		"headway":   headway,
	}
	return &VeryShortHeadwayNotice{
		BaseNotice: NewBaseNotice("very_short_headway", WARNING, context),
	}
}

// VeryLongHeadwayNotice represents very long headways
type VeryLongHeadwayNotice struct {
	*BaseNotice
}

func NewVeryLongHeadwayNotice(routeID, serviceID string, headway int) *VeryLongHeadwayNotice {
	context := map[string]interface{}{
		"routeId":   routeID,
		"serviceId": serviceID,
		"headway":   headway,
	}
	return &VeryLongHeadwayNotice{
		BaseNotice: NewBaseNotice("very_long_headway", INFO, context),
	}
}

// SchedulingSummaryNotice provides scheduling analysis summary
type SchedulingSummaryNotice struct {
	*BaseNotice
}

func NewSchedulingSummaryNotice(totalRoutes, totalServices, totalTrips int, avgServicesPerRoute, avgTripsPerService float64) *SchedulingSummaryNotice {
	context := map[string]interface{}{
		"totalRoutes":         totalRoutes,
		"totalServices":       totalServices,
		"totalTrips":          totalTrips,
		"avgServicesPerRoute": avgServicesPerRoute,
		"avgTripsPerService":  avgTripsPerService,
	}
	return &SchedulingSummaryNotice{
		BaseNotice: NewBaseNotice("scheduling_summary", INFO, context),
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

// ExcessiveRoutePatternVariationsNotice represents a route with too many pattern variations
type ExcessiveRoutePatternVariationsNotice struct {
	*BaseNotice
}

func NewExcessiveRoutePatternVariationsNotice(routeID string, directionID, variations, tripCount int) *ExcessiveRoutePatternVariationsNotice {
	context := map[string]interface{}{
		"routeId":     routeID,
		"directionId": directionID,
		"variations":  variations,
		"tripCount":   tripCount,
	}
	return &ExcessiveRoutePatternVariationsNotice{
		BaseNotice: NewBaseNotice("excessive_route_pattern_variations", WARNING, context),
	}
}

// UnbalancedDirectionTripsNotice represents unbalanced trips between directions
type UnbalancedDirectionTripsNotice struct {
	*BaseNotice
}

func NewUnbalancedDirectionTripsNotice(routeID string, dir1, trips1, dir2, trips2 int) *UnbalancedDirectionTripsNotice {
	context := map[string]interface{}{
		"routeId":        routeID,
		"direction1":     dir1,
		"direction1Trips": trips1,
		"direction2":     dir2,
		"direction2Trips": trips2,
	}
	return &UnbalancedDirectionTripsNotice{
		BaseNotice: NewBaseNotice("unbalanced_direction_trips", WARNING, context),
	}
}

// LimitedServiceVarietyNotice represents a route with very few services
type LimitedServiceVarietyNotice struct {
	*BaseNotice
}

func NewLimitedServiceVarietyNotice(routeID string, serviceCount, tripCount int) *LimitedServiceVarietyNotice {
	context := map[string]interface{}{
		"routeId":      routeID,
		"serviceCount": serviceCount,
		"tripCount":    tripCount,
	}
	return &LimitedServiceVarietyNotice{
		BaseNotice: NewBaseNotice("limited_service_variety", INFO, context),
	}
}

// LowRouteUsageNotice represents a route with very few trips
type LowRouteUsageNotice struct {
	*BaseNotice
}

func NewLowRouteUsageNotice(routeID string, tripCount, serviceCount int) *LowRouteUsageNotice {
	context := map[string]interface{}{
		"routeId":      routeID,
		"tripCount":    tripCount,
		"serviceCount": serviceCount,
	}
	return &LowRouteUsageNotice{
		BaseNotice: NewBaseNotice("low_route_usage", WARNING, context),
	}
}

// VeryLongRouteNotice represents a route with many stops
type VeryLongRouteNotice struct {
	*BaseNotice
}

func NewVeryLongRouteNotice(routeID string, stopCount, tripCount int) *VeryLongRouteNotice {
	context := map[string]interface{}{
		"routeId":   routeID,
		"stopCount": stopCount,
		"tripCount": tripCount,
	}
	return &VeryLongRouteNotice{
		BaseNotice: NewBaseNotice("very_long_route", INFO, context),
	}
}

// VeryShortRouteNotice represents a route with very few stops
type VeryShortRouteNotice struct {
	*BaseNotice
}

func NewVeryShortRouteNotice(routeID string, stopCount, tripCount int) *VeryShortRouteNotice {
	context := map[string]interface{}{
		"routeId":   routeID,
		"stopCount": stopCount,
		"tripCount": tripCount,
	}
	return &VeryShortRouteNotice{
		BaseNotice: NewBaseNotice("very_short_route", WARNING, context),
	}
}

// LowTimepointCoverageNotice represents low timepoint coverage
type LowTimepointCoverageNotice struct {
	*BaseNotice
}

func NewLowTimepointCoverageNotice(routeID string, tripsWithTimepoints, totalTrips int, coverage float64) *LowTimepointCoverageNotice {
	context := map[string]interface{}{
		"routeId":              routeID,
		"tripsWithTimepoints": tripsWithTimepoints,
		"totalTrips":          totalTrips,
		"coverage":            coverage,
	}
	return &LowTimepointCoverageNotice{
		BaseNotice: NewBaseNotice("low_timepoint_coverage", WARNING, context),
	}
}

// RouteNetworkSummaryNotice provides route network analysis summary
type RouteNetworkSummaryNotice struct {
	*BaseNotice
}

func NewRouteNetworkSummaryNotice(totalRoutes, totalTrips int, avgTripsPerRoute float64, routeTypes, agencies int) *RouteNetworkSummaryNotice {
	context := map[string]interface{}{
		"totalRoutes":       totalRoutes,
		"totalTrips":        totalTrips,
		"avgTripsPerRoute": avgTripsPerRoute,
		"routeTypes":        routeTypes,
		"agencies":          agencies,
	}
	return &RouteNetworkSummaryNotice{
		BaseNotice: NewBaseNotice("route_network_summary", INFO, context),
	}
}

// HighRouteTypeDiversityNotice represents high route type diversity in an agency
type HighRouteTypeDiversityNotice struct {
	*BaseNotice
}

func NewHighRouteTypeDiversityNotice(agencyID string, routeTypes, routeCount int) *HighRouteTypeDiversityNotice {
	context := map[string]interface{}{
		"agencyId":   agencyID,
		"routeTypes": routeTypes,
		"routeCount": routeCount,
	}
	return &HighRouteTypeDiversityNotice{
		BaseNotice: NewBaseNotice("high_route_type_diversity", INFO, context),
	}
}

// InvalidLatitudeNotice represents invalid latitude coordinates
type InvalidLatitudeNotice struct {
	*BaseNotice
}

func NewInvalidLatitudeNotice(stopID string, latitude float64, rowNumber int) *InvalidLatitudeNotice {
	context := map[string]interface{}{
		"stopId":    stopID,
		"latitude":  latitude,
		"rowNumber": rowNumber,
	}
	return &InvalidLatitudeNotice{
		BaseNotice: NewBaseNotice("invalid_latitude", ERROR, context),
	}
}

// InvalidLongitudeNotice represents invalid longitude coordinates
type InvalidLongitudeNotice struct {
	*BaseNotice
}

func NewInvalidLongitudeNotice(stopID string, longitude float64, rowNumber int) *InvalidLongitudeNotice {
	context := map[string]interface{}{
		"stopId":    stopID,
		"longitude": longitude,
		"rowNumber": rowNumber,
	}
	return &InvalidLongitudeNotice{
		BaseNotice: NewBaseNotice("invalid_longitude", ERROR, context),
	}
}

// VeryLargeFeedCoverageNotice represents feed with very large geographic coverage
type VeryLargeFeedCoverageNotice struct {
	*BaseNotice
}

func NewVeryLargeFeedCoverageNotice(minLat, maxLat, minLon, maxLon, latSpan, lonSpan float64) *VeryLargeFeedCoverageNotice {
	context := map[string]interface{}{
		"minLat":  minLat,
		"maxLat":  maxLat,
		"minLon":  minLon,
		"maxLon":  maxLon,
		"latSpan": latSpan,
		"lonSpan": lonSpan,
	}
	return &VeryLargeFeedCoverageNotice{
		BaseNotice: NewBaseNotice("very_large_feed_coverage", WARNING, context),
	}
}

// VerySmallFeedCoverageNotice represents feed with very small geographic coverage
type VerySmallFeedCoverageNotice struct {
	*BaseNotice
}

func NewVerySmallFeedCoverageNotice(minLat, maxLat, minLon, maxLon, latSpan, lonSpan float64) *VerySmallFeedCoverageNotice {
	context := map[string]interface{}{
		"minLat":  minLat,
		"maxLat":  maxLat,
		"minLon":  minLon,
		"maxLon":  maxLon,
		"latSpan": latSpan,
		"lonSpan": lonSpan,
	}
	return &VerySmallFeedCoverageNotice{
		BaseNotice: NewBaseNotice("very_small_feed_coverage", INFO, context),
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

// VeryCloseStopsNotice represents stops that are very close to each other
type VeryCloseStopsNotice struct {
	*BaseNotice
}

func NewVeryCloseStopsNotice(stop1ID, stop2ID string, distance float64, row1, row2 int) *VeryCloseStopsNotice {
	context := map[string]interface{}{
		"stop1Id":   stop1ID,
		"stop2Id":   stop2ID,
		"distance":  distance,
		"rowNumber1": row1,
		"rowNumber2": row2,
	}
	return &VeryCloseStopsNotice{
		BaseNotice: NewBaseNotice("very_close_stops", INFO, context),
	}
}

// ShapePointOutsideFeedBoundsNotice represents shape point outside feed bounds
type ShapePointOutsideFeedBoundsNotice struct {
	*BaseNotice
}

func NewShapePointOutsideFeedBoundsNotice(shapeID string, sequence int, lat, lon float64, rowNumber int) *ShapePointOutsideFeedBoundsNotice {
	context := map[string]interface{}{
		"shapeId":   shapeID,
		"sequence":  sequence,
		"latitude":  lat,
		"longitude": lon,
		"rowNumber": rowNumber,
	}
	return &ShapePointOutsideFeedBoundsNotice{
		BaseNotice: NewBaseNotice("shape_point_outside_feed_bounds", WARNING, context),
	}
}

// ShapeDistanceInconsistentWithGeographyNotice represents shape distance inconsistent with geography
type ShapeDistanceInconsistentWithGeographyNotice struct {
	*BaseNotice
}

func NewShapeDistanceInconsistentWithGeographyNotice(shapeID string, sequence int, providedDist, geoDist, difference float64, rowNumber int) *ShapeDistanceInconsistentWithGeographyNotice {
	context := map[string]interface{}{
		"shapeId":        shapeID,
		"sequence":       sequence,
		"providedDist":   providedDist,
		"geographicDist": geoDist,
		"difference":     difference,
		"rowNumber":      rowNumber,
	}
	return &ShapeDistanceInconsistentWithGeographyNotice{
		BaseNotice: NewBaseNotice("shape_distance_inconsistent_with_geography", WARNING, context),
	}
}

// HighStopDensityAreaNotice represents area with high stop density
type HighStopDensityAreaNotice struct {
	*BaseNotice
}

func NewHighStopDensityAreaNotice(centerLat, centerLon, radius float64, stopCount int) *HighStopDensityAreaNotice {
	context := map[string]interface{}{
		"centerLat": centerLat,
		"centerLon": centerLon,
		"radius":    radius,
		"stopCount": stopCount,
	}
	return &HighStopDensityAreaNotice{
		BaseNotice: NewBaseNotice("high_stop_density_area", INFO, context),
	}
}

// LowStopClusteringNotice represents feed with low stop clustering
type LowStopClusteringNotice struct {
	*BaseNotice
}

func NewLowStopClusteringNotice(clusterCount, totalStops int) *LowStopClusteringNotice {
	context := map[string]interface{}{
		"clusterCount": clusterCount,
		"totalStops":   totalStops,
	}
	return &LowStopClusteringNotice{
		BaseNotice: NewBaseNotice("low_stop_clustering", INFO, context),
	}
}

// GeospatialSummaryNotice provides geospatial analysis summary
type GeospatialSummaryNotice struct {
	*BaseNotice
}

func NewGeospatialSummaryNotice(totalStops, totalShapes, clusterCount int) *GeospatialSummaryNotice {
	context := map[string]interface{}{
		"totalStops":   totalStops,
		"totalShapes":  totalShapes,
		"clusterCount": clusterCount,
	}
	return &GeospatialSummaryNotice{
		BaseNotice: NewBaseNotice("geospatial_summary", INFO, context),
	}
}

// === NETWORK TOPOLOGY NOTICES ===

// IsolatedStopNotice represents stops with no connections
type IsolatedStopNotice struct {
	*BaseNotice
}

func NewIsolatedStopNotice(stopID string, tripCount int) *IsolatedStopNotice {
	context := map[string]interface{}{
		"stopId":    stopID,
		"tripCount": tripCount,
	}
	return &IsolatedStopNotice{
		BaseNotice: NewBaseNotice("isolated_stop", WARNING, context),
	}
}

// FragmentedNetworkNotice represents disconnected network components
type FragmentedNetworkNotice struct {
	*BaseNotice
}

func NewFragmentedNetworkNotice(componentCount, mainComponentSize, totalStops int) *FragmentedNetworkNotice {
	context := map[string]interface{}{
		"componentCount":     componentCount,
		"mainComponentSize":  mainComponentSize,
		"totalStops":         totalStops,
	}
	return &FragmentedNetworkNotice{
		BaseNotice: NewBaseNotice("fragmented_network", WARNING, context),
	}
}

// SmallNetworkComponentNotice represents small disconnected components
type SmallNetworkComponentNotice struct {
	*BaseNotice
}

func NewSmallNetworkComponentNotice(componentRank, stopCount, routeCount int) *SmallNetworkComponentNotice {
	context := map[string]interface{}{
		"componentRank": componentRank,
		"stopCount":     stopCount,
		"routeCount":    routeCount,
	}
	return &SmallNetworkComponentNotice{
		BaseNotice: NewBaseNotice("small_network_component", INFO, context),
	}
}

// LowNetworkConnectivityNotice represents poorly connected networks
type LowNetworkConnectivityNotice struct {
	*BaseNotice
}

func NewLowNetworkConnectivityNotice(totalStops, totalEdges int, avgConnectivity float64) *LowNetworkConnectivityNotice {
	context := map[string]interface{}{
		"totalStops":      totalStops,
		"totalEdges":      totalEdges,
		"avgConnectivity": avgConnectivity,
	}
	return &LowNetworkConnectivityNotice{
		BaseNotice: NewBaseNotice("low_network_connectivity", WARNING, context),
	}
}

// LowTransferOpportunityNotice represents networks with few transfer points
type LowTransferOpportunityNotice struct {
	*BaseNotice
}

func NewLowTransferOpportunityNotice(transferStops, totalStops int, transferRatio float64) *LowTransferOpportunityNotice {
	context := map[string]interface{}{
		"transferStops": transferStops,
		"totalStops":    totalStops,
		"transferRatio": transferRatio,
	}
	return &LowTransferOpportunityNotice{
		BaseNotice: NewBaseNotice("low_transfer_opportunity", INFO, context),
	}
}

// NetworkHubIdentifiedNotice represents major network hubs
type NetworkHubIdentifiedNotice struct {
	*BaseNotice
}

func NewNetworkHubIdentifiedNotice(stopID string, routeCount, connectionCount, totalHubs int) *NetworkHubIdentifiedNotice {
	context := map[string]interface{}{
		"stopId":          stopID,
		"routeCount":      routeCount,
		"connectionCount": connectionCount,
		"totalHubs":       totalHubs,
	}
	return &NetworkHubIdentifiedNotice{
		BaseNotice: NewBaseNotice("network_hub_identified", INFO, context),
	}
}

// MajorTransferPointNotice represents significant transfer opportunities
type MajorTransferPointNotice struct {
	*BaseNotice
}

func NewMajorTransferPointNotice(stopID string, routeCount, connectionCount int, transferValue float64) *MajorTransferPointNotice {
	context := map[string]interface{}{
		"stopId":          stopID,
		"routeCount":      routeCount,
		"connectionCount": connectionCount,
		"transferValue":   transferValue,
	}
	return &MajorTransferPointNotice{
		BaseNotice: NewBaseNotice("major_transfer_point", INFO, context),
	}
}

// OverlappingRoutesNotice represents routes with identical stop patterns
type OverlappingRoutesNotice struct {
	*BaseNotice
}

func NewOverlappingRoutesNotice(routeIDs []string, stopCount int) *OverlappingRoutesNotice {
	context := map[string]interface{}{
		"routeIds":  routeIDs,
		"stopCount": stopCount,
	}
	return &OverlappingRoutesNotice{
		BaseNotice: NewBaseNotice("overlapping_routes", WARNING, context),
	}
}

// NetworkTopologySummaryNotice provides network topology analysis summary
type NetworkTopologySummaryNotice struct {
	*BaseNotice
}

func NewNetworkTopologySummaryNotice(totalStops, totalEdges, totalRoutes, totalTrips, transferStops, componentCount int, avgConnectivity float64, maxRouteCount int) *NetworkTopologySummaryNotice {
	context := map[string]interface{}{
		"totalStops":      totalStops,
		"totalEdges":      totalEdges,
		"totalRoutes":     totalRoutes,
		"totalTrips":      totalTrips,
		"transferStops":   transferStops,
		"componentCount":  componentCount,
		"avgConnectivity": avgConnectivity,
		"maxRouteCount":   maxRouteCount,
	}
	return &NetworkTopologySummaryNotice{
		BaseNotice: NewBaseNotice("network_topology_summary", INFO, context),
	}
}

// === FEED EXPIRATION NOTICES ===

// FeedInfoEndDateMissingNotice represents missing feed end date
type FeedInfoEndDateMissingNotice struct {
	*BaseNotice
}

func NewFeedInfoEndDateMissingNotice(rowNumber int) *FeedInfoEndDateMissingNotice {
	context := map[string]interface{}{
		"csvRowNumber": rowNumber,
	}
	return &FeedInfoEndDateMissingNotice{
		BaseNotice: NewBaseNotice("feed_info_end_date_missing", WARNING, context),
	}
}

// FeedExpiredNotice represents an expired feed
type FeedExpiredNotice struct {
	*BaseNotice
}

func NewFeedExpiredNotice(endDate, currentDate string, daysExpired int) *FeedExpiredNotice {
	context := map[string]interface{}{
		"endDate":     endDate,
		"currentDate": currentDate,
		"daysExpired": daysExpired,
	}
	return &FeedExpiredNotice{
		BaseNotice: NewBaseNotice("feed_expired", ERROR, context),
	}
}

// FeedExpiresWithin7DaysNotice represents feed expiring within 7 days
type FeedExpiresWithin7DaysNotice struct {
	*BaseNotice
}

func NewFeedExpiresWithin7DaysNotice(endDate, currentDate string, daysUntilExpiration int) *FeedExpiresWithin7DaysNotice {
	context := map[string]interface{}{
		"endDate":              endDate,
		"currentDate":          currentDate,
		"daysUntilExpiration":  daysUntilExpiration,
	}
	return &FeedExpiresWithin7DaysNotice{
		BaseNotice: NewBaseNotice("feed_expires_within_7_days", ERROR, context),
	}
}

// FeedExpiresWithin30DaysNotice represents feed expiring within 30 days
type FeedExpiresWithin30DaysNotice struct {
	*BaseNotice
}

func NewFeedExpiresWithin30DaysNotice(endDate, currentDate string, daysUntilExpiration int) *FeedExpiresWithin30DaysNotice {
	context := map[string]interface{}{
		"endDate":              endDate,
		"currentDate":          currentDate,
		"daysUntilExpiration":  daysUntilExpiration,
	}
	return &FeedExpiresWithin30DaysNotice{
		BaseNotice: NewBaseNotice("feed_expires_within_30_days", WARNING, context),
	}
}

// ServiceExpiredNotice represents expired service based on calendar
type ServiceExpiredNotice struct {
	*BaseNotice
}

func NewServiceExpiredNotice(lastServiceDate, currentDate string, daysExpired int) *ServiceExpiredNotice {
	context := map[string]interface{}{
		"lastServiceDate": lastServiceDate,
		"currentDate":     currentDate,
		"daysExpired":     daysExpired,
	}
	return &ServiceExpiredNotice{
		BaseNotice: NewBaseNotice("service_expired", ERROR, context),
	}
}

// ServiceExpiresWithin7DaysNotice represents service expiring within 7 days
type ServiceExpiresWithin7DaysNotice struct {
	*BaseNotice
}

func NewServiceExpiresWithin7DaysNotice(lastServiceDate, currentDate string, daysUntilExpiration int) *ServiceExpiresWithin7DaysNotice {
	context := map[string]interface{}{
		"lastServiceDate":      lastServiceDate,
		"currentDate":          currentDate,
		"daysUntilExpiration":  daysUntilExpiration,
	}
	return &ServiceExpiresWithin7DaysNotice{
		BaseNotice: NewBaseNotice("service_expires_within_7_days", ERROR, context),
	}
}

// ServiceExpiresWithin30DaysNotice represents service expiring within 30 days
type ServiceExpiresWithin30DaysNotice struct {
	*BaseNotice
}

func NewServiceExpiresWithin30DaysNotice(lastServiceDate, currentDate string, daysUntilExpiration int) *ServiceExpiresWithin30DaysNotice {
	context := map[string]interface{}{
		"lastServiceDate":      lastServiceDate,
		"currentDate":          currentDate,
		"daysUntilExpiration":  daysUntilExpiration,
	}
	return &ServiceExpiresWithin30DaysNotice{
		BaseNotice: NewBaseNotice("service_expires_within_30_days", WARNING, context),
	}
}

// NoServiceDateFoundNotice represents no service dates found
type NoServiceDateFoundNotice struct {
	*BaseNotice
}

func NewNoServiceDateFoundNotice() *NoServiceDateFoundNotice {
	return &NoServiceDateFoundNotice{
		BaseNotice: NewBaseNotice("no_service_date_found", ERROR, map[string]interface{}{}),
	}
}

// NoServiceNext7DaysNotice represents no service in next 7 days
type NoServiceNext7DaysNotice struct {
	*BaseNotice
}

func NewNoServiceNext7DaysNotice(startDate, endDate string) *NoServiceNext7DaysNotice {
	context := map[string]interface{}{
		"startDate": startDate,
		"endDate":   endDate,
	}
	return &NoServiceNext7DaysNotice{
		BaseNotice: NewBaseNotice("no_service_next_7_days", ERROR, context),
	}
}

// NoTripsNext7DaysNotice represents no trips in next 7 days
type NoTripsNext7DaysNotice struct {
	*BaseNotice
}

func NewNoTripsNext7DaysNotice(startDate, endDate string, serviceCount int) *NoTripsNext7DaysNotice {
	context := map[string]interface{}{
		"startDate":    startDate,
		"endDate":      endDate,
		"serviceCount": serviceCount,
	}
	return &NoTripsNext7DaysNotice{
		BaseNotice: NewBaseNotice("no_trips_next_7_days", ERROR, context),
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
		BaseNotice: NewBaseNotice("duplicate_route_long_name", WARNING, context),
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
		BaseNotice: NewBaseNotice("duplicate_route_short_name", WARNING, context),
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
		BaseNotice: NewBaseNotice("duplicate_route_name_combination", WARNING, context),
	}
}

// === ROUTE COLOR CONTRAST NOTICES ===

// RouteColorContrastNotice represents insufficient color contrast
type RouteColorContrastNotice struct {
	*BaseNotice
}

func NewRouteColorContrastNotice(routeID, routeColor, routeTextColor string, actualContrast, minimumContrast float64, rowNumber int, severity SeverityLevel) *RouteColorContrastNotice {
	context := map[string]interface{}{
		"routeId":         routeID,
		"routeColor":      routeColor,
		"routeTextColor":  routeTextColor,
		"actualContrast":  actualContrast,
		"minimumContrast": minimumContrast,
		"csvRowNumber":    rowNumber,
	}
	return &RouteColorContrastNotice{
		BaseNotice: NewBaseNotice("route_color_contrast", severity, context),
	}
}

// LightTextOnLightBackgroundNotice represents light text on light background
type LightTextOnLightBackgroundNotice struct {
	*BaseNotice
}

func NewLightTextOnLightBackgroundNotice(routeID, routeColor, routeTextColor string, rowNumber int) *LightTextOnLightBackgroundNotice {
	context := map[string]interface{}{
		"routeId":        routeID,
		"routeColor":     routeColor,
		"routeTextColor": routeTextColor,
		"csvRowNumber":   rowNumber,
	}
	return &LightTextOnLightBackgroundNotice{
		BaseNotice: NewBaseNotice("light_text_on_light_background", WARNING, context),
	}
}

// DarkTextOnDarkBackgroundNotice represents dark text on dark background
type DarkTextOnDarkBackgroundNotice struct {
	*BaseNotice
}

func NewDarkTextOnDarkBackgroundNotice(routeID, routeColor, routeTextColor string, rowNumber int) *DarkTextOnDarkBackgroundNotice {
	context := map[string]interface{}{
		"routeId":        routeID,
		"routeColor":     routeColor,
		"routeTextColor": routeTextColor,
		"csvRowNumber":   rowNumber,
	}
	return &DarkTextOnDarkBackgroundNotice{
		BaseNotice: NewBaseNotice("dark_text_on_dark_background", WARNING, context),
	}
}

// SimilarColorsNotice represents colors that are too similar
type SimilarColorsNotice struct {
	*BaseNotice
}

func NewSimilarColorsNotice(routeID, routeColor, routeTextColor string, rowNumber int) *SimilarColorsNotice {
	context := map[string]interface{}{
		"routeId":        routeID,
		"routeColor":     routeColor,
		"routeTextColor": routeTextColor,
		"csvRowNumber":   rowNumber,
	}
	return &SimilarColorsNotice{
		BaseNotice: NewBaseNotice("similar_colors", WARNING, context),
	}
}

// RedGreenColorCombinationNotice represents red-green color combinations
type RedGreenColorCombinationNotice struct {
	*BaseNotice
}

func NewRedGreenColorCombinationNotice(routeID, routeColor, routeTextColor string, rowNumber int) *RedGreenColorCombinationNotice {
	context := map[string]interface{}{
		"routeId":        routeID,
		"routeColor":     routeColor,
		"routeTextColor": routeTextColor,
		"csvRowNumber":   rowNumber,
	}
	return &RedGreenColorCombinationNotice{
		BaseNotice: NewBaseNotice("red_green_color_combination", INFO, context),
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
		BaseNotice: NewBaseNotice("missing_required_stop_name", ERROR, context),
	}
}

// StopNameMissingButInheritedNotice represents stop name inherited from parent
type StopNameMissingButInheritedNotice struct {
	*BaseNotice
}

func NewStopNameMissingButInheritedNotice(stopID, parentStationID, parentName string, locationType, rowNumber int) *StopNameMissingButInheritedNotice {
	context := map[string]interface{}{
		"stopId":          stopID,
		"parentStationId": parentStationID,
		"parentName":      parentName,
		"locationType":    locationType,
		"csvRowNumber":    rowNumber,
	}
	return &StopNameMissingButInheritedNotice{
		BaseNotice: NewBaseNotice("stop_name_missing_but_inherited", INFO, context),
	}
}

// GenericStopNameNotice represents generic or placeholder stop names
type GenericStopNameNotice struct {
	*BaseNotice
}

func NewGenericStopNameNotice(stopID, stopName string, rowNumber int) *GenericStopNameNotice {
	context := map[string]interface{}{
		"stopId":       stopID,
		"stopName":     stopName,
		"csvRowNumber": rowNumber,
	}
	return &GenericStopNameNotice{
		BaseNotice: NewBaseNotice("generic_stop_name", WARNING, context),
	}
}

// StopNameTooLongNotice represents excessively long stop names
type StopNameTooLongNotice struct {
	*BaseNotice
}

func NewStopNameTooLongNotice(stopID, stopName string, actualLength, maxLength, rowNumber int, severity SeverityLevel) *StopNameTooLongNotice {
	context := map[string]interface{}{
		"stopId":       stopID,
		"stopName":     stopName,
		"actualLength": actualLength,
		"maxLength":    maxLength,
		"csvRowNumber": rowNumber,
	}
	return &StopNameTooLongNotice{
		BaseNotice: NewBaseNotice("stop_name_too_long", severity, context),
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

// StopNameContainsHTMLNotice represents HTML content in stop names
type StopNameContainsHTMLNotice struct {
	*BaseNotice
}

func NewStopNameContainsHTMLNotice(stopID, stopName string, rowNumber int) *StopNameContainsHTMLNotice {
	context := map[string]interface{}{
		"stopId":       stopID,
		"stopName":     stopName,
		"csvRowNumber": rowNumber,
	}
	return &StopNameContainsHTMLNotice{
		BaseNotice: NewBaseNotice("stop_name_contains_html", WARNING, context),
	}
}

// StopNameContainsURLNotice represents URL content in stop names
type StopNameContainsURLNotice struct {
	*BaseNotice
}

func NewStopNameContainsURLNotice(stopID, stopName string, rowNumber int) *StopNameContainsURLNotice {
	context := map[string]interface{}{
		"stopId":       stopID,
		"stopName":     stopName,
		"csvRowNumber": rowNumber,
	}
	return &StopNameContainsURLNotice{
		BaseNotice: NewBaseNotice("stop_name_contains_url", WARNING, context),
	}
}

// StopNameDescriptionDuplicateNotice represents identical stop name and description
type StopNameDescriptionDuplicateNotice struct {
	*BaseNotice
}

func NewStopNameDescriptionDuplicateNotice(stopID, stopName string, rowNumber int) *StopNameDescriptionDuplicateNotice {
	context := map[string]interface{}{
		"stopId":       stopID,
		"stopName":     stopName,
		"csvRowNumber": rowNumber,
	}
	return &StopNameDescriptionDuplicateNotice{
		BaseNotice: NewBaseNotice("stop_name_description_duplicate", INFO, context),
	}
}

// StopNameAllCapsNotice represents all-caps stop names
type StopNameAllCapsNotice struct {
	*BaseNotice
}

func NewStopNameAllCapsNotice(stopID, stopName string, rowNumber int) *StopNameAllCapsNotice {
	context := map[string]interface{}{
		"stopId":       stopID,
		"stopName":     stopName,
		"csvRowNumber": rowNumber,
	}
	return &StopNameAllCapsNotice{
		BaseNotice: NewBaseNotice("stop_name_all_caps", INFO, context),
	}
}

// StopNameRepeatedWordNotice represents repeated words in stop names
type StopNameRepeatedWordNotice struct {
	*BaseNotice
}

func NewStopNameRepeatedWordNotice(stopID, stopName, repeatedWord string, rowNumber int) *StopNameRepeatedWordNotice {
	context := map[string]interface{}{
		"stopId":       stopID,
		"stopName":     stopName,
		"repeatedWord": repeatedWord,
		"csvRowNumber": rowNumber,
	}
	return &StopNameRepeatedWordNotice{
		BaseNotice: NewBaseNotice("stop_name_repeated_word", WARNING, context),
	}
}

// === DATE TRIPS NOTICES ===

// NoServiceDefinedNotice represents no service defined in calendar
type NoServiceDefinedNotice struct {
	*BaseNotice
}

func NewNoServiceDefinedNotice() *NoServiceDefinedNotice {
	return &NoServiceDefinedNotice{
		BaseNotice: NewBaseNotice("no_service_defined", ERROR, map[string]interface{}{}),
	}
}

// InsufficientServiceNext7DaysNotice represents insufficient service coverage in next 7 days
type InsufficientServiceNext7DaysNotice struct {
	*BaseNotice
}

func NewInsufficientServiceNext7DaysNotice(daysWithService, totalDays int, startDate, endDate string) *InsufficientServiceNext7DaysNotice {
	context := map[string]interface{}{
		"daysWithService": daysWithService,
		"totalDays":       totalDays,
		"startDate":       startDate,
		"endDate":         endDate,
	}
	return &InsufficientServiceNext7DaysNotice{
		BaseNotice: NewBaseNotice("insufficient_service_next_7_days", WARNING, context),
	}
}

// LowTripVolumeNext7DaysNotice represents low trip volume in next 7 days
type LowTripVolumeNext7DaysNotice struct {
	*BaseNotice
}

func NewLowTripVolumeNext7DaysNotice(totalTrips, daysWithService int, avgTripsPerDay float64, startDate, endDate string) *LowTripVolumeNext7DaysNotice {
	context := map[string]interface{}{
		"totalTrips":      totalTrips,
		"daysWithService": daysWithService,
		"avgTripsPerDay":  avgTripsPerDay,
		"startDate":       startDate,
		"endDate":         endDate,
	}
	return &LowTripVolumeNext7DaysNotice{
		BaseNotice: NewBaseNotice("low_trip_volume_next_7_days", WARNING, context),
	}
}

// InsufficientServiceNext30DaysNotice represents insufficient service coverage in next 30 days
type InsufficientServiceNext30DaysNotice struct {
	*BaseNotice
}

func NewInsufficientServiceNext30DaysNotice(daysWithService, totalDays int, serviceRatio float64, startDate, endDate string) *InsufficientServiceNext30DaysNotice {
	context := map[string]interface{}{
		"daysWithService": daysWithService,
		"totalDays":       totalDays,
		"serviceRatio":    serviceRatio,
		"startDate":       startDate,
		"endDate":         endDate,
	}
	return &InsufficientServiceNext30DaysNotice{
		BaseNotice: NewBaseNotice("insufficient_service_next_30_days", WARNING, context),
	}
}

// === BIKES ALLOWANCE NOTICES ===

// MissingBikesAllowedForFerryNotice represents missing bike allowance for ferry trips
type MissingBikesAllowedForFerryNotice struct {
	*BaseNotice
}

func NewMissingBikesAllowedForFerryNotice(tripID, routeID string, rowNumber int) *MissingBikesAllowedForFerryNotice {
	context := map[string]interface{}{
		"tripId":       tripID,
		"routeId":      routeID,
		"csvRowNumber": rowNumber,
	}
	return &MissingBikesAllowedForFerryNotice{
		BaseNotice: NewBaseNotice("missing_bikes_allowed_for_ferry", WARNING, context),
	}
}

// InvalidBikesAllowedValueNotice represents invalid bikes_allowed value
type InvalidBikesAllowedValueNotice struct {
	*BaseNotice
}

func NewInvalidBikesAllowedValueNotice(tripID string, bikesAllowed, rowNumber int) *InvalidBikesAllowedValueNotice {
	context := map[string]interface{}{
		"tripId":       tripID,
		"bikesAllowed": bikesAllowed,
		"csvRowNumber": rowNumber,
	}
	return &InvalidBikesAllowedValueNotice{
		BaseNotice: NewBaseNotice("invalid_bikes_allowed_value", ERROR, context),
	}
}

// BikeWheelchairAccessibilityMismatchNotice represents potential accessibility mismatch
type BikeWheelchairAccessibilityMismatchNotice struct {
	*BaseNotice
}

func NewBikeWheelchairAccessibilityMismatchNotice(tripID, routeID string, bikesAllowed, wheelchairAccessible, rowNumber int) *BikeWheelchairAccessibilityMismatchNotice {
	context := map[string]interface{}{
		"tripId":              tripID,
		"routeId":             routeID,
		"bikesAllowed":        bikesAllowed,
		"wheelchairAccessible": wheelchairAccessible,
		"csvRowNumber":        rowNumber,
	}
	return &BikeWheelchairAccessibilityMismatchNotice{
		BaseNotice: NewBaseNotice("bike_wheelchair_accessibility_mismatch", INFO, context),
	}
}

// UnusualBikeAllowanceNotice represents unusual bike allowance for route type
type UnusualBikeAllowanceNotice struct {
	*BaseNotice
}

func NewUnusualBikeAllowanceNotice(tripID, routeID string, routeType, bikesAllowed, rowNumber int) *UnusualBikeAllowanceNotice {
	context := map[string]interface{}{
		"tripId":       tripID,
		"routeId":      routeID,
		"routeType":    routeType,
		"bikesAllowed": bikesAllowed,
		"csvRowNumber": rowNumber,
	}
	return &UnusualBikeAllowanceNotice{
		BaseNotice: NewBaseNotice("unusual_bike_allowance", INFO, context),
	}
}

// === SHAPE DISTANCE NOTICES ===

// ShapeDistanceDecreasingNotice represents decreasing shape distances
type ShapeDistanceDecreasingNotice struct {
	*BaseNotice
}

func NewShapeDistanceDecreasingNotice(shapeID string, prevSequence, currentSequence int, prevDistance, currentDistance float64, rowNumber int) *ShapeDistanceDecreasingNotice {
	context := map[string]interface{}{
		"shapeId":         shapeID,
		"prevSequence":    prevSequence,
		"currentSequence": currentSequence,
		"prevDistance":    prevDistance,
		"currentDistance": currentDistance,
		"csvRowNumber":    rowNumber,
	}
	return &ShapeDistanceDecreasingNotice{
		BaseNotice: NewBaseNotice("shape_distance_decreasing", ERROR, context),
	}
}

// ShapeDistanceNotIncreasingNotice represents non-increasing shape distances
type ShapeDistanceNotIncreasingNotice struct {
	*BaseNotice
}

func NewShapeDistanceNotIncreasingNotice(shapeID string, prevSequence, currentSequence int, distance float64, rowNumber int) *ShapeDistanceNotIncreasingNotice {
	context := map[string]interface{}{
		"shapeId":         shapeID,
		"prevSequence":    prevSequence,
		"currentSequence": currentSequence,
		"distance":        distance,
		"csvRowNumber":    rowNumber,
	}
	return &ShapeDistanceNotIncreasingNotice{
		BaseNotice: NewBaseNotice("shape_distance_not_increasing", WARNING, context),
	}
}

// UnrealisticShapeDistanceNotice represents unrealistic shape distances
type UnrealisticShapeDistanceNotice struct {
	*BaseNotice
}

func NewUnrealisticShapeDistanceNotice(shapeID string, prevSequence, currentSequence int, providedDistance, geoDistance, ratio float64, rowNumber int) *UnrealisticShapeDistanceNotice {
	context := map[string]interface{}{
		"shapeId":          shapeID,
		"prevSequence":     prevSequence,
		"currentSequence":  currentSequence,
		"providedDistance": providedDistance,
		"geoDistance":      geoDistance,
		"ratio":            ratio,
		"csvRowNumber":     rowNumber,
	}
	return &UnrealisticShapeDistanceNotice{
		BaseNotice: NewBaseNotice("unrealistic_shape_distance", WARNING, context),
	}
}

// IncompleteShapeDistanceNotice represents incomplete distance information
type IncompleteShapeDistanceNotice struct {
	*BaseNotice
}

func NewIncompleteShapeDistanceNotice(shapeID string, pointsWithDistance, totalPoints, missingCount int) *IncompleteShapeDistanceNotice {
	context := map[string]interface{}{
		"shapeId":           shapeID,
		"pointsWithDistance": pointsWithDistance,
		"totalPoints":       totalPoints,
		"missingCount":      missingCount,
	}
	return &IncompleteShapeDistanceNotice{
		BaseNotice: NewBaseNotice("incomplete_shape_distance", INFO, context),
	}
}

// ShapeDistanceNotStartingFromZeroNotice represents distance not starting from zero
type ShapeDistanceNotStartingFromZeroNotice struct {
	*BaseNotice
}

func NewShapeDistanceNotStartingFromZeroNotice(shapeID string, firstSequence int, firstDistance float64, rowNumber int) *ShapeDistanceNotStartingFromZeroNotice {
	context := map[string]interface{}{
		"shapeId":       shapeID,
		"firstSequence": firstSequence,
		"firstDistance": firstDistance,
		"csvRowNumber":  rowNumber,
	}
	return &ShapeDistanceNotStartingFromZeroNotice{
		BaseNotice: NewBaseNotice("shape_distance_not_starting_from_zero", INFO, context),
	}
}

// LargeShapeDistanceJumpNotice represents large jumps in shape distance
type LargeShapeDistanceJumpNotice struct {
	*BaseNotice
}

func NewLargeShapeDistanceJumpNotice(shapeID string, prevSequence, currentSequence int, jump, geoDistance float64, rowNumber int) *LargeShapeDistanceJumpNotice {
	context := map[string]interface{}{
		"shapeId":         shapeID,
		"prevSequence":    prevSequence,
		"currentSequence": currentSequence,
		"jump":            jump,
		"geoDistance":     geoDistance,
		"csvRowNumber":    rowNumber,
	}
	return &LargeShapeDistanceJumpNotice{
		BaseNotice: NewBaseNotice("large_shape_distance_jump", WARNING, context),
	}
}

// === ADDITIONAL FREQUENCY NOTICES ===

// SmallFrequencyGapNotice represents small gaps between frequency entries
type SmallFrequencyGapNotice struct {
	*BaseNotice
}

func NewSmallFrequencyGapNotice(tripID, endTime, startTime string, gapSeconds, rowNumber int) *SmallFrequencyGapNotice {
	context := map[string]interface{}{
		"tripId":       tripID,
		"endTime":      endTime,
		"startTime":    startTime,
		"gapSeconds":   gapSeconds,
		"csvRowNumber": rowNumber,
	}
	return &SmallFrequencyGapNotice{
		BaseNotice: NewBaseNotice("small_frequency_gap", INFO, context),
	}
}

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
		BaseNotice: NewBaseNotice("frequency_duration_shorter_than_headway", ERROR, context),
	}
}

// VeryLongFrequencyPeriodNotice represents very long frequency periods
type VeryLongFrequencyPeriodNotice struct {
	*BaseNotice
}

func NewVeryLongFrequencyPeriodNotice(tripID, startTime, endTime string, duration, rowNumber int) *VeryLongFrequencyPeriodNotice {
	context := map[string]interface{}{
		"tripId":       tripID,
		"startTime":    startTime,
		"endTime":      endTime,
		"duration":     duration,
		"csvRowNumber": rowNumber,
	}
	return &VeryLongFrequencyPeriodNotice{
		BaseNotice: NewBaseNotice("very_long_frequency_period", WARNING, context),
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
		BaseNotice: NewBaseNotice("attribution_without_role", ERROR, context),
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

// SingleTripBlockNotice represents block with only one trip
type SingleTripBlockNotice struct {
	*BaseNotice
}

func NewSingleTripBlockNotice(blockID, tripID string, rowNumber int) *SingleTripBlockNotice {
	context := map[string]interface{}{
		"blockId":      blockID,
		"tripId":       tripID,
		"csvRowNumber": rowNumber,
	}
	return &SingleTripBlockNotice{
		BaseNotice: NewBaseNotice("single_trip_block", INFO, context),
	}
}

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
		BaseNotice: NewBaseNotice("block_service_mismatch", ERROR, context),
	}
}

// BlockMultipleRoutesNotice represents block spanning multiple routes
type BlockMultipleRoutesNotice struct {
	*BaseNotice
}

func NewBlockMultipleRoutesNotice(blockID string, routeIDs []string, tripCount int) *BlockMultipleRoutesNotice {
	context := map[string]interface{}{
		"blockId":   blockID,
		"routeIds":  routeIDs,
		"tripCount": tripCount,
	}
	return &BlockMultipleRoutesNotice{
		BaseNotice: NewBaseNotice("block_multiple_routes", INFO, context),
	}
}

// BlockTooManyTripsNotice represents block with too many trips
type BlockTooManyTripsNotice struct {
	*BaseNotice
}

func NewBlockTooManyTripsNotice(blockID string, tripCount int) *BlockTooManyTripsNotice {
	context := map[string]interface{}{
		"blockId":   blockID,
		"tripCount": tripCount,
	}
	return &BlockTooManyTripsNotice{
		BaseNotice: NewBaseNotice("block_too_many_trips", WARNING, context),
	}
}


// === STOP TIME HEADSIGN NOTICES ===

// TooManyHeadsignsInTripNotice represents too many different headsigns in one trip
type TooManyHeadsignsInTripNotice struct {
	*BaseNotice
}

func NewTooManyHeadsignsInTripNotice(tripID string, count int, headsigns []string) *TooManyHeadsignsInTripNotice {
	context := map[string]interface{}{
		"tripId":    tripID,
		"count":     count,
		"headsigns": headsigns,
	}
	return &TooManyHeadsignsInTripNotice{
		BaseNotice: NewBaseNotice("too_many_headsigns_in_trip", WARNING, context),
	}
}

// HeadsignChangeWithinTripNotice represents headsign change within trip
type HeadsignChangeWithinTripNotice struct {
	*BaseNotice
}

func NewHeadsignChangeWithinTripNotice(tripID string, prevSequence, currentSequence int, prevHeadsign, currentHeadsign string, rowNumber int) *HeadsignChangeWithinTripNotice {
	context := map[string]interface{}{
		"tripId":           tripID,
		"prevSequence":     prevSequence,
		"currentSequence":  currentSequence,
		"prevHeadsign":     prevHeadsign,
		"currentHeadsign":  currentHeadsign,
		"csvRowNumber":     rowNumber,
	}
	return &HeadsignChangeWithinTripNotice{
		BaseNotice: NewBaseNotice("headsign_change_within_trip", INFO, context),
	}
}

// FrequentHeadsignChangesNotice represents frequent headsign changes in trip
type FrequentHeadsignChangesNotice struct {
	*BaseNotice
}

func NewFrequentHeadsignChangesNotice(tripID string, changeCount int) *FrequentHeadsignChangesNotice {
	context := map[string]interface{}{
		"tripId":      tripID,
		"changeCount": changeCount,
	}
	return &FrequentHeadsignChangesNotice{
		BaseNotice: NewBaseNotice("frequent_headsign_changes", WARNING, context),
	}
}

// StopTripHeadsignMismatchNotice represents mismatch between stop and trip headsign
type StopTripHeadsignMismatchNotice struct {
	*BaseNotice
}

func NewStopTripHeadsignMismatchNotice(tripID string, stopSequence int, stopHeadsign, tripHeadsign string, rowNumber int) *StopTripHeadsignMismatchNotice {
	context := map[string]interface{}{
		"tripId":        tripID,
		"stopSequence":  stopSequence,
		"stopHeadsign":  stopHeadsign,
		"tripHeadsign":  tripHeadsign,
		"csvRowNumber":  rowNumber,
	}
	return &StopTripHeadsignMismatchNotice{
		BaseNotice: NewBaseNotice("stop_trip_headsign_mismatch", WARNING, context),
	}
}

// VeryShortHeadsignNotice represents very short headsign
type VeryShortHeadsignNotice struct {
	*BaseNotice
}

func NewVeryShortHeadsignNotice(tripID string, stopSequence int, headsign string, rowNumber int) *VeryShortHeadsignNotice {
	context := map[string]interface{}{
		"tripId":       tripID,
		"stopSequence": stopSequence,
		"headsign":     headsign,
		"csvRowNumber": rowNumber,
	}
	return &VeryShortHeadsignNotice{
		BaseNotice: NewBaseNotice("very_short_headsign", WARNING, context),
	}
}

// VeryLongHeadsignNotice represents very long headsign
type VeryLongHeadsignNotice struct {
	*BaseNotice
}

func NewVeryLongHeadsignNotice(tripID string, stopSequence, length, rowNumber int) *VeryLongHeadsignNotice {
	context := map[string]interface{}{
		"tripId":       tripID,
		"stopSequence": stopSequence,
		"length":       length,
		"csvRowNumber": rowNumber,
	}
	return &VeryLongHeadsignNotice{
		BaseNotice: NewBaseNotice("very_long_headsign", WARNING, context),
	}
}

// AllCapsHeadsignNotice represents all caps headsign
type AllCapsHeadsignNotice struct {
	*BaseNotice
}

func NewAllCapsHeadsignNotice(tripID string, stopSequence int, headsign string, rowNumber int) *AllCapsHeadsignNotice {
	context := map[string]interface{}{
		"tripId":       tripID,
		"stopSequence": stopSequence,
		"headsign":     headsign,
		"csvRowNumber": rowNumber,
	}
	return &AllCapsHeadsignNotice{
		BaseNotice: NewBaseNotice("all_caps_headsign", INFO, context),
	}
}

// ExcessivePunctuationHeadsignNotice represents headsign with excessive punctuation
type ExcessivePunctuationHeadsignNotice struct {
	*BaseNotice
}

func NewExcessivePunctuationHeadsignNotice(tripID string, stopSequence int, headsign string, punctuationCount, rowNumber int) *ExcessivePunctuationHeadsignNotice {
	context := map[string]interface{}{
		"tripId":           tripID,
		"stopSequence":     stopSequence,
		"headsign":         headsign,
		"punctuationCount": punctuationCount,
		"csvRowNumber":     rowNumber,
	}
	return &ExcessivePunctuationHeadsignNotice{
		BaseNotice: NewBaseNotice("excessive_punctuation_headsign", WARNING, context),
	}
}

// SuspiciousHeadsignPatternNotice represents suspicious headsign pattern
type SuspiciousHeadsignPatternNotice struct {
	*BaseNotice
}

func NewSuspiciousHeadsignPatternNotice(tripID string, stopSequence int, headsign, pattern string, rowNumber int) *SuspiciousHeadsignPatternNotice {
	context := map[string]interface{}{
		"tripId":       tripID,
		"stopSequence": stopSequence,
		"headsign":     headsign,
		"pattern":      pattern,
		"csvRowNumber": rowNumber,
	}
	return &SuspiciousHeadsignPatternNotice{
		BaseNotice: NewBaseNotice("suspicious_headsign_pattern", WARNING, context),
	}
}

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

// UncommonRouteTypeNotice represents uncommon route_type value
type UncommonRouteTypeNotice struct {
	*BaseNotice
}

func NewUncommonRouteTypeNotice(routeID string, routeType int, description string, rowNumber int) *UncommonRouteTypeNotice {
	context := map[string]interface{}{
		"routeId":      routeID,
		"routeType":    routeType,
		"description":  description,
		"csvRowNumber": rowNumber,
	}
	return &UncommonRouteTypeNotice{
		BaseNotice: NewBaseNotice("uncommon_route_type", INFO, context),
	}
}

// RouteTypeNameMismatchNotice represents mismatch between route type and name
type RouteTypeNameMismatchNotice struct {
	*BaseNotice
}

func NewRouteTypeNameMismatchNotice(routeID string, routeType int, expectedMode, shortName, longName string, rowNumber int) *RouteTypeNameMismatchNotice {
	context := map[string]interface{}{
		"routeId":      routeID,
		"routeType":    routeType,
		"expectedMode": expectedMode,
		"shortName":    shortName,
		"longName":     longName,
		"csvRowNumber": rowNumber,
	}
	return &RouteTypeNameMismatchNotice{
		BaseNotice: NewBaseNotice("route_type_name_mismatch", WARNING, context),
	}
}

// AgencyMixedRouteTypesNotice represents agency with many different route types
type AgencyMixedRouteTypesNotice struct {
	*BaseNotice
}

func NewAgencyMixedRouteTypesNotice(agencyID string, typeCount int, routeTypes []int) *AgencyMixedRouteTypesNotice {
	context := map[string]interface{}{
		"agencyId":   agencyID,
		"typeCount":  typeCount,
		"routeTypes": routeTypes,
	}
	return &AgencyMixedRouteTypesNotice{
		BaseNotice: NewBaseNotice("agency_mixed_route_types", INFO, context),
	}
}

// SingleRouteTypeInFeedNotice represents feed with only one uncommon route type
type SingleRouteTypeInFeedNotice struct {
	*BaseNotice
}

func NewSingleRouteTypeInFeedNotice(routeType int, description string, routeCount int) *SingleRouteTypeInFeedNotice {
	context := map[string]interface{}{
		"routeType":   routeType,
		"description": description,
		"routeCount":  routeCount,
	}
	return &SingleRouteTypeInFeedNotice{
		BaseNotice: NewBaseNotice("single_route_type_in_feed", WARNING, context),
	}
}

// UnusualRouteTypeCombinationNotice represents unusual combination of route types
type UnusualRouteTypeCombinationNotice struct {
	*BaseNotice
}

func NewUnusualRouteTypeCombinationNotice(routeTypes []int, descriptions []string) *UnusualRouteTypeCombinationNotice {
	context := map[string]interface{}{
		"routeTypes":   routeTypes,
		"descriptions": descriptions,
	}
	return &UnusualRouteTypeCombinationNotice{
		BaseNotice: NewBaseNotice("unusual_route_type_combination", INFO, context),
	}
}

// === ADDITIONAL TRANSFER TIMING NOTICES ===

// LongDistanceTransferNotice represents transfer over long distance
type LongDistanceTransferNotice struct {
	*BaseNotice
}

func NewLongDistanceTransferNotice(fromStopID, toStopID string, distance float64, transferTypeDesc string, rowNumber int) *LongDistanceTransferNotice {
	context := map[string]interface{}{
		"fromStopId":       fromStopID,
		"toStopId":         toStopID,
		"distance":         distance,
		"transferTypeDesc": transferTypeDesc,
		"csvRowNumber":     rowNumber,
	}
	return &LongDistanceTransferNotice{
		BaseNotice: NewBaseNotice("long_distance_transfer", WARNING, context),
	}
}

// UnrealisticTransferTimeNotice represents unrealistic transfer time
type UnrealisticTransferTimeNotice struct {
	*BaseNotice
}

func NewUnrealisticTransferTimeNotice(fromStopID, toStopID string, minTime, expectedTime int, distance float64, rowNumber int) *UnrealisticTransferTimeNotice {
	context := map[string]interface{}{
		"fromStopId":   fromStopID,
		"toStopId":     toStopID,
		"minTime":      minTime,
		"expectedTime": expectedTime,
		"distance":     distance,
		"csvRowNumber": rowNumber,
	}
	return &UnrealisticTransferTimeNotice{
		BaseNotice: NewBaseNotice("unrealistic_transfer_time", ERROR, context),
	}
}

// VeryLongTransferTimeNotice represents very long transfer time
type VeryLongTransferTimeNotice struct {
	*BaseNotice
}

func NewVeryLongTransferTimeNotice(fromStopID, toStopID string, minTime, rowNumber int) *VeryLongTransferTimeNotice {
	context := map[string]interface{}{
		"fromStopId":   fromStopID,
		"toStopId":     toStopID,
		"minTime":      minTime,
		"csvRowNumber": rowNumber,
	}
	return &VeryLongTransferTimeNotice{
		BaseNotice: NewBaseNotice("very_long_transfer_time", WARNING, context),
	}
}

// VeryShortTransferTimeNotice represents very short transfer time
type VeryShortTransferTimeNotice struct {
	*BaseNotice
}

func NewVeryShortTransferTimeNotice(fromStopID, toStopID string, minTime, rowNumber int) *VeryShortTransferTimeNotice {
	context := map[string]interface{}{
		"fromStopId":   fromStopID,
		"toStopId":     toStopID,
		"minTime":      minTime,
		"csvRowNumber": rowNumber,
	}
	return &VeryShortTransferTimeNotice{
		BaseNotice: NewBaseNotice("very_short_transfer_time", WARNING, context),
	}
}

// CloseStopsNotPossibleTransferNotice represents close stops marked as not possible transfer
type CloseStopsNotPossibleTransferNotice struct {
	*BaseNotice
}

func NewCloseStopsNotPossibleTransferNotice(fromStopID, toStopID string, distance float64, rowNumber int) *CloseStopsNotPossibleTransferNotice {
	context := map[string]interface{}{
		"fromStopId":   fromStopID,
		"toStopId":     toStopID,
		"distance":     distance,
		"csvRowNumber": rowNumber,
	}
	return &CloseStopsNotPossibleTransferNotice{
		BaseNotice: NewBaseNotice("close_stops_not_possible_transfer", WARNING, context),
	}
}

// InconsistentBidirectionalTransferNotice represents inconsistent bidirectional transfers
type InconsistentBidirectionalTransferNotice struct {
	*BaseNotice
}

func NewInconsistentBidirectionalTransferNotice(fromStopID, toStopID string, transferType1, transferType2, rowNumber int) *InconsistentBidirectionalTransferNotice {
	context := map[string]interface{}{
		"fromStopId":    fromStopID,
		"toStopId":      toStopID,
		"transferType1": transferType1,
		"transferType2": transferType2,
		"csvRowNumber":  rowNumber,
	}
	return &InconsistentBidirectionalTransferNotice{
		BaseNotice: NewBaseNotice("inconsistent_bidirectional_transfer", WARNING, context),
	}
}

// === SERVICE CALENDAR NOTICES ===

// ServiceWithoutDefinitionNotice represents service without calendar definition
type ServiceWithoutDefinitionNotice struct {
	*BaseNotice
}

func NewServiceWithoutDefinitionNotice(serviceID string) *ServiceWithoutDefinitionNotice {
	context := map[string]interface{}{
		"serviceId": serviceID,
	}
	return &ServiceWithoutDefinitionNotice{
		BaseNotice: NewBaseNotice("service_without_definition", ERROR, context),
	}
}

// CalendarNoDaysSelectedNotice represents calendar with no days selected
type CalendarNoDaysSelectedNotice struct {
	*BaseNotice
}

func NewCalendarNoDaysSelectedNotice(serviceID string, rowNumber int) *CalendarNoDaysSelectedNotice {
	context := map[string]interface{}{
		"serviceId":    serviceID,
		"csvRowNumber": rowNumber,
	}
	return &CalendarNoDaysSelectedNotice{
		BaseNotice: NewBaseNotice("calendar_no_days_selected", ERROR, context),
	}
}

// CalendarEndBeforeStartNotice represents calendar end date before start date
type CalendarEndBeforeStartNotice struct {
	*BaseNotice
}

func NewCalendarEndBeforeStartNotice(serviceID, startDate, endDate string, rowNumber int) *CalendarEndBeforeStartNotice {
	context := map[string]interface{}{
		"serviceId":    serviceID,
		"startDate":    startDate,
		"endDate":      endDate,
		"csvRowNumber": rowNumber,
	}
	return &CalendarEndBeforeStartNotice{
		BaseNotice: NewBaseNotice("calendar_end_before_start", ERROR, context),
	}
}

// VeryLongServicePeriodNotice represents very long service period
type VeryLongServicePeriodNotice struct {
	*BaseNotice
}

func NewVeryLongServicePeriodNotice(serviceID, startDate, endDate string, durationDays, rowNumber int) *VeryLongServicePeriodNotice {
	context := map[string]interface{}{
		"serviceId":     serviceID,
		"startDate":     startDate,
		"endDate":       endDate,
		"durationDays":  durationDays,
		"csvRowNumber":  rowNumber,
	}
	return &VeryLongServicePeriodNotice{
		BaseNotice: NewBaseNotice("very_long_service_period", WARNING, context),
	}
}


// DuplicateCalendarDateNotice represents duplicate calendar date
type DuplicateCalendarDateNotice struct {
	*BaseNotice
}

func NewDuplicateCalendarDateNotice(serviceID, date string, firstRowNumber, duplicateRowNumber int) *DuplicateCalendarDateNotice {
	context := map[string]interface{}{
		"serviceId":          serviceID,
		"date":               date,
		"firstRowNumber":     firstRowNumber,
		"duplicateRowNumber": duplicateRowNumber,
	}
	return &DuplicateCalendarDateNotice{
		BaseNotice: NewBaseNotice("duplicate_calendar_date", ERROR, context),
	}
}


// InactiveServiceCurrentMonthNotice represents service inactive in current month
type InactiveServiceCurrentMonthNotice struct {
	*BaseNotice
}

func NewInactiveServiceCurrentMonthNotice(serviceID string) *InactiveServiceCurrentMonthNotice {
	context := map[string]interface{}{
		"serviceId": serviceID,
	}
	return &InactiveServiceCurrentMonthNotice{
		BaseNotice: NewBaseNotice("inactive_service_current_month", WARNING, context),
	}
}

// LowFrequencyServiceNotice represents low frequency service
type LowFrequencyServiceNotice struct {
	*BaseNotice
}

func NewLowFrequencyServiceNotice(serviceID string, activeDays int) *LowFrequencyServiceNotice {
	context := map[string]interface{}{
		"serviceId":  serviceID,
		"activeDays": activeDays,
	}
	return &LowFrequencyServiceNotice{
		BaseNotice: NewBaseNotice("low_frequency_service", INFO, context),
	}
}

// WeekendOnlyServiceNotice represents weekend-only service
type WeekendOnlyServiceNotice struct {
	*BaseNotice
}

func NewWeekendOnlyServiceNotice(serviceID string, rowNumber int) *WeekendOnlyServiceNotice {
	context := map[string]interface{}{
		"serviceId":    serviceID,
		"csvRowNumber": rowNumber,
	}
	return &WeekendOnlyServiceNotice{
		BaseNotice: NewBaseNotice("weekend_only_service", INFO, context),
	}
}

// SingleDayServiceNotice represents single day service
type SingleDayServiceNotice struct {
	*BaseNotice
}

func NewSingleDayServiceNotice(serviceID, dayName string, rowNumber int) *SingleDayServiceNotice {
	context := map[string]interface{}{
		"serviceId":    serviceID,
		"dayName":      dayName,
		"csvRowNumber": rowNumber,
	}
	return &SingleDayServiceNotice{
		BaseNotice: NewBaseNotice("single_day_service", INFO, context),
	}
}

// UnusualServicePatternNotice represents unusual service pattern
type UnusualServicePatternNotice struct {
	*BaseNotice
}

func NewUnusualServicePatternNotice(serviceID, pattern string, rowNumber int) *UnusualServicePatternNotice {
	context := map[string]interface{}{
		"serviceId":    serviceID,
		"pattern":      pattern,
		"csvRowNumber": rowNumber,
	}
	return &UnusualServicePatternNotice{
		BaseNotice: NewBaseNotice("unusual_service_pattern", INFO, context),
	}
}

// MostlyCalendarDatesServicesNotice represents feed with mostly calendar_dates services
type MostlyCalendarDatesServicesNotice struct {
	*BaseNotice
}

func NewMostlyCalendarDatesServicesNotice(calendarDatesCount, totalCount int) *MostlyCalendarDatesServicesNotice {
	context := map[string]interface{}{
		"calendarDatesCount": calendarDatesCount,
		"totalCount":         totalCount,
	}
	return &MostlyCalendarDatesServicesNotice{
		BaseNotice: NewBaseNotice("mostly_calendar_dates_services", INFO, context),
	}
}

// === VALIDATOR SYSTEM NOTICES ===

// ValidatorErrorNotice is generated when a validator encounters an error
type ValidatorErrorNotice struct {
	*BaseNotice
}

func NewValidatorErrorNotice(validatorName string, errorMessage string) *ValidatorErrorNotice {
	context := map[string]interface{}{
		"validatorName": validatorName,
		"errorMessage":  errorMessage,
	}
	return &ValidatorErrorNotice{
		BaseNotice: NewBaseNotice("validator_error", ERROR, context),
	}
}

// ValidationSummaryNotice is generated to summarize validation process
type ValidationSummaryNotice struct {
	*BaseNotice
}

func NewValidationSummaryNotice(totalValidators, validatorsRun, validatorsFailed int) *ValidationSummaryNotice {
	context := map[string]interface{}{
		"totalValidators":   totalValidators,
		"validatorsRun":     validatorsRun,
		"validatorsFailed":  validatorsFailed,
		"successRate":       float64(validatorsRun) / float64(totalValidators) * 100,
	}
	severity := INFO
	if validatorsFailed > 0 {
		severity = WARNING
	}
	return &ValidationSummaryNotice{
		BaseNotice: NewBaseNotice("validation_summary", severity, context),
	}
}