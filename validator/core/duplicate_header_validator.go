package core

import (
	"log"
	"strings"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/parser"
	"github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

// DuplicateHeaderValidator validates that CSV headers don't contain duplicates
type DuplicateHeaderValidator struct{}

// NewDuplicateHeaderValidator creates a new duplicate header validator
func NewDuplicateHeaderValidator() *DuplicateHeaderValidator {
	return &DuplicateHeaderValidator{}
}

// Validate checks for duplicate column headers in GTFS files
func (v *DuplicateHeaderValidator) Validate(loader *parser.FeedLoader, container *notice.NoticeContainer, config validator.Config) {
	files := loader.ListFiles()

	for _, filename := range files {
		v.validateFileHeaders(loader, container, filename)
	}
}

// validateFileHeaders checks for duplicate headers in a single file
func (v *DuplicateHeaderValidator) validateFileHeaders(loader *parser.FeedLoader, container *notice.NoticeContainer, filename string) {
	reader, err := loader.GetFile(filename)
	if err != nil {
		return // File doesn't exist, other validators handle this
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

	// Check for duplicate headers
	headers := csvFile.Headers
	headerCounts := make(map[string][]int)

	for i, header := range headers {
		headerName := strings.TrimSpace(header)
		headerCounts[headerName] = append(headerCounts[headerName], i)
	}

	// Report duplicates
	for headerName, positions := range headerCounts {
		if len(positions) > 1 {
			container.AddNotice(notice.NewDuplicateHeaderNotice(
				filename,
				headerName,
				positions,
			))
		}
	}
}
