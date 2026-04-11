package workers

import (
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/miketsu-inc/reservations/backend/internal/domain"
	"github.com/miketsu-inc/reservations/backend/internal/jobs/args"
	"github.com/miketsu-inc/reservations/backend/internal/service/booking"
	"github.com/miketsu-inc/reservations/backend/pkg/db"
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

	bookingService *booking.Service
	bookingRepo    domain.BookingRepository
	catalogRepo    domain.CatalogRepository
	txManager      db.TransactionManager
}

func NewBookingOccurrenceGenerator(bookingService *booking.Service, bookingRepo domain.BookingRepository, catalogRepo domain.CatalogRepository,
	txManager db.TransactionManager) *BookingOccurrenceGenerator {
	return &BookingOccurrenceGenerator{bookingService: bookingService, bookingRepo: bookingRepo, catalogRepo: catalogRepo, txManager: txManager}
}

func (w *BookingOccurrenceGenerator) Work(ctx context.Context, job *river.Job[args.BookingOccurrenceGenerator]) error {
	series, err := w.bookingRepo.GetBookingSeries(ctx, job.Args.BookingSeriesId)
	if err != nil {
		return err
	}

	seriesDetails, err := w.bookingRepo.GetBookingSeriesDetails(ctx, job.Args.BookingSeriesId)
	if err != nil {
		return err
	}

	seriesParticipants, err := w.bookingRepo.GetBookingSeriesParticipants(ctx, job.Args.BookingSeriesId)
	if err != nil {
		return err
	}

	service, err := w.catalogRepo.GetServiceWithPhases(ctx, series.ServiceId, series.MerchantId)
	if err != nil {
		return err
	}

	return w.txManager.WithTransaction(ctx, func(tx pgx.Tx) error {
		var generateFrom time.Time

		if series.GeneratedUntil != nil {
			generateFrom = *series.GeneratedUntil
		} else {
			slog.DebugContext(ctx, "BookingSeries's generated_until should not be nil here!")
			generateFrom = time.Now().UTC()
		}

		_, err = w.bookingService.GenerateRecurringBookings(ctx, tx, series, seriesDetails, seriesParticipants, service.Phases, generateFrom)
		if err != nil {
			return err
		}

		return nil
	})
}
