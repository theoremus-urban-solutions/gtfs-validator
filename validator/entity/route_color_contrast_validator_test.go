package entity

import (
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/testutil"

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
			expectedNoticeCodes: []string{"route_color_contrast"},
			description:         "Yellow text on white background has poor contrast",
		},
		{
			name: "extremely poor contrast light gray on white",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_type,route_color,route_text_color\n" +
					"route1,1,3,FFFFFF,F0F0F0",
			},
			expectedNoticeCodes: []string{"route_color_contrast"},
			description:         "Light gray on white has extremely poor contrast",
		},
		{
			name: "dark text on dark background",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_type,route_color,route_text_color\n" +
					"route1,1,3,000000,333333",
			},
			expectedNoticeCodes: []string{"route_color_contrast"},
			description:         "Dark gray on black has poor contrast",
		},
		{
			name: "default colors good contrast",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_type\n" +
					"route1,1,3",
			},
			expectedNoticeCodes: []string{},
			description:         "Neither color is given, so the route is not judged",
		},
		{
			name: "red green color combination",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_type,route_color,route_text_color\n" +
					"route1,1,3,FF0000,00FF00",
			},
			expectedNoticeCodes: []string{},
			description:         "Red and green differ in brightness by 74, just clear of the threshold; the rule measures legibility, not color blindness",
		},
		{
			name: "green background red text",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_type,route_color,route_text_color\n" +
					"route1,1,3,00FF00,FF0000",
			},
			expectedNoticeCodes: []string{},
			description:         "Same pair reversed, same brightness gap, same verdict",
		},
		{
			name: "similar colors too close",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_type,route_color,route_text_color\n" +
					"route1,1,3,FF0000,FE0101",
			},
			expectedNoticeCodes: []string{"route_color_contrast"},
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
			expectedNoticeCodes: []string{"route_color_contrast"},
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
			name: "only route_color specified",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_type,route_color\n" +
					"route1,1,3,FFFFFF",
			},
			expectedNoticeCodes: []string{},
			description:         "White is only a defect against the text color, which the agency never chose",
		},
		{
			name: "only route_text_color specified",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_type,route_text_color\n" +
					"route1,1,3,000000",
			},
			expectedNoticeCodes: []string{},
			description:         "Half a color pair says nothing about contrast",
		},
		{
			name: "black on unspecified background is not judged",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_type,route_text_color\n" +
					"route1,1,3,010101",
			},
			expectedNoticeCodes: []string{},
			description:         "Near-black text would clash with a black background, but the feed never asked for one",
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
			feedLoader := testutil.CreateTestFeedLoader(t, tt.files)

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

// TestRouteColorContrastValidator_RealFeedColorPairs pins the exact palette of
// a Sofia bus feed, the case that showed the WCAG ratio was the wrong measure:
// it condemned four of these five, of which only the yellow is hard to read.
func TestRouteColorContrastValidator_RealFeedColorPairs(t *testing.T) {
	tests := []struct {
		name           string
		routeColor     string
		routeTextColor string
		expectNotice   bool
		description    string
	}{
		{
			name:           "dark green on white",
			routeColor:     "008b02",
			routeTextColor: "ffffff",
			expectNotice:   false,
			description:    "Luma 82 against 255",
		},
		{
			name:           "blue on white",
			routeColor:     "004dcf",
			routeTextColor: "ffffff",
			expectNotice:   false,
			description:    "Luma 68 against 255; blue barely contributes to brightness",
		},
		{
			name:           "orange on white",
			routeColor:     "db3e00",
			routeTextColor: "ffffff",
			expectNotice:   false,
			description:    "Luma 102 against 255",
		},
		{
			name:           "yellow on white",
			routeColor:     "fccb00",
			routeTextColor: "ffffff",
			expectNotice:   true,
			description:    "Luma 195 against 255, a gap of 60 - white on yellow is the one nobody can read",
		},
		{
			name:           "salmon on white",
			routeColor:     "FF6666",
			routeTextColor: "ffffff",
			expectNotice:   false,
			description:    "Luma 147 against 255, a gap of 108 - light, but still legible",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files := map[string]string{
				"routes.txt": "route_id,route_short_name,route_type,route_color,route_text_color\n" +
					"route1,1,3," + tt.routeColor + "," + tt.routeTextColor,
			}

			container := notice.NewNoticeContainer()
			NewRouteColorContrastValidator().Validate(testutil.CreateTestFeedLoader(t, files), container, gtfsvalidator.Config{})

			gotNotice := len(container.GetNotices()) > 0
			if gotNotice != tt.expectNotice {
				t.Errorf("%s on %s: got notice %v, want %v (%s)",
					tt.routeTextColor, tt.routeColor, gotNotice, tt.expectNotice, tt.description)
			}
		})
	}
}

func TestRouteColorContrastValidator_ParseColor(t *testing.T) {
	validator := NewRouteColorContrastValidator()

	tests := []struct {
		name        string
		hexStr      string
		expectedRGB [3]int
		shouldPass  bool
	}{
		{
			name:        "valid uppercase hex",
			hexStr:      "FF0000",
			expectedRGB: [3]int{255, 0, 0},
			shouldPass:  true,
		},
		{
			name:        "valid lowercase hex",
			hexStr:      "00ff00",
			expectedRGB: [3]int{0, 255, 0},
			shouldPass:  true,
		},
		{
			name:        "valid hex with hash prefix",
			hexStr:      "#0000FF",
			expectedRGB: [3]int{0, 0, 255},
			shouldPass:  true,
		},
		{
			name:       "invalid hex too short",
			hexStr:     "FF00",
			shouldPass: false,
		},
		{
			name:       "invalid hex too long",
			hexStr:     "FF000000",
			shouldPass: false,
		},
		{
			name:       "invalid hex characters",
			hexStr:     "GGGGGG",
			shouldPass: false,
		},
		{
			name:       "empty string",
			hexStr:     "",
			shouldPass: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.parseColor(tt.hexStr)

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
			} else if result != nil {
				t.Errorf("Expected nil for invalid color, got %+v", result)
			}
		})
	}
}

func TestRouteColorContrastValidator_Rec601Luma(t *testing.T) {
	tests := []struct {
		name         string
		color        *ColorInfo
		expectedLuma int
	}{
		{name: "black", color: &ColorInfo{R: 0, G: 0, B: 0}, expectedLuma: 0},
		{name: "white", color: &ColorInfo{R: 255, G: 255, B: 255}, expectedLuma: 255},
		{name: "green carries the most weight", color: &ColorInfo{R: 0, G: 255, B: 0}, expectedLuma: 150},
		{name: "red carries less", color: &ColorInfo{R: 255, G: 0, B: 0}, expectedLuma: 76},
		{name: "blue carries almost none", color: &ColorInfo{R: 0, G: 0, B: 255}, expectedLuma: 28},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := rec601Luma(tt.color); got != tt.expectedLuma {
				t.Errorf("rec601Luma(%+v) = %d, want %d", tt.color, got, tt.expectedLuma)
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
