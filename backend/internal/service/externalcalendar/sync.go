package externalcalendar

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/miketsu-inc/reservations/backend/internal/domain"
	"github.com/miketsu-inc/reservations/backend/internal/jobs/args"
	"github.com/miketsu-inc/reservations/backend/internal/types"
	"github.com/miketsu-inc/reservations/backend/internal/utils"
	"github.com/miketsu-inc/reservations/backend/pkg/assert"
	"github.com/riverqueue/river"
	"golang.org/x/oauth2"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
)

func eventToBlockedTime(event *calendar.Event, merchantId uuid.UUID, calendarTz *time.Location) (domain.BlockedTime, error) {
	fromDate, toDate, isAllDay, err := parseEventDates(event, calendarTz)
	if err != nil {
		return domain.BlockedTime{}, err
	}

	return domain.BlockedTime{
		MerchantId:    merchantId,
		BlockedTypeId: nil,
		Name:          event.Summary,
		FromDate:      fromDate,
		ToDate:        toDate,
		AllDay:        isAllDay,
		Source:        &types.EventSourceGoogle,
	}, nil
}

func eventToExternalCalendarEvent(event *calendar.Event, extCalendarId int, calendarTz *time.Location) (domain.ExternalCalendarEvent, error) {
	fromDate, toDate, isAllDay, err := parseEventDates(event, calendarTz)
	if err != nil {
		return domain.ExternalCalendarEvent{}, err
	}

	if event.ExtendedProperties != nil {
		_, ok := event.ExtendedProperties.Private["internal_type"]
		if ok {
			assert.Never("Internal source events should not end up here!", event, extCalendarId)
		}
		_, ok = event.ExtendedProperties.Private["internal_id"]
		if ok {
			assert.Never("Internal source events should not end up here!", event, extCalendarId)
		}
	}

	return domain.ExternalCalendarEvent{
		ExternalCalendarId: extCalendarId,
		ExternalEventId:    event.Id,
		Etag:               event.Etag,
		Status:             event.Status,
		Title:              event.Summary,
		Description:        event.Description,
		FromDate:           fromDate,
		ToDate:             toDate,
		IsAllDay:           isAllDay,
		InternalId:         nil,
		InternalType:       nil,
		IsBlocking:         false,
		Source:             types.EventSourceGoogle,
	}, nil
}

func parseEventDates(event *calendar.Event, calendarTz *time.Location) (time.Time, time.Time, bool, error) {
	var fromDate time.Time
	var toDate time.Time
	var isAllDay bool

	if event.Start.Date != "" {
		isAllDay = true

		startLocal, err := time.ParseInLocation("2006-01-02", event.Start.Date, calendarTz)
		if err != nil {
			return time.Time{}, time.Time{}, false, err
		}

		endLocal, err := time.ParseInLocation("2006-01-02", event.End.Date, calendarTz)
		if err != nil {
			return time.Time{}, time.Time{}, false, err
		}

		fromDate = startLocal.UTC()
		toDate = endLocal.UTC()
	} else {
		isAllDay = false

		startLocal, err := time.Parse(time.RFC3339, event.Start.DateTime)
		if err != nil {
			return time.Time{}, time.Time{}, false, err
		}

		endLocal, err := time.Parse(time.RFC3339, event.End.DateTime)
		if err != nil {
			return time.Time{}, time.Time{}, false, err
		}

		fromDate = startLocal.UTC()
		toDate = endLocal.UTC()
	}

	return fromDate, toDate, isAllDay, nil
}

func (s *Service) initialCalendarSyncToDB(ctx context.Context, employeeId int, blockedTimes []domain.BlockedTime, blockingIdxs []int, externalEvents []domain.ExternalCalendarEvent) error {
	return s.txManager.WithTransaction(ctx, func(tx pgx.Tx) error {
		var btIds []int
		var err error

		if len(blockedTimes) > 0 {
			btIds, err = s.blockedTimeRepo.WithTx(tx).BulkInsertBlockedTime(ctx, blockedTimes)
			if err != nil {
				return err
			}

			err := s.blockedTimeRepo.WithTx(tx).BulkInsertEmployeeBlockedTime(ctx, btIds, utils.RepeatEach([]int{employeeId}, len(btIds)))
			if err != nil {
				return err
			}

			btPos := 0
			for _, idx := range blockingIdxs {
				externalEvents[idx].InternalId = &btIds[btPos]
				btPos++
			}
		}

		if len(externalEvents) > 0 {
			err = s.externalCalendarRepo.WithTx(tx).BulkInsertExternalCalendarEvent(ctx, externalEvents)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

func (s *Service) initialCalendarSync(ctx context.Context, service *calendar.Service, extCalendar domain.ExternalCalendar,
	calendarTz *time.Location, merchantId uuid.UUID) error {
	req := service.Events.List("primary").ShowDeleted(false).SingleEvents(true).TimeMin(time.Now().UTC().Format(time.RFC3339))

	const batchSize = 200

	var syncToken string
	var blockedTimes []domain.BlockedTime
	var blockingEventsIdxs []int
	var externalEvents []domain.ExternalCalendarEvent

	for {
		events, err := req.Do()
		if err != nil {
			return err
		}

		for _, ev := range events.Items {
			// skip birthday events as they are all-day non-blocking and unnecessary
			if ev.EventType == "birthday" {
				continue
			}

			isBlocking := ev.Transparency == "" || ev.Transparency == "opaque"

			ece, err := eventToExternalCalendarEvent(ev, extCalendar.Id, calendarTz)
			if err != nil {
				return err
			}

			// apparently 0 duration google events are valid so skip them
			if !ece.FromDate.Before(ece.ToDate) {
				continue
			}

			ece.IsBlocking = isBlocking

			if isBlocking {
				bt, err := eventToBlockedTime(ev, merchantId, calendarTz)
				if err != nil {
					return err
				}

				blockingEventsIdxs = append(blockingEventsIdxs, len(externalEvents))
				blockedTimes = append(blockedTimes, bt)
			}

			externalEvents = append(externalEvents, ece)

			if len(externalEvents) >= batchSize {
				err := s.initialCalendarSyncToDB(ctx, extCalendar.EmployeeId, blockedTimes, blockingEventsIdxs, externalEvents)
				if err != nil {
					return err
				}

				blockedTimes = blockedTimes[:0]
				externalEvents = externalEvents[:0]
				blockingEventsIdxs = blockingEventsIdxs[:0]
			}
		}

		if events.NextPageToken == "" {
			syncToken = events.NextSyncToken
			break
		}

		req.PageToken(events.NextPageToken)
	}

	err := s.initialCalendarSyncToDB(ctx, extCalendar.EmployeeId, blockedTimes, blockingEventsIdxs, externalEvents)
	if err != nil {
		return err
	}

	err = s.externalCalendarRepo.UpdateExternalCalendarSyncToken(ctx, extCalendar.Id, syncToken)
	if err != nil {
		return err
	}

	channelId := uuid.NewString()
	googleChannel, err := service.Events.Watch(extCalendar.CalendarId, &calendar.Channel{
		Id:      channelId,
		Type:    "web_hook",
		Address: "http://localhost:8080/api/v1/integrations/calendar/google/watch",
	}).Do()
	if err != nil {
		return err
	}

	channelExpiry := time.UnixMilli(googleChannel.Expiration)

	return s.externalCalendarRepo.UpdateExternalCalendarChannel(ctx, extCalendar.Id, googleChannel.Id, googleChannel.ResourceId, channelExpiry)
}

type externalEventBlockedTimeLink struct {
	ExternalEventIdx int
	BlockedTimeIdx   int
}

func (s *Service) incrementalCalendarSyncToDB(ctx context.Context, employeeId int, newBlockedTimes []domain.BlockedTime,
	updateBlockedTimes []domain.BlockedTime, deleteBlockedTimes []int, blockingIdxs []int, newExtEvents []domain.ExternalCalendarEvent,
	updateExtEvents []domain.ExternalCalendarEvent, pendingBlockingLinks []externalEventBlockedTimeLink) error {

	assert.True(len(blockingIdxs)+len(pendingBlockingLinks) == len(newBlockedTimes), "ExternalCalendarEvent and BlockedTime link mismatch!",
		len(blockingIdxs), len(pendingBlockingLinks), len(newBlockedTimes))

	return s.txManager.WithTransaction(ctx, func(tx pgx.Tx) error {
		var btIds []int
		var err error

		if len(newBlockedTimes) > 0 {
			btIds, err = s.blockedTimeRepo.WithTx(tx).BulkInsertBlockedTime(ctx, newBlockedTimes)
			if err != nil {
				return err
			}

			err := s.blockedTimeRepo.WithTx(tx).BulkInsertEmployeeBlockedTime(ctx, btIds, utils.RepeatEach([]int{employeeId}, len(btIds)))
			if err != nil {
				return err
			}

			btPos := 0
			for _, idx := range blockingIdxs {
				newExtEvents[idx].InternalId = &btIds[btPos]
				btPos++
			}

			for _, link := range pendingBlockingLinks {
				updateExtEvents[link.ExternalEventIdx].InternalId = &btIds[link.BlockedTimeIdx]
			}
		}

		if len(updateBlockedTimes) > 0 {
			err = s.blockedTimeRepo.WithTx(tx).BulkUpdateBlockedTime(ctx, updateBlockedTimes)
			if err != nil {
				return err
			}
		}

		if len(deleteBlockedTimes) > 0 {
			err = s.blockedTimeRepo.WithTx(tx).BulkDeleteBlockedTime(ctx, deleteBlockedTimes)
			if err != nil {
				return err
			}
		}

		if len(newExtEvents) > 0 {
			err = s.externalCalendarRepo.WithTx(tx).BulkInsertExternalCalendarEvent(ctx, newExtEvents)
			if err != nil {
				return err
			}
		}

		if len(updateExtEvents) > 0 {
			err = s.externalCalendarRepo.WithTx(tx).BulkUpdateExternalCalendarEvent(ctx, updateExtEvents)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

// Delete all external calendar related data (BlockedTime, ExternalCalendarEvent) and reset sync state
// should be called for 410 GONE response before full initial sync
func (s *Service) resetExternalCalendar(ctx context.Context, extCalendarId int) error {
	return s.txManager.WithTransaction(ctx, func(tx pgx.Tx) error {
		err := s.blockedTimeRepo.WithTx(tx).DeleteExternalCalendarBlockedTimes(ctx, extCalendarId)
		if err != nil {
			return err
		}

		err = s.externalCalendarRepo.WithTx(tx).DeleteAllExternalCalendarEvents(ctx, extCalendarId)
		if err != nil {
			return err
		}

		err = s.externalCalendarRepo.WithTx(tx).ResetExternalCalendarSyncState(ctx, extCalendarId)
		if err != nil {
			return err
		}

		return nil
	})
}

func (s *Service) IncrementalCalendarSync(ctx context.Context, extCalendar domain.ExternalCalendar) error {
	merchantId, err := s.teamRepo.GetMerchantIdByEmployee(ctx, extCalendar.EmployeeId)
	if err != nil {
		return err
	}

	ts := googleCalendarConf.TokenSource(ctx, &oauth2.Token{
		AccessToken:  extCalendar.AccessToken,
		RefreshToken: extCalendar.RefreshToken,
		Expiry:       extCalendar.TokenExpiry,
	})

	service, err := calendar.NewService(ctx, option.WithTokenSource(ts))
	if err != nil {
		return err
	}

	req := service.Events.List("primary").SyncToken(*extCalendar.SyncToken).ShowDeleted(true)

	calendarTz, err := time.LoadLocation(extCalendar.Timezone)
	if err != nil {
		return err
	}

	var (
		nextSyncToken string

		newExternalEvents    []domain.ExternalCalendarEvent
		newBlockedTimes      []domain.BlockedTime
		newBlockingEventIdxs []int

		pendingBlockingLinks []externalEventBlockedTimeLink

		deleteBlockedTimes   []int
		updateBlockedTimes   []domain.BlockedTime
		updateExternalEvents []domain.ExternalCalendarEvent
	)

	for {
		events, err := req.Do()
		if err != nil {
			// TODO: handle more errors
			if googleErr, ok := err.(*googleapi.Error); ok && googleErr.Code == 410 {
				// Stop channel, new gets created in initial sync
				err = service.Channels.Stop(&calendar.Channel{
					Id:         *extCalendar.ChannelId,
					ResourceId: *extCalendar.ResourceId,
				}).Do()
				if err != nil {
					return err
				}

				err := s.resetExternalCalendar(ctx, extCalendar.Id)
				if err != nil {
					return err
				}

				return s.initialCalendarSync(ctx, service, extCalendar, calendarTz, merchantId)
			}
			return err
		}

		if len(events.Items) == 0 {
			if events.NextPageToken == "" {
				nextSyncToken = events.NextSyncToken
				break
			}

			req.PageToken(events.NextPageToken)

			continue
		}

		eventIds := make([]string, 0, len(events.Items))
		for _, ev := range events.Items {
			eventIds = append(eventIds, ev.Id)
		}

		existingEvents, err := s.externalCalendarRepo.GetExternalCalendarEvents(ctx, extCalendar.Id, eventIds)
		if err != nil {
			return err
		}

		existingEventsMap := make(map[string]domain.ExternalCalendarEvent, len(existingEvents))

		for _, e := range existingEvents {
			existingEventsMap[e.ExternalEventId] = e
		}

		for _, ev := range events.Items {
			existing, ok := existingEventsMap[ev.Id]

			// skip events that came from us
			if ok && existing.Source == types.EventSourceInternal {
				continue
			}

			// skip birthday events as they are all-day non-blocking and unnecessary
			if ev.EventType == "birthday" {
				continue
			}

			ece, err := eventToExternalCalendarEvent(ev, extCalendar.Id, calendarTz)
			if err != nil {
				return err
			}

			// apparently 0 duration google events are valid so skip them
			if !ece.FromDate.Before(ece.ToDate) {
				continue
			}

			isBlocking := ev.Transparency == "" || ev.Transparency == "opaque"
			ece.IsBlocking = isBlocking

			// event has been cancelled, delete corresponding BlockedTime
			if ev.Status == "cancelled" {
				if ok {
					if existing.InternalId != nil {
						deleteBlockedTimes = append(deleteBlockedTimes, *existing.InternalId)
					}

					updateExternalEvents = append(updateExternalEvents, ece)
				}

				continue
			}

			// etag indicates wether the event has changed
			// apparently cancelling event does not trigger a change
			if ok && existing.Etag == ev.Etag {
				continue
			}

			var bt domain.BlockedTime
			if isBlocking {
				bt, err = eventToBlockedTime(ev, merchantId, calendarTz)
				if err != nil {
					return err
				}
			}

			// event does not exist, insert new rows
			if !ok {

				if isBlocking {
					newBlockedTimes = append(newBlockedTimes, bt)
					newBlockingEventIdxs = append(newBlockingEventIdxs, len(newExternalEvents))

					ece.InternalType = &types.EventInternalTypeBlockedTime
				}

				newExternalEvents = append(newExternalEvents, ece)

				continue
			}

			switch {
			// event was not blocking but now is, insert new BlockedTime
			case !existing.IsBlocking && isBlocking:
				newBlockedTimes = append(newBlockedTimes, bt)

				pendingBlockingLinks = append(pendingBlockingLinks, externalEventBlockedTimeLink{
					// the externalEvent to update is the next one that will be appended to updateExternalEvents
					ExternalEventIdx: len(updateExternalEvents),
					BlockedTimeIdx:   len(newBlockedTimes) - 1,
				})

				ece.InternalType = &types.EventInternalTypeBlockedTime

			// event was blocking but now isn't, delete corresponding BlockedTime
			case existing.IsBlocking && !isBlocking:
				if existing.InternalId != nil {
					deleteBlockedTimes = append(deleteBlockedTimes, *existing.InternalId)
				}

			// blocking event, update BlockedTime as it has probably changed
			case existing.IsBlocking && isBlocking:
				bt.Id = *existing.InternalId
				updateBlockedTimes = append(updateBlockedTimes, bt)

				ece.InternalId = existing.InternalId
				ece.InternalType = &types.EventInternalTypeBlockedTime
			}

			// It's important to note that the switch statement's first case relies on the fact
			// that this append happens after getting the length of updateExternalEvents
			updateExternalEvents = append(updateExternalEvents, ece)
		}

		if events.NextPageToken == "" {
			nextSyncToken = events.NextSyncToken
			break
		}

		req.PageToken(events.NextPageToken)
	}

	err = s.incrementalCalendarSyncToDB(ctx, extCalendar.EmployeeId, newBlockedTimes, updateBlockedTimes, deleteBlockedTimes,
		newBlockingEventIdxs, newExternalEvents, updateExternalEvents, pendingBlockingLinks)
	if err != nil {
		return err
	}

	return s.externalCalendarRepo.UpdateExternalCalendarSyncToken(ctx, extCalendar.Id, nextSyncToken)
}

func bookingToGoogleEvent(booking domain.BookingForExternalCalendar, tz string) *calendar.Event {
	var startDate *calendar.EventDateTime
	var endDate *calendar.EventDateTime

	startDate = &calendar.EventDateTime{
		DateTime: booking.FromDate.Format(time.RFC3339),
		TimeZone: tz,
	}

	endDate = &calendar.EventDateTime{
		DateTime: booking.ToDate.Format(time.RFC3339),
		TimeZone: tz,
	}

	return &calendar.Event{
		Summary:      booking.ServiceName,
		Description:  *booking.ServiceDescription,
		Start:        startDate,
		End:          endDate,
		Location:     booking.FormattedLocation,
		Transparency: "opaque",
		Visibility:   "private",
		Source: &calendar.EventSource{
			Title: "Reservations",
			Url:   "http://app.reservations.local:3000/calendar",
		},
		ExtendedProperties: &calendar.EventExtendedProperties{
			Private: map[string]string{
				"internal_type": types.EventInternalTypeBooking.String(),
				"internal_id":   strconv.Itoa(booking.Id),
			},
		},
	}
}

func blockedTimeToGoogleEvent(blockedTime domain.BlockedTime, tz string) *calendar.Event {
	var startDate *calendar.EventDateTime
	var endDate *calendar.EventDateTime

	if blockedTime.AllDay {
		startDate = &calendar.EventDateTime{
			Date: blockedTime.FromDate.Format("2006-01-02"),
		}

		endDate = &calendar.EventDateTime{
			Date: blockedTime.ToDate.Format("2006-01-02"),
		}
	} else {
		startDate = &calendar.EventDateTime{
			DateTime: blockedTime.FromDate.Format(time.RFC3339),
			TimeZone: tz,
		}

		endDate = &calendar.EventDateTime{
			DateTime: blockedTime.ToDate.Format(time.RFC3339),
			TimeZone: tz,
		}
	}

	return &calendar.Event{
		Summary:      blockedTime.Name,
		Start:        startDate,
		End:          endDate,
		Transparency: "opaque",
		Visibility:   "private",
		Source: &calendar.EventSource{
			Title: "Reservations",
			Url:   "http://app.reservations.local:3000/calendar",
		},
		ExtendedProperties: &calendar.EventExtendedProperties{
			Private: map[string]string{
				"internal_type": types.EventInternalTypeBlockedTime.String(),
				"internal_id":   strconv.Itoa(blockedTime.Id),
			},
		},
	}
}

func (s *Service) persistTokenIfRefreshed(ctx context.Context, extCalendar domain.ExternalCalendar, ts oauth2.TokenSource) error {
	newToken, err := ts.Token()
	if err != nil {
		return err
	}

	if newToken.AccessToken == extCalendar.AccessToken {
		return nil
	}

	return s.externalCalendarRepo.UpdateExternalCalendarAuthTokens(ctx, extCalendar.Id, newToken.AccessToken, newToken.RefreshToken, newToken.Expiry)
}

type syncType struct {
	ExternalEventId *string
	InternalType    types.EventInternalType
	InternalId      int
	Action          string
	FromDate        *time.Time
	ToDate          *time.Time
	IsAllDay        bool
	IsBlocking      bool
	GoogleEvent     *calendar.Event
}

func (s *Service) syncGoogleEvent(ctx context.Context, extCalendar domain.ExternalCalendar, sync syncType) error {
	ts := googleCalendarConf.TokenSource(ctx, &oauth2.Token{
		AccessToken:  extCalendar.AccessToken,
		RefreshToken: extCalendar.RefreshToken,
		Expiry:       extCalendar.TokenExpiry,
	})

	service, err := calendar.NewService(ctx, option.WithTokenSource(ts))
	if err != nil {
		return err
	}

	switch strings.ToUpper(sync.Action) {
	case "INSERT":
		googleEvent, err := service.Events.Insert(extCalendar.CalendarId, sync.GoogleEvent).SendUpdates("none").Do()
		if err != nil {
			return err
		}

		err = s.externalCalendarRepo.NewExternalCalendarEvent(ctx, domain.ExternalCalendarEvent{
			ExternalCalendarId: extCalendar.Id,
			ExternalEventId:    googleEvent.Id,
			Etag:               googleEvent.Etag,
			Status:             googleEvent.Status,
			Title:              googleEvent.Summary,
			Description:        googleEvent.Description,
			FromDate:           *sync.FromDate,
			ToDate:             *sync.ToDate,
			IsAllDay:           sync.IsAllDay,
			InternalId:         &sync.InternalId,
			InternalType:       &sync.InternalType,
			IsBlocking:         sync.IsBlocking,
			Source:             types.EventSourceInternal,
		})
		if err != nil {
			return err
		}
	case "UPDATE":
		googleEvent, err := service.Events.Patch(extCalendar.CalendarId, *sync.ExternalEventId, sync.GoogleEvent).SendUpdates("none").Do()
		if err != nil {
			return err
		}

		err = s.externalCalendarRepo.UpdateExternalCalendarEvent(ctx, domain.ExternalCalendarEvent{
			ExternalCalendarId: extCalendar.Id,
			ExternalEventId:    googleEvent.Id,
			Etag:               googleEvent.Etag,
			Status:             googleEvent.Status,
			Title:              googleEvent.Summary,
			Description:        googleEvent.Description,
			FromDate:           *sync.FromDate,
			ToDate:             *sync.ToDate,
			IsAllDay:           sync.IsAllDay,
			InternalId:         &sync.InternalId,
			InternalType:       &sync.InternalType,
			IsBlocking:         sync.IsBlocking,
			Source:             types.EventSourceInternal,
		})
		if err != nil {
			return err
		}
	case "DELETE":
		err := service.Events.Delete(extCalendar.CalendarId, *sync.ExternalEventId).SendUpdates("none").Do()
		if gErr, ok := err.(*googleapi.Error); ok && gErr.Code == 404 {
			// Event not found in the external calendar
		} else if err != nil {
			return err
		}

		err = s.externalCalendarRepo.DeleteExternalCalendarEvent(ctx, sync.InternalId)
		if err != nil {
			return err
		}
	}

	return s.persistTokenIfRefreshed(ctx, extCalendar, ts)
}

func (s *Service) SyncNewBooking(ctx context.Context, bookingId int) error {
	booking, err := s.bookingRepo.GetBookingForExternalCalendar(ctx, bookingId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}

		return err
	}

	if booking.EmployeeId == nil {
		return nil
	}

	extCalendar, err := s.externalCalendarRepo.GetExternalCalendarByEmployeeId(ctx, *booking.EmployeeId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}

		return err
	}

	return s.syncGoogleEvent(ctx, extCalendar, syncType{
		ExternalEventId: nil,
		InternalType:    types.EventInternalTypeBooking,
		InternalId:      bookingId,
		Action:          "INSERT",
		FromDate:        &booking.FromDate,
		ToDate:          &booking.ToDate,
		IsAllDay:        false,
		IsBlocking:      true,
		// TODO: merchant timezone is likely equal to extCalendar timezone but not guaranteed
		GoogleEvent: bookingToGoogleEvent(booking, extCalendar.Timezone),
	})
}

func (s *Service) SyncUpdateBooking(ctx context.Context, bookingId int) error {
	booking, err := s.bookingRepo.GetBookingForExternalCalendar(ctx, bookingId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}

		return err
	}

	events, err := s.externalCalendarRepo.GetExternalCalendarEventsByInternal(ctx, types.EventInternalTypeBooking, bookingId)
	if err != nil {
		return err
	}

	if len(events) == 0 {
		return nil
	}

	if len(events) > 1 {
		return fmt.Errorf("bookings and external calendar events should match 1 to 1")
	}

	extCalendar, err := s.externalCalendarRepo.GetExternalCalendar(ctx, events[0].ExternalCalendarId)
	if err != nil {
		return err
	}

	return s.syncGoogleEvent(ctx, extCalendar, syncType{
		ExternalEventId: &events[0].ExternalEventId,
		InternalType:    types.EventInternalTypeBooking,
		InternalId:      bookingId,
		Action:          "UPDATE",
		FromDate:        &booking.FromDate,
		ToDate:          &booking.ToDate,
		IsAllDay:        false,
		IsBlocking:      true,
		// TODO: merchant timezone is likely equal to extCalendar timezone but not guaranteed
		GoogleEvent: bookingToGoogleEvent(booking, extCalendar.Timezone),
	})
}

func (s *Service) SyncDeleteBooking(ctx context.Context, bookingId int) error {
	events, err := s.externalCalendarRepo.GetExternalCalendarEventsByInternal(ctx, types.EventInternalTypeBooking, bookingId)
	if err != nil {
		return err
	}

	if len(events) == 0 {
		return nil
	}

	if len(events) > 1 {
		return fmt.Errorf("bookings and external calendar events should match 1 to 1")
	}

	extCalendar, err := s.externalCalendarRepo.GetExternalCalendar(ctx, events[0].ExternalCalendarId)
	if err != nil {
		return err
	}

	return s.syncGoogleEvent(ctx, extCalendar, syncType{
		ExternalEventId: &events[0].ExternalEventId,
		InternalType:    types.EventInternalTypeBooking,
		InternalId:      bookingId,
		Action:          "DELETE",
		FromDate:        nil,
		ToDate:          nil,
		IsAllDay:        false,
		IsBlocking:      true,
		GoogleEvent:     nil,
	})
}

func (s *Service) SyncNewBlockedTimeDispatcher(ctx context.Context, blockedTimeId int) error {
	blockedTime, err := s.blockedTimeRepo.GetBlockedTimeEmployees(ctx, blockedTimeId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}

		return err
	}

	var employeeIds []int

	if len(blockedTime.EmployeeIds) == 0 {
		employees, err := s.teamRepo.GetActiveEmployees(ctx, blockedTime.MerchantId)
		if err != nil {
			return err
		}

		if len(employees) == 0 {
			return nil
		}

		for _, e := range employees {
			employeeIds = append(employeeIds, e.Id)
		}
	} else {
		employeeIds = blockedTime.EmployeeIds
	}

	insertParams := make([]river.InsertManyParams, len(employeeIds))
	for i, e := range employeeIds {
		insertParams[i] = river.InsertManyParams{
			Args: args.SyncNewBlockedTime{
				BlockedTimeId: blockedTimeId,
				EmployeeId:    e,
			},
		}
	}

	_, err = s.enqueuer.InsertManyFast(ctx, insertParams)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) SyncNewBlockedTime(ctx context.Context, blockedTimeId int, employeeId int) error {
	blockedTime, err := s.blockedTimeRepo.GetBlockedTimeForEmployee(ctx, blockedTimeId, employeeId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}

		return err
	}

	extCalendar, err := s.externalCalendarRepo.GetExternalCalendarByEmployeeId(ctx, employeeId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}

		return err
	}

	return s.syncGoogleEvent(ctx, extCalendar, syncType{
		ExternalEventId: nil,
		InternalType:    types.EventInternalTypeBlockedTime,
		InternalId:      blockedTimeId,
		Action:          "INSERT",
		FromDate:        &blockedTime.FromDate,
		ToDate:          &blockedTime.ToDate,
		IsAllDay:        blockedTime.AllDay,
		IsBlocking:      true,
		// TODO: merchant timezone is likely equal to extCalendar timezone but not guaranteed
		GoogleEvent: blockedTimeToGoogleEvent(blockedTime, extCalendar.Timezone),
	})
}

func (s *Service) SyncUpdateBlockedTimeDispatcher(ctx context.Context, blockedTimeId int) error {
	events, err := s.externalCalendarRepo.GetExternalCalendarEventsByInternal(ctx, types.EventInternalTypeBlockedTime, blockedTimeId)
	if err != nil {
		return err
	}

	if len(events) > 0 {
		insertParams := make([]river.InsertManyParams, len(events))

		for i, e := range events {
			insertParams[i] = river.InsertManyParams{
				Args: args.SyncUpdateBlockedTime{
					BlockedTimeId:           blockedTimeId,
					ExternalCalendarEventId: e.Id,
				},
			}
		}

		_, err = s.enqueuer.InsertManyFast(ctx, insertParams)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) SyncUpdateBlockedTime(ctx context.Context, blockedTimeId int, extCalendarEventId int) error {
	event, err := s.externalCalendarRepo.GetExternalCalendarEvent(ctx, extCalendarEventId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}

		return err
	}

	extCalendar, err := s.externalCalendarRepo.GetExternalCalendar(ctx, event.ExternalCalendarId)
	if err != nil {
		return err
	}

	blockedTime, err := s.blockedTimeRepo.GetBlockedTimeForEmployee(ctx, blockedTimeId, extCalendar.EmployeeId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}

		return err
	}

	return s.syncGoogleEvent(ctx, extCalendar, syncType{
		ExternalEventId: &event.ExternalEventId,
		InternalType:    types.EventInternalTypeBlockedTime,
		InternalId:      blockedTimeId,
		Action:          "UPDATE",
		FromDate:        &blockedTime.FromDate,
		ToDate:          &blockedTime.ToDate,
		IsAllDay:        blockedTime.AllDay,
		IsBlocking:      true,
		// TODO: merchant timezone is likely equal to extCalendar timezone but not guaranteed
		GoogleEvent: blockedTimeToGoogleEvent(blockedTime, extCalendar.Timezone),
	})
}

func (s *Service) SyncDeleteBlockedTimeDispatcher(ctx context.Context, blockedTimeId int) error {
	events, err := s.externalCalendarRepo.GetExternalCalendarEventsByInternal(ctx, types.EventInternalTypeBlockedTime, blockedTimeId)
	if err != nil {
		return err
	}

	if len(events) > 0 {
		insertParams := make([]river.InsertManyParams, len(events))

		for i, e := range events {
			insertParams[i] = river.InsertManyParams{
				Args: args.SyncDeleteBlockedTime{
					BlockedTimeId:           blockedTimeId,
					ExternalCalendarEventId: e.Id,
				},
			}
		}

		_, err = s.enqueuer.InsertManyFast(ctx, insertParams)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) SyncDeleteBlockedTime(ctx context.Context, blockedTimeId int, extCalendarEventId int) error {
	event, err := s.externalCalendarRepo.GetExternalCalendarEvent(ctx, extCalendarEventId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}

		return err
	}

	extCalendar, err := s.externalCalendarRepo.GetExternalCalendar(ctx, event.ExternalCalendarId)
	if err != nil {
		return err
	}

	return s.syncGoogleEvent(ctx, extCalendar, syncType{
		ExternalEventId: &event.ExternalEventId,
		InternalType:    types.EventInternalTypeBlockedTime,
		InternalId:      blockedTimeId,
		Action:          "DELETE",
		FromDate:        nil,
		ToDate:          nil,
		IsAllDay:        false,
		IsBlocking:      true,
		GoogleEvent:     nil,
	})
}

func (s *Service) HandleChannelExpiration(ctx context.Context) error {
	extCalendars, err := s.externalCalendarRepo.GetExpiringExternalCalendars(ctx, time.Now().UTC().Add(time.Hour*24))
	if err != nil {
		return err
	}

	for _, extCal := range extCalendars {
		ts := googleCalendarConf.TokenSource(ctx, &oauth2.Token{
			AccessToken:  extCal.AccessToken,
			RefreshToken: extCal.RefreshToken,
			Expiry:       extCal.TokenExpiry,
		})

		service, err := calendar.NewService(ctx, option.WithTokenSource(ts))
		if err != nil {
			return err
		}

		err = service.Channels.Stop(&calendar.Channel{
			Id:         *extCal.ChannelId,
			ResourceId: *extCal.ResourceId,
		}).Do()
		if err != nil {
			return err
		}

		channelId := uuid.NewString()
		googleChannel, err := service.Events.Watch(extCal.CalendarId, &calendar.Channel{
			Id:      channelId,
			Type:    "web_hook",
			Address: "http://localhost:8080/api/v1/integrations/calendar/google/watch",
		}).Do()
		if err != nil {
			return err
		}

		err = s.externalCalendarRepo.UpdateExternalCalendarChannel(ctx, extCal.Id, googleChannel.Id, googleChannel.ResourceId,
			time.UnixMilli(googleChannel.Expiration))
		if err != nil {
			return err
		}
	}

	return nil
}
