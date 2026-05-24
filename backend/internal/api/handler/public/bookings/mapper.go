package bookings

import (
	"fmt"
	"time"

	"github.com/miketsu-inc/reservations/backend/internal/domain"
	bookingServ "github.com/miketsu-inc/reservations/backend/internal/service/booking"
)

func mapToCreateByCustomerInput(in createBookingByCustomerReq) (bookingServ.CreateByCustomerInput, error) {
	timeStamp, err := time.Parse(time.RFC3339, in.TimeStamp)
	if err != nil {
		return bookingServ.CreateByCustomerInput{}, fmt.Errorf("timestamp could not be converted to time: %w", err)
	}

	return bookingServ.CreateByCustomerInput{
		MerchantName: in.MerchantName,
		ServiceId:    in.ServiceId,
		LocationId:   in.LocationId,
		TimeStamp:    timeStamp,
		CustomerNote: in.CustomerNote,
		BookingId:    in.BookingId,
	}, nil
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
		Price:             in.Price.ToFormatted(),
		PriceType:         in.PriceType,
		MerchantName:      in.MerchantName,
		Status:            in.Status,
	}
}
