package business

import (
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/testutil"
	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

func TestInSeatTransferValidator_Validate(t *testing.T) {
	stops := "stop_id,stop_name,stop_lat,stop_lon,location_type,parent_station\n" +
		"ST1,Central,1.0,1.0,1,\n" +
		"A,Platform A,1.0,1.0,0,ST1\n" +
		"B,Platform B,2.0,2.0,0,\n" +
		"C,Platform C,3.0,3.0,0,"
	stopTimes := "trip_id,arrival_time,departure_time,stop_id,stop_sequence\n" +
		"T1,08:00:00,08:00:00,B,1\n" +
		"T1,08:10:00,08:10:00,A,2\n" +
		"T2,08:15:00,08:15:00,A,1\n" +
		"T2,08:25:00,08:25:00,C,2"

	tests := []struct {
		name      string
		routes    string
		trips     string
		transfers string
		stopTimes string // defaults to stopTimes above
		expected  map[string]int
	}{
		{
			name:      "in-seat transfer joining the trips end to end",
			routes:    "route_id,route_short_name,route_type\nR1,1,3\nR2,2,3",
			trips:     "route_id,service_id,trip_id\nR1,SV1,T1\nR2,SV1,T2",
			transfers: "from_stop_id,to_stop_id,transfer_type,from_trip_id,to_trip_id\nA,A,4,T1,T2",
			expected:  map[string]int{},
		},
		{
			name:      "transfer naming the station rather than the platform",
			routes:    "route_id,route_short_name,route_type\nR1,1,3\nR2,2,3",
			trips:     "route_id,service_id,trip_id\nR1,SV1,T1\nR2,SV1,T2",
			transfers: "from_stop_id,to_stop_id,transfer_type,from_trip_id,to_trip_id\nST1,ST1,4,T1,T2",
			expected:  map[string]int{},
		},
		{
			name:      "trips on different modes",
			routes:    "route_id,route_short_name,route_type\nR1,1,3\nR2,2,0",
			trips:     "route_id,service_id,trip_id\nR1,SV1,T1\nR2,SV1,T2",
			transfers: "from_stop_id,to_stop_id,transfer_type,from_trip_id,to_trip_id\nA,A,4,T1,T2",
			expected:  map[string]int{"inconsistent_route_type_for_in_seat_transfer": 1},
		},
		{
			name:      "leaving the arriving trip before its last stop",
			routes:    "route_id,route_short_name,route_type\nR1,1,3\nR2,2,3",
			trips:     "route_id,service_id,trip_id\nR1,SV1,T1\nR2,SV1,T2",
			transfers: "from_stop_id,to_stop_id,transfer_type,from_trip_id,to_trip_id\nB,A,4,T1,T2",
			expected:  map[string]int{"transfer_with_suspicious_mid_trip_in_seat": 1},
		},
		{
			name:      "joining the departing trip after its first stop",
			routes:    "route_id,route_short_name,route_type\nR1,1,3\nR2,2,3",
			trips:     "route_id,service_id,trip_id\nR1,SV1,T1\nR2,SV1,T2",
			transfers: "from_stop_id,to_stop_id,transfer_type,from_trip_id,to_trip_id\nA,C,4,T1,T2",
			expected:  map[string]int{"transfer_with_suspicious_mid_trip_in_seat": 1},
		},
		{
			name:      "transfer type 5 is checked too",
			routes:    "route_id,route_short_name,route_type\nR1,1,3\nR2,2,3",
			trips:     "route_id,service_id,trip_id\nR1,SV1,T1\nR2,SV1,T2",
			transfers: "from_stop_id,to_stop_id,transfer_type,from_trip_id,to_trip_id\nB,A,5,T1,T2",
			expected:  map[string]int{"transfer_with_suspicious_mid_trip_in_seat": 1},
		},
		{
			name:      "ordinary transfer is not an in-seat transfer",
			routes:    "route_id,route_short_name,route_type\nR1,1,3\nR2,2,0",
			trips:     "route_id,service_id,trip_id\nR1,SV1,T1\nR2,SV1,T2",
			transfers: "from_stop_id,to_stop_id,transfer_type,from_trip_id,to_trip_id\nB,C,2,T1,T2",
			expected:  map[string]int{},
		},
		{
			name:      "stop times out of file order still find the ends",
			routes:    "route_id,route_short_name,route_type\nR1,1,3\nR2,2,3",
			trips:     "route_id,service_id,trip_id\nR1,SV1,T1\nR2,SV1,T2",
			transfers: "from_stop_id,to_stop_id,transfer_type,from_trip_id,to_trip_id\nA,A,4,T1,T2",
			stopTimes: "trip_id,arrival_time,departure_time,stop_id,stop_sequence\n" +
				"T1,08:10:00,08:10:00,A,2\n" +
				"T1,08:00:00,08:00:00,B,1\n" +
				"T2,08:25:00,08:25:00,C,2\n" +
				"T2,08:15:00,08:15:00,A,1",
			expected: map[string]int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tripStopTimes := tt.stopTimes
			if tripStopTimes == "" {
				tripStopTimes = stopTimes
			}
			loader := testutil.CreateTestFeedLoader(t, map[string]string{
				"stops.txt":      stops,
				"routes.txt":     tt.routes,
				"trips.txt":      tt.trips,
				"stop_times.txt": tripStopTimes,
				"transfers.txt":  tt.transfers,
			})
			container := notice.NewNoticeContainer()

			v := NewInSeatTransferValidator()
			v.Validate(loader, container, gtfsvalidator.Config{})

			counts := map[string]int{}
			for _, n := range container.GetNotices() {
				counts[n.Code()]++
			}
			for _, code := range []string{
				"inconsistent_route_type_for_in_seat_transfer",
				"transfer_with_suspicious_mid_trip_in_seat",
			} {
				if counts[code] != tt.expected[code] {
					t.Errorf("%s: got %d, want %d", code, counts[code], tt.expected[code])
				}
			}
		})
	}
}

func TestInSeatTransferValidator_NoTransfersFile(t *testing.T) {
	loader := testutil.CreateTestFeedLoader(t, map[string]string{
		"routes.txt":     "route_id,route_short_name,route_type\nR1,1,3",
		"trips.txt":      "route_id,service_id,trip_id\nR1,SV1,T1",
		"stop_times.txt": "trip_id,arrival_time,departure_time,stop_id,stop_sequence\nT1,08:00:00,08:00:00,A,1",
	})
	container := notice.NewNoticeContainer()

	v := NewInSeatTransferValidator()
	v.Validate(loader, container, gtfsvalidator.Config{})

	if len(container.GetNotices()) != 0 {
		t.Errorf("expected no notices without transfers.txt, got %d", len(container.GetNotices()))
	}
}
