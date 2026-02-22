package domain

import (
	"context"

	"github.com/google/uuid"
	"github.com/miketsu-inc/reservations/backend/internal/types"
)

type TeamRepository interface {
	NewEmployee(context.Context, uuid.UUID, PublicEmployee) error
	UpdateEmployeeById(context.Context, uuid.UUID, PublicEmployee) error
	DeleteEmployeeById(context.Context, uuid.UUID, int) error
	GetEmployeeById(context.Context, uuid.UUID, int) (PublicEmployee, error)

	GetEmployeesForCalendarByMerchant(context.Context, uuid.UUID) ([]EmployeeForCalendar, error)
	GetEmployeesByMerchant(context.Context, uuid.UUID) ([]PublicEmployee, error)

	GetMerchantIdByEmployee(context.Context, int) (uuid.UUID, error)
}

type PublicEmployee struct {
	Id          int                `json:"id" db:"id"`
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
