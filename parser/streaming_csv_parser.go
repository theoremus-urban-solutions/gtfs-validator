package parser

import (
	"bufio"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/theoremus-urban-solutions/gtfs-validator/pools"
)

// StreamingCSVParser provides streaming CSV parsing for large files
// without loading all data into memory at once
type StreamingCSVParser struct {
	filename     string
	headers      []string
	reader       *csv.Reader
	parser       *pools.PooledCSVParser
	rowCounter   int
	bufferSize   int
	lastReadTime time.Time
}

// StreamingCSVOptions configures the streaming CSV parser
type StreamingCSVOptions struct {
	// BufferSize sets the buffer size for reading (default: 64KB)
	BufferSize int

	// LazyQuotes allows lazy quote handling in CSV
	LazyQuotes bool

	// TrimLeadingSpace trims leading space in CSV fields
	TrimLeadingSpace bool
}

// DefaultStreamingCSVOptions returns sensible defaults for streaming CSV parsing
func DefaultStreamingCSVOptions() *StreamingCSVOptions {
	return &StreamingCSVOptions{
		BufferSize:       64 * 1024, // 64KB buffer
		LazyQuotes:       true,
		TrimLeadingSpace: false,
	}
}

// NewStreamingCSVParser creates a new streaming CSV parser
func NewStreamingCSVParser(reader io.Reader, filename string, opts *StreamingCSVOptions) (*StreamingCSVParser, error) {
	if opts == nil {
		opts = DefaultStreamingCSVOptions()
	}

	// Create buffered reader for better performance
	bufferedReader := bufio.NewReaderSize(reader, opts.BufferSize)

	csvReader := csv.NewReader(bufferedReader)
	csvReader.LazyQuotes = opts.LazyQuotes
	csvReader.TrimLeadingSpace = opts.TrimLeadingSpace

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

	return &StreamingCSVParser{
		filename:     filename,
		headers:      headers,
		reader:       csvReader,
		parser:       pools.NewPooledCSVParser(),
		rowCounter:   1, // Start at 1 (header is row 1)
		bufferSize:   opts.BufferSize,
		lastReadTime: time.Now(),
	}, nil
}

// Headers returns the CSV headers
func (s *StreamingCSVParser) Headers() []string {
	return s.headers
}

// Filename returns the filename being parsed
func (s *StreamingCSVParser) Filename() string {
	return s.filename
}

// RowCounter returns the current row number
func (s *StreamingCSVParser) RowCounter() int {
	return s.rowCounter
}

// ReadRow reads the next row from the CSV stream
// The returned CSVRow uses pooled memory and should be released with ReleaseRow()
func (s *StreamingCSVParser) ReadRow() (*CSVRow, error) {
	values, err := s.reader.Read()
	if err != nil {
		return nil, err
	}

	s.rowCounter++
	s.lastReadTime = time.Now()

	// Use pooled memory for the row values map
	rowValues := s.parser.ParseRecord(values, s.headers)
	if rowValues == nil {
		// Fallback to regular map creation if ParseRecord fails
		rowValues = make(map[string]string)
		for i, header := range s.headers {
			if i < len(values) {
				rowValues[header] = values[i]
			} else {
				rowValues[header] = ""
			}
		}
	}

	row := &CSVRow{
		RowNumber:     s.rowCounter,
		Values:        rowValues,
		RawFieldCount: len(values),
	}

	return row, nil
}

// ReleaseRow returns the memory used by a CSVRow back to the pool
func (s *StreamingCSVParser) ReleaseRow(row *CSVRow) {
	if row != nil && row.Values != nil {
		s.parser.ReturnRecord(row.Values)
		row.Values = nil
	}
}

// StreamingCSVProcessor defines the interface for processing streaming CSV rows
type StreamingCSVProcessor interface {
	// ProcessRow processes a single CSV row
	// The row will be automatically released after this method returns
	ProcessRow(row *CSVRow) error

	// ProcessingComplete is called when all rows have been processed
	ProcessingComplete() error
}

// StreamingCSVBatch represents a batch of CSV rows for batch processing
type StreamingCSVBatch struct {
	Rows      []*CSVRow
	BatchSize int
	StartRow  int
	EndRow    int
}

// StreamingCSVBatchProcessor processes CSV rows in batches
type StreamingCSVBatchProcessor interface {
	// ProcessBatch processes a batch of CSV rows
	// All rows in the batch will be automatically released after this method returns
	ProcessBatch(batch *StreamingCSVBatch) error

	// ProcessingComplete is called when all batches have been processed
	ProcessingComplete() error
}

// ProcessStream processes the entire CSV stream using the given processor
func (s *StreamingCSVParser) ProcessStream(ctx context.Context, processor StreamingCSVProcessor) error {
	defer func() {
		// Ensure processor cleanup is called
		if err := processor.ProcessingComplete(); err != nil {
			// Log error but don't return it as main processing might have succeeded
			fmt.Printf("Warning: ProcessingComplete failed: %v\n", err)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		row, err := s.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading row %d: %v", s.rowCounter, err)
		}

		// Process the row
		if err := processor.ProcessRow(row); err != nil {
			s.ReleaseRow(row) // Make sure to release on error
			return fmt.Errorf("error processing row %d: %v", row.RowNumber, err)
		}

		// Release the row memory back to pool
		s.ReleaseRow(row)
	}

	return nil
}

// ProcessStreamInBatches processes the CSV stream in batches using the given processor
func (s *StreamingCSVParser) ProcessStreamInBatches(ctx context.Context, processor StreamingCSVBatchProcessor, batchSize int) error {
	if batchSize <= 0 {
		batchSize = 1000 // Default batch size
	}

	defer func() {
		// Ensure processor cleanup is called
		if err := processor.ProcessingComplete(); err != nil {
			// Log error but don't return it as main processing might have succeeded
			fmt.Printf("Warning: ProcessingComplete failed: %v\n", err)
		}
	}()

	batch := &StreamingCSVBatch{
		Rows:      make([]*CSVRow, 0, batchSize),
		BatchSize: batchSize,
	}

	for {
		select {
		case <-ctx.Done():
			// Process any remaining rows in the current batch
			if len(batch.Rows) > 0 {
				batch.EndRow = batch.Rows[len(batch.Rows)-1].RowNumber
				if err := processor.ProcessBatch(batch); err != nil {
					s.releaseBatchRows(batch)
					return fmt.Errorf("error processing final batch: %v", err)
				}
				s.releaseBatchRows(batch)
			}
			return ctx.Err()
		default:
		}

		row, err := s.ReadRow()
		if err == io.EOF {
			// Process any remaining rows in the current batch
			if len(batch.Rows) > 0 {
				batch.EndRow = batch.Rows[len(batch.Rows)-1].RowNumber
				if err := processor.ProcessBatch(batch); err != nil {
					s.releaseBatchRows(batch)
					return fmt.Errorf("error processing final batch: %v", err)
				}
				s.releaseBatchRows(batch)
			}
			break
		}
		if err != nil {
			s.releaseBatchRows(batch)
			return fmt.Errorf("error reading row %d: %v", s.rowCounter, err)
		}

		// Add row to current batch
		if len(batch.Rows) == 0 {
			batch.StartRow = row.RowNumber
		}
		batch.Rows = append(batch.Rows, row)

		// Process batch when it's full
		if len(batch.Rows) >= batchSize {
			batch.EndRow = row.RowNumber
			if err := processor.ProcessBatch(batch); err != nil {
				s.releaseBatchRows(batch)
				return fmt.Errorf("error processing batch (rows %d-%d): %v", batch.StartRow, batch.EndRow, err)
			}

			// Release all rows in the batch and reset
			s.releaseBatchRows(batch)
			batch.Rows = batch.Rows[:0] // Reset slice but keep capacity
		}
	}

	return nil
}

// releaseBatchRows releases all rows in a batch back to the memory pool
func (s *StreamingCSVParser) releaseBatchRows(batch *StreamingCSVBatch) {
	for _, row := range batch.Rows {
		s.ReleaseRow(row)
	}
}

// Statistics returns parsing statistics
func (s *StreamingCSVParser) Statistics() StreamingCSVStatistics {
	return StreamingCSVStatistics{
		RowsProcessed: s.rowCounter - 1, // Subtract 1 for header row
		HeaderCount:   len(s.headers),
		LastReadTime:  s.lastReadTime,
		BufferSize:    s.bufferSize,
	}
}

// StreamingCSVStatistics contains statistics about the streaming parsing
type StreamingCSVStatistics struct {
	RowsProcessed int
	HeaderCount   int
	LastReadTime  time.Time
	BufferSize    int
}

// String returns a string representation of the statistics
func (s StreamingCSVStatistics) String() string {
	return fmt.Sprintf("StreamingCSVStatistics{RowsProcessed: %d, HeaderCount: %d, BufferSize: %d, LastReadTime: %s}",
		s.RowsProcessed, s.HeaderCount, s.BufferSize, s.LastReadTime.Format(time.RFC3339))
}

// Example processors

// CountingProcessor counts rows without storing them in memory
type CountingProcessor struct {
	Count int
}

// ProcessRow implements StreamingCSVProcessor
func (c *CountingProcessor) ProcessRow(row *CSVRow) error {
	c.Count++
	return nil
}

// ProcessingComplete implements StreamingCSVProcessor
func (c *CountingProcessor) ProcessingComplete() error {
	return nil
}

// ValidatingProcessor validates rows as they are streamed
type ValidatingProcessor struct {
	ErrorCount     int
	WarningCount   int
	RequiredFields []string
}

// ProcessRow implements StreamingCSVProcessor
func (v *ValidatingProcessor) ProcessRow(row *CSVRow) error {
	// Example validation: check required fields are present
	for _, field := range v.RequiredFields {
		if value, exists := row.Values[field]; !exists || strings.TrimSpace(value) == "" {
			v.ErrorCount++
			return fmt.Errorf("required field '%s' is missing or empty in row %d", field, row.RowNumber)
		}
	}
	return nil
}

// ProcessingComplete implements StreamingCSVProcessor
func (v *ValidatingProcessor) ProcessingComplete() error {
	return nil
}

// FilteringProcessor filters and processes only matching rows
type FilteringProcessor struct {
	FilterFunc    func(row *CSVRow) bool
	Processor     StreamingCSVProcessor
	ProcessedRows int
	FilteredRows  int
}

// ProcessRow implements StreamingCSVProcessor
func (f *FilteringProcessor) ProcessRow(row *CSVRow) error {
	if f.FilterFunc(row) {
		f.ProcessedRows++
		return f.Processor.ProcessRow(row)
	}
	f.FilteredRows++
	return nil
}

// ProcessingComplete implements StreamingCSVProcessor
func (f *FilteringProcessor) ProcessingComplete() error {
	return f.Processor.ProcessingComplete()
}
