package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/miketsu-inc/reservations/backend/pkg/db"
)

type FavoriteRepository interface {
	WithTx(tx db.DBTX) FavoriteRepository

	AddFavoriteLocation(ctx context.Context, userId uuid.UUID, merchantId uuid.UUID, locationId int) error
	RemoveFavoriteLocation(ctx context.Context, userId uuid.UUID, merchantId uuid.UUID, locationId int) error
	GetFavoriteLocations(ctx context.Context, userId uuid.UUID) ([]FavoriteLocation, error)
	IsFavoriteLocation(ctx context.Context, userId uuid.UUID, merchantId uuid.UUID, locationId int) (bool, error)
}

type Favorite struct {
	MerchantId uuid.UUID
	LocationId int
	UserId     uuid.UUID
	CreatedAt  time.Time
}

type FavoriteLocation struct {
	MerchantId        uuid.UUID  `db:"merchant_id"`
	MerchantName      string     `db:"merchant_name"`
	MerchantUrlName   string     `db:"merchant_url_name"`
	LocationId        int        `db:"location_id"`
	FormattedLocation string     `db:"formatted_location"`
	AverageRating     float64    `db:"average_rating"`
	ReviewCount       int        `db:"review_count"`
	MainImageId       *uuid.UUID `db:"main_image_id"`
}
