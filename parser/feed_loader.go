package parser

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// FeedLoader loads GTFS feeds from various sources
type FeedLoader struct {
	files     map[string]io.ReadCloser // For ZIP files (deprecated approach)
	filePaths map[string]string        // For directory files
	zipReader *zip.ReadCloser          // For ZIP files (new approach)
	zipFiles  map[string]*zip.File     // For ZIP files (new approach)
	isDir     bool                     // True if loading from directory
}

// LoadFromZip loads a GTFS feed from a zip file
func LoadFromZip(zipPath string) (*FeedLoader, error) {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open zip file: %v", err)
	}

	loader := &FeedLoader{
		files:     make(map[string]io.ReadCloser),
		filePaths: make(map[string]string),
		zipReader: reader,
		zipFiles:  make(map[string]*zip.File),
		isDir:     false,
	}

	// Map ZIP files for multiple access
	for _, file := range reader.File {
		if file.FileInfo().IsDir() {
			continue
		}

		// Only include .txt and .geojson files at root level
		name := filepath.Base(file.Name)
		if !strings.HasSuffix(name, ".txt") && !strings.HasSuffix(name, ".geojson") {
			continue
		}

		loader.zipFiles[name] = file
	}

	return loader, nil
}

// LoadFromDirectory loads a GTFS feed from a directory
func LoadFromDirectory(dirPath string) (*FeedLoader, error) {
	loader := &FeedLoader{
		files:     make(map[string]io.ReadCloser),
		filePaths: make(map[string]string),
		isDir:     true,
	}

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(name, ".txt") && !strings.HasSuffix(name, ".geojson") {
			continue
		}

		filePath := filepath.Join(dirPath, name)
		loader.filePaths[name] = filePath
	}

	return loader, nil
}

// GetFile returns a reader for the specified GTFS file
func (l *FeedLoader) GetFile(filename string) (io.ReadCloser, error) {
	if l.isDir {
		// For directory files, open a fresh reader each time
		filePath, exists := l.filePaths[filename]
		if !exists {
			return nil, fmt.Errorf("file not found: %s", filename)
		}
		return os.Open(filePath)
	} else {
		// For ZIP files, open a fresh reader each time
		zipFile, exists := l.zipFiles[filename]
		if !exists {
			return nil, fmt.Errorf("file not found: %s", filename)
		}
		return zipFile.Open()
	}
}

// HasFile returns true if the specified file exists in the feed
func (l *FeedLoader) HasFile(filename string) bool {
	if l.isDir {
		_, exists := l.filePaths[filename]
		return exists
	} else {
		_, exists := l.zipFiles[filename]
		return exists
	}
}

// ListFiles returns a list of all files in the feed
func (l *FeedLoader) ListFiles() []string {
	if l.isDir {
		files := make([]string, 0, len(l.filePaths))
		for filename := range l.filePaths {
			files = append(files, filename)
		}
		return files
	} else {
		files := make([]string, 0, len(l.zipFiles))
		for filename := range l.zipFiles {
			files = append(files, filename)
		}
		return files
	}
}

// Close closes all open file readers
func (l *FeedLoader) Close() error {
	var firstErr error

	// Close any old-style file readers
	for _, reader := range l.files {
		if err := reader.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	// Close ZIP reader if present
	if l.zipReader != nil {
		if err := l.zipReader.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	return firstErr
}

// RequiredFiles lists the required GTFS files
var RequiredFiles = []string{
	"agency.txt",
	"stops.txt",
	"routes.txt",
	"trips.txt",
	"stop_times.txt",
}

// ConditionallyRequiredFiles lists files that may be required based on feed content
var ConditionallyRequiredFiles = []string{
	"calendar.txt",
	"calendar_dates.txt",
	"feed_info.txt",
}

// OptionalFiles lists common optional GTFS files
var OptionalFiles = []string{
	"fare_attributes.txt",
	"fare_rules.txt",
	"shapes.txt",
	"frequencies.txt",
	"transfers.txt",
	"pathways.txt",
	"levels.txt",
	"translations.txt",
	"attributions.txt",
}
