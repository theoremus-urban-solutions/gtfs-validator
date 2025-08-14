package entity

import (
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

func TestRouteColorContrastValidator_Validate(t *testing.T) {
	tests := []struct {
		name                string
		files               map[string]string
		expectedNoticeCodes []string
		description         string
	}{
		{
			name: "good contrast black on white",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_type,route_color,route_text_color\n" +
					"route1,1,3,FFFFFF,000000",
			},
			expectedNoticeCodes: []string{},
			description:         "Black text on white background has excellent contrast",
		},
		{
			name: "good contrast white on black",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_type,route_color,route_text_color\n" +
					"route1,1,3,000000,FFFFFF",
			},
			expectedNoticeCodes: []string{},
			description:         "White text on black background has excellent contrast",
		},
		{
			name: "poor contrast yellow on white",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_type,route_color,route_text_color\n" +
					"route1,1,3,FFFFFF,FFFF00",
			},
			expectedNoticeCodes: []string{"route_color_contrast", "light_text_on_light_background"},
			description:         "Yellow text on white background has poor contrast",
		},
		{
			name: "extremely poor contrast light gray on white",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_type,route_color,route_text_color\n" +
					"route1,1,3,FFFFFF,F0F0F0",
			},
			expectedNoticeCodes: []string{"route_color_contrast", "light_text_on_light_background", "similar_colors"},
			description:         "Light gray on white has extremely poor contrast",
		},
		{
			name: "dark text on dark background",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_type,route_color,route_text_color\n" +
					"route1,1,3,000000,333333",
			},
			expectedNoticeCodes: []string{"route_color_contrast", "dark_text_on_dark_background"},
			description:         "Dark gray on black has poor contrast",
		},
		{
			name: "default colors good contrast",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_type\n" +
					"route1,1,3",
			},
			expectedNoticeCodes: []string{},
			description:         "Default white background with black text should have good contrast",
		},
		{
			name: "red green color combination",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_type,route_color,route_text_color\n" +
					"route1,1,3,FF0000,00FF00",
			},
			expectedNoticeCodes: []string{"route_color_contrast", "red_green_color_combination"},
			description:         "Red background with green text is problematic for colorblind users and has poor contrast",
		},
		{
			name: "green background red text",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_type,route_color,route_text_color\n" +
					"route1,1,3,00FF00,FF0000",
			},
			expectedNoticeCodes: []string{"route_color_contrast", "red_green_color_combination"},
			description:         "Green background with red text is problematic for colorblind users and has poor contrast",
		},
		{
			name: "similar colors too close",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_type,route_color,route_text_color\n" +
					"route1,1,3,FF0000,FE0101",
			},
			expectedNoticeCodes: []string{"route_color_contrast", "similar_colors"},
			description:         "Very similar red colors are hard to distinguish and have poor contrast",
		},
		{
			name: "multiple routes mixed contrast",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_type,route_color,route_text_color\n" +
					"route1,1,3,000000,FFFFFF\n" +
					"route2,2,3,FFFFFF,FFFF00\n" +
					"route3,3,3,0000FF,FFFFFF",
			},
			expectedNoticeCodes: []string{"route_color_contrast", "light_text_on_light_background"},
			description:         "Mixed route colors with one having poor contrast",
		},
		{
			name: "blue on white good contrast",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_type,route_color,route_text_color\n" +
					"route1,1,3,FFFFFF,0000FF",
			},
			expectedNoticeCodes: []string{},
			description:         "Blue text on white background should have adequate contrast",
		},
		{
			name: "only route_color specified uses default text",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_type,route_color\n" +
					"route1,1,3,FFFFFF",
			},
			expectedNoticeCodes: []string{},
			description:         "White background with default black text should be fine",
		},
		{
			name: "only route_text_color specified uses default background",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_type,route_text_color\n" +
					"route1,1,3,000000",
			},
			expectedNoticeCodes: []string{},
			description:         "Black text with default white background should be fine",
		},
		{
			name: "invalid hex colors ignored",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_type,route_color,route_text_color\n" +
					"route1,1,3,GGGGGG,000000\n" +
					"route2,2,3,FF0000,ZZZZZZ",
			},
			expectedNoticeCodes: []string{},
			description:         "Invalid hex colors should be ignored and not cause validation errors",
		},
		{
			name: "empty color fields use defaults",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_type,route_color,route_text_color\n" +
					"route1,1,3,,",
			},
			expectedNoticeCodes: []string{},
			description:         "Empty color fields should use defaults (white background, black text)",
		},
		{
			name: "colors with hash prefix",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_type,route_color,route_text_color\n" +
					"route1,1,3,#FFFFFF,#000000",
			},
			expectedNoticeCodes: []string{},
			description:         "Colors with hash prefix should be parsed correctly",
		},
		{
			name: "mixed case hex colors",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_type,route_color,route_text_color\n" +
					"route1,1,3,ffffff,000000\n" +
					"route2,2,3,000000,ffffff",
			},
			expectedNoticeCodes: []string{},
			description:         "Mixed case hex colors should be handled correctly",
		},
		{
			name: "navy blue on yellow good contrast",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_type,route_color,route_text_color\n" +
					"route1,1,3,FFFF00,000080",
			},
			expectedNoticeCodes: []string{},
			description:         "Navy blue text on yellow background should have good contrast",
		},
		{
			name: "no routes file",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,http://metro.example,America/Los_Angeles",
			},
			expectedNoticeCodes: []string{},
			description:         "Missing routes.txt file should not generate errors",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test feed loader
			feedLoader := CreateTestFeedLoader(t, tt.files)

			// Create notice container and validator
			container := notice.NewNoticeContainer()
			validator := NewRouteColorContrastValidator()
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

func TestRouteColorContrastValidator_ParseColor(t *testing.T) {
	validator := NewRouteColorContrastValidator()

	tests := []struct {
		name        string
		hexStr      string
		isDefault   bool
		expectedRGB [3]int
		shouldPass  bool
	}{
		{
			name:        "valid uppercase hex",
			hexStr:      "FF0000",
			isDefault:   false,
			expectedRGB: [3]int{255, 0, 0},
			shouldPass:  true,
		},
		{
			name:        "valid lowercase hex",
			hexStr:      "00ff00",
			isDefault:   false,
			expectedRGB: [3]int{0, 255, 0},
			shouldPass:  true,
		},
		{
			name:        "valid hex with hash prefix",
			hexStr:      "#0000FF",
			isDefault:   false,
			expectedRGB: [3]int{0, 0, 255},
			shouldPass:  true,
		},
		{
			name:        "default color",
			hexStr:      "FFFFFF",
			isDefault:   true,
			expectedRGB: [3]int{255, 255, 255},
			shouldPass:  true,
		},
		{
			name:       "invalid hex too short",
			hexStr:     "FF00",
			isDefault:  false,
			shouldPass: false,
		},
		{
			name:       "invalid hex too long",
			hexStr:     "FF000000",
			isDefault:  false,
			shouldPass: false,
		},
		{
			name:       "invalid hex characters",
			hexStr:     "GGGGGG",
			isDefault:  false,
			shouldPass: false,
		},
		{
			name:       "empty string",
			hexStr:     "",
			isDefault:  false,
			shouldPass: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.parseColor(tt.hexStr, tt.isDefault)

			if tt.shouldPass {
				if result == nil {
					t.Errorf("Expected valid color, got nil")
					return
				}

				if result.R != tt.expectedRGB[0] || result.G != tt.expectedRGB[1] || result.B != tt.expectedRGB[2] {
					t.Errorf("Expected RGB (%d, %d, %d), got (%d, %d, %d)",
						tt.expectedRGB[0], tt.expectedRGB[1], tt.expectedRGB[2],
						result.R, result.G, result.B)
				}

				if result.IsDefault != tt.isDefault {
					t.Errorf("Expected IsDefault %v, got %v", tt.isDefault, result.IsDefault)
				}
			} else if result != nil {
				t.Errorf("Expected nil for invalid color, got %+v", result)
			}
		})
	}
}

func TestRouteColorContrastValidator_CalculateContrastRatio(t *testing.T) {
	validator := NewRouteColorContrastValidator()

	tests := []struct {
		name          string
		color1        *ColorInfo
		color2        *ColorInfo
		expectedRatio float64
		tolerance     float64
	}{
		{
			name:          "black on white maximum contrast",
			color1:        &ColorInfo{R: 0, G: 0, B: 0},       // Black
			color2:        &ColorInfo{R: 255, G: 255, B: 255}, // White
			expectedRatio: 21.0,
			tolerance:     0.1,
		},
		{
			name:          "white on black maximum contrast",
			color1:        &ColorInfo{R: 255, G: 255, B: 255}, // White
			color2:        &ColorInfo{R: 0, G: 0, B: 0},       // Black
			expectedRatio: 21.0,
			tolerance:     0.1,
		},
		{
			name:          "identical colors minimum contrast",
			color1:        &ColorInfo{R: 128, G: 128, B: 128},
			color2:        &ColorInfo{R: 128, G: 128, B: 128},
			expectedRatio: 1.0,
			tolerance:     0.1,
		},
		{
			name:          "red on blue moderate contrast",
			color1:        &ColorInfo{R: 255, G: 0, B: 0}, // Red
			color2:        &ColorInfo{R: 0, G: 0, B: 255}, // Blue
			expectedRatio: 2.14,                           // Approximate expected value
			tolerance:     0.2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.calculateContrastRatio(tt.color1, tt.color2)

			if result < tt.expectedRatio-tt.tolerance || result > tt.expectedRatio+tt.tolerance {
				t.Errorf("Expected contrast ratio around %.2f (±%.2f), got %.2f",
					tt.expectedRatio, tt.tolerance, result)
			}
		})
	}
}

func TestRouteColorContrastValidator_ColorClassification(t *testing.T) {
	validator := NewRouteColorContrastValidator()

	tests := []struct {
		name    string
		color   *ColorInfo
		isLight bool
		isDark  bool
		isRed   bool
		isGreen bool
	}{
		{
			name:    "white is light",
			color:   &ColorInfo{R: 255, G: 255, B: 255},
			isLight: true,
			isDark:  false,
			isRed:   false,
			isGreen: false,
		},
		{
			name:    "black is dark",
			color:   &ColorInfo{R: 0, G: 0, B: 0},
			isLight: false,
			isDark:  true,
			isRed:   false,
			isGreen: false,
		},
		{
			name:    "bright red is reddish",
			color:   &ColorInfo{R: 255, G: 0, B: 0},
			isLight: false,
			isDark:  false,
			isRed:   true,
			isGreen: false,
		},
		{
			name:    "bright green is greenish",
			color:   &ColorInfo{R: 0, G: 255, B: 0},
			isLight: true,
			isDark:  false,
			isRed:   false,
			isGreen: true,
		},
		{
			name:    "gray is neither light nor dark",
			color:   &ColorInfo{R: 128, G: 128, B: 128},
			isLight: false,
			isDark:  false,
			isRed:   false,
			isGreen: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isLight := validator.isLightColor(tt.color)
			isDark := validator.isDarkColor(tt.color)
			isRed := validator.isRedish(tt.color)
			isGreen := validator.isGreenish(tt.color)

			if isLight != tt.isLight {
				t.Errorf("isLightColor: expected %v, got %v", tt.isLight, isLight)
			}
			if isDark != tt.isDark {
				t.Errorf("isDarkColor: expected %v, got %v", tt.isDark, isDark)
			}
			if isRed != tt.isRed {
				t.Errorf("isRedish: expected %v, got %v", tt.isRed, isRed)
			}
			if isGreen != tt.isGreen {
				t.Errorf("isGreenish: expected %v, got %v", tt.isGreen, isGreen)
			}
		})
	}
}

func TestRouteColorContrastValidator_ColorsAreTooSimilar(t *testing.T) {
	validator := NewRouteColorContrastValidator()

	tests := []struct {
		name            string
		color1          *ColorInfo
		color2          *ColorInfo
		expectedSimilar bool
	}{
		{
			name:            "identical colors are similar",
			color1:          &ColorInfo{R: 255, G: 0, B: 0},
			color2:          &ColorInfo{R: 255, G: 0, B: 0},
			expectedSimilar: true,
		},
		{
			name:            "very close colors are similar",
			color1:          &ColorInfo{R: 255, G: 0, B: 0},
			color2:          &ColorInfo{R: 254, G: 1, B: 1},
			expectedSimilar: true,
		},
		{
			name:            "distant colors are not similar",
			color1:          &ColorInfo{R: 255, G: 0, B: 0},
			color2:          &ColorInfo{R: 0, G: 255, B: 0},
			expectedSimilar: false,
		},
		{
			name:            "black and white are not similar",
			color1:          &ColorInfo{R: 0, G: 0, B: 0},
			color2:          &ColorInfo{R: 255, G: 255, B: 255},
			expectedSimilar: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.colorsAreTooSimilar(tt.color1, tt.color2)

			if result != tt.expectedSimilar {
				t.Errorf("Expected similarity %v, got %v", tt.expectedSimilar, result)
			}
		})
	}
}

func TestRouteColorContrastValidator_New(t *testing.T) {
	validator := NewRouteColorContrastValidator()
	if validator == nil {
		t.Error("NewRouteColorContrastValidator() returned nil")
	}
}
