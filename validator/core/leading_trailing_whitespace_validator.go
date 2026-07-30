package core

import (
	"bytes"
	"io"
	"log"
	"strings"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/parser"
	"github.com/theoremus-urban-solutions/gtfs-validator/schema"
	"github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

// LeadingTrailingWhitespaceValidator reports values that still carry
// whitespace at one end or the other once the CSV parser has done its work.
//
// The rule is narrower than it first looks. A CSV parser strips the whitespace
// around an unquoted value, so ` Metro ` written without quotes reaches every
// other validator as `Metro` and there is nothing left to report. Only
// whitespace inside double quotes survives parsing, and only that whitespace
// ends up in a stop name a rider sees or an identifier a join fails on. The
// canonical validator says as much in the rule's own description, and this
// check follows it: a value is reported when it was quoted and the quotes hold
// whitespace at either end.
//
// Every column the spec defines is checked, in every file it defines, because
// the rule is about the file's text rather than any particular field's
// meaning. Columns the spec does not define are left to unknown_column.
type LeadingTrailingWhitespaceValidator struct{}

// NewLeadingTrailingWhitespaceValidator creates a new whitespace validator
func NewLeadingTrailingWhitespaceValidator() *LeadingTrailingWhitespaceValidator {
	return &LeadingTrailingWhitespaceValidator{}
}

// Validate checks every file of the feed that the spec describes.
func (v *LeadingTrailingWhitespaceValidator) Validate(loader *parser.FeedLoader, container *notice.NoticeContainer, config validator.Config) {
	for _, filename := range loader.ListFiles() {
		columns, described := schema.KnownColumns(filename)
		if !described {
			continue
		}
		v.validateFile(loader, container, filename, columns)
	}
}

// validateFile makes one streaming pass over a file, reporting each quoted
// value that begins or ends in whitespace.
func (v *LeadingTrailingWhitespaceValidator) validateFile(loader *parser.FeedLoader, container *notice.NoticeContainer, filename string, columns []string) {
	reader, err := loader.GetFile(filename)
	if err != nil {
		return
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close reader %v", closeErr)
		}
	}()

	known := make(map[string]bool, len(columns))
	for _, column := range columns {
		known[column] = true
	}

	lines := newLineIndex(reader)
	csvFile, err := parser.NewCSVFile(lines, filename)
	if err != nil {
		return
	}

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}
		if row.RawFieldCount == 0 {
			continue
		}

		for i, header := range csvFile.Headers {
			if i >= row.RawFieldCount {
				break
			}
			if !known[header] {
				continue
			}
			value := row.Values[header]
			if value == trimSpace(value) {
				continue
			}
			if line, column := csvFile.FieldPos(i); !lines.quotedAt(line, column) {
				continue
			}
			container.AddNotice(notice.NewLeadingOrTrailingWhitespacesNotice(
				filename,
				header,
				value,
				row.RowNumber,
			))
		}

		// Nothing before this row will be asked about again.
		firstLine, _ := csvFile.FieldPos(0)
		lines.forget(firstLine)
	}
}

// trimSpace strips what the canonical validator strips: it compares a value
// against Java's String.trim, which removes every character at or below the
// space from both ends, rather than only the space and the tab.
func trimSpace(value string) string {
	return strings.TrimFunc(value, func(r rune) bool { return r <= ' ' })
}

// lineIndex sits between the file and the CSV parser and remembers the raw
// text of the lines the parser has not finished with. encoding/csv reports the
// line and column each field started at but not how it was written, and the
// byte at that position is a double quote exactly when the field was quoted.
//
// Only the current record's lines are kept — the parser reads a few kilobytes
// ahead and no further — so following a feed's largest file costs the same as
// following its smallest.
type lineIndex struct {
	src     io.Reader
	number  int            // line being filled, counting from 1
	partial []byte         // its bytes so far
	lines   map[int][]byte // complete lines not yet forgotten
}

func newLineIndex(src io.Reader) *lineIndex {
	return &lineIndex{src: src, number: 1, lines: make(map[int][]byte)}
}

// Read passes the bytes through, splitting them into lines on the way.
func (l *lineIndex) Read(p []byte) (int, error) {
	n, err := l.src.Read(p)

	chunk := p[:n]
	for {
		end := bytes.IndexByte(chunk, '\n')
		if end < 0 {
			l.partial = append(l.partial, chunk...)
			break
		}
		l.partial = append(l.partial, chunk[:end]...)
		l.lines[l.number] = l.partial
		l.number++
		l.partial = nil
		chunk = chunk[end+1:]
	}

	// A file whose last line has no newline still has that line.
	if err == io.EOF && len(l.partial) > 0 {
		l.lines[l.number] = l.partial
		l.partial = nil
	}

	return n, err
}

// quotedAt reports whether the field beginning at this position was written
// inside double quotes.
func (l *lineIndex) quotedAt(line int, column int) bool {
	text, held := l.lines[line]
	if !held || column < 1 || column > len(text) {
		return false
	}
	return text[column-1] == '"'
}

// forget drops the lines before the one given.
func (l *lineIndex) forget(before int) {
	for number := range l.lines {
		if number < before {
			delete(l.lines, number)
		}
	}
}
