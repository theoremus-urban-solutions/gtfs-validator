package gtfsvalidator

import (
	"archive/zip"
	"fmt"
	"os"
	"testing"
)

// BenchmarkValidateFile benchmarks the main validation function
func BenchmarkValidateFile(b *testing.B) {
	validator := New()

	// Create test ZIP once for all iterations
	zipPath := createBenchmarkZip(b, MinimalValidGTFS())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := validator.ValidateFile(zipPath)
		if err != nil {
			b.Fatalf("Validation failed: %v", err)
		}
	}
}

// BenchmarkValidateFile_Performance benchmarks performance mode
func BenchmarkValidateFile_Performance(b *testing.B) {
	validator := New(WithValidationMode(ValidationModePerformance))
	zipPath := createBenchmarkZip(b, MinimalValidGTFS())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := validator.ValidateFile(zipPath)
		if err != nil {
			b.Fatalf("Validation failed: %v", err)
		}
	}
}

// BenchmarkValidateFile_Comprehensive benchmarks comprehensive mode
func BenchmarkValidateFile_Comprehensive(b *testing.B) {
	validator := New(WithValidationMode(ValidationModeComprehensive))
	zipPath := createBenchmarkZip(b, MinimalValidGTFS())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := validator.ValidateFile(zipPath)
		if err != nil {
			b.Fatalf("Validation failed: %v", err)
		}
	}
}

// BenchmarkValidateFile_ParallelWorkers benchmarks different worker counts
func BenchmarkValidateFile_ParallelWorkers(b *testing.B) {
	workerCounts := []int{1, 2, 4, 8}

	for _, workers := range workerCounts {
		b.Run(fmt.Sprintf("workers_%d", workers), func(b *testing.B) {
			validator := New(WithParallelWorkers(workers))
			zipPath := createBenchmarkZip(b, MinimalValidGTFS())

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := validator.ValidateFile(zipPath)
				if err != nil {
					b.Fatalf("Validation failed: %v", err)
				}
			}
		})
	}
}

// createBenchmarkZip creates a temporary ZIP file for benchmarking
func createBenchmarkZip(b *testing.B, files map[string]string) string {
	b.Helper()

	tmpFile, err := os.CreateTemp("", "benchmark-gtfs-*.zip")
	if err != nil {
		b.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()

	zipFile, err := os.Create(tmpFile.Name())
	if err != nil {
		b.Fatalf("Failed to create zip file: %v", err)
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	for filename, content := range files {
		writer, err := zipWriter.Create(filename)
		if err != nil {
			b.Fatalf("Failed to create zip entry %s: %v", filename, err)
		}

		if _, err := writer.Write([]byte(content)); err != nil {
			b.Fatalf("Failed to write zip entry %s: %v", filename, err)
		}
	}

	// Register cleanup
	b.Cleanup(func() {
		os.Remove(tmpFile.Name())
	})

	return tmpFile.Name()
}
