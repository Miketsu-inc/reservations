package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/miketsu-inc/reservations/backend/internal/types"
	"github.com/miketsu-inc/reservations/backend/pkg/currencyx"
	"github.com/miketsu-inc/reservations/backend/pkg/db"
)

type BookingRepository interface {
	WithTx(tx db.DBTX) BookingRepository

	NewBooking(ctx context.Context, booking Booking) (int, error)
	NewBookings(ctx context.Context, bookings []Booking) ([]int, error)
	NewBookingPhases(ctx context.Context, bookingPhases []BookingPhase) error
	NewBookingDetails(ctx context.Context, bookingDetails BookingDetails) error
	NewBookingDetailsBatch(ctx context.Context, bookingDetails []BookingDetails) error
	NewBookingParticipants(ctx context.Context, bookingParticipants []BookingParticipant) error

	DeleteBookingPhases(ctx context.Context, bookingId int) error
	DeleteBookingParticipants(ctx context.Context, bookingId int, participantIds []uuid.UUID) error

	UpdateBookingStatus(ctx context.Context, merchantId uuid.UUID, bookingId int, status types.BookingStatus) error
	UpdateBookingCore(ctx context.Context, merchantId uuid.UUID, bookingId int, serviceId int, offset time.Duration, bookingType types.BookingType, status types.BookingStatus) error
	UpdateEmailIdForBooking(context.Context, int, string, uuid.UUID) error
	UpdateBookingTotalPrice(ctx context.Context, bookingId int, price, cost currencyx.Price) error
	UpdateBookingDetails(ctx context.Context, merchantId uuid.UUID, details BookingDetails) error
	UpdateBookingPhaseTime(ctx context.Context, bookingId int, offset time.Duration) error
	UpdateBookingParticipants(ctx context.Context, bookingId int, participantIds []uuid.UUID, status types.BookingStatus) error
	UpdateParticipantStatus(ctx context.Context, bookingId int, participantId int, status types.BookingStatus) error
	IncrementParticipantCount(ctx context.Context, bookingId int) (currencyx.Price, currencyx.Price, error)
	DecrementParticipantCount(ctx context.Context, bookingId int) error
	// decrements the participant count on every booking related to the customer
	DecrementEveryParticipantCountForCustomer(ctx context.Context, customerId uuid.UUID, merchantId uuid.UUID) error
	TransferDummyBookings(ctx context.Context, merchantId uuid.UUID, fromCustomerId uuid.UUID, toCustomerId uuid.UUID) error

	CancelBookingByMerchant(ctx context.Context, merchantId uuid.UUID, bookingId int, cancellationReason string) error
	CancelBookingByCustomer(ctx context.Context, bookingId int, customerId uuid.UUID) (types.BookingType, error)
	DeleteAppointmentsByCustomer(ctx context.Context, customerId uuid.UUID, merchantId uuid.UUID) error
	DeleteParticipantByCustomer(ctx context.Context, customerId uuid.UUID, merchantId uuid.UUID) error

	GetBooking(ctx context.Context, bookingId int) (Booking, error)
	GetPublicBooking(ctx context.Context, bookingId int) (PublicBooking, error)
	GetLatestBookings(ctx context.Context, merchantId uuid.UUID, afterDate time.Time, rowLimit int) ([]PublicBookingDetails, error)
	GetUpcomingBookings(ctx context.Context, merchantId uuid.UUID, afterDate time.Time, rowLimit int) ([]PublicBookingDetails, error)
	GetBookingsForCalendar(ctx context.Context, merchantId uuid.UUID, startTime, endTime string) ([]BookingForCalendar, error)
	GetBookingForExternalCalendar(ctx context.Context, bookingId int) (BookingForExternalCalendar, error)
	GetBookingForEmail(ctx context.Context, bookingId int, customerId uuid.UUID) (BookingForEmail, error)
	GetBookingDetails(ctx context.Context, bookingId int) (BookingDetails, error)
	GetBookingParticipants(ctx context.Context, bookingId int) ([]BookingParticipant, error)

	GetReservedTimes(ctx context.Context, merchantId uuid.UUID, locationId int, day time.Time) ([]BookingTime, error)
	GetReservedTimesForPeriod(ctx context.Context, merchantId uuid.UUID, locationiId int, startDate time.Time, endDate time.Time) ([]BookingTime, error)
	GetAvailableGroupBookingsForPeriod(ctx context.Context, merchantId uuid.UUID, serviceId int, locationId int, startDate time.Time, endDate time.Time) ([]BookingTime, error)

	NewBookingSeries(ctx context.Context, bookingSeries BookingSeries) (BookingSeries, error)
	NewBookingSeriesDetails(ctx context.Context, bookingSeriesDetails BookingSeriesDetails) (BookingSeriesDetails, error)
	NewBookingSeriesParticipants(ctx context.Context, bookingSeriesParticipants []BookingSeriesParticipant) ([]BookingSeriesParticipant, error)

	GetExistingOccurrenceDates(ctx context.Context, seriesId int, startDate time.Time, endDate time.Time) ([]time.Time, error)
}

type Booking struct {
	Id                 int
	Status             types.BookingStatus
	BookingType        types.BookingType
	IsRecurring        bool
	MerchantId         uuid.UUID
	EmployeeId         *int
	ServiceId          int
	LocationId         int
	BookingSeriesId    *int
	SeriesOriginalDate *time.Time
	FromDate           time.Time
	ToDate             time.Time
}

type BookingPhase struct {
	Id             int
	BookingId      int
	ServicePhaseId int
	FromDate       time.Time
	ToDate         time.Time
}

type BookingDetails struct {
	Id                    int
	BookingId             int
	PricePerPerson        currencyx.Price
	CostPerPerson         currencyx.Price
	TotalPrice            currencyx.Price
	TotalCost             currencyx.Price
	MerchantNote          *string
	MinParticipants       int
	MaxParticipants       int
	CurrentParticipants   int
	CancelledByMerchantOn *time.Time
	CancellationReason    *string
}

type BookingParticipant struct {
	Id                 int
	Status             types.BookingStatus
	BookingId          int
	CustomerId         *uuid.UUID
	CustomerNote       *string
	CancelledOn        *time.Time
	CancellationReason *string
	TransferredTo      *uuid.UUID
	EmailId            *uuid.UUID
}

type PublicBooking struct {
	FromDate          time.Time       `json:"from_date" db:"from_date"`
	ToDate            time.Time       `json:"to_date" db:"to_date"`
	ServiceName       string          `json:"service_name" db:"service_name"`
	CancelDeadline    int             `json:"cancel_deadline" db:"cancel_deadline"`
	FormattedLocation string          `json:"formatted_location" db:"formatted_location"`
	Price             currencyx.Price `json:"price" db:"price"`
	PriceType         types.PriceType `json:"price_type"`
	MerchantName      string          `json:"merchant_name" db:"merchant_name"`
	IsCancelled       bool            `json:"is_cancelled" db:"is_cancelled"`
}

type PublicBookingDetails struct {
	ID              int             `json:"id" db:"id"`
	FromDate        time.Time       `json:"from_date" db:"from_date"`
	ToDate          time.Time       `json:"to_date" db:"to_date"`
	CustomerNote    *string         `json:"customer_note" db:"customer_note"`
	MerchantNote    *string         `json:"merchant_note" db:"merchant_note"`
	ServiceName     string          `json:"service_name" db:"service_name"`
	ServiceColor    string          `json:"service_color" db:"service_color"`
	ServiceDuration int             `json:"service_duration" db:"service_duration"`
	Price           currencyx.Price `json:"price" db:"price"`
	Cost            currencyx.Price `json:"cost" db:"cost"`
	FirstName       *string         `json:"first_name" db:"first_name"`
	LastName        *string         `json:"last_name" db:"last_name"`
	PhoneNumber     *string         `json:"phone_number" db:"phone_number"`
}

type BookingForCalendar struct {
	ID              int                             `json:"id" db:"id"`
	BookingType     types.BookingType               `json:"booking_type" db:"booking_type"`
	BookingStatus   types.BookingStatus             `json:"booking_status" db:"booking_status"`
	FromDate        time.Time                       `json:"from_date" db:"from_date"`
	ToDate          time.Time                       `json:"to_date" db:"to_date"`
	IsRecurring     bool                            `json:"is_recurring" db:"is_recurring"`
	MerchantNote    *string                         `json:"merchant_note" db:"merchant_note"`
	ServiceId       int                             `json:"service_id" db:"service_id"`
	ServiceName     string                          `json:"service_name" db:"service_name"`
	ServiceColor    string                          `json:"service_color" db:"service_color"`
	MaxParticipants int                             `json:"max_participants" db:"max_participants"`
	Price           currencyx.Price                 `json:"price" db:"price"`
	Cost            currencyx.Price                 `json:"cost" db:"cost"`
	Participants    []BookingParticipantForCalendar `json:"participants" db:"participants"`
}

type BookingParticipantForCalendar struct {
	Id           int                 `json:"id" db:"id"`
	CustomerId   uuid.UUID           `json:"customer_id" db:"customer_id"`
	FirstName    *string             `json:"first_name" db:"first_name"`
	LastName     *string             `json:"last_name" db:"last_name"`
	CustomerNote *string             `json:"customer_note" db:"customer_note"`
	Status       types.BookingStatus `json:"participant_status" db:"participant_status"`
}

type CalendarEvents struct {
	Bookings     []BookingForCalendar `json:"bookings"`
	BlockedTimes []BlockedTimeEvent   `json:"blocked_times"`
}

type BookingTime struct {
	From_date time.Time `db:"from_date"`
	To_date   time.Time `db:"to_date"`
}

type BookingForEmail struct {
	Id                int                 `db:"id"`
	Status            types.BookingStatus `db:"status"`
	FromDate          time.Time           `db:"from_date"`
	ToDate            time.Time           `db:"to_date"`
	ServiceName       string              `db:"service_name"`
	ServiceId         int                 `db:"service_id"`
	MerchantName      string              `db:"merchant_name"`
	MerchantUrl       string              `db:"merchant_url"`
	Timezone          string              `db:"timezone"`
	CancelDeadline    int                 `db:"cancel_deadline"`
	FormattedLocation string              `db:"formatted_location"`
	CustomerId        uuid.UUID           `db:"customer_id"`
	CustomerEmail     *string             `db:"customer_email"`
	ParticipantStatus types.BookingStatus `db:"participant_status"`
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
