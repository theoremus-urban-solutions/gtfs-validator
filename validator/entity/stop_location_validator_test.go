package entity

import (
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/testutil"
	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

func TestStopLocationValidator_PlatformAndStopAccess(t *testing.T) {
	tests := []struct {
		name     string
		stops    string
		expected map[string]int
	}{
		{
			// platform_code is the stop declaring itself part of a larger stop,
			// which is exactly the parent it then fails to name.
			name: "stop with a platform_code and no parent",
			stops: "stop_id,stop_name,stop_lat,stop_lon,location_type,parent_station,platform_code\n" +
				"ST1,Central,1.0,1.0,1,,\n" +
				"P1,Platform 1,1.0,1.0,0,ST1,A\n" +
				"S1,Platform 2,2.0,2.0,0,,B",
			expected: map[string]int{"platform_without_parent_station": 1},
		},
		{
			name:     "platform inside a station",
			stops:    "stop_id,stop_name,stop_lat,stop_lon,location_type,parent_station,platform_code\nST1,Central,1.0,1.0,1,,\nS1,Platform 1,1.0,1.0,0,ST1,A",
			expected: map[string]int{"platform_without_parent_station": 0},
		},
		{
			name: "location type defaults to platform",
			stops: "stop_id,stop_name,stop_lat,stop_lon,location_type,parent_station,platform_code\n" +
				"ST1,Central,1.0,1.0,1,,\n" +
				"S1,Roadside,2.0,2.0,,,B",
			expected: map[string]int{"platform_without_parent_station": 1},
		},
		{
			// An ordinary roadside stop is not a platform of anything, so a
			// missing parent is not an omission — even where the feed models
			// stations elsewhere. Only platform_code separates the two cases.
			name: "parentless stops without a platform_code in a feed that has stations",
			stops: "stop_id,stop_name,stop_lat,stop_lon,location_type,parent_station,platform_code\n" +
				"ST1,Central,1.0,1.0,1,,\n" +
				"E1,North entrance,1.0,1.0,2,ST1,\n" +
				"P1,Platform 1,1.0,1.0,0,ST1,\n" +
				"S1,Roadside,2.0,2.0,0,,\n" +
				"S2,Kerbside,3.0,3.0,0,,",
			expected: map[string]int{"platform_without_parent_station": 0},
		},
		{
			// A flat feed has no hierarchy to be incomplete, so the advisory
			// stays silent rather than naming every stop in the feed.
			name: "flat feed with no stations at all",
			stops: "stop_id,stop_name,stop_lat,stop_lon,location_type,parent_station\n" +
				"S1,Central,1.0,1.0,0,\n" +
				"S2,Midtown,2.0,2.0,0,\n" +
				"S3,Airport,3.0,3.0,0,",
			expected: map[string]int{"platform_without_parent_station": 0},
		},
		{
			name:     "flat feed with no location_type column",
			stops:    "stop_id,stop_name,stop_lat,stop_lon\nS1,Central,1.0,1.0\nS2,Midtown,2.0,2.0",
			expected: map[string]int{"platform_without_parent_station": 0},
		},
		{
			// An empty platform_code carries no assertion either way.
			name:     "empty platform_code",
			stops:    "stop_id,stop_name,stop_lat,stop_lon,location_type,parent_station,platform_code\nST1,Central,1.0,1.0,1,,\nS1,Roadside,2.0,2.0,0,,",
			expected: map[string]int{"platform_without_parent_station": 0},
		},
		{
			// The signal only means anything on a stop; the other location
			// types have their own, stricter rule.
			name:     "platform_code on an entrance",
			stops:    "stop_id,stop_name,stop_lat,stop_lon,location_type,parent_station,platform_code\nE1,North entrance,1.0,1.0,2,,A",
			expected: map[string]int{"platform_without_parent_station": 0, "location_without_parent_station": 1},
		},
		{
			// An entrance is meaningless outside a station whether or not the
			// feed declares one, so its notice is not gated the same way.
			name:  "entrance without a parent in a feed with no stations",
			stops: "stop_id,stop_name,stop_lat,stop_lon,location_type,parent_station\nE1,North entrance,1.0,1.0,2,",
			expected: map[string]int{
				"location_without_parent_station": 1,
				"platform_without_parent_station": 0,
			},
		},
		{
			name:  "stop_access on a station",
			stops: "stop_id,stop_name,stop_lat,stop_lon,location_type,parent_station,stop_access\nST1,Central,1.0,1.0,1,,1",
			expected: map[string]int{
				"stop_access_specified_for_incorrect_location": 1,
			},
		},
		{
			name:  "stop_access on a platform without a station",
			stops: "stop_id,stop_name,stop_lat,stop_lon,location_type,parent_station,stop_access\nS1,Central,1.0,1.0,0,,1",
			expected: map[string]int{
				"stop_access_specified_for_stop_with_no_parent_station": 1,
			},
		},
		{
			name:  "stop_access on a platform inside a station",
			stops: "stop_id,stop_name,stop_lat,stop_lon,location_type,parent_station,stop_access\nST1,Central,1.0,1.0,1,,\nS1,Platform 1,1.0,1.0,0,ST1,1",
			expected: map[string]int{
				"stop_access_specified_for_incorrect_location":          0,
				"stop_access_specified_for_stop_with_no_parent_station": 0,
			},
		},
		{
			name:  "no stop_access column",
			stops: "stop_id,stop_name,stop_lat,stop_lon,location_type,parent_station\nST1,Central,1.0,1.0,1,\nS1,Platform 1,1.0,1.0,0,ST1",
			expected: map[string]int{
				"stop_access_specified_for_incorrect_location":          0,
				"stop_access_specified_for_stop_with_no_parent_station": 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := testutil.CreateTestFeedLoader(t, map[string]string{"stops.txt": tt.stops})
			container := notice.NewNoticeContainer()

			v := NewStopLocationValidator()
			v.Validate(loader, container, gtfsvalidator.Config{})

			counts := map[string]int{}
			for _, n := range container.GetNotices() {
				counts[n.Code()]++
			}
			for code, want := range tt.expected {
				if counts[code] != want {
					t.Errorf("%s: got %d, want %d", code, counts[code], want)
				}
			}
		})
	}
}

func TestStopLocationValidator_ZoneCoverage(t *testing.T) {
	stops := "stop_id,stop_name,stop_lat,stop_lon,location_type,parent_station,zone_id\n" +
		"S1,Central,1.0,1.0,0,,Z1\n" +
		"S2,Midtown,2.0,2.0,0,,\n" +
		"S3,Airport,3.0,3.0,0,,"
	trips := "route_id,service_id,trip_id\nR1,SV1,T1\nR2,SV1,T2"
	stopTimes := "trip_id,arrival_time,departure_time,stop_id,stop_sequence\n" +
		"T1,08:00:00,08:00:00,S1,1\n" +
		"T1,08:10:00,08:10:00,S2,2\n" +
		"T2,09:00:00,09:00:00,S3,1"

	tests := []struct {
		name      string
		fareRules string
		expected  int
	}{
		{
			name:      "no fare rules file",
			fareRules: "",
			expected:  0,
		},
		{
			name:      "fare rules without zone fields",
			fareRules: "fare_id,route_id\nF1,R1",
			expected:  0,
		},
		{
			name:      "zone rule on one route",
			fareRules: "fare_id,route_id,origin_id,destination_id\nF1,R1,Z1,Z1",
			expected:  1,
		},
		{
			name:      "zone rule without a route covers every route",
			fareRules: "fare_id,route_id,contains_id\nF1,,Z1",
			expected:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files := map[string]string{
				"stops.txt":      stops,
				"trips.txt":      trips,
				"stop_times.txt": stopTimes,
			}
			if tt.fareRules != "" {
				files["fare_rules.txt"] = tt.fareRules
			}

			loader := testutil.CreateTestFeedLoader(t, files)
			container := notice.NewNoticeContainer()

			v := NewStopLocationValidator()
			v.Validate(loader, container, gtfsvalidator.Config{})

			count := 0
			for _, n := range container.GetNotices() {
				if n.Code() == "stop_without_zone_id" {
					count++
				}
			}
			if count != tt.expected {
				t.Errorf("stop_without_zone_id: got %d, want %d", count, tt.expected)
			}
		})
	}
}
