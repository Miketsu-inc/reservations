package database

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Appointment struct {
	Id         int       `json:"ID"`
	ClientId   uuid.UUID `json:"client_id"`
	MerchantId uuid.UUID `json:"merchant_id"`
	ServiceId  int       `json:"service_id"`
	LocationId int       `json:"location_id"`
	FromDate   string    `json:"from_date"`
	ToDate     string    `json:"to_date"`
}

func (s *service) NewAppointment(ctx context.Context, app Appointment) error {
	query := `
	insert into "Appointment" (client_id, merchant_id, service_id, location_id, from_date, to_date)
	values ($1, $2, $3, $4, $5, $6)
	`

	_, err := s.db.ExecContext(ctx, query, app.ClientId, app.MerchantId, app.ServiceId, app.LocationId, app.FromDate, app.ToDate)
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
		if err := rows.Scan(&app.Id, &app.ClientId, &app.MerchantId, &app.ServiceId, &app.LocationId, &app.FromDate, &app.ToDate); err != nil {
			return nil, err
		}
		appointments = append(appointments, app)
	}

	return appointments, nil
}

func (s *service) GetAppointmentsByMerchant(ctx context.Context, merchantId uuid.UUID, start string, end string) ([]Appointment, error) {
	query := `
	select * from "Appointment"
	where merchant_id = $1 and from_date >= $2 and to_date <= $3
	`

	rows, err := s.db.QueryContext(ctx, query, merchantId, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var appointments []Appointment

	for rows.Next() {
		var app Appointment
		if err := rows.Scan(&app.Id, &app.ClientId, &app.MerchantId, &app.ServiceId, &app.LocationId, &app.FromDate, &app.ToDate); err != nil {
			return nil, err
		}
		appointments = append(appointments, app)
	}

	return appointments, nil
}

type AppointmentTime struct {
	From_date time.Time
	To_date   time.Time
}

func (s *service) GetReservedTimes(ctx context.Context, merchant_id uuid.UUID, location_id int, day time.Time) ([]AppointmentTime, error) {
	query := `
    select  from_date AT TIME ZONE 'UTC' AS from_date_utc, to_date AT TIME ZONE 'UTC' AS to_date_utc from "Appointment"
    where merchant_id = $1 and location_id = $2 and DATE(from_date) = $3
    ORDER BY from_date`

	rows, err := s.db.QueryContext(ctx, query, merchant_id, location_id, day)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bookedApps []AppointmentTime
	for rows.Next() {
		var app AppointmentTime
		if err := rows.Scan(&app.From_date, &app.To_date); err != nil {
			return nil, err
		}
		bookedApps = append(bookedApps, app)
	}

	return bookedApps, nil
}
