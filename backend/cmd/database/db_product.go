package database

import (
	"context"
	"strconv"

	"fmt"
	"strings"

	"github.com/google/uuid"
)

type Product struct {
	Id            int       `json:"ID"`
	MerchantId    uuid.UUID `json:"merchant_id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	Price         int       `json:"price"`
	StockQuantity int       `json:"stock_quantity"`
	UsagePerUnit  int       `json:"usage_per_unit"`
	ServiceIds    []int     `json:"service_ids"`
}

// Convert Go []int slice to PostgreSQL array string
func intSliceToPgArray(ints []int) string {
	if len(ints) == 0 {
		return "{}"
	}

	strInts := make([]string, len(ints))
	for i, num := range ints {
		strInts[i] = fmt.Sprintf("%d", num)
	}

	return "{" + strings.Join(strInts, ",") + "}"
}

// convert the PostgresSql array into []int  (array format: {1, NULL, 3, NULL, 5} )
func parsePgArray(arrayStr string) ([]int, error) {
	if arrayStr == "NULL" || arrayStr == "{}" {
		return []int{}, nil
	}

	trimmed := arrayStr[1 : len(arrayStr)-1]
	elements := strings.Split(trimmed, ",")

	result := make([]int, 0, len(elements))
	for _, elem := range elements {
		if elem == "NULL" {
			continue
		}
		val, err := strconv.Atoi(elem)
		if err != nil {
			return nil, err
		}
		result = append(result, val)
	}
	return result, nil
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
    	where s.merchant_id = $1
	)
	insert into "ServiceProduct" (product_id, service_id)
	select inserted_product.id, valid_services.service_id
	from inserted_product, valid_services;
	`

	_, err := s.db.ExecContext(ctx, query, prod.MerchantId, prod.Name, prod.Description, prod.Price, prod.StockQuantity, prod.UsagePerUnit, intSliceToPgArray(prod.ServiceIds))
	if err != nil {
		return err
	}

	return nil
}

func (s *service) UpdateProduct(ctx context.Context, newProduct Product) error {

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	// nolint: errcheck
	defer tx.Rollback()

	productUpdateQuery := `
	update "Product"
	set name = $3, description = $4, price = $5, stock_quantity = $6, usage_per_unit = $7
	where merchant_id = $1 and id = $2
	`
	_, err = tx.ExecContext(ctx, productUpdateQuery, newProduct.MerchantId, newProduct.Id, newProduct.Name, newProduct.Description, newProduct.Price, newProduct.StockQuantity, newProduct.UsagePerUnit)
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

	_, err = tx.ExecContext(ctx, insertServiceQuery, newProduct.MerchantId, newProduct.Id, intSliceToPgArray(newProduct.ServiceIds))
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
	_, err = tx.ExecContext(ctx, deleteServiceQuery, newProduct.MerchantId, newProduct.Id, intSliceToPgArray(newProduct.ServiceIds))
	if err != nil {
		return fmt.Errorf("failed to delete outdated service associations: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (s *service) DeleteProductById(ctx context.Context, merchantId uuid.UUID, productId int) error {
	query := `
		with deleted as (
			delete from "Product"
			where merchant_id = $1 AND id = $2
			returning id
		)
		delete from "ServiceProduct"
		where product_id in (select id from deleted);
	`

	_, err := s.db.ExecContext(ctx, query, merchantId, productId)
	if err != nil {
		return err
	}

	return nil
}

type PublicProduct struct {
	Id            int    `json:"id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	Price         int    `json:"price"`
	StockQuantity int    `json:"stock_quantity"`
	UsagePerUnit  int    `json:"usage_per_unit"`
	ServiceIds    []int  `json:"service_ids"`
}

func (s *service) GetProductsByMerchant(ctx context.Context, merchantId uuid.UUID) ([]PublicProduct, error) {
	query := `select
	p.id, p.name, p.description, p.price, p.stock_quantity, p.usage_per_unit, array_agg(sp.service_id) as service_ids
	from "Product" p
	left join "ServiceProduct" sp on p.id = sp.product_id
	where p.merchant_id = $1
	group by p.id, p.name, p.description, p.stock_quantity, p.usage_per_unit;`

	rows, err := s.db.QueryContext(ctx, query, merchantId)
	if err != nil {
		return []PublicProduct{}, err
	}
	defer rows.Close()

	var products []PublicProduct
	for rows.Next() {
		var prod PublicProduct
		var serviceIdsStr string
		if err := rows.Scan(&prod.Id, &prod.Name, &prod.Description, &prod.Price, &prod.StockQuantity, &prod.UsagePerUnit, &serviceIdsStr); err != nil {
			return []PublicProduct{}, err
		}
		prod.ServiceIds, err = parsePgArray(serviceIdsStr)
		if err != nil {
			return []PublicProduct{}, err
		}
		products = append(products, prod)
	}

	return products, nil
}
