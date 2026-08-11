// Large feeds memory optimization example shows how to validate
// large GTFS feeds efficiently with memory management
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator"
)

const memoryOptimizedMode = "memory-optimized"

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: large-feeds <large-gtfs-file> [mode]")
		fmt.Println("Modes: " + memoryOptimizedMode + " (default), fast, comprehensive")
		os.Exit(1)
	}

	gtfsFile := os.Args[1]
	mode := memoryOptimizedMode
	if len(os.Args) > 2 {
		mode = os.Args[2]
	}

	fmt.Printf("🔧 Large GTFS Feed Validation - Mode: %s\n", mode)
	fmt.Printf("📁 File: %s\n", gtfsFile)
	fmt.Println()

	// Check file size
	if info, err := os.Stat(gtfsFile); err == nil {
		sizeGB := float64(info.Size()) / (1024 * 1024 * 1024)
		fmt.Printf("📊 Feed size: %.2f GB\n", sizeGB)
		if sizeGB > 1.0 {
			fmt.Printf("⚠️  Large feed detected - using memory optimization strategies\n")
		}
	}

	// Show initial memory usage
	showMemoryUsage("Initial")

	// Create validator based on mode
	var validator gtfsvalidator.Validator
	var description string

	switch mode {
	case memoryOptimizedMode:
		validator, description = createMemoryOptimizedValidator()
	case "fast":
		validator, description = createFastValidator()
	case "comprehensive":
		validator, description = createComprehensiveValidator()
	default:
		log.Fatalf("Unknown mode: %s", mode)
	}

	fmt.Printf("🎯 Configuration: %s\n", description)
	fmt.Println()

	// Memory monitor
	memMonitor := &MemoryMonitor{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start memory monitoring
	go memMonitor.Start(ctx)

	// Perform validation with progress tracking
	fmt.Println("🚀 Starting validation...")
	startTime := time.Now()

	report, err := validator.ValidateFile(gtfsFile)

	elapsed := time.Since(startTime)
	cancel() // Stop memory monitoring

	if err != nil {
		fmt.Printf("❌ Validation failed: %v\n", err)
		os.Exit(1) //nolint:gocritic // cancel() already called on line 82
	}

	// Final results
	fmt.Println()
	fmt.Printf("✅ Validation completed in %.2f seconds\n", elapsed.Seconds())

	// Memory summary
	memMonitor.PrintSummary()

	// Validation results
	printValidationSummary(report, elapsed)

	// Memory optimization tips
	printMemoryOptimizationTips(mode, report)
}

// Memory-optimized configuration for large feeds
func createMemoryOptimizedValidator() (gtfsvalidator.Validator, string) {
	validator := gtfsvalidator.New(
		// Limit memory usage to 1GB
		gtfsvalidator.WithMaxMemory(1024*1024*1024),

		// Reduce parallelism to limit memory usage
		gtfsvalidator.WithParallelWorkers(2),

		// Limit notices to prevent memory growth
		gtfsvalidator.WithMaxNoticesPerType(25),

		// Add progress callback for monitoring
		gtfsvalidator.WithProgressCallback(func(info gtfsvalidator.ProgressInfo) {
			if int(info.PercentComplete)%25 == 0 {
				showMemoryUsage(fmt.Sprintf("Progress %.0f%%", info.PercentComplete))
			}
		}),
	)

	return validator, "Memory-optimized: 1GB limit, 2 workers, 25 notices/type"
}

// Fast configuration prioritizing speed over thoroughness
func createFastValidator() (gtfsvalidator.Validator, string) {
	validator := gtfsvalidator.New(
		// High parallelism for speed
		gtfsvalidator.WithParallelWorkers(8),

		// Generous memory limit
		gtfsvalidator.WithMaxMemory(2*1024*1024*1024), // 2GB

		// Minimal notices for quick results
		gtfsvalidator.WithMaxNoticesPerType(10),

		gtfsvalidator.WithProgressCallback(func(info gtfsvalidator.ProgressInfo) {
			if int(info.PercentComplete)%20 == 0 {
				fmt.Printf("⚡ Fast mode: %.0f%% complete\n", info.PercentComplete)
			}
		}),
	)

	return validator, "Fast: 8 workers, 2GB limit, minimal notices"
}

// Comprehensive configuration for thorough analysis
func createComprehensiveValidator() (gtfsvalidator.Validator, string) {
	validator := gtfsvalidator.New(
		// Moderate parallelism to balance speed and memory
		gtfsvalidator.WithParallelWorkers(4),

		// Higher memory limit for comprehensive analysis
		gtfsvalidator.WithMaxMemory(3*1024*1024*1024), // 3GB

		// More detailed reporting
		gtfsvalidator.WithMaxNoticesPerType(100),

		gtfsvalidator.WithProgressCallback(func(info gtfsvalidator.ProgressInfo) {
			if int(info.PercentComplete)%10 == 0 {
				fmt.Printf("🔍 Comprehensive: %.0f%% - %s\n",
					info.PercentComplete,
					formatValidatorName(info.CurrentValidator))
			}
		}),
	)

	return validator, "Comprehensive: All validators, 4 workers, 3GB limit, detailed reporting"
}

// MemoryMonitor tracks memory usage during validation
type MemoryMonitor struct {
	mutex     sync.Mutex
	samples   []MemorySample
	maxMemory uint64
	startTime time.Time
}

type MemorySample struct {
	timestamp time.Time
	alloc     uint64
	sys       uint64
	numGC     uint32
}

func (m *MemoryMonitor) Start(ctx context.Context) {
	m.mutex.Lock()
	m.startTime = time.Now()
	m.mutex.Unlock()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.takeSample()
		}
	}
}

func (m *MemoryMonitor) takeSample() {
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)

	m.mutex.Lock()
	defer m.mutex.Unlock()

	sample := MemorySample{
		timestamp: time.Now(),
		alloc:     stats.Alloc,
		sys:       stats.Sys,
		numGC:     stats.NumGC,
	}

	m.samples = append(m.samples, sample)

	if stats.Alloc > m.maxMemory {
		m.maxMemory = stats.Alloc
	}
}

func (m *MemoryMonitor) PrintSummary() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if len(m.samples) == 0 {
		return
	}

	fmt.Println()
	fmt.Println("📊 Memory Usage Summary")
	fmt.Println(strings.Repeat("-", 40))

	startSample := m.samples[0]
	endSample := m.samples[len(m.samples)-1]

	fmt.Printf("💾 Peak memory usage: %s\n", formatBytes(m.maxMemory))
	fmt.Printf("📈 Memory growth: %s → %s\n",
		formatBytes(startSample.alloc),
		formatBytes(endSample.alloc))
	fmt.Printf("🗑️  Garbage collections: %d\n",
		endSample.numGC-startSample.numGC)
	fmt.Printf("⏱️  Monitoring duration: %.1f seconds\n",
		endSample.timestamp.Sub(m.startTime).Seconds())
}

func showMemoryUsage(label string) {
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)

	fmt.Printf("💾 %s - Memory: %s allocated, %s system, %d GC cycles\n",
		label,
		formatBytes(stats.Alloc),
		formatBytes(stats.Sys),
		stats.NumGC)
}

func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func printValidationSummary(report *gtfsvalidator.ValidationReport, elapsed time.Duration) {
	fmt.Println()
	fmt.Println("🎯 Validation Results")
	fmt.Println(strings.Repeat("-", 40))

	// Feed statistics
	info := report.Summary.FeedInfo
	fmt.Printf("📁 Feed Statistics:\n")
	fmt.Printf("  Agencies: %s\n", formatNumber(info.AgencyCount))
	fmt.Printf("  Routes: %s\n", formatNumber(info.RouteCount))
	fmt.Printf("  Trips: %s\n", formatNumber(info.TripCount))
	fmt.Printf("  Stops: %s\n", formatNumber(info.StopCount))
	fmt.Printf("  Stop Times: %s\n", formatNumber(info.StopTimeCount))

	// Validation results
	counts := report.Summary.Counts
	fmt.Printf("\n🔍 Validation Results:\n")
	fmt.Printf("  Errors: %s\n", formatNumber(counts.Errors))
	fmt.Printf("  Warnings: %s\n", formatNumber(counts.Warnings))
	fmt.Printf("  Info: %s\n", formatNumber(counts.Infos))
	fmt.Printf("  Total: %s\n", formatNumber(counts.Total))

	// Performance metrics
	fmt.Printf("\n⚡ Performance:\n")
	fmt.Printf("  Validation time: %.2f seconds\n", elapsed.Seconds())
	if info.StopTimeCount > 0 {
		throughput := float64(info.StopTimeCount) / elapsed.Seconds()
		fmt.Printf("  Throughput: %.0f stop times/second\n", throughput)
	}

	// Final status
	fmt.Println()
	switch {
	case report.HasErrors():
		fmt.Printf("❌ VALIDATION FAILED - %s errors found\n", formatNumber(counts.Errors))
	case report.HasWarnings():
		fmt.Printf("⚠️  VALIDATION PASSED - %s warnings\n", formatNumber(counts.Warnings))
	default:
		fmt.Println("✅ VALIDATION PASSED - Feed is valid!")
	}
}

func printMemoryOptimizationTips(mode string, report *gtfsvalidator.ValidationReport) {
	fmt.Println()
	fmt.Println("💡 Memory Optimization Tips")
	fmt.Println(strings.Repeat("-", 40))

	feedSize := report.Summary.FeedInfo.StopTimeCount

	if feedSize > 1000000 { // >1M stop times
		fmt.Println("🔧 For very large feeds:")
		fmt.Println("  • Limit parallel workers to 2-4 to reduce memory usage")
		fmt.Println("  • Set MaxNoticesPerType to 10-25 to limit memory growth")
		fmt.Println("  • Set MaxMemory to appropriate limit (1-2GB)")
	}

	if mode != memoryOptimizedMode {
		fmt.Println("🔄 Try memory-optimized mode:")
		fmt.Printf("  go run examples/large-feeds/main.go %s "+memoryOptimizedMode+"\n", os.Args[1])
	}

	if report.Summary.Counts.Total > 10000 {
		fmt.Println("📊 For feeds with many validation issues:")
		fmt.Println("  • Reduce MaxNoticesPerType to limit memory usage")
		fmt.Println("  • Use streaming validation for real-time feedback")
		fmt.Println("  • Fix major issues first, then re-validate")
	}

	fmt.Println("\n📚 Additional optimizations:")
	fmt.Println("  • Use SSD storage for better I/O performance")
	fmt.Println("  • Increase available system RAM")
	fmt.Println("  • Close other memory-intensive applications")
	fmt.Println("  • Use streaming validation API for real-time progress")
}

func formatNumber(n int) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}
	if n < 1000000 {
		return fmt.Sprintf("%.1fK", float64(n)/1000)
	}
	return fmt.Sprintf("%.1fM", float64(n)/1000000)
}

func formatValidatorName(validatorName string) string {
	name := validatorName
	if len(name) > 30 {
		parts := strings.Split(name, ".")
		if len(parts) > 0 {
			name = parts[len(parts)-1]
		}
	}
	return name
}
