package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/miketsu-inc/reservations/backend/internal/domain"
	"github.com/miketsu-inc/reservations/backend/internal/types"
	"github.com/miketsu-inc/reservations/backend/pkg/assert"
)

type externalCalendarRepository struct {
	db *pgxpool.Pool
}

func NewExternalCalendarRepository(db *pgxpool.Pool) domain.ExternalCalendarRepository {
	return &externalCalendarRepository{db: db}
}

func (r *externalCalendarRepository) NewExternalCalendar(ctx context.Context, ec domain.ExternalCalendar) (int, error) {
	query := `
	insert into "ExternalCalendar" (employee_id, calendar_id, access_token, refresh_token, token_expiry, channel_id, resource_id, channel_expiry, timezone)
	values ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	returning id
	`

	var extCalendarId int
	err := r.db.QueryRow(ctx, query, ec.EmployeeId, ec.CalendarId, ec.AccessToken, ec.RefreshToken, ec.TokenExpiry, ec.ChannelId,
		ec.ResourceId, ec.ChannelExpiry, ec.Timezone).Scan(&extCalendarId)
	if err != nil {
		return 0, err
	}

	return extCalendarId, nil
}

func (r *externalCalendarRepository) UpdateExternalCalendarSyncToken(ctx context.Context, extCalendarId int, syncToken string) error {
	query := `
	update "ExternalCalendar"
	set sync_token = $2
	where id = $1
	`

	_, err := r.db.Exec(ctx, query, extCalendarId, syncToken)
	if err != nil {
		return err
	}

	return nil
}

func bulkInsertBlockedTime(ctx context.Context, tx pgx.Tx, bt []domain.BlockedTime) ([]int, error) {
	query := `
	insert into "BlockedTime" (merchant_id, employee_id, name, from_date, to_date, all_day, source)
	select $1, $2, unnest($3::text[]), unnest($4::timestamptz[]), unnest($5::timestamptz[]), unnest($6::boolean[]), $7
	returning id
	`

	merchantId := bt[0].MerchantId
	employeeId := bt[0].EmployeeId
	source := bt[0].Source

	names := make([]string, len(bt))
	fromDates := make([]time.Time, len(bt))
	toDates := make([]time.Time, len(bt))
	isAllDay := make([]bool, len(bt))

	for i, blockedTime := range bt {
		names[i] = blockedTime.Name
		fromDates[i] = blockedTime.FromDate
		toDates[i] = blockedTime.ToDate
		isAllDay[i] = blockedTime.AllDay
	}

	var btIds []int

	rows, _ := tx.Query(ctx, query, merchantId, employeeId, names, fromDates, toDates, isAllDay, source)
	btIds, err := pgx.CollectRows(rows, pgx.RowTo[int])
	if err != nil {
		return []int{}, err
	}

	return btIds, nil
}

func bulkInsertExternalCalendarEvent(ctx context.Context, tx pgx.Tx, externalEvents []domain.ExternalCalendarEvent) error {
	query := `
	insert into "ExternalCalendarEvent" (external_calendar_id, external_event_id, etag, status, title, description, from_date, to_date, is_all_day,
		internal_id, internal_type, is_blocking, source, last_synced_at)
	select $1, unnest($2::text[]), unnest($3::text[]), unnest($4::text[]), unnest($5::text[]), unnest($6::text[]), unnest($7::timestamptz[]),
		unnest($8::timestamptz[]), unnest($9::boolean[]), unnest($10::int[]), unnest($11::event_internal_type[]), unnest($12::boolean[]), $13, $14
	`

	extEventIds := make([]string, len(externalEvents))
	etags := make([]string, len(externalEvents))
	statuses := make([]string, len(externalEvents))
	titles := make([]string, len(externalEvents))
	descriptions := make([]string, len(externalEvents))
	fromDates := make([]time.Time, len(externalEvents))
	toDates := make([]time.Time, len(externalEvents))
	isAllDays := make([]bool, len(externalEvents))
	InternalIds := make([]*int, len(externalEvents))
	InternalTypes := make([]*string, len(externalEvents))
	isBlockings := make([]bool, len(externalEvents))

	for i, ee := range externalEvents {
		extEventIds[i] = ee.ExternalEventId
		etags[i] = ee.Etag
		statuses[i] = ee.Status
		titles[i] = ee.Title
		descriptions[i] = ee.Description
		fromDates[i] = ee.FromDate
		toDates[i] = ee.ToDate
		isAllDays[i] = ee.IsAllDay
		InternalIds[i] = ee.InternalId
		if ee.InternalType != nil {
			str := ee.InternalType.String()
			InternalTypes[i] = &str
		} else {
			InternalTypes[i] = nil
		}
		isBlockings[i] = ee.IsBlocking
	}

	_, err := tx.Exec(ctx, query, externalEvents[0].ExternalCalendarId, extEventIds, etags, statuses, titles, descriptions, fromDates, toDates,
		isAllDays, InternalIds, InternalTypes, isBlockings, externalEvents[0].Source, time.Now().UTC())
	if err != nil {
		return err
	}

	return nil
}

func (r *externalCalendarRepository) BulkInitialSyncExternalCalendarEvents(ctx context.Context, bt []domain.BlockedTime, blockingEventIdxs []int, ece []domain.ExternalCalendarEvent) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	// nolint:errcheck
	defer tx.Rollback(ctx)

	var btIds []int

	if len(bt) > 0 {
		btIds, err = bulkInsertBlockedTime(ctx, tx, bt)
		if err != nil {
			return err
		}

		btPos := 0
		for _, idx := range blockingEventIdxs {
			ece[idx].InternalId = &btIds[btPos]
			btPos++
		}
	}

	if len(ece) > 0 {
		err = bulkInsertExternalCalendarEvent(ctx, tx, ece)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func bulkUpdateBlockedTime(ctx context.Context, tx pgx.Tx, bt []domain.BlockedTime) error {
	query := `
	update "BlockedTime" b
	set name = u.name, from_date = u.from_date, to_date = u.to_date, all_day = u.all_day
	from unnest($1::int[], $2::text[], $3::timestamptz[], $4::timestamptz[], $5::boolean[])
	as u(id, name, from_date, to_date, all_day)
	where b.id = u.id
	`

	ids := make([]int, len(bt))
	names := make([]string, len(bt))
	fromDates := make([]time.Time, len(bt))
	toDates := make([]time.Time, len(bt))
	isAllDay := make([]bool, len(bt))

	for i, blockedTime := range bt {
		ids[i] = blockedTime.Id
		names[i] = blockedTime.Name
		fromDates[i] = blockedTime.FromDate
		toDates[i] = blockedTime.ToDate
		isAllDay[i] = blockedTime.AllDay
	}

	_, err := tx.Exec(ctx, query, ids, names, fromDates, toDates, isAllDay)
	if err != nil {
		return err
	}

	return nil
}

func bulkDeleteBlockedTime(ctx context.Context, tx pgx.Tx, btIds []int) error {
	query := `
	delete from "BlockedTime" b
	from unnest($1::int[]) as d(id)
	where b.id = d.id
	`

	_, err := tx.Exec(ctx, query, btIds)
	if err != nil {
		return err
	}

	return nil
}

func bulkUpdateExternalCalendarEvent(ctx context.Context, tx pgx.Tx, ece []domain.ExternalCalendarEvent) error {
	query := `
	update "ExternalCalendarEvent" e
	set etag = u.etag, status = u.status, title = u.title, description = u.description, from_date = u.from_date, to_date = u.to_date,
		is_all_day = u.is_all_day, internal_id = u.internal_id, internal_type = u.internal_type, is_blocking = u.is_blocking, last_synced_at = $13
	from unnest($2::int[], $3::text[], $4::text[], $5::text[], $6::text[], $7::timestamptz[], $8::timestamptz[], $9::boolean[], $10::int[], $11::event_internal_type[], $12::boolean[])
	as u(id, etag, status, title, description, from_date, to_date, is_all_day, internal_id, internal_type, is_blocking)
	where external_calendar_id = $1 and e.id = u.id
	`

	ids := make([]int, len(ece))
	etags := make([]string, len(ece))
	statuses := make([]string, len(ece))
	titles := make([]string, len(ece))
	descriptions := make([]string, len(ece))
	fromDates := make([]time.Time, len(ece))
	toDates := make([]time.Time, len(ece))
	isAllDays := make([]bool, len(ece))
	InternalIds := make([]*int, len(ece))
	InternalTypes := make([]*string, len(ece))
	isBlockings := make([]bool, len(ece))

	for i, event := range ece {
		ids[i] = event.Id
		etags[i] = event.Etag
		statuses[i] = event.Status
		titles[i] = event.Title
		descriptions[i] = event.Description
		fromDates[i] = event.FromDate
		toDates[i] = event.ToDate
		isAllDays[i] = event.IsAllDay
		InternalIds[i] = event.InternalId
		if event.InternalType != nil {
			str := event.InternalType.String()
			InternalTypes[i] = &str
		} else {
			InternalTypes[i] = nil
		}
		isBlockings[i] = event.IsBlocking
	}

	_, err := tx.Exec(ctx, query, ece[0].ExternalCalendarId, ids, etags, statuses, titles, descriptions, fromDates, toDates, isAllDays, InternalIds, InternalTypes,
		isBlockings, time.Now().UTC())
	if err != nil {
		return err
	}

	return nil
}

func (r *externalCalendarRepository) BulkIncrementalSyncExternalCalendarEvents(ctx context.Context, newBlockedTimes []domain.BlockedTime,
	updateBlockedTimes []domain.BlockedTime, deleteBlockedTimes []int, blockingEventIdxs []int, newExtEvents []domain.ExternalCalendarEvent,
	updateExtEvents []domain.ExternalCalendarEvent, pendingBlockingLinks []domain.ExternalEventBlockedTimeLink) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	assert.True(len(blockingEventIdxs)+len(pendingBlockingLinks) == len(newBlockedTimes), "ExternalCalendarEvent and BlockedTime link mismatch!",
		len(blockingEventIdxs), len(pendingBlockingLinks), len(newBlockedTimes))

	// nolint:errcheck
	defer tx.Rollback(ctx)

	var btIds []int

	if len(newBlockedTimes) > 0 {
		btIds, err = bulkInsertBlockedTime(ctx, tx, newBlockedTimes)
		if err != nil {
			return err
		}

		btPos := 0
		for _, idx := range blockingEventIdxs {
			newExtEvents[idx].InternalId = &btIds[btPos]
			btPos++
		}

		for _, link := range pendingBlockingLinks {
			updateExtEvents[link.ExternalEventIdx].InternalId = &btIds[link.BlockedTimeIdx]
		}
	}

	if len(updateBlockedTimes) > 0 {
		err = bulkUpdateBlockedTime(ctx, tx, updateBlockedTimes)
		if err != nil {
			return err
		}
	}

	if len(deleteBlockedTimes) > 0 {
		err = bulkDeleteBlockedTime(ctx, tx, deleteBlockedTimes)
		if err != nil {
			return err
		}
	}

	if len(newExtEvents) > 0 {
		err = bulkInsertExternalCalendarEvent(ctx, tx, newExtEvents)
		if err != nil {
			return err
		}
	}

	if len(updateExtEvents) > 0 {
		err = bulkUpdateExternalCalendarEvent(ctx, tx, updateExtEvents)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *externalCalendarRepository) GetExternalCalendarEventsByIds(ctx context.Context, extCalendarId int, eventIds []string) ([]domain.ExternalCalendarEvent, error) {
	query := `
	select *
	from "ExternalCalendarEvent"
	where external_calendar_id = $1 and external_event_id = any($2)
	`

	rows, _ := r.db.Query(ctx, query, extCalendarId, eventIds)
	events, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.ExternalCalendarEvent])
	if err != nil {
		return []domain.ExternalCalendarEvent{}, err
	}

	return events, nil
}

func (r *externalCalendarRepository) ResetExternalCalendar(ctx context.Context, extCalendarId int) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	// nolint:errcheck
	defer tx.Rollback(ctx)

	blockedTimesQuery := `
	delete from "BlockedTime"
	where id in (
		select internal_id
		from "ExternalCalendarEvent"
		where external_calendar_id = $1 and internal_id is not null and internal_type = 'blocked_time'
	)
	`

	_, err = tx.Exec(ctx, blockedTimesQuery, extCalendarId)
	if err != nil {
		return err
	}

	externalEventsQuery := `
	delete from "ExternalCalendarEvent"
	where external_calendar_id = $1
	`

	_, err = tx.Exec(ctx, externalEventsQuery, extCalendarId)
	if err != nil {
		return err
	}

	externalCalendarSyncStateQuery := `
	update "ExternalCalendar"
	set sync_token = null, channel_id = null, resource_id = null, channel_expiry = null
	where id = $1
	`

	_, err = tx.Exec(ctx, externalEventsQuery, externalCalendarSyncStateQuery)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *externalCalendarRepository) GetExternalCalendarByEmployeeId(ctx context.Context, employeeId int) (domain.ExternalCalendar, error) {
	query := `
	select * from "ExternalCalendar"
	where employee_id = $1
	`

	rows, _ := r.db.Query(ctx, query, employeeId)
	extCalendar, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[domain.ExternalCalendar])
	if err != nil {
		return domain.ExternalCalendar{}, err
	}

	return extCalendar, nil
}

func (r *externalCalendarRepository) UpdateExternalCalendarAuthTokens(ctx context.Context, extCalendarId int, accessToken, refreshToken string, tokenExpiry time.Time) error {
	query := `
	update "ExternalCalendar"
	set access_token = $2, refresh_token = $3, token_expiry = $4
	where id = $1
	`

	_, err := r.db.Exec(ctx, query, extCalendarId, accessToken, refreshToken, tokenExpiry)
	if err != nil {
		return err
	}

	return nil
}

func (r *externalCalendarRepository) NewExternalCalendarEvent(ctx context.Context, event domain.ExternalCalendarEvent) error {
	query := `
	insert into "ExternalCalendarEvent" (external_calendar_id, external_event_id, etag, status, title, description, from_date, to_date,
		is_all_day, internal_id, internal_type, is_blocking, source, last_synced_at)
	values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`

	_, err := r.db.Exec(ctx, query, event.ExternalCalendarId, event.ExternalEventId, event.Etag, event.Status, event.Title, event.Description,
		event.FromDate, event.ToDate, event.IsAllDay, event.InternalId, event.InternalType, event.IsBlocking, event.Source, time.Now().UTC())
	if err != nil {
		return err
	}

	return nil
}

func (r *externalCalendarRepository) UpdateExternalCalendarEvent(ctx context.Context, event domain.ExternalCalendarEvent) error {
	query := `
	update "ExternalCalendarEvent"
	set etag = $2, status = $3, title = $4, description = $5, from_date = $6, to_date = $7, is_all_day = $8, internal_id = $9,
		internal_type = $10, is_blocking = $11, last_synced_at = $12
	where id = $1
	`

	_, err := r.db.Exec(ctx, query, event.Id, event.Etag, event.Status, event.Title, event.Description, event.FromDate, event.ToDate, event.IsAllDay,
		event.InternalId, event.InternalType, event.IsBlocking, time.Now().UTC())
	if err != nil {
		return err
	}

	return nil
}

func (r *externalCalendarRepository) DeleteExternalCalendarEvent(ctx context.Context, extEventId int) error {
	query := `
	delete from "ExternalCalendarEvent"
	where id = $1
	`

	_, err := r.db.Exec(ctx, query, extEventId)
	if err != nil {
		return err
	}

	return nil
}

func (r *externalCalendarRepository) GetExternalCalendarEventByInternal(ctx context.Context, internalType types.EventInternalType, internalId int) (domain.ExternalCalendarEvent, error) {
	query := `
	select *
	from "ExternalCalendarEvent"
	where internal_type = $1 and internal_id = $2
	`

	rows, _ := r.db.Query(ctx, query, internalType, internalId)
	event, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[domain.ExternalCalendarEvent])
	if err != nil {
		return domain.ExternalCalendarEvent{}, err
	}

	return event, err
}

func (r *externalCalendarRepository) GetExternalCalendarById(ctx context.Context, calendarId int) (domain.ExternalCalendar, error) {
	query := `
	select *
	from "ExternalCalendar"
	where id = $1
	`

	rows, _ := r.db.Query(ctx, query, calendarId)
	calendar, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[domain.ExternalCalendar])
	if err != nil {
		return domain.ExternalCalendar{}, err
	}

	return calendar, nil
}

func (r *externalCalendarRepository) GetExternalCalendarByChannel(ctx context.Context, channelId string, resourceId string) (domain.ExternalCalendar, error) {
	query := `
	select *
	from "ExternalCalendar"
	where channel_id = $1 and resource_id = $2
	`

	rows, _ := r.db.Query(ctx, query, channelId, resourceId)
	extCalendar, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[domain.ExternalCalendar])
	if err != nil {
		return domain.ExternalCalendar{}, err
	}

	return extCalendar, nil
}

func (r *externalCalendarRepository) GetExpiringExternalCalendars(ctx context.Context) ([]domain.ExternalCalendar, error) {
	query := `
	select *
	from "ExternalCalendar"
	where channel_expiry < $1
	`

	rows, _ := r.db.Query(ctx, query, time.Now().UTC().Add(time.Hour*24))
	extCalendars, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.ExternalCalendar])
	if err != nil {
		return []domain.ExternalCalendar{}, err
	}

	return extCalendars, nil
}

func (r *externalCalendarRepository) UpdateExternalCalendarChannel(ctx context.Context, calendarId int, channelId string, resourceId string, channelExpiry time.Time) error {
	query := `
	update "ExternalCalendar"
	set channel_id = $2, resource_id = $3, channel_expiry = $4
	where id = $1
	`

	_, err := r.db.Exec(ctx, query, calendarId, channelId, resourceId, channelExpiry)
	if err != nil {
		return err
	}

	return nil
}
