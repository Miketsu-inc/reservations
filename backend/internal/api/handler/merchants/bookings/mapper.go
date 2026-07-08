package bookings

import (
	"fmt"
	"time"

	bookingServ "github.com/miketsu-inc/reservations/backend/internal/service/booking"
)

func mapToCreateByMerchantInput(in createByMerchantReq) (bookingServ.CreateByMerchantInput, error) {
	timeStamp, err := time.Parse(time.RFC3339, in.TimeStamp)
	if err != nil {
		return bookingServ.CreateByMerchantInput{}, fmt.Errorf("timestamp could not be converted to time: %s", err.Error())
	}

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
		EmployeeId:   in.EmployeeId,
		TimeStamp:    timeStamp,
		MerchantNote: in.MerchantNote,
		IsRecurring:  in.IsRecurring,
		Rrule: &bookingServ.RecurringRuleInput{
			Frequency: in.Rrule.Frequency,
			Interval:  in.Rrule.Interval,
			Weekdays:  in.Rrule.Weekdays,
			Until:     in.Rrule.Until,
		},
	}, nil
}

func mapToUpdateByMerchantInput(in updateByMerchantReq) (bookingServ.UpdateByMerchantInput, error) {
	timeStamp, err := time.Parse(time.RFC3339, in.TimeStamp)
	if err != nil {
		return bookingServ.UpdateByMerchantInput{}, fmt.Errorf("timestamp could not be converted to time: %s", err.Error())
	}

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
		TimeStamp:       timeStamp,
		ServiceName:     in.ServiceName,
		Price:           in.Price,
		PriceType:       in.PriceType,
		MerchantNote:    in.MerchantNote,
		EmployeeId:      in.EmployeeId,
		BookingStatus:   in.BookingStatus,
		UpdateAllFuture: in.UpdateAllFuture,
	}, nil
}

func mapToCancelByMerchantInput(in cancelByMerchantReq) bookingServ.CancelByMerchantInput {
	return bookingServ.CancelByMerchantInput{
		CancellationReason: in.CancellationReason,
		CancelFuture:       in.CancelFuture,
	}
}

func mapToUpdateParticipantStatusInput(in updatePaticipantStatusReq) bookingServ.UpdatePaticipantStatusInput {
	return bookingServ.UpdatePaticipantStatusInput{
		Status: in.Status,
	}
}
