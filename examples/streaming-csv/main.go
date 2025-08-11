// Streaming CSV example demonstrates how to process large GTFS CSV files
// without loading them entirely into memory using the streaming CSV parser
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/theoremus-urban-solutions/gtfs-validator/parser"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: streaming-csv <csv-file> [mode]")
		fmt.Println("Modes: count, validate, filter, batch (default: count)")
		fmt.Println()
		fmt.Println("Example:")
		fmt.Println("  streaming-csv stop_times.txt count")
		fmt.Println("  streaming-csv routes.txt validate")
		fmt.Println("  streaming-csv stops.txt filter")
		fmt.Println("  streaming-csv stop_times.txt batch")
		os.Exit(1)
	}

	csvFile := os.Args[1]
	mode := "count"
	if len(os.Args) > 2 {
		mode = os.Args[2]
	}

	fmt.Printf("🚀 Streaming CSV Processing - Mode: %s\n", mode)
	fmt.Printf("📁 File: %s\n", csvFile)
	fmt.Println()

	// Open file
	file, err := os.Open(csvFile)
	if err != nil {
		log.Fatalf("❌ Failed to open file: %v", err)
	}
	defer file.Close()

	// Get file size for progress reporting
	fileInfo, err := file.Stat()
	if err == nil {
		sizeGB := float64(fileInfo.Size()) / (1024 * 1024 * 1024)
		fmt.Printf("📊 File size: %.2f GB\n", sizeGB)
		if sizeGB > 0.5 {
			fmt.Println("⚡ Large file detected - streaming processing will keep memory usage low")
		}
	}

	// Configure streaming options for large files
	opts := &parser.StreamingCSVOptions{
		BufferSize:       128 * 1024, // 128KB buffer for better I/O performance
		LazyQuotes:       true,       // Handle malformed quotes gracefully
		TrimLeadingSpace: false,      // Preserve original data
	}

	// Create streaming parser
	streamParser, err := parser.NewStreamingCSVParser(file, csvFile, opts)
	if err != nil {
		log.Fatalf("❌ Failed to create streaming parser: %v", err)
	}

	fmt.Printf("📋 Headers found: %v\n", streamParser.Headers())
	fmt.Printf("🔧 Buffer size: %d KB\n", opts.BufferSize/1024)
	fmt.Println()

	startTime := time.Now()
	ctx := context.Background()

	switch mode {
	case "count":
		err = demonstrateCountingMode(ctx, streamParser)
	case "validate":
		err = demonstrateValidationMode(ctx, streamParser)
	case "filter":
		err = demonstrateFilteringMode(ctx, streamParser)
	case "batch":
		err = demonstrateBatchMode(ctx, streamParser)
	default:
		log.Fatalf("❌ Unknown mode: %s", mode)
	}

	elapsed := time.Since(startTime)

	if err != nil {
		log.Fatalf("❌ Processing failed: %v", err)
	}

	// Print final statistics
	stats := streamParser.Statistics()
	fmt.Println()
	fmt.Println("📊 Processing Complete!")
	fmt.Printf("⏱️  Total time: %.2f seconds\n", elapsed.Seconds())
	fmt.Printf("📝 Rows processed: %s\n", formatNumber(stats.RowsProcessed))
	if stats.RowsProcessed > 0 && elapsed.Seconds() > 0 {
		throughput := float64(stats.RowsProcessed) / elapsed.Seconds()
		fmt.Printf("🚄 Throughput: %.0f rows/second\n", throughput)
	}
	fmt.Printf("💾 Memory usage: Streaming (low memory footprint)\n")
}

func demonstrateCountingMode(ctx context.Context, streamParser *parser.StreamingCSVParser) error {
	fmt.Println("🔢 Running in COUNT mode - counting all rows without storing in memory")

	processor := &parser.CountingProcessor{}

	// Add progress reporting
	progressProcessor := &ProgressReportingProcessor{
		Processor:    processor,
		ReportEvery:  100000, // Report every 100k rows
		lastReported: time.Now(),
	}

	err := streamParser.ProcessStream(ctx, progressProcessor)
	if err != nil {
		return err
	}

	fmt.Printf("✅ Total rows counted: %s\n", formatNumber(processor.Count))
	return nil
}

func demonstrateValidationMode(ctx context.Context, streamParser *parser.StreamingCSVParser) error {
	fmt.Println("✅ Running in VALIDATE mode - validating required fields")

	// Determine required fields based on headers
	headers := streamParser.Headers()
	var requiredFields []string

	// Common GTFS required fields
	commonRequired := map[string]bool{
		"agency_id":      true,
		"route_id":       true,
		"trip_id":        true,
		"stop_id":        true,
		"service_id":     true,
		"departure_time": true,
		"arrival_time":   true,
		"stop_sequence":  true,
	}

	for _, header := range headers {
		if commonRequired[header] {
			requiredFields = append(requiredFields, header)
		}
	}

	if len(requiredFields) == 0 {
		// Fallback: require first field to be non-empty
		if len(headers) > 0 {
			requiredFields = []string{headers[0]}
		}
	}

	fmt.Printf("🔍 Validating required fields: %v\n", requiredFields)

	processor := &parser.ValidatingProcessor{
		RequiredFields: requiredFields,
	}

	progressProcessor := &ProgressReportingProcessor{
		Processor:    processor,
		ReportEvery:  50000,
		lastReported: time.Now(),
	}

	err := streamParser.ProcessStream(ctx, progressProcessor)
	if err != nil {
		fmt.Printf("❌ Validation failed: %v\n", err)
		fmt.Printf("📊 Validation stats: %d errors, %d warnings found before failure\n",
			processor.ErrorCount, processor.WarningCount)
		return err
	}

	fmt.Printf("✅ Validation completed successfully!\n")
	fmt.Printf("📊 Validation stats: %d errors, %d warnings found\n",
		processor.ErrorCount, processor.WarningCount)
	return nil
}

func demonstrateFilteringMode(ctx context.Context, streamParser *parser.StreamingCSVParser) error {
	fmt.Println("🔄 Running in FILTER mode - filtering and processing specific records")

	headers := streamParser.Headers()

	// Create a filter function based on file type
	var filterFunc func(row *parser.CSVRow) bool
	var filterDescription string

	// Detect common GTFS file patterns and create appropriate filters
	if contains(headers, "stop_lat") && contains(headers, "stop_lon") {
		// stops.txt - filter for stops with coordinates
		filterFunc = func(row *parser.CSVRow) bool {
			lat := strings.TrimSpace(row.Values["stop_lat"])
			lon := strings.TrimSpace(row.Values["stop_lon"])
			return lat != "" && lon != "" && lat != "0" && lon != "0"
		}
		filterDescription = "stops with valid coordinates"
	} else if contains(headers, "route_type") {
		// routes.txt - filter for specific route types
		filterFunc = func(row *parser.CSVRow) bool {
			routeType := strings.TrimSpace(row.Values["route_type"])
			// Filter for bus (3) and rail (1,2) routes
			return routeType == "1" || routeType == "2" || routeType == "3"
		}
		filterDescription = "bus and rail routes (types 1,2,3)"
	} else if contains(headers, "trip_id") && contains(headers, "stop_sequence") {
		// stop_times.txt - filter for first/last stops in trips
		filterFunc = func(row *parser.CSVRow) bool {
			sequence := strings.TrimSpace(row.Values["stop_sequence"])
			return sequence == "1" || sequence == "0" // First stops
		}
		filterDescription = "first stops in trips (sequence 0 or 1)"
	} else {
		// Generic filter - non-empty first field
		if len(headers) > 0 {
			firstField := headers[0]
			filterFunc = func(row *parser.CSVRow) bool {
				return strings.TrimSpace(row.Values[firstField]) != ""
			}
			filterDescription = fmt.Sprintf("rows with non-empty %s", firstField)
		}
	}

	fmt.Printf("🎯 Filter: %s\n", filterDescription)

	counter := &parser.CountingProcessor{}
	processor := &parser.FilteringProcessor{
		FilterFunc: filterFunc,
		Processor:  counter,
	}

	progressProcessor := &ProgressReportingProcessor{
		Processor:    processor,
		ReportEvery:  100000,
		lastReported: time.Now(),
	}

	err := streamParser.ProcessStream(ctx, progressProcessor)
	if err != nil {
		return err
	}

	fmt.Printf("✅ Filtering completed!\n")
	fmt.Printf("📊 Results: %s rows processed, %s rows filtered out\n",
		formatNumber(processor.ProcessedRows), formatNumber(processor.FilteredRows))
	fmt.Printf("🎯 Filtered data: %s matching rows found\n", formatNumber(counter.Count))
	return nil
}

func demonstrateBatchMode(ctx context.Context, streamParser *parser.StreamingCSVParser) error {
	fmt.Println("📦 Running in BATCH mode - processing data in batches")

	batchSize := 10000 // 10k rows per batch
	fmt.Printf("📊 Batch size: %s rows\n", formatNumber(batchSize))

	processor := &BatchStatsProcessor{
		BatchSize: batchSize,
	}

	err := streamParser.ProcessStreamInBatches(ctx, processor, batchSize)
	if err != nil {
		return err
	}

	fmt.Printf("✅ Batch processing completed!\n")
	fmt.Printf("📦 Total batches: %d\n", processor.BatchesProcessed)
	fmt.Printf("📊 Total rows: %s\n", formatNumber(processor.TotalRows))
	fmt.Printf("📈 Average batch size: %.1f rows\n", float64(processor.TotalRows)/float64(processor.BatchesProcessed))

	return nil
}

// Helper processors

type ProgressReportingProcessor struct {
	Processor    parser.StreamingCSVProcessor
	ReportEvery  int
	rowCount     int
	lastReported time.Time
}

func (p *ProgressReportingProcessor) ProcessRow(row *parser.CSVRow) error {
	p.rowCount++

	// Report progress periodically
	if p.rowCount%p.ReportEvery == 0 {
		now := time.Now()
		elapsed := now.Sub(p.lastReported)
		if elapsed > 0 {
			rate := float64(p.ReportEvery) / elapsed.Seconds()
			fmt.Printf("⏳ Progress: %s rows processed (%.0f rows/sec)\n",
				formatNumber(p.rowCount), rate)
		}
		p.lastReported = now
	}

	return p.Processor.ProcessRow(row)
}

func (p *ProgressReportingProcessor) ProcessingComplete() error {
	return p.Processor.ProcessingComplete()
}

type BatchStatsProcessor struct {
	BatchSize        int
	BatchesProcessed int
	TotalRows        int
	startTime        time.Time
}

func (p *BatchStatsProcessor) ProcessBatch(batch *parser.StreamingCSVBatch) error {
	if p.BatchesProcessed == 0 {
		p.startTime = time.Now()
	}

	p.BatchesProcessed++
	p.TotalRows += len(batch.Rows)

	// Report batch progress
	elapsed := time.Since(p.startTime)
	if elapsed > 0 {
		batchRate := float64(p.BatchesProcessed) / elapsed.Seconds()
		rowRate := float64(p.TotalRows) / elapsed.Seconds()

		fmt.Printf("📦 Batch %d: %d rows (%.1f batches/sec, %.0f rows/sec)\n",
			p.BatchesProcessed, len(batch.Rows), batchRate, rowRate)
	}

	return nil
}

func (p *BatchStatsProcessor) ProcessingComplete() error {
	return nil
}

// Helper functions

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func formatNumber(n int) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}
	if n < 1000000 {
		return fmt.Sprintf("%.1fK", float64(n)/1000)
	}
	return fmt.Sprintf("%.1fM", float64(n)/1000000)
}
