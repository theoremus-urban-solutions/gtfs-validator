// Configuration validation example shows how to properly configure
// the GTFS validator with validation and error handling
package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator"
)

func main() {
	fmt.Println("GTFS Validator Configuration Examples")
	fmt.Println("=====================================")
	fmt.Println()

	// Example 1: Default configuration (always valid)
	example1()

	// Example 2: Valid custom configuration
	example2()

	// Example 3: Invalid configurations (will be sanitized)
	example3()

	// Example 4: Configuration for different use cases
	example4()

	// Example 5: Interactive configuration
	if len(os.Args) > 1 {
		example5(os.Args[1])
	}
}

// Example 1: Default configuration
func example1() {
	fmt.Println("Example 1: Default Configuration")
	fmt.Println("--------------------------------")

	// Create validator with default settings
	validator := gtfsvalidator.New()

	// Show the configuration (we can't directly access it, but we can infer from behavior)
	fmt.Println("✅ Default validator created successfully")
	fmt.Println("  - Country Code: US (default)")
	fmt.Println("  - Validation Mode: Default")
	fmt.Println("  - Parallel Workers: 4 (default)")
	fmt.Println("  - Max Notices Per Type: 100 (default)")
	fmt.Println()

	_ = validator // Use the validator to avoid unused variable warning
}

// Example 2: Valid custom configuration
func example2() {
	fmt.Println("Example 2: Valid Custom Configuration")
	fmt.Println("------------------------------------")

	// Create validator with valid custom settings
	validator := gtfsvalidator.New(
		gtfsvalidator.WithCountryCode("GB"), // Valid 2-letter code
		gtfsvalidator.WithValidationMode(gtfsvalidator.ValidationModePerformance),
		gtfsvalidator.WithParallelWorkers(8),                       // Valid range
		gtfsvalidator.WithMaxMemory(512*1024*1024),                 // 512MB
		gtfsvalidator.WithMaxNoticesPerType(50),                    // Reasonable limit
		gtfsvalidator.WithCurrentDate(time.Now().AddDate(0, 6, 0)), // 6 months in future
	)

	fmt.Println("✅ Custom validator created successfully")
	fmt.Println("  - Country Code: GB")
	fmt.Println("  - Validation Mode: Performance")
	fmt.Println("  - Parallel Workers: 8")
	fmt.Println("  - Max Memory: 512MB")
	fmt.Println("  - Max Notices Per Type: 50")
	fmt.Println("  - Current Date: 6 months in future")
	fmt.Println()

	_ = validator
}

// Example 3: Invalid configurations (will be sanitized)
func example3() {
	fmt.Println("Example 3: Invalid Configurations (Sanitized)")
	fmt.Println("---------------------------------------------")

	// These configurations are invalid but will be sanitized to valid values
	validator := gtfsvalidator.New(
		gtfsvalidator.WithCountryCode("USA"),       // Invalid: too long
		gtfsvalidator.WithParallelWorkers(-5),      // Invalid: negative
		gtfsvalidator.WithMaxMemory(-1000),         // Invalid: negative
		gtfsvalidator.WithMaxNoticesPerType(50000), // Invalid: too high
	)

	fmt.Println("⚠️  Invalid configuration provided but validator created successfully")
	fmt.Println("   (Configuration was automatically sanitized to valid values)")
	fmt.Println("  - Country Code: 'USA' → sanitized to 'US'")
	fmt.Println("  - Parallel Workers: -5 → sanitized to 1")
	fmt.Println("  - Max Memory: -1000 → sanitized to 0 (no limit)")
	fmt.Println("  - Max Notices Per Type: 50000 → sanitized to 10000")
	fmt.Println()

	_ = validator
}

// Example 4: Configuration for different use cases
func example4() {
	fmt.Println("Example 4: Use Case Specific Configurations")
	fmt.Println("------------------------------------------")

	// Fast validation for CI/CD
	fmt.Println("🚀 Fast Validation (CI/CD):")
	fastValidator := gtfsvalidator.New(
		gtfsvalidator.WithValidationMode(gtfsvalidator.ValidationModePerformance),
		gtfsvalidator.WithParallelWorkers(8),
		gtfsvalidator.WithMaxNoticesPerType(10), // Limit notices for quick feedback
	)
	fmt.Println("  - Optimized for speed")
	fmt.Println("  - Performance mode only runs essential validators")
	fmt.Println("  - High parallelism for faster execution")
	fmt.Println("  - Limited notices to reduce output")
	fmt.Println()

	// Thorough validation for production
	fmt.Println("🔍 Thorough Validation (Production):")
	thoroughValidator := gtfsvalidator.New(
		gtfsvalidator.WithValidationMode(gtfsvalidator.ValidationModeComprehensive),
		gtfsvalidator.WithParallelWorkers(4),          // Moderate parallelism
		gtfsvalidator.WithMaxNoticesPerType(1000),     // More detailed reporting
		gtfsvalidator.WithMaxMemory(2*1024*1024*1024), // 2GB limit
	)
	fmt.Println("  - Comprehensive validation mode")
	fmt.Println("  - Includes all validators including expensive ones")
	fmt.Println("  - Detailed notice reporting")
	fmt.Println("  - Memory-bounded for stability")
	fmt.Println()

	// Memory-constrained validation
	fmt.Println("💾 Memory-Constrained Validation:")
	constrainedValidator := gtfsvalidator.New(
		gtfsvalidator.WithValidationMode(gtfsvalidator.ValidationModeDefault),
		gtfsvalidator.WithParallelWorkers(2),       // Fewer workers = less memory
		gtfsvalidator.WithMaxMemory(256*1024*1024), // 256MB limit
		gtfsvalidator.WithMaxNoticesPerType(50),    // Limit memory usage
	)
	fmt.Println("  - Conservative memory usage")
	fmt.Println("  - Reduced parallelism")
	fmt.Println("  - Notice limits to prevent memory growth")
	fmt.Println()

	_, _, _ = fastValidator, thoroughValidator, constrainedValidator
}

// Example 5: Interactive configuration
func example5(gtfsFile string) {
	fmt.Println("Example 5: Interactive Configuration & Validation")
	fmt.Println("------------------------------------------------")

	// Create validator with progress callback
	validator := gtfsvalidator.New(
		gtfsvalidator.WithValidationMode(gtfsvalidator.ValidationModeDefault),
		gtfsvalidator.WithProgressCallback(func(info gtfsvalidator.ProgressInfo) {
			// Only show progress every 20%
			if int(info.PercentComplete)%20 == 0 {
				fmt.Printf("📊 Progress: %.0f%% - %s\n",
					info.PercentComplete,
					formatValidatorName(info.CurrentValidator))
			}
		}),
		gtfsvalidator.WithCountryCode("US"),
		gtfsvalidator.WithParallelWorkers(4),
	)

	fmt.Printf("🎯 Validating: %s\n", gtfsFile)
	fmt.Println()

	startTime := time.Now()
	report, err := validator.ValidateFile(gtfsFile)
	elapsed := time.Since(startTime)

	if err != nil {
		log.Printf("❌ Validation failed: %v", err)
		return
	}

	fmt.Printf("\n⏱️  Validation completed in %.2f seconds\n", elapsed.Seconds())
	fmt.Printf("📊 Results: %d errors, %d warnings, %d infos\n",
		report.Summary.Counts.Errors,
		report.Summary.Counts.Warnings,
		report.Summary.Counts.Infos)

	if report.HasErrors() {
		fmt.Printf("❌ Validation FAILED with %d errors\n", report.ErrorCount())
	} else if report.HasWarnings() {
		fmt.Printf("⚠️  Validation PASSED with %d warnings\n", report.WarningCount())
	} else {
		fmt.Println("✅ Validation PASSED - Feed is perfect!")
	}
	fmt.Println()
}

// Helper functions

func formatValidatorName(validatorName string) string {
	// Remove package prefixes and make more readable
	name := validatorName
	if len(name) > 40 {
		// Extract just the struct name
		parts := strings.Split(name, ".")
		if len(parts) > 0 {
			name = parts[len(parts)-1]
		}
	}
	return name
}
