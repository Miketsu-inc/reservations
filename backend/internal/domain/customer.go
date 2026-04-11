package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
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
	GetCustomersForCalendar(ctx context.Context, merchantId uuid.UUID) ([]CustomerForCalendar, error)

	SetBlacklistStatusForCustomer(ctx context.Context, merchantId uuid.UUID, customerId uuid.UUID, isBlacklisted bool, blacklistReason *string) error

	GetCustomerIdByUserIdAndMerchantId(ctx context.Context, merchantId uuid.UUID, userId uuid.UUID) (uuid.UUID, error)
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
	IsDummy bool `json:"is_dummy"`
}

type CustomerStatistics struct {
	Customer
	IsDummy              bool            `json:"is_dummy"`
	IsBlacklisted        bool            `json:"is_blacklisted"`
	BlacklistReason      *string         `json:"blacklist_reason"`
	TimesBooked          int             `json:"times_booked"`
	TimesCancelledByUser int             `json:"times_cancelled_by_user"`
	TimesUpcoming        int             `json:"times_upcoming"`
	Bookings             []PublicBooking `json:"bookings"`
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
