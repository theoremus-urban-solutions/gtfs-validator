package gtfsvalidator

import (
	"embed"
	"html/template"
	"io"
	"log"
	"os"
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
	// Calculate severity counts and add descriptions
	severityCounts := make(map[string]int)
	noticesWithDesc := make([]NoticeWithDescription, len(report.Notices))

	for i, notice := range report.Notices {
		severity := strings.ToLower(notice.Severity)
		severityCounts[severity] += 1

		// Add severity information to notice
		noticesWithDesc[i] = NoticeWithDescription{
			NoticeGroup:  notice,
			SeverityInfo: GetSeverityInfo(notice.Severity),
		}
	}

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

// GenerateHTMLString generates an HTML report as a string
func (f *HTMLFormatter) GenerateHTMLString(report *ValidationReport) (string, error) {
	var buf strings.Builder
	err := f.GenerateHTML(report, &buf)
	return buf.String(), err
}
