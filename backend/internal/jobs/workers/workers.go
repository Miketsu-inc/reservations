package workers

import (
	"github.com/miketsu-inc/reservations/backend/internal/domain"
	"github.com/miketsu-inc/reservations/backend/internal/service/booking"
	"github.com/miketsu-inc/reservations/backend/internal/service/email"
	"github.com/miketsu-inc/reservations/backend/pkg/db"
	"github.com/riverqueue/river"
)

type Deps struct {
	BookingService *booking.Service
	EmailService   *email.Service
	BookingRepo    domain.BookingRepository
	TxManager      db.TransactionManager
}

func RegisterWorkers(workers *river.Workers, deps Deps) {
	river.AddWorker(workers, NewBookingConfirmationEmailWorker(deps.EmailService, deps.BookingRepo))
	river.AddWorker(workers, NewBookingReminderEmailWorker(deps.EmailService, deps.BookingRepo))
	river.AddWorker(workers, NewBookingCancellationEmail(deps.EmailService, deps.BookingRepo))
	river.AddWorker(workers, NewBookingModificationEmail(deps.EmailService, deps.BookingRepo))
}
