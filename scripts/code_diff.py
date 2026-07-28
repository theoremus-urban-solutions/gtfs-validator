#!/usr/bin/env python3
"""Diff the codes we emit now against a recorded baseline.

docs/validation-scope/baseline_rules.json is the 201-code set measured before
the scope rework began. It is a fixed snapshot and is never regenerated —
`scope_audit.py --json` rewrites current_rules.json, which tracks the live set.
Comparing the live scan against the baseline produces the add/remove/
severity-change table a breaking release needs in CHANGELOG.md.

    python3 scripts/code_diff.py            # summary
    python3 scripts/code_diff.py --markdown # CHANGELOG tables
"""

import json
import os
import sys

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
import scope_audit  # noqa: E402

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
BASELINE = os.path.join(ROOT, "docs", "validation-scope", "baseline_rules.json")


def main():
    markdown = "--markdown" in sys.argv

    baseline = json.load(open(BASELINE))
    current = scope_audit.scan_repo()
    canon = scope_audit.scrape()

    removed = sorted(set(baseline) - set(current))
    added = sorted(set(current) - set(baseline))
    changed = sorted(
        c for c in set(baseline) & set(current)
        if baseline[c] != current[c] and "COMPUTED" not in (baseline[c], current[c])
    )

    if not markdown:
        print(f"baseline {len(baseline)}  ->  current {len(current)}")
        print(f"  removed {len(removed)}")
        print(f"  added   {len(added)}  ({sum(1 for c in added if c in canon)} canonical)")
        print(f"  severity changed {len(changed)}")
        return

    print("#### Codes removed\n")
    for code in removed:
        print(f"- `{code}`")

    print("\n#### Codes added\n")
    for code in added:
        mark = "" if code in canon else "  (not canonical)"
        print(f"- `{code}` ({current[code]}){mark}")

    print("\n#### Severity changed\n")
    print("| code | was | now |")
    print("|---|---|---|")
    for code in changed:
        print(f"| `{code}` | {baseline[code]} | {current[code]} |")


if __name__ == "__main__":
    main()
