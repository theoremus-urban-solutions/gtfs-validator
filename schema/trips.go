package schema

// Trip represents a trip from trips.txt
type Trip struct {
	TripID               string `csv:"trip_id"`
	RouteID              string `csv:"route_id"`
	ServiceID            string `csv:"service_id"`
	TripHeadsign         string `csv:"trip_headsign"`
	TripShortName        string `csv:"trip_short_name"`
	DirectionID          string `csv:"direction_id"`
	BlockID              string `csv:"block_id"`
	ShapeID              string `csv:"shape_id"`
	WheelchairAccessible int    `csv:"wheelchair_accessible"`
	BikesAllowed         int    `csv:"bikes_allowed"`
}
