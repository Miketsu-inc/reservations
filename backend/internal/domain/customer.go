package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/miketsu-inc/reservations/backend/internal/types"
	"github.com/miketsu-inc/reservations/backend/pkg/currencyx"
	"github.com/miketsu-inc/reservations/backend/pkg/db"
)

type CustomerRepository interface {
	WithTx(tx db.DBTX) CustomerRepository

	NewCustomer(ctx context.Context, merchantId uuid.UUID, customer Customer) error
	NewCustomerFromUser(ctx context.Context, customerId, merchantId, userId uuid.UUID) (uuid.UUID, bool, bool, error)
	UpdateCustomer(ctx context.Context, merchantId uuid.UUID, customer Customer) error
	DeleteCustomer(ctx context.Context, customerId uuid.UUID, merchantId uuid.UUID) error

	GetCustomers(ctx context.Context, merchantId uuid.UUID, isBlacklisted bool) ([]PublicCustomer, error)
	GetCustomerInfo(ctx context.Context, merchantId uuid.UUID, customerId uuid.UUID) (CustomerInfo, error)
	GetCustomerStats(ctx context.Context, merchantId uuid.UUID, customerId uuid.UUID) (CustomerStatistics, error)
	GetCustomerBookings(ctx context.Context, merchantId uuid.UUID, customerId uuid.UUID, status string, limit int, cursorStart time.Time, cursorId int) ([]CustomerBooking, error)
	GetCustomersForCalendar(ctx context.Context, merchantId uuid.UUID) ([]CustomerForCalendar, error)

	SetBlacklistStatusForCustomer(ctx context.Context, merchantId uuid.UUID, customerId uuid.UUID, isBlacklisted bool, blacklistReason *string) error

	GetCustomerEmailById(ctx context.Context, merchantId uuid.UUID, customerId uuid.UUID) (*string, error)
}

type Customer struct {
	Id          uuid.UUID  `json:"id" db:"id"`
	FirstName   *string    `json:"first_name" db:"first_name"`
	LastName    *string    `json:"last_name" db:"last_name"`
	Email       *string    `json:"email" db:"email"`
	PhoneNumber *string    `json:"phone_number" db:"phone_number"`
	Birthday    *time.Time `json:"birthday" db:"birthday"`
	Note        *string    `json:"note" db:"note"`
}

type PublicCustomer struct {
	Customer
	IsDummy         bool    `json:"is_dummy" db:"is_dummy"`
	IsBlacklisted   bool    `json:"is_blacklisted" db:"is_blacklisted"`
	BlacklistReason *string `json:"blacklist_reason" db:"blacklist_reason"`
	TimesBooked     int     `json:"times_booked" db:"times_booked"`
	TimesCancelled  int     `json:"times_cancelled" db:"times_cancelled"`
}

type CustomerInfo struct {
	Customer
	IsDummy         bool    `json:"is_dummy" db:"is_dummy"`
	IsBlacklisted   bool    `json:"is_blacklisted" db:"is_blacklisted"`
	BlacklistReason *string `json:"blacklist_reason" db:"blacklist_reason"`
}

type CustomerStatistics struct {
	TimesBooked          int        `json:"times_booked" db:"times_booked"`
	TimesCancelledByUser int        `json:"times_cancelled_by_user" db:"times_cancelled_by_user"`
	TimesNoShow          int        `json:"times_no_show" db:"times_no_show"`
	TimesUpcoming        int        `json:"times_upcoming" db:"times_upcoming"`
	TimesCompleted       int        `json:"times_completed" db:"times_completed"`
	FirstBooking         *time.Time `json:"first_booking" db:"first_booking"`
	LastVisited          *time.Time `json:"last_visited" db:"last_visited"`
}

type CustomerBooking struct {
	Id                int                 `json:"id" db:"id"`
	BookingType       types.BookingType   `json:"booking_type" db:"booking_type"`
	IsRecurring       bool                `json:"is_recurring" db:"is_recurring"`
	FromDate          time.Time           `json:"from_date" db:"from_date"`
	ToDate            time.Time           `json:"to_date" db:"to_date"`
	ServiceName       string              `json:"service_name" db:"service_name"`
	ServiceColor      *string             `json:"service_color" db:"service_color"`
	FormattedLocation string              `json:"formatted_location" db:"formatted_location"`
	Price             currencyx.Price     `json:"price" db:"price"`
	PriceType         types.PriceType     `json:"price_type" db:"price_type"`
	Status            types.BookingStatus `json:"status" db:"status"`
	EmployeeFirstName *string             `json:"employee_first_name" db:"employee_first_name"`
	EmployeeLastName  *string             `json:"employee_last_name" db:"employee_last_name"`
}

type CustomerForCalendar struct {
	CustomerId  uuid.UUID  `json:"customer_id" db:"customer_id"`
	FirstName   string     `json:"first_name" db:"first_name"`
	LastName    string     `json:"last_name" db:"last_name"`
	Email       *string    `json:"email" db:"email"`
	PhoneNumber *string    `json:"phone_number" db:"phone_number"`
	BirthDay    *time.Time `json:"birthday" db:"birthday"`
	IsDummy     bool       `json:"is_dummy" db:"is_dummy"`
	LastVisited *time.Time `json:"last_visited" db:"last_visited"`
}
