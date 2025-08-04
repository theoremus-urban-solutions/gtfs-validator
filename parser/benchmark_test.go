package parser

import (
	"strings"
	"testing"
)

// BenchmarkCSVFile_ReadAll benchmarks CSV parsing performance
func BenchmarkCSVFile_ReadAll(b *testing.B) {
	// Create a reasonably sized CSV for benchmarking
	var csvBuilder strings.Builder
	csvBuilder.WriteString("id,name,value,description\n")
	for i := 0; i < 1000; i++ {
		csvBuilder.WriteString("id")
		csvBuilder.WriteString(string(rune('0' + i%10)))
		csvBuilder.WriteString(",name")
		csvBuilder.WriteString(string(rune('0' + i%10)))
		csvBuilder.WriteString(",value")
		csvBuilder.WriteString(string(rune('0' + i%10)))
		csvBuilder.WriteString(",description")
		csvBuilder.WriteString(string(rune('0' + i%10)))
		csvBuilder.WriteString("\n")
	}
	csvContent := csvBuilder.String()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader := strings.NewReader(csvContent)
		csvFile, err := NewCSVFile(reader, "benchmark.txt")
		if err != nil {
			b.Fatalf("Failed to create CSV file: %v", err)
		}

		if err := csvFile.ReadAll(); err != nil {
			b.Fatalf("Failed to read CSV: %v", err)
		}
	}
}

// BenchmarkCSVFile_ReadAll_LargeFile benchmarks with larger data
func BenchmarkCSVFile_ReadAll_LargeFile(b *testing.B) {
	// Create a large CSV for stress testing
	var csvBuilder strings.Builder
	csvBuilder.WriteString("stop_id,stop_name,stop_lat,stop_lon,stop_desc\n")
	for i := 0; i < 10000; i++ {
		csvBuilder.WriteString("stop_")
		csvBuilder.WriteString(string(rune('0' + (i/1000)%10)))
		csvBuilder.WriteString(string(rune('0' + (i/100)%10)))
		csvBuilder.WriteString(string(rune('0' + (i/10)%10)))
		csvBuilder.WriteString(string(rune('0' + i%10)))
		csvBuilder.WriteString(",Stop ")
		csvBuilder.WriteString(string(rune('0' + i%10)))
		csvBuilder.WriteString(",40.75")
		csvBuilder.WriteString(string(rune('0' + i%10)))
		csvBuilder.WriteString(",-73.98")
		csvBuilder.WriteString(string(rune('0' + i%10)))
		csvBuilder.WriteString(",Description for stop ")
		csvBuilder.WriteString(string(rune('0' + i%10)))
		csvBuilder.WriteString("\n")
	}
	csvContent := csvBuilder.String()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader := strings.NewReader(csvContent)
		csvFile, err := NewCSVFile(reader, "large_benchmark.txt")
		if err != nil {
			b.Fatalf("Failed to create CSV file: %v", err)
		}

		if err := csvFile.ReadAll(); err != nil {
			b.Fatalf("Failed to read CSV: %v", err)
		}
	}
}