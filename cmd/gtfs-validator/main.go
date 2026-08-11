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
structured logging, and comprehensive validation with 201 validation rules.`,
		Example: `  gtfs-validator -i feed.zip
  gtfs-validator -i ./gtfs-feed -f json -o report.json
  gtfs-validator -i feed.zip -f html -o report.html
  gtfs-validator -i feed.zip --progress`,
		Version: version,
		RunE:    runValidation,
	}

	// Add flags
	rootCmd.Flags().StringVarP(&inputPath, "input", "i", "", "Path to GTFS feed (ZIP file or directory) [required]")
	rootCmd.Flags().StringVarP(&outputFormat, "format", "f", "console", "Output format: console, json, summary, html")
	rootCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file path (default: stdout)")
	rootCmd.Flags().StringVarP(&countryCode, "country", "c", "", "Country code for phone number validation (e.g., BG, GB, DE); unset accepts any dialable length")
	rootCmd.Flags().Int64Var(&maxMemory, "memory", 0, "Maximum memory usage in MB (0 = no limit)")
	rootCmd.Flags().IntVarP(&workers, "workers", "w", 4, "Number of parallel workers")
	rootCmd.Flags().IntVar(&maxNotices, "max-notices", 0, "Maximum notices per type (0 = no limit, the default)")
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
comprehensive validation with 201 validation rules.`,
		Example: `  gtfs-validator validate feed.zip
  gtfs-validator validate ./gtfs-directory --format json
  gtfs-validator validate feed.zip --format html --output report.html
  gtfs-validator validate feed.zip --progress`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			inputPath = args[0]
			return runValidation(cmd, args)
		},
	}

	// Add the same flags as root command
	cmd.Flags().StringVarP(&outputFormat, "format", "f", "console", "Output format: console, json, summary, html")
	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file path (default: stdout)")
	cmd.Flags().StringVarP(&countryCode, "country", "c", "", "Country code for phone number validation (e.g., BG, GB, DE); unset accepts any dialable length")
	cmd.Flags().Int64Var(&maxMemory, "memory", 0, "Maximum memory usage in MB (0 = no limit)")
	cmd.Flags().IntVarP(&workers, "workers", "w", 4, "Number of parallel workers")
	cmd.Flags().IntVar(&maxNotices, "max-notices", 0, "Maximum notices per type (0 = no limit, the default)")
	cmd.Flags().DurationVarP(&timeout, "timeout", "t", 5*time.Minute, "Validation timeout")
	cmd.Flags().BoolVarP(&showProgress, "progress", "p", false, "Show progress bar")

	return cmd
}

func runValidation(cmd *cobra.Command, args []string) error {
	// Validate input
	if err := validateInput(inputPath, outputFormat); err != nil {
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
	if maxNotices > 0 {
		fmt.Fprintf(os.Stderr, "   Notice limit: %d per type\n", maxNotices)
	}
	fmt.Fprintf(os.Stderr, "\n")

	// Perform validation
	startTime := time.Now()
	report, err := validator.ValidateFileWithContext(ctx, inputPath)
	elapsed := time.Since(startTime)

	if err != nil {
		switch err {
		case context.Canceled:
			fmt.Fprintf(os.Stderr, "⚠️  Validation cancelled by user\n")
			cancel()
			os.Exit(1) //nolint:gocritic // cancel explicitly called above
		case context.DeadlineExceeded:
			fmt.Fprintf(os.Stderr, "⏰ Validation timed out after %v\n", timeout)
			cancel()
			os.Exit(1)
		default:
			return fmt.Errorf("❌ Validation Error: %v", err)
		}
	}

	fmt.Fprintf(os.Stderr, "✅ Validation completed in %.2fs\n\n", elapsed.Seconds())

	// Handle output
	output := os.Stdout
	var outputFileHandle *os.File
	if outputFile != "" {
		file, err := os.Create(outputFile) // #nosec G304 -- User-provided output file path
		if err != nil {
			return fmt.Errorf("❌ Output Error: Failed to create output file '%s': %v", outputFile, err)
		}
		outputFileHandle = file
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
	switch {
	case report.HasErrors():
		fmt.Fprintf(os.Stderr, "💀 Validation FAILED: %d errors found\n", report.ErrorCount())
		cancel()
		if outputFileHandle != nil {
			_ = outputFileHandle.Close() // Ignore error on program exit
		}
		os.Exit(1)
	case report.HasWarnings():
		fmt.Fprintf(os.Stderr, "⚠️  Validation completed with %d warnings\n", report.WarningCount())
	default:
		fmt.Fprintf(os.Stderr, "🎉 Validation PASSED: Feed is valid!\n")
	}

	return nil
}

func validateInput(inputPath, format string) error {
	// Check if input exists
	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		return fmt.Errorf("input error: path does not exist: '%s'", inputPath)
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

	if !write("GTFS Validation Summary\n") {
		return
	}
	if !write("======================\n\n") {
		return
	}
	if !write("Feed: %s\n", filepath.Base(inputPath)) {
		return
	}
	if !write("Validation Time: %.2fs\n\n", report.Summary.ValidationTime) {
		return
	}

	if !write("Feed Statistics:\n") {
		return
	}
	if !write("  Agencies: %d\n", report.Summary.FeedInfo.AgencyCount) {
		return
	}
	if !write("  Routes: %d\n", report.Summary.FeedInfo.RouteCount) {
		return
	}
	if !write("  Trips: %d\n", report.Summary.FeedInfo.TripCount) {
		return
	}
	if !write("  Stops: %d\n", report.Summary.FeedInfo.StopCount) {
		return
	}
	if !write("  Stop Times: %d\n", report.Summary.FeedInfo.StopTimeCount) {
		return
	}
	if report.Summary.FeedInfo.ServiceDateFrom != "" && report.Summary.FeedInfo.ServiceDateTo != "" {
		if !write("  Service Period: %s to %s\n", report.Summary.FeedInfo.ServiceDateFrom, report.Summary.FeedInfo.ServiceDateTo) {
			return
		}
	}

	if !write("\nValidation Results:\n") {
		return
	}
	if !write("  Errors: %d\n", report.Summary.Counts.Errors) {
		return
	}
	if !write("  Warnings: %d\n", report.Summary.Counts.Warnings) {
		return
	}
	if !write("  Infos: %d\n", report.Summary.Counts.Infos) {
		return
	}
	if !write("  Total: %d\n", report.Summary.Counts.Total) {
		return
	}

	switch {
	case report.HasErrors():
		write("\n❌ Validation FAILED - Feed contains errors\n")
	case report.HasWarnings():
		write("\n⚠️  Validation completed with warnings\n")
	default:
		write("\n✅ Validation PASSED\n")
	}
}

func outputConsole(output *os.File, report *gtfsvalidator.ValidationReport, inputPath string) {
	outputSummary(output, report, inputPath)

	// Show first few notices if any
	if len(report.Notices) > 0 {
		// Helper function to write output with error checking
		write := func(format string, args ...interface{}) {
			if _, err := fmt.Fprintf(output, format, args...); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: Failed to write console output: %v\n", err)
			}
		}

		write("\nSample Notices:\n")
		write("===============\n")

		errorCount := 0
		warningCount := 0

		// A group can hold more than one severity, so report it under every
		// severity it contains rather than under a single label.
		for _, group := range report.Notices {
			if errorCount >= 5 && warningCount >= 5 {
				break
			}

			if group.SeverityCounts.Errors > 0 && errorCount < 5 {
				write("ERROR: %s (%d instances)\n", group.Code, group.SeverityCounts.Errors)
				showFirstSampleOfSeverity(output, group, "ERROR")
				errorCount++
			}
			if group.SeverityCounts.Warnings > 0 && warningCount < 5 {
				write("WARNING: %s (%d instances)\n", group.Code, group.SeverityCounts.Warnings)
				showFirstSampleOfSeverity(output, group, "WARNING")
				warningCount++
			}
		}

		if len(report.Notices) > 10 {
			write("\n... and %d more notices (use -f json for full details)\n", len(report.Notices)-10)
		}
	}
}

// showFirstSampleOfSeverity prints the context of the first sample matching a
// severity, so a mixed group shows a genuine example of the level it is
// being reported under.
func showFirstSampleOfSeverity(output *os.File, group gtfsvalidator.NoticeGroup, severity string) {
	for _, sample := range group.SampleNotices {
		if sample["severity"] == severity {
			showNoticeContext(output, sample, group.AffectedFiles)
			return
		}
	}
}

func showNoticeContext(output *os.File, context map[string]interface{}, groupFiles []string) {
	details := []string{}

	// A notice names its own file when the check is tied to one. Otherwise
	// fall back to the files the code as a whole concerns, so a row number is
	// never shown without saying which file it is a row of.
	filename, hasFile := context["file"].(string)
	if !hasFile && len(groupFiles) > 0 {
		filename, hasFile = strings.Join(groupFiles, "|"), true
	}
	line, hasLine := intValue(context["line"])

	switch {
	case hasFile && hasLine:
		details = append(details, fmt.Sprintf("%s:%d", filename, line))
	case hasFile:
		details = append(details, fmt.Sprintf("file=%s", filename))
	case hasLine:
		details = append(details, fmt.Sprintf("row=%d", line))
	}
	if field, ok := context["fieldName"].(string); ok {
		details = append(details, fmt.Sprintf("field=%s", field))
	}
	if routeId, ok := context["routeId"].(string); ok {
		details = append(details, fmt.Sprintf("route=%s", routeId))
	}

	if len(details) > 0 {
		if _, err := fmt.Fprintf(output, "       (%s)\n", strings.Join(details, ", ")); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to write notice context: %v\n", err)
		}
	}
}

// intValue reads a numeric context value. Values are ints in process but
// arrive as float64 when a report is read back from JSON.
func intValue(value interface{}) (int, bool) {
	switch v := value.(type) {
	case int:
		return v, true
	case int64:
		return int(v), true
	case float64:
		return int(v), true
	default:
		return 0, false
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
