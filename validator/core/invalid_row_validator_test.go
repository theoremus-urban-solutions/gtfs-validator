package core

import (
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/testutil"
	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

// This validator checks a row's shape only. What the fields contain is
// covered by TestFieldTypeValidator_Validate.
func TestInvalidRowValidator_Validate(t *testing.T) {
	tests := []struct {
		name                string
		files               map[string]string
		expectedNoticeCodes []string
		description         string
	}{
		{
			name: "row matches its header",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,http://metro.example,America/Los_Angeles",
			},
			expectedNoticeCodes: []string{},
			description:         "A row with as many fields as the header is well formed",
		},
		{
			name: "row has too few fields",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,http://metro.example",
			},
			expectedNoticeCodes: []string{"invalid_row_length"},
			description:         "A short row leaves a field silently empty",
		},
		{
			name: "row has too many fields",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,http://metro.example,America/Los_Angeles,extra",
			},
			expectedNoticeCodes: []string{"invalid_row_length"},
			description:         "An extra field has no column to belong to",
		},
		{
			name: "field contents are not this validator's business",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_lat,stop_lon,location_type\n1,Main St,34.05,-118.25,5",
			},
			expectedNoticeCodes: []string{},
			description:         "An out-of-range enum is reported by the field type validator",
		},
		{
			name: "missing files are ignored",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,http://metro.example,America/Los_Angeles",
			},
			expectedNoticeCodes: []string{},
			description:         "Absent optional files are the missing-files validator's business",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := testutil.CreateTestFeedLoader(t, tt.files)
			container := notice.NewNoticeContainer()
			validator := NewInvalidRowValidator()

			validator.Validate(loader, container, gtfsvalidator.Config{})

			assertNoticeCodes(t, container, tt.expectedNoticeCodes, tt.description)
		})
	}
}

func TestInvalidRowValidator_New(t *testing.T) {
	if NewInvalidRowValidator() == nil {
		t.Error("NewInvalidRowValidator() returned nil")
	}
}

// assertNoticeCodes compares the notices a validator produced against the
// codes a case expects, counting duplicates.
func assertNoticeCodes(t *testing.T, container *notice.NoticeContainer, expected []string, description string) {
	t.Helper()

	notices := container.GetNotices()
	if len(notices) != len(expected) {
		codes := make([]string, 0, len(notices))
		for _, n := range notices {
			codes = append(codes, n.Code())
		}
		t.Errorf("expected %d notices %v, got %d %v: %s", len(expected), expected, len(notices), codes, description)
	}

	expectedCounts := make(map[string]int)
	for _, code := range expected {
		expectedCounts[code]++
	}
	actualCounts := make(map[string]int)
	for _, n := range notices {
		actualCounts[n.Code()]++
	}

	for code, want := range expectedCounts {
		if actualCounts[code] != want {
			t.Errorf("expected %d %q, got %d", want, code, actualCounts[code])
		}
	}
	for code := range actualCounts {
		if expectedCounts[code] == 0 {
			t.Errorf("unexpected notice code %q", code)
		}
	}
}
