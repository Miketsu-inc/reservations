package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/miketsu-inc/reservations/backend/internal/types"
)

type BlockedTimeRepository interface {
	NewBlockedTime(context.Context, uuid.UUID, []int, string, time.Time, time.Time, bool, *int) ([]int, error)
	UpdateBlockedTime(context.Context, BlockedTime) error
	DeleteBlockedTime(context.Context, int, uuid.UUID, int) error

	GetBlockedTimeById(context.Context, int) (BlockedTime, error)
	GetBlockedTimes(context.Context, uuid.UUID, time.Time, time.Time) ([]BlockedTimes, error)

	NewBlockedTimeType(context.Context, uuid.UUID, BlockedTimeType) error
	UpdateBlockedTimeType(context.Context, uuid.UUID, BlockedTimeType) error
	DeleteBlockedTimeType(context.Context, uuid.UUID, int) error

	GetAllBlockedTimeTypes(context.Context, uuid.UUID) ([]BlockedTimeType, error)
}

type BlockedTime struct {
	Id            int                `json:"id" db:"id"`
	MerchantId    uuid.UUID          `json:"merchant_id" db:"merchant_id"`
	EmployeeId    int                `json:"employee_id" db:"employee_id"`
	BlockedTypeId *int               `json:"blocked_type_id" db:"blocked_type_id"`
	Name          string             `json:"name" db:"name"`
	FromDate      time.Time          `json:"from_date" db:"from_date"`
	ToDate        time.Time          `json:"to_date" db:"to_date"`
	AllDay        bool               `json:"all_day" db:"all_day"`
	Source        *types.EventSource `json:"source" db:"source"`
}

type BlockedTimes struct {
	FromDate time.Time `db:"from_date"`
	ToDate   time.Time `db:"to_date"`
	AllDay   bool      `db:"all_day"`
}

type BlockedTimeType struct {
	Id       int    `json:"id" db:"id"`
	Name     string `json:"name" db:"name"`
	Duration int    `json:"duration" db:"duration"`
	Icon     string `json:"icon" db:"icon"`
}
