package database

import (
	"context"
	"fmt"
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

func (s *service) GetAvailableTimes(ctx context.Context, merchant_id uuid.UUID, service_duration, location_id int, day time.Time) ([]string, error) {

	query := `
    select  from_date AT TIME ZONE 'UTC' AS from_date_utc, to_date AT TIME ZONE 'UTC' AS to_date_utc from "Appointment"
    where merchant_id = $1 and location_id = $2 and DATE(from_date) = $3
    ORDER BY from_date`

	rows, err := s.db.QueryContext(ctx, query, merchant_id, location_id, day)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type bookedApps struct {
		from_date time.Time
		to_date   time.Time
	}

	var b_apps []bookedApps

	for rows.Next() {
		var b_app bookedApps
		if err := rows.Scan(&b_app.from_date, &b_app.to_date); err != nil {
			return nil, err
		}
		b_apps = append(b_apps, b_app)
	}

	year, month, day_ := day.Date()
	location := day.Location()

	businessStart := time.Date(year, month, day_, 8, 0, 0, 0, location)
	businessEnd := time.Date(year, month, day_, 17, 0, 0, 0, location)

	duration := time.Duration(service_duration) * time.Minute
	current := businessStart
	var availableTimes []time.Time

	for current.Add(duration).Before(businessEnd) || current.Add(duration).Equal(businessEnd) {
		timeEnd := current.Add(duration)
		available := true

		for _, appt := range b_apps {
			if timeEnd.After(appt.from_date) && current.Before(appt.to_date) {
				current = appt.to_date
				timeEnd = current.Add(duration)
				available = false
				break
			}
		}
		if available && timeEnd.Before(businessEnd) || timeEnd.Equal(businessEnd) {
			availableTimes = append(availableTimes, current)
			current = timeEnd
		}
	}

	var formattedTimes []string

	for _, slot := range availableTimes {
		formattedTimes = append(formattedTimes, fmt.Sprintf("%02d:%02d", slot.Hour(), slot.Minute()))
	}
	return formattedTimes, nil
}
