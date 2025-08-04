# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Comprehensive testing infrastructure with unit, integration, and CLI tests
- Performance benchmarks for validation operations
- Thread-safe notice container implementation
- GitHub Actions CI/CD pipeline with multi-platform testing
- Community contribution guidelines and templates
- Development tooling (Makefile, linting configuration)
- Cross-platform binary builds

### Changed
- Improved error handling and context propagation
- Enhanced test data with future-proof dates
- Updated documentation with comprehensive examples

### Fixed
- Thread safety issues in concurrent validation
- Import path inconsistencies
- UTF-8 BOM handling in CSV parsing

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