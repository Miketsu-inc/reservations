package workers

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/miketsu-inc/reservations/backend/internal/api/middleware/lang"
	"github.com/miketsu-inc/reservations/backend/internal/domain"
	"github.com/miketsu-inc/reservations/backend/internal/jobs/args"
	"github.com/miketsu-inc/reservations/backend/internal/service/email"
	"github.com/miketsu-inc/reservations/backend/internal/types"
	"github.com/riverqueue/river"
	"golang.org/x/text/language"
)

type BookingConfirmationEmail struct {
	river.WorkerDefaults[args.BookingConfirmationEmail]

	emailService *email.Service
	bookingRepo  domain.BookingRepository
}

func NewBookingConfirmationEmailWorker(emailService *email.Service, bookingRepo domain.BookingRepository) *BookingConfirmationEmail {
	return &BookingConfirmationEmail{emailService: emailService, bookingRepo: bookingRepo}
}

func (w *BookingConfirmationEmail) Work(ctx context.Context, job *river.Job[args.BookingConfirmationEmail]) error {
	booking, err := w.bookingRepo.GetBookingForEmail(ctx, job.Args.BookingId, job.Args.CustomerId)
	if err != nil {
		return err
	}

	// added customer without email
	if booking.CustomerEmail == nil {
		return nil
	}

	// participant was deleted before job could run
	if booking.ParticipantStatus == nil {
		return nil
	}

	if booking.Status == types.BookingStatusCancelled || *booking.ParticipantStatus == types.BookingStatusCancelled {
		return nil
	}

	if booking.Status == types.BookingStatusCompleted || *booking.ParticipantStatus == types.BookingStatusCompleted {
		return nil
	}

	if time.Now().UTC().After(booking.FromDate) {
		return nil
	}

	merchantTz, err := time.LoadLocation(booking.Timezone)
	if err != nil {
		return err
	}

	fromDateMerchantTz := booking.FromDate.In(merchantTz)
	toDateMerchantTz := booking.ToDate.In(merchantTz)

	lang := lang.GetDefaultLang()

	if booking.UserLanguage != nil {
		lang, err = language.Parse(*booking.UserLanguage)
		if err != nil {
			return err
		}
	}

	return w.emailService.BookingConfirmation(ctx, lang, *booking.CustomerEmail, email.BookingConfirmationData{
		Time:        fmt.Sprintf("%s - %s", fromDateMerchantTz.Format("15:04"), toDateMerchantTz.Format("15:04")),
		Date:        fromDateMerchantTz.Format("Monday, January 2"),
		Location:    booking.FormattedLocation,
		ServiceName: booking.ServiceName,
		TimeZone:    merchantTz.String(),
		ModifyLink:  fmt.Sprintf("http://reservations.local:3000/m/%s/cancel/%d", booking.MerchantUrl, booking.Id),
	})
}

type BookingStatusConfirmedEmail struct {
	river.WorkerDefaults[args.BookingStatusConfirmedEmail]

	emailService *email.Service
	bookingRepo  domain.BookingRepository
}

func NewBookingStatusConfirmedEmailWorker(emailService *email.Service, bookingRepo domain.BookingRepository) *BookingStatusConfirmedEmail {
	return &BookingStatusConfirmedEmail{emailService: emailService, bookingRepo: bookingRepo}
}

func (w *BookingStatusConfirmedEmail) Work(ctx context.Context, job *river.Job[args.BookingStatusConfirmedEmail]) error {
	booking, err := w.bookingRepo.GetBookingForEmail(ctx, job.Args.BookingId, job.Args.CustomerId)
	if err != nil {
		return err
	}

	// added customer without email
	if booking.CustomerEmail == nil {
		return nil
	}

	// participant was deleted before job could run
	if booking.ParticipantStatus == nil {
		return nil
	}

	if booking.Status == types.BookingStatusCancelled || *booking.ParticipantStatus == types.BookingStatusCancelled {
		return nil
	}

	if booking.Status == types.BookingStatusCompleted || *booking.ParticipantStatus == types.BookingStatusCompleted {
		return nil
	}

	if time.Now().UTC().After(booking.FromDate) {
		return nil
	}

	if booking.Status != types.BookingStatusConfirmed {
		return nil
	}

	merchantTz, err := time.LoadLocation(booking.Timezone)
	if err != nil {
		return err
	}

	fromDateMerchantTz := booking.FromDate.In(merchantTz)
	toDateMerchantTz := booking.ToDate.In(merchantTz)

	lang := lang.GetDefaultLang()

	if booking.UserLanguage != nil {
		lang, err = language.Parse(*booking.UserLanguage)
		if err != nil {
			return err
		}
	}

	return w.emailService.BookingStatusConfirmed(ctx, lang, *booking.CustomerEmail, email.BookingConfirmationData{
		Time:        fmt.Sprintf("%s - %s", fromDateMerchantTz.Format("15:04"), toDateMerchantTz.Format("15:04")),
		Date:        fromDateMerchantTz.Format("Monday, January 2"),
		Location:    booking.FormattedLocation,
		ServiceName: booking.ServiceName,
		TimeZone:    merchantTz.String(),
		ModifyLink:  fmt.Sprintf("http://reservations.local:3000/m/%s/cancel/%d", booking.MerchantUrl, booking.Id),
	})
}

type BookingReminderEmail struct {
	river.WorkerDefaults[args.BookingReminderEmail]

	emailService *email.Service
	bookingRepo  domain.BookingRepository
}

func NewBookingReminderEmailWorker(emailService *email.Service, bookingRepo domain.BookingRepository) *BookingReminderEmail {
	return &BookingReminderEmail{emailService: emailService, bookingRepo: bookingRepo}
}

func (w *BookingReminderEmail) Work(ctx context.Context, job *river.Job[args.BookingReminderEmail]) error {
	booking, err := w.bookingRepo.GetBookingForEmail(ctx, job.Args.BookingId, job.Args.CustomerId)
	if err != nil {
		return err
	}

	if booking.CustomerEmail == nil {
		return nil
	}

	if booking.ParticipantStatus == nil {
		return nil
	}

	if booking.Status == types.BookingStatusCancelled || *booking.ParticipantStatus == types.BookingStatusCancelled {
		return nil
	}

	if booking.Status == types.BookingStatusCompleted || *booking.ParticipantStatus == types.BookingStatusCompleted {
		return nil
	}

	if time.Now().UTC().After(booking.FromDate) {
		return nil
	}

	if !job.Args.ExpectedFromDate.Equal(booking.FromDate) {
		return nil
	}

	merchantTz, err := time.LoadLocation(booking.Timezone)
	if err != nil {
		return err
	}

	fromDateMerchantTz := booking.FromDate.In(merchantTz)
	toDateMerchantTz := booking.ToDate.In(merchantTz)

	hoursUntilBooking := time.Until(fromDateMerchantTz).Hours()

	if hoursUntilBooking < 24 {
		return nil
	}

	lang := lang.GetDefaultLang()

	if booking.UserLanguage != nil {
		lang, err = language.Parse(*booking.UserLanguage)
		if err != nil {
			return err
		}
	}

	return w.emailService.BookingReminder(ctx, lang, *booking.CustomerEmail, email.BookingConfirmationData{
		Time:        fmt.Sprintf("%s - %s", fromDateMerchantTz.Format("15:04"), toDateMerchantTz.Format("15:04")),
		Date:        fromDateMerchantTz.Format("Monday, January 2"),
		Location:    booking.FormattedLocation,
		ServiceName: booking.ServiceName,
		TimeZone:    merchantTz.String(),
		ModifyLink:  fmt.Sprintf("http://reservations.local:3000/m/%s/cancel/%d", booking.MerchantUrl, booking.Id),
	})
}

type BookingCancellationEmail struct {
	river.WorkerDefaults[args.BookingCancellationEmail]

	emailService *email.Service
	bookingRepo  domain.BookingRepository
}

func NewBookingCancellationEmail(emailService *email.Service, bookingRepo domain.BookingRepository) *BookingCancellationEmail {
	return &BookingCancellationEmail{emailService: emailService, bookingRepo: bookingRepo}
}

func (w *BookingCancellationEmail) Work(ctx context.Context, job *river.Job[args.BookingCancellationEmail]) error {
	booking, err := w.bookingRepo.GetBookingForEmail(ctx, job.Args.BookingId, job.Args.CustomerId)
	if err != nil {
		return err
	}

	if booking.CustomerEmail == nil {
		return nil
	}

	// TODO: how could we prevent duplicate cancellation emails?
	// if booking.Status == types.BookingStatusCancelled || booking.ParticipantStatus == types.BookingStatusCancelled {
	// 	return nil
	// }

	if booking.ParticipantStatus == nil {
		return nil
	}

	if booking.Status == types.BookingStatusCompleted || *booking.ParticipantStatus == types.BookingStatusCompleted {
		return nil
	}

	if time.Now().UTC().After(booking.FromDate) {
		return nil
	}

	merchantTz, err := time.LoadLocation(booking.Timezone)
	if err != nil {
		return err
	}

	fromDateMerchantTz := booking.FromDate.In(merchantTz)
	toDateMerchantTz := booking.ToDate.In(merchantTz)

	lang := lang.GetDefaultLang()

	if booking.UserLanguage != nil {
		lang, err = language.Parse(*booking.UserLanguage)
		if err != nil {
			return err
		}
	}

	return w.emailService.BookingCancellation(ctx, lang, *booking.CustomerEmail, email.BookingCancellationData{
		Time:           fmt.Sprintf("%s - %s", fromDateMerchantTz.Format("15:04"), toDateMerchantTz.Format("15:04")),
		Date:           fromDateMerchantTz.Format("Monday, January 2"),
		Location:       booking.FormattedLocation,
		ServiceName:    booking.ServiceName,
		TimeZone:       merchantTz.String(),
		Reason:         job.Args.CancellationReason,
		NewBookingLink: fmt.Sprintf("http://reservations.local:3000/m/%s", booking.MerchantUrl),
	})
}

type BookingModificationEmail struct {
	river.WorkerDefaults[args.BookingModificationEmail]

	emailService *email.Service
	bookingRepo  domain.BookingRepository
}

func NewBookingModificationEmail(emailService *email.Service, bookingRepo domain.BookingRepository) *BookingModificationEmail {
	return &BookingModificationEmail{emailService: emailService, bookingRepo: bookingRepo}
}

// TODO: the email should probably contain both the old and new service names
// also we should probably have either different email templates or some kind of checking for what's changed
func (w *BookingModificationEmail) Work(ctx context.Context, job *river.Job[args.BookingModificationEmail]) error {
	booking, err := w.bookingRepo.GetBookingForEmail(ctx, job.Args.BookingId, job.Args.CustomerId)
	if err != nil {
		return err
	}

	if booking.CustomerEmail == nil {
		return nil
	}

	if booking.ParticipantStatus == nil {
		return nil
	}

	if booking.Status == types.BookingStatusCancelled || *booking.ParticipantStatus == types.BookingStatusCancelled {
		return nil
	}

	if booking.Status == types.BookingStatusCompleted || *booking.ParticipantStatus == types.BookingStatusCompleted {
		return nil
	}

	if time.Now().UTC().After(booking.FromDate) {
		return nil
	}

	merchantTz, err := time.LoadLocation(booking.Timezone)
	if err != nil {
		return err
	}

	oldFromDateMerchantTz := job.Args.OldFromDate.In(merchantTz)
	oldToDateMerchantTz := job.Args.OldToDate.In(merchantTz)

	fromDateMerchantTz := booking.FromDate.In(merchantTz)
	toDateMerchantTz := booking.ToDate.In(merchantTz)

	lang := lang.GetDefaultLang()

	if booking.UserLanguage != nil {
		lang, err = language.Parse(*booking.UserLanguage)
		if err != nil {
			return err
		}
	}

	return w.emailService.BookingModification(ctx, lang, *booking.CustomerEmail, email.BookingModificationData{
		Time:        fmt.Sprintf("%s - %s", fromDateMerchantTz.Format("15:04"), toDateMerchantTz.Format("15:04")),
		Date:        fromDateMerchantTz.Format("Monday, January 2"),
		Location:    booking.FormattedLocation,
		ServiceName: job.Args.OldServiceName,
		TimeZone:    merchantTz.String(),
		ModifyLink:  fmt.Sprintf("http://reservations.local:3000/m/%s/cancel/%d", booking.MerchantUrl, booking.Id),
		OldTime:     fmt.Sprintf("%s - %s", oldFromDateMerchantTz.Format("15:04"), oldToDateMerchantTz.Format("15:04")),
		OldDate:     oldFromDateMerchantTz.Format("Monday, January 2"),
	})
}

type ForgotPasswordEmail struct {
	river.WorkerDefaults[args.ForgotPasswordEmail]

	emailService *email.Service
	userRepo     domain.UserRepository
}

func NewForgotPasswordEmail(emailService *email.Service, userRepo domain.UserRepository) *ForgotPasswordEmail {
	return &ForgotPasswordEmail{emailService: emailService, userRepo: userRepo}
}

func (w *ForgotPasswordEmail) Work(ctx context.Context, job *river.Job[args.ForgotPasswordEmail]) error {
	user, err := w.userRepo.GetUser(ctx, job.Args.UserId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}

		return err
	}

	if user.IsOauthUser() {
		return nil
	}

	return w.emailService.ForgotPassword(ctx, job.Args.Language, user.Email, email.ForgotPasswordData{
		PasswordLink: fmt.Sprintf("http://reservations.local:3000/reset-password?token=%s", job.Args.Token),
	})
}
