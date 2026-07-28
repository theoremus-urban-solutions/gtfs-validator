# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed — validation scope reworked (breaking)

The emitted rule set is now reconciled against the
[Canonical GTFS Schedule Validator](https://gtfs-validator.mobilitydata.org/rules.html).
**Notice codes have been added, removed and renamed**, and severities have
moved. Anything keying on specific codes needs the mapping tables below.

| | before | after |
|---|---|---|
| Codes emitted | 201 | 176 |
| — canonical | 40 | **133 (all of them)** |
| — our own | 161 | 43 |
| Codes that are ERROR but not canonical | 86 | **0** |
| Severities disagreeing with canonical | 8 | **0** |
| Registered validators | 58 | 50 |

Three things drive the change:

- **Every in-scope canonical rule is now implemented.** In scope means the 181
  published rules less 4 deprecated upstream, 27 GTFS-Flex, 11 GTFS-Fares v2 and
  6 that are artefacts of the canonical validator's own execution model.
- **Only canonical rules may be ERROR.** A code MobilityData does not define is
  our opinion, and an opinion should not fail someone's feed. 86 of our own
  codes were ERROR; a feed could fail on checks nobody else recognises.
- **Duplicate and superseded codes were retired.** Most removals below are not
  lost coverage: the same check is now reported under the canonical code, and
  the mapping is in `CANONICAL_PARITY.md`. Notably the ~15 per-field
  `invalid_*` enum codes collapse onto `unexpected_enum_value`, which also fixes
  a hole — the old checks were guarded by a successful `strconv.Atoi`, so an
  enum holding a non-numeric value reported nothing at all.

`python3 scripts/scope_audit.py` reproduces every number here.
`docs/validation-scope/VALIDATION_SCOPE_PROPOSAL.md` is the full argument, and
`VALIDATOR_RULES.md` lists every code by validator.

### Fixed

- `route_short_name_too_long` counted bytes rather than runes, so a Cyrillic or
  Greek short name tripped the 12-character limit at six characters.
- `date_trips_validator` re-read `trips.txt` once per service, which was
  quadratic in the size of the feed.
- `file_structure_validator` was never constructed, so `csv_parsing_failed` and
  `unknown_column` counted as implemented while never being emitted. It is now
  registered.
- `stop_sequence_gap` fired on gaps in `stop_sequence`, which the spec permits —
  values must increase, not be contiguous. Removed as a pure false positive.
- `stop_name_missing_but_inherited` downgraded a missing required `stop_name`
  when the parent station had one. Nothing in GTFS inherits `stop_name`.

### Performance

- Shape-matching checks (`stop_too_far_from_shape` and the rest of the
  geometric set) run only in comprehensive mode, behind `EnableShapeGeometry`.
- Removed the whole-feed graph build in `network_topology_validator`, the most
  expensive validator in the suite, along with two redundant passes over
  `shapes.txt` and several duplicated passes over `stop_times.txt`.

### Migration

#### Codes removed

- `calendar_end_before_start`
- `calendar_no_days_selected`
- `close_stops_not_possible_transfer`
- `decreasing_or_equal_shape_distance`
- `duplicate_calendar_date`
- `duplicate_shape_point`
- `duplicate_shape_sequence`
- `duplicate_stop_sequence`
- `equal_shape_distance`
- `expired_feed`
- `expired_service`
- `feed_expired`
- `feed_info_end_date_before_start_date`
- `feed_info_end_date_missing`
- `fragmented_network`
- `future_feed_start_date`
- `future_service`
- `incomplete_shape_distance`
- `inconsistent_bidirectional_transfer`
- `inconsistent_shape_distance`
- `insufficient_coordinate_precision`
- `insufficient_service_next_30_days`
- `insufficient_service_next_7_days`
- `insufficient_stop_times`
- `invalid_agency_reference`
- `invalid_bidirectional`
- `invalid_bikes_allowed`
- `invalid_bikes_allowed_value`
- `invalid_coordinate`
- `invalid_currency_code`
- `invalid_date_format`
- `invalid_day_value`
- `invalid_direction_id`
- `invalid_exact_times`
- `invalid_exception_type`
- `invalid_fare_price`
- `invalid_frequency_time_range`
- `invalid_headway`
- `invalid_latitude`
- `invalid_location_type`
- `invalid_longitude`
- `invalid_min_width`
- `invalid_parent_station_reference`
- `invalid_parent_station_type`
- `invalid_pathway_length`
- `invalid_pathway_mode`
- `invalid_payment_method`
- `invalid_route_type`
- `invalid_row`
- `invalid_service_date_range`
- `invalid_stair_count`
- `invalid_time_format`
- `invalid_timepoint`
- `invalid_transfer_duration`
- `invalid_transfer_type`
- `invalid_transfers`
- `invalid_traversal_time`
- `invalid_wheelchair_accessible`
- `invalid_wheelchair_boarding`
- `isolated_stop`
- `large_shape_distance_jump`
- `long_distance_transfer`
- `long_zone_id`
- `low_network_connectivity`
- `missing_agency_id`
- `missing_arrival_time`
- `missing_attribution_role`
- `missing_bikes_allowed_for_ferry`
- `missing_coordinates`
- `missing_departure_time`
- `missing_fare_attributes`
- `missing_feed_info`
- `missing_levels`
- `missing_parent_station`
- `missing_route_agency_id`
- `multiple_feed_info_entries`
- `negative_shape_distance`
- `negative_shape_sequence`
- `negative_stop_sequence`
- `no_service_date_found`
- `no_service_defined`
- `no_service_next_7_days`
- `no_trips_next_7_days`
- `non_increasing_shape_sequence`
- `non_increasing_stop_sequence`
- `orphaned_station`
- `pathway_to_same_stop`
- `same_name_and_description`
- `service_expired`
- `service_expires_within_30_days`
- `service_expires_within_7_days`
- `service_never_active`
- `service_without_definition`
- `shape_distance_decreasing`
- `shape_distance_inconsistent_with_geography`
- `shape_distance_not_increasing`
- `shape_distance_not_starting_from_zero`
- `shape_point_outside_feed_bounds`
- `small_network_component`
- `stop_name_contains_html`
- `stop_name_contains_url`
- `stop_name_missing_but_inherited`
- `stop_sequence_gap`
- `suspicious_coordinate`
- `undefined_service`
- `undefined_zone`
- `unexpected_bidirectional_gate`
- `unrealistic_shape_distance`
- `unrealistic_transfer_time`
- `unreasonable_level_index`
- `unreasonable_max_slope`
- `unreasonably_long_shape_segment`
- `unusual_route_type_combination`
- `validator_error`
- `very_long_transfer_time`
- `very_short_transfer_time`
- `whitespace_only_field`
- `zone_id_same_as_stop_id`

#### Codes added

- `bidirectional_exit_gate` (ERROR)
- `big_gap_in_service` (INFO)
- `empty_column_name` (ERROR)
- `empty_row` (WARNING)
- `equal_shape_distance_diff_coordinates` (ERROR)
- `equal_shape_distance_diff_coordinates_distance_below_threshold` (WARNING)
- `equal_shape_distance_same_coordinates` (WARNING)
- `expired_calendar` (WARNING)
- `fast_travel_between_far_stops` (WARNING)
- `feed_info_lang_and_agency_lang_mismatch` (WARNING)
- `feed_valid_beyond_total_service_window` (INFO)
- `forbidden_arrival_or_departure_time` (ERROR)
- `forbidden_continuous_pickup_drop_off` (ERROR)
- `forbidden_drop_off_type` (ERROR)
- `forbidden_pickup_type` (ERROR)
- `forbidden_shape_dist_traveled` (ERROR)
- `future_calendar` (INFO)
- `future_feed` (INFO)
- `inconsistent_agency_lang` (WARNING)
- `inconsistent_agency_timezone` (ERROR)
- `inconsistent_route_type_for_block_id` (WARNING)
- `inconsistent_route_type_for_in_seat_transfer` (WARNING)
- `invalid_character` (ERROR)
- `invalid_currency` (ERROR)
- `invalid_currency_amount` (ERROR)
- `invalid_date` (ERROR)
- `invalid_float` (ERROR)
- `invalid_integer` (ERROR)
- `invalid_phone_number` (ERROR)
- `invalid_time` (ERROR)
- `location_without_parent_station` (ERROR)
- `missing_bike_allowance` (WARNING)
- `missing_feed_contact_email_and_url` (WARNING)
- `missing_feed_info_date` (WARNING)
- `missing_level_id` (ERROR)
- `missing_recommended_file` (WARNING)
- `missing_required_agency_id` (ERROR)
- `missing_stop_times_record` (ERROR)
- `missing_timepoint_value` (WARNING)
- `mixed_case_recommended_field` (WARNING)
- `new_line_in_value` (ERROR)
- `non_ascii_or_non_printable_char` (WARNING)
- `number_out_of_range` (ERROR)
- `pathway_dangling_generic_node` (WARNING)
- `pathway_loop` (WARNING)
- `pathway_to_platform_with_boarding_areas` (ERROR)
- `pathway_to_stop_with_access_outside_of_station_pathways` (ERROR)
- `pathway_to_wrong_location_type` (ERROR)
- `pathway_unreachable_location` (ERROR)
- `platform_without_parent_station` (INFO)
- `point_near_origin` (ERROR)
- `point_near_pole` (ERROR)
- `route_long_name_contains_short_name` (WARNING)
- `runtime_exception_in_validator_error` (ERROR)
- `same_name_and_description_for_route` (WARNING)
- `same_name_and_description_for_stop` (WARNING)
- `same_route_and_agency_url` (WARNING)
- `same_stop_and_agency_url` (WARNING)
- `same_stop_and_route_url` (WARNING)
- `service_extends_far_in_the_future` (INFO)
- `service_window_outside_feed_period` (INFO)
- `start_and_end_range_equal` (ERROR)
- `start_and_end_range_out_of_order` (ERROR)
- `stop_access_specified_for_incorrect_location` (ERROR)
- `stop_access_specified_for_stop_with_no_parent_station` (ERROR)
- `stop_has_too_many_matches_for_shape` (WARNING)
- `stop_time_timepoint_without_times` (ERROR)
- `stop_time_with_only_arrival_or_departure_time` (ERROR)
- `stop_too_far_from_shape` (WARNING)
- `stop_too_far_from_shape_using_user_distance` (WARNING)
- `stop_without_location` (ERROR)
- `stop_without_stop_time` (WARNING)
- `stop_without_zone_id` (INFO)
- `stops_match_shape_out_of_order` (WARNING)
- `transfer_distance_above_2_km` (INFO)
- `transfer_distance_too_large` (WARNING)
- `transfer_with_invalid_stop_location_type` (ERROR)
- `transfer_with_invalid_trip_and_route` (ERROR)
- `transfer_with_invalid_trip_and_stop` (ERROR)
- `transfer_with_suspicious_mid_trip_in_seat` (WARNING)
- `translation_foreign_key_violation` (ERROR)
- `translation_unexpected_value` (ERROR)
- `translation_unknown_table_name` (WARNING)
- `trip_coverage_not_active_for_next7_days` (WARNING)
- `trip_distance_exceeds_shape_distance` (ERROR)
- `trip_distance_exceeds_shape_distance_below_threshold` (WARNING)
- `trip_headsign_matches_intermediate_stop` (INFO)
- `trip_with_shape_dist_traveled_but_no_shape_distances` (INFO)
- `unexpected_enum_value` (WARNING)
- `unsorted_stop_times` (INFO)
- `unused_station` (INFO)
- `unused_trip` (WARNING)
- `wrong_parent_location_type` (ERROR)

#### Severity changed

| code | was | now |
|---|---|---|
| `all_stops_no_drop_off` | ERROR | WARNING |
| `all_stops_no_pickup` | ERROR | WARNING |
| `attribution_without_role` | ERROR | WARNING |
| `block_service_mismatch` | ERROR | WARNING |
| `circular_station_reference` | ERROR | WARNING |
| `conflicting_attribution_scope` | ERROR | WARNING |
| `conflicting_calendar_exception` | ERROR | WARNING |
| `duplicate_composite_key` | ERROR | WARNING |
| `duplicate_level_index` | ERROR | WARNING |
| `duplicate_transfer` | ERROR | WARNING |
| `empty_file` | WARNING | ERROR |
| `feed_expiration_date7_days` | ERROR | WARNING |
| `frequency_duration_shorter_than_headway` | ERROR | WARNING |
| `invalid_field_format` | ERROR | WARNING |
| `invalid_language_code` | WARNING | ERROR |
| `missing_min_transfer_time` | ERROR | WARNING |
| `more_than_one_entity` | ERROR | WARNING |
| `negative_min_transfer_time` | ERROR | WARNING |
| `service_has_no_active_day_of_the_week` | ERROR | WARNING |
| `single_shape_point` | ERROR | WARNING |
| `stop_time_arrival_after_departure` | ERROR | WARNING |
| `stop_without_service` | ERROR | WARNING |
| `unusable_trip` | ERROR | WARNING |


### Added
- **Enhanced Error Descriptions**: Added comprehensive, user-friendly descriptions to all validation notices in both JSON and HTML outputs
- **Centralized Description System**: Created `notice_descriptions.go` with 180+ detailed descriptions covering all validation categories
- **Memory Pooling System**: Comprehensive memory pools for CSV parsing to reduce garbage collection overhead
- **Streaming CSV Parser**: High-performance streaming parser for processing massive CSV files (2-4M rows/sec)
- **Structured Logging**: JSON and text formatters with configurable levels (DEBUG, INFO, WARN, ERROR)
- **Configuration Validation**: Automatic config sanitization and bounds checking with detailed error messages
- **Performance Benchmarking**: Built-in benchmarking documentation and performance monitoring
- **Memory Optimization Examples**: Complete examples for processing large feeds with minimal memory usage
- **Streaming Validation**: Context-aware streaming validation with real-time progress reporting
- **Advanced Examples**: Streaming CSV processing, config validation, and large feed optimization examples
- Comprehensive testing infrastructure with unit, integration, and CLI tests
- Performance benchmarks for validation operations
- Thread-safe notice container implementation
- GitHub Actions CI/CD pipeline with multi-platform testing
- Community contribution guidelines and templates
- Development tooling (Makefile, linting configuration)
- Cross-platform binary builds
- Cobra-based CLI with subcommands, help, and autocompletion
- Missing validator test coverage (5 new test files created)

### Changed
- **NoticeGroup Structure**: Added `Description` field to `NoticeGroup` struct for comprehensive error descriptions
- **JSON Output Enhancement**: All JSON validation reports now include detailed error descriptions
- **HTML Report Enhancement**: HTML reports now include comprehensive error descriptions for better user experience
- **CSV Parser Integration**: Integrated memory pools into existing CSV parser for automatic memory optimization
- **Modern Go Practices**: Replaced deprecated `ioutil` functions with modern `os` equivalents throughout codebase
- **Enhanced Documentation**: Updated README, CLI help, and API docs to reflect streaming and memory features
- **Performance Improvements**: Memory pooling reduces GC pressure, streaming parser enables constant memory usage
- Improved error handling and context propagation
- CLI interface updated to use Cobra framework for better UX
- Documentation updated with modern CLI examples
- README enhanced to highlight comprehensive validation coverage (294+ rules vs ~60 official)

### Fixed
- GTFS time validation now correctly supports late-night service times (25:30:00+)
- Time parsing no longer rejects valid GTFS times beyond 24:00:00
- Thread safety issues in concurrent validation
- MissingColumnValidator test fixed to handle empty files correctly
- BikesAllowanceValidator test fixed with proper CSV formatting for empty fields
- CoordinateValidator test corrected with accurate expected notice counts
- CLI tests updated to match actual error message formats

### Performance
- **2-4M rows/sec** sustained throughput on large GTFS files (tested with Sofia GTFS 588K+ records)
- **Constant memory usage** regardless of file size with streaming CSV processing
- **Memory pool optimization** reduces garbage collection overhead during CSV parsing
- **Streaming processing** enables validation of feeds with millions of records without OOM errors

## [1.0.0] - Initial Release

### Added
- Complete GTFS validation library for Go
- 60+ validators covering all GTFS specification requirements
- Multiple validation modes (performance, default, comprehensive)
- CLI tool with rich output formats (console, JSON, summary)
- Context support for cancellation and timeouts
- Progress reporting with callbacks
- Configurable worker pools for parallel processing
- Comprehensive error reporting with notice system
- Examples for library usage, advanced features, and web API integration

### Core Features
- **File Structure Validation**: Required files, CSV format, encoding
- **Data Format Validation**: Field types, ranges, patterns
- **Entity Validation**: Routes, stops, trips, agencies consistency
- **Relationship Validation**: Foreign keys, stop sequences, service consistency
- **Business Logic Validation**: Travel speeds, transfers, frequency overlaps
- **Accessibility Validation**: Pathways, wheelchair access
- **Geospatial Validation**: Coordinate analysis, geographic clustering

### Performance
- Optimized for large feeds with parallel processing
- Memory efficient parsing and validation
- Configurable resource limits
- Benchmark results: ~10-12 seconds for performance mode on large feeds

### Documentation
- Complete API documentation
- Usage examples and tutorials
- CLI reference guide
- Contributing guidelines