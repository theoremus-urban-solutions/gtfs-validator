package entity

import (
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

func TestAgencyConsistencyValidator_Validate(t *testing.T) {
	tests := []struct {
		name                string
		files               map[string]string
		expectedNoticeCodes []string
		description         string
	}{
		{
			name: "single agency without agency_id",
			files: map[string]string{
				"agency.txt": "agency_name,agency_url,agency_timezone\nMetro,http://metro.example,America/Los_Angeles",
				"routes.txt": "route_id,route_short_name,route_type\n1,Red,3",
			},
			expectedNoticeCodes: []string{},
			description:         "Single agency without agency_id should be valid",
		},
		{
			name: "single agency with agency_id",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,http://metro.example,America/Los_Angeles",
				"routes.txt": "route_id,agency_id,route_short_name,route_type\nR1,1,Red,3",
			},
			expectedNoticeCodes: []string{},
			description:         "Single agency with agency_id should be valid",
		},
		{
			name: "multiple agencies all with agency_id",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,http://metro.example,America/Los_Angeles\n2,Bus,http://bus.example,America/Los_Angeles",
				"routes.txt": "route_id,agency_id,route_short_name,route_type\nR1,1,Red,3\nR2,2,Blue,3",
			},
			expectedNoticeCodes: []string{},
			description:         "Multiple agencies with valid agency_ids should be valid",
		},
		{
			name: "multiple agencies with missing agency_id",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,http://metro.example,America/Los_Angeles\n,Bus,http://bus.example,America/Los_Angeles", // Missing agency_id
			},
			expectedNoticeCodes: []string{"missing_agency_id"},
			description:         "Multiple agencies require agency_id for all",
		},
		{
			name: "route references invalid agency",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,http://metro.example,America/Los_Angeles",
				"routes.txt": "route_id,agency_id,route_short_name,route_type\nR1,999,Red,3", // Invalid agency reference
			},
			expectedNoticeCodes: []string{"invalid_agency_reference"},
			description:         "Route referencing non-existent agency should generate notice",
		},
		{
			name: "route missing agency_id with multiple agencies",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,http://metro.example,America/Los_Angeles\n2,Bus,http://bus.example,America/Los_Angeles",
				"routes.txt": "route_id,route_short_name,route_type\nR1,Red,3", // Missing agency_id
			},
			expectedNoticeCodes: []string{"missing_route_agency_id"},
			description:         "Route without agency_id when multiple agencies exist should generate notice",
		},
		{
			name: "route omits agency_id with single agency",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,http://metro.example,America/Los_Angeles",
				"routes.txt": "route_id,route_short_name,route_type\nR1,Red,3", // Can omit agency_id with single agency
			},
			expectedNoticeCodes: []string{},
			description:         "Route can omit agency_id when only one agency exists",
		},
		{
			name: "empty agency_id treated as missing",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\n,Metro,http://metro.example,America/Los_Angeles\n2,Bus,http://bus.example,America/Los_Angeles",
			},
			expectedNoticeCodes: []string{"missing_agency_id"},
			description:         "Empty agency_id should be treated as missing",
		},
		{
			name: "whitespace-only agency_id treated as missing",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\n   ,Metro,http://metro.example,America/Los_Angeles\n2,Bus,http://bus.example,America/Los_Angeles",
			},
			expectedNoticeCodes: []string{"missing_agency_id"},
			description:         "Whitespace-only agency_id should be treated as missing",
		},
		{
			name: "agency_id with whitespace padding",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\n 1 ,Metro,http://metro.example,America/Los_Angeles",
				"routes.txt": "route_id,agency_id,route_short_name,route_type\nR1, 1 ,Red,3", // Whitespace around agency_id
			},
			expectedNoticeCodes: []string{},
			description:         "Agency_id values should be trimmed for comparison",
		},
		{
			name: "route agency_id with whitespace",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,http://metro.example,America/Los_Angeles",
				"routes.txt": "route_id,agency_id,route_short_name,route_type\nR1,  1  ,Red,3", // Whitespace around route agency_id
			},
			expectedNoticeCodes: []string{},
			description:         "Route agency_id values should be trimmed for comparison",
		},
		{
			name: "multiple invalid agency references",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,http://metro.example,America/Los_Angeles",
				"routes.txt": "route_id,agency_id,route_short_name,route_type\nR1,999,Red,3\nR2,888,Blue,3", // Multiple invalid references
			},
			expectedNoticeCodes: []string{"invalid_agency_reference", "invalid_agency_reference"},
			description:         "Multiple invalid agency references should generate multiple notices",
		},
		{
			name: "mixed valid and invalid route references",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,http://metro.example,America/Los_Angeles\n2,Bus,http://bus.example,America/Los_Angeles",
				"routes.txt": "route_id,agency_id,route_short_name,route_type\nR1,1,Red,3\nR2,999,Blue,3\nR3,2,Green,3", // Mixed valid/invalid
			},
			expectedNoticeCodes: []string{"invalid_agency_reference"},
			description:         "Mix of valid and invalid references should only generate notices for invalid ones",
		},
		{
			name: "no agency.txt file",
			files: map[string]string{
				"routes.txt": "route_id,route_short_name,route_type\nR1,Red,3",
			},
			expectedNoticeCodes: []string{},
			description:         "Missing agency.txt should not cause validation errors in this validator",
		},
		{
			name: "no routes.txt file",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,http://metro.example,America/Los_Angeles\n2,Bus,http://bus.example,America/Los_Angeles",
			},
			expectedNoticeCodes: []string{},
			description:         "Missing routes.txt should not cause validation errors",
		},
		{
			name: "empty agency.txt file",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone", // Headers only
				"routes.txt": "route_id,route_short_name,route_type\nR1,Red,3",
			},
			expectedNoticeCodes: []string{},
			description:         "Empty agency.txt should not cause validation errors",
		},
		{
			name: "empty routes.txt file",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,http://metro.example,America/Los_Angeles",
				"routes.txt": "route_id,agency_id,route_short_name,route_type", // Headers only
			},
			expectedNoticeCodes: []string{},
			description:         "Empty routes.txt should not cause validation errors",
		},
		{
			name: "route without route_id ignored",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,http://metro.example,America/Los_Angeles",
				"routes.txt": "route_id,agency_id,route_short_name,route_type\n,1,Red,3", // Missing route_id
			},
			expectedNoticeCodes: []string{},
			description:         "Routes without route_id should be ignored for agency validation",
		},
		{
			name: "multiple agencies some without agency_id",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,http://metro.example,America/Los_Angeles\n,Bus,http://bus.example,America/Los_Angeles\n3,Rail,http://rail.example,America/Los_Angeles\n,Subway,http://subway.example,America/Los_Angeles", // Two missing agency_ids
			},
			expectedNoticeCodes: []string{"missing_agency_id", "missing_agency_id"},
			description:         "Multiple agencies with missing agency_ids should generate multiple notices",
		},
		{
			name: "case sensitive agency_id matching",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\nAgency1,Metro,http://metro.example,America/Los_Angeles",
				"routes.txt": "route_id,agency_id,route_short_name,route_type\nR1,agency1,Red,3", // Different case
			},
			expectedNoticeCodes: []string{"invalid_agency_reference"},
			description:         "Agency_id matching should be case-sensitive",
		},
		{
			name: "numeric agency_ids",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\n100,Metro,http://metro.example,America/Los_Angeles\n200,Bus,http://bus.example,America/Los_Angeles",
				"routes.txt": "route_id,agency_id,route_short_name,route_type\nR1,100,Red,3\nR2,200,Blue,3",
			},
			expectedNoticeCodes: []string{},
			description:         "Numeric agency_ids should work correctly",
		},
		{
			name: "complex agency_ids with special characters",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\nMETRO-LA-001,Metro,http://metro.example,America/Los_Angeles\nBUS_SYS_002,Bus,http://bus.example,America/Los_Angeles",
				"routes.txt": "route_id,agency_id,route_short_name,route_type\nR1,METRO-LA-001,Red,3\nR2,BUS_SYS_002,Blue,3",
			},
			expectedNoticeCodes: []string{},
			description:         "Complex agency_ids with special characters should work",
		},
		{
			name: "agency without agency_name field",
			files: map[string]string{
				"agency.txt": "agency_id,agency_url,agency_timezone\n1,http://metro.example,America/Los_Angeles\n2,http://bus.example,America/Los_Angeles", // Missing agency_name column
			},
			expectedNoticeCodes: []string{},
			description:         "Missing agency_name field should not cause validation errors",
		},
		{
			name: "route references empty string agency_id",
			files: map[string]string{
				"agency.txt": "agency_name,agency_url,agency_timezone\nMetro,http://metro.example,America/Los_Angeles", // No agency_id column
				"routes.txt": "route_id,agency_id,route_short_name,route_type\nR1,,Red,3",                              // Empty agency_id reference
			},
			expectedNoticeCodes: []string{},
			description:         "Route referencing empty agency_id with single agency should be valid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test components
			loader := CreateTestFeedLoader(t, tt.files)
			container := notice.NewNoticeContainer()
			v := NewAgencyConsistencyValidator()
			config := gtfsvalidator.Config{}

			// Run validation
			v.Validate(loader, container, config)

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

func TestAgencyConsistencyValidator_LoadAgencies(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		expected    []*AgencyInfo
		description string
	}{
		{
			name:    "single agency with agency_id",
			content: "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,http://metro.example,America/Los_Angeles",
			expected: []*AgencyInfo{
				{AgencyID: "1", AgencyName: "Metro", RowNumber: 2},
			},
			description: "Single agency should be loaded correctly",
		},
		{
			name:    "single agency without agency_id",
			content: "agency_name,agency_url,agency_timezone\nMetro,http://metro.example,America/Los_Angeles",
			expected: []*AgencyInfo{
				{AgencyID: "", AgencyName: "Metro", RowNumber: 2},
			},
			description: "Agency without agency_id should use empty string",
		},
		{
			name:    "multiple agencies",
			content: "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,http://metro.example,America/Los_Angeles\n2,Bus,http://bus.example,America/Los_Angeles",
			expected: []*AgencyInfo{
				{AgencyID: "1", AgencyName: "Metro", RowNumber: 2},
				{AgencyID: "2", AgencyName: "Bus", RowNumber: 3},
			},
			description: "Multiple agencies should be loaded correctly",
		},
		{
			name:    "agency with whitespace padding",
			content: "agency_id,agency_name,agency_url,agency_timezone\n 1 , Metro ,http://metro.example,America/Los_Angeles",
			expected: []*AgencyInfo{
				{AgencyID: "1", AgencyName: "Metro", RowNumber: 2},
			},
			description: "Whitespace should be trimmed from agency_id and agency_name",
		},
		{
			name:    "agency with empty agency_id",
			content: "agency_id,agency_name,agency_url,agency_timezone\n,Metro,http://metro.example,America/Los_Angeles",
			expected: []*AgencyInfo{
				{AgencyID: "", AgencyName: "Metro", RowNumber: 2},
			},
			description: "Empty agency_id should use empty string",
		},
		{
			name:    "agency with missing agency_name",
			content: "agency_id,agency_url,agency_timezone\n1,http://metro.example,America/Los_Angeles",
			expected: []*AgencyInfo{
				{AgencyID: "1", AgencyName: "", RowNumber: 2},
			},
			description: "Missing agency_name should result in empty name",
		},
		{
			name:        "empty file",
			content:     "agency_id,agency_name,agency_url,agency_timezone",
			expected:    []*AgencyInfo{},
			description: "Empty file should return empty slice",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files := map[string]string{"agency.txt": tt.content}
			loader := CreateTestFeedLoader(t, files)
			validator := NewAgencyConsistencyValidator()

			agencies := validator.loadAgencies(loader)

			if len(agencies) != len(tt.expected) {
				t.Errorf("Expected %d agencies, got %d", len(tt.expected), len(agencies))
			}

			for i, expectedAgency := range tt.expected {
				if i >= len(agencies) {
					t.Errorf("Expected agency at index %d not found", i)
					continue
				}

				actualAgency := agencies[i]
				if actualAgency.AgencyID != expectedAgency.AgencyID {
					t.Errorf("Expected AgencyID '%s', got '%s'", expectedAgency.AgencyID, actualAgency.AgencyID)
				}

				if actualAgency.AgencyName != expectedAgency.AgencyName {
					t.Errorf("Expected AgencyName '%s', got '%s'", expectedAgency.AgencyName, actualAgency.AgencyName)
				}

				if actualAgency.RowNumber != expectedAgency.RowNumber {
					t.Errorf("Expected RowNumber %d, got %d", expectedAgency.RowNumber, actualAgency.RowNumber)
				}
			}
		})
	}
}

func TestAgencyConsistencyValidator_ValidateAgencyIdRequirement(t *testing.T) {
	tests := []struct {
		name                string
		agencies            []*AgencyInfo
		expectedNoticeCount int
		description         string
	}{
		{
			name: "single agency without agency_id",
			agencies: []*AgencyInfo{
				{AgencyID: "", AgencyName: "Metro", RowNumber: 2},
			},
			expectedNoticeCount: 0,
			description:         "Single agency without agency_id should be valid",
		},
		{
			name: "single agency with agency_id",
			agencies: []*AgencyInfo{
				{AgencyID: "1", AgencyName: "Metro", RowNumber: 2},
			},
			expectedNoticeCount: 0,
			description:         "Single agency with agency_id should be valid",
		},
		{
			name: "multiple agencies all with agency_id",
			agencies: []*AgencyInfo{
				{AgencyID: "1", AgencyName: "Metro", RowNumber: 2},
				{AgencyID: "2", AgencyName: "Bus", RowNumber: 3},
			},
			expectedNoticeCount: 0,
			description:         "Multiple agencies with agency_ids should be valid",
		},
		{
			name: "multiple agencies with one missing agency_id",
			agencies: []*AgencyInfo{
				{AgencyID: "1", AgencyName: "Metro", RowNumber: 2},
				{AgencyID: "", AgencyName: "Bus", RowNumber: 2},
			},
			expectedNoticeCount: 1,
			description:         "Multiple agencies with missing agency_id should generate notice",
		},
		{
			name: "multiple agencies with multiple missing agency_ids",
			agencies: []*AgencyInfo{
				{AgencyID: "1", AgencyName: "Metro", RowNumber: 2},
				{AgencyID: "", AgencyName: "Bus", RowNumber: 2},
				{AgencyID: "", AgencyName: "Rail", RowNumber: 3},
			},
			expectedNoticeCount: 2,
			description:         "Multiple agencies with missing agency_ids should generate notices",
		},
		{
			name:                "no agencies",
			agencies:            []*AgencyInfo{},
			expectedNoticeCount: 0,
			description:         "No agencies should not generate notices",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			container := notice.NewNoticeContainer()
			validator := NewAgencyConsistencyValidator()

			validator.validateAgencyIdRequirement(container, tt.agencies)

			notices := container.GetNotices()
			if len(notices) != tt.expectedNoticeCount {
				t.Errorf("Expected %d notices, got %d for %s", tt.expectedNoticeCount, len(notices), tt.description)
			}
		})
	}
}

func TestAgencyConsistencyValidator_ValidateRouteAgencyReferences(t *testing.T) {
	tests := []struct {
		name                string
		routesContent       string
		agencies            []*AgencyInfo
		expectedNoticeCount int
		expectedCodes       []string
		description         string
	}{
		{
			name:          "valid route references",
			routesContent: "route_id,agency_id,route_short_name,route_type\nR1,1,Red,3\nR2,2,Blue,3",
			agencies: []*AgencyInfo{
				{AgencyID: "1", AgencyName: "Metro", RowNumber: 2},
				{AgencyID: "2", AgencyName: "Bus", RowNumber: 3},
			},
			expectedNoticeCount: 0,
			expectedCodes:       []string{},
			description:         "Valid route agency references should not generate notices",
		},
		{
			name:          "invalid route reference",
			routesContent: "route_id,agency_id,route_short_name,route_type\nR1,999,Red,3",
			agencies: []*AgencyInfo{
				{AgencyID: "1", AgencyName: "Metro", RowNumber: 2},
			},
			expectedNoticeCount: 1,
			expectedCodes:       []string{"invalid_agency_reference"},
			description:         "Invalid route agency reference should generate notice",
		},
		{
			name:          "route without agency_id with single agency",
			routesContent: "route_id,route_short_name,route_type\nR1,Red,3",
			agencies: []*AgencyInfo{
				{AgencyID: "1", AgencyName: "Metro", RowNumber: 2},
			},
			expectedNoticeCount: 0,
			expectedCodes:       []string{},
			description:         "Route without agency_id should be valid with single agency",
		},
		{
			name:          "route without agency_id with multiple agencies",
			routesContent: "route_id,route_short_name,route_type\nR1,Red,3",
			agencies: []*AgencyInfo{
				{AgencyID: "1", AgencyName: "Metro", RowNumber: 2},
				{AgencyID: "2", AgencyName: "Bus", RowNumber: 3},
			},
			expectedNoticeCount: 1,
			expectedCodes:       []string{"missing_route_agency_id"},
			description:         "Route without agency_id should generate notice with multiple agencies",
		},
		{
			name:          "route without route_id",
			routesContent: "route_id,agency_id,route_short_name,route_type\n,1,Red,3",
			agencies: []*AgencyInfo{
				{AgencyID: "1", AgencyName: "Metro", RowNumber: 2},
			},
			expectedNoticeCount: 0,
			expectedCodes:       []string{},
			description:         "Route without route_id should be ignored",
		},
		{
			name:          "empty routes file",
			routesContent: "route_id,agency_id,route_short_name,route_type",
			agencies: []*AgencyInfo{
				{AgencyID: "1", AgencyName: "Metro", RowNumber: 2},
			},
			expectedNoticeCount: 0,
			expectedCodes:       []string{},
			description:         "Empty routes file should not generate notices",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files := map[string]string{"routes.txt": tt.routesContent}
			loader := CreateTestFeedLoader(t, files)
			container := notice.NewNoticeContainer()
			validator := NewAgencyConsistencyValidator()

			validator.validateRouteAgencyReferences(loader, container, tt.agencies)

			notices := container.GetNotices()
			if len(notices) != tt.expectedNoticeCount {
				t.Errorf("Expected %d notices, got %d for %s", tt.expectedNoticeCount, len(notices), tt.description)
			}

			// Check notice codes
			actualCodes := make([]string, len(notices))
			for i, notice := range notices {
				actualCodes[i] = notice.Code()
			}

			if len(actualCodes) != len(tt.expectedCodes) {
				t.Errorf("Expected %d notice codes, got %d", len(tt.expectedCodes), len(actualCodes))
			}

			for i, expectedCode := range tt.expectedCodes {
				if i >= len(actualCodes) || actualCodes[i] != expectedCode {
					t.Errorf("Expected notice code '%s' at index %d, got '%v'", expectedCode, i, actualCodes)
				}
			}
		})
	}
}

func TestAgencyConsistencyValidator_New(t *testing.T) {
	validator := NewAgencyConsistencyValidator()
	if validator == nil {
		t.Error("NewAgencyConsistencyValidator() returned nil")
	}
}
