package entity

import (
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/testutil"
	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

func TestNameComparisonValidator_Validate(t *testing.T) {
	tests := []struct {
		name                string
		files               map[string]string
		expectedNoticeCodes []string
		description         string
	}{
		{
			name: "each entity has its own url",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\nA1,Metro,https://metro.example,America/Los_Angeles",
				"routes.txt": "route_id,route_url,route_short_name,route_type\nR1,https://metro.example/routes/1,1,3",
				"stops.txt":  "stop_id,stop_name,stop_url\nS1,Main St,https://metro.example/stops/1",
			},
			expectedNoticeCodes: []string{},
			description:         "Distinct URLs are what the fields are for",
		},
		{
			name: "route url copied from agency",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\nA1,Metro,https://metro.example,America/Los_Angeles",
				"routes.txt": "route_id,route_url,route_short_name,route_type\nR1,https://metro.example,1,3",
			},
			expectedNoticeCodes: []string{"same_route_and_agency_url"},
			description:         "route_url pointing at the agency page carries no route information",
		},
		{
			name: "stop url copied from agency",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\nA1,Metro,https://metro.example,America/Los_Angeles",
				"stops.txt":  "stop_id,stop_name,stop_url\nS1,Main St,https://metro.example",
			},
			expectedNoticeCodes: []string{"same_stop_and_agency_url"},
			description:         "stop_url pointing at the agency page carries no stop information",
		},
		{
			name: "stop url copied from route",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\nA1,Metro,https://metro.example,America/Los_Angeles",
				"routes.txt": "route_id,route_url,route_short_name,route_type\nR1,https://metro.example/routes/1,1,3",
				"stops.txt":  "stop_id,stop_name,stop_url\nS1,Main St,https://metro.example/routes/1",
			},
			expectedNoticeCodes: []string{"same_stop_and_route_url"},
			description:         "A stop is served by several routes, so it cannot link to one of them",
		},
		{
			name: "one url shared by all three files",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\nA1,Metro,https://metro.example,America/Los_Angeles",
				"routes.txt": "route_id,route_url,route_short_name,route_type\nR1,https://metro.example,1,3",
				"stops.txt":  "stop_id,stop_name,stop_url\nS1,Main St,https://metro.example",
			},
			expectedNoticeCodes: []string{"same_route_and_agency_url", "same_stop_and_agency_url", "same_stop_and_route_url"},
			description:         "Every pairing is reported once",
		},
		{
			name: "urls differing only in case and trailing slash",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\nA1,Metro,https://Metro.example/,America/Los_Angeles",
				"routes.txt": "route_id,route_url,route_short_name,route_type\nR1,https://metro.example,1,3",
			},
			expectedNoticeCodes: []string{"same_route_and_agency_url"},
			description:         "Host case and a trailing slash do not change which page is reached",
		},
		{
			name: "two agencies share the url a route copied",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\nA1,Metro,https://metro.example,America/Los_Angeles\nA2,Metro Rail,https://metro.example,America/Los_Angeles",
				"routes.txt": "route_id,route_url,route_short_name,route_type\nR1,https://metro.example,1,3",
			},
			expectedNoticeCodes: []string{"same_route_and_agency_url"},
			description:         "One copied URL is one defect, however many agencies share it",
		},
		{
			name: "no urls given",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\nA1,Metro,,America/Los_Angeles",
				"routes.txt": "route_id,route_url,route_short_name,route_type\nR1,,1,3",
				"stops.txt":  "stop_id,stop_name,stop_url\nS1,Main St,",
			},
			expectedNoticeCodes: []string{},
			description:         "Empty URLs are not copies of one another",
		},
		{
			name: "url columns absent",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_timezone\nA1,Metro,America/Los_Angeles",
				"routes.txt": "route_id,route_short_name,route_type\nR1,1,3",
				"stops.txt":  "stop_id,stop_name\nS1,Main St",
			},
			expectedNoticeCodes: []string{},
			description:         "The URL fields are optional",
		},
		{
			name: "no agency.txt",
			files: map[string]string{
				"routes.txt": "route_id,route_url,route_short_name,route_type\nR1,https://metro.example/routes/1,1,3",
				"stops.txt":  "stop_id,stop_name,stop_url\nS1,Main St,https://metro.example/routes/1",
			},
			expectedNoticeCodes: []string{"same_stop_and_route_url"},
			description:         "Stop and route URLs are still comparable without agency.txt",
		},
		{
			name: "agency name in upper case only",
			files: map[string]string{
				"agency.txt": "agency_id,agency_name,agency_url,agency_timezone\nA1,METRO TRANSIT,https://metro.example,America/Los_Angeles",
			},
			expectedNoticeCodes: []string{"mixed_case_recommended_field"},
			description:         "agency_name is customer-facing text",
		},
		{
			name: "no files at all",
			files: map[string]string{
				"trips.txt": "trip_id,route_id,service_id\nT1,R1,S1",
			},
			expectedNoticeCodes: []string{},
			description:         "Nothing to compare",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := testutil.CreateTestFeedLoader(t, tt.files)
			container := notice.NewNoticeContainer()
			v := NewNameComparisonValidator()
			config := gtfsvalidator.Config{}

			v.Validate(loader, container, config)

			notices := container.GetNotices()

			if len(notices) != len(tt.expectedNoticeCodes) {
				t.Errorf("Expected %d notices, got %d for case: %s", len(tt.expectedNoticeCodes), len(notices), tt.description)
			}

			expectedCodeCounts := make(map[string]int)
			for _, code := range tt.expectedNoticeCodes {
				expectedCodeCounts[code]++
			}

			actualCodeCounts := make(map[string]int)
			for _, n := range notices {
				actualCodeCounts[n.Code()]++
			}

			for expectedCode, expectedCount := range expectedCodeCounts {
				if actualCodeCounts[expectedCode] != expectedCount {
					t.Errorf("Expected %d notices with code '%s', got %d", expectedCount, expectedCode, actualCodeCounts[expectedCode])
				}
			}

			for actualCode := range actualCodeCounts {
				if expectedCodeCounts[actualCode] == 0 {
					t.Errorf("Unexpected notice code: %s", actualCode)
				}
			}
		})
	}
}

func TestNameComparisonValidator_NeedsMixedCase(t *testing.T) {
	tests := []struct {
		name   string
		value  string
		needed bool
	}{
		{name: "mixed case", value: "Red Hook/Atlantic Basin", needed: false},
		{name: "upper case only", value: "GALLERIA MALL", needed: true},
		{name: "lower case only", value: "campo grande norte", needed: true},
		{name: "upper case among digits", value: "3427 GG 17", needed: true},
		{name: "script without case", value: "東京駅", needed: false},
		{name: "single cased letter", value: "M4", needed: false},
		{name: "digits only", value: "604", needed: false},
		{name: "empty", value: "", needed: false},
		{name: "accented mixed case", value: "Gare du Nord", needed: false},
		{name: "accented upper case only", value: "GARE DU NORD", needed: true},

		// Cyrillic is cased, so the rule has something to say about it. Feeds
		// written in it are the reason this check cannot lean on ASCII.
		{name: "Cyrillic upper case only", value: "ЦЕНТРАЛНА ГАРА", needed: true},
		{name: "Cyrillic mixed case", value: "Централна гара", needed: false},
		{name: "Cyrillic lower case only", value: "централна гара", needed: true},
		{name: "Cyrillic upper case behind digits", value: "6ТМ", needed: true},
		{name: "Cyrillic abbreviations with one mixed word", value: "Ул. Ген. Гурко", needed: false},

		// Scripts with no case cannot answer the question, whether they are
		// written as one word or several.
		{name: "two caseless words", value: "東京 駅行", needed: false},
		{name: "Arabic", value: "محطة القطار", needed: false},
		{name: "Hebrew", value: "תחנה מרכזית", needed: false},

		// A word of one letter says nothing about the producer's intent.
		{name: "single letter", value: "A", needed: false},
		{name: "single letters only", value: "A B C", needed: false},
		{name: "single letter beside a word", value: "A Street", needed: false},

		// Canonical judges a lone word only when it is lower case, so an
		// all-caps single word is left alone while "GALLERIA MALL" is not.
		{name: "single upper case word", value: "GALLERIA", needed: false},
		{name: "single lower case word", value: "galleria", needed: true},
		{name: "mixed case single word", value: "GalleriaMall", needed: false},

		// One mixed-case word is enough to acquit the whole name.
		{name: "one mixed word among upper case", value: "Main ST NW", needed: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := needsMixedCase(tt.value); got != tt.needed {
				t.Errorf("needsMixedCase(%q) = %v, want %v", tt.value, got, tt.needed)
			}
		})
	}
}

func TestNameComparisonValidator_New(t *testing.T) {
	v := NewNameComparisonValidator()
	if v == nil {
		t.Error("NewNameComparisonValidator() returned nil")
	}
}
