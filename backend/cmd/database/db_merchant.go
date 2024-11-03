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

func (s *service) GetMerchantByUrlName(ctx context.Context, UrlName string) (Merchant, error) {
	query := `
	select * from "Merchant"
	where url_name = $1
	`

	var merchant Merchant
	err := s.db.QueryRowContext(ctx, query, UrlName).Scan(&merchant.Id, &merchant.Name, &merchant.UrlName, &merchant.OwnerId, &merchant.ContactEmail, &merchant.Settings)
	if err != nil {
		return Merchant{}, err
	}

	return merchant, nil
}
