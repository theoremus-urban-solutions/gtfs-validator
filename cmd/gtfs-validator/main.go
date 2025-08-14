// Command-line interface for the GTFS validator library
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator"
)

// Version information - this will be set during build
var version = "dev"

var (
	// Global flags
	inputPath    string
	outputFormat string
	outputFile   string
	countryCode  string
	maxMemory    int64
	workers      int
	mode         string
	maxNotices   int
	timeout      time.Duration
	showProgress bool
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "gtfs-validator [flags]",
		Short: "A comprehensive GTFS feed validator",
		Long: `GTFS Validator CLI - A comprehensive GTFS feed validator written in Go.

This tool validates General Transit Feed Specification (GTFS) feeds for compliance
with the GTFS specification and transit industry best practices.

Features memory optimization with streaming CSV processing for large feeds,
structured logging, and comprehensive validation with 294+ validation rules.`,
		Example: `  gtfs-validator -i feed.zip
  gtfs-validator -i ./gtfs-feed -f json -o report.json
  gtfs-validator -i feed.zip -f html -o report.html
  gtfs-validator -i feed.zip -m performance
  gtfs-validator -i feed.zip --progress`,
		Version: version,
		RunE:    runValidation,
	}

	// Add flags
	rootCmd.Flags().StringVarP(&inputPath, "input", "i", "", "Path to GTFS feed (ZIP file or directory) [required]")
	rootCmd.Flags().StringVarP(&outputFormat, "format", "f", "console", "Output format: console, json, summary, html")
	rootCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file path (default: stdout)")
	rootCmd.Flags().StringVarP(&countryCode, "country", "c", "US", "Country code for validation (e.g., US, GB, FR)")
	rootCmd.Flags().Int64Var(&maxMemory, "memory", 0, "Maximum memory usage in MB (0 = no limit)")
	rootCmd.Flags().IntVarP(&workers, "workers", "w", 4, "Number of parallel workers")
	rootCmd.Flags().StringVarP(&mode, "mode", "m", "default", "Validation mode: performance, default, comprehensive")
	rootCmd.Flags().IntVar(&maxNotices, "max-notices", 100, "Maximum notices per type (0 = no limit)")
	rootCmd.Flags().DurationVarP(&timeout, "timeout", "t", 5*time.Minute, "Validation timeout")
	rootCmd.Flags().BoolVarP(&showProgress, "progress", "p", false, "Show progress bar")

	// Mark input as required
	if err := rootCmd.MarkFlagRequired("input"); err != nil {
		fmt.Fprintf(os.Stderr, "Error marking input flag as required: %v\n", err)
		os.Exit(1)
	}

	// Add subcommands
	rootCmd.AddCommand(newVersionCmd())
	rootCmd.AddCommand(newValidateCmd())

	// Execute
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("GTFS Validator CLI v%s\n", version)
			fmt.Println("A comprehensive GTFS feed validator written in Go")
			fmt.Println("https://github.com/theoremus-urban-solutions/gtfs-validator")
		},
	}
}

func newValidateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate [flags] <input>",
		Short: "Validate a GTFS feed",
		Long: `Validate a GTFS feed for compliance with the GTFS specification.

The input can be either a ZIP file containing the GTFS feed or a directory
with the GTFS files.

Uses memory-efficient streaming processing for large feeds and provides
comprehensive validation with 294+ validation rules.`,
		Example: `  gtfs-validator validate feed.zip
  gtfs-validator validate ./gtfs-directory --format json
  gtfs-validator validate feed.zip --format html --output report.html
  gtfs-validator validate feed.zip --mode performance --progress`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			inputPath = args[0]
			return runValidation(cmd, args)
		},
	}

	// Add the same flags as root command
	cmd.Flags().StringVarP(&outputFormat, "format", "f", "console", "Output format: console, json, summary, html")
	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file path (default: stdout)")
	cmd.Flags().StringVarP(&countryCode, "country", "c", "US", "Country code for validation (e.g., US, GB, FR)")
	cmd.Flags().Int64Var(&maxMemory, "memory", 0, "Maximum memory usage in MB (0 = no limit)")
	cmd.Flags().IntVarP(&workers, "workers", "w", 4, "Number of parallel workers")
	cmd.Flags().StringVarP(&mode, "mode", "m", "default", "Validation mode: performance, default, comprehensive")
	cmd.Flags().IntVar(&maxNotices, "max-notices", 100, "Maximum notices per type (0 = no limit)")
	cmd.Flags().DurationVarP(&timeout, "timeout", "t", 5*time.Minute, "Validation timeout")
	cmd.Flags().BoolVarP(&showProgress, "progress", "p", false, "Show progress bar")

	return cmd
}

func runValidation(cmd *cobra.Command, args []string) error {
	// Validate input
	if err := validateInput(inputPath, mode, outputFormat); err != nil {
		return fmt.Errorf("❌ %v", err)
	}

	// Create context with timeout and cancellation
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
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
		gtfsvalidator.WithCountryCode(countryCode),
		gtfsvalidator.WithMaxMemory(maxMemory * 1024 * 1024), // Convert MB to bytes
		gtfsvalidator.WithParallelWorkers(workers),
		gtfsvalidator.WithMaxNoticesPerType(maxNotices),
	}

	// Set validation mode
	switch mode {
	case "performance":
		opts = append(opts, gtfsvalidator.WithValidationMode(gtfsvalidator.ValidationModePerformance))
	case "comprehensive":
		opts = append(opts, gtfsvalidator.WithValidationMode(gtfsvalidator.ValidationModeComprehensive))
	default:
		opts = append(opts, gtfsvalidator.WithValidationMode(gtfsvalidator.ValidationModeDefault))
	}

	// Add progress callback if requested
	if showProgress {
		progressBar := NewProgressBar()
		opts = append(opts, gtfsvalidator.WithProgressCallback(func(info gtfsvalidator.ProgressInfo) {
			progressBar.Update(info.PercentComplete, info.CurrentValidator)
		}))
	}

	// Create validator
	validator := gtfsvalidator.New(opts...)

	// Show startup message
	fmt.Fprintf(os.Stderr, "🚀 Starting GTFS validation...\n")
	fmt.Fprintf(os.Stderr, "   Feed: %s\n", filepath.Base(inputPath))
	fmt.Fprintf(os.Stderr, "   Mode: %s\n", mode)
	if maxNotices > 0 {
		fmt.Fprintf(os.Stderr, "   Notice limit: %d per type\n", maxNotices)
	}
	fmt.Fprintf(os.Stderr, "\n")

	// Perform validation
	startTime := time.Now()
	report, err := validator.ValidateFileWithContext(ctx, inputPath)
	elapsed := time.Since(startTime)

	if err != nil {
		if err == context.Canceled {
			fmt.Fprintf(os.Stderr, "⚠️  Validation cancelled by user\n")
			os.Exit(1)
		} else if err == context.DeadlineExceeded {
			fmt.Fprintf(os.Stderr, "⏰ Validation timed out after %v\n", timeout)
			os.Exit(1)
		} else {
			return fmt.Errorf("❌ Validation Error: %v", err)
		}
	}

	fmt.Fprintf(os.Stderr, "✅ Validation completed in %.2fs\n\n", elapsed.Seconds())

	// Handle output
	output := os.Stdout
	if outputFile != "" {
		file, err := os.Create(outputFile)
		if err != nil {
			return fmt.Errorf("❌ Output Error: Failed to create output file '%s': %v", outputFile, err)
		}
		defer func() {
			if err := file.Close(); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: Failed to close output file: %v\n", err)
			}
		}()
		output = file
		fmt.Fprintf(os.Stderr, "📄 Writing output to: %s\n", outputFile)
	}

	// Generate output based on format
	switch outputFormat {
	case "json":
		if err := json.NewEncoder(output).Encode(report); err != nil {
			return fmt.Errorf("❌ JSON Error: Failed to encode report: %v", err)
		}
	case "summary":
		outputSummary(output, report, inputPath)
	case "console":
		outputConsole(output, report, inputPath)
	case "html":
		if err := outputHTML(output, report, inputPath); err != nil {
			return fmt.Errorf("❌ HTML Error: Failed to generate HTML report: %v", err)
		}
	default:
		return fmt.Errorf("❌ Format Error: Unknown output format '%s'. Valid formats: console, json, summary, html", outputFormat)
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

	return nil
}

func validateInput(inputPath, mode, format string) error {
	// Check if input exists
	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		return fmt.Errorf("input error: path does not exist: '%s'", inputPath)
	}

	// Validate mode
	validModes := []string{"performance", "default", "comprehensive"}
	if !contains(validModes, mode) {
		return fmt.Errorf("invalid validation mode: '%s'. valid modes: %s", mode, strings.Join(validModes, ", "))
	}

	// Validate format
	validFormats := []string{"console", "json", "summary", "html"}
	if !contains(validFormats, format) {
		return fmt.Errorf("invalid output format: '%s'. valid formats: %s", format, strings.Join(validFormats, ", "))
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
	// Helper function to write output with error checking
	write := func(format string, args ...interface{}) bool {
		if _, err := fmt.Fprintf(output, format, args...); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to write output: %v\n", err)
			return false
		}
		return true
	}

	if !write("GTFS Validation Summary\n") { return }
	if !write("======================\n\n") { return }
	if !write("Feed: %s\n", filepath.Base(inputPath)) { return }
	if !write("Validation Time: %.2fs\n\n", report.Summary.ValidationTime) { return }

	if !write("Feed Statistics:\n") { return }
	if !write("  Agencies: %d\n", report.Summary.FeedInfo.AgencyCount) { return }
	if !write("  Routes: %d\n", report.Summary.FeedInfo.RouteCount) { return }
	if !write("  Trips: %d\n", report.Summary.FeedInfo.TripCount) { return }
	if !write("  Stops: %d\n", report.Summary.FeedInfo.StopCount) { return }
	if !write("  Stop Times: %d\n", report.Summary.FeedInfo.StopTimeCount) { return }
	if report.Summary.FeedInfo.ServiceDateFrom != "" && report.Summary.FeedInfo.ServiceDateTo != "" {
		if !write("  Service Period: %s to %s\n", report.Summary.FeedInfo.ServiceDateFrom, report.Summary.FeedInfo.ServiceDateTo) { return }
	}

	if !write("\nValidation Results:\n") { return }
	if !write("  Errors: %d\n", report.Summary.Counts.Errors) { return }
	if !write("  Warnings: %d\n", report.Summary.Counts.Warnings) { return }
	if !write("  Infos: %d\n", report.Summary.Counts.Infos) { return }
	if !write("  Total: %d\n", report.Summary.Counts.Total) { return }

	if report.HasErrors() {
		write("\n❌ Validation FAILED - Feed contains errors\n")
	} else if report.HasWarnings() {
		write("\n⚠️  Validation completed with warnings\n")
	} else {
		write("\n✅ Validation PASSED\n")
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
			fmt.Fprintf(output, "\n... and %d more notices (use -f json for full details)\n", len(report.Notices)-10)
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

func outputHTML(output *os.File, report *gtfsvalidator.ValidationReport, inputPath string) error {
	// Create HTML formatter
	formatter, err := gtfsvalidator.NewHTMLFormatter()
	if err != nil {
		return fmt.Errorf("failed to create HTML formatter: %v", err)
	}

	// Generate HTML report
	return formatter.GenerateHTML(report, output)
}
