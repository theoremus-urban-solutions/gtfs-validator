package entity

import (
	"github.com/theoremus-urban-solutions/gtfs-validator/testutil"
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/parser"
	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

func TestAttributionWithoutRoleValidator_Validate(t *testing.T) {
	tests := []struct {
		name                string
		files               map[string]string
		expectedNoticeCodes []string
		description         string
	}{
		{
			name: "valid attribution with single role",
			files: map[string]string{
				"attributions.txt": "attribution_id,organization_name,is_producer,is_operator,is_authority\n" +
					"attr1,Simple Transport Company,0,1,0",
			},
			expectedNoticeCodes: []string{},
			description:         "Attribution with operator role should be valid",
		},
		{
			name: "valid attribution with multiple roles",
			files: map[string]string{
				"attributions.txt": "attribution_id,organization_name,is_producer,is_operator,is_authority\n" +
					"attr1,Simple Company,0,1,1",
			},
			expectedNoticeCodes: []string{},
			description:         "Attribution with multiple roles should be valid",
		},
		{
			name: "attribution without any roles",
			files: map[string]string{
				"attributions.txt": "attribution_id,organization_name,is_producer,is_operator,is_authority\n" +
					"attr1,Simple Company,0,0,0",
			},
			expectedNoticeCodes: []string{"attribution_without_role"},
			description:         "Attribution without any roles should generate error",
		},
		{
			name: "attribution with all roles",
			files: map[string]string{
				"attributions.txt": "attribution_id,organization_name,is_producer,is_operator,is_authority\n" +
					"attr1,Simple Company,1,1,1",
			},
			expectedNoticeCodes: []string{"attribution_all_roles"},
			description:         "Attribution with all three roles should generate info notice",
		},
		{
			name: "multiple attributions mixed validity",
			files: map[string]string{
				"attributions.txt": "attribution_id,organization_name,is_producer,is_operator,is_authority\n" +
					"attr1,First Company,0,1,0\n" +
					"attr2,Second Company,0,0,0\n" +
					"attr3,Third Company,1,1,1",
			},
			expectedNoticeCodes: []string{"attribution_without_role", "attribution_all_roles"},
			description:         "Multiple attributions with mixed validity",
		},
		{
			name: "attribution without attribution_id",
			files: map[string]string{
				"attributions.txt": "organization_name,is_producer,is_operator,is_authority\n" +
					"Simple Company,0,1,0",
			},
			expectedNoticeCodes: []string{},
			description:         "Attribution without attribution_id but with role should be valid",
		},
		{
			name: "operator name with operator role mismatch",
			files: map[string]string{
				"attributions.txt": "attribution_id,organization_name,is_producer,is_operator,is_authority\n" +
					"attr1,Metro Transit Operator,1,0,0", // Has "operator" in name but is producer
			},
			expectedNoticeCodes: []string{"attribution_role_name_mismatch"},
			description:         "Organization name suggests operator but assigned different role",
		},
		{
			name: "authority name with authority role mismatch",
			files: map[string]string{
				"attributions.txt": "attribution_id,organization_name,is_producer,is_operator,is_authority\n" +
					"attr1,City Transport Authority,0,1,0", // Has "authority" in name but is operator
			},
			expectedNoticeCodes: []string{"attribution_role_name_mismatch"},
			description:         "Organization name suggests authority but assigned different role",
		},
		{
			name: "producer name with producer role mismatch",
			files: map[string]string{
				"attributions.txt": "attribution_id,organization_name,is_producer,is_operator,is_authority\n" +
					"attr1,Data Systems Solutions,0,0,1", // Has "systems" in name but is authority
			},
			expectedNoticeCodes: []string{"attribution_role_name_mismatch"},
			description:         "Organization name suggests producer but assigned different role",
		},
		{
			name: "matching operator name and role",
			files: map[string]string{
				"attributions.txt": "attribution_id,organization_name,is_producer,is_operator,is_authority\n" +
					"attr1,Metro Transit Operator,0,1,0",
			},
			expectedNoticeCodes: []string{},
			description:         "Operator name with operator role should not generate mismatch notice",
		},
		{
			name: "matching authority name and role",
			files: map[string]string{
				"attributions.txt": "attribution_id,organization_name,is_producer,is_operator,is_authority\n" +
					"attr1,Regional Authority,0,0,1",
			},
			expectedNoticeCodes: []string{},
			description:         "Authority name with authority role should not generate mismatch notice",
		},
		{
			name: "matching producer name and role",
			files: map[string]string{
				"attributions.txt": "attribution_id,organization_name,is_producer,is_operator,is_authority\n" +
					"attr1,Data Systems Solutions,1,0,0",
			},
			expectedNoticeCodes: []string{},
			description:         "Producer name with producer role should not generate mismatch notice",
		},
		{
			name: "empty role fields treated as no role",
			files: map[string]string{
				"attributions.txt": "attribution_id,organization_name,is_producer,is_operator,is_authority\n" +
					"attr1,Simple Company,,,",
			},
			expectedNoticeCodes: []string{"attribution_without_role"},
			description:         "Empty role fields should be treated as no roles assigned",
		},
		{
			name: "role fields with 0 values",
			files: map[string]string{
				"attributions.txt": "attribution_id,organization_name,is_producer,is_operator,is_authority\n" +
					"attr1,Simple Company,0,0,0",
			},
			expectedNoticeCodes: []string{"attribution_without_role"},
			description:         "Role fields with explicit 0 values should generate error",
		},
		{
			name: "role fields with 1 values",
			files: map[string]string{
				"attributions.txt": "attribution_id,organization_name,is_producer,is_operator,is_authority\n" +
					"attr1,Simple Company,1,0,1",
			},
			expectedNoticeCodes: []string{},
			description:         "Role fields with 1 values should be valid",
		},
		{
			name: "role fields with invalid values treated as 0",
			files: map[string]string{
				"attributions.txt": "attribution_id,organization_name,is_producer,is_operator,is_authority\n" +
					"attr1,Simple Company,yes,no,maybe",
			},
			expectedNoticeCodes: []string{"attribution_without_role"},
			description:         "Invalid role values should be treated as 0/false",
		},
		{
			name: "missing organization_name rows ignored",
			files: map[string]string{
				"attributions.txt": "attribution_id,organization_name,is_producer,is_operator,is_authority\n" +
					"attr1,,1,0,0\n" +
					"attr2,Valid Org,0,1,0",
			},
			expectedNoticeCodes: []string{},
			description:         "Rows without organization_name should be ignored",
		},
		{
			name: "no attributions file",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,http://metro.example,America/Los_Angeles",
			},
			expectedNoticeCodes: []string{},
			description:         "Missing attributions.txt file should not generate errors",
		},
		{
			name: "complex names with multiple keywords",
			files: map[string]string{
				"attributions.txt": "attribution_id,organization_name,is_producer,is_operator,is_authority\n" +
					"attr1,Complex Business Name,0,1,0", // Generic name without specific keywords
			},
			expectedNoticeCodes: []string{},
			description:         "Complex organization names should not generate false positives",
		},
		{
			name: "case insensitive keyword matching",
			files: map[string]string{
				"attributions.txt": "attribution_id,organization_name,is_producer,is_operator,is_authority\n" +
					"attr1,METRO TRANSPORT OPERATOR,1,0,0", // Uppercase should still match
			},
			expectedNoticeCodes: []string{"attribution_role_name_mismatch"},
			description:         "Keyword matching should be case insensitive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test feed loader
			feedLoader := testutil.CreateTestFeedLoader(t, tt.files)

			// Create notice container and validator
			container := notice.NewNoticeContainer()
			validator := NewAttributionWithoutRoleValidator()
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

func TestAttributionWithoutRoleValidator_ParseAttribution(t *testing.T) {
	tests := []struct {
		name     string
		rowData  map[string]string
		expected *AttributionRoleInfo
	}{
		{
			name: "all fields present and valid",
			rowData: map[string]string{
				"attribution_id":    "attr1",
				"organization_name": "Metro Transit",
				"is_producer":       "1",
				"is_operator":       "0",
				"is_authority":      "1",
			},
			expected: &AttributionRoleInfo{
				AttributionID:    "attr1",
				OrganizationName: "Metro Transit",
				IsProducer:       true,
				IsOperator:       false,
				IsAuthority:      true,
				RowNumber:        1,
			},
		},
		{
			name: "missing attribution_id",
			rowData: map[string]string{
				"organization_name": "Metro Transit",
				"is_operator":       "1",
			},
			expected: &AttributionRoleInfo{
				AttributionID:    "",
				OrganizationName: "Metro Transit",
				IsProducer:       false,
				IsOperator:       true,
				IsAuthority:      false,
				RowNumber:        1,
			},
		},
		{
			name: "missing organization_name",
			rowData: map[string]string{
				"attribution_id": "attr1",
				"is_producer":    "1",
			},
			expected: nil, // Should return nil when organization_name is missing
		},
		{
			name: "invalid role values",
			rowData: map[string]string{
				"organization_name": "Metro Transit",
				"is_producer":       "invalid",
				"is_operator":       "2", // Not 1, should be false
				"is_authority":      "",  // Empty, should be false
			},
			expected: &AttributionRoleInfo{
				AttributionID:    "",
				OrganizationName: "Metro Transit",
				IsProducer:       false,
				IsOperator:       false,
				IsAuthority:      false,
				RowNumber:        1,
			},
		},
	}

	validator := NewAttributionWithoutRoleValidator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock CSV row
			row := &parser.CSVRow{
				RowNumber: 1,
				Values:    tt.rowData,
			}

			result := validator.parseAttribution(row)

			if tt.expected == nil {
				if result != nil {
					t.Errorf("Expected nil result, got %+v", result)
				}
				return
			}

			if result == nil {
				t.Errorf("Expected result %+v, got nil", tt.expected)
				return
			}

			// Compare all fields
			if result.AttributionID != tt.expected.AttributionID {
				t.Errorf("AttributionID: expected '%s', got '%s'", tt.expected.AttributionID, result.AttributionID)
			}
			if result.OrganizationName != tt.expected.OrganizationName {
				t.Errorf("OrganizationName: expected '%s', got '%s'", tt.expected.OrganizationName, result.OrganizationName)
			}
			if result.IsProducer != tt.expected.IsProducer {
				t.Errorf("IsProducer: expected %v, got %v", tt.expected.IsProducer, result.IsProducer)
			}
			if result.IsOperator != tt.expected.IsOperator {
				t.Errorf("IsOperator: expected %v, got %v", tt.expected.IsOperator, result.IsOperator)
			}
			if result.IsAuthority != tt.expected.IsAuthority {
				t.Errorf("IsAuthority: expected %v, got %v", tt.expected.IsAuthority, result.IsAuthority)
			}
		})
	}
}

func TestAttributionWithoutRoleValidator_ContainsAnyKeyword(t *testing.T) {
	validator := NewAttributionWithoutRoleValidator()

	tests := []struct {
		text     string
		keywords []string
		expected bool
	}{
		{"metro transit operator", []string{"operator", "transport"}, true}, // Lowercase for accurate testing
		{"city authority", []string{"authority", "government"}, true},
		{"data systems inc", []string{"data", "technology"}, true},
		{"simple company", []string{"operator", "authority"}, false},
		{"", []string{"keyword"}, false},
		{"text", []string{}, false},
		{"bus transport service", []string{"bus", "metro"}, true}, // Should match 'bus'
		{"railway company", []string{"railway", "train"}, true},   // Should match 'railway'
	}

	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			result := validator.containsAnyKeyword(tt.text, tt.keywords)
			if result != tt.expected {
				t.Errorf("containsAnyKeyword('%s', %v) = %v, expected %v", tt.text, tt.keywords, result, tt.expected)
			}
		})
	}
}

func TestAttributionWithoutRoleValidator_New(t *testing.T) {
	validator := NewAttributionWithoutRoleValidator()
	if validator == nil {
		t.Error("NewAttributionWithoutRoleValidator() returned nil")
	}
}
