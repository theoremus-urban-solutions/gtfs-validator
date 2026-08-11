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
#
# Both of these sets are exclusions by name, which is a blunt instrument — a rule
# that merely reads like an extension gets dropped from the scope count without
# anyone re-deriving why. Two were wrong and hid ERROR rules for a full release:
# `location_with_unexpected_stop_time` sat in FLEX although it is core GTFS,
# emitted by the same canonical validator as `stop_without_stop_time`, and
# `invalid_input_files_in_subfolder` sat in DECLINED_RUNTIME although a zip whose
# files are not at the root is a packaging defect in the feed, not an artefact of
# the Java runner. Before adding to either set, check the rule against
# MobilityData's own exported notice schema rather than the rendered rules page.
FLEX = re.compile(
    r"geo_json|prior_notice|prior_day|booking|pickup_drop_off_window"
    r"|pickup_or_drop_off_window|geometry|feature_type|malformed_json"
    r"|invalid_geometry|missing_required_element"
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
}

IMPLEMENTATION_GO = os.path.join(ROOT, "implementation.go")


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


def strip_comments(text):
    """Blank out // line comments and /* */ blocks, preserving offsets loosely."""
    text = re.sub(r"/\*.*?\*/", "", text, flags=re.S)
    return re.sub(r"//[^\n]*", "", text)


def registered_constructors():
    """Return the set of validator constructors actually wired into the registry.

    Source presence is not reachability: a validator whose file compiles but
    which `initializeValidators` never constructs emits nothing at runtime. That
    is exactly how `leading_or_trailing_whitespaces` stayed commented out while
    every doc counted it as implemented, so this reads the registry body with
    comments stripped and takes only the constructors that survive.
    """
    text = open(IMPLEMENTATION_GO, encoding="utf-8", errors="replace").read()
    start = text.find("func (v *internalValidator) initializeValidators()")
    if start == -1:
        sys.exit("could not find initializeValidators in implementation.go")

    # Walk braces from the signature to the matching close, so the body is taken
    # exactly rather than by a fragile line count.
    brace = text.find("{", start)
    depth, end = 0, None
    for i in range(brace, len(text)):
        if text[i] == "{":
            depth += 1
        elif text[i] == "}":
            depth -= 1
            if depth == 0:
                end = i
                break
    if end is None:
        sys.exit("initializeValidators body is unbalanced")

    body = strip_comments(text[brace:end])
    return set(re.findall(r"\b(?:\w+\.)?(New\w+)\s*\(", body))


def go_sources():
    """Yield (path, comment-stripped text) for every non-test Go file."""
    for base, dirs, files in os.walk(ROOT):
        dirs[:] = [d for d in dirs if d not in (".git", "testdata", "docs", "vendor")]
        for fn in files:
            if not fn.endswith(".go") or fn.endswith("_test.go"):
                continue
            path = os.path.join(base, fn)
            text = open(path, encoding="utf-8", errors="replace").read()
            yield path, strip_comments(text)


def severity_of(match):
    sev = match.group(2).rsplit(".", 1)[-1]
    return sev if sev in ("ERROR", "WARNING", "INFO") else "COMPUTED"


def notice_constructors(notice_dir):
    """Return ({ctor: code}, {code: severity}) for the notice package.

    Most codes are produced by a `NewSomethingNotice` wrapper around a literal
    `NewBaseNotice("code", SEVERITY, ...)`, so the wrapper is the unit validators
    call and therefore the unit whose reachability matters. Only the notice
    package is scanned here; `NewBaseNotice` called directly from a validator is
    handled at its call site instead, where the enclosing file settles the
    question on its own.
    """
    ctor_code, severities = {}, {}
    for path, text in go_sources():
        if not path.startswith(notice_dir):
            continue
        # Split on top-level func declarations so a NewBaseNotice call is
        # attributed to the wrapper it sits inside.
        parts = re.split(r"^func\s+(New\w+)\s*\(", text, flags=re.M)
        for ctor, body in zip(parts[1::2], parts[2::2]):
            m = re.search(r'NewBaseNotice\("([a-z0-9_]+)",\s*([A-Za-z_.]+)', body)
            if not m:
                continue
            ctor_code[ctor] = m.group(1)
            severities.setdefault(m.group(1), severity_of(m))
    return ctor_code, severities


def scan_repo():
    """Return ({code: severity}, unreachable_codes) for the production build.

    Severity is the literal passed to NewBaseNotice, normalised across the
    `ERROR` and `notice.ERROR` spellings. Where it is computed at runtime the
    value is a variable name, reported as "COMPUTED" — those cannot be compared
    against canonical statically and are excluded from the mismatch count.

    Reachability, not source presence, decides whether a code counts as
    implemented. A notice constructor is reachable when something outside the
    notice package calls it from a live site: any file outside `validator/`, or
    a validator file whose own constructor the registry actually builds. A code
    whose every call site sits in an unregistered validator — or which nothing
    calls at all — emits nothing at runtime and is reported as missing.
    """
    validator_dir = os.path.join(ROOT, "validator") + os.sep
    notice_dir = os.path.join(ROOT, "notice") + os.sep

    registered = registered_constructors()
    ctor_code, severities = notice_constructors(notice_dir)

    reachable_codes = set()
    for path, text in go_sources():
        if path.startswith(notice_dir):
            continue  # definitions, not call sites
        if path.startswith(validator_dir):
            own = set(re.findall(r"^func (New\w+)\s*\(", text, re.M))
            # A validator file with constructors none of which the registry
            # builds is dead code; a file with no constructors is a shared
            # helper reached through whichever validator calls it.
            if own and not (own & registered):
                continue
        for ctor in re.findall(r"\b(?:\w+\.)?(New\w+Notice)\s*\(", text):
            if ctor in ctor_code:
                reachable_codes.add(ctor_code[ctor])
        # Codes built straight from NewBaseNotice at a live site, with no
        # wrapper in the notice package to trace.
        for m in re.finditer(
            r'NewBaseNotice\("([a-z0-9_]+)",\s*([A-Za-z_.]+)', text
        ):
            severities.setdefault(m.group(1), severity_of(m))
            reachable_codes.add(m.group(1))

    unreachable = set(severities) - reachable_codes
    ours = {c: severities[c] for c in reachable_codes}
    return ours, unreachable


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
    ours, unreachable = scan_repo()

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

    if unreachable:
        print("\nin source but never constructed — these emit nothing at runtime:")
        for code in sorted(unreachable):
            print(f"  {canon.get(code, '?'):<8} {code}")

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

    # Exit non-zero on the two conditions that make the parity claim a lie: a
    # validator present in source but missing from the registry, and a severity
    # that disagrees with canonical. Rules still to add are printed but do not
    # fail the run, so the script stays usable while a gap is being closed.
    if unreachable or sev_mismatch:
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
