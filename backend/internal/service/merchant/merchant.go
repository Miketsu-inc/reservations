package merchant

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/bojanz/currency"
	"github.com/jackc/pgx/v5"
	"github.com/miketsu-inc/reservations/backend/internal/api/middleware/actor"
	"github.com/miketsu-inc/reservations/backend/internal/domain"
	"github.com/miketsu-inc/reservations/backend/internal/types"
	"github.com/miketsu-inc/reservations/backend/pkg/apperr"
	"github.com/miketsu-inc/reservations/backend/pkg/currencyx"
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

func getPeriodDates(now time.Time, period int, timezone *time.Location) (time.Time, time.Time, time.Time) {
	localNow := now.In(timezone)
	periodEnd := time.Date(localNow.Year(), localNow.Month(), localNow.Day(), 0, 0, 0, 0, timezone)

	// -1 because the last date is the current date in the requested timezone.
	currPeriodStart := periodEnd.AddDate(0, 0, -(period - 1))
	prevPeriodStart := currPeriodStart.AddDate(0, 0, -period)

	return currPeriodStart, periodEnd, prevPeriodStart
}

func fillRevenueMissingDays(revenue []domain.RevenueStat, periodStart time.Time, periodEnd time.Time,
	currencyCode string) ([]domain.RevenueStat, error) {

	zeroAmount, err := currency.NewAmount("0", currencyCode)
	if err != nil {
		return nil, fmt.Errorf("format dashboard revenue zero: %w", err)
	}

	revenueByDay := make(map[string]domain.RevenueStat, len(revenue))
	for _, stat := range revenue {
		revenueByDay[stat.Day.Format(time.DateOnly)] = stat
	}

	completeRevenue := make([]domain.RevenueStat, 0)
	for day := periodStart; !day.After(periodEnd); day = day.AddDate(0, 0, 1) {
		stat, ok := revenueByDay[day.Format(time.DateOnly)]
		if !ok {
			stat = domain.RevenueStat{Value: currencyx.Price{Amount: zeroAmount}}
		}

		stat.Day = day
		completeRevenue = append(completeRevenue, stat)
	}

	return completeRevenue, nil
}

type GetDashboardStatisticsResult struct {
	PeriodStart time.Time
	PeriodEnd   time.Time
	Statistics  domain.DashboardStatistics
}

func (s *Service) GetDashboardStatistics(ctx context.Context, period int, timezone *time.Location) (GetDashboardStatisticsResult, error) {
	actor := actor.MustGetFromContext(ctx)

	now := time.Now()
	currPeriodStart, periodEnd, prevPeriodStart := getPeriodDates(now, period, timezone)

	statistics, err := s.merchantRepo.GetDashboardStats(ctx, actor.MerchantId, currPeriodStart.UTC(), now.UTC(), prevPeriodStart.UTC())
	if err != nil {
		return GetDashboardStatisticsResult{}, err
	}

	return GetDashboardStatisticsResult{
		PeriodStart: currPeriodStart,
		PeriodEnd:   periodEnd,
		Statistics:  statistics,
	}, nil
}

func (s *Service) GetDashboardRevenue(ctx context.Context, period int, timezone *time.Location) (domain.DashboardRevenue, error) {
	actor := actor.MustGetFromContext(ctx)

	now := time.Now()
	currPeriodStart, periodEnd, _ := getPeriodDates(now, period, timezone)

	revenue, err := s.merchantRepo.GetRevenueStats(ctx, actor.MerchantId, currPeriodStart.UTC(), now.UTC(), timezone)
	if err != nil {
		return domain.DashboardRevenue{}, err
	}

	currencyCode := ""
	if len(revenue) == 0 {
		currencyCode, err = s.merchantRepo.GetMerchantCurrency(ctx, actor.MerchantId)
		if err != nil {
			return domain.DashboardRevenue{}, err
		}
	} else {
		currencyCode = revenue[0].Value.CurrencyCode()
	}

	completeRevenue, err := fillRevenueMissingDays(revenue, currPeriodStart, periodEnd, currencyCode)
	if err != nil {
		return domain.DashboardRevenue{}, err
	}

	return domain.DashboardRevenue{
		PeriodStart: currPeriodStart,
		PeriodEnd:   periodEnd,
		Revenue:     completeRevenue,
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

func (s *Service) GetSettings(ctx context.Context) (domain.MerchantSettingsInfo, error) {
	actor := actor.MustGetFromContext(ctx)

	settings, err := s.merchantRepo.GetMerchantSettingsInfo(ctx, actor.MerchantId)
	if err != nil {
		return domain.MerchantSettingsInfo{}, err
	}

	return settings, nil
}

type UpdateSettingsInput struct {
	Introduction     string
	Announcement     string
	AboutUs          string
	ParkingInfo      string
	PaymentInfo      string
	CancelDeadline   int
	BookingWindowMin int
	BookingWindowMax int
	BufferTime       int
	ApprovalPolicy   types.ApprovalType
	BusinessHours    domain.BusinessHours
}

func (s *Service) UpdateSettings(ctx context.Context, input UpdateSettingsInput) error {
	actor := actor.MustGetFromContext(ctx)

	err := s.txManager.WithTransaction(ctx, func(tx pgx.Tx) error {
		err := s.merchantRepo.WithTx(tx).UpdateMerchantFields(ctx, actor.MerchantId, domain.MerchantSettingFields{
			Introduction:     input.Introduction,
			Announcement:     input.Announcement,
			AboutUs:          input.AboutUs,
			ParkingInfo:      input.ParkingInfo,
			PaymentInfo:      input.PaymentInfo,
			CancelDeadline:   input.CancelDeadline,
			BookingWindowMin: input.BookingWindowMin,
			BookingWindowMax: input.BookingWindowMax,
			BufferTime:       input.BufferTime,
			ApprovalPolicy:   input.ApprovalPolicy,
		})
		if err != nil {
			return err
		}

		err = s.merchantRepo.WithTx(tx).DeleteOutdatedBusinessHours(ctx, actor.MerchantId, input.BusinessHours)
		if err != nil {
			return err
		}

		err = s.merchantRepo.WithTx(tx).NewBusinessHours(ctx, actor.MerchantId, input.BusinessHours)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("error while updating reservation fileds for merchant: %s", err.Error())
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

type NewLocationInput struct {
	Country           *string
	City              *string
	PostalCode        *string
	Address           *string
	GeoPoint          types.GeoPoint
	PlaceId           *string
	FormattedLocation string
	IsPrimary         bool
	IsActive          bool
}

func (s *Service) NewLocation(ctx context.Context, req NewLocationInput) error {
	actor := actor.MustGetFromContext(ctx)

	err := s.merchantRepo.NewLocation(ctx, domain.Location{
		MerchantId:        actor.MerchantId,
		Country:           req.Country,
		City:              req.City,
		PostalCode:        req.PostalCode,
		Address:           req.Address,
		GeoPoint:          req.GeoPoint,
		PlaceId:           req.PlaceId,
		FormattedLocation: req.FormattedLocation,
		IsPrimary:         req.IsPrimary,
		IsActive:          req.IsActive,
	})
	if err != nil {
		return err
	}

	return nil
}
