package core

import (
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

func TestMissingFilesValidator_Validate(t *testing.T) {
	tests := []struct {
		name                string
		files               map[string]string
		expectedNoticeCodes []string
		description         string
	}{
		{
			name: "all required files present",
			files: map[string]string{
				"agency.txt":     "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,http://metro.example,America/Los_Angeles",
				"stops.txt":      "stop_id,stop_name,stop_lat,stop_lon\n1,Main St,34.05,-118.25",
				"routes.txt":     "route_id,agency_id,route_short_name,route_long_name,route_type\n1,1,1,Main Line,3",
				"trips.txt":      "route_id,service_id,trip_id\n1,S1,T1",
				"stop_times.txt": "trip_id,arrival_time,departure_time,stop_id,stop_sequence\nT1,08:00:00,08:00:00,1,1",
				"calendar.txt":   "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\nS1,1,1,1,1,1,0,0,20250101,20251231",
			},
			expectedNoticeCodes: []string{},
			description:         "Valid GTFS feed with all required files",
		},
		{
			name: "missing agency.txt",
			files: map[string]string{
				"stops.txt":      "stop_id,stop_name,stop_lat,stop_lon\n1,Main St,34.05,-118.25",
				"routes.txt":     "route_id,agency_id,route_short_name,route_long_name,route_type\n1,1,1,Main Line,3",
				"trips.txt":      "route_id,service_id,trip_id\n1,S1,T1",
				"stop_times.txt": "trip_id,arrival_time,departure_time,stop_id,stop_sequence\nT1,08:00:00,08:00:00,1,1",
			},
			expectedNoticeCodes: []string{"missing_required_file", "missing_calendar_and_calendar_date_files"},
			description:         "Missing required agency.txt file and calendar files",
		},
		{
			name: "missing multiple required files",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,http://metro.example,America/Los_Angeles",
			},
			expectedNoticeCodes: []string{"missing_required_file", "missing_required_file", "missing_required_file", "missing_required_file", "missing_calendar_and_calendar_date_files"},
			description:         "Missing stops.txt, routes.txt, trips.txt, stop_times.txt, and calendar files",
		},
		{
			name: "missing calendar files",
			files: map[string]string{
				"agency.txt":     "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,http://metro.example,America/Los_Angeles",
				"stops.txt":      "stop_id,stop_name,stop_lat,stop_lon\n1,Main St,34.05,-118.25",
				"routes.txt":     "route_id,agency_id,route_short_name,route_long_name,route_type\n1,1,1,Main Line,3",
				"trips.txt":      "route_id,service_id,trip_id\n1,S1,T1",
				"stop_times.txt": "trip_id,arrival_time,departure_time,stop_id,stop_sequence\nT1,08:00:00,08:00:00,1,1",
			},
			expectedNoticeCodes: []string{"missing_calendar_and_calendar_date_files"},
			description:         "Missing both calendar.txt and calendar_dates.txt",
		},
		{
			name: "calendar_dates.txt only (valid)",
			files: map[string]string{
				"agency.txt":         "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,http://metro.example,America/Los_Angeles",
				"stops.txt":          "stop_id,stop_name,stop_lat,stop_lon\n1,Main St,34.05,-118.25",
				"routes.txt":         "route_id,agency_id,route_short_name,route_long_name,route_type\n1,1,1,Main Line,3",
				"trips.txt":          "route_id,service_id,trip_id\n1,S1,T1",
				"stop_times.txt":     "trip_id,arrival_time,departure_time,stop_id,stop_sequence\nT1,08:00:00,08:00:00,1,1",
				"calendar_dates.txt": "service_id,date,exception_type\nS1,20250101,1",
			},
			expectedNoticeCodes: []string{},
			description:         "Valid with only calendar_dates.txt (no calendar.txt needed)",
		},
		{
			name: "translations.txt without feed_info.txt",
			files: map[string]string{
				"agency.txt":       "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,http://metro.example,America/Los_Angeles",
				"stops.txt":        "stop_id,stop_name,stop_lat,stop_lon\n1,Main St,34.05,-118.25",
				"routes.txt":       "route_id,agency_id,route_short_name,route_long_name,route_type\n1,1,1,Main Line,3",
				"trips.txt":        "route_id,service_id,trip_id\n1,S1,T1",
				"stop_times.txt":   "trip_id,arrival_time,departure_time,stop_id,stop_sequence\nT1,08:00:00,08:00:00,1,1",
				"calendar.txt":     "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\nS1,1,1,1,1,1,0,0,20250101,20251231",
				"translations.txt": "table_name,field_name,language,translation\nstops,stop_name,es,Calle Principal",
			},
			expectedNoticeCodes: []string{"missing_feed_info"},
			description:         "translations.txt requires feed_info.txt",
		},
		{
			name: "fare_rules.txt without fare_attributes.txt",
			files: map[string]string{
				"agency.txt":     "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,http://metro.example,America/Los_Angeles",
				"stops.txt":      "stop_id,stop_name,stop_lat,stop_lon\n1,Main St,34.05,-118.25",
				"routes.txt":     "route_id,agency_id,route_short_name,route_long_name,route_type\n1,1,1,Main Line,3",
				"trips.txt":      "route_id,service_id,trip_id\n1,S1,T1",
				"stop_times.txt": "trip_id,arrival_time,departure_time,stop_id,stop_sequence\nT1,08:00:00,08:00:00,1,1",
				"calendar.txt":   "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\nS1,1,1,1,1,1,0,0,20250101,20251231",
				"fare_rules.txt": "fare_id,route_id\nF1,1",
			},
			expectedNoticeCodes: []string{"missing_fare_attributes"},
			description:         "fare_rules.txt requires fare_attributes.txt",
		},
		{
			name: "pathways.txt without levels.txt",
			files: map[string]string{
				"agency.txt":     "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,http://metro.example,America/Los_Angeles",
				"stops.txt":      "stop_id,stop_name,stop_lat,stop_lon\n1,Main St,34.05,-118.25",
				"routes.txt":     "route_id,agency_id,route_short_name,route_long_name,route_type\n1,1,1,Main Line,3",
				"trips.txt":      "route_id,service_id,trip_id\n1,S1,T1",
				"stop_times.txt": "trip_id,arrival_time,departure_time,stop_id,stop_sequence\nT1,08:00:00,08:00:00,1,1",
				"calendar.txt":   "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\nS1,1,1,1,1,1,0,0,20250101,20251231",
				"pathways.txt":   "pathway_id,from_stop_id,to_stop_id,pathway_mode\nP1,1,2,1",
			},
			expectedNoticeCodes: []string{"missing_levels"},
			description:         "pathways.txt requires levels.txt",
		},
		{
			name: "complete valid feed with optional files",
			files: map[string]string{
				"agency.txt":          "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,http://metro.example,America/Los_Angeles",
				"stops.txt":           "stop_id,stop_name,stop_lat,stop_lon\n1,Main St,34.05,-118.25",
				"routes.txt":          "route_id,agency_id,route_short_name,route_long_name,route_type\n1,1,1,Main Line,3",
				"trips.txt":           "route_id,service_id,trip_id\n1,S1,T1",
				"stop_times.txt":      "trip_id,arrival_time,departure_time,stop_id,stop_sequence\nT1,08:00:00,08:00:00,1,1",
				"calendar.txt":        "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\nS1,1,1,1,1,1,0,0,20250101,20251231",
				"feed_info.txt":       "feed_publisher_name,feed_publisher_url,feed_lang\nMetro,http://metro.example,en",
				"translations.txt":    "table_name,field_name,language,translation\nstops,stop_name,es,Calle Principal",
				"fare_attributes.txt": "fare_id,price,currency_type,payment_method,transfers\nF1,2.50,USD,0,0",
				"fare_rules.txt":      "fare_id,route_id\nF1,1",
				"levels.txt":          "level_id,level_index,level_name\nL1,0,Ground Level",
				"pathways.txt":        "pathway_id,from_stop_id,to_stop_id,pathway_mode\nP1,1,2,1",
			},
			expectedNoticeCodes: []string{},
			description:         "Complete valid feed with all conditional requirements met",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test components
			loader := CreateTestFeedLoader(t, tt.files)
			container := notice.NewNoticeContainer()
			validator := NewMissingFilesValidator()
			config := gtfsvalidator.Config{}

			// Run validation
			validator.Validate(loader, container, config)

			// Get notices
			notices := container.GetNotices()

			// Check notice count
			if len(notices) != len(tt.expectedNoticeCodes) {
				t.Errorf("Expected %d notices, got %d", len(tt.expectedNoticeCodes), len(notices))
			}

			// Check notice codes
			actualCodes := make([]string, len(notices))
			for i, notice := range notices {
				actualCodes[i] = notice.Code()
			}

			// For cases with multiple notices of the same type, just count them
			expectedCodeCounts := make(map[string]int)
			for _, code := range tt.expectedNoticeCodes {
				expectedCodeCounts[code]++
			}

			actualCodeCounts := make(map[string]int)
			for _, code := range actualCodes {
				actualCodeCounts[code]++
			}

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

func TestMissingFilesValidator_New(t *testing.T) {
	validator := NewMissingFilesValidator()
	if validator == nil {
		t.Error("NewMissingFilesValidator() returned nil")
	}
}

func TestMissingFilesValidator_ValidateRequiredFiles(t *testing.T) {
	tests := []struct {
		name            string
		files           map[string]string
		expectedMissing []string
	}{
		{
			name:            "all required files missing",
			files:           map[string]string{},
			expectedMissing: []string{"agency.txt", "stops.txt", "routes.txt", "trips.txt", "stop_times.txt"},
		},
		{
			name: "some required files missing",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,http://metro.example,America/Los_Angeles",
				"stops.txt":  "stop_id,stop_name,stop_lat,stop_lon\n1,Main St,34.05,-118.25",
			},
			expectedMissing: []string{"routes.txt", "trips.txt", "stop_times.txt"},
		},
		{
			name: "all required files present",
			files: map[string]string{
				"agency.txt":     "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,http://metro.example,America/Los_Angeles",
				"stops.txt":      "stop_id,stop_name,stop_lat,stop_lon\n1,Main St,34.05,-118.25",
				"routes.txt":     "route_id,agency_id,route_short_name,route_long_name,route_type\n1,1,1,Main Line,3",
				"trips.txt":      "route_id,service_id,trip_id\n1,S1,T1",
				"stop_times.txt": "trip_id,arrival_time,departure_time,stop_id,stop_sequence\nT1,08:00:00,08:00:00,1,1",
			},
			expectedMissing: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := CreateTestFeedLoader(t, tt.files)
			container := notice.NewNoticeContainer()
			validator := NewMissingFilesValidator()

			validator.validateRequiredFiles(loader, container)

			notices := container.GetNotices()
			missingFileNotices := 0
			for _, notice := range notices {
				if notice.Code() == "missing_required_file" {
					missingFileNotices++
				}
			}

			if missingFileNotices != len(tt.expectedMissing) {
				t.Errorf("Expected %d missing file notices, got %d", len(tt.expectedMissing), missingFileNotices)
			}
		})
	}
}

func TestMissingFilesValidator_ValidateConditionalFiles(t *testing.T) {
	tests := []struct {
		name                string
		files               map[string]string
		expectedNoticeCodes []string
	}{
		{
			name:                "no calendar files",
			files:               map[string]string{},
			expectedNoticeCodes: []string{"missing_calendar_and_calendar_date_files"},
		},
		{
			name: "calendar.txt only",
			files: map[string]string{
				"calendar.txt": "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\nS1,1,1,1,1,1,0,0,20250101,20251231",
			},
			expectedNoticeCodes: []string{},
		},
		{
			name: "calendar_dates.txt only",
			files: map[string]string{
				"calendar_dates.txt": "service_id,date,exception_type\nS1,20250101,1",
			},
			expectedNoticeCodes: []string{},
		},
		{
			name: "both calendar files",
			files: map[string]string{
				"calendar.txt":       "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\nS1,1,1,1,1,1,0,0,20250101,20251231",
				"calendar_dates.txt": "service_id,date,exception_type\nS1,20250101,1",
			},
			expectedNoticeCodes: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := CreateTestFeedLoader(t, tt.files)
			container := notice.NewNoticeContainer()
			validator := NewMissingFilesValidator()

			validator.validateConditionalFiles(loader, container)

			notices := container.GetNotices()
			actualCodes := make([]string, len(notices))
			for i, notice := range notices {
				actualCodes[i] = notice.Code()
			}

			if len(actualCodes) != len(tt.expectedNoticeCodes) {
				t.Errorf("Expected %d notices, got %d. Expected: %v, Actual: %v", len(tt.expectedNoticeCodes), len(actualCodes), tt.expectedNoticeCodes, actualCodes)
			}

			for i, expectedCode := range tt.expectedNoticeCodes {
				if i >= len(actualCodes) || actualCodes[i] != expectedCode {
					t.Errorf("Expected notice code '%s' at index %d, got '%v'", expectedCode, i, actualCodes)
				}
			}
		})
	}
}
