package gtfsvalidator

import (
	"strings"
	"testing"
	"time"
)

func TestNewHTMLFormatter(t *testing.T) {
	formatter, err := NewHTMLFormatter()
	if err != nil {
		t.Fatalf("NewHTMLFormatter() failed: %v", err)
	}
	if formatter == nil {
		t.Fatal("NewHTMLFormatter() returned nil formatter")
	}
	if formatter.template == nil {
		t.Fatal("NewHTMLFormatter() returned formatter with nil template")
	}
}

func TestHTMLFormatter_GenerateHTML(t *testing.T) {
	formatter, err := NewHTMLFormatter()
	if err != nil {
		t.Fatalf("Failed to create formatter: %v", err)
	}

	// Create test report
	report := &ValidationReport{
		Summary: Summary{
			ValidatorVersion: "1.0.0",
			ValidationTime:   1.23,
			Date:             time.Now().Format(time.RFC3339),
			FeedInfo: FeedInfo{
				FeedPath:    "test_feed.zip",
				AgencyCount: 2,
				RouteCount:  10,
				TripCount:   100,
				StopCount:   50,
			},
			Counts: NoticeCounts{
				Errors:   2,
				Warnings: 3,
				Infos:    1,
				Total:    6,
			},
		},
		Notices: []NoticeGroup{
			{
				Code:         "missing_required_field",
				Severity:     "ERROR",
				TotalNotices: 2,
				SampleNotices: []map[string]interface{}{
					{
						"filename":     "stops.txt",
						"csvRowNumber": 5.0,
						"fieldName":    "stop_lat",
					},
				},
			},
			{
				Code:         "invalid_date",
				Severity:     "WARNING",
				TotalNotices: 3,
				SampleNotices: []map[string]interface{}{
					{
						"filename":     "calendar.txt",
						"csvRowNumber": 2.0,
						"fieldName":    "start_date",
						"fieldValue":   "20240229",
					},
				},
			},
			{
				Code:         "unused_shape",
				Severity:     "INFO",
				TotalNotices: 1,
				SampleNotices: []map[string]interface{}{
					{
						"filename": "shapes.txt",
						"shapeId":  "UNUSED_SHAPE",
					},
				},
			},
		},
	}

	// Generate HTML
	var output strings.Builder
	err = formatter.GenerateHTML(report, &output)
	if err != nil {
		t.Fatalf("GenerateHTML() failed: %v", err)
	}

	html := output.String()

	// Test HTML structure
	if !strings.Contains(html, "<!DOCTYPE html>") {
		t.Error("HTML output missing DOCTYPE declaration")
	}
	if !strings.Contains(html, "<html lang=\"en\">") {
		t.Error("HTML output missing html element with lang attribute")
	}
	if !strings.Contains(html, "</html>") {
		t.Error("HTML output missing closing html tag")
	}

	// Test feed information (using feed path since FeedName doesn't exist in this structure)
	if !strings.Contains(html, "test_feed.zip") {
		t.Error("HTML output missing feed path")
	}
	if !strings.Contains(html, "Agencies:") || !strings.Contains(html, ">2<") {
		t.Error("HTML output missing agency count")
	}
	if !strings.Contains(html, "Routes:") || !strings.Contains(html, ">10<") {
		t.Error("HTML output missing route count")
	}

	// Test validation results
	if !strings.Contains(html, "Errors:") || !strings.Contains(html, "color: #dc3545;\">2<") {
		t.Error("HTML output missing error count")
	}
	if !strings.Contains(html, "Warnings:") || !strings.Contains(html, "color: #ffc107;\">3<") {
		t.Error("HTML output missing warning count")
	}
	if !strings.Contains(html, "Infos:") || !strings.Contains(html, "color: #17a2b8;\">1<") {
		t.Error("HTML output missing info count")
	}

	// Test notices
	if !strings.Contains(html, "missing_required_field") {
		t.Error("HTML output missing error notice code")
	}
	if !strings.Contains(html, "invalid_date") {
		t.Error("HTML output missing warning notice code")
	}
	if !strings.Contains(html, "unused_shape") {
		t.Error("HTML output missing info notice code")
	}

	// Test notice descriptions
	if !strings.Contains(html, "A required field is missing from a GTFS file") {
		t.Error("HTML output missing error notice description")
	}
	if !strings.Contains(html, "A date field contains an invalid date") {
		t.Error("HTML output missing warning notice description")
	}

	// Test sample notice details
	if !strings.Contains(html, "stops.txt") {
		t.Error("HTML output missing sample notice filename")
	}
	if !strings.Contains(html, "stop_lat") {
		t.Error("HTML output missing sample notice field name")
	}

	// Test CSS and JavaScript presence
	if !strings.Contains(html, "<style>") {
		t.Error("HTML output missing CSS styles")
	}
	if !strings.Contains(html, "<script>") {
		t.Error("HTML output missing JavaScript")
	}

	// Test filter functionality elements
	if !strings.Contains(html, "filter-button") {
		t.Error("HTML output missing filter buttons")
	}
	if !strings.Contains(html, "search-box") {
		t.Error("HTML output missing search box")
	}

	// Test responsive design elements
	if !strings.Contains(html, "viewport") {
		t.Error("HTML output missing viewport meta tag")
	}
	if !strings.Contains(html, "@media") {
		t.Error("HTML output missing responsive CSS")
	}
}

func TestHTMLFormatter_GenerateHTMLString(t *testing.T) {
	formatter, err := NewHTMLFormatter()
	if err != nil {
		t.Fatalf("Failed to create formatter: %v", err)
	}

	// Create minimal test report
	report := &ValidationReport{
		Summary: Summary{
			ValidatorVersion: "1.0.0",
			ValidationTime:   0.5,
			Date:             time.Now().Format(time.RFC3339),
			FeedInfo: FeedInfo{
				FeedPath:    "minimal_feed.zip",
				AgencyCount: 1,
			},
			Counts: NoticeCounts{
				Total: 0,
			},
		},
		Notices: []NoticeGroup{},
	}

	html, err := formatter.GenerateHTMLString(report)
	if err != nil {
		t.Fatalf("GenerateHTMLString() failed: %v", err)
	}

	if len(html) == 0 {
		t.Error("GenerateHTMLString() returned empty string")
	}

	if !strings.Contains(html, "Validation PASSED") {
		t.Error("HTML output for clean feed should show validation passed")
	}

	if !strings.Contains(html, "No Issues Found") {
		t.Error("HTML output should show 'No Issues Found' for clean feed")
	}
}

func TestGetNoticeDescription(t *testing.T) {
	tests := []struct {
		code        string
		expected    string
		description string
	}{
		{
			code:        "missing_required_field",
			expected:    "A required field is missing from a GTFS file",
			description: "Should return predefined description for known notice codes",
		},
		{
			code:        "invalid_date",
			expected:    "A date field contains an invalid date",
			description: "Should return predefined description for date validation errors",
		},
		{
			code:        "unknown_notice_code",
			expected:    "Unknown Notice Code",
			description: "Should generate title-case description for unknown codes",
		},
		{
			code:        "some_custom_validation_error",
			expected:    "Some Custom Validation Error",
			description: "Should convert underscores to spaces and title case",
		},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			result := getNoticeDescription(tt.code)
			if result != tt.expected {
				t.Errorf("%s: expected '%s', got '%s'", tt.description, tt.expected, result)
			}
		})
	}
}

func TestHTMLFormatter_ValidationStatus(t *testing.T) {
	formatter, err := NewHTMLFormatter()
	if err != nil {
		t.Fatalf("Failed to create formatter: %v", err)
	}

	tests := []struct {
		name           string
		errorCount     int
		warningCount   int
		expectedStatus string
		description    string
	}{
		{
			name:           "validation_passed",
			errorCount:     0,
			warningCount:   0,
			expectedStatus: "Validation PASSED",
			description:    "Should show passed status when no errors or warnings",
		},
		{
			name:           "validation_with_warnings",
			errorCount:     0,
			warningCount:   2,
			expectedStatus: "Validation completed with 2 warning",
			description:    "Should show warning status when warnings but no errors",
		},
		{
			name:           "validation_failed",
			errorCount:     3,
			warningCount:   1,
			expectedStatus: "Validation FAILED: 3 error",
			description:    "Should show failed status when errors present",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := &ValidationReport{
				Summary: Summary{
					ValidatorVersion: "1.0.0",
					ValidationTime:   1.0,
					Date:             time.Now().Format(time.RFC3339),
					FeedInfo:         FeedInfo{FeedPath: "test.zip"},
					Counts: NoticeCounts{
						Errors:   tt.errorCount,
						Warnings: tt.warningCount,
						Total:    tt.errorCount + tt.warningCount,
					},
				},
				Notices: []NoticeGroup{},
			}

			html, err := formatter.GenerateHTMLString(report)
			if err != nil {
				t.Fatalf("GenerateHTMLString() failed: %v", err)
			}

			if !strings.Contains(html, tt.expectedStatus) {
				t.Errorf("%s: expected to contain '%s' in HTML output", tt.description, tt.expectedStatus)
			}
		})
	}
}

func TestHTMLFormatter_SeverityCounts(t *testing.T) {
	formatter, err := NewHTMLFormatter()
	if err != nil {
		t.Fatalf("Failed to create formatter: %v", err)
	}

	report := &ValidationReport{
		Summary: Summary{
			ValidatorVersion: "1.0.0",
			ValidationTime:   1.0,
			Date:             time.Now().Format(time.RFC3339),
			FeedInfo:         FeedInfo{FeedPath: "test.zip"},
			Counts:           NoticeCounts{Total: 6},
		},
		Notices: []NoticeGroup{
			{Code: "error1", Severity: "ERROR", TotalNotices: 1},
			{Code: "error2", Severity: "ERROR", TotalNotices: 2},
			{Code: "warning1", Severity: "WARNING", TotalNotices: 2},
			{Code: "info1", Severity: "INFO", TotalNotices: 1},
		},
	}

	html, err := formatter.GenerateHTMLString(report)
	if err != nil {
		t.Fatalf("GenerateHTMLString() failed: %v", err)
	}

	// Test filter button counts
	if !strings.Contains(html, "Errors (2)") {
		t.Error("HTML should show correct error count in filter buttons")
	}
	if !strings.Contains(html, "Warnings (1)") {
		t.Error("HTML should show correct warning count in filter buttons")
	}
	if !strings.Contains(html, "Infos (1)") {
		t.Error("HTML should show correct info count in filter buttons")
	}
	if !strings.Contains(html, "All (4)") {
		t.Error("HTML should show correct total notice group count in filter buttons")
	}
}
