package schema

// Transfer represents a transfer rule from transfers.txt
type Transfer struct {
	FromStopID      string `csv:"from_stop_id"`
	ToStopID        string `csv:"to_stop_id"`
	FromRouteID     string `csv:"from_route_id"`
	ToRouteID       string `csv:"to_route_id"`
	FromTripID      string `csv:"from_trip_id"`
	ToTripID        string `csv:"to_trip_id"`
	TransferType    int    `csv:"transfer_type"`
	MinTransferTime int    `csv:"min_transfer_time"`
}