package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type CustomerRepository interface {
	NewCustomer(context.Context, uuid.UUID, Customer) error
	UpdateCustomerById(context.Context, uuid.UUID, Customer) error
	DeleteCustomerById(context.Context, uuid.UUID, uuid.UUID) error

	GetCustomersByMerchantId(context.Context, uuid.UUID, bool) ([]PublicCustomer, error)
	GetCustomerInfoByMerchant(context.Context, uuid.UUID, uuid.UUID) (CustomerInfo, error)
	GetCustomerStatsByMerchant(context.Context, uuid.UUID, uuid.UUID) (CustomerStatistics, error)
	GetCustomersForCalendarByMerchant(context.Context, uuid.UUID) ([]CustomerForCalendar, error)

	SetBlacklistStatusForCustomer(context.Context, uuid.UUID, uuid.UUID, bool, *string) error

	GetCustomerIdByUserIdAndMerchantId(context.Context, uuid.UUID, uuid.UUID) (uuid.UUID, error)
	GetCustomerEmailById(context.Context, uuid.UUID, uuid.UUID) (*string, error)
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
	Id          uuid.UUID  `json:"id" db:"id"`
	FirstName   string     `json:"first_name" db:"first_name"`
	LastName    string     `json:"last_name" db:"last_name"`
	Email       *string    `json:"email" db:"email"`
	PhoneNumber *string    `json:"phone_number" db:"phone_number"`
	BirthDay    *time.Time `json:"birthday" db:"birthday"`
	IsDummy     bool       `json:"is_dummy" db:"is_dummy"`
	LastVisited *time.Time `json:"last_visited" db:"last_visited"`
}
