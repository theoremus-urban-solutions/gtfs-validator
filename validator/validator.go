package validator

import (
	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/parser"
)

// Config contains validation configuration
type Config struct {
	CountryCode     string
	CurrentDate     interface{} // time.Time, but avoiding import cycle
	MaxMemory       int64
	ParallelWorkers int
}

// Validator is the interface for all validators
type Validator interface {
	// Validate performs validation and adds notices to the container
	Validate(loader *parser.FeedLoader, container *notice.NoticeContainer, config Config)
}
