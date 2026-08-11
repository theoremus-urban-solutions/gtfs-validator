package core

import (
	"bufio"
	"io"
	"log"
	"strings"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/parser"
	"github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

// LeadingTrailingWhitespaceValidator reports fields padded with leading or
// trailing whitespace. The padding is rarely visible to the author but is
// significant to consumers: an id with a trailing space does not match the same
// id without one, so the reference silently fails to resolve.
//
// Only *quoted* values are reported, and that is the whole rule. In CSV an
// unquoted field's surrounding whitespace is layout, not content — a reader is
// entitled to strip it, and the canonical validator does — so reporting it
// produces a warning about how the file was formatted rather than about what it
// says. Inside quotes the whitespace is asserted to be part of the value, and
// that is the mistake worth naming. Testing the field name instead, as this
// check used to, gets both halves wrong: it fires on unquoted padding canonical
// ignores, and stays silent on quoted padding in any field the hand-written
// list happened to omit.
type LeadingTrailingWhitespaceValidator struct{}

// NewLeadingTrailingWhitespaceValidator creates a new whitespace validator
func NewLeadingTrailingWhitespaceValidator() *LeadingTrailingWhitespaceValidator {
	return &LeadingTrailingWhitespaceValidator{}
}

// Validate checks for leading and trailing whitespace in GTFS fields
func (v *LeadingTrailingWhitespaceValidator) Validate(loader *parser.FeedLoader, container *notice.NoticeContainer, config validator.Config) {
	for _, filename := range loader.ListFiles() {
		v.validateFile(loader, container, filename)
	}
}

// quotedField is one CSV field together with whether it was written in quotes.
type quotedField struct {
	Value  string
	Quoted bool
}

// validateFile reports every quoted, whitespace-padded value in one file.
func (v *LeadingTrailingWhitespaceValidator) validateFile(loader *parser.FeedLoader, container *notice.NoticeContainer, filename string) {
	reader, err := loader.GetFile(filename)
	if err != nil {
		return
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close reader %v", closeErr)
		}
	}()

	records, err := readQuotedRecords(reader)
	if err != nil || len(records) == 0 {
		return
	}

	headers := make([]string, len(records[0]))
	for i, field := range records[0] {
		headers[i] = strings.TrimSpace(field.Value)
	}

	// Row 1 is the header; data rows start at 2, matching every other notice.
	for rowIndex, record := range records[1:] {
		for i, field := range record {
			if i >= len(headers) || !field.Quoted {
				continue
			}
			if !hasSurroundingWhitespace(field.Value) {
				continue
			}
			container.AddNotice(notice.NewLeadingOrTrailingWhitespacesNotice(
				filename, headers[i], field.Value, rowIndex+2,
			))
		}
	}
}

// hasSurroundingWhitespace reports whether a value begins or ends with space or
// tab. A value that is nothing but whitespace does both and is reported once.
func hasSurroundingWhitespace(value string) bool {
	if value == "" {
		return false
	}
	return value != strings.Trim(value, " \t")
}

// readQuotedRecords parses CSV while remembering which fields were quoted,
// which encoding/csv does not expose and which is the only thing this rule
// turns on. Quotes, escaped quotes and embedded newlines are handled as the
// CSV grammar requires; anything malformed is left to the parsing checks.
func readQuotedRecords(r io.Reader) ([][]quotedField, error) {
	br := bufio.NewReader(r)

	var (
		records   [][]quotedField
		record    []quotedField
		value     strings.Builder
		quoted    bool // this field was opened with a quote
		inQuotes  bool // currently inside a quoted section
		fieldSeen bool // something has been read towards the current field
	)

	endField := func() {
		record = append(record, quotedField{Value: value.String(), Quoted: quoted})
		value.Reset()
		quoted, fieldSeen = false, false
	}
	endRecord := func() {
		endField()
		records = append(records, record)
		record = nil
	}

	for {
		c, _, err := br.ReadRune()
		if err != nil {
			break
		}

		switch {
		case inQuotes:
			if c != '"' {
				value.WriteRune(c)
				continue
			}
			// A doubled quote is a literal quote; a single one closes the field.
			next, _, peekErr := br.ReadRune()
			if peekErr == nil && next == '"' {
				value.WriteRune('"')
				continue
			}
			if peekErr == nil {
				_ = br.UnreadRune()
			}
			inQuotes = false

		case c == '"' && !fieldSeen:
			inQuotes, quoted, fieldSeen = true, true, true

		case c == ',':
			endField()

		case c == '\n':
			endRecord()

		case c == '\r':
			// Consumed with the newline that follows it.

		default:
			value.WriteRune(c)
			fieldSeen = true
		}
	}

	// A final line with no trailing newline still holds a record.
	if value.Len() > 0 || len(record) > 0 {
		endRecord()
	}

	return records, nil
}
