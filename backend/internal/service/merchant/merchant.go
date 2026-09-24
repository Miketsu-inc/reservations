package merchant

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/miketsu-inc/reservations/backend/internal/api/middleware/actor"
	"github.com/miketsu-inc/reservations/backend/internal/domain"
	"github.com/miketsu-inc/reservations/backend/internal/types"
	"github.com/miketsu-inc/reservations/backend/internal/utils"
	"github.com/miketsu-inc/reservations/backend/pkg/apperr"
	"github.com/miketsu-inc/reservations/backend/pkg/db"
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

func (s *Service) GetCalendarEvents(ctx context.Context, start string, end string) (domain.CalendarEvents, error) {
	actor := actor.MustGetFromContext(ctx)
	merchantTz, err := s.merchantRepo.GetMerchantTimezone(ctx, actor.MerchantId)
	if err != nil {
		return domain.CalendarEvents{}, err
	}

	startDate, err := time.Parse(time.DateOnly, start)
	if err != nil {
		return domain.CalendarEvents{}, fmt.Errorf("invalid calendar start date: %w", err)
	}

	endDate, err := time.Parse(time.DateOnly, end)
	if err != nil {
		return domain.CalendarEvents{}, fmt.Errorf("invalid calendar end date: %w", err)
	}

	if !endDate.After(startDate) {
		return domain.CalendarEvents{}, fmt.Errorf("calendar end date must be after start date")
	}

	startYear, startMonth, startDay := startDate.Date()
	endYear, endMonth, endDay := endDate.Date()
	startTime := time.Date(startYear, startMonth, startDay, 0, 0, 0, 0, merchantTz).UTC()
	endTime := time.Date(endYear, endMonth, endDay, 0, 0, 0, 0, merchantTz).UTC()

	var events domain.CalendarEvents

	events.Bookings, err = s.bookingRepo.GetBookingsForCalendar(ctx, actor.MerchantId, startTime, endTime)
	if err != nil {
		return domain.CalendarEvents{}, err
	}

	events.BlockedTimes, err = s.blockedTimeRepo.GetBlockedTimesForCalendar(ctx, actor.MerchantId, startDate, endDate, startTime, endTime)
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
