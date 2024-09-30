package database

import (
	"context"
)

type Appointment struct {
	User            string `json:"user"`
	Merchant        string `json:"merchant"`
	AppointmentType string `json:"type"`
	Location        string `json:"location"`
	FromDate        string `json:"from_date"`
	ToDate          string `json:"to_date"`
}

func (s *service) NewAppointment(ctx context.Context, app Appointment) error {
	query := `
	insert into appointment("user", merchant, type, location, from_date, to_date)
	values ($1, $2, $3, $4, $5, $6)
	`

	_, err := s.db.ExecContext(ctx, query, app.User, app.Merchant, app.AppointmentType, app.Location, app.FromDate, app.ToDate)
	if err != nil {
		return err
	}
	return nil
}

func (s *service) GetAppointmentsByUser(ctx context.Context, user string) ([]Appointment, error) {
	query := `
	select * from appointment where "user" = $1
	`

	rows, err := s.db.QueryContext(ctx, query, user)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var appointments []Appointment
	for rows.Next() {
		var app Appointment
		if err := rows.Scan(&app.User, &app.Merchant, &app.AppointmentType, &app.Location, &app.FromDate, &app.ToDate); err != nil {
			return nil, err
		}
		appointments = append(appointments, app)
	}

	return appointments, nil
}

func (s *service) GetAppointmentsByMerchant(ctx context.Context, merchant, start, end string) ([]Appointment, error) {
	query := `
	select * from appointment 
where merchant = $1
and from_date >= $2 
		and to_date <= $3
	`

	rows, err := s.db.QueryContext(ctx, query, merchant, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var appointments []Appointment

	for rows.Next() {
		var app Appointment
		if err := rows.Scan(&app.User, &app.Merchant, &app.AppointmentType, &app.Location, &app.FromDate, &app.ToDate); err != nil {
			return nil, err
		}
		appointments = append(appointments, app)
	}

	return appointments, nil
}
