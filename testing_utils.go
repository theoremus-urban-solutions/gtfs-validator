package gtfsvalidator

import (
	"archive/zip"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestDataPath returns the path to test data directory
func TestDataPath(subpath string) string {
	return filepath.Join("testdata", subpath)
}

// CreateTestZipFromDir creates a ZIP file from a directory for testing
func CreateTestZipFromDir(t *testing.T, srcDir, zipPath string) {
	t.Helper()

	zipFile, err := os.Create(zipPath) // #nosec G304 -- This is test code with controlled paths
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

	err = filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		zipEntry, err := zipWriter.Create(relPath)
		if err != nil {
			return err
		}

		file, err := os.Open(path) // #nosec G304 -- This is test code with controlled paths
		if err != nil {
			return err
		}
		defer func() {
			if closeErr := file.Close(); closeErr != nil {
				log.Printf("Warning: failed to close %v", closeErr)
			}
		}()

		_, err = io.Copy(zipEntry, file)
		return err
	})

	if err != nil {
		t.Fatalf("Failed to create zip file: %v", err)
	}
}

// AssertValidationReport provides assertions for validation reports
type AssertValidationReport struct {
	t      *testing.T
	report *ValidationReport
}

// NewAssertValidationReport creates a new validation report asserter
func NewAssertValidationReport(t *testing.T, report *ValidationReport) *AssertValidationReport {
	t.Helper()
	if report == nil {
		t.Fatal("ValidationReport is nil")
	}
	return &AssertValidationReport{t: t, report: report}
}

// HasErrors asserts the report has errors
func (a *AssertValidationReport) HasErrors() *AssertValidationReport {
	a.t.Helper()
	if !a.report.HasErrors() {
		a.t.Error("Expected validation report to have errors")
	}
	return a
}

// HasNoErrors asserts the report has no errors
func (a *AssertValidationReport) HasNoErrors() *AssertValidationReport {
	a.t.Helper()
	if a.report.HasErrors() {
		a.t.Errorf("Expected validation report to have no errors, but found %d errors", a.report.ErrorCount())
		a.logFirstFewNotices("ERROR")
	}
	return a
}

// HasWarnings asserts the report has warnings
func (a *AssertValidationReport) HasWarnings() *AssertValidationReport {
	a.t.Helper()
	if !a.report.HasWarnings() {
		a.t.Error("Expected validation report to have warnings")
	}
	return a
}

// HasNoWarnings asserts the report has no warnings
func (a *AssertValidationReport) HasNoWarnings() *AssertValidationReport {
	a.t.Helper()
	if a.report.HasWarnings() {
		a.t.Errorf("Expected validation report to have no warnings, but found %d warnings", a.report.WarningCount())
		a.logFirstFewNotices("WARNING")
	}
	return a
}

// ErrorCountEquals asserts the error count matches expected
func (a *AssertValidationReport) ErrorCountEquals(expected int) *AssertValidationReport {
	a.t.Helper()
	actual := a.report.ErrorCount()
	if actual != expected {
		a.t.Errorf("Expected %d errors, got %d", expected, actual)
		if actual > 0 {
			a.logFirstFewNotices("ERROR")
		}
	}
	return a
}

// WarningCountEquals asserts the warning count matches expected
func (a *AssertValidationReport) WarningCountEquals(expected int) *AssertValidationReport {
	a.t.Helper()
	actual := a.report.WarningCount()
	if actual != expected {
		a.t.Errorf("Expected %d warnings, got %d", expected, actual)
		if actual > 0 {
			a.logFirstFewNotices("WARNING")
		}
	}
	return a
}

// ContainsNotice asserts the report contains a notice with the given code
func (a *AssertValidationReport) ContainsNotice(code string) *AssertValidationReport {
	a.t.Helper()
	for _, notice := range a.report.Notices {
		if notice.Code == code {
			return a
		}
	}
	a.t.Errorf("Expected validation report to contain notice with code %q", code)
	a.logAllNoticeCodes()
	return a
}

// DoesNotContainNotice asserts the report does not contain a notice with the given code
func (a *AssertValidationReport) DoesNotContainNotice(code string) *AssertValidationReport {
	a.t.Helper()
	for _, notice := range a.report.Notices {
		if notice.Code == code {
			a.t.Errorf("Expected validation report to NOT contain notice with code %q", code)
			return a
		}
	}
	return a
}

// FeedInfoEquals asserts the feed info matches expected values
func (a *AssertValidationReport) FeedInfoEquals(expected FeedInfo) *AssertValidationReport {
	a.t.Helper()
	actual := a.report.Summary.FeedInfo

	if actual.AgencyCount != expected.AgencyCount {
		a.t.Errorf("Expected AgencyCount = %d, got %d", expected.AgencyCount, actual.AgencyCount)
	}
	if actual.RouteCount != expected.RouteCount {
		a.t.Errorf("Expected RouteCount = %d, got %d", expected.RouteCount, actual.RouteCount)
	}
	if actual.TripCount != expected.TripCount {
		a.t.Errorf("Expected TripCount = %d, got %d", expected.TripCount, actual.TripCount)
	}
	if actual.StopCount != expected.StopCount {
		a.t.Errorf("Expected StopCount = %d, got %d", expected.StopCount, actual.StopCount)
	}
	if actual.StopTimeCount != expected.StopTimeCount {
		a.t.Errorf("Expected StopTimeCount = %d, got %d", expected.StopTimeCount, actual.StopTimeCount)
	}

	return a
}

// Helper methods for logging

func (a *AssertValidationReport) logFirstFewNotices(severity string) {
	a.t.Helper()
	count := 0
	for _, notice := range a.report.Notices {
		if notice.Severity == severity && count < 5 {
			a.t.Logf("  %s: %s (%d instances)", severity, notice.Code, notice.TotalNotices)
			count++
		}
	}
	if count >= 5 {
		a.t.Logf("  ... and more %s notices", strings.ToLower(severity))
	}
}

func (a *AssertValidationReport) logAllNoticeCodes() {
	a.t.Helper()
	a.t.Log("Available notice codes:")
	for _, notice := range a.report.Notices {
		a.t.Logf("  - %s (%s)", notice.Code, notice.Severity)
	}
}

// CreateTempZip creates a temporary ZIP file for testing
func CreateTempZip(t *testing.T, files map[string]string) string {
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

// ValidateAndAssert is a helper that validates and returns an asserter
func ValidateAndAssert(t *testing.T, validator Validator, path string) *AssertValidationReport {
	t.Helper()

	report, err := validator.ValidateFile(path)
	if err != nil {
		t.Fatalf("Validation failed: %v", err)
	}

	return NewAssertValidationReport(t, report)
}

// Example test data generators

// MinimalValidGTFS returns the content for a minimal valid GTFS feed
func MinimalValidGTFS() map[string]string {
	return map[string]string{
		"agency.txt": `agency_id,agency_name,agency_url,agency_timezone
test_agency,Test Transit Agency,https://example.com,America/New_York`,
		"routes.txt": `route_id,agency_id,route_short_name,route_long_name,route_type
route_1,test_agency,1,Main Street Line,3`,
		"stops.txt": `stop_id,stop_name,stop_lat,stop_lon
stop_1,First Stop,40.7589,-73.9851
stop_2,Second Stop,40.7614,-73.9776`,
		"trips.txt": `route_id,service_id,trip_id,trip_headsign
route_1,service_1,trip_1,Downtown`,
		"stop_times.txt": `trip_id,arrival_time,departure_time,stop_id,stop_sequence
trip_1,08:00:00,08:00:00,stop_1,1
trip_1,08:15:00,08:15:00,stop_2,2`,
		"calendar.txt": `service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date
service_1,1,1,1,1,1,0,0,20250101,20251231`,
	}
}

// InvalidGTFS returns the content for an invalid GTFS feed with known errors
func InvalidGTFS() map[string]string {
	return map[string]string{
		"agency.txt": `agency_id,agency_name,agency_url,agency_timezone
test_agency,Test Transit Agency,not_a_valid_url,Invalid/Timezone`,
		"routes.txt": `route_id,agency_id,route_short_name,route_long_name,route_type
route_1,nonexistent_agency,1,Main Street Line,999`,
		"stops.txt": `stop_id,stop_name,stop_lat,stop_lon
stop_1,,999.0,-999.0
stop_2,Second Stop,invalid_lat,invalid_lon`,
		// Missing required files trips.txt, stop_times.txt, calendar.txt
	}
}
