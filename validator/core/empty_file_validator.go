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

	// empty_file means the file carries nothing at all — not even a header row.
	// A header with no data rows is a valid, legitimately empty table: the feed
	// is saying "this table exists and has no entries", which is a normal thing
	// to say and which canonical accepts without comment. Reporting it as an
	// empty file made a header-only stops.txt an ERROR here while canonical
	// loaded it and reported the references into it instead.
	if strings.TrimSpace(string(content)) == "" {
		container.AddNotice(notice.NewEmptyFileNotice(filename))
		return
	}

}
