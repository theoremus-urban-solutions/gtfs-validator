package report

import (
	"encoding/json"
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

// NoticeReport represents a group of notices with the same code
type NoticeReport struct {
	Code          string                   `json:"code"`
	Severity      string                   `json:"severity"`
	Description   string                   `json:"description"`
	TotalNotices  int                      `json:"totalNotices"`
	SampleNotices []map[string]interface{} `json:"sampleNotices"`
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
			Code:          code,
			Severity:      notices[0].Severity().String(),
			Description:   "", // Will be populated by the main package
			TotalNotices:  len(notices),
			SampleNotices: g.getSampleNotices(notices),
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

// getSampleNotices returns a sample of notices (limited to maxSamplesPerNotice)
func (g *ReportGenerator) getSampleNotices(notices []notice.Notice) []map[string]interface{} {
	limit := g.maxSamplesPerNotice
	if len(notices) < limit {
		limit = len(notices)
	}

	samples := make([]map[string]interface{}, limit)
	for i := 0; i < limit; i++ {
		samples[i] = notices[i].Context()
	}

	return samples
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
