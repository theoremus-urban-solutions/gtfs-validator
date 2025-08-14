package core

import (
	"io"
	"log"
	"strings"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/parser"
	"github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

// EmptyFileValidator validates that files are not empty
type EmptyFileValidator struct{}

// NewEmptyFileValidator creates a new empty file validator
func NewEmptyFileValidator() *EmptyFileValidator {
	return &EmptyFileValidator{}
}

// Validate checks that GTFS files are not empty
func (v *EmptyFileValidator) Validate(loader *parser.FeedLoader, container *notice.NoticeContainer, config validator.Config) {
	files := loader.ListFiles()

	for _, filename := range files {
		v.validateFileNotEmpty(loader, container, filename)
	}
}

// validateFileNotEmpty checks if a single file is empty
func (v *EmptyFileValidator) validateFileNotEmpty(loader *parser.FeedLoader, container *notice.NoticeContainer, filename string) {
	reader, err := loader.GetFile(filename)
	if err != nil {
		return // File doesn't exist, other validators handle this
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close reader %v", closeErr)
		}
	}()

	// Read the entire file content to check if it contains only headers and whitespace
	content, err := io.ReadAll(reader)
	if err != nil {
		return // Can't read file, skip validation
	}

	// Check if the file contains only headers and whitespace rows
	lines := strings.Split(string(content), "\n")
	hasDataRows := false
	hasHeaders := false

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			if i == 0 {
				// First non-empty line is likely headers
				hasHeaders = true
			} else {
				// Non-empty line after headers is a data row
				hasDataRows = true
				break
			}
		}
	}

	// If the file has headers but no data rows, it's empty
	if hasHeaders && !hasDataRows {
		container.AddNotice(notice.NewEmptyFileNotice(filename))
		return
	}

	// Reset reader for CSV parsing
	reader, err = loader.GetFile(filename)
	if err != nil {
		return
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close reader %v", closeErr)
		}
	}()

	csvFile, err := parser.NewCSVFile(reader, filename)
	if err != nil {
		return // File format issues, other validators handle this
	}

	// Iterate rows to find any non-empty data row
	hasData := false
	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			// CSV parsing errors should be handled by other validators
			return
		}

		// Check if any field contains non-whitespace content
		rowHasContent := false
		for _, val := range row.Values {
			if strings.TrimSpace(val) != "" {
				rowHasContent = true
				break
			}
		}
		if rowHasContent {
			hasData = true
			break
		}
	}

	if !hasData {
		container.AddNotice(notice.NewEmptyFileNotice(filename))
	}
}
