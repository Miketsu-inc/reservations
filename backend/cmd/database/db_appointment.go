package database

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Appointment struct {
	Id              int       `json:"ID"`
	ClientId        uuid.UUID `json:"client_id"`
	MerchantId      uuid.UUID `json:"merchant_id"`
	ServiceId       int       `json:"service_id"`
	LocationId      int       `json:"location_id"`
	FromDate        string    `json:"from_date"`
	ToDate          string    `json:"to_date"`
	UserComment     string    `json:"user_comment"`
	MerchantComment string    `json:"merchant_comment"`
}

func (s *service) NewAppointment(ctx context.Context, app Appointment) error {
	query := `
	insert into "Appointment" (client_id, merchant_id, service_id, location_id, from_date, to_date, user_comment, merchant_comment)
	values ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := s.db.ExecContext(ctx, query, app.ClientId, app.MerchantId, app.ServiceId, app.LocationId, app.FromDate, app.ToDate, app.UserComment, app.MerchantComment)
	if err != nil {
		return err
	}
	return nil
}

func (s *service) UpdateMerchantCommentById(ctx context.Context, app_id string, merchant_comment string) error {
	query := `
	update "Appointment" set merchant_comment = $1 where id = $2
	`
	_, err := s.db.ExecContext(ctx, query, merchant_comment, app_id)
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

type AppointmentDetails struct {
	ID              int    `json:"id"`
	FromDate        string `json:"from_date"`
	ToDate          string `json:"to_date"`
	UserComment     string `json:"user_comment"`
	MerchantComment string `json:"merchant_comment"`
	ServiceName     string `json:"service_name"`
	Price           int    `json:"price"`
	FirstName       string `json:"first_name"`
	LastName        string `json:"last_name"`
	PhoneNumber     string `json:"phone_number"`
}

func (s *service) GetAppointmentsByMerchant(ctx context.Context, merchantId uuid.UUID, start string, end string) ([]AppointmentDetails, error) {
	//
	query := `
		select a.id , a.from_date, a.to_date, a.user_comment, a.merchant_comment,
		s.name as service_name, s.price, u.first_name, u.last_name, u.phone_number
		from "Appointment" a
		join "Service" s on a.service_id = s.id
		join "User" u on a.client_id = u.id
		where a.merchant_id = $1 and a.from_date >= $2 AND a.to_date <= $3`

	rows, err := s.db.QueryContext(ctx, query, merchantId, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var appointments []AppointmentDetails

	for rows.Next() {
		var app AppointmentDetails
		if err := rows.Scan(&app.ID, &app.FromDate, &app.ToDate, &app.UserComment, &app.MerchantComment, &app.ServiceName, &app.Price, &app.FirstName, &app.LastName, &app.PhoneNumber); err != nil {
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
