package customers

import (
	"github.com/miketsu-inc/reservations/backend/internal/domain"
	customerServ "github.com/miketsu-inc/reservations/backend/internal/service/customer"
	"github.com/miketsu-inc/reservations/backend/pkg/currencyx"
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
		TimesBooked:     in.TimesBooked,
		TimesConfirmed:  in.TimesConfirmed,
		TimesCompleted:  in.TimesCompleted,
		TimesCancelled:  in.TimesCancelled,
		TimesNoShow:     in.TimesNoShow,
		FirstBooking:    in.FirstBooking,
		CompletedValues: mapToCompletedValuesResp(in.CompletedValues),
		FavoriteService: in.FavoriteService,
		NextBooking:     in.NextBooking,
	}
}

func mapToCompletedValuesResp(in []currencyx.Price) []completedValueResp {
	out := make([]completedValueResp, len(in))
	for i, value := range in {
		out[i] = completedValueResp{
			Value:          value,
			FormattedValue: value.ToFormatted(),
		}
	}

	return out
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
