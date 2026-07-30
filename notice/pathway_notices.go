package notice

// Notices for pathways.txt and levels.txt, defined by the Canonical GTFS
// Schedule Validator. Codes and severities match the published rules; the
// checks are our own implementations.
//
// pathways.txt and levels.txt are core GTFS, not an extension: they are what
// a consumer needs to route a passenger through a station rather than to its
// door, and a station whose pathway graph is broken is worse than one with no
// pathways at all, because the consumer trusts it.

// PathwayLoopNotice reports a pathway whose two endpoints are the same
// location. It describes a walk that arrives where it started, so it can only
// ever add cost to a route.
type PathwayLoopNotice struct {
	*BaseNotice
}

func NewPathwayLoopNotice(pathwayID string, rowNumber int, stopID string) *PathwayLoopNotice {
	context := map[string]interface{}{
		"pathwayId":    pathwayID,
		"csvRowNumber": rowNumber,
		"stopId":       stopID,
	}
	return &PathwayLoopNotice{
		BaseNotice: NewBaseNotice("pathway_loop", WARNING, context),
	}
}

// PathwayDanglingGenericNodeNotice reports a generic node with a single
// incident location. A generic node exists to join two or more parts of a
// station; one with a single neighbour is a dead end no passenger has reason
// to walk into.
type PathwayDanglingGenericNodeNotice struct {
	*BaseNotice
}

func NewPathwayDanglingGenericNodeNotice(stopID string, rowNumber int, stopName string, parentStation string) *PathwayDanglingGenericNodeNotice {
	context := map[string]interface{}{
		"stopId":        stopID,
		"csvRowNumber":  rowNumber,
		"stopName":      stopName,
		"parentStation": parentStation,
	}
	return &PathwayDanglingGenericNodeNotice{
		BaseNotice: NewBaseNotice("pathway_dangling_generic_node", WARNING, context),
	}
}

// PathwayUnreachableLocationNotice reports a location that cannot be walked to
// from any entrance of its station, or from which no exit can be walked to.
// Either way the station's pathway graph claims a platform a passenger can
// never reach or never leave.
//
// Reported for platforms, boarding areas and generic nodes only. An entrance
// is reachable by definition — it is the way in — and a station is a container
// rather than a point on the graph.
type PathwayUnreachableLocationNotice struct {
	*BaseNotice
}

func NewPathwayUnreachableLocationNotice(stopID string, rowNumber int, stopName string, locationType int, parentStation string, hasEntrance bool, hasExit bool) *PathwayUnreachableLocationNotice {
	context := map[string]interface{}{
		"stopId":        stopID,
		"csvRowNumber":  rowNumber,
		"stopName":      stopName,
		"locationType":  locationType,
		"parentStation": parentStation,
		"hasEntrance":   hasEntrance,
		"hasExit":       hasExit,
	}
	return &PathwayUnreachableLocationNotice{
		BaseNotice: NewBaseNotice("pathway_unreachable_location", ERROR, context),
	}
}

// PathwayToWrongLocationTypeNotice reports a pathway endpoint that is a
// station. A station is the whole building; a pathway has to start and end
// somewhere a passenger can stand, so endpoints must be platforms,
// entrances/exits, generic nodes or boarding areas.
type PathwayToWrongLocationTypeNotice struct {
	*BaseNotice
}

func NewPathwayToWrongLocationTypeNotice(pathwayID string, rowNumber int, fieldName string, stopID string, locationType int) *PathwayToWrongLocationTypeNotice {
	context := map[string]interface{}{
		"pathwayId":    pathwayID,
		"csvRowNumber": rowNumber,
		"fieldName":    fieldName,
		"stopId":       stopID,
		"locationType": locationType,
	}
	return &PathwayToWrongLocationTypeNotice{
		BaseNotice: NewBaseNotice("pathway_to_wrong_location_type", ERROR, context),
	}
}

// PathwayToPlatformWithBoardingAreasNotice reports a pathway endpoint that is
// a platform subdivided into boarding areas. Once a platform has boarding
// areas it is a parent object rather than a point, and the pathway has to name
// which boarding area it reaches.
type PathwayToPlatformWithBoardingAreasNotice struct {
	*BaseNotice
}

func NewPathwayToPlatformWithBoardingAreasNotice(pathwayID string, rowNumber int, fieldName string, stopID string) *PathwayToPlatformWithBoardingAreasNotice {
	context := map[string]interface{}{
		"pathwayId":    pathwayID,
		"csvRowNumber": rowNumber,
		"fieldName":    fieldName,
		"stopId":       stopID,
	}
	return &PathwayToPlatformWithBoardingAreasNotice{
		BaseNotice: NewBaseNotice("pathway_to_platform_with_boarding_areas", ERROR, context),
	}
}

// PathwayToStopWithAccessOutsideOfStationPathwaysNotice reports a pathway
// endpoint that is a stop declaring stop_access=1. Such a stop is reached from
// the street and is deliberately outside the station's pathway graph, so a
// pathway leading to it contradicts the access it declares.
type PathwayToStopWithAccessOutsideOfStationPathwaysNotice struct {
	*BaseNotice
}

func NewPathwayToStopWithAccessOutsideOfStationPathwaysNotice(pathwayID string, rowNumber int, fieldName string, stopID string) *PathwayToStopWithAccessOutsideOfStationPathwaysNotice {
	context := map[string]interface{}{
		"pathwayId":    pathwayID,
		"csvRowNumber": rowNumber,
		"fieldName":    fieldName,
		"stopId":       stopID,
	}
	return &PathwayToStopWithAccessOutsideOfStationPathwaysNotice{
		BaseNotice: NewBaseNotice("pathway_to_stop_with_access_outside_of_station_pathways", ERROR, context),
	}
}

// BidirectionalExitGateNotice reports an exit gate declared bidirectional. An
// exit gate (pathway_mode=7) is a gate that can only be crossed on the way
// out; a consumer that believes it works both ways will route passengers in
// through a barrier they cannot open.
type BidirectionalExitGateNotice struct {
	*BaseNotice
}

func NewBidirectionalExitGateNotice(pathwayID string, rowNumber int, fromStopID string, toStopID string) *BidirectionalExitGateNotice {
	context := map[string]interface{}{
		"pathwayId":    pathwayID,
		"csvRowNumber": rowNumber,
		"fromStopId":   fromStopID,
		"toStopId":     toStopID,
	}
	return &BidirectionalExitGateNotice{
		BaseNotice: NewBaseNotice("bidirectional_exit_gate", ERROR, context),
	}
}

// MissingLevelIdNotice reports a location joined by an elevator pathway that
// does not say which level it is on. An elevator's whole purpose is to move
// between levels, so without level_id at both ends the ride has no meaning.
type MissingLevelIdNotice struct {
	*BaseNotice
}

// The notice names the stop rather than the elevator, so a location served by
// several elevators is reported once.
func NewMissingLevelIdNotice(stopID string, rowNumber int, stopName string) *MissingLevelIdNotice {
	context := map[string]interface{}{
		"stopId":       stopID,
		"csvRowNumber": rowNumber,
		"stopName":     stopName,
	}
	return &MissingLevelIdNotice{
		BaseNotice: NewBaseNotice("missing_level_id", ERROR, context),
	}
}
