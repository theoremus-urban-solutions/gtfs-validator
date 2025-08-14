package parser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCSVFile_Basic(t *testing.T) {
	csvContent := `header1,header2,header3
value1,value2,value3
value4,value5,value6`

	reader := strings.NewReader(csvContent)
	csvFile, err := NewCSVFile(reader, "test.txt")
	if err != nil {
		t.Fatalf("Failed to create CSV file: %v", err)
	}

	if err := csvFile.ReadAll(); err != nil {
		t.Fatalf("Failed to read CSV: %v", err)
	}

	// Test headers
	expectedHeaders := []string{"header1", "header2", "header3"}
	if len(csvFile.Headers) != len(expectedHeaders) {
		t.Errorf("Expected %d headers, got %d", len(expectedHeaders), len(csvFile.Headers))
	}
	for i, expected := range expectedHeaders {
		if i >= len(csvFile.Headers) || csvFile.Headers[i] != expected {
			t.Errorf("Expected header[%d] = %s, got %s", i, expected, csvFile.Headers[i])
		}
	}

	// Test rows
	if csvFile.RowCount() != 2 {
		t.Errorf("Expected 2 rows, got %d", csvFile.RowCount())
	}

	// Test first row
	if len(csvFile.Rows) < 1 {
		t.Fatal("Expected at least 1 row")
	}
	row1 := csvFile.Rows[0]
	if row1.RowNumber != 2 { // Header is row 1
		t.Errorf("Expected row number 2, got %d", row1.RowNumber)
	}
	if value, exists := row1.Values["header1"]; !exists || value != "value1" {
		t.Errorf("Expected header1 = value1, got %s", value)
	}
	if value, exists := row1.Values["header2"]; !exists || value != "value2" {
		t.Errorf("Expected header2 = value2, got %s", value)
	}
}

func TestCSVFile_EmptyFile(t *testing.T) {
	reader := strings.NewReader("")
	_, err := NewCSVFile(reader, "empty.txt")
	if err == nil {
		t.Error("Expected error for empty file")
	}
	if !strings.Contains(err.Error(), "empty file") {
		t.Errorf("Expected 'empty file' error, got: %v", err)
	}
}

func TestCSVFile_OnlyHeaders(t *testing.T) {
	csvContent := `header1,header2,header3`

	reader := strings.NewReader(csvContent)
	csvFile, err := NewCSVFile(reader, "headers_only.txt")
	if err != nil {
		t.Fatalf("Failed to create CSV file: %v", err)
	}

	if err := csvFile.ReadAll(); err != nil {
		t.Fatalf("Failed to read CSV: %v", err)
	}

	if len(csvFile.Headers) != 3 {
		t.Errorf("Expected 3 headers, got %d", len(csvFile.Headers))
	}

	if csvFile.RowCount() != 0 {
		t.Errorf("Expected 0 data rows, got %d", csvFile.RowCount())
	}
}

func TestCSVFile_MissingValues(t *testing.T) {
	csvContent := `header1,header2,header3
value1,,value3
,value5,`

	reader := strings.NewReader(csvContent)
	csvFile, err := NewCSVFile(reader, "missing_values.txt")
	if err != nil {
		t.Fatalf("Failed to create CSV file: %v", err)
	}

	if err := csvFile.ReadAll(); err != nil {
		t.Fatalf("Failed to read CSV: %v", err)
	}

	if csvFile.RowCount() != 2 {
		t.Errorf("Expected 2 rows, got %d", csvFile.RowCount())
	}

	// Test missing values are empty strings
	row1 := csvFile.Rows[0]
	if value, exists := row1.Values["header2"]; !exists || value != "" {
		t.Errorf("Expected empty string for missing value, got %q", value)
	}

	row2 := csvFile.Rows[1]
	if value, exists := row2.Values["header1"]; !exists || value != "" {
		t.Errorf("Expected empty string for missing value, got %q", value)
	}
	if value, exists := row2.Values["header3"]; !exists || value != "" {
		t.Errorf("Expected empty string for missing value, got %q", value)
	}
}

func TestCSVFile_ExtraCommas(t *testing.T) {
	csvContent := `header1,header2,header3
value1,value2,value3,extra1`

	reader := strings.NewReader(csvContent)
	csvFile, err := NewCSVFile(reader, "extra_commas.txt")
	if err != nil {
		t.Fatalf("Failed to create CSV file: %v", err)
	}

	// Reading should fail due to wrong number of fields
	if err := csvFile.ReadAll(); err == nil {
		t.Error("Expected error when reading CSV with extra fields")
	}
}

func TestCSVFile_UTF8BOM(t *testing.T) {
	// CSV with UTF-8 BOM
	csvContent := "\ufeffheader1,header2\nvalue1,value2"

	reader := strings.NewReader(csvContent)
	csvFile, err := NewCSVFile(reader, "bom.txt")
	if err != nil {
		t.Fatalf("Failed to create CSV file: %v", err)
	}

	if err := csvFile.ReadAll(); err != nil {
		t.Fatalf("Failed to read CSV with BOM: %v", err)
	}

	// BOM should be stripped from first header
	if len(csvFile.Headers) < 1 {
		t.Fatal("Expected at least 1 header")
	}
	if csvFile.Headers[0] != "header1" {
		t.Errorf("Expected first header to be 'header1' (BOM stripped), got %q", csvFile.Headers[0])
	}
}

func TestFeedLoader_FromDirectory(t *testing.T) {
	// Create temporary directory with test files
	tmpDir := t.TempDir()

	// Create test files
	testFiles := map[string]string{
		"agency.txt": "agency_id,agency_name\ntest_agency,Test Agency",
		"routes.txt": "route_id,route_short_name\nroute_1,1",
	}

	for filename, content := range testFiles {
		filePath := filepath.Join(tmpDir, filename)
		if err := os.WriteFile(filePath, []byte(content), 0600); err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
	}

	// Test loading from directory
	loader, err := LoadFromDirectory(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load from directory: %v", err)
	}
	defer loader.Close()

	// Test HasFile
	if !loader.HasFile("agency.txt") {
		t.Error("Expected agency.txt to exist")
	}
	if !loader.HasFile("routes.txt") {
		t.Error("Expected routes.txt to exist")
	}
	if loader.HasFile("nonexistent.txt") {
		t.Error("Expected nonexistent.txt to not exist")
	}

	// Test GetFile
	reader, err := loader.GetFile("agency.txt")
	if err != nil {
		t.Fatalf("Failed to get agency.txt: %v", err)
	}
	defer reader.Close()

	csvFile, err := NewCSVFile(reader, "agency.txt")
	if err != nil {
		t.Fatalf("Failed to create CSV from agency.txt: %v", err)
	}

	if err := csvFile.ReadAll(); err != nil {
		t.Fatalf("Failed to read CSV: %v", err)
	}

	if csvFile.RowCount() != 1 {
		t.Errorf("Expected 1 row in agency.txt, got %d", csvFile.RowCount())
	}
}

func TestFeedLoader_NonExistentDirectory(t *testing.T) {
	_, err := LoadFromDirectory("/non/existent/directory")
	if err == nil {
		t.Error("Expected error for non-existent directory")
	}
}

func TestFeedLoader_GetNonExistentFile(t *testing.T) {
	tmpDir := t.TempDir()

	loader, err := LoadFromDirectory(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load from directory: %v", err)
	}
	defer loader.Close()

	_, err = loader.GetFile("nonexistent.txt")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

// Test the required files constant
func TestRequiredFiles(t *testing.T) {
	expectedFiles := []string{
		"agency.txt",
		"routes.txt",
		"trips.txt",
		"stop_times.txt",
		"stops.txt",
	}

	if len(RequiredFiles) != len(expectedFiles) {
		t.Errorf("Expected %d required files, got %d", len(expectedFiles), len(RequiredFiles))
	}

	for _, expected := range expectedFiles {
		found := false
		for _, required := range RequiredFiles {
			if required == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected %s to be in required files", expected)
		}
	}
}

func TestCSVFile_RowNumbering(t *testing.T) {
	csvContent := `header1,header2
row1_val1,row1_val2
row2_val1,row2_val2
row3_val1,row3_val2`

	reader := strings.NewReader(csvContent)
	csvFile, err := NewCSVFile(reader, "test.txt")
	if err != nil {
		t.Fatalf("Failed to create CSV file: %v", err)
	}

	if err := csvFile.ReadAll(); err != nil {
		t.Fatalf("Failed to read CSV: %v", err)
	}

	if csvFile.RowCount() != 3 {
		t.Errorf("Expected 3 rows, got %d", csvFile.RowCount())
	}

	// Check row numbers (should start from 2, as 1 is the header)
	expectedRowNumbers := []int{2, 3, 4}
	for i, expectedRowNum := range expectedRowNumbers {
		if i >= len(csvFile.Rows) {
			t.Errorf("Missing row %d", i)
			continue
		}
		if csvFile.Rows[i].RowNumber != expectedRowNum {
			t.Errorf("Expected row %d to have row number %d, got %d", i, expectedRowNum, csvFile.Rows[i].RowNumber)
		}
	}
}

func TestCSVFile_Filename(t *testing.T) {
	reader := strings.NewReader("header\nvalue")
	csvFile, err := NewCSVFile(reader, "test_filename.txt")
	if err != nil {
		t.Fatalf("Failed to create CSV file: %v", err)
	}

	if csvFile.Filename != "test_filename.txt" {
		t.Errorf("Expected filename 'test_filename.txt', got %s", csvFile.Filename)
	}
}
