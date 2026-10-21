package report

import (
	"encoding/json"
	"fmt"
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

// DefaultMaxSamplesPerNotice is how many samples a notice group carries unless the caller asks for more.
const DefaultMaxSamplesPerNotice = 5

// NewReportGenerator creates a new report generator
func NewReportGenerator(validatorVersion string) *ReportGenerator {
	return NewReportGeneratorWithSamples(validatorVersion, DefaultMaxSamplesPerNotice)
}

// NewReportGeneratorWithSamples creates a report generator that keeps up to maxSamples samples per notice code, or the default when maxSamples is not positive.
func NewReportGeneratorWithSamples(validatorVersion string, maxSamples int) *ReportGenerator {
	if maxSamples <= 0 {
		maxSamples = DefaultMaxSamplesPerNotice
	}
	return &ReportGenerator{
		validatorVersion:    validatorVersion,
		maxSamplesPerNotice: maxSamples,
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

// getSampleNotices returns a sample of notices (limited to maxSamplesPerNotice), in SortNotices order.
func (g *ReportGenerator) getSampleNotices(code string, notices []notice.Notice) []map[string]interface{} {
	ordered := SortNotices(notices)

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

// SortNotices returns a sorted copy: most severe first, then by file, row and context, so the sample never depends on the order validators ran in.
func SortNotices(notices []notice.Notice) []notice.Notice {
	type keyed struct {
		n       notice.Notice
		file    string
		row     int64
		context string
	}
	keys := make([]keyed, len(notices))
	for i, n := range notices {
		ctx := n.Context()
		keys[i] = keyed{n: n, file: fmt.Sprint(ctx["filename"]), row: contextInt(ctx["csvRowNumber"]), context: contextKey(ctx)}
	}
	sort.SliceStable(keys, func(i, j int) bool {
		a, b := keys[i], keys[j]
		if a.n.Severity() != b.n.Severity() {
			return a.n.Severity() > b.n.Severity()
		}
		if a.file != b.file {
			return a.file < b.file
		}
		if a.row != b.row {
			return a.row < b.row
		}
		return a.context < b.context
	})
	ordered := make([]notice.Notice, len(keys))
	for i, k := range keys {
		ordered[i] = k.n
	}
	return ordered
}

// contextInt reads a row number of any integer or float type, or -1 when there is none.
func contextInt(v interface{}) int64 {
	switch n := v.(type) {
	case int:
		return int64(n)
	case int32:
		return int64(n)
	case int64:
		return n
	case float64:
		return int64(n)
	default:
		return -1
	}
}

// contextKey renders a notice context as JSON, whose map keys are sorted, so equal contexts give equal keys.
func contextKey(ctx map[string]interface{}) string {
	if b, err := json.Marshal(ctx); err == nil {
		return string(b)
	}
	return fmt.Sprint(ctx)
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
