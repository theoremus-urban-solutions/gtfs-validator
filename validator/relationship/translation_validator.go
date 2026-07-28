package relationship

import (
	"io"
	"log"
	"strings"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/parser"
	"github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

// TranslationValidator checks translations.txt against the table it points at:
// that the table can be translated at all, that the row selects its target the
// one way the spec allows, and that the target exists.
type TranslationValidator struct{}

// NewTranslationValidator creates a new translation validator
func NewTranslationValidator() *TranslationValidator {
	return &TranslationValidator{}
}

// translatableTables maps each table_name the spec permits to its file and the
// column holding its primary key. feed_info has a single row and therefore no
// key, which is why it is the one table whose translations may not name a
// record.
var translatableTables = map[string]struct {
	filename string
	keyField string
}{
	"agency":       {"agency.txt", "agency_id"},
	"stops":        {"stops.txt", "stop_id"},
	"routes":       {"routes.txt", "route_id"},
	"trips":        {"trips.txt", "trip_id"},
	"stop_times":   {"stop_times.txt", "trip_id"},
	"pathways":     {"pathways.txt", "pathway_id"},
	"levels":       {"levels.txt", "level_id"},
	"feed_info":    {"feed_info.txt", ""},
	"attributions": {"attributions.txt", "attribution_id"},
}

// translationRef is one translations row's reference to a record, kept so the
// violation can be reported against the row that made it.
type translationRef struct {
	RowNumber   int
	RecordID    string
	RecordSubID string
}

// Validate checks the translations.txt rows and their references.
func (v *TranslationValidator) Validate(loader *parser.FeedLoader, container *notice.NoticeContainer, config validator.Config) {
	reader, err := loader.GetFile("translations.txt")
	if err != nil {
		return
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close reader %v", closeErr)
		}
	}()

	csvFile, err := parser.NewCSVFile(reader, "translations.txt")
	if err != nil {
		return
	}

	// References are collected rather than resolved as they are read: the
	// referenced tables are the large ones, and gathering first means each is
	// streamed once no matter how many translations point into it.
	refs := make(map[string][]translationRef)

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		field := func(name string) string {
			return strings.TrimSpace(row.Values[name])
		}
		tableName := field("table_name")
		if tableName == "" {
			continue // Reported as a missing required field by the field layer.
		}

		table, known := translatableTables[tableName]
		if !known || !loader.HasFile(table.filename) {
			container.AddNotice(notice.NewTranslationUnknownTableNameNotice(
				row.RowNumber, tableName,
			))
			continue
		}

		recordID := field("record_id")
		recordSubID := field("record_sub_id")
		fieldValue := field("field_value")

		if !v.validateSelection(container, row.RowNumber, tableName, recordID, recordSubID, fieldValue) {
			continue
		}
		if recordID == "" {
			continue // Selected by field_value, so there is no key to resolve.
		}

		refs[tableName] = append(refs[tableName], translationRef{
			RowNumber:   row.RowNumber,
			RecordID:    recordID,
			RecordSubID: recordSubID,
		})
	}

	v.validateReferences(loader, container, refs)
}

// validateSelection checks that the row picks its target either by record id
// or by field value, and reports each field carrying a value the spec forbids.
// It returns false when the row is malformed enough that resolving its
// reference would be meaningless.
func (v *TranslationValidator) validateSelection(container *notice.NoticeContainer, rowNumber int, tableName string, recordID string, recordSubID string, fieldValue string) bool {
	// feed_info holds one row, so naming a record within it is never valid.
	if tableName == "feed_info" {
		ok := true
		for _, forbidden := range []struct {
			name  string
			value string
		}{
			{"record_id", recordID},
			{"record_sub_id", recordSubID},
			{"field_value", fieldValue},
		} {
			if forbidden.value == "" {
				continue
			}
			container.AddNotice(notice.NewTranslationUnexpectedValueNotice(
				rowNumber, forbidden.name, forbidden.value,
			))
			ok = false
		}
		return ok
	}

	valid := true
	// record_id and field_value are two ways of naming the same thing, and a
	// row giving both leaves a consumer to guess which one it meant.
	if recordID != "" && fieldValue != "" {
		container.AddNotice(notice.NewTranslationUnexpectedValueNotice(
			rowNumber, "field_value", fieldValue,
		))
		valid = false
	}
	if recordID == "" && recordSubID != "" {
		container.AddNotice(notice.NewTranslationUnexpectedValueNotice(
			rowNumber, "record_sub_id", recordSubID,
		))
		valid = false
	}
	return valid
}

// validateReferences streams each referenced table once and reports the
// references that were not matched by any of its rows.
func (v *TranslationValidator) validateReferences(loader *parser.FeedLoader, container *notice.NoticeContainer, refs map[string][]translationRef) {
	for tableName, tableRefs := range refs {
		table := translatableTables[tableName]
		if table.keyField == "" {
			continue
		}

		// stop_times is keyed by a trip and a position within it, so both parts
		// have to match; every other table is keyed by a single id.
		wanted := make(map[string][]translationRef, len(tableRefs))
		for _, ref := range tableRefs {
			key := ref.RecordID
			if tableName == "stop_times" {
				key = ref.RecordID + "\x00" + ref.RecordSubID
			}
			wanted[key] = append(wanted[key], ref)
		}

		v.consumeTable(loader, table.filename, func(row *parser.CSVRow) {
			key := strings.TrimSpace(row.Values[table.keyField])
			if tableName == "stop_times" {
				key += "\x00" + strings.TrimSpace(row.Values["stop_sequence"])
			}
			delete(wanted, key)
		})

		for _, unmatched := range wanted {
			for _, ref := range unmatched {
				container.AddNotice(notice.NewTranslationForeignKeyViolationNotice(
					ref.RowNumber, tableName, ref.RecordID, ref.RecordSubID,
				))
			}
		}
	}
}

// consumeTable calls visit for every row of the named file, if it can be read.
func (v *TranslationValidator) consumeTable(loader *parser.FeedLoader, filename string, visit func(row *parser.CSVRow)) {
	reader, err := loader.GetFile(filename)
	if err != nil {
		return
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close reader %v", closeErr)
		}
	}()

	csvFile, err := parser.NewCSVFile(reader, filename)
	if err != nil {
		return
	}
	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			return
		}
		if err != nil {
			continue
		}
		visit(row)
	}
}
