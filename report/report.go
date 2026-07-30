package report

import (
	"encoding/json"
	"sort"
	"time"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
)

// ValidationReport represents the complete validation report
type ValidationReport struct {
	Summary Summary        `json:"summary"`
	Notices []NoticeReport `json:"notices"`
}

// Summary contains summary information about the validation
type Summary struct {
	ValidatorVersion string       `json:"validatorVersion"`
	ValidationTime   float64      `json:"validationTimeSeconds"`
	Date             string       `json:"date"`
	FeedInfo         FeedInfo     `json:"feedInfo"`
	Counts           NoticeCounts `json:"counts"`
}

// FeedInfo contains information about the validated feed
type FeedInfo struct {
	FeedPath        string `json:"feedPath"`
	FeedName        string `json:"feedName,omitempty"`
	AgencyCount     int    `json:"agencyCount"`
	RouteCount      int    `json:"routeCount"`
	TripCount       int    `json:"tripCount"`
	StopCount       int    `json:"stopCount"`
	StopTimeCount   int    `json:"stopTimeCount"`
	ServiceDateFrom string `json:"serviceDateFrom,omitempty"`
	ServiceDateTo   string `json:"serviceDateTo,omitempty"`
}

// NoticeCounts contains counts of notices by severity
type NoticeCounts struct {
	Errors   int `json:"errors"`
	Warnings int `json:"warnings"`
	Infos    int `json:"infos"`
	Total    int `json:"total"`
}

// NoticeReport represents a group of notices with the same code.
//
// A group deliberately has no single severity: one code can produce notices of
// different severities (route_color_contrast is a WARNING below the WCAG
// threshold but an ERROR when the text is unreadable), and labelling the group
// with any one of them hides the others. SeverityCounts reports the breakdown
// and each sample carries its own severity.
type NoticeReport struct {
	Code           string                   `json:"code"`
	SeverityCounts NoticeCounts             `json:"severityCounts"`
	Description    string                   `json:"description"`
	AffectedFiles  []string                 `json:"affectedFiles,omitempty"`
	TotalNotices   int                      `json:"totalNotices"`
	SampleNotices  []map[string]interface{} `json:"sampleNotices"`
}

// HighestSeverity returns the most severe level present in the group. Used for
// sorting and colouring, never as a label for the group as a whole.
func (n *NoticeReport) HighestSeverity() string {
	switch {
	case n.SeverityCounts.Errors > 0:
		return notice.ERROR.String()
	case n.SeverityCounts.Warnings > 0:
		return notice.WARNING.String()
	default:
		return notice.INFO.String()
	}
}

// ReportGenerator generates validation reports
type ReportGenerator struct {
	validatorVersion    string
	maxSamplesPerNotice int
}

// NewReportGenerator creates a new report generator
func NewReportGenerator(validatorVersion string) *ReportGenerator {
	return &ReportGenerator{
		validatorVersion:    validatorVersion,
		maxSamplesPerNotice: 5, // Limit samples to prevent huge reports
	}
}

// GenerateReport generates a validation report from a notice container
func (g *ReportGenerator) GenerateReport(container *notice.NoticeContainer, feedInfo FeedInfo, validationTime float64) *ValidationReport {
	// Group notices by code
	noticeGroups := g.groupNoticesByCode(container.GetNotices())

	// Create notice reports
	noticeReports := make([]NoticeReport, 0, len(noticeGroups))
	for code, notices := range noticeGroups {
		if len(notices) == 0 {
			continue
		}

		report := NoticeReport{
			Code:           code,
			SeverityCounts: countBySeverity(notices),
			Description:    "", // Will be populated by the main package
			AffectedFiles:  notice.AffectedFilesForCode(code),
			TotalNotices:   len(notices),
			SampleNotices:  g.getSampleNotices(code, notices),
		}
		noticeReports = append(noticeReports, report)
	}

	// Calculate counts
	counts := container.CountBySeverity()
	noticeCounts := NoticeCounts{
		Errors:   counts[notice.ERROR],
		Warnings: counts[notice.WARNING],
		Infos:    counts[notice.INFO],
		Total:    len(container.GetNotices()),
	}

	// Create summary
	summary := Summary{
		ValidatorVersion: g.validatorVersion,
		ValidationTime:   validationTime,
		Date:             time.Now().Format(time.RFC3339),
		FeedInfo:         feedInfo,
		Counts:           noticeCounts,
	}

	return &ValidationReport{
		Summary: summary,
		Notices: noticeReports,
	}
}

// groupNoticesByCode groups notices by their code
func (g *ReportGenerator) groupNoticesByCode(notices []notice.Notice) map[string][]notice.Notice {
	groups := make(map[string][]notice.Notice)
	for _, n := range notices {
		code := n.Code()
		groups[code] = append(groups[code], n)
	}
	return groups
}

// countBySeverity counts the notices in a group by severity level.
func countBySeverity(notices []notice.Notice) NoticeCounts {
	counts := NoticeCounts{Total: len(notices)}
	for _, n := range notices {
		switch n.Severity() {
		case notice.ERROR:
			counts.Errors++
		case notice.WARNING:
			counts.Warnings++
		case notice.INFO:
			counts.Infos++
		}
	}
	return counts
}

// getSampleNotices returns a sample of notices (limited to maxSamplesPerNotice).
//
// Samples are ordered most severe first so that the cap can never hide the
// errors in a group that is mostly warnings. Order within a severity is the
// order the notices were reported in.
func (g *ReportGenerator) getSampleNotices(code string, notices []notice.Notice) []map[string]interface{} {
	ordered := make([]notice.Notice, len(notices))
	copy(ordered, notices)
	sort.SliceStable(ordered, func(i, j int) bool {
		return ordered[i].Severity() > ordered[j].Severity()
	})

	limit := g.maxSamplesPerNotice
	if len(ordered) < limit {
		limit = len(ordered)
	}

	samples := make([]map[string]interface{}, limit)
	for i := 0; i < limit; i++ {
		samples[i] = DescribeNotice(code, ordered[i])
	}

	return samples
}

// DescribeNotice renders a single notice as a map, adding the severity, file
// and line that identify it. The notice's own context is copied rather than
// mutated so that the container is left untouched.
func DescribeNotice(code string, n notice.Notice) map[string]interface{} {
	context := n.Context()
	described := make(map[string]interface{}, len(context)+3)
	for k, v := range context {
		described[k] = v
	}

	described["severity"] = n.Severity().String()
	if file, ok := notice.FileName(code, context); ok {
		described["file"] = file
	}
	if line, ok := notice.LineNumber(context); ok {
		described["line"] = line
	}

	return described
}

// ToJSON converts the report to JSON
func (r *ValidationReport) ToJSON() ([]byte, error) {
	return json.MarshalIndent(r, "", "  ")
}

// ToJSONCompact converts the report to compact JSON
func (r *ValidationReport) ToJSONCompact() ([]byte, error) {
	return json.Marshal(r)
}

// HasErrors returns true if the report contains any errors
func (r *ValidationReport) HasErrors() bool {
	return r.Summary.Counts.Errors > 0
}

// HasWarnings returns true if the report contains any warnings
func (r *ValidationReport) HasWarnings() bool {
	return r.Summary.Counts.Warnings > 0
}
