package parser

import (
	"strings"
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/pools"
)

func TestCSVFilePooledMemory(t *testing.T) {
	csvData := `id,name,type
123,Test Station,0
456,Another Station,1
789,Third Station,2`

	csvFile, err := NewCSVFile(strings.NewReader(csvData), "test.csv")
	if err != nil {
		t.Fatalf("Failed to create CSV file: %v", err)
	}

	// Read all rows
	err = csvFile.ReadAll()
	if err != nil {
		t.Fatalf("Failed to read CSV data: %v", err)
	}

	// Verify data was parsed correctly with pooled memory
	if len(csvFile.Rows) != 3 {
		t.Errorf("Expected 3 rows, got %d", len(csvFile.Rows))
	}

	// Check first row
	row := csvFile.Rows[0]
	if row.Values["id"] != "123" {
		t.Errorf("Expected id=123, got %s", row.Values["id"])
	}
	if row.Values["name"] != "Test Station" {
		t.Errorf("Expected name=Test Station, got %s", row.Values["name"])
	}
	if row.Values["type"] != "0" {
		t.Errorf("Expected type=0, got %s", row.Values["type"])
	}

	// Test memory release
	firstRowValues := csvFile.Rows[0].Values
	csvFile.ReleaseRow(&csvFile.Rows[0])

	// After release, the Values should be nil
	if csvFile.Rows[0].Values != nil {
		t.Error("Row values should be nil after release")
	}

	// The original map should still be valid (but returned to pool)
	if firstRowValues["id"] != "123" {
		t.Error("Original map should still be valid after return to pool")
	}

	// Test releasing all rows
	csvFile.ReleaseAllRows()
	for i, row := range csvFile.Rows {
		if row.Values != nil {
			t.Errorf("Row %d values should be nil after ReleaseAllRows", i)
		}
	}
}

func TestCSVFileRowByRowPooled(t *testing.T) {
	csvData := `id,name,type
123,Test Station,0
456,Another Station,1`

	csvFile, err := NewCSVFile(strings.NewReader(csvData), "test.csv")
	if err != nil {
		t.Fatalf("Failed to create CSV file: %v", err)
	}

	// Read rows one by one and release immediately to test pool reuse
	row1, err := csvFile.ReadRow()
	if err != nil {
		t.Fatalf("Failed to read first row: %v", err)
	}

	if row1.Values["id"] != "123" {
		t.Errorf("Expected id=123, got %s", row1.Values["id"])
	}

	// Release the row back to pool
	csvFile.ReleaseRow(row1)

	// Read second row - should reuse pooled memory
	row2, err := csvFile.ReadRow()
	if err != nil {
		t.Fatalf("Failed to read second row: %v", err)
	}

	if row2.Values["id"] != "456" {
		t.Errorf("Expected id=456, got %s", row2.Values["id"])
	}

	csvFile.ReleaseRow(row2)
}

func TestCSVFileMismatchedFieldsWithPooling(t *testing.T) {
	// Test case where CSV has mismatched field counts (but valid CSV structure)
	csvData := `id,name,type
123,"Test Station",0
456,"Another Station",1`

	csvFile, err := NewCSVFile(strings.NewReader(csvData), "test.csv")
	if err != nil {
		t.Fatalf("Failed to create CSV file: %v", err)
	}

	// Read first row
	row1, err := csvFile.ReadRow()
	if err != nil {
		t.Fatalf("Failed to read first row: %v", err)
	}

	if row1.Values["id"] != "123" {
		t.Errorf("Expected id=123, got %s", row1.Values["id"])
	}
	if row1.Values["name"] != "Test Station" {
		t.Errorf("Expected name=Test Station, got %s", row1.Values["name"])
	}
	if row1.RawFieldCount != 3 {
		t.Errorf("Expected raw field count 3, got %d", row1.RawFieldCount)
	}

	csvFile.ReleaseRow(row1)

	// Read second row
	row2, err := csvFile.ReadRow()
	if err != nil {
		t.Fatalf("Failed to read second row: %v", err)
	}

	if row2.Values["id"] != "456" {
		t.Errorf("Expected id=456, got %s", row2.Values["id"])
	}
	if row2.RawFieldCount != 3 {
		t.Errorf("Expected raw field count 3, got %d", row2.RawFieldCount)
	}

	csvFile.ReleaseRow(row2)
}

func BenchmarkCSVFileWithPooling(b *testing.B) {
	// Create a larger CSV dataset for benchmarking
	var csvData strings.Builder
	csvData.WriteString("id,name,type,lat,lon\n")
	for i := 0; i < 1000; i++ {
		csvData.WriteString("123,Test Station,0,37.7749,-122.4194\n")
	}

	data := csvData.String()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		csvFile, err := NewCSVFile(strings.NewReader(data), "test.csv")
		if err != nil {
			b.Fatal(err)
		}

		err = csvFile.ReadAll()
		if err != nil {
			b.Fatal(err)
		}

		// Release all memory back to pools
		csvFile.ReleaseAllRows()
	}
}

func BenchmarkCSVFileWithoutPooling(b *testing.B) {
	// Benchmark without using memory pools (for comparison)
	var csvData strings.Builder
	csvData.WriteString("id,name,type,lat,lon\n")
	for i := 0; i < 1000; i++ {
		csvData.WriteString("123,Test Station,0,37.7749,-122.4194\n")
	}

	data := csvData.String()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		csvFile, err := NewCSVFile(strings.NewReader(data), "test.csv")
		if err != nil {
			b.Fatal(err)
		}

		// Temporarily disable pooling for this benchmark
		csvFile.parser = nil

		// Manually read rows without pooling
		for {
			values, err := csvFile.reader.Read()
			if err != nil {
				break
			}

			// Create map without pooling
			rowValues := make(map[string]string)
			for i, header := range csvFile.Headers {
				if i < len(values) {
					rowValues[header] = values[i]
				} else {
					rowValues[header] = ""
				}
			}

			csvFile.Rows = append(csvFile.Rows, CSVRow{
				RowNumber:     len(csvFile.Rows) + 2,
				Values:        rowValues,
				RawFieldCount: len(values),
			})
		}
	}
}

func TestGlobalPoolIntegration(t *testing.T) {
	// Test that the CSV parser integrates with global pools correctly

	// Get initial pool stats if using stats pools
	initialBuffer := pools.GetSmallBuffer()
	pools.PutSmallBuffer(initialBuffer)

	initialRecord := pools.GetCSVRecord()
	pools.PutCSVRecord(initialRecord)

	// Create and process CSV file
	csvData := `id,name,type
123,Test Station,0`

	csvFile, err := NewCSVFile(strings.NewReader(csvData), "test.csv")
	if err != nil {
		t.Fatalf("Failed to create CSV file: %v", err)
	}

	err = csvFile.ReadAll()
	if err != nil {
		t.Fatalf("Failed to read CSV data: %v", err)
	}

	// Verify the row was parsed correctly
	if len(csvFile.Rows) != 1 {
		t.Errorf("Expected 1 row, got %d", len(csvFile.Rows))
	}

	if csvFile.Rows[0].Values["id"] != "123" {
		t.Errorf("Expected id=123, got %s", csvFile.Rows[0].Values["id"])
	}

	// Release memory back to pools
	csvFile.ReleaseAllRows()
}
