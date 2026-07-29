package relationship

import (
	"fmt"
	"sort"
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/testutil"
	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

// unsorted_stop_times is reported once per trip and names the trip's whole span
// in the file, so what matters is which trips are named and with what rows, not
// how many rows inside a trip break the order.
func TestStopTimeFieldValidator_UnsortedStopTimes(t *testing.T) {
	const header = "trip_id,arrival_time,departure_time,stop_id,stop_sequence,timepoint\n"

	tests := []struct {
		name        string
		stopTimes   string
		want        []string // "tripId:startCsvRowNumber-endCsvRowNumber"
		description string
	}{
		{
			name: "sorted and contiguous",
			stopTimes: "T1,08:00:00,08:00:00,A,1,1\n" +
				"T1,08:05:00,08:05:00,B,2,1\n" +
				"T2,09:00:00,09:00:00,A,1,1\n" +
				"T2,09:05:00,09:05:00,B,2,1",
			want:        nil,
			description: "Nothing to report about a file already in order",
		},
		{
			name: "the last stop listed second",
			stopTimes: "T1,08:00:00,08:00:00,A,1,1\n" +
				"T1,08:20:00,08:20:00,D,4,1\n" +
				"T1,08:05:00,08:05:00,B,2,1\n" +
				"T1,08:10:00,08:10:00,C,3,1",
			want:        []string{"T1:2-5"},
			description: "One notice for the trip, spanning every row it owns",
		},
		{
			name: "two descents inside one trip",
			stopTimes: "T1,08:00:00,08:00:00,A,3,1\n" +
				"T1,08:05:00,08:05:00,B,1,1\n" +
				"T1,08:10:00,08:10:00,C,4,1\n" +
				"T1,08:15:00,08:15:00,D,2,1",
			want:        []string{"T1:2-5"},
			description: "The trip is reported once however many rows break the order",
		},
		{
			name: "stop_sequence repeated",
			stopTimes: "T1,08:00:00,08:00:00,A,1,1\n" +
				"T1,08:05:00,08:05:00,B,1,1",
			want:        []string{"T1:2-3"},
			description: "A sequence that does not advance is not sorted either",
		},
		{
			name: "a trip resuming after another",
			stopTimes: "T1,08:00:00,08:00:00,A,1,1\n" +
				"T1,08:05:00,08:05:00,B,2,1\n" +
				"T2,09:00:00,09:00:00,A,1,1\n" +
				"T1,08:10:00,08:10:00,C,3,1",
			want:        []string{"T1:2-5"},
			description: "Sequences ascend throughout, but a consumer streaming the file has already finished T1",
		},
		{
			name: "one bad trip among good ones",
			stopTimes: "T1,08:00:00,08:00:00,A,1,1\n" +
				"T1,08:05:00,08:05:00,B,2,1\n" +
				"T2,09:00:00,09:00:00,A,2,1\n" +
				"T2,09:05:00,09:05:00,B,1,1\n" +
				"T3,10:00:00,10:00:00,A,1,1\n" +
				"T3,10:05:00,10:05:00,B,2,1",
			want:        []string{"T2:4-5"},
			description: "Only the trip that is out of order is named",
		},
		{
			name: "every trip out of order",
			stopTimes: "T1,08:00:00,08:00:00,A,2,1\n" +
				"T1,08:05:00,08:05:00,B,1,1\n" +
				"T2,09:00:00,09:00:00,A,2,1\n" +
				"T2,09:05:00,09:05:00,B,1,1",
			want:        []string{"T1:2-3", "T2:4-5"},
			description: "One notice each, not one per offending row",
		},
		{
			name:        "a single row",
			stopTimes:   "T1,08:00:00,08:00:00,A,1,1",
			want:        nil,
			description: "One row cannot be out of order with itself",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := testutil.CreateTestFeedLoader(t, map[string]string{
				"trips.txt":      "route_id,service_id,trip_id\nR1,S1,T1\nR1,S1,T2\nR1,S1,T3",
				"stops.txt":      "stop_id,stop_name\nA,Stop A\nB,Stop B\nC,Stop C\nD,Stop D",
				"stop_times.txt": header + tt.stopTimes,
			})
			container := notice.NewNoticeContainer()

			NewStopTimeFieldValidator().Validate(loader, container, gtfsvalidator.Config{})

			var got []string
			for _, n := range container.GetNotices() {
				if n.Code() != "unsorted_stop_times" {
					continue
				}
				context := n.Context()
				got = append(got, fmt.Sprintf("%v:%v-%v",
					context["tripId"], context["startCsvRowNumber"], context["endCsvRowNumber"]))
			}

			sort.Strings(got)
			want := append([]string(nil), tt.want...)
			sort.Strings(want)

			if fmt.Sprint(got) != fmt.Sprint(want) {
				t.Errorf("got %v, want %v: %s", got, want, tt.description)
			}
		})
	}
}
