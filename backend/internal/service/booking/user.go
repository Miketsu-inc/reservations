package booking

import (
	"context"
	"fmt"
	"time"

	"github.com/miketsu-inc/reservations/backend/internal/api/middleware/jwt"
	"github.com/miketsu-inc/reservations/backend/internal/domain"
	"github.com/miketsu-inc/reservations/backend/pkg/cursor"
)

type bookingCursor struct {
	Id       int       `json:"id"`
	FromDate time.Time `json:"from_date"`
}

type GetForUserResult struct {
	Bookings    []domain.BookingForUser
	NextCursor  *string
	HasNextPage bool
}

func (s *Service) GetForUser(ctx context.Context, status string, cursorStr string, pageSize int) (GetForUserResult, error) {
	userId := jwt.MustGetUserIDFromContext(ctx)

	var bookings []domain.BookingForUser
	var err error

	// +1 so we can check if there is another page
	limit := pageSize + 1

	decodedCursor, err := cursor.Decode[bookingCursor](cursorStr)
	if err != nil {
		return GetForUserResult{}, fmt.Errorf("error during cursor decoding: %s", err.Error())
	}

	switch status {
	case "upcoming":
		bookings, err = s.bookingRepo.GetUpcomingBookingsForUser(ctx, userId, limit, decodedCursor.FromDate, decodedCursor.Id)
	case "completed":
		bookings, err = s.bookingRepo.GetCompletedBookingsForUser(ctx, userId, limit, decodedCursor.FromDate, decodedCursor.Id)
	case "cancelled":
		bookings, err = s.bookingRepo.GetCancelledBookingsForUser(ctx, userId, limit, decodedCursor.FromDate, decodedCursor.Id)
	default:
		return GetForUserResult{}, fmt.Errorf("invalid status query parameter")
	}
	if err != nil {
		return GetForUserResult{}, err
	}

	var nextCursor *string

	hasNextPage := len(bookings) > pageSize

	if hasNextPage {
		encodedCursor, err := cursor.Encode(bookingCursor{
			Id:       bookings[pageSize-1].Id,
			FromDate: bookings[pageSize-1].FromDate,
		})
		if err != nil {
			return GetForUserResult{}, fmt.Errorf("error during cursor encoding: %s", err.Error())
		}

		nextCursor = &encodedCursor
		bookings = bookings[:pageSize]
	}

	return GetForUserResult{
		Bookings:    bookings,
		NextCursor:  nextCursor,
		HasNextPage: hasNextPage,
	}, nil
}
