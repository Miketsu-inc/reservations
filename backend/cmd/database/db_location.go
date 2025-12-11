package database

import (
	"context"

	"github.com/google/uuid"
	"github.com/miketsu-inc/reservations/backend/cmd/types"
)

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

func (s *service) NewLocation(ctx context.Context, location Location) error {
	query := `
	insert into "Location" (merchant_id, country, city, postal_code, address, geo_point, place_id, formatted_location, is_primary, is_active)
	values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := s.db.Exec(ctx, query, location.MerchantId, location.Country, location.City, location.PostalCode, location.Address, location.GeoPoint,
		location.PlaceId, location.FormattedLocation, location.IsPrimary, location.IsActive)
	if err != nil {
		return err
	}

	return nil
}

func (s *service) GetLocationById(ctx context.Context, locationId int, merchantId uuid.UUID) (Location, error) {
	query := `
	select * from "Location"
	where id = $1 and merchant_id = $2
	`

	var location Location
	err := s.db.QueryRow(ctx, query, locationId, merchantId).Scan(&location.Id, &location.MerchantId, &location.Country, &location.City, &location.PostalCode,
		&location.Address, &location.GeoPoint, &location.PlaceId, &location.FormattedLocation, &location.IsPrimary, &location.IsActive)
	if err != nil {
		return Location{}, err
	}

	return location, nil
}
