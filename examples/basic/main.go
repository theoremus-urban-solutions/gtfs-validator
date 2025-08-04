// Basic example of using the GTFS validator library
package main

import (
	"fmt"
	"log"
	"os"

	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: basic <gtfs-file>")
		os.Exit(1)
	}

	gtfsFile := os.Args[1]

	// Create a validator with default settings
	validator := gtfsvalidator.New()

	fmt.Printf("Validating GTFS feed: %s\n", gtfsFile)

	// Validate the file
	report, err := validator.ValidateFile(gtfsFile)
	if err != nil {
		log.Fatalf("Failed to validate: %v", err)
	}

	// Display summary
	fmt.Printf("\nValidation Summary:\n")
	fmt.Printf("==================\n")
	fmt.Printf("Feed: %s\n", report.Summary.FeedInfo.FeedPath)
	fmt.Printf("Validation time: %.2fs\n", report.Summary.ValidationTime)
	fmt.Printf("\nFeed Statistics:\n")
	fmt.Printf("  Agencies: %d\n", report.Summary.FeedInfo.AgencyCount)
	fmt.Printf("  Routes: %d\n", report.Summary.FeedInfo.RouteCount)
	fmt.Printf("  Trips: %d\n", report.Summary.FeedInfo.TripCount)
	fmt.Printf("  Stops: %d\n", report.Summary.FeedInfo.StopCount)
	fmt.Printf("\nResults:\n")
	fmt.Printf("  Errors: %d\n", report.Summary.Counts.Errors)
	fmt.Printf("  Warnings: %d\n", report.Summary.Counts.Warnings)
	fmt.Printf("  Infos: %d\n", report.Summary.Counts.Infos)

	// Display first few errors if any
	if report.HasErrors() {
		fmt.Printf("\nErrors found:\n")
		errorCount := 0
		for _, notice := range report.Notices {
			if notice.Severity == "ERROR" && errorCount < 5 {
				fmt.Printf("  - %s (%d instances)\n", notice.Code, notice.TotalNotices)
				errorCount++
			}
		}
		if errorCount < report.ErrorCount() {
			fmt.Printf("  ... and %d more errors\n", report.ErrorCount()-errorCount)
		}
		os.Exit(1)
	} else {
		fmt.Println("\n✅ Validation passed!")
	}
}
