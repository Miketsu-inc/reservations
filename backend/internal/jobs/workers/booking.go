package workers

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/miketsu-inc/reservations/backend/internal/domain"
	"github.com/miketsu-inc/reservations/backend/internal/jobs/args"
	bookingServ "github.com/miketsu-inc/reservations/backend/internal/service/booking"
	"github.com/riverqueue/river"
)

type RecurringBookingScheduler struct {
	river.WorkerDefaults[args.RecurringBookingScheduler]

	bookingRepo domain.BookingRepository
}

func NewRecurringBookingScheduler(bookingRepo domain.BookingRepository) *RecurringBookingScheduler {
	return &RecurringBookingScheduler{bookingRepo: bookingRepo}
}

func (w *RecurringBookingScheduler) Work(ctx context.Context, job *river.Job[args.RecurringBookingScheduler]) error {
	tresholdTime := time.Now().UTC().AddDate(0, 3, 0)

	seriesIds, err := w.bookingRepo.GetActiveBookingSeriesIds(ctx, tresholdTime)
	if err != nil {
		return err
	}

	if len(seriesIds) == 0 {
		return nil
	}

	client := river.ClientFromContext[pgx.Tx](ctx)

	insertParams := make([]river.InsertManyParams, len(seriesIds))
	for i, id := range seriesIds {
		insertParams[i] = river.InsertManyParams{
			Args: args.BookingOccurrenceGenerator{BookingSeriesId: id},
		}
	}

	_, err = client.InsertMany(ctx, insertParams)
	if err != nil {
		return err
	}

	return nil
}

type BookingOccurrenceGenerator struct {
	river.WorkerDefaults[args.BookingOccurrenceGenerator]

	bookingService *bookingServ.Service
	bookingRepo    domain.BookingRepository
	catalogRepo    domain.CatalogRepository
}

func NewBookingOccurrenceGenerator(bookingService *bookingServ.Service, bookingRepo domain.BookingRepository, catalogRepo domain.CatalogRepository) *BookingOccurrenceGenerator {
	return &BookingOccurrenceGenerator{bookingService: bookingService, bookingRepo: bookingRepo, catalogRepo: catalogRepo}
}

func (w *BookingOccurrenceGenerator) Work(ctx context.Context, job *river.Job[args.BookingOccurrenceGenerator]) error {
	series, err := w.bookingRepo.GetBookingSeries(ctx, job.Args.BookingSeriesId)
	if err != nil {
		return err
	}

	if !series.IsActive {
		return nil
	}

	seriesParticipants, err := w.bookingRepo.GetBookingSeriesParticipants(ctx, job.Args.BookingSeriesId)
	if err != nil {
		return err
	}

	service, err := w.catalogRepo.GetServiceWithPhases(ctx, series.ServiceId, series.MerchantId)
	if err != nil {
		return err
	}

	var generateFrom time.Time

	// only nil on the first generation, which is triggered 'manually'
	if series.GeneratedUntil != nil {
		generateFrom = *series.GeneratedUntil
	} else {
		generateFrom = job.Args.GenerateFrom
	}

	return w.bookingService.GenerateRecurringBookings(ctx, series, seriesParticipants, service, generateFrom)
}

type UpdateFutureBookingOccurrences struct {
	river.WorkerDefaults[args.UpdateFutureBookingOccurrences]

	bookingService *bookingServ.Service
	bookingRepo    domain.BookingRepository
	catalogRepo    domain.CatalogRepository
}

func NewUpdateFutureBookingOccurrences(bookingService *bookingServ.Service, bookingRepo domain.BookingRepository, catalogRepo domain.CatalogRepository) *UpdateFutureBookingOccurrences {
	return &UpdateFutureBookingOccurrences{bookingService: bookingService, bookingRepo: bookingRepo, catalogRepo: catalogRepo}
}

func (w *UpdateFutureBookingOccurrences) Work(ctx context.Context, job *river.Job[args.UpdateFutureBookingOccurrences]) error {
	series, err := w.bookingRepo.GetBookingSeries(ctx, job.Args.BookingSeriesId)
	if err != nil {
		return err
	}

	service, err := w.catalogRepo.GetServiceWithPhases(ctx, series.ServiceId, series.MerchantId)
	if err != nil {
		return err
	}

	return w.bookingService.UpdateFutureBookingOccurrences(ctx, series, service, job.Args.OriginalFromDate, job.Args.FromDateOffset, job.Args.PriceChanged,
		job.Args.ParticipantsToInsert, job.Args.ParticipantsToDelete)
}
