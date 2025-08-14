package core

import (
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

func TestCurrencyValidator_Validate(t *testing.T) {
	tests := []struct {
		name                string
		files               map[string]string
		expectedNoticeCodes []string
		description         string
	}{
		{
			name: "valid currency codes",
			files: map[string]string{
				FareAttributesFile: "fare_id,price,currency_type,payment_method,transfers\nF1,2.50,USD,0,0\nF2,3.00,EUR,0,0\nF3,300,JPY,0,0",
			},
			expectedNoticeCodes: []string{},
			description:         "Valid ISO 4217 currency codes should not generate notices",
		},
		{
			name: "invalid currency code - too short",
			files: map[string]string{
				FareAttributesFile: "fare_id,price,currency_type,payment_method,transfers\nF1,2.50,US,0,0", // Only 2 characters
			},
			expectedNoticeCodes: []string{"invalid_currency_code"},
			description:         "Currency codes must be exactly 3 characters",
		},
		{
			name: "invalid currency code - too long",
			files: map[string]string{
				FareAttributesFile: "fare_id,price,currency_type,payment_method,transfers\nF1,2.50,USDX,0,0", // 4 characters
			},
			expectedNoticeCodes: []string{"invalid_currency_code"},
			description:         "Currency codes must be exactly 3 characters",
		},
		{
			name: "invalid currency code - unknown",
			files: map[string]string{
				FareAttributesFile: "fare_id,price,currency_type,payment_method,transfers\nF1,2.50,XYZ,0,0", // Not a valid ISO code
			},
			expectedNoticeCodes: []string{"invalid_currency_code"},
			description:         "Unknown ISO 4217 currency codes should generate notices",
		},
		{
			name: "invalid currency code - lowercase",
			files: map[string]string{
				FareAttributesFile: "fare_id,price,currency_type,payment_method,transfers\nF1,2.50,usd,0,0", // Lowercase
			},
			expectedNoticeCodes: []string{"invalid_currency_code"},
			description:         "Currency codes should be uppercase",
		},
		{
			name: "invalid currency code - mixed case",
			files: map[string]string{
				FareAttributesFile: "fare_id,price,currency_type,payment_method,transfers\nF1,2.50,Usd,0,0", // Mixed case
			},
			expectedNoticeCodes: []string{"invalid_currency_code"},
			description:         "Currency codes should be uppercase",
		},
		{
			name: "multiple invalid currency codes",
			files: map[string]string{
				FareAttributesFile: "fare_id,price,currency_type,payment_method,transfers\nF1,2.50,US,0,0\nF2,3.00,XYZ,0,0\nF3,300,jpy,0,0", // All invalid
			},
			expectedNoticeCodes: []string{"invalid_currency_code", "invalid_currency_code", "invalid_currency_code"},
			description:         "Multiple invalid currency codes should generate multiple notices",
		},
		{
			name: "mixed valid and invalid currency codes",
			files: map[string]string{
				FareAttributesFile: "fare_id,price,currency_type,payment_method,transfers\nF1,2.50,USD,0,0\nF2,3.00,XYZ,0,0\nF3,300,EUR,0,0", // Mixed valid/invalid
			},
			expectedNoticeCodes: []string{"invalid_currency_code"},
			description:         "Mix of valid and invalid currency codes",
		},
		{
			name: "empty currency field ignored",
			files: map[string]string{
				FareAttributesFile: "fare_id,price,currency_type,payment_method,transfers\nF1,2.50,,0,0", // Empty currency
			},
			expectedNoticeCodes: []string{},
			description:         "Empty currency fields should not generate validation errors",
		},
		{
			name: "whitespace-only currency field ignored",
			files: map[string]string{
				FareAttributesFile: "fare_id,price,currency_type,payment_method,transfers\nF1,2.50,   ,0,0", // Whitespace currency
			},
			expectedNoticeCodes: []string{},
			description:         "Whitespace-only currency fields should not generate validation errors",
		},
		{
			name: "currency field with whitespace padding",
			files: map[string]string{
				FareAttributesFile: "fare_id,price,currency_type,payment_method,transfers\nF1,2.50, USD ,0,0", // Whitespace around currency
			},
			expectedNoticeCodes: []string{},
			description:         "Whitespace around currency values should be trimmed",
		},
		{
			name: "missing FareAttributesFile file",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,http://metro.example,America/Los_Angeles",
			},
			expectedNoticeCodes: []string{},
			description:         "Missing FareAttributesFile should not cause validation errors",
		},
		{
			name: "FareAttributesFile without currency_type field",
			files: map[string]string{
				FareAttributesFile: "fare_id,price,payment_method,transfers\nF1,2.50,0,0", // Missing currency_type column
			},
			expectedNoticeCodes: []string{},
			description:         "Missing currency_type field should not cause validation errors",
		},
		{
			name: "common world currencies",
			files: map[string]string{
				FareAttributesFile: "fare_id,price,currency_type,payment_method,transfers\nF1,2.50,USD,0,0\nF2,3.00,EUR,0,0\nF3,300,JPY,0,0\nF4,2.00,GBP,0,0\nF5,3.50,CAD,0,0\nF6,4.00,AUD,0,0",
			},
			expectedNoticeCodes: []string{},
			description:         "Common world currencies should all be valid",
		},
		{
			name: "emerging market currencies",
			files: map[string]string{
				FareAttributesFile: "fare_id,price,currency_type,payment_method,transfers\nF1,50,CNY,0,0\nF2,75,INR,0,0\nF3,5000,KRW,0,0\nF4,15,BRL,0,0\nF5,25,RUB,0,0",
			},
			expectedNoticeCodes: []string{},
			description:         "Emerging market currencies should all be valid",
		},
		{
			name: "cryptocurrency codes invalid",
			files: map[string]string{
				FareAttributesFile: "fare_id,price,currency_type,payment_method,transfers\nF1,0.001,BTC,0,0\nF2,0.05,ETH,0,0", // Cryptocurrencies not in ISO 4217
			},
			expectedNoticeCodes: []string{"invalid_currency_code", "invalid_currency_code"},
			description:         "Cryptocurrency codes are not valid ISO 4217 codes",
		},
		{
			name: "obsolete currency codes",
			files: map[string]string{
				FareAttributesFile: "fare_id,price,currency_type,payment_method,transfers\nF1,2.50,DEM,0,0\nF2,3000,ITL,0,0", // Pre-Euro currencies
			},
			expectedNoticeCodes: []string{"invalid_currency_code", "invalid_currency_code"},
			description:         "Obsolete currency codes should be invalid",
		},
		{
			name: "special drawing rights and metals",
			files: map[string]string{
				FareAttributesFile: "fare_id,price,currency_type,payment_method,transfers\nF1,2.50,XDR,0,0\nF2,1800,XAU,0,0", // XDR is valid, XAU might not be
			},
			expectedNoticeCodes: []string{"invalid_currency_code"}, // XAU (gold) might not be in our list
			description:         "Special currencies and precious metals",
		},
		{
			name: "numeric characters in currency",
			files: map[string]string{
				FareAttributesFile: "fare_id,price,currency_type,payment_method,transfers\nF1,2.50,US1,0,0", // Numbers in currency
			},
			expectedNoticeCodes: []string{"invalid_currency_code"},
			description:         "Currency codes with numbers should be invalid",
		},
		{
			name: "special characters in currency",
			files: map[string]string{
				FareAttributesFile: "fare_id,price,currency_type,payment_method,transfers\nF1,2.50,US$,0,0", // Special characters
			},
			expectedNoticeCodes: []string{"invalid_currency_code"},
			description:         "Currency codes with special characters should be invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test components
			loader := CreateTestFeedLoader(t, tt.files)
			container := notice.NewNoticeContainer()
			validator := NewCurrencyValidator()
			config := gtfsvalidator.Config{}

			// Run validation
			validator.Validate(loader, container, config)

			// Get notices
			notices := container.GetNotices()

			// Check notice count
			if len(notices) != len(tt.expectedNoticeCodes) {
				t.Errorf("Expected %d notices, got %d for case: %s", len(tt.expectedNoticeCodes), len(notices), tt.description)
			}

			// Count notice codes
			expectedCodeCounts := make(map[string]int)
			for _, code := range tt.expectedNoticeCodes {
				expectedCodeCounts[code]++
			}

			actualCodeCounts := make(map[string]int)
			for _, notice := range notices {
				actualCodeCounts[notice.Code()]++
			}

			// Verify expected codes
			for expectedCode, expectedCount := range expectedCodeCounts {
				actualCount := actualCodeCounts[expectedCode]
				if actualCount != expectedCount {
					t.Errorf("Expected %d notices with code '%s', got %d", expectedCount, expectedCode, actualCount)
				}
			}

			// Check for unexpected notice codes
			for actualCode := range actualCodeCounts {
				if expectedCodeCounts[actualCode] == 0 {
					t.Errorf("Unexpected notice code: %s", actualCode)
				}
			}
		})
	}
}

func TestCurrencyValidator_ValidateCurrencyCode(t *testing.T) {
	tests := []struct {
		name           string
		currencyCode   string
		expectNotice   bool
		expectedReason string
		description    string
	}{
		{
			name:         "valid USD",
			currencyCode: "USD",
			expectNotice: false,
			description:  "USD should be valid",
		},
		{
			name:         "valid EUR",
			currencyCode: "EUR",
			expectNotice: false,
			description:  "EUR should be valid",
		},
		{
			name:         "valid JPY",
			currencyCode: "JPY",
			expectNotice: false,
			description:  "JPY should be valid",
		},
		{
			name:           "too short",
			currencyCode:   "US",
			expectNotice:   true,
			expectedReason: "Currency code must be exactly 3 characters",
			description:    "2-character codes should be invalid",
		},
		{
			name:           "too long",
			currencyCode:   "USDX",
			expectNotice:   true,
			expectedReason: "Currency code must be exactly 3 characters",
			description:    "4-character codes should be invalid",
		},
		{
			name:           "unknown code",
			currencyCode:   "XYZ",
			expectNotice:   true,
			expectedReason: "Unknown ISO 4217 currency code",
			description:    "Unknown currency codes should be invalid",
		},
		{
			name:           "lowercase",
			currencyCode:   "usd",
			expectNotice:   true,
			expectedReason: "Currency code should be uppercase",
			description:    "Lowercase codes should be invalid",
		},
		{
			name:           "mixed case",
			currencyCode:   "Usd",
			expectNotice:   true,
			expectedReason: "Currency code should be uppercase",
			description:    "Mixed case codes should be invalid",
		},
		{
			name:           "empty string",
			currencyCode:   "",
			expectNotice:   true,
			expectedReason: "Currency code must be exactly 3 characters",
			description:    "Empty string should be invalid",
		},
		{
			name:           "with numbers",
			currencyCode:   "US1",
			expectNotice:   true,
			expectedReason: "Unknown ISO 4217 currency code",
			description:    "Codes with numbers should be invalid",
		},
		{
			name:           "with special characters",
			currencyCode:   "US$",
			expectNotice:   true,
			expectedReason: "Unknown ISO 4217 currency code",
			description:    "Codes with special characters should be invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			container := notice.NewNoticeContainer()
			validator := NewCurrencyValidator()

			validator.validateCurrencyCode(container, FareAttributesFile, "currency_type", tt.currencyCode, 1)

			notices := container.GetNotices()
			hasNotice := len(notices) > 0

			if hasNotice != tt.expectNotice {
				t.Errorf("Expected notice: %v, got notice: %v for %s", tt.expectNotice, hasNotice, tt.description)
			}

			if hasNotice && tt.expectNotice {
				// Verify notice details
				notice := notices[0]
				if notice.Code() != "invalid_currency_code" {
					t.Errorf("Expected notice code 'invalid_currency_code', got '%s'", notice.Code())
				}

				context := notice.Context()
				if filename, ok := context["filename"]; !ok || filename != FareAttributesFile {
					t.Errorf("Expected filename '%s' in context, got '%v'", FareAttributesFile, filename)
				}
				if fieldName, ok := context["fieldName"]; !ok || fieldName != "currency_type" {
					t.Errorf("Expected fieldName 'currency_type' in context, got '%v'", fieldName)
				}
				if currencyCode, ok := context["currencyCode"]; !ok || currencyCode != tt.currencyCode {
					t.Errorf("Expected currencyCode '%s' in context, got '%v'", tt.currencyCode, currencyCode)
				}
				if rowNumber, ok := context["csvRowNumber"]; !ok || rowNumber != 1 {
					t.Errorf("Expected csvRowNumber 1 in context, got '%v'", rowNumber)
				}
				if reason, ok := context["reason"]; ok && tt.expectedReason != "" {
					if reason != tt.expectedReason {
						t.Errorf("Expected reason '%s' in context, got '%v'", tt.expectedReason, reason)
					}
				}
			}
		})
	}
}

func TestCurrencyValidator_ValidCurrencyCodes(t *testing.T) {
	// Test that major world currencies are in validCurrencyCodes
	majorCurrencies := []string{
		"USD", "EUR", "JPY", "GBP", "CAD", "AUD", "CHF", "CNY", "SEK", "NZD",
		"MXN", "SGD", "HKD", "NOK", "KRW", "TRY", "RUB", "INR", "BRL", "ZAR",
	}

	for _, currency := range majorCurrencies {
		if !validCurrencyCodes[currency] {
			t.Errorf("Expected major currency '%s' to be in validCurrencyCodes", currency)
		}
	}

	// Test that some invalid codes are not in the map
	invalidCodes := []string{
		"XXX", "ABC", "123", "US", "USDX", "", "BTC", "ETH",
	}

	for _, code := range invalidCodes {
		if validCurrencyCodes[code] {
			t.Errorf("Invalid code '%s' should not be in validCurrencyCodes", code)
		}
	}
}

func TestCurrencyValidator_ValidateFileCurrency(t *testing.T) {
	tests := []struct {
		name                string
		filename            string
		content             string
		fieldName           string
		expectedNoticeCount int
		description         string
	}{
		{
			name:                "valid currencies",
			filename:            FareAttributesFile,
			content:             "fare_id,price,currency_type,payment_method,transfers\nF1,2.50,USD,0,0\nF2,3.00,EUR,0,0",
			fieldName:           "currency_type",
			expectedNoticeCount: 0,
			description:         "Valid currencies should not generate notices",
		},
		{
			name:                "invalid currencies",
			filename:            FareAttributesFile,
			content:             "fare_id,price,currency_type,payment_method,transfers\nF1,2.50,XYZ,0,0\nF2,3.00,ABC,0,0",
			fieldName:           "currency_type",
			expectedNoticeCount: 2,
			description:         "Invalid currencies should generate notices",
		},
		{
			name:                "missing currency field in data",
			filename:            FareAttributesFile,
			content:             "fare_id,price,payment_method,transfers\nF1,2.50,0,0", // Missing currency_type column
			fieldName:           "currency_type",
			expectedNoticeCount: 0,
			description:         "Missing field should not generate validation errors",
		},
		{
			name:                "empty currency fields",
			filename:            FareAttributesFile,
			content:             "fare_id,price,currency_type,payment_method,transfers\nF1,2.50,,0,0",
			fieldName:           "currency_type",
			expectedNoticeCount: 0,
			description:         "Empty currency fields should be ignored",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files := map[string]string{tt.filename: tt.content}
			loader := CreateTestFeedLoader(t, files)
			container := notice.NewNoticeContainer()
			validator := NewCurrencyValidator()

			validator.validateFileCurrency(loader, container, tt.filename, tt.fieldName)

			notices := container.GetNotices()
			if len(notices) != tt.expectedNoticeCount {
				t.Errorf("Expected %d notices, got %d for %s", tt.expectedNoticeCount, len(notices), tt.description)
			}
		})
	}
}

func TestCurrencyValidator_CaseSensitivity(t *testing.T) {
	// Test that the validator properly handles case sensitivity
	validator := NewCurrencyValidator()

	testCases := []struct{ input, reason string }{
		{"USD", ""}, // Valid - should be no notice
		{"usd", "Currency code should be uppercase"},
		{"Usd", "Currency code should be uppercase"},
		{"UsD", "Currency code should be uppercase"},
	}

	for _, tc := range testCases {
		container := notice.NewNoticeContainer() // Reset container
		validator.validateCurrencyCode(container, FareAttributesFile, "currency_type", tc.input, 1)

		notices := container.GetNotices()

		if tc.reason == "" {
			// Should be valid
			if len(notices) != 0 {
				t.Errorf("Expected no notices for '%s', got %d", tc.input, len(notices))
			}
		} else {
			// Should be invalid
			if len(notices) != 1 {
				t.Errorf("Expected 1 notice for '%s', got %d", tc.input, len(notices))
			} else {
				context := notices[0].Context()
				if reason, ok := context["reason"]; !ok || reason != tc.reason {
					t.Errorf("Expected reason '%s' for '%s', got '%v'", tc.reason, tc.input, reason)
				}
			}
		}
	}
}

func TestCurrencyValidator_New(t *testing.T) {
	validator := NewCurrencyValidator()
	if validator == nil {
		t.Error("NewCurrencyValidator() returned nil")
	}
}
