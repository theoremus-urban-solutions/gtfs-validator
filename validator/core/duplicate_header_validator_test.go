package core

import (
	"github.com/theoremus-urban-solutions/gtfs-validator/testutil"
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

const duplicateHeaderCode = "duplicate_header"

func TestDuplicateHeaderValidator_Validate(t *testing.T) {
	tests := []struct {
		name                string
		files               map[string]string
		expectedNoticeCodes []string
		description         string
	}{
		{
			name: "no duplicate headers",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,http://metro.example,America/Los_Angeles",
				"stops.txt":  "stop_id,stop_name,stop_lat,stop_lon\n1,Main St,34.05,-118.25",
			},
			expectedNoticeCodes: []string{},
			description:         "All files have unique headers",
		},
		{
			name: "single file with duplicate header",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_id,agency_timezone\n1,Metro,1,America/Los_Angeles",
			},
			expectedNoticeCodes: []string{duplicateHeaderCode},
			description:         "agency.txt has duplicate agency_id header",
		},
		{
			name: "multiple files with duplicate headers",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_id,agency_timezone\n1,Metro,1,America/Los_Angeles",
				"stops.txt":  "stop_id,stop_name,stop_lat,stop_lat\n1,Main St,34.05,-118.25",
			},
			expectedNoticeCodes: []string{duplicateHeaderCode, duplicateHeaderCode},
			description:         "Both files have duplicate headers",
		},
		{
			name: "multiple duplicates in single file",
			files: map[string]string{
				"routes.txt": "route_id,route_id,agency_id,route_short_name,route_short_name,route_type\n1,1,1,Red,Red,3",
			},
			expectedNoticeCodes: []string{duplicateHeaderCode, duplicateHeaderCode},
			description:         "Single file with multiple duplicate header pairs",
		},
		{
			name: "triplicate header",
			files: map[string]string{
				"trips.txt": "trip_id,trip_id,trip_id,route_id,service_id\nT1,T1,T1,R1,S1",
			},
			expectedNoticeCodes: []string{duplicateHeaderCode},
			description:         "Single header appears three times",
		},
		{
			name: "headers with whitespace",
			files: map[string]string{
				"agency.txt": "agency_id, agency_id ,agency_name,agency_timezone\n1,1,Metro,America/Los_Angeles",
			},
			expectedNoticeCodes: []string{duplicateHeaderCode},
			description:         "Duplicate headers with different whitespace should be detected",
		},
		{
			name: "case sensitive duplicates",
			files: map[string]string{
				"agency.txt": "agency_id,Agency_ID,agency_name,agency_timezone\n1,1,Metro,America/Los_Angeles",
			},
			expectedNoticeCodes: []string{},
			description:         "Headers with different cases should not be considered duplicates",
		},
		{
			name: "empty header duplicates",
			files: map[string]string{
				"custom.txt": "field1,,field2,\nvalue1,,value2,",
			},
			expectedNoticeCodes: []string{duplicateHeaderCode},
			description:         "Empty headers should be detected as duplicates",
		},
		{
			name: "mixed valid and invalid files",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,http://metro.example,America/Los_Angeles",
				"stops.txt":  "stop_id,stop_name,stop_id,stop_lon\n1,Main St,1,-118.25",
				"routes.txt": "route_id,agency_id,route_short_name,route_type\n1,1,Red,3",
			},
			expectedNoticeCodes: []string{duplicateHeaderCode},
			description:         "Mix of files with and without duplicate headers",
		},
		{
			name: "headers only file",
			files: map[string]string{
				"test.txt": "field1,field1,field2",
			},
			expectedNoticeCodes: []string{duplicateHeaderCode},
			description:         "File with only headers and duplicates",
		},
		{
			name: "complex duplicate pattern",
			files: map[string]string{
				"complex.txt": "a,b,a,c,b,d,a\nval1,val2,val3,val4,val5,val6,val7",
			},
			expectedNoticeCodes: []string{duplicateHeaderCode, duplicateHeaderCode},
			description:         "Multiple headers with various duplication patterns",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test components
			loader := testutil.CreateTestFeedLoader(t, tt.files)
			container := notice.NewNoticeContainer()
			validator := NewDuplicateHeaderValidator()
			config := gtfsvalidator.Config{}

			// Run validation
			validator.Validate(loader, container, config)

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

func TestDuplicateHeaderValidator_ValidateFileHeaders(t *testing.T) {
	tests := []struct {
		name              string
		filename          string
		content           string
		expectDuplicates  bool
		expectedHeader    string
		expectedPositions []int
	}{
		{
			name:             "no duplicates",
			filename:         "agency.txt",
			content:          "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,http://metro.example,America/Los_Angeles",
			expectDuplicates: false,
		},
		{
			name:              "simple duplicate",
			filename:          "stops.txt",
			content:           "stop_id,stop_name,stop_id,stop_lon\n1,Main St,1,-118.25",
			expectDuplicates:  true,
			expectedHeader:    "stop_id",
			expectedPositions: []int{0, 2},
		},
		{
			name:              "duplicate with whitespace",
			filename:          "routes.txt",
			content:           "route_id, route_id ,agency_id,route_type\n1,1,1,3",
			expectDuplicates:  true,
			expectedHeader:    "route_id",
			expectedPositions: []int{0, 1},
		},
		{
			name:              "triple duplicate",
			filename:          "trips.txt",
			content:           "trip_id,trip_id,trip_id,route_id\nT1,T1,T1,R1",
			expectDuplicates:  true,
			expectedHeader:    "trip_id",
			expectedPositions: []int{0, 1, 2},
		},
		{
			name:              "empty header duplicates",
			filename:          "test.txt",
			content:           "field1,,field2,\nval1,,val2,",
			expectDuplicates:  true,
			expectedHeader:    "",
			expectedPositions: []int{1, 3},
		},
		{
			name:             "case sensitive - no duplicates",
			filename:         "case.txt",
			content:          "Field,field,FIELD\nval1,val2,val3",
			expectDuplicates: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test components
			files := map[string]string{tt.filename: tt.content}
			loader := testutil.CreateTestFeedLoader(t, files)
			container := notice.NewNoticeContainer()
			validator := NewDuplicateHeaderValidator()

			// Run validation on specific file
			validator.validateFileHeaders(loader, container, tt.filename)

			// Get notices
			notices := container.GetNotices()

			if tt.expectDuplicates {
				if len(notices) == 0 {
					t.Error("Expected duplicate header notice, but got none")
					return
				}

				// Find the duplicate header notice
				found := false
				for _, n := range notices {
					if n.Code() == duplicateHeaderCode {
						found = true
						context := n.Context()

						// Check filename
						if filename, ok := context["filename"]; !ok || filename != tt.filename {
							t.Errorf("Expected filename '%s' in context, got '%v'", tt.filename, filename)
						}

						// Check header name
						if headerName, ok := context["headerName"]; !ok || headerName != tt.expectedHeader {
							t.Errorf("Expected header name '%s' in context, got '%v'", tt.expectedHeader, headerName)
						}

						// Check positions
						if positions, ok := context["positions"]; ok {
							if posSlice, ok := positions.([]int); ok {
								if len(posSlice) != len(tt.expectedPositions) {
									t.Errorf("Expected %d positions, got %d", len(tt.expectedPositions), len(posSlice))
								} else {
									for i, expectedPos := range tt.expectedPositions {
										if i < len(posSlice) && posSlice[i] != expectedPos {
											t.Errorf("Expected position %d at index %d, got %d", expectedPos, i, posSlice[i])
										}
									}
								}
							} else {
								t.Errorf("Expected positions to be []int, got %T", positions)
							}
						} else {
							t.Error("Expected positions in context")
						}
						break
					}
				}

				if !found {
					t.Error("Expected " + duplicateHeaderCode + " notice but didn't find one")
				}
			} else {
				// Should not have any duplicate header notices
				for _, n := range notices {
					if n.Code() == duplicateHeaderCode {
						t.Error("Did not expect duplicate header notice, but got one")
					}
				}
			}
		})
	}
}

func TestDuplicateHeaderValidator_New(t *testing.T) {
	validator := NewDuplicateHeaderValidator()
	if validator == nil {
		t.Error("NewDuplicateHeaderValidator() returned nil")
	}
}

func TestDuplicateHeaderValidator_FileNotExists(t *testing.T) {
	// Test behavior when file doesn't exist
	loader := testutil.CreateTestFeedLoader(t, map[string]string{}) // No files
	container := notice.NewNoticeContainer()
	validator := NewDuplicateHeaderValidator()

	// Try to validate a non-existent file
	validator.validateFileHeaders(loader, container, "nonexistent.txt")

	// Should not generate any notices (other validators handle missing files)
	notices := container.GetNotices()
	if len(notices) != 0 {
		t.Errorf("Expected no notices for non-existent file, got %d", len(notices))
	}
}

func TestDuplicateHeaderValidator_MalformedCSV(t *testing.T) {
	// Test behavior with malformed CSV content
	files := map[string]string{
		"malformed.txt": "field1,field2\n\"unclosed quote", // Malformed CSV
	}
	loader := testutil.CreateTestFeedLoader(t, files)
	container := notice.NewNoticeContainer()
	validator := NewDuplicateHeaderValidator()

	validator.validateFileHeaders(loader, container, "malformed.txt")

	// Should not generate duplicate header notices (CSV parsing errors handled elsewhere)
	notices := container.GetNotices()
	for _, notice := range notices {
		if notice.Code() == duplicateHeaderCode {
			t.Error("Should not generate " + duplicateHeaderCode + " notice for malformed CSV")
		}
	}
}

func TestDuplicateHeaderValidator_WhitespaceHandling(t *testing.T) {
	tests := []struct {
		name            string
		content         string
		expectDuplicate bool
		description     string
	}{
		{
			name:            "leading whitespace",
			content:         "field1, field1,field2\nval1,val2,val3",
			expectDuplicate: true,
			description:     "Headers with leading whitespace should be trimmed and detected as duplicates",
		},
		{
			name:            "trailing whitespace",
			content:         "field1,field1 ,field2\nval1,val2,val3",
			expectDuplicate: true,
			description:     "Headers with trailing whitespace should be trimmed and detected as duplicates",
		},
		{
			name:            "both leading and trailing whitespace",
			content:         "field1, field1 ,field2\nval1,val2,val3",
			expectDuplicate: true,
			description:     "Headers with both leading and trailing whitespace should be trimmed",
		},
		{
			name:            "different whitespace",
			content:         "field1,\tfield1\t,field2\nval1,val2,val3",
			expectDuplicate: true,
			description:     "Headers with tabs should be trimmed and detected as duplicates",
		},
		{
			name:            "no duplicates with different content",
			content:         "field1,field2,field3\nval1,val2,val3",
			expectDuplicate: false,
			description:     "Different headers should not be flagged",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files := map[string]string{"test.txt": tt.content}
			loader := testutil.CreateTestFeedLoader(t, files)
			container := notice.NewNoticeContainer()
			validator := NewDuplicateHeaderValidator()

			validator.validateFileHeaders(loader, container, "test.txt")

			notices := container.GetNotices()
			hasDuplicate := false
			for _, notice := range notices {
				if notice.Code() == duplicateHeaderCode {
					hasDuplicate = true
					break
				}
			}

			if hasDuplicate != tt.expectDuplicate {
				t.Errorf("%s: expected duplicate=%v, got duplicate=%v", tt.description, tt.expectDuplicate, hasDuplicate)
			}
		})
	}
}
