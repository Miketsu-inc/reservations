package customers

import (
	"github.com/miketsu-inc/reservations/backend/internal/domain"
	customerServ "github.com/miketsu-inc/reservations/backend/internal/service/customer"
)

func mapToNewInput(in newReq) customerServ.NewInput {
	return customerServ.NewInput{
		FirstName:   in.FirstName,
		LastName:    in.LastName,
		Email:       in.Email,
		PhoneNumber: in.PhoneNumber,
		Birthday:    in.Birthday,
		Note:        in.Note,
	}
}

func mapToUpdateInput(in updateReq) customerServ.UpdateInput {
	return customerServ.UpdateInput{
		Id:          in.Id,
		FirstName:   in.FirstName,
		LastName:    in.LastName,
		Email:       in.Email,
		PhoneNumber: in.PhoneNumber,
		Birthday:    in.Birthday,
		Note:        in.Note,
	}
}

func mapToGetResp(in domain.CustomerInfo) getResp {
	return getResp{
		Id:              in.Id,
		FirstName:       in.FirstName,
		LastName:        in.LastName,
		Email:           in.Email,
		PhoneNumber:     in.PhoneNumber,
		Birthday:        in.Birthday,
		Note:            in.Note,
		IsDummy:         in.IsDummy,
		IsBlacklisted:   in.IsBlacklisted,
		BlacklistReason: in.BlacklistReason,
	}
}

func mapToGetStatsResp(in domain.CustomerStatistics) getStatsResp {
	return getStatsResp{
		TimesBooked:          in.TimesBooked,
		TimesCancelledByUser: in.TimesCancelledByUser,
		TimesNoShow:          in.TimesNoShow,
		TimesUpcoming:        in.TimesUpcoming,
		TimesCompleted:       in.TimesCompleted,
		FirstBooking:         in.FirstBooking,
		LastVisited:          in.LastVisited,
	}
}

func mapToGetBookingsResp(in customerServ.GetBookingsResult) getBookingsResp {
	bookings := make([]customerBookingsResp, len(in.Bookings))
	for i, b := range in.Bookings {
		bookings[i] = customerBookingsResp{
			Id:                b.Id,
			BookingType:       b.BookingType,
			IsRecurring:       b.IsRecurring,
			FromDate:          b.FromDate,
			ToDate:            b.ToDate,
			ServiceName:       b.ServiceName,
			ServiceColor:      b.ServiceColor,
			FormattedLocation: b.FormattedLocation,
			Price:             b.Price.ToFormatted(),
			PriceType:         b.PriceType,
			Status:            b.Status,
			EmployeeFirstName: b.EmployeeFirstName,
			EmployeeLastName:  b.EmployeeLastName,
		}
	}

	return getBookingsResp{
		Bookings:    bookings,
		HasNextPage: in.HasNextPage,
		NextCursor:  in.NextCursor,
	}
}

func mapToBlacklistInput(in blacklistReq) customerServ.BlacklistInput {
	return customerServ.BlacklistInput{
		CustomerId:      in.CustomerId,
		BlacklistReason: in.BlacklistReason,
	}
}

func mapToGetAllResp(in []domain.PublicCustomer) []getAllResp {
	out := make([]getAllResp, len(in))

	for i, c := range in {
		out[i] = getAllResp{
			Id:              c.Id,
			FirstName:       c.FirstName,
			LastName:        c.LastName,
			Email:           c.Email,
			PhoneNumber:     c.PhoneNumber,
			Birthday:        c.Birthday,
			Note:            c.Note,
			IsDummy:         c.IsDummy,
			IsBlacklisted:   c.IsBlacklisted,
			BlacklistReason: c.BlacklistReason,
			TimesBooked:     c.TimesBooked,
			TimesCancelled:  c.TimesCancelled,
		}
	}

	return out
}

func mapToTransferBookingsInput(in transferBookingsReq) customerServ.TransferBookingsInput {
	return customerServ.TransferBookingsInput{
		FromCustomerId: in.FromCustomerId,
		ToCustomerId:   in.ToCustomerId,
	}
}
