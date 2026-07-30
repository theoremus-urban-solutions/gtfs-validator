package entity

import (
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/testutil"
)

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
