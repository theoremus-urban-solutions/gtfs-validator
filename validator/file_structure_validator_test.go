package validator

import (
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/schema"
	"github.com/theoremus-urban-solutions/gtfs-validator/testutil"
)

func TestFileStructureValidator_UnknownColumn(t *testing.T) {
	tests := []struct {
		name     string
		files    map[string]string
		expected int
	}{
		{
			name: "calendar.txt spelled exactly as the spec does",
			files: map[string]string{
				"calendar.txt": "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\n" +
					"SV1,1,1,1,1,1,0,0,20250101,20251231",
			},
			expected: 0,
		},
		{
			name: "stops.txt using the fields added after the original spec",
			files: map[string]string{
				"stops.txt": "stop_id,stop_name,tts_stop_name,stop_lat,stop_lon,stop_access,platform_code,level_id\n" +
					"S1,Central,Central Station,1.0,1.0,1,A,L1",
			},
			expected: 0,
		},
		{
			name: "stop_times.txt using the flex fields",
			files: map[string]string{
				"stop_times.txt": "trip_id,stop_id,stop_sequence,location_group_id,location_id," +
					"start_pickup_drop_off_window,end_pickup_drop_off_window," +
					"pickup_booking_rule_id,drop_off_booking_rule_id\n" +
					"T1,S1,1,,,08:00:00,09:00:00,BR1,BR1",
			},
			expected: 0,
		},
		{
			name: "trips.txt and routes.txt using the newest fields",
			files: map[string]string{
				"trips.txt":  "route_id,service_id,trip_id,cars_allowed\nR1,SV1,T1,1",
				"routes.txt": "route_id,route_type,network_id,cemv_support\nR1,3,N1,1",
			},
			expected: 0,
		},
		{
			name: "a column the spec does not define",
			files: map[string]string{
				"agency.txt": "agency_name,agency_url,agency_timezone,agency_colour\n" +
					"Metro,https://example.com,Europe/Sofia,red",
			},
			expected: 1,
		},
		{
			name: "a known column padded with whitespace",
			files: map[string]string{
				"agency.txt": "agency_name, agency_url ,agency_timezone\n" +
					"Metro,https://example.com,Europe/Sofia",
			},
			expected: 0,
		},
		{
			// The file is reported once as unknown_file; naming its columns
			// too would just restate that per column.
			name: "a file outside the spec entirely",
			files: map[string]string{
				"vehicle_positions.txt": "vehicle_id,latitude,longitude\nV1,1.0,1.0",
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := testutil.CreateTestFeedLoader(t, tt.files)
			container := notice.NewNoticeContainer()

			NewFileStructureValidator().Validate(loader, container, Config{})

			count := 0
			for _, n := range container.GetNotices() {
				if n.Code() == "unknown_column" {
					count++
				}
			}
			if count != tt.expected {
				t.Errorf("unknown_column: got %d, want %d", count, tt.expected)
			}
		})
	}
}

// TestFileStructureValidator_KnownColumnsCoverSpecFiles guards the failure mode
// this table has: a file the validator reads but the table does not describe is
// checked against nothing, so every one of its columns is reported.
func TestFileStructureValidator_KnownColumnsCoverSpecFiles(t *testing.T) {
	required := []string{
		"agency.txt", "stops.txt", "routes.txt", "trips.txt", "stop_times.txt",
		"calendar.txt", "calendar_dates.txt", "fare_attributes.txt", "fare_rules.txt",
		"shapes.txt", "frequencies.txt", "transfers.txt", "pathways.txt", "levels.txt",
		"feed_info.txt", "translations.txt", "attributions.txt",
	}

	for _, filename := range required {
		if _, ok := schema.KnownColumns(filename); !ok {
			t.Errorf("the generated spec table has no entry for %s", filename)
		}
	}
}
