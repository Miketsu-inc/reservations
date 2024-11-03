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
	insert into "Location" (ID, merchant_id, country, city, postal_code, address)
	values ($1, $2, $3, $4, $5, $6)
	`

	_, err := s.db.ExecContext(ctx, query, location.Id, location.MerchantId, location.Country, location.City, location.PostalCode, location.Address)
	if err != nil {
		return err
	}

	return nil
}
