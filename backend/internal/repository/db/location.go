package db

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/miketsu-inc/reservations/backend/internal/domain"
	"github.com/miketsu-inc/reservations/backend/pkg/db"
)

type locationRepository struct {
	db db.DBTX
}

func NewLocationRepository(db db.DBTX) domain.LocationRepository {
	return &locationRepository{db: db}
}

func (r *locationRepository) WithTx(tx db.DBTX) domain.LocationRepository {
	return &locationRepository{db: tx}
}

func (r *locationRepository) NewLocation(ctx context.Context, location domain.Location) error {
	query := `
	insert into "Location" (merchant_id, country, city, postal_code, address, geo_point, place_id, formatted_location, is_primary, is_active)
	values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := r.db.Exec(ctx, query, location.MerchantId, location.Country, location.City, location.PostalCode, location.Address,
		location.GeoPoint, location.PlaceId, location.FormattedLocation, location.IsPrimary, location.IsActive)
	if err != nil {
		return fmt.Errorf("NewLocation: %w", err)
	}

	return nil
}

func (r *locationRepository) GetLocation(ctx context.Context, locationId int, merchantId uuid.UUID) (domain.Location, error) {
	query := `
	select * from "Location"
	where id = $1 and merchant_id = $2
	`

	var location domain.Location
	err := r.db.QueryRow(ctx, query, locationId, merchantId).Scan(&location.Id, &location.MerchantId, &location.Country, &location.City,
		&location.PostalCode, &location.Address, &location.GeoPoint, &location.PlaceId, &location.FormattedLocation, &location.IsPrimary,
		&location.IsActive)
	if err != nil {
		return domain.Location{}, fmt.Errorf("GetLocation: %w", err)
	}

	return location, nil
}
