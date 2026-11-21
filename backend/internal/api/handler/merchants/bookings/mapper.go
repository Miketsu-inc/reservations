package bookings

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/miketsu-inc/reservations/backend/internal/domain"
	bookingServ "github.com/miketsu-inc/reservations/backend/internal/service/booking"
	mediaServ "github.com/miketsu-inc/reservations/backend/internal/service/media"
)

func mapToGetBookingsResp(in bookingServ.GetBookingsResult) getBookingsResp {
	bookings := make([]bookingListItemResp, len(in.Bookings))
	for i, booking := range in.Bookings {
		bookings[i] = bookingListItemResp{
			Id:                booking.Id,
			BookingType:       booking.BookingType,
			IsRecurring:       booking.IsRecurring,
			FromDate:          booking.FromDate,
			ToDate:            booking.ToDate,
			ServiceName:       booking.ServiceName,
			ServiceColor:      booking.ServiceColor,
			FormattedLocation: booking.FormattedLocation,
			Price:             booking.Price.ToFormatted(),
			PriceType:         booking.PriceType,
			BookingStatus:     booking.BookingStatus,
			ParticipantStatus: booking.ParticipantStatus,
			Employee:          mapToEmployeeResp(booking.EmployeeId, booking.EmployeeFirstName, booking.EmployeeLastName, booking.EmployeeAvatarImageId),
		}
	}

	return getBookingsResp{
		Bookings:    bookings,
		HasNextPage: in.HasNextPage,
		NextCursor:  in.NextCursor,
	}
}

func mapToEmployeeResp(id *int, firstName, lastName *string, avatarImageId *uuid.UUID) *employeeResp {
	if id == nil {
		return nil
	}

	return &employeeResp{
		Id:        *id,
		FirstName: firstName,
		LastName:  lastName,
		AvatarUrl: mediaServ.NewURL(avatarImageId, mediaServ.VariantSmall),
	}
}

func mapToGetByMerchantResp(in domain.BookingForCalendar) getByMerchantResp {
	participants := make([]bookingParticipantResp, len(in.Participants))
	for i, p := range in.Participants {
		participants[i] = bookingParticipantResp{
			Id:           p.Id,
			CustomerId:   p.CustomerId,
			FirstName:    p.FirstName,
			LastName:     p.LastName,
			CustomerNote: p.CustomerNote,
			Status:       p.Status,
			AvatarUrl:    mediaServ.NewURL(p.AvatarImageId, mediaServ.VariantSmall),
		}
	}

	phases := make([]bookingPhaseResp, len(in.Phases))
	for i, phase := range in.Phases {
		phases[i] = bookingPhaseResp{
			Duration:  int(phase.ToDate.Sub(phase.FromDate).Minutes()),
			PhaseType: phase.PhaseType,
		}
	}

	return getByMerchantResp{
		ID:              in.ID,
		BookingType:     in.BookingType,
		BookingStatus:   in.BookingStatus,
		FromDate:        in.FromDate,
		ToDate:          in.ToDate,
		IsRecurring:     in.IsRecurring,
		MerchantNote:    in.MerchantNote,
		EmployeeId:      in.EmployeeId,
		IsUnassigned:    in.IsUnassigned,
		ServiceId:       in.ServiceId,
		ServiceName:     in.ServiceName,
		ServiceColor:    in.ServiceColor,
		MaxParticipants: in.MaxParticipants,
		PricePerPerson:  in.PricePerPerson.ToFormatted(),
		PriceType:       in.PriceType,
		Participants:    participants,
		Phases:          phases,
	}
}

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
			UntilDate: in.Rrule.UntilDate,
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
