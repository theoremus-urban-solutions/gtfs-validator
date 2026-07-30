# GTFS Validator Go

[![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Validation Rules](https://img.shields.io/badge/Validation%20Rules-176-brightgreen.svg)](https://github.com/theoremus-urban-solutions/gtfs-validator)
[![Canonical Parity](https://img.shields.io/badge/Canonical%20Parity-133%2F133-brightgreen.svg)](CANONICAL_PARITY.md)
[![Validators](https://img.shields.io/badge/Validators-55-brightgreen.svg)](VALIDATOR_RULES.md)
[![Performance](https://img.shields.io/badge/Performance-685k%20stop%20times%20in%20~20s-orange.svg)](https://github.com/theoremus-urban-solutions/gtfs-validator)

A fast, comprehensive GTFS (General Transit Feed Specification) validator library for Go. 176 validation rules: every applicable rule from the Canonical GTFS Schedule Validator, plus business-logic checks that matter to downstream consumers such as OpenTripPlanner.

> **📊 Scope**: 176 rules. **All 133 applicable rules from the
> [Canonical GTFS Schedule Validator](https://gtfs-validator.mobilitydata.org/rules.html)
> are implemented**, at the severity it gives them. The 48 canonical rules not
> implemented are GTFS-Flex, GTFS-Fares v2, deprecated upstream, or artefacts of
> that validator's own execution model rather than feed defects.
>
> The other 43 rules are ours, covering failure modes the canonical set does not
> model — several of which break OpenTripPlanner graph builds. None of them is
> ERROR: a code MobilityData does not define is our opinion, and an opinion
> should not fail your feed. See [CANONICAL_PARITY.md](CANONICAL_PARITY.md) and
> [VALIDATOR_RULES.md](VALIDATOR_RULES.md).

## Features

- **🚀 Fast Validation**: Optimized for large feeds with parallel processing and memory pools
- **📋 Comprehensive**: 176 validation rules across 55 validators, in full parity with the canonical validator
- **🔧 Multiple Modes**: Performance, default, and comprehensive validation modes
- **⚡ Concurrent**: Thread-safe with configurable worker pools
- **⏰ Context Support**: Cancellation, timeouts, and progress reporting
- **💾 Memory Aware**: Memory pooling, a streaming CSV parser, and a hard cap you can set — see the measured figures below
- **📊 Rich Reports**: JSON, console, and summary output formats with comprehensive error descriptions
- **🎯 Streaming Processing**: Process massive CSV files without loading into memory
- **📦 Dual Purpose**: Use as Go library or standalone CLI tool
- **🛠️ Modern CLI**: Cobra-powered interface with subcommands, help, and autocompletion
- **⏱️ Late-night Support**: Proper GTFS time validation for 24+ hour formats
- **📈 Advanced Analytics**: Structured logging, benchmarking, and performance monitoring

## Installation

```bash
# As a library
go get github.com/theoremus-urban-solutions/gtfs-validator

# CLI tool
go install github.com/theoremus-urban-solutions/gtfs-validator/cmd/gtfs-validator@latest
```

## Quick Start

### Library

```go
import gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator"

// Basic validation
validator := gtfsvalidator.New()
report, err := validator.ValidateFile("feed.zip")

if report.HasErrors() {
    fmt.Printf("❌ %d errors found\n", report.ErrorCount())
} else {
    fmt.Println("✅ Feed is valid!")
}
```

### CLI

```bash
# Validate a feed (short flags)
gtfs-validator -i feed.zip

# Validate with subcommand
gtfs-validator validate feed.zip

# Fast mode with JSON output  
gtfs-validator -i feed.zip -m performance -f json

# With progress and output file
gtfs-validator validate feed.zip --mode performance --progress -o report.json
```

## Validation Modes

A mode chooses how much of the feed is examined, not how carefully. Every rule
a mode runs is the same rule at the same severity.

| Mode | Validators | What it runs | Use case |
|------|-----------|--------------|----------|
| **Performance** | 25 of 55 | Structure, field types, references and feed metadata | CI gates, where a broken reference should fail the build and an opinion should not |
| **Default** | 52 of 55 | The above plus entity, business, accessibility and fare rules | Day-to-day validation |
| **Comprehensive** | 55 of 55 | Everything, adding the three whole-feed passes: shape geometry, geospatial checks and per-date trip coverage | Deep analysis before publishing a feed |

The three validators held back from the default are the ones whose cost grows
with the whole feed rather than with one file: shape geometry, geospatial
proximity and per-date trip coverage. Comprehensive mode is how you ask for
them, and their codes — `stop_too_far_from_shape`, `big_gap_in_service` and the
rest of those sets — cannot appear in a default run.

## Advanced Usage

### Custom Configuration

```go
validator := gtfsvalidator.New(
    gtfsvalidator.WithValidationMode(gtfsvalidator.ValidationModePerformance),
    gtfsvalidator.WithCountryCode("UK"),
    gtfsvalidator.WithMaxNoticesPerType(50),
    gtfsvalidator.WithParallelWorkers(8),
    gtfsvalidator.WithMaxMemory(1024 * 1024 * 1024), // 1GB memory limit
    gtfsvalidator.WithProgressCallback(func(info gtfsvalidator.ProgressInfo) {
        fmt.Printf("Progress: %.1f%% - %s\n", info.PercentComplete, info.CurrentValidator)
    }),
)

ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
defer cancel()

report, err := validator.ValidateFileWithContext(ctx, "large-feed.zip")
```

### Streaming CSV Processing

```go
// For processing very large CSV files without loading into memory
parser, err := parser.NewStreamingCSVParser(file, "stop_times.txt", nil)
if err != nil {
    log.Fatal(err)
}

processor := &parser.CountingProcessor{}
err = parser.ProcessStream(context.Background(), processor)
fmt.Printf("Processed %d rows with minimal memory usage\n", processor.Count)
```

### Web API Integration

```go
func validateHandler(w http.ResponseWriter, r *http.Request) {
    file, _, err := r.FormFile("gtfs")
    if err != nil {
        http.Error(w, "Failed to read file", http.StatusBadRequest)
        return
    }
    defer func() {
		if closeErr := file.Close(); closeErr != nil {
			log.Printf("Warning: failed to close %v", closeErr)
		}
	}()
    
    validator := gtfsvalidator.New(
        gtfsvalidator.WithValidationMode(gtfsvalidator.ValidationModePerformance),
    )
    
    report, err := validator.ValidateReader(file)
    if err != nil {
        http.Error(w, "Validation failed", http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(report)
}
```

## CLI Commands and Options

### Commands

```bash
gtfs-validator [flags]                    # Validate with flags (legacy style)
gtfs-validator validate <input> [flags]   # Validate with subcommand
gtfs-validator version                     # Show version information
gtfs-validator help                        # Show help
```

### Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--input` | `-i` | Path to GTFS feed (ZIP or directory) | *required* |
| `--mode` | `-m` | Validation mode: `performance`, `default`, `comprehensive` | `default` |
| `--format` | `-f` | Output format: `console`, `json`, `summary` | `console` |
| `--output` | `-o` | Output file path | `stdout` |
| `--country` | `-c` | Country code for validation | `US` |
| `--workers` | `-w` | Number of parallel workers | `4` |
| `--max-notices` | | Maximum notices per type (0 = no limit) | `0` |
| `--progress` | `-p` | Show progress bar | `false` |
| `--timeout` | `-t` | Validation timeout | `5m` |
| `--memory` | | Maximum memory usage in MB (0 = no limit) | `0` |

### Examples

```bash
# Basic validation with short flags
gtfs-validator -i feed.zip

# Subcommand with long flags
gtfs-validator validate feed.zip --mode performance --progress

# JSON output to file
gtfs-validator -i feed.zip -f json -o validation-report.json

# Custom settings
gtfs-validator validate feed.zip -m comprehensive -w 8 -t 10m

# Show help for specific command
gtfs-validator validate --help
```

## Validation Coverage

| | count |
|---|---|
| Codes emitted | **176** |
| — canonical | **133**, every rule in scope |
| — ours | 43 |
| Our codes that are ERROR | **0** |
| Severities disagreeing with canonical | **0** |

Scope is the 181 rules the canonical validator publishes, less 4 deprecated
upstream, 27 GTFS-Flex, 11 GTFS-Fares v2 and 6 that describe the canonical
validator's own execution model rather than anything about a feed. That leaves
133, all implemented, each at the severity MobilityData gives it.

`python3 scripts/scope_audit.py` scrapes the published rule set, scans the codes
this repo emits and reproduces every number above.
[CANONICAL_PARITY.md](CANONICAL_PARITY.md) explains what was declined and why.

### Validators

| Group | Count | What it looks at |
|---|---|---|
| **Core** | 13 | File presence, CSV structure, column names, field types, required fields |
| **Entity** | 16 | One record at a time: routes, stops, shapes, calendars, agencies |
| **Relationship** | 11 | References between files: foreign keys, stop times, translations, attributions |
| **Business** | 11 | Operational sense across the feed: speeds, transfers, frequencies, blocks, geometry |
| **Accessibility** | 2 | Pathways and levels |
| **Fare** | 1 | Fare attributes and rules |
| **Meta** | 1 | feed_info.txt |

[VALIDATOR_RULES.md](VALIDATOR_RULES.md) lists every validator with the codes it
emits and their severities. It is generated from the source, so it cannot drift
from what the code does.

### The 43 rules that are ours

They cover failure modes the canonical set does not model — several of them
break OpenTripPlanner graph builds, such as a trip whose stops repeat mid-route
or a station no stop is a child of. None of them is ERROR. A code MobilityData
does not define is our opinion, and an opinion should not fail someone's feed.

## Notice Types

| Severity | Meaning | Example |
|----------|---------|---------|
| **ERROR** | A canonical rule is broken; the feed is invalid | `missing_required_file` |
| **WARNING** | Worth fixing, but the feed still loads | `route_without_trips` |
| **INFO** | Observation, no action implied | `unused_station` |

Only canonical rules may be ERROR, so a feed that fails here is failing a rule
MobilityData enforces as well — never one we invented.

## Enhanced Error Descriptions

The validator provides comprehensive, user-friendly descriptions for all validation issues:

Notices are grouped by code, but severity belongs to the individual notice, not
to the group. One code can produce notices of different severities — a route
colour contrast below the WCAG threshold is a WARNING, while unreadable text is
an ERROR — so a group reports a breakdown and each sample carries its own
severity, file and line.

```json
{
  "code": "route_color_contrast",
  "severityCounts": { "errors": 1, "warnings": 3, "infos": 0, "total": 4 },
  "description": "Route colors have insufficient contrast for accessibility compliance.",
  "affectedFiles": ["routes.txt"],
  "totalNotices": 4,
  "sampleNotices": [
    { "severity": "ERROR", "file": "routes.txt", "line": 5, "routeId": "6", "actualContrast": 1.0 },
    { "severity": "WARNING", "file": "routes.txt", "line": 2, "routeId": "2", "actualContrast": 4.47 }
  ]
}
```

Samples are ordered most severe first, so the sample cap never hides the errors
in a group that is mostly warnings. `affectedFiles` lists the GTFS files a code
concerns, the file its rows belong to first; feed-wide summary notices have
neither a file nor a line.

**Features:**
Descriptions come from the Canonical GTFS Schedule Validator rule registry, so
the wording matches what MobilityData publishes. Regenerate them with:

```bash
python3 scripts/gen_notice_descriptions.py
```

Codes we emit that MobilityData does not define fall back to the code name.

## What changed in 1.1.5 and 1.1.6

[CHANGELOG.md](CHANGELOG.md) has the full mapping of every code added, removed,
renamed and re-graded. In short:

### 1.1.5 — the emitted rule set was reconciled against the canonical validator (breaking)

- **176 codes, down from 201.** All 133 in-scope canonical rules are now
  implemented, up from 40. Most of the removals are not lost coverage: the same
  check is reported under the canonical code.
- **No non-canonical code is ERROR any more**, down from 86. A feed could
  previously fail on checks nobody else recognises.
- **The per-field enum codes collapsed onto `unexpected_enum_value`.** That also
  closed a hole: the old checks ran only after a successful integer parse, so an
  enum holding a non-numeric value reported nothing at all.
- **Anything keying on specific codes needs the mapping tables in the
  changelog.** Codes were added, removed and renamed, and severities moved.
- **The expensive geometry checks moved behind comprehensive mode**, and the
  whole-feed graph build in the network topology validator — the costliest
  validator in the suite — was removed along with several duplicated passes over
  `stop_times.txt` and `shapes.txt`.
- Fixes worth naming: `route_short_name_too_long` counted bytes rather than
  runes, so a Cyrillic name tripped the 12-character limit at six characters;
  `file_structure_validator` was never constructed, so two canonical codes
  counted as implemented while never being emitted; `stop_sequence_gap` and
  `stop_name_missing_but_inherited` were pure false positives and are gone.

### 1.1.6 — the field tables are derived from the spec

- **Which fields a file has, and whether each is required, is now generated**
  from the GTFS Schedule reference into `schema/field_presence.go`: 31 files, 218
  fields. Kept by hand these lists drift, and they had — three fields the spec
  calls Optional were being reported as missing recommended fields.
- **`missing_recommended_field` now fires for the three Recommended fields the
  spec actually has**, all of them in `feed_info.txt`, and for nothing else.
- **A blank in a Required field whose value list offers "empty" is a value, not
  an omission** — `fare_attributes.transfers` empty means unlimited, and
  `transfers.transfer_type` empty means a recommended transfer point.
- **`leading_or_trailing_whitespaces` is emitted at last.** Its validator had
  never been registered, so the code counted as implemented while no feed could
  produce it. It now covers every column the spec defines, in every file it
  defines, and reports only the whitespace that survives parsing — the
  whitespace inside double quotes, which is what the canonical rule means and
  the only kind any other validator ever sees.

## Examples

See the [examples/](examples/) directory:
- [Basic Usage](examples/basic/) - Simple validation
- [Advanced Features](examples/advanced/) - Progress tracking, cancellation
- [Web API Server](examples/api-server/) - HTTP API integration
- [Streaming CSV Processing](examples/streaming-csv/) - Memory-efficient large file processing
- [Configuration Validation](examples/config-validation/) - Configuration sanitization and validation
- [Large Feeds Optimization](examples/large-feeds/) - Memory optimization for massive feeds

## Project Structure

```
.
├── validator.go           # Public API
├── implementation.go      # Internal logic  
├── doc.go                # Package docs
├── cmd/gtfs-validator/   # CLI tool
├── examples/             # Usage examples
├── notice/               # Notice system
├── parser/               # GTFS parsing (including streaming CSV parser)
├── pools/                # Memory pooling for performance optimization
├── logging/              # Structured logging system
├── report/               # Report generation
├── validator/            # Individual validators
├── schema/               # GTFS data types, and the spec tables generated from the reference
├── scripts/              # Generators and the scope audit
├── docs/validation-scope/ # The scope proposal and the scraped canonical rule set
└── types/                # Custom types
```

## Performance & Reliability

### **Real-world performance**

Measured on the Sofia feed — 198 routes, 30,069 trips, 4,447 stops, 684,740 stop
times — on a 10-core laptop:

| Mode | Wall clock | Peak memory | Notices |
|---|---|---|---|
| Performance | ~14 s | ~1.0 GB | 0 errors, 1,191 warnings, 5,110 infos |
| Default | ~19 s | ~1.0 GB | 104 errors, 29,169 warnings, 5,110 infos |
| Comprehensive | ~22 s | ~0.9 GB | 104 errors, 29,169 warnings, 6,729 infos |

Memory is spent on holding the feed, not on running the checks, which is why
performance mode is no lighter than comprehensive. Budget roughly a gigabyte for
a feed of this size and cap it with `WithMaxMemory` if that matters.

Every one of those 104 errors is a canonical rule being broken, since no rule of
ours may be ERROR.

### **Production ready**
- ✅ **Every validator has tests** — 55 of 55, with package statement coverage between 65% and 90%
- ✅ **Thread-safe**: concurrent validation with configurable worker pools
- ✅ **Streaming processing**: handles feeds with millions of records without loading a file whole
- ✅ **Enterprise features**: timeouts, cancellation, progress reporting, memory limits
- ✅ **Structured logging**: JSON/text logging with configurable levels
- ✅ **Performance monitoring**: built-in benchmarking and statistics tracking

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Roadmap

See [ROADMAP.md](ROADMAP.md) for planned features and future development direction.

## Contributing

Contributions welcome! Please see contributing guidelines for development setup and pull request process.