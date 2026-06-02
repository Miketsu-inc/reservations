package merchant

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/miketsu-inc/reservations/backend/internal/domain"
	"github.com/miketsu-inc/reservations/backend/internal/types"
)

func (s *Service) GetInfo(ctx context.Context, merchantName string) (domain.MerchantInfo, error) {
	merchantId, err := s.merchantRepo.GetMerchantIdByUrlName(ctx, strings.ToLower(merchantName))
	if err != nil {
		return domain.MerchantInfo{}, fmt.Errorf("error while retrieving the merchant's id: %s", err.Error())
	}

	merchantInfo, err := s.merchantRepo.GetAllMerchantInfo(ctx, merchantId)
	if err != nil {
		return domain.MerchantInfo{}, fmt.Errorf("error while accessing merchant info: %s", err.Error())
	}

	merchantInfo.BusinessHoursStatus = CalculateBusinessStatus(merchantInfo.BusinessHours)

	return merchantInfo, nil
}

func CalculateBusinessStatus(businessHours domain.BusinessHours) domain.BusinessHoursStatus {
	now := time.Now().In(time.UTC)
	year, month, day := now.Date()
	today := int(now.Weekday())
	shiftsToday := businessHours[today]

	status := domain.BusinessHoursStatus{
		IsOpen: false,
	}

	for _, shift := range shiftsToday {
		businessStart := time.Date(year, month, day, shift.StartTime.Hour(), shift.StartTime.Minute(), 0, 0, time.UTC)
		businessEnd := time.Date(year, month, day, shift.EndTime.Hour(), shift.EndTime.Minute(), 0, 0, time.UTC)

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
			businessStart := time.Date(year, month, day, shift.StartTime.Hour(), shift.StartTime.Minute(), 0, 0, time.UTC)
			if now.Before(businessStart) {
				val := today
				status.NextOpenDay = &val
				foundNextOpen = true
				break
			}
		}

		if !foundNextOpen {
			for i := 0; i <= 6; i++ {
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

func (s *Service) GetServicesGroupedByCategories(ctx context.Context, merchantName string) ([]domain.MerchantPageServicesGroupedByCategory, error) {
	merchantId, err := s.merchantRepo.GetMerchantIdByUrlName(ctx, strings.ToLower(merchantName))
	if err != nil {
		return []domain.MerchantPageServicesGroupedByCategory{}, fmt.Errorf("error while retrieving the merchant's id: %s", err.Error())
	}

	services, err := s.catalogRepo.GetServicesForMerchantPage(ctx, merchantId)
	if err != nil {
		return []domain.MerchantPageServicesGroupedByCategory{}, fmt.Errorf("error while getting service for the merchant: %s", err.Error())
	}

	return services, nil

}

func (s *Service) GetTeam(ctx context.Context, merchantName string) ([]domain.PublicEmployee, error) {
	merchantId, err := s.merchantRepo.GetMerchantIdByUrlName(ctx, strings.ToLower(merchantName))
	if err != nil {
		return []domain.PublicEmployee{}, fmt.Errorf("error while retrieving the merchant's id: %s", err.Error())
	}

	employees, err := s.teamRepo.GetActiveEmployees(ctx, merchantId)
	if err != nil {
		return []domain.PublicEmployee{}, fmt.Errorf("error while getting employees for merchant: %s", err.Error())
	}

	return employees, nil
}

func (s *Service) GetServiceDetails(ctx context.Context, merchantName string, serviceId, locationId int) (domain.PublicServiceDetails, error) {
	merchantId, err := s.merchantRepo.GetMerchantIdByUrlName(ctx, strings.ToLower(merchantName))
	if err != nil {
		return domain.PublicServiceDetails{}, fmt.Errorf("error while retrieving the merchant's id: %s", err.Error())
	}

	serviceDetails, err := s.catalogRepo.GetServiceDetailsForMerchantPage(ctx, merchantId, serviceId, locationId)
	if err != nil {
		return domain.PublicServiceDetails{}, fmt.Errorf("error while retrieving service info: %s", err.Error())
	}

	return serviceDetails, nil
}

func (s *Service) GetSummary(ctx context.Context, merchantName string, serviceId, locationId int) (domain.MinimalServiceInfo, error) {
	merchantId, err := s.merchantRepo.GetMerchantIdByUrlName(ctx, strings.ToLower(merchantName))
	if err != nil {
		return domain.MinimalServiceInfo{}, fmt.Errorf("error while retrieving the merchant's id: %s", err.Error())
	}

	serviceInfo, err := s.catalogRepo.GetMinimalServiceInfo(ctx, merchantId, serviceId, locationId)
	if err != nil {
		return domain.MinimalServiceInfo{}, fmt.Errorf("error while retrieving minimal service info: %s", err.Error())
	}

	return serviceInfo, nil
}

func (s *Service) GetAvailability(ctx context.Context, merchantName string, serviceId, locationId int, startDate, endDate time.Time) ([]MultiDayAvailableTimes, error) {
	merchantId, err := s.merchantRepo.GetMerchantIdByUrlName(ctx, strings.ToLower(merchantName))
	if err != nil {
		return []MultiDayAvailableTimes{}, fmt.Errorf("error while retrieving the merchant's id: %s", err.Error())
	}

	service, err := s.catalogRepo.GetServiceWithPhases(ctx, serviceId, merchantId)
	if err != nil {
		return []MultiDayAvailableTimes{}, fmt.Errorf("error while retrieving service: %s", err.Error())
	}

	if service.MerchantId != merchantId {
		return []MultiDayAvailableTimes{}, fmt.Errorf("this service id does not belong to this merchant")
	}

	merchantTz, err := s.merchantRepo.GetMerchantTimezone(ctx, merchantId)
	if err != nil {
		return []MultiDayAvailableTimes{}, fmt.Errorf("error while getting merchant's timezone: %s", err.Error())
	}

	var availableSlots []MultiDayAvailableTimes

	startDate = startDate.UTC()
	endDate = endDate.UTC()

	if service.BookingType == types.BookingTypeAppointment {

		bookingSettings, err := s.merchantRepo.GetBookingSettingsByMerchantAndService(ctx, merchantId, service.Id)
		if err != nil {
			return []MultiDayAvailableTimes{}, fmt.Errorf("error while getting booking settings for merchant: %s", err.Error())
		}

		reservedTimes, err := s.bookingRepo.GetReservedTimesForPeriod(ctx, merchantId, locationId, startDate, endDate)
		if err != nil {
			return []MultiDayAvailableTimes{}, fmt.Errorf("error while calculating available time slots: %s", err.Error())
		}

		blockedTimes, err := s.blockedTimeRepo.GetBlockedTimes(ctx, merchantId, startDate, endDate)
		if err != nil {
			return []MultiDayAvailableTimes{}, fmt.Errorf("error while getting blocked times for merchant: %s", err.Error())
		}

		businessHours, err := s.merchantRepo.GetBusinessHours(ctx, merchantId)
		if err != nil {
			return []MultiDayAvailableTimes{}, fmt.Errorf("error while getting business hours: %s", err.Error())
		}

		now := time.Now()
		availableSlots = CalculateAvailableTimesPeriod(reservedTimes, blockedTimes, service.Phases, service.TotalDuration, bookingSettings.BufferTime, bookingSettings.BookingWindowMin, startDate, endDate, businessHours, now, merchantTz)

	} else {

		groupBookings, err := s.bookingRepo.GetAvailableGroupBookingsForPeriod(ctx, merchantId, serviceId, locationId, startDate, endDate)
		if err != nil {
			return []MultiDayAvailableTimes{}, fmt.Errorf("error while getting available group bookings for period: %s", err.Error())
		}

		bookingsByDate := make(map[string][]time.Time)
		for _, b := range groupBookings {
			fromDate := b.FromDate.In(merchantTz)
			date := fromDate.Format("2006-01-02")

			bookingsByDate[date] = append(bookingsByDate[date], fromDate)
		}

		for d := startDate.In(merchantTz); !d.After(endDate.In(merchantTz)); d = d.AddDate(0, 0, 1) {
			date := d.Format("2006-01-02")

			var morning []string
			var afternoon []string

			times, ok := bookingsByDate[date]
			if ok {
				for _, t := range times {
					formattedTime := fmt.Sprintf("%02d:%02d", t.Hour(), t.Minute())

					if t.Hour() < 12 {
						morning = append(morning, formattedTime)
					} else if t.Hour() >= 12 {
						afternoon = append(afternoon, formattedTime)
					}
				}
			}

			isAvailable := len(morning) > 0 || len(afternoon) > 0

			availableSlots = append(availableSlots, MultiDayAvailableTimes{
				Date:        date,
				IsAvailable: isAvailable,
				Morning:     morning,
				Afternoon:   afternoon,
			})
		}
	}

	return availableSlots, nil
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
		return NextAvailable{}, fmt.Errorf("error while retrieving the merchant's id: %s", err.Error())
	}

	service, err := s.catalogRepo.GetServiceWithPhases(ctx, serviceId, merchantId)
	if err != nil {
		return NextAvailable{}, fmt.Errorf("error while retrieving service: %s", err.Error())
	}

	bookingSettings, err := s.merchantRepo.GetBookingSettingsByMerchantAndService(ctx, merchantId, service.Id)
	if err != nil {
		return NextAvailable{}, fmt.Errorf("error while getting booking setting for merchant: %s", err.Error())
	}

	merchantTz, err := s.merchantRepo.GetMerchantTimezone(ctx, merchantId)
	if err != nil {
		return NextAvailable{}, fmt.Errorf("error while getting merchant's timezone: %s", err.Error())
	}

	now := time.Now().In(time.UTC)

	if service.BookingType == types.BookingTypeAppointment {

		startDate := now
		endDate := startDate.AddDate(0, 3, 0)

		reservedTimes, err := s.bookingRepo.GetReservedTimesForPeriod(ctx, merchantId, locationId, startDate, endDate)
		if err != nil {
			return NextAvailable{}, fmt.Errorf("error while calculating available time slots: %s", err.Error())
		}

		blockedTimes, err := s.blockedTimeRepo.GetBlockedTimes(ctx, merchantId, startDate, endDate)
		if err != nil {
			return NextAvailable{}, fmt.Errorf("error while getting blocked times for merchant: %s", err.Error())
		}

		businessHours, err := s.merchantRepo.GetBusinessHours(ctx, merchantId)
		if err != nil {
			return NextAvailable{}, fmt.Errorf("error while getting business hours: %s", err.Error())
		}

		availableSlots := CalculateAvailableTimesPeriod(reservedTimes, blockedTimes, service.Phases, service.TotalDuration, bookingSettings.BufferTime, bookingSettings.BookingWindowMin, startDate, endDate, businessHours, now, merchantTz)

		var na NextAvailable
		var dateStr, timeStr string

		for _, day := range availableSlots {
			if len(day.Morning) > 0 {
				dateStr, timeStr = day.Date, day.Morning[0]
				break
			}
			if len(day.Afternoon) > 0 {
				dateStr, timeStr = day.Date, day.Afternoon[0]
				break
			}
		}

		if dateStr != "" && timeStr != "" {
			timeString := fmt.Sprintf("%s %s", dateStr, timeStr)
			parsedTime, err := time.ParseInLocation("2006-01-02 15:04", timeString, merchantTz)
			if err == nil {
				na.FromDate = &parsedTime
			}
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
			return NextAvailable{}, fmt.Errorf("error finding group booking: %w", err)
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
		return DisabledDays{}, fmt.Errorf("error while retrieving the merchant's id: %s", err.Error())
	}

	bookingSettings, err := s.merchantRepo.GetBookingSettingsByMerchantAndService(ctx, merchantId, serviceId)
	if err != nil {
		return DisabledDays{}, fmt.Errorf("error while retrieving booking settings by merchant id: %s", err.Error())
	}

	merchantTz, err := s.merchantRepo.GetMerchantTimezone(ctx, merchantId)
	if err != nil {
		return DisabledDays{}, fmt.Errorf("error while getting merchant's timezone: %s", err.Error())
	}

	now := time.Now().In(merchantTz)

	minDate := now.Add(time.Duration(bookingSettings.BookingWindowMin) * time.Minute)
	maxDate := now.AddDate(0, bookingSettings.BookingWindowMax, 0)

	businessHours, err := s.merchantRepo.GetNormalizedBusinessHours(ctx, merchantId)
	if err != nil {
		return DisabledDays{}, fmt.Errorf("error while retrieving business hours by merchant id: %s", err.Error())
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
