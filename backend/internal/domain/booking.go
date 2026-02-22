package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/miketsu-inc/reservations/backend/internal/types"
	"github.com/miketsu-inc/reservations/backend/pkg/currencyx"
)

type BookingRepository interface {
	NewBookingByCustomer(context.Context, NewCustomerBooking) (int, error)
	NewBookingByMerchant(context.Context, NewMerchantBooking) (int, error)

	UpdateBookingData(context.Context, uuid.UUID, int, string, time.Duration) error

	CancelBookingByMerchant(context.Context, uuid.UUID, int, string) error
	CancelBookingByCustomer(context.Context, uuid.UUID, int) (uuid.UUID, error)

	GetCalendarEventsByMerchant(context.Context, uuid.UUID, string, string) (CalendarEvents, error)
	GetReservedTimes(context.Context, uuid.UUID, int, time.Time) ([]BookingTime, error)
	GetReservedTimesForPeriod(context.Context, uuid.UUID, int, time.Time, time.Time) ([]BookingTime, error)
	GetAvailableGroupBookingsForPeriod(context.Context, uuid.UUID, int, int, time.Time, time.Time) ([]BookingTime, error)

	GetPublicBooking(context.Context, int) (PublicBooking, error)
	GetBookingDataForEmail(context.Context, int) (BookingEmailData, error)
	UpdateEmailIdForBooking(context.Context, int, string, uuid.UUID) error

	TransferDummyBookings(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) error

	NewBookingSeries(context.Context, NewBookingSeries) (CompleteBookingSeries, error)
	BatchCreateRecurringBookings(context.Context, NewRecurringBookings) (int, error)
	GetExistingOccurrenceDates(context.Context, int, time.Time, time.Time) ([]time.Time, error)

	GetBookingForExternalCalendar(context.Context, int) (BookingForExternalCalendar, error)
}

type NewCustomerBooking struct {
	Status         types.BookingStatus  `json:"status"`
	BookingType    types.BookingType    `json:"booking_type"`
	BookingId      *int                 `json:"booking_id"`
	MerchantId     uuid.UUID            `json:"merchant_id"`
	ServiceId      int                  `json:"service_id"`
	LocationId     int                  `json:"location_id"`
	FromDate       time.Time            `json:"from_date"`
	ToDate         time.Time            `json:"to_date"`
	CustomerNote   *string              `json:"customer_note"`
	PricePerPerson currencyx.Price      `json:"price_per_person"`
	CostPerPerson  currencyx.Price      `json:"cost_per_person"`
	UserId         uuid.UUID            `json:"user_id"`
	CustomerId     uuid.UUID            `json:"customer_id"`
	Phases         []PublicServicePhase `json:"phases"`
}

type NewMerchantBooking struct {
	Status          types.BookingStatus  `json:"status"`
	BookingType     types.BookingType    `json:"booking_type"`
	MerchantId      uuid.UUID            `json:"merchant_id"`
	ServiceId       int                  `json:"service_id"`
	LocationId      int                  `json:"location_id"`
	FromDate        time.Time            `json:"from_date"`
	ToDate          time.Time            `json:"to_date"`
	MerchantNote    *string              `json:"merchant_note"`
	PricePerPerson  currencyx.Price      `json:"price_per_person"`
	CostPerPerson   currencyx.Price      `json:"cost_per_person"`
	MinParticipants int                  `json:"min_participants"`
	MaxParticipants int                  `json:"max_participants"`
	Participants    []*uuid.UUID         `json:"participants"`
	Phases          []PublicServicePhase `json:"phases"`
}

type PublicBooking struct {
	FromDate          time.Time                `json:"from_date" db:"from_date"`
	ToDate            time.Time                `json:"to_date" db:"to_date"`
	ServiceName       string                   `json:"service_name" db:"service_name"`
	CancelDeadline    int                      `json:"cancel_deadline" db:"cancel_deadline"`
	FormattedLocation string                   `json:"formatted_location" db:"formatted_location"`
	Price             currencyx.FormattedPrice `json:"price" db:"price"`
	PriceType         types.PriceType          `json:"price_type"`
	MerchantName      string                   `json:"merchant_name" db:"merchant_name"`
	IsCancelled       bool                     `json:"is_cancelled" db:"is_cancelled"`
}

type PublicBookingDetails struct {
	ID              int                      `json:"id" db:"id"`
	FromDate        time.Time                `json:"from_date" db:"from_date"`
	ToDate          time.Time                `json:"to_date" db:"to_date"`
	CustomerNote    *string                  `json:"customer_note" db:"customer_note"`
	MerchantNote    *string                  `json:"merchant_note" db:"merchant_note"`
	ServiceName     string                   `json:"service_name" db:"service_name"`
	ServiceColor    string                   `json:"service_color" db:"service_color"`
	ServiceDuration int                      `json:"service_duration" db:"service_duration"`
	Price           currencyx.FormattedPrice `json:"price" db:"price"`
	Cost            currencyx.FormattedPrice `json:"cost" db:"cost"`
	FirstName       *string                  `json:"first_name" db:"first_name"`
	LastName        *string                  `json:"last_name" db:"last_name"`
	PhoneNumber     *string                  `json:"phone_number" db:"phone_number"`
}

type BlockedTimeEvent struct {
	ID            int       `json:"id" db:"id"`
	EmployeeId    int       `json:"employee_id" db:"employee_id"`
	Name          string    `json:"name" db:"name"`
	FromDate      time.Time `json:"from_date" db:"from_date"`
	ToDate        time.Time `json:"to_date" db:"to_date"`
	AllDay        bool      `json:"all_day" db:"all_day"`
	Icon          *string   `json:"icon" db:"icon"`
	BlockedTypeId *int      `json:"blocked_type_id" db:"blocked_type_id"`
}

type CalendarEvents struct {
	Bookings     []PublicBookingDetails `json:"bookings"`
	BlockedTimes []BlockedTimeEvent     `json:"blocked_times"`
}

type BookingTime struct {
	From_date time.Time `db:"from_date"`
	To_date   time.Time `db:"to_date"`
}

type BookingEmailData struct {
	FromDate          time.Time              `json:"from_date" db:"from_date"`
	ToDate            time.Time              `json:"to_date" db:"to_date"`
	ServiceName       string                 `json:"service_name" db:"service_name"`
	FormattedLocation string                 `json:"formatted_location" db:"formatted_location"`
	MerchantName      string                 `json:"merchant_name" db:"merchant_name"`
	CancelDeadline    int                    `json:"cancel_deadline" db:"cancel_deadline"`
	Participants      []ParticipantEmailData `json:"participants"`
}

type ParticipantEmailData struct {
	CustomerId uuid.UUID `josn:"customer_id" db:"customer_id"`
	Email      *string   `json:"email" db:"email"`
	EmailId    uuid.UUID `json:"email_id" db:"email_id"`
}

type NewRecurringBookings struct {
	BookingSeriesId int
	BookingStatus   types.BookingStatus
	BookingType     types.BookingType
	MerchantId      uuid.UUID
	EmployeeId      int
	ServiceId       int
	LocationId      int
	FromDates       []time.Time
	ToDates         []time.Time
	Phases          []PublicServicePhase `json:"phases"`
	Details         BookingSeriesDetails
	Participants    []BookingSeriesParticipant
}

type BookingSeries struct {
	Id          int               `json:"id" db:"id"`
	BookingType types.BookingType `json:"booking_type" db:"booking_type"`
	MerchantId  uuid.UUID         `json:"merchant_id" db:"merchant_id"`
	EmployeeId  int               `json:"employee_id" db:"employee_id"`
	ServiceId   int               `json:"service_id" db:"service_id"`
	LocationId  int               `json:"location_id" db:"location_id"`
	Rrule       string            `json:"rrule" db:"rrule"`
	Dstart      time.Time         `json:"dstart" db:"dstart"`
	Timezone    string            `json:"timezone" db:"timezone"`
	IsActive    bool              `json:"is_active" db:"is_active"`
}

type BookingSeriesDetails struct {
	Id                  int             `json:"id" db:"id"`
	BookingSeriesId     int             `json:"booking_series_id" db:"booking_series_id"`
	PricePerPerson      currencyx.Price `json:"price_per_person" db:"price_per_person"`
	CostPerPerson       currencyx.Price `json:"cost_per_person" db:"cost_per_person"`
	TotalPrice          currencyx.Price `json:"total_price" db:"total_price"`
	TotalCost           currencyx.Price `json:"total_cost" db:"total_cost"`
	MinParticipants     int             `json:"min_participants" db:"min_participants"`
	MaxParticipants     int             `json:"max_participants" db:"max_participants"`
	CurrentParticipants int             `json:"current_participants" db:"current_participants"`
}

type BookingSeriesParticipant struct {
	Id              int        `json:"id" db:"id"`
	BookingSeriesId int        `json:"booking_series_id" db:"booking_series_id"`
	CustomerId      *uuid.UUID `json:"customer_id" db:"customer_id"`
	IsActive        bool       `json:"is_active" db:"is_active"`
	DroppedOutOn    *time.Time `json:"dropped_out_on" db:"dropped_out_on"`
}

type CompleteBookingSeries struct {
	BookingSeries
	Details      BookingSeriesDetails
	Participants []BookingSeriesParticipant
}

type NewBookingSeries struct {
	BookingType     types.BookingType
	MerchantId      uuid.UUID
	EmployeeId      int
	ServiceId       int
	LocationId      int
	Rrule           string
	Dstart          time.Time
	Timezone        *time.Location
	PricePerPerson  currencyx.Price
	CostPerPerson   currencyx.Price
	MinParticipants int
	MaxParticipants int
	Participants    []*uuid.UUID
}

type BookingForExternalCalendar struct {
	Id                  int                 `json:"id" db:"id"`
	Status              types.BookingStatus `json:"status" db:"status"`
	BookingType         types.BookingType   `json:"booking_type" db:"booking_type"`
	EmployeeId          *int                `json:"employee_id" db:"employee_id"`
	ServiceName         string              `json:"service_name" db:"service_name"`
	ServiceDescription  *string             `json:"service_description" db:"service_description"`
	PriceType           types.PriceType     `json:"price_type" db:"price_type"`
	FormattedLocation   string              `json:"formatted_location" db:"formatted_location"`
	FromDate            time.Time           `json:"from_date" db:"from_date"`
	ToDate              time.Time           `json:"to_date" db:"to_date"`
	TotalPrice          currencyx.Price     `json:"total_price" db:"total_price"`
	TotalCost           currencyx.Price     `json:"total_cost" db:"total_cost"`
	MerchantNote        *string             `json:"merchant_note" db:"merchant_note"`
	CurrentParticipants int                 `json:"current_participants" db:"current_participants"`
}
