package core

import (
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/testutil"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

func TestMissingColumnValidator_Validate(t *testing.T) {
	tests := []struct {
		name                string
		files               map[string]string
		expectedNoticeCodes []string
		description         string
	}{
		{
			name: "all required columns present",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,http://metro.example,America/Los_Angeles",
				"stops.txt":  "stop_id,stop_name,stop_lat,stop_lon\n1,Main St,34.05,-118.25",
				"routes.txt": "route_id,agency_id,route_short_name,route_long_name,route_type\n1,1,1,Main Line,3",
			},
			expectedNoticeCodes: []string{},
			description:         "All files have their required columns",
		},
		{
			name: "agency.txt missing required columns",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name\n1,Metro", // Missing agency_url, agency_timezone
			},
			expectedNoticeCodes: []string{"missing_required_column", "missing_required_column"},
			description:         "agency.txt missing agency_url and agency_timezone",
		},
		{
			name: "stops.txt missing stop_id",
			files: map[string]string{
				"stops.txt": "stop_name,stop_lat,stop_lon\nMain St,34.05,-118.25", // Missing stop_id
			},
			expectedNoticeCodes: []string{"missing_required_column"},
			description:         "stops.txt missing required stop_id column",
		},
		{
			name: "routes.txt missing multiple required columns",
			files: map[string]string{
				"routes.txt": "agency_id,route_short_name,route_long_name\n1,1,Main Line", // Missing route_id, route_type
			},
			expectedNoticeCodes: []string{"missing_required_column", "missing_required_column"},
			description:         "routes.txt missing route_id and route_type",
		},
		{
			name: "trips.txt missing all required columns",
			files: map[string]string{
				"trips.txt": "trip_headsign,direction_id\nDowntown,0", // Missing route_id, service_id, trip_id
			},
			expectedNoticeCodes: []string{"missing_required_column", "missing_required_column", "missing_required_column"},
			description:         "trips.txt missing all required columns",
		},
		{
			name: "stop_times.txt missing some required columns",
			files: map[string]string{
				"stop_times.txt": "trip_id,arrival_time,departure_time,stop_id\nT1,08:00:00,08:00:00,1", // Missing stop_sequence
			},
			expectedNoticeCodes: []string{"missing_required_column"},
			description:         "stop_times.txt missing stop_sequence",
		},
		{
			name: "calendar.txt missing weekday columns",
			files: map[string]string{
				"calendar.txt": "service_id,start_date,end_date\nS1,20250101,20251231", // Missing all weekday columns
			},
			expectedNoticeCodes: []string{"missing_required_column", "missing_required_column", "missing_required_column", "missing_required_column", "missing_required_column", "missing_required_column", "missing_required_column"},
			description:         "calendar.txt missing all weekday columns",
		},
		{
			name: "calendar_dates.txt missing columns",
			files: map[string]string{
				"calendar_dates.txt": "service_id,date\nS1,20250101", // Missing exception_type
			},
			expectedNoticeCodes: []string{"missing_required_column"},
			description:         "calendar_dates.txt missing exception_type",
		},
		{
			name: "fare_attributes.txt missing columns",
			files: map[string]string{
				"fare_attributes.txt": "fare_id,price\nF1,2.50", // Missing currency_type
			},
			expectedNoticeCodes: []string{"missing_required_column"},
			description:         "fare_attributes.txt missing currency_type",
		},
		{
			name: "shapes.txt missing columns",
			files: map[string]string{
				"shapes.txt": "shape_id,shape_pt_lat,shape_pt_lon\nS1,34.05,-118.25", // Missing shape_pt_sequence
			},
			expectedNoticeCodes: []string{"missing_required_column"},
			description:         "shapes.txt missing shape_pt_sequence",
		},
		{
			name: "frequencies.txt missing columns",
			files: map[string]string{
				"frequencies.txt": "trip_id,start_time,end_time\nT1,08:00:00,10:00:00", // Missing headway_secs
			},
			expectedNoticeCodes: []string{"missing_required_column"},
			description:         "frequencies.txt missing headway_secs",
		},
		{
			name: "transfers.txt missing columns",
			files: map[string]string{
				"transfers.txt": "from_stop_id,to_stop_id\n1,2", // Missing transfer_type
			},
			expectedNoticeCodes: []string{"missing_required_column"},
			description:         "transfers.txt missing transfer_type",
		},
		{
			name: "pathways.txt missing columns",
			files: map[string]string{
				"pathways.txt": "pathway_id,from_stop_id,to_stop_id\nP1,1,2", // Missing pathway_mode, is_bidirectional
			},
			expectedNoticeCodes: []string{"missing_required_column", "missing_required_column"},
			description:         "pathways.txt missing pathway_mode and is_bidirectional",
		},
		{
			name: "levels.txt missing columns",
			files: map[string]string{
				"levels.txt": "level_id\nL1", // Missing level_index
			},
			expectedNoticeCodes: []string{"missing_required_column"},
			description:         "levels.txt missing level_index",
		},
		{
			name: "feed_info.txt missing columns",
			files: map[string]string{
				"feed_info.txt": "feed_publisher_name,feed_publisher_url\nMetro,http://metro.example", // Missing feed_lang
			},
			expectedNoticeCodes: []string{"missing_required_column"},
			description:         "feed_info.txt missing feed_lang",
		},
		{
			name: "files without required column definitions",
			files: map[string]string{
				"translations.txt": "table_name,field_name,language,translation\nstops,stop_name,es,Calle Principal",
				"attributions.txt": "attribution_id,organization_name\nA1,Metro",
				"custom_file.txt":  "custom_field\nvalue",
			},
			expectedNoticeCodes: []string{},
			description:         "Files without defined required columns should not generate notices",
		},
		{
			name: "mixed valid and invalid files",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,http://metro.example,America/Los_Angeles",
				"stops.txt":  "stop_name,stop_lat,stop_lon\nMain St,34.05,-118.25", // Missing stop_id
				"routes.txt": "route_id,agency_id,route_short_name,route_long_name,route_type\n1,1,1,Main Line,3",
			},
			expectedNoticeCodes: []string{"missing_required_column"},
			description:         "Mix of valid and invalid files",
		},
		{
			name: "columns with different case",
			files: map[string]string{
				"agency.txt": "Agency_ID,Agency_Name,Agency_URL,Agency_Timezone\n1,Metro,http://metro.example,America/Los_Angeles", // Wrong case
			},
			expectedNoticeCodes: []string{"missing_required_column", "missing_required_column", "missing_required_column"},
			description:         "Column names are case-sensitive",
		},
		{
			name: "columns with extra whitespace",
			files: map[string]string{
				"agency.txt": " agency_name , agency_url , agency_timezone \n1,Metro,http://metro.example,America/Los_Angeles", // Extra whitespace
			},
			expectedNoticeCodes: []string{"missing_required_column", "missing_required_column", "missing_required_column"},
			description:         "Column names with whitespace should not match required columns",
		},
		{
			name: "additional optional columns present",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone,agency_lang,agency_phone\n1,Metro,http://metro.example,America/Los_Angeles,en,555-1234",
			},
			expectedNoticeCodes: []string{},
			description:         "Additional optional columns should not cause issues",
		},
		{
			name: "fare_rules.txt with only required column",
			files: map[string]string{
				"fare_rules.txt": "fare_id\nF1", // Only has fare_id (which is required)
			},
			expectedNoticeCodes: []string{},
			description:         "fare_rules.txt only requires fare_id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test components
			loader := testutil.CreateTestFeedLoader(t, tt.files)
			container := notice.NewNoticeContainer()
			validator := NewMissingColumnValidator()
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

func TestMissingColumnValidator_ValidateFileColumns(t *testing.T) {
	tests := []struct {
		name            string
		filename        string
		content         string
		expectedMissing []string
		description     string
	}{
		{
			name:            "agency.txt all columns present",
			filename:        "agency.txt",
			content:         "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,http://metro.example,America/Los_Angeles",
			expectedMissing: []string{},
			description:     "All required columns present",
		},
		{
			name:            "agency.txt missing single column",
			filename:        "agency.txt",
			content:         "agency_id,agency_name,agency_url\n1,Metro,http://metro.example", // Missing agency_timezone
			expectedMissing: []string{"agency_timezone"},
			description:     "Missing agency_timezone",
		},
		{
			name:            "agency.txt missing multiple columns",
			filename:        "agency.txt",
			content:         "agency_id,agency_name\n1,Metro", // Missing agency_url, agency_timezone
			expectedMissing: []string{"agency_url", "agency_timezone"},
			description:     "Missing agency_url and agency_timezone",
		},
		{
			name:            "stops.txt valid",
			filename:        "stops.txt",
			content:         "stop_id,stop_name,stop_lat,stop_lon\n1,Main St,34.05,-118.25",
			expectedMissing: []string{},
			description:     "Only stop_id is required for stops.txt",
		},
		{
			name:            "stops.txt missing required column",
			filename:        "stops.txt",
			content:         "stop_name,stop_lat,stop_lon\nMain St,34.05,-118.25", // Missing stop_id
			expectedMissing: []string{"stop_id"},
			description:     "Missing required stop_id",
		},
		{
			name:            "calendar.txt all weekdays present",
			filename:        "calendar.txt",
			content:         "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\nS1,1,1,1,1,1,0,0,20250101,20251231",
			expectedMissing: []string{},
			description:     "All required columns including all weekdays",
		},
		{
			name:            "calendar.txt missing weekdays",
			filename:        "calendar.txt",
			content:         "service_id,monday,tuesday,start_date,end_date\nS1,1,1,20250101,20251231", // Missing wed-sun
			expectedMissing: []string{"wednesday", "thursday", "friday", "saturday", "sunday"},
			description:     "Missing multiple weekday columns",
		},
		{
			name:            "file without required columns defined",
			filename:        "translations.txt",
			content:         "table_name,field_name\nstops,stop_name", // No required columns defined for this file
			expectedMissing: []string{},
			description:     "Files without defined requirements should not generate notices",
		},
		{
			name:            "empty file",
			filename:        "agency.txt",
			content:         "", // Empty file - will cause CSV parsing error
			expectedMissing: []string{},
			description:     "Empty files should be handled gracefully (CSV parser will fail)",
		},
		{
			name:            "headers only file",
			filename:        "routes.txt",
			content:         "route_id,agency_id,route_short_name,route_long_name,route_type",
			expectedMissing: []string{},
			description:     "File with all required headers but no data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test components
			files := map[string]string{tt.filename: tt.content}
			loader := testutil.CreateTestFeedLoader(t, files)
			container := notice.NewNoticeContainer()
			validator := NewMissingColumnValidator()

			// Run validation on specific file
			validator.validateFileColumns(loader, container, tt.filename)

			// Get notices
			notices := container.GetNotices()

			// Check notice count
			if len(notices) != len(tt.expectedMissing) {
				t.Errorf("Expected %d missing column notices, got %d for %s", len(tt.expectedMissing), len(notices), tt.description)
			}

			// Verify specific missing columns
			foundMissing := make(map[string]bool)
			for _, notice := range notices {
				if notice.Code() == "missing_required_column" {
					context := notice.Context()
					if columnName, ok := context["columnName"]; ok {
						foundMissing[columnName.(string)] = true
						// Also verify filename in context
						if filename, ok := context["filename"]; !ok || filename != tt.filename {
							t.Errorf("Expected filename '%s' in context, got '%v'", tt.filename, filename)
						}
					}
				}
			}

			// Check that all expected missing columns were found
			for _, expectedColumn := range tt.expectedMissing {
				if !foundMissing[expectedColumn] {
					t.Errorf("Expected missing column notice for '%s' but didn't find it", expectedColumn)
				}
			}

			// Check for unexpected missing columns
			for foundColumn := range foundMissing {
				expected := false
				for _, expectedColumn := range tt.expectedMissing {
					if foundColumn == expectedColumn {
						expected = true
						break
					}
				}
				if !expected {
					t.Errorf("Found unexpected missing column notice for '%s'", foundColumn)
				}
			}
		})
	}
}

func TestMissingColumnValidator_New(t *testing.T) {
	validator := NewMissingColumnValidator()
	if validator == nil {
		t.Error("NewMissingColumnValidator() returned nil")
	}
}

func TestMissingColumnValidator_FileRequiredColumns(t *testing.T) {
	// Test that the fileRequiredColumns map contains expected files and columns
	expectedFileColumns := map[string][]string{
		"agency.txt":          {"agency_name", "agency_url", "agency_timezone"},
		"stops.txt":           {"stop_id"},
		"routes.txt":          {"route_id", "route_type"},
		"trips.txt":           {"route_id", "service_id", "trip_id"},
		"stop_times.txt":      {"trip_id", "stop_id", "stop_sequence"},
		"calendar.txt":        {"service_id", "monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday", "start_date", "end_date"},
		"calendar_dates.txt":  {"service_id", "date", "exception_type"},
		"fare_attributes.txt": {"fare_id", "price", "currency_type"},
		"fare_rules.txt":      {"fare_id"},
		"shapes.txt":          {"shape_id", "shape_pt_lat", "shape_pt_lon", "shape_pt_sequence"},
		"frequencies.txt":     {"trip_id", "start_time", "end_time", "headway_secs"},
		"transfers.txt":       {"from_stop_id", "to_stop_id", "transfer_type"},
		"pathways.txt":        {"pathway_id", "from_stop_id", "to_stop_id", "pathway_mode", "is_bidirectional"},
		"levels.txt":          {"level_id", "level_index"},
		"feed_info.txt":       {"feed_publisher_name", "feed_publisher_url", "feed_lang"},
	}

	for filename, expectedColumns := range expectedFileColumns {
		actualColumns, exists := fileRequiredColumns[filename]
		if !exists {
			t.Errorf("Expected file '%s' to be in fileRequiredColumns map", filename)
			continue
		}

		if len(actualColumns) != len(expectedColumns) {
			t.Errorf("File '%s': expected %d required columns, got %d", filename, len(expectedColumns), len(actualColumns))
			continue
		}

		// Create maps for easier comparison
		expectedMap := make(map[string]bool)
		for _, col := range expectedColumns {
			expectedMap[col] = true
		}

		actualMap := make(map[string]bool)
		for _, col := range actualColumns {
			actualMap[col] = true
		}

		// Check that all expected columns are present
		for _, expectedCol := range expectedColumns {
			if !actualMap[expectedCol] {
				t.Errorf("File '%s': expected required column '%s' not found", filename, expectedCol)
			}
		}

		// Check for unexpected columns
		for _, actualCol := range actualColumns {
			if !expectedMap[actualCol] {
				t.Errorf("File '%s': unexpected required column '%s'", filename, actualCol)
			}
		}
	}
}

func TestMissingColumnValidator_FileNotExists(t *testing.T) {
	// Test behavior when file doesn't exist
	loader := testutil.CreateTestFeedLoader(t, map[string]string{}) // No files
	container := notice.NewNoticeContainer()
	validator := NewMissingColumnValidator()

	// Try to validate a non-existent file
	validator.validateFileColumns(loader, container, "agency.txt")

	// Should not generate any notices (other validators handle missing files)
	notices := container.GetNotices()
	if len(notices) != 0 {
		t.Errorf("Expected no notices for non-existent file, got %d", len(notices))
	}
}

func TestMissingColumnValidator_MalformedCSV(t *testing.T) {
	// Test behavior with malformed CSV that prevents header parsing
	// Use malformed CSV that will cause the header parsing itself to fail
	files := map[string]string{
		"agency.txt": "", // Empty file - headers cannot be parsed
	}
	loader := testutil.CreateTestFeedLoader(t, files)
	container := notice.NewNoticeContainer()
	validator := NewMissingColumnValidator()

	validator.validateFileColumns(loader, container, "agency.txt")

	// Should not generate missing column notices for files that can't be parsed
	notices := container.GetNotices()
	for _, notice := range notices {
		if notice.Code() == "missing_required_column" {
			t.Error("Should not generate missing_required_column notice when CSV headers cannot be parsed")
		}
	}
}
