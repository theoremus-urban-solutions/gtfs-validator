package business

import (
	"math"
	"sort"
)

// Matching stops against shapes has to answer "which part of this polyline runs
// near this stop" for every stop of every distinct trip pattern in the feed.
// Testing every stop against every shape point is quadratic in feed size, which
// a feed with millions of stop times will not survive, so shape segments are
// bucketed into a fixed latitude/longitude grid and a query touches only the
// cells its search radius covers.

const (
	metresPerDegreeLat = 110574.0
	metresPerDegreeLon = 111320.0 // at the equator, scaled by cos(latitude)

	// gridCellDegrees sizes a bucket at roughly 1.1 km north to south: well
	// above the 100 m match radius, so a query reads a handful of cells, and
	// small enough that a city-sized shape spreads across many of them.
	gridCellDegrees = 0.01

	// maxCellsPerSegment caps how many buckets one segment may be written to.
	// A shape point every 5 m and a shape point every 50 km are both legal
	// GTFS; indexing the latter cell by cell would cost more than testing it
	// outright, so oversized segments go in a list every query checks.
	maxCellsPerSegment = 32

	// maxQueryCells bounds the cells one query may walk. Only reached near the
	// poles, where a degree of longitude shrinks to metres and the search
	// window spans absurdly many columns; a full scan is cheaper there.
	maxQueryCells = 4096
)

// shapePoint is one vertex of a shape polyline. Dist is the point's
// shape_dist_traveled, in whatever unit the feed chose, and is absent for
// shapes that do not declare distances.
type shapePoint struct {
	Lat  float64
	Lon  float64
	Dist *float64
}

// shapeMatch is a point on a shape close to a queried location: how far from
// the start of the shape it lies, and how far the queried location was from it.
type shapeMatch struct {
	Along  float64 // metres travelled along the shape to reach the match
	Metres float64 // distance from the queried location to the match
}

type gridCell struct {
	lat int32
	lon int32
}

// shapeIndex is a shape polyline prepared for nearest-point queries.
type shapeIndex struct {
	points []shapePoint

	// cumulative[i] is the ground distance in metres from the first point to
	// points[i], which turns a match on a segment into a position along the
	// shape.
	cumulative []float64

	cells     map[gridCell][]int
	oversized []int

	// userDist holds the shape's own shape_dist_traveled values and userIdx
	// the points they belong to, present only when every value is non-
	// decreasing. A shape whose distances go backwards is already reported by
	// decreasing_shape_distance and cannot be interpolated meaningfully.
	userDist []float64
	userIdx  []int
}

// newShapeIndex prepares points for querying. Points must already be sorted by
// shape_pt_sequence.
func newShapeIndex(points []shapePoint) *shapeIndex {
	idx := &shapeIndex{
		points:     points,
		cumulative: make([]float64, len(points)),
		cells:      make(map[gridCell][]int),
	}

	for i := 1; i < len(points); i++ {
		idx.cumulative[i] = idx.cumulative[i-1] + haversineMetres(
			points[i-1].Lat, points[i-1].Lon,
			points[i].Lat, points[i].Lon,
		)
	}

	for i := 0; i+1 < len(points); i++ {
		idx.indexSegment(i)
	}

	idx.buildUserDistances()
	return idx
}

// indexSegment files a segment under every cell its bounding box touches, or
// under oversized when that would be too many cells to be worth it.
func (s *shapeIndex) indexSegment(i int) {
	a, b := s.points[i], s.points[i+1]
	low := cellOf(math.Min(a.Lat, b.Lat), math.Min(a.Lon, b.Lon))
	high := cellOf(math.Max(a.Lat, b.Lat), math.Max(a.Lon, b.Lon))

	span := (int64(high.lat-low.lat) + 1) * (int64(high.lon-low.lon) + 1)
	if span > maxCellsPerSegment {
		s.oversized = append(s.oversized, i)
		return
	}

	for lat := low.lat; lat <= high.lat; lat++ {
		for lon := low.lon; lon <= high.lon; lon++ {
			cell := gridCell{lat: lat, lon: lon}
			s.cells[cell] = append(s.cells[cell], i)
		}
	}
}

// buildUserDistances records the shape's declared distances, but only if they
// never decrease — interpolating a position out of values that go backwards
// would invent a location the feed does not describe.
func (s *shapeIndex) buildUserDistances() {
	dist := make([]float64, 0, len(s.points))
	idx := make([]int, 0, len(s.points))

	for i, point := range s.points {
		if point.Dist == nil {
			continue
		}
		if len(dist) > 0 && *point.Dist < dist[len(dist)-1] {
			return
		}
		dist = append(dist, *point.Dist)
		idx = append(idx, i)
	}

	if len(dist) < 2 {
		return
	}
	s.userDist = dist
	s.userIdx = idx
}

// maxUserDistance returns the largest shape_dist_traveled the shape declares.
func (s *shapeIndex) maxUserDistance() (float64, bool) {
	if len(s.userDist) == 0 {
		return 0, false
	}
	return s.userDist[len(s.userDist)-1], true
}

// locateByUserDistance returns the coordinates the shape's own
// shape_dist_traveled values place at userDistance. Distances beyond either end
// clamp to that end, which is where trip_distance_exceeds_shape_distance takes
// over.
func (s *shapeIndex) locateByUserDistance(userDistance float64) (lat float64, lon float64, ok bool) {
	if len(s.userDist) < 2 {
		return 0, 0, false
	}

	at := sort.SearchFloat64s(s.userDist, userDistance)
	switch {
	case at == 0:
		point := s.points[s.userIdx[0]]
		return point.Lat, point.Lon, true
	case at >= len(s.userDist):
		point := s.points[s.userIdx[len(s.userIdx)-1]]
		return point.Lat, point.Lon, true
	}

	before, after := s.points[s.userIdx[at-1]], s.points[s.userIdx[at]]
	span := s.userDist[at] - s.userDist[at-1]
	if span <= 0 {
		return before.Lat, before.Lon, true
	}

	fraction := (userDistance - s.userDist[at-1]) / span
	return before.Lat + (after.Lat-before.Lat)*fraction,
		before.Lon + (after.Lon-before.Lon)*fraction,
		true
}

// nearby appends every segment match within radiusMetres of the location to
// out. A segment filed under several cells can appear twice; both entries
// describe the same place on the shape, so the caller's clustering folds them
// together.
func (s *shapeIndex) nearby(lat float64, lon float64, radiusMetres float64, out []shapeMatch) []shapeMatch {
	if len(s.points) < 2 {
		return out
	}

	low, high, ok := s.queryWindow(lat, lon, radiusMetres)
	if !ok {
		// The window is too wide to walk cell by cell, so test everything.
		for i := 0; i+1 < len(s.points); i++ {
			out = s.appendIfWithin(out, lat, lon, i, radiusMetres)
		}
		return out
	}

	for cellLat := low.lat; cellLat <= high.lat; cellLat++ {
		for cellLon := low.lon; cellLon <= high.lon; cellLon++ {
			for _, i := range s.cells[gridCell{lat: cellLat, lon: cellLon}] {
				out = s.appendIfWithin(out, lat, lon, i, radiusMetres)
			}
		}
	}
	for _, i := range s.oversized {
		out = s.appendIfWithin(out, lat, lon, i, radiusMetres)
	}

	return out
}

// nearest returns the closest point on the whole shape. It scans every segment,
// so it is reserved for locations the grid found nothing near — that is, the
// ones about to be reported anyway.
func (s *shapeIndex) nearest(lat float64, lon float64) shapeMatch {
	best := shapeMatch{Metres: math.Inf(1)}
	for i := 0; i+1 < len(s.points); i++ {
		metres, fraction := s.segmentDistance(lat, lon, i)
		if metres < best.Metres {
			best = shapeMatch{Along: s.alongAt(i, fraction), Metres: metres}
		}
	}
	return best
}

func (s *shapeIndex) appendIfWithin(out []shapeMatch, lat float64, lon float64, i int, radiusMetres float64) []shapeMatch {
	metres, fraction := s.segmentDistance(lat, lon, i)
	if metres > radiusMetres {
		return out
	}
	return append(out, shapeMatch{Along: s.alongAt(i, fraction), Metres: metres})
}

// alongAt converts a position part-way along segment i into metres travelled
// from the start of the shape.
func (s *shapeIndex) alongAt(i int, fraction float64) float64 {
	return s.cumulative[i] + fraction*(s.cumulative[i+1]-s.cumulative[i])
}

// queryWindow returns the range of cells covering radiusMetres around the
// location, or ok=false when that range is too large to be worth walking.
func (s *shapeIndex) queryWindow(lat float64, lon float64, radiusMetres float64) (low gridCell, high gridCell, ok bool) {
	latSpan := radiusMetres / metresPerDegreeLat

	lonSpan := 180.0
	if cosLat := math.Cos(lat * math.Pi / 180); cosLat > 1e-6 {
		lonSpan = math.Min(180.0, radiusMetres/(metresPerDegreeLon*cosLat))
	}

	low = cellOf(lat-latSpan, lon-lonSpan)
	high = cellOf(lat+latSpan, lon+lonSpan)
	span := (int64(high.lat-low.lat) + 1) * (int64(high.lon-low.lon) + 1)
	if span > maxQueryCells {
		return low, high, false
	}
	return low, high, true
}

// segmentDistance returns the distance in metres from the location to segment i
// and how far along that segment the closest point lies, as a fraction.
// Coordinates are projected onto a plane centred on the location, which over
// the length of a shape segment costs far less accuracy than the coordinates
// themselves carry.
func (s *shapeIndex) segmentDistance(lat float64, lon float64, i int) (metres float64, fraction float64) {
	a, b := s.points[i], s.points[i+1]

	scaleLon := metresPerDegreeLon * math.Cos(lat*math.Pi/180)
	ax, ay := (a.Lon-lon)*scaleLon, (a.Lat-lat)*metresPerDegreeLat
	bx, by := (b.Lon-lon)*scaleLon, (b.Lat-lat)*metresPerDegreeLat

	dx, dy := bx-ax, by-ay
	if lengthSq := dx*dx + dy*dy; lengthSq > 0 {
		fraction = math.Max(0, math.Min(1, -(ax*dx+ay*dy)/lengthSq))
	}

	return math.Hypot(ax+fraction*dx, ay+fraction*dy), fraction
}

func cellOf(lat float64, lon float64) gridCell {
	return gridCell{
		lat: int32(math.Floor(lat / gridCellDegrees)),
		lon: int32(math.Floor(lon / gridCellDegrees)),
	}
}

// haversineMetres is the great-circle distance between two coordinates.
func haversineMetres(lat1 float64, lon1 float64, lat2 float64, lon2 float64) float64 {
	const earthRadius = 6371000.0

	lat1Rad := lat1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	deltaLat := (lat2 - lat1) * math.Pi / 180
	deltaLon := (lon2 - lon1) * math.Pi / 180

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*math.Sin(deltaLon/2)*math.Sin(deltaLon/2)

	return earthRadius * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}
