package report

import (
	"encoding/json"
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
)

func TestReportGenerator_GenerateReport(t *testing.T) {
	container := notice.NewNoticeContainer()

	// Add a mix of notices
	for i := 0; i < 7; i++ { // exceed default sample cap (5)
		container.AddNotice(notice.NewBaseNotice("invalid_url", notice.ERROR, map[string]interface{}{
			"filename":     "agency.txt",
			"fieldName":    "agency_url",
			"fieldValue":   "not-a-url",
			"csvRowNumber": i + 2,
		}))
	}
	for i := 0; i < 3; i++ {
		container.AddNotice(notice.NewBaseNotice("whitespace_only_field", notice.WARNING, map[string]interface{}{
			"filename":     "stops.txt",
			"fieldName":    "stop_desc",
			"csvRowNumber": i + 2,
		}))
	}

	gen := NewReportGenerator("v0.0.0-test")
	feed := FeedInfo{FeedPath: "test.zip", AgencyCount: 1}
	r := gen.GenerateReport(container, feed, 0.123)

	if r == nil {
		t.Fatal("GenerateReport returned nil")
	}

	// Summary checks
	if r.Summary.ValidatorVersion != "v0.0.0-test" {
		t.Errorf("expected validator version v0.0.0-test, got %s", r.Summary.ValidatorVersion)
	}
	if r.Summary.FeedInfo.FeedPath != "test.zip" {
		t.Errorf("expected feed path test.zip, got %s", r.Summary.FeedInfo.FeedPath)
	}

	// Notice count checks
	if r.Summary.Counts.Total != 10 {
		t.Errorf("expected total notices 10, got %d", r.Summary.Counts.Total)
	}
	if r.Summary.Counts.Errors != 7 {
		t.Errorf("expected error count 7, got %d", r.Summary.Counts.Errors)
	}
	if r.Summary.Counts.Warnings != 3 {
		t.Errorf("expected warning count 3, got %d", r.Summary.Counts.Warnings)
	}

	// Build a map of code -> report
	byCode := map[string]NoticeReport{}
	for _, nr := range r.Notices {
		byCode[nr.Code] = nr
	}

	invURL, ok := byCode["invalid_url"]
	if !ok {
		t.Fatalf("missing invalid_url notice group")
	}
	if invURL.TotalNotices != 7 {
		t.Errorf("expected 7 invalid_url notices, got %d", invURL.TotalNotices)
	}
	if invURL.Severity != notice.ERROR.String() {
		t.Errorf("invalid_url severity mismatch: %s", invURL.Severity)
	}
	if len(invURL.SampleNotices) != 5 { // capped
		t.Errorf("expected 5 sample notices, got %d", len(invURL.SampleNotices))
	}

	wsOnly, ok := byCode["whitespace_only_field"]
	if !ok {
		t.Fatalf("missing whitespace_only_field notice group")
	}
	if wsOnly.TotalNotices != 3 {
		t.Errorf("expected 3 whitespace_only_field notices, got %d", wsOnly.TotalNotices)
	}
}

func TestValidationReport_JSON(t *testing.T) {
	container := notice.NewNoticeContainer()
	container.AddNotice(notice.NewBaseNotice("test_notice", notice.INFO, map[string]interface{}{"k": "v"}))

	gen := NewReportGenerator("v1")
	r := gen.GenerateReport(container, FeedInfo{FeedPath: "p"}, 1.5)

	pretty, err := r.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON error: %v", err)
	}
	if len(pretty) == 0 {
		t.Fatal("ToJSON returned empty")
	}

	compact, err := r.ToJSONCompact()
	if err != nil {
		t.Fatalf("ToJSONCompact error: %v", err)
	}
	if len(compact) == 0 {
		t.Fatal("ToJSONCompact returned empty")
	}

	// Ensure compact JSON decodes back
	var decoded ValidationReport
	if err := json.Unmarshal(compact, &decoded); err != nil {
		t.Fatalf("unmarshal compact failed: %v", err)
	}
	if decoded.Summary.FeedInfo.FeedPath != "p" {
		t.Errorf("unexpected feed path: %s", decoded.Summary.FeedInfo.FeedPath)
	}
}

func TestValidationReport_HasFlags(t *testing.T) {
	container := notice.NewNoticeContainer()
	container.AddNotice(notice.NewBaseNotice("a", notice.ERROR, map[string]interface{}{}))
	container.AddNotice(notice.NewBaseNotice("b", notice.WARNING, map[string]interface{}{}))

	gen := NewReportGenerator("v1")
	r := gen.GenerateReport(container, FeedInfo{}, 0)
	if !r.HasErrors() {
		t.Error("expected HasErrors true")
	}
	if !r.HasWarnings() {
		t.Error("expected HasWarnings true")
	}
}
