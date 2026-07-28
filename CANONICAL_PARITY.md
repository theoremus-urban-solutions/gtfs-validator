# Parity with the Canonical GTFS Schedule Validator

Reference: <https://gtfs-validator.mobilitydata.org/rules.html> (177 rules)

| | count |
|---|---|
| Rules we emit | 201 |
| Canonical rules | 177 |
| Shared codes | 40 |
| Canonical rules not implemented | 137 |
| Our rules with no canonical equivalent | 161 |

The canonical validator is Java under Apache-2.0; this project is MIT. Its code
cannot be vendored without making this a mixed-license repository, so anything
adopted from it is reimplemented from the published rule descriptions, which are
statements about GTFS rather than copyrightable expression.
`scripts/gen_notice_descriptions.py` pulls the wording; the checks are our own.

## Renamed to canonical codes

Verified by comparing our implementation against the published rule text, then
applied with `scripts/rename_to_canonical.py`. The Go type and constructor names
were left alone; only the emitted code changed.

| was | now |
|---|---|
| `block_trips_overlap` | `block_trips_with_overlapping_stop_times` |
| `duplicate_header` | `duplicated_column` |
| `excessive_travel_speed` | `fast_travel_between_consecutive_stops` |
| `feed_expires_within_7_days` | `feed_expiration_date7_days` |
| `feed_expires_within_30_days` | `feed_expiration_date30_days` |
| `insufficient_shape_points` | `single_shape_point` |
| `missing_required_stop_name` | `missing_stop_name` |
| `missing_route_name` | `route_both_short_and_long_name_missing` |
| `multiple_records_in_single_record_file` | `more_than_one_entity` |
| `service_without_active_days` | `service_has_no_active_day_of_the_week` |
| `stop_time_decreasing_time` | `stop_time_with_arrival_before_previous_departure_time` |
| `trip_usability` | `unusable_trip` |
| `wrong_number_of_fields` | `invalid_row_length` |

Three canonical codes each absorb several of ours, matching how upstream reports
them:

| was | now |
|---|---|
| `duplicate_route_short_name`, `duplicate_route_long_name`, `duplicate_route_name_combination` | `duplicate_route_name` |
| `leading_whitespace`, `trailing_whitespace` | `leading_or_trailing_whitespaces` |
| `missing_trip_first_time`, `missing_trip_last_time` | `missing_trip_edge` |

The whitespace merge is a behaviour change: a value with whitespace on both
sides used to produce two notices and now produces one, which is what upstream
does. Notice deduplication in `notice.NoticeContainer` collapses the pair.

## Deliberately not renamed

These looked equivalent by name but check something different. Renaming them
would claim canonical compliance for a check that does not match.

| ours | canonical | why not |
|---|---|---|
| `validator_error` | `i_o_error` | ours is any validator failure, not I/O |
| `invalid_row` | `empty_row` | ours is a generic CSV parse failure |
| `invalid_field_format` | `unexpected_enum_value` | ours is any malformed field, not enums |
| `invalid_timepoint` | `missing_timepoint_value` | ours fires on a value outside 0/1; theirs on an absent value |
| `duplicate_route_long_name` | `route_long_name_contains_short_name` | ours compares two routes; theirs compares two fields of one route |
| `invalid_timezone` | `inconsistent_agency_timezone` | ours is an unparseable timezone; theirs is agencies disagreeing |
| `invalid_coordinate` | `invalid_geometry` | theirs is GeoJSON-specific |
| `shape_distance_inconsistent_with_geography` | `stop_too_far_from_shape` | different computations |
| `stop_without_service` | `stop_without_stop_time` | ours is pickup and drop-off both forbidden; theirs is a stop no trip references |
| `stop_name_contains_control_character` | `non_ascii_or_non_printable_char` | ours covers stop names only |
| `long_distance_transfer` | `transfer_distance_too_large` | ours triggers at 500 m, theirs at 10 km |
| `expired_service` | `expired_calendar` | ours triggers 30 days after expiry, theirs on any expiry |

`insufficient_stop_times` duplicates `unusable_trip` — both fire on a trip with
fewer than two stops. Only one was renamed; the other should be removed.

## Not implemented, by area

### Highest value first

These three are what the canonical validator reports on a feed that we report
nothing for, and `unsorted_stop_times` alone accounts for the bulk of it:

- `unsorted_stop_times`
- `mixed_case_recommended_field`
- `trip_coverage_not_active_for_next7_days`

### Core GTFS (85)

Ordinary rules over files we already parse.

**File & field structure (33)** — `empty_column_name`, `empty_row`,
`forbidden_arrival_or_departure_time`, `forbidden_continuous_pickup_drop_off`,
`forbidden_drop_off_type`, `forbidden_pickup_type`, `i_o_error`,
`invalid_character`, `invalid_float`, `invalid_input_files_in_subfolder`,
`invalid_integer`, `invalid_phone_number`, `invalid_time`, `malformed_json`,
`missing_bike_allowance`, `missing_feed_contact_email_and_url`,
`missing_pickup_or_drop_off_window`, `missing_recommended_file`,
`missing_required_element`, `mixed_case_recommended_field`, `new_line_in_value`,
`non_ascii_or_non_printable_char`, `number_out_of_range`, `point_near_origin`,
`point_near_pole`, `runtime_exception_in_loader_error`,
`runtime_exception_in_validator_error`, `start_and_end_range_equal`,
`start_and_end_range_out_of_order`, `thread_execution_error`, `too_many_rows`,
`u_r_i_syntax_error`, `unexpected_enum_value`

**Stops & stop_times (16)** — `fast_travel_between_far_stops`,
`missing_stop_times_record`, `missing_timepoint_value`,
`platform_without_parent_station`, `same_name_and_description_for_stop`,
`same_stop_and_agency_url`, `same_stop_and_route_url`,
`stop_access_specified_for_stop_with_no_parent_station`,
`stop_time_timepoint_without_times`,
`stop_time_with_only_arrival_or_departure_time`, `stop_without_stop_time`,
`stop_without_zone_id`, `transfer_with_invalid_trip_and_stop`,
`trip_headsign_matches_intermediate_stop`, `unsorted_stop_times`,
`unused_station`

**Routes, trips & agencies (13)** — `feed_info_lang_and_agency_lang_mismatch`,
`inconsistent_agency_lang`, `inconsistent_agency_timezone`,
`inconsistent_route_type_for_block_id`,
`inconsistent_route_type_for_in_seat_transfer`, `missing_required_agency_id`,
`route_long_name_contains_short_name`,
`route_networks_specified_in_more_than_one_file`,
`same_name_and_description_for_route`, `same_route_and_agency_url`,
`transfer_with_invalid_trip_and_route`,
`transfer_with_suspicious_mid_trip_in_seat`, `unused_trip`

**Shapes (11)** — `equal_shape_distance_diff_coordinates`,
`equal_shape_distance_diff_coordinates_distance_below_threshold`,
`equal_shape_distance_same_coordinates`, `forbidden_shape_dist_traveled`,
`stop_has_too_many_matches_for_shape`, `stop_too_far_from_shape`,
`stop_too_far_from_shape_using_user_distance`, `stops_match_shape_out_of_order`,
`trip_distance_exceeds_shape_distance`,
`trip_distance_exceeds_shape_distance_below_threshold`,
`trip_with_shape_dist_traveled_but_no_shape_distances`

**Calendar & service dates (10)** — `big_gap_in_service`, `expired_calendar`,
`feed_valid_beyond_total_service_window`, `future_calendar`, `future_feed`,
`invalid_date`, `missing_feed_info_date`, `service_extends_far_in_the_future`,
`service_window_outside_feed_period`, `trip_coverage_not_active_for_next7_days`

**Transfers (2)** — `transfer_distance_above_2_km`, `transfer_distance_too_large`

### Extensions (52)

Only relevant to feeds using these features. A feed of agency, stops, routes,
trips, stop_times and calendar needs none of them.

- **GTFS-Flex / GeoJSON (31)** — `locations.geojson`, booking rules, pickup and
  drop-off windows
- **GTFS-Fares v2 (12)** — fare media, fare products, transfer rules, timeframes
- **Pathways & levels (6)** — `bidirectional_exit_gate`, `missing_level_id`,
  `pathway_dangling_generic_node`, `pathway_loop`,
  `pathway_to_platform_with_boarding_areas`,
  `pathway_to_stop_with_access_outside_of_station_pathways`
- **Translations (3)** — `translation_foreign_key_violation`,
  `translation_unexpected_value`, `translation_unknown_table_name`

## Our rules with no canonical equivalent

161 codes, down from 226. Sixty-seven were removed as neither canonical nor
useful to any consumer — see below. What remains covers failure modes that
matter to consumers the canonical validator does not model — OpenTripPlanner drops trips on `TripDegenerate` and
`TripUndefinedService`, and has historically failed graph builds on unsupported
route types, none of which the canonical rule set flags. Keeping them is
deliberate; they are additional coverage, not divergence.

### Removed: 67 codes that served no consumer

Deleted because they are absent from the canonical rule set *and* correspond to
nothing OpenTripPlanner acts on. None were ERROR level.

- **Editorial opinions about service design (18)** — `low_service_usage`,
  `low_route_usage`, `low_trip_volume_next_7_days`, `low_transfer_opportunity`,
  `limited_service_variety`, `excessive_service_variety`,
  `excessive_route_pattern_variations`, `unusual_service_pattern`,
  `unbalanced_direction_trips`, `single_route_type_in_feed`,
  `high_route_type_diversity`, `agency_mixed_route_types`,
  `mostly_calendar_dates_services`, `overlapping_routes`,
  `same_origin_destination`, `block_multiple_routes`, `block_too_many_trips`,
  `long_trip_pattern`. A weekend-only service or a one-bus-a-day route is an
  operational choice, not a data defect.
- **Invented magnitude thresholds (24)** — every `very_*` and `unreasonable_*`
  pair. No defensible number backs "very long headway", and both directions
  fired, so any value was wrong to somebody.
- **Text cosmetics (13)** — `stop_name_too_long`, `route_long_name_too_long`,
  `stop_name_contains_html`, `stop_name_contains_url`, `stop_name_repeated_word`,
  `generic_stop_name`, and the headsign style checks. The canonical
  `mixed_case_recommended_field` covers this properly and is on the list above.
- **Broken or duplicated (12)** — chiefly `timepoint_without_times`, which fired
  on `timepoint=0` with times present. That is normal GTFS, so it emitted one
  notice per `stop_times.txt` row: 684,740 of them on a 30,000-trip feed.

Two validators became entirely dead and were deleted:
`stop_time_headsign_validator.go` (already unregistered, commented out for
hanging on large feeds) and `feed_expiration_validator.go` (never registered,
and a duplicate of `feed_expiration_date_validator.go`).

### Still outstanding

Length checks count bytes rather than runes (`len` on a Go string), so any
non-Latin script measures at two or three times its real length. The surviving
case is `route_short_name_too_long` in `route_name_validator.go` and
`route_consistency_validator.go`, where the 12-character limit is real and a
Cyrillic short name exceeds it at six characters.
