package workers

import (
	"context"

	"github.com/miketsu-inc/reservations/backend/internal/domain"
	"github.com/miketsu-inc/reservations/backend/internal/jobs/args"
	"github.com/miketsu-inc/reservations/backend/internal/service/externalcalendar"
	"github.com/riverqueue/river"
)

type IncrementalCalendarSync struct {
	river.WorkerDefaults[args.IncrementalCalendarSync]

	externalCalendarService *externalcalendar.Service
	externalCalendarRepo    domain.ExternalCalendarRepository
}

func NewIncrementalCalendarSync(extCalendarService *externalcalendar.Service, extCalendarRepo domain.ExternalCalendarRepository) *IncrementalCalendarSync {
	return &IncrementalCalendarSync{externalCalendarService: extCalendarService, externalCalendarRepo: extCalendarRepo}
}

func (w *IncrementalCalendarSync) Work(ctx context.Context, job *river.Job[args.IncrementalCalendarSync]) error {
	extCalendar, err := w.externalCalendarRepo.GetExternalCalendar(ctx, job.Args.ExternalCalendarId)
	if err != nil {
		return err
	}

	err = w.externalCalendarService.IncrementalCalendarSync(ctx, extCalendar)
	if err != nil {
		return err
	}

	return nil
}

type SyncNewBooking struct {
	river.WorkerDefaults[args.SyncNewBooking]

	externalCalendarService *externalcalendar.Service
}

func NewSyncNewBooking(extCalendarService *externalcalendar.Service) *SyncNewBooking {
	return &SyncNewBooking{externalCalendarService: extCalendarService}
}

func (w *SyncNewBooking) Work(ctx context.Context, job *river.Job[args.SyncNewBooking]) error {
	return w.externalCalendarService.SyncNewBooking(ctx, job.Args.BookingId)
}

type SyncUpdateBooking struct {
	river.WorkerDefaults[args.SyncUpdateBooking]

	externalCalendarService *externalcalendar.Service
}

func NewSyncUpdateBooking(extCalendarService *externalcalendar.Service) *SyncUpdateBooking {
	return &SyncUpdateBooking{externalCalendarService: extCalendarService}
}

func (w *SyncUpdateBooking) Work(ctx context.Context, job *river.Job[args.SyncUpdateBooking]) error {
	return w.externalCalendarService.SyncUpdateBooking(ctx, job.Args.BookingId)
}

type SyncDeleteBooking struct {
	river.WorkerDefaults[args.SyncDeleteBooking]

	externalCalendarService *externalcalendar.Service
}

func NewSyncDeleteBooking(extCalendarService *externalcalendar.Service) *SyncDeleteBooking {
	return &SyncDeleteBooking{externalCalendarService: extCalendarService}
}

func (w *SyncDeleteBooking) Work(ctx context.Context, job *river.Job[args.SyncDeleteBooking]) error {
	return w.externalCalendarService.SyncDeleteBooking(ctx, job.Args.BookingId)
}

type SyncNewBlockedTime struct {
	river.WorkerDefaults[args.SyncNewBlockedTime]

	externalCalendarService *externalcalendar.Service
}

func NewSyncNewBlockedTime(extCalendarService *externalcalendar.Service) *SyncNewBlockedTime {
	return &SyncNewBlockedTime{externalCalendarService: extCalendarService}
}

func (w *SyncNewBlockedTime) Work(ctx context.Context, job *river.Job[args.SyncNewBlockedTime]) error {
	return w.externalCalendarService.SyncNewBlockedTime(ctx, job.Args.BlockedTimeId)
}

type SyncUpdateBlockedTime struct {
	river.WorkerDefaults[args.SyncUpdateBlockedTime]

	externalCalendarService *externalcalendar.Service
}

func NewSyncUpdateBlockedTime(extCalendarService *externalcalendar.Service) *SyncUpdateBlockedTime {
	return &SyncUpdateBlockedTime{externalCalendarService: extCalendarService}
}

func (w *SyncUpdateBlockedTime) Work(ctx context.Context, job *river.Job[args.SyncUpdateBlockedTime]) error {
	return w.externalCalendarService.SyncUpdateBlockedTime(ctx, job.Args.BlockedTimeId)
}

type SyncDeleteBlockedTime struct {
	river.WorkerDefaults[args.SyncDeleteBlockedTime]

	externalCalendarService *externalcalendar.Service
}

func NewSyncDeleteBlockedTime(extCalendarService *externalcalendar.Service) *SyncDeleteBlockedTime {
	return &SyncDeleteBlockedTime{externalCalendarService: extCalendarService}
}

func (w *SyncDeleteBlockedTime) Work(ctx context.Context, job *river.Job[args.SyncDeleteBlockedTime]) error {
	return w.externalCalendarService.SyncDeleteBlockedTime(ctx, job.Args.BlockedTimeId)
}

type HandleChannelExpiration struct {
	river.WorkerDefaults[args.HandleChannelExpiration]

	externalCalendarService *externalcalendar.Service
}

func NewHandleChannelExpiration(extCalendarService *externalcalendar.Service) *HandleChannelExpiration {
	return &HandleChannelExpiration{externalCalendarService: extCalendarService}
}

func (w *HandleChannelExpiration) Work(ctx context.Context, job *river.Job[args.HandleChannelExpiration]) error {
	return w.externalCalendarService.HandleChannelExpiration(ctx)
}
