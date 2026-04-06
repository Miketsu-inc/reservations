package schedule

import "time"

type DailyMidnight struct {
	loc *time.Location
}

func NewDailyMidnight(loc *time.Location) *DailyMidnight {
	return &DailyMidnight{loc: loc}
}

func (s *DailyMidnight) Next(t time.Time) time.Time {
	t = t.In(s.loc)
	return time.Date(t.Year(), t.Month(), t.Day()+1, 0, 0, 0, 0, s.loc)
}
