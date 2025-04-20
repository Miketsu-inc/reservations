package database

import (
	"context"

	"github.com/google/uuid"
)

type Location struct {
	Id         int       `json:"ID"`
	MerchantId uuid.UUID `json:"merchant_id"`
	Country    string    `json:"country"`
	City       string    `json:"city"`
	PostalCode string    `json:"postal_code"`
	Address    string    `json:"address"`
}

func (s *service) NewLocation(ctx context.Context, location Location) error {
	query := `
	insert into "Location" (merchant_id, country, city, postal_code, address)
	values ($1, $2, $3, $4, $5)
	`

	_, err := s.db.Exec(ctx, query, location.MerchantId, location.Country, location.City, location.PostalCode, location.Address)
	if err != nil {
		return err
	}

	return nil
}

func (s *service) GetLocationById(ctx context.Context, locationId int) (Location, error) {
	query := `
	select * from "Location"
	where id = $1
	`

	var location Location
	err := s.db.QueryRow(ctx, query, locationId).Scan(&location.Id, &location.MerchantId, &location.Country, &location.City,
		&location.PostalCode, &location.Address)
	if err != nil {
		return Location{}, err
	}

	return location, nil
}
