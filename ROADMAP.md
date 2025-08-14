# GTFS Validator Library Roadmap

This document outlines the future development plans for the GTFS Validator Go library and CLI tool.

## Vision

To provide a high-performance, comprehensive GTFS validation library that can be easily integrated into other applications and services, with a simple CLI for quick validation tasks.

## Current Features ✅

### Already Implemented
- **61 validators** across 6 categories (Core, Entity, Relationship, Business, Accessibility, Fare)
- **294+ validation rules** - more comprehensive than official MobilityData validator
- **Multiple validation modes**: Performance, Default, Comprehensive
- **Streaming CSV processing** for large files (2-4M rows/sec)
- **Memory pooling** for efficient memory usage
- **Parallel processing** with configurable workers
- **Context support** for cancellation and timeouts
- **Progress reporting** with callbacks
- **Multiple output formats**: Console, JSON, Summary, HTML
- **CLI tool** with Cobra framework
- **100% test coverage** for all validators
- **Benchmarking suite** included
- **GTFS-Pathways** validation (via PathwayValidator)
- **Accessibility compliance** checking (LevelValidator, PathwayValidator)
- **Geospatial validation** (GeospatialValidator in comprehensive mode)

## Planned Enhancements

### Performance Optimizations
- [ ] **Intelligent caching** for repeated validations of the same feed
- [ ] **Incremental validation** for feed updates (validate only changed files)
- [ ] **Memory optimization** for feeds > 1GB
- [ ] **Parallel ZIP extraction** for faster startup

### Library Enhancements
- [ ] **Validation profiles** - Pre-configured rule sets for different agency types
- [ ] **Custom validation rules** - Plugin system for agency-specific validators
- [ ] **Validation rule configuration** - Enable/disable specific rules via config
- [ ] **Diff validation** - Compare two feed versions and validate changes
- [ ] **Auto-fix capabilities** - Programmatically fix common issues
- [ ] **Better error recovery** - Continue validation after encountering malformed files

### Output Improvements
- [ ] **Structured validation reports** - Machine-readable report formats
- [ ] **CSV export** for notices
- [ ] **SARIF format** - Static Analysis Results Interchange Format for CI/CD
- [ ] **JUnit XML** - Test result format for CI integration

### GTFS Extensions Support
- [ ] **GTFS-Realtime validation** - Validate GTFS-RT feeds
- [ ] **GTFS-RT/Static cross-validation** - Ensure RT matches static data
- [ ] **GTFS-Flex v2.1** - Demand-responsive transit validation
- [ ] **GTFS-Fares v2** - Enhanced fare validation
- [ ] **GTFS-Vehicles** - Vehicle information validation

### Developer Experience
- [ ] **Go modules** for individual validator categories
- [ ] **Validation middleware** - Easy integration with HTTP servers
- [ ] **gRPC support** - For microservice architectures
- [ ] **WebAssembly build** - Run validator in browsers
- [ ] **Better API documentation** - GoDoc improvements
- [ ] **More code examples** - Common integration patterns

### CLI Improvements
- [ ] **Watch mode** - Monitor directory for changes
- [ ] **Batch validation** - Validate multiple feeds
- [ ] **Config file support** - Store common CLI flags
- [ ] **Shell completion** - Better autocomplete support

## Technical Debt

### Code Quality
- [ ] **Mutation testing** - Ensure test quality
- [ ] **Property-based testing** - For complex validators
- [ ] **Fuzz testing** - For parser robustness
- [ ] **Static analysis** - Additional linting rules

### Performance
- [ ] **Optimize validator execution order** - Run fast validators first
- [ ] **Result caching** - Cache expensive computations
- [ ] **Better memory management** - Reduce allocations

### Architecture
- [ ] **Validator plugin system** - Dynamic validator loading
- [ ] **Validation pipeline** - Composable validation stages
- [ ] **Event-driven notifications** - Real-time validation events
- [ ] **Better error types** - More specific error handling

## Library Integration Examples

### Priority Integration Patterns
1. **REST API Server** - Example HTTP server wrapping the library
2. **gRPC Service** - Example microservice implementation
3. **Lambda Function** - Serverless validation example
4. **Kubernetes Job** - Batch processing example
5. **GitHub Action** - CI/CD integration example

## Success Metrics

### Performance Goals
- Validate 10M stop times in < 30 seconds
- Support feeds up to 10GB with streaming
- < 100MB memory footprint for streaming mode
- < 100ms library initialization time

### Library Adoption Goals
- 1,000+ GitHub stars
- 50+ projects using as dependency
- < 24hr issue response time
- Zero breaking API changes
- Semantic versioning compliance

### Code Quality Goals
- Maintain 100% test coverage
- Zero critical bugs
- All validators documented
- Examples for every public API

## Non-Goals

This project focuses on being a validation library and simple CLI tool. The following are explicitly out of scope:

- ❌ Web UI (separate project)
- ❌ Database storage
- ❌ User authentication
- ❌ Multi-tenancy
- ❌ Cloud hosting service
- ❌ Feed monitoring service
- ❌ Visualization features
- ❌ Feed generation/editing

These features should be implemented in applications that use this library.

## Contributing

Priority areas for contribution:

1. **New validators** - Additional validation rules
2. **GTFS extensions** - Support for GTFS-RT, GTFS-Flex
3. **Performance** - Optimizations and benchmarks
4. **Documentation** - API docs and examples
5. **Test cases** - Edge cases and real-world feeds

## Compatibility Commitment

- Semantic versioning (SemVer)
- No breaking changes in minor releases
- Deprecation notices for 2 versions
- Migration guides for major changes
- Backward compatibility for CLI flags

## Release Schedule

- **Patch releases**: As needed for bug fixes
- **Minor releases**: Quarterly with new features
- **Major releases**: Annually (if needed)

---