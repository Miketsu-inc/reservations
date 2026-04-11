package workers

import (
	"time"

	"github.com/miketsu-inc/reservations/backend/internal/domain"
	"github.com/miketsu-inc/reservations/backend/internal/jobs/args"
	"github.com/miketsu-inc/reservations/backend/internal/jobs/schedule"
	"github.com/miketsu-inc/reservations/backend/internal/service/booking"
	"github.com/miketsu-inc/reservations/backend/internal/service/email"
	"github.com/miketsu-inc/reservations/backend/internal/service/externalcalendar"
	"github.com/miketsu-inc/reservations/backend/pkg/db"
	"github.com/riverqueue/river"
)

type Deps struct {
	BookingService     *booking.Service
	EmailService       *email.Service
	ExtCalendarService *externalcalendar.Service
	BookingRepo        domain.BookingRepository
	CatalogRepo        domain.CatalogRepository
	ExtCalendarRepo    domain.ExternalCalendarRepository
	TxManager          db.TransactionManager
}

func RegisterWorkers(workers *river.Workers, deps Deps) {
	river.AddWorker(workers, NewBookingConfirmationEmailWorker(deps.EmailService, deps.BookingRepo))
	river.AddWorker(workers, NewBookingReminderEmailWorker(deps.EmailService, deps.BookingRepo))
	river.AddWorker(workers, NewBookingCancellationEmail(deps.EmailService, deps.BookingRepo))
	river.AddWorker(workers, NewBookingModificationEmail(deps.EmailService, deps.BookingRepo))

	river.AddWorker(workers, NewIncrementalCalendarSync(deps.ExtCalendarService, deps.ExtCalendarRepo))
	river.AddWorker(workers, NewSyncNewBooking(deps.ExtCalendarService))
	river.AddWorker(workers, NewSyncUpdateBooking(deps.ExtCalendarService))
	river.AddWorker(workers, NewSyncDeleteBooking(deps.ExtCalendarService))
	river.AddWorker(workers, NewSyncNewBlockedTime(deps.ExtCalendarService))
	river.AddWorker(workers, NewSyncUpdateBlockedTime(deps.ExtCalendarService))
	river.AddWorker(workers, NewSyncDeleteBlockedTime(deps.ExtCalendarService))
	river.AddWorker(workers, NewHandleChannelExpiration(deps.ExtCalendarService))

	river.AddWorker(workers, NewRecurringBookingScheduler(deps.BookingRepo))
	river.AddWorker(workers, NewBookingOccurrenceGenerator(deps.BookingService, deps.BookingRepo, deps.CatalogRepo, deps.TxManager))
}

func GetPeriodicJobs() []*river.PeriodicJob {
	return []*river.PeriodicJob{
		river.NewPeriodicJob(schedule.NewDailyMidnight(time.UTC),
			func() (river.JobArgs, *river.InsertOpts) {
				return args.HandleChannelExpiration{}, nil
			}, &river.PeriodicJobOpts{RunOnStart: true},
		),
		river.NewPeriodicJob(schedule.NewDailyMidnight(time.UTC),
			func() (river.JobArgs, *river.InsertOpts) {
				return args.RecurringBookingScheduler{}, nil
			}, &river.PeriodicJobOpts{RunOnStart: true},
		),
	}
}
