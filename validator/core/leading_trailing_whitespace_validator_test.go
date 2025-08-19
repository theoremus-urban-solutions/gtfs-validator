package core

import (
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/testutil"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

func TestLeadingTrailingWhitespaceValidator_Validate(t *testing.T) {
	tests := []struct {
		name                string
		files               map[string]string
		expectedNoticeCodes []string
		description         string
	}{
		{
			name: "no whitespace issues",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,http://metro.example,America/Los_Angeles",
				"stops.txt":  "stop_id,stop_name,stop_lat,stop_lon\n1,Main St,34.05,-118.25",
			},
			expectedNoticeCodes: []string{},
			description:         "Fields without whitespace issues should not generate notices",
		},
		{
			name: "leading whitespace in agency_name",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\n1, Metro,http://metro.example,America/Los_Angeles", // Leading space
			},
			expectedNoticeCodes: []string{"leading_whitespace"},
			description:         "Leading whitespace should generate notice",
		},
		{
			name: "trailing whitespace in stop_name",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_lat,stop_lon\n1,Main St ,34.05,-118.25", // Trailing space
			},
			expectedNoticeCodes: []string{"trailing_whitespace"},
			description:         "Trailing whitespace should generate notice",
		},
		{
			name: "both leading and trailing whitespace",
			files: map[string]string{
				"routes.txt": "route_id,agency_id,route_short_name,route_long_name,route_type\n1,1, Red Line ,Red Metro Line,3", // Both spaces
			},
			expectedNoticeCodes: []string{"leading_whitespace", "trailing_whitespace"},
			description:         "Both leading and trailing whitespace should generate separate notices",
		},
		{
			name: "tab characters as whitespace",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\n1,\tMetro\t,http://metro.example,America/Los_Angeles", // Tab characters
			},
			expectedNoticeCodes: []string{"leading_whitespace", "trailing_whitespace"},
			description:         "Tab characters should be detected as whitespace",
		},
		{
			name: "whitespace-only field",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_desc,stop_lat,stop_lon\n1,Main St,   ,34.05,-118.25", // Whitespace-only description
			},
			expectedNoticeCodes: []string{"leading_whitespace", "trailing_whitespace", "excessive_whitespace", "whitespace_only_field"},
			description:         "Fields containing only whitespace should generate multiple notices",
		},
		{
			name: "excessive internal whitespace",
			files: map[string]string{
				"routes.txt": "route_id,agency_id,route_short_name,route_long_name,route_type\n1,1,Red,Red  Line,3", // Double space
			},
			expectedNoticeCodes: []string{"excessive_whitespace"},
			description:         "Multiple consecutive spaces should generate notice",
		},
		{
			name: "multiple whitespace issues in same field",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\n1, Metro  Transit ,http://metro.example,America/Los_Angeles", // Leading, trailing, and excessive
			},
			expectedNoticeCodes: []string{"leading_whitespace", "trailing_whitespace", "excessive_whitespace"},
			description:         "Multiple whitespace issues in same field should generate multiple notices",
		},
		{
			name: "whitespace issues across multiple files",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\n1, Metro,http://metro.example,America/Los_Angeles", // Leading space
				"stops.txt":  "stop_id,stop_name,stop_lat,stop_lon\n1,Main St ,34.05,-118.25",                                       // Trailing space
			},
			expectedNoticeCodes: []string{"leading_whitespace", "trailing_whitespace"},
			description:         "Whitespace issues in different files should each generate notices",
		},
		{
			name: "multiple rows with whitespace issues",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\n1, Metro,http://metro.example,America/Los_Angeles\n2,Bus ,http://bus.example,America/Los_Angeles", // Multiple rows
			},
			expectedNoticeCodes: []string{"leading_whitespace", "trailing_whitespace"},
			description:         "Multiple rows with whitespace issues should generate multiple notices",
		},
		{
			name: "empty fields ignored",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_desc,stop_lat,stop_lon\n1,Main St,,34.05,-118.25", // Empty description
			},
			expectedNoticeCodes: []string{},
			description:         "Empty fields should not generate whitespace notices",
		},
		{
			name: "numeric fields not validated",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_lat,stop_lon,location_type\n1,Main St, 34.05 , -118.25 ,0", // Whitespace around numeric values
			},
			expectedNoticeCodes: []string{},
			description:         "Numeric fields should not be validated for whitespace",
		},
		{
			name: "ID fields with whitespace",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\n 1 ,Metro,http://metro.example,America/Los_Angeles", // Whitespace around ID
			},
			expectedNoticeCodes: []string{"leading_whitespace", "trailing_whitespace"},
			description:         "ID fields should be validated for whitespace",
		},
		{
			name: "URL fields with whitespace",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\n1,Metro, http://metro.example ,America/Los_Angeles", // Whitespace around URL
			},
			expectedNoticeCodes: []string{"leading_whitespace", "trailing_whitespace"},
			description:         "URL fields should be validated for whitespace",
		},
		{
			name: "timezone fields with whitespace",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,http://metro.example, America/Los_Angeles ", // Whitespace around timezone
			},
			expectedNoticeCodes: []string{"leading_whitespace", "trailing_whitespace"},
			description:         "Timezone fields should be validated for whitespace",
		},
		{
			name: "time fields with whitespace",
			files: map[string]string{
				"stop_times.txt": "trip_id,arrival_time,departure_time,stop_id,stop_sequence\nT1, 08:00:00 , 08:00:00 ,1,1", // Whitespace around times
			},
			expectedNoticeCodes: []string{"leading_whitespace", "trailing_whitespace", "leading_whitespace", "trailing_whitespace"},
			description:         "Time fields should be validated for whitespace",
		},
		{
			name: "date fields with whitespace",
			files: map[string]string{
				"calendar.txt": "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\nS1,1,1,1,1,1,0,0, 20250101 , 20251231 ", // Whitespace around dates
			},
			expectedNoticeCodes: []string{"leading_whitespace", "trailing_whitespace", "leading_whitespace", "trailing_whitespace"},
			description:         "Date fields should be validated for whitespace",
		},
		{
			name: "boolean/enum fields not validated",
			files: map[string]string{
				"calendar.txt": "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\nS1, 1 , 1 , 1 , 1 , 1 , 0 , 0 ,20250101,20251231", // Whitespace around boolean values
			},
			expectedNoticeCodes: []string{},
			description:         "Boolean/enum fields should not be validated for whitespace",
		},
		{
			name: "headsign fields with whitespace",
			files: map[string]string{
				"trips.txt": "route_id,service_id,trip_id,trip_headsign,direction_id\n1,S1,T1, Downtown ,0", // Whitespace around headsign
			},
			expectedNoticeCodes: []string{"leading_whitespace", "trailing_whitespace"},
			description:         "Headsign fields should be validated for whitespace",
		},
		{
			name: "color fields with whitespace",
			files: map[string]string{
				"routes.txt": "route_id,agency_id,route_short_name,route_type,route_color,route_text_color\n1,1,Red,3, FF0000 , FFFFFF ", // Whitespace around colors
			},
			expectedNoticeCodes: []string{"leading_whitespace", "trailing_whitespace", "leading_whitespace", "trailing_whitespace"},
			description:         "Color fields should be validated for whitespace",
		},
		{
			name: "feed_info fields with whitespace",
			files: map[string]string{
				"feed_info.txt": "feed_publisher_name,feed_publisher_url,feed_lang\n Metro Transit , http://metro.example ,en", // Whitespace in feed info
			},
			expectedNoticeCodes: []string{"leading_whitespace", "trailing_whitespace", "leading_whitespace", "trailing_whitespace"},
			description:         "Feed info fields should be validated for whitespace",
		},
		{
			name: "missing files ignored",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,http://metro.example,America/Los_Angeles",
			},
			expectedNoticeCodes: []string{},
			description:         "Missing optional files should not cause validation errors",
		},
		{
			name: "files without significant fields use text field detection",
			files: map[string]string{
				"custom.txt": "text_field,numeric_field,id_field\nvalue1,123,ID1", // Custom file not in validator's field map
			},
			expectedNoticeCodes: []string{},
			description:         "Unknown files should validate based on field type detection",
		},
		{
			name: "mixed valid and invalid fields",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\n1, Metro,http://metro.example,America/Los_Angeles\n2,Bus,http://bus.example,America/Los_Angeles", // Mixed valid/invalid
			},
			expectedNoticeCodes: []string{"leading_whitespace"},
			description:         "Mix of valid and invalid fields should only generate notices for invalid ones",
		},
		{
			name: "all GTFS files with various whitespace issues",
			files: map[string]string{
				"agency.txt":          "agency_id,agency_name,agency_url,agency_timezone\n1, Metro ,http://metro.example,America/Los_Angeles",
				"stops.txt":           "stop_id,stop_name,stop_lat,stop_lon\n1, Main St ,34.05,-118.25",
				"routes.txt":          "route_id,agency_id,route_short_name,route_type\n1,1, Red ,3",
				"trips.txt":           "route_id,service_id,trip_id,trip_headsign\n1,S1, T1 , Downtown ",
				"stop_times.txt":      "trip_id,arrival_time,departure_time,stop_id,stop_sequence\n T1 , 08:00:00 , 08:00:00 ,1,1",
				"calendar.txt":        "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\n S1 ,1,1,1,1,1,0,0, 20250101 , 20251231 ",
				"calendar_dates.txt":  "service_id,date,exception_type\n S1 , 20250101 ,1",
				"fare_attributes.txt": "fare_id,price,currency_type\n F1 , 2.50 , USD ",
				"fare_rules.txt":      "fare_id,route_id\n F1 , R1 ",
				"shapes.txt":          "shape_id,shape_pt_lat,shape_pt_lon,shape_pt_sequence\n S1 ,34.05,-118.25,1",
				"frequencies.txt":     "trip_id,start_time,end_time,headway_secs\n T1 , 06:00:00 , 22:00:00 ,600",
				"transfers.txt":       "from_stop_id,to_stop_id,transfer_type\n 1 , 2 ,0",
				"pathways.txt":        "pathway_id,from_stop_id,to_stop_id,pathway_mode\n P1 , 1 , 2 ,1",
				"levels.txt":          "level_id,level_index,level_name\n L1 ,0, Ground Floor ",
				"feed_info.txt":       "feed_publisher_name,feed_publisher_url,feed_lang\n Metro Transit , http://metro.example ,en",
				"attributions.txt":    "attribution_id,organization_name\n A1 , Metro ",
			},
			expectedNoticeCodes: []string{
				"leading_whitespace", "trailing_whitespace", // agency
				"leading_whitespace", "trailing_whitespace", // stops
				"leading_whitespace", "trailing_whitespace", // routes
				"leading_whitespace", "trailing_whitespace", "leading_whitespace", "trailing_whitespace", // trips
				"leading_whitespace", "trailing_whitespace", "leading_whitespace", "trailing_whitespace", "leading_whitespace", "trailing_whitespace", // stop_times
				"leading_whitespace", "trailing_whitespace", "leading_whitespace", "trailing_whitespace", "leading_whitespace", "trailing_whitespace", // calendar
				"leading_whitespace", "trailing_whitespace", "leading_whitespace", "trailing_whitespace", // calendar_dates
				"leading_whitespace", "trailing_whitespace", "leading_whitespace", "trailing_whitespace", "leading_whitespace", "trailing_whitespace", // fare_attributes
				"leading_whitespace", "trailing_whitespace", "leading_whitespace", "trailing_whitespace", // fare_rules
				"leading_whitespace", "trailing_whitespace", // shapes
				"leading_whitespace", "trailing_whitespace", "leading_whitespace", "trailing_whitespace", "leading_whitespace", "trailing_whitespace", // frequencies
				"leading_whitespace", "trailing_whitespace", "leading_whitespace", "trailing_whitespace", // transfers
				"leading_whitespace", "trailing_whitespace", "leading_whitespace", "trailing_whitespace", "leading_whitespace", "trailing_whitespace", // pathways
				"leading_whitespace", "trailing_whitespace", "leading_whitespace", "trailing_whitespace", // levels
				"leading_whitespace", "trailing_whitespace", "leading_whitespace", "trailing_whitespace", // feed_info
				"leading_whitespace", "trailing_whitespace", "leading_whitespace", "trailing_whitespace", // attributions
			},
			description: "All GTFS files with whitespace issues should generate appropriate notices",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test components
			loader := testutil.CreateTestFeedLoader(t, tt.files)
			container := notice.NewNoticeContainer()
			validator := NewLeadingTrailingWhitespaceValidator()
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

func TestLeadingTrailingWhitespaceValidator_ValidateFieldWhitespace(t *testing.T) {
	tests := []struct {
		name            string
		filename        string
		fieldName       string
		fieldValue      string
		expectedNotices []string
		description     string
	}{
		{
			name:            "clean field value",
			filename:        "agency.txt",
			fieldName:       "agency_name",
			fieldValue:      "Metro",
			expectedNotices: []string{},
			description:     "Clean field value should not generate notices",
		},
		{
			name:            "leading space",
			filename:        "agency.txt",
			fieldName:       "agency_name",
			fieldValue:      " Metro",
			expectedNotices: []string{"leading_whitespace"},
			description:     "Leading space should generate notice",
		},
		{
			name:            "trailing space",
			filename:        "agency.txt",
			fieldName:       "agency_name",
			fieldValue:      "Metro ",
			expectedNotices: []string{"trailing_whitespace"},
			description:     "Trailing space should generate notice",
		},
		{
			name:            "both leading and trailing spaces",
			filename:        "agency.txt",
			fieldName:       "agency_name",
			fieldValue:      " Metro ",
			expectedNotices: []string{"leading_whitespace", "trailing_whitespace"},
			description:     "Both leading and trailing spaces should generate separate notices",
		},
		{
			name:            "leading tab",
			filename:        "agency.txt",
			fieldName:       "agency_name",
			fieldValue:      "\tMetro",
			expectedNotices: []string{"leading_whitespace"},
			description:     "Leading tab should generate notice",
		},
		{
			name:            "trailing tab",
			filename:        "agency.txt",
			fieldName:       "agency_name",
			fieldValue:      "Metro\t",
			expectedNotices: []string{"trailing_whitespace"},
			description:     "Trailing tab should generate notice",
		},
		{
			name:            "whitespace-only field",
			filename:        "agency.txt",
			fieldName:       "agency_name",
			fieldValue:      "   ",
			expectedNotices: []string{"leading_whitespace", "trailing_whitespace", "excessive_whitespace", "whitespace_only_field"},
			description:     "Whitespace-only field should generate multiple notices",
		},
		{
			name:            "tab-only field",
			filename:        "agency.txt",
			fieldName:       "agency_name",
			fieldValue:      "\t\t",
			expectedNotices: []string{"leading_whitespace", "trailing_whitespace", "whitespace_only_field"},
			description:     "Tab-only field should generate multiple notices",
		},
		{
			name:            "excessive internal whitespace",
			filename:        "agency.txt",
			fieldName:       "agency_name",
			fieldValue:      "Metro  Transit",
			expectedNotices: []string{"excessive_whitespace"},
			description:     "Double spaces should generate excessive whitespace notice",
		},
		{
			name:            "triple spaces",
			filename:        "agency.txt",
			fieldName:       "agency_name",
			fieldValue:      "Metro   Transit",
			expectedNotices: []string{"excessive_whitespace"},
			description:     "Triple spaces should generate excessive whitespace notice",
		},
		{
			name:            "multiple double spaces",
			filename:        "agency.txt",
			fieldName:       "agency_name",
			fieldValue:      "Metro  Transit  Authority",
			expectedNotices: []string{"excessive_whitespace"},
			description:     "Multiple double spaces should generate one excessive whitespace notice",
		},
		{
			name:            "all whitespace issues combined",
			filename:        "agency.txt",
			fieldName:       "agency_name",
			fieldValue:      " Metro  Transit ",
			expectedNotices: []string{"leading_whitespace", "trailing_whitespace", "excessive_whitespace"},
			description:     "All whitespace issues should generate separate notices",
		},
		{
			name:            "mixed whitespace types",
			filename:        "agency.txt",
			fieldName:       "agency_name",
			fieldValue:      "\t Metro  Transit \t",
			expectedNotices: []string{"leading_whitespace", "trailing_whitespace", "excessive_whitespace"},
			description:     "Mixed tabs and spaces should be detected",
		},
		{
			name:            "single space between words is valid",
			filename:        "agency.txt",
			fieldName:       "agency_name",
			fieldValue:      "Metro Transit",
			expectedNotices: []string{},
			description:     "Single space between words should be valid",
		},
		{
			name:            "newline characters as whitespace",
			filename:        "agency.txt",
			fieldName:       "agency_name",
			fieldValue:      "\nMetro\n",
			expectedNotices: []string{},
			description:     "Newlines are not considered leading/trailing whitespace in this context",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			container := notice.NewNoticeContainer()
			validator := NewLeadingTrailingWhitespaceValidator()

			validator.validateFieldWhitespace(container, tt.filename, tt.fieldName, tt.fieldValue, 1)

			notices := container.GetNotices()
			actualCodes := make([]string, len(notices))
			for i, notice := range notices {
				actualCodes[i] = notice.Code()
			}

			// Check notice count
			if len(actualCodes) != len(tt.expectedNotices) {
				t.Errorf("Expected %d notices, got %d for %s", len(tt.expectedNotices), len(actualCodes), tt.description)
			}

			// Count notice codes
			expectedCodeCounts := make(map[string]int)
			for _, code := range tt.expectedNotices {
				expectedCodeCounts[code]++
			}

			actualCodeCounts := make(map[string]int)
			for _, code := range actualCodes {
				actualCodeCounts[code]++
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

func TestLeadingTrailingWhitespaceValidator_IsTextField(t *testing.T) {
	validator := NewLeadingTrailingWhitespaceValidator()

	textFields := []string{
		"agency_name", "agency_url", "stop_name", "stop_desc", "route_short_name",
		"route_long_name", "trip_headsign", "stop_headsign", "service_id", "trip_id",
		"stop_id", "route_id", "agency_id", "shape_id", "fare_id", "zone_id",
	}

	numericFields := []string{
		"stop_lat", "stop_lon", "route_type", "direction_id", "location_type",
		"wheelchair_boarding", "wheelchair_accessible", "bikes_allowed", "stop_sequence",
		"pickup_type", "drop_off_type", "timepoint", "monday", "tuesday", "exception_type",
		"payment_method", "transfers", "shape_pt_lat", "shape_pt_lon", "shape_pt_sequence",
		"headway_secs", "exact_times", "transfer_type", "min_transfer_time",
	}

	// Test text fields
	for _, field := range textFields {
		if !validator.isTextField(field) {
			t.Errorf("Expected '%s' to be identified as a text field", field)
		}
	}

	// Test numeric fields
	for _, field := range numericFields {
		if validator.isTextField(field) {
			t.Errorf("Expected '%s' to be identified as a numeric field", field)
		}
	}
}

func TestLeadingTrailingWhitespaceValidator_ShouldValidateField(t *testing.T) {
	validator := NewLeadingTrailingWhitespaceValidator()

	tests := []struct {
		fieldName         string
		significantFields map[string]bool
		expected          bool
		description       string
	}{
		{
			fieldName:         "agency_name",
			significantFields: map[string]bool{"agency_name": true, "agency_id": true},
			expected:          true,
			description:       "Field in significant fields should be validated",
		},
		{
			fieldName:         "agency_phone",
			significantFields: map[string]bool{"agency_name": true, "agency_id": true},
			expected:          false,
			description:       "Field not in significant fields should not be validated",
		},
		{
			fieldName:         "agency_name",
			significantFields: map[string]bool{}, // Empty map - use text field detection
			expected:          true,
			description:       "With empty significant fields, text fields should be validated",
		},
		{
			fieldName:         "stop_lat",
			significantFields: map[string]bool{}, // Empty map - use text field detection
			expected:          false,
			description:       "With empty significant fields, numeric fields should not be validated",
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			result := validator.shouldValidateField(tt.fieldName, tt.significantFields)
			if result != tt.expected {
				t.Errorf("Expected shouldValidateField('%s') to return %v, got %v", tt.fieldName, tt.expected, result)
			}
		})
	}
}

func TestLeadingTrailingWhitespaceValidator_GetSignificantFields(t *testing.T) {
	validator := NewLeadingTrailingWhitespaceValidator()

	tests := []struct {
		filename         string
		expectedFields   []string
		unexpectedFields []string
		description      string
	}{
		{
			filename:         "agency.txt",
			expectedFields:   []string{"agency_id", "agency_name", "agency_url", "agency_timezone"},
			unexpectedFields: []string{"stop_id", "route_id"},
			description:      "Agency file should have agency-specific fields",
		},
		{
			filename:         "stops.txt",
			expectedFields:   []string{"stop_id", "stop_name", "zone_id", "parent_station"},
			unexpectedFields: []string{"agency_id", "route_id"},
			description:      "Stops file should have stop-specific fields",
		},
		{
			filename:         "routes.txt",
			expectedFields:   []string{"route_id", "agency_id", "route_short_name", "route_long_name"},
			unexpectedFields: []string{"stop_id", "trip_id"},
			description:      "Routes file should have route-specific fields",
		},
		{
			filename:         "unknown_file.txt",
			expectedFields:   []string{},
			unexpectedFields: []string{"agency_id", "stop_id"},
			description:      "Unknown files should return empty significant fields map",
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			significantFields := validator.getSignificantFields(tt.filename)

			// Check expected fields are present
			for _, field := range tt.expectedFields {
				if !significantFields[field] {
					t.Errorf("Expected field '%s' to be significant for %s", field, tt.filename)
				}
			}

			// Check unexpected fields are not present
			for _, field := range tt.unexpectedFields {
				if significantFields[field] {
					t.Errorf("Did not expect field '%s' to be significant for %s", field, tt.filename)
				}
			}
		})
	}
}

func TestLeadingTrailingWhitespaceValidator_ValidateFile(t *testing.T) {
	tests := []struct {
		name                string
		filename            string
		content             string
		expectedNoticeCount int
		description         string
	}{
		{
			name:                "file with no whitespace issues",
			filename:            "agency.txt",
			content:             "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,http://metro.example,America/Los_Angeles",
			expectedNoticeCount: 0,
			description:         "Clean file should generate no notices",
		},
		{
			name:                "file with multiple whitespace issues",
			filename:            "agency.txt",
			content:             "agency_id,agency_name,agency_url,agency_timezone\n 1 , Metro , http://metro.example ,America/Los_Angeles",
			expectedNoticeCount: 6, // Leading and trailing for first 3 fields
			description:         "File with multiple issues should generate multiple notices",
		},
		{
			name:                "file with CSV parsing error",
			filename:            "agency.txt",
			content:             "agency_id,agency_name,agency_url,agency_timezone\n1,Metro", // Incomplete row
			expectedNoticeCount: 0,
			description:         "CSV parsing errors should not cause crashes",
		},
		{
			name:                "empty file",
			filename:            "agency.txt",
			content:             "agency_id,agency_name,agency_url,agency_timezone", // Headers only
			expectedNoticeCount: 0,
			description:         "Empty file (headers only) should generate no notices",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files := map[string]string{tt.filename: tt.content}
			loader := testutil.CreateTestFeedLoader(t, files)
			container := notice.NewNoticeContainer()
			validator := NewLeadingTrailingWhitespaceValidator()

			validator.validateFile(loader, container, tt.filename)

			notices := container.GetNotices()
			if len(notices) != tt.expectedNoticeCount {
				t.Errorf("Expected %d notices, got %d for %s", tt.expectedNoticeCount, len(notices), tt.description)
			}
		})
	}
}

func TestLeadingTrailingWhitespaceValidator_New(t *testing.T) {
	validator := NewLeadingTrailingWhitespaceValidator()
	if validator == nil {
		t.Error("NewLeadingTrailingWhitespaceValidator() returned nil")
	}
}
