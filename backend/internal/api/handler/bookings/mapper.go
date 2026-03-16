package bookings

import (
	"github.com/miketsu-inc/reservations/backend/internal/domain"
	bookingServ "github.com/miketsu-inc/reservations/backend/internal/service/booking"
)

func mapToCreateByCustomerInput(in createBookingByCustomerReq) bookingServ.CreateByCustomerInput {
	return bookingServ.CreateByCustomerInput{
		MerchantName: in.MerchantName,
		ServiceId:    in.ServiceId,
		LocationId:   in.LocationId,
		TimeStamp:    in.TimeStamp,
		CustomerNote: in.CustomerNote,
		BookingId:    in.BookingId,
	}
}

func mapToCancelByCustomerInput(in cancelByCustomerReq) bookingServ.CancelByCustomerInput {
	return bookingServ.CancelByCustomerInput{
		BookingId:    in.BookingId,
		MerchantName: in.MerchantName,
	}
}

func mapToGetByCustomerResp(in domain.PublicBooking) getByCustomerResp {
	return getByCustomerResp{
		FromDate:          in.FromDate,
		ToDate:            in.ToDate,
		ServiceName:       in.ServiceName,
		CancelDeadline:    in.CancelDeadline,
		FormattedLocation: in.FormattedLocation,
		Price:             in.Price,
		PriceType:         in.PriceType,
		MerchantName:      in.MerchantName,
		IsCancelled:       in.IsCancelled,
	}
}

func mapToCreateByMerchantInput(in createByMerchantReq) bookingServ.CreateByMerchantInput {
	customers := make([]bookingServ.CustomerInput, len(in.Customers))

	for i, c := range in.Customers {
		customers[i] = bookingServ.CustomerInput{
			CustomerId:  c.CustomerId,
			FirstName:   c.FirstName,
			LastName:    c.LastName,
			Email:       c.Email,
			PhoneNumber: c.PhoneNumber,
		}
	}

	return bookingServ.CreateByMerchantInput{
		Customers:    customers,
		ServiceId:    in.ServiceId,
		TimeStamp:    in.TimeStamp,
		MerchantNote: in.MerchantNote,
		IsRecurring:  in.IsRecurring,
		Rrule: &bookingServ.RecurringRuleInput{
			Frequency: in.Rrule.Frequency,
			Interval:  in.Rrule.Interval,
			Weekdays:  in.Rrule.Weekdays,
			Until:     in.Rrule.Until,
		},
	}
}

func mapToUpdateByMerchantInput(in updateByMerchantReq) bookingServ.UpdateByMerchantInput {
	customers := make([]bookingServ.CustomerInput, len(in.Customers))

	for i, c := range in.Customers {
		customers[i] = bookingServ.CustomerInput{
			CustomerId:  c.CustomerId,
			FirstName:   c.FirstName,
			LastName:    c.LastName,
			Email:       c.Email,
			PhoneNumber: c.PhoneNumber,
		}
	}

	return bookingServ.UpdateByMerchantInput{
		Customers:       customers,
		ServiceId:       in.ServiceId,
		TimeStamp:       in.TimeStamp,
		MerchantNote:    in.MerchantNote,
		BookingStatus:   in.BookingStatus,
		UpdateAllFuture: in.UpdateAllFuture,
	}
}

func mapToCancelByMerchantInput(in cancelByMerchantReq) bookingServ.CancelByMerchantInput {
	return bookingServ.CancelByMerchantInput{
		CancellationReason: in.CancellationReason,
	}
}

func mapToUpdateParticipantStatusInput(in updatePaticipantStatusReq) bookingServ.UpdatePaticipantStatusInput {
	return bookingServ.UpdatePaticipantStatusInput{
		Status: in.Status,
	}
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
			Price:           b.Price,
			Cost:            b.Cost,
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
