package entity

import (
	"io"
	"log"
	"strings"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/parser"
	"github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

// AgencyConsistencyValidator validates agency references and consistency
type AgencyConsistencyValidator struct{}

// NewAgencyConsistencyValidator creates a new agency consistency validator
func NewAgencyConsistencyValidator() *AgencyConsistencyValidator {
	return &AgencyConsistencyValidator{}
}

// Validate checks agency consistency rules
func (v *AgencyConsistencyValidator) Validate(loader *parser.FeedLoader, container *notice.NoticeContainer, config validator.Config) {
	// Load agencies
	agencies := v.loadAgencies(loader)
	if len(agencies) == 0 {
		return // No agencies to validate
	}

	// Check if agency_id is required
	v.validateAgencyIdRequirement(container, agencies)

	// Every time in the feed is read in the agency timezone, so two of them
	// leave the schedule ambiguous.
	v.validateAgencyTimezones(container, agencies)

	// The agencies describe one body of text and should agree on its language,
	// as should feed_info.txt.
	v.validateAgencyLanguages(loader, container, agencies)

	// Check route agency references
	v.validateRouteAgencyReferences(loader, container, agencies)

	// fare_attributes.agency_id is required for the same reason routes.agency_id is.
	v.validateFareAgencyReferences(loader, container, agencies)
}

// AgencyInfo represents agency information
type AgencyInfo struct {
	AgencyID   string
	AgencyName string
	Timezone   string
	Language   string
	RowNumber  int
}

// loadAgencies loads agency information from agency.txt
func (v *AgencyConsistencyValidator) loadAgencies(loader *parser.FeedLoader) []*AgencyInfo {
	var agencies []*AgencyInfo

	reader, err := loader.GetFile("agency.txt")
	if err != nil {
		return agencies
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close reader %v", closeErr)
		}
	}()

	csvFile, err := parser.NewCSVFile(reader, "agency.txt")
	if err != nil {
		return agencies
	}

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}

		agencyID, hasAgencyID := row.Values["agency_id"]
		agencyName, hasAgencyName := row.Values["agency_name"]

		// Use empty string if no agency_id provided
		key := ""
		if hasAgencyID {
			key = strings.TrimSpace(agencyID)
		}

		name := ""
		if hasAgencyName {
			name = strings.TrimSpace(agencyName)
		}

		agencies = append(agencies, &AgencyInfo{
			AgencyID:   key,
			AgencyName: name,
			Timezone:   strings.TrimSpace(row.Values["agency_timezone"]),
			Language:   strings.TrimSpace(row.Values["agency_lang"]),
			RowNumber:  row.RowNumber,
		})
	}

	return agencies
}

// validateAgencyIdRequirement checks if agency_id is required
func (v *AgencyConsistencyValidator) validateAgencyIdRequirement(container *notice.NoticeContainer, agencies []*AgencyInfo) {
	// With one agency the reference is unambiguous, so agency_id may be
	// omitted. With more than one it cannot be.
	if len(agencies) <= 1 {
		return
	}
	for _, agency := range agencies {
		if agency.AgencyID == "" {
			container.AddNotice(notice.NewMissingRequiredAgencyIDNotice(
				"agency.txt",
				"agency_id",
				agency.RowNumber,
			))
		}
	}
}

// validateAgencyTimezones reports agencies that disagree about the timezone.
// The first agency in the file is taken as the expected one.
func (v *AgencyConsistencyValidator) validateAgencyTimezones(container *notice.NoticeContainer, agencies []*AgencyInfo) {
	expected := ""
	for _, agency := range agencies {
		if agency.Timezone == "" {
			continue
		}
		if expected == "" {
			expected = agency.Timezone
			continue
		}
		if agency.Timezone != expected {
			container.AddNotice(notice.NewInconsistentAgencyTimezoneNotice(
				expected,
				agency.Timezone,
				agency.RowNumber,
			))
		}
	}
}

// validateAgencyLanguages reports agencies that disagree about the language the
// feed is written in, and agencies that disagree with feed_info.feed_lang.
// The first agency declaring a language is taken as the expected one.
func (v *AgencyConsistencyValidator) validateAgencyLanguages(loader *parser.FeedLoader, container *notice.NoticeContainer, agencies []*AgencyInfo) {
	expected := ""
	for _, agency := range agencies {
		if agency.Language == "" {
			continue
		}
		if expected == "" {
			expected = agency.Language
			continue
		}
		if !sameLanguage(agency.Language, expected) {
			container.AddNotice(notice.NewInconsistentAgencyLangNotice(
				agency.RowNumber,
				expected,
				agency.Language,
			))
		}
	}

	feedLang := v.loadFeedLang(loader)
	if feedLang == "" {
		return
	}
	// "mul" declares the feed multilingual, so no single agency_lang can match
	// it and none of them is wrong for failing to.
	if primaryLanguageSubtag(feedLang) == "mul" {
		return
	}

	for _, agency := range agencies {
		if agency.Language == "" || sameLanguage(agency.Language, feedLang) {
			continue
		}
		container.AddNotice(notice.NewFeedInfoLangAndAgencyLangMismatchNotice(
			agency.RowNumber,
			agency.AgencyID,
			agency.AgencyName,
			agency.Language,
			feedLang,
		))
	}
}

// loadFeedLang returns feed_info.feed_lang, or an empty string when the file is
// absent or does not declare one. Only the first row is read; a feed_info.txt
// with more than one row is reported elsewhere.
func (v *AgencyConsistencyValidator) loadFeedLang(loader *parser.FeedLoader) string {
	reader, err := loader.GetFile("feed_info.txt")
	if err != nil {
		return ""
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close reader %v", closeErr)
		}
	}()

	csvFile, err := parser.NewCSVFile(reader, "feed_info.txt")
	if err != nil {
		return ""
	}

	row, err := csvFile.ReadRow()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(row.Values["feed_lang"])
}

// sameLanguage compares two language tags by the language they name. Tags are
// case-insensitive, and a region subtag ("en-GB" against "en") narrows a
// language rather than contradicting it.
func sameLanguage(a string, b string) bool {
	return primaryLanguageSubtag(a) == primaryLanguageSubtag(b)
}

// primaryLanguageSubtag returns the language part of a BCP-47 tag, lowercased.
func primaryLanguageSubtag(tag string) string {
	lowered := strings.ToLower(strings.TrimSpace(tag))
	if index := strings.IndexAny(lowered, "-_"); index >= 0 {
		return lowered[:index]
	}
	return lowered
}

// validateFareAgencyReferences reports fare_attributes rows omitting agency_id
// in a feed with more than one agency.
func (v *AgencyConsistencyValidator) validateFareAgencyReferences(loader *parser.FeedLoader, container *notice.NoticeContainer, agencies []*AgencyInfo) {
	if len(agencies) <= 1 {
		return
	}

	reader, err := loader.GetFile("fare_attributes.txt")
	if err != nil {
		return
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close reader %v", closeErr)
		}
	}()

	csvFile, err := parser.NewCSVFile(reader, "fare_attributes.txt")
	if err != nil {
		return
	}

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}
		if strings.TrimSpace(row.Values["agency_id"]) == "" {
			container.AddNotice(notice.NewMissingRequiredAgencyIDNotice(
				"fare_attributes.txt",
				"agency_id",
				row.RowNumber,
			))
		}
	}
}

// validateRouteAgencyReferences checks that routes reference valid agencies
func (v *AgencyConsistencyValidator) validateRouteAgencyReferences(loader *parser.FeedLoader, container *notice.NoticeContainer, agencies []*AgencyInfo) {
	reader, err := loader.GetFile("routes.txt")
	if err != nil {
		return // No routes file
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close reader %v", closeErr)
		}
	}()

	csvFile, err := parser.NewCSVFile(reader, "routes.txt")
	if err != nil {
		return
	}

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}

		_, hasRouteID := row.Values["route_id"]
		agencyID, hasAgencyID := row.Values["agency_id"]

		if !hasRouteID {
			continue
		}

		// Determine which agency this route should reference
		expectedAgencyID := ""
		if hasAgencyID {
			expectedAgencyID = strings.TrimSpace(agencyID)
		}

		// If no agency_id specified in route, use the single agency if there's only one
		if expectedAgencyID == "" && len(agencies) == 1 {
			// This is valid - routes can omit agency_id if there's only one agency
			continue
		}

		// If multiple agencies exist but route has no agency_id, that's an error
		if expectedAgencyID == "" && len(agencies) > 1 {
			container.AddNotice(notice.NewMissingRequiredAgencyIDNotice(
				"routes.txt",
				"agency_id",
				row.RowNumber,
			))
			continue
		}

		// A route naming an agency that agency.txt does not define is reported
		// by relationship/foreign_key_validator.go as foreign_key_violation.
	}
}
