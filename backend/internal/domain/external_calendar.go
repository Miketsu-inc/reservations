package domain

import (
	"context"
	"time"

	"github.com/miketsu-inc/reservations/backend/internal/types"
)

type ExternalCalendarRepository interface {
	NewExternalCalendar(context.Context, ExternalCalendar) (int, error)
	UpdateExternalCalendarSyncToken(context.Context, int, string) error
	UpdateExternalCalendarAuthTokens(context.Context, int, string, string, time.Time) error
	UpdateExternalCalendarChannel(context.Context, int, string, string, time.Time) error
	// Delete all external calendar related data (BlockedTime, ExternalCalendarEvent) and reset sync state
	// should be called for 410 GONE response before full initial sync
	ResetExternalCalendar(context.Context, int) error

	GetExternalCalendarById(context.Context, int) (ExternalCalendar, error)
	GetExternalCalendarByChannel(context.Context, string, string) (ExternalCalendar, error)
	GetExternalCalendarByEmployeeId(context.Context, int) (ExternalCalendar, error)

	BulkInitialSyncExternalCalendarEvents(context.Context, []BlockedTime, []int, []ExternalCalendarEvent) error
	BulkIncrementalSyncExternalCalendarEvents(context.Context, []BlockedTime, []BlockedTime, []int, []int, []ExternalCalendarEvent,
		[]ExternalCalendarEvent, []ExternalEventBlockedTimeLink) error

	GetExternalCalendarEventsByIds(context.Context, int, []string) ([]ExternalCalendarEvent, error)

	NewExternalCalendarEvent(context.Context, ExternalCalendarEvent) error
	UpdateExternalCalendarEvent(context.Context, ExternalCalendarEvent) error
	DeleteExternalCalendarEvent(context.Context, int) error

	GetExternalCalendarEventByInternal(context.Context, types.EventInternalType, int) (ExternalCalendarEvent, error)
	// Get external calendars that have a channel expiry of less than 24 hours
	GetExpiringExternalCalendars(context.Context) ([]ExternalCalendar, error)
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

type ExternalEventBlockedTimeLink struct {
	ExternalEventIdx int
	BlockedTimeIdx   int
}
