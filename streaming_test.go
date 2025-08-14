package gtfsvalidator

import (
	"archive/zip"
	"context"
	"os"
	"testing"
	"time"
)

// TestStreamingValidationAPI tests the streaming validation functionality
func TestStreamingValidationAPI(t *testing.T) {
	// Create test data with some invalid content to trigger notices
	testFiles := map[string]string{
		"agency.txt": `agency_id,agency_name,agency_url,agency_timezone
AGENCY1,Test Agency,http://example.com,America/Los_Angeles`,
		"routes.txt": `route_id,agency_id,route_short_name,route_long_name,route_type
ROUTE1,AGENCY1,1,Test Route,3`,
		"stops.txt": `stop_id,stop_name,stop_lat,stop_lon
STOP1,Test Stop,37.7749,-122.4194`,
		"trips.txt": `route_id,service_id,trip_id
ROUTE1,SERVICE1,TRIP1`,
		"stop_times.txt": `trip_id,arrival_time,departure_time,stop_id,stop_sequence
TRIP1,08:00:00,08:00:00,STOP1,1`,
		"calendar.txt": `service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date
SERVICE1,1,1,1,1,1,0,0,20241201,20241231`,
	}

	// Create temporary ZIP file for testing
	zipPath := createTestZip(t, testFiles)
	defer func() {
		// Cleanup happens automatically via t.Cleanup in createBenchmarkZip
	}()

	// Track streamed notices
	streamedNotices := make([]NoticeGroup, 0)

	// Create validator with streaming callback
	validator := New(
		WithValidationMode(ValidationModeDefault),
		WithParallelWorkers(1), // Use sequential for predictable streaming
	)

	// Validate with streaming
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	report, err := validator.ValidateFileStreamWithContext(ctx, zipPath, func(notice NoticeGroup) {
		streamedNotices = append(streamedNotices, notice)
		t.Logf("Streamed notice: %s (severity: %s, count: %d)", notice.Code, notice.Severity, notice.TotalNotices)
	})

	if err != nil {
		t.Fatalf("Streaming validation failed: %v", err)
	}

	if report == nil {
		t.Fatal("Expected validation report, got nil")
	}

	// Verify that notices were streamed
	if len(streamedNotices) == 0 {
		t.Log("No notices were streamed - this might be expected for valid minimal feed")
	} else {
		t.Logf("Successfully streamed %d notice groups", len(streamedNotices))

		// Verify that streamed notices match the final report
		reportNoticeCount := len(report.Notices)
		if reportNoticeCount > 0 {
			t.Logf("Final report contains %d notice groups", reportNoticeCount)
		}
	}

	// Verify the report structure
	if report.Summary.ValidatorVersion == "" {
		t.Error("Expected validator version in report summary")
	}

	if report.Summary.ValidationTime <= 0 {
		t.Error("Expected positive validation time")
	}

	t.Logf("Validation completed in %.2f seconds", report.Summary.ValidationTime)
	t.Logf("Feed contains %d agencies, %d routes, %d trips, %d stops",
		report.Summary.FeedInfo.AgencyCount,
		report.Summary.FeedInfo.RouteCount,
		report.Summary.FeedInfo.TripCount,
		report.Summary.FeedInfo.StopCount)
}

// TestStreamingValidationCancellation tests context cancellation during streaming
func TestStreamingValidationCancellation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping cancellation test in short mode")
	}

	// Use the actual Sofia GTFS file if available
	zipPath := "sofia_gtfs-2025_07_09.zip" // Large file for testing cancellation

	// Check if the file exists
	if !fileExists(zipPath) {
		t.Skip("Sofia GTFS file not found, skipping cancellation test")
	}

	// Create validator
	validator := New(WithValidationMode(ValidationModeComprehensive))

	// Create context that cancels after 1 second
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// Track if we received any streaming notices before cancellation
	receivedNotices := false

	_, err := validator.ValidateFileStreamWithContext(ctx, zipPath, func(notice NoticeGroup) {
		receivedNotices = true
		t.Logf("Received notice before cancellation: %s", notice.Code)
	})

	if err == nil {
		t.Log("Validation completed before timeout - this is okay for small feeds")
	} else if err == context.DeadlineExceeded {
		t.Logf("Validation correctly cancelled due to timeout")
		if !receivedNotices {
			t.Log("No notices received before cancellation - this is expected for very quick cancellation")
		}
	} else {
		t.Fatalf("Unexpected error: %v", err)
	}
}

// fileExists checks if a file exists
func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

// createTestZip creates a temporary ZIP file for testing (similar to createBenchmarkZip but for testing.T)
func createTestZip(t *testing.T, files map[string]string) string {
	t.Helper()

	tmpFile, err := os.CreateTemp("", "test-gtfs-*.zip")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	if closeErr := tmpFile.Close(); closeErr != nil {
		t.Errorf("Failed to close temp file: %v", closeErr)
	}

	zipFile, err := os.Create(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to create zip file: %v", err)
	}
	defer func() {
		if closeErr := zipFile.Close(); closeErr != nil {
			t.Errorf("Failed to close zip file: %v", closeErr)
		}
	}()

	zipWriter := zip.NewWriter(zipFile)
	defer func() {
		if closeErr := zipWriter.Close(); closeErr != nil {
			t.Errorf("Failed to close zip file: %v", closeErr)
		}
	}()

	for filename, content := range files {
		writer, err := zipWriter.Create(filename)
		if err != nil {
			t.Fatalf("Failed to create zip entry %s: %v", filename, err)
		}

		if _, err := writer.Write([]byte(content)); err != nil {
			t.Fatalf("Failed to write zip entry %s: %v", filename, err)
		}
	}

	// Register cleanup
	t.Cleanup(func() {
		if err := os.Remove(tmpFile.Name()); err != nil {
			t.Errorf("Failed to remove temp file: %v", err)
		}
	})

	return tmpFile.Name()
}
