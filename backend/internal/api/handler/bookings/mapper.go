package bookings

import (
	"fmt"
	"time"

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

func mapToUpdateByMerchantInput(in updateByMerchantReq) (bookingServ.UpdateByMerchantInput, error) {
	fromDate, err := time.Parse(time.RFC3339, in.FromDate)
	if err != nil {
		return bookingServ.UpdateByMerchantInput{}, fmt.Errorf("invalid from date: %s", err.Error())
	}

	toDate, err := time.Parse(time.RFC3339, in.ToDate)
	if err != nil {
		return bookingServ.UpdateByMerchantInput{}, fmt.Errorf("invalid to date: %s", err.Error())
	}

	return bookingServ.UpdateByMerchantInput{
		MerchantNote: in.MerchantNote,
		FromDate:     fromDate,
		ToDate:       toDate,
	}, nil
}

func mapToCancelByMerchantInput(in cancelByMerchantReq) bookingServ.CancelByMerchantInput {
	return bookingServ.CancelByMerchantInput{
		CancellationReason: in.CancellationReason,
	}
}

func mapToGetCalendarEventsResp(in domain.CalendarEvents) getCalendarEventsResp {
	bookings := make([]bookingDetails, len(in.Bookings))

	for i, b := range in.Bookings {
		bookings[i] = bookingDetails{
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
