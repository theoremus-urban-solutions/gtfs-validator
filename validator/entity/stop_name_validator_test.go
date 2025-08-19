package entity

import (
	"github.com/theoremus-urban-solutions/gtfs-validator/testutil"
	"fmt"
	"strings"
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

func TestStopNameValidator_Validate(t *testing.T) {
	tests := []struct {
		name                string
		files               map[string]string
		expectedNoticeCodes []string
		description         string
	}{
		{
			name: "valid stop names",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_lat,stop_lon,location_type\n" +
					"stop1,Central Station,34.0522,-118.2437,0\n" +
					"stop2,Main Street Platform,34.0525,-118.2440,1\n" +
					"stop3,North Entrance,34.0523,-118.2438,2",
			},
			expectedNoticeCodes: []string{},
			description:         "Valid stop names should not generate notices",
		},
		{
			name: "missing required stop names",
			files: map[string]string{
				"stops.txt": "stop_id,stop_lat,stop_lon,location_type\n" +
					"stop1,34.0522,-118.2437,0\n" +
					"stop2,34.0525,-118.2440,1\n" +
					"stop3,34.0523,-118.2438,2",
			},
			expectedNoticeCodes: []string{"missing_required_stop_name", "missing_required_stop_name", "missing_required_stop_name"},
			description:         "Missing required stop names should generate errors",
		},
		{
			name: "optional stop names missing",
			files: map[string]string{
				"stops.txt": "stop_id,stop_lat,stop_lon,location_type\n" +
					"stop1,34.0522,-118.2437,3\n" +
					"stop2,34.0525,-118.2440,4",
			},
			expectedNoticeCodes: []string{},
			description:         "Missing optional stop names (generic node, boarding area) should not generate notices",
		},
		{
			name: "generic stop names",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_lat,stop_lon,location_type\n" +
					"stop1,Stop,34.0522,-118.2437,0\n" +
					"stop2,STATION,34.0525,-118.2440,1\n" +
					"stop3,platform,34.0523,-118.2438,0\n" +
					"stop4,Test,34.0524,-118.2439,0",
			},
			expectedNoticeCodes: []string{"generic_stop_name", "generic_stop_name", "stop_name_all_caps", "generic_stop_name", "generic_stop_name"},
			description:         "Generic stop names should generate warnings",
		},
		{
			name: "child stop inheriting parent name",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_lat,stop_lon,location_type,parent_station\n" +
					"station1,Central Station,34.0525,-118.2440,1,\n" +
					"stop1,,34.0522,-118.2437,0,station1",
			},
			expectedNoticeCodes: []string{"stop_name_missing_but_inherited"},
			description:         "Child stop without name should inherit from parent with info notice",
		},
		{
			name: "child stop with missing parent name",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_lat,stop_lon,location_type,parent_station\n" +
					"station1,,34.0525,-118.2440,1,\n" +
					"stop1,,34.0522,-118.2437,0,station1",
			},
			expectedNoticeCodes: []string{"missing_required_stop_name", "missing_required_stop_name"},
			description:         "Child stop without name and parent without name should generate errors",
		},
		{
			name: "long stop names",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_lat,stop_lon,location_type\n" +
					"stop1,This is a very long stop name that exceeds one hundred characters and should generate a warning notice for being too long for readability,34.0522,-118.2437,0\n" +
					"stop2," + strings.Repeat("A", 300) + ",34.0525,-118.2440,1",
			},
			expectedNoticeCodes: []string{"stop_name_too_long", "stop_name_too_long", "stop_name_all_caps"},
			description:         "Excessively long stop names should generate warnings/errors",
		},
		{
			name: "all caps stop names",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_lat,stop_lon,location_type\n" +
					"stop1,MAIN STREET STATION,34.0522,-118.2437,0\n" +
					"stop2,DOWNTOWN PLATFORM,34.0525,-118.2440,1\n" +
					"stop3,GPS,34.0523,-118.2438,0", // Short abbreviation - OK
			},
			expectedNoticeCodes: []string{"stop_name_all_caps", "stop_name_all_caps"},
			description:         "All caps stop names should generate warnings, except short abbreviations",
		},
		{
			name: "repeated words in names",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_lat,stop_lon,location_type\n" +
					"stop1,Station Station,34.0522,-118.2437,0\n" +
					"stop2,Main Main Street,34.0525,-118.2440,1\n" +
					"stop3,The The Plaza,34.0523,-118.2438,0",
			},
			expectedNoticeCodes: []string{"generic_stop_name", "stop_name_repeated_word", "stop_name_repeated_word", "stop_name_repeated_word"},
			description:         "Repeated consecutive words should generate warnings",
		},
		{
			name: "identical name and description",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_desc,stop_lat,stop_lon,location_type\n" +
					"stop1,Central Station,Central Station,34.0522,-118.2437,0\n" +
					"stop2,Main Platform,Different Description,34.0525,-118.2440,1",
			},
			expectedNoticeCodes: []string{"stop_name_description_duplicate"},
			description:         "Identical stop name and description should generate warning",
		},
		{
			name: "problematic characters in names",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_lat,stop_lon,location_type\n" +
					"stop1,<b>Bold Station</b>,34.0522,-118.2437,0\n" +
					"stop2,Visit http://example.com Station,34.0525,-118.2440,1\n" +
					"stop3,Control" + string(rune(7)) + "Character,34.0523,-118.2438,0",
			},
			expectedNoticeCodes: []string{"stop_name_contains_html", "stop_name_contains_url", "stop_name_contains_control_character"},
			description:         "Problematic characters should generate warnings",
		},
		{
			name: "mixed validation issues",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_desc,stop_lat,stop_lon,location_type\n" +
					"stop1,STOP,STOP,34.0522,-118.2437,0", // Generic + all caps + duplicate desc
			},
			expectedNoticeCodes: []string{
				"generic_stop_name",
				"stop_name_all_caps",
				"stop_name_description_duplicate",
			},
			description: "Multiple validation issues should generate multiple notices",
		},
		{
			name: "valid edge cases",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_lat,stop_lon,location_type\n" +
					"stop1,42nd St-Times Sq,34.0522,-118.2437,0\n" +
					"stop2,St. Mary's Hospital,34.0525,-118.2440,1\n" +
					"stop3,O'Hare Airport,34.0523,-118.2438,0\n" +
					"stop4,Müller-Straße,34.0524,-118.2439,0",
			},
			expectedNoticeCodes: []string{},
			description:         "Valid stop names with numbers, apostrophes, and unicode should be fine",
		},
		{
			name: "empty stop names",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_lat,stop_lon,location_type\n" +
					"stop1,,34.0522,-118.2437,0\n" +
					"stop2,   ,34.0525,-118.2440,1\n" + // Whitespace only
					"stop3,,34.0523,-118.2438,3", // Optional for generic node
			},
			expectedNoticeCodes: []string{"missing_required_stop_name", "missing_required_stop_name"},
			description:         "Empty and whitespace-only names should be treated as missing for required types",
		},
		{
			name: "various location types",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_lat,stop_lon,location_type\n" +
					"stop1,Platform A,34.0522,-118.2437,0\n" +
					"stop2,Central Station,34.0525,-118.2440,1\n" +
					"stop3,Main Entrance,34.0523,-118.2438,2\n" +
					"stop4,,34.0524,-118.2439,3\n" + // Generic node - optional name
					"stop5,,34.0526,-118.2441,4", // Boarding area - optional name
			},
			expectedNoticeCodes: []string{},
			description:         "Different location types with appropriate naming should be valid",
		},
		{
			name: "placeholder names",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_lat,stop_lon,location_type\n" +
					"stop1,Placeholder,34.0522,-118.2437,0\n" +
					"stop2,TBD,34.0525,-118.2440,1\n" +
					"stop3,TODO,34.0523,-118.2438,0\n" +
					"stop4,XXX,34.0524,-118.2439,0",
			},
			expectedNoticeCodes: []string{"generic_stop_name", "generic_stop_name", "generic_stop_name", "stop_name_all_caps", "generic_stop_name"},
			description:         "Placeholder names should be flagged as generic",
		},
		{
			name: "whitespace handling",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_lat,stop_lon,location_type\n" +
					"stop1, Central Station ,34.0522,-118.2437,0\n" +
					"stop2,  Main   Street  Platform  ,34.0525,-118.2440,1",
			},
			expectedNoticeCodes: []string{},
			description:         "Whitespace should be trimmed and handled properly",
		},
		{
			name: "no stops file",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,http://metro.example,America/Los_Angeles",
			},
			expectedNoticeCodes: []string{},
			description:         "Missing stops.txt file should not generate errors",
		},
		{
			name: "stops without stop_id ignored",
			files: map[string]string{
				"stops.txt": "stop_name,stop_lat,stop_lon,location_type\n" +
					"Central Station,34.0522,-118.2437,0",
			},
			expectedNoticeCodes: []string{},
			description:         "Stops without stop_id should be ignored",
		},
		{
			name: "repeated generic words",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_lat,stop_lon,location_type\n" +
					"stop1,stop stop,34.0522,-118.2437,0\n" +
					"stop2,station station,34.0525,-118.2440,1",
			},
			expectedNoticeCodes: []string{"generic_stop_name", "stop_name_repeated_word", "generic_stop_name", "stop_name_repeated_word"},
			description:         "Repeated generic words should be caught",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test feed loader
			feedLoader := testutil.CreateTestFeedLoader(t, tt.files)

			// Create notice container and validator
			container := notice.NewNoticeContainer()
			validator := NewStopNameValidator()
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

func TestStopNameValidator_CheckGenericStopName(t *testing.T) {
	validator := NewStopNameValidator()

	tests := []struct {
		name      string
		stopName  string
		isGeneric bool
	}{
		{"normal name", "Central Station", false},
		{"generic stop", "Stop", true},
		{"generic station", "STATION", true},
		{"generic platform", "platform", true},
		{"generic test", "Test", true},
		{"generic placeholder", "Placeholder", true},
		{"generic tbd", "TBD", true},
		{"generic unknown", "unknown", true},
		{"repeated generic", "stop stop", true},
		{"partial match", "Central Stop", false}, // Not just "stop"
		{"valid name with generic word", "42nd Street Stop", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			container := notice.NewNoticeContainer()
			stop := &StopNameInfo{
				StopID:    "test_stop",
				StopName:  tt.stopName,
				RowNumber: 1,
			}

			validator.checkGenericStopName(container, stop)

			notices := container.GetNotices()
			hasGenericNotice := false
			for _, notice := range notices {
				if notice.Code() == "generic_stop_name" {
					hasGenericNotice = true
					break
				}
			}

			if hasGenericNotice != tt.isGeneric {
				t.Errorf("Stop name '%s': expected generic=%v, got %v", tt.stopName, tt.isGeneric, hasGenericNotice)
			}
		})
	}
}

func TestStopNameValidator_CheckStopNameLength(t *testing.T) {
	validator := NewStopNameValidator()

	tests := []struct {
		name         string
		stopName     string
		expectNotice bool
		noticeLevel  string
	}{
		{"short name", "Central", false, ""},
		{"normal name", "Central Station Platform", false, ""},
		{"at recommended limit", strings.Repeat("A", 100), false, ""},
		{"over recommended limit", strings.Repeat("A", 150), true, "WARNING"},
		{"at maximum limit", strings.Repeat("A", 255), true, "WARNING"},
		{"over maximum limit", strings.Repeat("A", 300), true, "ERROR"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			container := notice.NewNoticeContainer()
			stop := &StopNameInfo{
				StopID:    "test_stop",
				StopName:  tt.stopName,
				RowNumber: 1,
			}

			validator.checkStopNameLength(container, stop)

			notices := container.GetNotices()
			hasLengthNotice := false
			for _, notice := range notices {
				if notice.Code() == "stop_name_too_long" {
					hasLengthNotice = true
					break
				}
			}

			if hasLengthNotice != tt.expectNotice {
				t.Errorf("Stop name length %d: expected notice=%v, got %v", len(tt.stopName), tt.expectNotice, hasLengthNotice)
			}
		})
	}
}

func TestStopNameValidator_CheckAllCapsName(t *testing.T) {
	validator := NewStopNameValidator()

	tests := []struct {
		name          string
		stopName      string
		expectAllCaps bool
	}{
		{"normal name", "Central Station", false},
		{"all caps long", "CENTRAL STATION", true},
		{"all caps short", "GPS", false}, // Abbreviations are OK
		{"mixed case", "Central STATION", false},
		{"numbers and caps", "42ND STREET", true},
		{"caps with punctuation", "O'HARE AIRPORT", true},
		{"single letter", "A", false}, // Too short
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			container := notice.NewNoticeContainer()
			stop := &StopNameInfo{
				StopID:    "test_stop",
				StopName:  tt.stopName,
				RowNumber: 1,
			}

			validator.checkAllCapsName(container, stop)

			notices := container.GetNotices()
			hasAllCapsNotice := false
			for _, notice := range notices {
				if notice.Code() == "stop_name_all_caps" {
					hasAllCapsNotice = true
					break
				}
			}

			if hasAllCapsNotice != tt.expectAllCaps {
				t.Errorf("Stop name '%s': expected all caps notice=%v, got %v", tt.stopName, tt.expectAllCaps, hasAllCapsNotice)
			}
		})
	}
}

func TestStopNameValidator_CheckRepeatedWords(t *testing.T) {
	validator := NewStopNameValidator()

	tests := []struct {
		name           string
		stopName       string
		expectRepeated bool
	}{
		{"normal name", "Central Station Platform", false},
		{"repeated word", "Station Station", true},
		{"repeated long word", "Platform Platform A", true},
		{"repeated short word", "A A Street", false}, // Too short to trigger
		{"case insensitive", "Main main Street", true},
		{"similar but not repeated", "Main Maine Street", false},
		{"single word", "Central", false},
		{"three repeated", "Stop Stop Stop", true}, // Should catch first repetition
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			container := notice.NewNoticeContainer()
			stop := &StopNameInfo{
				StopID:    "test_stop",
				StopName:  tt.stopName,
				RowNumber: 1,
			}

			validator.checkRepeatedWords(container, stop)

			notices := container.GetNotices()
			hasRepeatedNotice := false
			for _, notice := range notices {
				if notice.Code() == "stop_name_repeated_word" {
					hasRepeatedNotice = true
					break
				}
			}

			if hasRepeatedNotice != tt.expectRepeated {
				t.Errorf("Stop name '%s': expected repeated word notice=%v, got %v", tt.stopName, tt.expectRepeated, hasRepeatedNotice)
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
