package timeutil

import "time"

type DateRange struct {
	// Date part of start in the provided timezone converted to UTC.
	// Meaning it might differ from the exact date. Useful for querying UTC stored dates
	StartTime time.Time
	// Date part of end in the provided timezone converted to UTC.
	// Meaning it might differ from the exact date. Useful for querying UTC stored dates
	EndTime time.Time
	// Date part of start in UTC (no conversion).
	// Meaning it is the exact date as start just in UTC. Useful for querying all day dates
	StartDay time.Time
	// Date part of end in UTC (no conversion).
	// Meaning it is the exact date as end just in UTC. Useful for querying all day dates
	EndDay time.Time
}

// StartDay and EndDay are the UTC equivalent of the start and end time's dates
// while StartTime and EndTime are the dates in the provided timezone converted to UTC
func NewDateRange(start, end time.Time, tz *time.Location) DateRange {
	startYear, startMonth, startDay := start.Date()
	endYear, endMonth, endDay := end.Date()

	return DateRange{
		StartTime: time.Date(startYear, startMonth, startDay, 0, 0, 0, 0, tz).UTC(),
		EndTime:   time.Date(endYear, endMonth, endDay, 0, 0, 0, 0, tz).UTC(),
		StartDay:  time.Date(startYear, startMonth, startDay, 0, 0, 0, 0, time.UTC),
		EndDay:    time.Date(endYear, endMonth, endDay, 0, 0, 0, 0, time.UTC),
	}
}

// NewDateRangeFromInclusiveEnd converts the local dates containing start and end into
// a half-open range by adding a day to the end.
func NewDateRangeFromInclusiveEnd(start, end time.Time, tz *time.Location) DateRange {
	localStart := start.In(tz)
	localEnd := end.In(tz).AddDate(0, 0, 1)

	return NewDateRange(localStart, localEnd, tz)
}
