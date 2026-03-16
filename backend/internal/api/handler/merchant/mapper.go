package merchant

import (
	"github.com/miketsu-inc/reservations/backend/internal/domain"
	merchantServ "github.com/miketsu-inc/reservations/backend/internal/service/merchant"
)

func mapToUpdateNameInput(in updateNameReq) merchantServ.UpdateNameInput {
	return merchantServ.UpdateNameInput{
		Name: in.Name,
	}
}

func mapToCheckUrlInput(in checkUrlReq) merchantServ.CheckUrlInput {
	return merchantServ.CheckUrlInput{
		Name: in.Name,
	}
}

func mapToGetDashboardResp(in domain.DashboardData) getDashboardResp {
	upcomingBookings := make([]bookingDetailsResp, len(in.UpcomingBookings))

	for i, b := range in.UpcomingBookings {
		upcomingBookings[i] = bookingDetailsResp{
			ID:              b.ID,
			FromDate:        b.FromDate,
			ToDate:          b.ToDate,
			CustomerNote:    b.CustomerNote,
			MerchantNote:    b.MerchantNote,
			ServiceName:     b.ServiceName,
			ServiceColor:    b.ServiceColor,
			ServiceDuration: b.ServiceDuration,
			Price:           b.Price,
			Cost:            b.Cost,
			FirstName:       b.FirstName,
			LastName:        b.LastName,
			PhoneNumber:     b.PhoneNumber,
		}
	}

	latestBookings := make([]bookingDetailsResp, len(in.LatestBookings))

	for i, b := range in.LatestBookings {
		latestBookings[i] = bookingDetailsResp{
			ID:              b.ID,
			FromDate:        b.FromDate,
			ToDate:          b.ToDate,
			CustomerNote:    b.CustomerNote,
			MerchantNote:    b.MerchantNote,
			ServiceName:     b.ServiceName,
			ServiceColor:    b.ServiceColor,
			ServiceDuration: b.ServiceDuration,
			Price:           b.Price,
			Cost:            b.Cost,
			FirstName:       b.FirstName,
			LastName:        b.LastName,
			PhoneNumber:     b.PhoneNumber,
		}
	}

	lowStockProducts := make([]lowStockProductResp, len(in.LowStockProducts))

	for i, p := range in.LowStockProducts {
		lowStockProducts[i] = lowStockProductResp{
			Id:            p.Id,
			Name:          p.Name,
			MaxAmount:     p.MaxAmount,
			CurrentAmount: p.CurrentAmount,
			Unit:          p.Unit,
			FillRatio:     p.FillRatio,
		}
	}

	revenueStats := make([]revenueStatResp, len(in.Statistics.Revenue))

	for i, r := range in.Statistics.Revenue {
		revenueStats[i] = revenueStatResp{
			Value: r.Value,
			Day:   r.Day,
		}
	}

	return getDashboardResp{
		PeriodStart:      in.PeriodStart,
		PeriodEnd:        in.PeriodEnd,
		UpcomingBookings: upcomingBookings,
		LatestBookings:   latestBookings,
		LowStockProducts: lowStockProducts,
		Statistics: dashboardStatisticsResp{
			Revenue:               revenueStats,
			RevenueSum:            in.Statistics.RevenueSum,
			RevenueChange:         in.Statistics.RevenueChange,
			Bookings:              in.Statistics.Bookings,
			BookingsChange:        in.Statistics.BookingsChange,
			Cancellations:         in.Statistics.Cancellations,
			CancellationsChange:   in.Statistics.CancellationsChange,
			AverageDuration:       in.Statistics.AverageDuration,
			AverageDurationChange: in.Statistics.AverageDurationChange,
		},
	}
}

func mapToCheckUrlResp(in merchantServ.CheckUrlInput) checkUrlResp {
	return checkUrlResp{
		Name: in.Name,
	}
}

func mapToGetSettingsResp(in domain.MerchantSettingsInfo) getSettingsResp {
	businessHours := make(map[int][]timeSlotResp, len(in.BusinessHours))

	for day, slots := range in.BusinessHours {
		timeSlots := make([]timeSlotResp, len(slots))

		for i, s := range slots {
			timeSlots[i] = timeSlotResp{
				StartTime: s.StartTime,
				EndTime:   s.EndTime,
			}
		}

		businessHours[day] = timeSlots
	}

	return getSettingsResp{
		Name:              in.Name,
		ContactEmail:      in.ContactEmail,
		Introduction:      in.Introduction,
		Announcement:      in.Announcement,
		AboutUs:           in.AboutUs,
		ParkingInfo:       in.ParkingInfo,
		PaymentInfo:       in.PaymentInfo,
		CancelDeadline:    in.CancelDeadline,
		BookingWindowMin:  in.BookingWindowMin,
		BookingWindowMax:  in.BookingWindowMax,
		BufferTime:        in.BufferTime,
		Timezone:          in.Timezone,
		BusinessHours:     businessHours,
		LocationId:        in.LocationId,
		Country:           in.Country,
		City:              in.City,
		PostalCode:        in.PostalCode,
		Address:           in.Address,
		FormattedLocation: in.FormattedLocation,
	}
}

func mapToUpdateSettingsInput(in updateSettingsReq) merchantServ.UpdateSettingsInput {
	businessHours := make(map[int][]domain.TimeSlot, len(in.BusinessHours))

	for day, slots := range in.BusinessHours {
		timeSlots := make([]domain.TimeSlot, len(slots))

		for i, s := range slots {
			timeSlots[i] = domain.TimeSlot{
				StartTime: s.StartTime,
				EndTime:   s.EndTime,
			}
		}

		businessHours[day] = timeSlots
	}

	return merchantServ.UpdateSettingsInput{
		Introduction:     in.Introduction,
		Announcement:     in.Announcement,
		AboutUs:          in.AboutUs,
		ParkingInfo:      in.ParkingInfo,
		PaymentInfo:      in.PaymentInfo,
		CancelDeadline:   in.CancelDeadline,
		BookingWindowMin: in.BookingWindowMin,
		BookingWindowMax: in.BookingWindowMax,
		BufferTime:       in.BufferTime,
		BusinessHours:    businessHours,
	}
}

func mapToGetNormalizedBusinessHoursResp(in map[int]domain.TimeSlot) map[int]timeSlotResp {
	businessHours := make(map[int]timeSlotResp, len(in))

	for day, ts := range in {
		businessHours[day] = timeSlotResp{
			StartTime: ts.StartTime,
			EndTime:   ts.EndTime,
		}
	}

	return businessHours
}

func mapToGetPreferencesResp(in domain.PreferenceData) getPreferencesResp {
	return getPreferencesResp{
		FirstDayOfWeek:     in.FirstDayOfWeek,
		TimeFormat:         in.TimeFormat,
		CalendarView:       in.CalendarView,
		CalendarViewMobile: in.CalendarViewMobile,
		StartHour:          in.StartHour.String(),
		EndHour:            in.EndHour.String(),
		TimeFrequency:      in.TimeFrequency.String(),
	}
}

func mapToUpdatePreferencesInput(in updatePreferencesReq) merchantServ.UpdatePreferencesInput {
	return merchantServ.UpdatePreferencesInput{
		FirstDayOfWeek:     in.FirstDayOfWeek,
		TimeFormat:         in.TimeFormat,
		CalendarView:       in.CalendarView,
		CalendarViewMobile: in.CalendarViewMobile,
		StartHour:          domain.TimeString(in.StartHour),
		EndHour:            domain.TimeString(in.EndHour),
		TimeFrequency:      domain.TimeString(in.TimeFrequency),
	}
}

func mapToGetTeamMembersForCalendarResp(in []domain.EmployeeForCalendar) []getTeamMembersForCalendarResp {
	teamMembers := make([]getTeamMembersForCalendarResp, len(in))

	for i, m := range in {
		teamMembers[i] = getTeamMembersForCalendarResp{
			Id:        m.Id,
			FirstName: m.FirstName,
			LastName:  m.LastName,
		}
	}

	return teamMembers
}

func mapToGetServicesForCalendarResp(in []domain.ServicesGroupedByCategoriesForCalendar) []getServicesForCalendarResp {
	servicesGroupedByCategories := make([]getServicesForCalendarResp, len(in))

	for i, c := range in {
		services := make([]calendarServiceResp, len(c.Services))

		for j, s := range c.Services {
			services[j] = calendarServiceResp{
				Id:              s.Id,
				Name:            s.Name,
				Duration:        s.Duration,
				Price:           s.Price,
				PriceType:       s.PriceType,
				Color:           s.Color,
				BookingType:     s.BookingType,
				MaxParticipants: s.MaxParticipants,
			}
		}

		servicesGroupedByCategories[i] = getServicesForCalendarResp{
			Id:       c.Id,
			Name:     c.Name,
			Services: services,
		}
	}

	return servicesGroupedByCategories
}

func mapToGetCustomersForCalendarResp(in []domain.CustomerForCalendar) []getCustomersForCalendarResp {
	customers := make([]getCustomersForCalendarResp, len(in))

	for i, m := range in {
		customers[i] = getCustomersForCalendarResp{
			CustomerId:  m.CustomerId,
			FirstName:   m.FirstName,
			LastName:    m.LastName,
			Email:       m.Email,
			PhoneNumber: m.PhoneNumber,
			BirthDay:    m.BirthDay,
			IsDummy:     m.IsDummy,
			LastVisited: m.LastVisited,
		}
	}

	return customers
}
