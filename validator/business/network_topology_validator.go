package business

import (
	"io"
	"sort"
	"strconv"
	"strings"

	"github.com/theoremus-urban-solutions/gtfs-validator/notice"
	"github.com/theoremus-urban-solutions/gtfs-validator/parser"
	"github.com/theoremus-urban-solutions/gtfs-validator/validator"
)

// NetworkTopologyValidator validates network connectivity and topology
type NetworkTopologyValidator struct{}

// NewNetworkTopologyValidator creates a new network topology validator
func NewNetworkTopologyValidator() *NetworkTopologyValidator {
	return &NetworkTopologyValidator{}
}

// NetworkNode represents a stop in the network graph
type NetworkNode struct {
	StopID      string
	Connections map[string]*NetworkEdge
	RouteCount  int
	TripCount   int
	IsTransfer  bool
	Centrality  float64
}

// NetworkEdge represents a connection between two stops
type NetworkEdge struct {
	FromStopID string
	ToStopID   string
	RouteIDs   map[string]bool
	TripCount  int
	Weight     float64
}

// NetworkGraph represents the complete transit network
type NetworkGraph struct {
	Nodes    map[string]*NetworkNode
	Edges    map[string]*NetworkEdge
	RouteMap map[string][]string // route_id -> stop_ids
	TripMap  map[string][]string // trip_id -> stop_ids
}

// ConnectedComponent represents a connected component in the network
type ConnectedComponent struct {
	StopIDs    []string
	StopCount  int
	RouteIDs   map[string]bool
	RouteCount int
}

// TransferOpportunity represents potential transfer points
type TransferOpportunity struct {
	StopID         string
	RouteCount     int
	ConnectedStops []string
	TransferValue  float64
}

// Validate performs comprehensive network topology validation
func (v *NetworkTopologyValidator) Validate(loader *parser.FeedLoader, container *notice.NoticeContainer, config validator.Config) {
	// Build network graph
	graph := v.buildNetworkGraph(loader)
	if len(graph.Nodes) == 0 {
		return
	}

	// Validate network connectivity
	v.validateNetworkConnectivity(container, graph)

	// Analyze network topology
	v.analyzeNetworkTopology(container, graph)

	// Identify transfer opportunities
	v.identifyTransferOpportunities(container, graph)

	// Validate routing efficiency
	v.validateRoutingEfficiency(container, graph)

	// Generate network summary
	v.generateNetworkSummary(container, graph)
}

// buildNetworkGraph constructs the network graph from GTFS data
func (v *NetworkTopologyValidator) buildNetworkGraph(loader *parser.FeedLoader) *NetworkGraph {
	graph := &NetworkGraph{
		Nodes:    make(map[string]*NetworkNode),
		Edges:    make(map[string]*NetworkEdge),
		RouteMap: make(map[string][]string),
		TripMap:  make(map[string][]string),
	}

	// Load trip patterns from stop_times.txt
	tripPatterns := v.loadTripPatterns(loader)

	// Load trip-route mapping from trips.txt
	tripRoutes := v.loadTripRoutes(loader)

	// Build graph from trip patterns
	for tripID, stopSequence := range tripPatterns {
		graph.TripMap[tripID] = stopSequence

		if routeID := tripRoutes[tripID]; routeID != "" {
			graph.RouteMap[routeID] = append(graph.RouteMap[routeID], stopSequence...)
		}

		// Create nodes for stops
		for _, stopID := range stopSequence {
			if graph.Nodes[stopID] == nil {
				graph.Nodes[stopID] = &NetworkNode{
					StopID:      stopID,
					Connections: make(map[string]*NetworkEdge),
				}
			}
			graph.Nodes[stopID].TripCount++
		}

		// Create edges between consecutive stops
		for i := 1; i < len(stopSequence); i++ {
			fromStop := stopSequence[i-1]
			toStop := stopSequence[i]

			edgeKey := fromStop + "->" + toStop

			if graph.Edges[edgeKey] == nil {
				graph.Edges[edgeKey] = &NetworkEdge{
					FromStopID: fromStop,
					ToStopID:   toStop,
					RouteIDs:   make(map[string]bool),
					Weight:     1.0,
				}

				// Add edge to node connections
				graph.Nodes[fromStop].Connections[toStop] = graph.Edges[edgeKey]
			}

			graph.Edges[edgeKey].TripCount++
			if routeID := tripRoutes[tripID]; routeID != "" {
				graph.Edges[edgeKey].RouteIDs[routeID] = true
			}
		}
	}

	// Calculate route counts for nodes
	for _, stopList := range graph.RouteMap {
		uniqueStops := make(map[string]bool)
		for _, stopID := range stopList {
			uniqueStops[stopID] = true
		}

		for stopID := range uniqueStops {
			if node := graph.Nodes[stopID]; node != nil {
				node.RouteCount++
			}
		}
	}

	// Identify transfer nodes (served by multiple routes)
	for _, node := range graph.Nodes {
		if node.RouteCount > 1 {
			node.IsTransfer = true
		}
	}

	return graph
}

// loadTripPatterns loads trip patterns from stop_times.txt
func (v *NetworkTopologyValidator) loadTripPatterns(loader *parser.FeedLoader) map[string][]string {
	patterns := make(map[string][]string)

	reader, err := loader.GetFile("stop_times.txt")
	if err != nil {
		return patterns
	}
	defer reader.Close()

	csvFile, err := parser.NewCSVFile(reader, "stop_times.txt")
	if err != nil {
		return patterns
	}

	// Temporary storage for sorting
	tripStops := make(map[string][]StopSequence)

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		stopTime := v.parseStopTimeForNetwork(row)
		if stopTime != nil {
			tripStops[stopTime.TripID] = append(tripStops[stopTime.TripID], *stopTime)
		}
	}

	// Sort by stop_sequence and build patterns
	for tripID, stops := range tripStops {
		sort.Slice(stops, func(i, j int) bool {
			return stops[i].StopSequence < stops[j].StopSequence
		})

		stopSequence := make([]string, len(stops))
		for i, stop := range stops {
			stopSequence[i] = stop.StopID
		}
		patterns[tripID] = stopSequence
	}

	return patterns
}

// StopSequence represents a stop time for network analysis
type StopSequence struct {
	TripID       string
	StopID       string
	StopSequence int
}

// parseStopTimeForNetwork parses stop time for network analysis
func (v *NetworkTopologyValidator) parseStopTimeForNetwork(row *parser.CSVRow) *StopSequence {
	tripID, hasTripID := row.Values["trip_id"]
	stopID, hasStopID := row.Values["stop_id"]
	stopSeqStr, hasStopSeq := row.Values["stop_sequence"]

	if !hasTripID || !hasStopID || !hasStopSeq {
		return nil
	}

	stopSeq, err := strconv.Atoi(strings.TrimSpace(stopSeqStr))
	if err != nil {
		return nil
	}

	return &StopSequence{
		TripID:       strings.TrimSpace(tripID),
		StopID:       strings.TrimSpace(stopID),
		StopSequence: stopSeq,
	}
}

// loadTripRoutes loads trip-route mapping from trips.txt
func (v *NetworkTopologyValidator) loadTripRoutes(loader *parser.FeedLoader) map[string]string {
	tripRoutes := make(map[string]string)

	reader, err := loader.GetFile("trips.txt")
	if err != nil {
		return tripRoutes
	}
	defer reader.Close()

	csvFile, err := parser.NewCSVFile(reader, "trips.txt")
	if err != nil {
		return tripRoutes
	}

	for {
		row, err := csvFile.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		tripID, hasTripID := row.Values["trip_id"]
		routeID, hasRouteID := row.Values["route_id"]

		if hasTripID && hasRouteID {
			tripRoutes[strings.TrimSpace(tripID)] = strings.TrimSpace(routeID)
		}
	}

	return tripRoutes
}

// validateNetworkConnectivity validates network connectivity
func (v *NetworkTopologyValidator) validateNetworkConnectivity(container *notice.NoticeContainer, graph *NetworkGraph) {
	// Find connected components
	components := v.findConnectedComponents(graph)

	// Check for isolated stops (no connections)
	for stopID, node := range graph.Nodes {
		if len(node.Connections) == 0 {
			container.AddNotice(notice.NewIsolatedStopNotice(
				stopID,
				node.TripCount,
			))
		}
	}

	// Check for network fragmentation
	if len(components) > 1 {
		// Sort components by size
		sort.Slice(components, func(i, j int) bool {
			return components[i].StopCount > components[j].StopCount
		})

		mainComponentSize := components[0].StopCount
		totalStops := len(graph.Nodes)

		// If main component has less than 80% of stops, network is fragmented
		if float64(mainComponentSize)/float64(totalStops) < 0.8 {
			container.AddNotice(notice.NewFragmentedNetworkNotice(
				len(components),
				mainComponentSize,
				totalStops,
			))

			// Report smaller components
			for i := 1; i < len(components) && i < 5; i++ {
				component := components[i]
				container.AddNotice(notice.NewSmallNetworkComponentNotice(
					i+1,
					component.StopCount,
					component.RouteCount,
				))
			}
		}
	}
}

// findConnectedComponents finds connected components in the network
func (v *NetworkTopologyValidator) findConnectedComponents(graph *NetworkGraph) []*ConnectedComponent {
	visited := make(map[string]bool)
	var components []*ConnectedComponent

	for stopID := range graph.Nodes {
		if !visited[stopID] {
			component := v.dfsComponent(graph, stopID, visited)
			if component.StopCount > 0 {
				components = append(components, component)
			}
		}
	}

	return components
}

// dfsComponent performs DFS to find a connected component
func (v *NetworkTopologyValidator) dfsComponent(graph *NetworkGraph, startStop string, visited map[string]bool) *ConnectedComponent {
	component := &ConnectedComponent{
		StopIDs:  []string{},
		RouteIDs: make(map[string]bool),
	}

	stack := []string{startStop}

	for len(stack) > 0 {
		stopID := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if visited[stopID] {
			continue
		}

		visited[stopID] = true
		component.StopIDs = append(component.StopIDs, stopID)
		component.StopCount++

		// Add connected stops to stack
		if node := graph.Nodes[stopID]; node != nil {
			for connectedStop := range node.Connections {
				if !visited[connectedStop] {
					stack = append(stack, connectedStop)
				}
			}
		}

		// Track routes serving this component
		for routeID, stopList := range graph.RouteMap {
			for _, routeStopID := range stopList {
				if routeStopID == stopID {
					component.RouteIDs[routeID] = true
					break
				}
			}
		}
	}

	component.RouteCount = len(component.RouteIDs)
	return component
}

// analyzeNetworkTopology analyzes network topological properties
func (v *NetworkTopologyValidator) analyzeNetworkTopology(container *notice.NoticeContainer, graph *NetworkGraph) {
	// Calculate basic network metrics
	totalStops := len(graph.Nodes)
	totalEdges := len(graph.Edges)

	if totalStops == 0 {
		return
	}

	// Calculate average connectivity
	totalConnections := 0
	transferStops := 0
	for _, node := range graph.Nodes {
		totalConnections += len(node.Connections)
		if node.IsTransfer {
			transferStops++
		}
	}

	avgConnectivity := float64(totalConnections) / float64(totalStops)

	// Check for poorly connected network
	if avgConnectivity < 1.5 {
		container.AddNotice(notice.NewLowNetworkConnectivityNotice(
			totalStops,
			totalEdges,
			avgConnectivity,
		))
	}

	// Check transfer stop ratio
	transferRatio := float64(transferStops) / float64(totalStops)
	if transferRatio < 0.1 {
		container.AddNotice(notice.NewLowTransferOpportunityNotice(
			transferStops,
			totalStops,
			transferRatio,
		))
	}

	// Find hub stops (high connectivity)
	var hubStops []*NetworkNode
	for _, node := range graph.Nodes {
		if node.RouteCount >= 5 {
			hubStops = append(hubStops, node)
		}
	}

	// Sort hubs by route count
	sort.Slice(hubStops, func(i, j int) bool {
		return hubStops[i].RouteCount > hubStops[j].RouteCount
	})

	// Report major hubs
	if len(hubStops) > 0 {
		container.AddNotice(notice.NewNetworkHubIdentifiedNotice(
			hubStops[0].StopID,
			hubStops[0].RouteCount,
			len(hubStops[0].Connections),
			len(hubStops),
		))
	}
}

// identifyTransferOpportunities identifies potential transfer improvements
func (v *NetworkTopologyValidator) identifyTransferOpportunities(container *notice.NoticeContainer, graph *NetworkGraph) {
	var opportunities []*TransferOpportunity

	for stopID, node := range graph.Nodes {
		if node.RouteCount >= 2 {
			// Calculate transfer value based on route count and connectivity
			transferValue := float64(node.RouteCount) * float64(len(node.Connections))

			connectedStops := make([]string, 0, len(node.Connections))
			for connectedStop := range node.Connections {
				connectedStops = append(connectedStops, connectedStop)
			}

			opportunities = append(opportunities, &TransferOpportunity{
				StopID:         stopID,
				RouteCount:     node.RouteCount,
				ConnectedStops: connectedStops,
				TransferValue:  transferValue,
			})
		}
	}

	// Sort by transfer value
	sort.Slice(opportunities, func(i, j int) bool {
		return opportunities[i].TransferValue > opportunities[j].TransferValue
	})

	// Report top transfer opportunities
	for i := 0; i < len(opportunities) && i < 10; i++ {
		opp := opportunities[i]
		if opp.RouteCount >= 3 {
			container.AddNotice(notice.NewMajorTransferPointNotice(
				opp.StopID,
				opp.RouteCount,
				len(opp.ConnectedStops),
				opp.TransferValue,
			))
		}
	}
}

// validateRoutingEfficiency validates routing patterns for efficiency
func (v *NetworkTopologyValidator) validateRoutingEfficiency(container *notice.NoticeContainer, graph *NetworkGraph) {
	// Check for route overlaps (same stop sequence)
	routePatterns := make(map[string][]string) // pattern -> route_ids

	for routeID, stopList := range graph.RouteMap {
		// Create unique stop sequence
		uniqueStops := make(map[string]bool)
		for _, stopID := range stopList {
			uniqueStops[stopID] = true
		}

		stopSequence := make([]string, 0, len(uniqueStops))
		for stopID := range uniqueStops {
			stopSequence = append(stopSequence, stopID)
		}
		sort.Strings(stopSequence)

		pattern := strings.Join(stopSequence, "|")
		routePatterns[pattern] = append(routePatterns[pattern], routeID)
	}

	// Report overlapping routes
	for pattern, routeIDs := range routePatterns {
		if len(routeIDs) > 1 {
			container.AddNotice(notice.NewOverlappingRoutesNotice(
				routeIDs,
				len(strings.Split(pattern, "|")),
			))
		}
	}

	// Check for very short routes (< 3 stops)
	for routeID, stopList := range graph.RouteMap {
		uniqueStops := make(map[string]bool)
		for _, stopID := range stopList {
			uniqueStops[stopID] = true
		}

		if len(uniqueStops) < 2 {
			container.AddNotice(notice.NewVeryShortRouteNotice(
				routeID,
				len(uniqueStops),
				len(stopList), // trip count approximation
			))
		}
	}
}

// generateNetworkSummary generates comprehensive network analysis summary
func (v *NetworkTopologyValidator) generateNetworkSummary(container *notice.NoticeContainer, graph *NetworkGraph) {
	totalStops := len(graph.Nodes)
	totalEdges := len(graph.Edges)
	totalRoutes := len(graph.RouteMap)
	totalTrips := len(graph.TripMap)

	// Calculate connectivity metrics
	transferStops := 0
	maxRouteCount := 0
	totalConnections := 0

	for _, node := range graph.Nodes {
		if node.IsTransfer {
			transferStops++
		}
		if node.RouteCount > maxRouteCount {
			maxRouteCount = node.RouteCount
		}
		totalConnections += len(node.Connections)
	}

	avgConnectivity := 0.0
	if totalStops > 0 {
		avgConnectivity = float64(totalConnections) / float64(totalStops)
	}

	components := v.findConnectedComponents(graph)

	container.AddNotice(notice.NewNetworkTopologySummaryNotice(
		totalStops,
		totalEdges,
		totalRoutes,
		totalTrips,
		transferStops,
		len(components),
		avgConnectivity,
		maxRouteCount,
	))
}
