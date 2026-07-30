package entity

import (
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/parser"
)

func TestTripPatternValidator_ParseStopTime(t *testing.T) {
	validator := &TripPatternValidator{}

	tests := []struct {
		name     string
		row      *parser.CSVRow
		expected *TripStopTime
	}{
		{
			name: "valid stop time",
			row: &parser.CSVRow{
				Values: map[string]string{
					"trip_id":        "trip_1",
					"stop_id":        "stop_1",
					"stop_sequence":  "1",
					"arrival_time":   "08:00:00",
					"departure_time": "08:00:00",
				},
				RowNumber: 2,
			},
			expected: &TripStopTime{
				TripID:        "trip_1",
				StopID:        "stop_1",
				StopSequence:  1,
				ArrivalTime:   "08:00:00",
				DepartureTime: "08:00:00",
				RowNumber:     2,
			},
		},
		{
			name: "stop time with whitespace",
			row: &parser.CSVRow{
				Values: map[string]string{
					"trip_id":       "  trip_1  ",
					"stop_id":       "  stop_1  ",
					"stop_sequence": "  1  ",
				},
				RowNumber: 2,
			},
			expected: &TripStopTime{
				TripID:       "trip_1",
				StopID:       "stop_1",
				StopSequence: 1,
				RowNumber:    2,
			},
		},
		{
			name: "missing required field",
			row: &parser.CSVRow{
				Values: map[string]string{
					"trip_id": "trip_1",
					"stop_id": "stop_1",
				},
				RowNumber: 2,
			},
			expected: nil,
		},
		{
			name: "invalid stop sequence",
			row: &parser.CSVRow{
				Values: map[string]string{
					"trip_id":       "trip_1",
					"stop_id":       "stop_1",
					"stop_sequence": "not_a_number",
				},
				RowNumber: 2,
			},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.parseStopTime(tt.row)

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
				result.StopID != tt.expected.StopID ||
				result.StopSequence != tt.expected.StopSequence ||
				result.ArrivalTime != tt.expected.ArrivalTime ||
				result.DepartureTime != tt.expected.DepartureTime ||
				result.RowNumber != tt.expected.RowNumber {
				t.Errorf("Expected %+v, got %+v", tt.expected, result)
			}
		})
	}
}

func TestTripPatternValidator_GroupStopTimesByTrip(t *testing.T) {
	validator := &TripPatternValidator{}

	stopTimes := []*TripStopTime{
		{TripID: "trip_1", StopID: "stop_1", StopSequence: 2},
		{TripID: "trip_1", StopID: "stop_2", StopSequence: 1},
		{TripID: "trip_1", StopID: "stop_3", StopSequence: 3},
		{TripID: "trip_2", StopID: "stop_1", StopSequence: 1},
		{TripID: "trip_2", StopID: "stop_2", StopSequence: 2},
	}

	grouped := validator.groupStopTimesByTrip(stopTimes)

	// Check trip_1 is sorted by stop_sequence
	if len(grouped["trip_1"]) != 3 {
		t.Errorf("Expected 3 stops for trip_1, got %d", len(grouped["trip_1"]))
	}
	if grouped["trip_1"][0].StopSequence != 1 {
		t.Errorf("Expected first stop sequence to be 1, got %d", grouped["trip_1"][0].StopSequence)
	}
	if grouped["trip_1"][1].StopSequence != 2 {
		t.Errorf("Expected second stop sequence to be 2, got %d", grouped["trip_1"][1].StopSequence)
	}
	if grouped["trip_1"][2].StopSequence != 3 {
		t.Errorf("Expected third stop sequence to be 3, got %d", grouped["trip_1"][2].StopSequence)
	}

	// Check trip_2
	if len(grouped["trip_2"]) != 2 {
		t.Errorf("Expected 2 stops for trip_2, got %d", len(grouped["trip_2"]))
	}
}
