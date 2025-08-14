package relationship

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/parser"
)

// CreateTestFeedLoader creates a real FeedLoader from a map of test files
// The map key is the filename (e.g., "agency.txt") and value is the file content
func CreateTestFeedLoader(t *testing.T, files map[string]string) *parser.FeedLoader {
	t.Helper()

	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "gtfs-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Clean up the directory after test
	t.Cleanup(func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Errorf("Failed to remove temp dir: %v", err)
		}
	})

	// Write test files to the temporary directory
	for filename, content := range files {
		filePath := filepath.Join(tmpDir, filename)
		if err := os.WriteFile(filePath, []byte(content), 0600); err != nil {
			t.Fatalf("Failed to write test file %s: %v", filename, err)
		}
	}

	// Create and return the FeedLoader
	loader, err := parser.LoadFromDirectory(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create FeedLoader: %v", err)
	}

	// Clean up the loader after test
	t.Cleanup(func() {
		if err := loader.Close(); err != nil {
			t.Errorf("Failed to close loader: %v", err)
		}
	})

	return loader
}
