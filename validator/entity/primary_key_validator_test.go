package entity

import (
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/parser"
	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

func TestPrimaryKeyValidator_Validate(t *testing.T) {
	tests := []struct {
		name                string
		files               map[string]string
		expectedNoticeCodes []string
		description         string
	}{
		{
			name: "no duplicate keys",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,http://metro.example,America/Los_Angeles\n2,Bus,http://bus.example,America/Los_Angeles",
				"stops.txt":  "stop_id,stop_name,stop_lat,stop_lon\n1,Main St,34.05,-118.25\n2,Second St,34.06,-118.26",
			},
			expectedNoticeCodes: []string{},
			description:         "Files with unique primary keys should not generate notices",
		},
		{
			name: "duplicate agency_id",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,http://metro.example,America/Los_Angeles\n1,Bus,http://bus.example,America/Los_Angeles",
			},
			expectedNoticeCodes: []string{"duplicate_key"},
			description:         "Duplicate single primary key should generate notice",
		},
		{
			name: "duplicate composite key in stop_times",
			files: map[string]string{
				"stop_times.txt": "trip_id,arrival_time,departure_time,stop_id,stop_sequence\nT1,08:00:00,08:00:00,1,1\nT1,08:01:00,08:01:00,2,1",
			},
			expectedNoticeCodes: []string{"duplicate_key"},
			description:         "Duplicate composite key should generate notice",
		},
		{
			name: "multiple duplicate keys",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,http://metro.example,America/Los_Angeles\n1,Bus,http://bus.example,America/Los_Angeles",
				"stops.txt":  "stop_id,stop_name,stop_lat,stop_lon\n1,Main St,34.05,-118.25\n1,First St,34.06,-118.26",
			},
			expectedNoticeCodes: []string{"duplicate_key", "duplicate_key"},
			description:         "Multiple files with duplicates should generate multiple notices",
		},
		{
			name: "valid composite keys with different combinations",
			files: map[string]string{
				"stop_times.txt": "trip_id,arrival_time,departure_time,stop_id,stop_sequence\nT1,08:00:00,08:00:00,1,1\nT1,08:01:00,08:01:00,2,2\nT2,08:00:00,08:00:00,1,1",
			},
			expectedNoticeCodes: []string{},
			description:         "Different composite key combinations should be valid",
		},
		{
			name: "file without primary key definition",
			files: map[string]string{
				"unknown_file.txt": "field1,field2\nvalue1,value2\nvalue1,value3",
			},
			expectedNoticeCodes: []string{},
			description:         "Files without primary key definition should not generate notices",
		},
		{
			name: "empty files",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone",
			},
			expectedNoticeCodes: []string{},
			description:         "Empty files should not generate notices",
		},
		{
			name: "triplicate primary keys",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,http://metro.example,America/Los_Angeles\n1,Bus,http://bus.example,America/Los_Angeles\n1,Rail,http://rail.example,America/Los_Angeles",
			},
			expectedNoticeCodes: []string{"duplicate_key", "duplicate_key"},
			description:         "Three instances of same key should generate two duplicate notices",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := CreateTestFeedLoader(t, tt.files)
			container := notice.NewNoticeContainer()
			validator := NewPrimaryKeyValidator()
			config := gtfsvalidator.Config{}

			validator.Validate(loader, container, config)

			notices := container.GetNotices()

			if len(notices) != len(tt.expectedNoticeCodes) {
				t.Errorf("Expected %d notices, got %d for case: %s", len(tt.expectedNoticeCodes), len(notices), tt.description)
			}

			expectedCodeCounts := make(map[string]int)
			for _, code := range tt.expectedNoticeCodes {
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

func TestPrimaryKeyValidator_GetPrimaryKeyFields(t *testing.T) {
	validator := NewPrimaryKeyValidator()

	tests := []struct {
		filename     string
		expectedKeys []string
		description  string
	}{
		{"agency.txt", []string{"agency_id"}, "Agency should have single key"},
		{"stops.txt", []string{"stop_id"}, "Stops should have single key"},
		{"stop_times.txt", []string{"trip_id", "stop_sequence"}, "Stop times should have composite key"},
		{"calendar_dates.txt", []string{"service_id", "date"}, "Calendar dates should have composite key"},
		{"unknown_file.txt", []string{}, "Unknown files should have no key"},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			keys := validator.getPrimaryKeyFields(tt.filename)

			if len(keys) != len(tt.expectedKeys) {
				t.Errorf("Expected %d keys for %s, got %d", len(tt.expectedKeys), tt.filename, len(keys))
			}

			for i, expectedKey := range tt.expectedKeys {
				if i >= len(keys) || keys[i] != expectedKey {
					t.Errorf("Expected key '%s' at index %d for %s, got %v", expectedKey, i, tt.filename, keys)
				}
			}
		})
	}
}

func TestPrimaryKeyValidator_BuildCompositeKey(t *testing.T) {
	validator := NewPrimaryKeyValidator()

	tests := []struct {
		name        string
		rowValues   map[string]string
		fields      []string
		expected    string
		description string
	}{
		{
			name:        "single field key",
			rowValues:   map[string]string{"agency_id": "1", "agency_name": "Metro"},
			fields:      []string{"agency_id"},
			expected:    "1",
			description: "Single field should return field value",
		},
		{
			name:        "composite key two fields",
			rowValues:   map[string]string{"trip_id": "T1", "stop_sequence": "1"},
			fields:      []string{"trip_id", "stop_sequence"},
			expected:    "T1|1",
			description: "Two fields should be joined with pipe",
		},
		{
			name:        "composite key multiple fields",
			rowValues:   map[string]string{"fare_id": "F1", "route_id": "R1", "origin_id": "O1"},
			fields:      []string{"fare_id", "route_id", "origin_id"},
			expected:    "F1|R1|O1",
			description: "Multiple fields should be joined with pipes",
		},
		{
			name:        "empty field values",
			rowValues:   map[string]string{"trip_id": "", "stop_sequence": "1"},
			fields:      []string{"trip_id", "stop_sequence"},
			expected:    "|1",
			description: "Empty values should be preserved in composite key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock row
			row := &parser.CSVRow{
				RowNumber: 1,
				Values:    tt.rowValues,
			}

			result := validator.buildCompositeKey(row, tt.fields)
			if result != tt.expected {
				t.Errorf("Expected composite key '%s', got '%s' for %s", tt.expected, result, tt.description)
			}
		})
	}
}

func TestPrimaryKeyValidator_New(t *testing.T) {
	validator := NewPrimaryKeyValidator()
	if validator == nil {
		t.Error("NewPrimaryKeyValidator() returned nil")
	}
}
