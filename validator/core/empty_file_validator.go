package core

import (
	"io"
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
	defer reader.Close()

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
			// Malformed CSV is handled by other validators; do not flag empty file here
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
