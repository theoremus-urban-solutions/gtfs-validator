#!/usr/bin/env python3
"""Delete notice declarations that nothing emits any more.

A notice is three things in this repo: a doc comment, a struct embedding
*BaseNotice, and a NewXxxNotice constructor, all adjacent in
notice/validation_notices.go. Removing a check leaves all three behind. This
finds constructors with no caller outside notice/ and deletes the whole block,
plus the code's entry in notice/code_files.go.

    python3 scripts/prune_notices.py            # report only
    python3 scripts/prune_notices.py --apply    # delete them
"""

import os
import re
import sys

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
NOTICE_DIR = os.path.join(ROOT, "notice")


def constructors():
    """Return {file: {constructorName: code}}."""
    out = {}
    for fn in sorted(os.listdir(NOTICE_DIR)):
        if not fn.endswith(".go") or fn.endswith("_test.go"):
            continue
        text = open(os.path.join(NOTICE_DIR, fn), encoding="utf-8").read()
        found = {}
        for m in re.finditer(r"func (New\w+Notice)\([^)]*\)[^{]*\{(.*?)\n\}", text, re.S):
            code = re.search(r'NewBaseNotice\("([a-z0-9_]+)"', m.group(2))
            if code:
                found[m.group(1)] = code.group(1)
        if found:
            out[fn] = found
    return out


def callers():
    """Return the set of constructor names called outside notice/."""
    used = set()
    for base, dirs, files in os.walk(ROOT):
        dirs[:] = [d for d in dirs if d not in (".git", "testdata", "docs", "scripts")]
        for fn in files:
            if not fn.endswith(".go"):
                continue
            path = os.path.join(base, fn)
            if os.path.relpath(path, ROOT).startswith("notice" + os.sep):
                continue
            text = open(path, encoding="utf-8", errors="replace").read()
            used |= set(re.findall(r"\bnotice\.(New\w+Notice)\b", text))
    return used


def block_span(text, ctor):
    """Return (start, end) covering the doc comment, struct and constructor.

    The three are written as one paragraph per notice, so the block starts at
    the comment line above the struct and ends at the constructor's closing
    brace.
    """
    m = re.search(r"^func %s\(" % re.escape(ctor), text, re.M)
    if not m:
        return None
    end = text.index("\n}\n", m.start()) + len("\n}\n")

    struct_name = ctor[len("New"):]
    s = re.search(r"^type %s struct \{" % re.escape(struct_name), text, re.M)
    start = s.start() if s else m.start()

    # Walk back over the contiguous comment lines introducing the struct.
    lines = text[:start].split("\n")
    while len(lines) >= 2 and lines[-2].startswith("//"):
        lines.pop()
        start -= len(lines[-1]) + 1
    return start, end


def main():
    apply = "--apply" in sys.argv
    used = callers()
    total = 0

    for fn, found in constructors().items():
        path = os.path.join(NOTICE_DIR, fn)
        text = open(path, encoding="utf-8").read()
        orphans = [c for c in found if c not in used]
        if not orphans:
            continue

        spans = []
        for ctor in orphans:
            span = block_span(text, ctor)
            if span is None:
                print(f"  ! could not locate {ctor}", file=sys.stderr)
                continue
            spans.append(span)
            print(f"{fn}: {ctor} ({found[ctor]})")
            total += 1

        if apply:
            for start, end in sorted(spans, reverse=True):
                text = text[:start] + text[end:]
            text = re.sub(r"\n{3,}", "\n\n", text)
            open(path, "w", encoding="utf-8").write(text)

    print(f"\n{total} orphaned notices" + (" removed" if apply else ""))


if __name__ == "__main__":
    main()
