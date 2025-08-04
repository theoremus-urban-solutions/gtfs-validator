package schema

// Translation represents a translation from translations.txt
type Translation struct {
	TableName   string `csv:"table_name"`
	FieldName   string `csv:"field_name"`
	Language    string `csv:"language"`
	Translation string `csv:"translation"`
	RecordID    string `csv:"record_id"`
	RecordSubID string `csv:"record_sub_id"`
	FieldValue  string `csv:"field_value"`
}