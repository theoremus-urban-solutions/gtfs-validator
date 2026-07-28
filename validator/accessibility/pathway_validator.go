package accessibility

import (
	"io"
	"log"
	"strconv"
	"strings"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/parser"
	"github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

// Location types from stops.txt. Named because the pathway rules turn on them
// constantly and the bare numbers read as noise.
const (
	locationTypeStop         = 0
	locationTypeStation      = 1
	locationTypeEntrance     = 2
	locationTypeGenericNode  = 3
	locationTypeBoardingArea = 4
)

// Pathway modes that carry a rule of their own.
const (
	pathwayModeStairs    = 2
	pathwayModeEscalator = 4
	pathwayModeElevator  = 5
	pathwayModeExitGate  = 7
)

// parentChainLimit bounds the walk up parent_station. A feed with a cycle in
// its station hierarchy is reported by circular_station_reference; here the
// bound only needs to stop this validator spinning on one.
const parentChainLimit = 8

// PathwayValidator validates pathway definitions for accessibility
type PathwayValidator struct{}

// NewPathwayValidator creates a new pathway validator
func NewPathwayValidator() *PathwayValidator {
	return &PathwayValidator{}
}

// PathwayInfo represents pathway information.
//
// pathway_mode and is_bidirectional are pointers because a row that omits them
// or writes them as something other than a number still has endpoints worth
// checking. The absent value is reported by the required-field and type
// layers, not here.
type PathwayInfo struct {
	PathwayID            string
	FromStopID           string
	ToStopID             string
	PathwayMode          *int
	IsBidirectional      *int
	Length               *float64
	TraversalTime        *int
	StairCount           *int
	MaxSlope             *float64
	MinWidth             *float64
	SignpostedAs         string
	ReversedSignpostedAs string
	RowNumber            int
}

// pathwayStop is what the pathway rules need to know about a stops.txt row.
type pathwayStop struct {
	StopID           string
	StopName         string
	LocationType     int
	ParentStation    string
	StopAccess       string
	LevelID          string
	HasBoardingAreas bool
	RowNumber        int
}

// Validate checks pathway definitions
func (v *PathwayValidator) Validate(loader *parser.FeedLoader, container *notice.NoticeContainer, config validator.Config) {
	pathways := v.loadPathways(loader)
	if len(pathways) == 0 {
		return // No pathways to validate
	}

	stops := v.loadStopsForPathways(loader)

	for _, pathway := range pathways {
		v.validatePathway(container, pathway, stops)
	}

	// Check for duplicate pathways
	v.validateDuplicatePathways(container, pathways)

	// Validate bidirectional consistency
	v.validateBidirectionalConsistency(container, pathways)

	// The reachability rules need the graph rather than single rows.
	v.validateStationGraphs(container, pathways, stops)
}

// loadPathways loads pathway information from pathways.txt
func (v *PathwayValidator) loadPathways(loader *parser.FeedLoader) []*PathwayInfo {
	var pathways []*PathwayInfo

	reader, err := loader.GetFile("pathways.txt")
	if err != nil {
		return pathways
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close %v", closeErr)
		}
	}()

	csvFile, err := parser.NewCSVFile(reader, "pathways.txt")
	if err != nil {
		return pathways
	}

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}

		pathway := v.parsePathway(row)
		if pathway != nil {
			pathways = append(pathways, pathway)
		}
	}

	return pathways
}

// parsePathway parses a pathway record
func (v *PathwayValidator) parsePathway(row *parser.CSVRow) *PathwayInfo {
	pathwayID, hasPathwayID := row.Values["pathway_id"]
	fromStopID, hasFromStopID := row.Values["from_stop_id"]
	toStopID, hasToStopID := row.Values["to_stop_id"]

	if !hasPathwayID || !hasFromStopID || !hasToStopID {
		return nil
	}

	pathway := &PathwayInfo{
		PathwayID:       strings.TrimSpace(pathwayID),
		FromStopID:      strings.TrimSpace(fromStopID),
		ToStopID:        strings.TrimSpace(toStopID),
		PathwayMode:     optionalInt(row.Values["pathway_mode"]),
		IsBidirectional: optionalInt(row.Values["is_bidirectional"]),
		RowNumber:       row.RowNumber,
	}

	pathway.Length = optionalFloat(row.Values["length"])
	pathway.TraversalTime = optionalInt(row.Values["traversal_time"])
	pathway.StairCount = optionalInt(row.Values["stair_count"])
	pathway.MaxSlope = optionalFloat(row.Values["max_slope"])
	pathway.MinWidth = optionalFloat(row.Values["min_width"])
	pathway.SignpostedAs = strings.TrimSpace(row.Values["signposted_as"])
	pathway.ReversedSignpostedAs = strings.TrimSpace(row.Values["reversed_signposted_as"])

	return pathway
}

// hasMode reports whether the pathway declares the given mode.
func (p *PathwayInfo) hasMode(mode int) bool {
	return p.PathwayMode != nil && *p.PathwayMode == mode
}

// optionalInt returns the parsed value, or nil when the field is absent,
// blank, or not a number — the last of which the type layer reports.
func optionalInt(value string) *int {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	parsed, err := strconv.Atoi(trimmed)
	if err != nil {
		return nil
	}
	return &parsed
}

// optionalFloat is optionalInt for the decimal fields.
func optionalFloat(value string) *float64 {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	parsed, err := strconv.ParseFloat(trimmed, 64)
	if err != nil {
		return nil
	}
	return &parsed
}

// loadStopsForPathways loads stop information needed for pathway validation
func (v *PathwayValidator) loadStopsForPathways(loader *parser.FeedLoader) map[string]*pathwayStop {
	stops := make(map[string]*pathwayStop)

	reader, err := loader.GetFile("stops.txt")
	if err != nil {
		return stops
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Printf("Warning: failed to close %v", closeErr)
		}
	}()

	csvFile, err := parser.NewCSVFile(reader, "stops.txt")
	if err != nil {
		return stops
	}

	// A platform only becomes a parent object once a boarding area names it,
	// which may happen on a row we have not read yet.
	boardingAreaParents := make(map[string]bool)

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}

		stopID := strings.TrimSpace(row.Values["stop_id"])
		if stopID == "" {
			continue
		}

		stop := &pathwayStop{
			StopID:        stopID,
			StopName:      strings.TrimSpace(row.Values["stop_name"]),
			ParentStation: strings.TrimSpace(row.Values["parent_station"]),
			StopAccess:    strings.TrimSpace(row.Values["stop_access"]),
			LevelID:       strings.TrimSpace(row.Values["level_id"]),
			RowNumber:     row.RowNumber,
		}
		// An absent location_type means a platform; an unparseable one is the
		// type layer's to report, and treating it as a platform here keeps the
		// pathway rules from firing on a value nobody has validated yet.
		if locationType := optionalInt(row.Values["location_type"]); locationType != nil {
			stop.LocationType = *locationType
		}
		if stop.LocationType == locationTypeBoardingArea && stop.ParentStation != "" {
			boardingAreaParents[stop.ParentStation] = true
		}

		stops[stopID] = stop
	}

	for parentID := range boardingAreaParents {
		if parent, exists := stops[parentID]; exists {
			parent.HasBoardingAreas = true
		}
	}

	return stops
}

// validatePathway validates a single pathway record
func (v *PathwayValidator) validatePathway(container *notice.NoticeContainer, pathway *PathwayInfo, stops map[string]*pathwayStop) {
	// pathway_mode, is_bidirectional and the numeric fields are checked
	// against the spec's types by core/field_type_validator.go.

	// Validate stop references
	if _, exists := stops[pathway.FromStopID]; !exists {
		container.AddNotice(notice.NewForeignKeyViolationNotice(
			"pathways.txt",
			"from_stop_id",
			pathway.FromStopID,
			pathway.RowNumber,
			"stops.txt",
			"stop_id",
		))
	}

	if _, exists := stops[pathway.ToStopID]; !exists {
		container.AddNotice(notice.NewForeignKeyViolationNotice(
			"pathways.txt",
			"to_stop_id",
			pathway.ToStopID,
			pathway.RowNumber,
			"stops.txt",
			"stop_id",
		))
	}

	if pathway.FromStopID == pathway.ToStopID {
		container.AddNotice(notice.NewPathwayLoopNotice(
			pathway.PathwayID,
			pathway.RowNumber,
			pathway.FromStopID,
		))
	}

	// An exit gate only opens outwards, so a bidirectional one invites
	// consumers to route passengers in through a barrier.
	if pathway.hasMode(pathwayModeExitGate) && pathway.IsBidirectional != nil && *pathway.IsBidirectional == 1 {
		container.AddNotice(notice.NewBidirectionalExitGateNotice(
			pathway.PathwayID,
			pathway.RowNumber,
			pathway.FromStopID,
			pathway.ToStopID,
		))
	}

	v.validateEndpoints(container, pathway, stops)
	v.validateElevatorLevels(container, pathway, stops)
	v.validatePathwaySpecificRequirements(container, pathway)
}

// validateEndpoints checks what each end of a pathway is allowed to be.
func (v *PathwayValidator) validateEndpoints(container *notice.NoticeContainer, pathway *PathwayInfo, stops map[string]*pathwayStop) {
	for _, endpoint := range []struct {
		fieldName string
		stopID    string
	}{
		{"from_stop_id", pathway.FromStopID},
		{"to_stop_id", pathway.ToStopID},
	} {
		stop, exists := stops[endpoint.stopID]
		if !exists {
			continue // already reported as a foreign key violation
		}

		// A station is the whole building rather than a point somebody can
		// walk to, so it cannot terminate a pathway.
		if stop.LocationType == locationTypeStation {
			container.AddNotice(notice.NewPathwayToWrongLocationTypeNotice(
				pathway.PathwayID,
				pathway.RowNumber,
				endpoint.fieldName,
				stop.StopID,
				stop.LocationType,
			))
		}

		// Subdividing a platform into boarding areas makes the platform a
		// parent object; the pathway has to name the boarding area it reaches.
		if stop.LocationType == locationTypeStop && stop.HasBoardingAreas {
			container.AddNotice(notice.NewPathwayToPlatformWithBoardingAreasNotice(
				pathway.PathwayID,
				pathway.RowNumber,
				endpoint.fieldName,
				stop.StopID,
			))
		}

		// stop_access=1 says the stop is reached from the street, independently
		// of the station's pathways — which a pathway to it contradicts.
		if stop.StopAccess == "1" {
			container.AddNotice(notice.NewPathwayToStopWithAccessOutsideOfStationPathwaysNotice(
				pathway.PathwayID,
				pathway.RowNumber,
				endpoint.fieldName,
				stop.StopID,
			))
		}
	}
}

// validateElevatorLevels reports the endpoints of an elevator that do not say
// which level they are on. levels.txt is required as soon as a feed has an
// elevator, and an elevator between two unknown levels conveys nothing.
func (v *PathwayValidator) validateElevatorLevels(container *notice.NoticeContainer, pathway *PathwayInfo, stops map[string]*pathwayStop) {
	if !pathway.hasMode(pathwayModeElevator) {
		return
	}

	for _, stopID := range []string{pathway.FromStopID, pathway.ToStopID} {
		stop, exists := stops[stopID]
		if !exists || stop.LevelID != "" {
			continue
		}
		container.AddNotice(notice.NewMissingLevelIdNotice(
			stop.StopID,
			stop.RowNumber,
			stop.StopName,
		))
	}
}

// validatePathwaySpecificRequirements validates requirements specific to pathway modes
func (v *PathwayValidator) validatePathwaySpecificRequirements(container *notice.NoticeContainer, pathway *PathwayInfo) {
	if pathway.PathwayMode == nil {
		return
	}

	switch *pathway.PathwayMode {
	case pathwayModeStairs, pathwayModeEscalator:
		// Stairs and escalators should say how many steps they are, since that
		// is what tells a consumer whether the pathway is usable.
		if pathway.StairCount == nil {
			container.AddNotice(notice.NewMissingRecommendedFieldNotice(
				"pathways.txt",
				"stair_count",
				pathway.RowNumber,
			))
		}
	}
}

// validateDuplicatePathways checks for duplicate pathway definitions
func (v *PathwayValidator) validateDuplicatePathways(container *notice.NoticeContainer, pathways []*PathwayInfo) {
	pathwayMap := make(map[string]*PathwayInfo)

	for _, pathway := range pathways {
		key := pathway.FromStopID + "->" + pathway.ToStopID

		if existingPathway, exists := pathwayMap[key]; exists {
			container.AddNotice(notice.NewDuplicatePathwayNotice(
				pathway.PathwayID,
				pathway.FromStopID,
				pathway.ToStopID,
				pathway.RowNumber,
				existingPathway.RowNumber,
			))
		} else {
			pathwayMap[key] = pathway
		}
	}
}

// validateBidirectionalConsistency checks bidirectional pathway consistency
func (v *PathwayValidator) validateBidirectionalConsistency(container *notice.NoticeContainer, pathways []*PathwayInfo) {
	// Create map of pathway directions
	pathwayMap := make(map[string]*PathwayInfo)

	for _, pathway := range pathways {
		forward := pathway.FromStopID + "->" + pathway.ToStopID
		reverse := pathway.ToStopID + "->" + pathway.FromStopID

		pathwayMap[forward] = pathway

		// Check if there's a reverse pathway when this one is not bidirectional
		if pathway.IsBidirectional != nil && *pathway.IsBidirectional == 0 {
			if reversePathway, hasReverse := pathwayMap[reverse]; hasReverse {
				// Both directions exist as separate pathways - this is valid
				// But check if they have consistent properties
				if !samePathwayMode(pathway, reversePathway) {
					container.AddNotice(notice.NewInconsistentBidirectionalPathwayNotice(
						pathway.PathwayID,
						reversePathway.PathwayID,
						pathway.RowNumber,
						reversePathway.RowNumber,
					))
				}
			}
		}
	}
}

// samePathwayMode compares two pathways' modes, treating an unstated mode as
// nothing to disagree about.
func samePathwayMode(a *PathwayInfo, b *PathwayInfo) bool {
	if a.PathwayMode == nil || b.PathwayMode == nil {
		return true
	}
	return *a.PathwayMode == *b.PathwayMode
}

// stationGraph is one station's pathway graph: the locations the station's
// pathways touch, and the directed edges between them.
type stationGraph struct {
	nodes map[string]bool
	out   map[string][]string
	in    map[string][]string
}

func newStationGraph() *stationGraph {
	return &stationGraph{
		nodes: make(map[string]bool),
		out:   make(map[string][]string),
		in:    make(map[string][]string),
	}
}

// addEdge records a directed traversal from one location to another.
func (g *stationGraph) addEdge(from string, to string) {
	g.nodes[from] = true
	g.nodes[to] = true
	g.out[from] = append(g.out[from], to)
	g.in[to] = append(g.in[to], from)
}

// validateStationGraphs reports the rules that depend on how a station's
// pathways connect rather than on any single row.
//
// The graph is built per station and walked there, so a feed with a thousand
// stations does a thousand small traversals rather than one over everything.
func (v *PathwayValidator) validateStationGraphs(container *notice.NoticeContainer, pathways []*PathwayInfo, stops map[string]*pathwayStop) {
	stationOf := make(map[string]string, len(stops))
	graphs := make(map[string]*stationGraph)

	for _, pathway := range pathways {
		from, fromExists := stops[pathway.FromStopID]
		to, toExists := stops[pathway.ToStopID]
		if !fromExists || !toExists {
			continue // dangling reference, already reported
		}

		// Both ends belong to the same station in a well-formed feed; keying on
		// the from end keeps a malformed pathway in one graph rather than two.
		station := rootStation(stops, stationOf, from.StopID)
		graph, exists := graphs[station]
		if !exists {
			graph = newStationGraph()
			graphs[station] = graph
		}

		graph.addEdge(from.StopID, to.StopID)
		if pathway.IsBidirectional != nil && *pathway.IsBidirectional == 1 {
			graph.addEdge(to.StopID, from.StopID)
		}
	}

	for _, graph := range graphs {
		v.validateStationGraph(container, graph, stops)
	}
}

// validateStationGraph walks one station's graph for reachability and for
// generic nodes that lead nowhere.
func (v *PathwayValidator) validateStationGraph(container *notice.NoticeContainer, graph *stationGraph, stops map[string]*pathwayStop) {
	var entrances []string
	for stopID := range graph.nodes {
		if stop, exists := stops[stopID]; exists && stop.LocationType == locationTypeEntrance {
			entrances = append(entrances, stopID)
		}
	}

	// Reachability is asymmetric: walking in from an entrance and walking out
	// to one are different questions, and a one-way pathway can answer them
	// differently. The entrances are both the sources and the sinks.
	reachableFromEntrance := reachable(graph.out, entrances)
	canReachExit := reachable(graph.in, entrances)

	for stopID := range graph.nodes {
		stop, exists := stops[stopID]
		if !exists {
			continue
		}

		switch stop.LocationType {
		case locationTypeStop, locationTypeGenericNode, locationTypeBoardingArea:
		default:
			// An entrance is reachable by definition and a station is not a
			// point on the graph, so neither is reported.
			continue
		}

		hasEntrance := reachableFromEntrance[stopID]
		hasExit := canReachExit[stopID]
		if !hasEntrance || !hasExit {
			container.AddNotice(notice.NewPathwayUnreachableLocationNotice(
				stop.StopID,
				stop.RowNumber,
				stop.StopName,
				stop.LocationType,
				stop.ParentStation,
				hasEntrance,
				hasExit,
			))
		}

		// A generic node joins parts of a station together. One with a single
		// neighbour joins nothing, so no route ever has reason to pass through
		// it.
		if stop.LocationType == locationTypeGenericNode && incidentLocations(graph, stopID) == 1 {
			container.AddNotice(notice.NewPathwayDanglingGenericNodeNotice(
				stop.StopID,
				stop.RowNumber,
				stop.StopName,
				stop.ParentStation,
			))
		}
	}
}

// incidentLocations counts the distinct locations a node is joined to,
// ignoring direction and ignoring a loop back to itself.
func incidentLocations(graph *stationGraph, stopID string) int {
	neighbours := make(map[string]bool)
	for _, list := range [][]string{graph.out[stopID], graph.in[stopID]} {
		for _, neighbour := range list {
			if neighbour != stopID {
				neighbours[neighbour] = true
			}
		}
	}
	return len(neighbours)
}

// reachable returns every node the walk reaches from any of the given starts,
// following the supplied adjacency.
func reachable(adjacency map[string][]string, starts []string) map[string]bool {
	seen := make(map[string]bool, len(starts))
	queue := make([]string, 0, len(starts))

	for _, start := range starts {
		if !seen[start] {
			seen[start] = true
			queue = append(queue, start)
		}
	}

	for len(queue) > 0 {
		current := queue[len(queue)-1]
		queue = queue[:len(queue)-1]
		for _, next := range adjacency[current] {
			if !seen[next] {
				seen[next] = true
				queue = append(queue, next)
			}
		}
	}

	return seen
}

// rootStation walks parent_station up to the station a location belongs to,
// memoising the answer since a station's every platform asks for it.
func rootStation(stops map[string]*pathwayStop, cache map[string]string, stopID string) string {
	if station, known := cache[stopID]; known {
		return station
	}

	current := stopID
	for range parentChainLimit {
		stop, exists := stops[current]
		if !exists || stop.ParentStation == "" {
			break
		}
		current = stop.ParentStation
	}

	cache[stopID] = current
	return current
}
