package entity

import (
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

func TestBikesAllowanceValidator_Validate(t *testing.T) {
	tests := []struct {
		name                string
		files               map[string]string
		expectedNoticeCodes []string
		description         string
	}{
		{
			name: "ferry trip with bikes_allowed specified",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_type\nF1,Ferry,4",
				"trips.txt":  "route_id,service_id,trip_id,bikes_allowed\nF1,S1,T1,1",
			},
			expectedNoticeCodes: []string{},
			description:         "Ferry trip with bikes_allowed should be valid",
		},
		{
			name: "ferry trip missing bikes_allowed",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_type\nF1,Ferry,4",
				"trips.txt":  "route_id,service_id,trip_id\nF1,S1,T1",
			},
			expectedNoticeCodes: []string{"missing_bikes_allowed_for_ferry"},
			description:         "Ferry trip without bikes_allowed should generate notice",
		},
		{
			name: "non-ferry trip no bikes_allowed required",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_type\nB1,Bus,3",
				"trips.txt":  "route_id,service_id,trip_id\nB1,S1,T1",
			},
			expectedNoticeCodes: []string{},
			description:         "Non-ferry trip without bikes_allowed should be valid",
		},
		{
			name: "invalid bikes_allowed value",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_type\nF1,Ferry,4",
				"trips.txt":  "route_id,service_id,trip_id,bikes_allowed\nF1,S1,T1,3",
			},
			expectedNoticeCodes: []string{"invalid_bikes_allowed_value"},
			description:         "Invalid bikes_allowed value should generate notice",
		},
		{
			name: "valid bikes_allowed values",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_type\nF1,Ferry,4\nF2,Ferry,4\nF3,Ferry,4",
				"trips.txt":  "route_id,service_id,trip_id,bikes_allowed\nF1,S1,T1,0\nF2,S1,T2,1\nF3,S1,T3,2",
			},
			expectedNoticeCodes: []string{},
			description:         "All valid bikes_allowed values (0,1,2) should be accepted",
		},
		{
			name: "bikes allowed on uncommon route type",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_type\nM1,Metro,1",
				"trips.txt":  "route_id,service_id,trip_id,bikes_allowed\nM1,S1,T1,1",
			},
			expectedNoticeCodes: []string{"unusual_bike_allowance"},
			description:         "Bikes allowed on subway should generate unusual notice",
		},
		{
			name: "bike wheelchair accessibility mismatch",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_type\nF1,Ferry,4",
				"trips.txt":  "route_id,service_id,trip_id,bikes_allowed,wheelchair_accessible\nF1,S1,T1,2,1",
			},
			expectedNoticeCodes: []string{"bike_wheelchair_accessibility_mismatch"},
			description:         "Bikes not allowed but wheelchair accessible should generate info notice",
		},
		{
			name: "empty bikes_allowed ignored",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_type\nF1,Ferry,4",
				"trips.txt":  "route_id,service_id,trip_id,bikes_allowed\nF1,S1,T1,",
			},
			expectedNoticeCodes: []string{"missing_bikes_allowed_for_ferry"},
			description:         "Empty bikes_allowed should be treated as missing",
		},
		{
			name: "whitespace bikes_allowed ignored",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_type\nF1,Ferry,4",
				"trips.txt":  "route_id,service_id,trip_id,bikes_allowed\nF1,S1,T1,   ",
			},
			expectedNoticeCodes: []string{"missing_bikes_allowed_for_ferry"},
			description:         "Whitespace-only bikes_allowed should be treated as missing",
		},
		{
			name: "bikes_allowed with whitespace trimmed",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_type\nF1,Ferry,4",
				"trips.txt":  "route_id,service_id,trip_id,bikes_allowed\nF1,S1,T1, 1 ",
			},
			expectedNoticeCodes: []string{},
			description:         "Whitespace around bikes_allowed should be trimmed",
		},
		{
			name: "non-numeric bikes_allowed ignored",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_type\nF1,Ferry,4",
				"trips.txt":  "route_id,service_id,trip_id,bikes_allowed\nF1,S1,T1,invalid",
			},
			expectedNoticeCodes: []string{"missing_bikes_allowed_for_ferry"},
			description:         "Non-numeric bikes_allowed should be treated as missing",
		},
		{
			name: "missing route referenced by trip",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_type\nF1,Ferry,4",
				"trips.txt":  "route_id,service_id,trip_id,bikes_allowed\nF2,S1,T1,1",
			},
			expectedNoticeCodes: []string{},
			description:         "Trip referencing missing route should be ignored",
		},
		{
			name: "missing trip_id ignored",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_type\nF1,Ferry,4",
				"trips.txt":  "route_id,service_id,trip_id,bikes_allowed\nF1,S1,,1",
			},
			expectedNoticeCodes: []string{},
			description:         "Trip without trip_id should be ignored",
		},
		{
			name: "missing route_id in trip ignored",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_type\nF1,Ferry,4",
				"trips.txt":  "route_id,service_id,trip_id,bikes_allowed\n,S1,T1,1",
			},
			expectedNoticeCodes: []string{},
			description:         "Trip without route_id should be ignored",
		},
		{
			name: "multiple ferry trips some missing bikes_allowed",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_type\nF1,Ferry,4",
				"trips.txt":  "route_id,service_id,trip_id,bikes_allowed\nF1,S1,T1,1\nF1,S1,T2,\nF1,S1,T3,2",
			},
			expectedNoticeCodes: []string{"missing_bikes_allowed_for_ferry"},
			description:         "Multiple ferry trips with some missing bikes_allowed",
		},
		{
			name: "no routes.txt file",
			files: map[string]string{
				"trips.txt": "route_id,service_id,trip_id,bikes_allowed\nF1,S1,T1,1",
			},
			expectedNoticeCodes: []string{},
			description:         "Missing routes.txt should not cause errors",
		},
		{
			name: "no trips.txt file",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_type\nF1,Ferry,4",
			},
			expectedNoticeCodes: []string{},
			description:         "Missing trips.txt should not cause errors",
		},
		{
			name: "invalid route_type in routes ignored",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_type\nF1,Ferry,invalid",
				"trips.txt":  "route_id,service_id,trip_id,bikes_allowed\nF1,S1,T1,1",
			},
			expectedNoticeCodes: []string{},
			description:         "Routes with invalid route_type should be ignored",
		},
		{
			name: "buses with bikes allowed - common route type",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_type\nB1,Bus,3",
				"trips.txt":  "route_id,service_id,trip_id,bikes_allowed\nB1,S1,T1,1",
			},
			expectedNoticeCodes: []string{},
			description:         "Bikes allowed on buses should not generate unusual notice",
		},
		{
			name: "complex scenario with multiple route types",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_type\nF1,Ferry,4\nB1,Bus,3\nM1,Metro,1\nT1,Tram,0",
				"trips.txt":  "route_id,service_id,trip_id,bikes_allowed,wheelchair_accessible\nF1,S1,T1,,\nB1,S1,T2,1,1\nM1,S1,T3,1,0\nT1,S1,T4,2,1",
			},
			expectedNoticeCodes: []string{"missing_bikes_allowed_for_ferry", "unusual_bike_allowance", "unusual_bike_allowance", "bike_wheelchair_accessibility_mismatch"},
			description:         "Complex scenario with multiple validation cases",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := CreateTestFeedLoader(t, tt.files)
			container := notice.NewNoticeContainer()
			validator := NewBikesAllowanceValidator()
			config := gtfsvalidator.Config{}

			validator.Validate(loader, container, config)

			notices := container.GetNotices()

			if len(notices) != len(tt.expectedNoticeCodes) {
				t.Errorf("Expected %d notices, got %d for case: %s", len(tt.expectedNoticeCodes), len(notices), tt.description)
			}

			expectedCodeCounts := make(map[string]int)
			for _, code := range tt.expectedNoticeCodes {
				expectedCodeCounts[code]++
			}

			actualCodeCounts := make(map[string]int)
			for _, notice := range notices {
				actualCodeCounts[notice.Code()]++
			}

			for expectedCode, expectedCount := range expectedCodeCounts {
				actualCount := actualCodeCounts[expectedCode]
				if actualCount != expectedCount {
					t.Errorf("Expected %d notices with code '%s', got %d", expectedCount, expectedCode, actualCount)
				}
			}

			for actualCode := range actualCodeCounts {
				if expectedCodeCounts[actualCode] == 0 {
					t.Errorf("Unexpected notice code: %s", actualCode)
				}
			}
		})
	}
}

func TestBikesAllowanceValidator_IsBikeUncommonRouteType(t *testing.T) {
	validator := NewBikesAllowanceValidator()

	uncommonTypes := []int{0, 1, 2, 5, 6, 7} // Tram, Subway, Rail, Cable tram, Aerial lift, Funicular
	commonTypes := []int{3, 4, 11, 12}       // Bus, Ferry, Trolleybus, Monorail

	for _, routeType := range uncommonTypes {
		if !validator.isBikeUncommonRouteType(routeType) {
			t.Errorf("Expected route type %d to be uncommon for bikes", routeType)
		}
	}

	for _, routeType := range commonTypes {
		if validator.isBikeUncommonRouteType(routeType) {
			t.Errorf("Expected route type %d to be common for bikes", routeType)
		}
	}
}

func TestBikesAllowanceValidator_New(t *testing.T) {
	validator := NewBikesAllowanceValidator()
	if validator == nil {
		t.Error("NewBikesAllowanceValidator() returned nil")
	}
}
