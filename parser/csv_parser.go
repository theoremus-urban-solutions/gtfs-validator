package parser

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"
)

// CSVFile represents a parsed CSV file with headers and rows
type CSVFile struct {
	Filename   string
	Headers    []string
	Rows       []CSVRow
	reader     *csv.Reader
	rowCounter int // Track the current row being read
}

// CSVRow represents a single row in a CSV file
type CSVRow struct {
	RowNumber int
	Values    map[string]string
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
		Rows:       make([]CSVRow, 0),
		reader:     csvReader,
		rowCounter: 1, // Start at 1 (header is row 1)
	}, nil
}

// ReadRow reads the next row from the CSV file
func (f *CSVFile) ReadRow() (*CSVRow, error) {
	values, err := f.reader.Read()
	if err != nil {
		return nil, err
	}

	f.rowCounter++ // Increment counter for each data row
	row := &CSVRow{
		RowNumber: f.rowCounter,
		Values:    make(map[string]string),
	}

	for i, header := range f.Headers {
		if i < len(values) {
			row.Values[header] = values[i]
		} else {
			row.Values[header] = ""
		}
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