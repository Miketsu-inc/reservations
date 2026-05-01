package merchants

import (
	"time"

	"github.com/miketsu-inc/reservations/backend/internal/api/middleware/actor"
	"github.com/miketsu-inc/reservations/backend/internal/domain"
	merchantServ "github.com/miketsu-inc/reservations/backend/internal/service/merchant"
	"github.com/miketsu-inc/reservations/backend/pkg/currencyx"
)

func mapToMeResp(in actor.EmployeeContext) meResp {
	return meResp{
		UserId:     in.UserId,
		MerchantId: in.MerchantId,
		LocationId: in.LocationId,
		EmployeeId: in.EmployeeId,
		Role:       in.Role,
	}
}

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
			Status:          b.Status,
			FromDate:        b.FromDate,
			ToDate:          b.ToDate,
			CustomerNote:    b.CustomerNote,
			MerchantNote:    b.MerchantNote,
			ServiceName:     b.ServiceName,
			ServiceColor:    b.ServiceColor,
			ServiceDuration: b.ServiceDuration,
			Price:           b.Price.ToFormatted(),
			Cost:            b.Cost.ToFormatted(),
			FirstName:       b.FirstName,
			LastName:        b.LastName,
			PhoneNumber:     b.PhoneNumber,
		}
	}

	latestBookings := make([]bookingDetailsResp, len(in.LatestBookings))

	for i, b := range in.LatestBookings {
		latestBookings[i] = bookingDetailsResp{
			ID:              b.ID,
			Status:          b.Status,
			FromDate:        b.FromDate,
			ToDate:          b.ToDate,
			CustomerNote:    b.CustomerNote,
			MerchantNote:    b.MerchantNote,
			ServiceName:     b.ServiceName,
			ServiceColor:    b.ServiceColor,
			ServiceDuration: b.ServiceDuration,
			Price:           b.Price.ToFormatted(),
			Cost:            b.Cost.ToFormatted(),
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
				StartTime: s.StartTime.Format("15:04"),
				EndTime:   s.EndTime.Format("15:04"),
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
		ApprovalPolicy:    in.ApprovalPolicy,
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

func mapToUpdateSettingsInput(in updateSettingsReq) (merchantServ.UpdateSettingsInput, error) {
	businessHours := make(domain.BusinessHours, len(in.BusinessHours))

	for day, slots := range in.BusinessHours {
		timeSlots := make([]domain.TimeSlot, len(slots))

		for i, s := range slots {
			startTime, err := time.Parse("15:04", s.StartTime)
			if err != nil {
				return merchantServ.UpdateSettingsInput{}, err
			}

			endTime, err := time.Parse("15:04", s.EndTime)
			if err != nil {
				return merchantServ.UpdateSettingsInput{}, err
			}

			timeSlots[i] = domain.TimeSlot{
				StartTime: startTime,
				EndTime:   endTime,
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
		ApprovalPolicy:   in.ApprovalPolicy,
		BusinessHours:    businessHours,
	}, nil
}

func mapToGetNormalizedBusinessHoursResp(in domain.BusinessHours) map[int]timeSlotResp {
	businessHours := make(map[int]timeSlotResp, len(in))

	for day, ts := range in {
		businessHours[day] = timeSlotResp{
			StartTime: ts[0].StartTime.Format("15:04"),
			EndTime:   ts[0].EndTime.Format("15:04"),
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
		StartHour:          in.StartHour.Format("15:04"),
		EndHour:            in.EndHour.Format("15:04"),
		TimeFrequency:      in.TimeFrequency.Format("15:04"),
	}
}

func mapToUpdatePreferencesInput(in updatePreferencesReq) (merchantServ.UpdatePreferencesInput, error) {
	startHour, err := time.Parse("15:04", in.StartHour)
	if err != nil {
		return merchantServ.UpdatePreferencesInput{}, err
	}

	endHour, err := time.Parse("15:04", in.EndHour)
	if err != nil {
		return merchantServ.UpdatePreferencesInput{}, err
	}

	timeFreq, err := time.Parse("15:04", in.TimeFrequency)
	if err != nil {
		return merchantServ.UpdatePreferencesInput{}, err
	}

	return merchantServ.UpdatePreferencesInput{
		FirstDayOfWeek:     in.FirstDayOfWeek,
		TimeFormat:         in.TimeFormat,
		CalendarView:       in.CalendarView,
		CalendarViewMobile: in.CalendarViewMobile,
		StartHour:          startHour,
		EndHour:            endHour,
		TimeFrequency:      timeFreq,
	}, nil
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
				Price:           currencyx.FormatPrice(s.Price),
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

func mapToGetCalendarEventsResp(in domain.CalendarEvents) getCalendarEventsResp {
	bookings := make([]bookingForCalendar, len(in.Bookings))

	for i, b := range in.Bookings {
		bookings[i] = bookingForCalendar{
			ID:              b.ID,
			BookingType:     b.BookingType,
			BookingStatus:   b.BookingStatus,
			FromDate:        b.FromDate,
			ToDate:          b.ToDate,
			IsRecurring:     b.IsRecurring,
			MerchantNote:    b.MerchantNote,
			ServiceId:       b.ServiceId,
			ServiceName:     b.ServiceName,
			ServiceColor:    b.ServiceColor,
			MaxParticipants: b.MaxParticipants,
			Price:           b.Price.ToFormatted(),
			Cost:            b.Cost.ToFormatted(),
		}

		participants := make([]bookingParticipantForCalendar, len(b.Participants))
		for j, p := range b.Participants {
			participants[j] = bookingParticipantForCalendar{
				Id:           p.Id,
				CustomerId:   p.CustomerId,
				FirstName:    p.FirstName,
				LastName:     p.LastName,
				CustomerNote: p.CustomerNote,
				Status:       p.Status,
			}
		}

		bookings[i].Participants = participants
	}

	blockedTimes := make([]blockedTime, len(in.BlockedTimes))

	for i, b := range in.BlockedTimes {
		blockedTimes[i] = blockedTime{
			ID:            b.ID,
			EmployeeId:    b.EmployeeId,
			Name:          b.Name,
			FromDate:      b.FromDate,
			ToDate:        b.ToDate,
			AllDay:        b.AllDay,
			Icon:          b.Icon,
			BlockedTypeId: b.BlockedTypeId,
		}
	}

	return getCalendarEventsResp{
		Bookings:     bookings,
		BlockedTimes: blockedTimes,
	}
}
