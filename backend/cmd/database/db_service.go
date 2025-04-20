package database

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Service struct {
	Id          int       `json:"ID"`
	MerchantId  uuid.UUID `json:"merchant_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Color       string    `json:"color"`
	Duration    int       `json:"duration"`
	Price       int       `json:"price"`
	Cost        int       `json:"cost"`
	DeletedOn   *string   `json:"deleted_on"`
}

func (s *service) NewService(ctx context.Context, serv Service) error {
	query := `
	insert into "Service" (merchant_id, name, description, color, duration, price, cost)
	values ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := s.db.Exec(ctx, query, serv.MerchantId, serv.Name, serv.Description, serv.Color, serv.Duration, serv.Price, serv.Cost)
	if err != nil {
		return err
	}

	return nil
}

func (s *service) GetServiceById(ctx context.Context, serviceID int) (Service, error) {
	query := `
	select * from "Service"
	where id = $1
	`

	var serv Service
	err := s.db.QueryRow(ctx, query, serviceID).Scan(&serv.Id, &serv.MerchantId, &serv.Name, &serv.Description, &serv.Color,
		&serv.Duration, &serv.Price, &serv.Cost, &serv.DeletedOn)
	if err != nil {
		return Service{}, err
	}

	return serv, nil
}

type PublicService struct {
	Id          int    `json:"id" db:"id"`
	Name        string `json:"name" db:"name"`
	Description string `json:"description" db:"description"`
	Color       string `json:"color" db:"color"`
	Duration    int    `json:"duration" db:"duration"`
	Price       int    `json:"price" db:"price"`
	Cost        int    `json:"cost" db:"cost"`
}

func (s *service) GetServicesByMerchantId(ctx context.Context, merchantId uuid.UUID) ([]PublicService, error) {
	query := `
	select id, name, description, color, duration, price, cost from "Service"
	where merchant_id = $1 and deleted_on is null
	`

	rows, _ := s.db.Query(ctx, query, merchantId)
	services, err := pgx.CollectRows(rows, pgx.RowToStructByName[PublicService])
	if err != nil {
		return []PublicService{}, err
	}

	// if services array is empty the encoded json field will be null
	// unless an empty slice is supplied to it
	if len(services) == 0 {
		services = []PublicService{}
	}

	return services, nil
}

func (s *service) DeleteServiceById(ctx context.Context, merchantId uuid.UUID, serviceId int) error {
	query := `
	update "Service"
	set deleted_on = $1
	where merchant_id = $2 and ID = $3
	`

	_, err := s.db.Exec(ctx, query, time.Now().UTC(), merchantId, serviceId)
	if err != nil {
		return err
	}

	return nil
}

func (s *service) UpdateServiceById(ctx context.Context, serv Service) error {
	query := `
	update "Service"
	set name = $3, description = $4, color = $5, duration = $6, price = $7, cost = $8
	where ID = $1 and merchant_id = $2 and deleted_on is null
	`

	_, err := s.db.Exec(ctx, query, serv.Id, serv.MerchantId, serv.Name, serv.Description, serv.Color, serv.Duration, serv.Price, serv.Cost)
	if err != nil {
		return err
	}

	return nil
}
