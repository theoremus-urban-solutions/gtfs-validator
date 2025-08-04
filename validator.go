// Package gtfsvalidator provides a comprehensive GTFS feed validation library.
//
// The validator checks GTFS feeds against the official specification and provides
// detailed reports on errors, warnings, and informational notices.
//
// Basic usage:
//
//	validator := gtfsvalidator.New()
//	report, err := validator.ValidateFile("feed.zip")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	
//	if report.HasErrors() {
//	    fmt.Printf("Validation failed with %d errors\n", report.ErrorCount)
//	}
package gtfsvalidator

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"
)

// NoticeCallback is called for each notice group during streaming validation.
type NoticeCallback func(notice NoticeGroup)

// Validator is the main GTFS validator interface.
type Validator interface {
	// ValidateFile validates a GTFS feed from a file path (ZIP or directory).
	ValidateFile(path string) (*ValidationReport, error)
	
	// ValidateFileWithContext validates with cancellation support.
	ValidateFileWithContext(ctx context.Context, path string) (*ValidationReport, error)
	
	// ValidateReader validates a GTFS feed from an io.Reader (ZIP format).
	ValidateReader(reader io.Reader) (*ValidationReport, error)
	
	// ValidateReaderWithContext validates a reader with cancellation support.
	ValidateReaderWithContext(ctx context.Context, reader io.Reader) (*ValidationReport, error)
	
	// ValidateFileStream validates with streaming notice delivery.
	ValidateFileStream(path string, callback NoticeCallback) (*ValidationReport, error)
	
	// ValidateFileStreamWithContext validates with streaming and cancellation.
	ValidateFileStreamWithContext(ctx context.Context, path string, callback NoticeCallback) (*ValidationReport, error)
}

// Config contains configuration options for the validator.
type Config struct {
	// CountryCode for phone number validation (e.g., "US", "GB").
	CountryCode string
	
	// CurrentDate for date validation (defaults to today).
	CurrentDate time.Time
	
	// MaxMemory limits memory usage in bytes (0 = no limit).
	MaxMemory int64
	
	// ParallelWorkers for concurrent validation (0 = auto).
	ParallelWorkers int
	
	// ValidatorVersion for reports.
	ValidatorVersion string
	
	// ProgressCallback is called during validation to report progress.
	ProgressCallback func(progress ProgressInfo)
	
	// ValidationMode configures which validators to run.
	ValidationMode ValidationMode
	
	// MaxNoticesPerType limits notices per type (0 = no limit).
	MaxNoticesPerType int
}

// ValidationMode defines preset validation configurations.
type ValidationMode string

const (
	// ValidationModePerformance runs only essential validators for speed.
	ValidationModePerformance ValidationMode = "performance"
	
	// ValidationModeDefault runs standard validators.
	ValidationModeDefault ValidationMode = "default"
	
	// ValidationModeComprehensive runs all validators including expensive ones.
	ValidationModeComprehensive ValidationMode = "comprehensive"
)

// ProgressInfo contains information about validation progress.
type ProgressInfo struct {
	// CurrentValidator is the name of the currently running validator.
	CurrentValidator string
	
	// TotalValidators is the total number of validators to run.
	TotalValidators int
	
	// CompletedValidators is the number of validators completed.
	CompletedValidators int
	
	// PercentComplete is the overall completion percentage.
	PercentComplete float64
	
	// ElapsedTime is the time elapsed since validation started.
	ElapsedTime time.Duration
}

// ValidationReport contains the complete validation results.
type ValidationReport struct {
	// Summary contains high-level information about the validation.
	Summary Summary `json:"summary"`
	
	// Notices contains all validation notices grouped by type.
	Notices []NoticeGroup `json:"notices"`
	
	// mu protects concurrent access to the report.
	mu sync.RWMutex
}

// Summary contains summary information about the validation.
type Summary struct {
	// ValidatorVersion is the version of the validator used.
	ValidatorVersion string `json:"validatorVersion"`
	
	// ValidationTime is the total validation time in seconds.
	ValidationTime float64 `json:"validationTimeSeconds"`
	
	// Date is the validation timestamp.
	Date string `json:"date"`
	
	// FeedInfo contains information about the validated feed.
	FeedInfo FeedInfo `json:"feedInfo"`
	
	// Counts contains notice counts by severity.
	Counts NoticeCounts `json:"counts"`
}

// FeedInfo contains information about the validated GTFS feed.
type FeedInfo struct {
	// FeedPath is the path to the validated feed.
	FeedPath string `json:"feedPath"`
	
	// AgencyCount is the number of agencies in the feed.
	AgencyCount int `json:"agencyCount"`
	
	// RouteCount is the number of routes in the feed.
	RouteCount int `json:"routeCount"`
	
	// TripCount is the number of trips in the feed.
	TripCount int `json:"tripCount"`
	
	// StopCount is the number of stops in the feed.
	StopCount int `json:"stopCount"`
	
	// StopTimeCount is the number of stop times in the feed.
	StopTimeCount int `json:"stopTimeCount"`
	
	// ServiceDateFrom is the start date of service.
	ServiceDateFrom string `json:"serviceDateFrom,omitempty"`
	
	// ServiceDateTo is the end date of service.
	ServiceDateTo string `json:"serviceDateTo,omitempty"`
}

// NoticeCounts contains counts of notices by severity.
type NoticeCounts struct {
	// Errors is the count of error-level notices.
	Errors int `json:"errors"`
	
	// Warnings is the count of warning-level notices.
	Warnings int `json:"warnings"`
	
	// Infos is the count of info-level notices.
	Infos int `json:"infos"`
	
	// Total is the total count of all notices.
	Total int `json:"total"`
}

// NoticeGroup represents a group of notices with the same type.
type NoticeGroup struct {
	// Code is the notice type code (e.g., "missing_required_field").
	Code string `json:"code"`
	
	// Severity is the notice severity (ERROR, WARNING, INFO).
	Severity string `json:"severity"`
	
	// TotalNotices is the total count of this notice type.
	TotalNotices int `json:"totalNotices"`
	
	// SampleNotices contains sample instances of this notice.
	SampleNotices []map[string]interface{} `json:"sampleNotices"`
}

// HasErrors returns true if the report contains any errors.
func (r *ValidationReport) HasErrors() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.Summary.Counts.Errors > 0
}

// HasWarnings returns true if the report contains any warnings.
func (r *ValidationReport) HasWarnings() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.Summary.Counts.Warnings > 0
}

// ErrorCount returns the number of errors in the report.
func (r *ValidationReport) ErrorCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.Summary.Counts.Errors
}

// WarningCount returns the number of warnings in the report.
func (r *ValidationReport) WarningCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.Summary.Counts.Warnings
}

// Option is a functional option for configuring the validator.
type Option func(*Config)

// WithCountryCode sets the country code for validation.
func WithCountryCode(code string) Option {
	return func(c *Config) {
		c.CountryCode = code
	}
}

// WithCurrentDate sets the current date for validation.
func WithCurrentDate(date time.Time) Option {
	return func(c *Config) {
		c.CurrentDate = date
	}
}

// WithMaxMemory sets the maximum memory usage in bytes.
func WithMaxMemory(bytes int64) Option {
	return func(c *Config) {
		c.MaxMemory = bytes
	}
}

// WithParallelWorkers sets the number of parallel workers.
func WithParallelWorkers(workers int) Option {
	return func(c *Config) {
		c.ParallelWorkers = workers
	}
}

// WithProgressCallback sets a progress callback function.
func WithProgressCallback(callback func(ProgressInfo)) Option {
	return func(c *Config) {
		c.ProgressCallback = callback
	}
}

// WithValidationMode sets the validation mode.
func WithValidationMode(mode ValidationMode) Option {
	return func(c *Config) {
		c.ValidationMode = mode
	}
}

// WithMaxNoticesPerType sets the maximum notices per type.
func WithMaxNoticesPerType(max int) Option {
	return func(c *Config) {
		c.MaxNoticesPerType = max
	}
}

// New creates a new GTFS validator with the given options.
func New(opts ...Option) Validator {
	config := &Config{
		CountryCode:       "US",
		CurrentDate:       time.Now(),
		ParallelWorkers:   4,
		ValidatorVersion:  "1.0.0",
		ValidationMode:    ValidationModeDefault,
		MaxNoticesPerType: 100,
	}
	
	for _, opt := range opts {
		opt(config)
	}
	
	return &validatorImpl{
		config: config,
	}
}

// validatorImpl is the concrete implementation of the Validator interface.
type validatorImpl struct {
	config *Config
	mu     sync.Mutex
}

// ValidateFile validates a GTFS feed from a file path.
func (v *validatorImpl) ValidateFile(path string) (*ValidationReport, error) {
	return v.ValidateFileWithContext(context.Background(), path)
}

// ValidateReader validates a GTFS feed from an io.Reader.
func (v *validatorImpl) ValidateReader(reader io.Reader) (*ValidationReport, error) {
	return v.ValidateReaderWithContext(context.Background(), reader)
}

// ValidateFileStream validates with streaming notice delivery.
func (v *validatorImpl) ValidateFileStream(path string, callback NoticeCallback) (*ValidationReport, error) {
	return v.ValidateFileStreamWithContext(context.Background(), path, callback)
}

// ValidateFileStreamWithContext validates with streaming and cancellation.
func (v *validatorImpl) ValidateFileStreamWithContext(ctx context.Context, path string, callback NoticeCallback) (*ValidationReport, error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	
	// This would use a modified implementation.go that calls the callback
	// for each notice group as it's created during validation
	return nil, fmt.Errorf("streaming validation not implemented yet")
}