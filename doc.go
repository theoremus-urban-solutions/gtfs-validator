/*
Package gtfsvalidator provides a comprehensive GTFS (General Transit Feed Specification)
validation library for Go.

This library validates GTFS feeds against the official specification and provides
detailed reports on errors, warnings, and informational notices. It supports both
ZIP files and directories containing GTFS data.

Features:
  - 179 validation rules across 54 validators, all of which always run
  - Full parity with the 135 applicable Canonical GTFS Schedule Validator rules
  - Thread-safe concurrent processing with memory pools
  - Context-based cancellation support
  - Progress reporting with structured logging
  - Configurable notice limits and memory management
  - Streaming CSV processing for massive feeds (2-4M rows/sec)
  - Memory-efficient processing with constant memory usage
  - Enhanced error descriptions for every notice code

Basic Usage:

	import gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator"

	// Create a validator with default settings
	validator := gtfsvalidator.New()

	// Validate a GTFS ZIP file
	report, err := validator.ValidateFile("transit-feed.zip")
	if err != nil {
		log.Fatal(err)
	}

	if report.HasErrors() {
		fmt.Printf("Validation failed with %d errors\n", report.ErrorCount())
	}

Advanced Usage with Options:

	// Create a validator with custom configuration
	validator := gtfsvalidator.New(
		gtfsvalidator.WithCountryCode("UK"),
		gtfsvalidator.WithMaxNoticesPerType(50),
		gtfsvalidator.WithProgressCallback(func(info gtfsvalidator.ProgressInfo) {
			fmt.Printf("Progress: %.1f%% - %s\n",
				info.PercentComplete,
				info.CurrentValidator)
		}),
	)

	// Validate with context for cancellation
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	report, err := validator.ValidateFileWithContext(ctx, "large-feed.zip")

Validation Scope:

Every check the library implements runs on every feed. There is no mode or
preset that selects a subset: a Sofia feed of 685k stop times validates in
about 9.5s with all 54 validators running, and a caller cannot tell in advance
which of the rules a preset drops would have been the one that mattered.

Checks that read a table stand down when that table did not load — absent,
empty, unparseable, or missing the ids other files join it on — and record why.
This keeps one defect from being restated once per row that references it: an
emptied stops.txt reports one empty file rather than four thousand dangling
references.

Thread Safety:

The validator is thread-safe and can be used concurrently. Each validation
operation is independent and does not affect other concurrent validations.

Memory Management:

For large feeds, set memory limits and tune the worker count. The streaming
CSV parser automatically handles feeds with millions of records:

	validator := gtfsvalidator.New(
		gtfsvalidator.WithMaxMemory(512 * 1024 * 1024), // 512MB limit
		gtfsvalidator.WithParallelWorkers(4), // Optimize for your CPU cores
	)

Streaming CSV Processing:

For processing individual CSV files efficiently:

	parser, _ := parser.NewStreamingCSVParser(file, "stop_times.txt", nil)
	processor := &parser.CountingProcessor{}
	parser.ProcessStream(context.Background(), processor)
	// Processes millions of rows with constant memory usage

Error Handling:

The library distinguishes between validation errors (issues in the GTFS data)
and operational errors (file access, memory issues, etc.):

	report, err := validator.ValidateFile("feed.zip")
	if err != nil {
		// Operational error - couldn't process the file
		log.Fatal("Failed to validate:", err)
	}

	if report.HasErrors() {
		// Validation errors - issues found in the GTFS data
		// A group can hold more than one severity, so check the breakdown
		// rather than a single label.
		for _, group := range report.Notices {
			if group.SeverityCounts.Errors > 0 {
				fmt.Printf("Error: %s (%d instances)\n",
					group.Code,
					group.SeverityCounts.Errors)
			}
		}
	}

Integration with APIs:

The library is designed for easy integration with web services:

	func validateHandler(w http.ResponseWriter, r *http.Request) {
		// Parse multipart form for file upload
		file, _, err := r.FormFile("gtfs")
		if err != nil {
			http.Error(w, "Failed to read file", http.StatusBadRequest)
			return
		}
		defer file.Close()

		// Validate the uploaded file
		validator := gtfsvalidator.New()
		report, err := validator.ValidateReader(file)
		if err != nil {
			http.Error(w, "Validation failed", http.StatusInternalServerError)
			return
		}

		// Return JSON report
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(report)
	}

Notice Types:

Notices are categorized by severity:
  - ERROR: Specification violations that prevent feed usage
  - WARNING: Issues that may cause problems but don't break compatibility
  - INFO: Informational notices about feed characteristics

For more information about GTFS, see: https://gtfs.org/
*/
package gtfsvalidator
