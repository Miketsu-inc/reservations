package merchant

import (
	"context"
	"fmt"
	"time"

	"github.com/miketsu-inc/reservations/backend/internal/api/middleware/jwt"
	"github.com/miketsu-inc/reservations/backend/internal/domain"
	"github.com/miketsu-inc/reservations/backend/internal/types"
	"github.com/miketsu-inc/reservations/backend/pkg/validate"
)

type Service struct {
	bookingRepo     domain.BookingRepository
	catalogRepo     domain.CatalogRepository
	merchantRepo    domain.MerchantRepository
	customerRepo    domain.CustomerRepository
	blockedTimeRepo domain.BlockedTimeRepository
	teamRepo        domain.TeamRepository
}

func NewService(booking domain.BookingRepository, catalog domain.CatalogRepository, merchant domain.MerchantRepository,
	customer domain.CustomerRepository, blockedTime domain.BlockedTimeRepository,
	team domain.TeamRepository) *Service {
	return &Service{
		bookingRepo:     booking,
		catalogRepo:     catalog,
		merchantRepo:    merchant,
		customerRepo:    customer,
		blockedTimeRepo: blockedTime,
		teamRepo:        team,
	}
}

func (s *Service) Delete(ctx context.Context) error {
	employee := jwt.MustGetEmployeeFromContext(ctx)

	err := s.merchantRepo.DeleteMerchant(ctx, employee.Id, employee.MerchantId)
	if err != nil {
		return fmt.Errorf("error while deleting merchant: %s", err.Error())
	}

	return nil
}

type UpdateNameInput struct {
	Name string
}

func (s *Service) UpdateName(ctx context.Context, input UpdateNameInput) error {
	urlName, err := validate.MerchantNameToUrlName(input.Name)
	if err != nil {
		return fmt.Errorf("unexpected error during merchant url name conversion: %s", err.Error())
	}

	unique, err := s.merchantRepo.IsMerchantUrlUnique(ctx, urlName)
	if err != nil {
		return err
	}

	if !unique {
		return ErrMerchantUrlNotUnique{URL: urlName}
	}

	employee := jwt.MustGetEmployeeFromContext(ctx)

	err = s.merchantRepo.ChangeMerchantNameAndURL(ctx, employee.MerchantId, input.Name, urlName)
	if err != nil {
		return fmt.Errorf("error while updating merchant's name: %s", err.Error())
	}

	return nil
}

func (s *Service) GetDashboard(ctx context.Context, date time.Time, period int) (domain.DashboardData, error) {
	if period != 7 && period != 30 {
		return domain.DashboardData{}, fmt.Errorf("invalid period: %d", period)
	}

	employee := jwt.MustGetEmployeeFromContext(ctx)

	dashboard, err := s.merchantRepo.GetDashboardData(ctx, employee.MerchantId, date, period)
	if err != nil {
		return domain.DashboardData{}, fmt.Errorf("error while retrieving dashboard data: %s", err.Error())
	}

	return dashboard, nil
}

type ErrMerchantUrlNotUnique struct {
	URL string
}

func (e ErrMerchantUrlNotUnique) Error() string {
	return "this merchant url is already used"
}

type CheckUrlInput struct {
	Name string
}

func (s *Service) CheckUrl(ctx context.Context, input CheckUrlInput) (CheckUrlInput, error) {
	urlName, err := validate.MerchantNameToUrlName(input.Name)
	if err != nil {
		return CheckUrlInput{}, fmt.Errorf("unexpected error during merchant url name conversion: %s", err.Error())
	}

	unique, err := s.merchantRepo.IsMerchantUrlUnique(ctx, urlName)
	if err != nil {
		return CheckUrlInput{Name: urlName}, err
	}

	if !unique {
		return CheckUrlInput{Name: urlName}, ErrMerchantUrlNotUnique{URL: urlName}
	}

	return CheckUrlInput{Name: urlName}, nil
}

func (s *Service) GetSettings(ctx context.Context) (domain.MerchantSettingsInfo, error) {
	employee := jwt.MustGetEmployeeFromContext(ctx)

	settings, err := s.merchantRepo.GetMerchantSettingsInfo(ctx, employee.MerchantId)
	if err != nil {
		return domain.MerchantSettingsInfo{}, fmt.Errorf("error while accessing settings merchant info: %s", err.Error())
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
	BusinessHours    map[int][]domain.TimeSlot
}

func (s *Service) UpdateSettings(ctx context.Context, input UpdateSettingsInput) error {
	employee := jwt.MustGetEmployeeFromContext(ctx)

	err := s.merchantRepo.UpdateMerchantFieldsById(ctx, employee.MerchantId, domain.MerchantSettingFields{
		Introduction:     input.Introduction,
		Announcement:     input.Announcement,
		AboutUs:          input.AboutUs,
		ParkingInfo:      input.ParkingInfo,
		PaymentInfo:      input.PaymentInfo,
		CancelDeadline:   input.CancelDeadline,
		BookingWindowMin: input.BookingWindowMin,
		BookingWindowMax: input.BookingWindowMax,
		BufferTime:       input.BufferTime,
		BusinessHours:    input.BusinessHours,
	})
	if err != nil {
		return fmt.Errorf("error while updating reservation fileds for merchant: %s", err.Error())
	}

	return nil
}

func (s *Service) GetNormalizedBusinessHours(ctx context.Context) (map[int]domain.TimeSlot, error) {
	employee := jwt.MustGetEmployeeFromContext(ctx)

	businessHours, err := s.merchantRepo.GetNormalizedBusinessHours(ctx, employee.MerchantId)
	if err != nil {
		return map[int]domain.TimeSlot{}, fmt.Errorf("error while retrieving business hours by merchant id: %s", err.Error())
	}

	return businessHours, nil
}

func (s *Service) GetPreferences(ctx context.Context) (domain.PreferenceData, error) {
	employee := jwt.MustGetEmployeeFromContext(ctx)

	preferences, err := s.merchantRepo.GetPreferencesByMerchantId(ctx, employee.MerchantId)
	if err != nil {
		return domain.PreferenceData{}, fmt.Errorf("error while accessing merchant preferences: %s", err.Error())
	}

	return preferences, nil
}

type UpdatePreferencesInput struct {
	FirstDayOfWeek     string
	TimeFormat         string
	CalendarView       string
	CalendarViewMobile string
	StartHour          domain.TimeString
	EndHour            domain.TimeString
	TimeFrequency      domain.TimeString
}

func (s *Service) UpdatePreferences(ctx context.Context, input UpdatePreferencesInput) error {
	employee := jwt.MustGetEmployeeFromContext(ctx)

	err := s.merchantRepo.UpdatePreferences(ctx, employee.MerchantId, domain.PreferenceData{
		FirstDayOfWeek:     input.FirstDayOfWeek,
		TimeFormat:         input.TimeFormat,
		CalendarView:       input.CalendarView,
		CalendarViewMobile: input.CalendarViewMobile,
		StartHour:          input.StartHour,
		EndHour:            input.EndHour,
		TimeFrequency:      input.TimeFrequency,
	})
	if err != nil {
		return fmt.Errorf("error while updating preferences: %s", err.Error())
	}

	return nil
}

func (s *Service) GetTeamMembersForCalendar(ctx context.Context) ([]domain.EmployeeForCalendar, error) {
	employee := jwt.MustGetEmployeeFromContext(ctx)

	teamMember, err := s.teamRepo.GetEmployeesForCalendarByMerchant(ctx, employee.MerchantId)
	if err != nil {
		return []domain.EmployeeForCalendar{}, fmt.Errorf("error while retrieving employees for merchant: %s", err.Error())
	}

	return teamMember, nil
}

func (s *Service) GetServicesForCalendar(ctx context.Context) ([]domain.ServicesGroupedByCategoriesForCalendar, error) {
	employee := jwt.MustGetEmployeeFromContext(ctx)

	services, err := s.catalogRepo.GetServicesForCalendarByMerchant(ctx, employee.MerchantId)
	if err != nil {
		return []domain.ServicesGroupedByCategoriesForCalendar{}, fmt.Errorf("error while retrieving services for merchant: %s", err.Error())
	}

	return services, nil
}

func (s *Service) GetCustomersForCalendar(ctx context.Context) ([]domain.CustomerForCalendar, error) {
	employee := jwt.MustGetEmployeeFromContext(ctx)

	customers, err := s.customerRepo.GetCustomersForCalendarByMerchant(ctx, employee.MerchantId)
	if err != nil {
		return []domain.CustomerForCalendar{}, fmt.Errorf("error while retrieving customers for merchant: %s", err.Error())
	}

	return customers, nil
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
	employee := jwt.MustGetEmployeeFromContext(ctx)

	err := s.merchantRepo.NewLocation(ctx, domain.Location{
		MerchantId:        employee.MerchantId,
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
		return fmt.Errorf("unexpected error during adding location to database: %s", err.Error())
	}

	return nil
}
