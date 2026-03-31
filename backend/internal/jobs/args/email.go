package args

import (
	"time"

	"github.com/google/uuid"
	"github.com/riverqueue/river"
	"golang.org/x/text/language"
)

type BookingConfirmationEmail struct {
	Language   language.Tag `json:"language"`
	BookingId  int          `json:"booking_id"`
	CustomerId uuid.UUID    `json:"customer_id"`
}

func (BookingConfirmationEmail) Kind() string { return "booking_confirmation_email" }

func (BookingConfirmationEmail) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue: "email",
	}
}

type BookingReminderEmail struct {
	Language         language.Tag `json:"language"`
	BookingId        int          `json:"booking_id"`
	CustomerId       uuid.UUID    `json:"customer_id"`
	ExpectedFromDate time.Time    `json:"expected_from_date"`
}

func (BookingReminderEmail) Kind() string { return "booking_reminder_email" }

func (BookingReminderEmail) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue: "email",
	}
}

type BookingCancellationEmail struct {
	Language           language.Tag `json:"language"`
	BookingId          int          `json:"booking_id"`
	CustomerId         uuid.UUID    `json:"customer_id"`
	CancellationReason string       `json:"cancellation_reason"`
}

func (BookingCancellationEmail) Kind() string { return "booking_cancellation_email" }

func (BookingCancellationEmail) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue: "email",
	}
}

type BookingModificationEmail struct {
	Language       language.Tag `json:"language"`
	BookingId      int          `json:"booking_id"`
	CustomerId     uuid.UUID    `json:"customer_id"`
	OldServiceName string       `json:"service_name"`
	OldFromDate    time.Time    `json:"old_from_date"`
	OldToDate      time.Time    `json:"old_to_date"`
}

func (BookingModificationEmail) Kind() string { return "booking_modification_email" }

func (BookingModificationEmail) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue: "email",
	}
}
