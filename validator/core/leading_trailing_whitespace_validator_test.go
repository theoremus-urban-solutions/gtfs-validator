package core

import (
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/testutil"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

func TestLeadingTrailingWhitespaceValidator(t *testing.T) {
	tests := []struct {
		name     string
		files    map[string]string
		expected int
		values   []string
	}{
		{
			name: "clean file",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,http://metro.example,America/Los_Angeles\n",
			},
			expected: 0,
		},
		{
			name: "unquoted whitespace is the parser's, not the feed's",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\n 1 , Metro ,http://metro.example,America/Los_Angeles\n",
			},
			expected: 0,
		},
		{
			name: "quoted whitespace survives parsing and is reported",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\n1,\" Metro \",http://metro.example,America/Los_Angeles\n",
			},
			expected: 1,
			values:   []string{" Metro "},
		},
		{
			name: "one notice per value, not per end",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_lat,stop_lon\n\" S1 \",\"\tCentral\n\",1.0,2.0\n",
			},
			expected: 2,
			values:   []string{" S1 ", "\tCentral\n"},
		},
		{
			name: "whitespace on a later line of a quoted value",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,stop_lat,stop_lon\nS1,\"Central\nStation \",1.0,2.0\nS2,\" Airport\",3.0,4.0\n",
			},
			expected: 2,
			values:   []string{"Central\nStation ", " Airport"},
		},
		{
			name: "columns the spec does not define belong to unknown_column",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone,agency_notes\n1,Metro,http://metro.example,America/Los_Angeles,\" internal \"\n",
			},
			expected: 0,
		},
		{
			name: "files the spec does not describe are not read",
			files: map[string]string{
				"agency.txt":    "agency_id,agency_name,agency_url,agency_timezone\n1,Metro,http://metro.example,America/Los_Angeles\n",
				"operators.txt": "operator_id,operator_name\n1,\" Metro \"\n",
			},
			expected: 0,
		},
		{
			name: "every described file is read, not a chosen few",
			files: map[string]string{
				"translations.txt": "table_name,field_name,language,translation,record_id\nstops,stop_name,en,\"Central \",S1\n",
			},
			expected: 1,
			values:   []string{"Central "},
		},
		{
			name: "short rows do not run past their fields",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\n1,Metro\n",
			},
			expected: 0,
		},
		{
			name: "headers only",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\n",
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := testutil.CreateTestFeedLoader(t, tt.files)
			container := notice.NewNoticeContainer()

			NewLeadingTrailingWhitespaceValidator().Validate(loader, container, validator.Config{})

			notices := container.GetNotices()
			if len(notices) != tt.expected {
				t.Fatalf("expected %d notices, got %d: %v", tt.expected, len(notices), notices)
			}

			reported := make(map[string]bool, len(notices))
			for _, n := range notices {
				if n.Code() != "leading_or_trailing_whitespaces" {
					t.Errorf("expected code leading_or_trailing_whitespaces, got %s", n.Code())
				}
				value, ok := n.Context()["fieldValue"].(string)
				if !ok {
					t.Fatalf("notice carries no fieldValue: %v", n.Context())
				}
				reported[value] = true
			}
			for _, want := range tt.values {
				if !reported[want] {
					t.Errorf("expected a notice for %q, got %v", want, reported)
				}
			}
		})
	}
}

func TestTrimSpace(t *testing.T) {
	tests := []struct {
		value string
		want  string
	}{
		{"Metro", "Metro"},
		{" Metro ", "Metro"},
		{"\tMetro\r\n", "Metro"},
		{"   ", ""},
		{"Ц Metro Ц", "Ц Metro Ц"},
		{"\u00a0Metro\u00a0", "\u00a0Metro\u00a0"}, // Java's trim leaves a non-breaking space alone
	}

	for _, tt := range tests {
		if got := trimSpace(tt.value); got != tt.want {
			t.Errorf("trimSpace(%q) = %q, want %q", tt.value, got, tt.want)
		}
	}
}
