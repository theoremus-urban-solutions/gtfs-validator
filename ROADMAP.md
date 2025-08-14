# GTFS Validator Library Roadmap

This document outlines the future development plans for the GTFS Validator Go library and CLI tool.

## Vision

To provide a high-performance, comprehensive GTFS validation library that can be easily integrated into other applications and services, with a simple CLI for quick validation tasks.

### Library Enhancements
- [ ] **Validation profiles** - Pre-configured rule sets for different agency types
- [ ] **Custom validation rules** - Plugin system for agency-specific validators
- [ ] **Validation rule configuration** - Enable/disable specific rules via config
- [ ] **Diff validation** - Compare two feed versions and validate changes
- [ ] **Auto-fix capabilities** - Programmatically fix common issues
- [ ] **Better error recovery** - Continue validation after encountering malformed files

### GTFS Extensions Support
- [ ] **GTFS-Realtime validation** - Validate GTFS-RT feeds
- [ ] **GTFS-RT/Static cross-validation** - Ensure RT matches static data
- [ ] **GTFS-Flex v2.1** - Demand-responsive transit validation
- [ ] **GTFS-Fares v2** - Enhanced fare validation
- [ ] **GTFS-Vehicles** - Vehicle information validation

### Developer Experience
- [ ] **Go modules** for individual validator categories
- [ ] **Validation middleware** - Easy integration with HTTP servers

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

---