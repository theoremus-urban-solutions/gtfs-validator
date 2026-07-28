package notice

// The generic type layer's notices, as defined by the Canonical GTFS Schedule
// Validator. Every one of these describes a field whose value does not match
// the type the spec gives it, so they all carry the same four context keys:
// the file, the row, the field and the value as written.

// fieldContext is the context every type notice carries.
func fieldContext(filename string, rowNumber int, fieldName string, fieldValue string) map[string]interface{} {
	return map[string]interface{}{
		"filename":     filename,
		"csvRowNumber": rowNumber,
		"fieldName":    fieldName,
		"fieldValue":   fieldValue,
	}
}

// UnexpectedEnumValueNotice reports an enum field holding a value the spec
// does not define. It replaces the per-field invalid_* codes, which said the
// same thing once per field and emitted nothing at all when the value was not
// a number.
type UnexpectedEnumValueNotice struct {
	*BaseNotice
}

func NewUnexpectedEnumValueNotice(filename string, rowNumber int, fieldName string, fieldValue string) *UnexpectedEnumValueNotice {
	return &UnexpectedEnumValueNotice{
		BaseNotice: NewBaseNotice("unexpected_enum_value", WARNING, fieldContext(filename, rowNumber, fieldName, fieldValue)),
	}
}

// InvalidDateNotice reports a field that is not a YYYYMMDD date.
type InvalidDateNotice struct {
	*BaseNotice
}

func NewInvalidDateNotice(filename string, rowNumber int, fieldName string, fieldValue string) *InvalidDateNotice {
	return &InvalidDateNotice{
		BaseNotice: NewBaseNotice("invalid_date", ERROR, fieldContext(filename, rowNumber, fieldName, fieldValue)),
	}
}

// InvalidTimeNotice reports a field that is not an H:MM:SS, HH:MM:SS or
// HHH:MM:SS time. Hours past 24 are legal and mean the trip runs past midnight.
type InvalidTimeNotice struct {
	*BaseNotice
}

func NewInvalidTimeNotice(filename string, rowNumber int, fieldName string, fieldValue string) *InvalidTimeNotice {
	return &InvalidTimeNotice{
		BaseNotice: NewBaseNotice("invalid_time", ERROR, fieldContext(filename, rowNumber, fieldName, fieldValue)),
	}
}

// InvalidIntegerNotice reports a field that cannot be parsed as an integer.
type InvalidIntegerNotice struct {
	*BaseNotice
}

func NewInvalidIntegerNotice(filename string, rowNumber int, fieldName string, fieldValue string) *InvalidIntegerNotice {
	return &InvalidIntegerNotice{
		BaseNotice: NewBaseNotice("invalid_integer", ERROR, fieldContext(filename, rowNumber, fieldName, fieldValue)),
	}
}

// InvalidFloatNotice reports a field that cannot be parsed as a floating point
// number.
type InvalidFloatNotice struct {
	*BaseNotice
}

func NewInvalidFloatNotice(filename string, rowNumber int, fieldName string, fieldValue string) *InvalidFloatNotice {
	return &InvalidFloatNotice{
		BaseNotice: NewBaseNotice("invalid_float", ERROR, fieldContext(filename, rowNumber, fieldName, fieldValue)),
	}
}

// NumberOutOfRangeNotice reports a numeric field outside the range the spec
// allows for it — a latitude past the pole, a negative sequence, a headway of
// zero.
type NumberOutOfRangeNotice struct {
	*BaseNotice
}

func NewNumberOutOfRangeNotice(filename string, rowNumber int, fieldName string, fieldValue string, fieldType string) *NumberOutOfRangeNotice {
	context := fieldContext(filename, rowNumber, fieldName, fieldValue)
	context["fieldType"] = fieldType
	return &NumberOutOfRangeNotice{
		BaseNotice: NewBaseNotice("number_out_of_range", ERROR, context),
	}
}

// InvalidCurrencyNotice reports a currency code that is not ISO 4217.
type InvalidCurrencyNotice struct {
	*BaseNotice
}

func NewInvalidCurrencyNotice(filename string, rowNumber int, fieldName string, fieldValue string) *InvalidCurrencyNotice {
	return &InvalidCurrencyNotice{
		BaseNotice: NewBaseNotice("invalid_currency", ERROR, fieldContext(filename, rowNumber, fieldName, fieldValue)),
	}
}

// InvalidCurrencyAmountNotice reports an amount whose decimal places do not
// match what its currency uses — 1.5 USD where 1.50 is meant, or 100.00 JPY
// where the yen has no subunit.
type InvalidCurrencyAmountNotice struct {
	*BaseNotice
}

func NewInvalidCurrencyAmountNotice(filename string, rowNumber int, fieldName string, fieldValue string, currency string) *InvalidCurrencyAmountNotice {
	context := fieldContext(filename, rowNumber, fieldName, fieldValue)
	context["currency"] = currency
	return &InvalidCurrencyAmountNotice{
		BaseNotice: NewBaseNotice("invalid_currency_amount", ERROR, context),
	}
}

// InvalidPhoneNumberNotice reports a malformed phone number.
type InvalidPhoneNumberNotice struct {
	*BaseNotice
}

func NewInvalidPhoneNumberNotice(filename string, rowNumber int, fieldName string, fieldValue string) *InvalidPhoneNumberNotice {
	return &InvalidPhoneNumberNotice{
		BaseNotice: NewBaseNotice("invalid_phone_number", ERROR, fieldContext(filename, rowNumber, fieldName, fieldValue)),
	}
}

// InvalidCharacterNotice reports a value containing the Unicode replacement
// character, which means the file was decoded as something other than the
// UTF-8 the spec requires.
type InvalidCharacterNotice struct {
	*BaseNotice
}

func NewInvalidCharacterNotice(filename string, rowNumber int, fieldName string, fieldValue string) *InvalidCharacterNotice {
	return &InvalidCharacterNotice{
		BaseNotice: NewBaseNotice("invalid_character", ERROR, fieldContext(filename, rowNumber, fieldName, fieldValue)),
	}
}

// NewLineInValueNotice reports a newline inside a value, which usually means a
// quote was left unclosed and the next line was read as a continuation.
type NewLineInValueNotice struct {
	*BaseNotice
}

func NewNewLineInValueNotice(filename string, rowNumber int, fieldName string, fieldValue string) *NewLineInValueNotice {
	return &NewLineInValueNotice{
		BaseNotice: NewBaseNotice("new_line_in_value", ERROR, fieldContext(filename, rowNumber, fieldName, fieldValue)),
	}
}

// NonAsciiOrNonPrintableCharNotice reports an ID containing characters outside
// printable ASCII. IDs travel through URLs and other systems that may not
// carry them intact.
type NonAsciiOrNonPrintableCharNotice struct {
	*BaseNotice
}

func NewNonAsciiOrNonPrintableCharNotice(filename string, rowNumber int, fieldName string, fieldValue string) *NonAsciiOrNonPrintableCharNotice {
	return &NonAsciiOrNonPrintableCharNotice{
		BaseNotice: NewBaseNotice("non_ascii_or_non_printable_char", WARNING, fieldContext(filename, rowNumber, fieldName, fieldValue)),
	}
}

// EmptyColumnNameNotice reports a header cell with no name. Such a column
// cannot be referred to, so its values are skipped.
type EmptyColumnNameNotice struct {
	*BaseNotice
}

func NewEmptyColumnNameNotice(filename string, columnIndex int) *EmptyColumnNameNotice {
	context := map[string]interface{}{
		"filename": filename,
		"index":    columnIndex,
	}
	return &EmptyColumnNameNotice{
		BaseNotice: NewBaseNotice("empty_column_name", ERROR, context),
	}
}

// EmptyRowNotice reports a row holding nothing but whitespace.
type EmptyRowNotice struct {
	*BaseNotice
}

func NewEmptyRowNotice(filename string, rowNumber int) *EmptyRowNotice {
	context := map[string]interface{}{
		"filename":     filename,
		"csvRowNumber": rowNumber,
	}
	return &EmptyRowNotice{
		BaseNotice: NewBaseNotice("empty_row", WARNING, context),
	}
}

// StartAndEndRangeEqualNotice reports a range whose two ends are the same
// value, leaving what it covers undefined.
type StartAndEndRangeEqualNotice struct {
	*BaseNotice
}

func NewStartAndEndRangeEqualNotice(filename string, rowNumber int, entityID string, startFieldName string, endFieldName string, value string) *StartAndEndRangeEqualNotice {
	context := map[string]interface{}{
		"filename":       filename,
		"csvRowNumber":   rowNumber,
		"entityId":       entityID,
		"startFieldName": startFieldName,
		"endFieldName":   endFieldName,
		"value":          value,
	}
	return &StartAndEndRangeEqualNotice{
		BaseNotice: NewBaseNotice("start_and_end_range_equal", ERROR, context),
	}
}

// StartAndEndRangeOutOfOrderNotice reports a range that ends before it starts.
type StartAndEndRangeOutOfOrderNotice struct {
	*BaseNotice
}

func NewStartAndEndRangeOutOfOrderNotice(filename string, rowNumber int, entityID string, startFieldName string, startValue string, endFieldName string, endValue string) *StartAndEndRangeOutOfOrderNotice {
	context := map[string]interface{}{
		"filename":       filename,
		"csvRowNumber":   rowNumber,
		"entityId":       entityID,
		"startFieldName": startFieldName,
		"startValue":     startValue,
		"endFieldName":   endFieldName,
		"endValue":       endValue,
	}
	return &StartAndEndRangeOutOfOrderNotice{
		BaseNotice: NewBaseNotice("start_and_end_range_out_of_order", ERROR, context),
	}
}
