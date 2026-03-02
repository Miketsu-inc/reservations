package domain

import (
	"context"

	"github.com/google/uuid"
	"github.com/miketsu-inc/reservations/backend/internal/types"
	"github.com/miketsu-inc/reservations/backend/pkg/db"
)

type TeamRepository interface {
	WithTx(tx db.DBTX) TeamRepository

	NewEmployee(ctx context.Context, merchantId uuid.UUID, employee PublicEmployee) error
	UpdateEmployee(ctx context.Context, merchantId uuid.UUID, employee PublicEmployee) error
	DeleteEmployee(ctx context.Context, merchantId uuid.UUID, employeeId int) error
	GetEmployee(ctx context.Context, merchantId uuid.UUID, employeeId int) (PublicEmployee, error)

	GetEmployees(ctx context.Context, merchantId uuid.UUID) ([]PublicEmployee, error)
	GetEmployeesForCalendar(ctx context.Context, merchantId uuid.UUID) ([]EmployeeForCalendar, error)

	GetMerchantIdByEmployee(ctx context.Context, employeeId int) (uuid.UUID, error)
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

type EmployeeForCalendar struct {
	Id        int    `json:"id" db:"id"`
	FirstName string `json:"first_name" db:"first_name"`
	LastName  string `json:"last_name" db:"last_name"`
}
