package entity

import (
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/testutil"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

func TestStopTimeHeadsignValidator_Validate(t *testing.T) {
	tests := []struct {
		name                string
		files               map[string]string
		expectedNoticeCodes []string
		description         string
	}{
		{
			name: "consistent headsigns",
			files: map[string]string{
				"stop_times.txt": "trip_id,stop_id,stop_sequence,arrival_time,departure_time,stop_headsign\n" +
					"trip1,stop1,1,08:00:00,08:00:00,Downtown\n" +
					"trip1,stop2,2,08:05:00,08:05:00,Downtown\n" +
					"trip1,stop3,3,08:10:00,08:10:00,Downtown",
				"trips.txt": "trip_id,route_id,service_id,trip_headsign\n" +
					"trip1,route1,service1,Downtown",
			},
			expectedNoticeCodes: []string{},
			description:         "Consistent headsigns should not generate notices",
		},
		{
			name: "headsign change within trip",
			files: map[string]string{
				"stop_times.txt": "trip_id,stop_id,stop_sequence,arrival_time,departure_time,stop_headsign\n" +
					"trip1,stop1,1,08:00:00,08:00:00,Downtown\n" +
					"trip1,stop2,2,08:05:00,08:05:00,Uptown\n" +
					"trip1,stop3,3,08:10:00,08:10:00,Downtown",
				"trips.txt": "trip_id,route_id,service_id,trip_headsign\n" +
					"trip1,route1,service1,Downtown",
			},
			expectedNoticeCodes: []string{"headsign_change_within_trip", "headsign_change_within_trip"},
			description:         "Headsign changes within trip should generate info notices",
		},
		{
			name: "too many headsigns in trip",
			files: map[string]string{
				"stop_times.txt": "trip_id,stop_id,stop_sequence,arrival_time,departure_time,stop_headsign\n" +
					"trip1,stop1,1,08:00:00,08:00:00,Downtown\n" +
					"trip1,stop2,2,08:05:00,08:05:00,Uptown\n" +
					"trip1,stop3,3,08:10:00,08:10:00,Eastside\n" +
					"trip1,stop4,4,08:15:00,08:15:00,Westside\n" +
					"trip1,stop5,5,08:20:00,08:20:00,Northside\n" +
					"trip1,stop6,6,08:25:00,08:25:00,Southside",
				"trips.txt": "trip_id,route_id,service_id,trip_headsign\n" +
					"trip1,route1,service1,Downtown",
			},
			expectedNoticeCodes: []string{"too_many_headsigns_in_trip"},
			description:         "Too many different headsigns should generate warning",
		},
		{
			name: "frequent headsign changes",
			files: map[string]string{
				"stop_times.txt": "trip_id,stop_id,stop_sequence,arrival_time,departure_time,stop_headsign\n" +
					"trip1,stop1,1,08:00:00,08:00:00,A\n" +
					"trip1,stop2,2,08:05:00,08:05:00,B\n" +
					"trip1,stop3,3,08:10:00,08:10:00,C\n" +
					"trip1,stop4,4,08:15:00,08:15:00,D\n" +
					"trip1,stop5,5,08:20:00,08:20:00,E",
				"trips.txt": "trip_id,route_id,service_id,trip_headsign\n" +
					"trip1,route1,service1,Downtown",
			},
			expectedNoticeCodes: []string{"frequent_headsign_changes"},
			description:         "Frequent headsign changes should generate warning",
		},
		{
			name: "stop trip headsign mismatch",
			files: map[string]string{
				"stop_times.txt": "trip_id,stop_id,stop_sequence,arrival_time,departure_time,stop_headsign\n" +
					"trip1,stop1,1,08:00:00,08:00:00,Uptown\n" +
					"trip1,stop2,2,08:05:00,08:05:00,Eastside",
				"trips.txt": "trip_id,route_id,service_id,trip_headsign\n" +
					"trip1,route1,service1,Downtown",
			},
			expectedNoticeCodes: []string{"stop_trip_headsign_mismatch", "stop_trip_headsign_mismatch"},
			description:         "Stop headsigns conflicting with trip headsign should generate warnings",
		},
		{
			name: "consistent headsign variations",
			files: map[string]string{
				"stop_times.txt": "trip_id,stop_id,stop_sequence,arrival_time,departure_time,stop_headsign\n" +
					"trip1,stop1,1,08:00:00,08:00:00,Main Street\n" +
					"trip1,stop2,2,08:05:00,08:05:00,Main St",
				"trips.txt": "trip_id,route_id,service_id,trip_headsign\n" +
					"trip1,route1,service1,Main Street Station",
			},
			expectedNoticeCodes: []string{},
			description:         "Consistent headsign variations should not generate mismatches",
		},
		{
			name: "very short headsigns",
			files: map[string]string{
				"stop_times.txt": "trip_id,stop_id,stop_sequence,arrival_time,departure_time,stop_headsign\n" +
					"trip1,stop1,1,08:00:00,08:00:00,A\n" +
					"trip1,stop2,2,08:05:00,08:05:00,NB",
				"trips.txt": "trip_id,route_id,service_id,trip_headsign\n" +
					"trip1,route1,service1,Downtown",
			},
			expectedNoticeCodes: []string{"very_short_headsign", "very_short_headsign"},
			description:         "Very short headsigns should generate warnings",
		},
		{
			name: "very long headsigns",
			files: map[string]string{
				"stop_times.txt": "trip_id,stop_id,stop_sequence,arrival_time,departure_time,stop_headsign\n" +
					"trip1,stop1,1,08:00:00,08:00:00," + string(make([]rune, 120)) + "",
				"trips.txt": "trip_id,route_id,service_id,trip_headsign\n" +
					"trip1,route1,service1,Downtown",
			},
			expectedNoticeCodes: []string{"very_long_headsign"},
			description:         "Very long headsigns should generate warnings",
		},
		{
			name: "all caps headsigns",
			files: map[string]string{
				"stop_times.txt": "trip_id,stop_id,stop_sequence,arrival_time,departure_time,stop_headsign\n" +
					"trip1,stop1,1,08:00:00,08:00:00,DOWNTOWN STATION\n" +
					"trip1,stop2,2,08:05:00,08:05:00,GPS", // Short caps - OK
				"trips.txt": "trip_id,route_id,service_id,trip_headsign\n" +
					"trip1,route1,service1,Downtown Station",
			},
			expectedNoticeCodes: []string{"all_caps_headsign"},
			description:         "All caps headsigns should generate warnings, except short ones",
		},
		{
			name: "excessive punctuation",
			files: map[string]string{
				"stop_times.txt": "trip_id,stop_id,stop_sequence,arrival_time,departure_time,stop_headsign\n" +
					"trip1,stop1,1,08:00:00,08:00:00,Downtown!!!???###",
				"trips.txt": "trip_id,route_id,service_id,trip_headsign\n" +
					"trip1,route1,service1,Downtown",
			},
			expectedNoticeCodes: []string{"excessive_punctuation_headsign"},
			description:         "Excessive punctuation in headsigns should generate warnings",
		},
		{
			name: "suspicious headsign patterns",
			files: map[string]string{
				"stop_times.txt": "trip_id,stop_id,stop_sequence,arrival_time,departure_time,stop_headsign\n" +
					"trip1,stop1,1,08:00:00,08:00:00,NULL\n" +
					"trip1,stop2,2,08:05:00,08:05:00,N/A\n" +
					"trip1,stop3,3,08:10:00,08:10:00,Test Data",
				"trips.txt": "trip_id,route_id,service_id,trip_headsign\n" +
					"trip1,route1,service1,Downtown",
			},
			expectedNoticeCodes: []string{"suspicious_headsign_pattern", "suspicious_headsign_pattern", "suspicious_headsign_pattern"},
			description:         "Suspicious headsign patterns should generate warnings",
		},
		{
			name: "no headsigns",
			files: map[string]string{
				"stop_times.txt": "trip_id,stop_id,stop_sequence,arrival_time,departure_time\n" +
					"trip1,stop1,1,08:00:00,08:00:00\n" +
					"trip1,stop2,2,08:05:00,08:05:00",
				"trips.txt": "trip_id,route_id,service_id,trip_headsign\n" +
					"trip1,route1,service1,Downtown",
			},
			expectedNoticeCodes: []string{},
			description:         "No headsigns should not generate errors",
		},
		{
			name: "empty headsigns",
			files: map[string]string{
				"stop_times.txt": "trip_id,stop_id,stop_sequence,arrival_time,departure_time,stop_headsign\n" +
					"trip1,stop1,1,08:00:00,08:00:00,\n" +
					"trip1,stop2,2,08:05:00,08:05:00,   ",
				"trips.txt": "trip_id,route_id,service_id,trip_headsign\n" +
					"trip1,route1,service1,Downtown",
			},
			expectedNoticeCodes: []string{},
			description:         "Empty headsigns should be ignored",
		},
		{
			name: "mixed headsign issues",
			files: map[string]string{
				"stop_times.txt": "trip_id,stop_id,stop_sequence,arrival_time,departure_time,stop_headsign\n" +
					"trip1,stop1,1,08:00:00,08:00:00,DOWNTOWN!!!\n" +
					"trip1,stop2,2,08:05:00,08:05:00,A\n" +
					"trip1,stop3,3,08:10:00,08:10:00,NULL",
				"trips.txt": "trip_id,route_id,service_id,trip_headsign\n" +
					"trip1,route1,service1,Uptown",
			},
			expectedNoticeCodes: []string{
				"all_caps_headsign", "excessive_punctuation_headsign", "stop_trip_headsign_mismatch",
				"very_short_headsign", "stop_trip_headsign_mismatch",
				"suspicious_headsign_pattern", "stop_trip_headsign_mismatch",
				"headsign_change_within_trip", "headsign_change_within_trip",
			},
			description: "Multiple headsign issues should generate multiple notices",
		},
		{
			name: "common abbreviations match",
			files: map[string]string{
				"stop_times.txt": "trip_id,stop_id,stop_sequence,arrival_time,departure_time,stop_headsign\n" +
					"trip1,stop1,1,08:00:00,08:00:00,Downtown Terminal\n" +
					"trip1,stop2,2,08:05:00,08:05:00,Dtown Term",
				"trips.txt": "trip_id,route_id,service_id,trip_headsign\n" +
					"trip1,route1,service1,Downtown Terminal",
			},
			expectedNoticeCodes: []string{},
			description:         "Common abbreviations should be recognized as consistent",
		},
		{
			name: "multiple trips",
			files: map[string]string{
				"stop_times.txt": "trip_id,stop_id,stop_sequence,arrival_time,departure_time,stop_headsign\n" +
					"trip1,stop1,1,08:00:00,08:00:00,Downtown\n" +
					"trip1,stop2,2,08:05:00,08:05:00,Downtown\n" +
					"trip2,stop1,1,09:00:00,09:00:00,A\n" +
					"trip2,stop2,2,09:05:00,09:05:00,B",
				"trips.txt": "trip_id,route_id,service_id,trip_headsign\n" +
					"trip1,route1,service1,Downtown\n" +
					"trip2,route1,service1,Uptown",
			},
			expectedNoticeCodes: []string{
				"very_short_headsign", "stop_trip_headsign_mismatch",
				"very_short_headsign", "stop_trip_headsign_mismatch",
				"headsign_change_within_trip",
			},
			description: "Multiple trips should be validated independently",
		},
		{
			name: "no stop_times file",
			files: map[string]string{
				"trips.txt": "trip_id,route_id,service_id,trip_headsign\n" +
					"trip1,route1,service1,Downtown",
			},
			expectedNoticeCodes: []string{},
			description:         "Missing stop_times file should not generate errors",
		},
		{
			name: "no trips file",
			files: map[string]string{
				"stop_times.txt": "trip_id,stop_id,stop_sequence,arrival_time,departure_time,stop_headsign\n" +
					"trip1,stop1,1,08:00:00,08:00:00,Downtown",
			},
			expectedNoticeCodes: []string{},
			description:         "Missing trips file should not cause errors",
		},
		{
			name: "case insensitive consistency",
			files: map[string]string{
				"stop_times.txt": "trip_id,stop_id,stop_sequence,arrival_time,departure_time,stop_headsign\n" +
					"trip1,stop1,1,08:00:00,08:00:00,downtown\n" +
					"trip1,stop2,2,08:05:00,08:05:00,DOWNTOWN",
				"trips.txt": "trip_id,route_id,service_id,trip_headsign\n" +
					"trip1,route1,service1,Downtown",
			},
			expectedNoticeCodes: []string{"all_caps_headsign"},
			description:         "Case insensitive matching should work, but all caps still flagged",
		},
		{
			name: "whitespace handling",
			files: map[string]string{
				"stop_times.txt": "trip_id,stop_id,stop_sequence,arrival_time,departure_time,stop_headsign\n" +
					" trip1 , stop1 , 1 , 08:00:00 , 08:00:00 , Downtown \n" +
					" trip1 , stop2 , 2 , 08:05:00 , 08:05:00 ,  Downtown  ",
				"trips.txt": "trip_id,route_id,service_id,trip_headsign\n" +
					" trip1 , route1 , service1 ,  Downtown ",
			},
			expectedNoticeCodes: []string{},
			description:         "Whitespace should be trimmed properly",
		},
		{
			name: "invalid stop sequence ignored",
			files: map[string]string{
				"stop_times.txt": "trip_id,stop_id,stop_sequence,arrival_time,departure_time,stop_headsign\n" +
					"trip1,stop1,invalid,08:00:00,08:00:00,Downtown\n" +
					"trip1,stop2,2,08:05:00,08:05:00,Downtown",
				"trips.txt": "trip_id,route_id,service_id,trip_headsign\n" +
					"trip1,route1,service1,Downtown",
			},
			expectedNoticeCodes: []string{},
			description:         "Invalid stop sequences should be ignored gracefully",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test feed loader
			feedLoader := testutil.CreateTestFeedLoader(t, tt.files)

			// Create notice container and validator
			container := notice.NewNoticeContainer()
			validator := NewStopTimeHeadsignValidator()
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

func TestStopTimeHeadsignValidator_LoadStopTimeHeadsigns(t *testing.T) {
	validator := NewStopTimeHeadsignValidator()

	tests := []struct {
		name     string
		csvData  string
		expected []*StopTimeHeadsignInfo
	}{
		{
			name: "basic headsign loading",
			csvData: "trip_id,stop_id,stop_sequence,arrival_time,departure_time,stop_headsign\n" +
				"trip1,stop1,1,08:00:00,08:00:00,Downtown\n" +
				"trip1,stop2,2,08:05:00,08:05:00,Downtown",
			expected: []*StopTimeHeadsignInfo{
				{
					TripID:       "trip1",
					StopSequence: 1,
					StopHeadsign: "Downtown",
					RowNumber:    2,
				},
				{
					TripID:       "trip1",
					StopSequence: 2,
					StopHeadsign: "Downtown",
					RowNumber:    3,
				},
			},
		},
		{
			name: "without headsigns",
			csvData: "trip_id,stop_id,stop_sequence,arrival_time,departure_time\n" +
				"trip1,stop1,1,08:00:00,08:00:00\n" +
				"trip1,stop2,2,08:05:00,08:05:00",
			expected: []*StopTimeHeadsignInfo{
				{
					TripID:       "trip1",
					StopSequence: 1,
					StopHeadsign: "",
					RowNumber:    2,
				},
				{
					TripID:       "trip1",
					StopSequence: 2,
					StopHeadsign: "",
					RowNumber:    3,
				},
			},
		},
		{
			name: "whitespace trimming",
			csvData: "trip_id,stop_id,stop_sequence,arrival_time,departure_time,stop_headsign\n" +
				" trip1 , stop1 , 1 , 08:00:00 , 08:00:00 ,  Downtown  ",
			expected: []*StopTimeHeadsignInfo{
				{
					TripID:       "trip1",
					StopSequence: 1,
					StopHeadsign: "Downtown",
					RowNumber:    2,
				},
			},
		},
		{
			name: "mixed trips",
			csvData: "trip_id,stop_id,stop_sequence,arrival_time,departure_time,stop_headsign\n" +
				"trip1,stop1,1,08:00:00,08:00:00,Downtown\n" +
				"trip2,stop1,1,09:00:00,09:00:00,Uptown",
			expected: []*StopTimeHeadsignInfo{
				{
					TripID:       "trip1",
					StopSequence: 1,
					StopHeadsign: "Downtown",
					RowNumber:    2,
				},
				{
					TripID:       "trip2",
					StopSequence: 1,
					StopHeadsign: "Uptown",
					RowNumber:    3,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			feedLoader := testutil.CreateTestFeedLoader(t, map[string]string{
				"stop_times.txt": tt.csvData,
			})

			result := validator.loadStopTimeHeadsigns(feedLoader)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d headsigns, got %d", len(tt.expected), len(result))
			}

			for i, expectedHeadsign := range tt.expected {
				if i >= len(result) {
					t.Errorf("Expected headsign at index %d not found", i)
					continue
				}

				actualHeadsign := result[i]
				if actualHeadsign.TripID != expectedHeadsign.TripID {
					t.Errorf("Headsign %d: expected TripID %s, got %s", i, expectedHeadsign.TripID, actualHeadsign.TripID)
				}
				if actualHeadsign.StopSequence != expectedHeadsign.StopSequence {
					t.Errorf("Headsign %d: expected StopSequence %d, got %d", i, expectedHeadsign.StopSequence, actualHeadsign.StopSequence)
				}
				if actualHeadsign.StopHeadsign != expectedHeadsign.StopHeadsign {
					t.Errorf("Headsign %d: expected StopHeadsign %s, got %s", i, expectedHeadsign.StopHeadsign, actualHeadsign.StopHeadsign)
				}
				if actualHeadsign.RowNumber != expectedHeadsign.RowNumber {
					t.Errorf("Headsign %d: expected RowNumber %d, got %d", i, expectedHeadsign.RowNumber, actualHeadsign.RowNumber)
				}
			}
		})
	}
}

func TestStopTimeHeadsignValidator_LoadTripHeadsigns(t *testing.T) {
	validator := NewStopTimeHeadsignValidator()

	tests := []struct {
		name     string
		csvData  string
		expected map[string]*TripHeadsignInfo
	}{
		{
			name: "basic trip loading",
			csvData: "trip_id,route_id,service_id,trip_headsign\n" +
				"trip1,route1,service1,Downtown",
			expected: map[string]*TripHeadsignInfo{
				"trip1": {
					TripID:       "trip1",
					TripHeadsign: "Downtown",
					RouteID:      "route1",
				},
			},
		},
		{
			name: "without trip headsigns",
			csvData: "trip_id,route_id,service_id\n" +
				"trip1,route1,service1",
			expected: map[string]*TripHeadsignInfo{
				"trip1": {
					TripID:       "trip1",
					TripHeadsign: "",
					RouteID:      "route1",
				},
			},
		},
		{
			name: "whitespace trimming",
			csvData: "trip_id,route_id,service_id,trip_headsign\n" +
				" trip1 , route1 , service1 ,  Downtown  ",
			expected: map[string]*TripHeadsignInfo{
				"trip1": {
					TripID:       "trip1",
					TripHeadsign: "Downtown",
					RouteID:      "route1",
				},
			},
		},
		{
			name: "multiple trips",
			csvData: "trip_id,route_id,service_id,trip_headsign\n" +
				"trip1,route1,service1,Downtown\n" +
				"trip2,route1,service1,Uptown",
			expected: map[string]*TripHeadsignInfo{
				"trip1": {
					TripID:       "trip1",
					TripHeadsign: "Downtown",
					RouteID:      "route1",
				},
				"trip2": {
					TripID:       "trip2",
					TripHeadsign: "Uptown",
					RouteID:      "route1",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			feedLoader := testutil.CreateTestFeedLoader(t, map[string]string{
				"trips.txt": tt.csvData,
			})

			result := validator.loadTripHeadsigns(feedLoader)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d trip headsigns, got %d", len(tt.expected), len(result))
			}

			for tripID, expectedTrip := range tt.expected {
				actualTrip, exists := result[tripID]
				if !exists {
					t.Errorf("Expected trip %s not found", tripID)
					continue
				}

				if actualTrip.TripID != expectedTrip.TripID {
					t.Errorf("Trip %s: expected TripID %s, got %s", tripID, expectedTrip.TripID, actualTrip.TripID)
				}
				if actualTrip.TripHeadsign != expectedTrip.TripHeadsign {
					t.Errorf("Trip %s: expected TripHeadsign %s, got %s", tripID, expectedTrip.TripHeadsign, actualTrip.TripHeadsign)
				}
				if actualTrip.RouteID != expectedTrip.RouteID {
					t.Errorf("Trip %s: expected RouteID %s, got %s", tripID, expectedTrip.RouteID, actualTrip.RouteID)
				}
			}
		})
	}
}

func TestStopTimeHeadsignValidator_AreHeadsignsConsistent(t *testing.T) {
	validator := NewStopTimeHeadsignValidator()

	tests := []struct {
		stop        string
		trip        string
		consistent  bool
		description string
	}{
		{"Downtown", "Downtown", true, "Exact match"},
		{"downtown", "DOWNTOWN", true, "Case insensitive match"},
		{"Downtown Station", "Downtown", true, "Trip contains stop"},
		{"Downtown", "Downtown Terminal", true, "Stop contains trip"},
		{"Main St", "Main Street", true, "Street abbreviation"},
		{"Downtown Term", "Downtown Terminal", true, "Terminal abbreviation"},
		{"Dtown", "Downtown", true, "Downtown abbreviation"},
		{"University", "Univ", true, "University abbreviation"},
		{"Different", "Destination", false, "Completely different"},
		{"", "Downtown", false, "Empty stop headsign"},
		{"Downtown", "", false, "Empty trip headsign"},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			result := validator.areHeadsignsConsistent(tt.stop, tt.trip)
			if result != tt.consistent {
				t.Errorf("Headsigns '%s' and '%s': expected consistent=%v, got %v",
					tt.stop, tt.trip, tt.consistent, result)
			}
		})
	}
}

func TestStopTimeHeadsignValidator_AreHeadsignVariations(t *testing.T) {
	validator := NewStopTimeHeadsignValidator()

	tests := []struct {
		headsign1   string
		headsign2   string
		variations  bool
		description string
	}{
		{"main street", "main st", true, "Street abbreviation"},
		{"central avenue", "central ave", true, "Avenue abbreviation"},
		{"downtown boulevard", "downtown blvd", true, "Boulevard abbreviation"},
		{"downtown", "dtown", true, "Downtown abbreviation"},
		{"university hospital", "univ hosp", true, "Multiple abbreviations"},
		{"airport terminal", "apt trml", true, "Airport terminal abbreviation"},
		{"completely different", "words here", false, "No abbreviation match"},
		{"street", "road", false, "Different street types"},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			result := validator.areHeadsignVariations(tt.headsign1, tt.headsign2)
			if result != tt.variations {
				t.Errorf("Headsigns '%s' and '%s': expected variations=%v, got %v",
					tt.headsign1, tt.headsign2, tt.variations, result)
			}
		})
	}
}

func TestStopTimeHeadsignValidator_ContainsLowerCase(t *testing.T) {
	validator := NewStopTimeHeadsignValidator()

	tests := []struct {
		input    string
		expected bool
	}{
		{"DOWNTOWN", false},
		{"Downtown", true},
		{"MAIN St", true},
		{"123", false},
		{"GPS", false},
		{"Mix3d", true},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := validator.containsLowerCase(tt.input)
			if result != tt.expected {
				t.Errorf("String '%s': expected contains lowercase=%v, got %v",
					tt.input, tt.expected, result)
			}
		})
	}
}

func TestStopTimeHeadsignValidator_New(t *testing.T) {
	validator := NewStopTimeHeadsignValidator()
	if validator == nil {
		t.Error("NewStopTimeHeadsignValidator() returned nil")
	}
}
