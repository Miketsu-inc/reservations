package domain

import (
	"context"

	"github.com/google/uuid"
	"github.com/miketsu-inc/reservations/backend/pkg/currencyx"
)

type ProductRepository interface {
	NewProduct(context.Context, Product) error
	UpdateProduct(context.Context, Product) error
	DeleteProductById(context.Context, uuid.UUID, int) error

	GetProductsByMerchant(context.Context, uuid.UUID) ([]ProductInfo, error)
}

type Product struct {
	Id            int              `json:"ID"`
	MerchantId    uuid.UUID        `json:"merchant_id"`
	Name          string           `json:"name"`
	Description   string           `json:"description"`
	Price         *currencyx.Price `json:"price"`
	Unit          string           `json:"unit"`
	MaxAmount     int              `json:"max_amount"`
	CurrentAmount int              `json:"current_amount"`
	DeletedOn     *string          `json:"deleted_on"`
}

type ProductInfo struct {
	Id            int                      `json:"id" db:"id"`
	Name          string                   `json:"name" db:"name"`
	Description   string                   `json:"description" db:"description"`
	Price         *currencyx.Price         `json:"price" db:"price"`
	Unit          string                   `json:"unit" db:"unit"`
	MaxAmount     int                      `json:"max_amount" db:"max_amount"`
	CurrentAmount int                      `json:"current_amount" db:"current_amount"`
	Services      []ServiceInfoForProducts `json:"services" db:"services"`
}

type MinimalProductInfo struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
	Unit string `json:"unit"`
}

type MinimalProductInfoWithUsage struct {
	Id         int    `json:"id"`
	Name       string `json:"name"`
	Unit       string `json:"unit"`
	AmountUsed int    `json:"amount_used"`
}
