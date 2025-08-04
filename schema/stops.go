package schema

// Stop represents a stop/station from stops.txt
type Stop struct {
	StopID             string  `csv:"stop_id"`
	StopCode           string  `csv:"stop_code"`
	StopName           string  `csv:"stop_name"`
	StopDesc           string  `csv:"stop_desc"`
	StopLat            float64 `csv:"stop_lat"`
	StopLon            float64 `csv:"stop_lon"`
	LocationType       int     `csv:"location_type"`
	ParentStation      string  `csv:"parent_station"`
	StopTimezone       string  `csv:"stop_timezone"`
	LevelID            string  `csv:"level_id"`
	StopURL            string  `csv:"stop_url"`
	WheelchairBoarding int     `csv:"wheelchair_boarding"`
	PlatformCode       string  `csv:"platform_code"`
	ZoneID             string  `csv:"zone_id"`
}
