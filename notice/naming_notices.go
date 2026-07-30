package notice

// Canonical Tier 3 quality warnings about the names and URLs a feed shows to
// riders. Codes and severities match
// <https://gtfs-validator.mobilitydata.org/rules.html>; the checks themselves
// are our own implementations.
//
// None of these break a trip planner. They are warnings because what they
// catch is a feed that reads badly — a stop announced as "GALLERIA MALL", a
// route whose name is printed twice because the producer put it in both the
// name and the description, a "more information" link that goes to the
// agency's front page instead of the route's timetable.

// MixedCaseRecommendedFieldNotice reports a customer-facing name written
// entirely in one case. Screen readers pronounce all-caps text as initialisms
// and rider-facing displays cannot restore the casing, so the name reaches the
// rider degraded.
type MixedCaseRecommendedFieldNotice struct {
	*BaseNotice
}

func NewMixedCaseRecommendedFieldNotice(filename string, fieldName string, fieldValue string, rowNumber int) *MixedCaseRecommendedFieldNotice {
	context := map[string]interface{}{
		"filename":     filename,
		"fieldName":    fieldName,
		"fieldValue":   fieldValue,
		"csvRowNumber": rowNumber,
	}
	return &MixedCaseRecommendedFieldNotice{
		BaseNotice: NewBaseNotice("mixed_case_recommended_field", WARNING, context),
	}
}

// RouteLongNameContainsShortNameNotice reports a route_long_name that repeats
// route_short_name. Consumers routinely render the two together, so "14" plus
// "Route 14" is displayed as "14 Route 14".
type RouteLongNameContainsShortNameNotice struct {
	*BaseNotice
}

func NewRouteLongNameContainsShortNameNotice(routeID string, routeShortName string, routeLongName string, rowNumber int) *RouteLongNameContainsShortNameNotice {
	context := map[string]interface{}{
		"routeId":        routeID,
		"routeShortName": routeShortName,
		"routeLongName":  routeLongName,
		"csvRowNumber":   rowNumber,
	}
	return &RouteLongNameContainsShortNameNotice{
		BaseNotice: NewBaseNotice("route_long_name_contains_short_name", WARNING, context),
	}
}

// SameNameAndDescriptionForRouteNotice reports a route_desc that only repeats
// the route's name. The spec asks route_desc for information the name does not
// already carry; a duplicate costs the rider a line of screen and tells them
// nothing.
type SameNameAndDescriptionForRouteNotice struct {
	*BaseNotice
}

func NewSameNameAndDescriptionForRouteNotice(routeID string, routeDesc string, specifiedField string, rowNumber int) *SameNameAndDescriptionForRouteNotice {
	context := map[string]interface{}{
		"routeId":        routeID,
		"routeDesc":      routeDesc,
		"specifiedField": specifiedField,
		"csvRowNumber":   rowNumber,
	}
	return &SameNameAndDescriptionForRouteNotice{
		BaseNotice: NewBaseNotice("same_name_and_description_for_route", WARNING, context),
	}
}

// SameNameAndDescriptionForStopNotice reports a stop_desc that only repeats
// stop_name, for the same reason as the route case.
type SameNameAndDescriptionForStopNotice struct {
	*BaseNotice
}

func NewSameNameAndDescriptionForStopNotice(stopID string, stopDesc string, rowNumber int) *SameNameAndDescriptionForStopNotice {
	context := map[string]interface{}{
		"stopId":       stopID,
		"stopDesc":     stopDesc,
		"csvRowNumber": rowNumber,
	}
	return &SameNameAndDescriptionForStopNotice{
		BaseNotice: NewBaseNotice("same_name_and_description_for_stop", WARNING, context),
	}
}

// SameRouteAndAgencyUrlNotice reports a route_url pointing at the agency's own
// URL. route_url is meant to be the page for that route; when it is the
// agency's home page every route in the feed links to the same place and the
// field carries no information.
type SameRouteAndAgencyUrlNotice struct {
	*BaseNotice
}

func NewSameRouteAndAgencyUrlNotice(routeID string, routeURL string, routeRowNumber int, agencyID string, agencyName string, agencyRowNumber int) *SameRouteAndAgencyUrlNotice {
	context := map[string]interface{}{
		"routeId":            routeID,
		"routeUrl":           routeURL,
		"routeCsvRowNumber":  routeRowNumber,
		"agencyId":           agencyID,
		"agencyName":         agencyName,
		"agencyCsvRowNumber": agencyRowNumber,
	}
	return &SameRouteAndAgencyUrlNotice{
		BaseNotice: NewBaseNotice("same_route_and_agency_url", WARNING, context),
	}
}

// SameStopAndAgencyUrlNotice reports a stop_url pointing at the agency's own
// URL, which leaves the rider without the stop's page.
type SameStopAndAgencyUrlNotice struct {
	*BaseNotice
}

func NewSameStopAndAgencyUrlNotice(stopID string, stopName string, stopURL string, stopRowNumber int, agencyID string, agencyName string, agencyRowNumber int) *SameStopAndAgencyUrlNotice {
	context := map[string]interface{}{
		"stopId":             stopID,
		"stopName":           stopName,
		"stopUrl":            stopURL,
		"stopCsvRowNumber":   stopRowNumber,
		"agencyId":           agencyID,
		"agencyName":         agencyName,
		"agencyCsvRowNumber": agencyRowNumber,
	}
	return &SameStopAndAgencyUrlNotice{
		BaseNotice: NewBaseNotice("same_stop_and_agency_url", WARNING, context),
	}
}

// SameStopAndRouteUrlNotice reports a stop_url pointing at a route's URL. A
// stop is served by several routes, so linking it to one of them misdirects
// riders travelling on the others.
type SameStopAndRouteUrlNotice struct {
	*BaseNotice
}

func NewSameStopAndRouteUrlNotice(stopID string, stopName string, stopURL string, stopRowNumber int, routeID string, routeRowNumber int) *SameStopAndRouteUrlNotice {
	context := map[string]interface{}{
		"stopId":            stopID,
		"stopName":          stopName,
		"stopUrl":           stopURL,
		"stopCsvRowNumber":  stopRowNumber,
		"routeId":           routeID,
		"routeCsvRowNumber": routeRowNumber,
	}
	return &SameStopAndRouteUrlNotice{
		BaseNotice: NewBaseNotice("same_stop_and_route_url", WARNING, context),
	}
}
