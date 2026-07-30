#!/usr/bin/env python3
"""Apply the severity rule: only canonical rules may be ERROR.

A code MobilityData does not define is our opinion, and an opinion should not
fail someone's feed. Every non-canonical notice is therefore demoted to
WARNING, and the handful of shared codes whose severity disagrees with
canonical are moved onto canonical's value.

    python3 scripts/apply_severity_rule.py            # report
    python3 scripts/apply_severity_rule.py --apply    # rewrite
"""

import os
import re
import sys

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
import scope_audit  # noqa: E402

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
SOURCES = ["notice/validation_notices.go", "validator/file_structure_validator.go"]


def main():
    apply = "--apply" in sys.argv
    canon = scope_audit.scrape()

    changes = []
    for rel in SOURCES:
        path = os.path.join(ROOT, rel)
        text = open(path, encoding="utf-8").read()

        def rewrite(m):
            prefix, code, sev = m.group(1), m.group(2), m.group(3)
            bare = sev.rsplit(".", 1)[-1]
            if bare not in ("ERROR", "WARNING", "INFO"):
                return m.group(0)  # computed at runtime

            target = canon.get(code, "WARNING" if bare == "ERROR" else bare)
            if target == bare:
                return m.group(0)

            changes.append((code, bare, target))
            qualified = sev[: -len(bare)] + target
            return f'{prefix}"{code}", {qualified}'

        new = re.sub(
            r'(NewBaseNotice\()"([a-z0-9_]+)", ([A-Za-z_.]+)', rewrite, text
        )
        if apply and new != text:
            open(path, "w", encoding="utf-8").write(new)

    for code, was, now in sorted(changes):
        kind = "canonical" if code in canon else "non-canonical"
        print(f"  {code:<52} {was:>8} -> {now:<8} ({kind})")
    print(f"\n{len(changes)} severities changed" + ("" if apply else " (dry run)"))


if __name__ == "__main__":
    main()
