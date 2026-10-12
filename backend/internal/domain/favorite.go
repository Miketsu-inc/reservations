package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/miketsu-inc/reservations/backend/pkg/db"
)

type FavoriteRepository interface {
	WithTx(tx db.DBTX) FavoriteRepository

	AddFavoriteMerchant(ctx context.Context, userId uuid.UUID, merchantId uuid.UUID) error
	RemoveFavoriteMerchant(ctx context.Context, userId uuid.UUID, merchantId uuid.UUID) error
	GetFavoriteMerchants(ctx context.Context, userId uuid.UUID) ([]FavoriteMerchant, error)
	IsFavoriteMerchant(ctx context.Context, userId uuid.UUID, merchantId uuid.UUID) (bool, error)
}

type Favorite struct {
	MerchantId uuid.UUID
	UserId     uuid.UUID
	CreatedAt  time.Time
}

type FavoriteMerchant struct {
	MerchantId        uuid.UUID `db:"merchant_id"`
	MerchantName      string    `db:"merchant_name"`
	MerchantUrlName   string    `db:"merchant_url_name"`
	FormattedLocation *string   `db:"formatted_location"`
}
