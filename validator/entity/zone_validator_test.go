package entity

import (
	"github.com/theoremus-urban-solutions/gtfs-validator/testutil"
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

func TestZoneValidator_Validate(t *testing.T) {
	tests := []struct {
		name                string
		files               map[string]string
		expectedNoticeCodes []string
		description         string
	}{
		{
			name: "valid zones with multiple stops and usage",
			files: map[string]string{
				"stops.txt":      "stop_id,stop_name,stop_lat,stop_lon,zone_id\n1,Stop A,34.05,-118.25,Z1\n2,Stop B,34.06,-118.26,Z1\n3,Stop C,34.07,-118.27,Z2",
				"fare_rules.txt": "fare_id,origin_id,destination_id\nF1,Z1,Z2",
			},
			expectedNoticeCodes: []string{"single_stop_zone"},
			description:         "Valid zones with multiple stops and proper usage should not generate notices",
		},
		{
			name: "single stop zone",
			files: map[string]string{
				"stops.txt":      "stop_id,stop_name,stop_lat,stop_lon,zone_id\n1,Stop A,34.05,-118.25,Z1\n2,Stop B,34.06,-118.26,Z2",
				"fare_rules.txt": "fare_id,origin_id,destination_id\nF1,Z1,Z2",
			},
			expectedNoticeCodes: []string{"single_stop_zone", "single_stop_zone"},
			description:         "Zones with single stops should generate notices",
		},
		{
			name: "unused zone",
			files: map[string]string{
				"stops.txt":      "stop_id,stop_name,stop_lat,stop_lon,zone_id\n1,Stop A,34.05,-118.25,Z1\n2,Stop B,34.06,-118.26,Z2",
				"fare_rules.txt": "fare_id,origin_id,destination_id\nF1,Z1,Z1", // Z2 not used
			},
			expectedNoticeCodes: []string{"single_stop_zone", "single_stop_zone", "unused_zone"},
			description:         "Zones defined but not used in fare rules should generate notices",
		},
		{
			name: "undefined zone referenced",
			files: map[string]string{
				"stops.txt":      "stop_id,stop_name,stop_lat,stop_lon,zone_id\n1,Stop A,34.05,-118.25,Z1",
				"fare_rules.txt": "fare_id,origin_id,destination_id\nF1,Z1,Z2", // Z2 not defined
			},
			expectedNoticeCodes: []string{"single_stop_zone", "undefined_zone"},
			description:         "Zones referenced but not defined should generate notices",
		},
		{
			name: "zone_id same as stop_id",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_lat,stop_lon,zone_id\nSTOP_001,Stop A,34.05,-118.25,STOP_001",
			},
			expectedNoticeCodes: []string{"single_stop_zone", "unused_zone", "zone_id_same_as_stop_id"},
			description:         "Zone ID same as stop ID should generate notices",
		},
		{
			name: "long zone_id",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_lat,stop_lon,zone_id\n1,Stop A,34.05,-118.25,THIS_IS_A_VERY_LONG_ZONE_ID_THAT_EXCEEDS_FIFTY_CHARACTERS_AND_SHOULD_TRIGGER_A_NOTICE",
			},
			expectedNoticeCodes: []string{"single_stop_zone", "unused_zone", "long_zone_id"},
			description:         "Very long zone IDs should generate notices",
		},
		{
			name: "empty zone_id ignored",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_lat,stop_lon,zone_id\n1,Stop A,34.05,-118.25,\n2,Stop B,34.06,-118.26,Z1",
			},
			expectedNoticeCodes: []string{"single_stop_zone", "unused_zone"},
			description:         "Empty zone_id should be ignored",
		},
		{
			name: "whitespace zone_id ignored",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_lat,stop_lon,zone_id\n1,Stop A,34.05,-118.25,   \n2,Stop B,34.06,-118.26,Z1",
			},
			expectedNoticeCodes: []string{"single_stop_zone", "unused_zone"},
			description:         "Whitespace-only zone_id should be ignored",
		},
		{
			name: "zone_id with whitespace trimmed",
			files: map[string]string{
				"stops.txt":      "stop_id,stop_name,stop_lat,stop_lon,zone_id\n1,Stop A,34.05,-118.25, Z1 \n2,Stop B,34.06,-118.26, Z1 ",
				"fare_rules.txt": "fare_id,origin_id,destination_id\nF1, Z1 , Z1 ",
			},
			expectedNoticeCodes: []string{},
			description:         "Zone IDs with whitespace should be properly trimmed",
		},
		{
			name: "multiple zone types in fare_rules",
			files: map[string]string{
				"stops.txt":      "stop_id,stop_name,stop_lat,stop_lon,zone_id\n1,Stop A,34.05,-118.25,Z1\n2,Stop B,34.06,-118.26,Z2\n3,Stop C,34.07,-118.27,Z3",
				"fare_rules.txt": "fare_id,origin_id,destination_id,contains_id\nF1,Z1,Z2,Z3",
			},
			expectedNoticeCodes: []string{"single_stop_zone", "single_stop_zone", "single_stop_zone"},
			description:         "All zone reference fields in fare_rules should be checked",
		},
		{
			name: "no stops.txt file",
			files: map[string]string{
				"fare_rules.txt": "fare_id,origin_id,destination_id\nF1,Z1,Z2",
			},
			expectedNoticeCodes: []string{"undefined_zone", "undefined_zone"},
			description:         "Missing stops.txt should still check fare rule zone references",
		},
		{
			name: "no fare_rules.txt file",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_lat,stop_lon,zone_id\n1,Stop A,34.05,-118.25,Z1\n2,Stop B,34.06,-118.26,Z2",
			},
			expectedNoticeCodes: []string{"single_stop_zone", "single_stop_zone", "unused_zone", "unused_zone"},
			description:         "Missing fare_rules.txt should mark all zones as unused",
		},
		{
			name: "stops without zone_id field",
			files: map[string]string{
				"stops.txt":      "stop_id,stop_name,stop_lat,stop_lon\n1,Stop A,34.05,-118.25\n2,Stop B,34.06,-118.26",
				"fare_rules.txt": "fare_id,origin_id,destination_id\nF1,Z1,Z2",
			},
			expectedNoticeCodes: []string{"undefined_zone", "undefined_zone"},
			description:         "Stops without zone_id field should not define zones",
		},
		{
			name: "mixed case zone IDs",
			files: map[string]string{
				"stops.txt":      "stop_id,stop_name,stop_lat,stop_lon,zone_id\n1,Stop A,34.05,-118.25,z1\n2,Stop B,34.06,-118.26,z1",
				"fare_rules.txt": "fare_id,origin_id,destination_id\nF1,Z1,z1", // Different case
			},
			expectedNoticeCodes: []string{"undefined_zone"},
			description:         "Zone ID matching should be case sensitive",
		},
		{
			name: "empty fare_rules values ignored",
			files: map[string]string{
				"stops.txt":      "stop_id,stop_name,stop_lat,stop_lon,zone_id\n1,Stop A,34.05,-118.25,Z1",
				"fare_rules.txt": "fare_id,origin_id,destination_id,contains_id\nF1,,Z1,", // Empty origin and contains
			},
			expectedNoticeCodes: []string{"single_stop_zone"},
			description:         "Empty fare rule zone fields should be ignored",
		},
		{
			name: "complex zone usage scenario",
			files: map[string]string{
				"stops.txt":      "stop_id,stop_name,stop_lat,stop_lon,zone_id\n1,Stop A,34.05,-118.25,DOWNTOWN\n2,Stop B,34.06,-118.26,DOWNTOWN\n3,Stop C,34.07,-118.27,AIRPORT\n4,Stop D,34.08,-118.28,UNUSED_ZONE",
				"fare_rules.txt": "fare_id,route_id,origin_id,destination_id,contains_id\nF1,,DOWNTOWN,AIRPORT,\nF2,R1,,,DOWNTOWN",
			},
			expectedNoticeCodes: []string{"single_stop_zone", "single_stop_zone", "unused_zone"},
			description:         "Complex scenario with multiple zone usages",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := testutil.CreateTestFeedLoader(t, tt.files)
			container := notice.NewNoticeContainer()
			validator := NewZoneValidator()
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

func TestZoneValidator_LoadZones(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		expected    map[string][]*ZoneInfo
		description string
	}{
		{
			name:    "single zone multiple stops",
			content: "stop_id,stop_name,stop_lat,stop_lon,zone_id\n1,Stop A,34.05,-118.25,Z1\n2,Stop B,34.06,-118.26,Z1",
			expected: map[string][]*ZoneInfo{
				"Z1": {
					{ZoneID: "Z1", StopID: "1", RowNumber: 2},
					{ZoneID: "Z1", StopID: "2", RowNumber: 3},
				},
			},
			description: "Multiple stops in same zone should be grouped",
		},
		{
			name:    "multiple zones",
			content: "stop_id,stop_name,stop_lat,stop_lon,zone_id\n1,Stop A,34.05,-118.25,Z1\n2,Stop B,34.06,-118.26,Z2",
			expected: map[string][]*ZoneInfo{
				"Z1": {{ZoneID: "Z1", StopID: "1", RowNumber: 2}},
				"Z2": {{ZoneID: "Z2", StopID: "2", RowNumber: 3}},
			},
			description: "Different zones should be separate",
		},
		{
			name:        "no zone_id field",
			content:     "stop_id,stop_name,stop_lat,stop_lon\n1,Stop A,34.05,-118.25\n2,Stop B,34.06,-118.26",
			expected:    map[string][]*ZoneInfo{},
			description: "Missing zone_id field should result in empty zones",
		},
		{
			name:        "empty zones ignored",
			content:     "stop_id,stop_name,stop_lat,stop_lon,zone_id\n1,Stop A,34.05,-118.25,\n2,Stop B,34.06,-118.26,   ",
			expected:    map[string][]*ZoneInfo{},
			description: "Empty or whitespace zone_ids should be ignored",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files := map[string]string{"stops.txt": tt.content}
			loader := testutil.CreateTestFeedLoader(t, files)
			validator := NewZoneValidator()

			zones := validator.loadZones(loader)

			if len(zones) != len(tt.expected) {
				t.Errorf("Expected %d zones, got %d", len(tt.expected), len(zones))
			}

			for zoneID, expectedInfos := range tt.expected {
				actualInfos, exists := zones[zoneID]
				if !exists {
					t.Errorf("Expected zone '%s' not found", zoneID)
					continue
				}

				if len(actualInfos) != len(expectedInfos) {
					t.Errorf("Expected %d stops in zone '%s', got %d", len(expectedInfos), zoneID, len(actualInfos))
					continue
				}

				for i, expectedInfo := range expectedInfos {
					if i >= len(actualInfos) {
						t.Errorf("Missing stop info at index %d for zone '%s'", i, zoneID)
						continue
					}

					actualInfo := actualInfos[i]
					if actualInfo.ZoneID != expectedInfo.ZoneID {
						t.Errorf("Expected ZoneID '%s', got '%s'", expectedInfo.ZoneID, actualInfo.ZoneID)
					}
					if actualInfo.StopID != expectedInfo.StopID {
						t.Errorf("Expected StopID '%s', got '%s'", expectedInfo.StopID, actualInfo.StopID)
					}
					if actualInfo.RowNumber != expectedInfo.RowNumber {
						t.Errorf("Expected RowNumber %d, got %d", expectedInfo.RowNumber, actualInfo.RowNumber)
					}
				}
			}
		})
	}
}

func TestZoneValidator_LoadUsedZones(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		expected    map[string]bool
		description string
	}{
		{
			name:        "origin and destination zones",
			content:     "fare_id,origin_id,destination_id\nF1,Z1,Z2",
			expected:    map[string]bool{"Z1": true, "Z2": true},
			description: "Origin and destination zones should be marked as used",
		},
		{
			name:        "contains_id zones",
			content:     "fare_id,contains_id\nF1,Z1\nF2,Z2",
			expected:    map[string]bool{"Z1": true, "Z2": true},
			description: "Contains zones should be marked as used",
		},
		{
			name:        "all zone field types",
			content:     "fare_id,origin_id,destination_id,contains_id\nF1,Z1,Z2,Z3",
			expected:    map[string]bool{"Z1": true, "Z2": true, "Z3": true},
			description: "All zone field types should be checked",
		},
		{
			name:        "empty values ignored",
			content:     "fare_id,origin_id,destination_id,contains_id\nF1,,Z2,",
			expected:    map[string]bool{"Z2": true},
			description: "Empty zone fields should be ignored",
		},
		{
			name:        "no fare_rules file",
			content:     "",
			expected:    map[string]bool{},
			description: "Missing fare_rules should result in no used zones",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files := map[string]string{}
			if tt.content != "" {
				files["fare_rules.txt"] = tt.content
			}
			loader := testutil.CreateTestFeedLoader(t, files)
			validator := NewZoneValidator()

			usedZones := validator.loadUsedZones(loader)

			if len(usedZones) != len(tt.expected) {
				t.Errorf("Expected %d used zones, got %d", len(tt.expected), len(usedZones))
			}

			for zoneID, expectedUsed := range tt.expected {
				actualUsed := usedZones[zoneID]
				if actualUsed != expectedUsed {
					t.Errorf("Expected zone '%s' used=%v, got used=%v", zoneID, expectedUsed, actualUsed)
				}
			}
		})
	}
}

func TestZoneValidator_New(t *testing.T) {
	validator := NewZoneValidator()
	if validator == nil {
		t.Error("NewZoneValidator() returned nil")
	}
}
