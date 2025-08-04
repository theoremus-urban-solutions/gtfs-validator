package gtfsvalidator

import (
	"context"
	"io"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name string
		opts []Option
		want *Config
	}{
		{
			name: "default configuration",
			opts: nil,
			want: &Config{
				CountryCode:       "US",
				ParallelWorkers:   4,
				ValidatorVersion:  "1.0.0",
				ValidationMode:    ValidationModeDefault,
				MaxNoticesPerType: 100,
			},
		},
		{
			name: "custom country code",
			opts: []Option{WithCountryCode("UK")},
			want: &Config{
				CountryCode:       "UK",
				ParallelWorkers:   4,
				ValidatorVersion:  "1.0.0",
				ValidationMode:    ValidationModeDefault,
				MaxNoticesPerType: 100,
			},
		},
		{
			name: "performance mode",
			opts: []Option{WithValidationMode(ValidationModePerformance)},
			want: &Config{
				CountryCode:       "US",
				ParallelWorkers:   4,
				ValidatorVersion:  "1.0.0",
				ValidationMode:    ValidationModePerformance,
				MaxNoticesPerType: 100,
			},
		},
		{
			name: "custom workers",
			opts: []Option{WithParallelWorkers(8)},
			want: &Config{
				CountryCode:       "US",
				ParallelWorkers:   8,
				ValidatorVersion:  "1.0.0",
				ValidationMode:    ValidationModeDefault,
				MaxNoticesPerType: 100,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := New(tt.opts...)
			impl, ok := validator.(*validatorImpl)
			if !ok {
				t.Fatal("New() should return *validatorImpl")
			}

			config := impl.config
			if config.CountryCode != tt.want.CountryCode {
				t.Errorf("CountryCode = %q, want %q", config.CountryCode, tt.want.CountryCode)
			}
			if config.ParallelWorkers != tt.want.ParallelWorkers {
				t.Errorf("ParallelWorkers = %d, want %d", config.ParallelWorkers, tt.want.ParallelWorkers)
			}
			if config.ValidatorVersion != tt.want.ValidatorVersion {
				t.Errorf("ValidatorVersion = %q, want %q", config.ValidatorVersion, tt.want.ValidatorVersion)
			}
			if config.ValidationMode != tt.want.ValidationMode {
				t.Errorf("ValidationMode = %q, want %q", config.ValidationMode, tt.want.ValidationMode)
			}
			if config.MaxNoticesPerType != tt.want.MaxNoticesPerType {
				t.Errorf("MaxNoticesPerType = %d, want %d", config.MaxNoticesPerType, tt.want.MaxNoticesPerType)
			}
		})
	}
}

func TestWithOptions(t *testing.T) {
	t.Run("WithCountryCode", func(t *testing.T) {
		validator := New(WithCountryCode("FR"))
		impl := validator.(*validatorImpl)
		if impl.config.CountryCode != "FR" {
			t.Errorf("Expected country code FR, got %s", impl.config.CountryCode)
		}
	})

	t.Run("WithCurrentDate", func(t *testing.T) {
		testDate := time.Date(2023, 12, 25, 0, 0, 0, 0, time.UTC)
		validator := New(WithCurrentDate(testDate))
		impl := validator.(*validatorImpl)
		if !impl.config.CurrentDate.Equal(testDate) {
			t.Errorf("Expected date %v, got %v", testDate, impl.config.CurrentDate)
		}
	})

	t.Run("WithMaxMemory", func(t *testing.T) {
		validator := New(WithMaxMemory(1024 * 1024 * 512)) // 512MB
		impl := validator.(*validatorImpl)
		if impl.config.MaxMemory != 1024*1024*512 {
			t.Errorf("Expected max memory 536870912, got %d", impl.config.MaxMemory)
		}
	})

	t.Run("WithParallelWorkers", func(t *testing.T) {
		validator := New(WithParallelWorkers(16))
		impl := validator.(*validatorImpl)
		if impl.config.ParallelWorkers != 16 {
			t.Errorf("Expected 16 workers, got %d", impl.config.ParallelWorkers)
		}
	})

	t.Run("WithMaxNoticesPerType", func(t *testing.T) {
		validator := New(WithMaxNoticesPerType(50))
		impl := validator.(*validatorImpl)
		if impl.config.MaxNoticesPerType != 50 {
			t.Errorf("Expected 50 max notices, got %d", impl.config.MaxNoticesPerType)
		}
	})

	t.Run("WithProgressCallback", func(t *testing.T) {
		called := false
		callback := func(info ProgressInfo) {
			called = true
		}
		validator := New(WithProgressCallback(callback))
		impl := validator.(*validatorImpl)
		
		if impl.config.ProgressCallback == nil {
			t.Error("Expected progress callback to be set")
		}
		
		// Test callback works
		impl.config.ProgressCallback(ProgressInfo{})
		if !called {
			t.Error("Expected progress callback to be called")
		}
	})
}

func TestValidationModes(t *testing.T) {
	modes := []ValidationMode{
		ValidationModePerformance,
		ValidationModeDefault,
		ValidationModeComprehensive,
	}

	for _, mode := range modes {
		t.Run(string(mode), func(t *testing.T) {
			validator := New(WithValidationMode(mode))
			impl := validator.(*validatorImpl)
			if impl.config.ValidationMode != mode {
				t.Errorf("Expected mode %s, got %s", mode, impl.config.ValidationMode)
			}
		})
	}
}

func TestValidationReport(t *testing.T) {
	t.Run("HasErrors", func(t *testing.T) {
		report := &ValidationReport{
			Summary: Summary{
				Counts: NoticeCounts{Errors: 5, Warnings: 3, Infos: 1},
			},
		}
		if !report.HasErrors() {
			t.Error("Expected HasErrors() to return true")
		}

		report.Summary.Counts.Errors = 0
		if report.HasErrors() {
			t.Error("Expected HasErrors() to return false")
		}
	})

	t.Run("HasWarnings", func(t *testing.T) {
		report := &ValidationReport{
			Summary: Summary{
				Counts: NoticeCounts{Errors: 0, Warnings: 3, Infos: 1},
			},
		}
		if !report.HasWarnings() {
			t.Error("Expected HasWarnings() to return true")
		}

		report.Summary.Counts.Warnings = 0
		if report.HasWarnings() {
			t.Error("Expected HasWarnings() to return false")
		}
	})

	t.Run("ErrorCount", func(t *testing.T) {
		report := &ValidationReport{
			Summary: Summary{
				Counts: NoticeCounts{Errors: 42, Warnings: 3, Infos: 1},
			},
		}
		if count := report.ErrorCount(); count != 42 {
			t.Errorf("Expected ErrorCount() = 42, got %d", count)
		}
	})

	t.Run("WarningCount", func(t *testing.T) {
		report := &ValidationReport{
			Summary: Summary{
				Counts: NoticeCounts{Errors: 5, Warnings: 33, Infos: 1},
			},
		}
		if count := report.WarningCount(); count != 33 {
			t.Errorf("Expected WarningCount() = 33, got %d", count)
		}
	})
}

// Test that the public API methods exist and have correct signatures
func TestValidatorInterface(t *testing.T) {
	validator := New()

	// These should compile without error
	var _ func(string) (*ValidationReport, error) = validator.ValidateFile
	var _ func(context.Context, string) (*ValidationReport, error) = validator.ValidateFileWithContext
	var _ func(io.Reader) (*ValidationReport, error) = validator.ValidateReader
	var _ func(context.Context, io.Reader) (*ValidationReport, error) = validator.ValidateReaderWithContext
	var _ func(string, NoticeCallback) (*ValidationReport, error) = validator.ValidateFileStream
	var _ func(context.Context, string, NoticeCallback) (*ValidationReport, error) = validator.ValidateFileStreamWithContext
}

func TestValidateFileErrors(t *testing.T) {
	validator := New()

	t.Run("non-existent file", func(t *testing.T) {
		_, err := validator.ValidateFile("/non/existent/path.zip")
		if err == nil {
			t.Error("Expected error for non-existent file")
		}
	})

	t.Run("empty path", func(t *testing.T) {
		_, err := validator.ValidateFile("")
		if err == nil {
			t.Error("Expected error for empty path")
		}
	})
}

func TestContextCancellation(t *testing.T) {
	validator := New()
	
	// Test context cancellation
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately
	
	_, err := validator.ValidateFileWithContext(ctx, "/some/path.zip")
	if err != context.Canceled {
		t.Errorf("Expected context canceled error, got %v", err)
	}
}

func TestProgressInfo(t *testing.T) {
	info := ProgressInfo{
		CurrentValidator:    "TestValidator",
		TotalValidators:     10,
		CompletedValidators: 5,
		PercentComplete:     50.0,
		ElapsedTime:         time.Second * 30,
	}

	if info.CurrentValidator != "TestValidator" {
		t.Errorf("Expected CurrentValidator = TestValidator, got %s", info.CurrentValidator)
	}
	if info.TotalValidators != 10 {
		t.Errorf("Expected TotalValidators = 10, got %d", info.TotalValidators)
	}
	if info.CompletedValidators != 5 {
		t.Errorf("Expected CompletedValidators = 5, got %d", info.CompletedValidators)
	}
	if info.PercentComplete != 50.0 {
		t.Errorf("Expected PercentComplete = 50.0, got %f", info.PercentComplete)
	}
	if info.ElapsedTime != time.Second*30 {
		t.Errorf("Expected ElapsedTime = 30s, got %v", info.ElapsedTime)
	}
}

func TestNoticeCallback(t *testing.T) {
	var receivedNotice NoticeGroup
	callback := func(notice NoticeGroup) {
		receivedNotice = notice
	}

	testNotice := NoticeGroup{
		Code:         "test_code",
		Severity:     "ERROR",
		TotalNotices: 1,
	}

	callback(testNotice)

	if receivedNotice.Code != "test_code" {
		t.Errorf("Expected notice code test_code, got %s", receivedNotice.Code)
	}
	if receivedNotice.Severity != "ERROR" {
		t.Errorf("Expected notice severity ERROR, got %s", receivedNotice.Severity)
	}
}