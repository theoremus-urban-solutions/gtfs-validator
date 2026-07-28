package gtfsvalidator

import (
	"embed"
	"html/template"
	"io"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

//go:embed templates/*.html
var templateFS embed.FS

// NoticeWithDescription extends NoticeGroup with severity information
type NoticeWithDescription struct {
	NoticeGroup
	SeverityInfo SeverityInfo `json:"severityInfo"`

	// Severity is the most severe level in the group, used for the badge and
	// for ordering. It is not a label for every notice in the group.
	Severity string `json:"severity"`

	// SeverityClasses lists every severity present in the group, lowercased,
	// so the filter buttons can match a group on any of them.
	SeverityClasses string `json:"severityClasses"`

	// IsMixedSeverity is true when the group holds more than one severity, in
	// which case the badge shows the breakdown rather than a single label.
	IsMixedSeverity bool `json:"isMixedSeverity"`
}

// HTMLTemplateData represents the data structure passed to HTML templates
type HTMLTemplateData struct {
	Summary        Summary                 `json:"summary"`
	Notices        []NoticeWithDescription `json:"notices"`
	GeneratedAt    string
	SeverityCounts map[string]int
}

// HTMLFormatter handles HTML report generation
type HTMLFormatter struct {
	template *template.Template
}

// NewHTMLFormatter creates a new HTML formatter with embedded templates
func NewHTMLFormatter() (*HTMLFormatter, error) {
	// Parse the embedded template
	caser := cases.Title(language.English)
	tmpl, err := template.New("report.html").Funcs(template.FuncMap{
		"title": caser.String,
		"lower": strings.ToLower,
	}).ParseFS(templateFS, "templates/report.html")
	if err != nil {
		return nil, err
	}

	return &HTMLFormatter{
		template: tmpl,
	}, nil
}

// GenerateHTML generates an HTML report from the validation results
func (f *HTMLFormatter) GenerateHTML(report *ValidationReport, writer io.Writer) error {
	// Count individual notices rather than groups, so the filter totals match
	// the summary counts even when a group holds more than one severity.
	severityCounts := make(map[string]int)
	noticesWithDesc := make([]NoticeWithDescription, len(report.Notices))

	for i, group := range report.Notices {
		severityCounts["error"] += group.SeverityCounts.Errors
		severityCounts["warning"] += group.SeverityCounts.Warnings
		severityCounts["info"] += group.SeverityCounts.Infos

		classes := make([]string, 0, 3)
		if group.SeverityCounts.Errors > 0 {
			classes = append(classes, "error")
		}
		if group.SeverityCounts.Warnings > 0 {
			classes = append(classes, "warning")
		}
		if group.SeverityCounts.Infos > 0 {
			classes = append(classes, "info")
		}

		highest := group.HighestSeverity()
		noticesWithDesc[i] = NoticeWithDescription{
			NoticeGroup:     group,
			SeverityInfo:    GetSeverityInfo(highest),
			Severity:        highest,
			SeverityClasses: strings.Join(classes, " "),
			IsMixedSeverity: len(classes) > 1,
		}
	}

	// Most severe groups first, so errors are never buried below warnings.
	sort.SliceStable(noticesWithDesc, func(i, j int) bool {
		return severityRank(noticesWithDesc[i].Severity) > severityRank(noticesWithDesc[j].Severity)
	})

	// Prepare template data
	data := HTMLTemplateData{
		Summary:        report.Summary,
		Notices:        noticesWithDesc,
		GeneratedAt:    time.Now().Format("January 2, 2006 at 3:04 PM"),
		SeverityCounts: severityCounts,
	}

	// Execute template
	return f.template.Execute(writer, data)
}

// GenerateHTMLToFile generates an HTML report and writes it to a file
func (f *HTMLFormatter) GenerateHTMLToFile(report *ValidationReport, filename string) error {
	file, err := os.Create(filename) // #nosec G304 -- User-provided output filename
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			log.Printf("Warning: failed to close %v", closeErr)
		}
	}()

	return f.GenerateHTML(report, file)
}

// severityRank orders severity levels from most to least severe.
func severityRank(severity string) int {
	switch severity {
	case SeverityLevelError:
		return 3
	case SeverityLevelWarning:
		return 2
	case SeverityLevelInfo:
		return 1
	default:
		return 0
	}
}

// GenerateHTMLString generates an HTML report as a string
func (f *HTMLFormatter) GenerateHTMLString(report *ValidationReport) (string, error) {
	var buf strings.Builder
	err := f.GenerateHTML(report, &buf)
	return buf.String(), err
}
