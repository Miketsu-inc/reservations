package database

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Appointment struct {
	Id                    int       `json:"ID" db:"id"`
	UserId                uuid.UUID `json:"user_id" db:"user_id"`
	MerchantId            uuid.UUID `json:"merchant_id" db:"merchant_id"`
	ServiceId             int       `json:"service_id" db:"service_id"`
	LocationId            int       `json:"location_id" db:"location_id"`
	FromDate              string    `json:"from_date" db:"from_date"`
	ToDate                string    `json:"to_date" db:"to_date"`
	UserNote              string    `json:"user_note" db:"user_note"`
	MerchantNote          string    `json:"merchant_note" db:"merchant_note"`
	PriceThen             int       `json:"price_then" db:"price_then"`
	CostThen              int       `json:"cost_then" db:"cost_then"`
	CancelledByUserOn     string    `json:"cancelled_by_user_on" db:"cancelled_by_merchant_on"`
	CancelledByMerchantOn string    `json:"cancelled_by_merchant_on" db:"cancelled_by_merchant_on"`
	CancellationReason    string    `json:"cancellation_reason" db:"cancellation_reason"`
	TransferredTo         uuid.UUID `json:"transferred_to" db:"transferred_to"`
	EmailId               uuid.UUID `json:"email_id" db:"email_id"`
}

func (s *service) NewAppointment(ctx context.Context, app Appointment) (int, error) {
	query := `
	insert into "Appointment" (user_id, merchant_id, service_id, location_id, from_date, to_date, user_note, merchant_note, price_then, cost_then)
	values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	returning id`

	var id int
	err := s.db.QueryRow(ctx, query, app.UserId, app.MerchantId, app.ServiceId, app.LocationId, app.FromDate, app.ToDate, app.UserNote, app.MerchantNote, app.PriceThen, app.CostThen).Scan(&id)

	if err != nil {
		return 0, err
	}
	return id, nil
}

func (s *service) UpdateAppointmentData(ctx context.Context, merchantId uuid.UUID, appointmentId int, merchant_note string, from_date string, to_date string) error {
	query := `
	update "Appointment" set merchant_note = $1, from_date = $2, to_date = $3
	where id = $4 and merchant_id = $5 and cancelled_by_user_on is null and cancelled_by_merchant_on is null
	`
	_, err := s.db.Exec(ctx, query, merchant_note, from_date, to_date, appointmentId, merchantId)
	if err != nil {
		return err
	}

	return nil
}

type AppointmentDetails struct {
	ID              int       `json:"id" db:"id"`
	FromDate        time.Time `json:"from_date" db:"from_date"`
	ToDate          time.Time `json:"to_date" db:"to_date"`
	UserNote        string    `json:"user_note" db:"user_note"`
	MerchantNote    string    `json:"merchant_note" db:"merchant_note"`
	ServiceName     string    `json:"service_name" db:"service_name"`
	ServiceColor    string    `json:"service_color" db:"service_color"`
	ServiceDuration int       `json:"service_duration" db:"service_duration"`
	Price           int       `json:"price" db:"price"`
	Cost            int       `json:"cost" db:"cost"`
	FirstName       string    `json:"first_name" db:"first_name"`
	LastName        string    `json:"last_name" db:"last_name"`
	PhoneNumber     string    `json:"phone_number" db:"phone_number"`
}

func (s *service) GetAppointmentsByMerchant(ctx context.Context, merchantId uuid.UUID, start string, end string) ([]AppointmentDetails, error) {
	query := `
	select a.id, a.from_date, a.to_date, a.user_note, a.merchant_note, a.price_then as price, a.cost_then as cost,
	s.name as service_name, s.color as service_color, s.duration as service_duration, u.first_name, u.last_name, u.phone_number
	from "Appointment" a
	join "Service" s on a.service_id = s.id
	join "User" u on a.user_id = u.id
	where a.merchant_id = $1 and a.from_date >= $2 AND a.to_date <= $3 AND a.cancelled_by_user_on is null and a.cancelled_by_merchant_on is null`

	rows, _ := s.db.Query(ctx, query, merchantId, start, end)
	appointments, err := pgx.CollectRows(rows, pgx.RowToStructByName[AppointmentDetails])
	if err != nil {
		return nil, err
	}

	return appointments, nil
}

type AppointmentTime struct {
	From_date time.Time
	To_date   time.Time
}

func (s *service) GetReservedTimes(ctx context.Context, merchant_id uuid.UUID, location_id int, day time.Time) ([]AppointmentTime, error) {
	query := `
    select from_date, to_date from "Appointment"
    where merchant_id = $1 and location_id = $2 and DATE(from_date) = $3 and cancelled_by_user_on is null and cancelled_by_merchant_on is null
    ORDER BY from_date`

	rows, _ := s.db.Query(ctx, query, merchant_id, location_id, day)
	bookedApps, err := pgx.CollectRows(rows, pgx.RowToStructByName[AppointmentTime])
	if err != nil {
		return nil, err
	}

	return bookedApps, nil
}

func (s *service) TransferDummyAppointments(ctx context.Context, merchantId uuid.UUID, fromUser uuid.UUID, toUser uuid.UUID) error {
	query := `
	update "Appointment" a
	set transferred_to = $3
	from "User" u
	where a.user_id = u.id and a.merchant_id = $1 and a.user_id = $2 and u.is_dummy = true
	`

	_, err := s.db.Exec(ctx, query, merchantId, fromUser, toUser)
	if err != nil {
		return err
	}

	return nil
}

func (s *service) CancelAppointmentByMerchant(ctx context.Context, merchantId uuid.UUID, appointmentId int, cancellationReason string) error {
	query := `
	update "Appointment"
	set cancelled_by_merchant_on = $1, cancellation_reason = $2
	where merchant_id = $3 and id = $4 and cancelled_by_user_on is null
	`

	_, err := s.db.Exec(ctx, query, time.Now().UTC(), cancellationReason, merchantId, appointmentId)
	if err != nil {
		return err
	}

	return nil
}

func (s *service) UpdateEmailIdForAppointment(ctx context.Context, appointmentId int, emailId string) error {
	emailUUID, err := uuid.Parse(emailId)
	if err != nil {
		return err
	}

	query := `
	update "Appointment" set email_id = $1 where ID = $2`

	_, err = s.db.Exec(ctx, query, emailUUID, appointmentId)
	if err != nil {
		return err
	}

	return nil
}

type AppointmentEmailData struct {
	FromDate      time.Time `json:"from_date" db:"from_date"`
	ToDate        time.Time `json:"to_date" db:"to_date"`
	ServiceName   string    `json:"service_name" db:"service_name"`
	ShortLocation string    `json:"short_location" db:"short_location"`
	UserEmail     string    `json:"user_email" db:"user_email"`
	EmailId       uuid.UUID `json:"email_id" db:"email_id"`
	MerchantName  string    `json:"merchant_name" db:"merchant_name"`
}

func (s *service) GetAppointmentDataForEmail(ctx context.Context, appointmentId int) (AppointmentEmailData, error) {
	query := `
	select a.from_date, a.to_date, a.email_id, s.name as service_name, u.email as user_email, m.name as merchant_name, 
	l.address || ', ' || l.city || ', ' || l.postal_code || ', ' || l.country as short_location from "Appointment" a
	join "Service" s on s.id = a.service_id
	join "User" u on u.id = a.user_id
	join "Merchant" m on m.id = a.merchant_id
	join "Location" l on l.id = a.location_id
	where a.id = $1`

	var data AppointmentEmailData
	err := s.db.QueryRow(ctx, query, appointmentId).Scan(&data.FromDate, &data.ToDate, &data.EmailId, &data.ServiceName, &data.UserEmail, &data.MerchantName, &data.ShortLocation)
	if err != nil {
		return AppointmentEmailData{}, err
	}

	return data, nil
}
