package booking

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/miketsu-inc/reservations/backend/internal/api/middleware/actor"
	"github.com/miketsu-inc/reservations/backend/internal/domain"
	"github.com/miketsu-inc/reservations/backend/internal/types"
	"github.com/miketsu-inc/reservations/backend/pkg/cursor"
)

type GetBookingsInput struct {
	CustomerId *uuid.UUID
	EmployeeId *int
	Statuses   []types.BookingStatus
	Before     *time.Time
	Cursor     string
	PageSize   int
}

type GetBookingsResult struct {
	Bookings    []domain.BookingListItem
	NextCursor  *string
	HasNextPage bool
}

func (s *Service) GetBookings(ctx context.Context, input GetBookingsInput) (GetBookingsResult, error) {
	if (input.CustomerId == nil) == (input.EmployeeId == nil) {
		return GetBookingsResult{}, fmt.Errorf("exactly one booking subject must be provided")
	}

	decodedCursor, err := cursor.Decode[bookingCursor](input.Cursor)
	if err != nil {
		return GetBookingsResult{}, fmt.Errorf("decode booking cursor: %w", err)
	}

	if input.Cursor == "" {
		decodedCursor = bookingCursor{
			Id:       math.MaxInt32,
			FromDate: time.Date(9999, time.December, 31, 23, 59, 59, 0, time.UTC),
		}
		if input.Before != nil {
			decodedCursor.FromDate = *input.Before
		}
	}

	requestActor := actor.MustGetFromContext(ctx)
	var bookings []domain.BookingListItem
	if input.CustomerId != nil {
		bookings, err = s.bookingRepo.GetBookingsForCustomer(
			ctx,
			requestActor.MerchantId,
			*input.CustomerId,
			input.Statuses,
			input.PageSize+1,
			decodedCursor.FromDate,
			decodedCursor.Id,
		)
	} else {
		bookings, err = s.bookingRepo.GetBookingsForEmployee(
			ctx,
			requestActor.MerchantId,
			*input.EmployeeId,
			input.Statuses,
			input.PageSize+1,
			decodedCursor.FromDate,
			decodedCursor.Id,
		)
	}
	if err != nil {
		return GetBookingsResult{}, err
	}

	hasNextPage := len(bookings) > input.PageSize
	var nextCursor *string
	if hasNextPage {
		lastBooking := bookings[input.PageSize-1]
		encodedCursor, err := cursor.Encode(bookingCursor{
			Id:       lastBooking.Id,
			FromDate: lastBooking.FromDate,
		})
		if err != nil {
			return GetBookingsResult{}, fmt.Errorf("encode booking cursor: %w", err)
		}

		nextCursor = &encodedCursor
		bookings = bookings[:input.PageSize]
	}

	return GetBookingsResult{
		Bookings:    bookings,
		NextCursor:  nextCursor,
		HasNextPage: hasNextPage,
	}, nil
}
