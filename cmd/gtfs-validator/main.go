// Command-line interface for the GTFS validator library
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"time"
	
	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator"
)

const version = "1.0.0"

func main() {
	// Command line flags
	var (
		inputPath    = flag.String("input", "", "Path to GTFS feed (ZIP file or directory)")
		outputFormat = flag.String("format", "console", "Output format: console, json, summary")
		outputFile   = flag.String("output", "", "Output file path (default: stdout)")
		countryCode  = flag.String("country", "US", "Country code for validation (e.g., US, GB, FR)")
		maxMemory    = flag.Int64("memory", 0, "Maximum memory usage in MB (0 = no limit)")
		workers      = flag.Int("workers", 4, "Number of parallel workers")
		mode         = flag.String("mode", "default", "Validation mode: performance, default, comprehensive")
		maxNotices   = flag.Int("max-notices", 100, "Maximum notices per type (0 = no limit)")
		timeout      = flag.Duration("timeout", 5*time.Minute, "Validation timeout")
		showProgress = flag.Bool("progress", false, "Show progress bar")
		help         = flag.Bool("help", false, "Show help message")
		showVersion  = flag.Bool("version", false, "Show version information")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "GTFS Validator CLI - A comprehensive GTFS feed validator\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [options] -input <path>\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s -input feed.zip\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -input ./gtfs-feed -format json -output report.json\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -input feed.zip -mode performance\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -input feed.zip -progress\n", os.Args[0])
	}

	flag.Parse()

	if *help {
		flag.Usage()
		return
	}

	if *showVersion {
		fmt.Printf("GTFS Validator CLI v%s\n", version)
		fmt.Println("A comprehensive GTFS feed validator written in Go")
		return
	}

	if *inputPath == "" {
		fmt.Fprintf(os.Stderr, "Error: -input flag is required\n\n")
		flag.Usage()
		os.Exit(1)
	}

	// Validate input
	if err := validateInput(*inputPath, *mode, *outputFormat); err != nil {
		fmt.Fprintf(os.Stderr, "❌ %v\n", err)
		os.Exit(1)
	}

	// Create context with timeout and cancellation
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	// Handle interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	go func() {
		<-sigChan
		fmt.Fprintf(os.Stderr, "\n⚠️  Cancelling validation...\n")
		cancel()
	}()

	// Configure validator options
	opts := []gtfsvalidator.Option{
		gtfsvalidator.WithCountryCode(*countryCode),
		gtfsvalidator.WithMaxMemory(*maxMemory * 1024 * 1024), // Convert MB to bytes
		gtfsvalidator.WithParallelWorkers(*workers),
		gtfsvalidator.WithMaxNoticesPerType(*maxNotices),
	}

	// Set validation mode
	switch *mode {
	case "performance":
		opts = append(opts, gtfsvalidator.WithValidationMode(gtfsvalidator.ValidationModePerformance))
	case "comprehensive":
		opts = append(opts, gtfsvalidator.WithValidationMode(gtfsvalidator.ValidationModeComprehensive))
	default:
		opts = append(opts, gtfsvalidator.WithValidationMode(gtfsvalidator.ValidationModeDefault))
	}

	// Add progress callback if requested
	if *showProgress {
		progressBar := NewProgressBar()
		opts = append(opts, gtfsvalidator.WithProgressCallback(func(info gtfsvalidator.ProgressInfo) {
			progressBar.Update(info.PercentComplete, info.CurrentValidator)
		}))
	}

	// Create validator
	validator := gtfsvalidator.New(opts...)

	// Show startup message
	fmt.Fprintf(os.Stderr, "🚀 Starting GTFS validation...\n")
	fmt.Fprintf(os.Stderr, "   Feed: %s\n", filepath.Base(*inputPath))
	fmt.Fprintf(os.Stderr, "   Mode: %s\n", *mode)
	if *maxNotices > 0 {
		fmt.Fprintf(os.Stderr, "   Notice limit: %d per type\n", *maxNotices)
	}
	fmt.Fprintf(os.Stderr, "\n")

	// Perform validation
	startTime := time.Now()
	report, err := validator.ValidateFileWithContext(ctx, *inputPath)
	elapsed := time.Since(startTime)

	if err != nil {
		if err == context.Canceled {
			fmt.Fprintf(os.Stderr, "⚠️  Validation cancelled by user\n")
			os.Exit(1)
		} else if err == context.DeadlineExceeded {
			fmt.Fprintf(os.Stderr, "⏰ Validation timed out after %v\n", *timeout)
			os.Exit(1)
		} else {
			fmt.Fprintf(os.Stderr, "❌ Validation Error: %v\n", err)
			os.Exit(1)
		}
	}

	fmt.Fprintf(os.Stderr, "✅ Validation completed in %.2fs\n\n", elapsed.Seconds())

	// Handle output
	output := os.Stdout
	if *outputFile != "" {
		file, err := os.Create(*outputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Output Error: Failed to create output file '%s': %v\n", *outputFile, err)
			os.Exit(1)
		}
		defer file.Close()
		output = file
		fmt.Fprintf(os.Stderr, "📄 Writing output to: %s\n", *outputFile)
	}

	// Generate output based on format
	switch *outputFormat {
	case "json":
		if err := json.NewEncoder(output).Encode(report); err != nil {
			fmt.Fprintf(os.Stderr, "❌ JSON Error: Failed to encode report: %v\n", err)
			os.Exit(1)
		}
	case "summary":
		outputSummary(output, report, *inputPath)
	case "console":
		outputConsole(output, report, *inputPath)
	default:
		fmt.Fprintf(os.Stderr, "❌ Format Error: Unknown output format '%s'\n", *outputFormat)
		fmt.Fprintf(os.Stderr, "   Valid formats: console, json, summary\n")
		os.Exit(1)
	}

	// Final status and exit
	if report.HasErrors() {
		fmt.Fprintf(os.Stderr, "💀 Validation FAILED: %d errors found\n", report.ErrorCount())
		os.Exit(1)
	} else if report.HasWarnings() {
		fmt.Fprintf(os.Stderr, "⚠️  Validation completed with %d warnings\n", report.WarningCount())
	} else {
		fmt.Fprintf(os.Stderr, "🎉 Validation PASSED: Feed is valid!\n")
	}
}

func validateInput(inputPath, mode, format string) error {
	// Check if input exists
	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		return fmt.Errorf("Input Error: Path does not exist: '%s'", inputPath)
	}

	// Validate mode
	validModes := []string{"performance", "default", "comprehensive"}
	if !contains(validModes, mode) {
		return fmt.Errorf("Invalid validation mode: '%s'. Valid modes: %s", mode, strings.Join(validModes, ", "))
	}

	// Validate format
	validFormats := []string{"console", "json", "summary"}
	if !contains(validFormats, format) {
		return fmt.Errorf("Invalid output format: '%s'. Valid formats: %s", format, strings.Join(validFormats, ", "))
	}

	return nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func outputSummary(output *os.File, report *gtfsvalidator.ValidationReport, inputPath string) {
	fmt.Fprintf(output, "GTFS Validation Summary\n")
	fmt.Fprintf(output, "======================\n\n")
	fmt.Fprintf(output, "Feed: %s\n", filepath.Base(inputPath))
	fmt.Fprintf(output, "Validation Time: %.2fs\n\n", report.Summary.ValidationTime)

	fmt.Fprintf(output, "Feed Statistics:\n")
	fmt.Fprintf(output, "  Agencies: %d\n", report.Summary.FeedInfo.AgencyCount)
	fmt.Fprintf(output, "  Routes: %d\n", report.Summary.FeedInfo.RouteCount)
	fmt.Fprintf(output, "  Trips: %d\n", report.Summary.FeedInfo.TripCount)
	fmt.Fprintf(output, "  Stops: %d\n", report.Summary.FeedInfo.StopCount)
	fmt.Fprintf(output, "  Stop Times: %d\n", report.Summary.FeedInfo.StopTimeCount)
	if report.Summary.FeedInfo.ServiceDateFrom != "" && report.Summary.FeedInfo.ServiceDateTo != "" {
		fmt.Fprintf(output, "  Service Period: %s to %s\n", report.Summary.FeedInfo.ServiceDateFrom, report.Summary.FeedInfo.ServiceDateTo)
	}

	fmt.Fprintf(output, "\nValidation Results:\n")
	fmt.Fprintf(output, "  Errors: %d\n", report.Summary.Counts.Errors)
	fmt.Fprintf(output, "  Warnings: %d\n", report.Summary.Counts.Warnings)
	fmt.Fprintf(output, "  Infos: %d\n", report.Summary.Counts.Infos)
	fmt.Fprintf(output, "  Total: %d\n", report.Summary.Counts.Total)

	if report.HasErrors() {
		fmt.Fprintf(output, "\n❌ Validation FAILED - Feed contains errors\n")
	} else if report.HasWarnings() {
		fmt.Fprintf(output, "\n⚠️  Validation completed with warnings\n")
	} else {
		fmt.Fprintf(output, "\n✅ Validation PASSED\n")
	}
}

func outputConsole(output *os.File, report *gtfsvalidator.ValidationReport, inputPath string) {
	outputSummary(output, report, inputPath)

	// Show first few notices if any
	if len(report.Notices) > 0 {
		fmt.Fprintf(output, "\nSample Notices:\n")
		fmt.Fprintf(output, "===============\n")

		errorCount := 0
		warningCount := 0

		for _, notice := range report.Notices {
			if errorCount >= 5 && warningCount >= 5 {
				break
			}

			if notice.Severity == "ERROR" && errorCount < 5 {
				fmt.Fprintf(output, "ERROR: %s (%d instances)\n", notice.Code, notice.TotalNotices)
				if len(notice.SampleNotices) > 0 {
					showNoticeContext(output, notice.SampleNotices[0])
				}
				errorCount++
			} else if notice.Severity == "WARNING" && warningCount < 5 {
				fmt.Fprintf(output, "WARNING: %s (%d instances)\n", notice.Code, notice.TotalNotices)
				if len(notice.SampleNotices) > 0 {
					showNoticeContext(output, notice.SampleNotices[0])
				}
				warningCount++
			}
		}

		if len(report.Notices) > 10 {
			fmt.Fprintf(output, "\n... and %d more notices (use -format json for full details)\n", len(report.Notices)-10)
		}
	}
}

func showNoticeContext(output *os.File, context map[string]interface{}) {
	details := []string{}
	
	if filename, ok := context["filename"].(string); ok {
		details = append(details, fmt.Sprintf("file=%s", filename))
	}
	if row, ok := context["csvRowNumber"].(float64); ok {
		details = append(details, fmt.Sprintf("row=%d", int(row)))
	}
	if field, ok := context["fieldName"].(string); ok {
		details = append(details, fmt.Sprintf("field=%s", field))
	}
	if routeId, ok := context["routeId"].(string); ok {
		details = append(details, fmt.Sprintf("route=%s", routeId))
	}
	
	if len(details) > 0 {
		fmt.Fprintf(output, "       (%s)\n", strings.Join(details, ", "))
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
		return
	}
	p.lastPercent = currentPercent

	// Create progress bar
	barWidth := 40
	filled := int(float64(barWidth) * percent / 100)
	bar := strings.Repeat("=", filled) + strings.Repeat(" ", barWidth-filled)

	// Truncate status if too long
	if len(status) > 30 {
		status = status[:27] + "..."
	}

	fmt.Fprintf(os.Stderr, "\r[%s] %3d%% %s", bar, currentPercent, status)
}