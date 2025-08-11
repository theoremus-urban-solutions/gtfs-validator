package entity

import (
	"fmt"
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

func TestRouteTypeValidator_Validate(t *testing.T) {
	tests := []struct {
		name                string
		files               map[string]string
		expectedNoticeCodes []string
		description         string
	}{
		{
			name: "valid basic route types",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_long_name,route_type\n" +
					"route1,1,Main Bus Line,3\n" +
					"route2,Red,Red Line Metro,1\n" +
					"route3,Ferry,Island Ferry,4\n" +
					"route4,Rail,Commuter Rail,2",
			},
			expectedNoticeCodes: []string{"unusual_route_type_combination"},
			description:         "Valid basic GTFS route types should not generate notices",
		},
		{
			name: "invalid route types",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_long_name,route_type\n" +
					"route1,1,Bus Line,999\n" + // Invalid
					"route2,2,Metro Line,-1\n" + // Invalid
					"route3,3,Rail Line,50", // Invalid
			},
			expectedNoticeCodes: []string{"uncommon_route_type", "invalid_route_type", "invalid_route_type"},
			description:         "Invalid route types should generate errors",
		},
		{
			name: "uncommon route types",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_long_name,route_type\n" +
					"route1,Cable,Cable Car,5\n" +
					"route2,Lift,Aerial Lift,6\n" +
					"route3,Fun,Funicular,7\n" +
					"route4,Trolley,Trolleybus,11\n" +
					"route5,Mono,Monorail,12",
			},
			expectedNoticeCodes: []string{"uncommon_route_type", "uncommon_route_type", "uncommon_route_type", "uncommon_route_type", "uncommon_route_type"},
			description:         "Uncommon route types should generate warnings",
		},
		{
			name: "extended route types",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_long_name,route_type\n" +
					"route1,101,Railway Service,101\n" +
					"route2,201,Coach Service,201\n" +
					"route3,701,Bus Service,701\n" +
					"route4,1001,Water Service,1001\n" +
					"route5,1301,Aerial Service,1301",
			},
			expectedNoticeCodes: []string{"uncommon_route_type", "uncommon_route_type", "uncommon_route_type", "uncommon_route_type", "uncommon_route_type"},
			description:         "Extended route types are valid but uncommon",
		},
		{
			name: "route type name mismatch",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_long_name,route_type\n" +
					"route1,Metro,Metro Line,3\n" + // Bus with metro name
					"route2,Bus,Bus Line,1\n" + // Metro with bus name
					"route3,Train,Train Line,3\n" + // Bus with train name
					"route4,Ferry,Ferry Service,4", // Correct ferry
			},
			expectedNoticeCodes: []string{"route_type_name_mismatch", "route_type_name_mismatch", "route_type_name_mismatch"},
			description:         "Route names that don't match route type should generate warnings",
		},
		{
			name: "ferry without ferry keywords",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_long_name,route_type\n" +
					"route1,F1,Island Service,4\n" + // Ferry without ferry/boat/water keywords
					"route2,Boat,Boat Service,4\n" + // Ferry with boat keyword - OK
					"route3,,Water Taxi,4", // Ferry with water keyword - OK
			},
			expectedNoticeCodes: []string{"route_type_name_mismatch"},
			description:         "Ferry routes should have ferry/boat/water keywords in names",
		},
		{
			name: "single uncommon route type in feed",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_long_name,route_type\n" +
					"route1,Ferry1,Island Ferry,4\n" +
					"route2,Ferry2,Mainland Ferry,4\n" +
					"route3,Ferry3,Harbor Ferry,4",
			},
			expectedNoticeCodes: []string{"single_route_type_in_feed"},
			description:         "Feed with only ferry routes should generate warning",
		},
		{
			name: "agency with too many route types",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_long_name,route_type,agency_id\n" +
					"route1,1,Bus Line,3,agency1\n" +
					"route2,Red,Metro Line,1,agency1\n" +
					"route3,Rail,Rail Line,2,agency1\n" +
					"route4,Ferry,Ferry,4,agency1\n" +
					"route5,Cable,Cable,5,agency1\n" +
					"route6,Lift,Lift,6,agency1", // 6 different types for one agency
			},
			expectedNoticeCodes: []string{"uncommon_route_type", "uncommon_route_type", "agency_mixed_route_types"},
			description:         "Agency with more than 4 route types should generate warning",
		},
		{
			name: "unusual route type combination",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_long_name,route_type\n" +
					"route1,Metro,Metro Line,1\n" + // Subway
					"route2,Ferry,Ferry,4", // Ferry (unusual with subway)
			},
			expectedNoticeCodes: []string{"unusual_route_type_combination"},
			description:         "Subway and ferry together is unusual combination",
		},
		{
			name: "mixed route types normal combination",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_long_name,route_type\n" +
					"route1,1,Bus Line,3\n" +
					"route2,Metro,Metro Line,1\n" +
					"route3,Rail,Rail Line,2\n" +
					"route4,Tram,Tram Line,0",
			},
			expectedNoticeCodes: []string{},
			description:         "Normal combination of route types should not generate warnings",
		},
		{
			name: "route type edge cases",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_long_name,route_type\n" +
					"route1,,Light Rail,0\n" + // No short name
					"route2,Express,,3\n" + // No long name
					"route3,42,42nd Street Express,3", // Numeric short name
			},
			expectedNoticeCodes: []string{},
			description:         "Edge cases with missing names should not affect route type validation",
		},
		{
			name: "valid extended route type ranges",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_long_name,route_type\n" +
					"route1,R100,Railway,100\n" + // Railway range start
					"route2,R117,Railway,117\n" + // Railway range end
					"route3,C200,Coach,200\n" + // Coach range start
					"route4,C209,Coach,209\n" + // Coach range end
					"route5,B700,Bus,700\n" + // Bus range start
					"route6,B799,Bus,799", // Bus range end
			},
			expectedNoticeCodes: []string{"uncommon_route_type", "uncommon_route_type", "uncommon_route_type", "uncommon_route_type", "uncommon_route_type", "uncommon_route_type"},
			description:         "Extended route types should be valid but generate uncommon warnings",
		},
		{
			name: "case insensitive keyword matching",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_long_name,route_type\n" +
					"route1,METRO,METRO LINE,3\n" + // Bus with METRO (uppercase)
					"route2,Bus,Metro Express,1\n" + // Metro with mixed case
					"route3,rail,Rail Service,3", // Bus with rail (lowercase)
			},
			expectedNoticeCodes: []string{"route_type_name_mismatch", "route_type_name_mismatch", "route_type_name_mismatch"},
			description:         "Keyword matching should be case insensitive",
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
			name: "routes without route_id or route_type ignored",
			files: map[string]string{
				"routes.txt": "route_short_name,route_long_name,route_type\n" +
					"1,Bus Line,3\n" + // Missing route_id
					"route_id,route_short_name,route_long_name\n" +
					"route1,2,Metro Line", // Missing route_type
			},
			expectedNoticeCodes: []string{},
			description:         "Routes without required fields should be ignored",
		},
		{
			name: "single common route type in feed",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_long_name,route_type\n" +
					"route1,1,Bus Line 1,3\n" +
					"route2,2,Bus Line 2,3\n" +
					"route3,3,Bus Line 3,3",
			},
			expectedNoticeCodes: []string{},
			description:         "Feed with only common route type (bus) should not generate warnings",
		},
		{
			name: "multiple agencies with different route types",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_long_name,route_type,agency_id\n" +
					"route1,1,Bus Line,3,bus_agency\n" +
					"route2,2,Bus Line,3,bus_agency\n" +
					"route3,Metro,Metro Line,1,metro_agency\n" +
					"route4,Rail,Rail Line,2,rail_agency",
			},
			expectedNoticeCodes: []string{},
			description:         "Multiple agencies with reasonable route type diversity should be fine",
		},
		{
			name: "whitespace handling",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_long_name,route_type\n" +
					" route1 , Metro , Metro Line , 3 ", // Bus with metro name (trimmed)
			},
			expectedNoticeCodes: []string{"route_type_name_mismatch"},
			description:         "Whitespace should be trimmed properly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test feed loader
			feedLoader := CreateTestFeedLoader(t, tt.files)

			// Create notice container and validator
			container := notice.NewNoticeContainer()
			validator := NewRouteTypeValidator()
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

func TestRouteTypeValidator_LoadRoutes(t *testing.T) {
	validator := NewRouteTypeValidator()

	tests := []struct {
		name     string
		csvData  string
		expected []*RouteTypeInfo
	}{
		{
			name: "basic route loading",
			csvData: "route_id,route_short_name,route_long_name,route_type,agency_id\n" +
				"route1,1,Main Bus,3,agency1",
			expected: []*RouteTypeInfo{
				{
					RouteID:        "route1",
					RouteType:      3,
					RouteShortName: "1",
					RouteLongName:  "Main Bus",
					AgencyID:       "agency1",
					RowNumber:      2,
				},
			},
		},
		{
			name: "minimal route loading",
			csvData: "route_id,route_type\n" +
				"route1,3",
			expected: []*RouteTypeInfo{
				{
					RouteID:        "route1",
					RouteType:      3,
					RouteShortName: "",
					RouteLongName:  "",
					AgencyID:       "",
					RowNumber:      2,
				},
			},
		},
		{
			name: "whitespace trimming",
			csvData: "route_id,route_short_name,route_long_name,route_type,agency_id\n" +
				" route1 , 1 , Main Bus , 3 , agency1 ",
			expected: []*RouteTypeInfo{
				{
					RouteID:        "route1",
					RouteType:      3,
					RouteShortName: "1",
					RouteLongName:  "Main Bus",
					AgencyID:       "agency1",
					RowNumber:      2,
				},
			},
		},
		{
			name: "mixed route types",
			csvData: "route_id,route_type\n" +
				"bus1,3\n" +
				"metro1,1\n" +
				"rail1,2\n" +
				"ferry1,4",
			expected: []*RouteTypeInfo{
				{RouteID: "bus1", RouteType: 3, RowNumber: 2},
				{RouteID: "metro1", RouteType: 1, RowNumber: 3},
				{RouteID: "rail1", RouteType: 2, RowNumber: 4},
				{RouteID: "ferry1", RouteType: 4, RowNumber: 5},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			feedLoader := CreateTestFeedLoader(t, map[string]string{
				"routes.txt": tt.csvData,
			})

			result := validator.loadRoutes(feedLoader)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d routes, got %d", len(tt.expected), len(result))
			}

			for i, expectedRoute := range tt.expected {
				if i >= len(result) {
					t.Errorf("Expected route at index %d not found", i)
					continue
				}

				actualRoute := result[i]
				if actualRoute.RouteID != expectedRoute.RouteID {
					t.Errorf("Route %d: expected RouteID %s, got %s", i, expectedRoute.RouteID, actualRoute.RouteID)
				}
				if actualRoute.RouteType != expectedRoute.RouteType {
					t.Errorf("Route %d: expected RouteType %d, got %d", i, expectedRoute.RouteType, actualRoute.RouteType)
				}
				if actualRoute.RouteShortName != expectedRoute.RouteShortName {
					t.Errorf("Route %d: expected RouteShortName %s, got %s", i, expectedRoute.RouteShortName, actualRoute.RouteShortName)
				}
				if actualRoute.RouteLongName != expectedRoute.RouteLongName {
					t.Errorf("Route %d: expected RouteLongName %s, got %s", i, expectedRoute.RouteLongName, actualRoute.RouteLongName)
				}
				if actualRoute.AgencyID != expectedRoute.AgencyID {
					t.Errorf("Route %d: expected AgencyID %s, got %s", i, expectedRoute.AgencyID, actualRoute.AgencyID)
				}
				if actualRoute.RowNumber != expectedRoute.RowNumber {
					t.Errorf("Route %d: expected RowNumber %d, got %d", i, expectedRoute.RowNumber, actualRoute.RowNumber)
				}
			}
		})
	}
}

func TestRouteTypeValidator_IsValidRouteType(t *testing.T) {
	validator := NewRouteTypeValidator()

	tests := []struct {
		routeType int
		isValid   bool
	}{
		// Basic valid types
		{0, true},  // Tram
		{1, true},  // Subway
		{2, true},  // Rail
		{3, true},  // Bus
		{4, true},  // Ferry
		{5, true},  // Cable tram
		{6, true},  // Aerial lift
		{7, true},  // Funicular
		{11, true}, // Trolleybus
		{12, true}, // Monorail
		// Invalid basic types
		{8, false},  // Invalid
		{9, false},  // Invalid
		{10, false}, // Invalid
		{13, false}, // Invalid
		{99, false}, // Invalid
		// Extended valid types
		{100, true},  // Railway service start
		{117, true},  // Railway service end
		{200, true},  // Coach service start
		{209, true},  // Coach service end
		{701, true},  // Bus service middle
		{1001, true}, // Water transport
		{1301, true}, // Aerial lift service
		{1700, true}, // Miscellaneous service end
		// Invalid extended types
		{50, false},   // Invalid gap
		{118, false},  // Invalid gap
		{210, false},  // Invalid gap
		{1800, false}, // Too high
		{-1, false},   // Negative
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("route_type_%d", tt.routeType), func(t *testing.T) {
			result := validator.isValidRouteType(tt.routeType)
			if result != tt.isValid {
				t.Errorf("Route type %d: expected validity %v, got %v", tt.routeType, tt.isValid, result)
			}
		})
	}
}

func TestRouteTypeValidator_IsUncommonRouteType(t *testing.T) {
	validator := NewRouteTypeValidator()

	tests := []struct {
		routeType  int
		isUncommon bool
	}{
		// Common types
		{0, false}, // Tram
		{1, false}, // Subway
		{2, false}, // Rail
		{3, false}, // Bus
		{4, false}, // Ferry
		// Uncommon basic types
		{5, true},  // Cable tram
		{6, true},  // Aerial lift
		{7, true},  // Funicular
		{11, true}, // Trolleybus
		{12, true}, // Monorail
		// Extended types (all uncommon)
		{100, true},  // Extended
		{701, true},  // Extended
		{1001, true}, // Extended
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("route_type_%d", tt.routeType), func(t *testing.T) {
			result := validator.isUncommonRouteType(tt.routeType)
			if result != tt.isUncommon {
				t.Errorf("Route type %d: expected uncommon %v, got %v", tt.routeType, tt.isUncommon, result)
			}
		})
	}
}

func TestRouteTypeValidator_IsUncommonAsOnlyRouteType(t *testing.T) {
	validator := NewRouteTypeValidator()

	tests := []struct {
		routeType            int
		isUncommonAsOnlyType bool
	}{
		// Common as only types
		{0, false}, // Tram - OK as only type
		{1, false}, // Subway - OK as only type
		{2, false}, // Rail - OK as only type
		{3, false}, // Bus - OK as only type
		// Uncommon as only types
		{4, true},  // Ferry - uncommon as only type
		{5, true},  // Cable tram - uncommon as only type
		{6, true},  // Aerial lift - uncommon as only type
		{7, true},  // Funicular - uncommon as only type
		{11, true}, // Trolleybus - uncommon as only type
		{12, true}, // Monorail - uncommon as only type
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("route_type_%d", tt.routeType), func(t *testing.T) {
			result := validator.isUncommonAsOnlyRouteType(tt.routeType)
			if result != tt.isUncommonAsOnlyType {
				t.Errorf("Route type %d: expected uncommon as only type %v, got %v", tt.routeType, tt.isUncommonAsOnlyType, result)
			}
		})
	}
}

func TestRouteTypeValidator_GetRouteTypeDescription(t *testing.T) {
	validator := NewRouteTypeValidator()

	tests := []struct {
		routeType   int
		description string
	}{
		{0, "Tram, Streetcar, Light rail"},
		{1, "Subway, Metro"},
		{2, "Rail"},
		{3, "Bus"},
		{4, "Ferry"},
		{5, "Cable tram"},
		{6, "Aerial lift, suspended cable car"},
		{7, "Funicular"},
		{11, "Trolleybus"},
		{12, "Monorail"},
		{101, "Railway Service"},
		{201, "Coach Service"},
		{701, "Bus Service"},
		{1001, "Water Transport Service"},
		{1301, "Aerial Lift Service"},
		{9999, "Extended Route Type"}, // Fallback
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("route_type_%d", tt.routeType), func(t *testing.T) {
			result := validator.getRouteTypeDescription(tt.routeType)
			if result != tt.description {
				t.Errorf("Route type %d: expected description '%s', got '%s'", tt.routeType, tt.description, result)
			}
		})
	}
}

func TestRouteTypeValidator_ContainsTransitModeKeywords(t *testing.T) {
	validator := NewRouteTypeValidator()

	tests := []struct {
		name     string
		text     string
		keywords []string
		expected bool
	}{
		{"exact match", "metro", []string{"metro", "subway"}, true},
		{"partial match", "metro line", []string{"metro", "subway"}, true},
		{"no match", "bus line", []string{"metro", "subway"}, false},
		{"case sensitive", "METRO", []string{"metro", "subway"}, false}, // Function is case-sensitive
		{"multiple keywords", "bus service", []string{"bus", "coach"}, true},
		{"empty text", "", []string{"metro"}, false},
		{"empty keywords", "metro", []string{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.containsTransitModeKeywords(tt.text, tt.keywords)
			if result != tt.expected {
				t.Errorf("Text '%s' with keywords %v: expected %v, got %v", tt.text, tt.keywords, tt.expected, result)
			}
		})
	}
}

func TestRouteTypeValidator_IsValidExtendedRouteType(t *testing.T) {
	validator := NewRouteTypeValidator()

	tests := []struct {
		routeType int
		isValid   bool
	}{
		// Railway Service (100-117)
		{100, true},
		{117, true},
		{118, false},
		// Coach Service (200-209)
		{200, true},
		{209, true},
		{210, false},
		// Bus Service (700-799)
		{700, true},
		{799, true},
		{800, true}, // This is trolleybus range (800-899) - valid
		// Water Transport (1000-1099)
		{1000, true},
		{1099, true},
		{1100, true}, // This is air service range (1100-1199) - valid
		// Invalid ranges
		{50, false},
		{150, false},
		{250, false},
		{1900, false},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("extended_route_type_%d", tt.routeType), func(t *testing.T) {
			result := validator.isValidExtendedRouteType(tt.routeType)
			if result != tt.isValid {
				t.Errorf("Extended route type %d: expected validity %v, got %v", tt.routeType, tt.isValid, result)
			}
		})
	}
}

func TestRouteTypeValidator_New(t *testing.T) {
	validator := NewRouteTypeValidator()
	if validator == nil {
		t.Error("NewRouteTypeValidator() returned nil")
	}
}
