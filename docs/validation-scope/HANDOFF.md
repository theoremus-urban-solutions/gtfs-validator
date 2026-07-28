# Handoff: validation scope rework

You are picking up an audit that is **complete and agreed**. Nothing in
`validator/` or `notice/` has been touched yet. Your job is to execute it.

Read `VALIDATION_SCOPE_PROPOSAL.md` in this directory first — it is the spec.
This file tells you how to work through it.

## The two criteria, in priority order

1. **Canonical MobilityData rules, minus extensions.** These are the contract.
2. **Cheap extras OTP acts on.** Only if it costs one pass over data already
   loaded.

Anything that satisfies neither gets deleted. When in doubt, delete — the audit
found the repo had drifted to 201 codes of which 161 were nobody's standard.

## Decisions already taken — do not relitigate

- **Only canonical rules may be ERROR.** Every non-canonical check is WARNING at
  most. Currently violated by 86 codes; the target is 0.
- Delete `validator/business/network_topology_validator.go` outright.
- Collapse the per-field enum notices onto canonical `unexpected_enum_value`.
- Rename `validator_error` → `runtime_exception_in_validator_error` (stays ERROR).
- Pathways and levels are **core GTFS, not an extension**. `CANONICAL_PARITY.md`
  says otherwise and is wrong on this point; fix it as you go.

## Files here

| file | what it is |
|---|---|
| `VALIDATION_SCOPE_PROPOSAL.md` | the spec: what to remove (§1), what to add (§2), what to decline (§3), the severity rule (§5), execution order (§6) |
| `canonical_rules.json` | 181 canonical codes → severity, scraped |
| `current_rules.json` | 201 codes we emit → severity, scanned from source |
| `mobilitydata-rules.html` | cached scrape of the rules page — **use this, do not re-fetch** |
| `../../scripts/scope_audit.py` | regenerates all of the above and prints the reconciliation |

## Your feedback loop

```
python3 scripts/scope_audit.py            # summary + severity mismatches
python3 scripts/scope_audit.py --list     # every code, bucketed
python3 scripts/scope_audit.py --json     # rewrite the two JSON files
python3 scripts/scope_audit.py --refresh  # re-scrape upstream (rarely needed)
```

Run it after every step. It is the scoreboard. Current baseline:

```
in-scope canonical        133  (68 ERROR)
we emit                   201
  canonical, in scope      40
  non-canonical           161
    of which ERROR         86   <-- target: 0
canonical rules to add     93
severity mismatches         8
```

Target end state: 133 canonical + 14 non-canonical = 147 codes, 68 ERROR, all
canonical, 0 duplicate emissions.

`route_color_contrast` computes its severity at runtime and is reported as
COMPUTED rather than compared — that is correct behaviour, not a bug to fix.

## Order of work

Follow §6 of the proposal. The sequencing constraint that is easy to get wrong:

> **Tier 1's three `equal_shape_distance_*` rules must land BEFORE §1c deletes
> the shape cluster.** §1c collapses twelve codes onto four canonical ones;
> three of those four do not exist yet. Delete first and you lose coverage.

Otherwise: deletions (§1a–1c, 1f) → severity (§1g, §5) → Tier 1 → Tier 2 + the
enum collapse (§1d) → Tier 3 → Tier 4 → §1e → Tiers 5–6.

Steps 1, 2 and 4 are breaking changes to emitted codes. Land them in one release
with a mapping table in `CHANGELOG.md`.

## Ground rules

- **Reimplement, never vendor.** The canonical validator is Apache-2.0 Java;
  this repo is MIT. Work from the published rule text in
  `mobilitydata-rules.html`, which states facts about GTFS rather than
  copyrightable expression. `scripts/gen_notice_descriptions.py` already pulls
  the wording.
- **Every code change needs its test updated.** Each validator has a
  `_test.go` beside it. Renaming a code breaks its test — that is the test
  doing its job.
- Run `go build ./... && go test ./...` before declaring any step done. Report
  failures rather than working around them.
- `scripts/rename_to_canonical.py` exists from the previous pass and handles
  code renames across source and tests. Reuse it for §1g and the enum collapse.
- Keep `CANONICAL_PARITY.md` and this proposal in sync as you land changes, or
  the next agent inherits a stale map.

## Known traps

- **`notice.NoticeContainer` deduplicates.** Merging two codes into one can
  silently reduce the notice count in a test fixture. Expected — check the
  fixture's intent before "fixing" it.
- **Severity is sometimes a variable**, not a literal (`route_color_contrast`).
  Grep for `NewBaseNotice(` rather than assuming the severity is inline.
- **Enum checks are guarded by `strconv.Atoi(...); err == nil`**
  (`validator/core/invalid_row_validator.go:204-300`). A non-numeric enum value
  falls through and emits nothing. The Tier 2 rewrite must not preserve this —
  it is the hole the collapse is meant to close.
- **`validator_error` is a panic handler**, not a feed check
  (`implementation.go:456` serial, `:512` parallel). Both loops `recover()` and
  emit it. Do not treat it as a validation rule.
- Length checks use `len()` on Go strings, so they count bytes, not runes. A
  Cyrillic `route_short_name` trips the 12-char limit at 6 characters. Live bug
  in `route_name_validator.go` and `route_consistency_validator.go`; fix it if
  you touch those files.

## Open question for the user

The two `stop_access_*` rules are parked in Tier 6 on the assumption that no
feed we validate populates `stop_access`. If that turns out to be wrong they
move up. Worth confirming before Tier 6.
