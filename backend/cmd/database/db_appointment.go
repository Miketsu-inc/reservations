package database

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/miketsu-inc/reservations/backend/pkg/currencyx"
)

type Appointment struct {
	Id                    int             `json:"ID" db:"id"`
	CustomerId            uuid.UUID       `json:"customer_id" db:"customer_id"`
	MerchantId            uuid.UUID       `json:"merchant_id" db:"merchant_id"`
	ServiceId             int             `json:"service_id" db:"service_id"`
	ServicePhaseId        int             `json:"service_phase_id" db:"service_phase_id"`
	LocationId            int             `json:"location_id" db:"location_id"`
	GroupId               int             `json:"group_id" db:"group_id"`
	FromDate              time.Time       `json:"from_date" db:"from_date"`
	ToDate                time.Time       `json:"to_date" db:"to_date"`
	CustomerNote          string          `json:"" db:"customer_note"`
	MerchantNote          string          `json:"merchant_note" db:"merchant_note"`
	PriceThen             currencyx.Price `json:"price_then" db:"price_then"`
	CostThen              currencyx.Price `json:"cost_then" db:"cost_then"`
	CancelledByUserOn     string          `json:"cancelled_by_user_on" db:"cancelled_by_user_on"`
	CancelledByMerchantOn string          `json:"cancelled_by_merchant_on" db:"cancelled_by_merchant_on"`
	CancellationReason    string          `json:"cancellation_reason" db:"cancellation_reason"`
	TransferredTo         uuid.UUID       `json:"transferred_to" db:"transferred_to"`
	EmailId               uuid.UUID       `json:"email_id" db:"email_id"`
}

// every appointment needs a group_id because otherwise
// they would get grouped together as null
func (s *service) NewAppointment(ctx context.Context, app Appointment, phases []PublicServicePhase, UserId uuid.UUID, newCustomerId uuid.UUID) (int, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return 0, err
	}
	// nolint:errcheck
	defer tx.Rollback(ctx)

	var IsBlacklisted bool
	ensureCustomerQuery := `
	insert into "Customer" (id, merchant_id, user_id) values ($1, $2, $3)
	on conflict (merchant_id, user_id) do update
	set merchant_id = excluded.merchant_id
	returning id, is_blacklisted`

	err = tx.QueryRow(ctx, ensureCustomerQuery, newCustomerId, app.MerchantId, UserId).Scan(&app.CustomerId, &IsBlacklisted)
	if err != nil {
		return 0, err
	}
	if IsBlacklisted {
		return 0, fmt.Errorf("you are blacklisted, please contact the merchant by email or phone to book an appointment")
	}

	insertQuery := `
	insert into "Appointment" (customer_id, merchant_id, service_id, service_phase_id, location_id, group_id, from_date, to_date,
		customer_note, merchant_note, price_then, cost_then)
	values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	var id int
	appStart := app.FromDate

	for index, phase := range phases {
		phaseDuration := time.Duration(phase.Duration) * time.Minute
		appEnd := appStart.Add(phaseDuration)

		// get the first appointment's id for the group_id column
		if index == 0 {
			err = tx.QueryRow(ctx, insertQuery+` returning id`, app.CustomerId, app.MerchantId, app.ServiceId, phase.Id, app.LocationId, app.GroupId,
				appStart, appEnd, app.CustomerNote, app.MerchantNote, app.PriceThen, app.CostThen).Scan(&id)
			if err != nil {
				return 0, err
			}

		} else {
			_, err = tx.Exec(ctx, insertQuery, app.CustomerId, app.MerchantId, app.ServiceId, phase.Id, app.LocationId, id,
				appStart, appEnd, app.CustomerNote, app.MerchantNote, app.PriceThen, app.CostThen)
			if err != nil {
				return 0, err
			}
		}

		appStart = appEnd
	}

	updateGroupIdQuery := `
	update "Appointment"
	set group_id = $1
	where id = $2
	`
	_, err = tx.Exec(ctx, updateGroupIdQuery, id, id)
	if err != nil {
		return 0, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (s *service) UpdateAppointmentData(ctx context.Context, merchantId uuid.UUID, appointmentId int, merchant_note string, offset time.Duration) error {
	query := `
	update "Appointment"
	set merchant_note = $1, from_date = from_date + $2, to_date = to_date + $2
	where group_id = $3 and merchant_id = $4 and cancelled_by_user_on is null and cancelled_by_merchant_on is null
	`

	_, err := s.db.Exec(ctx, query, merchant_note, offset, appointmentId, merchantId)
	if err != nil {
		return err
	}

	return nil
}

type AppointmentDetails struct {
	ID              int                      `json:"id" db:"id"`
	GroupId         int                      `json:"group_id" db:"group_id"`
	FromDate        time.Time                `json:"from_date" db:"from_date"`
	ToDate          time.Time                `json:"to_date" db:"to_date"`
	CustomerNote    string                   `json:"customer_note" db:"customer_note"`
	MerchantNote    string                   `json:"merchant_note" db:"merchant_note"`
	ServiceName     string                   `json:"service_name" db:"service_name"`
	ServiceColor    string                   `json:"service_color" db:"service_color"`
	ServiceDuration int                      `json:"service_duration" db:"service_duration"`
	Price           currencyx.FormattedPrice `json:"price" db:"price"`
	Cost            currencyx.FormattedPrice `json:"cost" db:"cost"`
	FirstName       string                   `json:"first_name" db:"first_name"`
	LastName        string                   `json:"last_name" db:"last_name"`
	PhoneNumber     string                   `json:"phone_number" db:"phone_number"`
}

func (s *service) GetAppointmentsByMerchant(ctx context.Context, merchantId uuid.UUID, start string, end string) ([]AppointmentDetails, error) {
	query := `
	select distinct on (a.group_id) a.id, a.group_id,
		min(a.from_date) over (partition by a.group_id) as from_date,
		max(a.to_date) over (partition by a.group_id) as to_date,
		a.customer_note, a.merchant_note, a.price_then as price, a.cost_then as cost,
	s.name as service_name, s.color as service_color, s.total_duration as service_duration,
	coalesce(c.first_name, u.first_name) as first_name, coalesce(c.last_name, u.last_name) as last_name, coalesce(c.phone_number, u.phone_number) as phone_number
	from "Appointment" a
	join "Service" s on a.service_id = s.id
	join "Customer" c on a.customer_id = c.id
	left join "User" u on c.user_id = u.id
	where a.merchant_id = $1 and a.from_date >= $2 AND a.to_date <= $3 AND a.cancelled_by_user_on is null and a.cancelled_by_merchant_on is null
	order by a.group_id, a.id
	`

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
    select a.from_date, a.to_date from "Appointment" a
	inner join "ServicePhase" sp on a.service_phase_id = sp.id
    where a.merchant_id = $1 and a.location_id = $2 and DATE(a.from_date) = $3 and a.cancelled_by_user_on is null
		and a.cancelled_by_merchant_on is null and sp.phase_type = 'active'
    ORDER BY a.from_date`

	rows, _ := s.db.Query(ctx, query, merchant_id, location_id, day)
	bookedApps, err := pgx.CollectRows(rows, pgx.RowToStructByName[AppointmentTime])
	if err != nil {
		return nil, err
	}

	return bookedApps, nil
}

func (s *service) GetReservedTimesForPeriod(ctx context.Context, merchantId uuid.UUID, locationId int, startDate time.Time, endDate time.Time) ([]AppointmentTime, error) {
	query := `
	select a.from_date, a.to_date from "Appointment" a
	inner join "ServicePhase" sp on a.service_phase_id = sp.id
	where a.merchant_id = $1 and a.location_id = $2 and DATE(a.from_date) >= $3 and DATE(to_date) <= $4
		and a.cancelled_by_merchant_on is null and a.cancelled_by_user_on is null and sp.phase_type = 'active'
	order by a.from_date`

	rows, _ := s.db.Query(ctx, query, merchantId, locationId, startDate, endDate)
	bookedApps, err := pgx.CollectRows(rows, pgx.RowToStructByName[AppointmentTime])
	if err != nil {
		return nil, err
	}

	return bookedApps, nil
}

func (s *service) TransferDummyAppointments(ctx context.Context, merchantId uuid.UUID, fromCustomer uuid.UUID, toCustomer uuid.UUID) error {
	query := `
	update "Appointment" a
	set transferred_to = $3
	from "Customer" c
	where a.customer_id = c.id and a.merchant_id = $1 and a.customer_id = $2 and c.user_id is null
	`

	_, err := s.db.Exec(ctx, query, merchantId, fromCustomer, toCustomer)
	if err != nil {
		return err
	}

	return nil
}

func (s *service) CancelAppointmentByMerchant(ctx context.Context, merchantId uuid.UUID, appointmentId int, cancellationReason string) error {
	query := `
	update "Appointment"
	set cancelled_by_merchant_on = $1, cancellation_reason = $2
	where merchant_id = $3 and group_id = $4 and cancelled_by_user_on is null and cancelled_by_merchant_on is null and from_date > $1
	`

	_, err := s.db.Exec(ctx, query, time.Now().UTC(), cancellationReason, merchantId, appointmentId)
	if err != nil {
		return err
	}

	return nil
}

func (s *service) CancelAppointmentByUser(ctx context.Context, customerId uuid.UUID, appointmentId int) (uuid.UUID, error) {
	query := `
	update "Appointment"
	set cancelled_by_user_on = $1
	where customer_id = $2 and group_id = $3 and cancelled_by_merchant_on is null and cancelled_by_user_on is null and from_date > $1
	returning email_id`

	var emailId uuid.UUID
	err := s.db.QueryRow(ctx, query, time.Now().UTC(), customerId, appointmentId).Scan(&emailId)
	if err != nil {
		return uuid.Nil, err
	}

	return emailId, nil
}

func (s *service) UpdateEmailIdForAppointment(ctx context.Context, appointmentId int, emailId string) error {
	emailUUID, err := uuid.Parse(emailId)
	if err != nil {
		return err
	}

	query := `
	update "Appointment" set email_id = $1 where group_id = $2`

	_, err = s.db.Exec(ctx, query, emailUUID, appointmentId)
	if err != nil {
		return err
	}

	return nil
}

type AppointmentEmailData struct {
	FromDate       time.Time `json:"from_date" db:"from_date"`
	ToDate         time.Time `json:"to_date" db:"to_date"`
	ServiceName    string    `json:"service_name" db:"service_name"`
	ShortLocation  string    `json:"short_location" db:"short_location"`
	CustomerEmail  string    `json:"customer_email" db:"customer_email"`
	EmailId        uuid.UUID `json:"email_id" db:"email_id"`
	MerchantName   string    `json:"merchant_name" db:"merchant_name"`
	CancelDeadline int       `json:"cancel_deadline" db:"cancel_deadline"`
}

func (s *service) GetAppointmentDataForEmail(ctx context.Context, appointmentId int) (AppointmentEmailData, error) {
	query := `
	select distinct on (a.group_id)
		min(a.from_date) over (partition by a.group_id) as from_date,
		max(a.to_date) over (partition by a.group_id) as to_date,
		a.email_id, s.name as service_name, coalesce(u.email, c.email) as customer_email, m.name as merchant_name, m.cancel_deadline,
	l.address || ', ' || l.city || ', ' || l.postal_code || ', ' || l.country as short_location from "Appointment" a
	join "Service" s on s.id = a.service_id
	join "Customer" c on c.id = a.customer_id
	left join "User" u on u.id = c.user_id
	join "Merchant" m on m.id = a.merchant_id
	join "Location" l on l.id = a.location_id
	where a.group_id = $1
	`

	var data AppointmentEmailData
	err := s.db.QueryRow(ctx, query, appointmentId).Scan(&data.FromDate, &data.ToDate, &data.EmailId, &data.ServiceName,
		&data.CustomerEmail, &data.MerchantName, &data.CancelDeadline, &data.ShortLocation)
	if err != nil {
		return AppointmentEmailData{}, err
	}

	return data, nil
}

type PublicAppointmentInfo struct {
	FromDate            time.Time                `json:"from_date" db:"from_date"`
	ToDate              time.Time                `json:"to_date" db:"to_date"`
	ServiceName         string                   `json:"service_name" db:"service_name"`
	CancelDeadline      int                      `json:"cancel_deadline" db:"cancel_deadline"`
	ShortLocation       string                   `json:"short_location" db:"short_location"`
	Price               currencyx.FormattedPrice `json:"price" db:"price"`
	PriceNote           *string                  `json:"price_note"`
	MerchantName        string                   `json:"merchant_name" db:"merchant_name"`
	CancelledByUser     bool                     `json:"cancelled_by_user" db:"cancelled_by_user"`
	CancelledByMerchant bool                     `json:"cancelled_by_merchant" db:"cancelled_by_merchant"`
}

func (s *service) GetPublicAppointmentInfo(ctx context.Context, appointmentId int) (PublicAppointmentInfo, error) {
	query := `
	select distinct on (a.group_id)
		min(a.from_date) over (partition by a.group_id) as from_date,
		max(a.to_date) over (partition by a.group_id) as to_date,
		a.price_then as price, m.name as merchant_name, s.name as service_name, m.cancel_deadline, s.price_note,
	a.cancelled_by_user_on is not null as cancelled_by_user,
	a.cancelled_by_merchant_on is not null as cancelled_by_merchant,
	l.address || ', ' || l.city || ' ' || l.postal_code || ', ' || l.country as short_location
	from "Appointment" a
	join "Service" s on s.id = a.service_id
	join "Merchant" m on m.id = a.merchant_id
	join "Location" l on l.id = a.location_id
	where a.group_id = $1
	`

	var data PublicAppointmentInfo
	err := s.db.QueryRow(ctx, query, appointmentId).Scan(&data.FromDate, &data.ToDate, &data.Price, &data.MerchantName,
		&data.ServiceName, &data.CancelDeadline, &data.PriceNote, &data.CancelledByUser, &data.CancelledByMerchant, &data.ShortLocation)
	if err != nil {
		return PublicAppointmentInfo{}, err
	}

	return data, nil
}
