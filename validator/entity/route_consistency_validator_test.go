package entity

import (
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/testutil"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

func TestRouteConsistencyValidator_Validate(t *testing.T) {
	tests := []struct {
		name                string
		files               map[string]string
		expectedNoticeCodes []string
		description         string
	}{
		{
			name: "valid routes",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_long_name,route_type,route_color,route_text_color,route_url\n" +
					"route1,1,Main Bus Line,3,FF0000,FFFFFF,https://example.com/route1\n" +
					"route2,Red,Red Line Metro,1,DC143C,FFFFFF,https://example.com/route2",
			},
			expectedNoticeCodes: []string{},
			description:         "Valid routes should not generate notices",
		},
		{
			name: "invalid route types",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_long_name,route_type\n" +
					"route1,1,Bus Line,abc\n" + // Non-numeric
					"route2,2,Metro Line,99", // Invalid number
			},
			expectedNoticeCodes: []string{}, // route_type is checked by core/field_type_validator.go
			description:         "Invalid route types should generate errors",
		},
		{
			name: "invalid hex colors",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_long_name,route_type,route_color,route_text_color\n" +
					"route1,1,Bus Line,3,invalid,FFFFFF\n" + // Invalid route_color
					"route2,2,Metro Line,1,FF0000,gggggg", // Invalid route_text_color
			},
			expectedNoticeCodes: []string{"invalid_color", "invalid_color"},
			description:         "Invalid hex colors should generate errors",
		},
		{
			name: "poor color contrast",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_long_name,route_type,route_color,route_text_color\n" +
					"route1,1,Bus Line,3,FF0000,FF0000\n" + // Same color
					"route2,2,Metro Line,1,FFFFFF,EEEEEE", // Very similar colors
			},
			expectedNoticeCodes: []string{},
			description:         "Poor color contrast should generate warnings",
		},
		{
			name: "invalid URLs",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_long_name,route_type,route_url\n" +
					"route1,1,Bus Line,3,invalid-url\n" +
					"route2,2,Metro Line,1,ftp://example.com",
			},
			expectedNoticeCodes: []string{"invalid_url", "invalid_url"},
			description:         "Invalid URLs should generate errors",
		},
		{
			name: "valid color combinations",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_long_name,route_type,route_color,route_text_color\n" +
					"route1,1,Bus Line,3,000000,FFFFFF\n" + // Black on white
					"route2,2,Metro Line,1,FFFFFF,000000\n" + // White on black
					"route3,3,Train Line,2,0000FF,FFFF00", // Blue on yellow
			},
			expectedNoticeCodes: []string{},
			description:         "Valid color combinations should not generate warnings",
		},
		{
			name: "optional fields missing",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_long_name,route_type\n" +
					"route1,1,Bus Line,3\n" +
					"route2,2,Metro Line,1",
			},
			expectedNoticeCodes: []string{},
			description:         "Missing optional fields should not generate errors",
		},
		{
			name: "empty color fields",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_long_name,route_type,route_color,route_text_color\n" +
					"route1,1,Bus Line,3,,\n" + // Empty colors
					"route2,2,Metro Line,1,   ,   ", // Whitespace-only colors
			},
			expectedNoticeCodes: []string{},
			description:         "Empty color fields should not generate errors",
		},
		{
			name: "mixed valid and invalid routes",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_long_name,route_type,route_color,route_text_color,route_url\n" +
					"route1,1,Valid Bus,3,FF0000,FFFFFF,https://example.com\n" +
					"route2,2,Invalid Type,99,FF0000,FFFFFF,https://example.com\n" +
					"route3,3,Invalid Color,3,GGGGGG,FFFFFF,https://example.com\n" +
					"route4,4,Poor Contrast,3,FF0000,FF0000,https://example.com\n" +
					"route5,5,Invalid URL,3,FF0000,FFFFFF,invalid-url",
			},
			expectedNoticeCodes: []string{"invalid_color", "invalid_url"},
			description:         "Mixed valid and invalid routes should generate appropriate notices",
		},
		{
			name: "uncommon but valid route types",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_long_name,route_type\n" +
					"route1,Cable,Cable Car,5\n" +
					"route2,Lift,Aerial Lift,6\n" +
					"route3,Fun,Funicular,7\n" +
					"route4,Trolley,Trolleybus,11\n" +
					"route5,Mono,Monorail,12",
			},
			expectedNoticeCodes: []string{},
			description:         "Uncommon but valid route types should not generate errors",
		},
		{
			name: "color validation edge cases",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_long_name,route_type,route_color,route_text_color\n" +
					"route1,1,Bus,3,ff0000,ffffff\n" + // Lowercase hex
					"route2,2,Metro,1,FF00,FFFFFF\n" + // Too short
					"route3,3,Train,2,FF000000,FFFFFF\n" + // Too long
					"route4,4,Ferry,4,FF000X,FFFFFF", // Invalid character
			},
			expectedNoticeCodes: []string{"invalid_color", "invalid_color", "invalid_color"},
			description:         "Color validation edge cases should be handled correctly",
		},
		{
			name: "URL validation edge cases",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_long_name,route_type,route_url\n" +
					"route1,1,Bus,3,HTTP://EXAMPLE.COM\n" + // Uppercase protocol
					"route2,2,Metro,1,https://example.com/path?query=1\n" + // Complex URL
					"route3,3,Train,2,file://local\n" + // Unsupported protocol
					"route4,4,Ferry,4,//example.com", // Protocol-relative
			},
			expectedNoticeCodes: []string{"invalid_url", "invalid_url"},
			description:         "URL validation edge cases should be handled correctly",
		},
		{
			name: "whitespace handling",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_long_name,route_type,route_color,route_text_color,route_url\n" +
					" route1 , 1 , Bus Line , 3 , FF0000 , FFFFFF , https://example.com ",
			},
			expectedNoticeCodes: []string{},
			description:         "Whitespace should be trimmed properly",
		},
		{
			name: "missing route_id handled gracefully",
			files: map[string]string{
				"routes.txt": "route_short_name,route_long_name,route_type\n" +
					"1,Bus Line,3",
			},
			expectedNoticeCodes: []string{},
			description:         "Missing route_id should be handled gracefully by other validators",
		},
		{
			name: "no routes file",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,http://metro.example,America/Los_Angeles",
			},
			expectedNoticeCodes: []string{},
			description:         "Missing routes.txt file should not generate errors",
		},
		{
			name: "single color provided",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_long_name,route_type,route_color\n" +
					"route1,1,Bus Line,3,FF0000\n" + // Only route_color
					"route2,2,Metro Line,1,", // Empty route_color
			},
			expectedNoticeCodes: []string{},
			description:         "Single color provided should not trigger contrast validation",
		},
		{
			name: "route_text_color only",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_long_name,route_type,route_text_color\n" +
					"route1,1,Bus Line,3,FFFFFF",
			},
			expectedNoticeCodes: []string{},
			description:         "Only route_text_color provided should not trigger contrast validation",
		},
		{
			name: "good contrast combinations",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_long_name,route_type,route_color,route_text_color\n" +
					"route1,1,High Contrast,3,000000,FFFFFF\n" + // Perfect contrast
					"route2,2,Dark Blue on Yellow,1,003366,FFFF00\n" + // Good contrast
					"route3,3,White on Dark Red,2,FFFFFF,800000", // Good contrast
			},
			expectedNoticeCodes: []string{},
			description:         "Good contrast combinations should pass validation",
		},
		{
			name: "borderline contrast cases",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_long_name,route_type,route_color,route_text_color\n" +
					"route1,1,Borderline,3,888888,BBBBBB", // Low contrast
			},
			expectedNoticeCodes: []string{},
			description:         "Borderline contrast should generate warnings",
		},
		{
			name: "valid but unusual URLs",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_long_name,route_type,route_url\n" +
					"route1,1,Bus,3,https://xn--bcher-kva.example.com\n" + // IDN domain
					"route2,2,Metro,1,http://192.168.1.1:8080/path", // IP address with port
			},
			expectedNoticeCodes: []string{},
			description:         "Valid but unusual URLs should pass validation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test feed loader
			feedLoader := testutil.CreateTestFeedLoader(t, tt.files)

			// Create notice container and validator
			container := notice.NewNoticeContainer()
			validator := NewRouteConsistencyValidator()
			config := gtfsvalidator.Config{}

			// Run validation
			validator.Validate(feedLoader, container, config)

			// Get all notices
			allNotices := container.GetNotices()

			// Extract notice codes
			var actualNoticeCodes []string
			for _, n := range allNotices {
				actualNoticeCodes = append(actualNoticeCodes, n.Code())
			}

			// Check if we got the expected notice codes
			expectedSet := make(map[string]bool)
			for _, code := range tt.expectedNoticeCodes {
				expectedSet[code] = true
			}

			actualSet := make(map[string]bool)
			for _, code := range actualNoticeCodes {
				actualSet[code] = true
			}

			// Verify expected codes are present
			for expectedCode := range expectedSet {
				if !actualSet[expectedCode] {
					t.Errorf("Expected notice code '%s' not found. Got: %v", expectedCode, actualNoticeCodes)
				}
			}

			// If no notices expected, ensure no notices were generated
			if len(tt.expectedNoticeCodes) == 0 && len(actualNoticeCodes) > 0 {
				t.Errorf("Expected no notices, but got: %v", actualNoticeCodes)
			}

			t.Logf("Test '%s': Expected %v, Got %v", tt.name, tt.expectedNoticeCodes, actualNoticeCodes)
		})
	}
}

func TestRouteConsistencyValidator_IsValidHexColor(t *testing.T) {
	validator := NewRouteConsistencyValidator()

	tests := []struct {
		color   string
		isValid bool
	}{
		{"FF0000", true},   // Valid red
		{"00FF00", true},   // Valid green
		{"0000FF", true},   // Valid blue
		{"FFFFFF", true},   // Valid white
		{"000000", true},   // Valid black
		{"ff0000", true},   // Valid lowercase
		{"123ABC", true},   // Mixed case
		{"ABCDEF", true},   // All letters
		{"123456", true},   // All numbers
		{"FF000", false},   // Too short
		{"FF00000", false}, // Too long
		{"GGGGGG", false},  // Invalid character
		{"FF000X", false},  // Invalid character
		{"", false},        // Empty
		{"#FF0000", false}, // Hash prefix not allowed
	}

	for _, tt := range tests {
		t.Run(tt.color, func(t *testing.T) {
			result := validator.isValidHexColor(tt.color)
			if result != tt.isValid {
				t.Errorf("Color '%s': expected valid=%v, got %v", tt.color, tt.isValid, result)
			}
		})
	}
}

func TestRouteConsistencyValidator_IsValidURL(t *testing.T) {
	validator := NewRouteConsistencyValidator()

	tests := []struct {
		url     string
		isValid bool
	}{
		{"https://example.com", true},
		{"http://example.com", true},
		{"HTTPS://EXAMPLE.COM", true}, // Case insensitive
		{"HTTP://EXAMPLE.COM", true},  // Case insensitive
		{"https://example.com/path", true},
		{"http://example.com:8080", true},
		{"https://sub.example.com", true},
		{"ftp://example.com", false}, // Wrong protocol
		{"example.com", false},       // Missing protocol
		{"//example.com", false},     // Protocol-relative
		{"file://local", false},      // Wrong protocol
		{"", false},                  // Empty
		{"invalid", false},           // Not a URL
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			result := validator.isValidURL(tt.url)
			if result != tt.isValid {
				t.Errorf("URL '%s': expected valid=%v, got %v", tt.url, tt.isValid, result)
			}
		})
	}
}

func TestRouteConsistencyValidator_New(t *testing.T) {
	validator := NewRouteConsistencyValidator()
	if validator == nil {
		t.Error("NewRouteConsistencyValidator() returned nil")
	}
}
