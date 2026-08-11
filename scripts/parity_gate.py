#!/usr/bin/env python3
"""Fail unless every canonical rule matches the canonical validator exactly.

`scope_audit.py` compares codes and severities by NAME. That is not sufficient
and has now been fooled twice: a canonical ERROR emitted under a fork-owned name
at WARNING satisfies "no non-canonical code is ERROR" instead of tripping
"severity mismatch", and a validator present in source but absent from the
registry counted as implemented. Both were caught only by running the two
validators side by side.

This script is that comparison, made repeatable. It builds a corpus of feeds —
one per defect shape, each isolating a single rule — runs both validators over
each, and fails on any disagreement about a canonical code. It also fails when
the two disagree on whether a feed is valid at all, which is naming-independent
and so catches a re-badged rule however it is spelled.

    python3 scripts/parity_gate.py --jar path/to/gtfs-validator-cli.jar

The jar is the official MobilityData release, not this repo's scrape of its
rules page:

    https://github.com/MobilityData/gtfs-validator/releases

Requires java on PATH and a built ./gtfs-validator binary (make build).
"""

import argparse
import json
import os
import shutil
import subprocess
import sys
import tempfile

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
sys.path.insert(0, os.path.join(ROOT, "scripts"))
import scope_audit  # noqa: E402

# A minimal feed every case starts from. Cases override individual files.
BASE = {
    "agency.txt": (
        "agency_id,agency_name,agency_url,agency_timezone\n"
        "A1,Test Transit,https://example.com,America/New_York\n"
    ),
    "stops.txt": (
        "stop_id,stop_name,stop_lat,stop_lon\n"
        "S1,First Stop,40.7589,-73.9851\n"
        "S2,Second Stop,40.7614,-73.9776\n"
    ),
    "routes.txt": (
        "route_id,agency_id,route_short_name,route_long_name,route_type\n"
        "R1,A1,1,Main Street Line,3\n"
    ),
    "trips.txt": "route_id,service_id,trip_id,trip_headsign\nR1,SV1,T1,Downtown\n",
    "stop_times.txt": (
        "trip_id,arrival_time,departure_time,stop_id,stop_sequence\n"
        "T1,08:00:00,08:00:00,S1,1\n"
        "T1,08:15:00,08:15:00,S2,2\n"
    ),
    "calendar.txt": (
        "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,"
        "start_date,end_date\n"
        "SV1,1,1,1,1,1,0,0,20240101,20241231\n"
    ),
}


def case(**overrides):
    files = dict(BASE)
    for name, content in overrides.items():
        key = name.replace("__", ".")
        if content is None:
            files.pop(key, None)
        else:
            files[key] = content
    return files


# Each case isolates one rule, chosen because the two validators once disagreed
# on it. Add a case whenever a new divergence is found; that is what stops it
# coming back.
CASES = {
    "clean": case(),

    # Table states. Canonical distinguishes a file with no content from a file
    # with a header and no rows: the first suppresses dependent checks, the
    # second is a valid empty table and suppresses nothing.
    "stops_zerobyte": case(stops__txt=""),
    "stops_headeronly": case(stops__txt="stop_id,stop_name,stop_lat,stop_lon\n"),
    "stops_nocolumn": case(stops__txt="stop_name,stop_lat,stop_lon\nFirst,40.7,-73.9\n"),
    "stops_blankkey": case(
        stops__txt="stop_id,stop_name,stop_lat,stop_lon\n,First,40.7,-73.9\nS2,Second,40.76,-73.97\n"
    ),
    "stops_absent": case(stops__txt=None),
    "trips_absent": case(trips__txt=None),
    "trips_headeronly": case(trips__txt="route_id,service_id,trip_id,trip_headsign\n"),
    "routes_headeronly": case(
        routes__txt="route_id,agency_id,route_short_name,route_long_name,route_type\n"
    ),

    # Duplicate primary keys, single and composite. Canonical reports every one
    # of these as duplicate_key at ERROR.
    "dup_stop_id": case(
        stops__txt="stop_id,stop_name,stop_lat,stop_lon\nS1,A,40.7,-73.9\nS1,B,40.8,-73.8\n"
    ),
    "dup_stoptimes": case(
        stop_times__txt=(
            "trip_id,arrival_time,departure_time,stop_id,stop_sequence\n"
            "T1,08:00:00,08:00:00,S1,1\nT1,08:15:00,08:15:00,S2,1\n"
        )
    ),
    "dup_caldate": case(
        calendar_dates__txt="service_id,date,exception_type\nSV1,20240115,2\nSV1,20240115,2\n"
    ),
    "dup_shape": case(
        shapes__txt=(
            "shape_id,shape_pt_lat,shape_pt_lon,shape_pt_sequence\n"
            "SH1,40.75,-73.98,1\nSH1,40.76,-73.97,1\n"
        )
    ),
    "dup_transfer": case(
        transfers__txt="from_stop_id,to_stop_id,transfer_type\nS1,S2,0\nS1,S2,0\n"
    ),
    "dup_frequency": case(
        frequencies__txt=(
            "trip_id,start_time,end_time,headway_secs\n"
            "T1,08:00:00,09:00:00,600\nT1,08:00:00,10:00:00,900\n"
        )
    ),

    # Start/end pairs. arrival after departure is a canonical ERROR.
    "arr_after_dep": case(
        stop_times__txt=(
            "trip_id,arrival_time,departure_time,stop_id,stop_sequence\n"
            "T1,08:30:00,08:15:00,S1,1\nT1,09:00:00,09:00:00,S2,2\n"
        )
    ),
    "cal_range": case(
        calendar__txt=(
            "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,"
            "start_date,end_date\nSV1,1,1,1,1,1,0,0,20241231,20240101\n"
        )
    ),

    # timepoint: an absent column is not a missing value.
    "tp_absent": case(),
    "tp_blank": case(
        stop_times__txt=(
            "trip_id,arrival_time,departure_time,stop_id,stop_sequence,timepoint\n"
            "T1,08:00:00,08:00:00,S1,1,\nT1,08:15:00,08:15:00,S2,2,\n"
        )
    ),

    # Whitespace is reported for quoted values only; unquoted padding is layout.
    "ws_unquoted": case(
        stops__txt="stop_id,stop_name,stop_lat,stop_lon\nS1, First,40.7,-73.9\nS2,Second ,40.76,-73.97\n"
    ),
    "ws_quoted": case(
        stops__txt='stop_id,stop_name,stop_lat,stop_lon\nS1," First",40.7,-73.9\nS2,"Second ",40.76,-73.97\n'
    ),

    # Phone numbers are checked for possible length in the feed's country.
    "phone_ok": case(
        agency__txt=(
            "agency_id,agency_name,agency_url,agency_timezone,agency_phone\n"
            "A1,Test,https://example.com,America/New_York,202-456-1111\n"
        )
    ),
    "phone_short": case(
        agency__txt=(
            "agency_id,agency_name,agency_url,agency_timezone,agency_phone\n"
            "A1,Test,https://example.com,America/New_York,123\n"
        )
    ),
    "phone_vanity": case(
        agency__txt=(
            "agency_id,agency_name,agency_url,agency_timezone,agency_phone\n"
            "A1,Test,https://example.com,America/New_York,1-800-FLOWERS\n"
        )
    ),

    # route_color_contrast is WARNING however faint the contrast.
    "color_contrast": case(
        routes__txt=(
            "route_id,agency_id,route_short_name,route_long_name,route_type,"
            "route_color,route_text_color\nR1,A1,1,Main,3,FFFFFF,FEFEFE\n"
        )
    ),

    # A location that is not a stop, referenced by stop_times.
    "station_ref": case(
        stops__txt=(
            "stop_id,stop_name,stop_lat,stop_lon,location_type,parent_station\n"
            "S1,First,40.7589,-73.9851,0,ST1\nS2,Second,40.7614,-73.9776,0,ST1\n"
            "ST1,Main Station,40.76,-73.98,1,\n"
        ),
        stop_times__txt=(
            "trip_id,arrival_time,departure_time,stop_id,stop_sequence\n"
            "T1,08:00:00,08:00:00,S1,1\nT1,08:15:00,08:15:00,ST1,2\n"
        ),
    ),

    # A trip edge missing one of its two times.
    "trip_edge": case(
        stop_times__txt=(
            "trip_id,arrival_time,departure_time,stop_id,stop_sequence\n"
            "T1,,08:00:00,S1,1\nT1,08:15:00,,S2,2\n"
        )
    ),
}


def write_case(directory, files):
    os.makedirs(directory, exist_ok=True)
    for name, content in files.items():
        with open(os.path.join(directory, name), "w", encoding="utf-8") as f:
            f.write(content)


def country_flag(country):
    """The -c flag, or nothing when no country is being asserted.

    An empty country is not the same as a country named "". Both validators
    take the absence of the flag to mean "do not hold phone numbers to any
    one numbering plan", so the flag has to be absent rather than empty.
    """
    return ["-c", country] if country else []


def run_fork(binary, path, country, date):
    out = subprocess.run(
        [binary, "-i", path, "-f", "json"] + country_flag(country),
        capture_output=True, text=True,
    )
    # The CLI exits 1 when the feed has errors, which is not a failure to run.
    try:
        report = json.loads(out.stdout)
    except json.JSONDecodeError:
        return None
    codes, errors = {}, report["summary"]["counts"]["errors"]
    for n in report["notices"]:
        counts = n["severityCounts"]
        severity = ("ERROR" if counts["errors"] else
                    "WARNING" if counts["warnings"] else "INFO")
        codes[n["code"]] = (severity, n["totalNotices"])
    return codes, errors


def run_canonical(jar, path, country, date, outdir):
    os.makedirs(outdir, exist_ok=True)
    subprocess.run(
        ["java", "-jar", jar, "-i", path, "-o", outdir, "-d", date]
        + country_flag(country),
        capture_output=True, text=True,
    )
    report_path = os.path.join(outdir, "report.json")
    if not os.path.exists(report_path):
        return None
    report = json.load(open(report_path, encoding="utf-8"))
    codes = {n["code"]: (n["severity"], n["totalNotices"]) for n in report["notices"]}
    errors = sum(n["totalNotices"] for n in report["notices"] if n["severity"] == "ERROR")
    return codes, errors


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--jar", required=True, help="MobilityData gtfs-validator CLI jar")
    ap.add_argument("--binary", default=os.path.join(ROOT, "gtfs-validator"))
    ap.add_argument("--country", default="",
                    help="passed to both validators, so a country-dependent "
                         "rule cannot disagree merely because the two were "
                         "asked about different places. Empty by default, "
                         "which is both validators' own default, so the gate "
                         "tests what an unflagged run actually does")
    ap.add_argument("--date", default="2026-08-11",
                    help="fixed validation date, so time-based rules cannot "
                         "disagree merely because the runs straddled midnight")
    ap.add_argument("--feeds", nargs="*", default=[],
                    help="extra real feeds (zip or directory) to compare")
    ap.add_argument("--keep", action="store_true", help="keep the work directory")
    args = ap.parse_args()

    for path, what in ((args.jar, "jar"), (args.binary, "binary")):
        if not os.path.exists(path):
            sys.exit("%s not found: %s" % (what, path))

    workdir = tempfile.mkdtemp(prefix="gtfs-parity-")
    published = scope_audit.scrape()
    in_scope = {c for c in published if scope_audit.bucket(c) == "in-scope"}

    inputs = []
    for name, files in sorted(CASES.items()):
        directory = os.path.join(workdir, "cases", name)
        write_case(directory, files)
        inputs.append((name, directory))
    for feed in args.feeds:
        inputs.append(("feed:" + os.path.basename(feed), feed))

    failures, checked = [], 0
    for name, path in inputs:
        fork = run_fork(args.binary, path, args.country, args.date)
        canon = run_canonical(args.jar, path, args.country, args.date,
                              os.path.join(workdir, "out", name.replace(":", "_")))
        if fork is None or canon is None:
            failures.append((name, "RUN", "one side produced no report", ""))
            continue
        fork_codes, fork_errors = fork
        canon_codes, canon_errors = canon
        checked += 1

        # Naming-independent: catches a canonical ERROR re-badged under a
        # fork-owned code, which the per-code loop below cannot see.
        if (fork_errors > 0) != (canon_errors > 0):
            failures.append((name, "VERDICT",
                             "fork errors=%d" % fork_errors,
                             "canonical errors=%d" % canon_errors))

        for code in sorted(in_scope & (set(fork_codes) | set(canon_codes))):
            ours, theirs = fork_codes.get(code), canon_codes.get(code)
            if ours != theirs:
                failures.append((name, code, "fork=%s" % (ours,),
                                 "canonical=%s" % (theirs,)))

    if args.keep:
        print("work directory: %s" % workdir)
    else:
        shutil.rmtree(workdir, ignore_errors=True)

    print("inputs checked: %d" % checked)
    if not failures:
        print("PASS - every canonical rule matches canonical exactly")
        return 0

    print("FAIL - %d divergence(s)\n" % len(failures))
    width = max(len(f[0]) for f in failures)
    for name, code, ours, theirs in failures:
        print("  %-*s  %-46s %-30s %s" % (width, name, code, ours, theirs))
    return 1


if __name__ == "__main__":
    sys.exit(main())
