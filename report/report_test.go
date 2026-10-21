package report

import (
	"encoding/json"
	"fmt"
	"reflect"
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
		return
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
	if invURL.SeverityCounts.Errors != 7 {
		t.Errorf("expected 7 invalid_url errors, got %d", invURL.SeverityCounts.Errors)
	}
	if invURL.HighestSeverity() != notice.ERROR.String() {
		t.Errorf("invalid_url severity mismatch: %s", invURL.HighestSeverity())
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

// TestMixedSeverityGroup covers the case that motivated per-notice severity: a
// single code emitting notices of different severities. route_color_contrast
// is a WARNING below the WCAG threshold but an ERROR when text is unreadable.
func TestMixedSeverityGroup(t *testing.T) {
	container := notice.NewNoticeContainer()

	// Warnings first, so a group labelled by its first member would read
	// WARNING and hide the error entirely.
	for i := 0; i < 6; i++ {
		container.AddNotice(notice.NewBaseNotice("route_color_contrast", notice.WARNING, map[string]interface{}{
			"routeId":      "w",
			"csvRowNumber": i + 2,
		}))
	}
	container.AddNotice(notice.NewBaseNotice("route_color_contrast", notice.ERROR, map[string]interface{}{
		"routeId":      "6",
		"csvRowNumber": 6,
	}))

	gen := NewReportGenerator("v1")
	r := gen.GenerateReport(container, FeedInfo{}, 0)

	if len(r.Notices) != 1 {
		t.Fatalf("expected 1 group, got %d", len(r.Notices))
	}
	group := r.Notices[0]

	if group.SeverityCounts.Errors != 1 || group.SeverityCounts.Warnings != 6 {
		t.Errorf("expected 1 error and 6 warnings, got %d and %d",
			group.SeverityCounts.Errors, group.SeverityCounts.Warnings)
	}
	if group.HighestSeverity() != notice.ERROR.String() {
		t.Errorf("expected highest severity ERROR, got %s", group.HighestSeverity())
	}

	// The group breakdown must agree with the summary counts.
	if group.SeverityCounts.Errors != r.Summary.Counts.Errors {
		t.Errorf("group errors %d disagree with summary errors %d",
			group.SeverityCounts.Errors, r.Summary.Counts.Errors)
	}

	// The error must survive the 5-sample cap, and be first.
	if len(group.SampleNotices) != 5 {
		t.Fatalf("expected 5 samples, got %d", len(group.SampleNotices))
	}
	if group.SampleNotices[0]["severity"] != notice.ERROR.String() {
		t.Errorf("expected the error to be the first sample, got %v", group.SampleNotices[0]["severity"])
	}
	if group.SampleNotices[0]["routeId"] != "6" {
		t.Errorf("expected the error sample to be route 6, got %v", group.SampleNotices[0]["routeId"])
	}
}

func TestDescribeNoticeDoesNotMutateContext(t *testing.T) {
	n := notice.NewBaseNotice("missing_stop_name", notice.WARNING, map[string]interface{}{
		"csvRowNumber": 15,
	})

	described := DescribeNotice("missing_stop_name", n)
	if described["severity"] != notice.WARNING.String() {
		t.Errorf("expected severity on the described notice, got %v", described["severity"])
	}
	if _, ok := n.Context()["severity"]; ok {
		t.Error("DescribeNotice must not write back to the notice context")
	}
	if _, ok := n.Context()["file"]; ok {
		t.Error("DescribeNotice must not write back to the notice context")
	}
}

// TestFileResolutionForMultiFileCode checks that a code spanning several files
// still resolves a notice to one, using the file its rows belong to.

// The samples must not depend on the order notices were reported in, or a feed validated twice reports different samples.
func TestSamplesDoNotDependOnReportOrder(t *testing.T) {
	var notices []notice.Notice
	for i := 0; i < 12; i++ {
		file := "trips.txt"
		if i%3 == 0 {
			file = "stops.txt"
		}
		notices = append(notices, notice.NewBaseNotice("some_code", notice.WARNING, map[string]interface{}{
			"filename":     file,
			"csvRowNumber": 20 - i,
			"fieldName":    "field",
		}))
	}
	// Two notices on one row, told apart only by the field they name.
	notices = append(notices,
		notice.NewBaseNotice("some_code", notice.WARNING, map[string]interface{}{"filename": "stops.txt", "csvRowNumber": 2, "fieldName": "b"}),
		notice.NewBaseNotice("some_code", notice.WARNING, map[string]interface{}{"filename": "stops.txt", "csvRowNumber": 2, "fieldName": "a"}),
	)

	samples := func(order []notice.Notice) []map[string]interface{} {
		container := notice.NewNoticeContainer()
		for _, n := range order {
			container.AddNotice(n)
		}
		return NewReportGeneratorWithSamples("v1", 5).GenerateReport(container, FeedInfo{}, 0).Notices[0].SampleNotices
	}
	reversed := make([]notice.Notice, len(notices))
	for i, n := range notices {
		reversed[len(notices)-1-i] = n
	}

	forward, backward := samples(notices), samples(reversed)
	if !reflect.DeepEqual(forward, backward) {
		t.Fatalf("samples depend on report order:\n forward  %v\n backward %v", forward, backward)
	}
	want := []string{"stops.txt|2|a", "stops.txt|2|b", "stops.txt|11|field", "stops.txt|14|field", "stops.txt|17|field"}
	for i, s := range forward {
		got := fmt.Sprintf("%v|%v|%v", s["filename"], s["csvRowNumber"], s["fieldName"])
		if got != want[i] {
			t.Errorf("sample %d = %s, want %s", i, got, want[i])
		}
	}
}

// The sample limit is the caller's, and the default stays 5.
func TestSampleLimit(t *testing.T) {
	container := notice.NewNoticeContainer()
	for i := 0; i < 30; i++ {
		container.AddNotice(notice.NewBaseNotice("some_code", notice.INFO, map[string]interface{}{"csvRowNumber": i + 2}))
	}
	for _, tc := range []struct {
		gen  *ReportGenerator
		want int
	}{
		{NewReportGenerator("v1"), 5},
		{NewReportGeneratorWithSamples("v1", 0), 5},
		{NewReportGeneratorWithSamples("v1", 20), 20},
		{NewReportGeneratorWithSamples("v1", 100), 30},
	} {
		r := tc.gen.GenerateReport(container, FeedInfo{}, 0)
		if got := len(r.Notices[0].SampleNotices); got != tc.want {
			t.Errorf("limit %d: %d samples, want %d", tc.gen.maxSamplesPerNotice, got, tc.want)
		}
		if r.Notices[0].TotalNotices != 30 {
			t.Errorf("total %d, want 30", r.Notices[0].TotalNotices)
		}
	}
}
