package args

import (
	"time"

	"github.com/riverqueue/river"
)

type IncrementalCalendarSync struct {
	ExternalCalendarId int `json:"external_calendar_id"`
}

func (IncrementalCalendarSync) Kind() string { return "incremental_calendar_sync" }

type SyncNewBooking struct {
	BookingId int `json:"booking_id"`
}

func (SyncNewBooking) Kind() string { return "sync_new_booking" }

type SyncUpdateBooking struct {
	BookingId int `json:"booking_id"`
}

func (SyncUpdateBooking) Kind() string { return "sync_update_booking" }

type SyncDeleteBooking struct {
	BookingId int `json:"booking_id"`
}

func (SyncDeleteBooking) Kind() string { return "sync_delete_booking" }

type SyncNewBlockedTimeDispatcher struct {
	BlockedTimeId int `json:"blocked_time_id"`
}

func (SyncNewBlockedTimeDispatcher) Kind() string { return "sync_new_blocked_time_dispatcher" }

type SyncNewBlockedTime struct {
	BlockedTimeId int `json:"blocked_time_id"`
	EmployeeId    int `json:"employee_id"`
}

func (SyncNewBlockedTime) Kind() string { return "sync_new_blocked_time" }

type SyncUpdateBlockedTimeDispatcher struct {
	BlockedTimeId int `json:"blocked_time_id"`
}

func (SyncUpdateBlockedTimeDispatcher) Kind() string { return "sync_update_blocked_time_dispatcher" }

type SyncUpdateBlockedTime struct {
	BlockedTimeId           int `json:"blocked_time_id"`
	ExternalCalendarEventId int `json:"external_calendar_event_id"`
}

func (SyncUpdateBlockedTime) Kind() string { return "sync_update_blocked_time" }

type SyncDeleteBlockedTimeDispatcher struct {
	BlockedTimeId int `json:"blocked_time_id"`
}

func (SyncDeleteBlockedTimeDispatcher) Kind() string { return "sync_delete_blocked_time_dispatcher" }

type SyncDeleteBlockedTime struct {
	BlockedTimeId           int `json:"blocked_time_id"`
	ExternalCalendarEventId int `json:"external_calendar_event_id"`
}

func (SyncDeleteBlockedTime) Kind() string { return "sync_delete_blocked_time" }

type HandleChannelExpiration struct{}

func (HandleChannelExpiration) Kind() string { return "handle_channel_expiration" }

func (HandleChannelExpiration) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		UniqueOpts: river.UniqueOpts{
			ByPeriod: time.Hour * 24,
		},
	}
}
