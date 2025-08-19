package core

import (
	"github.com/theoremus-urban-solutions/gtfs-validator/testutil"
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/parser"
	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

func TestInvalidRowValidator_Validate(t *testing.T) {
	tests := []struct {
		name                string
		files               map[string]string
		expectedNoticeCodes []string
		description         string
	}{
		{
			name: "all rows valid",
			files: map[string]string{
				"agency.txt":     "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,http://metro.example,America/Los_Angeles",
				"stops.txt":      "stop_id,stop_name,stop_lat,stop_lon,location_type,wheelchair_boarding\n1,Main St,34.05,-118.25,0,1",
				"routes.txt":     "route_id,agency_id,route_short_name,route_type\n1,1,Red,3",
				"trips.txt":      "route_id,service_id,trip_id,direction_id,wheelchair_accessible,bikes_allowed\n1,S1,T1,0,1,1",
				"stop_times.txt": "trip_id,arrival_time,departure_time,stop_id,stop_sequence\nT1,08:00:00,08:00:00,1,1",
			},
			expectedNoticeCodes: []string{},
			description:         "All rows should be valid",
		},
		{
			name: "wrong number of fields",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,http://metro.example", // Missing timezone
			},
			expectedNoticeCodes: []string{"wrong_number_of_fields"},
			description:         "Row with missing fields should generate notice",
		},
		{
			name: "extra fields in row",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,http://metro.example,America/Los_Angeles,extra_field",
			},
			expectedNoticeCodes: []string{"wrong_number_of_fields"},
			description:         "Row with extra fields should generate notice",
		},
		{
			name: "negative stop_sequence",
			files: map[string]string{
				"stop_times.txt": "trip_id,arrival_time,departure_time,stop_id,stop_sequence\nT1,08:00:00,08:00:00,1,-1",
			},
			expectedNoticeCodes: []string{"negative_stop_sequence"},
			description:         "Negative stop_sequence should generate notice",
		},
		{
			name: "negative shape_dist_traveled",
			files: map[string]string{
				"stop_times.txt": "trip_id,arrival_time,departure_time,stop_id,stop_sequence,shape_dist_traveled\nT1,08:00:00,08:00:00,1,1,-100.5",
			},
			expectedNoticeCodes: []string{"negative_shape_distance"},
			description:         "Negative shape_dist_traveled should generate notice",
		},
		{
			name: "invalid location_type",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_lat,stop_lon,location_type\n1,Main St,34.05,-118.25,5", // Invalid location_type
			},
			expectedNoticeCodes: []string{"invalid_location_type"},
			description:         "Invalid location_type (> 4) should generate notice",
		},
		{
			name: "invalid wheelchair_boarding in stops",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_lat,stop_lon,wheelchair_boarding\n1,Main St,34.05,-118.25,3", // Invalid wheelchair_boarding
			},
			expectedNoticeCodes: []string{"invalid_wheelchair_boarding"},
			description:         "Invalid wheelchair_boarding (> 2) should generate notice",
		},
		{
			name: "invalid route_type",
			files: map[string]string{
				"routes.txt": "route_id,agency_id,route_short_name,route_type\n1,1,Red,99", // Invalid route_type
			},
			expectedNoticeCodes: []string{"invalid_route_type"},
			description:         "Invalid route_type should generate notice",
		},
		{
			name: "valid extended route_type",
			files: map[string]string{
				"routes.txt": "route_id,agency_id,route_short_name,route_type\n1,1,Red,100", // Valid extended route_type
			},
			expectedNoticeCodes: []string{},
			description:         "Extended route_type (100-1700) should be valid",
		},
		{
			name: "invalid direction_id",
			files: map[string]string{
				"trips.txt": "route_id,service_id,trip_id,direction_id\n1,S1,T1,2", // Invalid direction_id
			},
			expectedNoticeCodes: []string{"invalid_direction_id"},
			description:         "Invalid direction_id (> 1) should generate notice",
		},
		{
			name: "invalid wheelchair_accessible in trips",
			files: map[string]string{
				"trips.txt": "route_id,service_id,trip_id,wheelchair_accessible\n1,S1,T1,3", // Invalid wheelchair_accessible
			},
			expectedNoticeCodes: []string{"invalid_wheelchair_accessible"},
			description:         "Invalid wheelchair_accessible (> 2) should generate notice",
		},
		{
			name: "invalid bikes_allowed",
			files: map[string]string{
				"trips.txt": "route_id,service_id,trip_id,bikes_allowed\n1,S1,T1,3", // Invalid bikes_allowed
			},
			expectedNoticeCodes: []string{"invalid_bikes_allowed"},
			description:         "Invalid bikes_allowed (> 2) should generate notice",
		},
		{
			name: "invalid calendar day values",
			files: map[string]string{
				"calendar.txt": "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\nS1,2,1,1,1,1,0,0,20250101,20251231", // Invalid monday value
			},
			expectedNoticeCodes: []string{"invalid_day_value"},
			description:         "Invalid day value (not 0 or 1) should generate notice",
		},
		{
			name: "invalid exception_type",
			files: map[string]string{
				"calendar_dates.txt": "service_id,date,exception_type\nS1,20250101,3", // Invalid exception_type
			},
			expectedNoticeCodes: []string{"invalid_exception_type"},
			description:         "Invalid exception_type (> 2) should generate notice",
		},
		{
			name: "negative shape_pt_sequence",
			files: map[string]string{
				"shapes.txt": "shape_id,shape_pt_lat,shape_pt_lon,shape_pt_sequence\nS1,34.05,-118.25,-1", // Negative sequence
			},
			expectedNoticeCodes: []string{"negative_shape_sequence"},
			description:         "Negative shape_pt_sequence should generate notice",
		},
		{
			name: "invalid payment_method",
			files: map[string]string{
				"fare_attributes.txt": "fare_id,price,currency_type,payment_method,transfers\nF1,2.50,USD,2,0", // Invalid payment_method
			},
			expectedNoticeCodes: []string{"invalid_payment_method"},
			description:         "Invalid payment_method (> 1) should generate notice",
		},
		{
			name: "invalid transfers value",
			files: map[string]string{
				"fare_attributes.txt": "fare_id,price,currency_type,payment_method,transfers\nF1,2.50,USD,0,3", // Invalid transfers
			},
			expectedNoticeCodes: []string{"invalid_transfers"},
			description:         "Invalid transfers (> 2) should generate notice",
		},
		{
			name: "invalid headway_secs",
			files: map[string]string{
				"frequencies.txt": "trip_id,start_time,end_time,headway_secs\nT1,06:00:00,22:00:00,0", // Invalid headway_secs
			},
			expectedNoticeCodes: []string{"invalid_headway"},
			description:         "Invalid headway_secs (<= 0) should generate notice",
		},
		{
			name: "invalid exact_times",
			files: map[string]string{
				"frequencies.txt": "trip_id,start_time,end_time,headway_secs,exact_times\nT1,06:00:00,22:00:00,600,2", // Invalid exact_times
			},
			expectedNoticeCodes: []string{"invalid_exact_times"},
			description:         "Invalid exact_times (> 1) should generate notice",
		},
		{
			name: "invalid transfer_type",
			files: map[string]string{
				"transfers.txt": "from_stop_id,to_stop_id,transfer_type\n1,2,4", // Invalid transfer_type
			},
			expectedNoticeCodes: []string{"invalid_transfer_type"},
			description:         "Invalid transfer_type (> 3) should generate notice",
		},
		{
			name: "multiple invalid values in same row",
			files: map[string]string{
				"trips.txt": "route_id,service_id,trip_id,direction_id,wheelchair_accessible,bikes_allowed\n1,S1,T1,2,3,3", // Multiple invalid values
			},
			expectedNoticeCodes: []string{"invalid_direction_id", "invalid_wheelchair_accessible", "invalid_bikes_allowed"},
			description:         "Multiple invalid values in same row should generate multiple notices",
		},
		{
			name: "multiple invalid rows",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_lat,stop_lon,location_type\n1,Main St,34.05,-118.25,5\n2,Second St,34.06,-118.26,6", // Two invalid location_types
			},
			expectedNoticeCodes: []string{"invalid_location_type", "invalid_location_type"},
			description:         "Multiple invalid rows should generate multiple notices",
		},
		{
			name: "empty optional fields ignored",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_lat,stop_lon,location_type,wheelchair_boarding\n1,Main St,34.05,-118.25,,", // Empty optional fields
			},
			expectedNoticeCodes: []string{},
			description:         "Empty optional fields should not generate validation errors",
		},
		{
			name: "whitespace-only fields ignored",
			files: map[string]string{
				"trips.txt": "route_id,service_id,trip_id,direction_id\n1,S1,T1,   ", // Whitespace direction_id
			},
			expectedNoticeCodes: []string{},
			description:         "Whitespace-only optional fields should not generate validation errors",
		},
		{
			name: "valid boundary values",
			files: map[string]string{
				"stops.txt":     "stop_id,stop_name,stop_lat,stop_lon,location_type,wheelchair_boarding\n1,Main St,34.05,-118.25,4,2", // Max valid values
				"trips.txt":     "route_id,service_id,trip_id,direction_id,wheelchair_accessible,bikes_allowed\n1,S1,T1,1,2,2",
				"transfers.txt": "from_stop_id,to_stop_id,transfer_type\n1,2,3",
				"calendar.txt":  "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\nS1,1,1,1,1,1,1,1,20250101,20251231",
			},
			expectedNoticeCodes: []string{},
			description:         "Maximum valid boundary values should not generate notices",
		},
		{
			name: "minimum valid boundary values",
			files: map[string]string{
				"stops.txt":           "stop_id,stop_name,stop_lat,stop_lon,location_type,wheelchair_boarding\n1,Main St,34.05,-118.25,0,0", // Min valid values
				"trips.txt":           "route_id,service_id,trip_id,direction_id,wheelchair_accessible,bikes_allowed\n1,S1,T1,0,0,0",
				"transfers.txt":       "from_stop_id,to_stop_id,transfer_type\n1,2,0",
				"calendar_dates.txt":  "service_id,date,exception_type\nS1,20250101,1",
				"fare_attributes.txt": "fare_id,price,currency_type,payment_method,transfers\nF1,2.50,USD,0,0",
			},
			expectedNoticeCodes: []string{},
			description:         "Minimum valid boundary values should not generate notices",
		},
		{
			name: "non-numeric values in numeric fields",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_lat,stop_lon,location_type\n1,Main St,34.05,-118.25,invalid", // Non-numeric location_type
			},
			expectedNoticeCodes: []string{},
			description:         "Non-numeric values in numeric fields should be ignored (other validators handle this)",
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
			name: "files without specific validation rules",
			files: map[string]string{
				"attributions.txt": "attribution_id,organization_name\nA1,Metro", // No specific validation rules
			},
			expectedNoticeCodes: []string{},
			description:         "Files without specific validation rules should not generate row content errors",
		},
		{
			name: "valid standard route types",
			files: map[string]string{
				"routes.txt": "route_id,agency_id,route_short_name,route_type\n1,1,Tram,0\n2,1,Subway,1\n3,1,Rail,2\n4,1,Bus,3\n5,1,Ferry,4\n6,1,Cable,5\n7,1,Gondola,6\n8,1,Funicular,7\n9,1,Monorail,12", // All standard types
			},
			expectedNoticeCodes: []string{},
			description:         "All standard GTFS route types should be valid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test components
			loader := testutil.CreateTestFeedLoader(t, tt.files)
			container := notice.NewNoticeContainer()
			validator := NewInvalidRowValidator()
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

func TestInvalidRowValidator_ValidateRowContent(t *testing.T) {
	tests := []struct {
		name            string
		filename        string
		rowData         map[string]string
		expectedNotices []string
		description     string
	}{
		{
			name:            "valid stop_times row",
			filename:        "stop_times.txt",
			rowData:         map[string]string{"trip_id": "T1", "stop_sequence": "1", "shape_dist_traveled": "100.5"},
			expectedNotices: []string{},
			description:     "Valid stop_times row should not generate notices",
		},
		{
			name:            "stop_times with negative stop_sequence",
			filename:        "stop_times.txt",
			rowData:         map[string]string{"trip_id": "T1", "stop_sequence": "-1"},
			expectedNotices: []string{"negative_stop_sequence"},
			description:     "Negative stop_sequence should generate notice",
		},
		{
			name:            "stop_times with negative shape_dist_traveled",
			filename:        "stop_times.txt",
			rowData:         map[string]string{"trip_id": "T1", "stop_sequence": "1", "shape_dist_traveled": "-100.5"},
			expectedNotices: []string{"negative_shape_distance"},
			description:     "Negative shape_dist_traveled should generate notice",
		},
		{
			name:            "valid stops row",
			filename:        "stops.txt",
			rowData:         map[string]string{"stop_id": "S1", "location_type": "2", "wheelchair_boarding": "1"},
			expectedNotices: []string{},
			description:     "Valid stops row should not generate notices",
		},
		{
			name:            "stops with invalid location_type",
			filename:        "stops.txt",
			rowData:         map[string]string{"stop_id": "S1", "location_type": "5"},
			expectedNotices: []string{"invalid_location_type"},
			description:     "Invalid location_type should generate notice",
		},
		{
			name:            "routes with valid extended route_type",
			filename:        "routes.txt",
			rowData:         map[string]string{"route_id": "R1", "route_type": "1500"},
			expectedNotices: []string{},
			description:     "Extended route_type should be valid",
		},
		{
			name:            "routes with invalid route_type",
			filename:        "routes.txt",
			rowData:         map[string]string{"route_id": "R1", "route_type": "99"},
			expectedNotices: []string{"invalid_route_type"},
			description:     "Invalid route_type should generate notice",
		},
		{
			name:            "calendar with valid day values",
			filename:        "calendar.txt",
			rowData:         map[string]string{"service_id": "S1", "monday": "1", "tuesday": "0", "wednesday": "1"},
			expectedNotices: []string{},
			description:     "Valid day values (0 and 1) should not generate notices",
		},
		{
			name:            "calendar with invalid day value",
			filename:        "calendar.txt",
			rowData:         map[string]string{"service_id": "S1", "monday": "2"},
			expectedNotices: []string{"invalid_day_value"},
			description:     "Invalid day value should generate notice",
		},
		{
			name:            "file without specific validation",
			filename:        "agency.txt",
			rowData:         map[string]string{"agency_id": "A1", "agency_name": "Metro"},
			expectedNotices: []string{},
			description:     "Files without specific validation should not generate notices",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			container := notice.NewNoticeContainer()
			validator := NewInvalidRowValidator()

			// Create mock row
			row := &parser.CSVRow{
				RowNumber: 1,
				Values:    tt.rowData,
			}

			validator.validateRowContent(container, tt.filename, row)

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

func TestInvalidRowValidator_ValidateStopTimeRow(t *testing.T) {
	validator := NewInvalidRowValidator()

	tests := []struct {
		name            string
		rowData         map[string]string
		expectedNotices []string
	}{
		{
			name:            "valid stop_sequence and shape_dist",
			rowData:         map[string]string{"trip_id": "T1", "stop_sequence": "1", "shape_dist_traveled": "100.5"},
			expectedNotices: []string{},
		},
		{
			name:            "negative stop_sequence",
			rowData:         map[string]string{"trip_id": "T1", "stop_sequence": "-1"},
			expectedNotices: []string{"negative_stop_sequence"},
		},
		{
			name:            "zero stop_sequence valid",
			rowData:         map[string]string{"trip_id": "T1", "stop_sequence": "0"},
			expectedNotices: []string{},
		},
		{
			name:            "negative shape_dist_traveled",
			rowData:         map[string]string{"trip_id": "T1", "stop_sequence": "1", "shape_dist_traveled": "-100.5"},
			expectedNotices: []string{"negative_shape_distance"},
		},
		{
			name:            "zero shape_dist_traveled valid",
			rowData:         map[string]string{"trip_id": "T1", "stop_sequence": "1", "shape_dist_traveled": "0"},
			expectedNotices: []string{},
		},
		{
			name:            "empty shape_dist_traveled ignored",
			rowData:         map[string]string{"trip_id": "T1", "stop_sequence": "1", "shape_dist_traveled": ""},
			expectedNotices: []string{},
		},
		{
			name:            "non-numeric values ignored",
			rowData:         map[string]string{"trip_id": "T1", "stop_sequence": "invalid", "shape_dist_traveled": "not_a_number"},
			expectedNotices: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			container := notice.NewNoticeContainer()
			row := &parser.CSVRow{RowNumber: 1, Values: tt.rowData}

			validator.validateStopTimeRow(container, row)

			notices := container.GetNotices()
			actualCodes := make([]string, len(notices))
			for i, notice := range notices {
				actualCodes[i] = notice.Code()
			}

			if len(actualCodes) != len(tt.expectedNotices) {
				t.Errorf("Expected %d notices, got %d", len(tt.expectedNotices), len(actualCodes))
			}

			for i, expectedCode := range tt.expectedNotices {
				if i >= len(actualCodes) || actualCodes[i] != expectedCode {
					t.Errorf("Expected notice code '%s' at index %d, got %v", expectedCode, i, actualCodes)
				}
			}
		})
	}
}

func TestInvalidRowValidator_ValidateRouteRow(t *testing.T) {
	validator := NewInvalidRowValidator()

	tests := []struct {
		name        string
		routeType   string
		expectValid bool
		description string
	}{
		// Standard GTFS route types
		{"tram", "0", true, "Tram/Light rail"},
		{"subway", "1", true, "Subway/Metro"},
		{"rail", "2", true, "Rail"},
		{"bus", "3", true, "Bus"},
		{"ferry", "4", true, "Ferry"},
		{"cable_tram", "5", true, "Cable tram"},
		{"aerial_lift", "6", true, "Aerial lift"},
		{"funicular", "7", true, "Funicular"},
		{"trolleybus", "11", true, "Trolleybus"},
		{"monorail", "12", true, "Monorail"},

		// Extended route types (100-1700)
		{"extended_100", "100", true, "Extended type 100"},
		{"extended_1000", "1000", true, "Extended type 1000"},
		{"extended_1700", "1700", true, "Extended type 1700"},

		// Invalid route types
		{"invalid_8", "8", false, "Invalid standard type 8"},
		{"invalid_13", "13", false, "Invalid standard type 13"},
		{"invalid_99", "99", false, "Invalid type 99"},
		{"invalid_1701", "1701", false, "Extended type too high"},
		{"negative", "-1", false, "Negative route type"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			container := notice.NewNoticeContainer()
			rowData := map[string]string{"route_id": "R1", "route_type": tt.routeType}
			row := &parser.CSVRow{RowNumber: 1, Values: rowData}

			validator.validateRouteRow(container, row)

			notices := container.GetNotices()
			hasInvalidNotice := false
			for _, notice := range notices {
				if notice.Code() == "invalid_route_type" {
					hasInvalidNotice = true
					break
				}
			}

			if tt.expectValid && hasInvalidNotice {
				t.Errorf("Expected route_type %s to be valid but got invalid_route_type notice (%s)", tt.routeType, tt.description)
			} else if !tt.expectValid && !hasInvalidNotice {
				t.Errorf("Expected route_type %s to be invalid but got no invalid_route_type notice (%s)", tt.routeType, tt.description)
			}
		})
	}
}

func TestInvalidRowValidator_New(t *testing.T) {
	validator := NewInvalidRowValidator()
	if validator == nil {
		t.Error("NewInvalidRowValidator() returned nil")
	}
}
