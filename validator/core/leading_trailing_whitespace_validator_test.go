package core

import (
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/testutil"
	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

// The rule turns on whether a value was quoted, not on which field it is in.
// Unquoted padding is CSV layout that any reader may strip; quoted padding is
// asserted to be part of the value.
func TestLeadingTrailingWhitespaceValidator_OnlyReportsQuotedValues(t *testing.T) {
	tests := []struct {
		name     string
		stops    string
		expected int
	}{
		{
			name:     "unquoted padding is layout, not content",
			stops:    "stop_id,stop_name,stop_lat,stop_lon\nS1, First,40.0,-70.0\nS2,Second ,41.0,-71.0",
			expected: 0,
		},
		{
			name:     "quoted padding is part of the value",
			stops:    "stop_id,stop_name,stop_lat,stop_lon\nS1,\" First\",40.0,-70.0\nS2,\"Second \",41.0,-71.0",
			expected: 2,
		},
		{
			name:     "quoted id padding breaks joins and is reported",
			stops:    "stop_id,stop_name,stop_lat,stop_lon\n\" S1\",First,40.0,-70.0\nS2,Second,41.0,-71.0",
			expected: 1,
		},
		{
			name:     "a quoted value with no padding is fine",
			stops:    "stop_id,stop_name,stop_lat,stop_lon\nS1,\"First, and only\",40.0,-70.0",
			expected: 0,
		},
		{
			name:     "one notice covers a value padded at both ends",
			stops:    "stop_id,stop_name,stop_lat,stop_lon\nS1,\" First \",40.0,-70.0",
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := testutil.CreateTestFeedLoader(t, map[string]string{"stops.txt": tt.stops})
			container := notice.NewNoticeContainer()

			NewLeadingTrailingWhitespaceValidator().Validate(loader, container, gtfsvalidator.Config{})

			got := 0
			for _, n := range container.GetNotices() {
				if n.Code() == "leading_or_trailing_whitespaces" {
					got++
				}
			}
			if got != tt.expected {
				t.Errorf("expected %d leading_or_trailing_whitespaces, got %d", tt.expected, got)
			}
		})
	}
}

// A quoted field may hold commas, escaped quotes and newlines. The scanner has
// to agree with a real CSV reader about where fields end, or it would report
// whitespace that is not there.
func TestLeadingTrailingWhitespaceValidator_HandlesCSVGrammar(t *testing.T) {
	stops := "stop_id,stop_name,stop_lat,stop_lon\n" +
		"S1,\"Has, comma\",40.0,-70.0\n" +
		"S2,\"Has \"\"quote\"\"\",41.0,-71.0\n" +
		"S3,\"Has\nnewline\",42.0,-72.0"

	loader := testutil.CreateTestFeedLoader(t, map[string]string{"stops.txt": stops})
	container := notice.NewNoticeContainer()

	NewLeadingTrailingWhitespaceValidator().Validate(loader, container, gtfsvalidator.Config{})

	for _, n := range container.GetNotices() {
		if n.Code() == "leading_or_trailing_whitespaces" {
			t.Errorf("no value is padded, but got a notice: %v", n.Context())
		}
	}
}

func TestLeadingTrailingWhitespaceValidator_New(t *testing.T) {
	if NewLeadingTrailingWhitespaceValidator() == nil {
		t.Error("expected a validator")
	}
}
