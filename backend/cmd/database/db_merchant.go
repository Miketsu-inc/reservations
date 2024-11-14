package database

import (
	"context"

	"github.com/google/uuid"
)

type Merchant struct {
	Id           uuid.UUID `json:"ID"`
	Name         string    `json:"name"`
	UrlName      string    `json:"url_name"`
	OwnerId      uuid.UUID `json:"owner_id"`
	ContactEmail string    `json:"contact_email"`
	// Settings     map[string]bool `json:"settings"`
}

func (s *service) NewMerchant(ctx context.Context, merchant Merchant) error {
	query := `
	insert into "Merchant" (ID, name, url_name, owner_id, contact_email)
	values ($1, $2, $3, $4, $5)
	`

	_, err := s.db.ExecContext(ctx, query, merchant.Id, merchant.Name, merchant.UrlName, merchant.OwnerId, merchant.ContactEmail)
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
	err := s.db.QueryRowContext(ctx, query, merchantId).Scan(&merchant.Id, &merchant.Name, &merchant.UrlName, &merchant.OwnerId, &merchant.ContactEmail)
	if err != nil {
		return Merchant{}, err
	}

	return merchant, nil
}

type MerchantInfo struct {
	Name         string `json:"merchant_name"`
	UrlName      string `json:"url_name"`
	ContactEmail string `json:"contact_email"`

	LocationId int    `json:"location_id"`
	Country    string `json:"country"`
	City       string `json:"city"`
	PostalCode string `json:"postal_code"`
	Address    string `json:"address"`

	Services []Service `json:"services"`
}

// this should and will be refactored
func (s *service) GetAllMerchantInfo(ctx context.Context, merchantId uuid.UUID) (MerchantInfo, error) {
	query := `
	select m.name, m.url_name, m.contact_email, l.id as location_id, l.country, l.city, l.postal_code, l.address from "Merchant" m
	inner join "Location" l on m.id = l.merchant_id
	`

	var mi MerchantInfo
	err := s.db.QueryRowContext(ctx, query).Scan(&mi.Name, &mi.UrlName, &mi.ContactEmail, &mi.LocationId, &mi.Country, &mi.City, &mi.PostalCode, &mi.Address)
	if err != nil {
		return MerchantInfo{}, err
	}

	query = `
	select * from "Service"
	where merchant_id = $1
	`

	rows, err := s.db.QueryContext(ctx, query, merchantId)
	if err != nil {
		return MerchantInfo{}, err
	}
	defer rows.Close()

	var services []Service
	for rows.Next() {
		var s Service
		if err := rows.Scan(&s.Id, &s.MerchantId, &s.Name, &s.Duration, &s.Price); err != nil {
			return MerchantInfo{}, err
		}

		services = append(services, s)
	}

	mi.Services = services
	return mi, nil
}
