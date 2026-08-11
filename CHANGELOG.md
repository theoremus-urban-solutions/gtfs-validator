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
| Codes emitted | 201 | 177 |
| — canonical | 40 | **135 (all of them)** |
| — our own | 161 | 42 |
| Codes that are ERROR but not canonical | 86 | **0** |
| Severities disagreeing with canonical | 8 | **0** |
| Registered validators | 58 | 54 |

Three things drive the change:

- **Every in-scope canonical rule is now implemented and registered.** In scope
  means the 181 published rules less 4 deprecated upstream, 26 GTFS-Flex, 11
  GTFS-Fares v2 and 5 that are artefacts of the canonical validator's own
  execution model. "Registered" is load-bearing — see below.
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

**Parity is now verified by running the canonical validator, not by comparing
code names.** `scripts/parity_gate.py` builds a feed per defect shape, runs both
validators over each, and fails on any disagreement about a canonical rule — or
on any disagreement about whether the feed is valid at all, which catches a
canonical rule re-badged under a fork-owned name however it is spelled. The
name-based audit could not see that, and was fooled by it twice; see below.

### Removed — validation modes and the parsed-feed cache (breaking)

`WithValidationMode`, the `ValidationMode` type and its three constants,
`Config.ValidationMode`, `WithCaching`, `Config.EnableCaching` and the CLI's
`--mode`/`-m` flag are all gone. There is no replacement knob. Every registered
validator runs on every feed.

**Modes.** The presets were defended by a cost that measurement does not
support: the comprehensive set is about 1.15x the old default on the largest
feed available (Sofia, 685k stop times, 8.59s to 9.86s), not the "2+ minutes"
the README claimed or the "5-30 minutes" in `BENCHMARKS.md`. Neither figure was
ever a measurement. What the split did buy was silence — performance mode ran 24
of the 54 validators, so a feed could pass having never been checked against 30
rules this tool implements. Callers on performance mode gain those 30 checks and
should expect new findings.

**The cache.** It was consulted by 3 of the ~50 validators; 26 validator files
read `stop_times.txt` directly and bypassed it. It saved a constant 3.98 GB of
allocation (2.5-3.4%) regardless of how much validation ran, with no wall-time
or peak-RSS benefit. Its three cached paths were also worse than the direct ones
they shadowed: they could not carry a row number, so row-bearing notices
reported row 0; they skipped the trimming and row-validity filters the direct
path applies, so a row with an unparseable stop sequence sorted to the front and
was treated as the trip's first stop; and the cached foreign-key path skipped a
check entirely when a defining file was missing. All three validators now use
their direct path.

### Added — cascade suppression

A defect in one table no longer restates itself once per row that references it.
The loader classifies each file it is joined on — parsed, missing, empty,
unparseable, missing its key column, or carrying rows with a blank key — and
checks that depend on a table stand down when it did not load, emitting an INFO
`validator_skipped` naming the check, the file and the reason.

Measured on a 177-stop feed, and checked against MobilityData v8.0.1 on each:

| fixture | before | after | canonical v8.0.1 |
|---|---|---|---|
| `stops.txt` zero bytes | 4,044 errors | **1** | 1 |
| `stops.txt` header only, no rows | 4,044 errors | **4,043** | 4,043 |
| `stop_id` column removed | 4,221 errors | **1** | 1 |
| one blank `stop_id` | 52 errors | **1** | 1 |

The distinction in the first two rows is canonical's and is the whole rule: a
file with no content at all did not load, and its dependants stand down; a file
with a valid header and no data rows loaded perfectly well and happens to be
empty, so every reference into it is a real dangling reference and is reported.
An earlier revision of this branch collapsed the two, which suppressed 4,043
canonical ERRORs and was caught only by running both validators side by side.

A single bad row does not silence a whole check — only a table that failed to
load does. The absence of a file usually stands its dependants down too, except
for `stops.txt`, which canonical treats as conditionally required and whose
absence therefore still strands every stop reference.

### Fixed

- **Two canonical ERROR rules were emitted under fork-owned names at WARNING**,
  so six defect shapes passed with zero errors that canonical rejects. Both were
  invisible to the name-based audit: renaming a canonical rule satisfies "no
  non-canonical code is ERROR" rather than tripping "severity mismatch".
  - `duplicate_key` (ERROR) was emitted as `duplicate_composite_key` (WARNING)
    for every file with a multi-column primary key — `stop_times.txt`,
    `calendar_dates.txt`, `fare_rules.txt`, `shapes.txt`, `frequencies.txt` and
    `transfers.txt`. Canonical draws no single/composite distinction; a
    duplicated key is the same defect either way. The payload now matches
    canonical's (`fieldName1`, `fieldValue1`, `oldCsvRowNumber`,
    `newCsvRowNumber`), and `duplicate_composite_key` is gone.
  - `start_and_end_range_out_of_order` (ERROR) never ran on
    `stop_times.arrival_time`/`departure_time`, which was covered instead by a
    fork-owned `stop_time_arrival_after_departure` (WARNING). A stop time whose
    arrival is after its own departure is now the canonical ERROR, and the
    fork-owned code is gone.
- `empty_file` fired on any file with no data rows. Canonical means a file with
  no content at all: a valid header with zero rows is a legitimately empty
  table. Three separate code paths raised it, and notice deduplication made them
  look like one.
- `leading_or_trailing_whitespaces` was decided by a hand-written list of field
  names, which was the wrong axis in both directions — it reported unquoted
  padding canonical ignores, and stayed silent on quoted padding in any field
  the list omitted. The rule now turns on whether the value was quoted, which is
  what canonical tests: unquoted surrounding whitespace is CSV layout any reader
  may strip, while quoted whitespace is asserted to be part of the value.
- `missing_timepoint_value` could not tell an absent `timepoint` column from a
  blank value, because the field was read out of a map. Every feed without the
  column was reported once per timed row.
- `route_color_contrast` escalated itself to ERROR below a second luma
  threshold. Canonical defines the rule at WARNING; the runtime severity also
  made it the one rule the audit could not compare statically, and it reported
  it as "not comparable" rather than as the disagreement it was.
- `Config.CountryCode` was read by nothing at all, so `-c`/`--country` and
  `WithCountryCode` were inert. It is now plumbed into field validation, where
  `invalid_phone_number` needs it: the check was a country-blind shape
  heuristic that both missed impossible numbers (`123` passed) and rejected
  valid ones (`1-800-FLOWERS`, extensions). It is now a possible-length test
  per country, which is the question canonical asks.
- A `calendar.txt` row whose `end_date` precedes its `start_date` enumerated no
  dates, so the service dropped out of every date-based rule and a plainly
  expired calendar went unreported. The inverted range is still reported on its
  own; the end date it nominates is now used for the service window.
- **Three canonical rules were counted as implemented while emitting nothing**,
  two of them ERROR, so this validator passed feeds MobilityData rejects.
  `scripts/scope_audit.py` excluded rules by matching their names and confirmed
  implementation by scanning source files without consulting the registry. It
  now reads the registry and exits non-zero when a validator exists in source
  but is never constructed.
  - `location_with_unexpected_stop_time` (ERROR) — a station referenced by
    `stop_times.stop_id` was 1 error in canonical and 0 here. It had been filed
    as GTFS-Flex; it is core GTFS, emitted by the same canonical validator as
    `stop_without_stop_time`, which was already implemented.
  - `invalid_input_files_in_subfolder` (ERROR) — a zip with its files in a
    subfolder was 7 errors in canonical and 0 here. `LoadFromZip` keyed on
    `filepath.Base`, silently flattening `gtfs/stops.txt` to `stops.txt` and
    validating the nested feed as though correctly packaged, which suppressed
    both the packaging error and the missing-file errors that follow from it.
    Only root-level entries are loaded now.
  - `leading_or_trailing_whitespaces` (WARNING) — commented out of the registry
    as "PROBLEMATIC: Hangs with large datasets (Sofia)". It does not hang. Sofia
    validates in 9.5s with it enabled, and a synthetic feed carrying 684,740
    whitespace defects completes in 9.2s. The read loop did have a real latent
    spin — it continued rather than stopped on a non-parse read error, which is
    returned again forever without consuming input — and that is fixed.
- `missing_trip_edge` required *both* `arrival_time` and `departure_time` to be
  absent before reporting. Canonical tests each field on its own and emits one
  notice per missing field per edge, so the most common form of the defect — a
  terminus with an arrival but no departure — was silently accepted. The payload
  now carries `stopSequence` and `specifiedField` as canonical does, and the
  first/last notice constructors are replaced by a single `MissingTripEdgeNotice`.
- `feed_valid_beyond_total_service_window` tested only whether the declared feed
  end ran past the service window. The condition is two-sided; a period starting
  well before the first day of service overstates coverage just as much. The
  payload is now canonical's four dates (`feedStartDate`, `feedEndDate`,
  `serviceWindowStartDate`, `serviceWindowEndDate`) instead of a single "days
  beyond", which is ambiguous once either side can trigger it.
- Shape geometry differed from canonical in four ways, all fixed: rail first and
  last stops now get four times the base 100 m tolerance (a platform sits
  alongside the track, and at a terminus the shape stops at the buffer);
  out-of-order pairs are reported once per trip rather than once per pair, the
  later ones being downstream of the same disagreement; notices are emitted once
  per shape-and-pattern rather than once per trip, and stop-distance notices once
  per shape-and-stop, so a shape served by fifty trips no longer multiplies one
  misplaced stop by fifty; and the two stop ids in the
  `stops_match_shape_out_of_order` payload were in the opposite order from
  canonical. On a railway feed this moved `stop_too_far_from_shape` from 45 to
  35 and `stops_match_shape_out_of_order` from 21 to 20, both matching canonical.
- `missing_required_field` was reported for every row of a file whose column was
  absent entirely, on top of the single `missing_required_column` that explains
  it. Removing `stop_id` from a 177-row `stops.txt` produced 177 redundant
  errors.

- `WithCaching(true)` returned different results from the same feed than
  `WithCaching(false)`, in both directions. (Superseded within this release: the
  cache is removed outright, so this fix survives only as the reason not to
  reintroduce one.) The parsed-feed cache recorded a
  file as loaded without keeping what it had parsed whenever an index was the
  first thing asked for, so whichever accessor a validator reached for first
  decided whether the cache held the feed or was permanently empty — with
  `GetStopTimesByTrip` first, every later reader saw a feed with no stop times
  at all and the checks over them silently reported nothing. Separately, the
  cached foreign-key path built the `service_id` and `shape_id` lookups out of
  `trips.txt`, the file being checked against them, so a service defined in
  `calendar.txt` but not yet used by any trip was reported as a broken
  reference at ERROR, while a trip naming a service or shape that does not
  exist could never fail.
- `fast_travel_between_consecutive_stops` over-reported on timetables written to
  the whole minute. The canonical rule reads a one-minute hop as up to two, on
  the grounds that a scheduling system emitting minute resolution has already
  rounded; without that allowance ordinary suburban services were reported as
  supersonic.
- `stops_match_shape_out_of_order` reported alignments that are in order. Where
  a shape runs out to a terminus and back, both legs stay within tolerance of
  the stops between them, and candidate matches were folded together by their
  spacing along the shape — which chains an out-and-back into a single pass and
  leaves each stop pinned at one point it must be at both before and after its
  neighbours. Passes are now cut at each local minimum, as the canonical
  validator cuts them. On one 34-route city feed this was 21 notices, all false.
- Travel-speed limits per `route_type` did not match the canonical validator's
  for six modes: light rail (was 500, now 100), subway (500 → 150), ferry
  (100 → 80), cable tram (150 → 30), funicular (150 → 50) and monorail
  (500 → 150). `fast_travel_between_far_stops` applied a flat 200 km/h rather
  than the mode's own limit, and could report the same trip once per stop; it
  now reports each trip at most once. A trip whose route cannot be resolved is
  skipped rather than held to the bus limit, the broken reference being another
  rule's to report.
- `stop_has_too_many_matches_for_shape` called a stop ambiguous at 5 passes of
  the shape, where the canonical threshold is 20. A stop in the middle of a
  dense city alignment legitimately collects a great many.
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

- Removed the whole-feed graph build in `network_topology_validator`, the most
  expensive validator in the suite, along with two redundant passes over
  `shapes.txt` and several duplicated passes over `stop_times.txt`.
- Shape-matching checks were briefly gated behind comprehensive mode. That
  gating is gone with the modes; see "Removed" below.

### Migration

#### Options removed

| before | after |
|---|---|
| `WithValidationMode(ValidationModePerformance)` | delete the option; 30 more validators now run |
| `WithValidationMode(ValidationModeDefault)` | delete the option; 3 more validators now run |
| `WithValidationMode(ValidationModeComprehensive)` | delete the option; no behaviour change |
| `WithCaching(true)` / `WithCaching(false)` | delete the option |
| CLI `--mode` / `-m` | delete the flag |

`Config.ValidationMode` and `Config.EnableCaching` are gone from the struct, so
code constructing `Config` literally will not compile until those fields are
removed.

Callers that gate on the error count should revalidate their feeds against this
build before deploying it. Two new ERROR rules and the corrected `missing_trip_edge`
condition can fail a feed that previously passed, and the shape and cascade
changes move counts in the other direction.

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