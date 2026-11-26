package favorite

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/miketsu-inc/reservations/backend/internal/api/middleware/jwt"
	"github.com/miketsu-inc/reservations/backend/internal/domain"
)

type Service struct {
	favoriteRepo domain.FavoriteRepository
	locationRepo domain.LocationRepository
}

func NewService(favoriteRepo domain.FavoriteRepository, locationRepo domain.LocationRepository) *Service {
	return &Service{favoriteRepo: favoriteRepo, locationRepo: locationRepo}
}

func (s *Service) AddFavorite(ctx context.Context, merchantId uuid.UUID, locationId int) error {
	userId := jwt.MustGetUserIDFromContext(ctx)

	location, err := s.locationRepo.GetLocation(ctx, locationId, merchantId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrLocationNotFound
		}

		return err
	}

	if !location.IsActive {
		return ErrLocationNotFound
	}

	return s.favoriteRepo.AddFavoriteLocation(ctx, userId, merchantId, locationId)
}

func (s *Service) RemoveFavorite(ctx context.Context, merchantId uuid.UUID, locationId int) error {
	userId := jwt.MustGetUserIDFromContext(ctx)

	return s.favoriteRepo.RemoveFavoriteLocation(ctx, userId, merchantId, locationId)
}

func (s *Service) GetFavorites(ctx context.Context) ([]domain.FavoriteLocation, error) {
	userId := jwt.MustGetUserIDFromContext(ctx)

	return s.favoriteRepo.GetFavoriteLocations(ctx, userId)
}

func (s *Service) IsFavorite(ctx context.Context, merchantId uuid.UUID, locationId int) (bool, error) {
	userId := jwt.MustGetUserIDFromContext(ctx)

	return s.favoriteRepo.IsFavoriteLocation(ctx, userId, merchantId, locationId)
}
