package database

import (
	"context"

	"github.com/google/uuid"
)

type Merchant struct {
	Id           uuid.UUID       `json:"ID"`
	Name         string          `json:"name"`
	UrlName      string          `json:"url_name"`
	OwnerId      uuid.UUID       `json:"owner_id"`
	ContactEmail string          `json:"contact_email"`
	Settings     map[string]bool `json:"settings"`
}

func (s *service) NewMerchant(ctx context.Context, merchant Merchant) error {
	query := `
	insert into "Merchant" (ID, name, url_name, owner_id, contact_email, settings)
	values ($1, $2, $3, $4, $5, $6)
	`

	_, err := s.db.ExecContext(ctx, query, merchant.Id, merchant.Name, merchant.UrlName, merchant.OwnerId, merchant.ContactEmail, merchant.Settings)
	if err != nil {
		return err
	}

	return nil
}

func (s *service) GetMerchantIdByUrlName(ctx context.Context, UrlName string) (uuid.UUID, error) {
	query := `
	select id from "Merchant"
	where url_name = $1
	`

	var merchantId uuid.UUID
	err := s.db.QueryRowContext(ctx, query, UrlName).Scan(&merchantId)
	if err != nil {
		return uuid.Nil, err
	}

	return merchantId, nil
}

func (s *service) GetMerchantIdByOwnerId(ctx context.Context, ownerId uuid.UUID) (uuid.UUID, error) {
	query := `
	select id from "Merchant"
	where owner_id = $1
	`

	var merchantId uuid.UUID
	err := s.db.QueryRowContext(ctx, query, ownerId).Scan(&merchantId)
	if err != nil {
		return uuid.UUID{}, err
	}

	return merchantId, nil
}

func (s *service) GetMerchantById(ctx context.Context, merchantId uuid.UUID) (Merchant, error) {
	query := `
	select * from "Merchant"
	where id = $1
	`

	var merchant Merchant
	err := s.db.QueryRowContext(ctx, query, merchantId).Scan(&merchant.Id, &merchant.Name, &merchant.UrlName, &merchant.OwnerId, &merchant.ContactEmail, &merchant.Settings)
	if err != nil {
		return Merchant{}, err
	}

	return merchant, nil
}
