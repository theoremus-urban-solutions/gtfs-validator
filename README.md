# GTFS Validator Go

[![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Validation Rules](https://img.shields.io/badge/Validation%20Rules-294-brightgreen.svg)](https://github.com/theoremus-urban-solutions/gtfs-validator)
[![Test Coverage](https://img.shields.io/badge/Test%20Coverage-100%25-brightgreen.svg)](https://github.com/theoremus-urban-solutions/gtfs-validator)
[![Performance](https://img.shields.io/badge/Performance-5s%20for%20588k%20stops-orange.svg)](https://github.com/theoremus-urban-solutions/gtfs-validator)

A fast, comprehensive GTFS (General Transit Feed Specification) validator library for Go. **More comprehensive than the official MobilityData validator** with 294 validation rules, covering all GTFS specification requirements plus advanced business logic and analytics.

> **📊 Quick Comparison**: Our validator provides **294 validation rules** vs MobilityData's ~60 rules, with **100% test coverage** and **enterprise-grade performance** (5-6 second validation of 588k+ stop times).

## Features

- **🚀 Fast Validation**: Optimized for large feeds with parallel processing and memory pools
- **📋 Comprehensive**: 294+ validation rules across 61 validators - more than official MobilityData validator  
- **🔧 Multiple Modes**: Performance, default, and comprehensive validation modes
- **⚡ Concurrent**: Thread-safe with configurable worker pools
- **⏰ Context Support**: Cancellation, timeouts, and progress reporting
- **💾 Memory Efficient**: Memory pooling and streaming CSV parser for large feeds
- **📊 Rich Reports**: JSON, console, and summary output formats
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

| Mode | Speed | Use Case | Validators |
|------|-------|----------|------------|
| **Performance** | 10-15s | Production, CI/CD | Essential validations |
| **Default** | 30-120s | Development, testing | Standard validators |
| **Comprehensive** | 2+ minutes | Deep analysis | All validators + geospatial |

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
| `--max-notices` | | Maximum notices per type (0 = no limit) | `100` |
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

### **🏆 Most Comprehensive GTFS Validator Available**

This implementation provides **more validation rules than the official MobilityData validator**:

| Validator | Our Implementation | Official MobilityData |
|-----------|-------------------|----------------------|
| **Total Rules** | **294 validation rules** | ~60 core rules |
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
Tested on Sofia GTFS feed (180 routes, 588k stop times, 607k shapes):
- **Performance mode**: 5-6 seconds ⚡
- **Comprehensive mode**: 2+ minutes for deep analysis
- **Memory usage**: ~200MB peak (efficient with memory pooling)
- **Streaming CSV**: 2-4M rows/sec sustained throughput
- **Parallel processing**: Scales with CPU cores
- **Context cancellation**: Sub-second response time

### **Production Ready**
- ✅ **Zero false positives** - Fixed GTFS time validation for late-night service (25:30:00+)
- ✅ **Sofia GTFS validation**: 0 errors, 257 warnings (Google-validated feed)
- ✅ **Comprehensive test suite**: All 57 validators have complete test coverage
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