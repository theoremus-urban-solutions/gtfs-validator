package parser

import (
	"context"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"
)

const (
	testStationName = "Test Station"
	testCSVSingle   = `id,name,type
123,Test Station,0`
	testCSVSmall = `id,name,type
123,Test Station,0
456,Another Station,1`
	testCSVLarge = `id,name,type
123,Test Station,0
456,Another Station,1
789,Third Station,2`
)

func TestNewStreamingCSVParser(t *testing.T) {
	csvData := testCSVSmall

	parser, err := NewStreamingCSVParser(strings.NewReader(csvData), "test.csv", nil)
	if err != nil {
		t.Fatalf("Failed to create streaming CSV parser: %v", err)
	}

	if parser.Filename() != "test.csv" {
		t.Errorf("Expected filename 'test.csv', got '%s'", parser.Filename())
	}

	headers := parser.Headers()
	if len(headers) != 3 {
		t.Errorf("Expected 3 headers, got %d", len(headers))
	}

	expectedHeaders := []string{"id", "name", "type"}
	for i, expected := range expectedHeaders {
		if headers[i] != expected {
			t.Errorf("Expected header[%d] = '%s', got '%s'", i, expected, headers[i])
		}
	}

	if parser.RowCounter() != 1 {
		t.Errorf("Expected initial row counter 1, got %d", parser.RowCounter())
	}
}

func TestStreamingCSVParserWithOptions(t *testing.T) {
	csvData := testCSVSingle

	opts := &StreamingCSVOptions{
		BufferSize:       32 * 1024,
		LazyQuotes:       false,
		TrimLeadingSpace: true,
	}

	parser, err := NewStreamingCSVParser(strings.NewReader(csvData), "test.csv", opts)
	if err != nil {
		t.Fatalf("Failed to create streaming CSV parser with options: %v", err)
	}

	if parser.bufferSize != opts.BufferSize {
		t.Errorf("Expected buffer size %d, got %d", opts.BufferSize, parser.bufferSize)
	}
}

func TestStreamingCSVParserReadRow(t *testing.T) {
	csvData := testCSVLarge

	parser, err := NewStreamingCSVParser(strings.NewReader(csvData), "test.csv", nil)
	if err != nil {
		t.Fatalf("Failed to create streaming CSV parser: %v", err)
	}

	// Read first row
	row1, err := parser.ReadRow()
	if err != nil {
		t.Fatalf("Failed to read first row: %v", err)
	}

	if row1.RowNumber != 2 {
		t.Errorf("Expected row number 2, got %d", row1.RowNumber)
	}
	if row1.Values["id"] != "123" {
		t.Errorf("Expected id=123, got %s", row1.Values["id"])
	}
	if row1.Values["name"] != testStationName {
		t.Errorf("Expected name=%s, got %s", testStationName, row1.Values["name"])
	}
	if row1.Values["type"] != "0" {
		t.Errorf("Expected type=0, got %s", row1.Values["type"])
	}

	parser.ReleaseRow(row1)

	// Read second row
	row2, err := parser.ReadRow()
	if err != nil {
		t.Fatalf("Failed to read second row: %v", err)
	}

	if row2.RowNumber != 3 {
		t.Errorf("Expected row number 3, got %d", row2.RowNumber)
	}
	if row2.Values["id"] != "456" {
		t.Errorf("Expected id=456, got %s", row2.Values["id"])
	}

	parser.ReleaseRow(row2)

	// Read third row
	row3, err := parser.ReadRow()
	if err != nil {
		t.Fatalf("Failed to read third row: %v", err)
	}

	if row3.RowNumber != 4 {
		t.Errorf("Expected row number 4, got %d", row3.RowNumber)
	}
	if row3.Values["id"] != "789" {
		t.Errorf("Expected id=789, got %s", row3.Values["id"])
	}

	parser.ReleaseRow(row3)

	// Try to read beyond end
	row4, err := parser.ReadRow()
	if err != io.EOF {
		t.Errorf("Expected EOF, got: %v", err)
	}
	if row4 != nil {
		t.Error("Expected nil row at EOF")
	}
}

func TestCountingProcessor(t *testing.T) {
	csvData := testCSVLarge

	parser, err := NewStreamingCSVParser(strings.NewReader(csvData), "test.csv", nil)
	if err != nil {
		t.Fatalf("Failed to create streaming CSV parser: %v", err)
	}

	processor := &CountingProcessor{}
	ctx := context.Background()

	err = parser.ProcessStream(ctx, processor)
	if err != nil {
		t.Fatalf("Failed to process stream: %v", err)
	}

	if processor.Count != 3 {
		t.Errorf("Expected count 3, got %d", processor.Count)
	}

	stats := parser.Statistics()
	if stats.RowsProcessed != 3 {
		t.Errorf("Expected 3 rows processed, got %d", stats.RowsProcessed)
	}
}

func TestValidatingProcessor(t *testing.T) {
	csvData := `id,name,type
123,Test Station,0
456,,1
789,Third Station,`

	parser, err := NewStreamingCSVParser(strings.NewReader(csvData), "test.csv", nil)
	if err != nil {
		t.Fatalf("Failed to create streaming CSV parser: %v", err)
	}

	processor := &ValidatingProcessor{
		RequiredFields: []string{"id", "name", "type"},
	}
	ctx := context.Background()

	err = parser.ProcessStream(ctx, processor)
	if err == nil {
		t.Error("Expected validation error, but got none")
	}

	// Should have found the first validation error (missing name in row 2)
	if !strings.Contains(err.Error(), "name") {
		t.Errorf("Expected error about missing name, got: %v", err)
	}
}

func TestBatchProcessor(t *testing.T) {
	// Create a larger dataset
	var csvData strings.Builder
	csvData.WriteString("id,name,type\n")
	for i := 1; i <= 25; i++ {
		csvData.WriteString(fmt.Sprintf("%d,Station %d,0\n", i, i))
	}

	parser, err := NewStreamingCSVParser(strings.NewReader(csvData.String()), "test.csv", nil)
	if err != nil {
		t.Fatalf("Failed to create streaming CSV parser: %v", err)
	}

	processor := &MockBatchProcessor{}
	ctx := context.Background()
	batchSize := 10

	err = parser.ProcessStreamInBatches(ctx, processor, batchSize)
	if err != nil {
		t.Fatalf("Failed to process stream in batches: %v", err)
	}

	// Should have processed 3 batches: 10 + 10 + 5
	if len(processor.Batches) != 3 {
		t.Errorf("Expected 3 batches, got %d", len(processor.Batches))
	}

	// First two batches should have 10 rows each
	for i := 0; i < 2; i++ {
		if len(processor.Batches[i].Rows) != 10 {
			t.Errorf("Expected batch %d to have 10 rows, got %d", i, len(processor.Batches[i].Rows))
		}
	}

	// Last batch should have 5 rows
	if len(processor.Batches[2].Rows) != 5 {
		t.Errorf("Expected last batch to have 5 rows, got %d", len(processor.Batches[2].Rows))
	}

	// Check total rows processed
	if processor.TotalRows != 25 {
		t.Errorf("Expected 25 total rows processed, got %d", processor.TotalRows)
	}
}

func TestFilteringProcessor(t *testing.T) {
	csvData := `id,name,type
123,Test Station,0
456,Another Station,1
789,Third Station,0
101,Fourth Station,1`

	parser, err := NewStreamingCSVParser(strings.NewReader(csvData), "test.csv", nil)
	if err != nil {
		t.Fatalf("Failed to create streaming CSV parser: %v", err)
	}

	// Filter for type "0" only
	filterFunc := func(row *CSVRow) bool {
		return row.Values["type"] == "0"
	}

	counter := &CountingProcessor{}
	processor := &FilteringProcessor{
		FilterFunc: filterFunc,
		Processor:  counter,
	}

	ctx := context.Background()
	err = parser.ProcessStream(ctx, processor)
	if err != nil {
		t.Fatalf("Failed to process filtered stream: %v", err)
	}

	// Should process 2 rows (type "0") and filter 2 rows (type "1")
	if processor.ProcessedRows != 2 {
		t.Errorf("Expected 2 processed rows, got %d", processor.ProcessedRows)
	}
	if processor.FilteredRows != 2 {
		t.Errorf("Expected 2 filtered rows, got %d", processor.FilteredRows)
	}
	if counter.Count != 2 {
		t.Errorf("Expected counter to have 2, got %d", counter.Count)
	}
}

func TestStreamingCSVParserEmptyFile(t *testing.T) {
	_, err := NewStreamingCSVParser(strings.NewReader(""), "empty.csv", nil)
	if err == nil {
		t.Error("Expected error for empty file")
	}
	if !strings.Contains(err.Error(), "empty file") {
		t.Errorf("Expected empty file error, got: %v", err)
	}
}

func TestStreamingCSVParserBOM(t *testing.T) {
	csvData := "\ufeffid,name,type\n123,Test Station,0"

	parser, err := NewStreamingCSVParser(strings.NewReader(csvData), "test.csv", nil)
	if err != nil {
		t.Fatalf("Failed to create streaming CSV parser: %v", err)
	}

	headers := parser.Headers()
	if headers[0] != "id" {
		t.Errorf("Expected first header to be 'id' after BOM removal, got '%s'", headers[0])
	}
}

func TestStreamingCSVParserContextCancellation(t *testing.T) {
	// Create a large dataset
	var csvData strings.Builder
	csvData.WriteString("id,name,type\n")
	for i := 1; i <= 10000; i++ {
		csvData.WriteString(fmt.Sprintf("%d,Station %d,0\n", i, i))
	}

	parser, err := NewStreamingCSVParser(strings.NewReader(csvData.String()), "test.csv", nil)
	if err != nil {
		t.Fatalf("Failed to create streaming CSV parser: %v", err)
	}

	processor := &SlowCountingProcessor{delay: 1 * time.Millisecond}
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err = parser.ProcessStream(ctx, processor)
	if err != context.DeadlineExceeded {
		t.Errorf("Expected context deadline exceeded, got: %v", err)
	}

	// Should have processed some rows but not all
	if processor.Count >= 10000 {
		t.Error("Should not have processed all rows due to timeout")
	}
	if processor.Count == 0 {
		t.Error("Should have processed some rows before timeout")
	}
}

func TestDefaultStreamingCSVOptions(t *testing.T) {
	opts := DefaultStreamingCSVOptions()

	if opts.BufferSize != 64*1024 {
		t.Errorf("Expected default buffer size 64KB, got %d", opts.BufferSize)
	}
	if !opts.LazyQuotes {
		t.Error("Expected default LazyQuotes to be true")
	}
	if opts.TrimLeadingSpace {
		t.Error("Expected default TrimLeadingSpace to be false")
	}
}

func TestStreamingCSVStatistics(t *testing.T) {
	csvData := testCSVSingle

	parser, err := NewStreamingCSVParser(strings.NewReader(csvData), "test.csv", nil)
	if err != nil {
		t.Fatalf("Failed to create streaming CSV parser: %v", err)
	}

	// Read a row
	row, err := parser.ReadRow()
	if err != nil {
		t.Fatalf("Failed to read row: %v", err)
	}
	parser.ReleaseRow(row)

	stats := parser.Statistics()
	if stats.RowsProcessed != 1 {
		t.Errorf("Expected 1 row processed, got %d", stats.RowsProcessed)
	}
	if stats.HeaderCount != 3 {
		t.Errorf("Expected 3 headers, got %d", stats.HeaderCount)
	}
	if stats.BufferSize != 64*1024 {
		t.Errorf("Expected buffer size 64KB, got %d", stats.BufferSize)
	}

	// Test string representation
	statsStr := stats.String()
	if !strings.Contains(statsStr, "RowsProcessed: 1") {
		t.Errorf("Expected stats string to contain row count, got: %s", statsStr)
	}
}

// Test helper processors

type MockBatchProcessor struct {
	Batches   []*StreamingCSVBatch
	TotalRows int
}

func (p *MockBatchProcessor) ProcessBatch(batch *StreamingCSVBatch) error {
	// Make a copy of the batch to store (since original will be reused)
	batchCopy := &StreamingCSVBatch{
		Rows:      make([]*CSVRow, len(batch.Rows)),
		BatchSize: batch.BatchSize,
		StartRow:  batch.StartRow,
		EndRow:    batch.EndRow,
	}

	// Copy row data (but don't copy Values maps since they're pooled)
	for i, row := range batch.Rows {
		batchCopy.Rows[i] = &CSVRow{
			RowNumber:     row.RowNumber,
			RawFieldCount: row.RawFieldCount,
			// Don't copy Values since it's pooled memory
		}
	}

	p.Batches = append(p.Batches, batchCopy)
	p.TotalRows += len(batch.Rows)
	return nil
}

func (p *MockBatchProcessor) ProcessingComplete() error {
	return nil
}

type SlowCountingProcessor struct {
	CountingProcessor
	delay time.Duration
}

func (s *SlowCountingProcessor) ProcessRow(row *CSVRow) error {
	time.Sleep(s.delay)
	return s.CountingProcessor.ProcessRow(row)
}

// Benchmarks

func BenchmarkStreamingCSVParser(b *testing.B) {
	// Create CSV data
	var csvData strings.Builder
	csvData.WriteString("id,name,type,lat,lon\n")
	for i := 0; i < 10000; i++ {
		csvData.WriteString(fmt.Sprintf("%d,Station %d,0,37.7749,-122.4194\n", i, i))
	}

	data := csvData.String()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parser, err := NewStreamingCSVParser(strings.NewReader(data), "bench.csv", nil)
		if err != nil {
			b.Fatal(err)
		}

		processor := &CountingProcessor{}
		ctx := context.Background()

		err = parser.ProcessStream(ctx, processor)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkStreamingCSVParserBatches(b *testing.B) {
	// Create CSV data
	var csvData strings.Builder
	csvData.WriteString("id,name,type,lat,lon\n")
	for i := 0; i < 10000; i++ {
		csvData.WriteString(fmt.Sprintf("%d,Station %d,0,37.7749,-122.4194\n", i, i))
	}

	data := csvData.String()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parser, err := NewStreamingCSVParser(strings.NewReader(data), "bench.csv", nil)
		if err != nil {
			b.Fatal(err)
		}

		processor := &BenchBatchProcessor{}
		ctx := context.Background()

		err = parser.ProcessStreamInBatches(ctx, processor, 1000)
		if err != nil {
			b.Fatal(err)
		}
	}
}

type BenchBatchProcessor struct{}

func (p *BenchBatchProcessor) ProcessBatch(batch *StreamingCSVBatch) error {
	// Just count rows without storing anything
	return nil
}

func (p *BenchBatchProcessor) ProcessingComplete() error {
	return nil
}
