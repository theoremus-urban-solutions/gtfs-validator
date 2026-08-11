package core

import (
	"io"
	"log"
	"strconv"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/parser"
	"github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

// FieldTypeValidator checks every field against the type the spec gives it,
// driven by the table in field_types.go.
//
// One implementation replaces the per-field checks that used to live in each
// typed validator. It also closes the hole they shared: those checks were
// guarded by a successful strconv.Atoi, so an enum holding a non-numeric value
// fell through and reported nothing at all.
type FieldTypeValidator struct{}

// NewFieldTypeValidator creates a new field type validator
func NewFieldTypeValidator() *FieldTypeValidator {
	return &FieldTypeValidator{}
}

// Validate checks the typed fields of every file the table describes.
func (v *FieldTypeValidator) Validate(loader *parser.FeedLoader, container *notice.NoticeContainer, config validator.Config) {
	files := make([]string, 0, len(fieldSpecs))
	for filename := range fieldSpecs {
		if loader.HasFile(filename) {
			files = append(files, filename)
		}
	}

	workers := config.ParallelWorkers
	if workers > 1 && len(files) >= 4 {
		v.validateFilesParallel(loader, container, files, workers, config.CountryCode)
		return
	}
	for _, filename := range files {
		v.validateFile(loader, container, filename, config.CountryCode)
	}
}

// validateFilesParallel spreads the files across a worker pool.
func (v *FieldTypeValidator) validateFilesParallel(loader *parser.FeedLoader, container *notice.NoticeContainer, files []string, workers int, countryCode string) {
	fileChan := make(chan string, len(files))
	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for filename := range fileChan {
				v.validateFile(loader, container, filename, countryCode)
			}
		}()
	}

	for _, filename := range files {
		fileChan <- filename
	}
	close(fileChan)
	wg.Wait()
}

// validateFile walks one file.
func (v *FieldTypeValidator) validateFile(loader *parser.FeedLoader, container *notice.NoticeContainer, filename string, countryCode string) {
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

	for index, header := range csvFile.Headers {
		if strings.TrimSpace(header) == "" {
			container.AddNotice(notice.NewEmptyColumnNameNotice(filename, index))
		}
	}

	specs := fieldSpecs[filename]
	ranges := rangeSpecs[filename]

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		if isBlankRow(row) {
			container.AddNotice(notice.NewEmptyRowNotice(filename, row.RowNumber))
			continue
		}

		for field, value := range row.Values {
			v.validateEncoding(container, filename, row.RowNumber, field, value)
		}

		for i := range specs {
			v.validateField(container, filename, row, &specs[i], countryCode)
		}
		for _, r := range ranges {
			v.validateRange(container, filename, row, r)
		}
	}
}

// isBlankRow reports a row whose every cell is empty or whitespace. Such rows
// parse as a record but describe nothing.
func isBlankRow(row *parser.CSVRow) bool {
	for _, value := range row.Values {
		if strings.TrimSpace(value) != "" {
			return false
		}
	}
	return true
}

// validateEncoding reports values that survived decoding but should not have:
// a replacement character means the file was not the UTF-8 the spec requires,
// and an embedded newline means a quote was left open.
func (v *FieldTypeValidator) validateEncoding(container *notice.NoticeContainer, filename string, rowNumber int, field string, value string) {
	if value == "" {
		return
	}
	if strings.ContainsRune(value, utf8.RuneError) || strings.ContainsRune(value, '�') {
		container.AddNotice(notice.NewInvalidCharacterNotice(filename, rowNumber, field, value))
	}
	if strings.ContainsAny(value, "\n\r") {
		container.AddNotice(notice.NewNewLineInValueNotice(filename, rowNumber, field, value))
	}
}

// validateField checks one field of one row against its spec.
func (v *FieldTypeValidator) validateField(container *notice.NoticeContainer, filename string, row *parser.CSVRow, spec *fieldSpec, countryCode string) {
	raw, present := row.Values[spec.Name]
	if !present {
		return
	}
	value := strings.TrimSpace(raw)
	if value == "" {
		return // Absence is the required-field validator's business.
	}

	switch spec.Type {
	case typeID:
		if !isPrintableASCII(value) {
			container.AddNotice(notice.NewNonAsciiOrNonPrintableCharNotice(filename, row.RowNumber, spec.Name, value))
		}

	case typeEnum:
		// An enum is an integer first and a constrained set second, and the two
		// failures are different findings. A value that is not an integer at all
		// cannot be out of range — it has no range — so it is reported as a bad
		// integer, at ERROR, exactly as canonical does. Only a value that parses
		// and then falls outside the set is an unexpected enum, at WARNING.
		// Reporting a non-numeric enum as merely unexpected let a feed with
		// `friday=ZZZ` pass with no errors while canonical rejected it.
		if _, err := strconv.Atoi(value); err != nil {
			container.AddNotice(notice.NewInvalidIntegerNotice(filename, row.RowNumber, spec.Name, value))
		} else if !spec.admits(value) {
			container.AddNotice(notice.NewUnexpectedEnumValueNotice(filename, row.RowNumber, spec.Name, value))
		}

	case typeDate:
		if !isGTFSDate(value) {
			container.AddNotice(notice.NewInvalidDateNotice(filename, row.RowNumber, spec.Name, value))
		}

	case typeTime:
		if !isGTFSTime(value) {
			container.AddNotice(notice.NewInvalidTimeNotice(filename, row.RowNumber, spec.Name, value))
		}

	case typeInteger:
		parsed, err := strconv.Atoi(value)
		if err != nil {
			container.AddNotice(notice.NewInvalidIntegerNotice(filename, row.RowNumber, spec.Name, value))
			return
		}
		v.checkRange(container, filename, row.RowNumber, spec, value, float64(parsed), "integer")

	case typeFloat:
		parsed, err := strconv.ParseFloat(value, 64)
		if err != nil {
			container.AddNotice(notice.NewInvalidFloatNotice(filename, row.RowNumber, spec.Name, value))
			return
		}
		v.checkRange(container, filename, row.RowNumber, spec, value, parsed, "float")

	case typeCurrencyCode:
		if !isISO4217(value) {
			container.AddNotice(notice.NewInvalidCurrencyNotice(filename, row.RowNumber, spec.Name, value))
		}

	case typeCurrencyAmount:
		v.validateCurrencyAmount(container, filename, row, spec, value)

	case typePhone:
		// Without a country there is no numbering plan to measure against, and
		// canonical skips the check outright rather than guessing at one. A
		// length that is impossible in Bulgaria is ordinary in Austria, so a
		// country-free check could only reject numbers that dial fine
		// somewhere.
		if countryCode == "" {
			return
		}
		if !isPossiblePhoneNumber(value, countryCode) {
			container.AddNotice(notice.NewInvalidPhoneNumberNotice(filename, row.RowNumber, spec.Name, value))
		}
	}
}

// checkRange reports a number outside the bounds the spec gives its field.
func (v *FieldTypeValidator) checkRange(container *notice.NoticeContainer, filename string, rowNumber int, spec *fieldSpec, raw string, value float64, kind string) {
	if spec.HasMin && value < spec.Min {
		container.AddNotice(notice.NewNumberOutOfRangeNotice(filename, rowNumber, spec.Name, raw, kind))
		return
	}
	if spec.HasMax && value > spec.Max {
		container.AddNotice(notice.NewNumberOutOfRangeNotice(filename, rowNumber, spec.Name, raw, kind))
	}
}

// validateCurrencyAmount checks an amount's decimal places against the subunit
// its currency actually has: 1.5 USD is wrong where 1.50 is meant, and 100.00
// JPY is wrong because the yen has no subunit.
//
// This is the spec's Currency amount type, which only the Fares v2 amounts
// carry. A field the spec types as a plain float — fare_attributes.price — is
// not held to it, however its currency is written.
func (v *FieldTypeValidator) validateCurrencyAmount(container *notice.NoticeContainer, filename string, row *parser.CSVRow, spec *fieldSpec, value string) {
	if _, err := strconv.ParseFloat(value, 64); err != nil {
		container.AddNotice(notice.NewInvalidFloatNotice(filename, row.RowNumber, spec.Name, value))
		return
	}

	currency := strings.ToUpper(strings.TrimSpace(row.Values[spec.CurrencyField]))
	if currency == "" || !isISO4217(currency) {
		return // The currency code's own notice says what is wrong.
	}

	decimals := 0
	if dot := strings.Index(value, "."); dot >= 0 {
		decimals = len(value) - dot - 1
	}
	if decimals != currencyDecimals(currency) {
		container.AddNotice(notice.NewInvalidCurrencyAmountNotice(filename, row.RowNumber, spec.Name, value, currency))
	}
}

// validateRange checks a pair of fields describing the two ends of one range.
func (v *FieldTypeValidator) validateRange(container *notice.NoticeContainer, filename string, row *parser.CSVRow, spec rangeSpec) {
	start := strings.TrimSpace(row.Values[spec.Start])
	end := strings.TrimSpace(row.Values[spec.End])
	if start == "" || end == "" {
		return
	}

	var startValue, endValue int
	switch spec.Kind {
	case typeDate:
		if !isGTFSDate(start) || !isGTFSDate(end) {
			return // Reported as invalid_date.
		}
		startValue, _ = strconv.Atoi(start)
		endValue, _ = strconv.Atoi(end)
	case typeTime:
		var ok bool
		if startValue, ok = gtfsTimeSeconds(start); !ok {
			return
		}
		if endValue, ok = gtfsTimeSeconds(end); !ok {
			return
		}
	default:
		return
	}

	entityID := strings.TrimSpace(row.Values[spec.EntityField])

	switch {
	case endValue < startValue:
		container.AddNotice(notice.NewStartAndEndRangeOutOfOrderNotice(
			filename, row.RowNumber, entityID, spec.Start, start, spec.End, end,
		))
	case endValue == startValue && spec.EqualIsError:
		container.AddNotice(notice.NewStartAndEndRangeEqualNotice(
			filename, row.RowNumber, entityID, spec.Start, spec.End, start,
		))
	}
}

// admits reports whether an enum spec allows a value as written. The check is
// on the string, not on a parsed int, so a non-numeric value is reported
// rather than silently skipped.
func (s *fieldSpec) admits(value string) bool {
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return false
	}
	for _, allowed := range s.Values {
		if parsed == allowed {
			return true
		}
	}
	if s.ExtendedEnumMax > 0 && parsed >= s.ExtendedEnumMin && parsed <= s.ExtendedEnumMax {
		return true
	}
	return false
}

// isPrintableASCII reports whether every rune is printable ASCII.
func isPrintableASCII(value string) bool {
	for _, r := range value {
		if r > unicode.MaxASCII || !unicode.IsPrint(r) {
			return false
		}
	}
	return true
}

// isGTFSDate reports whether a value is a YYYYMMDD date.
func isGTFSDate(value string) bool {
	if len(value) != 8 {
		return false
	}
	for _, r := range value {
		if r < '0' || r > '9' {
			return false
		}
	}
	month, _ := strconv.Atoi(value[4:6])
	day, _ := strconv.Atoi(value[6:8])
	return month >= 1 && month <= 12 && day >= 1 && day <= 31
}

// gtfsTimeSeconds parses H:MM:SS, HH:MM:SS or HHH:MM:SS into seconds since
// midnight. Hours may exceed 24: a trip that departs at 25:10:00 leaves after
// midnight on the service day it belongs to.
func gtfsTimeSeconds(value string) (int, bool) {
	parts := strings.Split(value, ":")
	if len(parts) != 3 {
		return 0, false
	}
	if len(parts[0]) < 1 || len(parts[0]) > 3 || len(parts[1]) != 2 || len(parts[2]) != 2 {
		return 0, false
	}

	hours, err := strconv.Atoi(parts[0])
	if err != nil || hours < 0 {
		return 0, false
	}
	minutes, err := strconv.Atoi(parts[1])
	if err != nil || minutes < 0 || minutes > 59 {
		return 0, false
	}
	seconds, err := strconv.Atoi(parts[2])
	if err != nil || seconds < 0 || seconds > 59 {
		return 0, false
	}

	return hours*3600 + minutes*60 + seconds, true
}

// isGTFSTime reports whether a value is a GTFS time.
func isGTFSTime(value string) bool {
	_, ok := gtfsTimeSeconds(value)
	return ok
}
