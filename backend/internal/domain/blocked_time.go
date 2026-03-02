package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/miketsu-inc/reservations/backend/internal/types"
	"github.com/miketsu-inc/reservations/backend/pkg/db"
)

type BlockedTimeRepository interface {
	WithTx(db.DBTX) BlockedTimeRepository

	NewBlockedTime(ctx context.Context, merchantId uuid.UUID, employeeIds []int, name string, fromDate time.Time, toDate time.Time, allDay bool, blockedTypeId *int) ([]int, error)
	BulkInsertBlockedTime(ctx context.Context, blockedTimes []BlockedTime) ([]int, error)
	UpdateBlockedTime(ctx context.Context, blockedTime BlockedTime) error
	BulkUpdateBlockedTime(ctx context.Context, blockedTime []BlockedTime) error
	DeleteBlockedTime(ctx context.Context, blockedTimeId int, merchantId uuid.UUID, employeeId int) error
	BulkDeleteBlockedTime(ctx context.Context, blockedTimeIds []int) error
	DeleteExternalCalendarBlockedTimes(ctx context.Context, extCalendarId int) error

	GetBlockedTime(ctx context.Context, blockedTimeId int) (BlockedTime, error)
	GetBlockedTimesForCalendar(ctx context.Context, merchantId uuid.UUID, startTime string, endTime string) ([]BlockedTimeEvent, error)
	GetBlockedTimes(ctx context.Context, merchantId uuid.UUID, start time.Time, end time.Time) ([]BlockedTimes, error)

	NewBlockedTimeType(ctx context.Context, merchantId uuid.UUID, blockedTimeType BlockedTimeType) error
	UpdateBlockedTimeType(ctx context.Context, merchantId uuid.UUID, blockedTimeType BlockedTimeType) error
	DeleteBlockedTimeType(ctx context.Context, merchantId uuid.UUID, blockedTimeId int) error

	GetAllBlockedTimeTypes(ctx context.Context, merchantId uuid.UUID) ([]BlockedTimeType, error)
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

type BlockedTimeEvent struct {
	ID            int       `json:"id" db:"id"`
	EmployeeId    int       `json:"employee_id" db:"employee_id"`
	Name          string    `json:"name" db:"name"`
	FromDate      time.Time `json:"from_date" db:"from_date"`
	ToDate        time.Time `json:"to_date" db:"to_date"`
	AllDay        bool      `json:"all_day" db:"all_day"`
	Icon          *string   `json:"icon" db:"icon"`
	BlockedTypeId *int      `json:"blocked_type_id" db:"blocked_type_id"`
}
