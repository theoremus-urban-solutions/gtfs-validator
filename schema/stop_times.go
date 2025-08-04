package schema

// StopTime represents a stop time from stop_times.txt
type StopTime struct {
	TripID            string `csv:"trip_id"`
	ArrivalTime       string `csv:"arrival_time"`
	DepartureTime     string `csv:"departure_time"`
	StopID            string `csv:"stop_id"`
	StopSequence      int    `csv:"stop_sequence"`
	StopHeadsign      string `csv:"stop_headsign"`
	PickupType        string `csv:"pickup_type"`
	DropOffType       string `csv:"drop_off_type"`
	ShapeDistTraveled string `csv:"shape_dist_traveled"`
	ContinuousPickup  string `csv:"continuous_pickup"`
	ContinuousDropOff string `csv:"continuous_drop_off"`
	Timepoint         int    `csv:"timepoint"`
}