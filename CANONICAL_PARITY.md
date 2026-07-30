# Parity with the Canonical GTFS Schedule Validator

Reference: <https://gtfs-validator.mobilitydata.org/rules.html>

Run `python3 scripts/scope_audit.py` to reproduce every number in this
document. It scrapes the published rule set, scans the codes this repo emits,
and prints the reconciliation.

## Where we stand

| | count |
|---|---|
| Canonical rules published | 181 |
| — less deprecated upstream | 4 |
| — less GTFS-Flex / GeoJSON | 27 |
| — less GTFS-Fares v2 | 11 |
| — less declined runtime notices | 6 |
| **In scope** | **133** |
| **Implemented** | **133** |
| Severity disagreements with canonical | 0 |

Scope is set by two criteria, in priority order: the canonical rules minus the
extensions we do not consume, and cheap extras that OpenTripPlanner acts on.
`docs/validation-scope/VALIDATION_SCOPE_PROPOSAL.md` is the full argument.

The canonical validator is Java under Apache-2.0; this project is MIT. Its code
cannot be vendored without making this a mixed-license repository, so
everything adopted from it is reimplemented from the published rule
descriptions, which are statements about GTFS rather than copyrightable
expression. `scripts/gen_notice_descriptions.py` pulls the wording; the checks
are our own.

## Deliberately not implemented

| group | count | reason |
|---|---|---|
| GTFS-Flex / GeoJSON | 27 | extension, unused by the feeds we validate |
| GTFS-Fares v2 | 11 | extension, unused |
| Deprecated upstream | 4 | withdrawn from the canonical rule set |
| Runtime / infrastructure | 6 | `i_o_error`, `thread_execution_error`, `runtime_exception_in_loader_error`, `u_r_i_syntax_error`, `too_many_rows`, `invalid_input_files_in_subfolder` — artefacts of the Java implementation's execution model, not feed defects |

One rule is carved out of that last group: `runtime_exception_in_validator_error`.
We already implement exactly that failure mode — both validator loops wrap each
`Validate` call in a `defer`/`recover` and emit it on a panic — so the "Java
execution artefact" reasoning does not apply. It means one of our validators
crashed and its checks did not run: a hole in the report rather than a defect in
the feed. It is the one ERROR that is not about the feed, and that is
deliberate; at WARNING a crashed validator would sit unnoticed among ordinary
feed warnings.

Note that `pathways.txt` and `levels.txt` are **core GTFS, not an extension**.
An earlier version of this document classified them as one and used that to
justify skipping eight rules. All eight are now implemented.

## The severity rule

**Only canonical rules may be ERROR. Every check of our own is WARNING at most.**

A code MobilityData does not define is our opinion, and an opinion should not
fail someone's feed. Before this rule, 86 of our own codes were ERROR, so a feed
could fail on 86 checks nobody else recognises. That count is now zero, and
`scripts/apply_severity_rule.py` enforces it.

## Our own rules

What remains beyond the canonical set covers failure modes that matter to
consumers the canonical rule set does not model. OpenTripPlanner drops routes
and trips silently in several of these cases, and has historically failed graph
builds on deprecated route types.

They are additional coverage, not divergence: every one is WARNING or INFO, so
none of them can fail a feed that canonical would pass.

`scripts/gen_validator_rules.py` regenerates `VALIDATOR_RULES.md`, which lists
every code by validator with its severity and whether it is canonical.

## Renamed onto canonical codes

Verified by comparing our implementation against the published rule text. The Go
type and constructor names were left alone; only the emitted code changed.

| was | now |
|---|---|
| `block_trips_overlap` | `block_trips_with_overlapping_stop_times` |
| `duplicate_header` | `duplicated_column` |
| `excessive_travel_speed` | `fast_travel_between_consecutive_stops` |
| `feed_expires_within_7_days` | `feed_expiration_date7_days` |
| `feed_expires_within_30_days` | `feed_expiration_date30_days` |
| `insufficient_shape_points` | `single_shape_point` |
| `missing_agency_id`, `missing_route_agency_id` | `missing_required_agency_id` |
| `missing_required_stop_name` | `missing_stop_name` |
| `missing_route_name` | `route_both_short_and_long_name_missing` |
| `multiple_records_in_single_record_file` | `more_than_one_entity` |
| `service_without_active_days` | `service_has_no_active_day_of_the_week` |
| `stop_time_decreasing_time` | `stop_time_with_arrival_before_previous_departure_time` |
| `trip_usability` | `unusable_trip` |
| `validator_error` | `runtime_exception_in_validator_error` |
| `wrong_number_of_fields` | `invalid_row_length` |

Several canonical codes each absorb a group of ours, matching how upstream
reports them:

| was | now |
|---|---|
| `duplicate_route_short_name`, `duplicate_route_long_name`, `duplicate_route_name_combination` | `duplicate_route_name` |
| `leading_whitespace`, `trailing_whitespace` | `leading_or_trailing_whitespaces` |
| `missing_trip_first_time`, `missing_trip_last_time` | `missing_trip_edge` |
| `missing_arrival_time`, `missing_departure_time` | `stop_time_with_only_arrival_or_departure_time` |
| the ~15 per-field `invalid_*` enum codes | `unexpected_enum_value` |
| `invalid_date_format` | `invalid_date` |
| `invalid_time_format` | `invalid_time` |
| `invalid_currency_code` | `invalid_currency` |
| `invalid_fare_price` | `invalid_currency_amount` |
| `invalid_coordinate`, `invalid_latitude`, `invalid_longitude` | `invalid_float`, `number_out_of_range` |
| `suspicious_coordinate` | `point_near_origin`, `point_near_pole` |
| `invalid_service_date_range`, `calendar_end_before_start`, `feed_info_end_date_before_start_date` | `start_and_end_range_out_of_order` |
| `invalid_frequency_time_range` | `start_and_end_range_equal`, `start_and_end_range_out_of_order` |
| `same_name_and_description` | `same_name_and_description_for_route`, `same_name_and_description_for_stop` |
| `pathway_to_same_stop` | `pathway_loop` |
| `unexpected_bidirectional_gate` | `bidirectional_exit_gate` |
| `missing_bikes_allowed_for_ferry` | `missing_bike_allowance` |
| `long_distance_transfer` | `transfer_distance_too_large` |
| the 15-code service and feed expiry cluster | `expired_calendar`, `future_calendar`, `future_feed`, `feed_expiration_date7_days`, `feed_expiration_date30_days`, `trip_coverage_not_active_for_next7_days` |

The whitespace and enum merges are behaviour changes: a value with whitespace on
both sides used to produce two notices and now produces one, and a row with
several bad enum values now produces one notice per field under a single code
rather than one code per field. Both match upstream.

`leading_or_trailing_whitespaces` reports only whitespace inside double quotes.
This is not a shortcut but the rule: a CSV parser strips the whitespace around
an unquoted value, so ` Metro ` written bare reaches every other check as
`Metro` and there is nothing left to report. Upstream's parser does the same and
its rule description says so. Go's `encoding/csv` keeps both kinds, so the
validator asks the reader where each field began and looks at the byte there —
a double quote exactly when the field was quoted. Without that distinction a
feed with padded unquoted values collects notices the canonical validator does
not emit.

The enum merge also closed a hole. Every per-field enum check was guarded by a
successful `strconv.Atoi`, so a field holding a non-numeric value fell through
and reported nothing at all. `unexpected_enum_value` compares the value as
written.

## Where the published rules and the canonical implementation disagree

Running a real feed through both validators and comparing turned up two rules
where the rules page and the reference implementation do not say the same
thing. We follow the implementation, because parity is what consumers actually
compare against: a feed that passes upstream and fails here is a support ticket,
whichever reading is better on paper. Both are worth knowing about.

**`route_long_name_contains_short_name`** is published as a containment rule
with three bad examples, two of which — `"14"`/`"Route 14"` and
`"2"`/`"Route 2: Bellows Falls In-Town"` — put the short name after a generic
word. The implementation tests only whether the long name *begins* with the
short name followed by a separator, so it does not report either of its own
examples. The narrowing is deliberate; the source carries a comment and a design
discussion for it. We match the implementation, and the cost is one-directional:
the common `"14"`/`"Route 14"` shape goes unreported by both validators.
`TestRouteNameValidator_LongNameLeadsWithShortName` pins all seven documented
examples, with the two we knowingly miss recorded as such.

**`route_color_contrast`** is not a WCAG contrast ratio, despite the name. The
implementation compares Rec. 601 luma — `0.30R + 0.59G + 0.11B` — and fires
below a difference of 72, with a source comment noting that W3C's recommended
125 is meant for body text and is "too harsh for big colored logos like line
names". We had used a WCAG 4.5 ratio, which failed white-on-dark-green: legible
by eye, and passed by canonical.

## Deliberately not renamed

These look equivalent by name but check something different. Renaming them would
claim canonical compliance for a check that does not match.

| ours | canonical | why not |
|---|---|---|
| `invalid_field_format` | `unexpected_enum_value` | ours is any malformed field, not enums |
| `stop_without_service` | `stop_without_stop_time` | ours is pickup and drop-off both forbidden at every visit; theirs is a stop no trip references |
| `first_stop_no_pickup`, `last_stop_no_drop_off` | `forbidden_pickup_type`, `forbidden_drop_off_type` | the canonical pair is about GTFS-Flex pickup/drop-off windows, not about the first and last stop of a trip. `VALIDATION_SCOPE_PROPOSAL.md` §2 asserts the equivalence and is wrong; both canonical rules are implemented, and ours are kept because nothing canonical covers what they check |
| `child_station_too_far_from_parent` | `stop_too_far_from_shape` | different computations |
