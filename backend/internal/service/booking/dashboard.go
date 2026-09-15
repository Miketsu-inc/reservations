package booking

import (
	"context"
	"fmt"
	"time"

	"github.com/miketsu-inc/reservations/backend/internal/api/middleware/actor"
	"github.com/miketsu-inc/reservations/backend/internal/domain"
)

const (
	dashboardUpcomingBookingLimit = 1
	dashboardLatestBookingLimit   = 3
)

func (s *Service) GetDashboardBookings(ctx context.Context, view string) ([]domain.PublicBookingDetails, error) {
	actor := actor.MustGetFromContext(ctx)
	afterDate := time.Now().UTC()

	var bookings []domain.PublicBookingDetails
	var err error

	switch view {
	case "latest":
		bookings, err = s.bookingRepo.GetLatestBookings(ctx, actor.MerchantId, afterDate, dashboardLatestBookingLimit)
	case "upcoming":
		bookings, err = s.bookingRepo.GetUpcomingBookings(ctx, actor.MerchantId, afterDate, dashboardUpcomingBookingLimit)
	default:
		return []domain.PublicBookingDetails{}, fmt.Errorf("invalid booking view")
	}
	if err != nil {
		return []domain.PublicBookingDetails{}, err
	}

	return bookings, nil
}
