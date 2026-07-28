package relationship

import (
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/testutil"
	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

func TestTranslationValidator_Validate(t *testing.T) {
	baseFiles := map[string]string{
		"stops.txt":      "stop_id,stop_name,stop_lat,stop_lon\nS1,First,1.0,1.0\nS2,Second,2.0,2.0",
		"routes.txt":     "route_id,route_short_name,route_type\nR1,1,3",
		"trips.txt":      "route_id,service_id,trip_id\nR1,SV1,T1",
		"stop_times.txt": "trip_id,arrival_time,departure_time,stop_id,stop_sequence\nT1,08:00:00,08:00:00,S1,1\nT1,08:10:00,08:10:00,S2,2",
		"feed_info.txt":  "feed_publisher_name,feed_publisher_url,feed_lang\nPub,https://example.com,en",
	}

	tests := []struct {
		name         string
		translations string
		expected     map[string]int
	}{
		{
			name:         "valid record reference",
			translations: "table_name,field_name,language,translation,record_id\nstops,stop_name,fr,Premier,S1",
			expected:     map[string]int{},
		},
		{
			name:         "valid field value reference",
			translations: "table_name,field_name,language,translation,field_value\nstops,stop_name,fr,Premier,First",
			expected:     map[string]int{},
		},
		{
			name:         "valid stop_times reference",
			translations: "table_name,field_name,language,translation,record_id,record_sub_id\nstop_times,stop_headsign,fr,Vers,T1,2",
			expected:     map[string]int{},
		},
		{
			name:         "unknown table name",
			translations: "table_name,field_name,language,translation,record_id\nvehicles,name,fr,Bus,V1",
			expected:     map[string]int{"translation_unknown_table_name": 1},
		},
		{
			name:         "table not in feed",
			translations: "table_name,field_name,language,translation,record_id\nlevels,level_name,fr,Etage,L1",
			expected:     map[string]int{"translation_unknown_table_name": 1},
		},
		{
			name:         "record id and field value both given",
			translations: "table_name,field_name,language,translation,record_id,field_value\nstops,stop_name,fr,Premier,S1,First",
			expected:     map[string]int{"translation_unexpected_value": 1},
		},
		{
			name:         "record sub id without record id",
			translations: "table_name,field_name,language,translation,record_id,record_sub_id,field_value\nstop_times,stop_headsign,fr,Vers,,2,First",
			expected:     map[string]int{"translation_unexpected_value": 1},
		},
		{
			name:         "feed_info naming a record",
			translations: "table_name,field_name,language,translation,record_id,field_value\nfeed_info,feed_publisher_name,fr,Editeur,FI1,Pub",
			expected:     map[string]int{"translation_unexpected_value": 2},
		},
		{
			name:         "missing record",
			translations: "table_name,field_name,language,translation,record_id\nstops,stop_name,fr,Premier,S99",
			expected:     map[string]int{"translation_foreign_key_violation": 1},
		},
		{
			name:         "missing stop_times sub record",
			translations: "table_name,field_name,language,translation,record_id,record_sub_id\nstop_times,stop_headsign,fr,Vers,T1,9",
			expected:     map[string]int{"translation_foreign_key_violation": 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files := make(map[string]string, len(baseFiles)+1)
			for name, content := range baseFiles {
				files[name] = content
			}
			files["translations.txt"] = tt.translations

			loader := testutil.CreateTestFeedLoader(t, files)
			container := notice.NewNoticeContainer()

			v := NewTranslationValidator()
			v.Validate(loader, container, gtfsvalidator.Config{})

			counts := map[string]int{}
			for _, n := range container.GetNotices() {
				counts[n.Code()]++
			}
			for _, code := range []string{
				"translation_unknown_table_name",
				"translation_unexpected_value",
				"translation_foreign_key_violation",
			} {
				if counts[code] != tt.expected[code] {
					t.Errorf("%s: got %d, want %d", code, counts[code], tt.expected[code])
				}
			}
		})
	}
}

func TestTranslationValidator_NoTranslationsFile(t *testing.T) {
	loader := testutil.CreateTestFeedLoader(t, map[string]string{
		"stops.txt": "stop_id,stop_name,stop_lat,stop_lon\nS1,First,1.0,1.0",
	})
	container := notice.NewNoticeContainer()

	v := NewTranslationValidator()
	v.Validate(loader, container, gtfsvalidator.Config{})

	if len(container.GetNotices()) != 0 {
		t.Errorf("expected no notices without translations.txt, got %d", len(container.GetNotices()))
	}
}
