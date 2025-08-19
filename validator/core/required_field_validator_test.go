package core

import (
	"github.com/theoremus-urban-solutions/gtfs-validator/testutil"
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

func TestRequiredFieldValidator_Validate(t *testing.T) {
	tests := []struct {
		name                string
		files               map[string]string
		expectedNoticeCodes []string
		description         string
	}{
		{
			name: "all required fields present",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,http://metro.example,America/Los_Angeles",
				"stops.txt":  "stop_id,stop_name,stop_lat,stop_lon\n1,Main St,34.05,-118.25",
			},
			expectedNoticeCodes: []string{},
			description:         "All required fields have values",
		},
		{
			name: "agency.txt missing required field values",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\n1,,http://metro.example,America/Los_Angeles", // Missing agency_name
			},
			expectedNoticeCodes: []string{"missing_required_field"},
			description:         "agency_name is empty",
		},
		{
			name: "multiple missing required field values",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\n1,,,", // Missing name, url, timezone
			},
			expectedNoticeCodes: []string{"missing_required_field", "missing_required_field", "missing_required_field"},
			description:         "Multiple required fields are empty",
		},
		{
			name: "stops.txt missing stop_id",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_lat,stop_lon\n,Main St,34.05,-118.25", // Missing stop_id
			},
			expectedNoticeCodes: []string{"missing_required_field"},
			description:         "stop_id is empty",
		},
		{
			name: "stops.txt missing stop_name for regular stop",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_lat,stop_lon,location_type\n1,,34.05,-118.25,0", // Missing stop_name for location_type 0
			},
			expectedNoticeCodes: []string{"missing_required_field"},
			description:         "stop_name is required for regular stops (location_type 0)",
		},
		{
			name: "stops.txt missing stop_name for generic node (location_type 3)",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_lat,stop_lon,location_type\n1,,34.05,-118.25,3", // Missing stop_name for location_type 3
			},
			expectedNoticeCodes: []string{"missing_recommended_field"},
			description:         "stop_name is optional for generic nodes (location_type 3) - generates warning",
		},
		{
			name: "stops.txt missing stop_name for boarding area with parent",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_lat,stop_lon,location_type,parent_station\n1,,34.05,-118.25,4,STATION1", // Missing stop_name but has parent
			},
			expectedNoticeCodes: []string{"missing_recommended_field"},
			description:         "stop_name is optional for boarding areas with parent station",
		},
		{
			name: "stops.txt missing stop_name for boarding area without parent",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_lat,stop_lon,location_type,parent_station\n1,,34.05,-118.25,4,", // Missing stop_name and no parent
			},
			expectedNoticeCodes: []string{"missing_required_field"},
			description:         "stop_name is required for boarding areas without parent station",
		},
		{
			name: "routes.txt missing required fields",
			files: map[string]string{
				"routes.txt": "route_id,agency_id,route_short_name,route_long_name,route_type\n,1,1,Main Line,", // Missing route_id and route_type
			},
			expectedNoticeCodes: []string{"missing_required_field", "missing_required_field"},
			description:         "route_id and route_type are required",
		},
		{
			name: "trips.txt missing required fields",
			files: map[string]string{
				"trips.txt": "route_id,service_id,trip_id,trip_headsign\n,,T1,Downtown", // Missing route_id and service_id
			},
			expectedNoticeCodes: []string{"missing_required_field", "missing_required_field"},
			description:         "route_id and service_id are required",
		},
		{
			name: "stop_times.txt missing required fields",
			files: map[string]string{
				"stop_times.txt": "trip_id,arrival_time,departure_time,stop_id,stop_sequence\nT1,08:00:00,08:00:00,,", // Missing stop_id and stop_sequence
			},
			expectedNoticeCodes: []string{"missing_required_field", "missing_required_field"},
			description:         "stop_id and stop_sequence are required",
		},
		{
			name: "calendar.txt missing weekday fields",
			files: map[string]string{
				"calendar.txt": "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\nS1,,,1,1,1,0,0,20250101,20251231", // Missing monday and tuesday
			},
			expectedNoticeCodes: []string{"missing_required_field", "missing_required_field"},
			description:         "All weekday fields are required",
		},
		{
			name: "calendar_dates.txt missing required fields",
			files: map[string]string{
				"calendar_dates.txt": "service_id,date,exception_type\nS1,,1", // Missing date
			},
			expectedNoticeCodes: []string{"missing_required_field"},
			description:         "date is required in calendar_dates.txt",
		},
		{
			name: "fare_attributes.txt missing required fields",
			files: map[string]string{
				"fare_attributes.txt": "fare_id,price,currency_type,payment_method,transfers\nF1,,USD,0,0", // Missing price
			},
			expectedNoticeCodes: []string{"missing_required_field"},
			description:         "price is required in fare_attributes.txt",
		},
		{
			name: "shapes.txt missing required fields",
			files: map[string]string{
				"shapes.txt": "shape_id,shape_pt_lat,shape_pt_lon,shape_pt_sequence\nS1,34.05,,1", // Missing shape_pt_lon
			},
			expectedNoticeCodes: []string{"missing_required_field"},
			description:         "shape_pt_lon is required",
		},
		{
			name: "frequencies.txt missing required fields",
			files: map[string]string{
				"frequencies.txt": "trip_id,start_time,end_time,headway_secs\nT1,08:00:00,,300", // Missing end_time
			},
			expectedNoticeCodes: []string{"missing_required_field"},
			description:         "end_time is required in frequencies.txt",
		},
		{
			name: "transfers.txt missing required fields",
			files: map[string]string{
				"transfers.txt": "from_stop_id,to_stop_id,transfer_type\n1,2,", // Missing transfer_type
			},
			expectedNoticeCodes: []string{"missing_required_field"},
			description:         "transfer_type is required",
		},
		{
			name: "pathways.txt missing required fields",
			files: map[string]string{
				"pathways.txt": "pathway_id,from_stop_id,to_stop_id,pathway_mode,is_bidirectional\nP1,1,2,,", // Missing pathway_mode and is_bidirectional
			},
			expectedNoticeCodes: []string{"missing_required_field", "missing_required_field"},
			description:         "pathway_mode and is_bidirectional are required",
		},
		{
			name: "levels.txt missing required fields",
			files: map[string]string{
				"levels.txt": "level_id,level_index\nL1,", // Missing level_index
			},
			expectedNoticeCodes: []string{"missing_required_field"},
			description:         "level_index is required",
		},
		{
			name: "feed_info.txt missing required fields",
			files: map[string]string{
				"feed_info.txt": "feed_publisher_name,feed_publisher_url,feed_lang\nMetro,,en", // Missing feed_publisher_url
			},
			expectedNoticeCodes: []string{"missing_required_field"},
			description:         "feed_publisher_url is required",
		},
		{
			name: "whitespace-only fields treated as empty",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\n1,   ,http://metro.example,America/Los_Angeles", // agency_name is just whitespace
			},
			expectedNoticeCodes: []string{"missing_required_field"},
			description:         "Fields with only whitespace should be treated as empty",
		},
		{
			name: "multiple rows with missing fields",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_lat,stop_lon\n1,Main St,34.05,-118.25\n,Second St,34.06,-118.26\n3,,34.07,-118.27", // Row 2 missing stop_id, Row 3 missing stop_name
			},
			expectedNoticeCodes: []string{"missing_required_field", "missing_required_field"},
			description:         "Multiple rows with different missing required fields",
		},
		{
			name: "files without required field definitions",
			files: map[string]string{
				"translations.txt": "table_name,field_name,language,translation\n,,es,Calle Principal", // Empty fields but no requirements defined
				"custom_file.txt":  "custom_field\n",                                                   // Empty field
			},
			expectedNoticeCodes: []string{},
			description:         "Files without defined required fields should not generate notices",
		},
		{
			name: "mixed valid and invalid rows",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,http://metro.example,America/Los_Angeles\n2,,http://bus.example,America/Los_Angeles", // Second row missing agency_name
			},
			expectedNoticeCodes: []string{"missing_required_field"},
			description:         "Mix of valid and invalid rows",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test components
			loader := testutil.CreateTestFeedLoader(t, tt.files)
			container := notice.NewNoticeContainer()
			validator := NewRequiredFieldValidator()
			config := gtfsvalidator.Config{}

			// Run validation
			validator.Validate(loader, container, config)

			// Get notices
			notices := container.GetNotices()

			// Check notice count
			if len(notices) != len(tt.expectedNoticeCodes) {
				t.Errorf("Expected %d notices, got %d for case: %s", len(tt.expectedNoticeCodes), len(notices), tt.description)
			}

			// Count notice codes
			expectedCodeCounts := make(map[string]int)
			for _, code := range tt.expectedNoticeCodes {
				expectedCodeCounts[code]++
			}

			actualCodeCounts := make(map[string]int)
			for _, notice := range notices {
				actualCodeCounts[notice.Code()]++
			}

			// Verify expected codes
			for expectedCode, expectedCount := range expectedCodeCounts {
				actualCount := actualCodeCounts[expectedCode]
				if actualCount != expectedCount {
					t.Errorf("Expected %d notices with code '%s', got %d", expectedCount, expectedCode, actualCount)
				}
			}

			// Check for unexpected notice codes
			for actualCode := range actualCodeCounts {
				if expectedCodeCounts[actualCode] == 0 {
					t.Errorf("Unexpected notice code: %s", actualCode)
				}
			}
		})
	}
}

func TestRequiredFieldValidator_GetRequiredFields(t *testing.T) {
	validator := NewRequiredFieldValidator()

	tests := []struct {
		filename       string
		expectedFields []string
		description    string
	}{
		{
			filename:       "agency.txt",
			expectedFields: []string{"agency_name", "agency_url", "agency_timezone"},
			description:    "Agency file required fields",
		},
		{
			filename:       "stops.txt",
			expectedFields: []string{"stop_id", "stop_name"},
			description:    "Stops file required fields",
		},
		{
			filename:       "routes.txt",
			expectedFields: []string{"route_id", "route_type"},
			description:    "Routes file required fields",
		},
		{
			filename:       "trips.txt",
			expectedFields: []string{"route_id", "service_id", "trip_id"},
			description:    "Trips file required fields",
		},
		{
			filename:       "stop_times.txt",
			expectedFields: []string{"trip_id", "stop_id", "stop_sequence"},
			description:    "Stop times file required fields",
		},
		{
			filename:       "calendar.txt",
			expectedFields: []string{"service_id", "monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday", "start_date", "end_date"},
			description:    "Calendar file required fields",
		},
		{
			filename:       "calendar_dates.txt",
			expectedFields: []string{"service_id", "date", "exception_type"},
			description:    "Calendar dates file required fields",
		},
		{
			filename:       "fare_attributes.txt",
			expectedFields: []string{"fare_id", "price", "currency_type"},
			description:    "Fare attributes file required fields",
		},
		{
			filename:       "shapes.txt",
			expectedFields: []string{"shape_id", "shape_pt_lat", "shape_pt_lon", "shape_pt_sequence"},
			description:    "Shapes file required fields",
		},
		{
			filename:       "frequencies.txt",
			expectedFields: []string{"trip_id", "start_time", "end_time", "headway_secs"},
			description:    "Frequencies file required fields",
		},
		{
			filename:       "transfers.txt",
			expectedFields: []string{"from_stop_id", "to_stop_id", "transfer_type"},
			description:    "Transfers file required fields",
		},
		{
			filename:       "pathways.txt",
			expectedFields: []string{"pathway_id", "from_stop_id", "to_stop_id", "pathway_mode", "is_bidirectional"},
			description:    "Pathways file required fields",
		},
		{
			filename:       "levels.txt",
			expectedFields: []string{"level_id", "level_index"},
			description:    "Levels file required fields",
		},
		{
			filename:       "feed_info.txt",
			expectedFields: []string{"feed_publisher_name", "feed_publisher_url", "feed_lang"},
			description:    "Feed info file required fields",
		},
		{
			filename:       "unknown_file.txt",
			expectedFields: []string{},
			description:    "Unknown files should have no required fields",
		},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			actualFields := validator.getRequiredFields(tt.filename)

			if len(actualFields) != len(tt.expectedFields) {
				t.Errorf("Expected %d required fields for %s, got %d", len(tt.expectedFields), tt.filename, len(actualFields))
			}

			// Create maps for easier comparison
			expectedMap := make(map[string]bool)
			for _, field := range tt.expectedFields {
				expectedMap[field] = true
			}

			actualMap := make(map[string]bool)
			for _, field := range actualFields {
				actualMap[field] = true
			}

			// Check all expected fields are present
			for _, expectedField := range tt.expectedFields {
				if !actualMap[expectedField] {
					t.Errorf("Expected required field '%s' for %s but didn't find it", expectedField, tt.filename)
				}
			}

			// Check for unexpected fields
			for _, actualField := range actualFields {
				if !expectedMap[actualField] {
					t.Errorf("Found unexpected required field '%s' for %s", actualField, tt.filename)
				}
			}
		})
	}
}

func TestRequiredFieldValidator_IsStopNameOptionalForLocationType(t *testing.T) {
	validator := NewRequiredFieldValidator()

	tests := []struct {
		name        string
		rowValues   map[string]string
		expected    bool
		description string
	}{
		{
			name:        "regular stop (location_type 0)",
			rowValues:   map[string]string{"location_type": "0"},
			expected:    false,
			description: "Regular stops require stop_name",
		},
		{
			name:        "station (location_type 1)",
			rowValues:   map[string]string{"location_type": "1"},
			expected:    false,
			description: "Stations require stop_name",
		},
		{
			name:        "entrance/exit (location_type 2)",
			rowValues:   map[string]string{"location_type": "2"},
			expected:    false,
			description: "Entrances/exits require stop_name",
		},
		{
			name:        "generic node (location_type 3)",
			rowValues:   map[string]string{"location_type": "3"},
			expected:    true,
			description: "Generic nodes don't require stop_name",
		},
		{
			name:        "boarding area with parent (location_type 4)",
			rowValues:   map[string]string{"location_type": "4", "parent_station": "STATION1"},
			expected:    true,
			description: "Boarding areas with parent stations don't require stop_name",
		},
		{
			name:        "boarding area without parent (location_type 4)",
			rowValues:   map[string]string{"location_type": "4", "parent_station": ""},
			expected:    false,
			description: "Boarding areas without parent stations require stop_name",
		},
		{
			name:        "boarding area no parent field (location_type 4)",
			rowValues:   map[string]string{"location_type": "4"},
			expected:    false,
			description: "Boarding areas without parent_station field require stop_name",
		},
		{
			name:        "no location_type field",
			rowValues:   map[string]string{},
			expected:    false,
			description: "Missing location_type defaults to 0, which requires stop_name",
		},
		{
			name:        "invalid location_type",
			rowValues:   map[string]string{"location_type": "invalid"},
			expected:    false,
			description: "Invalid location_type should require stop_name",
		},
		{
			name:        "boarding area with whitespace-only parent",
			rowValues:   map[string]string{"location_type": "4", "parent_station": "   "},
			expected:    false,
			description: "Boarding areas with whitespace-only parent require stop_name",
		},
		{
			name:        "future location type",
			rowValues:   map[string]string{"location_type": "5"},
			expected:    false,
			description: "Unknown location types should require stop_name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.isStopNameOptionalForLocationType(tt.rowValues)
			if result != tt.expected {
				t.Errorf("Expected %v for %s, got %v", tt.expected, tt.description, result)
			}
		})
	}
}

func TestRequiredFieldValidator_ValidateFile(t *testing.T) {
	tests := []struct {
		name            string
		filename        string
		content         string
		expectedNotices []string
		description     string
	}{
		{
			name:            "valid agency file",
			filename:        "agency.txt",
			content:         "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,http://metro.example,America/Los_Angeles",
			expectedNotices: []string{},
			description:     "All required fields present and filled",
		},
		{
			name:            "agency file with missing required field",
			filename:        "agency.txt",
			content:         "agency_id,agency_name,agency_url,agency_timezone\n1,,http://metro.example,America/Los_Angeles", // Missing agency_name
			expectedNotices: []string{"missing_required_field"},
			description:     "agency_name is empty",
		},
		{
			name:            "stops file with generic node",
			filename:        "stops.txt",
			content:         "stop_id,stop_name,stop_lat,stop_lon,location_type\n1,,34.05,-118.25,3", // Generic node without name
			expectedNotices: []string{"missing_recommended_field"},
			description:     "Generic node generates warning for missing stop_name",
		},
		{
			name:            "stops file with boarding area and parent",
			filename:        "stops.txt",
			content:         "stop_id,stop_name,stop_lat,stop_lon,location_type,parent_station\n1,,34.05,-118.25,4,STATION1", // Boarding area with parent
			expectedNotices: []string{"missing_recommended_field"},
			description:     "Boarding area with parent generates warning",
		},
		{
			name:            "file without required field definitions",
			filename:        "translations.txt",
			content:         "table_name,field_name,language,translation\n,,es,", // Empty fields
			expectedNotices: []string{},
			description:     "Files without defined requirements generate no notices",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test components
			files := map[string]string{tt.filename: tt.content}
			loader := testutil.CreateTestFeedLoader(t, files)
			container := notice.NewNoticeContainer()
			validator := NewRequiredFieldValidator()

			// Run validation on specific file
			validator.validateFile(loader, container, tt.filename)

			// Get notices
			notices := container.GetNotices()

			// Count notice codes
			expectedCodeCounts := make(map[string]int)
			for _, code := range tt.expectedNotices {
				expectedCodeCounts[code]++
			}

			actualCodeCounts := make(map[string]int)
			for _, notice := range notices {
				actualCodeCounts[notice.Code()]++
			}

			// Verify expected codes
			for expectedCode, expectedCount := range expectedCodeCounts {
				actualCount := actualCodeCounts[expectedCode]
				if actualCount != expectedCount {
					t.Errorf("Expected %d notices with code '%s', got %d for %s", expectedCount, expectedCode, actualCount, tt.description)
				}
			}

			// Check for unexpected notice codes
			for actualCode := range actualCodeCounts {
				if expectedCodeCounts[actualCode] == 0 {
					t.Errorf("Unexpected notice code: %s for %s", actualCode, tt.description)
				}
			}
		})
	}
}

func TestRequiredFieldValidator_New(t *testing.T) {
	validator := NewRequiredFieldValidator()
	if validator == nil {
		t.Error("NewRequiredFieldValidator() returned nil")
	}
}
