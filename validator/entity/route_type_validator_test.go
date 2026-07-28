package entity

import (
	"fmt"
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/testutil"
)

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
			feedLoader := testutil.CreateTestFeedLoader(t, map[string]string{
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
