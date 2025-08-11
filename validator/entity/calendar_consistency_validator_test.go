package entity

import (
	"testing"
	"time"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

func TestCalendarConsistencyValidator_Validate(t *testing.T) {
	// Use a fixed current date for testing
	currentDate := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name                string
		files               map[string]string
		expectedNoticeCodes []string
		description         string
	}{
		{
			name: "valid calendar service",
			files: map[string]string{
				"calendar.txt": "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\n" +
					"service1,1,1,1,1,1,0,0,20240601,20241201",
				"trips.txt": "trip_id,route_id,service_id\n" +
					"trip1,route1,service1",
			},
			expectedNoticeCodes: []string{},
			description:         "Valid calendar service should not generate notices",
		},
		{
			name: "service never active",
			files: map[string]string{
				"calendar.txt": "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\n" +
					"service1,0,0,0,0,0,0,0,20240601,20241201",
				"trips.txt": "trip_id,route_id,service_id\n" +
					"trip1,route1,service1",
			},
			expectedNoticeCodes: []string{"service_never_active"},
			description:         "Service that runs on no days should generate error",
		},
		{
			name: "invalid service date range",
			files: map[string]string{
				"calendar.txt": "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\n" +
					"service1,1,1,1,1,1,0,0,20241201,20240601", // End before start
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
			name: "future service",
			files: map[string]string{
				"calendar.txt": "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\n" +
					"service1,1,1,1,1,1,0,0,20270601,20271201", // Starts more than 2 years in future
				"trips.txt": "trip_id,route_id,service_id\n" +
					"trip1,route1,service1",
			},
			expectedNoticeCodes: []string{"future_service"},
			description:         "Service starting more than 2 years in future should generate warning",
		},
		{
			name: "valid calendar dates",
			files: map[string]string{
				"calendar_dates.txt": "service_id,date,exception_type\n" +
					"service1,20240704,2\n" + // Service removed (holiday)
					"service1,20241225,2", // Service removed (Christmas)
				"trips.txt": "trip_id,route_id,service_id\n" +
					"trip1,route1,service1",
			},
			expectedNoticeCodes: []string{},
			description:         "Valid calendar dates should not generate notices",
		},
		{
			name: "invalid exception type",
			files: map[string]string{
				"calendar_dates.txt": "service_id,date,exception_type\n" +
					"service1,20240704,3", // Invalid exception type
				"trips.txt": "trip_id,route_id,service_id\n" +
					"trip1,route1,service1",
			},
			expectedNoticeCodes: []string{"invalid_exception_type"},
			description:         "Invalid exception type should generate error",
		},
		{
			name: "very old calendar date",
			files: map[string]string{
				"calendar_dates.txt": "service_id,date,exception_type\n" +
					"service1,20180101,2", // More than 5 years ago
				"trips.txt": "trip_id,route_id,service_id\n" +
					"trip1,route1,service1",
			},
			expectedNoticeCodes: []string{"very_old_calendar_date"},
			description:         "Calendar date more than 5 years ago should generate warning",
		},
		{
			name: "very future calendar date",
			files: map[string]string{
				"calendar_dates.txt": "service_id,date,exception_type\n" +
					"service1,20300101,2", // More than 5 years in future
				"trips.txt": "trip_id,route_id,service_id\n" +
					"trip1,route1,service1",
			},
			expectedNoticeCodes: []string{"very_future_calendar_date"},
			description:         "Calendar date more than 5 years in future should generate warning",
		},
		{
			name: "undefined service",
			files: map[string]string{
				"trips.txt": "trip_id,route_id,service_id\n" +
					"trip1,route1,undefined_service",
			},
			expectedNoticeCodes: []string{"undefined_service"},
			description:         "Trip using undefined service should generate error",
		},
		{
			name: "unused service",
			files: map[string]string{
				"calendar.txt": "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\n" +
					"service1,1,1,1,1,1,0,0,20240601,20241201\n" +
					"unused_service,1,1,1,1,1,0,0,20240601,20241201",
				"trips.txt": "trip_id,route_id,service_id\n" +
					"trip1,route1,service1",
			},
			expectedNoticeCodes: []string{"unused_service"},
			description:         "Service defined but not used should generate warning",
		},
		{
			name: "duplicate calendar exception",
			files: map[string]string{
				"calendar_dates.txt": "service_id,date,exception_type\n" +
					"service1,20240704,2\n" +
					"service1,20240704,2", // Same service, date, and exception type
				"trips.txt": "trip_id,route_id,service_id\n" +
					"trip1,route1,service1",
			},
			expectedNoticeCodes: []string{"duplicate_calendar_exception"},
			description:         "Duplicate calendar exceptions should generate warning",
		},
		{
			name: "conflicting calendar exception",
			files: map[string]string{
				"calendar_dates.txt": "service_id,date,exception_type\n" +
					"service1,20240704,1\n" +
					"service1,20240704,2", // Same service and date, different exception types
				"trips.txt": "trip_id,route_id,service_id\n" +
					"trip1,route1,service1",
			},
			expectedNoticeCodes: []string{"conflicting_calendar_exception"},
			description:         "Conflicting calendar exceptions should generate error",
		},
		{
			name: "service defined in both calendar and calendar_dates",
			files: map[string]string{
				"calendar.txt": "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\n" +
					"service1,1,1,1,1,1,0,0,20240601,20241201",
				"calendar_dates.txt": "service_id,date,exception_type\n" +
					"service1,20240704,2", // Holiday exception for same service
				"trips.txt": "trip_id,route_id,service_id\n" +
					"trip1,route1,service1",
			},
			expectedNoticeCodes: []string{},
			description:         "Service defined in both files should be valid",
		},
		{
			name: "multiple validation issues",
			files: map[string]string{
				"calendar.txt": "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\n" +
					"service1,0,0,0,0,0,0,0,20241201,20240601", // Never active + invalid date range
				"calendar_dates.txt": "service_id,date,exception_type\n" +
					"service2,20180101,3", // Old date + invalid exception type
				"trips.txt": "trip_id,route_id,service_id\n" +
					"trip1,route1,service1\n" +
					"trip2,route2,service2",
			},
			expectedNoticeCodes: []string{"service_never_active", "invalid_service_date_range", "very_old_calendar_date", "invalid_exception_type"},
			description:         "Multiple validation issues should generate multiple notices",
		},
		{
			name: "weekend only service",
			files: map[string]string{
				"calendar.txt": "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\n" +
					"weekend_service,0,0,0,0,0,1,1,20240601,20241201",
				"trips.txt": "trip_id,route_id,service_id\n" +
					"trip1,route1,weekend_service",
			},
			expectedNoticeCodes: []string{},
			description:         "Weekend-only service should be valid",
		},
		{
			name: "no calendar files",
			files: map[string]string{
				"trips.txt": "trip_id,route_id,service_id\n" +
					"trip1,route1,service1",
			},
			expectedNoticeCodes: []string{"undefined_service"},
			description:         "Missing calendar files should still validate service references",
		},
		{
			name: "calendar service without trips",
			files: map[string]string{
				"calendar.txt": "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\n" +
					"service1,1,1,1,1,1,0,0,20240601,20241201",
				"trips.txt": "trip_id,route_id,service_id\n" +
					"trip1,route1,different_service",
			},
			expectedNoticeCodes: []string{"unused_service", "undefined_service"},
			description:         "Service defined but not used should generate warning, undefined service should generate error",
		},
		{
			name: "calendar dates only service",
			files: map[string]string{
				"calendar_dates.txt": "service_id,date,exception_type\n" +
					"holidays_only,20240704,1\n" +
					"holidays_only,20241225,1", // Service only defined via exceptions
				"trips.txt": "trip_id,route_id,service_id\n" +
					"trip1,route1,holidays_only",
			},
			expectedNoticeCodes: []string{},
			description:         "Service defined only in calendar_dates should be valid",
		},
		{
			name: "edge case dates",
			files: map[string]string{
				"calendar.txt": "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\n" +
					"service1,1,1,1,1,1,0,0,20240829,20240829", // Single day service (not expired)
				"calendar_dates.txt": "service_id,date,exception_type\n" +
					"service2,20240829,1", // Same date exception
				"trips.txt": "trip_id,route_id,service_id\n" +
					"trip1,route1,service1\n" +
					"trip2,route2,service2",
			},
			expectedNoticeCodes: []string{},
			description:         "Edge case dates should be handled correctly",
		},
		{
			name: "whitespace handling",
			files: map[string]string{
				"calendar.txt": "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\n" +
					" service1 , 1 , 1 , 1 , 1 , 1 , 0 , 0 , 20240601 , 20241201 ",
				"calendar_dates.txt": "service_id,date,exception_type\n" +
					" service2 , 20240704 , 2 ",
				"trips.txt": "trip_id,route_id,service_id\n" +
					"trip1,route1, service1 \n" +
					"trip2,route2, service2 ",
			},
			expectedNoticeCodes: []string{},
			description:         "Whitespace should be trimmed properly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test feed loader
			feedLoader := CreateTestFeedLoader(t, tt.files)

			// Create notice container and validator
			container := notice.NewNoticeContainer()
			validator := NewCalendarConsistencyValidator()
			config := gtfsvalidator.Config{
				CurrentDate: currentDate,
			}

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

func TestCalendarConsistencyValidator_LoadCalendarServices(t *testing.T) {
	validator := NewCalendarConsistencyValidator()

	tests := []struct {
		name     string
		csvData  string
		expected map[string]*CalendarService
	}{
		{
			name: "basic calendar loading",
			csvData: "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\n" +
				"service1,1,1,1,1,1,0,0,20240601,20241201",
			expected: map[string]*CalendarService{
				"service1": {
					ServiceID: "service1",
					Monday:    true,
					Tuesday:   true,
					Wednesday: true,
					Thursday:  true,
					Friday:    true,
					Saturday:  false,
					Sunday:    false,
					StartDate: "20240601",
					EndDate:   "20241201",
					RowNumber: 2,
				},
			},
		},
		{
			name: "weekend service",
			csvData: "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\n" +
				"weekend,0,0,0,0,0,1,1,20240101,20241231",
			expected: map[string]*CalendarService{
				"weekend": {
					ServiceID: "weekend",
					Monday:    false,
					Tuesday:   false,
					Wednesday: false,
					Thursday:  false,
					Friday:    false,
					Saturday:  true,
					Sunday:    true,
					StartDate: "20240101",
					EndDate:   "20241231",
					RowNumber: 2,
				},
			},
		},
		{
			name: "whitespace trimming",
			csvData: "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\n" +
				" service1 , 1 , 0 , 1 , 0 , 1 , 0 , 1 , 20240601 , 20241201 ",
			expected: map[string]*CalendarService{
				"service1": {
					ServiceID: "service1",
					Monday:    true,
					Tuesday:   false,
					Wednesday: true,
					Thursday:  false,
					Friday:    true,
					Saturday:  false,
					Sunday:    true,
					StartDate: "20240601",
					EndDate:   "20241201",
					RowNumber: 2,
				},
			},
		},
		{
			name: "multiple services",
			csvData: "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\n" +
				"weekday,1,1,1,1,1,0,0,20240101,20241231\n" +
				"weekend,0,0,0,0,0,1,1,20240101,20241231",
			expected: map[string]*CalendarService{
				"weekday": {
					ServiceID: "weekday",
					Monday:    true,
					Tuesday:   true,
					Wednesday: true,
					Thursday:  true,
					Friday:    true,
					Saturday:  false,
					Sunday:    false,
					StartDate: "20240101",
					EndDate:   "20241231",
					RowNumber: 2,
				},
				"weekend": {
					ServiceID: "weekend",
					Monday:    false,
					Tuesday:   false,
					Wednesday: false,
					Thursday:  false,
					Friday:    false,
					Saturday:  true,
					Sunday:    true,
					StartDate: "20240101",
					EndDate:   "20241231",
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
				if actualService.Monday != expectedService.Monday {
					t.Errorf("Service %s: expected Monday %v, got %v", serviceID, expectedService.Monday, actualService.Monday)
				}
				if actualService.Tuesday != expectedService.Tuesday {
					t.Errorf("Service %s: expected Tuesday %v, got %v", serviceID, expectedService.Tuesday, actualService.Tuesday)
				}
				if actualService.Wednesday != expectedService.Wednesday {
					t.Errorf("Service %s: expected Wednesday %v, got %v", serviceID, expectedService.Wednesday, actualService.Wednesday)
				}
				if actualService.Thursday != expectedService.Thursday {
					t.Errorf("Service %s: expected Thursday %v, got %v", serviceID, expectedService.Thursday, actualService.Thursday)
				}
				if actualService.Friday != expectedService.Friday {
					t.Errorf("Service %s: expected Friday %v, got %v", serviceID, expectedService.Friday, actualService.Friday)
				}
				if actualService.Saturday != expectedService.Saturday {
					t.Errorf("Service %s: expected Saturday %v, got %v", serviceID, expectedService.Saturday, actualService.Saturday)
				}
				if actualService.Sunday != expectedService.Sunday {
					t.Errorf("Service %s: expected Sunday %v, got %v", serviceID, expectedService.Sunday, actualService.Sunday)
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
			}
		})
	}
}

func TestCalendarConsistencyValidator_LoadCalendarDates(t *testing.T) {
	validator := NewCalendarConsistencyValidator()

	tests := []struct {
		name     string
		csvData  string
		expected []*CalendarDate
	}{
		{
			name: "basic calendar dates loading",
			csvData: "service_id,date,exception_type\n" +
				"service1,20240704,2\n" +
				"service1,20241225,2",
			expected: []*CalendarDate{
				{
					ServiceID:     "service1",
					Date:          "20240704",
					ExceptionType: 2,
					RowNumber:     2,
				},
				{
					ServiceID:     "service1",
					Date:          "20241225",
					ExceptionType: 2,
					RowNumber:     3,
				},
			},
		},
		{
			name: "mixed exception types",
			csvData: "service_id,date,exception_type\n" +
				"service1,20240704,2\n" +
				"service1,20240705,1",
			expected: []*CalendarDate{
				{
					ServiceID:     "service1",
					Date:          "20240704",
					ExceptionType: 2,
					RowNumber:     2,
				},
				{
					ServiceID:     "service1",
					Date:          "20240705",
					ExceptionType: 1,
					RowNumber:     3,
				},
			},
		},
		{
			name: "whitespace trimming",
			csvData: "service_id,date,exception_type\n" +
				" service1 , 20240704 , 2 ",
			expected: []*CalendarDate{
				{
					ServiceID:     "service1",
					Date:          "20240704",
					ExceptionType: 2,
					RowNumber:     2,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			feedLoader := CreateTestFeedLoader(t, map[string]string{
				"calendar_dates.txt": tt.csvData,
			})

			result := validator.loadCalendarDates(feedLoader)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d calendar dates, got %d", len(tt.expected), len(result))
			}

			for i, expectedDate := range tt.expected {
				if i >= len(result) {
					t.Errorf("Expected calendar date at index %d not found", i)
					continue
				}

				actualDate := result[i]
				if actualDate.ServiceID != expectedDate.ServiceID {
					t.Errorf("Date %d: expected ServiceID %s, got %s", i, expectedDate.ServiceID, actualDate.ServiceID)
				}
				if actualDate.Date != expectedDate.Date {
					t.Errorf("Date %d: expected Date %s, got %s", i, expectedDate.Date, actualDate.Date)
				}
				if actualDate.ExceptionType != expectedDate.ExceptionType {
					t.Errorf("Date %d: expected ExceptionType %d, got %d", i, expectedDate.ExceptionType, actualDate.ExceptionType)
				}
				if actualDate.RowNumber != expectedDate.RowNumber {
					t.Errorf("Date %d: expected RowNumber %d, got %d", i, expectedDate.RowNumber, actualDate.RowNumber)
				}
			}
		})
	}
}

func TestCalendarConsistencyValidator_GetUsedServiceIDs(t *testing.T) {
	validator := NewCalendarConsistencyValidator()

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

			result := validator.getUsedServiceIDs(feedLoader)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d used services, got %d", len(tt.expected), len(result))
			}

			for serviceID, expected := range tt.expected {
				if actual, exists := result[serviceID]; !exists || actual != expected {
					t.Errorf("Service %s: expected %v, got %v", serviceID, expected, actual)
				}
			}
		})
	}
}

func TestCalendarConsistencyValidator_New(t *testing.T) {
	validator := NewCalendarConsistencyValidator()
	if validator == nil {
		t.Error("NewCalendarConsistencyValidator() returned nil")
	}
}
