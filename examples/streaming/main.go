// Streaming validation example shows how to use the streaming validation API
// to receive real-time feedback during GTFS validation
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: streaming <gtfs-file>")
		os.Exit(1)
	}

	gtfsFile := os.Args[1]

	// Create context that can be cancelled with Ctrl+C
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

	// Track streaming statistics
	stats := &StreamingStats{}

	// Create validator with streaming configuration
	validator := gtfsvalidator.New(
		// Use default mode for comprehensive validation
		gtfsvalidator.WithValidationMode(gtfsvalidator.ValidationModeDefault),

		// Limit notices to prevent overwhelming output
		gtfsvalidator.WithMaxNoticesPerType(10),

		// Use 4 workers for parallel validation
		gtfsvalidator.WithParallelWorkers(4),

		// Add progress callback
		gtfsvalidator.WithProgressCallback(func(info gtfsvalidator.ProgressInfo) {
			stats.UpdateProgress(info)
		}),
	)

	fmt.Printf("Starting streaming validation of: %s\n", gtfsFile)
	fmt.Println("Press Ctrl+C to cancel")
	fmt.Println()

	startTime := time.Now()

	// Validate with streaming
	report, err := validator.ValidateFileStreamWithContext(ctx, gtfsFile, func(notice gtfsvalidator.NoticeGroup) {
		stats.ProcessNotice(notice)
	})

	if err != nil {
		if err == context.Canceled {
			fmt.Println("\nValidation was cancelled by user")
			stats.PrintFinalSummary(time.Since(startTime), nil)
			cancel()
			os.Exit(1) //nolint:gocritic // cancel() already called above
		}
		log.Fatalf("Validation failed: %v", err)
	}

	elapsed := time.Since(startTime)
	stats.PrintFinalSummary(elapsed, report)
}

// StreamingStats tracks statistics during streaming validation
type StreamingStats struct {
	mutex          sync.Mutex
	totalNotices   int
	errorCount     int
	warningCount   int
	infoCount      int
	lastProgress   float64
	lastValidator  string
	noticeCodes    map[string]int
	severityCounts map[string]int
	startTime      time.Time
}

func (s *StreamingStats) ProcessNotice(notice gtfsvalidator.NoticeGroup) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.noticeCodes == nil {
		s.noticeCodes = make(map[string]int)
		s.severityCounts = make(map[string]int)
		s.startTime = time.Now()
	}

	// Update counts
	s.totalNotices++
	s.noticeCodes[notice.Code] += notice.TotalNotices
	s.severityCounts[notice.Severity] += notice.TotalNotices

	switch notice.Severity {
	case "ERROR":
		s.errorCount += notice.TotalNotices
	case "WARNING":
		s.warningCount += notice.TotalNotices
	case "INFO":
		s.infoCount += notice.TotalNotices
	}

	// Print notice information
	severity := notice.Severity
	switch severity {
	case "ERROR":
		severity = "🔴 ERROR"
	case "WARNING":
		severity = "🟡 WARNING"
	case "INFO":
		severity = "🔵 INFO"
	}

	fmt.Printf("  %s %s: %d instances\n", severity, notice.Code, notice.TotalNotices)

	// Print sample context if available
	if len(notice.SampleNotices) > 0 {
		sample := notice.SampleNotices[0]
		if filename, ok := sample["filename"].(string); ok {
			fmt.Printf("    📁 File: %s", filename)
			if row, ok := sample["csvRowNumber"].(float64); ok {
				fmt.Printf(", Row: %d", int(row))
			}
			if field, ok := sample["fieldName"].(string); ok {
				fmt.Printf(", Field: %s", field)
			}
			fmt.Println()
		}
	}
}

func (s *StreamingStats) UpdateProgress(info gtfsvalidator.ProgressInfo) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.lastProgress = info.PercentComplete
	s.lastValidator = info.CurrentValidator

	// Print progress updates every 10%
	if int(info.PercentComplete)%10 == 0 && int(info.PercentComplete) != int(s.lastProgress) {
		fmt.Printf("\n📊 Progress: %.0f%% - Running %s (%d/%d validators)\n",
			info.PercentComplete,
			formatValidatorName(info.CurrentValidator),
			info.CompletedValidators,
			info.TotalValidators)
	}
}

func (s *StreamingStats) PrintFinalSummary(elapsed time.Duration, report *gtfsvalidator.ValidationReport) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	fmt.Printf("\n" + strings.Repeat("=", 60) + "\n")
	fmt.Printf("🏁 VALIDATION COMPLETE\n")
	fmt.Printf(strings.Repeat("=", 60) + "\n")

	if report != nil {
		fmt.Printf("⏱️  Validation time: %.2f seconds\n", elapsed.Seconds())
		fmt.Printf("📂 Feed path: %s\n", report.Summary.FeedInfo.FeedPath)

		// Feed statistics
		fmt.Println("\n📈 Feed Statistics:")
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
	}

	// Validation results
	fmt.Println("\n🔍 Validation Results:")
	fmt.Printf("  Total Notice Types: %d\n", s.totalNotices)
	fmt.Printf("  🔴 Errors: %d\n", s.errorCount)
	fmt.Printf("  🟡 Warnings: %d\n", s.warningCount)
	fmt.Printf("  🔵 Info: %d\n", s.infoCount)
	fmt.Printf("  📊 Total Issues: %d\n", s.errorCount+s.warningCount+s.infoCount)

	// Top issues
	if len(s.noticeCodes) > 0 {
		fmt.Println("\n🔝 Top Issues:")
		type noticeCount struct {
			code  string
			count int
		}

		var notices []noticeCount
		for code, count := range s.noticeCodes {
			notices = append(notices, noticeCount{code, count})
		}

		// Simple sort by count (descending)
		for i := 0; i < len(notices); i++ {
			for j := i + 1; j < len(notices); j++ {
				if notices[j].count > notices[i].count {
					notices[i], notices[j] = notices[j], notices[i]
				}
			}
		}

		// Show top 5
		limit := 5
		if len(notices) < limit {
			limit = len(notices)
		}

		for i := 0; i < limit; i++ {
			fmt.Printf("  %d. %s: %d instances\n", i+1, notices[i].code, notices[i].count)
		}

		if len(notices) > limit {
			fmt.Printf("  ... and %d more issue types\n", len(notices)-limit)
		}
	}

	// Final verdict
	fmt.Println()
	if s.errorCount == 0 {
		if s.warningCount == 0 {
			fmt.Println("✅ VALIDATION PASSED - Feed is valid!")
		} else {
			fmt.Printf("⚠️  VALIDATION PASSED with %d warnings\n", s.warningCount)
		}
	} else {
		fmt.Printf("❌ VALIDATION FAILED - %d errors found\n", s.errorCount)
	}
}

// formatValidatorName makes validator names more readable
func formatValidatorName(validatorName string) string {
	// Remove package prefixes and make more readable
	name := validatorName
	if len(name) > 50 {
		name = name[len(name)-47:] + "..."
	}
	return name
}
