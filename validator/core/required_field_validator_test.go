package core

import (
	"slices"
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/testutil"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/schema"
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
				"stops.txt": "stop_id,stop_name,stop_lat,stop_lon,location_type\n1,,34.05,-118.25,0",
			},
			expectedNoticeCodes: []string{},
			description:         "stop_name is conditionally required, and reported as missing_stop_name elsewhere",
		},
		{
			name: "stops.txt missing stop_name for generic node (location_type 3)",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_lat,stop_lon,location_type\n1,,34.05,-118.25,3", // Missing stop_name for location_type 3
			},
			expectedNoticeCodes: []string{},
			description:         "stop_name is optional for generic nodes (location_type 3) - generates warning",
		},
		{
			name: "stops.txt missing stop_name for boarding area with parent",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_lat,stop_lon,location_type,parent_station\n1,,34.05,-118.25,4,STATION1", // Missing stop_name but has parent
			},
			expectedNoticeCodes: []string{},
			description:         "stop_name is optional for boarding areas with parent station",
		},
		{
			name: "stops.txt missing stop_name for boarding area without parent",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_lat,stop_lon,location_type,parent_station\n1,,34.05,-118.25,4,",
			},
			expectedNoticeCodes: []string{},
			description:         "stop_name is conditionally required, and reported as missing_stop_name elsewhere",
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
				"stop_times.txt": "trip_id,arrival_time,departure_time,stop_id,stop_sequence\nT1,08:00:00,08:00:00,,",
			},
			// stop_sequence is Required; stop_id is only Conditionally Required,
			// because a Flex row may name a location group or GeoJSON location
			// instead. That condition belongs to a rule that knows it.
			expectedNoticeCodes: []string{"missing_required_field"},
			description:         "stop_sequence is required outright, stop_id conditionally",
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
				"transfers.txt": "from_stop_id,to_stop_id,transfer_type\n1,2,",
			},
			// transfer_type is Required, but its value list offers empty as a
			// choice meaning a recommended transfer point, so a blank here is a
			// value rather than an omission.
			expectedNoticeCodes: []string{},
			description:         "an empty transfer_type is a value, not a missing field",
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
				"feed_info.txt": "feed_publisher_name,feed_publisher_url,feed_lang\nMetro,,en",
			},
			// feed_info.txt carries the only three Recommended fields in the
			// spec, and this row omits all of them alongside the required URL.
			expectedNoticeCodes: []string{
				"missing_required_field",
				"missing_recommended_field", "missing_recommended_field", "missing_recommended_field",
			},
			description: "feed_publisher_url is required; the three feed dates and version are recommended",
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
				"stops.txt": "stop_id,stop_name,stop_lat,stop_lon\n1,Main St,34.05,-118.25\n,Second St,34.06,-118.26\n3,,34.07,-118.27",
			},
			// Only the missing stop_id is this validator's: an absent stop_name
			// is conditionally required and belongs to missing_stop_name.
			expectedNoticeCodes: []string{"missing_required_field"},
			description:         "Multiple rows with different missing required fields",
		},
		{
			name: "file outside the spec",
			files: map[string]string{
				"custom_file.txt": "custom_field\n",
			},
			// The spec table answers for every file it describes; anything else
			// has no required fields to check.
			expectedNoticeCodes: []string{},
			description:         "A file the spec does not describe generates no notices",
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

// The required and recommended field lists now come from the generated spec
// table, so what is worth testing is that the table is wired in correctly and
// that the two presence classes stay distinct — not the contents of a list this
// package no longer owns.
func TestRequiredFieldValidator_UsesSpecPresence(t *testing.T) {
	required, known := schema.FieldsWithPresence("agency.txt", schema.PresenceRequired)
	if !known {
		t.Fatal("agency.txt should be described by the spec table")
	}
	if !slices.Contains(required, "agency_name") {
		t.Errorf("agency_name is Required in the spec, got %v", required)
	}
	// Conditionally Required belongs to the rule that knows the condition, so
	// it must not leak into the generic required list.
	if slices.Contains(required, "agency_id") {
		t.Error("agency_id is Conditionally Required and must not be treated as Required")
	}

	// feed_info.txt holds the only Recommended fields in the whole spec.
	recommended, _ := schema.FieldsWithPresence("feed_info.txt", schema.PresenceRecommended)
	if !slices.Contains(recommended, "feed_version") {
		t.Errorf("feed_version is Recommended in the spec, got %v", recommended)
	}

	if _, known := schema.FieldsWithPresence("not_a_gtfs_file.txt", schema.PresenceRequired); known {
		t.Error("a file the spec does not describe should not be known")
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
			expectedNotices: []string{},
			description:     "Generic node generates warning for missing stop_name",
		},
		{
			name:            "stops file with boarding area and parent",
			filename:        "stops.txt",
			content:         "stop_id,stop_name,stop_lat,stop_lon,location_type,parent_station\n1,,34.05,-118.25,4,STATION1", // Boarding area with parent
			expectedNotices: []string{},
			description:     "Boarding area with parent generates warning",
		},
		{
			name:            "file the spec does not describe",
			filename:        "custom_extension.txt",
			content:         "some_field,another\n,",
			expectedNotices: []string{},
			description:     "A file outside the spec has no required fields to check",
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
