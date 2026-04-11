package args

import (
	"time"

	"github.com/riverqueue/river"
)

type RecurringBookingScheduler struct{}

func (RecurringBookingScheduler) Kind() string { return "recurring_booking_scheduler" }

func (RecurringBookingScheduler) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		UniqueOpts: river.UniqueOpts{
			ByPeriod: time.Hour * 24,
		},
	}
}

type BookingOccurrenceGenerator struct {
	BookingSeriesId int `json:"booking_series_id"`
}

func (BookingOccurrenceGenerator) Kind() string { return "booking_occurrence_generator" }

func (BookingOccurrenceGenerator) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		UniqueOpts: river.UniqueOpts{
			ByArgs: true,
		},
	}
}
