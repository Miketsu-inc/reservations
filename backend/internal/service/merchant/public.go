package merchant

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/miketsu-inc/reservations/backend/internal/domain"
	"github.com/miketsu-inc/reservations/backend/internal/service/catalog"
	"github.com/miketsu-inc/reservations/backend/internal/types"
	"github.com/miketsu-inc/reservations/backend/pkg/apperr"
	"github.com/miketsu-inc/reservations/backend/pkg/timeutil"
)

func (s *Service) GetInfo(ctx context.Context, merchantName string) (domain.MerchantInfo, error) {
	merchantId, err := s.merchantRepo.GetMerchantIdByUrlName(ctx, strings.ToLower(merchantName))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.MerchantInfo{}, apperr.Wrap(ErrMerchantNotFound, err)
		}
		return domain.MerchantInfo{}, err
	}

	merchantInfo, err := s.merchantRepo.GetAllMerchantInfo(ctx, merchantId)
	if err != nil {
		return domain.MerchantInfo{}, err
	}

	merchantTz, err := time.LoadLocation(merchantInfo.Timezone)
	if err != nil {
		return domain.MerchantInfo{}, fmt.Errorf("invalid merchant timezone: %w", err)
	}

	now := time.Now().In(merchantTz)
	merchantInfo.BusinessHoursStatus = CalculateBusinessStatus(merchantInfo.BusinessHours, now)

	return merchantInfo, nil
}

func CalculateBusinessStatus(businessHours domain.BusinessHours, now time.Time) domain.BusinessHoursStatus {
	year, month, day := now.Date()
	today := int(now.Weekday())
	shiftsToday := businessHours[today]
	tz := now.Location()

	status := domain.BusinessHoursStatus{
		IsOpen: false,
	}

	for _, shift := range shiftsToday {
		businessStart := time.Date(year, month, day, shift.StartTime.Hour(), shift.StartTime.Minute(), 0, 0, tz)
		businessEnd := time.Date(year, month, day, shift.EndTime.Hour(), shift.EndTime.Minute(), 0, 0, tz)

		if (now.Equal(businessStart) || now.After(businessStart)) && now.Before(businessEnd) {
			status.IsOpen = true
			formattedCloseTime := shift.EndTime.Format("15:04")
			status.CloseTime = &formattedCloseTime
			break
		}
	}

	if !status.IsOpen {
		foundNextOpen := false

		for _, shift := range shiftsToday {
			businessStart := time.Date(year, month, day, shift.StartTime.Hour(), shift.StartTime.Minute(), 0, 0, tz)
			if now.Before(businessStart) {
				val := today
				status.NextOpenDay = &val
				foundNextOpen = true
				break
			}
		}

		if !foundNextOpen {
			for i := 1; i <= 7; i++ {
				nextDay := (today + i) % 7
				if len(businessHours[nextDay]) > 0 {
					val := nextDay
					status.NextOpenDay = &val
					break
				}
			}
		}
	}

	return status
}

type BookingSummary struct {
	MerchantName string
	Location     string
	Service      *domain.MinimalServiceInfo
	Employee     *domain.Employee
}

// this function should be optimized later
func (s *Service) GetSummary(ctx context.Context, merchantName string, locationId int, serviceId, employeeId *int) (BookingSummary, error) {
	merchantId, err := s.merchantRepo.GetMerchantIdByUrlName(ctx, strings.ToLower(merchantName))
	if err != nil {
		return BookingSummary{}, err
	}

	merchantName, location, err := s.merchantRepo.GetMerchantNameAndLocation(ctx, merchantId, locationId)
	if err != nil {
		return BookingSummary{}, err
	}

	summary := BookingSummary{
		MerchantName: merchantName,
		Location:     location,
	}

	if serviceId != nil {
		serviceInfo, err := s.catalogRepo.GetMinimalServiceInfo(ctx, merchantId, *serviceId, locationId, employeeId)
		if err != nil {
			return BookingSummary{}, err
		}

		serviceInfo.Price, serviceInfo.PriceType, err = catalog.BuildServicePricing(serviceInfo.Price, serviceInfo.PriceType, serviceInfo.PriceOverrides)
		if err != nil {
			return BookingSummary{}, err
		}
		summary.Service = &serviceInfo
	}

	if employeeId != nil {
		employee, err := s.teamRepo.GetEmployee(ctx, merchantId, *employeeId)
		if err != nil {
			return BookingSummary{}, err
		}
		summary.Employee = &employee
	}

	return summary, nil
}

type employeeAvailabilityData struct {
	EmpService    domain.Service
	ReservedTimes []domain.BookingSlot
	BlockedTimes  []domain.BlockedTimes
}

func (s *Service) fetchEmployeeAvailabilityData(ctx context.Context, serviceId int, employeeIds []int, merchantId uuid.UUID, locationId int, dateRange timeutil.DateRange) (map[int]employeeAvailabilityData, error) {
	empServices, err := s.catalogRepo.GetServiceWithPhasesForEmployees(ctx, serviceId, employeeIds)
	if err != nil {
		return nil, err
	}

	reservedByEmp, err := s.bookingRepo.GetReservedTimesByEmployees(ctx, merchantId, locationId, employeeIds, dateRange.StartTime, dateRange.EndTime)
	if err != nil {
		return nil, err
	}

	blockedByEmp, err := s.blockedTimeRepo.GetBlockedTimesByEmployees(ctx, merchantId, employeeIds, dateRange.StartTime, dateRange.EndTime, dateRange.StartDay, dateRange.EndDay)
	if err != nil {
		return nil, err
	}

	result := make(map[int]employeeAvailabilityData, len(employeeIds))
	for _, empId := range employeeIds {
		result[empId] = employeeAvailabilityData{
			EmpService:    empServices[empId],
			ReservedTimes: reservedByEmp[empId],
			BlockedTimes:  blockedByEmp[empId],
		}
	}

	return result, nil
}

func (s *Service) resolveEmployeeIds(ctx context.Context, serviceId int, employeeId *int) ([]int, error) {
	if employeeId != nil {
		return []int{*employeeId}, nil
	}

	employeeIds, err := s.catalogRepo.GetEmployeeIdsForService(ctx, serviceId)
	if err != nil {
		return nil, err
	}

	if len(employeeIds) == 0 {
		return nil, ErrEmployeeNotFound
	}

	return employeeIds, nil
}

func ResolveEmployeeBufferTime(empService domain.Service, defaultBufferTime int) int {
	if empService.BufferTime != nil {
		return *empService.BufferTime
	}
	return defaultBufferTime
}

type NextAvailable struct {
	FromDate            *time.Time
	ToDate              *time.Time
	CurrentParticipants *int
	Employee            *int
}

func (s *Service) GetNextAvailability(ctx context.Context, merchantName string, serviceId, locationId int) (NextAvailable, error) {
	merchantId, err := s.merchantRepo.GetMerchantIdByUrlName(ctx, strings.ToLower(merchantName))
	if err != nil {
		return NextAvailable{}, err
	}

	service, err := s.catalogRepo.GetServiceWithPhases(ctx, serviceId, merchantId)
	if err != nil {
		return NextAvailable{}, err
	}

	bookingSettings, err := s.merchantRepo.GetBookingSettingsByMerchantAndService(ctx, merchantId, service.Id)
	if err != nil {
		return NextAvailable{}, err
	}

	merchantTz, err := s.merchantRepo.GetMerchantTimezone(ctx, merchantId)
	if err != nil {
		return NextAvailable{}, err
	}

	now := time.Now().In(time.UTC)

	if service.BookingType == types.BookingTypeAppointment {
		startDate := now
		endDate := startDate.AddDate(0, 3, 0)
		dateRange := timeutil.NewDateRangeFromInclusiveEnd(startDate, endDate, merchantTz)

		businessHours, err := s.merchantRepo.GetBusinessHours(ctx, merchantId)
		if err != nil {
			return NextAvailable{}, err
		}

		employeeIds, err := s.catalogRepo.GetEmployeeIdsForService(ctx, service.Id)
		if err != nil {
			return NextAvailable{}, err
		}

		var na NextAvailable
		var earliestStart *time.Time
		var earliestEnd *time.Time
		var earliestEmpId *int

		empDataByEmp, err := s.fetchEmployeeAvailabilityData(ctx, service.Id, employeeIds, merchantId, locationId, dateRange)
		if err != nil {
			return NextAvailable{}, err
		}

		for _, empId := range employeeIds {
			empData := empDataByEmp[empId]

			empBufferTime := ResolveEmployeeBufferTime(empData.EmpService, bookingSettings.BufferTime)

			availableSlots := CalculateAvailableTimesPeriod(empData.ReservedTimes, empData.BlockedTimes, empData.EmpService.Phases, empData.EmpService.TotalDuration, empBufferTime, bookingSettings.BookingWindowMin, dateRange.StartTime, dateRange.EndTime, businessHours, now, merchantTz)

			for _, day := range availableSlots {
				var firstTimeStr string
				if len(day.Morning) > 0 {
					firstTimeStr = day.Morning[0]
				} else if len(day.Afternoon) > 0 {
					firstTimeStr = day.Afternoon[0]
				}

				if firstTimeStr != "" {
					timeString := fmt.Sprintf("%s %s", day.Date, firstTimeStr)
					parsedTime, err := time.ParseInLocation("2006-01-02 15:04", timeString, merchantTz)
					if err == nil {
						parsedTimeUTC := parsedTime.UTC()
						if earliestStart == nil || parsedTimeUTC.Before(*earliestStart) {
							earliestStart = &parsedTimeUTC
							endTimeUTC := parsedTimeUTC.Add(time.Duration(empData.EmpService.TotalDuration) * time.Minute)
							earliestEnd = &endTimeUTC
							currEmpId := empId
							earliestEmpId = &currEmpId
						}
					}
					break
				}
			}
		}

		if earliestStart != nil {
			na.FromDate = earliestStart
			na.ToDate = earliestEnd
			na.Employee = earliestEmpId
		}
		return na, nil

	} else {

		searchStart := now.Add(time.Duration(bookingSettings.BookingWindowMin) * time.Minute)
		searchEnd := now.AddDate(0, bookingSettings.BookingWindowMax, 0)

		booking, err := s.bookingRepo.GetClosestAvailableGroupBooking(ctx, merchantId, serviceId, locationId, searchStart, searchEnd)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return NextAvailable{}, nil
			}
			return NextAvailable{}, err
		}

		fromDateMechantTz := booking.FromDate.In(merchantTz)
		toDateMerchantTz := booking.ToDate.In(merchantTz)

		return NextAvailable{
			FromDate:            &fromDateMechantTz,
			ToDate:              &toDateMerchantTz,
			CurrentParticipants: &booking.CurrentParticipants,
			Employee:            booking.EmployeeId,
		}, nil
	}
}

type DisabledDays struct {
	ClosedDays []int
	MinDate    time.Time
	MaxDate    time.Time
}

// TODO: location id should be used later for location specific services/availability
func (s *Service) GetDisabledDays(ctx context.Context, merchantName string, serviceId, locationId int) (DisabledDays, error) {
	merchantId, err := s.merchantRepo.GetMerchantIdByUrlName(ctx, strings.ToLower(merchantName))
	if err != nil {
		return DisabledDays{}, err
	}

	bookingSettings, err := s.merchantRepo.GetBookingSettingsByMerchantAndService(ctx, merchantId, serviceId)
	if err != nil {
		return DisabledDays{}, err
	}

	merchantTz, err := s.merchantRepo.GetMerchantTimezone(ctx, merchantId)
	if err != nil {
		return DisabledDays{}, err
	}

	now := time.Now().In(merchantTz)

	minDate := now.Add(time.Duration(bookingSettings.BookingWindowMin) * time.Minute)
	maxDate := now.AddDate(0, bookingSettings.BookingWindowMax, 0)

	businessHours, err := s.merchantRepo.GetNormalizedBusinessHours(ctx, merchantId)
	if err != nil {
		return DisabledDays{}, err
	}

	closedDays := []int{}

	for i := 0; i <= 6; i++ {
		if _, ok := businessHours[i]; !ok {
			closedDays = append(closedDays, i)
		}
	}

	return DisabledDays{
		ClosedDays: closedDays,
		MinDate:    minDate,
		MaxDate:    maxDate,
	}, nil
}

func (s *Service) GetDayAvailability(ctx context.Context, merchantName string, serviceId int, locationId int, employeeId *int) ([]DayAvailability, error) {
	merchantId, err := s.merchantRepo.GetMerchantIdByUrlName(ctx, strings.ToLower(merchantName))
	if err != nil {
		return []DayAvailability{}, err
	}

	bookingSettings, err := s.merchantRepo.GetBookingSettingsByMerchantAndService(ctx, merchantId, serviceId)
	if err != nil {
		return []DayAvailability{}, err
	}

	employeeIds, err := s.resolveEmployeeIds(ctx, serviceId, employeeId)
	if err != nil {
		return []DayAvailability{}, err
	}

	merchantTz, err := s.merchantRepo.GetMerchantTimezone(ctx, merchantId)
	if err != nil {
		return []DayAvailability{}, err
	}

	businessHours, err := s.merchantRepo.GetBusinessHours(ctx, merchantId)
	if err != nil {
		return []DayAvailability{}, err
	}

	now := time.Now().In(time.UTC)
	startDate := now
	endDate := now.AddDate(0, bookingSettings.BookingWindowMax, 0)
	dateRange := timeutil.NewDateRangeFromInclusiveEnd(startDate, endDate, merchantTz)

	dayAvailabilityMap := make(map[string]bool)

	empDataByEmp, err := s.fetchEmployeeAvailabilityData(ctx, serviceId, employeeIds, merchantId, locationId, dateRange)
	if err != nil {
		return []DayAvailability{}, err
	}

	for _, empId := range employeeIds {
		empData := empDataByEmp[empId]

		if empData.EmpService.MerchantId != merchantId {
			return []DayAvailability{}, ErrMerchantServiceMismatch
		}

		if empData.EmpService.BookingType != types.BookingTypeAppointment {
			return []DayAvailability{}, ErrInvalidBookingType
		}

		empBufferTime := ResolveEmployeeBufferTime(empData.EmpService, bookingSettings.BufferTime)

		empDayAvailability := CalculateAvailableDays(empData.ReservedTimes, empData.BlockedTimes, empData.EmpService.Phases, empData.EmpService.TotalDuration, empBufferTime, bookingSettings.BookingWindowMin, dateRange.StartTime, dateRange.EndTime, businessHours, now, merchantTz)

		for _, day := range empDayAvailability {
			dayAvailabilityMap[day.Date] = dayAvailabilityMap[day.Date] || day.IsAvailable
		}
	}

	keys := make([]string, 0, len(dayAvailabilityMap))
	for k := range dayAvailabilityMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var finalAvailability []DayAvailability
	for _, key := range keys {
		finalAvailability = append(finalAvailability, DayAvailability{
			Date:        key,
			IsAvailable: dayAvailabilityMap[key],
		})
	}

	return finalAvailability, nil
}

func (s *Service) GetAvailabilityForDay(ctx context.Context, merchantName string, serviceId int, locationId int, employeeId *int, bookingDay time.Time) (FormattedAvailableTimes, error) {
	merchantId, err := s.merchantRepo.GetMerchantIdByUrlName(ctx, strings.ToLower(merchantName))
	if err != nil {
		return FormattedAvailableTimes{}, err
	}

	bookingSettings, err := s.merchantRepo.GetBookingSettingsByMerchantAndService(ctx, merchantId, serviceId)
	if err != nil {
		return FormattedAvailableTimes{}, err
	}

	employeeIds, err := s.resolveEmployeeIds(ctx, serviceId, employeeId)
	if err != nil {
		return FormattedAvailableTimes{}, err
	}

	merchantTz, err := s.merchantRepo.GetMerchantTimezone(ctx, merchantId)
	if err != nil {
		return FormattedAvailableTimes{}, err
	}

	businessHours, err := s.merchantRepo.GetBusinessHours(ctx, merchantId)
	if err != nil {
		return FormattedAvailableTimes{}, err
	}

	year, month, day := bookingDay.Date()
	bookingDay = time.Date(year, month, day, 0, 0, 0, 0, merchantTz)

	dayOfWeek := int(bookingDay.Weekday())
	bookingDayBusinessHours := businessHours[dayOfWeek]

	dateRange := timeutil.NewDateRange(bookingDay, bookingDay.AddDate(0, 0, 1), merchantTz)
	now := time.Now()

	uniqueAvailableTimesMap := make(map[string]time.Time)

	empDataByEmp, err := s.fetchEmployeeAvailabilityData(ctx, serviceId, employeeIds, merchantId, locationId, dateRange)
	if err != nil {
		return FormattedAvailableTimes{}, err
	}

	for _, empId := range employeeIds {
		empData := empDataByEmp[empId]

		if empData.EmpService.MerchantId != merchantId {
			return FormattedAvailableTimes{}, ErrMerchantServiceMismatch
		}

		if empData.EmpService.BookingType != types.BookingTypeAppointment {
			return FormattedAvailableTimes{}, ErrInvalidBookingType
		}

		empBufferTime := ResolveEmployeeBufferTime(empData.EmpService, bookingSettings.BufferTime)

		empAvailableTimes := CalculateAvailableTimes(empData.ReservedTimes, empData.BlockedTimes, empData.EmpService.Phases, empData.EmpService.TotalDuration, empBufferTime, bookingSettings.BookingWindowMin, bookingDay, bookingDayBusinessHours, now, merchantTz)

		for _, slot := range empAvailableTimes {
			key := slot.Format("15:04")
			if _, exists := uniqueAvailableTimesMap[key]; !exists {
				uniqueAvailableTimesMap[key] = slot
			}
		}
	}

	keys := make([]string, 0, len(uniqueAvailableTimesMap))
	for k := range uniqueAvailableTimesMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	finalAvailability := FormattedAvailableTimes{
		Morning:   []string{},
		Afternoon: []string{},
	}

	for _, key := range keys {
		if uniqueAvailableTimesMap[key].Hour() < 12 {
			finalAvailability.Morning = append(finalAvailability.Morning, key)
		} else {
			finalAvailability.Afternoon = append(finalAvailability.Afternoon, key)
		}
	}

	return finalAvailability, nil
}
