package schema

// Route represents a transit route from routes.txt
type Route struct {
	RouteID           string `csv:"route_id"`
	AgencyID          string `csv:"agency_id"`
	RouteShortName    string `csv:"route_short_name"`
	RouteLongName     string `csv:"route_long_name"`
	RouteDesc         string `csv:"route_desc"`
	RouteType         int    `csv:"route_type"`
	RouteURL          string `csv:"route_url"`
	RouteColor        string `csv:"route_color"`
	RouteTextColor    string `csv:"route_text_color"`
	RouteSortOrder    string `csv:"route_sort_order"`
	ContinuousPickup  string `csv:"continuous_pickup"`
	ContinuousDropOff string `csv:"continuous_drop_off"`
}
