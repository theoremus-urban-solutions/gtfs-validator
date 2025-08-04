package schema

// Frequency represents feed metadata from frequencies.txt
type Frequency struct {
	TripID      string `csv:"trip_id"`
	StartTime   string `csv:"start_time"`
	EndTime     string `csv:"end_time"`
	HeadwaySecs int    `csv:"headway_secs"`
	ExactTimes  int    `csv:"exact_times"`
}
