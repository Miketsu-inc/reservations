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

func mapToGetDashboardBookingsResp(in []domain.PublicBookingDetails) getDashboardBookingsResp {
	bookings := make([]dashboardBookingResp, len(in))

	for i, booking := range in {
		bookings[i] = dashboardBookingResp{
			ID:                  booking.ID,
			BookingType:         booking.BookingType,
			BookingStatus:       booking.BookingStatus,
			ParticipantStatus:   booking.ParticipantStatus,
			IsRecurring:         booking.IsRecurring,
			FromDate:            booking.FromDate,
			ToDate:              booking.ToDate,
			CustomerNote:        booking.CustomerNote,
			MerchantNote:        booking.MerchantNote,
			ServiceName:         booking.ServiceName,
			ServiceColor:        booking.ServiceColor,
			Price:               booking.Price.ToFormatted(),
			PriceType:           booking.PriceType,
			CurrentParticipants: booking.CurrentParticipants,
			MaxParticipants:     booking.MaxParticipants,
			CustomerFirstName:   booking.CustomerFirstName,
			CustomerLastName:    booking.CustomerLastName,
			EmployeeFirstName:   booking.EmployeeFirstName,
			EmployeeLastName:    booking.EmployeeLastName,
		}
	}

	return getDashboardBookingsResp{
		Bookings: bookings,
	}
}

func mapToDashboardStatisticsResp(in domain.DashboardStatistics) dashboardStatisticsResp {
	return dashboardStatisticsResp{
		RevenueSum:            in.RevenueSum,
		RevenueChange:         in.RevenueChange,
		Bookings:              in.Bookings,
		BookingsChange:        in.BookingsChange,
		Cancellations:         in.Cancellations,
		CancellationsChange:   in.CancellationsChange,
		AverageDuration:       in.AverageDuration,
		AverageDurationChange: in.AverageDurationChange,
	}
}

func mapToDashboardRevenueResp(in domain.DashboardRevenue) dashboardRevenueResp {
	revenue := make([]revenueStatResp, len(in.Revenue))
	for i, stat := range in.Revenue {
		revenue[i] = revenueStatResp{
			Value: stat.Value,
			Day:   stat.Day,
		}
	}

	return dashboardRevenueResp{
		PeriodStart: in.PeriodStart,
		PeriodEnd:   in.PeriodEnd,
		Revenue:     revenue,
	}
}

func mapToCheckUrlInput(in checkUrlReq) merchantServ.CheckUrlInput {
	return merchantServ.CheckUrlInput{
		Name: in.Name,
	}
}

func mapToCheckUrlResp(in string) checkUrlResp {
	return checkUrlResp{
		Name: in,
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

func mapToGetTeamMembersForCalendarResp(in []domain.Employee) []getTeamMembersForCalendarResp {
	teamMembers := make([]getTeamMembersForCalendarResp, len(in))

	for i, m := range in {
		teamMembers[i] = getTeamMembersForCalendarResp{
			Id:        m.Id,
			FirstName: *m.FirstName,
			LastName:  *m.LastName,
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
		bookings[i] = mapToBookingForCalendarResp(b)
	}

	blockedTimes := make([]blockedTime, len(in.BlockedTimes))

	for i, b := range in.BlockedTimes {
		blockedTimes[i] = blockedTime{
			ID:            b.ID,
			EmployeeIds:   b.EmployeeIds,
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

func mapToBookingForCalendarResp(in domain.BookingForCalendar) bookingForCalendar {
	participants := make([]bookingParticipantForCalendar, len(in.Participants))
	for i, p := range in.Participants {
		participants[i] = bookingParticipantForCalendar{
			Id:           p.Id,
			CustomerId:   p.CustomerId,
			FirstName:    p.FirstName,
			LastName:     p.LastName,
			CustomerNote: p.CustomerNote,
			Status:       p.Status,
		}
	}

	return bookingForCalendar{
		ID:              in.ID,
		BookingType:     in.BookingType,
		BookingStatus:   in.BookingStatus,
		FromDate:        in.FromDate,
		ToDate:          in.ToDate,
		IsRecurring:     in.IsRecurring,
		Duration:        int(in.ToDate.Sub(in.FromDate).Minutes()),
		MerchantNote:    in.MerchantNote,
		EmployeeId:      in.EmployeeId,
		ServiceId:       in.ServiceId,
		ServiceName:     in.ServiceName,
		ServiceColor:    in.ServiceColor,
		MaxParticipants: in.MaxParticipants,
		Price:           in.Price.ToFormatted(),
		PriceType:       in.PriceType,
		Participants:    participants,
	}
}
