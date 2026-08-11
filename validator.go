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
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/theoremus-urban-solutions/gtfs-validator/report"
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

	// MaxNoticesPerType limits notices per type (0 = no limit, the default).
	//
	// A cap silently discards notices once reached, and since notices arrive
	// in file order rather than severity order, the discarded ones can be the
	// errors. Only set this if you knowingly want a truncated report.
	MaxNoticesPerType int
}

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

// Severity levels as they appear in reports.
const (
	SeverityLevelError   = "ERROR"
	SeverityLevelWarning = "WARNING"
	SeverityLevelInfo    = "INFO"
)

// NoticeGroup represents a group of notices with the same type.
//
// A group has no single severity. One code can produce notices of different
// severities, so a single label would hide all but one of them. Use
// SeverityCounts for the breakdown, HighestSeverity for sorting and display,
// and the "severity" key on each sample notice for the individual instance.
type NoticeGroup struct {
	// Code is the notice type code (e.g., "missing_required_field").
	Code string `json:"code"`

	// SeverityCounts breaks the group down by severity level.
	SeverityCounts NoticeCounts `json:"severityCounts"`

	// Description is a comprehensive, user-friendly description of the notice type.
	// This helps feed producers understand and fix validation issues.
	Description string `json:"description"`

	// GTFSReference provides a link to the relevant GTFS specification section.
	GTFSReference string `json:"gtfsReference,omitempty"`

	// AffectedFiles lists the GTFS files that may contain this issue.
	AffectedFiles []string `json:"affectedFiles,omitempty"`

	// AffectedFields lists the specific fields that may have this issue.
	AffectedFields []string `json:"affectedFields,omitempty"`

	// ExampleFix provides a concrete example of how to fix this issue.
	ExampleFix string `json:"exampleFix,omitempty"`

	// TotalNotices is the total count of this notice type.
	TotalNotices int `json:"totalNotices"`

	// SampleNotices contains sample instances of this notice. Each sample
	// carries its own "severity", plus "file" and "line" where they could be
	// determined. Samples are ordered most severe first.
	SampleNotices []map[string]interface{} `json:"sampleNotices"`
}

// HighestSeverity returns the most severe level present in the group. Use it
// for sorting and colouring, never as a label for the group as a whole.
func (g *NoticeGroup) HighestSeverity() string {
	switch {
	case g.SeverityCounts.Errors > 0:
		return SeverityLevelError
	case g.SeverityCounts.Warnings > 0:
		return SeverityLevelWarning
	default:
		return SeverityLevelInfo
	}
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

// WithMaxNoticesPerType sets the maximum notices per type.
func WithMaxNoticesPerType(max int) Option {
	return func(c *Config) {
		c.MaxNoticesPerType = max
	}
}

// New creates a new GTFS validator with the given options.
//
// Every check the validator implements runs on every feed. There is no mode or
// preset to select a subset: a rule worth shipping is worth running, and a
// caller cannot tell which of the rules a preset drops would have been the one
// that mattered for their feed.
func New(opts ...Option) Validator {
	config := &Config{
		// No country unless one is asked for, which is what the canonical
		// validator does. The country decides how agency_phone is measured,
		// and assuming the wrong one fails working numbers: a Bulgarian
		// service line read as a US number is short by two digits. Left empty,
		// a number only has to be dialable somewhere. Pass BG to hold a
		// Bulgarian feed to the Bulgarian numbering plan.
		CountryCode:       "",
		CurrentDate:       time.Now(),
		ParallelWorkers:   4,
		ValidatorVersion:  "1.0.0",
		MaxNoticesPerType: 0, // No limit: capping per type silently drops findings
	}

	for _, opt := range opts {
		opt(config)
	}

	// Validate and sanitize configuration
	if err := validateConfig(config); err != nil {
		// For backward compatibility, we'll use default values for invalid config
		// In a future version, this could return an error
		sanitizeConfig(config)
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

	// Check context cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	startTime := time.Now()

	// Create internal validator with streaming enabled
	internalConfig := v.createInternalConfig()
	validationConfig := v.createValidationConfig()
	internalValidator := newInternalValidatorWithStreaming(internalConfig, validationConfig, callback)

	// Set up progress reporting
	if v.config.ProgressCallback != nil {
		internalValidator.progressCallback = v.config.ProgressCallback
	}

	// Validate based on file type
	var internalReport *report.ValidationReport
	var err error

	if strings.HasSuffix(strings.ToLower(path), ".zip") {
		internalReport, err = internalValidator.ValidateZipWithContext(ctx, path)
	} else {
		// Check if it's a directory
		info, statErr := os.Stat(path)
		if statErr != nil {
			return nil, fmt.Errorf("cannot access path: %w", statErr)
		}
		if !info.IsDir() {
			return nil, fmt.Errorf("path must be a ZIP file or directory")
		}
		internalReport, err = internalValidator.ValidateDirectoryWithContext(ctx, path)
	}

	if err != nil {
		return nil, err
	}

	// Convert internal report to public API format
	return v.convertReport(internalReport, time.Since(startTime)), nil
}

// validateConfig validates the configuration and returns an error if invalid.
func validateConfig(config *Config) error {
	var errs []error

	// Validate CountryCode: either a 2-letter ISO code, or empty for a feed
	// whose country is not being asserted.
	if config.CountryCode != "" && len(config.CountryCode) != 2 {
		errs = append(errs, fmt.Errorf("CountryCode must be a 2-letter ISO code, got: %s", config.CountryCode))
	}

	// Validate CurrentDate (should not be in the distant future)
	if config.CurrentDate.After(time.Now().AddDate(10, 0, 0)) {
		errs = append(errs, fmt.Errorf("CurrentDate is too far in the future: %v", config.CurrentDate))
	}

	// Validate MaxMemory (should be reasonable)
	if config.MaxMemory < 0 {
		errs = append(errs, fmt.Errorf("MaxMemory cannot be negative: %d", config.MaxMemory))
	}
	if config.MaxMemory > 0 && config.MaxMemory < 10*1024*1024 { // Less than 10MB
		errs = append(errs, fmt.Errorf("MaxMemory is too small (minimum 10MB): %d", config.MaxMemory))
	}

	// Validate ParallelWorkers (should be reasonable)
	if config.ParallelWorkers < 0 {
		errs = append(errs, fmt.Errorf("ParallelWorkers cannot be negative: %d", config.ParallelWorkers))
	}
	if config.ParallelWorkers > 100 {
		errs = append(errs, fmt.Errorf("ParallelWorkers is too high (maximum 100): %d", config.ParallelWorkers))
	}

	// Validate ValidatorVersion (should not be empty)
	if strings.TrimSpace(config.ValidatorVersion) == "" {
		errs = append(errs, errors.New("ValidatorVersion cannot be empty"))
	}

	// Validate MaxNoticesPerType (should be reasonable)
	if config.MaxNoticesPerType < 0 {
		errs = append(errs, fmt.Errorf("MaxNoticesPerType cannot be negative: %d", config.MaxNoticesPerType))
	}
	if config.MaxNoticesPerType > 10000 {
		errs = append(errs, fmt.Errorf("MaxNoticesPerType is too high (maximum 10000): %d", config.MaxNoticesPerType))
	}

	// Combine errors if any
	if len(errs) > 0 {
		var errStr string
		for i, err := range errs {
			if i > 0 {
				errStr += "; "
			}
			errStr += err.Error()
		}
		return errors.New("configuration validation failed: " + errStr)
	}

	return nil
}

// sanitizeConfig fixes invalid configuration values by setting them to defaults.
func sanitizeConfig(config *Config) {
	// Sanitize CountryCode. A value we cannot read becomes no country rather
	// than a guessed one: measuring a number against the wrong country's plan
	// fails working feeds, which is worse than not measuring it at all.
	if len(config.CountryCode) != 2 {
		config.CountryCode = ""
	}

	// Sanitize CurrentDate
	if config.CurrentDate.After(time.Now().AddDate(10, 0, 0)) {
		config.CurrentDate = time.Now()
	}

	// Sanitize MaxMemory
	if config.MaxMemory < 0 {
		config.MaxMemory = 0 // 0 means no limit
	} else if config.MaxMemory > 0 && config.MaxMemory < 10*1024*1024 {
		config.MaxMemory = 10 * 1024 * 1024 // 10MB minimum
	}

	// Sanitize ParallelWorkers
	switch {
	case config.ParallelWorkers < 0:
		config.ParallelWorkers = 1
	case config.ParallelWorkers > 100:
		config.ParallelWorkers = 100
	case config.ParallelWorkers == 0:
		config.ParallelWorkers = 4 // Default
	}

	// Sanitize ValidatorVersion
	if strings.TrimSpace(config.ValidatorVersion) == "" {
		config.ValidatorVersion = "1.0.0"
	}

	// Sanitize MaxNoticesPerType
	if config.MaxNoticesPerType < 0 {
		config.MaxNoticesPerType = 0 // 0 means no limit
	} else if config.MaxNoticesPerType > 10000 {
		config.MaxNoticesPerType = 10000
	}
	// 0 is valid (no limit), no action needed
}
