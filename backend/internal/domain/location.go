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
	Id                int            `json:"ID"`
	MerchantId        uuid.UUID      `json:"merchant_id"`
	Country           *string        `json:"country"`
	City              *string        `json:"city"`
	PostalCode        *string        `json:"postal_code"`
	Address           *string        `json:"address"`
	GeoPoint          types.GeoPoint `json:"geo_point"`
	PlaceId           *string        `json:"place_id"`
	FormattedLocation string         `json:"formatted_location"`
	IsPrimary         bool           `json:"is_primary"`
	IsActive          bool           `json:"is_active"`
}
