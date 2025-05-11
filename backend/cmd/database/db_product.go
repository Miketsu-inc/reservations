package database

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Product struct {
	Id            int       `json:"ID"`
	MerchantId    uuid.UUID `json:"merchant_id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	Price         int       `json:"price"`
	StockQuantity int       `json:"stock_quantity"`
	UsagePerUnit  int       `json:"usage_per_unit"`
	ServiceIds    []*int    `json:"service_ids"`
	DeletedOn     *string   `json:"deleted_on"`
}

func (s *service) NewProduct(ctx context.Context, prod Product) error {

	query := `
	with inserted_product as (
	    insert into "Product" (merchant_id, name, description, price, stock_quantity, usage_per_unit)
	    values ($1, $2, $3, $4, $5, $6)
	    returning ID
	),
	valid_services as (
    	select s.id as service_id
    	from unnest($7::bigint[]) as input_id
    	join "Service" s on s.id = input_id
    	where s.merchant_id = $1 and s.deleted_on is null
	)
	insert into "ServiceProduct" (product_id, service_id)
	select inserted_product.id, valid_services.service_id
	from inserted_product, valid_services;
	`

	_, err := s.db.Exec(ctx, query, prod.MerchantId, prod.Name, prod.Description, prod.Price, prod.StockQuantity, prod.UsagePerUnit, prod.ServiceIds)
	if err != nil {
		return err
	}

	return nil
}

func (s *service) UpdateProduct(ctx context.Context, newProduct Product) error {

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	// nolint: errcheck
	defer tx.Rollback(ctx)

	productUpdateQuery := `
	update "Product"
	set name = $3, description = $4, price = $5, stock_quantity = $6, usage_per_unit = $7
	where merchant_id = $1 and id = $2 and deleted_on is null
	`
	_, err = tx.Exec(ctx, productUpdateQuery, newProduct.MerchantId, newProduct.Id, newProduct.Name, newProduct.Description, newProduct.Price, newProduct.StockQuantity, newProduct.UsagePerUnit)
	if err != nil {
		return fmt.Errorf("failed to update product: %v", err)
	}

	insertServiceQuery := `
	with valid_services as (
    	select s.id as service_id
    	from unnest($3::bigint[]) as input_id
    	join "Service" s on s.id = input_id
    	where s.merchant_id = $1
	)
	insert into "ServiceProduct" (product_id, service_id)
	select $2, service_id
	from valid_services
	on conflict (product_id, service_id) do nothing;
	`

	_, err = tx.Exec(ctx, insertServiceQuery, newProduct.MerchantId, newProduct.Id, newProduct.ServiceIds)
	if err != nil {
		return fmt.Errorf("failed to insert new service links: %v", err)
	}

	deleteServiceQuery := `
	with valid_services as (
    	select s.id as service_id
    	from unnest($3::bigint[]) as input_id
    	join "Service" s on s.id = input_id
    	where s.merchant_id = $1
	)
	delete from "ServiceProduct"
	where product_id = $2
	and not exists (
		select 1
		from valid_services
		where valid_services.service_id = "ServiceProduct".service_id
	);
	`
	_, err = tx.Exec(ctx, deleteServiceQuery, newProduct.MerchantId, newProduct.Id, newProduct.ServiceIds)
	if err != nil {
		return fmt.Errorf("failed to delete outdated service associations: %v", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

func (s *service) DeleteProductById(ctx context.Context, merchantId uuid.UUID, productId int) error {
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

	_, err := s.db.Exec(ctx, query, time.Now().UTC(), merchantId, productId)
	if err != nil {
		return err
	}

	return nil
}

type PublicProduct struct {
	Id            int    `json:"id" db:"id"`
	Name          string `json:"name" db:"name"`
	Description   string `json:"description" db:"description"`
	Price         int    `json:"price" db:"price"`
	StockQuantity int    `json:"stock_quantity" db:"stock_quantity"`
	UsagePerUnit  int    `json:"usage_per_unit" db:"usage_per_unit"`
	ServiceIds    []*int `json:"service_ids" db:"service_ids"`
}

// TODO: this should use pgx helpers
func (s *service) GetProductsByMerchant(ctx context.Context, merchantId uuid.UUID) ([]PublicProduct, error) {
	query := `
	select p.id, p.name, p.description, p.price, p.stock_quantity, p.usage_per_unit, array_agg(sp.service_id) as service_ids
	from "Product" p
	left join "ServiceProduct" sp on p.id = sp.product_id
	where p.merchant_id = $1 and deleted_on is null
	group by p.id, p.name, p.description, p.stock_quantity, p.usage_per_unit`

	rows, _ := s.db.Query(ctx, query, merchantId)
	products, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (PublicProduct, error) {
		var product PublicProduct
		err := row.Scan(&product.Id, &product.Name, &product.Description, &product.Price, &product.StockQuantity, &product.UsagePerUnit, &product.ServiceIds)

		if product.ServiceIds[0] == nil {
			product.ServiceIds = []*int{}
		}

		return product, err
	})
	if err != nil {
		return []PublicProduct{}, err
	}

	// if products array is empty the encoded json field will be null
	// unless an empty slice is supplied to it
	if len(products) == 0 {
		products = []PublicProduct{}
	}

	return products, nil
}
