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
	NewBookingParticipants(ctx context.Context, bookingParticipants []BookingParticipant) error

	DeleteBookingPhasesBatch(ctx context.Context, bookingIds []int) error
	DeleteBookingParticipantsBatch(ctx context.Context, bookingIds []int, participantIds []uuid.UUID) error

	UpdateBookingStatus(ctx context.Context, merchantId uuid.UUID, bookingId int, status types.BookingStatus) error
	UpdateBookingCoreBatch(ctx context.Context, merchantId uuid.UUID, bookingIds []int, serviceId int, fromDates []time.Time, toDates []time.Time, bookingType types.BookingType, status types.BookingStatus) error
	UpdateBookingTotalPrice(ctx context.Context, bookingId int, price, cost currencyx.Price) error
	UpdateBookingDetailsBatch(ctx context.Context, merchantId uuid.UUID, bookingIds []int, details BookingDetails) error
	UpdateBookingParticipants(ctx context.Context, participants []BookingParticipant) error
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
	GetBookingParticipants(ctx context.Context, bookingId int) ([]BookingParticipant, error)

	GetReservedTimes(ctx context.Context, merchantId uuid.UUID, locationId int, day time.Time) ([]BookingTime, error)
	GetReservedTimesForPeriod(ctx context.Context, merchantId uuid.UUID, locationiId int, startDate time.Time, endDate time.Time) ([]BookingTime, error)
	GetAvailableGroupBookingsForPeriod(ctx context.Context, merchantId uuid.UUID, serviceId int, locationId int, startDate time.Time, endDate time.Time) ([]BookingTime, error)

	NewBookingSeries(ctx context.Context, bookingSeries BookingSeries) (BookingSeries, error)
	NewBookingSeriesParticipants(ctx context.Context, bookingSeriesParticipants []BookingSeriesParticipant) ([]BookingSeriesParticipant, error)

	UpdateBookingSeriesCore(ctx context.Context, seriesId int, serviceId int, bookingType types.BookingType, rrule string, dstart time.Time) error
	UpdateBookingSeriesGeneratedUntil(ctx context.Context, seriesId int, generatedUntil time.Time) error
	DeactivateBookingSeries(ctx context.Context, seriesId int) error
	UpdateBookingSeriesDetails(ctx context.Context, seriesId int, details BookingDetails) error

	DeleteBookingSeriesParticipants(ctx context.Context, seriesId int, customerIds []uuid.UUID) error

	GetBookingSeries(ctx context.Context, seriesId int) (BookingSeries, error)
	GetActiveBookingSeriesIds(ctx context.Context, tresholdTime time.Time) ([]int, error)
	GetFutureSeriesBookings(ctx context.Context, seriesId int, fromDate time.Time) ([]Booking, error)
	GetBookingSeriesParticipants(ctx context.Context, seriesId int) ([]BookingSeriesParticipant, error)
}

type Booking struct {
	Id                    int                 `db:"id"`
	Status                types.BookingStatus `db:"status"`
	BookingType           types.BookingType   `db:"booking_type"`
	IsRecurring           bool                `db:"is_recurring"`
	MerchantId            uuid.UUID           `db:"merchant_id"`
	EmployeeId            *int                `db:"employee_id"`
	ServiceId             int                 `db:"service_id"`
	LocationId            int                 `db:"location_id"`
	BookingSeriesId       *int                `db:"booking_series_id"`
	SeriesOriginalDate    *time.Time          `db:"series_original_date"`
	FromDate              time.Time           `db:"from_date"`
	ToDate                time.Time           `db:"to_date"`
	PricePerPerson        currencyx.Price     `db:"price_per_person"`
	CostPerPerson         currencyx.Price     `db:"cost_per_person"`
	TotalPrice            currencyx.Price     `db:"total_price"`
	TotalCost             currencyx.Price     `db:"total_cost"`
	MerchantNote          *string             `db:"merchant_note"`
	MinParticipants       int                 `db:"min_participants"`
	MaxParticipants       int                 `db:"max_participants"`
	CurrentParticipants   int                 `db:"current_participants"`
	CancelledByMerchantOn *time.Time          `db:"cancelled_by_merchant_on"`
	CancellationReason    *string             `db:"cancellation_reason"`
}

// struct just for db inserts and updates
type BookingDetails struct {
	PricePerPerson      currencyx.Price
	CostPerPerson       currencyx.Price
	TotalPrice          currencyx.Price
	TotalCost           currencyx.Price
	MerchantNote        *string
	MinParticipants     int
	MaxParticipants     int
	CurrentParticipants int
}

type BookingPhase struct {
	Id             int
	BookingId      int
	ServicePhaseId int
	FromDate       time.Time
	ToDate         time.Time
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
	ID              int                 `json:"id" db:"id"`
	Status          types.BookingStatus `json:"status" db:"status"`
	FromDate        time.Time           `json:"from_date" db:"from_date"`
	ToDate          time.Time           `json:"to_date" db:"to_date"`
	CustomerNote    *string             `json:"customer_note" db:"customer_note"`
	MerchantNote    *string             `json:"merchant_note" db:"merchant_note"`
	ServiceName     string              `json:"service_name" db:"service_name"`
	ServiceColor    string              `json:"service_color" db:"service_color"`
	ServiceDuration int                 `json:"service_duration" db:"service_duration"`
	Price           currencyx.Price     `json:"price" db:"price"`
	Cost            currencyx.Price     `json:"cost" db:"cost"`
	FirstName       *string             `json:"first_name" db:"first_name"`
	LastName        *string             `json:"last_name" db:"last_name"`
	PhoneNumber     *string             `json:"phone_number" db:"phone_number"`
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
	Id                  int               `db:"id"`
	BookingType         types.BookingType `db:"booking_type"`
	MerchantId          uuid.UUID         `db:"merchant_id"`
	EmployeeId          int               `db:"employee_id"`
	ServiceId           int               `db:"service_id"`
	LocationId          int               `db:"location_id"`
	Rrule               string            `db:"rrule"`
	Dstart              time.Time         `db:"dstart"`
	Timezone            string            `db:"timezone"`
	IsActive            bool              `db:"is_active"`
	GeneratedUntil      *time.Time        `db:"generated_until"`
	PricePerPerson      currencyx.Price   `db:"price_per_person"`
	CostPerPerson       currencyx.Price   `db:"cost_per_person"`
	TotalPrice          currencyx.Price   `db:"total_price"`
	TotalCost           currencyx.Price   `db:"total_cost"`
	MinParticipants     int               `db:"min_participants"`
	MaxParticipants     int               `db:"max_participants"`
	CurrentParticipants int               `db:"current_participants"`
}

type BookingSeriesParticipant struct {
	Id              int        `db:"id"`
	BookingSeriesId int        `db:"booking_series_id"`
	CustomerId      *uuid.UUID `db:"customer_id"`
	IsActive        bool       `db:"is_active"`
	DroppedOutOn    *time.Time `db:"dropped_out_on"`
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
