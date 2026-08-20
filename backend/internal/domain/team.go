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

	NewEmployee(ctx context.Context, merchantId uuid.UUID, employee PublicEmployee) (int, error)
	UpdateEmployee(ctx context.Context, merchantId uuid.UUID, employee PublicEmployee) error
	DeleteEmployee(ctx context.Context, merchantId uuid.UUID, employeeId int) error
	GetEmployee(ctx context.Context, merchantId uuid.UUID, employeeId int) (PublicEmployee, error)

	GetEmployees(ctx context.Context, merchantId uuid.UUID) ([]PublicEmployee, error)

	GetActiveEmployees(ctx context.Context, merchantId uuid.UUID) ([]PublicEmployee, error)

	GetMerchantIdByEmployee(ctx context.Context, employeeId int) (uuid.UUID, error)

	NewEmployeePreferences(ctx context.Context, employeeId int) error
	UpdateEmployeePreferences(ctx context.Context, employeeId int, preferences EmployeePreferences) error
	GetEmployeePreferences(ctx context.Context, employeeId int) (EmployeePreferences, error)
}

type PublicEmployee struct {
	Id          int                `json:"id" db:"id"`
	UserId      *uuid.UUID         `db:"user_id"`
	Role        types.EmployeeRole `json:"role" db:"role"`
	FirstName   *string            `json:"first_name" db:"first_name"`
	LastName    *string            `json:"last_name" db:"last_name"`
	Email       *string            `json:"email" db:"email"`
	PhoneNumber *string            `json:"phone_number" db:"phone_number"`
	IsActive    bool               `json:"is_active" db:"is_active"`
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
