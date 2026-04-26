package merchant

import (
	"context"
	"fmt"
	"strings"
	"time"

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

	return merchantInfo, nil
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
	Date string
	Time string
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

	startDate := time.Now().In(time.UTC)
	endDate := startDate.AddDate(0, 3, 0)

	reservedTimes, err := s.bookingRepo.GetReservedTimesForPeriod(ctx, merchantId, locationId, startDate, endDate)
	if err != nil {
		return NextAvailable{}, fmt.Errorf("error while calculating available time slots: %s", err.Error())
	}

	blockedTimes, err := s.blockedTimeRepo.GetBlockedTimes(ctx, merchantId, startDate, endDate)
	if err != nil {
		return NextAvailable{}, fmt.Errorf("error while getting blocked times for merchant: %s", err.Error())
	}

	merchantTz, err := s.merchantRepo.GetMerchantTimezone(ctx, merchantId)
	if err != nil {
		return NextAvailable{}, fmt.Errorf("error while getting merchant's timezone: %s", err.Error())
	}

	businessHours, err := s.merchantRepo.GetBusinessHours(ctx, merchantId)
	if err != nil {
		return NextAvailable{}, fmt.Errorf("error while getting business hours: %s", err.Error())
	}

	now := time.Now()
	availableSlots := CalculateAvailableTimesPeriod(reservedTimes, blockedTimes, service.Phases, service.TotalDuration, bookingSettings.BufferTime, bookingSettings.BookingWindowMin, startDate, endDate, businessHours, now, merchantTz)

	var na NextAvailable

	for _, day := range availableSlots {
		if len(day.Morning) > 0 {
			na.Time = day.Morning[0]
			na.Date = day.Date
			break
		}
		if len(day.Afternoon) > 0 {
			na.Time = day.Afternoon[0]
			na.Date = day.Date
			break
		}
	}

	return na, nil
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
