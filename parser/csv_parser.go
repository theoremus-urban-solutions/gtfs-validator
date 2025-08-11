package parser

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"

	"github.com/theoremus-urban-solutions/gtfs-validator/pools"
)

// CSVFile represents a parsed CSV file with headers and rows
type CSVFile struct {
	Filename   string
	Headers    []string
	rawHeaders int
	Rows       []CSVRow
	reader     *csv.Reader
	rowCounter int                    // Track the current row being read
	parser     *pools.PooledCSVParser // Memory-efficient CSV parsing
}

// CSVRow represents a single row in a CSV file
type CSVRow struct {
	RowNumber     int
	Values        map[string]string
	RawFieldCount int
}

// NewCSVFile creates a new CSV file parser
func NewCSVFile(reader io.Reader, filename string) (*CSVFile, error) {
	csvReader := csv.NewReader(reader)
	csvReader.LazyQuotes = true
	csvReader.TrimLeadingSpace = false // We'll handle whitespace validation

	// Read headers
	headers, err := csvReader.Read()
	if err != nil {
		if err == io.EOF {
			return nil, fmt.Errorf("empty file: %s", filename)
		}
		return nil, fmt.Errorf("failed to read headers from %s: %v", filename, err)
	}

	// Clean headers (remove BOM if present)
	if len(headers) > 0 {
		headers[0] = strings.TrimPrefix(headers[0], "\ufeff")
	}

	return &CSVFile{
		Filename:   filename,
		Headers:    headers,
		rawHeaders: len(headers),
		Rows:       make([]CSVRow, 0),
		reader:     csvReader,
		rowCounter: 1, // Start at 1 (header is row 1)
		parser:     pools.NewPooledCSVParser(),
	}, nil
}

// ReadRow reads the next row from the CSV file
func (f *CSVFile) ReadRow() (*CSVRow, error) {
	values, err := f.reader.Read()
	if err != nil {
		return nil, err
	}

	f.rowCounter++ // Increment counter for each data row; first data row should be row 2

	// Use pooled memory for the row values map
	rowValues := make(map[string]string)

	// Try to use pooled memory if field counts match
	if len(values) == len(f.Headers) {
		pooledMap := f.parser.ParseRecord(values, f.Headers)
		if pooledMap != nil {
			rowValues = pooledMap
		}
	}

	// Fill in the map (either pooled or regular)
	if len(rowValues) == 0 {
		// Manual mapping for mismatched field counts or pool failure
		for i, header := range f.Headers {
			if i < len(values) {
				rowValues[header] = values[i]
			} else {
				rowValues[header] = ""
			}
		}
	}

	row := &CSVRow{
		RowNumber:     f.rowCounter,
		Values:        rowValues,
		RawFieldCount: len(values),
	}

	return row, nil
}

// ReadAll reads all remaining rows from the CSV file
func (f *CSVFile) ReadAll() error {
	for {
		row, err := f.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading row %d: %v", len(f.Rows)+2, err)
		}
		f.Rows = append(f.Rows, *row)
	}
	return nil
}

// GetHeaderIndex returns the index of a header, or -1 if not found
func (f *CSVFile) GetHeaderIndex(header string) int {
	for i, h := range f.Headers {
		if h == header {
			return i
		}
	}
	return -1
}

// HasHeader returns true if the file has the specified header
func (f *CSVFile) HasHeader(header string) bool {
	return f.GetHeaderIndex(header) != -1
}

// RowCount returns the number of data rows (excluding header)
func (f *CSVFile) RowCount() int {
	return len(f.Rows)
}

// IsEmpty returns true if the file has no data rows
func (f *CSVFile) IsEmpty() bool {
	return len(f.Rows) == 0
}

// ReleaseRow returns the memory used by a CSVRow back to the pool
// This should be called when the row is no longer needed to reduce memory usage
func (f *CSVFile) ReleaseRow(row *CSVRow) {
	if row != nil && row.Values != nil {
		f.parser.ReturnRecord(row.Values)
		row.Values = nil // Clear reference to prevent accidental reuse
	}
}

// ReleaseAllRows returns all row memory back to the pool
// This should be called when the entire CSV file is no longer needed
func (f *CSVFile) ReleaseAllRows() {
	for i := range f.Rows {
		f.ReleaseRow(&f.Rows[i])
	}
}
