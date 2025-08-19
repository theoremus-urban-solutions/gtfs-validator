package core

import (
	"strings"
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/testutil"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

func TestUnknownFileValidator_Validate(t *testing.T) {
	tests := []struct {
		name                string
		files               map[string]string
		expectedNoticeCodes []string
		description         string
	}{
		{
			name: "all known GTFS files",
			files: map[string]string{
				AgencyFile:           "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,http://metro.example,America/Los_Angeles",
				"stops.txt":          "stop_id,stop_name,stop_lat,stop_lon\n1,Main St,34.05,-118.25",
				"routes.txt":         "route_id,agency_id,route_short_name,route_long_name,route_type\n1,1,1,Main Line,3",
				"trips.txt":          "route_id,service_id,trip_id\n1,S1,T1",
				"stop_times.txt":     "trip_id,arrival_time,departure_time,stop_id,stop_sequence\nT1,08:00:00,08:00:00,1,1",
				"calendar.txt":       "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\nS1,1,1,1,1,1,0,0,20250101,20251231",
				"calendar_dates.txt": "service_id,date,exception_type\nS1,20250101,1",
				"feed_info.txt":      "feed_publisher_name,feed_publisher_url,feed_lang\nMetro,http://metro.example,en",
			},
			expectedNoticeCodes: []string{},
			description:         "All files are recognized GTFS files",
		},
		{
			name: "single unknown file",
			files: map[string]string{
				AgencyFile:        "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,http://metro.example,America/Los_Angeles",
				"custom_data.txt": "id,name,value\n1,test,123",
			},
			expectedNoticeCodes: []string{"unknown_file"},
			description:         "One unknown file in feed",
		},
		{
			name: "multiple unknown files",
			files: map[string]string{
				AgencyFile:        "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,http://metro.example,America/Los_Angeles",
				"custom_data.txt": "id,name,value\n1,test,123",
				"extra_info.txt":  "field1,field2\na,b",
				"vendor_data.txt": "vendor_field\nvalue",
			},
			expectedNoticeCodes: []string{"unknown_file", "unknown_file", "unknown_file"},
			description:         "Multiple unknown files in feed",
		},
		{
			name: "unknown geojson file",
			files: map[string]string{
				AgencyFile:              "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,http://metro.example,America/Los_Angeles",
				"custom_shapes.geojson": `{"type": "FeatureCollection", "features": []}`,
			},
			expectedNoticeCodes: []string{"unknown_file"},
			description:         "Unknown GeoJSON file",
		},
		{
			name: "known geojson file",
			files: map[string]string{
				AgencyFile:          "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,http://metro.example,America/Los_Angeles",
				"locations.geojson": `{"type": "FeatureCollection", "features": []}`,
			},
			expectedNoticeCodes: []string{},
			description:         "Known GeoJSON file (locations.geojson)",
		},
		{
			name: "non-text files ignored",
			files: map[string]string{
				AgencyFile:  "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,http://metro.example,America/Los_Angeles",
				"readme.md": "# GTFS Feed\nThis is a sample feed",
				"data.json": `{"version": "1.0"}`,
				"image.png": "binary image data",
				"style.css": "body { margin: 0; }",
			},
			expectedNoticeCodes: []string{},
			description:         "Non-text files should be ignored",
		},
		{
			name: "hidden files ignored",
			files: map[string]string{
				AgencyFile:    "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,http://metro.example,America/Los_Angeles",
				".hidden.txt": "hidden content",
				".DS_Store":   "mac system file",
				".gitignore":  "*.tmp",
			},
			expectedNoticeCodes: []string{},
			description:         "Hidden files (starting with .) should be ignored",
		},
		{
			name: "mixed known and unknown files",
			files: map[string]string{
				AgencyFile:            "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,http://metro.example,America/Los_Angeles",
				"stops.txt":           "stop_id,stop_name,stop_lat,stop_lon\n1,Main St,34.05,-118.25",
				"fare_attributes.txt": "fare_id,price,currency_type,payment_method,transfers\nF1,2.50,USD,0,0",
				"pathways.txt":        "pathway_id,from_stop_id,to_stop_id,pathway_mode\nP1,1,2,1",
				"custom_routes.txt":   "custom_field\nvalue",
				"vendor_stops.txt":    "vendor_field\nvalue",
			},
			expectedNoticeCodes: []string{"unknown_file", "unknown_file"},
			description:         "Mix of known GTFS files and unknown files",
		},
		{
			name: "all GTFS fare v2 files",
			files: map[string]string{
				AgencyFile:                "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,http://metro.example,America/Los_Angeles",
				"fare_media.txt":          "fare_media_id,fare_media_name,fare_media_type\nM1,Card,1",
				"fare_products.txt":       "fare_product_id,fare_product_name,amount,currency\nP1,Regular,2.50,USD",
				"fare_leg_rules.txt":      "leg_group_id,network_id,from_area_id,to_area_id,fare_product_id\nLG1,N1,A1,A2,P1",
				"fare_transfer_rules.txt": "from_leg_group_id,to_leg_group_id,fare_product_id,transfer_count,duration_limit\nLG1,LG2,P1,1,7200",
			},
			expectedNoticeCodes: []string{},
			description:         "All GTFS Fare v2 files are recognized",
		},
		{
			name: "GTFS flex files",
			files: map[string]string{
				AgencyFile:                 "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,http://metro.example,America/Los_Angeles",
				"areas.txt":                "area_id,area_name\nA1,Downtown",
				"stop_areas.txt":           "area_id,stop_id\nA1,S1",
				"booking_rules.txt":        "booking_rule_id,booking_type,prior_notice_duration_min\nBR1,1,60",
				"location_groups.txt":      "location_group_id,location_group_name\nLG1,City Center",
				"location_group_stops.txt": "location_group_id,stop_id\nLG1,S1",
			},
			expectedNoticeCodes: []string{},
			description:         "GTFS Flex files are recognized",
		},
		{
			name: "network files",
			files: map[string]string{
				AgencyFile:           "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,http://metro.example,America/Los_Angeles",
				"networks.txt":       "network_id,network_name\nN1,Metro Network",
				"route_networks.txt": "network_id,route_id\nN1,R1",
			},
			expectedNoticeCodes: []string{},
			description:         "Network files are recognized",
		},
		{
			name: "shapes geojson file",
			files: map[string]string{
				AgencyFile:           "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,http://metro.example,America/Los_Angeles",
				"shapes_geojson.txt": `{"type": "FeatureCollection", "features": []}`,
			},
			expectedNoticeCodes: []string{},
			description:         "shapes_geojson.txt is recognized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test components
			loader := testutil.CreateTestFeedLoader(t, tt.files)
			container := notice.NewNoticeContainer()
			validator := NewUnknownFileValidator()
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

func TestUnknownFileValidator_ValidateKnownFile(t *testing.T) {
	tests := []struct {
		name                string
		filename            string
		expectUnknownNotice bool
		description         string
	}{
		{
			name:                "known core file",
			filename:            AgencyFile,
			expectUnknownNotice: false,
			description:         AgencyFile + " is a core GTFS file",
		},
		{
			name:                "known optional file",
			filename:            "feed_info.txt",
			expectUnknownNotice: false,
			description:         "feed_info.txt is an optional GTFS file",
		},
		{
			name:                "unknown txt file",
			filename:            "custom_data.txt",
			expectUnknownNotice: true,
			description:         "custom_data.txt is not a recognized GTFS file",
		},
		{
			name:                "known geojson file",
			filename:            "locations.geojson",
			expectUnknownNotice: false,
			description:         "locations.geojson is a recognized GTFS file",
		},
		{
			name:                "unknown geojson file",
			filename:            "custom_shapes.geojson",
			expectUnknownNotice: true,
			description:         "custom_shapes.geojson is not recognized",
		},
		{
			name:                "non-text file ignored",
			filename:            "readme.md",
			expectUnknownNotice: false,
			description:         "Non-text files should be ignored",
		},
		{
			name:                "hidden file ignored",
			filename:            ".hidden.txt",
			expectUnknownNotice: false,
			description:         "Hidden files should be ignored",
		},
		{
			name:                "system file ignored",
			filename:            ".DS_Store",
			expectUnknownNotice: false,
			description:         "System files should be ignored",
		},
		{
			name:                "file with path",
			filename:            "data/custom_data.txt",
			expectUnknownNotice: true,
			description:         "Files with paths should use base filename for validation",
		},
		{
			name:                "known file with path",
			filename:            "gtfs/agency.txt",
			expectUnknownNotice: false,
			description:         "Known files with paths should be recognized",
		},
		{
			name:                "fare v2 file",
			filename:            "fare_media.txt",
			expectUnknownNotice: false,
			description:         "GTFS Fare v2 files should be recognized",
		},
		{
			name:                "flex file",
			filename:            "areas.txt",
			expectUnknownNotice: false,
			description:         "GTFS Flex files should be recognized",
		},
		{
			name:                "network file",
			filename:            "networks.txt",
			expectUnknownNotice: false,
			description:         "Network files should be recognized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			container := notice.NewNoticeContainer()
			validator := NewUnknownFileValidator()

			// Run validation on specific file
			validator.validateKnownFile(container, tt.filename)

			// Get notices
			notices := container.GetNotices()

			// Check if unknown file notice was generated as expected
			hasUnknownNotice := false
			for _, notice := range notices {
				if notice.Code() == "unknown_file" {
					hasUnknownNotice = true
					// Verify context contains correct filename (base filename)
					context := notice.Context()
					if filename, ok := context["filename"]; ok {
						expectedFilename := tt.filename
						// If the input had a path, the notice should contain the base filename
						switch tt.filename {
						case "data/custom_data.txt":
							expectedFilename = "custom_data.txt"
						case "gtfs/agency.txt":
							expectedFilename = AgencyFile
						}
						if filename != expectedFilename {
							t.Errorf("Expected filename '%s' in notice context, got '%v'", expectedFilename, filename)
						}
					} else {
						t.Error("Expected filename in notice context")
					}
				}
			}

			if hasUnknownNotice != tt.expectUnknownNotice {
				t.Errorf("Expected unknown notice: %v, got unknown notice: %v for %s", tt.expectUnknownNotice, hasUnknownNotice, tt.description)
			}
		})
	}
}

func TestUnknownFileValidator_New(t *testing.T) {
	validator := NewUnknownFileValidator()
	if validator == nil {
		t.Error("NewUnknownFileValidator() returned nil")
	}
}

func TestUnknownFileValidator_KnownGTFSFiles(t *testing.T) {
	// Test that all expected GTFS files are in the known files list
	expectedFiles := []string{
		// Core files
		"agency.txt",
		"stops.txt",
		"routes.txt",
		"trips.txt",
		"stop_times.txt",

		// Calendar files
		"calendar.txt",
		"calendar_dates.txt",

		// Optional files
		"fare_attributes.txt",
		"fare_rules.txt",
		"shapes.txt",
		"frequencies.txt",
		"transfers.txt",
		"pathways.txt",
		"levels.txt",
		"feed_info.txt",
		"translations.txt",
		"attributions.txt",

		// GTFS Fare v2
		"fare_media.txt",
		"fare_products.txt",
		"fare_leg_rules.txt",
		"fare_transfer_rules.txt",

		// GTFS Flex
		"areas.txt",
		"stop_areas.txt",
		"booking_rules.txt",
		"location_groups.txt",
		"location_group_stops.txt",

		// Networks
		"networks.txt",
		"route_networks.txt",

		// GeoJSON
		"shapes_geojson.txt",
		"locations.geojson",
	}

	for _, filename := range expectedFiles {
		if !knownGTFSFiles[filename] {
			t.Errorf("Expected file '%s' to be in knownGTFSFiles map", filename)
		}
	}
}

func TestUnknownFileValidator_FileExtensions(t *testing.T) {
	tests := []struct {
		filename string
		expected bool
	}{
		{"agency.txt", true},        // Standard GTFS file
		{"locations.geojson", true}, // GeoJSON file
		{"readme.md", false},        // Markdown file (ignored)
		{"data.json", false},        // JSON file (ignored)
		{"image.png", false},        // Image file (ignored)
		{"script.js", false},        // JavaScript file (ignored)
		{"style.css", false},        // CSS file (ignored)
		{"document.pdf", false},     // PDF file (ignored)
		{".hidden.txt", false},      // Hidden text file (ignored)
		{".DS_Store", false},        // System file (ignored)
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			container := notice.NewNoticeContainer()
			validator := NewUnknownFileValidator()

			validator.validateKnownFile(container, tt.filename)

			notices := container.GetNotices()
			hasNotice := len(notices) > 0

			// Files should only generate notices if they're .txt or .geojson and not hidden
			shouldCheck := (strings.HasSuffix(tt.filename, ".txt") || strings.HasSuffix(tt.filename, ".geojson")) &&
				!strings.HasPrefix(tt.filename, ".")

			if shouldCheck {
				// For files that should be checked, notice depends on whether file is known
				expectedNotice := !knownGTFSFiles[tt.filename]
				if hasNotice != expectedNotice {
					t.Errorf("File %s: expected notice=%v, got notice=%v", tt.filename, expectedNotice, hasNotice)
				}
			} else if hasNotice {
				// Files that shouldn't be checked should never generate notices
				t.Errorf("File %s should be ignored but generated a notice", tt.filename)
			}
		})
	}
}
