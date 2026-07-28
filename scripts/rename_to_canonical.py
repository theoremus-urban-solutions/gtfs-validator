#!/usr/bin/env python3
"""Rename notice codes to their Canonical GTFS Schedule Validator equivalents.

Each pair below was verified by comparing our implementation against the rule
description published at https://gtfs-validator.mobilitydata.org/rules.html.
Pairs whose semantics only looked similar are deliberately absent — see
CANONICAL_PARITY.md.

Several canonical codes cover more than one of ours: MobilityData reports a
single duplicate_route_name for short, long and combination clashes, a single
leading_or_trailing_whitespaces, and a single missing_trip_edge for the first
and last stop. Those merge here.
"""

import os
import re
import subprocess

RENAMES = {
    # exact one-to-one
    "block_trips_overlap": "block_trips_with_overlapping_stop_times",
    "duplicate_header": "duplicated_column",
    "feed_expires_within_7_days": "feed_expiration_date7_days",
    "feed_expires_within_30_days": "feed_expiration_date30_days",
    "missing_required_stop_name": "missing_stop_name",
    "multiple_records_in_single_record_file": "more_than_one_entity",
    "missing_route_name": "route_both_short_and_long_name_missing",
    "service_without_active_days": "service_has_no_active_day_of_the_week",
    "stop_time_decreasing_time": "stop_time_with_arrival_before_previous_departure_time",
    "trip_usability": "unusable_trip",
    "insufficient_shape_points": "single_shape_point",
    "excessive_travel_speed": "fast_travel_between_consecutive_stops",
    "wrong_number_of_fields": "invalid_row_length",
    # many-to-one merges
    "duplicate_route_short_name": "duplicate_route_name",
    "duplicate_route_long_name": "duplicate_route_name",
    "duplicate_route_name_combination": "duplicate_route_name",
    "leading_whitespace": "leading_or_trailing_whitespaces",
    "trailing_whitespace": "leading_or_trailing_whitespaces",
    "missing_trip_first_time": "missing_trip_edge",
    "missing_trip_last_time": "missing_trip_edge",
}


def sources():
    """Every source file in the repo, tracked or not — new files count too."""
    out = subprocess.run(
        ["git", "ls-files", "--cached", "--others", "--exclude-standard",
         "*.go", "*.md", "*.html"],
        capture_output=True, text=True)
    return [p for p in out.stdout.split() if os.path.exists(p)]


def main():
    changed = {}
    for path in sources():
        src = open(path, encoding="utf-8").read()
        out = src
        for old, new in RENAMES.items():
            # Only rewrite the code as a quoted string; Go type names are untouched.
            out = re.sub(r'"%s"' % re.escape(old), '"%s"' % new, out)
        if out != src:
            open(path, "w", encoding="utf-8").write(out)
            changed[path] = sum(1 for o in RENAMES if '"%s"' % o in src)

    for path, n in sorted(changed.items()):
        print("  %-58s %d" % (path, n))
    print("rewrote %d codes across %d files" % (len(RENAMES), len(changed)))


if __name__ == "__main__":
    main()
