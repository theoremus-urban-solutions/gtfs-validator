package validator

import (
	"io"
	"log"
	"strings"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/parser"
	"github.com/theoremus-urban-solutions/gtfs-validator/schema"
)

// FileStructureValidator validates the structure of GTFS files
type FileStructureValidator struct{}

// NewFileStructureValidator creates a new file structure validator
func NewFileStructureValidator() *FileStructureValidator {
	return &FileStructureValidator{}
}

// Validate checks the structure of all GTFS files
func (v *FileStructureValidator) Validate(loader *parser.FeedLoader, container *notice.NoticeContainer, config Config) {
	files := loader.ListFiles()

	for _, filename := range files {
		v.validateFile(loader, container, filename)
	}
}

// validateFile validates a single file's structure
func (v *FileStructureValidator) validateFile(loader *parser.FeedLoader, container *notice.NoticeContainer, filename string) {
	reader, err := loader.GetFile(filename)
	if err != nil {
		// This shouldn't happen as we're iterating over existing files
		return
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close reader %v", closeErr)
		}
	}()

	csvFile, err := parser.NewCSVFile(reader, filename)
	if err != nil {
		if strings.Contains(err.Error(), "empty file") {
			container.AddNotice(notice.NewEmptyFileNotice(filename))
		} else {
			// Add a CSV parsing error notice
			container.AddNotice(notice.NewBaseNotice("csv_parsing_failed", notice.ERROR, map[string]interface{}{
				"filename": filename,
				"error":    err.Error(),
			}))
		}
		return
	}

	// Check for empty file (no data rows)
	err = csvFile.ReadAll()
	if err != nil && err != io.EOF {
		container.AddNotice(notice.NewBaseNotice("csv_parsing_failed", notice.ERROR, map[string]interface{}{
			"filename": filename,
			"error":    err.Error(),
		}))
		return
	}

	if csvFile.IsEmpty() {
		container.AddNotice(notice.NewEmptyFileNotice(filename))
		return
	}

	// Check for unknown columns based on file type
	v.checkUnknownColumns(csvFile, container)
}

// checkUnknownColumns checks for columns not defined in GTFS spec
func (v *FileStructureValidator) checkUnknownColumns(csvFile *parser.CSVFile, container *notice.NoticeContainer) {
	knownColumns, isSpecFile := schema.KnownColumns(csvFile.Filename)
	if !isSpecFile {
		// Nothing is known about this file's columns, so every one of them
		// would be reported. The file itself is already reported once as
		// unknown_file, which is the finding; repeating it per column is not.
		return
	}

	for i, header := range csvFile.Headers {
		// Surrounding whitespace is a defect of the header itself and is
		// reported as such elsewhere. Matching on the trimmed name stops a
		// merely untidy column from also being read as an unknown one.
		if !contains(knownColumns, strings.TrimSpace(header)) {
			container.AddNotice(notice.NewUnknownColumnNotice(
				csvFile.Filename,
				header,
				i,
			))
		}
	}
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
