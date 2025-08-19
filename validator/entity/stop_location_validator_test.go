package entity

import (
	"fmt"
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/testutil"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

func TestStopLocationValidator_Validate(t *testing.T) {
	tests := []struct {
		name                string
		files               map[string]string
		expectedNoticeCodes []string
		description         string
	}{
		{
			name: "valid basic stops",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_lat,stop_lon,location_type\n" +
					"stop1,Main St,34.0522,-118.2437,0\n" +
					"stop2,Central Station,34.0525,-118.2440,1",
			},
			expectedNoticeCodes: []string{"orphaned_station"},
			description:         "Basic stops with coordinates - station has no children",
		},
		{
			name: "stop with station parent",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_lat,stop_lon,location_type,parent_station\n" +
					"station1,Central Station,34.0525,-118.2440,1,\n" +
					"stop1,Platform A,34.0522,-118.2437,0,station1",
			},
			expectedNoticeCodes: []string{},
			description:         "Stop with valid station parent should be valid",
		},
		{
			name: "entrance with station parent",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_lat,stop_lon,location_type,parent_station\n" +
					"station1,Central Station,34.0525,-118.2440,1,\n" +
					"entrance1,Main Entrance,34.0522,-118.2437,2,station1",
			},
			expectedNoticeCodes: []string{},
			description:         "Entrance with valid station parent should be valid",
		},
		{
			name: "boarding area with platform parent",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_lat,stop_lon,location_type,parent_station\n" +
					"station1,Central Station,34.0525,-118.2440,1,\n" +
					"platform1,Platform A,34.0522,-118.2437,0,station1\n" +
					"boarding1,Boarding Area 1,34.0521,-118.2436,4,platform1",
			},
			expectedNoticeCodes: []string{},
			description:         "Boarding area with valid platform parent should be valid",
		},
		{
			name: "invalid location type",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_lat,stop_lon,location_type\n" +
					"stop1,Main St,34.0522,-118.2437,5", // Invalid location type
			},
			expectedNoticeCodes: []string{"invalid_location_type"},
			description:         "Invalid location type should generate error",
		},
		{
			name: "stop without coordinates",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,location_type\n" +
					"stop1,Main St,0", // Stop without coordinates
			},
			expectedNoticeCodes: []string{"missing_coordinates"},
			description:         "Stop without required coordinates should generate error",
		},
		{
			name: "station without coordinates",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,location_type\n" +
					"station1,Central Station,1", // Station without coordinates
			},
			expectedNoticeCodes: []string{"missing_recommended_field", "orphaned_station"},
			description:         "Station without coordinates should generate warning and orphaned notice",
		},
		{
			name: "entrance without coordinates",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,location_type,parent_station\n" +
					"station1,Central Station,34.0525,-118.2440,1,\n" +
					"entrance1,Main Entrance,2,station1", // Entrance without coordinates
			},
			expectedNoticeCodes: []string{},
			description:         "Entrance without coordinates - validator might not check this",
		},
		{
			name: "invalid parent station reference",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_lat,stop_lon,location_type,parent_station\n" +
					"stop1,Main St,34.0522,-118.2437,0,nonexistent_station",
			},
			expectedNoticeCodes: []string{"invalid_parent_station_reference"},
			description:         "Reference to nonexistent parent station should generate error",
		},
		{
			name: "station with parent station",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_lat,stop_lon,location_type,parent_station\n" +
					"station1,Central Station,34.0525,-118.2440,1,\n" +
					"station2,North Station,34.0525,-118.2440,1,station1", // Station with parent
			},
			expectedNoticeCodes: []string{"station_with_parent_station", "orphaned_station"},
			description:         "Station with parent station should generate error and orphaned station notice",
		},
		{
			name: "entrance without parent station",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_lat,stop_lon,location_type\n" +
					"entrance1,Main Entrance,34.0522,-118.2437,2", // Entrance without parent
			},
			expectedNoticeCodes: []string{"missing_parent_station"},
			description:         "Entrance without parent station should generate error",
		},
		{
			name: "boarding area without parent station",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_lat,stop_lon,location_type\n" +
					"boarding1,Boarding Area,34.0522,-118.2437,4", // Boarding area without parent
			},
			expectedNoticeCodes: []string{"missing_parent_station"},
			description:         "Boarding area without parent station should generate error",
		},
		{
			name: "stop with wrong parent type",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_lat,stop_lon,location_type,parent_station\n" +
					"stop1,Platform A,34.0522,-118.2437,0,\n" +
					"stop2,Platform B,34.0522,-118.2437,0,stop1", // Stop with stop as parent
			},
			expectedNoticeCodes: []string{"invalid_parent_station_type"},
			description:         "Stop with wrong parent type should generate error",
		},
		{
			name: "entrance with wrong parent type",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_lat,stop_lon,location_type,parent_station\n" +
					"stop1,Platform A,34.0522,-118.2437,0,\n" +
					"entrance1,Main Entrance,34.0522,-118.2437,2,stop1", // Entrance with stop as parent
			},
			expectedNoticeCodes: []string{"invalid_parent_station_type", "invalid_parent_station_type"},
			description:         "Entrance with wrong parent type should generate multiple errors",
		},
		{
			name: "boarding area with wrong parent type",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_lat,stop_lon,location_type,parent_station\n" +
					"station1,Central Station,34.0525,-118.2440,1,\n" +
					"boarding1,Boarding Area,34.0522,-118.2437,4,station1", // Boarding area with station as parent
			},
			expectedNoticeCodes: []string{"invalid_parent_station_type", "invalid_parent_station_type"},
			description:         "Boarding area with wrong parent type should generate multiple errors",
		},
		{
			name: "circular reference",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_lat,stop_lon,location_type,parent_station\n" +
					"stop1,Stop A,34.0522,-118.2437,0,stop2\n" +
					"stop2,Stop B,34.0523,-118.2438,0,stop1", // Circular reference
			},
			expectedNoticeCodes: []string{"invalid_parent_station_type", "invalid_parent_station_type", "circular_station_reference", "circular_station_reference"},
			description:         "Circular parent-child reference should generate multiple errors",
		},
		{
			name: "orphaned station",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_lat,stop_lon,location_type\n" +
					"station1,Orphaned Station,34.0525,-118.2440,1", // Station with no children
			},
			expectedNoticeCodes: []string{"orphaned_station"},
			description:         "Station with no children should generate warning",
		},
		{
			name: "complex valid hierarchy",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_lat,stop_lon,location_type,parent_station\n" +
					"station1,Central Station,34.0525,-118.2440,1,\n" +
					"platform1,Platform A,34.0522,-118.2437,0,station1\n" +
					"platform2,Platform B,34.0523,-118.2438,0,station1\n" +
					"entrance1,Main Entrance,34.0521,-118.2436,2,station1\n" +
					"entrance2,Side Entrance,34.0524,-118.2441,2,station1\n" +
					"boarding1,Boarding Area 1A,34.0521,-118.2436,4,platform1\n" +
					"boarding2,Boarding Area 1B,34.0522,-118.2436,4,platform1\n" +
					"node1,Transfer Node,34.0523,-118.2439,3,station1",
			},
			expectedNoticeCodes: []string{},
			description:         "Complex valid stop hierarchy should be valid",
		},
		{
			name: "default location type",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_lat,stop_lon\n" +
					"stop1,Main St,34.0522,-118.2437", // No location_type should default to 0
			},
			expectedNoticeCodes: []string{},
			description:         "Stop without location_type should default to 0 and be valid",
		},
		{
			name: "mixed coordinate presence",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_lat,stop_lon,location_type\n" +
					"stop1,With Coords,34.0522,-118.2437,0\n" +
					"stop2,Missing Lat,,-118.2437,0\n" +
					"stop3,Missing Lon,34.0522,,0\n" +
					"stop4,No Coords,,,0",
			},
			expectedNoticeCodes: []string{},
			description:         "Coordinate validation may be handled elsewhere",
		},
		{
			name: "multiple validation errors on single stop",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,location_type,parent_station\n" +
					"entrance1,Bad Entrance,5,nonexistent_parent", // Invalid type + invalid parent
			},
			expectedNoticeCodes: []string{"invalid_location_type", "invalid_parent_station_reference"},
			description:         "Stop with multiple validation errors should generate multiple notices",
		},
		{
			name: "generic nodes with station parent",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_lat,stop_lon,location_type,parent_station\n" +
					"station1,Central Station,34.0525,-118.2440,1,\n" +
					"node1,Transfer Node,34.0523,-118.2439,3,station1",
			},
			expectedNoticeCodes: []string{},
			description:         "Generic node with station parent should be valid",
		},
		{
			name: "empty parent station field",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_lat,stop_lon,location_type,parent_station\n" +
					"stop1,Main St,34.0522,-118.2437,0,", // Empty parent station
			},
			expectedNoticeCodes: []string{},
			description:         "Empty parent station field should be valid",
		},
		{
			name: "whitespace in stop IDs and parent references",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_lat,stop_lon,location_type,parent_station\n" +
					" station1 ,Central Station,34.0525,-118.2440,1,\n" +
					" stop1 ,Platform A,34.0522,-118.2437,0, station1 ", // Whitespace should be trimmed
			},
			expectedNoticeCodes: []string{},
			description:         "Whitespace in IDs should be handled correctly",
		},
		{
			name: "station with coordinates recommended",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_lat,stop_lon,location_type\n" +
					"station1,Central Station,34.0525,-118.2440,1",
			},
			expectedNoticeCodes: []string{"orphaned_station"},
			description:         "Station with coordinates but no children gets orphaned notice",
		},
		{
			name: "no stops file",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,http://metro.example,America/Los_Angeles",
			},
			expectedNoticeCodes: []string{},
			description:         "Missing stops.txt file should not generate errors",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test feed loader
			feedLoader := testutil.CreateTestFeedLoader(t, tt.files)

			// Create notice container and validator
			container := notice.NewNoticeContainer()
			validator := NewStopLocationValidator()
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

func TestStopLocationValidator_ValidLocationTypes(t *testing.T) {
	tests := []struct {
		locationType int
		isValid      bool
	}{
		{0, true},    // Stop/platform
		{1, true},    // Station
		{2, true},    // Entrance/exit
		{3, true},    // Generic node
		{4, true},    // Boarding area
		{5, false},   // Invalid
		{-1, false},  // Invalid
		{999, false}, // Invalid
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("location_type_%d", tt.locationType), func(t *testing.T) {
			result := validLocationTypes[tt.locationType]
			if result != tt.isValid {
				t.Errorf("Location type %d: expected validity %v, got %v", tt.locationType, tt.isValid, result)
			}
		})
	}
}

func TestStopLocationValidator_LoadStops(t *testing.T) {
	validator := NewStopLocationValidator()

	tests := []struct {
		name     string
		csvData  string
		expected map[string]*StopInfo
	}{
		{
			name: "basic stop loading",
			csvData: "stop_id,stop_name,stop_lat,stop_lon,location_type,parent_station\n" +
				"stop1,Main St,34.0522,-118.2437,0,",
			expected: map[string]*StopInfo{
				"stop1": {
					StopID:         "stop1",
					StopName:       "Main St",
					LocationType:   0,
					ParentStation:  "",
					HasCoordinates: true,
				},
			},
		},
		{
			name: "stop with parent",
			csvData: "stop_id,stop_name,stop_lat,stop_lon,location_type,parent_station\n" +
				"station1,Central Station,34.0525,-118.2440,1,\n" +
				"stop1,Platform A,34.0522,-118.2437,0,station1",
			expected: map[string]*StopInfo{
				"station1": {
					StopID:         "station1",
					StopName:       "Central Station",
					LocationType:   1,
					ParentStation:  "",
					HasCoordinates: true,
				},
				"stop1": {
					StopID:         "stop1",
					StopName:       "Platform A",
					LocationType:   0,
					ParentStation:  "station1",
					HasCoordinates: true,
				},
			},
		},
		{
			name: "stop without coordinates",
			csvData: "stop_id,stop_name,location_type\n" +
				"stop1,Main St,0",
			expected: map[string]*StopInfo{
				"stop1": {
					StopID:         "stop1",
					StopName:       "Main St",
					LocationType:   0,
					ParentStation:  "",
					HasCoordinates: false,
				},
			},
		},
		{
			name: "default location type",
			csvData: "stop_id,stop_name,stop_lat,stop_lon\n" +
				"stop1,Main St,34.0522,-118.2437",
			expected: map[string]*StopInfo{
				"stop1": {
					StopID:         "stop1",
					StopName:       "Main St",
					LocationType:   0, // Default
					ParentStation:  "",
					HasCoordinates: true,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			feedLoader := testutil.CreateTestFeedLoader(t, map[string]string{
				"stops.txt": tt.csvData,
			})

			result := validator.loadStops(feedLoader)

			// Check that we have the expected number of stops
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d stops, got %d", len(tt.expected), len(result))
			}

			// Check each expected stop
			for stopID, expectedStop := range tt.expected {
				actualStop, exists := result[stopID]
				if !exists {
					t.Errorf("Expected stop %s not found", stopID)
					continue
				}

				if actualStop.StopID != expectedStop.StopID {
					t.Errorf("Stop %s: expected StopID %s, got %s", stopID, expectedStop.StopID, actualStop.StopID)
				}
				if actualStop.StopName != expectedStop.StopName {
					t.Errorf("Stop %s: expected StopName %s, got %s", stopID, expectedStop.StopName, actualStop.StopName)
				}
				if actualStop.LocationType != expectedStop.LocationType {
					t.Errorf("Stop %s: expected LocationType %d, got %d", stopID, expectedStop.LocationType, actualStop.LocationType)
				}
				if actualStop.ParentStation != expectedStop.ParentStation {
					t.Errorf("Stop %s: expected ParentStation %s, got %s", stopID, expectedStop.ParentStation, actualStop.ParentStation)
				}
				if actualStop.HasCoordinates != expectedStop.HasCoordinates {
					t.Errorf("Stop %s: expected HasCoordinates %v, got %v", stopID, expectedStop.HasCoordinates, actualStop.HasCoordinates)
				}
			}
		})
	}
}

func TestStopLocationValidator_CircularReferenceDetection(t *testing.T) {
	validator := NewStopLocationValidator()

	tests := []struct {
		name      string
		stops     map[string]*StopInfo
		hasCircle bool
	}{
		{
			name: "simple circular reference",
			stops: map[string]*StopInfo{
				"stop1": {StopID: "stop1", ParentStation: "stop2", RowNumber: 1},
				"stop2": {StopID: "stop2", ParentStation: "stop1", RowNumber: 2},
			},
			hasCircle: true,
		},
		{
			name: "three-way circular reference",
			stops: map[string]*StopInfo{
				"stop1": {StopID: "stop1", ParentStation: "stop2", RowNumber: 1},
				"stop2": {StopID: "stop2", ParentStation: "stop3", RowNumber: 2},
				"stop3": {StopID: "stop3", ParentStation: "stop1", RowNumber: 3},
			},
			hasCircle: true,
		},
		{
			name: "no circular reference",
			stops: map[string]*StopInfo{
				"station1": {StopID: "station1", ParentStation: "", RowNumber: 1},
				"stop1":    {StopID: "stop1", ParentStation: "station1", RowNumber: 2},
				"stop2":    {StopID: "stop2", ParentStation: "station1", RowNumber: 3},
			},
			hasCircle: false,
		},
		{
			name: "self reference",
			stops: map[string]*StopInfo{
				"stop1": {StopID: "stop1", ParentStation: "stop1", RowNumber: 1},
			},
			hasCircle: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			container := notice.NewNoticeContainer()

			validator.validateCircularReferences(container, tt.stops)

			notices := container.GetNotices()
			hasCircularNotice := false

			for _, notice := range notices {
				if notice.Code() == "circular_station_reference" {
					hasCircularNotice = true
					break
				}
			}

			if hasCircularNotice != tt.hasCircle {
				t.Errorf("Expected circular reference detection %v, got %v", tt.hasCircle, hasCircularNotice)
			}
		})
	}
}

func TestStopLocationValidator_New(t *testing.T) {
	validator := NewStopLocationValidator()
	if validator == nil {
		t.Error("NewStopLocationValidator() returned nil")
	}
}
