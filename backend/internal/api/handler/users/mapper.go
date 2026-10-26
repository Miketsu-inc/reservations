package users

import (
	"github.com/miketsu-inc/reservations/backend/internal/domain"
	authServ "github.com/miketsu-inc/reservations/backend/internal/service/auth"
	bookingServ "github.com/miketsu-inc/reservations/backend/internal/service/booking"
	mediaServ "github.com/miketsu-inc/reservations/backend/internal/service/media"
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
			MerchantLogoUrl:   mediaServ.NewURL(b.MerchantLogoImageId, mediaServ.VariantSmall),
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

func mapToGetBookingDetailsResp(in domain.BookingDetailsForUser) bookingDetailsResp {
	images := make([]imageResp, len(in.Images))
	for i, img := range in.Images {
		images[i] = imageResp{
			Url:          mediaServ.NewVariantURL(img.ImageId, mediaServ.VariantLarge),
			ThumbnailUrl: mediaServ.NewVariantURL(img.ImageId, mediaServ.VariantSmall),
			AltText:      img.AltText,
			Width:        img.Width,
			Height:       img.Height,
		}
	}

	return bookingDetailsResp{
		Id:                in.Id,
		Status:            in.Status.String(),
		ParticipantStatus: in.ParticipantStatus.String(),
		BookingType:       in.BookingType.String(),
		IsRecurring:       in.IsRecurring,
		FromDate:          in.FromDate,
		ToDate:            in.ToDate,
		Price:             in.PricePerPerson.ToFormatted(),
		PriceType:         in.PriceType.String(),
		ServiceName:       in.ServiceName,
		MerchantName:      in.MerchantName,
		MerchantUrl:       in.MerchantUrl,
		CancelDeadline:    in.CancelDeadline,
		FormattedLocation: in.FormattedLocation,
		Longitude:         in.Longitude,
		Latitude:          in.Latitude,
		EmployeeFirstName: in.EmployeeFirstName,
		EmployeeLastName:  in.EmployeeLastName,
		EmployeeAvatarUrl: mediaServ.NewURL(in.EmployeeAvatarImageId, mediaServ.VariantSmall),
		Images:            images,
	}
}

func mapToGetBookingCountsResp(in domain.BookingCountsForUser) getBookingCountsResp {
	return getBookingCountsResp{
		Upcoming:  in.Upcoming,
		Completed: in.Completed,
		Cancelled: in.Cancelled,
	}
}

func mapToUpdatePasswordInput(in updatePasswordReq) authServ.UpdatePasswordInput {
	return authServ.UpdatePasswordInput{
		OldPassword: in.OldPassword,
		NewPassword: in.NewPassword,
	}
}

func mapToUploadInput(in createUploadReq) mediaServ.UploadInput {
	return mediaServ.UploadInput{
		ContentType: in.ContentType,
		Size:        in.Size,
	}
}

func mapToUploadResp(in mediaServ.UploadResult) uploadResp {
	headers := make(map[string]string, len(in.Request.Headers))
	for key := range in.Request.Headers {
		headers[key] = in.Request.Headers.Get(key)
	}

	return uploadResp{
		ImageId:   in.ImageId,
		UploadUrl: in.Request.URL,
		Method:    in.Request.Method,
		Headers:   headers,
		ExpiresAt: in.Request.ExpiresAt,
	}
}
