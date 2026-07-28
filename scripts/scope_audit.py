#!/usr/bin/env python3
"""Reconcile our emitted notice codes against the canonical MobilityData rules.

Every count in docs/validation-scope/VALIDATION_SCOPE_PROPOSAL.md comes from
this script. Re-run it after any change to the notice set to see the numbers
move.

    python3 scripts/scope_audit.py              # summary
    python3 scripts/scope_audit.py --list       # + every code, bucketed
    python3 scripts/scope_audit.py --json       # rewrite the two JSON files
    python3 scripts/scope_audit.py --refresh    # re-scrape the rules page first

The scraped page is cached at docs/validation-scope/mobilitydata-rules.html so
the upstream site is not hit on every run.
"""

import argparse
import json
import os
import re
import sys
from collections import Counter

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
DOCS = os.path.join(ROOT, "docs", "validation-scope")
RULES_HTML = os.path.join(DOCS, "mobilitydata-rules.html")
CANONICAL_JSON = os.path.join(DOCS, "canonical_rules.json")
CURRENT_JSON = os.path.join(DOCS, "current_rules.json")
RULES_URL = "https://gtfs-validator.mobilitydata.org/rules.html"

# Deprecated upstream; listed in the page's "Table of deprecated notices".
DEPRECATED = {
    "fare_transfer_rule_missing_transfer_count",
    "missing_prior_day_booking_field_value",
    "missing_recommended_column",
    "unused_parent_station",
}

# Extensions we do not consume, and runtime notices that are artefacts of the
# Java implementation's execution model rather than feed defects. See §3 of the
# proposal. `runtime_exception_in_validator_error` is deliberately NOT declined:
# we already implement it as `validator_error`.
FLEX = re.compile(
    r"geo_json|prior_notice|prior_day|booking|pickup_drop_off_window"
    r"|pickup_or_drop_off_window|geometry|feature_type|malformed_json"
    r"|invalid_geometry|missing_required_element|location_with_unexpected_stop_time"
    r"|geography_id"
)
FARES = re.compile(
    r"fare_media|fare_product|fare_transfer_rule|timeframe|route_networks_specified"
)
DECLINED_RUNTIME = {
    "i_o_error",
    "u_r_i_syntax_error",
    "runtime_exception_in_loader_error",
    "thread_execution_error",
    "too_many_rows",
    "invalid_input_files_in_subfolder",
}


def scrape(refresh=False):
    """Return {code: severity} for every published canonical rule."""
    if refresh or not os.path.exists(RULES_HTML):
        import urllib.request

        print(f"fetching {RULES_URL} ...", file=sys.stderr)
        with urllib.request.urlopen(RULES_URL) as r:
            html = r.read().decode("utf-8")
        os.makedirs(DOCS, exist_ok=True)
        with open(RULES_HTML, "w", encoding="utf-8") as f:
            f.write(html)
    else:
        html = open(RULES_HTML, encoding="utf-8").read()

    # Severities live in an embedded JSON blob, not in the rendered tables.
    blob = html.replace('\\"', '"')
    pairs = re.findall(
        r'"code":"([a-z0-9_]+)".{0,4000}?"severityLevel":"(ERROR|WARNING|INFO)"',
        blob,
    )
    rules = dict(pairs)
    if len(rules) < 150:
        sys.exit(f"only parsed {len(rules)} rules — the page format likely changed")
    return rules


def scan_repo():
    """Return {code: severity} for every notice the production build can emit.

    Severity is the literal passed to NewBaseNotice, normalised across the
    `ERROR` and `notice.ERROR` spellings. Where it is computed at runtime the
    value is a variable name, reported as "COMPUTED" — those cannot be compared
    against canonical statically and are excluded from the mismatch count.
    """
    ours = {}
    for base, dirs, files in os.walk(ROOT):
        dirs[:] = [d for d in dirs if d not in (".git", "testdata", "docs")]
        for fn in files:
            if not fn.endswith(".go") or fn.endswith("_test.go"):
                continue
            text = open(os.path.join(base, fn), encoding="utf-8", errors="replace").read()
            for m in re.finditer(r'NewBaseNotice\("([a-z0-9_]+)",\s*([A-Za-z_.]+)', text):
                sev = m.group(2).rsplit(".", 1)[-1]
                if sev not in ("ERROR", "WARNING", "INFO"):
                    sev = "COMPUTED"
                ours.setdefault(m.group(1), sev)
    return ours


def bucket(code):
    if code in DEPRECATED:
        return "deprecated"
    if FLEX.search(code):
        return "flex"
    if FARES.search(code):
        return "fares-v2"
    if code in DECLINED_RUNTIME:
        return "declined-runtime"
    return "in-scope"


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--list", action="store_true", help="print every code")
    ap.add_argument("--json", action="store_true", help="rewrite the JSON files")
    ap.add_argument("--refresh", action="store_true", help="re-scrape the rules page")
    args = ap.parse_args()

    canon = scrape(args.refresh)
    ours = scan_repo()

    buckets = {}
    for code in canon:
        buckets.setdefault(bucket(code), set()).add(code)
    inscope = buckets.get("in-scope", set())

    shared = set(ours) & set(canon)
    implemented = inscope & set(ours)
    to_add = inscope - set(ours)
    non_canonical = set(ours) - set(canon)
    nc_errors = {c for c in non_canonical if ours[c] == "ERROR"}
    sev_mismatch = {
        c: (ours[c], canon[c])
        for c in shared
        if ours[c] != canon[c] and ours[c] != "COMPUTED"
    }
    computed = {c for c in ours if ours[c] == "COMPUTED"}

    print(f"canonical published      {len(canon):4d}  {dict(Counter(canon.values()))}")
    for name in ("deprecated", "flex", "fares-v2", "declined-runtime"):
        print(f"  less {name:<20} {len(buckets.get(name, set())):4d}")
    print(f"in-scope canonical       {len(inscope):4d}  "
          f"({sum(1 for c in inscope if canon[c] == 'ERROR')} ERROR)")
    print()
    print(f"we emit                  {len(ours):4d}")
    print(f"  canonical, in scope    {len(implemented):4d}")
    print(f"  canonical, out of scope{len(shared - inscope):4d}")
    print(f"  non-canonical          {len(non_canonical):4d}")
    print(f"    of which ERROR       {len(nc_errors):4d}   <-- target: 0")
    print()
    print(f"canonical rules to add   {len(to_add):4d}")
    print(f"severity mismatches      {len(sev_mismatch):4d}")
    if computed:
        print(f"severity computed at run {len(computed):4d}  "
              f"({', '.join(sorted(computed))}) — not statically comparable")

    if sev_mismatch:
        print("\nseverity mismatches (ours -> canonical):")
        for code, (o, c) in sorted(sev_mismatch.items()):
            print(f"  {code:<52} {o:>8} -> {c}")

    if args.list:
        print("\n=== canonical rules to add (in scope, not implemented) ===")
        for code in sorted(to_add):
            print(f"  {canon[code]:<8} {code}")
        print("\n=== non-canonical codes we emit ===")
        for code in sorted(non_canonical):
            mark = "  <-- ERROR" if code in nc_errors else ""
            print(f"  {ours[code]:<8} {code}{mark}")

    if args.json:
        os.makedirs(DOCS, exist_ok=True)
        with open(CANONICAL_JSON, "w") as f:
            json.dump(canon, f, indent=1, sort_keys=True)
            f.write("\n")
        with open(CURRENT_JSON, "w") as f:
            json.dump(ours, f, indent=1, sort_keys=True)
            f.write("\n")
        print(f"\nwrote {CANONICAL_JSON}\nwrote {CURRENT_JSON}")


if __name__ == "__main__":
    main()
