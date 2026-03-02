package db

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/miketsu-inc/reservations/backend/internal/domain"
	"github.com/miketsu-inc/reservations/backend/pkg/db"
)

type productRepository struct {
	db db.DBTX
}

func NewProductRepository(db db.DBTX) domain.ProductRepository {
	return &productRepository{db: db}
}

func (r *productRepository) WithTx(tx db.DBTX) domain.ProductRepository {
	return &productRepository{db: tx}
}

func (r *productRepository) NewProduct(ctx context.Context, prod domain.Product) error {

	query := `
	insert into "Product" (merchant_id, name, description, price, unit, max_amount, current_amount)
	values ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.db.Exec(ctx, query, prod.MerchantId, prod.Name, prod.Description, prod.Price, prod.Unit, prod.MaxAmount, prod.CurrentAmount)
	if err != nil {
		return err
	}

	return nil
}

func (r *productRepository) UpdateProduct(ctx context.Context, newProduct domain.Product) error {

	query := `
	update "Product"
	set name = $3, description = $4, price = $5, unit = $6, max_amount = $7, current_amount = $8
	where merchant_id = $1 and id = $2 and deleted_on is null
	`
	_, err := r.db.Exec(ctx, query, newProduct.MerchantId, newProduct.Id, newProduct.Name, newProduct.Description, newProduct.Price, newProduct.Unit, newProduct.MaxAmount, newProduct.CurrentAmount)
	if err != nil {
		return fmt.Errorf("failed to update product: %v", err)
	}

	return nil
}

func (r *productRepository) DeleteProduct(ctx context.Context, merchantId uuid.UUID, productId int) error {
	query := `
		with deleted as (
			update "Product"
			set deleted_on = $1
			where merchant_id = $2 AND id = $3
			returning id
		)
		delete from "ServiceProduct"
		where product_id in (select id from deleted);
	`

	_, err := r.db.Exec(ctx, query, time.Now().UTC(), merchantId, productId)
	if err != nil {
		return err
	}

	return nil
}

// TODO: this should use pgx helpers
func (r *productRepository) GetProducts(ctx context.Context, merchantId uuid.UUID) ([]domain.ProductInfo, error) {
	query := `
	select p.id, p.name, p.description, p.price, p.unit, p.max_amount, p.current_amount,
	coalesce(
        json_agg(
		    json_build_object(
	            'id', s.id,
	            'name', s.name,
	            'color', s.color
			)
	    ) filter (where s.id is not null),
	'[]'::json) as services
	from "Product" p
	left join "ServiceProduct" sp on p.id = sp.product_id
	left join "Service" s on sp.service_id = s.id and s.deleted_on is null
	where p.merchant_id = $1 and p.deleted_on is null
	group by p.id`

	rows, _ := r.db.Query(ctx, query, merchantId)
	products, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (domain.ProductInfo, error) {
		var product domain.ProductInfo
		var servicesJSON []byte

		err := row.Scan(
			&product.Id,
			&product.Name,
			&product.Description,
			&product.Price,
			&product.Unit,
			&product.MaxAmount,
			&product.CurrentAmount,
			&servicesJSON,
		)
		if err != nil {
			return product, err
		}

		// Parse the JSON
		if len(servicesJSON) > 0 {
			err = json.Unmarshal(servicesJSON, &product.Services)
			if err != nil {
				return product, err
			}
		} else {
			product.Services = []domain.ServiceInfoForProducts{}
		}

		return product, nil
	})

	if err != nil {
		return []domain.ProductInfo{}, err
	}

	// if products array is empty the encoded json field will be null
	// unless an empty slice is supplied to it
	if len(products) == 0 {
		products = []domain.ProductInfo{}
	}
	return products, nil
}

func (r *productRepository) GetLowStockProducts(ctx context.Context, merchantId uuid.UUID) ([]domain.LowStockProduct, error) {
	query := `
	select p.id, p.name, p.max_amount, p.current_amount, p.unit, (p.current_amount::float / p.max_amount) as fill_ratio from "Product" p
	where  p.merchant_id = $1 and p.deleted_on is null and p.max_amount > 0 and (p.current_amount::float / p.max_amount) < 0.4
	order by fill_ratio asc
	`

	rows, _ := r.db.Query(ctx, query, merchantId)
	products, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.LowStockProduct])
	if err != nil {
		return []domain.LowStockProduct{}, err
	}

	return products, nil
}
