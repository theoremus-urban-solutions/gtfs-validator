package business

import (
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	gtfsvalidator "github.com/theoremus-urban-solutions/gtfs-validator/validator"
	"github.com/theoremus-urban-solutions/gtfs-validator/testutil"
)

func TestNetworkTopologyValidator_Validate(t *testing.T) {
	files := map[string]string{
		// Two routes, sparse connectivity to trigger isolated/low connectivity
		"trips.txt":      "route_id,service_id,trip_id\nR1,S1,T1\nR2,S1,T2",
		"stop_times.txt": "trip_id,arrival_time,departure_time,stop_id,stop_sequence\nT1,08:00:00,08:00:00,A,1\nT1,08:10:00,08:10:00,B,2\nT2,09:00:00,09:00:00,C,1",
	}

	loader := testutil.CreateTestFeedLoader(t, files)
	container := notice.NewNoticeContainer()

	v := NewNetworkTopologyValidator()
	v.Validate(loader, container, gtfsvalidator.Config{})

	codes := map[string]int{}
	for _, n := range container.GetNotices() {
		codes[n.Code()]++
	}

	if codes["isolated_stop"] == 0 && codes["low_network_connectivity"] == 0 && codes["network_topology_summary"] == 0 {
		t.Errorf("expected at least one topology-related notice (isolated_stop/low_network_connectivity/summary)")
	}
}
