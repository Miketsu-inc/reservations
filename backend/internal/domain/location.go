package domain

import (
	"context"

	"github.com/google/uuid"
	"github.com/miketsu-inc/reservations/backend/internal/types"
	"github.com/miketsu-inc/reservations/backend/pkg/db"
)

type LocationRepository interface {
	WithTx(tx db.DBTX) LocationRepository

	NewLocation(ctx context.Context, location Location) error
	GetLocation(ctx context.Context, locationId int, merchantId uuid.UUID) (Location, error)
}

type Location struct {
	Id                int            `db:"id"`
	MerchantId        uuid.UUID      `db:"merchant_id"`
	Country           *string        `db:"country"`
	City              *string        `db:"city"`
	PostalCode        *string        `db:"postal_code"`
	Address           *string        `db:"address"`
	GeoPoint          types.GeoPoint `db:"geo_point"`
	PlaceId           *string        `db:"place_id"`
	FormattedLocation string         `db:"formatted_location"`
	IsPrimary         bool           `db:"is_primary"`
	IsActive          bool           `db:"is_active"`
}
