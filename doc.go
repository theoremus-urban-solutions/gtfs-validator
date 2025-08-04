/*
Package gtfsvalidator provides a comprehensive GTFS (General Transit Feed Specification) 
validation library for Go.

This library validates GTFS feeds against the official specification and provides 
detailed reports on errors, warnings, and informational notices. It supports both 
ZIP files and directories containing GTFS data.

Features:
  - Comprehensive validation with 60+ validators
  - Multiple validation modes (performance, default, comprehensive)
  - Thread-safe concurrent processing
  - Context-based cancellation support
  - Progress reporting
  - Configurable notice limits
  - Memory-efficient processing

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
		gtfsvalidator.WithValidationMode(gtfsvalidator.ValidationModePerformance),
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

Validation Modes:

The library supports three validation modes:

  - Performance: Runs only essential validators for fast validation (10-15s for large feeds)
  - Default: Runs standard validators excluding expensive ones (30-120s)
  - Comprehensive: Runs all validators including geospatial analysis (2+ minutes)

Thread Safety:

The validator is thread-safe and can be used concurrently. Each validation 
operation is independent and does not affect other concurrent validations.

Memory Management:

For large feeds, use the performance mode and set memory limits:

	validator := gtfsvalidator.New(
		gtfsvalidator.WithValidationMode(gtfsvalidator.ValidationModePerformance),
		gtfsvalidator.WithMaxMemory(512 * 1024 * 1024), // 512MB limit
	)

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
		for _, notice := range report.Notices {
			if notice.Severity == "ERROR" {
				fmt.Printf("Error: %s (%d instances)\n", 
					notice.Code, 
					notice.TotalNotices)
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