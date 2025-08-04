package schema

// CalendarDate represents service exceptions from calendar_dates.txt
type CalendarDate struct {
	ServiceID     string `csv:"service_id"`
	Date          string `csv:"date"`
	ExceptionType int    `csv:"exception_type"`
}
