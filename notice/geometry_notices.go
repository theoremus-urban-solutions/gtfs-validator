package notice

// Geometric notices defined by the Canonical GTFS Schedule Validator
// <https://gtfs-validator.mobilitydata.org/rules.html>. Each of these needs the
// shape polyline, a spatial pass over the feed, or both, which is why the
// checks behind them are gated rather than run on every validation. Codes and
// severities match the published rules; the checks are our own implementations.

// StopTooFarFromShapeNotice reports a stop the trip's shape never comes near.
// GTFS Best Practices put route alignments within 100 m of the stops the trip
// serves, so a larger gap means either the stop or the shape is in the wrong
// place, and a consumer snapping vehicles to the alignment will show them
// passing a stop they are supposedly calling at.
type StopTooFarFromShapeNotice struct {
	*BaseNotice
}

func NewStopTooFarFromShapeNotice(tripID string, stopID string, stopSequence int, shapeID string, distanceMetres float64, rowNumber int) *StopTooFarFromShapeNotice {
	context := map[string]interface{}{
		"tripId":             tripID,
		"stopId":             stopID,
		"stopSequence":       stopSequence,
		"shapeId":            shapeID,
		"geoDistanceToShape": distanceMetres,
		"csvRowNumber":       rowNumber,
	}
	return &StopTooFarFromShapeNotice{
		BaseNotice: NewBaseNotice("stop_too_far_from_shape", WARNING, context),
	}
}

// StopTooFarFromShapeUsingUserDistanceNotice reports a stop whose declared
// shape_dist_traveled points at a part of the shape far from where the stop
// actually is. The geometry and the distance values disagree, so a consumer
// that trusts shape_dist_traveled to cut the alignment into per-stop legs gets
// the wrong leg.
type StopTooFarFromShapeUsingUserDistanceNotice struct {
	*BaseNotice
}

func NewStopTooFarFromShapeUsingUserDistanceNotice(tripID string, stopID string, stopSequence int, shapeID string, distanceMetres float64, rowNumber int) *StopTooFarFromShapeUsingUserDistanceNotice {
	context := map[string]interface{}{
		"tripId":             tripID,
		"stopId":             stopID,
		"stopSequence":       stopSequence,
		"shapeId":            shapeID,
		"geoDistanceToShape": distanceMetres,
		"csvRowNumber":       rowNumber,
	}
	return &StopTooFarFromShapeUsingUserDistanceNotice{
		BaseNotice: NewBaseNotice("stop_too_far_from_shape_using_user_distance", WARNING, context),
	}
}

// StopsMatchShapeOutOfOrderNotice reports two stops the vehicle reaches in one
// order according to stop_times.txt and the opposite order according to where
// they sit along the shape. One of the two orderings is wrong, and which one
// determines whether the fault is in the stop sequence, the stop locations or
// the path of the shape.
type StopsMatchShapeOutOfOrderNotice struct {
	*BaseNotice
}

func NewStopsMatchShapeOutOfOrderNotice(tripID string, shapeID string, stopID1 string, stopSequence1 int, stopID2 string, stopSequence2 int) *StopsMatchShapeOutOfOrderNotice {
	context := map[string]interface{}{
		"tripId":        tripID,
		"shapeId":       shapeID,
		"stopId1":       stopID1,
		"stopSequence1": stopSequence1,
		"stopId2":       stopID2,
		"stopSequence2": stopSequence2,
	}
	return &StopsMatchShapeOutOfOrderNotice{
		BaseNotice: NewBaseNotice("stops_match_shape_out_of_order", WARNING, context),
	}
}

// StopHasTooManyMatchesForShapeNotice reports a stop that sits close to the
// shape in many separate places, so there is no telling which pass of the
// alignment actually serves it. A stop on a genuine loop matches twice; many
// more than that usually means the shape doubles back on itself or the stop is
// in the wrong place.
type StopHasTooManyMatchesForShapeNotice struct {
	*BaseNotice
}

func NewStopHasTooManyMatchesForShapeNotice(tripID string, stopID string, stopSequence int, shapeID string, matchCount int, rowNumber int) *StopHasTooManyMatchesForShapeNotice {
	context := map[string]interface{}{
		"tripId":       tripID,
		"stopId":       stopID,
		"stopSequence": stopSequence,
		"shapeId":      shapeID,
		"matchCount":   matchCount,
		"csvRowNumber": rowNumber,
	}
	return &StopHasTooManyMatchesForShapeNotice{
		BaseNotice: NewBaseNotice("stop_has_too_many_matches_for_shape", WARNING, context),
	}
}

// TripDistanceExceedsShapeDistanceNotice reports a trip whose last stop claims
// a shape_dist_traveled beyond the end of its own shape. The trip travels
// further than the alignment it is drawn on, so every distance-based position
// past the end of the shape is undefined.
type TripDistanceExceedsShapeDistanceNotice struct {
	*BaseNotice
}

func NewTripDistanceExceedsShapeDistanceNotice(tripID string, shapeID string, maxTripDistance float64, maxShapeDistance float64) *TripDistanceExceedsShapeDistanceNotice {
	context := map[string]interface{}{
		"tripId":                   tripID,
		"shapeId":                  shapeID,
		"maxTripDistanceTraveled":  maxTripDistance,
		"maxShapeDistanceTraveled": maxShapeDistance,
	}
	return &TripDistanceExceedsShapeDistanceNotice{
		BaseNotice: NewBaseNotice("trip_distance_exceeds_shape_distance", ERROR, context),
	}
}

// TripDistanceExceedsShapeDistanceBelowThresholdNotice is the same defect as
// TripDistanceExceedsShapeDistanceNotice, but by less than 11.1 m — four
// decimal places of latitude. At that size the overshoot is a rounding
// artefact rather than a trip that outruns its shape.
type TripDistanceExceedsShapeDistanceBelowThresholdNotice struct {
	*BaseNotice
}

func NewTripDistanceExceedsShapeDistanceBelowThresholdNotice(tripID string, shapeID string, maxTripDistance float64, maxShapeDistance float64) *TripDistanceExceedsShapeDistanceBelowThresholdNotice {
	context := map[string]interface{}{
		"tripId":                   tripID,
		"shapeId":                  shapeID,
		"maxTripDistanceTraveled":  maxTripDistance,
		"maxShapeDistanceTraveled": maxShapeDistance,
	}
	return &TripDistanceExceedsShapeDistanceBelowThresholdNotice{
		BaseNotice: NewBaseNotice("trip_distance_exceeds_shape_distance_below_threshold", WARNING, context),
	}
}

// FastTravelBetweenFarStopsNotice reports a vehicle covering more than 10 km of
// a trip faster than any transit mode manages. Distinct from the consecutive
// stop check: over a stretch that long a single mistyped time cannot explain
// it, so the trip's whole timetable or its stop locations are suspect.
type FastTravelBetweenFarStopsNotice struct {
	*BaseNotice
}

func NewFastTravelBetweenFarStopsNotice(tripID string, stopID1 string, stopSequence1 int, stopID2 string, stopSequence2 int, speedKph float64, distanceKm float64, rowNumber1 int, rowNumber2 int) *FastTravelBetweenFarStopsNotice {
	context := map[string]interface{}{
		"tripId":        tripID,
		"stopId1":       stopID1,
		"stopSequence1": stopSequence1,
		"stopId2":       stopID2,
		"stopSequence2": stopSequence2,
		"speedKph":      speedKph,
		"distanceKm":    distanceKm,
		"csvRowNumber1": rowNumber1,
		"csvRowNumber2": rowNumber2,
	}
	return &FastTravelBetweenFarStopsNotice{
		BaseNotice: NewBaseNotice("fast_travel_between_far_stops", WARNING, context),
	}
}

// UnusedShapeNotice reports a shape in shapes.txt that no trip's shape_id
// names. Nothing draws it, so it is dead weight in a file that is usually the
// largest in the feed — and more often than not it means the trip that was
// meant to reference it does not.
type UnusedShapeNotice struct {
	*BaseNotice
}

func NewUnusedShapeNotice(shapeID string, rowNumber int) *UnusedShapeNotice {
	context := map[string]interface{}{
		"shapeId":      shapeID,
		"csvRowNumber": rowNumber,
	}
	return &UnusedShapeNotice{
		BaseNotice: NewBaseNotice("unused_shape", WARNING, context),
	}
}

// TransferDistanceTooLargeNotice reports a transfer spanning more than 10 km.
// No passenger walks that, so the pair is almost always a wrong stop_id rather
// than a connection anyone can make.
type TransferDistanceTooLargeNotice struct {
	*BaseNotice
}

func NewTransferDistanceTooLargeNotice(fromStopID string, toStopID string, distanceMetres float64, rowNumber int) *TransferDistanceTooLargeNotice {
	context := map[string]interface{}{
		"fromStopId":       fromStopID,
		"toStopId":         toStopID,
		"distanceInMeters": distanceMetres,
		"csvRowNumber":     rowNumber,
	}
	return &TransferDistanceTooLargeNotice{
		BaseNotice: NewBaseNotice("transfer_distance_too_large", WARNING, context),
	}
}

// TransferDistanceAbove2KmNotice reports a transfer spanning more than 2 km.
// Long, but not impossible for a station complex or a timed connection, so this
// is reported for review rather than as a defect.
type TransferDistanceAbove2KmNotice struct {
	*BaseNotice
}

func NewTransferDistanceAbove2KmNotice(fromStopID string, toStopID string, distanceMetres float64, rowNumber int) *TransferDistanceAbove2KmNotice {
	context := map[string]interface{}{
		"fromStopId":       fromStopID,
		"toStopId":         toStopID,
		"distanceInMeters": distanceMetres,
		"csvRowNumber":     rowNumber,
	}
	return &TransferDistanceAbove2KmNotice{
		BaseNotice: NewBaseNotice("transfer_distance_above_2_km", INFO, context),
	}
}
