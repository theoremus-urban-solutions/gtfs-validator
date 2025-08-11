package business

import (
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/parser"
)

func TestOverlappingFrequencyValidator_ParseFrequency(t *testing.T) {
	validator := &OverlappingFrequencyValidator{}

	tests := []struct {
		name     string
		row      *parser.CSVRow
		expected *FrequencyEntry
	}{
		{
			name: "valid frequency entry",
			row: &parser.CSVRow{
				Values: map[string]string{
					"trip_id":      "trip_1",
					"start_time":   "08:00:00",
					"end_time":     "10:00:00",
					"headway_secs": "600",
					"exact_times":  "0",
				},
				RowNumber: 2,
			},
			expected: &FrequencyEntry{
				TripID:      "trip_1",
				StartTime:   28800,
				EndTime:     36000,
				HeadwaySecs: 600,
				ExactTimes:  0,
				RowNumber:   2,
			},
		},
		{
			name: "frequency without exact_times",
			row: &parser.CSVRow{
				Values: map[string]string{
					"trip_id":      "trip_1",
					"start_time":   "08:00:00",
					"end_time":     "10:00:00",
					"headway_secs": "600",
				},
				RowNumber: 2,
			},
			expected: &FrequencyEntry{
				TripID:      "trip_1",
				StartTime:   28800,
				EndTime:     36000,
				HeadwaySecs: 600,
				ExactTimes:  0,
				RowNumber:   2,
			},
		},
		{
			name: "frequency with whitespace",
			row: &parser.CSVRow{
				Values: map[string]string{
					"trip_id":      "  trip_1  ",
					"start_time":   "  08:00:00  ",
					"end_time":     "  10:00:00  ",
					"headway_secs": "  600  ",
				},
				RowNumber: 2,
			},
			expected: &FrequencyEntry{
				TripID:      "trip_1",
				StartTime:   28800,
				EndTime:     36000,
				HeadwaySecs: 600,
				ExactTimes:  0,
				RowNumber:   2,
			},
		},
		{
			name: "invalid time format",
			row: &parser.CSVRow{
				Values: map[string]string{
					"trip_id":      "trip_1",
					"start_time":   "invalid",
					"end_time":     "10:00:00",
					"headway_secs": "600",
				},
				RowNumber: 2,
			},
			expected: nil,
		},
		{
			name: "invalid headway",
			row: &parser.CSVRow{
				Values: map[string]string{
					"trip_id":      "trip_1",
					"start_time":   "08:00:00",
					"end_time":     "10:00:00",
					"headway_secs": "not_a_number",
				},
				RowNumber: 2,
			},
			expected: nil,
		},
		{
			name: "missing required field",
			row: &parser.CSVRow{
				Values: map[string]string{
					"trip_id":    "trip_1",
					"start_time": "08:00:00",
					"end_time":   "10:00:00",
				},
				RowNumber: 2,
			},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.parseFrequency(tt.row)

			if tt.expected == nil {
				if result != nil {
					t.Errorf("Expected nil, got %+v", result)
				}
				return
			}

			if result == nil {
				t.Errorf("Expected %+v, got nil", tt.expected)
				return
			}

			if result.TripID != tt.expected.TripID ||
				result.StartTime != tt.expected.StartTime ||
				result.EndTime != tt.expected.EndTime ||
				result.HeadwaySecs != tt.expected.HeadwaySecs ||
				result.ExactTimes != tt.expected.ExactTimes ||
				result.RowNumber != tt.expected.RowNumber {
				t.Errorf("Expected %+v, got %+v", tt.expected, result)
			}
		})
	}
}

func TestOverlappingFrequencyValidator_DoFrequenciesOverlap(t *testing.T) {
	validator := &OverlappingFrequencyValidator{}

	tests := []struct {
		name     string
		freq1    *FrequencyEntry
		freq2    *FrequencyEntry
		expected bool
	}{
		{
			name: "non-overlapping frequencies",
			freq1: &FrequencyEntry{
				StartTime: 28800, // 08:00:00
				EndTime:   36000, // 10:00:00
			},
			freq2: &FrequencyEntry{
				StartTime: 36000, // 10:00:00
				EndTime:   43200, // 12:00:00
			},
			expected: false,
		},
		{
			name: "overlapping frequencies",
			freq1: &FrequencyEntry{
				StartTime: 28800, // 08:00:00
				EndTime:   36000, // 10:00:00
			},
			freq2: &FrequencyEntry{
				StartTime: 32400, // 09:00:00
				EndTime:   39600, // 11:00:00
			},
			expected: true,
		},
		{
			name: "contained frequency",
			freq1: &FrequencyEntry{
				StartTime: 28800, // 08:00:00
				EndTime:   43200, // 12:00:00
			},
			freq2: &FrequencyEntry{
				StartTime: 32400, // 09:00:00
				EndTime:   36000, // 10:00:00
			},
			expected: true,
		},
		{
			name: "identical frequencies",
			freq1: &FrequencyEntry{
				StartTime: 28800, // 08:00:00
				EndTime:   36000, // 10:00:00
			},
			freq2: &FrequencyEntry{
				StartTime: 28800, // 08:00:00
				EndTime:   36000, // 10:00:00
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.doFrequenciesOverlap(tt.freq1, tt.freq2)
			if result != tt.expected {
				t.Errorf("Expected overlap=%v for [%d-%d] and [%d-%d], got %v",
					tt.expected, tt.freq1.StartTime, tt.freq1.EndTime,
					tt.freq2.StartTime, tt.freq2.EndTime, result)
			}
		})
	}
}

func TestOverlappingFrequencyValidator_ParseGTFSTime(t *testing.T) {
	validator := &OverlappingFrequencyValidator{}

	tests := []struct {
		name     string
		timeStr  string
		expected int
	}{
		{"valid morning time", "08:30:15", 30615},
		{"valid afternoon time", "14:45:00", 53100},
		{"midnight", "00:00:00", 0},
		{"late night service", "25:30:00", 91800},
		{"invalid format - missing seconds", "08:30", -1},
		{"invalid format - not a time", "invalid", -1},
		{"invalid minutes", "08:65:00", -1},
		{"invalid seconds", "08:30:65", -1},
		{"empty string", "", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.parseGTFSTime(tt.timeStr)
			if result != tt.expected {
				t.Errorf("Expected %d seconds for time %s, got %d", tt.expected, tt.timeStr, result)
			}
		})
	}
}
