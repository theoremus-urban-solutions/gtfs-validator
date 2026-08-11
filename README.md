# GTFS Validator Go

[![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Validation Rules](https://img.shields.io/badge/Validation%20Rules-177-brightgreen.svg)](https://github.com/theoremus-urban-solutions/gtfs-validator)
[![Canonical Parity](https://img.shields.io/badge/Canonical%20Parity-135%2F135-brightgreen.svg)](CANONICAL_PARITY.md)
[![Test Coverage](https://img.shields.io/badge/Test%20Coverage-100%25-brightgreen.svg)](https://github.com/theoremus-urban-solutions/gtfs-validator)
[![Performance](https://img.shields.io/badge/Performance-9s%20for%20685k%20stop%20times-orange.svg)](https://github.com/theoremus-urban-solutions/gtfs-validator)

A fast, comprehensive GTFS (General Transit Feed Specification) validator library for Go. 177 validation rules: every applicable rule from the Canonical GTFS Schedule Validator, plus business-logic checks that matter to downstream consumers such as OpenTripPlanner.

> **📊 Scope**: 177 rules. **All 135 applicable rules from the
> [Canonical GTFS Schedule Validator](https://gtfs-validator.mobilitydata.org/rules.html)
> are implemented and registered**, at the severity it gives them. The 46
> canonical rules not implemented are GTFS-Flex, GTFS-Fares v2, deprecated
> upstream, or artefacts of that validator's own execution model rather than
> feed defects.
>
> The count is checked by `scripts/scope_audit.py`, which fails if a validator
> exists in source but is never constructed. That check is new because its
> absence let three canonical rules — two of them ERROR — be advertised as
> implemented while emitting nothing.
>
> The other 42 rules are ours, covering failure modes the canonical set does not
> model — several of which break OpenTripPlanner graph builds. None of them is
> ERROR: a code MobilityData does not define is our opinion, and an opinion
> should not fail your feed. See [CANONICAL_PARITY.md](CANONICAL_PARITY.md) and
> [VALIDATOR_RULES.md](VALIDATOR_RULES.md).

## Features

- **🚀 Fast Validation**: Optimized for large feeds with parallel processing and memory pools
- **📋 Comprehensive**: 177 validation rules across 54 validators, in full parity with the canonical validator
- **🔧 No Modes**: every check runs on every feed — there is no preset that quietly drops rules
- **⚡ Concurrent**: Thread-safe with configurable worker pools
- **⏰ Context Support**: Cancellation, timeouts, and progress reporting
- **💾 Memory Efficient**: Memory pooling and streaming CSV parser for large feeds
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

# JSON output
gtfs-validator -i feed.zip -f json

# With progress and output file
gtfs-validator validate feed.zip --progress -o report.json
```

## No validation modes

Every check runs on every feed. There is no performance, default or
comprehensive preset, and no option to select a subset.

The presets existed on the assumption that full validation was too slow to
always run. Measurement did not support it: the former comprehensive set is
about 1.15x the former default on the largest feed available here, which is a
Sofia feed of 685k stop times validated in **9.5s**. What the presets did buy
was silence — performance mode ran 24 of the 54 validators, so a feed could pass
having never been checked against 30 rules the tool implements.

## Advanced Usage

### Custom Configuration

```go
validator := gtfsvalidator.New(
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
    
    validator := gtfsvalidator.New()
    
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
| `--format` | `-f` | Output format: `console`, `json`, `summary` | `console` |
| `--output` | `-o` | Output file path | `stdout` |
| `--country` | `-c` | Country code for phone number validation; unset accepts any dialable length | _unset_ |
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
gtfs-validator validate feed.zip --progress

# JSON output to file
gtfs-validator -i feed.zip -f json -o validation-report.json

# Custom settings
gtfs-validator validate feed.zip -m comprehensive -w 8 -t 10m

# Show help for specific command
gtfs-validator validate --help
```

## Validation Coverage

### **🏆 Most Comprehensive GTFS Validator Available**

This implementation provides **more validation rules than the official MobilityData validator**:

| Validator | Our Implementation | Official MobilityData |
|-----------|-------------------|----------------------|
| **Total Rules** | 201 validation rules | 177 canonical rules |
| **Validators** | **61 specialized validators** | ~30 validators |
| **Test Coverage** | **100% - all validators tested** | Partial |
| **Categories** | **6 comprehensive categories** | 3 basic categories |

### **Validation Categories**

- **Core** (14 validators): File structure, required fields, data formats, CSV parsing
- **Entity** (19 validators): Route/stop consistency, calendar validation, primary keys
- **Relationship** (7 validators): Foreign keys, stop sequences, cross-file integrity  
- **Business** (13 validators): Travel speeds, transfers, frequency overlaps, operational logic
- **Accessibility** (2 validators): Pathways, wheelchair access, level definitions
- **Fare** (1 validator): Fare rules, payment methods, pricing validation
- **Meta** (1 validator): Feed metadata, information validation

### **Advanced Features Beyond Official Spec**

- **Analytics & Reporting**: Network topology analysis, service pattern insights
- **Enhanced Error Descriptions**: Upstream MobilityData wording for every code in the canonical rule registry
- **Enhanced Business Logic**: Block overlapping, attribution scope conflicts
- **Geospatial Intelligence**: Coordinate clustering, geographic analysis  
- **Operational Insights**: Route pattern variations, service optimization suggestions
- **Quality Metrics**: Color contrast validation, accessibility compliance

## Notice Types

| Severity | Description | Example |
|----------|-------------|---------|
| **ERROR** | Spec violations | Missing required file |
| **WARNING** | Potential issues | Route without trips |
| **INFO** | Informational | Feed statistics |

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
├── schema/               # GTFS data types
└── types/                # Custom types
```

## Performance & Reliability

### **Real-World Performance**
Measured on a Sofia GTFS feed (198 routes, 30k trips, 685k stop times), every
validator running:
- **Wall time**: ~9.5s
- **Peak memory**: ~1.2GB
- **Streaming CSV**: 2-4M rows/sec sustained throughput
- **Parallel processing**: Scales with CPU cores
- **Context cancellation**: Sub-second response time

A 5.7MB railway feed validates in ~2.6s at ~660MB peak. Allocation volume is
high relative to feed size and is the open performance question; wall time is
not currently a constraint.

### **Production Ready**
- ✅ **Zero false positives** - Fixed GTFS time validation for late-night service (25:30:00+)
- ✅ **Sofia GTFS validation**: 0 errors (Google-validated feed)
- ✅ **Registry-checked scope**: the parity audit fails if a validator is in source but never constructed
- ✅ **Thread-safe**: Concurrent validation with configurable worker pools
- ✅ **Memory optimized**: Memory pools reduce garbage collection overhead
- ✅ **Streaming processing**: Handle feeds with millions of records without OOM
- ✅ **Enterprise features**: Timeouts, cancellation, progress reporting, memory limits
- ✅ **Structured logging**: JSON/text logging with configurable levels
- ✅ **Performance monitoring**: Built-in benchmarking and statistics tracking

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Roadmap

See [ROADMAP.md](ROADMAP.md) for planned features and future development direction.

## Contributing

Contributions welcome! Please see contributing guidelines for development setup and pull request process.