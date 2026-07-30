package notice

// Notices for the low-severity advisories and the translations trio, defined
// by the Canonical GTFS Schedule Validator. Codes and severities match the
// published rules; the checks are our own implementations.
//
// Most of what is here is INFO: a feed that trips these is not wrong, but a
// consumer given the missing information could do more with it — place a
// platform inside its station, price a journey by zone, align a stop to a
// shape. The two translations errors and the two stop_access errors are the
// exception: those describe rows the spec forbids outright.

// TranslationUnknownTableNameNotice reports a translations row naming a table
// the spec does not allow to be translated. Consumers cannot apply the
// translation to anything, so the row is dead weight.
type TranslationUnknownTableNameNotice struct {
	*BaseNotice
}

func NewTranslationUnknownTableNameNotice(rowNumber int, tableName string) *TranslationUnknownTableNameNotice {
	context := map[string]interface{}{
		"csvRowNumber": rowNumber,
		"tableName":    tableName,
	}
	return &TranslationUnknownTableNameNotice{
		BaseNotice: NewBaseNotice("translation_unknown_table_name", WARNING, context),
	}
}

// TranslationUnexpectedValueNotice reports a translations field that carries a
// value where the spec requires it to be empty. A row selects its target
// either by record id or by field value, never both, and feed_info has a
// single row so it selects by neither.
type TranslationUnexpectedValueNotice struct {
	*BaseNotice
}

func NewTranslationUnexpectedValueNotice(rowNumber int, fieldName string, fieldValue string) *TranslationUnexpectedValueNotice {
	context := map[string]interface{}{
		"csvRowNumber": rowNumber,
		"fieldName":    fieldName,
		"fieldValue":   fieldValue,
	}
	return &TranslationUnexpectedValueNotice{
		BaseNotice: NewBaseNotice("translation_unexpected_value", ERROR, context),
	}
}

// TranslationForeignKeyViolationNotice reports a translations row whose
// record_id and record_sub_id name a record that is not in the referenced
// table. The translation applies to nothing.
type TranslationForeignKeyViolationNotice struct {
	*BaseNotice
}

func NewTranslationForeignKeyViolationNotice(rowNumber int, tableName string, recordID string, recordSubID string) *TranslationForeignKeyViolationNotice {
	context := map[string]interface{}{
		"csvRowNumber": rowNumber,
		"tableName":    tableName,
		"recordId":     recordID,
		"recordSubId":  recordSubID,
	}
	return &TranslationForeignKeyViolationNotice{
		BaseNotice: NewBaseNotice("translation_foreign_key_violation", ERROR, context),
	}
}

// PlatformWithoutParentStationNotice reports a platform that is not attached
// to a station. It is legal — a lone bus stop is a platform with no station —
// but where a station does exist, a platform outside it cannot be reached by
// the station's pathways and does not group with its siblings in a trip
// planner.
type PlatformWithoutParentStationNotice struct {
	*BaseNotice
}

func NewPlatformWithoutParentStationNotice(stopID string, rowNumber int) *PlatformWithoutParentStationNotice {
	context := map[string]interface{}{
		"stopId":       stopID,
		"csvRowNumber": rowNumber,
	}
	return &PlatformWithoutParentStationNotice{
		BaseNotice: NewBaseNotice("platform_without_parent_station", INFO, context),
	}
}

// StopWithoutZoneIDNotice reports a platform served by a route that is priced
// by zone, but which declares no zone of its own. The fare rule cannot be
// applied to a journey through that stop.
type StopWithoutZoneIDNotice struct {
	*BaseNotice
}

func NewStopWithoutZoneIDNotice(stopID string, stopName string, rowNumber int) *StopWithoutZoneIDNotice {
	context := map[string]interface{}{
		"stopId":       stopID,
		"stopName":     stopName,
		"csvRowNumber": rowNumber,
	}
	return &StopWithoutZoneIDNotice{
		BaseNotice: NewBaseNotice("stop_without_zone_id", INFO, context),
	}
}

// StopAccessSpecifiedForIncorrectLocationNotice reports stop_access on a
// location that is not a platform. The field says how a passenger reaches a
// platform within its station, which is meaningless for a station, entrance,
// generic node or boarding area.
type StopAccessSpecifiedForIncorrectLocationNotice struct {
	*BaseNotice
}

func NewStopAccessSpecifiedForIncorrectLocationNotice(stopID string, rowNumber int, locationType int, stopAccess string) *StopAccessSpecifiedForIncorrectLocationNotice {
	context := map[string]interface{}{
		"stopId":       stopID,
		"csvRowNumber": rowNumber,
		"locationType": locationType,
		"stopAccess":   stopAccess,
	}
	return &StopAccessSpecifiedForIncorrectLocationNotice{
		BaseNotice: NewBaseNotice("stop_access_specified_for_incorrect_location", ERROR, context),
	}
}

// StopAccessSpecifiedForStopWithNoParentStationNotice reports stop_access on a
// platform that belongs to no station. The field describes access relative to
// a station, so without one there is nothing for it to be relative to.
type StopAccessSpecifiedForStopWithNoParentStationNotice struct {
	*BaseNotice
}

func NewStopAccessSpecifiedForStopWithNoParentStationNotice(stopID string, rowNumber int, stopAccess string) *StopAccessSpecifiedForStopWithNoParentStationNotice {
	context := map[string]interface{}{
		"stopId":       stopID,
		"csvRowNumber": rowNumber,
		"stopAccess":   stopAccess,
	}
	return &StopAccessSpecifiedForStopWithNoParentStationNotice{
		BaseNotice: NewBaseNotice("stop_access_specified_for_stop_with_no_parent_station", ERROR, context),
	}
}

// TripHeadsignMatchesIntermediateStopNotice reports a trip whose headsign
// names a stop the vehicle passes through rather than the one it terminates
// at. A passenger boarding after that stop reads the headsign as a promise the
// vehicle no longer keeps.
type TripHeadsignMatchesIntermediateStopNotice struct {
	*BaseNotice
}

func NewTripHeadsignMatchesIntermediateStopNotice(tripID string, rowNumber int, tripHeadsign string, stopID string, stopSequence int) *TripHeadsignMatchesIntermediateStopNotice {
	context := map[string]interface{}{
		"tripId":       tripID,
		"csvRowNumber": rowNumber,
		"tripHeadsign": tripHeadsign,
		"stopId":       stopID,
		"stopSequence": stopSequence,
	}
	return &TripHeadsignMatchesIntermediateStopNotice{
		BaseNotice: NewBaseNotice("trip_headsign_matches_intermediate_stop", INFO, context),
	}
}

// TripWithShapeDistTraveledButNoShapeDistancesNotice reports a trip whose stop
// times carry distances along a shape that does not carry them on every point.
// The two sets of distances cannot be compared, so a consumer cannot use the
// stop time values to place stops on the geometry.
type TripWithShapeDistTraveledButNoShapeDistancesNotice struct {
	*BaseNotice
}

func NewTripWithShapeDistTraveledButNoShapeDistancesNotice(tripID string, shapeID string, rowNumber int) *TripWithShapeDistTraveledButNoShapeDistancesNotice {
	context := map[string]interface{}{
		"tripId":       tripID,
		"shapeId":      shapeID,
		"csvRowNumber": rowNumber,
	}
	return &TripWithShapeDistTraveledButNoShapeDistancesNotice{
		BaseNotice: NewBaseNotice("trip_with_shape_dist_traveled_but_no_shape_distances", INFO, context),
	}
}

// InconsistentRouteTypeForBlockIDNotice reports a block whose trips run on
// routes of different modes. A block is one vehicle working through the day,
// and a vehicle does not change from a bus into a tram between trips.
type InconsistentRouteTypeForBlockIDNotice struct {
	*BaseNotice
}

func NewInconsistentRouteTypeForBlockIDNotice(blockID string, tripID1 string, routeID1 string, routeType1 int, tripID2 string, routeID2 string, routeType2 int, rowNumber int) *InconsistentRouteTypeForBlockIDNotice {
	context := map[string]interface{}{
		"blockId":      blockID,
		"tripId1":      tripID1,
		"routeId1":     routeID1,
		"routeType1":   routeType1,
		"tripId2":      tripID2,
		"routeId2":     routeID2,
		"routeType2":   routeType2,
		"csvRowNumber": rowNumber,
	}
	return &InconsistentRouteTypeForBlockIDNotice{
		BaseNotice: NewBaseNotice("inconsistent_route_type_for_block_id", WARNING, context),
	}
}

// InconsistentRouteTypeForInSeatTransferNotice reports an in-seat transfer
// between trips on routes of different modes. Staying in one's seat implies
// staying in one vehicle, which cannot be a bus for one trip and a tram for
// the next.
type InconsistentRouteTypeForInSeatTransferNotice struct {
	*BaseNotice
}

func NewInconsistentRouteTypeForInSeatTransferNotice(rowNumber int, fromTripID string, fromRouteID string, fromRouteType int, toTripID string, toRouteID string, toRouteType int) *InconsistentRouteTypeForInSeatTransferNotice {
	context := map[string]interface{}{
		"csvRowNumber":  rowNumber,
		"fromTripId":    fromTripID,
		"fromRouteId":   fromRouteID,
		"fromRouteType": fromRouteType,
		"toTripId":      toTripID,
		"toRouteId":     toRouteID,
		"toRouteType":   toRouteType,
	}
	return &InconsistentRouteTypeForInSeatTransferNotice{
		BaseNotice: NewBaseNotice("inconsistent_route_type_for_in_seat_transfer", WARNING, context),
	}
}

// TransferWithSuspiciousMidTripInSeatNotice reports an in-seat transfer that
// joins a trip somewhere other than at its ends. The expected shape is the
// last stop of the arriving trip continuing into the first stop of the
// departing one; anything else models a passenger staying seated while the
// vehicle carries on past them, which most consumers do not implement.
type TransferWithSuspiciousMidTripInSeatNotice struct {
	*BaseNotice
}

func NewTransferWithSuspiciousMidTripInSeatNotice(rowNumber int, tripID string, stopID string, tripIDFieldName string) *TransferWithSuspiciousMidTripInSeatNotice {
	context := map[string]interface{}{
		"csvRowNumber":    rowNumber,
		"tripId":          tripID,
		"stopId":          stopID,
		"tripIdFieldName": tripIDFieldName,
	}
	return &TransferWithSuspiciousMidTripInSeatNotice{
		BaseNotice: NewBaseNotice("transfer_with_suspicious_mid_trip_in_seat", WARNING, context),
	}
}
