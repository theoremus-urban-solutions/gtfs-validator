package entity

import (
	"io"
	"log"
	"strings"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/parser"
	"github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

// MixedCaseNameValidator applies the Mixed Case rule to the rider-facing text
// that lives outside agency.txt, routes.txt and stops.txt.
//
// Those three files are already walked for other reasons — by the agency, route
// and stop name validators — so their fields are checked in place rather than
// read a second time here. Everything else the rule covers has no other reader,
// and that is what this validator is for. trips.txt dominates the cost: it is
// the largest file the rule touches, and trip_headsign is by some distance the
// field producers most often shout.
type MixedCaseNameValidator struct{}

// NewMixedCaseNameValidator creates a new mixed case name validator
func NewMixedCaseNameValidator() *MixedCaseNameValidator {
	return &MixedCaseNameValidator{}
}

// mixedCaseFile names one file and the fields in it that carry text a rider
// reads. The set is not ours to choose: it is the field set the canonical
// validator annotates, and a field missing here is a notice a producer gets
// upstream and not from us.
type mixedCaseFile struct {
	filename string
	fields   []string
}

var mixedCaseFiles = []mixedCaseFile{
	{"trips.txt", []string{"trip_headsign", "trip_short_name"}},
	{"levels.txt", []string{"level_name"}},
	{"pathways.txt", []string{"signposted_as", "reversed_signposted_as"}},
	{"networks.txt", []string{"network_name"}},
	{"location_groups.txt", []string{"location_group_name"}},
	{"booking_rules.txt", []string{"message", "pickup_message", "drop_off_message"}},
}

// Validate reports single-case text in each file the rule covers.
func (v *MixedCaseNameValidator) Validate(loader *parser.FeedLoader, container *notice.NoticeContainer, config validator.Config) {
	for _, file := range mixedCaseFiles {
		v.validateFile(loader, container, file)
	}
}

// validateFile makes one streaming pass over a file. Nothing here needs a row
// to be remembered, so trips.txt costs a read and no memory.
func (v *MixedCaseNameValidator) validateFile(loader *parser.FeedLoader, container *notice.NoticeContainer, file mixedCaseFile) {
	reader, err := loader.GetFile(file.filename)
	if err != nil {
		return // Optional file, absent from this feed
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close reader %v", closeErr)
		}
	}()

	csvFile, err := parser.NewCSVFile(reader, file.filename)
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

		reportMixedCase(container, file.filename, row, file.fields)
	}
}

// reportMixedCase reports every field of a row that a rider reads and that was
// written in a single case.
func reportMixedCase(container *notice.NoticeContainer, filename string, row *parser.CSVRow, fields []string) {
	for _, field := range fields {
		value := strings.TrimSpace(row.Values[field])
		if value == "" {
			continue
		}
		if needsMixedCase(value) {
			container.AddNotice(notice.NewMixedCaseRecommendedFieldNotice(
				filename,
				field,
				value,
				row.RowNumber,
			))
		}
	}
}
