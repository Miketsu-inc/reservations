package merchant

import (
	"context"
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

func dashboardPeriod(utcDate time.Time, period int) (time.Time, time.Time, time.Time, error) {
	if period != 7 && period != 30 {
		return time.Time{}, time.Time{}, time.Time{}, fmt.Errorf("invalid period: %d", period)
	}

	// -1 because the last is the current day
	currPeriodStart := utils.TruncateToDay(utcDate.AddDate(0, 0, -(period - 1)))
	prevPeriodStart := utils.TruncateToDay(currPeriodStart.AddDate(0, 0, -period))

	return currPeriodStart, utils.TruncateToDay(utcDate), prevPeriodStart, nil
}

func (s *Service) GetDashboardStatistics(ctx context.Context, period int) (domain.DashboardStatistics, error) {
	utcDate := time.Now().UTC()
	currPeriodStart, _, prevPeriodStart, err := dashboardPeriod(utcDate, period)
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
	currPeriodStart, periodEnd, _, err := dashboardPeriod(utcDate, period)
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

func (s *Service) GetTeamForCalendar(ctx context.Context) ([]domain.Employee, error) {
	actor := actor.MustGetFromContext(ctx)

	team, err := s.teamRepo.GetEmployees(ctx, actor.MerchantId)
	if err != nil {
		return []domain.Employee{}, err
	}

	return team, nil
}

func (s *Service) GetServicesForCalendar(ctx context.Context) ([]domain.ServicesGroupedByCategoriesForCalendar, error) {
	actor := actor.MustGetFromContext(ctx)

	services, err := s.catalogRepo.GetServicesForCalendar(ctx, actor.MerchantId)
	if err != nil {
		return []domain.ServicesGroupedByCategoriesForCalendar{}, err
	}

	return services, nil
}

func (s *Service) GetCustomersForCalendar(ctx context.Context) ([]domain.CustomerForCalendar, error) {
	actor := actor.MustGetFromContext(ctx)

	customers, err := s.customerRepo.GetCustomersForCalendar(ctx, actor.MerchantId)
	if err != nil {
		return []domain.CustomerForCalendar{}, err
	}

	return customers, nil
}

func (s *Service) GetCalendarEvents(ctx context.Context, start string, end string) (domain.CalendarEvents, error) {
	actor := actor.MustGetFromContext(ctx)

	var events domain.CalendarEvents
	var err error

	events.Bookings, err = s.bookingRepo.GetBookingsForCalendar(ctx, actor.MerchantId, start, end)
	if err != nil {
		return domain.CalendarEvents{}, err
	}

	events.BlockedTimes, err = s.blockedTimeRepo.GetBlockedTimesForCalendar(ctx, actor.MerchantId, start, end)
	if err != nil {
		return domain.CalendarEvents{}, err
	}

	return events, nil
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
