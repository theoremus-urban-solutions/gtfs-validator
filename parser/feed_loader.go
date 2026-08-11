package parser

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// FeedLoader loads GTFS feeds from various sources
type FeedLoader struct {
	files            map[string]io.ReadCloser // For ZIP files (deprecated approach)
	filePaths        map[string]string        // For directory files
	zipReader        *zip.ReadCloser          // For ZIP files (new approach)
	zipFiles         map[string]*zip.File     // For ZIP files (new approach)
	isDir            bool                     // True if loading from directory
	filesInSubfolder bool                     // GTFS files found below the root

	// Per-file state, computed on demand and memoised; see feed_state.go.
	stateOnce sync.Once
	stateMu   *sync.Mutex
	states    map[string]FileState
}

// isArchiveMetadata reports whether a zip entry is packaging noise rather than
// part of the dataset. Archives zipped on macOS carry a __MACOSX sidecar tree;
// treating that as "GTFS files in a subfolder" would fail almost every feed
// produced on a Mac.
func isArchiveMetadata(name string) bool {
	return strings.HasPrefix(name, "__MACOSX/") || strings.HasPrefix(filepath.Base(name), "._")
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

	// Map ZIP files for multiple access. Only entries at the archive root are
	// part of the dataset: the spec requires the files to sit directly at the
	// root, and a subfolder is a packaging error the caller has to fix. This
	// used to key on filepath.Base, which silently flattened `gtfs/stops.txt`
	// to `stops.txt` and validated the nested feed as though it were correct —
	// hiding both the packaging error and the missing-file errors that follow
	// from it.
	for _, file := range reader.File {
		if file.FileInfo().IsDir() || isArchiveMetadata(file.Name) {
			continue
		}

		// Zip entry names always use forward slashes, whatever wrote them.
		name := strings.TrimPrefix(file.Name, "./")
		if !strings.HasSuffix(name, ".txt") && !strings.HasSuffix(name, ".geojson") {
			continue
		}

		if strings.Contains(name, "/") {
			// A stray text file in a subfolder is not a GTFS packaging error;
			// only a file the spec knows about is.
			if isGTFSFilename(filepath.Base(name)) {
				loader.filesInSubfolder = true
			}
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
			// Same packaging error as the zip case, reached by unpacking one.
			// Only the immediate children are examined: the error being
			// described is "the containing folder was handed over instead of
			// its contents", which is always exactly one level.
			if subEntries, err := os.ReadDir(filepath.Join(dirPath, entry.Name())); err == nil {
				for _, sub := range subEntries {
					if !sub.IsDir() && isGTFSFilename(sub.Name()) {
						loader.filesInSubfolder = true
						break
					}
				}
			}
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

// HasFilesInSubfolder reports whether GTFS files were found below the root of
// the archive or directory. Those files are deliberately not loaded, so the
// feed also reports every required file as missing — which is the same thing
// the canonical validator does, and the reason the root is treated as empty
// rather than being quietly substituted with the subfolder's contents.
func (l *FeedLoader) HasFilesInSubfolder() bool {
	return l.filesInSubfolder
}

// isGTFSFilename reports whether a name is a file the GTFS spec defines.
func isGTFSFilename(name string) bool {
	for _, group := range [][]string{RequiredFiles, ConditionallyRequiredFiles, OptionalFiles} {
		for _, known := range group {
			if name == known {
				return true
			}
		}
	}
	return false
}

// GetFile returns a reader for the specified GTFS file
func (l *FeedLoader) GetFile(filename string) (io.ReadCloser, error) {
	if l.isDir {
		// For directory files, open a fresh reader each time
		filePath, exists := l.filePaths[filename]
		if !exists {
			return nil, fmt.Errorf("file not found: %s", filename)
		}
		return os.Open(filePath) // #nosec G304 -- GTFS file path from validated directory
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
