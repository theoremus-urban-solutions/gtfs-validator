package entity

import (
	"testing"
	"time"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

func TestServiceValidationValidator_Validate(t *testing.T) {
	tests := []struct {
		name                string
		files               map[string]string
		expectedNoticeCodes []string
		description         string
	}{
		{
			name: "valid service",
			files: map[string]string{
				"calendar.txt": "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\n" +
					"service1,1,1,1,1,1,0,0,20250801,20251201",
				"trips.txt": "trip_id,route_id,service_id\n" +
					"trip1,route1,service1",
			},
			expectedNoticeCodes: []string{},
			description:         "Valid service should not generate notices",
		},
		{
			name: "service without active days",
			files: map[string]string{
				"calendar.txt": "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\n" +
					"service1,0,0,0,0,0,0,0,20250801,20251201",
				"trips.txt": "trip_id,route_id,service_id\n" +
					"trip1,route1,service1",
			},
			expectedNoticeCodes: []string{"service_without_active_days"},
			description:         "Service with no active days should generate error",
		},
		{
			name: "invalid date range",
			files: map[string]string{
				"calendar.txt": "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\n" +
					"service1,1,1,1,1,1,0,0,20251201,20250801", // End before start
				"trips.txt": "trip_id,route_id,service_id\n" +
					"trip1,route1,service1",
			},
			expectedNoticeCodes: []string{"invalid_service_date_range"},
			description:         "Service with end date before start date should generate error",
		},
		{
			name: "expired service",
			files: map[string]string{
				"calendar.txt": "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\n" +
					"service1,1,1,1,1,1,0,0,20230101,20230228", // Ended more than 30 days ago
				"trips.txt": "trip_id,route_id,service_id\n" +
					"trip1,route1,service1",
			},
			expectedNoticeCodes: []string{"expired_service"},
			description:         "Service that ended more than 30 days ago should generate warning",
		},
		{
			name: "unused calendar service",
			files: map[string]string{
				"calendar.txt": "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\n" +
					"service1,1,1,1,1,1,0,0,20250801,20251201\n" +
					"unused_service,1,1,1,1,1,0,0,20250801,20251201",
				"trips.txt": "trip_id,route_id,service_id\n" +
					"trip1,route1,service1",
			},
			expectedNoticeCodes: []string{"unused_service"},
			description:         "Unused service should generate warning",
		},
		{
			name: "unused calendar_dates service",
			files: map[string]string{
				"calendar_dates.txt": "service_id,date,exception_type\n" +
					"used_service,20240704,2\n" +
					"unused_service,20241225,2",
				"trips.txt": "trip_id,route_id,service_id\n" +
					"trip1,route1,used_service",
			},
			expectedNoticeCodes: []string{"unused_service"},
			description:         "Unused calendar_dates service should generate warning",
		},
		{
			name: "weekend only service",
			files: map[string]string{
				"calendar.txt": "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\n" +
					"weekend,0,0,0,0,0,1,1,20250801,20251201",
				"trips.txt": "trip_id,route_id,service_id\n" +
					"trip1,route1,weekend",
			},
			expectedNoticeCodes: []string{},
			description:         "Weekend-only service should be valid",
		},
		{
			name: "service without dates",
			files: map[string]string{
				"calendar.txt": "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday\n" +
					"service1,1,1,1,1,1,0,0",
				"trips.txt": "trip_id,route_id,service_id\n" +
					"trip1,route1,service1",
			},
			expectedNoticeCodes: []string{},
			description:         "Service without dates should not generate date-related errors",
		},
		{
			name: "mixed services",
			files: map[string]string{
				"calendar.txt": "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\n" +
					"valid_service,1,1,1,1,1,0,0,20250801,20251201\n" +
					"no_days,0,0,0,0,0,0,0,20250801,20251201\n" +
					"invalid_range,1,1,1,1,1,0,0,20251201,20250801",
				"trips.txt": "trip_id,route_id,service_id\n" +
					"trip1,route1,valid_service\n" +
					"trip2,route2,no_days\n" +
					"trip3,route3,invalid_range",
			},
			expectedNoticeCodes: []string{"service_without_active_days", "invalid_service_date_range"},
			description:         "Multiple services with different issues should generate appropriate notices",
		},
		{
			name: "service only in calendar_dates",
			files: map[string]string{
				"calendar_dates.txt": "service_id,date,exception_type\n" +
					"holiday_service,20240704,1\n" +
					"holiday_service,20241225,1",
				"trips.txt": "trip_id,route_id,service_id\n" +
					"trip1,route1,holiday_service",
			},
			expectedNoticeCodes: []string{},
			description:         "Service defined only in calendar_dates should be valid",
		},
		{
			name: "service in both calendar and calendar_dates",
			files: map[string]string{
				"calendar.txt": "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\n" +
					"service1,1,1,1,1,1,0,0,20250801,20251201",
				"calendar_dates.txt": "service_id,date,exception_type\n" +
					"service1,20250704,2", // Holiday exception
				"trips.txt": "trip_id,route_id,service_id\n" +
					"trip1,route1,service1",
			},
			expectedNoticeCodes: []string{},
			description:         "Service in both files should be valid",
		},
		{
			name: "current date service",
			files: map[string]string{
				"calendar.txt": "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\n" +
					"current_service,1,1,1,1,1,0,0,20250801,20251201", // Should not be expired
				"trips.txt": "trip_id,route_id,service_id\n" +
					"trip1,route1,current_service",
			},
			expectedNoticeCodes: []string{},
			description:         "Current date service should not be marked as expired",
		},
		{
			name: "no calendar files",
			files: map[string]string{
				"trips.txt": "trip_id,route_id,service_id\n" +
					"trip1,route1,service1",
			},
			expectedNoticeCodes: []string{},
			description:         "Missing calendar files should not generate errors",
		},
		{
			name: "empty calendar file",
			files: map[string]string{
				"calendar.txt": "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\n",
				"trips.txt": "trip_id,route_id,service_id\n" +
					"trip1,route1,service1",
			},
			expectedNoticeCodes: []string{},
			description:         "Empty calendar file should not generate errors",
		},
		{
			name: "service with invalid date format ignored",
			files: map[string]string{
				"calendar.txt": "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\n" +
					"service1,1,1,1,1,1,0,0,invalid,20251201",
				"trips.txt": "trip_id,route_id,service_id\n" +
					"trip1,route1,service1",
			},
			expectedNoticeCodes: []string{},
			description:         "Service with invalid date format should not cause date-related validations",
		},
		{
			name: "single day service",
			files: map[string]string{
				"calendar.txt": "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\n" +
					"single_day,0,0,0,0,1,0,0,20250829,20250829", // Single Friday
				"trips.txt": "trip_id,route_id,service_id\n" +
					"trip1,route1,single_day",
			},
			expectedNoticeCodes: []string{},
			description:         "Single day service should be valid",
		},
		{
			name: "whitespace handling",
			files: map[string]string{
				"calendar.txt": "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\n" +
					" service1 , 1 , 1 , 1 , 1 , 1 , 0 , 0 , 20250801 , 20251201 ",
				"calendar_dates.txt": "service_id,date,exception_type\n" +
					" service2 , 20250704 , 2 ",
				"trips.txt": "trip_id,route_id,service_id\n" +
					"trip1,route1, service1 \n" +
					"trip2,route2, service2 ",
			},
			expectedNoticeCodes: []string{},
			description:         "Whitespace should be trimmed properly",
		},
		{
			name: "service without service_id ignored",
			files: map[string]string{
				"calendar.txt": "monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\n" +
					"1,1,1,1,1,0,0,20240601,20241201",
				"trips.txt": "trip_id,route_id,service_id\n" +
					"trip1,route1,service1",
			},
			expectedNoticeCodes: []string{},
			description:         "Calendar entries without service_id should be ignored",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test feed loader
			feedLoader := CreateTestFeedLoader(t, tt.files)

			// Create notice container and validator
			container := notice.NewNoticeContainer()
			validator := NewServiceValidationValidator()
			config := gtfsvalidator.Config{}

			// Run validation
			validator.Validate(feedLoader, container, config)

			// Get all notices
			allNotices := container.GetNotices()

			// Extract notice codes
			var actualNoticeCodes []string
			for _, n := range allNotices {
				actualNoticeCodes = append(actualNoticeCodes, n.Code())
			}

			// Check if we got the expected notice codes
			expectedSet := make(map[string]bool)
			for _, code := range tt.expectedNoticeCodes {
				expectedSet[code] = true
			}

			actualSet := make(map[string]bool)
			for _, code := range actualNoticeCodes {
				actualSet[code] = true
			}

			// Verify expected codes are present
			for expectedCode := range expectedSet {
				if !actualSet[expectedCode] {
					t.Errorf("Expected notice code '%s' not found. Got: %v", expectedCode, actualNoticeCodes)
				}
			}

			// If no notices expected, ensure no notices were generated
			if len(tt.expectedNoticeCodes) == 0 && len(actualNoticeCodes) > 0 {
				t.Errorf("Expected no notices, but got: %v", actualNoticeCodes)
			}

			t.Logf("Test '%s': Expected %v, Got %v", tt.name, tt.expectedNoticeCodes, actualNoticeCodes)
		})
	}
}

func TestServiceValidationValidator_LoadCalendarServices(t *testing.T) {
	validator := NewServiceValidationValidator()

	tests := []struct {
		name     string
		csvData  string
		expected map[string]*ServiceInfo
	}{
		{
			name: "basic service loading",
			csvData: "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\n" +
				"service1,1,1,1,1,1,0,0,20250801,20251201",
			expected: map[string]*ServiceInfo{
				"service1": {
					ServiceID: "service1",
					StartDate: "20250801",
					EndDate:   "20251201",
					Days: map[string]bool{
						"monday":    true,
						"tuesday":   true,
						"wednesday": true,
						"thursday":  true,
						"friday":    true,
						"saturday":  false,
						"sunday":    false,
					},
					RowNumber: 2,
				},
			},
		},
		{
			name: "weekend service",
			csvData: "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\n" +
				"weekend,0,0,0,0,0,1,1,20250101,20251231",
			expected: map[string]*ServiceInfo{
				"weekend": {
					ServiceID: "weekend",
					StartDate: "20250101",
					EndDate:   "20251231",
					Days: map[string]bool{
						"monday":    false,
						"tuesday":   false,
						"wednesday": false,
						"thursday":  false,
						"friday":    false,
						"saturday":  true,
						"sunday":    true,
					},
					RowNumber: 2,
				},
			},
		},
		{
			name: "service without dates",
			csvData: "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday\n" +
				"service1,1,0,1,0,1,0,1",
			expected: map[string]*ServiceInfo{
				"service1": {
					ServiceID: "service1",
					StartDate: "",
					EndDate:   "",
					Days: map[string]bool{
						"monday":    true,
						"tuesday":   false,
						"wednesday": true,
						"thursday":  false,
						"friday":    true,
						"saturday":  false,
						"sunday":    true,
					},
					RowNumber: 2,
				},
			},
		},
		{
			name: "whitespace trimming",
			csvData: "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\n" +
				" service1 , 1 , 0 , 1 , 0 , 1 , 0 , 1 , 20250801 , 20251201 ",
			expected: map[string]*ServiceInfo{
				"service1": {
					ServiceID: "service1",
					StartDate: "20250801",
					EndDate:   "20251201",
					Days: map[string]bool{
						"monday":    true,
						"tuesday":   false,
						"wednesday": true,
						"thursday":  false,
						"friday":    true,
						"saturday":  false,
						"sunday":    true,
					},
					RowNumber: 2,
				},
			},
		},
		{
			name: "multiple services",
			csvData: "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\n" +
				"weekday,1,1,1,1,1,0,0,20250101,20251231\n" +
				"weekend,0,0,0,0,0,1,1,20250101,20251231",
			expected: map[string]*ServiceInfo{
				"weekday": {
					ServiceID: "weekday",
					StartDate: "20250101",
					EndDate:   "20251231",
					Days: map[string]bool{
						"monday": true, "tuesday": true, "wednesday": true, "thursday": true,
						"friday": true, "saturday": false, "sunday": false,
					},
					RowNumber: 2,
				},
				"weekend": {
					ServiceID: "weekend",
					StartDate: "20250101",
					EndDate:   "20251231",
					Days: map[string]bool{
						"monday": false, "tuesday": false, "wednesday": false, "thursday": false,
						"friday": false, "saturday": true, "sunday": true,
					},
					RowNumber: 3,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			feedLoader := CreateTestFeedLoader(t, map[string]string{
				"calendar.txt": tt.csvData,
			})

			result := validator.loadCalendarServices(feedLoader)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d services, got %d", len(tt.expected), len(result))
			}

			for serviceID, expectedService := range tt.expected {
				actualService, exists := result[serviceID]
				if !exists {
					t.Errorf("Expected service %s not found", serviceID)
					continue
				}

				if actualService.ServiceID != expectedService.ServiceID {
					t.Errorf("Service %s: expected ServiceID %s, got %s", serviceID, expectedService.ServiceID, actualService.ServiceID)
				}
				if actualService.StartDate != expectedService.StartDate {
					t.Errorf("Service %s: expected StartDate %s, got %s", serviceID, expectedService.StartDate, actualService.StartDate)
				}
				if actualService.EndDate != expectedService.EndDate {
					t.Errorf("Service %s: expected EndDate %s, got %s", serviceID, expectedService.EndDate, actualService.EndDate)
				}
				if actualService.RowNumber != expectedService.RowNumber {
					t.Errorf("Service %s: expected RowNumber %d, got %d", serviceID, expectedService.RowNumber, actualService.RowNumber)
				}

				// Check days
				for day, expectedValue := range expectedService.Days {
					if actualValue, exists := actualService.Days[day]; !exists || actualValue != expectedValue {
						t.Errorf("Service %s day %s: expected %v, got %v", serviceID, day, expectedValue, actualValue)
					}
				}
			}
		})
	}
}

func TestServiceValidationValidator_LoadCalendarDateServices(t *testing.T) {
	validator := NewServiceValidationValidator()

	tests := []struct {
		name     string
		csvData  string
		expected map[string]bool
	}{
		{
			name: "single service",
			csvData: "service_id,date,exception_type\n" +
				"service1,20240704,2",
			expected: map[string]bool{
				"service1": true,
			},
		},
		{
			name: "multiple services",
			csvData: "service_id,date,exception_type\n" +
				"service1,20240704,2\n" +
				"service2,20241225,2\n" +
				"service1,20240101,1", // Duplicate service - should still appear once
			expected: map[string]bool{
				"service1": true,
				"service2": true,
			},
		},
		{
			name: "whitespace trimming",
			csvData: "service_id,date,exception_type\n" +
				" service1 , 20240704 , 2 ",
			expected: map[string]bool{
				"service1": true,
			},
		},
		{
			name:     "empty file",
			csvData:  "service_id,date,exception_type\n",
			expected: map[string]bool{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			feedLoader := CreateTestFeedLoader(t, map[string]string{
				"calendar_dates.txt": tt.csvData,
			})

			result := validator.loadCalendarDateServices(feedLoader)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d services, got %d", len(tt.expected), len(result))
			}

			for serviceID, expectedValue := range tt.expected {
				if actualValue, exists := result[serviceID]; !exists || actualValue != expectedValue {
					t.Errorf("Service %s: expected %v, got %v", serviceID, expectedValue, actualValue)
				}
			}
		})
	}
}

func TestServiceValidationValidator_LoadUsedServices(t *testing.T) {
	validator := NewServiceValidationValidator()

	tests := []struct {
		name     string
		csvData  string
		expected map[string]bool
	}{
		{
			name: "single service",
			csvData: "trip_id,route_id,service_id\n" +
				"trip1,route1,service1",
			expected: map[string]bool{
				"service1": true,
			},
		},
		{
			name: "multiple services",
			csvData: "trip_id,route_id,service_id\n" +
				"trip1,route1,service1\n" +
				"trip2,route2,service2\n" +
				"trip3,route1,service1", // Duplicate service
			expected: map[string]bool{
				"service1": true,
				"service2": true,
			},
		},
		{
			name: "whitespace trimming",
			csvData: "trip_id,route_id,service_id\n" +
				"trip1,route1, service1 ",
			expected: map[string]bool{
				"service1": true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			feedLoader := CreateTestFeedLoader(t, map[string]string{
				"trips.txt": tt.csvData,
			})

			result := validator.loadUsedServices(feedLoader)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d used services, got %d", len(tt.expected), len(result))
			}

			for serviceID, expectedValue := range tt.expected {
				if actualValue, exists := result[serviceID]; !exists || actualValue != expectedValue {
					t.Errorf("Service %s: expected %v, got %v", serviceID, expectedValue, actualValue)
				}
			}
		})
	}
}

func TestServiceValidationValidator_ParseGTFSDate(t *testing.T) {
	validator := NewServiceValidationValidator()

	tests := []struct {
		dateStr     string
		shouldError bool
		expected    time.Time
	}{
		{"20240601", false, time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)},
		{"20241225", false, time.Date(2024, 12, 25, 0, 0, 0, 0, time.UTC)},
		{"20240229", false, time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC)}, // Leap year
		{"2024060", true, time.Time{}},                                    // Too short
		{"202406011", true, time.Time{}},                                  // Too long
		{"invalid", true, time.Time{}},                                    // Invalid format
		{"20241301", false, time.Date(2024, 13, 1, 0, 0, 0, 0, time.UTC)}, // Invalid month but parseable
		{"", true, time.Time{}},                                           // Empty
	}

	for _, tt := range tests {
		t.Run(tt.dateStr, func(t *testing.T) {
			result, err := validator.parseGTFSDate(tt.dateStr)

			if tt.shouldError {
				if err == nil {
					t.Errorf("Expected error for date '%s', but got none", tt.dateStr)
				}
			} else {
				switch {
				case err != nil:
					t.Errorf("Expected no error for date '%s', but got: %v", tt.dateStr, err)
				case result == nil:
					t.Errorf("Expected non-nil result for date '%s'", tt.dateStr)
				case !result.Equal(tt.expected):
					t.Errorf("Date '%s': expected %v, got %v", tt.dateStr, tt.expected, *result)
				}
			}
		})
	}
}

func TestServiceValidationValidator_New(t *testing.T) {
	validator := NewServiceValidationValidator()
	if validator == nil {
		t.Error("NewServiceValidationValidator() returned nil")
	}
}
