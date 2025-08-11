// Advanced example showing progress tracking, cancellation, and custom configuration
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: advanced <gtfs-file>")
		os.Exit(1)
	}

	gtfsFile := os.Args[1]

	// Create a context that can be cancelled with Ctrl+C
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	go func() {
		<-sigChan
		fmt.Println("\nCancelling validation...")
		cancel()
	}()

	// Create progress bar
	progressBar := NewProgressBar()

	// Create validator with advanced configuration
	validator := gtfsvalidator.New(
		// Set validation mode for performance
		gtfsvalidator.WithValidationMode(gtfsvalidator.ValidationModePerformance),

		// Limit notices to prevent huge reports
		gtfsvalidator.WithMaxNoticesPerType(50),

		// Set country code for phone number validation
		gtfsvalidator.WithCountryCode("US"),

		// Add progress callback
		gtfsvalidator.WithProgressCallback(func(info gtfsvalidator.ProgressInfo) {
			progressBar.Update(info.PercentComplete, info.CurrentValidator)
		}),

		// Limit memory usage to 512MB
		gtfsvalidator.WithMaxMemory(512*1024*1024),

		// Use 8 parallel workers for faster validation
		gtfsvalidator.WithParallelWorkers(8),
	)

	fmt.Printf("Validating GTFS feed: %s\n", gtfsFile)
	fmt.Println("Mode: Performance (fast validation)")
	fmt.Println("Press Ctrl+C to cancel")

	startTime := time.Now()

	// Validate with context
	report, err := validator.ValidateFileWithContext(ctx, gtfsFile)
	if err != nil {
		if err == context.Canceled {
			fmt.Println("\nValidation cancelled by user")
			os.Exit(1)
		}
		log.Fatalf("Failed to validate: %v", err)
	}

	progressBar.Complete()
	elapsed := time.Since(startTime)

	// Display detailed results
	fmt.Printf("\n\nValidation completed in %.2f seconds\n", elapsed.Seconds())
	fmt.Println("=" + string(make([]byte, 50)) + "=")

	// Feed info
	fmt.Printf("\nFeed Information:\n")
	fmt.Printf("  Agencies: %d\n", report.Summary.FeedInfo.AgencyCount)
	fmt.Printf("  Routes: %d\n", report.Summary.FeedInfo.RouteCount)
	fmt.Printf("  Trips: %d\n", report.Summary.FeedInfo.TripCount)
	fmt.Printf("  Stops: %d\n", report.Summary.FeedInfo.StopCount)
	fmt.Printf("  Stop Times: %d\n", report.Summary.FeedInfo.StopTimeCount)

	if report.Summary.FeedInfo.ServiceDateFrom != "" {
		fmt.Printf("  Service Period: %s to %s\n",
			report.Summary.FeedInfo.ServiceDateFrom,
			report.Summary.FeedInfo.ServiceDateTo)
	}

	// Notice summary
	fmt.Printf("\nValidation Results:\n")
	fmt.Printf("  Total Notices: %d\n", report.Summary.Counts.Total)
	fmt.Printf("    - Errors: %d\n", report.Summary.Counts.Errors)
	fmt.Printf("    - Warnings: %d\n", report.Summary.Counts.Warnings)
	fmt.Printf("    - Info: %d\n", report.Summary.Counts.Infos)

	// Group notices by severity
	var errors, warnings, infos []gtfsvalidator.NoticeGroup
	for _, notice := range report.Notices {
		switch notice.Severity {
		case "ERROR":
			errors = append(errors, notice)
		case "WARNING":
			warnings = append(warnings, notice)
		case "INFO":
			infos = append(infos, notice)
		}
	}

	// Display top issues
	if len(errors) > 0 {
		fmt.Printf("\nTop Errors:\n")
		displayTopNotices(errors, 5)
	}

	if len(warnings) > 0 {
		fmt.Printf("\nTop Warnings:\n")
		displayTopNotices(warnings, 5)
	}

	// Exit code based on errors
	if report.HasErrors() {
		fmt.Printf("\n❌ Validation FAILED with %d errors\n", report.ErrorCount())
		os.Exit(1)
	} else if report.HasWarnings() {
		fmt.Printf("\n⚠️  Validation completed with %d warnings\n", report.WarningCount())
	} else {
		fmt.Println("\n✅ Validation PASSED - Feed is valid!")
	}
}

func displayTopNotices(notices []gtfsvalidator.NoticeGroup, limit int) {
	for i, notice := range notices {
		if i >= limit {
			break
		}
		fmt.Printf("  %d. %s (%d instances)\n", i+1, notice.Code, notice.TotalNotices)

		// Show sample context for the first instance
		if len(notice.SampleNotices) > 0 {
			sample := notice.SampleNotices[0]
			fmt.Printf("     Example: ")

			// Display relevant context fields
			if filename, ok := sample["filename"].(string); ok {
				fmt.Printf("file=%s ", filename)
			}
			if row, ok := sample["csvRowNumber"].(float64); ok {
				fmt.Printf("row=%d ", int(row))
			}
			if fieldName, ok := sample["fieldName"].(string); ok {
				fmt.Printf("field=%s ", fieldName)
			}
			fmt.Println()
		}
	}

	if len(notices) > limit {
		fmt.Printf("  ... and %d more\n", len(notices)-limit)
	}
}

// ProgressBar provides a simple progress indicator
type ProgressBar struct {
	lastPercent int
}

func NewProgressBar() *ProgressBar {
	return &ProgressBar{lastPercent: -1}
}

func (p *ProgressBar) Update(percent float64, status string) {
	currentPercent := int(percent)
	if currentPercent == p.lastPercent {
		return // Don't update if percentage hasn't changed
	}
	p.lastPercent = currentPercent

	// Create progress bar
	barWidth := 50
	filled := int(float64(barWidth) * percent / 100)
	bar := string(make([]byte, filled))
	for i := range bar {
		bar = bar[:i] + "=" + bar[i+1:]
	}

	// Truncate status if too long
	maxStatusLen := 40
	if len(status) > maxStatusLen {
		status = status[:maxStatusLen-3] + "..."
	}

	fmt.Printf("\r[%-*s] %3d%% %s", barWidth, bar, currentPercent, status)
}

func (p *ProgressBar) Complete() {
	fmt.Println()
}
