package favorite

import (
	"context"

	"github.com/google/uuid"
	"github.com/miketsu-inc/reservations/backend/internal/api/middleware/jwt"
	"github.com/miketsu-inc/reservations/backend/internal/domain"
)

type Service struct {
	favoriteRepo domain.FavoriteRepository
}

func NewService(favoriteRepo domain.FavoriteRepository) *Service {
	return &Service{favoriteRepo: favoriteRepo}
}

func (s *Service) AddFavorite(ctx context.Context, merchantId uuid.UUID) error {
	userId := jwt.MustGetUserIDFromContext(ctx)

	err := s.favoriteRepo.AddFavoriteMerchant(ctx, userId, merchantId)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) RemoveFavorite(ctx context.Context, merchantId uuid.UUID) error {
	userId := jwt.MustGetUserIDFromContext(ctx)

	err := s.favoriteRepo.RemoveFavoriteMerchant(ctx, userId, merchantId)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) GetFavorites(ctx context.Context) ([]domain.FavoriteMerchant, error) {
	userId := jwt.MustGetUserIDFromContext(ctx)

	favorites, err := s.favoriteRepo.GetFavoriteMerchants(ctx, userId)
	if err != nil {
		return nil, err
	}

	return favorites, nil
}

func (s *Service) IsFavorite(ctx context.Context, merchantId uuid.UUID) (bool, error) {
	userId := jwt.MustGetUserIDFromContext(ctx)

	isFavourite, err := s.favoriteRepo.IsFavoriteMerchant(ctx, userId, merchantId)
	if err != nil {
		return false, err
	}

	return isFavourite, nil
}
