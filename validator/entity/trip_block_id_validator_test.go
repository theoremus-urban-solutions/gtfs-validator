package entity

import (
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/parser"
	"github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

func TestTripBlockIdValidator_Validate(t *testing.T) {
	tests := []struct {
		name         string
		files        map[string]string
		expectedCode []string
		notExpected  []string
	}{
		{
			name: "valid block with consistent service",
			files: map[string]string{
				"trips.txt": `trip_id,route_id,service_id,block_id
trip_1,route_1,service_1,block_1
trip_2,route_1,service_1,block_1`,
				"stop_times.txt": `trip_id,stop_id,stop_sequence,arrival_time,departure_time
trip_1,stop_1,1,08:00:00,08:00:00
trip_1,stop_2,2,08:30:00,08:30:00
trip_2,stop_2,1,09:00:00,09:00:00
trip_2,stop_3,2,09:30:00,09:30:00`,
			},
			expectedCode: []string{},
			notExpected:  []string{"block_service_mismatch", "block_trips_overlap", "single_trip_block"},
		},
		{
			name: "block with service mismatch",
			files: map[string]string{
				"trips.txt": `trip_id,route_id,service_id,block_id
trip_1,route_1,service_1,block_1
trip_2,route_1,service_2,block_1`,
			},
			expectedCode: []string{"block_service_mismatch"},
			notExpected:  []string{},
		},
		{
			name: "single trip in block",
			files: map[string]string{
				"trips.txt": `trip_id,route_id,service_id,block_id
trip_1,route_1,service_1,block_1
trip_2,route_1,service_1,block_2`,
			},
			expectedCode: []string{"single_trip_block"},
			notExpected:  []string{"block_service_mismatch"},
		},
		{
			name: "block with multiple routes",
			files: map[string]string{
				"trips.txt": `trip_id,route_id,service_id,block_id
trip_1,route_1,service_1,block_1
trip_2,route_2,service_1,block_1
trip_3,route_3,service_1,block_1`,
			},
			expectedCode: []string{"block_multiple_routes"},
			notExpected:  []string{"block_service_mismatch"},
		},
		{
			name: "block with too many trips",
			files: map[string]string{
				"trips.txt": `trip_id,route_id,service_id,block_id
trip_1,route_1,service_1,block_1
trip_2,route_1,service_1,block_1
trip_3,route_1,service_1,block_1
trip_4,route_1,service_1,block_1
trip_5,route_1,service_1,block_1
trip_6,route_1,service_1,block_1
trip_7,route_1,service_1,block_1
trip_8,route_1,service_1,block_1
trip_9,route_1,service_1,block_1
trip_10,route_1,service_1,block_1
trip_11,route_1,service_1,block_1
trip_12,route_1,service_1,block_1
trip_13,route_1,service_1,block_1
trip_14,route_1,service_1,block_1
trip_15,route_1,service_1,block_1
trip_16,route_1,service_1,block_1
trip_17,route_1,service_1,block_1
trip_18,route_1,service_1,block_1
trip_19,route_1,service_1,block_1
trip_20,route_1,service_1,block_1
trip_21,route_1,service_1,block_1`,
			},
			expectedCode: []string{"block_too_many_trips"},
			notExpected:  []string{"block_service_mismatch"},
		},
		{
			name: "block with overlapping trips",
			files: map[string]string{
				"trips.txt": `trip_id,route_id,service_id,block_id
trip_1,route_1,service_1,block_1
trip_2,route_1,service_1,block_1`,
				"stop_times.txt": `trip_id,stop_id,stop_sequence,arrival_time,departure_time
trip_1,stop_1,1,08:00:00,08:00:00
trip_1,stop_2,2,08:30:00,08:30:00
trip_2,stop_2,1,08:15:00,08:15:00
trip_2,stop_3,2,08:45:00,08:45:00`,
			},
			expectedCode: []string{"block_trips_overlap"},
			notExpected:  []string{"block_service_mismatch"},
		},
		{
			name: "trips without blocks",
			files: map[string]string{
				"trips.txt": `trip_id,route_id,service_id,block_id
trip_1,route_1,service_1,
trip_2,route_1,service_1,`,
			},
			expectedCode: []string{},
			notExpected:  []string{"single_trip_block", "block_service_mismatch"},
		},
		{
			name: "empty trips file",
			files: map[string]string{
				"trips.txt": `trip_id,route_id,service_id,block_id`,
			},
			expectedCode: []string{},
			notExpected:  []string{"single_trip_block", "block_service_mismatch"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := CreateTestFeedLoader(t, tt.files)
			container := notice.NewNoticeContainer()
			v := NewTripBlockIdValidator()
			config := validator.Config{}

			v.Validate(loader, container, config)

			notices := container.GetNotices()
			noticeMap := make(map[string]bool)
			for _, n := range notices {
				noticeMap[n.Code()] = true
			}

			// Check expected notices
			for _, code := range tt.expectedCode {
				if !noticeMap[code] {
					t.Errorf("Expected notice with code %s, but not found", code)
				}
			}

			// Check unexpected notices
			for _, code := range tt.notExpected {
				if noticeMap[code] {
					t.Errorf("Did not expect notice with code %s, but found it", code)
				}
			}
		})
	}
}

func TestTripBlockIdValidator_ParseGTFSTime(t *testing.T) {
	validator := &TripBlockIdValidator{}

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

func TestTripBlockIdValidator_FormatGTFSTime(t *testing.T) {
	validator := &TripBlockIdValidator{}

	tests := []struct {
		name     string
		seconds  int
		expected string
	}{
		{"morning time", 30615, "8:30:15"},
		{"afternoon time", 53100, "14:45:0"},
		{"midnight", 0, "0:0:0"},
		{"late night service", 91800, "25:30:0"},
		{"negative time", -1, "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.formatGTFSTime(tt.seconds)
			if result != tt.expected {
				t.Errorf("Expected %s for %d seconds, got %s", tt.expected, tt.seconds, result)
			}
		})
	}
}

func TestTripBlockIdValidator_DoTripsOverlap(t *testing.T) {
	validator := &TripBlockIdValidator{}

	tests := []struct {
		name     string
		start1   int
		end1     int
		start2   int
		end2     int
		expected bool
	}{
		{"no overlap - trip2 after trip1", 100, 200, 300, 400, false},
		{"no overlap - trip1 after trip2", 300, 400, 100, 200, false},
		{"overlap - partial", 100, 300, 200, 400, true},
		{"overlap - trip2 within trip1", 100, 400, 200, 300, true},
		{"overlap - trip1 within trip2", 200, 300, 100, 400, true},
		{"adjacent trips - no overlap", 100, 200, 200, 300, false},
		{"same time", 100, 200, 100, 200, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.doTripsOverlap(tt.start1, tt.end1, tt.start2, tt.end2)
			if result != tt.expected {
				t.Errorf("Expected overlap=%v for ranges [%d,%d] and [%d,%d], got %v",
					tt.expected, tt.start1, tt.end1, tt.start2, tt.end2, result)
			}
		})
	}
}

func TestTripBlockIdValidator_ParseTrip(t *testing.T) {
	validator := &TripBlockIdValidator{}

	tests := []struct {
		name     string
		row      *parser.CSVRow
		expected *TripBlockInfo
	}{
		{
			name: "valid trip with block",
			row: &parser.CSVRow{
				Values: map[string]string{
					"trip_id":    "trip_1",
					"route_id":   "route_1",
					"service_id": "service_1",
					"block_id":   "block_1",
				},
				RowNumber: 2,
			},
			expected: &TripBlockInfo{
				TripID:    "trip_1",
				RouteID:   "route_1",
				ServiceID: "service_1",
				BlockID:   "block_1",
				RowNumber: 2,
			},
		},
		{
			name: "valid trip without block",
			row: &parser.CSVRow{
				Values: map[string]string{
					"trip_id":    "trip_1",
					"route_id":   "route_1",
					"service_id": "service_1",
				},
				RowNumber: 2,
			},
			expected: &TripBlockInfo{
				TripID:    "trip_1",
				RouteID:   "route_1",
				ServiceID: "service_1",
				BlockID:   "",
				RowNumber: 2,
			},
		},
		{
			name: "trip with whitespace",
			row: &parser.CSVRow{
				Values: map[string]string{
					"trip_id":    "  trip_1  ",
					"route_id":   "  route_1  ",
					"service_id": " service_1 ",
					"block_id":   " block_1 ",
				},
				RowNumber: 2,
			},
			expected: &TripBlockInfo{
				TripID:    "trip_1",
				RouteID:   "route_1",
				ServiceID: "service_1",
				BlockID:   "block_1",
				RowNumber: 2,
			},
		},
		{
			name: "missing required field",
			row: &parser.CSVRow{
				Values: map[string]string{
					"trip_id":  "trip_1",
					"route_id": "route_1",
				},
				RowNumber: 2,
			},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.parseTrip(tt.row)

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
				result.RouteID != tt.expected.RouteID ||
				result.ServiceID != tt.expected.ServiceID ||
				result.BlockID != tt.expected.BlockID ||
				result.RowNumber != tt.expected.RowNumber {
				t.Errorf("Expected %+v, got %+v", tt.expected, result)
			}
		})
	}
}

func TestTripBlockIdValidator_ParseStopTime(t *testing.T) {
	validator := &TripBlockIdValidator{}

	tests := []struct {
		name     string
		row      *parser.CSVRow
		expected *StopTimeInfo
	}{
		{
			name: "valid stop time with times",
			row: &parser.CSVRow{
				Values: map[string]string{
					"trip_id":        "trip_1",
					"stop_sequence":  "1",
					"arrival_time":   "08:30:00",
					"departure_time": "08:31:00",
				},
			},
			expected: &StopTimeInfo{
				TripID:        "trip_1",
				StopSequence:  1,
				ArrivalTime:   30600,
				DepartureTime: 30660,
			},
		},
		{
			name: "valid stop time without times",
			row: &parser.CSVRow{
				Values: map[string]string{
					"trip_id":       "trip_1",
					"stop_sequence": "1",
				},
			},
			expected: &StopTimeInfo{
				TripID:        "trip_1",
				StopSequence:  1,
				ArrivalTime:   -1,
				DepartureTime: -1,
			},
		},
		{
			name: "invalid stop sequence",
			row: &parser.CSVRow{
				Values: map[string]string{
					"trip_id":       "trip_1",
					"stop_sequence": "not_a_number",
				},
			},
			expected: nil,
		},
		{
			name: "missing required fields",
			row: &parser.CSVRow{
				Values: map[string]string{
					"trip_id": "trip_1",
				},
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
				result.StopSequence != tt.expected.StopSequence ||
				result.ArrivalTime != tt.expected.ArrivalTime ||
				result.DepartureTime != tt.expected.DepartureTime {
				t.Errorf("Expected %+v, got %+v", tt.expected, result)
			}
		})
	}
}
