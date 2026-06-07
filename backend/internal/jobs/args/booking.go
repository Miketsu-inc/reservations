package args

import (
	"time"

	"github.com/google/uuid"
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
	BookingSeriesId int       `json:"booking_series_id"`
	GenerateFrom    time.Time `json:"generate_from"`
}

func (BookingOccurrenceGenerator) Kind() string { return "booking_occurrence_generator" }

func (BookingOccurrenceGenerator) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		UniqueOpts: river.UniqueOpts{
			ByArgs: true,
		},
	}
}

type UpdateFutureBookingOccurrences struct {
	BookingSeriesId          int           `json:"booking_series_id"`
	OccurrenceIndex          int           `json:"occurrence_index"`
	SeriesOriginalDateOffset time.Duration `json:"series_original_date_offset"`
	StatusChangedToCancelled bool          `json:"status_changed_to_cancelled"`
	PriceChanged             bool          `json:"price_changed"`
	ParticipantsToInsert     []uuid.UUID   `json:"particiapnts_to_insert"`
	ParticipantsToDelete     []uuid.UUID   `json:"particiapnts_to_delete"`
}

func (UpdateFutureBookingOccurrences) Kind() string { return "update_booking_occurrences" }
