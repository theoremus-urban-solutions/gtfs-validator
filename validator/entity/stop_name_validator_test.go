package entity

import (
	"fmt"
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/testutil"
)

func TestStopNameValidator_LoadStops(t *testing.T) {
	validator := NewStopNameValidator()

	tests := []struct {
		name     string
		csvData  string
		expected []*StopNameInfo
	}{
		{
			name: "basic stop loading",
			csvData: "stop_id,stop_name,stop_lat,stop_lon,location_type\n" +
				"stop1,Central Station,34.0522,-118.2437,1",
			expected: []*StopNameInfo{
				{
					StopID:       "stop1",
					StopName:     "Central Station",
					LocationType: 1,
					RowNumber:    2, // Header is row 1
				},
			},
		},
		{
			name: "stop with description and parent",
			csvData: "stop_id,stop_name,stop_desc,location_type,parent_station\n" +
				"stop1,Platform A,Main platform,0,station1",
			expected: []*StopNameInfo{
				{
					StopID:        "stop1",
					StopName:      "Platform A",
					StopDesc:      "Main platform",
					LocationType:  0,
					ParentStation: "station1",
					RowNumber:     2,
				},
			},
		},
		{
			name: "stops with whitespace trimming",
			csvData: "stop_id,stop_name,stop_desc,parent_station\n" +
				" stop1 , Central Station , Main station , station2 ",
			expected: []*StopNameInfo{
				{
					StopID:        "stop1",
					StopName:      "Central Station",
					StopDesc:      "Main station",
					LocationType:  0, // Default
					ParentStation: "station2",
					RowNumber:     2,
				},
			},
		},
		{
			name: "mixed location types",
			csvData: "stop_id,stop_name,location_type\n" +
				"stop1,Platform,0\n" +
				"stop2,Station,1\n" +
				"stop3,Entrance,2\n" +
				"stop4,Node,3\n" +
				"stop5,Boarding,4",
			expected: []*StopNameInfo{
				{StopID: "stop1", StopName: "Platform", LocationType: 0, RowNumber: 2},
				{StopID: "stop2", StopName: "Station", LocationType: 1, RowNumber: 3},
				{StopID: "stop3", StopName: "Entrance", LocationType: 2, RowNumber: 4},
				{StopID: "stop4", StopName: "Node", LocationType: 3, RowNumber: 5},
				{StopID: "stop5", StopName: "Boarding", LocationType: 4, RowNumber: 6},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			feedLoader := testutil.CreateTestFeedLoader(t, map[string]string{
				"stops.txt": tt.csvData,
			})

			result := validator.loadStops(feedLoader)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d stops, got %d", len(tt.expected), len(result))
			}

			for i, expectedStop := range tt.expected {
				if i >= len(result) {
					t.Errorf("Expected stop at index %d not found", i)
					continue
				}

				actualStop := result[i]
				if actualStop.StopID != expectedStop.StopID {
					t.Errorf("Stop %d: expected StopID %s, got %s", i, expectedStop.StopID, actualStop.StopID)
				}
				if actualStop.StopName != expectedStop.StopName {
					t.Errorf("Stop %d: expected StopName %s, got %s", i, expectedStop.StopName, actualStop.StopName)
				}
				if actualStop.StopDesc != expectedStop.StopDesc {
					t.Errorf("Stop %d: expected StopDesc %s, got %s", i, expectedStop.StopDesc, actualStop.StopDesc)
				}
				if actualStop.LocationType != expectedStop.LocationType {
					t.Errorf("Stop %d: expected LocationType %d, got %d", i, expectedStop.LocationType, actualStop.LocationType)
				}
				if actualStop.ParentStation != expectedStop.ParentStation {
					t.Errorf("Stop %d: expected ParentStation %s, got %s", i, expectedStop.ParentStation, actualStop.ParentStation)
				}
				if actualStop.RowNumber != expectedStop.RowNumber {
					t.Errorf("Stop %d: expected RowNumber %d, got %d", i, expectedStop.RowNumber, actualStop.RowNumber)
				}
			}
		})
	}
}

func TestStopNameValidator_ValidateStopName(t *testing.T) {
	validator := NewStopNameValidator()

	tests := []struct {
		name          string
		stop          *StopNameInfo
		expectedCodes []string
	}{
		{
			name:          "name and description say different things",
			stop:          &StopNameInfo{StopID: "stop1", StopName: "Main St", StopDesc: "Northbound platform"},
			expectedCodes: []string{},
		},
		{
			name:          "description duplicates the name",
			stop:          &StopNameInfo{StopID: "stop1", StopName: "Main St", StopDesc: "Main St"},
			expectedCodes: []string{"same_name_and_description_for_stop"},
		},
		{
			name:          "description duplicates the name in another case",
			stop:          &StopNameInfo{StopID: "stop1", StopName: "Main St", StopDesc: "MAIN ST"},
			expectedCodes: []string{"same_name_and_description_for_stop"},
		},
		{
			name:          "no description",
			stop:          &StopNameInfo{StopID: "stop1", StopName: "Main St"},
			expectedCodes: []string{},
		},
		{
			name:          "name in upper case only",
			stop:          &StopNameInfo{StopID: "stop1", StopName: "GALLERIA MALL"},
			expectedCodes: []string{"mixed_case_recommended_field"},
		},
		{
			name:          "name in lower case only",
			stop:          &StopNameInfo{StopID: "stop1", StopName: "central station"},
			expectedCodes: []string{"mixed_case_recommended_field"},
		},
		{
			name:          "name in a script without case",
			stop:          &StopNameInfo{StopID: "stop1", StopName: "東京駅"},
			expectedCodes: []string{},
		},
		{
			name:          "single case name duplicated in the description",
			stop:          &StopNameInfo{StopID: "stop1", StopName: "GALLERIA MALL", StopDesc: "GALLERIA MALL"},
			expectedCodes: []string{"mixed_case_recommended_field", "same_name_and_description_for_stop"},
		},
		{
			name:          "missing name on a platform",
			stop:          &StopNameInfo{StopID: "stop1", LocationType: 0},
			expectedCodes: []string{"missing_stop_name"},
		},
		{
			name:          "missing name on a generic node",
			stop:          &StopNameInfo{StopID: "stop1", LocationType: 3},
			expectedCodes: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			container := notice.NewNoticeContainer()

			validator.validateStopName(container, tt.stop, map[string]*StopNameInfo{})

			expectedCounts := make(map[string]int)
			for _, code := range tt.expectedCodes {
				expectedCounts[code]++
			}

			actualCounts := make(map[string]int)
			for _, n := range container.GetNotices() {
				actualCounts[n.Code()]++
			}

			for code, expected := range expectedCounts {
				if actualCounts[code] != expected {
					t.Errorf("Expected %d notices with code '%s', got %d", expected, code, actualCounts[code])
				}
			}

			for code := range actualCounts {
				if expectedCounts[code] == 0 {
					t.Errorf("Unexpected notice code: %s", code)
				}
			}
		})
	}
}

func TestStopNameValidator_IsStopNameRequired(t *testing.T) {
	validator := NewStopNameValidator()

	tests := []struct {
		locationType int
		required     bool
	}{
		{0, true},  // Stop/platform - required
		{1, true},  // Station - required
		{2, true},  // Entrance/exit - required
		{3, false}, // Generic node - optional
		{4, false}, // Boarding area - optional
		{5, false}, // Unknown types - not required
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("location_type_%d", tt.locationType), func(t *testing.T) {
			result := validator.isStopNameRequired(tt.locationType)
			if result != tt.required {
				t.Errorf("Location type %d: expected required=%v, got %v", tt.locationType, tt.required, result)
			}
		})
	}
}

func TestStopNameValidator_New(t *testing.T) {
	validator := NewStopNameValidator()
	if validator == nil {
		t.Error("NewStopNameValidator() returned nil")
	}
}
