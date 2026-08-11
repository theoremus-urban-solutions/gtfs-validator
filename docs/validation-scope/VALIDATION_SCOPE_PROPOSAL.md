# Validation scope proposal

> **Status: executed.** All 133 in-scope canonical rules are implemented, no
> non-canonical code is ERROR, and no severity disagrees with canonical. Run
> `python3 scripts/scope_audit.py` for the live reconciliation and see
> `CANONICAL_PARITY.md` for the resulting parity statement.
>
> Three things in this document turned out to be wrong, and were corrected
> rather than followed:
>
> 1. **§2 Tier 1 says `forbidden_pickup_type` / `forbidden_drop_off_type` are
>    "what our `first_stop_no_pickup` / `last_stop_no_drop_off` become".** They
>    are not. Both canonical rules are about GTFS-Flex pickup/drop-off windows
>    and say nothing about the first or last stop of a trip. The same applies to
>    `missing_stop_times_record`, `forbidden_arrival_or_departure_time`,
>    `forbidden_shape_dist_traveled` and `forbidden_continuous_pickup_drop_off`,
>    which §3 would otherwise have declined as Flex. All six are implemented to
>    the canonical meaning, which makes them inert on non-Flex feeds; the two
>    private codes are kept, at WARNING, because nothing canonical covers them.
> 2. **`unused_shape` is a canonical code**, not one of ours. §1c dropped it as
>    part of the shape-cluster collapse; it has been restored.
> 3. **§4's "duplicate emissions: 0" is not reachable as stated.** The 44 figure
>    counts generic codes such as `foreign_key_violation`,
>    `missing_required_field` and `invalid_url`, which are legitimately raised
>    from many validators with different context. Genuine duplicates — the same
>    check written twice — are gone.
>
> §5's list of 14 surviving non-canonical codes was also written against the
> pre-rework tree and undercounts. Its stated test is the right one — "does a
> canonical rule already cover this?" — and that test, not the list, is what was
> applied. Codes it expected to disappear but which nothing canonical covers
> were kept at WARNING rather than deleted, so no coverage was lost silently.

Scope criteria, in priority order:

1. **Canonical MobilityData rules, minus extensions.** These are the contract.
2. **Cheap extras that OTP acts on.** Anything OTP silently drops or fails a
   graph build over, provided it costs a single pass over data we already load.

Everything else is out of scope.

Baseline measured against <https://gtfs-validator.mobilitydata.org/rules.html>
(181 codes: 111 ERROR / 53 WARNING / 17 INFO, incl. 4 deprecated).

| | count |
|---|---|
| Codes we emit (production) | 201 |
| Codes shared with canonical | 40 |
| Canonical codes not implemented | 141 |
| Our codes with no canonical equivalent | 161 |
| **Codes emitted by more than one registered validator** | **44** |

`CANONICAL_PARITY.md` already enumerates the canonical gap. This document adds
the part it does not cover — our own internal redundancy — and turns both into
an add/remove decision.

### Decisions taken

- **Severity rule: only canonical rules may be ERROR.** Every non-canonical
  check is WARNING at most. Currently violated by 86 codes. (§5)
- **Delete `business/network_topology_validator.go`.** Not canonical, most
  expensive validator we run, and OTP acts on none of its four codes. (§1f)
- **Collapse the per-field enum notices onto canonical `unexpected_enum_value`.**
  ~15 codes into one, and it closes a live hole: non-numeric enum values
  currently emit nothing. (§1d, Tier 2)
- **Rename `validator_error` → `runtime_exception_in_validator_error`**, keeping
  ERROR. Leaves zero non-canonical ERROR codes. (§5)

---

## 1. Remove: internal redundancy

44 of our 201 codes are emitted from two or more registered validators. This is
the bulk of the "redundant checks" problem: the same feed defect is parsed,
walked and reported two or three times, at two or three times the cost.

### 1a. Validators that are wholly subsumed by another

Every code these emit is already emitted elsewhere. Delete them and unregister.

| validator | subsumed by |
|---|---|
| `entity/calendar_consistency_validator.go` (9 codes) | `business/service_consistency_validator.go` + `business/service_calendar_validator.go` |
| `business/service_consistency_validator.go` (6 codes) | `entity/calendar_consistency_validator.go` |
| `entity/primary_key_validator.go` (1 code) | `core/duplicate_key_validator.go` |
| `entity/calendar_validator.go` (1 code) | `core/missing_files_validator.go` |

The first two are mutual — they are the same validator written twice. Keep one.
Recommend keeping `business/service_consistency_validator.go` and folding
`service_calendar_validator`'s unique codes into it.

### 1b. Validator pairs with heavy overlap

Merge each pair into one; the second column is the handful of codes that would
otherwise be lost.

| pair | shared codes | unique to the one being merged in |
|---|---|---|
| `business/transfer_timing_validator.go` ↔ `business/transfer_validator.go` | 6 | `close_stops_not_possible_transfer`, `long_distance_transfer`, `unrealistic_transfer_time`, `negative_min_transfer_time` |
| `business/frequency_validator.go` ↔ `business/overlapping_frequency_validator.go` | 2 | `cross_trip_frequency_overlap`, `frequency_duration_shorter_than_headway` |
| `entity/route_name_validator.go` ↔ `relationship/route_consistency_validator.go` | 3 | `route_without_trips` |
| `business/schedule_consistency_validator.go` ↔ `relationship/stop_time_sequence_time_validator.go` + `relationship/stop_time_consistency_validator.go` | 4 | `stop_without_service` |
| `core/coordinate_validator.go` ↔ `business/geospatial_validator.go` | 2 | — |
| `business/block_overlapping_validator.go` ↔ `entity/trip_block_id_validator.go` | 1 | `block_service_mismatch` |
| `entity/shape_validator.go` ↔ `relationship/shape_distance_validator.go` ↔ `relationship/shape_increasing_distance_validator.go` | see 1c | — |

### 1c. The shape-distance cluster

Three validators emit **twelve** codes for what canonical expresses in four:

Ours: `decreasing_shape_distance`, `decreasing_or_equal_shape_distance`,
`shape_distance_decreasing`, `shape_distance_not_increasing`,
`equal_shape_distance`, `inconsistent_shape_distance`,
`incomplete_shape_distance`, `large_shape_distance_jump`,
`negative_shape_distance`, `shape_distance_not_starting_from_zero`,
`unrealistic_shape_distance`, `shape_distance_inconsistent_with_geography`.

Canonical: `decreasing_shape_distance` (we have it),
`equal_shape_distance_same_coordinates`,
`equal_shape_distance_diff_coordinates`,
`equal_shape_distance_diff_coordinates_distance_below_threshold`.

Collapse to one validator emitting the four canonical codes. That removes nine
codes and two full passes over `shapes.txt`.

### 1d. Per-field enum notices

`core/invalid_row_validator.go` emits 18 codes, 10 of which restate a check a
typed validator already performs (`invalid_route_type`, `invalid_location_type`,
`invalid_transfer_type`, `invalid_payment_method`, `invalid_exception_type`,
`invalid_exact_times`, `invalid_headway`, `invalid_transfers`,
`invalid_bikes_allowed`, `invalid_wheelchair_accessible`,
`invalid_wheelchair_boarding`).

Canonical reports all of these as one code, `unexpected_enum_value`, with the
field name in the context. Adopting that collapses ~15 of our codes into one and
removes the duplication in a single change. Recommended.

### 1e. Service/expiry cluster

Fifteen codes across four validators for feed and service date coverage:
`expired_feed`, `feed_expired`, `expired_service`, `service_expired`,
`service_expires_within_7_days`, `service_expires_within_30_days`,
`insufficient_service_next_7_days`, `insufficient_service_next_30_days`,
`no_service_next_7_days`, `no_trips_next_7_days`, `future_feed_start_date`,
`future_service`, `service_never_active`, `no_service_date_found`,
`no_service_defined`.

Canonical covers the same ground with `feed_expiration_date7_days`,
`feed_expiration_date30_days`, `expired_calendar`, `future_calendar`,
`trip_coverage_not_active_for_next7_days`. Map onto those five and drop the rest.

### 1f. Out of scope on the criteria

`business/network_topology_validator.go` — `fragmented_network`,
`low_network_connectivity`, `small_network_component`, `isolated_stop`.

Not canonical, and OTP does not act on any of them: it builds a graph from a
disconnected feed without complaint. It is also the most expensive validator we
run (whole-feed graph construction). Recommend deleting outright.

Same reasoning, cheaper to keep but still nothing consumes them: `long_zone_id`,
`zone_id_same_as_stop_id`, `unusual_route_type_combination`,
`insufficient_coordinate_precision`, `stop_name_contains_html`,
`stop_name_contains_url`, `unreasonable_level_index`, `unreasonable_max_slope`.

### 1g. Severity corrections

Eight shared codes disagree with canonical severity. Canonical is the contract,
so ours should move:

| code | ours | canonical |
|---|---|---|
| `attribution_without_role` | ERROR | WARNING |
| `empty_file` | WARNING | **ERROR** |
| `feed_expiration_date7_days` | ERROR | WARNING |
| `invalid_language_code` | WARNING | **ERROR** |
| `more_than_one_entity` | ERROR | WARNING |
| `service_has_no_active_day_of_the_week` | ERROR | WARNING |
| `single_shape_point` | ERROR | WARNING |
| `unusable_trip` | ERROR | WARNING |

---

## 2. Add: canonical rules worth implementing

141 canonical codes are unimplemented. Excluding extensions and the declined
runtime notices leaves **93 rules** to add, one of which
(`runtime_exception_in_validator_error`) is a rename of something we already
emit. Ranked; tier sizes sum to 93.

### Tier 1 — cheap, high signal, OTP-relevant (24)

Each is a predicate on rows we already parse. No extra pass, no geometry.

`unsorted_stop_times`, `missing_stop_times_record`,
`stop_time_with_only_arrival_or_departure_time`,
`stop_time_timepoint_without_times`, `missing_timepoint_value`,
`forbidden_arrival_or_departure_time`, `stop_without_stop_time`, `unused_trip`,
`missing_required_agency_id`, `inconsistent_agency_timezone`,
`wrong_parent_location_type`, `location_without_parent_station`,
`stop_without_location`, `forbidden_shape_dist_traveled`,
`transfer_with_invalid_stop_location_type`,
`transfer_with_invalid_trip_and_route`, `transfer_with_invalid_trip_and_stop`,
`trip_coverage_not_active_for_next7_days`, `forbidden_pickup_type`,
`forbidden_drop_off_type`, `forbidden_continuous_pickup_drop_off`.

Plus the three that §1c collapses the shape cluster onto, which must exist
before that deletion can land: `equal_shape_distance_same_coordinates`,
`equal_shape_distance_diff_coordinates`,
`equal_shape_distance_diff_coordinates_distance_below_threshold`.

`unsorted_stop_times` is the single largest gap — canonical reports it on feeds
we pass clean, and OTP's stop-time ordering assumptions depend on it.

`forbidden_pickup_type` / `forbidden_drop_off_type` are what our
`first_stop_no_pickup` / `last_stop_no_drop_off` become.

### Tier 2 — cheap field/structure checks (16)

Canonical's generic type layer, which we currently approximate with ad-hoc
per-field codes. Implementing these is what makes 1d possible.

`unexpected_enum_value`, `invalid_date`, `invalid_time`, `invalid_integer`,
`invalid_float`, `number_out_of_range`, `invalid_currency`,
`invalid_currency_amount`, `invalid_phone_number`, `invalid_character`,
`new_line_in_value`, `non_ascii_or_non_printable_char`, `empty_column_name`,
`empty_row`, `start_and_end_range_equal`, `start_and_end_range_out_of_order`.

### Tier 3 — cheap quality warnings (17)

Single-pass string and date comparisons.

`mixed_case_recommended_field`, `route_long_name_contains_short_name`,
`same_name_and_description_for_route`, `same_name_and_description_for_stop`,
`same_route_and_agency_url`, `same_stop_and_agency_url`,
`same_stop_and_route_url`, `expired_calendar`, `future_calendar`, `future_feed`,
`missing_feed_info_date`, `missing_feed_contact_email_and_url`,
`missing_recommended_file`, `inconsistent_agency_lang`,
`feed_info_lang_and_agency_lang_mismatch`, `missing_bike_allowance`,
`unused_station`.

Note `same_name_and_description_for_route` / `_for_stop` replace our single
`same_name_and_description`, which is emitted by two validators (see 1b).

### Tier 4 — pathways and levels (8)

**`CANONICAL_PARITY.md` classifies these as an extension. They are not** —
`pathways.txt` and `levels.txt` are core GTFS, and we already ship
`accessibility/pathway_validator.go` and `accessibility/level_validator.go` to
hang them on. Cheap relative to what those validators already do.

`pathway_loop`, `pathway_dangling_generic_node`, `pathway_unreachable_location`,
`pathway_to_wrong_location_type`, `pathway_to_platform_with_boarding_areas`,
`pathway_to_stop_with_access_outside_of_station_pathways`,
`bidirectional_exit_gate`, plus `missing_level_id`.

### Tier 5 — geometric (11)

Correct and canonical, but each needs shape interpolation or a spatial index
over the whole feed. This tier was the argument for the performance/
comprehensive mode split.

**Superseded.** The modes were removed and all of these now always run. The
cost the split was defending turned out not to exist: running the full set is
about 1.15x the old default on the largest feed to hand, not the 2x to 20x the
docs asserted. See `BENCHMARKS.md`.

`stop_too_far_from_shape`, `stop_too_far_from_shape_using_user_distance`,
`stops_match_shape_out_of_order`, `stop_has_too_many_matches_for_shape`,
`trip_distance_exceeds_shape_distance`,
`trip_distance_exceeds_shape_distance_below_threshold`,
`fast_travel_between_far_stops`, `transfer_distance_too_large`,
`transfer_distance_above_2_km`, `point_near_origin`, `point_near_pole`.

`transfer_distance_too_large` (10 km) should replace our `long_distance_transfer`
(500 m) — ours fires on ordinary same-street transfers.

### Tier 6 — low value, implement last (17)

INFO-level advisories, the translations trio, and the one runtime notice we do
adopt — `runtime_exception_in_validator_error`, which is a rename of our existing
`validator_error` and can land any time (see §5).

`big_gap_in_service`, `feed_valid_beyond_total_service_window`,
`service_extends_far_in_the_future`, `service_window_outside_feed_period`,
`platform_without_parent_station`, `stop_without_zone_id`,
`trip_headsign_matches_intermediate_stop`,
`trip_with_shape_dist_traveled_but_no_shape_distances`,
`inconsistent_route_type_for_block_id`,
`inconsistent_route_type_for_in_seat_transfer`,
`transfer_with_suspicious_mid_trip_in_seat`, plus
`translation_foreign_key_violation`, `translation_unexpected_value`,
`translation_unknown_table_name`.

Also `stop_access_specified_for_incorrect_location` and
`stop_access_specified_for_stop_with_no_parent_station`. Both fire only on feeds
that populate `stop_access`, which ours do not — inert either way, so they cost
nothing to add and nothing to skip.

---

## 3. Do not implement

| group | count | reason |
|---|---|---|
| GTFS-Flex / GeoJSON | 27 | extension, unused (incl. `missing_pickup_or_drop_off_window`) |
| GTFS-Fares v2 | 11 | extension, unused |
| Runtime/infrastructure | 5 | `i_o_error`, `thread_execution_error`, `runtime_exception_in_loader_error`, `u_r_i_syntax_error`, `too_many_rows` — artefacts of the Java implementation's execution model, not feed defects. `invalid_input_files_in_subfolder` was listed here and should not have been: a zip whose files sit below the root is a defect in the feed's packaging, and it is now implemented. |

**One exception, carved out of that group:
`runtime_exception_in_validator_error`.** We already implement exactly this
failure mode as `validator_error`, so the "Java execution artefact" reasoning
does not apply — see §5.

---

## 4. Net effect

In-scope canonical surface is **133 rules** (181 published, less 4 deprecated,
27 Flex, 11 Fares v2 and 6 declined runtime notices). We implement 40.

| | now | after Tiers 1–4 | full scope |
|---|---|---|---|
| Codes emitted | 201 | ~119 | 147 |
| — canonical | 40 | 105 | 133 |
| — non-canonical | 161 | 14 | 14 |
| Duplicate emissions | 44 | 0 | 0 |
| ERROR codes | 118 | ~58 | 68 |
| — non-canonical ERROR | 86 | **0** | **0** |
| Registered validators | 58 | ~48 | ~48 |

Codes drop by roughly 40% while canonical coverage nearly triples, because most
of what comes out is the same check counted twice or under a private name.
Tiers 5 and 6 (28 rules) are deferred, not dropped.

## 5. The remaining non-canonical checks

**Severity rule: only canonical rules may be ERROR. Every non-canonical check is
WARNING at most.** A code MobilityData does not define is our opinion, and an
opinion should not fail someone's feed.

Today 86 of our 161 non-canonical codes are ERROR. That is the rule's whole
justification: we currently fail feeds on 86 checks nobody else recognises.

### What survives

Of the 161 non-canonical codes:

| | count |
|---|---|
| Superseded once Tier 1–3 canonical rules land | 77 |
| Removed by the cluster/duplicate/scope cuts in §1 | 66 |
| Renamed onto a canonical code (`validator_error`) | 1 |
| **Genuinely novel and worth keeping** | **14** |

So the answer to "are the rest worthwhile" is mostly no — not because they check
nothing, but because 77 of them re-check something canonical already covers under
a different name. `invalid_date_format` is `invalid_date`; `missing_agency_id` is
`missing_required_agency_id`; `invalid_latitude` is `number_out_of_range`;
`non_increasing_stop_sequence` is `unsorted_stop_times`. They are worthwhile
checks and unworthwhile *codes*.

### The 14 keepers, all at WARNING

Each covers a failure mode OTP acts on and canonical does not model.

| code | now | proposed | why it earns its place |
|---|---|---|---|
| `all_stops_no_pickup` | ERROR | WARNING | no stop on the trip allows boarding — trip is unroutable |
| `all_stops_no_drop_off` | ERROR | WARNING | same, for alighting |
| `block_service_mismatch` | ERROR | WARNING | `block_id` spans services with different calendars; breaks OTP interlining |
| `circular_station_reference` | ERROR | WARNING | `parent_station` cycle; OTP fails the graph build |
| `frequency_duration_shorter_than_headway` | ERROR | WARNING | frequency window shorter than the headway generates zero trips |
| `cross_trip_frequency_overlap` | WARNING | WARNING | overlapping frequency windows across trips |
| `stop_without_service` | ERROR | WARNING | pickup and drop-off both forbidden at every visit |
| `pathway_to_same_stop` | ERROR | WARNING | degenerate pathway |
| `consecutive_duplicate_stops` | WARNING | WARNING | zero-distance hop; OTP routing artefact |
| `duplicate_stop_in_trip` | WARNING | WARNING | same |
| `deprecated_route_type` | WARNING | WARNING | OTP has historically failed graph builds on these |
| `route_without_trips` | WARNING | WARNING | OTP drops the route silently |
| `duplicate_pathway` | WARNING | WARNING | not covered by canonical's pathway set |
| `inconsistent_bidirectional_pathway` | WARNING | WARNING | same |

### `validator_error` → `runtime_exception_in_validator_error`

`validator_error` is not a feed check at all. Both validator loops
(`implementation.go:456` serial, `:512` parallel) wrap each `Validate` call in a
`defer`/`recover`; on a panic the recover emits this notice with the validator's
Go type name and the panic value, then the run continues. It means *one of our
validators crashed and its checks did not run* — a hole in the report, not a
defect in the feed.

That is precisely canonical `runtime_exception_in_validator_error` (ERROR).
Renaming onto it:

- removes the only exception to the severity rule — after the rename there are
  **zero** non-canonical ERROR codes, with no carve-out to remember;
- costs one string change plus a `notice_descriptions.go` entry;
- corrects §3, which would otherwise decline a canonical rule we already
  implement.

Keep it at ERROR, which is both canonical and correct: at WARNING a crashed
validator would sit unnoticed among ordinary feed warnings.

**Worth fixing at the same time:** the notice names the panicking Go type and
nothing else, so a consumer cannot tell *which rules* went unchecked. If
`shape_validator` panics on row 1, every shape check silently vanishes from the
report. Adding the validator's emitted-code list to the notice context would make
the gap legible. Not required for the rename — noting it as follow-up.

### Four to drop outright

| code | why |
|---|---|
| `unexpected_bidirectional_gate` | duplicates canonical `bidirectional_exit_gate` (Tier 4) |
| `missing_bikes_allowed_for_ferry` | duplicates canonical `missing_bike_allowance` (Tier 3) |
| `stop_sequence_gap` | gaps in `stop_sequence` are legal GTFS — values must increase, not be contiguous. Pure false positive. |
| `stop_name_missing_but_inherited` | contradicts canonical `missing_stop_name`, which we already emit as ERROR on the same rows |

### Net severity effect

| | now | proposed |
|---|---|---|
| ERROR codes | 118 | 69, all canonical |
| Non-canonical ERROR codes | 86 | **0** |

## 6. Suggested order

1. **1a + 1b + 1c** — delete the subsumed validators and collapse the shape
   cluster. Pure deletion, no new behaviour, immediate speedup.
2. **1g + §5 severity rule** — align the 8 shared codes to canonical severity,
   rename `validator_error` → `runtime_exception_in_validator_error`, and demote
   every remaining non-canonical ERROR to WARNING. All one-line changes, and
   together they make the rest of the work safe to land incrementally: once no
   non-canonical code can be ERROR, nothing added later can newly fail a feed
   unless canonical says so.
3. **Tier 1** — the OTP-relevant gap, and the highest-value work here.
4. **Tier 2 + 1d** — add the generic type layer, then retire the per-field enum
   codes onto `unexpected_enum_value`.
5. **Tier 3 + Tier 4**.
6. **1e** — remap the service/expiry cluster once `expired_calendar` and
   `future_calendar` exist to remap onto.
7. **Tier 5** — done, and unconditional: there is no longer a mode flag to gate it behind. **Tier 6** last.

Steps 1, 2 and 4 are breaking changes to emitted codes and should land together
in one release with a mapping table in `CHANGELOG.md`.
