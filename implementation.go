package gtfsvalidator

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/parser"
	"github.com/theoremus-urban-solutions/gtfs-validator/report"
	"github.com/theoremus-urban-solutions/gtfs-validator/validator"
	"github.com/theoremus-urban-solutions/gtfs-validator/validator/accessibility"
	"github.com/theoremus-urban-solutions/gtfs-validator/validator/business"
	"github.com/theoremus-urban-solutions/gtfs-validator/validator/core"
	"github.com/theoremus-urban-solutions/gtfs-validator/validator/entity"
	"github.com/theoremus-urban-solutions/gtfs-validator/validator/fare"
	"github.com/theoremus-urban-solutions/gtfs-validator/validator/meta"
	"github.com/theoremus-urban-solutions/gtfs-validator/validator/relationship"
)

// ValidateFileWithContext implements the main validation logic with context support.
func (v *validatorImpl) ValidateFileWithContext(ctx context.Context, path string) (*ValidationReport, error) {
	// Check context cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	startTime := time.Now()

	// Create internal validator with configuration
	internalConfig := v.createInternalConfig()
	validationConfig := v.createValidationConfig()
	internalValidator := newInternalValidator(internalConfig, validationConfig)

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

// ValidateReaderWithContext validates a GTFS feed from an io.Reader.
func (v *validatorImpl) ValidateReaderWithContext(ctx context.Context, reader io.Reader) (*ValidationReport, error) {
	// Check context cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Create temporary file for the ZIP content
	tmpFile, err := os.CreateTemp("", "gtfs-*.zip")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer func() {
		if err := os.Remove(tmpFile.Name()); err != nil {
			log.Printf("Warning: failed to remove temp file: %v", err)
		}
	}()
	defer func() {
		if closeErr := tmpFile.Close(); closeErr != nil {
			log.Printf("Warning: failed to close %v", closeErr)
		}
	}()

	// Copy reader content to temporary file
	_, err = io.Copy(tmpFile, reader)
	if err != nil {
		return nil, fmt.Errorf("failed to write temporary file: %w", err)
	}

	// Close the file to ensure all data is written
	if err := tmpFile.Close(); err != nil {
		return nil, fmt.Errorf("failed to close temporary file: %w", err)
	}

	// Validate the temporary file
	return v.ValidateFileWithContext(ctx, tmpFile.Name())
}

// createInternalConfig creates the internal validator configuration.
func (v *validatorImpl) createInternalConfig() Config {
	return Config{
		CountryCode:      v.config.CountryCode,
		CurrentDate:      v.config.CurrentDate,
		MaxMemory:        v.config.MaxMemory,
		ParallelWorkers:  v.config.ParallelWorkers,
		ValidatorVersion: v.config.ValidatorVersion,
	}
}

// createValidationConfig creates the validation configuration.
func (v *validatorImpl) createValidationConfig() validationConfig {
	return validationConfig{MaxNoticesPerType: v.config.MaxNoticesPerType}
}

// convertReport converts internal report format to public API format.
func (v *validatorImpl) convertReport(internal *report.ValidationReport, elapsed time.Duration) *ValidationReport {
	// Group notices by code
	noticeGroups := make(map[string]*NoticeGroup)

	for _, n := range internal.Notices {
		if group, exists := noticeGroups[n.Code]; exists {
			// This shouldn't happen with the current implementation
			// but handle it gracefully
			group.TotalNotices += n.TotalNotices
			group.SeverityCounts.Errors += n.SeverityCounts.Errors
			group.SeverityCounts.Warnings += n.SeverityCounts.Warnings
			group.SeverityCounts.Infos += n.SeverityCounts.Infos
			group.SeverityCounts.Total += n.SeverityCounts.Total
			group.SampleNotices = append(group.SampleNotices, n.SampleNotices...)
		} else {
			enhanced := GetEnhancedNoticeDescription(n.Code)
			noticeGroups[n.Code] = &NoticeGroup{
				Code: n.Code,
				SeverityCounts: NoticeCounts{
					Errors:   n.SeverityCounts.Errors,
					Warnings: n.SeverityCounts.Warnings,
					Infos:    n.SeverityCounts.Infos,
					Total:    n.SeverityCounts.Total,
				},
				Description:    enhanced.Description,
				GTFSReference:  enhanced.GTFSReference,
				AffectedFiles:  affectedFiles(n.Code, enhanced),
				AffectedFields: enhanced.AffectedFields,
				ExampleFix:     enhanced.ExampleFix,
				TotalNotices:   n.TotalNotices,
				SampleNotices:  n.SampleNotices,
			}
		}
	}

	// Convert map to slice
	notices := make([]NoticeGroup, 0, len(noticeGroups))
	for _, group := range noticeGroups {
		notices = append(notices, *group)
	}

	return &ValidationReport{
		Summary: Summary{
			ValidatorVersion: internal.Summary.ValidatorVersion,
			ValidationTime:   internal.Summary.ValidationTime,
			Date:             internal.Summary.Date,
			FeedInfo: FeedInfo{
				FeedPath:        internal.Summary.FeedInfo.FeedPath,
				AgencyCount:     internal.Summary.FeedInfo.AgencyCount,
				RouteCount:      internal.Summary.FeedInfo.RouteCount,
				TripCount:       internal.Summary.FeedInfo.TripCount,
				StopCount:       internal.Summary.FeedInfo.StopCount,
				StopTimeCount:   internal.Summary.FeedInfo.StopTimeCount,
				ServiceDateFrom: internal.Summary.FeedInfo.ServiceDateFrom,
				ServiceDateTo:   internal.Summary.FeedInfo.ServiceDateTo,
			},
			Counts: NoticeCounts{
				Errors:   internal.Summary.Counts.Errors,
				Warnings: internal.Summary.Counts.Warnings,
				Infos:    internal.Summary.Counts.Infos,
				Total:    internal.Summary.Counts.Total,
			},
		},
		Notices: notices,
	}
}

// Internal types that mirror the existing implementation

// validationConfig is what the internal validator needs beyond the public
// Config. The per-category enable flags are gone with the validation modes:
// every registered validator runs on every feed.
type validationConfig struct {
	MaxNoticesPerType int
}

// internalValidator wraps the existing validator implementation.
type internalValidator struct {
	config           Config
	validationConfig validationConfig
	noticeContainer  *notice.NoticeContainer
	feedLoader       *parser.FeedLoader
	validators       []validator.Validator
	progressCallback func(ProgressInfo)
	noticeCallback   NoticeCallback // For streaming validation
	streamedCount    int            // Track how many notices we've already streamed
	streamMutex      sync.Mutex     // Protect streaming state in parallel mode
}

// newInternalValidator creates a new internal validator.
func newInternalValidator(config Config, validationConfig validationConfig) *internalValidator {
	var noticeContainer *notice.NoticeContainer
	if validationConfig.MaxNoticesPerType > 0 {
		noticeContainer = notice.NewNoticeContainerWithLimit(validationConfig.MaxNoticesPerType)
	} else {
		noticeContainer = notice.NewNoticeContainer()
	}

	return &internalValidator{
		config:           config,
		validationConfig: validationConfig,
		noticeContainer:  noticeContainer,
	}
}

// newInternalValidatorWithStreaming creates a new internal validator with streaming support.
func newInternalValidatorWithStreaming(config Config, validationConfig validationConfig, callback NoticeCallback) *internalValidator {
	var noticeContainer *notice.NoticeContainer
	if validationConfig.MaxNoticesPerType > 0 {
		noticeContainer = newStreamingNoticeContainerWithLimit(validationConfig.MaxNoticesPerType, callback)
	} else {
		noticeContainer = newStreamingNoticeContainer(callback)
	}

	return &internalValidator{
		config:           config,
		validationConfig: validationConfig,
		noticeContainer:  noticeContainer,
		noticeCallback:   callback,
	}
}

// ValidateZipWithContext validates a ZIP file with context support.
func (v *internalValidator) ValidateZipWithContext(ctx context.Context, zipPath string) (*report.ValidationReport, error) {
	startTime := time.Now()

	// Load the feed
	loader, err := parser.LoadFromZip(zipPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load zip file: %w", err)
	}
	defer func() {
		if err := loader.Close(); err != nil {
			log.Printf("Warning: failed to close loader: %v", err)
		}
	}()

	v.feedLoader = loader

	// Run validation with context
	feedInfo, err := v.validateWithContext(ctx)
	if err != nil {
		return nil, err
	}
	feedInfo.FeedPath = zipPath

	// Generate report
	validationTime := time.Since(startTime).Seconds()
	reportGen := report.NewReportGenerator(v.config.ValidatorVersion)
	return reportGen.GenerateReport(v.noticeContainer, feedInfo, validationTime), nil
}

// ValidateDirectoryWithContext validates a directory with context support.
func (v *internalValidator) ValidateDirectoryWithContext(ctx context.Context, dirPath string) (*report.ValidationReport, error) {
	startTime := time.Now()

	// Load the feed
	loader, err := parser.LoadFromDirectory(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load directory: %w", err)
	}
	defer func() {
		if err := loader.Close(); err != nil {
			log.Printf("Warning: failed to close loader: %v", err)
		}
	}()

	v.feedLoader = loader

	// Run validation with context
	feedInfo, err := v.validateWithContext(ctx)
	if err != nil {
		return nil, err
	}
	feedInfo.FeedPath = dirPath

	// Generate report
	validationTime := time.Since(startTime).Seconds()
	reportGen := report.NewReportGenerator(v.config.ValidatorVersion)
	return reportGen.GenerateReport(v.noticeContainer, feedInfo, validationTime), nil
}

// validateWithContext performs the actual validation with context support.
func (v *internalValidator) validateWithContext(ctx context.Context) (report.FeedInfo, error) {
	startTime := time.Now()
	feedInfo := report.FeedInfo{}

	// Check for required files
	v.checkRequiredFiles()

	// Initialize validators
	v.initializeValidators()

	// Run validators with context and progress reporting
	validatorConfig := validator.Config{
		CountryCode:     v.config.CountryCode,
		CurrentDate:     v.config.CurrentDate,
		MaxMemory:       v.config.MaxMemory,
		ParallelWorkers: v.config.ParallelWorkers,
	}

	// Two passes. The foundation checks decide which tables actually loaded;
	// only then can the rest be stood down against that answer. Running them
	// together would judge a table by a verdict not yet reached.
	foundation, rest := v.partitionValidators()
	v.validators = foundation
	if err := v.runValidators(ctx, validatorConfig, startTime, len(foundation)); err != nil {
		return feedInfo, err
	}
	v.validators = v.skipValidatorsWithUnusableFiles(rest, v.filesWithRowErrors())

	totalValidators := len(v.validators)
	if err := v.runValidators(ctx, validatorConfig, startTime, totalValidators); err != nil {
		return feedInfo, err
	}

	// Final progress report
	if v.progressCallback != nil {
		v.progressCallback(ProgressInfo{
			CurrentValidator:    "Complete",
			TotalValidators:     totalValidators,
			CompletedValidators: totalValidators,
			PercentComplete:     100,
			ElapsedTime:         time.Since(startTime),
		})
	}

	// Collect feed statistics
	feedInfo = v.collectFeedStatistics()

	return feedInfo, nil
}

// foundationValidators are the checks that establish whether each table loaded:
// its presence, its header, its columns and the types of its values. They read
// only the file in front of them, so nothing they report depends on another
// table having survived, and they must run before anything is stood down.
var foundationValidators = map[string]bool{
	"*core.MissingFilesValidator":              true,
	"*core.EmptyFileValidator":                 true,
	"*core.UnknownFileValidator":               true,
	"*core.DuplicateHeaderValidator":           true,
	"*core.MissingColumnValidator":             true,
	"*core.RequiredFieldValidator":             true,
	"*core.FieldFormatValidator":               true,
	"*core.CoordinateValidator":                true,
	"*core.DuplicateKeyValidator":              true,
	"*core.InvalidRowValidator":                true,
	"*core.FieldTypeValidator":                 true,
	"*validator.FileStructureValidator":        true,
	"*core.LeadingTrailingWhitespaceValidator": true,
}

// partitionValidators splits the registry into the foundation checks and the
// rest, preserving order within each.
func (v *internalValidator) partitionValidators() (foundation []validator.Validator, rest []validator.Validator) {
	for _, validatorImpl := range v.validators {
		if foundationValidators[fmt.Sprintf("%T", validatorImpl)] {
			foundation = append(foundation, validatorImpl)
		} else {
			rest = append(rest, validatorImpl)
		}
	}
	return foundation, rest
}

// filesWithRowErrors returns the tables that produced a row-level error while
// being read.
//
// A value that does not parse as its declared type makes the row it sits in
// unusable, and canonical treats one such row as poisoning the whole table for
// every check that depends on it — a single `friday=ZZZ` in calendar.txt stops
// it reporting on the service window at all. Only errors that name a row count:
// a missing file or a missing column is a fact about the table's shape, handled
// by the file state, not about a row inside it.
func (v *internalValidator) filesWithRowErrors() map[string]bool {
	poisoned := make(map[string]bool)
	for _, n := range v.noticeContainer.GetNotices() {
		if !parseErrorCodes[n.Code()] {
			continue
		}
		context := n.Context()
		if _, hasRow := notice.LineNumber(context); !hasRow {
			continue
		}
		if filename, ok := notice.FileName(n.Code(), context); ok {
			poisoned[filename] = true
		}
	}
	return poisoned
}

// parseErrorCodes are the errors that mean a row could not be read as the types
// it declares. Only these poison a table.
//
// The distinction is between a value the feed could not express and a value it
// expressed and got wrong. `friday=ZZZ` is the first: there is no integer there,
// so nothing downstream can reason about that service at all. A start_date after
// its end_date is the second — both dates parsed, the row is legible, and the
// checks that read it still have something true to say. Treating the second as
// unreadable silences rules canonical still reports.
var parseErrorCodes = map[string]bool{
	"invalid_integer":         true,
	"invalid_float":           true,
	"invalid_date":            true,
	"invalid_time":            true,
	"invalid_color":           true,
	"invalid_url":             true,
	"invalid_email":           true,
	"invalid_phone_number":    true,
	"invalid_timezone":        true,
	"invalid_language_code":   true,
	"invalid_currency_code":   true,
	"invalid_currency_amount": true,
	"missing_required_field":  true,
	"invalid_row_length":      true,
}

// runValidators runs whatever is currently registered, in parallel when that is
// configured and there is enough to spread.
func (v *internalValidator) runValidators(ctx context.Context, validatorConfig validator.Config, startTime time.Time, total int) error {
	if total == 0 {
		return nil
	}
	if v.config.ParallelWorkers > 1 && total > 1 {
		return v.runValidatorsParallel(ctx, validatorConfig, startTime, total)
	}
	return v.runValidatorsSequential(ctx, validatorConfig, startTime, total)
}

// skipValidatorsWithUnusableFiles drops the validators whose source files did
// not load, recording each one so the report says what was not checked and why.
//
// Filtering here rather than inside each run loop keeps the sequential and
// parallel paths identical, and means the skip is decided once per run rather
// than re-derived per worker.
func (v *internalValidator) skipValidatorsWithUnusableFiles(candidates []validator.Validator, poisoned map[string]bool) []validator.Validator {
	runnable := candidates[:0]
	for _, validatorImpl := range candidates {
		name := fmt.Sprintf("%T", validatorImpl)
		skipped := false
		for _, filename := range requiredFiles[name] {
			state := v.feedLoader.FileState(filename)
			if !state.LoadFailed() && !poisoned[filename] {
				continue
			}
			if poisoned[filename] && !state.LoadFailed() {
				state = parser.FileStateInvalidRows
			}
			v.noticeContainer.AddNotice(notice.NewValidatorSkippedNotice(
				strings.TrimPrefix(name, "*"), filename, state.Reason(),
			))
			skipped = true
			break
		}
		if !skipped {
			runnable = append(runnable, validatorImpl)
		}
	}
	return runnable
}

// runValidatorsSequential runs validators one after another (thread-safe).
func (v *internalValidator) runValidatorsSequential(ctx context.Context, validatorConfig validator.Config, startTime time.Time, totalValidators int) error {
	for i, validatorImpl := range v.validators {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Report progress if callback is set
		if v.progressCallback != nil {
			v.progressCallback(ProgressInfo{
				CurrentValidator:    fmt.Sprintf("%T", validatorImpl),
				TotalValidators:     totalValidators,
				CompletedValidators: i,
				PercentComplete:     float64(i) / float64(totalValidators) * 100,
				ElapsedTime:         time.Since(startTime),
			})
		}

		// Run validator with error recovery
		func() {
			defer func() {
				if r := recover(); r != nil {
					// Log the panic but continue with other validators
					v.noticeContainer.AddNotice(notice.NewValidatorErrorNotice(
						fmt.Sprintf("%T", validatorImpl),
						fmt.Sprintf("Validator panic: %v", r),
					))
				}
			}()

			validatorImpl.Validate(v.feedLoader, v.noticeContainer, validatorConfig)

			// Stream notice groups after each validator if streaming is enabled
			if v.noticeCallback != nil {
				v.streamNoticeGroups()
			}
		}()
	}
	return nil
}

// runValidatorsParallel runs validators in parallel using worker goroutines (thread-safe).
func (v *internalValidator) runValidatorsParallel(ctx context.Context, validatorConfig validator.Config, startTime time.Time, totalValidators int) error {
	workers := v.config.ParallelWorkers
	if workers > totalValidators {
		workers = totalValidators
	}

	// Create channels for work distribution
	validatorChan := make(chan validator.Validator, totalValidators)

	// Populate work queue
	for _, validatorImpl := range v.validators {
		validatorChan <- validatorImpl
	}
	close(validatorChan)

	var wg sync.WaitGroup
	var completed int64

	// Start worker goroutines
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for validatorImpl := range validatorChan {
				// Check context cancellation
				select {
				case <-ctx.Done():
					return
				default:
				}

				// Run validator with error recovery
				func() {
					defer func() {
						if r := recover(); r != nil {
							// Log the panic but continue with other validators
							// NoticeContainer is thread-safe
							v.noticeContainer.AddNotice(notice.NewValidatorErrorNotice(
								fmt.Sprintf("%T", validatorImpl),
								fmt.Sprintf("Validator panic: %v", r),
							))
						}
					}()

					validatorImpl.Validate(v.feedLoader, v.noticeContainer, validatorConfig)

					// Stream notice groups after each validator if streaming is enabled
					// Note: In parallel mode, this will stream notices as they become available
					if v.noticeCallback != nil {
						v.streamNoticeGroups()
					}
				}()

				// Update progress atomically
				completedCount := atomic.AddInt64(&completed, 1)
				if v.progressCallback != nil {
					v.progressCallback(ProgressInfo{
						CurrentValidator:    fmt.Sprintf("%T", validatorImpl),
						TotalValidators:     totalValidators,
						CompletedValidators: int(completedCount),
						PercentComplete:     float64(completedCount) / float64(totalValidators) * 100,
						ElapsedTime:         time.Since(startTime),
					})
				}
			}
		}()
	}

	// Wait for all workers to complete or context cancellation
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
		return nil
	}
}

// checkRequiredFiles checks for required GTFS files.
func (v *internalValidator) checkRequiredFiles() {
	for _, filename := range parser.RequiredFiles {
		if !v.feedLoader.HasFile(filename) {
			v.noticeContainer.AddNotice(
				notice.NewMissingRequiredFileNotice(filename),
			)
		}
	}
}

// collectFeedStatistics collects statistics about the GTFS feed.
func (v *internalValidator) collectFeedStatistics() report.FeedInfo {
	feedInfo := report.FeedInfo{}

	// Count agencies
	if v.feedLoader.HasFile("agency.txt") {
		feedInfo.AgencyCount = v.countRowsInFile("agency.txt")
	}

	// Count routes
	if v.feedLoader.HasFile("routes.txt") {
		feedInfo.RouteCount = v.countRowsInFile("routes.txt")
	}

	// Count trips
	if v.feedLoader.HasFile("trips.txt") {
		feedInfo.TripCount = v.countRowsInFile("trips.txt")
	}

	// Count stops
	if v.feedLoader.HasFile("stops.txt") {
		feedInfo.StopCount = v.countRowsInFile("stops.txt")
	}

	// Count stop times
	if v.feedLoader.HasFile("stop_times.txt") {
		feedInfo.StopTimeCount = v.countRowsInFile("stop_times.txt")
	}

	// Extract service date range from feed_info.txt if available
	if v.feedLoader.HasFile("feed_info.txt") {
		v.extractServiceDates(&feedInfo)
	}

	return feedInfo
}

// countRowsInFile counts the number of data rows in a file (excluding header).
func (v *internalValidator) countRowsInFile(filename string) int {
	reader, err := v.feedLoader.GetFile(filename)
	if err != nil {
		return 0
	}

	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close reader %v", closeErr)
		}
	}()

	csvFile, err := parser.NewCSVFile(reader, filename)
	if err != nil {
		return 0
	}

	err = csvFile.ReadAll()
	if err != nil {
		return 0
	}

	return csvFile.RowCount()
}

// extractServiceDates extracts service date range from feed_info.txt.
func (v *internalValidator) extractServiceDates(feedInfo *report.FeedInfo) {
	reader, err := v.feedLoader.GetFile("feed_info.txt")
	if err != nil {
		return
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close %v", closeErr)
		}
	}()

	csvFile, err := parser.NewCSVFile(reader, "feed_info.txt")
	if err != nil {
		return
	}

	err = csvFile.ReadAll()
	if err != nil || len(csvFile.Rows) == 0 {
		return
	}

	// Get the first (and usually only) row
	row := csvFile.Rows[0]

	if startDate, exists := row.Values["feed_start_date"]; exists {
		feedInfo.ServiceDateFrom = startDate
	}

	if endDate, exists := row.Values["feed_end_date"]; exists {
		feedInfo.ServiceDateTo = endDate
	}
}

// requiredFiles lists, for the validators that cannot work without them, the
// files whose contents they read. A validator is skipped when one of its files
// is absent, empty, unparseable or missing the ids it is joined on, because
// every finding it would produce in that state restates the same defect once
// per row: with trips.txt emptied, the route consistency check reported all 18
// routes as having no trips, on top of the one empty_file error that explains
// it.
//
// Only total dependencies belong here. A validator that reads a file for extra
// detail, or that has something useful to say about the file itself, must keep
// running — the point is to suppress restatement, not coverage.
var requiredFiles = map[string][]string{
	"*business.DateTripsValidator":             {"calendar.txt", "trips.txt"},
	"*entity.ServiceValidationValidator":       {"calendar.txt"},
	"*business.TripUsabilityValidator":         {"trips.txt", "stop_times.txt"},
	"*relationship.RouteConsistencyValidator":  {"routes.txt", "trips.txt"},
	"*relationship.UsageValidator":             {"stops.txt", "trips.txt"},
	"*entity.ZoneValidator":                    {"stops.txt"},
	"*business.ShapeGeometryValidator":         {"shapes.txt", "trips.txt", "stops.txt"},
	"*business.TravelSpeedValidator":           {"stops.txt", "trips.txt"},
	"*business.BlockOverlappingValidator":      {"trips.txt"},
	"*relationship.TripShapeDistanceValidator": {"shapes.txt", "trips.txt"},
}

// initializeValidators sets up the validators. Every one of them runs on every
// feed: this list is the validator's behaviour, and the scope audit in
// scripts/scope_audit.py reads it to decide which canonical rules are actually
// reachable, so a check that is not constructed here does not exist.
func (v *internalValidator) initializeValidators() {
	v.validators = []validator.Validator{
		// Core: file structure, field types and required fields.
		core.NewMissingFilesValidator(),
		core.NewEmptyFileValidator(),
		core.NewUnknownFileValidator(),
		core.NewDuplicateHeaderValidator(),
		core.NewMissingColumnValidator(),
		core.NewRequiredFieldValidator(),
		core.NewFieldFormatValidator(),
		core.NewCoordinateValidator(),
		core.NewDuplicateKeyValidator(),
		core.NewInvalidRowValidator(),
		core.NewFieldTypeValidator(),
		core.NewLeadingTrailingWhitespaceValidator(),
		// Registered here rather than in the core package because it lives in
		// the validator package itself.
		validator.NewFileStructureValidator(),

		// Entity: properties and constraints of a single record.
		entity.NewAgencyConsistencyValidator(),
		entity.NewRouteConsistencyValidator(),
		entity.NewServiceValidationValidator(),
		entity.NewStopLocationValidator(),
		entity.NewShapeValidator(),
		entity.NewZoneValidator(),
		entity.NewRouteNameValidator(),
		entity.NewTripPatternValidator(),
		entity.NewDuplicateRouteNameValidator(),
		entity.NewRouteColorContrastValidator(),
		entity.NewStopNameValidator(),
		entity.NewAttributionWithoutRoleValidator(),
		entity.NewRouteTypeValidator(),
		entity.NewNameComparisonValidator(),
		entity.NewMixedCaseNameValidator(),
		entity.NewBikeAllowanceValidator(),

		// Relationship: references between files.
		relationship.NewForeignKeyValidator(),
		relationship.NewStopTimeSequenceValidator(),
		relationship.NewStopTimeSequenceTimeValidator(),
		relationship.NewStopTimeFieldValidator(),
		relationship.NewUsageValidator(),
		relationship.NewTranslationValidator(),
		relationship.NewTripHeadsignValidator(),
		relationship.NewTripShapeDistanceValidator(),
		relationship.NewStopTimeConsistencyValidator(),
		relationship.NewAttributionValidator(),
		relationship.NewRouteConsistencyValidator(),

		// Business: operational consistency across the feed. The last three
		// were previously reserved for the comprehensive mode on the strength
		// of a cost that measurement did not support — running everything is
		// about 1.15x the old default on the largest feed we have.
		business.NewFrequencyValidator(),
		business.NewFeedExpirationDateValidator(),
		business.NewTransferValidator(),
		business.NewTripUsabilityValidator(),
		business.NewTravelSpeedValidator(),
		business.NewBlockOverlappingValidator(),
		business.NewServiceConsistencyValidator(),
		business.NewInSeatTransferValidator(),
		business.NewGeospatialValidator(),
		business.NewShapeGeometryValidator(),
		business.NewDateTripsValidator(),

		// Accessibility: pathways and levels.
		accessibility.NewPathwayValidator(),
		accessibility.NewLevelValidator(),

		// Fare: fare rules and attributes.
		fare.NewFareValidator(),

		// Meta: feed metadata.
		meta.NewFeedInfoValidator(),
	}
}

// For streaming validation, we'll implement a post-validation streaming approach
// where we stream notice groups after each validator completes.

// streamNoticeGroups converts and streams only new notice groups from the container
func (v *internalValidator) streamNoticeGroups() {
	if v.noticeCallback == nil {
		return
	}

	v.streamMutex.Lock()
	defer v.streamMutex.Unlock()

	// Get all notices from the container
	notices := v.noticeContainer.GetNotices()

	// Only process new notices (those beyond our streamed count)
	if len(notices) <= v.streamedCount {
		return // No new notices to stream
	}

	newNotices := notices[v.streamedCount:]
	v.streamedCount = len(notices)

	// Group new notices by code for streaming
	noticeGroups := make(map[string][]notice.Notice)
	for _, n := range newNotices {
		code := n.Code()
		noticeGroups[code] = append(noticeGroups[code], n)
	}

	// Stream each notice group
	for code, groupNotices := range noticeGroups {
		if len(groupNotices) == 0 {
			continue
		}

		// Order most severe first so the sample cap cannot hide the errors
		// in a group that is mostly warnings.
		ordered := make([]notice.Notice, len(groupNotices))
		copy(ordered, groupNotices)
		sort.SliceStable(ordered, func(i, j int) bool {
			return ordered[i].Severity() > ordered[j].Severity()
		})

		// Create sample notices (limit to 5 samples)
		sampleNotices := make([]map[string]interface{}, 0)
		sampleLimit := 5
		counts := NoticeCounts{Total: len(ordered)}
		for i, n := range ordered {
			if i < sampleLimit {
				sampleNotices = append(sampleNotices, report.DescribeNotice(code, n))
			}
			switch n.Severity() {
			case notice.ERROR:
				counts.Errors++
			case notice.WARNING:
				counts.Warnings++
			case notice.INFO:
				counts.Infos++
			}
		}

		// Create notice group for streaming
		enhanced := GetEnhancedNoticeDescription(code)
		noticeGroup := NoticeGroup{
			Code:           code,
			SeverityCounts: counts,
			Description:    enhanced.Description,
			GTFSReference:  enhanced.GTFSReference,
			AffectedFiles:  affectedFiles(code, enhanced),
			AffectedFields: enhanced.AffectedFields,
			ExampleFix:     enhanced.ExampleFix,
			TotalNotices:   len(ordered),
			SampleNotices:  sampleNotices,
		}

		// Stream the notice group
		v.noticeCallback(noticeGroup)
	}
}

// newStreamingNoticeContainer creates a standard notice container for streaming validation
// The streaming happens via the streamNoticeGroups method called periodically
func newStreamingNoticeContainer(callback NoticeCallback) *notice.NoticeContainer {
	return notice.NewNoticeContainer()
}

// newStreamingNoticeContainerWithLimit creates a standard notice container with limit for streaming
func newStreamingNoticeContainerWithLimit(maxPerType int, callback NoticeCallback) *notice.NoticeContainer {
	return notice.NewNoticeContainerWithLimit(maxPerType)
}
