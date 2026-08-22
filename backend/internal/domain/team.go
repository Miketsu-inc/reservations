package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/miketsu-inc/reservations/backend/internal/types"
	"github.com/miketsu-inc/reservations/backend/pkg/db"
)

type TeamRepository interface {
	WithTx(tx db.DBTX) TeamRepository

	NewEmployee(ctx context.Context, employee Employee) (int, error)
	UpdateEmployee(ctx context.Context, employee Employee) error
	DeleteEmployee(ctx context.Context, merchantId uuid.UUID, employeeId int) error
	GetEmployee(ctx context.Context, merchantId uuid.UUID, employeeId int) (Employee, error)

	GetEmployees(ctx context.Context, merchantId uuid.UUID) ([]Employee, error)
	GetActiveEmployees(ctx context.Context, merchantId uuid.UUID) ([]Employee, error)

	GetMerchantIdByEmployee(ctx context.Context, employeeId int) (uuid.UUID, error)

	NewEmployeeInvitation(ctx context.Context, inv EmployeeInvitation) (int, error)
	UpdateEmployeeInvitation(ctx context.Context, inv EmployeeInvitation) error
	UpdateEmployeeInvitationsStatus(ctx context.Context, invitationIds []int, status types.EmployeeInvitationStatus) error
	AcceptEmployeeInvitation(ctx context.Context, invitationId int) error
	DeclineEmployeeInvitation(ctx context.Context, invitationId int) error
	RevokeEmployeeInvitation(ctx context.Context, invitationId int) error

	GetEmployeeInvitation(ctx context.Context, invitationId int) (EmployeeInvitation, error)
	GetEmployeeInvitationByEmail(ctx context.Context, merchantId uuid.UUID, email string) (EmployeeInvitation, error)
	GetEmployeeInvitationByToken(ctx context.Context, token string) (EmployeeInvitation, error)
	GetEmployeeInvitations(ctx context.Context, merchantId uuid.UUID) ([]EmployeeInvitation, error)
	// Get invitations that are expired but their status in not yet in 'expired'
	GetExpiredEmployeeInvitations(ctx context.Context, expiry time.Time, limit int) ([]EmployeeInvitation, error)

	NewEmployeePreferences(ctx context.Context, employeeId int) error
	UpdateEmployeePreferences(ctx context.Context, employeeId int, preferences EmployeePreferences) error
	GetEmployeePreferences(ctx context.Context, employeeId int) (EmployeePreferences, error)
}

type Employee struct {
	Id          int                `db:"id"`
	UserId      *uuid.UUID         `db:"user_id"`
	MerchantId  uuid.UUID          `db:"merchant_id"`
	Role        types.EmployeeRole `db:"role"`
	FirstName   *string            `db:"first_name"`
	LastName    *string            `db:"last_name"`
	Email       *string            `db:"email"`
	PhoneNumber *string            `db:"phone_number"`
	IsActive    bool               `db:"is_active"`
}

type EmployeeInvitation struct {
	Id         int                            `db:"id"`
	Status     types.EmployeeInvitationStatus `db:"status"`
	MerchantId uuid.UUID                      `db:"merchant_id"`
	Email      string                         `db:"email"`
	Role       types.EmployeeInvitationRole   `db:"role"`
	Token      string                         `db:"token"`
	InvitedBy  *int                           `db:"invited_by"`
	InvitedAt  time.Time                      `db:"invited_at"`
	ExpiresAt  time.Time                      `db:"expires_at"`
	AcceptedAt *time.Time                     `db:"accepted_at"`
	RevokedAt  *time.Time                     `db:"revoked_at"`
	DeclinedAt *time.Time                     `db:"declined_at"`
}

func (e EmployeeInvitation) IsExpired() bool {
	return e.ExpiresAt.Before(time.Now().UTC())
}

func (e EmployeeInvitation) IsPending() bool {
	return e.Status == types.EmployeeInvitationStatusPending && !e.IsExpired()
}

type EmployeePreferences struct {
	EmployeeId         int       `db:"employee_id"`
	FirstDayOfWeek     string    `db:"first_day_of_week"`
	TimeFormat         string    `db:"time_format"`
	CalendarView       string    `db:"calendar_view"`
	CalendarViewMobile string    `db:"calendar_view_mobile"`
	StartHour          time.Time `db:"start_hour"`
	EndHour            time.Time `db:"end_hour"`
	TimeFrequency      time.Time `db:"time_frequency"`
}
