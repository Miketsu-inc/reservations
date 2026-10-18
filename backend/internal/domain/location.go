package domain

import (
	"context"

	"github.com/google/uuid"
	"github.com/miketsu-inc/reservations/backend/internal/types"
	"github.com/miketsu-inc/reservations/backend/pkg/db"
)

type LocationRepository interface {
	WithTx(tx db.DBTX) LocationRepository

	NewLocation(ctx context.Context, location Location) (int, error)
	GetLocation(ctx context.Context, locationId int, merchantId uuid.UUID) (Location, error)
	LockLocation(ctx context.Context, locationId int, merchantId uuid.UUID) error
	GetDefaultLocation(ctx context.Context, merchantId uuid.UUID) (Location, error)
	GetLocations(ctx context.Context, merchantId uuid.UUID) ([]Location, error)
	UpdateLocation(ctx context.Context, location Location) error
	UpdateParkingInfo(ctx context.Context, locationId int, merchantId uuid.UUID, parkingInfo string) error
}

type Location struct {
	Id                int            `db:"id"`
	MerchantId        uuid.UUID      `db:"merchant_id"`
	Name              string         `db:"name"`
	Sequence          int            `db:"sequence"`
	Country           *string        `db:"country"`
	CountryCode       *string        `db:"country_code"`
	Region            *string        `db:"region"`
	City              *string        `db:"city"`
	PostalCode        *string        `db:"postal_code"`
	Address           *string        `db:"address"`
	GeoPoint          types.GeoPoint `db:"geo_point"`
	GeocodingProvider *string        `db:"geocoding_provider"`
	ProviderPlaceId   *string        `db:"provider_place_id"`
	FormattedLocation string         `db:"formatted_location"`
	ContactEmail      *string        `db:"contact_email"`
	PhoneNumber       *string        `db:"phone_number"`
	ParkingInfo       string         `db:"parking_info"`
	ArrivalInfo       string         `db:"arrival_info"`
	IsActive          bool           `db:"is_active"`
}
