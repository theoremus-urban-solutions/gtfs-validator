package entity

import (
	"testing"
	"time"

	"github.com/theoremus-urban-solutions/gtfs-validator/testutil"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

func TestServiceValidationValidator_Validate(t *testing.T) {
	currentDate := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	const calendarHeader = "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\n"

	tests := []struct {
		name        string
		files       map[string]string
		wantCodes   []string
		description string
	}{
		{
			name: "valid service",
			files: map[string]string{
				"calendar.txt": calendarHeader + "service1,1,1,1,1,1,0,0,20240101,20241201",
				"trips.txt":    "trip_id,route_id,service_id\ntrip1,route1,service1",
			},
			wantCodes:   nil,
			description: "A used service still running generates nothing",
		},
		{
			name: "service without active days",
			files: map[string]string{
				"calendar.txt": calendarHeader + "service1,0,0,0,0,0,0,0,20240101,20241201",
				"trips.txt":    "trip_id,route_id,service_id\ntrip1,route1,service1",
			},
			wantCodes:   []string{"service_has_no_active_day_of_the_week"},
			description: "A service running on no weekday never runs at all",
		},
		{
			name: "expired service",
			files: map[string]string{
				"calendar.txt": calendarHeader + "service1,1,1,1,1,1,0,0,20230101,20230228",
				"trips.txt":    "trip_id,route_id,service_id\ntrip1,route1,service1",
			},
			wantCodes:   []string{"expired_calendar"},
			description: "A service whose end_date has passed can no longer be planned on",
		},
		{
			name: "added exception keeps an otherwise expired service alive",
			files: map[string]string{
				"calendar.txt":       calendarHeader + "service1,1,1,1,1,1,0,0,20230101,20230228",
				"calendar_dates.txt": "service_id,date,exception_type\nservice1,20241225,1",
				"trips.txt":          "trip_id,route_id,service_id\ntrip1,route1,service1",
			},
			wantCodes:   nil,
			description: "calendar_dates.txt additions extend the last active date past end_date",
		},
		{
			name: "removed exception does not extend a service",
			files: map[string]string{
				"calendar.txt":       calendarHeader + "service1,1,1,1,1,1,0,0,20230101,20230228",
				"calendar_dates.txt": "service_id,date,exception_type\nservice1,20241225,2",
				"trips.txt":          "trip_id,route_id,service_id\ntrip1,route1,service1",
			},
			wantCodes:   []string{"expired_calendar"},
			description: "A removal adds no service, so the service is still expired",
		},
		{
			name: "calendar_dates-only service entirely in the past",
			files: map[string]string{
				"calendar_dates.txt": "service_id,date,exception_type\nholiday,20240101,1\nholiday,20240215,1",
				"trips.txt":          "trip_id,route_id,service_id\ntrip1,route1,holiday",
			},
			wantCodes:   []string{"expired_calendar"},
			description: "The last added date is what such a service expires on",
		},
		{
			name: "calendar_dates-only service still to come",
			files: map[string]string{
				"calendar_dates.txt": "service_id,date,exception_type\nholiday,20240704,1\nholiday,20241225,1",
				"trips.txt":          "trip_id,route_id,service_id\ntrip1,route1,holiday",
			},
			wantCodes:   nil,
			description: "Additions in the future are exactly what the feed is for",
		},
		{
			name: "service ending more than two years out",
			files: map[string]string{
				"calendar.txt": calendarHeader + "service1,1,1,1,1,1,0,0,20240101,20281231",
				"trips.txt":    "trip_id,route_id,service_id\ntrip1,route1,service1",
			},
			wantCodes:   []string{"service_extends_far_in_the_future"},
			description: "Dates that far ahead are a placeholder, not a schedule anyone has checked",
		},
		{
			name: "service ending just inside two years",
			files: map[string]string{
				"calendar.txt": calendarHeader + "service1,1,1,1,1,1,0,0,20240101,20260101",
				"trips.txt":    "trip_id,route_id,service_id\ntrip1,route1,service1",
			},
			wantCodes:   nil,
			description: "Two years of published schedule is committed, not speculative",
		},
		{
			name: "calendar_dates-only service far in the future",
			files: map[string]string{
				"calendar_dates.txt": "service_id,date,exception_type\nholiday,20281225,1",
				"trips.txt":          "trip_id,route_id,service_id\ntrip1,route1,holiday",
			},
			wantCodes:   []string{"service_extends_far_in_the_future"},
			description: "The last added date decides this for such a service too",
		},
		{
			name: "unused calendar service",
			files: map[string]string{
				"calendar.txt": calendarHeader +
					"service1,1,1,1,1,1,0,0,20240101,20241201\n" +
					"spare,1,1,1,1,1,0,0,20240101,20241201",
				"trips.txt": "trip_id,route_id,service_id\ntrip1,route1,service1",
			},
			wantCodes:   []string{"unused_service"},
			description: "A calendar entry no trip references is dead weight",
		},
		{
			name: "unused calendar_dates service",
			files: map[string]string{
				"calendar_dates.txt": "service_id,date,exception_type\n" +
					"used,20241225,1\n" +
					"spare,20241225,1",
				"trips.txt": "trip_id,route_id,service_id\ntrip1,route1,used",
			},
			wantCodes:   []string{"unused_service"},
			description: "Same for a service defined only in calendar_dates.txt",
		},
		{
			name: "service present in both files is counted once",
			files: map[string]string{
				"calendar.txt":       calendarHeader + "service1,1,1,1,1,1,0,0,20240101,20241201",
				"calendar_dates.txt": "service_id,date,exception_type\nservice1,20240704,2",
				"trips.txt":          "trip_id,route_id,service_id\ntrip1,route1,service1",
			},
			wantCodes:   nil,
			description: "A service is not unused just because it appears twice",
		},
		{
			name: "service without dates",
			files: map[string]string{
				"calendar.txt": "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday\nservice1,1,1,1,1,1,0,0",
				"trips.txt":    "trip_id,route_id,service_id\ntrip1,route1,service1",
			},
			wantCodes:   nil,
			description: "Missing dates are a required-field problem, not an expiry one",
		},
		{
			name: "unparseable end date",
			files: map[string]string{
				"calendar.txt": calendarHeader + "service1,1,1,1,1,1,0,0,20240101,not-a-date",
				"trips.txt":    "trip_id,route_id,service_id\ntrip1,route1,service1",
			},
			wantCodes:   nil,
			description: "Reported as invalid_date by the type layer",
		},
		{
			name:        "no calendar files",
			files:       map[string]string{"trips.txt": "trip_id,route_id,service_id\ntrip1,route1,service1"},
			wantCodes:   nil,
			description: "Reported as missing_calendar_and_calendar_date_files elsewhere",
		},
		{
			name: "whitespace handling",
			files: map[string]string{
				"calendar.txt":       calendarHeader + " service1 , 1 , 1 , 1 , 1 , 1 , 0 , 0 , 20240101 , 20241201 ",
				"calendar_dates.txt": "service_id,date,exception_type\n service2 , 20241225 , 1 ",
				"trips.txt":          "trip_id,route_id,service_id\ntrip1,route1, service1 \ntrip2,route2, service2 ",
			},
			wantCodes:   nil,
			description: "Whitespace should be trimmed on both sides of every join",
		},
		{
			name: "calendar entry without service_id ignored",
			files: map[string]string{
				"calendar.txt": "monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\n1,1,1,1,1,0,0,20240101,20241201",
				"trips.txt":    "trip_id,route_id,service_id\ntrip1,route1,service1",
			},
			wantCodes:   nil,
			description: "A row with no key cannot be reported against a service",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			feedLoader := testutil.CreateTestFeedLoader(t, tt.files)
			container := notice.NewNoticeContainer()

			v := NewServiceValidationValidator()
			v.Validate(feedLoader, container, gtfsvalidator.Config{CurrentDate: currentDate})

			got := map[string]int{}
			for _, n := range container.GetNotices() {
				got[n.Code()]++
			}

			for _, code := range tt.wantCodes {
				if got[code] == 0 {
					t.Errorf("expected %s: %s (got %v)", code, tt.description, got)
				}
			}
			if len(got) != len(tt.wantCodes) {
				t.Errorf("expected exactly %v, got %v: %s", tt.wantCodes, got, tt.description)
			}
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
				"service1,1,1,1,1,1,0,0,20260801,20261201",
			expected: map[string]*ServiceInfo{
				"service1": {
					ServiceID: "service1",
					StartDate: "20260801",
					EndDate:   "20261201",
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
				" service1 , 1 , 0 , 1 , 0 , 1 , 0 , 1 , 20260801 , 20261201 ",
			expected: map[string]*ServiceInfo{
				"service1": {
					ServiceID: "service1",
					StartDate: "20260801",
					EndDate:   "20261201",
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
				"weekday,1,1,1,1,1,0,0,20260101,20261231\n" +
				"weekend,0,0,0,0,0,1,1,20260101,20261231",
			expected: map[string]*ServiceInfo{
				"weekday": {
					ServiceID: "weekday",
					StartDate: "20260101",
					EndDate:   "20261231",
					Days: map[string]bool{
						"monday": true, "tuesday": true, "wednesday": true, "thursday": true,
						"friday": true, "saturday": false, "sunday": false,
					},
					RowNumber: 2,
				},
				"weekend": {
					ServiceID: "weekend",
					StartDate: "20260101",
					EndDate:   "20261231",
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
			feedLoader := testutil.CreateTestFeedLoader(t, map[string]string{
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
		name        string
		csvData     string
		expected    map[string]string // service_id -> last added date, "" for none
		description string
	}{
		{
			name:        "removal only",
			csvData:     "service_id,date,exception_type\nservice1,20240704,2",
			expected:    map[string]string{"service1": ""},
			description: "A service only ever removed has no active date",
		},
		{
			name: "latest addition wins",
			csvData: "service_id,date,exception_type\n" +
				"service1,20240704,1\n" +
				"service2,20241225,1\n" +
				"service1,20240101,1",
			expected:    map[string]string{"service1": "20240704", "service2": "20241225"},
			description: "Rows arrive in no particular order",
		},
		{
			name:        "removals do not count as additions",
			csvData:     "service_id,date,exception_type\nservice1,20240101,1\nservice1,20241225,2",
			expected:    map[string]string{"service1": "20240101"},
			description: "A later removal must not extend the service",
		},
		{
			name:        "whitespace trimming",
			csvData:     "service_id,date,exception_type\n service1 , 20240704 , 1 ",
			expected:    map[string]string{"service1": "20240704"},
			description: "Values are padded in real feeds",
		},
		{
			name:        "empty file",
			csvData:     "service_id,date,exception_type\n",
			expected:    map[string]string{},
			description: "A header-only file defines no services",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			feedLoader := testutil.CreateTestFeedLoader(t, map[string]string{
				"calendar_dates.txt": tt.csvData,
			})

			result := validator.loadCalendarDateServices(feedLoader)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d services, got %d: %s", len(tt.expected), len(result), tt.description)
			}

			for serviceID, expectedDate := range tt.expected {
				service, exists := result[serviceID]
				if !exists {
					t.Errorf("Expected service %s not found: %s", serviceID, tt.description)
					continue
				}

				got := ""
				if service.LastAddedDate != nil {
					got = service.LastAddedDate.Format("20060102")
				}
				if got != expectedDate {
					t.Errorf("Service %s: expected last added date %q, got %q: %s", serviceID, expectedDate, got, tt.description)
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
			feedLoader := testutil.CreateTestFeedLoader(t, map[string]string{
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
