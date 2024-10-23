package database

import (
	"context"

	"github.com/google/uuid"
)

type Appointment struct {
	Id           int       `json:"ID"`
	ClientId     uuid.UUID `json:"client_id"`
	MerchantName string    `json:"merchant_name"`
	TypeName     string    `json:"type_name"`
	LocationName string    `json:"location_name"`
	FromDate     string    `json:"from_date"`
	ToDate       string    `json:"to_date"`
}

func (s *service) NewAppointment(ctx context.Context, app Appointment) error {
	query := `
	insert into "Appointment" (client_id, merchant_name, type_name, location_name, from_date, to_date)
	values ($1, $2, $3, $4, $5, $6)
	`

	_, err := s.db.ExecContext(ctx, query, app.ClientId, app.MerchantName, app.TypeName, app.LocationName, app.FromDate, app.ToDate)
	if err != nil {
		return err
	}
	return nil
}

func (s *service) GetAppointmentsByUser(ctx context.Context, user_id uuid.UUID) ([]Appointment, error) {
	query := `
	select * from "Appointment"
	where "User" = $1
	`

	rows, err := s.db.QueryContext(ctx, query, user_id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var appointments []Appointment
	for rows.Next() {
		var app Appointment
		if err := rows.Scan(&app.ClientId, &app.MerchantName, &app.TypeName, &app.LocationName, &app.FromDate, &app.ToDate); err != nil {
			return nil, err
		}
		appointments = append(appointments, app)
	}

	return appointments, nil
}

func (s *service) GetAppointmentsByMerchant(ctx context.Context, merchant string, start string, end string) ([]Appointment, error) {
	query := `
	select * from "Appointment"
	where merchant_name = $1 and from_date >= $2 and to_date <= $3
	`

	rows, err := s.db.QueryContext(ctx, query, merchant, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var appointments []Appointment

	for rows.Next() {
		var app Appointment
		if err := rows.Scan(&app.ClientId, &app.MerchantName, &app.TypeName, &app.LocationName, &app.FromDate, &app.ToDate); err != nil {
			return nil, err
		}
		appointments = append(appointments, app)
	}

	return appointments, nil
}
