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
	insert into "Service" (ID, merchant_id, name, duration, price)
	values 
	`
	values := []string{}
	args := []interface{}{}
	for i, service := range services {
		//placeholder for values of each row
		values = append(values, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d)", i*5+1, i*5+2, i*5+3, i*5+4, i*5+5))
		args = append(args, service.Id, service.MerchantId, service.Name, service.Duration, service.Price)
	}
	query += strings.Join(values, ",")

	_, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	return nil
}