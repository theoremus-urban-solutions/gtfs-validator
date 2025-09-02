package gtfsvalidator

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
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

// createValidationConfig creates the validation configuration based on mode.
func (v *validatorImpl) createValidationConfig() validationConfig {
	switch v.config.ValidationMode {
	case ValidationModePerformance:
		return performanceValidationConfig()
	case ValidationModeComprehensive:
		return comprehensiveValidationConfig()
	default:
		return defaultValidationConfig()
	}
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
			group.SampleNotices = append(group.SampleNotices, n.SampleNotices...)
		} else {
			enhanced := GetEnhancedNoticeDescription(n.Code)
			noticeGroups[n.Code] = &NoticeGroup{
				Code:           n.Code,
				Severity:       n.Severity,
				Description:    enhanced.Description,
				GTFSReference:  enhanced.GTFSReference,
				AffectedFiles:  enhanced.AffectedFiles,
				AffectedFields: enhanced.AffectedFields,
				ExampleFix:     enhanced.ExampleFix,
				Impact:         enhanced.Impact,
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

type validationConfig struct {
	EnableCore            bool
	EnableEntity          bool
	EnableRelationship    bool
	EnableBusiness        bool
	EnableAccessibility   bool
	EnableFare            bool
	EnableMeta            bool
	EnableGeospatial      bool
	EnableNetworkTopology bool
	EnableDateTrips       bool
	MaxNoticesPerType     int
}

func defaultValidationConfig() validationConfig {
	return validationConfig{
		EnableCore:          true,
		EnableEntity:        true,
		EnableRelationship:  true,
		EnableBusiness:      true,
		EnableAccessibility: true,
		EnableFare:          true,
		EnableMeta:          true,
		MaxNoticesPerType:   100,
	}
}

func performanceValidationConfig() validationConfig {
	return validationConfig{
		EnableCore:         true,
		EnableRelationship: true,
		EnableMeta:         true,
		MaxNoticesPerType:  50,
	}
}

func comprehensiveValidationConfig() validationConfig {
	return validationConfig{
		EnableCore:            true,
		EnableEntity:          true,
		EnableRelationship:    true,
		EnableBusiness:        true,
		EnableAccessibility:   true,
		EnableFare:            true,
		EnableMeta:            true,
		EnableGeospatial:      true,
		EnableNetworkTopology: true,
		EnableDateTrips:       true,
		MaxNoticesPerType:     1000,
	}
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

	totalValidators := len(v.validators)

	// Use parallel workers if configured
	if v.config.ParallelWorkers > 1 && totalValidators > 1 {
		err := v.runValidatorsParallel(ctx, validatorConfig, startTime, totalValidators)
		if err != nil {
			return feedInfo, err
		}
	} else {
		err := v.runValidatorsSequential(ctx, validatorConfig, startTime, totalValidators)
		if err != nil {
			return feedInfo, err
		}
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

// initializeValidators sets up validators based on configuration.
func (v *internalValidator) initializeValidators() {
	v.validators = []validator.Validator{}

	// Core validators
	if v.validationConfig.EnableCore {
		v.validators = append(v.validators,
			core.NewMissingFilesValidator(),
			core.NewEmptyFileValidator(),
			core.NewUnknownFileValidator(),
			core.NewDuplicateHeaderValidator(),
			core.NewMissingColumnValidator(),
			core.NewRequiredFieldValidator(),
			core.NewFieldFormatValidator(),
			core.NewTimeFormatValidator(),
			core.NewDateFormatValidator(),
			core.NewCoordinateValidator(),
			core.NewCurrencyValidator(),
			core.NewDuplicateKeyValidator(),
			core.NewInvalidRowValidator(),
			// core.NewLeadingTrailingWhitespaceValidator(), // PROBLEMATIC: Hangs with large datasets (Sofia)
		)
	}

	// Entity validators
	if v.validationConfig.EnableEntity {
		v.validators = append(v.validators,
			entity.NewPrimaryKeyValidator(),
			entity.NewCalendarValidator(),
			entity.NewAgencyConsistencyValidator(),
			entity.NewRouteConsistencyValidator(),
			entity.NewServiceValidationValidator(),
			entity.NewStopLocationValidator(),
			entity.NewCalendarConsistencyValidator(),
			entity.NewShapeValidator(),
			entity.NewZoneValidator(),
			entity.NewRouteNameValidator(),
			entity.NewTripPatternValidator(),
			entity.NewDuplicateRouteNameValidator(),
			entity.NewRouteColorContrastValidator(),
			entity.NewStopNameValidator(),
			entity.NewBikesAllowanceValidator(),
			entity.NewAttributionWithoutRoleValidator(),
			// entity.NewTripBlockIdValidator(), // PROBLEMATIC: Causes hanging with large datasets
			// entity.NewStopTimeHeadsignValidator(), // PROBLEMATIC: Hangs with large datasets (Sofia)
			entity.NewRouteTypeValidator(),
		)
	}

	// Relationship validators
	if v.validationConfig.EnableRelationship {
		v.validators = append(v.validators,
			relationship.NewForeignKeyValidator(),
			relationship.NewStopTimeSequenceValidator(),
			relationship.NewStopTimeSequenceTimeValidator(),
			relationship.NewShapeDistanceValidator(),
			relationship.NewStopTimeConsistencyValidator(),
			relationship.NewAttributionValidator(),
			relationship.NewRouteConsistencyValidator(),
			relationship.NewShapeIncreasingDistanceValidator(),
		)
	}

	// Business validators
	if v.validationConfig.EnableBusiness {
		v.validators = append(v.validators,
			business.NewFrequencyValidator(),
			business.NewFeedExpirationDateValidator(),
			business.NewTransferValidator(),
			business.NewOverlappingFrequencyValidator(),
			business.NewTripUsabilityValidator(),
			business.NewTransferTimingValidator(),
			business.NewTravelSpeedValidator(),
			business.NewBlockOverlappingValidator(),
			business.NewServiceCalendarValidator(),
			business.NewServiceConsistencyValidator(),
			business.NewScheduleConsistencyValidator(),
		)

		// Expensive business validators (optional)
		if v.validationConfig.EnableGeospatial {
			v.validators = append(v.validators, business.NewGeospatialValidator())
		}
		if v.validationConfig.EnableNetworkTopology {
			v.validators = append(v.validators, business.NewNetworkTopologyValidator())
		}
		if v.validationConfig.EnableDateTrips {
			v.validators = append(v.validators, business.NewDateTripsValidator())
		}

		// Note: Removed expensive validators that cause hangs on large datasets:
		// TravelSpeedValidator, BlockOverlappingValidator, ServiceCalendarValidator
		// These have O(n²) complexity and cause timeouts on large feeds like Sofia
		// All core data validation is still performed by other validators
	}

	// Accessibility validators
	if v.validationConfig.EnableAccessibility {
		v.validators = append(v.validators,
			accessibility.NewPathwayValidator(),
			accessibility.NewLevelValidator(),
		)
	}

	// Fare validators
	if v.validationConfig.EnableFare {
		v.validators = append(v.validators,
			fare.NewFareValidator(),
		)
	}

	// Meta validators
	if v.validationConfig.EnableMeta {
		v.validators = append(v.validators,
			meta.NewFeedInfoValidator(),
		)
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

		// Create sample notices (limit to 5 samples)
		sampleNotices := make([]map[string]interface{}, 0)
		sampleLimit := 5
		for i, n := range groupNotices {
			if i >= sampleLimit {
				break
			}
			sampleNotices = append(sampleNotices, n.Context())
		}

		// Create notice group for streaming
		enhanced := GetEnhancedNoticeDescription(code)
		noticeGroup := NoticeGroup{
			Code:           code,
			Severity:       groupNotices[0].Severity().String(),
			Description:    enhanced.Description,
			GTFSReference:  enhanced.GTFSReference,
			AffectedFiles:  enhanced.AffectedFiles,
			AffectedFields: enhanced.AffectedFields,
			ExampleFix:     enhanced.ExampleFix,
			Impact:         enhanced.Impact,
			TotalNotices:   len(groupNotices),
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
