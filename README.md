# GTFS Validator Go

[![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A fast, comprehensive GTFS (General Transit Feed Specification) validator library for Go. Validates transit feeds against the official specification with detailed error reporting and multiple output formats.

## Features

- **🚀 Fast Validation**: Optimized for large feeds with parallel processing
- **📋 Comprehensive**: 60+ validators covering all GTFS specification requirements  
- **🔧 Multiple Modes**: Performance, default, and comprehensive validation modes
- **⚡ Concurrent**: Thread-safe with configurable worker pools
- **⏰ Context Support**: Cancellation, timeouts, and progress reporting
- **📊 Rich Reports**: JSON, console, and summary output formats
- **📦 Dual Purpose**: Use as Go library or standalone CLI tool

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
# Validate a feed
gtfs-validator -input feed.zip

# Fast mode with JSON output  
gtfs-validator -input feed.zip -mode performance -format json
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
    gtfsvalidator.WithProgressCallback(func(info gtfsvalidator.ProgressInfo) {
        fmt.Printf("Progress: %.1f%% - %s\n", info.PercentComplete, info.CurrentValidator)
    }),
)

ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
defer cancel()

report, err := validator.ValidateFileWithContext(ctx, "large-feed.zip")
```

### Web API Integration

```go
func validateHandler(w http.ResponseWriter, r *http.Request) {
    file, _, err := r.FormFile("gtfs")
    if err != nil {
        http.Error(w, "Failed to read file", http.StatusBadRequest)
        return
    }
    defer file.Close()
    
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

## CLI Options

```
-input string        Path to GTFS feed (ZIP or directory) [required]
-mode string         Validation mode: performance, default, comprehensive
-format string       Output format: console, json, summary  
-output string       Output file (default: stdout)
-country string      Country code for validation (default: US)
-workers int         Parallel workers (default: 4)
-max-notices int     Notice limit per type (default: 100)
-progress            Show progress bar
-timeout duration    Validation timeout (default: 5m)
```

## Validation Categories

- **Core**: File structure, required fields, data formats
- **Entity**: Route/stop consistency, calendar validation
- **Relationship**: Foreign keys, stop sequences, cross-file integrity  
- **Business**: Travel speeds, transfers, frequency overlaps
- **Accessibility**: Pathways, wheelchair access
- **Geospatial**: Coordinate analysis, geographic clustering

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

## Project Structure

```
.
├── validator.go           # Public API
├── implementation.go      # Internal logic  
├── doc.go                # Package docs
├── cmd/gtfs-validator/   # CLI tool
├── examples/             # Usage examples
├── notice/               # Notice system
├── parser/               # GTFS parsing
├── report/               # Report generation
├── validator/            # Individual validators
├── schema/               # GTFS data types
└── types/                # Custom types
```

## Performance

Tested on large feeds (180 routes, 588k stop times):
- **Performance mode**: 10-12 seconds
- **Memory usage**: ~200MB peak
- **Parallel processing**: Scales with CPU cores
- **Context cancellation**: Sub-second response time

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Contributing

Contributions welcome! Please see contributing guidelines for development setup and pull request process.