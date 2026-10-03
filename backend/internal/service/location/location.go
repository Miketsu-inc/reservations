package location

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/miketsu-inc/reservations/backend/internal/api/middleware/actor"
	"github.com/miketsu-inc/reservations/backend/internal/domain"
	"github.com/miketsu-inc/reservations/backend/internal/types"
)

type Service struct {
	locationRepo domain.LocationRepository
}

func NewService(location domain.LocationRepository) *Service {
	return &Service{locationRepo: location}
}

type NewInput struct {
	Country           *string
	City              *string
	PostalCode        *string
	Address           *string
	GeoPoint          types.GeoPoint
	PlaceId           *string
	FormattedLocation string
	IsPrimary         bool
	IsActive          bool
}

func (s *Service) New(ctx context.Context, input NewInput) error {
	actor := actor.MustGetFromContext(ctx)

	if err := s.locationRepo.NewLocation(ctx, domain.Location{
		MerchantId:        actor.MerchantId,
		Country:           input.Country,
		City:              input.City,
		PostalCode:        input.PostalCode,
		Address:           input.Address,
		GeoPoint:          input.GeoPoint,
		PlaceId:           input.PlaceId,
		FormattedLocation: input.FormattedLocation,
		IsPrimary:         input.IsPrimary,
		IsActive:          input.IsActive,
	}); err != nil {
		return err
	}

	return nil
}

func (s *Service) Get(ctx context.Context) (domain.Location, error) {
	actor := actor.MustGetFromContext(ctx)

	location, err := s.locationRepo.GetLocation(ctx, actor.LocationId, actor.MerchantId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Location{}, ErrLocationNotFound
		}

		return domain.Location{}, err
	}

	return location, nil
}
