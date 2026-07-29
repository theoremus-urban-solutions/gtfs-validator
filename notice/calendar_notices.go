package notice

// Notices about the span of dates a feed and its calendars remain useful for,
// from the Canonical GTFS Schedule Validator's rule set. Codes and severities
// match the published rules; the checks are our own implementations.

// ExpiredCalendarNotice reports a service whose last active date has already
// passed. Nothing can be planned on it, and a feed carrying a pile of them
// looks better covered than it is. The last active date is the date the service
// really last runs on, which both calendar.txt and calendar_dates.txt have a
// say in; see the emitting validator for how it is arrived at.
type ExpiredCalendarNotice struct {
	*BaseNotice
}

func NewExpiredCalendarNotice(rowNumber int, serviceID string) *ExpiredCalendarNotice {
	context := map[string]interface{}{
		"csvRowNumber": rowNumber,
		"serviceId":    serviceID,
	}
	return &ExpiredCalendarNotice{
		BaseNotice: NewBaseNotice("expired_calendar", WARNING, context),
	}
}

// FutureCalendarNotice reports that every service in the feed starts after
// today, so none of them covers the date the feed is read on. Publishing ahead
// of a timetable change is legitimate; a consumer loading the feed today still
// gets an empty schedule out of it.
type FutureCalendarNotice struct {
	*BaseNotice
}

func NewFutureCalendarNotice(currentDate string, earliestServiceDate string) *FutureCalendarNotice {
	context := map[string]interface{}{
		"currentDate":         currentDate,
		"earliestServiceDate": earliestServiceDate,
	}
	return &FutureCalendarNotice{
		BaseNotice: NewBaseNotice("future_calendar", INFO, context),
	}
}

// FutureFeedNotice reports a feed_info.txt declaring a feed_start_date later
// than today. The feed says of itself that it covers the future only, leaving
// the present uncovered.
type FutureFeedNotice struct {
	*BaseNotice
}

func NewFutureFeedNotice(rowNumber int, feedStartDate string, currentDate string) *FutureFeedNotice {
	context := map[string]interface{}{
		"csvRowNumber":  rowNumber,
		"feedStartDate": feedStartDate,
		"currentDate":   currentDate,
	}
	return &FutureFeedNotice{
		BaseNotice: NewBaseNotice("future_feed", INFO, context),
	}
}

// BigGapInServiceNotice reports a stretch of more than 13 days inside the feed's
// service window on which nothing runs. A hole that long is usually a calendar
// somebody forgot to extend rather than a real shutdown, and a trip planner
// returns nothing at all for those dates.
//
// The gap is measured across the feed as a whole rather than per service, so an
// ordinary holiday-only calendar does not report one on its own.
type BigGapInServiceNotice struct {
	*BaseNotice
}

func NewBigGapInServiceNotice(previousServiceDate string, nextServiceDate string, gapDays int) *BigGapInServiceNotice {
	context := map[string]interface{}{
		"previousServiceDate": previousServiceDate,
		"nextServiceDate":     nextServiceDate,
		"gapDays":             gapDays,
	}
	return &BigGapInServiceNotice{
		BaseNotice: NewBaseNotice("big_gap_in_service", INFO, context),
	}
}

// FeedValidBeyondTotalServiceWindowNotice reports a feed_info.txt whose
// feed_end_date runs more than 14 days past the last date any trip is scheduled
// on. The feed claims a validity it has no service to back, so a consumer
// trusting feed_end_date plans against an empty schedule.
type FeedValidBeyondTotalServiceWindowNotice struct {
	*BaseNotice
}

func NewFeedValidBeyondTotalServiceWindowNotice(rowNumber int, feedEndDate string, lastServiceDate string, daysBeyond int) *FeedValidBeyondTotalServiceWindowNotice {
	context := map[string]interface{}{
		"csvRowNumber":    rowNumber,
		"feedEndDate":     feedEndDate,
		"lastServiceDate": lastServiceDate,
		"daysBeyond":      daysBeyond,
	}
	return &FeedValidBeyondTotalServiceWindowNotice{
		BaseNotice: NewBaseNotice("feed_valid_beyond_total_service_window", INFO, context),
	}
}

// ServiceWindowOutsideFeedPeriodNotice reports one service whose active dates
// reach outside the validity period feed_info.txt declares for the feed. The
// two disagree about what the feed contains, and a consumer that honours
// feed_start_date and feed_end_date drops whatever falls outside them.
//
// The subject is the individual service rather than the feed as a whole,
// because that is what a publisher has to go and fix: one notice names one
// calendar to correct, and the number of them is how far the two are apart.
type ServiceWindowOutsideFeedPeriodNotice struct {
	*BaseNotice
}

func NewServiceWindowOutsideFeedPeriodNotice(serviceID string, serviceWindowStartDate string, serviceWindowEndDate string, daysBeforeFeedStart int, daysAfterFeedEnd int) *ServiceWindowOutsideFeedPeriodNotice {
	context := map[string]interface{}{
		"serviceId":              serviceID,
		"serviceWindowStartDate": serviceWindowStartDate,
		"serviceWindowEndDate":   serviceWindowEndDate,
		"daysBeforeFeedStart":    daysBeforeFeedStart,
		"daysAfterFeedEnd":       daysAfterFeedEnd,
	}
	return &ServiceWindowOutsideFeedPeriodNotice{
		BaseNotice: NewBaseNotice("service_window_outside_feed_period", INFO, context),
	}
}

// ServiceExtendsFarInTheFutureNotice reports a service whose last active date is
// more than two years out. At that distance the dates are a placeholder rather
// than a plan, and they make the feed look committed far past where anyone has
// actually checked the schedule.
type ServiceExtendsFarInTheFutureNotice struct {
	*BaseNotice
}

func NewServiceExtendsFarInTheFutureNotice(rowNumber int, serviceID string, endDate string, currentDate string) *ServiceExtendsFarInTheFutureNotice {
	context := map[string]interface{}{
		"csvRowNumber": rowNumber,
		"serviceId":    serviceID,
		"endDate":      endDate,
		"currentDate":  currentDate,
	}
	return &ServiceExtendsFarInTheFutureNotice{
		BaseNotice: NewBaseNotice("service_extends_far_in_the_future", INFO, context),
	}
}
