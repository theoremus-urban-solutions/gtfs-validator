# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- **Memory Pooling System**: Comprehensive memory pools for CSV parsing to reduce garbage collection overhead
- **Streaming CSV Parser**: High-performance streaming parser for processing massive CSV files (2-4M rows/sec)
- **Structured Logging**: JSON and text formatters with configurable levels (DEBUG, INFO, WARN, ERROR)
- **Configuration Validation**: Automatic config sanitization and bounds checking with detailed error messages
- **Performance Benchmarking**: Built-in benchmarking documentation and performance monitoring
- **Memory Optimization Examples**: Complete examples for processing large feeds with minimal memory usage
- **Streaming Validation**: Context-aware streaming validation with real-time progress reporting
- **Advanced Examples**: Streaming CSV processing, config validation, and large feed optimization examples
- Comprehensive testing infrastructure with unit, integration, and CLI tests
- Performance benchmarks for validation operations
- Thread-safe notice container implementation
- GitHub Actions CI/CD pipeline with multi-platform testing
- Community contribution guidelines and templates
- Development tooling (Makefile, linting configuration)
- Cross-platform binary builds
- Cobra-based CLI with subcommands, help, and autocompletion
- Missing validator test coverage (5 new test files created)

### Changed
- **CSV Parser Integration**: Integrated memory pools into existing CSV parser for automatic memory optimization
- **Modern Go Practices**: Replaced deprecated `ioutil` functions with modern `os` equivalents throughout codebase
- **Enhanced Documentation**: Updated README, CLI help, and API docs to reflect streaming and memory features
- **Performance Improvements**: Memory pooling reduces GC pressure, streaming parser enables constant memory usage
- Improved error handling and context propagation
- CLI interface updated to use Cobra framework for better UX
- Documentation updated with modern CLI examples
- README enhanced to highlight comprehensive validation coverage (294+ rules vs ~60 official)

### Fixed
- GTFS time validation now correctly supports late-night service times (25:30:00+)
- Time parsing no longer rejects valid GTFS times beyond 24:00:00
- Thread safety issues in concurrent validation

### Performance
- **2-4M rows/sec** sustained throughput on large GTFS files (tested with Sofia GTFS 588K+ records)
- **Constant memory usage** regardless of file size with streaming CSV processing
- **Memory pool optimization** reduces garbage collection overhead during CSV parsing
- **Streaming processing** enables validation of feeds with millions of records without OOM errors

## [1.0.0] - Initial Release

### Added
- Complete GTFS validation library for Go
- 60+ validators covering all GTFS specification requirements
- Multiple validation modes (performance, default, comprehensive)
- CLI tool with rich output formats (console, JSON, summary)
- Context support for cancellation and timeouts
- Progress reporting with callbacks
- Configurable worker pools for parallel processing
- Comprehensive error reporting with notice system
- Examples for library usage, advanced features, and web API integration

### Core Features
- **File Structure Validation**: Required files, CSV format, encoding
- **Data Format Validation**: Field types, ranges, patterns
- **Entity Validation**: Routes, stops, trips, agencies consistency
- **Relationship Validation**: Foreign keys, stop sequences, service consistency
- **Business Logic Validation**: Travel speeds, transfers, frequency overlaps
- **Accessibility Validation**: Pathways, wheelchair access
- **Geospatial Validation**: Coordinate analysis, geographic clustering

### Performance
- Optimized for large feeds with parallel processing
- Memory efficient parsing and validation
- Configurable resource limits
- Benchmark results: ~10-12 seconds for performance mode on large feeds

### Documentation
- Complete API documentation
- Usage examples and tutorials
- CLI reference guide
- Contributing guidelines