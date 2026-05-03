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
		Id:          in.Id,
		FirstName:   in.FirstName,
		LastName:    in.LastName,
		Email:       in.Email,
		PhoneNumber: in.PhoneNumber,
		Birthday:    in.Birthday,
		Note:        in.Note,
		IsDummy:     in.IsDummy,
	}
}

func mapToGetStatsResp(in domain.CustomerStatistics) getStatsResp {
	bookings := make([]customerBookingsResp, len(in.Bookings))

	for i, b := range in.Bookings {
		bookings[i] = customerBookingsResp{
			FromDate:          b.FromDate,
			ToDate:            b.ToDate,
			ServiceName:       b.ServiceName,
			CancelDeadline:    b.CancelDeadline,
			FormattedLocation: b.FormattedLocation,
			Price:             b.Price.ToFormatted(),
			PriceType:         b.PriceType,
			MerchantName:      b.MerchantName,
			Status:            b.Status,
		}
	}

	return getStatsResp{
		Id:                   in.Id,
		FirstName:            in.FirstName,
		LastName:             in.LastName,
		Email:                in.Email,
		PhoneNumber:          in.PhoneNumber,
		Birthday:             in.Birthday,
		Note:                 in.Note,
		IsDummy:              in.IsDummy,
		IsBlacklisted:        in.IsBlacklisted,
		BlacklistReason:      in.BlacklistReason,
		TimesBooked:          in.TimesBooked,
		TimesCancelledByUser: in.TimesCancelledByUser,
		TimesUpcoming:        in.TimesUpcoming,
		TimesCompleted:       in.TimesCompleted,
		Bookings:             bookings,
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
