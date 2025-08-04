package schema

// Calendar represents regular service schedules from calendar.txt
type Calendar struct {
	ServiceID string `csv:"service_id"`
	Monday    int    `csv:"monday"`
	Tuesday   int    `csv:"tuesday"`
	Wednesday int    `csv:"wednesday"`
	Thursday  int    `csv:"thursday"`
	Friday    int    `csv:"friday"`
	Saturday  int    `csv:"saturday"`
	Sunday    int    `csv:"sunday"`
	StartDate string `csv:"start_date"`
	EndDate   string `csv:"end_date"`
}
