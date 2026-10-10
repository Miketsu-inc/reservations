package db

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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

func (r *locationRepository) NewLocation(ctx context.Context, location domain.Location) (int, error) {
	query := `
	insert into "Location" (merchant_id, name, sequence, country, country_code, region, city, postal_code, address,
		geo_point, geocoding_provider, provider_place_id, formatted_location, contact_email, phone_number, parking_info, arrival_info, is_active)
	values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
	returning id
	`

	var id int
	err := r.db.QueryRow(ctx, query, location.MerchantId, location.Name, location.Sequence, location.Country, location.CountryCode,
		location.Region, location.City, location.PostalCode, location.Address, location.GeoPoint, location.GeocodingProvider,
		location.ProviderPlaceId, location.FormattedLocation, location.ContactEmail, location.PhoneNumber, location.ParkingInfo,
		location.ArrivalInfo, location.IsActive).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("NewLocation: %w", err)
	}

	return id, nil
}

func (r *locationRepository) GetLocation(ctx context.Context, locationId int, merchantId uuid.UUID) (domain.Location, error) {
	query := `
	select * from "Location"
	where id = $1 and merchant_id = $2
	`

	rows, err := r.db.Query(ctx, query, locationId, merchantId)
	if err != nil {
		return domain.Location{}, fmt.Errorf("GetLocation: %w", err)
	}

	location, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[domain.Location])
	if err != nil {
		return domain.Location{}, fmt.Errorf("GetLocation: %w", err)
	}

	return location, nil
}

func (r *locationRepository) GetFirstActiveLocation(ctx context.Context, merchantId uuid.UUID) (domain.Location, error) {
	query := `
	select * from "Location"
	where merchant_id = $1 and is_active is true
	order by sequence, id
	limit 1
	`

	rows, err := r.db.Query(ctx, query, merchantId)
	if err != nil {
		return domain.Location{}, fmt.Errorf("GetFirstActiveLocation: %w", err)
	}

	location, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[domain.Location])
	if err != nil {
		return domain.Location{}, fmt.Errorf("GetFirstActiveLocation: %w", err)
	}

	return location, nil
}

func (r *locationRepository) GetLocations(ctx context.Context, merchantId uuid.UUID) ([]domain.Location, error) {
	query := `
	select * from "Location"
	where merchant_id = $1
	order by is_active desc, sequence, id
	`

	rows, err := r.db.Query(ctx, query, merchantId)
	if err != nil {
		return nil, fmt.Errorf("GetLocations: %w", err)
	}

	locations, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.Location])
	if err != nil {
		return nil, fmt.Errorf("GetLocations: %w", err)
	}

	return locations, nil
}

func (r *locationRepository) UpdateLocation(ctx context.Context, location domain.Location) error {
	query := `
	update "Location"
	set country = $1, country_code = $2, region = $3, city = $4, postal_code = $5, address = $6,
		geo_point = $7, geocoding_provider = $8, provider_place_id = $9, formatted_location = $10
	where id = $11 and merchant_id = $12
	returning id
	`

	var id int
	err := r.db.QueryRow(ctx, query, location.Country, location.CountryCode, location.Region, location.City, location.PostalCode,
		location.Address, location.GeoPoint, location.GeocodingProvider, location.ProviderPlaceId, location.FormattedLocation,
		location.Id, location.MerchantId).Scan(&id)
	if err != nil {
		return fmt.Errorf("UpdateLocation: %w", err)
	}

	return nil
}

func (r *locationRepository) UpdateParkingInfo(ctx context.Context, locationId int, merchantId uuid.UUID, parkingInfo string) error {
	query := `
	update "Location"
	set parking_info = $1
	where id = $2 and merchant_id = $3
	`

	if _, err := r.db.Exec(ctx, query, parkingInfo, locationId, merchantId); err != nil {
		return fmt.Errorf("UpdateParkingInfo: %w", err)
	}

	return nil
}
