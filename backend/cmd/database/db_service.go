package database

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type Service struct {
	Id         int       `json:"ID"`
	MerchantId uuid.UUID `json:"merchant_id"`
	Name       string    `json:"name"`
	Duration   string    `json:"duration"`
	Price      string    `json:"price"`
}

func (s *service) NewServices(ctx context.Context, services []Service) error {
	query := `
	insert into "Service" (merchant_id, name, duration, price)
	values
	`

	values := []string{}
	args := []interface{}{}
	for i, service := range services {
		//placeholder for values of each row
		values = append(values, fmt.Sprintf("($%d, $%d, $%d, $%d)", i*4+1, i*4+2, i*4+3, i*4+4))
		args = append(args, service.MerchantId, service.Name, service.Duration, service.Price)
	}
	query += strings.Join(values, ",")

	_, err := s.db.ExecContext(ctx, query, args...)
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
	err := s.db.QueryRowContext(ctx, query, serviceID).Scan(&serv.Id, &serv.MerchantId, &serv.Name, &serv.Duration, &serv.Price)
	if err != nil {
		return Service{}, err
	}

	return serv, nil
}
