package core

import (
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/parser"
	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

func TestFieldFormatValidator_Validate(t *testing.T) {
	tests := []struct {
		name                string
		files               map[string]string
		expectedNoticeCodes []string
		description         string
	}{
		{
			name: "all fields valid format",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone,agency_email\n1,Metro,http://metro.example,America/Los_Angeles,contact@metro.example",
				"stops.txt":  "stop_id,stop_name,stop_url,stop_timezone\n1,Main St,https://stops.example/1,America/Los_Angeles",
			},
			expectedNoticeCodes: []string{},
			description:         "All fields have valid formats",
		},
		{
			name: "agency.txt invalid URL",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,invalid-url,America/Los_Angeles",
			},
			expectedNoticeCodes: []string{"invalid_url"},
			description:         "Invalid agency_url format",
		},
		{
			name: "agency.txt invalid email",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone,agency_email\n1,Metro,http://metro.example,America/Los_Angeles,invalid-email",
			},
			expectedNoticeCodes: []string{"invalid_email"},
			description:         "Invalid agency_email format",
		},
		{
			name: "agency.txt invalid timezone",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,http://metro.example,Invalid/Timezone",
			},
			expectedNoticeCodes: []string{"invalid_timezone"},
			description:         "Invalid agency_timezone",
		},
		{
			name: "stops.txt invalid URL",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_url\n1,Main St,ftp://invalid.protocol",
			},
			expectedNoticeCodes: []string{"invalid_url"},
			description:         "Invalid stop_url (FTP not allowed)",
		},
		{
			name: "stops.txt invalid timezone",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_timezone\n1,Main St,Invalid/Zone",
			},
			expectedNoticeCodes: []string{"invalid_timezone"},
			description:         "Invalid stop_timezone",
		},
		{
			name: "routes.txt invalid URL",
			files: map[string]string{
				"routes.txt": "route_id,agency_id,route_short_name,route_type,route_url\n1,1,Red,3,not-a-url",
			},
			expectedNoticeCodes: []string{"invalid_url"},
			description:         "Invalid route_url",
		},
		{
			name: "routes.txt invalid color",
			files: map[string]string{
				"routes.txt": "route_id,agency_id,route_short_name,route_type,route_color\n1,1,Red,3,GGGGGG",
			},
			expectedNoticeCodes: []string{"invalid_field_format"},
			description:         "Invalid route_color (non-hex characters)",
		},
		{
			name: "routes.txt invalid text color",
			files: map[string]string{
				"routes.txt": "route_id,agency_id,route_short_name,route_type,route_text_color\n1,1,Red,3,12345",
			},
			expectedNoticeCodes: []string{"invalid_field_format"},
			description:         "Invalid route_text_color (too short)",
		},
		{
			name: "stop_times.txt invalid arrival time",
			files: map[string]string{
				"stop_times.txt": "trip_id,arrival_time,departure_time,stop_id,stop_sequence\nT1,25:30:00,08:00:00,1,1",
			},
			expectedNoticeCodes: []string{},
			description:         "Valid arrival_time format (25 hours is valid in GTFS for next-day service)",
		},
		{
			name: "stop_times.txt invalid departure time",
			files: map[string]string{
				"stop_times.txt": "trip_id,arrival_time,departure_time,stop_id,stop_sequence\nT1,08:00:00,8:0:0,1,1",
			},
			expectedNoticeCodes: []string{"invalid_field_format"},
			description:         "Invalid departure_time format (not zero-padded)",
		},
		{
			name: "calendar.txt invalid start date",
			files: map[string]string{
				"calendar.txt": "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\nS1,1,1,1,1,1,0,0,2025-01-01,20251231",
			},
			expectedNoticeCodes: []string{"invalid_field_format"},
			description:         "Invalid start_date format (with hyphens)",
		},
		{
			name: "calendar.txt invalid end date",
			files: map[string]string{
				"calendar.txt": "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\nS1,1,1,1,1,1,0,0,20250101,25251231",
			},
			expectedNoticeCodes: []string{},
			description:         "Valid end_date format (year 2525 is valid according to GTFS specification)",
		},
		{
			name: "calendar_dates.txt invalid date",
			files: map[string]string{
				"calendar_dates.txt": "service_id,date,exception_type\nS1,20250230,1", // Feb 30th doesn't exist
			},
			expectedNoticeCodes: []string{"invalid_field_format"},
			description:         "Invalid date in calendar_dates.txt",
		},
		{
			name: "multiple invalid formats in one file",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone,agency_email\n1,Metro,invalid-url,Invalid/Zone,bad-email",
			},
			expectedNoticeCodes: []string{"invalid_url", "invalid_timezone", "invalid_email"},
			description:         "Multiple format errors in single row",
		},
		{
			name: "multiple rows with format errors",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,invalid-url-1,America/Los_Angeles\n2,Bus,invalid-url-2,Invalid/Zone",
			},
			expectedNoticeCodes: []string{"invalid_url", "invalid_url", "invalid_timezone"},
			description:         "Format errors across multiple rows",
		},
		{
			name: "valid HTTPS URLs",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,https://metro.example,America/Los_Angeles",
				"stops.txt":  "stop_id,stop_name,stop_url\n1,Main St,https://stops.example/1",
			},
			expectedNoticeCodes: []string{},
			description:         "HTTPS URLs should be valid",
		},
		{
			name: "valid HTTP URLs",
			files: map[string]string{
				"routes.txt": "route_id,agency_id,route_short_name,route_type,route_url\n1,1,Red,3,http://routes.example/red",
			},
			expectedNoticeCodes: []string{},
			description:         "HTTP URLs should be valid",
		},
		{
			name: "valid colors",
			files: map[string]string{
				"routes.txt": "route_id,agency_id,route_short_name,route_type,route_color,route_text_color\n1,1,Red,3,FF0000,FFFFFF",
			},
			expectedNoticeCodes: []string{},
			description:         "Valid hex colors",
		},
		{
			name: "valid times",
			files: map[string]string{
				"stop_times.txt": "trip_id,arrival_time,departure_time,stop_id,stop_sequence\nT1,08:00:00,08:01:00,1,1\nT1,24:00:00,24:00:00,2,2", // 24:00:00 is valid in GTFS
			},
			expectedNoticeCodes: []string{},
			description:         "Valid GTFS time formats including 24:00:00",
		},
		{
			name: "valid dates",
			files: map[string]string{
				"calendar.txt": "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\nS1,1,1,1,1,1,0,0,20250101,20251231",
			},
			expectedNoticeCodes: []string{},
			description:         "Valid YYYYMMDD date format",
		},
		{
			name: "empty optional fields",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone,agency_email\n1,Metro,http://metro.example,America/Los_Angeles,", // Empty email
				"stops.txt":  "stop_id,stop_name,stop_url,stop_timezone\n1,Main St,,",                                                            // Empty URL and timezone
			},
			expectedNoticeCodes: []string{},
			description:         "Empty optional fields should not generate format errors",
		},
		{
			name: "files without format validation",
			files: map[string]string{
				"trips.txt":        "route_id,service_id,trip_id\n1,S1,T1",
				"translations.txt": "table_name,field_name,language,translation\nstops,stop_name,es,Calle Principal",
			},
			expectedNoticeCodes: []string{},
			description:         "Files without specific format rules should not generate errors",
		},
		{
			name: "valid timezones",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,http://metro.example,America/New_York\n2,Bus,http://bus.example,Europe/London\n3,Rail,http://rail.example,Asia/Tokyo",
			},
			expectedNoticeCodes: []string{},
			description:         "Various valid timezone formats",
		},
		{
			name: "edge case URLs",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,http://localhost:8080/path,America/Los_Angeles\n2,Bus,https://example.com:443/,America/Los_Angeles",
			},
			expectedNoticeCodes: []string{},
			description:         "URLs with ports and paths should be valid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test components
			loader := CreateTestFeedLoader(t, tt.files)
			container := notice.NewNoticeContainer()
			validator := NewFieldFormatValidator()
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

func TestFieldFormatValidator_IsValidURL(t *testing.T) {
	validator := NewFieldFormatValidator()

	tests := []struct {
		url      string
		expected bool
	}{
		{"http://example.com", true},
		{"https://example.com", true},
		{"http://localhost", true},
		{"https://example.com:8080", true},
		{"http://example.com/path", true},
		{"https://example.com/path?query=value", true},
		{"ftp://example.com", false},       // FTP not allowed
		{"mailto:test@example.com", false}, // mailto not allowed
		{"example.com", false},             // Missing scheme
		{"http://", false},                 // Missing host
		{"not-a-url", false},               // Invalid format
		{"", false},                        // Empty string
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			result := validator.isValidURL(tt.url)
			if result != tt.expected {
				t.Errorf("isValidURL(%q) = %v, expected %v", tt.url, result, tt.expected)
			}
		})
	}
}

func TestFieldFormatValidator_IsValidEmail(t *testing.T) {
	validator := NewFieldFormatValidator()

	tests := []struct {
		email    string
		expected bool
	}{
		{"test@example.com", true},
		{"user.name@example.com", true},
		{"user+tag@example.com", true},
		{"user@subdomain.example.com", true},
		{"invalid-email", false},
		{"@example.com", false},          // Missing local part
		{"test@", false},                 // Missing domain
		{"test@.com", false},             // Invalid domain
		{"test test@example.com", false}, // Space in local part
		{"", false},                      // Empty string
	}

	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			result := validator.isValidEmail(tt.email)
			if result != tt.expected {
				t.Errorf("isValidEmail(%q) = %v, expected %v", tt.email, result, tt.expected)
			}
		})
	}
}

func TestFieldFormatValidator_IsValidTimezone(t *testing.T) {
	validator := NewFieldFormatValidator()

	tests := []struct {
		timezone string
		expected bool
	}{
		{"America/New_York", true},
		{"Europe/London", true},
		{"Asia/Tokyo", true},
		{"UTC", true},
		{"America/Los_Angeles", true},
		{"Invalid/Timezone", false},
		{"America/NonExistent", false},
		{"", false},        // Empty string
		{"America", false}, // Incomplete timezone
	}

	for _, tt := range tests {
		t.Run(tt.timezone, func(t *testing.T) {
			result := validator.isValidTimezone(tt.timezone)
			if result != tt.expected {
				t.Errorf("isValidTimezone(%q) = %v, expected %v", tt.timezone, result, tt.expected)
			}
		})
	}
}

func TestFieldFormatValidator_IsValidPhoneNumber(t *testing.T) {
	validator := NewFieldFormatValidator()

	tests := []struct {
		phone    string
		expected bool
	}{
		{"555-123-4567", true},
		{"(555) 123-4567", true},
		{"+1-555-123-4567", true},
		{"5551234567", true},
		{"555.123.4567", true},
		{"123", false},           // Too short
		{"abcd-efg-hijk", false}, // Non-numeric
		{"", false},              // Empty
	}

	for _, tt := range tests {
		t.Run(tt.phone, func(t *testing.T) {
			result := validator.isValidPhoneNumber(tt.phone)
			if result != tt.expected {
				t.Errorf("isValidPhoneNumber(%q) = %v, expected %v", tt.phone, result, tt.expected)
			}
		})
	}
}

func TestFieldFormatValidator_ValidateAgencyFields(t *testing.T) {
	tests := []struct {
		name            string
		rowData         map[string]string
		expectedNotices []string
	}{
		{
			name: "all valid agency fields",
			rowData: map[string]string{
				"agency_url":      "https://metro.example",
				"agency_email":    "contact@metro.example",
				"agency_timezone": "America/Los_Angeles",
			},
			expectedNotices: []string{},
		},
		{
			name: "invalid URL",
			rowData: map[string]string{
				"agency_url": "invalid-url",
			},
			expectedNotices: []string{"invalid_url"},
		},
		{
			name: "invalid email",
			rowData: map[string]string{
				"agency_email": "invalid-email",
			},
			expectedNotices: []string{"invalid_email"},
		},
		{
			name: "invalid timezone",
			rowData: map[string]string{
				"agency_timezone": "Invalid/Zone",
			},
			expectedNotices: []string{"invalid_timezone"},
		},
		{
			name: "multiple invalid fields",
			rowData: map[string]string{
				"agency_url":      "not-a-url",
				"agency_email":    "bad-email",
				"agency_timezone": "Bad/Zone",
			},
			expectedNotices: []string{"invalid_url", "invalid_email", "invalid_timezone"},
		},
		{
			name:            "empty fields",
			rowData:         map[string]string{},
			expectedNotices: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			container := notice.NewNoticeContainer()
			validator := NewFieldFormatValidator()

			// Create mock row
			row := &parser.CSVRow{
				RowNumber: 1,
				Values:    tt.rowData,
			}

			validator.validateAgencyFields(row, container, "agency.txt")

			notices := container.GetNotices()

			// Check notice count and codes
			expectedCodeCounts := make(map[string]int)
			for _, code := range tt.expectedNotices {
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
		})
	}
}

func TestFieldFormatValidator_ValidateRouteFields(t *testing.T) {
	tests := []struct {
		name            string
		rowData         map[string]string
		expectedNotices []string
	}{
		{
			name: "valid route fields",
			rowData: map[string]string{
				"route_url":        "https://routes.example/red",
				"route_color":      "FF0000",
				"route_text_color": "FFFFFF",
			},
			expectedNotices: []string{},
		},
		{
			name: "invalid route color",
			rowData: map[string]string{
				"route_color": "GGGGGG", // Invalid hex
			},
			expectedNotices: []string{"invalid_field_format"},
		},
		{
			name: "invalid route text color",
			rowData: map[string]string{
				"route_text_color": "12345", // Too short
			},
			expectedNotices: []string{"invalid_field_format"},
		},
		{
			name: "invalid route URL",
			rowData: map[string]string{
				"route_url": "not-a-url",
			},
			expectedNotices: []string{"invalid_url"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			container := notice.NewNoticeContainer()
			validator := NewFieldFormatValidator()

			row := &parser.CSVRow{
				RowNumber: 1,
				Values:    tt.rowData,
			}

			validator.validateRouteFields(row, container, "routes.txt")

			notices := container.GetNotices()

			expectedCodeCounts := make(map[string]int)
			for _, code := range tt.expectedNotices {
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
		})
	}
}

func TestFieldFormatValidator_New(t *testing.T) {
	validator := NewFieldFormatValidator()
	if validator == nil {
		t.Error("NewFieldFormatValidator() returned nil")
	}
}
