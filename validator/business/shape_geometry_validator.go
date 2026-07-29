package business

import (
	"io"
	"log"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/parser"
	"github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

const (
	// maxStopToShapeMetres is the GTFS Best Practices alignment tolerance: a
	// shape should pass within 100 m of every stop its trip serves.
	maxStopToShapeMetres = 100.0

	// matchSeparationMetres is how far apart along the shape two candidate
	// matches must be before they count as separate passes of the stop rather
	// than neighbouring segments of the same pass.
	matchSeparationMetres = 250.0

	// maxMatchesPerStop is how many separate passes a stop may have before the
	// match is called ambiguous. A loop route passes a stop twice and a
	// figure-of-eight three times; beyond that the shape is doubling back on
	// itself or the stop is in the wrong place.
	maxMatchesPerStop = 5

	// maxOrderingCandidates bounds how many passes per stop the order check will
	// weigh against each other. The check costs the square of this per stop, so
	// it has to be a constant: a pathological shape that grazes one stop a
	// hundred times must not turn a pattern walk into a hundred-fold one. Set
	// above maxMatchesPerStop, since a stop with more passes than that is
	// already reported as ambiguous.
	maxOrderingCandidates = 8

	// tripShapeOvershootThreshold is the shape_dist_traveled overshoot below
	// which a trip running past the end of its shape is treated as rounding.
	// 11.1 units is 1e-4 degrees of latitude, the smallest difference four
	// decimal places of coordinate can express — the canonical rule's
	// threshold, applied in whatever unit the feed declared its distances in.
	tripShapeOvershootThreshold = 11.1
)

// ShapeGeometryValidator checks trips against the path their shape draws:
// whether the alignment passes the stops it serves, in the order it serves
// them, and whether the distances the feed declares agree with that geometry.
//
// Gated behind EnableShapeGeometry because it is the only check here that has
// to hold the whole shape and stop-time set in memory at once.
type ShapeGeometryValidator struct{}

// NewShapeGeometryValidator creates a new shape geometry validator
func NewShapeGeometryValidator() *ShapeGeometryValidator {
	return &ShapeGeometryValidator{}
}

// tripStop is one call of a trip, carrying only what the shape checks need.
// Location is nil for a stop_id stops.txt does not place, which the foreign
// key and coordinate checks already report.
type tripStop struct {
	StopID       string
	StopSequence int
	Dist         *float64
	Location     *StopLocation
	RowNumber    int
}

// stopPasses is a stop the shape does come near, together with the places along
// the shape it could be served from.
type stopPasses struct {
	Stop   *tripStop
	Passes []shapeMatch
}

// assignmentCost ranks one way of placing the stops along the shape. An
// assignment that keeps the trip moving forwards beats one that does not
// however far off the alignment it sits, because that is what the rule is
// about; total distance from the shape only separates assignments that go
// backwards equally often, where it picks the reading closest to the geometry.
type assignmentCost struct {
	breaks    int
	deviation float64
}

func (c assignmentCost) better(other assignmentCost) bool {
	if c.breaks != other.breaks {
		return c.breaks < other.breaks
	}
	return c.deviation < other.deviation
}

// sequencedPoint keeps shape_pt_sequence alongside a point long enough to sort
// by it. The index only ever needs the ordered polyline.
type sequencedPoint struct {
	Point    shapePoint
	Sequence int
}

// stopPattern is a shape and the stops a trip calls at along it, plus every
// trip that repeats that arrangement. A timetable runs the same pattern dozens
// of times a day and the geometry does not change between runs, so it is
// walked once and reported against each of them.
type stopPattern struct {
	ShapeID string
	Stops   []tripStop
	TripIDs []string
}

// Validate checks every distinct trip pattern against its shape
func (v *ShapeGeometryValidator) Validate(loader *parser.FeedLoader, container *notice.NoticeContainer, config validator.Config) {
	shapes := v.loadShapes(loader)
	if len(shapes) == 0 {
		return
	}

	tripShapes := v.loadTripShapes(loader, shapes)
	if len(tripShapes) == 0 {
		return
	}

	stopLocations := v.loadStopLocations(loader)
	if len(stopLocations) == 0 {
		return
	}

	patterns := v.loadPatterns(loader, tripShapes, stopLocations)

	// Sorted so a feed always produces its notices in the same order; map
	// iteration would shuffle the report between runs.
	keys := make([]string, 0, len(patterns))
	for key := range patterns {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	indexes := make(map[string]*shapeIndex, len(shapes))
	for _, key := range keys {
		pattern := patterns[key]
		index, built := indexes[pattern.ShapeID]
		if !built {
			index = newShapeIndex(shapes[pattern.ShapeID])
			indexes[pattern.ShapeID] = index
		}
		v.validatePattern(container, pattern, index)
	}
}

// validatePattern runs every geometric check over one distinct pattern.
func (v *ShapeGeometryValidator) validatePattern(container *notice.NoticeContainer, pattern *stopPattern, index *shapeIndex) {
	v.validateStopsAgainstShape(container, pattern, index)
	v.validateUserDistances(container, pattern, index)
	v.validateTripDistance(container, pattern, index)
}

// validateStopsAgainstShape matches each stop of the pattern onto the shape and
// reports the three ways that can go wrong: no match, too many matches, or a
// match that goes backwards along a shape the trip travels forwards.
func (v *ShapeGeometryValidator) validateStopsAgainstShape(container *notice.NoticeContainer, pattern *stopPattern, index *shapeIndex) {
	if len(index.points) < 2 {
		return
	}

	matches := make([]shapeMatch, 0, 16)
	matched := make([]stopPasses, 0, len(pattern.Stops))

	for i := range pattern.Stops {
		stop := &pattern.Stops[i]
		location := stop.Location
		if location == nil {
			continue
		}

		matches = index.nearby(location.Latitude, location.Longitude, maxStopToShapeMetres, matches[:0])
		passes := clusterMatches(matches)

		if len(passes) == 0 {
			nearest := index.nearest(location.Latitude, location.Longitude)
			report(container, pattern.TripIDs, func(tripID string) notice.Notice {
				return notice.NewStopTooFarFromShapeNotice(
					tripID, stop.StopID, stop.StopSequence,
					pattern.ShapeID, nearest.Metres, stop.RowNumber,
				)
			})
			continue
		}

		if len(passes) > maxMatchesPerStop {
			report(container, pattern.TripIDs, func(tripID string) notice.Notice {
				return notice.NewStopHasTooManyMatchesForShapeNotice(
					tripID, stop.StopID, stop.StopSequence,
					pattern.ShapeID, len(passes), stop.RowNumber,
				)
			})
		}

		matched = append(matched, stopPasses{Stop: stop, Passes: boundCandidates(passes)})
	}

	v.reportOutOfOrder(container, pattern, matched)
}

// reportOutOfOrder decides where along the shape the pattern serves each of its
// stops, and reports the consecutive pairs left back to front.
//
// Deciding one stop at a time — the nearest pass at or ahead of the stop before
// it — is what a shape that doubles back defeats. Where an out-and-back spur
// brings the alignment past itself, a stop on the outbound leg lies within
// tolerance of the inbound leg too, and one locally nearest choice can strand
// every following stop apparently behind it. Whether the stops are in order is
// a property of the sequence as a whole, so it is settled over the sequence as
// a whole: of all the ways to place these stops along this shape, take the one
// that goes backwards fewest times, and report only what that cannot avoid.
func (v *ShapeGeometryValidator) reportOutOfOrder(container *notice.NoticeContainer, pattern *stopPattern, matched []stopPasses) {
	if len(matched) < 2 {
		return
	}

	// The first stop has nothing behind it to travel forwards from. Anchoring it
	// at the earliest place it touches the shape can never push a later stop
	// backwards — no other choice starts further back — so it needs no weighing
	// against the rest and is settled first.
	anchor := chooseMatch(matched[0].Passes)
	matched[0].Passes = []shapeMatch{anchor}

	// costs[j] is the best an assignment ending with the current stop matched at
	// its pass j can do; parents[i][j] records which pass of the stop before it
	// that assignment came through, so the winner can be walked back afterwards.
	costs := []assignmentCost{{deviation: anchor.Metres}}
	parents := make([][]int, len(matched))

	for i := 1; i < len(matched); i++ {
		previous := matched[i-1].Passes
		current := matched[i].Passes

		next := make([]assignmentCost, len(current))
		parent := make([]int, len(current))

		for j, pass := range current {
			for k, before := range previous {
				candidate := costs[k]
				if pass.Along < before.Along {
					candidate.breaks++
				}
				candidate.deviation += pass.Metres

				if k == 0 || candidate.better(next[j]) {
					next[j], parent[j] = candidate, k
				}
			}
		}

		costs, parents[i] = next, parent
	}

	best := 0
	for j := range costs {
		if costs[j].better(costs[best]) {
			best = j
		}
	}
	if costs[best].breaks == 0 {
		return
	}

	picks := make([]int, len(matched))
	picks[len(matched)-1] = best
	for i := len(matched) - 1; i > 0; i-- {
		picks[i-1] = parents[i][picks[i]]
	}

	for i := 1; i < len(matched); i++ {
		if matched[i].Passes[picks[i]].Along >= matched[i-1].Passes[picks[i-1]].Along {
			continue
		}
		earlier, later := matched[i-1].Stop, matched[i].Stop
		report(container, pattern.TripIDs, func(tripID string) notice.Notice {
			return notice.NewStopsMatchShapeOutOfOrderNotice(
				tripID, pattern.ShapeID,
				earlier.StopID, earlier.StopSequence,
				later.StopID, later.StopSequence,
			)
		})
	}
}

// boundCandidates trims a stop's passes to the ones the order check will weigh,
// keeping the closest and leaving them in order along the shape. A stop the
// alignment reaches from this many places is already past the point where the
// geometry says which one the trip means, and the passes dropped are the ones
// furthest off the alignment.
func boundCandidates(passes []shapeMatch) []shapeMatch {
	if len(passes) <= maxOrderingCandidates {
		return passes
	}

	closest := make([]shapeMatch, len(passes))
	copy(closest, passes)
	sort.SliceStable(closest, func(i, j int) bool { return closest[i].Metres < closest[j].Metres })

	kept := closest[:maxOrderingCandidates]
	sort.Slice(kept, func(i, j int) bool { return kept[i].Along < kept[j].Along })
	return kept
}

// validateUserDistances checks the stop times that declare a
// shape_dist_traveled against the place on the shape that distance points at.
// Unlike the geometric match this trusts the feed's own numbers, so it catches
// distances measured in the wrong unit or against the wrong shape.
func (v *ShapeGeometryValidator) validateUserDistances(container *notice.NoticeContainer, pattern *stopPattern, index *shapeIndex) {
	for i := range pattern.Stops {
		stop := &pattern.Stops[i]
		if stop.Dist == nil || stop.Location == nil {
			continue
		}

		lat, lon, ok := index.locateByUserDistance(*stop.Dist)
		if !ok {
			return // the shape declares no usable distances to locate against
		}

		metres := haversineMetres(lat, lon, stop.Location.Latitude, stop.Location.Longitude)
		if metres <= maxStopToShapeMetres {
			continue
		}

		report(container, pattern.TripIDs, func(tripID string) notice.Notice {
			return notice.NewStopTooFarFromShapeUsingUserDistanceNotice(
				tripID, stop.StopID, stop.StopSequence,
				pattern.ShapeID, metres, stop.RowNumber,
			)
		})
	}
}

// validateTripDistance checks that the trip does not claim to travel further
// than its own shape defines. Past the end of the shape there is no alignment
// left to place the vehicle on.
//
// The published wording for this rule describes a distance between two points
// rather than between two shape_dist_traveled values, which does not match its
// own name or its threshold. Read literally it would fail any feed whose shape
// happens to run a few metres past its last stop, which is both common and
// harmless, so the comparison here is the one the rule is named for.
func (v *ShapeGeometryValidator) validateTripDistance(container *notice.NoticeContainer, pattern *stopPattern, index *shapeIndex) {
	maxShape, ok := index.maxUserDistance()
	if !ok {
		return
	}

	maxTrip, ok := maxStopDistance(pattern.Stops)
	if !ok {
		return
	}

	overshoot := maxTrip - maxShape
	if overshoot <= 0 {
		return
	}

	report(container, pattern.TripIDs, func(tripID string) notice.Notice {
		if overshoot >= tripShapeOvershootThreshold {
			return notice.NewTripDistanceExceedsShapeDistanceNotice(tripID, pattern.ShapeID, maxTrip, maxShape)
		}
		return notice.NewTripDistanceExceedsShapeDistanceBelowThresholdNotice(tripID, pattern.ShapeID, maxTrip, maxShape)
	})
}

// report records one finding against every trip that shares the pattern.
func report(container *notice.NoticeContainer, tripIDs []string, build func(tripID string) notice.Notice) {
	for _, tripID := range tripIDs {
		container.AddNotice(build(tripID))
	}
}

// clusterMatches folds the raw segment matches into one entry per pass of the
// shape. Consecutive segments all report the same stop, so without this a
// straight run past a stop would look like a dozen separate matches.
func clusterMatches(matches []shapeMatch) []shapeMatch {
	if len(matches) == 0 {
		return nil
	}

	ordered := make([]shapeMatch, len(matches))
	copy(ordered, matches)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].Along < ordered[j].Along })

	passes := []shapeMatch{ordered[0]}
	lastAlong := ordered[0].Along

	for _, match := range ordered[1:] {
		current := &passes[len(passes)-1]
		if match.Along-lastAlong > matchSeparationMetres {
			passes = append(passes, match)
		} else if match.Metres < current.Metres {
			*current = match
		}
		lastAlong = match.Along
	}

	return passes
}

// chooseMatch anchors the stop the trip starts from at the earliest place it
// touches the shape, rather than the nearest one. On a loop the first stop sits
// within metres of both ends; anchoring at the near end would put the whole trip
// behind its own starting point and report every following stop as out of order.
func chooseMatch(passes []shapeMatch) shapeMatch {
	earliest := passes[0]
	for _, pass := range passes[1:] {
		if pass.Along < earliest.Along {
			earliest = pass
		}
	}
	return earliest
}

// maxStopDistance returns the furthest shape_dist_traveled the pattern's stop
// times declare.
func maxStopDistance(stops []tripStop) (float64, bool) {
	furthest := math.Inf(-1)
	found := false
	for _, stop := range stops {
		if stop.Dist != nil && *stop.Dist > furthest {
			furthest = *stop.Dist
			found = true
		}
	}
	return furthest, found
}

// loadShapes reads shapes.txt into polylines sorted by shape_pt_sequence.
func (v *ShapeGeometryValidator) loadShapes(loader *parser.FeedLoader) map[string][]shapePoint {
	shapes := make(map[string][]shapePoint)

	reader, err := loader.GetFile("shapes.txt")
	if err != nil {
		return shapes
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close reader %v", closeErr)
		}
	}()

	csvFile, err := parser.NewCSVFile(reader, "shapes.txt")
	if err != nil {
		return shapes
	}

	sequenced := make(map[string][]sequencedPoint)

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}

		shapeID := strings.TrimSpace(row.Values["shape_id"])
		lat, latErr := strconv.ParseFloat(strings.TrimSpace(row.Values["shape_pt_lat"]), 64)
		lon, lonErr := strconv.ParseFloat(strings.TrimSpace(row.Values["shape_pt_lon"]), 64)
		sequence, seqErr := strconv.Atoi(strings.TrimSpace(row.Values["shape_pt_sequence"]))
		if shapeID == "" || latErr != nil || lonErr != nil || seqErr != nil {
			continue
		}

		sequenced[shapeID] = append(sequenced[shapeID], sequencedPoint{
			Point: shapePoint{
				Lat:  lat,
				Lon:  lon,
				Dist: parseOptionalFloat(row.Values["shape_dist_traveled"]),
			},
			Sequence: sequence,
		})
	}

	for shapeID, points := range sequenced {
		sort.SliceStable(points, func(i, j int) bool { return points[i].Sequence < points[j].Sequence })

		polyline := make([]shapePoint, len(points))
		for i, point := range points {
			polyline[i] = point.Point
		}
		shapes[shapeID] = polyline
	}

	return shapes
}

// loadTripShapes maps each trip to its shape, skipping trips whose shape_id
// names a shape that is not in the feed — foreign key violations are reported
// elsewhere and there is no geometry here to check against.
func (v *ShapeGeometryValidator) loadTripShapes(loader *parser.FeedLoader, shapes map[string][]shapePoint) map[string]string {
	tripShapes := make(map[string]string)

	reader, err := loader.GetFile("trips.txt")
	if err != nil {
		return tripShapes
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close reader %v", closeErr)
		}
	}()

	csvFile, err := parser.NewCSVFile(reader, "trips.txt")
	if err != nil {
		return tripShapes
	}

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}

		tripID := strings.TrimSpace(row.Values["trip_id"])
		shapeID := strings.TrimSpace(row.Values["shape_id"])
		if tripID == "" || shapeID == "" {
			continue
		}
		if len(shapes[shapeID]) < 2 {
			continue
		}
		tripShapes[tripID] = shapeID
	}

	return tripShapes
}

// loadStopLocations reads the coordinates of every stop.
func (v *ShapeGeometryValidator) loadStopLocations(loader *parser.FeedLoader) map[string]*StopLocation {
	locations := make(map[string]*StopLocation)

	reader, err := loader.GetFile("stops.txt")
	if err != nil {
		return locations
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close reader %v", closeErr)
		}
	}()

	csvFile, err := parser.NewCSVFile(reader, "stops.txt")
	if err != nil {
		return locations
	}

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}

		stopID := strings.TrimSpace(row.Values["stop_id"])
		lat, latErr := strconv.ParseFloat(strings.TrimSpace(row.Values["stop_lat"]), 64)
		lon, lonErr := strconv.ParseFloat(strings.TrimSpace(row.Values["stop_lon"]), 64)
		if stopID == "" || latErr != nil || lonErr != nil {
			continue
		}
		locations[stopID] = &StopLocation{Latitude: lat, Longitude: lon}
	}

	return locations
}

// loadPatterns reads stop_times.txt and collapses the trips onto the distinct
// arrangements of shape, stops and declared distances they use. This is what
// keeps the cost proportional to the timetable's variety rather than its size:
// a feed with 40 000 trips typically has a few hundred patterns.
func (v *ShapeGeometryValidator) loadPatterns(loader *parser.FeedLoader, tripShapes map[string]string, stopLocations map[string]*StopLocation) map[string]*stopPattern {
	patterns := make(map[string]*stopPattern)

	reader, err := loader.GetFile("stop_times.txt")
	if err != nil {
		return patterns
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close reader %v", closeErr)
		}
	}()

	csvFile, err := parser.NewCSVFile(reader, "stop_times.txt")
	if err != nil {
		return patterns
	}

	byTrip := make(map[string][]tripStop)

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}

		tripID := strings.TrimSpace(row.Values["trip_id"])
		if _, wanted := tripShapes[tripID]; !wanted {
			continue
		}

		stopID := strings.TrimSpace(row.Values["stop_id"])
		sequence, seqErr := strconv.Atoi(strings.TrimSpace(row.Values["stop_sequence"]))
		if stopID == "" || seqErr != nil {
			continue
		}

		byTrip[tripID] = append(byTrip[tripID], tripStop{
			StopID:       stopID,
			StopSequence: sequence,
			Dist:         parseOptionalFloat(row.Values["shape_dist_traveled"]),
			Location:     stopLocations[stopID],
			RowNumber:    row.RowNumber,
		})
	}

	for tripID, stops := range byTrip {
		if len(stops) < 2 {
			continue
		}
		sort.SliceStable(stops, func(i, j int) bool { return stops[i].StopSequence < stops[j].StopSequence })

		key := patternKey(tripShapes[tripID], stops)
		if pattern, seen := patterns[key]; seen {
			pattern.TripIDs = append(pattern.TripIDs, tripID)
			continue
		}
		patterns[key] = &stopPattern{
			ShapeID: tripShapes[tripID],
			Stops:   stops,
			TripIDs: []string{tripID},
		}
	}

	for _, pattern := range patterns {
		sort.Strings(pattern.TripIDs)
	}

	return patterns
}

// patternKey identifies trips whose geometry checks would produce identical
// findings. Row numbers are deliberately left out: they differ between trips
// that are otherwise the same, and including them would defeat the grouping.
func patternKey(shapeID string, stops []tripStop) string {
	var key strings.Builder
	key.WriteString(shapeID)
	for _, stop := range stops {
		key.WriteByte(0)
		key.WriteString(stop.StopID)
		key.WriteByte(0)
		key.WriteString(strconv.Itoa(stop.StopSequence))
		if stop.Dist != nil {
			key.WriteByte(0)
			key.WriteString(strconv.FormatFloat(*stop.Dist, 'f', -1, 64))
		}
	}
	return key.String()
}

func parseOptionalFloat(raw string) *float64 {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil
	}
	value, err := strconv.ParseFloat(trimmed, 64)
	if err != nil {
		return nil
	}
	return &value
}
