package domain

import (
	"context"
	"time"

	"github.com/miketsu-inc/reservations/backend/internal/types"
	"github.com/miketsu-inc/reservations/backend/pkg/db"
)

type ExternalCalendarRepository interface {
	WithTx(tx db.DBTX) ExternalCalendarRepository

	NewExternalCalendar(ctx context.Context, extCalendar ExternalCalendar) (int, error)
	UpdateExternalCalendarSyncToken(ctx context.Context, extCalendarId int, syncToken string) error
	UpdateExternalCalendarAuthTokens(ctx context.Context, extCalendarId int, accessToken string, refreshToken string, tokenExpiry time.Time) error
	UpdateExternalCalendarChannel(ctx context.Context, calendarId int, channelId string, resourceId string, channelExpiry time.Time) error
	ResetExternalCalendarSyncState(ctx context.Context, extCalendarId int) error

	GetExternalCalendar(ctx context.Context, extCalendarId int) (ExternalCalendar, error)
	GetExternalCalendarByChannel(ctx context.Context, channelId string, resourceId string) (ExternalCalendar, error)
	GetExternalCalendarByEmployeeId(ctx context.Context, employeeId int) (ExternalCalendar, error)

	NewExternalCalendarEvent(ctx context.Context, externalEvent ExternalCalendarEvent) error
	BulkInsertExternalCalendarEvent(ctx context.Context, externalEvents []ExternalCalendarEvent) error
	UpdateExternalCalendarEvent(ctx context.Context, externalEvent ExternalCalendarEvent) error
	BulkUpdateExternalCalendarEvent(ctx context.Context, externalEvents []ExternalCalendarEvent) error
	DeleteExternalCalendarEvent(ctx context.Context, externalEventId int) error
	DeleteAllExternalCalendarEvents(ctx context.Context, extCalendarId int) error

	GetExternalCalendarEvents(ctx context.Context, extCalendarIds int, eventIds []string) ([]ExternalCalendarEvent, error)
	GetExternalCalendarEventByInternal(ctx context.Context, internalType types.EventInternalType, internalId int) (ExternalCalendarEvent, error)
	GetExpiringExternalCalendars(ctx context.Context, timeLeft time.Time) ([]ExternalCalendar, error)
}

type ExternalCalendar struct {
	Id            int        `json:"id" db:"id"`
	EmployeeId    int        `json:"employee_id" db:"employee_id"`
	CalendarId    string     `json:"calendar_id" db:"calendar_id"`
	AccessToken   string     `json:"access_token" db:"access_token"`
	RefreshToken  string     `json:"refresh_token" db:"refresh_token"`
	TokenExpiry   time.Time  `json:"token_expiry" db:"token_expiry"`
	SyncToken     *string    `json:"sync_token" db:"sync_token"`
	ChannelId     *string    `json:"channel_id" db:"channel_id"`
	ResourceId    *string    `json:"resource_id" db:"resource_id"`
	ChannelExpiry *time.Time `json:"channel_expiry" db:"channel_expiry"`
	Timezone      string     `json:"timezone" db:"timezone"`
}

type ExternalCalendarEvent struct {
	Id                 int                      `json:"id" db:"id"`
	ExternalCalendarId int                      `json:"external_calendar_id" db:"external_calendar_id"`
	ExternalEventId    string                   `json:"external_event_id" db:"external_event_id"`
	Etag               string                   `json:"etag" db:"etag"`
	Status             string                   `json:"status" db:"status"`
	Title              string                   `json:"title" db:"title"`
	Description        string                   `json:"description" db:"description"`
	FromDate           time.Time                `json:"from_date" db:"from_date"`
	ToDate             time.Time                `json:"to_date" db:"to_date"`
	IsAllDay           bool                     `json:"is_all_day" db:"is_all_day"`
	InternalId         *int                     `json:"internal_id" db:"internal_id"`
	InternalType       *types.EventInternalType `json:"internal_type" db:"internal_type"`
	IsBlocking         bool                     `json:"is_blocking" db:"is_blocking"`
	Source             types.EventSource        `json:"source" db:"source"`
	LastSyncedAt       time.Time                `json:"last_synced_at" db:"last_synced_at"`
}
