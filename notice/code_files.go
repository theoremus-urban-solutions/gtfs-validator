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
	// agency.txt
	"missing_agency_id":        agency,
	"invalid_agency_reference": {"routes.txt", "agency.txt"},
	"invalid_timezone":         {"agency.txt", "stops.txt"},
	"invalid_language_code":    {"agency.txt", "feed_info.txt"},

	// attributions.txt
	"attribution_all_roles":          attributions,
	"attribution_role_name_mismatch": attributions,
	"attribution_without_role":       attributions,
	"conflicting_attribution_scope":  attributions,
	"duplicate_attribution_scope":    attributions,
	"missing_attribution_contact":    attributions,
	"missing_attribution_role":       attributions,
	"multiple_attribution_scopes":    attributions,

	// calendar.txt
	"invalid_day_value":                     calendar,
	"invalid_service_date_range":            calendar,
	"service_has_no_active_day_of_the_week": calendar,

	// calendar_dates.txt
	"conflicting_calendar_exception": calendarDates,
	"duplicate_calendar_exception":   calendarDates,
	"invalid_exception_type":         calendarDates,

	// calendar.txt + calendar_dates.txt
	"expired_service":                          calendars,
	"service_expired":                          calendars,
	"service_never_active":                     calendars,
	"future_service":                           calendars,
	"service_expires_within_7_days":            calendars,
	"service_expires_within_30_days":           calendars,
	"missing_calendar_and_calendar_date_files": calendars,
	"no_service_defined":                       calendars,
	"no_service_date_found":                    calendars,
	"no_service_next_7_days":                   calendars,
	"insufficient_service_next_7_days":         calendars,
	"insufficient_service_next_30_days":        calendars,
	"undefined_service":                        {"trips.txt", "calendar.txt"},
	"unused_service":                           {"calendar.txt", "trips.txt"},

	// fares
	"invalid_fare_price":            fareAttrs,
	"invalid_payment_method":        fareAttrs,
	"invalid_currency_code":         fareAttrs,
	"unused_fare_attribute":         fareAttrs,
	"missing_fare_attributes":       fareAttrs,
	"invalid_transfer_duration":     fareAttrs,
	"unnecessary_transfer_duration": fareAttrs,
	"empty_fare_rule":               fareRules,
	"conflicting_fare_rule_fields":  fareRules,
	"undefined_zone":                {"fare_rules.txt", "stops.txt"},
	"unused_zone":                   {"stops.txt", "fare_rules.txt"},

	// feed_info.txt
	"feed_info_end_date_before_start_date": feedInfo,
	"feed_info_end_date_missing":           feedInfo,
	"missing_feed_info":                    feedInfo,
	"multiple_feed_info_entries":           feedInfo,
	"feed_expired":                         feedInfo,
	"expired_feed":                         feedInfo,
	"feed_expiration_date7_days":           feedInfo,
	"feed_expiration_date30_days":          feedInfo,
	"future_feed_start_date":               feedInfo,

	// frequencies.txt
	"invalid_exact_times":                     frequencies,
	"invalid_frequency_time_range":            frequencies,
	"invalid_headway":                         frequencies,
	"frequency_duration_shorter_than_headway": frequencies,
	"overlapping_frequency":                   frequencies,
	"cross_trip_frequency_overlap":            frequencies,

	// levels.txt
	"duplicate_level_index": levels,
	"unused_level":          levels,
	"missing_levels":        levels,

	// pathways.txt
	"duplicate_pathway":                  pathways,
	"inconsistent_bidirectional_pathway": pathways,
	"invalid_bidirectional":              pathways,
	"invalid_pathway_length":             pathways,
	"invalid_pathway_mode":               pathways,
	"invalid_stair_count":                pathways,
	"invalid_traversal_time":             pathways,
	"invalid_min_width":                  pathways,
	"pathway_to_same_stop":               pathways,
	"unexpected_bidirectional_gate":      pathways,

	// routes.txt
	"deprecated_route_type":                  routes,
	"invalid_route_type":                     routes,
	"duplicate_route_name":                   routes,
	"route_short_name_too_long":              routes,
	"route_both_short_and_long_name_missing": routes,
	"missing_route_agency_id":                routes,
	"same_name_and_description":              routes,
	"invalid_color":                          routes,
	"route_color_contrast":                   routes,
	"route_without_trips":                    routesTrips,

	// shapes.txt
	"decreasing_or_equal_shape_distance":         shapes,
	"decreasing_shape_distance":                  shapes,
	"shape_distance_decreasing":                  shapes,
	"shape_distance_not_increasing":              shapes,
	"shape_distance_not_starting_from_zero":      shapes,
	"equal_shape_distance":                       shapes,
	"negative_shape_distance":                    shapes,
	"large_shape_distance_jump":                  shapes,
	"unrealistic_shape_distance":                 shapes,
	"incomplete_shape_distance":                  shapes,
	"inconsistent_shape_distance":                shapes,
	"shape_distance_inconsistent_with_geography": shapes,
	"duplicate_shape_point":                      shapes,
	"duplicate_shape_sequence":                   shapes,
	"negative_shape_sequence":                    shapes,
	"non_increasing_shape_sequence":              shapes,
	"single_shape_point":                         shapes,
	"shape_point_outside_feed_bounds":            shapes,
	"unreasonably_long_shape_segment":            shapes,
	"unused_shape":                               shapes,
	"inconsistent_stop_time_shape_distance":      {"stop_times.txt", "shapes.txt"},

	// stops.txt
	"missing_stop_name":                    stops,
	"stop_name_contains_control_character": stops,
	"stop_name_missing_but_inherited":      stops,
	"invalid_latitude":                     stops,
	"invalid_longitude":                    stops,
	"invalid_coordinate":                   stops,
	"missing_coordinates":                  stops,
	"suspicious_coordinate":                stops,
	"invalid_location_type":                stops,
	"invalid_parent_station_reference":     stops,
	"invalid_parent_station_type":          stops,
	"missing_parent_station":               stops,
	"station_with_parent_station":          stops,
	"circular_station_reference":           stops,
	"orphaned_station":                     stops,
	"child_station_too_far_from_parent":    stops,
	"invalid_wheelchair_boarding":          stops,
	"stop_without_service":                 {"stops.txt", "stop_times.txt"},

	// stop_times.txt
	"missing_arrival_time":                                  stopTimes,
	"missing_departure_time":                                stopTimes,
	"missing_trip_edge":                                     stopTimes,
	"stop_time_arrival_after_departure":                     stopTimes,
	"stop_time_with_arrival_before_previous_departure_time": stopTimes,
	"invalid_timepoint":                                     stopTimes,
	"duplicate_stop_sequence":                               stopTimes,
	"negative_stop_sequence":                                stopTimes,
	"non_increasing_stop_sequence":                          stopTimes,
	"stop_sequence_gap":                                     stopTimes,
	"duplicate_stop_in_trip":                                stopTimes,
	"consecutive_duplicate_stops":                           stopTimes,
	"insufficient_stop_times":                               stopTimes,
	"all_stops_no_pickup":                                   stopTimes,
	"all_stops_no_drop_off":                                 stopTimes,
	"first_stop_no_pickup":                                  stopTimes,
	"last_stop_no_drop_off":                                 stopTimes,
	"decreasing_or_equal_stop_time_distance":                stopTimes,
	"fast_travel_between_consecutive_stops":                 stopTimesTrip,

	// transfers.txt
	"duplicate_transfer":             transfers,
	"invalid_transfer_type":          transfers,
	"invalid_transfers":              transfers,
	"transfer_to_same_stop":          transfers,
	"missing_min_transfer_time":      transfers,
	"negative_min_transfer_time":     transfers,
	"unnecessary_min_transfer_time":  transfers,
	"unreasonable_min_transfer_time": transfers,

	// trips.txt
	"invalid_direction_id":                    trips,
	"invalid_bikes_allowed":                   trips,
	"invalid_bikes_allowed_value":             trips,
	"missing_bikes_allowed_for_ferry":         trips,
	"invalid_wheelchair_accessible":           trips,
	"block_service_mismatch":                  trips,
	"block_trips_with_overlapping_stop_times": trips,
	"unusable_trip":                           trips,
	"no_trips_next_7_days":                    trips,

	// network-wide geometry, spanning stops and their connections
}
