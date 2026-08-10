package parser_test

import (
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-validator/parser"
	"github.com/theoremus-urban-solutions/gtfs-validator/testutil"
)

func cachedFeed(t *testing.T) *parser.ParsedFeedCache {
	t.Helper()

	loader := testutil.CreateTestFeedLoader(t, map[string]string{
		"stops.txt":  "stop_id,stop_name,stop_lat,stop_lon\nS1,Stop One,0,0\nS2,Stop Two,0,0.01",
		"routes.txt": "route_id,route_type\nR1,3",
		"trips.txt":  "route_id,service_id,trip_id\nR1,SV1,T1\nR1,SV1,T2",
		"stop_times.txt": "trip_id,arrival_time,departure_time,stop_id,stop_sequence\n" +
			"T1,08:00:00,08:00:00,S1,1\n" +
			"T1,08:10:00,08:10:00,S2,2\n" +
			"T2,09:00:00,09:00:00,S1,1\n" +
			"T2,09:10:00,09:10:00,S2,2",
	})
	loader.EnableCaching()
	return loader.GetCache()
}

// TestParsedFeedCache_IndexAccessorsPopulateTheCache reaches for each index
// before the bulk read that used to be the only thing that filled the cache.
// An index built over a file the cache had recorded as read but not kept left
// every later reader looking at an empty feed, which reads as a feed missing
// everything in it rather than as a cache that lost it.
func TestParsedFeedCache_IndexAccessorsPopulateTheCache(t *testing.T) {
	tests := []struct {
		name  string
		check func(t *testing.T, c *parser.ParsedFeedCache)
	}{
		{
			name: "stop times by trip, then all stop times",
			check: func(t *testing.T, c *parser.ParsedFeedCache) {
				byTrip, err := c.GetStopTimesByTrip()
				if err != nil {
					t.Fatalf("GetStopTimesByTrip: %v", err)
				}
				if len(byTrip["T1"]) != 2 {
					t.Errorf("trip T1 has %d stop times, want 2", len(byTrip["T1"]))
				}
				if all, err := c.GetStopTimes(); err != nil || len(all) != 4 {
					t.Errorf("GetStopTimes = %d rows (err %v), want 4", len(all), err)
				}
			},
		},
		{
			name: "trips by route, then all trips",
			check: func(t *testing.T, c *parser.ParsedFeedCache) {
				byRoute, err := c.GetTripsByRoute()
				if err != nil {
					t.Fatalf("GetTripsByRoute: %v", err)
				}
				if len(byRoute["R1"]) != 2 {
					t.Errorf("route R1 has %d trips, want 2", len(byRoute["R1"]))
				}
				if all, err := c.GetTrips(); err != nil || len(all) != 2 {
					t.Errorf("GetTrips = %d rows (err %v), want 2", len(all), err)
				}
			},
		},
		{
			name: "trip by id, then all trips",
			check: func(t *testing.T, c *parser.ParsedFeedCache) {
				if _, ok := c.GetTripByID("T1"); !ok {
					t.Error("GetTripByID(T1) found nothing")
				}
				if all, err := c.GetTrips(); err != nil || len(all) != 2 {
					t.Errorf("GetTrips = %d rows (err %v), want 2", len(all), err)
				}
			},
		},
		{
			name: "stop by id, then all stops",
			check: func(t *testing.T, c *parser.ParsedFeedCache) {
				if _, ok := c.GetStopByID("S1"); !ok {
					t.Error("GetStopByID(S1) found nothing")
				}
				if all, err := c.GetStops(); err != nil || len(all) != 2 {
					t.Errorf("GetStops = %d rows (err %v), want 2", len(all), err)
				}
			},
		},
		{
			name: "route by id, then all routes",
			check: func(t *testing.T, c *parser.ParsedFeedCache) {
				if _, ok := c.GetRouteByID("R1"); !ok {
					t.Error("GetRouteByID(R1) found nothing")
				}
				if all, err := c.GetRoutes(); err != nil || len(all) != 1 {
					t.Errorf("GetRoutes = %d rows (err %v), want 1", len(all), err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.check(t, cachedFeed(t))
		})
	}
}
