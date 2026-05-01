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
		Price:             in.Price.ToFormatted(),
		PriceType:         in.PriceType,
		MerchantName:      in.MerchantName,
		IsCancelled:       in.IsCancelled,
	}
}
