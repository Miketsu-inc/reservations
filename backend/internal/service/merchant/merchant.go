package merchant

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/miketsu-inc/reservations/backend/internal/api/middleware/actor"
	"github.com/miketsu-inc/reservations/backend/internal/domain"
	"github.com/miketsu-inc/reservations/backend/internal/utils"
	"github.com/miketsu-inc/reservations/backend/pkg/apperr"
	"github.com/miketsu-inc/reservations/backend/pkg/db"
	"github.com/miketsu-inc/reservations/backend/pkg/timeutil"
	"github.com/miketsu-inc/reservations/backend/pkg/validate"
)

type Service struct {
	bookingRepo     domain.BookingRepository
	catalogRepo     domain.CatalogRepository
	merchantRepo    domain.MerchantRepository
	customerRepo    domain.CustomerRepository
	blockedTimeRepo domain.BlockedTimeRepository
	teamRepo        domain.TeamRepository
	txManager       db.TransactionManager
}

func NewService(booking domain.BookingRepository, catalog domain.CatalogRepository, merchant domain.MerchantRepository,
	customer domain.CustomerRepository, blockedTime domain.BlockedTimeRepository, team domain.TeamRepository,
	txManager db.TransactionManager) *Service {
	return &Service{
		bookingRepo:     booking,
		catalogRepo:     catalog,
		merchantRepo:    merchant,
		customerRepo:    customer,
		blockedTimeRepo: blockedTime,
		teamRepo:        team,
		txManager:       txManager,
	}
}

func (s *Service) Delete(ctx context.Context) error {
	actor := actor.MustGetFromContext(ctx)

	err := s.merchantRepo.DeleteMerchant(ctx, actor.EmployeeId, actor.MerchantId)
	if err != nil {
		return fmt.Errorf("error while deleting merchant: %s", err.Error())
	}

	return nil
}

type UpdateNameInput struct {
	Name string
}

func (s *Service) UpdateName(ctx context.Context, input UpdateNameInput) error {
	actor := actor.MustGetFromContext(ctx)

	urlName, err := validate.MerchantNameToUrlName(input.Name)
	if err != nil {
		return fmt.Errorf("unexpected error during merchant url name conversion: %s", err.Error())
	}

	currentURL, err := s.merchantRepo.GetMerchantUrlName(ctx, actor.MerchantId)
	if err != nil {
		return err
	}

	if urlName != currentURL {
		unique, err := s.merchantRepo.IsMerchantUrlUnique(ctx, urlName)
		if err != nil {
			return err
		}

		if !unique {
			return apperr.Wrap(ErrMerchantUrlNotUnique, nil).With("merchant_url", urlName)
		}
	}

	err = s.merchantRepo.ChangeMerchantNameAndURL(ctx, actor.MerchantId, input.Name, urlName)
	if err != nil {
		return err
	}

	return nil
}

func getPeriodDates(utcDate time.Time, period int) (time.Time, time.Time, time.Time, error) {
	// -1 because the last is the current day
	currPeriodStart := utils.TruncateToDay(utcDate.AddDate(0, 0, -(period - 1)))
	prevPeriodStart := utils.TruncateToDay(currPeriodStart.AddDate(0, 0, -period))

	return currPeriodStart, utils.TruncateToDay(utcDate), prevPeriodStart, nil
}

func (s *Service) GetDashboardStatistics(ctx context.Context, period int) (domain.DashboardStatistics, error) {
	utcDate := time.Now().UTC()

	currPeriodStart, _, prevPeriodStart, err := getPeriodDates(utcDate, period)
	if err != nil {
		return domain.DashboardStatistics{}, err
	}

	actor := actor.MustGetFromContext(ctx)

	statistics, err := s.merchantRepo.GetDashboardStats(ctx, actor.MerchantId, currPeriodStart, utcDate, prevPeriodStart)
	if err != nil {
		return domain.DashboardStatistics{}, err
	}

	return statistics, nil
}

func (s *Service) GetDashboardRevenue(ctx context.Context, period int) (domain.DashboardRevenue, error) {
	utcDate := time.Now().UTC()

	currPeriodStart, periodEnd, _, err := getPeriodDates(utcDate, period)
	if err != nil {
		return domain.DashboardRevenue{}, err
	}

	actor := actor.MustGetFromContext(ctx)

	revenue, err := s.merchantRepo.GetRevenueStats(ctx, actor.MerchantId, currPeriodStart, utcDate)
	if err != nil {
		return domain.DashboardRevenue{}, err
	}

	return domain.DashboardRevenue{
		PeriodStart: currPeriodStart,
		PeriodEnd:   periodEnd,
		Revenue:     revenue,
	}, nil
}

type CheckUrlInput struct {
	Name string
}

func (s *Service) CheckUrl(ctx context.Context, input CheckUrlInput) (string, error) {
	urlName, err := validate.MerchantNameToUrlName(input.Name)
	if err != nil {
		return "", fmt.Errorf("unexpected error during merchant url name conversion: %w", err)
	}

	unique, err := s.merchantRepo.IsMerchantUrlUnique(ctx, urlName)
	if err != nil {
		return "", err
	}

	if !unique {
		return "", apperr.Wrap(ErrMerchantUrlNotUnique, nil).With("merchant_url", urlName)
	}

	return urlName, nil
}

func (s *Service) GetBusinessProfileSettings(ctx context.Context) (domain.BusinessProfileSettings, error) {
	actor := actor.MustGetFromContext(ctx)

	settings, err := s.merchantRepo.GetBusinessProfileSettings(ctx, actor.MerchantId)
	if err != nil {
		return domain.BusinessProfileSettings{}, err
	}

	return settings, nil
}

func (s *Service) UpdateBusinessProfileSettings(ctx context.Context, settings domain.BusinessProfileSettings) error {
	actor := actor.MustGetFromContext(ctx)

	if err := s.merchantRepo.UpdateBusinessProfileSettings(ctx, actor.MerchantId, settings); err != nil {
		return fmt.Errorf("error while updating business profile settings: %w", err)
	}

	return nil
}

func (s *Service) GetSchedulingSettings(ctx context.Context) (domain.SchedulingSettings, error) {
	actor := actor.MustGetFromContext(ctx)

	settings, err := s.merchantRepo.GetSchedulingSettings(ctx, actor.MerchantId)
	if err != nil {
		return domain.SchedulingSettings{}, err
	}

	return settings, nil
}

func (s *Service) UpdateSchedulingSettings(ctx context.Context, settings domain.SchedulingSettings) error {
	actor := actor.MustGetFromContext(ctx)

	if err := s.merchantRepo.UpdateSchedulingSettings(ctx, actor.MerchantId, settings); err != nil {
		return fmt.Errorf("error while updating scheduling settings: %w", err)
	}

	return nil
}

func (s *Service) GetBusinessHoursSettings(ctx context.Context) (domain.BusinessHours, error) {
	actor := actor.MustGetFromContext(ctx)

	businessHours, err := s.merchantRepo.GetBusinessHours(ctx, actor.MerchantId)
	if err != nil {
		return nil, err
	}

	return businessHours, nil
}

func (s *Service) UpdateBusinessHoursSettings(ctx context.Context, businessHours domain.BusinessHours) error {
	actor := actor.MustGetFromContext(ctx)

	err := s.txManager.WithTransaction(ctx, func(tx pgx.Tx) error {
		err := s.merchantRepo.WithTx(tx).DeleteOutdatedBusinessHours(ctx, actor.MerchantId, businessHours)
		if err != nil {
			return err
		}

		err = s.merchantRepo.WithTx(tx).NewBusinessHours(ctx, actor.MerchantId, businessHours)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("error while updating business hours: %w", err)
	}

	return nil
}

func (s *Service) GetNormalizedBusinessHours(ctx context.Context) (domain.BusinessHours, error) {
	actor := actor.MustGetFromContext(ctx)

	businessHours, err := s.merchantRepo.GetNormalizedBusinessHours(ctx, actor.MerchantId)
	if err != nil {
		return domain.BusinessHours{}, err
	}

	return businessHours, nil
}

type GetNormalizedBusinessHoursPublicInput struct {
	MerchantUrl string
	LocationId  int
}

func (s *Service) GetNormalizedBusinessHoursPublic(ctx context.Context, input GetNormalizedBusinessHoursPublicInput) (domain.BusinessHours, error) {
	merchantId, err := s.merchantRepo.GetMerchantIdByUrlName(ctx, input.MerchantUrl)
	if err != nil {
		return domain.BusinessHours{}, err
	}

	businessHours, err := s.merchantRepo.GetNormalizedBusinessHours(ctx, merchantId)
	if err != nil {
		return domain.BusinessHours{}, err
	}

	return businessHours, nil
}

func (s *Service) GetCalendarEvents(ctx context.Context, startDate, endDate time.Time, timezone *time.Location) (domain.CalendarEvents, error) {
	actor := actor.MustGetFromContext(ctx)

	dateRange := timeutil.NewDateRange(startDate, endDate, timezone)

	var events domain.CalendarEvents
	var err error

	events.Bookings, err = s.bookingRepo.GetBookingsForCalendar(ctx, actor.MerchantId, dateRange.StartTime, dateRange.EndTime)
	if err != nil {
		return domain.CalendarEvents{}, err
	}

	events.BlockedTimes, err = s.blockedTimeRepo.GetBlockedTimesForCalendar(ctx, actor.MerchantId,
		dateRange.StartTime, dateRange.EndTime, dateRange.StartDay, dateRange.EndDay)
	if err != nil {
		return domain.CalendarEvents{}, err
	}

	return events, nil
}

func (s *Service) GetBookingForCalendar(ctx context.Context, bookingId int) (domain.BookingForCalendar, error) {
	actor := actor.MustGetFromContext(ctx)

	booking, err := s.bookingRepo.GetBookingForCalendar(ctx, actor.MerchantId, bookingId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.BookingForCalendar{}, domain.ErrBookingNotFound
		}

		return domain.BookingForCalendar{}, err
	}

	return booking, nil
}
