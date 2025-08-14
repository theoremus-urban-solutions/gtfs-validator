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

// NoticeWithDescription extends NoticeGroup with a human-readable description
type NoticeWithDescription struct {
	NoticeGroup
	Description string `json:"description"`
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

		// Add description to notice
		noticesWithDesc[i] = NoticeWithDescription{
			NoticeGroup: notice,
			Description: getNoticeDescription(notice.Code),
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

// getNoticeDescription returns a human-readable description for a notice code
// This is a fallback since notice codes might not have built-in descriptions
func getNoticeDescription(code string) string {
	descriptions := map[string]string{
		// Common GTFS validation errors
		"missing_required_file":                   "A required GTFS file is missing from the feed",
		"missing_required_field":                  "A required field is missing from a GTFS file",
		"invalid_row_length":                      "A row has the wrong number of fields",
		"invalid_date":                            "A date field contains an invalid date",
		"invalid_time":                            "A time field contains an invalid time",
		"invalid_color":                           "A color field contains an invalid color code",
		"invalid_email":                           "An email field contains an invalid email address",
		"invalid_phone_number":                    "A phone number field contains an invalid phone number",
		"invalid_timezone":                        "A timezone field contains an invalid timezone",
		"invalid_currency_code":                   "A currency code field contains an invalid currency code",
		"invalid_language_code":                   "A language code field contains an invalid language code",
		"duplicate_key":                           "A record has a duplicate primary key",
		"unused_shape":                            "A shape is defined but never referenced by trips",
		"unused_route":                            "A route is defined but has no trips",
		"unreachable_stop":                        "A stop cannot be reached by any trip",
		"single_stop_zone":                        "A fare zone contains only a single stop",
		"unused_zone":                             "A fare zone is defined but not used in fare rules",
		"undefined_zone":                          "A fare rule references an undefined zone",
		"zone_id_same_as_stop_id":                 "A zone ID is the same as a stop ID",
		"long_zone_id":                            "A zone ID exceeds the recommended length",
		"overlapping_frequency":                   "Trip frequencies have overlapping time periods",
		"invalid_route_type":                      "A route type field contains an invalid route type",
		"missing_trip_edge":                       "A trip is missing required connections between stops",
		"decreasing_or_equal_shape_dist_traveled": "Shape distance values are not strictly increasing",
		"stop_too_far_from_trip_shape":            "A stop is too far from the trip's shape",
		"fast_travel_between_stops":               "Travel time between stops is unrealistically fast",
		"slow_travel_between_stops":               "Travel time between stops is unrealistically slow",
		"backwards_time_travel":                   "Departure time is before arrival time",
		"overlapping_trip_frequencies":            "Trip frequencies overlap in time",
		"feed_expiration_date":                    "The feed has expired or will expire soon",
	}

	if desc, exists := descriptions[code]; exists {
		return desc
	}

	// Generate a description from the code name
	words := strings.Split(code, "_")
	caser := cases.Title(language.English)
	for i, word := range words {
		words[i] = caser.String(word)
	}
	return strings.Join(words, " ")
}
