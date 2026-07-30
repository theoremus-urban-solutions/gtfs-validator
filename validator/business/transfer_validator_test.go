package business

import (
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/testutil"
	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

func TestTransferValidator_Validate(t *testing.T) {
	files := map[string]string{
		"stops.txt":     "stop_id,stop_name\nA,Stop A\nB,Stop B",
		"transfers.txt": "from_stop_id,to_stop_id,transfer_type,min_transfer_time\nA,B,4,\nA,A,0,\nA,B,2,",
	}

	loader := testutil.CreateTestFeedLoader(t, files)
	container := notice.NewNoticeContainer()

	v := NewTransferValidator()
	v.Validate(loader, container, gtfsvalidator.Config{})

	codes := map[string]int{}
	for _, n := range container.GetNotices() {
		codes[n.Code()]++
	}

	if codes["transfer_to_same_stop"] == 0 {
		t.Errorf("expected transfer_to_same_stop notice")
	}
	if codes["missing_min_transfer_time"] == 0 {
		t.Errorf("expected missing_min_transfer_time notice for type=2 without time")
	}
}

// Fixtures sit on the equator, where 0.01 degrees of longitude is 1113 m.
func TestTransferValidator_TransferDistance(t *testing.T) {
	tests := []struct {
		name         string
		stops        string
		expectedCode string
		description  string
	}{
		{
			name:         "across the street",
			stops:        "stop_id,stop_name,stop_lat,stop_lon\nA,Stop A,0,0\nB,Stop B,0,0.001",
			expectedCode: "",
			description:  "111 m is an ordinary same-street transfer and must stay silent",
		},
		{
			name:         "across town",
			stops:        "stop_id,stop_name,stop_lat,stop_lon\nA,Stop A,0,0\nB,Stop B,0,0.03",
			expectedCode: "transfer_distance_above_2_km",
			description:  "3.3 km is long but a timed connection could mean it",
		},
		{
			name:         "across the region",
			stops:        "stop_id,stop_name,stop_lat,stop_lon\nA,Stop A,0,0\nB,Stop B,0,0.1",
			expectedCode: "transfer_distance_too_large",
			description:  "11 km is beyond anything a passenger transfers on foot",
		},
		{
			name:         "stop with no coordinates",
			stops:        "stop_id,stop_name,stop_lat,stop_lon\nA,Stop A,0,0\nB,Stop B,,",
			expectedCode: "",
			description:  "there is no distance to measure without both ends",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := testutil.CreateTestFeedLoader(t, map[string]string{
				"stops.txt":     tt.stops,
				"transfers.txt": "from_stop_id,to_stop_id,transfer_type\nA,B,0",
			})
			container := notice.NewNoticeContainer()

			NewTransferValidator().Validate(loader, container, gtfsvalidator.Config{})

			distanceCodes := map[string]bool{
				"transfer_distance_above_2_km": true,
				"transfer_distance_too_large":  true,
			}

			var fired []string
			for _, n := range container.GetNotices() {
				if distanceCodes[n.Code()] {
					fired = append(fired, n.Code())
				}
			}

			switch {
			case tt.expectedCode == "" && len(fired) > 0:
				t.Errorf("expected no distance notice, got %v (%s)", fired, tt.description)
			case tt.expectedCode != "" && len(fired) != 1:
				t.Errorf("expected only %q, got %v (%s)", tt.expectedCode, fired, tt.description)
			case tt.expectedCode != "" && fired[0] != tt.expectedCode:
				t.Errorf("expected %q, got %v (%s)", tt.expectedCode, fired, tt.description)
			}
		})
	}
}
