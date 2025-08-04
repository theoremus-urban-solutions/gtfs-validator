package schema

// Pathway represents pathway connections from pathways.txt
type Pathway struct {
	PathwayID            string   `csv:"pathway_id"`
	FromStopID           string   `csv:"from_stop_id"`
	ToStopID             string   `csv:"to_stop_id"`
	PathwayMode          int      `csv:"pathway_mode"`
	IsBidirectional      int      `csv:"is_bidirectional"`
	Length               *float64 `csv:"length"`
	TraversalTime        *int     `csv:"traversal_time"`
	StairCount           *int     `csv:"stair_count"`
	MaxSlope             *float64 `csv:"max_slope"`
	MinWidth             *float64 `csv:"min_width"`
	SignpostedAs         string   `csv:"signposted_as"`
	ReversedSignpostedAs string   `csv:"reversed_signposted_as"`
}