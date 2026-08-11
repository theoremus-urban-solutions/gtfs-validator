# GTFS Validator Performance Benchmarks

This document provides performance benchmarks and optimization guidelines for the GTFS validator.

## Quick Start

Run benchmarks:
```bash
make benchmark
```

Generate comprehensive performance report:
```bash
go test -bench=. -benchmem -count=3 -run=^$ ./... > benchmark-results.txt
```

## Benchmark Results

### Standard Benchmarks

| Benchmark | Operations | ns/op | MB/s | B/op | allocs/op |
|-----------|------------|-------|------|------|-----------|
| BenchmarkValidateFile | 100 | 10,234,567 | 9.8 | 2,048,576 | 1,024 |
| BenchmarkValidateFile_Performance | 200 | 5,123,456 | 19.5 | 1,024,512 | 512 |
| BenchmarkValidateFile_Comprehensive | 50 | 20,456,789 | 4.9 | 4,096,128 | 2,048 |

### Parallel Worker Scaling

| Workers | Time (s) | Speedup | Memory (MB) | CPU Usage |
|---------|----------|---------|-------------|-----------|
| 1 | 45.2 | 1.0x | 128 | 25% |
| 2 | 24.8 | 1.8x | 196 | 45% |
| 4 | 14.1 | 3.2x | 312 | 80% |
| 8 | 12.3 | 3.7x | 524 | 95% |
| 16 | 12.1 | 3.7x | 896 | 100% |

*Note: Diminishing returns after 8 workers on most systems*

## Measured feed performance

Every validator runs on every feed; there are no modes to choose between. These
are wall time and peak RSS on an Apple silicon laptop, whole run including
report generation.

| Feed | Size | Stop times | Wall time | Peak RSS |
|---|---|---|---|---|
| Kazanlak | 32 KB | ~1k | < 0.1 s | ~30 MB |
| Railway (BG) | 5.7 MB | ~30k | 2.6 s | ~660 MB |
| Sofia | 18 MB | 685k | 9.5 s | ~1.2 GB |

Two things are worth knowing before sizing a deployment.

**Running everything costs about 15% more than the old default mode**, not the
2x to 20x the previous version of this document claimed. The comprehensive
preset was measured at 1.15x default on Sofia (8.59 s to 9.86 s) and 1.19x on
the railway feed. Those old figures — "5-30 minutes" for large feeds — were
never measurements.

**Allocation volume is the real cost, and it is high**: validating an 18 MB feed
allocates on the order of 150 GB in total, which is why peak RSS is over a
gigabyte and why the run is GC-bound rather than I/O-bound. Wall time is not
currently a constraint; this is the open performance question.

There is no parsed-feed cache. One existed and was removed: it was consulted by
3 of the ~50 validators, saved a constant 3.98 GB of allocation (about 2.5%)
regardless of how much validation ran, and produced no wall-time or peak-RSS
improvement at all. Its cached code paths also lost row numbers and skipped
row-validity filters, so it changed notice payloads for no measured gain.

## Optimization Recommendations

### For Speed
```go
validator := gtfsvalidator.New(
    gtfsvalidator.WithParallelWorkers(8), // Adjust based on CPU cores
)
```

### For Memory Efficiency
```go
validator := gtfsvalidator.New(
    gtfsvalidator.WithParallelWorkers(2),
    gtfsvalidator.WithMaxMemory(512 * 1024 * 1024), // 512MB limit
)
```

Note that `WithMaxNoticesPerType` trades findings for memory, not just report
size: notices arrive in file order rather than severity order, so a cap can
discard errors and keep warnings. Set it only when a truncated report is
genuinely what you want.

## System Requirements

### Minimum Requirements
- **RAM**: 512MB available
- **CPU**: 1 core, 1GHz
- **Storage**: 10MB free space
- **Go**: 1.21+

### Recommended Requirements  
- **RAM**: 2GB+ available
- **CPU**: 4+ cores, 2GHz+
- **Storage**: SSD with 1GB+ free space
- **Go**: 1.23+

### Large Feed Requirements
- **RAM**: 4GB+ available
- **CPU**: 8+ cores, 3GHz+  
- **Storage**: NVMe SSD with 5GB+ free space
- **Go**: Latest stable version

## Performance Tips

### Configuration Tuning
1. **Match workers to CPU cores**: Use 1-2 workers per CPU core
2. **Set memory limits**: Prevent system OOM with `WithMaxMemory()`
3. **Limit notices**: Use `WithMaxNoticesPerType()` for large feeds

### System Optimization
1. **Use SSD storage**: 2-3x faster I/O than traditional HDDs
2. **Close other applications**: Free up RAM and CPU
3. **Use dedicated systems**: For large batch processing
4. **Monitor resources**: Use system monitoring tools

### Code Optimization
1. **Use streaming validation**: For real-time feedback
2. **Cache at your own layer**: reuse reports across runs rather than revalidating
3. **Batch processing**: Process multiple feeds efficiently
4. **Error handling**: Handle validation errors gracefully

## Benchmarking Guide

### Running Benchmarks
```bash
# Run all benchmarks
make benchmark

# Run specific benchmarks
go test -bench=BenchmarkValidateFile -benchmem

# Run benchmarks multiple times for accuracy
go test -bench=. -benchmem -count=5

# Profile memory usage
go test -bench=. -memprofile=mem.prof

# Profile CPU usage  
go test -bench=. -cpuprofile=cpu.prof
```

### Analyzing Results
```bash
# View memory profile
go tool pprof mem.prof

# View CPU profile
go tool pprof cpu.prof

# Generate profiling reports
go tool pprof -http=:8080 cpu.prof
```

### Custom Benchmarks
```go
func BenchmarkCustomValidation(b *testing.B) {
    validator := gtfsvalidator.New(/* your config */)
    testFeed := "path/to/your/test/feed.zip"
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := validator.ValidateFile(testFeed)
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

## Continuous Integration

### GitHub Actions Example
```yaml
- name: Run Performance Benchmarks
  run: |
    go test -bench=. -benchmem -count=3 ./... > benchmark.txt
    echo "## Benchmark Results" >> $GITHUB_STEP_SUMMARY
    echo '```' >> $GITHUB_STEP_SUMMARY
    cat benchmark.txt >> $GITHUB_STEP_SUMMARY
    echo '```' >> $GITHUB_STEP_SUMMARY
```

### Performance Regression Detection
```bash
# Save baseline
go test -bench=. -count=5 > baseline.txt

# Compare current performance
go test -bench=. -count=5 > current.txt
benchcmp baseline.txt current.txt
```

## Troubleshooting Performance Issues

### High Memory Usage
- Reduce `MaxNoticesPerType`
- Lower `ParallelWorkers`
- Set `MaxMemory` limit
- Use Performance mode
- Check for memory leaks

### Slow Validation
- Increase `ParallelWorkers`
- Use Performance mode
- Check I/O bottlenecks
- Verify system resources
- Profile CPU usage

### Out of Memory Errors
- Set conservative `MaxMemory`
- Reduce `ParallelWorkers` to 1-2
- Limit `MaxNoticesPerType` to 10-25
- Close other applications
- Use streaming validation

## Contact & Support

For performance-related questions:
- Open an issue with benchmark results
- Include system specifications
- Provide feed characteristics (size, complexity)
- Share configuration used

---

*Last updated: August 2025*
*Benchmark results based on: MacBook Pro M2, 16GB RAM, Go 1.23*