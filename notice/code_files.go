package notice

// This file maps notice codes to the GTFS files they concern, and normalises
// the row number a notice points at.
//
// Notice contexts are free-form maps built independently by ~300 constructors,
// so the row number lives under one of a dozen key names and the file is
// usually absent entirely. The helpers here give the report layer a single
// place to resolve both.

// rowNumberKeys lists the context keys that hold a row number, in the order
// they should be preferred. The first key present on a notice wins.
// Notices that compare two records (a duplicate, an overlap, a transfer) carry
// both rows; the first is the one to report.
var rowNumberKeys = []string{
	"csvRowNumber",
	"rowNumber",
	"csvRowNumber1",
	"rowNumber1",
	"firstRowNumber",
	"firstRow",
	"fromRowNumber",
	"trip1RowNumber",
	"prevCsvRowNumber",
	"duplicateRowNumber",
	"duplicateRow",
	"csvRowNumber2",
	"rowNumber2",
	"trip2RowNumber",
	"toRowNumber",
	"lastRowNumber",
}

// filenameKey is the context key that holds a per-notice file name.
const filenameKey = "filename"

// LineNumber returns the row number a notice points at, and whether one was
// found. Notices that describe a whole route, service or feed have no single
// row and return false.
func LineNumber(context map[string]interface{}) (int, bool) {
	for _, key := range rowNumberKeys {
		value, ok := context[key]
		if !ok {
			continue
		}
		if row, ok := toInt(value); ok && row > 0 {
			return row, true
		}
	}
	return 0, false
}

// toInt coerces the numeric types a context value may hold. Values are ints
// in process but arrive as float64 when a report is read back from JSON.
func toInt(value interface{}) (int, bool) {
	switch v := value.(type) {
	case int:
		return v, true
	case int32:
		return int(v), true
	case int64:
		return int(v), true
	case float64:
		return int(v), true
	case float32:
		return int(v), true
	default:
		return 0, false
	}
}

// FileName returns the GTFS file a single notice concerns, and whether it
// could be determined. A file name carried on the notice itself wins over the
// per-code mapping.
//
// A code that spans several files still resolves to one, because the mapping
// lists the file a notice's row belongs to first. Codes with no row and no
// single file — feed-wide summaries and network-shape checks — resolve to
// nothing; use AffectedFilesForCode for those.
func FileName(code string, context map[string]interface{}) (string, bool) {
	if name, ok := context[filenameKey].(string); ok && name != "" {
		return name, true
	}
	files := AffectedFilesForCode(code)
	if len(files) == 1 {
		return files[0], true
	}
	if _, hasRow := LineNumber(context); hasRow && len(files) > 0 {
		return files[0], true
	}
	return "", false
}

// AffectedFilesForCode returns the GTFS files a notice code concerns, the file
// its rows belong to first. It returns nil for structural codes that can apply
// to any file — those carry their file on the notice itself — and for
// feed-wide summary notices.
func AffectedFilesForCode(code string) []string {
	return codeFiles[code]
}

var (
	agency        = []string{"agency.txt"}
	attributions  = []string{"attributions.txt"}
	calendar      = []string{"calendar.txt"}
	calendarDates = []string{"calendar_dates.txt"}
	calendars     = []string{"calendar.txt", "calendar_dates.txt"}
	fareAttrs     = []string{"fare_attributes.txt"}
	fareRules     = []string{"fare_rules.txt"}
	feedInfo      = []string{"feed_info.txt"}
	frequencies   = []string{"frequencies.txt"}
	levels        = []string{"levels.txt"}
	pathways      = []string{"pathways.txt"}
	routes        = []string{"routes.txt"}
	routesTrips   = []string{"routes.txt", "trips.txt"}
	shapes        = []string{"shapes.txt"}
	stops         = []string{"stops.txt"}
	stopTimes     = []string{"stop_times.txt"}
	stopTimesTrip = []string{"stop_times.txt", "trips.txt"}
	transfers     = []string{"transfers.txt"}
	trips         = []string{"trips.txt"}
	network       = []string{"stops.txt", "transfers.txt"}
)

// codeFiles maps every notice code to the files it concerns. Codes absent from
// this map either carry their file in the notice context (structural checks
// that run over every file) or describe the feed as a whole (summaries).
var codeFiles = map[string][]string{
	// Added by the scope rework. Codes that carry their own filename in the
	// notice context are deliberately absent: FileName() resolves those from
	// the notice itself.
	"inconsistent_agency_lang":                agency,
	"feed_info_lang_and_agency_lang_mismatch": {"agency.txt", "feed_info.txt"},
	"future_feed":                                             feedInfo,
	"missing_feed_info_date":                                  feedInfo,
	"missing_feed_contact_email_and_url":                      feedInfo,
	"big_gap_in_service":                                      calendars,
	"expired_calendar":                                        calendars,
	"future_calendar":                                         calendars,
	"service_extends_far_in_the_future":                       calendars,
	"feed_valid_beyond_total_service_window":                  {"feed_info.txt", "calendar.txt"},
	"service_window_outside_feed_period":                      {"feed_info.txt", "calendar.txt"},
	"bidirectional_exit_gate":                                 pathways,
	"pathway_loop":                                            pathways,
	"pathway_dangling_generic_node":                           pathways,
	"pathway_unreachable_location":                            pathways,
	"pathway_to_wrong_location_type":                          {"pathways.txt", "stops.txt"},
	"pathway_to_platform_with_boarding_areas":                 {"pathways.txt", "stops.txt"},
	"pathway_to_stop_with_access_outside_of_station_pathways": {"pathways.txt", "stops.txt"},
	"missing_level_id":                                        {"stops.txt", "levels.txt"},
	"route_long_name_contains_short_name":                     routes,
	"same_name_and_description_for_route":                     routes,
	"same_route_and_agency_url":                               {"routes.txt", "agency.txt"},
	"same_name_and_description_for_stop":                      stops,
	"same_stop_and_agency_url":                                {"stops.txt", "agency.txt"},
	"same_stop_and_route_url":                                 {"stops.txt", "routes.txt"},
	"platform_without_parent_station":                         stops,
	"stop_without_zone_id":                                    stops,
	"stop_access_specified_for_incorrect_location":            stops,
	"stop_access_specified_for_stop_with_no_parent_station":   stops,
	"unused_shape":                                            {"shapes.txt", "trips.txt"},
	"stop_too_far_from_shape":                                 {"shapes.txt", "stop_times.txt"},
	"stop_too_far_from_shape_using_user_distance":             {"shapes.txt", "stop_times.txt"},
	"stops_match_shape_out_of_order":                          {"shapes.txt", "stop_times.txt"},
	"stop_has_too_many_matches_for_shape":                     {"shapes.txt", "stop_times.txt"},
	"trip_distance_exceeds_shape_distance":                    {"shapes.txt", "stop_times.txt"},
	"trip_distance_exceeds_shape_distance_below_threshold":    {"shapes.txt", "stop_times.txt"},
	"fast_travel_between_far_stops":                           stopTimesTrip,
	"trip_headsign_matches_intermediate_stop":                 stopTimesTrip,
	"trip_with_shape_dist_traveled_but_no_shape_distances":    stopTimesTrip,
	"trip_coverage_not_active_for_next7_days":                 trips,
	"inconsistent_route_type_for_block_id":                    routesTrips,
	"missing_bike_allowance":                                  routesTrips,
	"inconsistent_route_type_for_in_seat_transfer":            {"transfers.txt", "routes.txt"},
	"transfer_with_suspicious_mid_trip_in_seat":               {"transfers.txt", "stop_times.txt"},
	"transfer_distance_too_large":                             network,
	"transfer_distance_above_2_km":                            network,
	"translation_foreign_key_violation":                       {"translations.txt"},
	"translation_unexpected_value":                            {"translations.txt"},
	"translation_unknown_table_name":                          {"translations.txt"},

	// agency.txt
	"invalid_timezone":             {"agency.txt", "stops.txt"},
	"invalid_language_code":        {"agency.txt", "feed_info.txt"},
	"inconsistent_agency_timezone": {"agency.txt"},

	// attributions.txt
	"attribution_all_roles":          attributions,
	"attribution_role_name_mismatch": attributions,
	"attribution_without_role":       attributions,
	"conflicting_attribution_scope":  attributions,
	"duplicate_attribution_scope":    attributions,
	"missing_attribution_contact":    attributions,
	"multiple_attribution_scopes":    attributions,

	// calendar.txt
	"service_has_no_active_day_of_the_week": calendar,

	// calendar_dates.txt
	"conflicting_calendar_exception": calendarDates,
	"duplicate_calendar_exception":   calendarDates,

	// calendar.txt + calendar_dates.txt
	"missing_calendar_and_calendar_date_files": calendars,
	"unused_service": {"calendar.txt", "trips.txt"},

	// fares
	"unused_fare_attribute":         fareAttrs,
	"unnecessary_transfer_duration": fareAttrs,
	"empty_fare_rule":               fareRules,
	"conflicting_fare_rule_fields":  fareRules,
	"unused_zone":                   {"stops.txt", "fare_rules.txt"},

	// feed_info.txt
	"feed_expiration_date7_days":  feedInfo,
	"feed_expiration_date30_days": feedInfo,

	// frequencies.txt
	"frequency_duration_shorter_than_headway": frequencies,
	"overlapping_frequency":                   frequencies,
	"cross_trip_frequency_overlap":            frequencies,

	// levels.txt
	"duplicate_level_index": levels,
	"unused_level":          levels,

	// pathways.txt
	"duplicate_pathway":                  pathways,
	"inconsistent_bidirectional_pathway": pathways,

	// routes.txt
	"deprecated_route_type":                  routes,
	"duplicate_route_name":                   routes,
	"route_short_name_too_long":              routes,
	"route_both_short_and_long_name_missing": routes,
	"invalid_color":                          routes,
	"route_color_contrast":                   routes,
	"route_without_trips":                    routesTrips,

	// shapes.txt
	"decreasing_shape_distance":                                      shapes,
	"equal_shape_distance_same_coordinates":                          shapes,
	"equal_shape_distance_diff_coordinates":                          shapes,
	"equal_shape_distance_diff_coordinates_distance_below_threshold": shapes,
	"single_shape_point":                                             shapes,
	"inconsistent_stop_time_shape_distance":                          {"stop_times.txt", "shapes.txt"},

	// stops.txt
	"missing_stop_name":                    stops,
	"stop_name_contains_control_character": stops,
	"station_with_parent_station":          stops,
	"circular_station_reference":           stops,
	"child_station_too_far_from_parent":    stops,
	"stop_without_service":                 {"stops.txt", "stop_times.txt"},
	"stop_without_location":                stops,
	"location_without_parent_station":      stops,
	"wrong_parent_location_type":           stops,
	"unused_station":                       stops,
	"stop_without_stop_time":               {"stops.txt", "stop_times.txt"},

	// stop_times.txt
	"missing_trip_edge":                                     stopTimes,
	"stop_time_arrival_after_departure":                     stopTimes,
	"stop_time_with_arrival_before_previous_departure_time": stopTimes,
	"duplicate_stop_in_trip":                                stopTimes,
	"consecutive_duplicate_stops":                           stopTimes,
	"all_stops_no_pickup":                                   stopTimes,
	"all_stops_no_drop_off":                                 stopTimes,
	"first_stop_no_pickup":                                  stopTimes,
	"last_stop_no_drop_off":                                 stopTimes,
	"decreasing_or_equal_stop_time_distance":                stopTimes,
	"fast_travel_between_consecutive_stops":                 stopTimesTrip,
	"unsorted_stop_times":                                   stopTimes,
	"stop_time_with_only_arrival_or_departure_time":         stopTimes,
	"stop_time_timepoint_without_times":                     stopTimes,
	"missing_timepoint_value":                               stopTimes,
	"forbidden_arrival_or_departure_time":                   stopTimes,
	"forbidden_shape_dist_traveled":                         stopTimes,
	"forbidden_pickup_type":                                 stopTimes,
	"forbidden_drop_off_type":                               stopTimes,
	"forbidden_continuous_pickup_drop_off":                  {"stop_times.txt", "routes.txt"},
	"missing_stop_times_record":                             stopTimes,

	// transfers.txt
	"duplicate_transfer":                       transfers,
	"transfer_to_same_stop":                    transfers,
	"missing_min_transfer_time":                transfers,
	"negative_min_transfer_time":               transfers,
	"unnecessary_min_transfer_time":            transfers,
	"unreasonable_min_transfer_time":           transfers,
	"transfer_with_invalid_stop_location_type": {"transfers.txt", "stops.txt"},
	"transfer_with_invalid_trip_and_route":     {"transfers.txt", "trips.txt"},
	"transfer_with_invalid_trip_and_stop":      {"transfers.txt", "stop_times.txt"},

	// trips.txt
	"block_service_mismatch":                  trips,
	"block_trips_with_overlapping_stop_times": trips,
	"unusable_trip":                           trips,
	"unused_trip":                             {"trips.txt", "stop_times.txt"},

	// network-wide geometry, spanning stops and their connections
}
