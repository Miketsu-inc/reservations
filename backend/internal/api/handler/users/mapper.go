package users

import (
	authServ "github.com/miketsu-inc/reservations/backend/internal/service/auth"
	bookingServ "github.com/miketsu-inc/reservations/backend/internal/service/booking"
	userServ "github.com/miketsu-inc/reservations/backend/internal/service/user"
)

func mapToEditInput(in editReq) userServ.EditInput {
	return userServ.EditInput{
		FirstName:   in.FirstName,
		LastName:    in.LastName,
		PhoneNumber: in.PhoneNumber,
		Email:       in.Email,
	}
}

func mapToGetBookingsResp(in bookingServ.GetForUserResult) getBookingsResp {
	bookings := make([]bookingForUser, len(in.Bookings))

	for i, b := range in.Bookings {
		bookings[i] = bookingForUser{
			Id:                b.Id,
			Status:            b.Status.String(),
			BookingType:       b.BookingType.String(),
			IsRecurring:       b.IsRecurring,
			FromDate:          b.FromDate,
			ToDate:            b.ToDate,
			Price:             b.PricePerPerson.ToFormatted(),
			MerchantName:      b.MerchantName,
			MerchantUrl:       b.MerchantUrl,
			FormattedLocation: b.FormattedLocation,
			ServiceName:       b.ServiceName,
			EmployeeFirstName: b.EmployeeFirstName,
			EmployeeLastName:  b.EmployeeLastName,
		}
	}

	return getBookingsResp{
		Bookings:    bookings,
		HasNextpage: in.HasNextPage,
		NextCursor:  in.NextCursor,
	}
}

func mapToUpdatePasswordInput(in updatePasswordReq) authServ.UpdatePasswordInput {
	return authServ.UpdatePasswordInput{
		OldPassword:        in.OldPassword,
		NewPassword:        in.NewPassword,
		ConfirmNewPassword: in.ConfirmNewPassword,
	}
}
