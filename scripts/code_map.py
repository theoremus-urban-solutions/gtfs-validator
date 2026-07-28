#!/usr/bin/env python3
"""Map notice codes to the validator files that emit them.

Notice codes are declared once in notice/validation_notices.go via
NewBaseNotice, wrapped in a NewXxxNotice constructor. Validators call the
constructor, never NewBaseNotice, so a static map from constructor to code is
needed to answer "which validator emits this code".

    python3 scripts/code_map.py               # code -> files
    python3 scripts/code_map.py --by-file     # file -> codes
    python3 scripts/code_map.py --dupes       # codes emitted by 2+ files
    python3 scripts/code_map.py FILE...       # codes emitted by these files
"""

import os
import re
import sys
from collections import defaultdict

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
NOTICES = os.path.join(ROOT, "notice")


def constructor_codes():
    """Return {constructorName: code} for every notice constructor."""
    out = {}
    for fn in os.listdir(NOTICES):
        if not fn.endswith(".go") or fn.endswith("_test.go"):
            continue
        text = open(os.path.join(NOTICES, fn), encoding="utf-8").read()
        # func NewFooNotice(...) *FooNotice { ... NewBaseNotice("code", ...
        for m in re.finditer(
            r"func (New\w+Notice)\([^)]*\)[^{]*\{(.*?)\n\}", text, re.S
        ):
            code = re.search(r'NewBaseNotice\("([a-z0-9_]+)"', m.group(2))
            if code:
                out[m.group(1)] = code.group(1)
    return out


def emitters(ctors):
    """Return {code: {relative file path}} across the whole tree."""
    out = defaultdict(set)
    for base, dirs, files in os.walk(ROOT):
        dirs[:] = [d for d in dirs if d not in (".git", "testdata", "docs", "scripts")]
        for fn in files:
            if not fn.endswith(".go") or fn.endswith("_test.go"):
                continue
            path = os.path.join(base, fn)
            rel = os.path.relpath(path, ROOT)
            if rel.startswith("notice/"):
                continue
            text = open(path, encoding="utf-8", errors="replace").read()
            for name in set(re.findall(r"\bnotice\.(New\w+Notice)\b", text)):
                if name in ctors:
                    out[ctors[name]].add(rel)
    return out


def main():
    args = [a for a in sys.argv[1:]]
    ctors = constructor_codes()
    emit = emitters(ctors)

    if "--dupes" in args:
        for code, files in sorted(emit.items()):
            if len(files) > 1:
                print(f"{code}\n    " + "\n    ".join(sorted(files)))
        return

    if "--by-file" in args:
        by_file = defaultdict(set)
        for code, files in emit.items():
            for f in files:
                by_file[f].add(code)
        for f, codes in sorted(by_file.items()):
            print(f"{f}  ({len(codes)})")
            for c in sorted(codes):
                print(f"    {c}")
        return

    targets = [a for a in args if not a.startswith("--")]
    if targets:
        for t in targets:
            codes = sorted(c for c, files in emit.items() if t in files)
            print(f"=== {t} ({len(codes)}) ===")
            for c in codes:
                others = sorted(emit[c] - {t})
                mark = "   also: " + ", ".join(others) if others else "   UNIQUE"
                print(f"  {c}{mark}")
        return

    for code, files in sorted(emit.items()):
        print(f"{code}: {', '.join(sorted(files))}")


if __name__ == "__main__":
    main()
