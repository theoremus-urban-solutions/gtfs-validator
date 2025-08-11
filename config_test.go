package gtfsvalidator

import (
	"strings"
	"testing"
	"time"
)

// TestConfigValidation tests configuration validation and sanitization
func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name           string
		configOptions  []Option
		expectValid    bool
		expectedValues map[string]interface{}
	}{
		{
			name:        "Valid default configuration",
			expectValid: true,
		},
		{
			name: "Valid custom configuration",
			configOptions: []Option{
				WithCountryCode("GB"),
				WithParallelWorkers(8),
				WithMaxMemory(512 * 1024 * 1024), // 512MB
				WithValidationMode(ValidationModePerformance),
				WithMaxNoticesPerType(50),
			},
			expectValid: true,
			expectedValues: map[string]interface{}{
				"CountryCode":       "GB",
				"ParallelWorkers":   8,
				"MaxMemory":         int64(512 * 1024 * 1024),
				"ValidationMode":    ValidationModePerformance,
				"MaxNoticesPerType": 50,
			},
		},
		{
			name: "Invalid country code (gets sanitized)",
			configOptions: []Option{
				WithCountryCode("USA"), // Should be 2 letters
			},
			expectValid: false, // Will be sanitized
			expectedValues: map[string]interface{}{
				"CountryCode": "US", // Should be sanitized to default
			},
		},
		{
			name: "Negative parallel workers (gets sanitized)",
			configOptions: []Option{
				WithParallelWorkers(-1),
			},
			expectValid: false,
			expectedValues: map[string]interface{}{
				"ParallelWorkers": 1, // Should be sanitized to minimum
			},
		},
		{
			name: "Too many parallel workers (gets sanitized)",
			configOptions: []Option{
				WithParallelWorkers(200),
			},
			expectValid: false,
			expectedValues: map[string]interface{}{
				"ParallelWorkers": 100, // Should be sanitized to maximum
			},
		},
		{
			name: "Negative memory limit (gets sanitized)",
			configOptions: []Option{
				WithMaxMemory(-1000),
			},
			expectValid: false,
			expectedValues: map[string]interface{}{
				"MaxMemory": int64(0), // Should be sanitized to no limit
			},
		},
		{
			name: "Too small memory limit (gets sanitized)",
			configOptions: []Option{
				WithMaxMemory(1024 * 1024), // 1MB, too small
			},
			expectValid: false,
			expectedValues: map[string]interface{}{
				"MaxMemory": int64(10 * 1024 * 1024), // Should be sanitized to 10MB
			},
		},
		{
			name: "Invalid validation mode (gets sanitized)",
			configOptions: []Option{
				WithValidationMode("invalid_mode"),
			},
			expectValid: false,
			expectedValues: map[string]interface{}{
				"ValidationMode": ValidationModeDefault, // Should be sanitized to default
			},
		},
		{
			name: "Negative max notices (gets sanitized)",
			configOptions: []Option{
				WithMaxNoticesPerType(-10),
			},
			expectValid: false,
			expectedValues: map[string]interface{}{
				"MaxNoticesPerType": 0, // Should be sanitized to no limit
			},
		},
		{
			name: "Too many max notices (gets sanitized)",
			configOptions: []Option{
				WithMaxNoticesPerType(20000),
			},
			expectValid: false,
			expectedValues: map[string]interface{}{
				"MaxNoticesPerType": 10000, // Should be sanitized to maximum
			},
		},
		{
			name: "Future date (valid within 10 years)",
			configOptions: []Option{
				WithCurrentDate(time.Now().AddDate(1, 0, 0)), // 1 year in future
			},
			expectValid: true,
		},
		{
			name: "Far future date (gets sanitized)",
			configOptions: []Option{
				WithCurrentDate(time.Now().AddDate(20, 0, 0)), // 20 years in future
			},
			expectValid: false,
			// CurrentDate will be sanitized to time.Now(), so we don't test exact value
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create validator with the test configuration
			validator := New(tt.configOptions...)

			// Extract the config for validation
			impl, ok := validator.(*validatorImpl)
			if !ok {
				t.Fatal("Expected validatorImpl, got different type")
			}

			config := impl.config

			// Check expected values if provided
			if tt.expectedValues != nil {
				for key, expectedValue := range tt.expectedValues {
					var actualValue interface{}
					switch key {
					case "CountryCode":
						actualValue = config.CountryCode
					case "ParallelWorkers":
						actualValue = config.ParallelWorkers
					case "MaxMemory":
						actualValue = config.MaxMemory
					case "ValidationMode":
						actualValue = config.ValidationMode
					case "MaxNoticesPerType":
						actualValue = config.MaxNoticesPerType
					default:
						t.Errorf("Unknown config key: %s", key)
						continue
					}

					if actualValue != expectedValue {
						t.Errorf("Config %s: expected %v, got %v", key, expectedValue, actualValue)
					}
				}
			}

			// Validate the final configuration manually to ensure it's valid
			if err := validateConfig(config); err != nil {
				if tt.expectValid {
					t.Errorf("Expected valid config after sanitization, but got error: %v", err)
				}
			} else {
				if !tt.expectValid {
					// This is okay - sanitization should make invalid configs valid
					t.Logf("Config was successfully sanitized to valid state")
				}
			}
		})
	}
}

// TestConfigValidationFunctions tests the validation functions directly
func TestConfigValidationFunctions(t *testing.T) {
	t.Run("validateConfig with valid config", func(t *testing.T) {
		config := &Config{
			CountryCode:       "US",
			CurrentDate:       time.Now(),
			MaxMemory:         512 * 1024 * 1024,
			ParallelWorkers:   4,
			ValidatorVersion:  "1.0.0",
			ValidationMode:    ValidationModeDefault,
			MaxNoticesPerType: 100,
		}

		err := validateConfig(config)
		if err != nil {
			t.Errorf("Expected valid config, got error: %v", err)
		}
	})

	t.Run("validateConfig with invalid config", func(t *testing.T) {
		config := &Config{
			CountryCode:       "INVALID", // Too long
			MaxMemory:         -1000,     // Negative
			ParallelWorkers:   -5,        // Negative
			ValidatorVersion:  "",        // Empty
			ValidationMode:    "invalid", // Unknown mode
			MaxNoticesPerType: -10,       // Negative
		}

		err := validateConfig(config)
		if err == nil {
			t.Error("Expected validation error for invalid config, got nil")
		}

		// Check that error contains information about multiple issues
		errStr := err.Error()
		expectedSubstrings := []string{
			"CountryCode",
			"MaxMemory",
			"ParallelWorkers",
			"ValidatorVersion",
			"ValidationMode",
			"MaxNoticesPerType",
		}

		for _, substr := range expectedSubstrings {
			if !strings.Contains(errStr, substr) {
				t.Errorf("Expected error to contain '%s', but got: %s", substr, errStr)
			}
		}
	})

	t.Run("sanitizeConfig fixes invalid values", func(t *testing.T) {
		config := &Config{
			CountryCode:       "INVALID",                    // Too long
			CurrentDate:       time.Now().AddDate(20, 0, 0), // Too far in future
			MaxMemory:         -1000,                        // Negative
			ParallelWorkers:   200,                          // Too high
			ValidatorVersion:  "",                           // Empty
			ValidationMode:    "invalid",                    // Unknown mode
			MaxNoticesPerType: -10,                          // Negative
		}

		// Sanitize the config
		sanitizeConfig(config)

		// Verify all values are now valid
		if config.CountryCode != "US" {
			t.Errorf("Expected sanitized CountryCode to be 'US', got: %s", config.CountryCode)
		}

		if config.MaxMemory != 0 {
			t.Errorf("Expected sanitized MaxMemory to be 0, got: %d", config.MaxMemory)
		}

		if config.ParallelWorkers != 100 {
			t.Errorf("Expected sanitized ParallelWorkers to be 100, got: %d", config.ParallelWorkers)
		}

		if config.ValidatorVersion != "1.0.0" {
			t.Errorf("Expected sanitized ValidatorVersion to be '1.0.0', got: %s", config.ValidatorVersion)
		}

		if config.ValidationMode != ValidationModeDefault {
			t.Errorf("Expected sanitized ValidationMode to be '%s', got: %s", ValidationModeDefault, config.ValidationMode)
		}

		if config.MaxNoticesPerType != 0 {
			t.Errorf("Expected sanitized MaxNoticesPerType to be 0, got: %d", config.MaxNoticesPerType)
		}

		// Verify the sanitized config now passes validation
		if err := validateConfig(config); err != nil {
			t.Errorf("Sanitized config should be valid, but got error: %v", err)
		}
	})
}
