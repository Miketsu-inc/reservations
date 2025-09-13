package database

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/miketsu-inc/reservations/backend/pkg/currencyx"
)

type Booking struct {
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

// every booking needs a group_id because otherwise
// they would get grouped together as null
func (s *service) NewBooking(ctx context.Context, book Booking, phases []PublicServicePhase, UserId uuid.UUID, newCustomerId uuid.UUID) (int, error) {
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

	err = tx.QueryRow(ctx, ensureCustomerQuery, newCustomerId, book.MerchantId, UserId).Scan(&book.CustomerId, &IsBlacklisted)
	if err != nil {
		return 0, err
	}
	if IsBlacklisted {
		return 0, fmt.Errorf("you are blacklisted, please contact the merchant by email or phone to make a booking")
	}

	insertQuery := `
	insert into "Booking" (customer_id, merchant_id, service_id, service_phase_id, location_id, group_id, from_date, to_date,
		customer_note, merchant_note, price_then, cost_then)
	values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	var id int
	bookingStart := book.FromDate

	for index, phase := range phases {
		phaseDuration := time.Duration(phase.Duration) * time.Minute
		bookingEnd := bookingStart.Add(phaseDuration)

		// get the first booking's id for the group_id column
		if index == 0 {
			err = tx.QueryRow(ctx, insertQuery+` returning id`, book.CustomerId, book.MerchantId, book.ServiceId, phase.Id, book.LocationId, book.GroupId,
				bookingStart, bookingEnd, book.CustomerNote, book.MerchantNote, book.PriceThen, book.CostThen).Scan(&id)
			if err != nil {
				return 0, err
			}

		} else {
			_, err = tx.Exec(ctx, insertQuery, book.CustomerId, book.MerchantId, book.ServiceId, phase.Id, book.LocationId, id,
				bookingStart, bookingEnd, book.CustomerNote, book.MerchantNote, book.PriceThen, book.CostThen)
			if err != nil {
				return 0, err
			}
		}

		bookingStart = bookingEnd
	}

	updateGroupIdQuery := `
	update "Booking"
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

func (s *service) UpdateBookingData(ctx context.Context, merchantId uuid.UUID, bookingId int, merchant_note string, offset time.Duration) error {
	query := `
	update "Booking"
	set merchant_note = $1, from_date = from_date + $2, to_date = to_date + $2
	where group_id = $3 and merchant_id = $4 and cancelled_by_user_on is null and cancelled_by_merchant_on is null
	`

	_, err := s.db.Exec(ctx, query, merchant_note, offset, bookingId, merchantId)
	if err != nil {
		return err
	}

	return nil
}

type BookingDetails struct {
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

func (s *service) GetBookingsByMerchant(ctx context.Context, merchantId uuid.UUID, start string, end string) ([]BookingDetails, error) {
	query := `
	select distinct on (b.group_id) b.id, b.group_id,
		min(b.from_date) over (partition by b.group_id) as from_date,
		max(b.to_date) over (partition by b.group_id) as to_date,
		b.customer_note, b.merchant_note, b.price_then as price, b.cost_then as cost,
	s.name as service_name, s.color as service_color, s.total_duration as service_duration,
	coalesce(c.first_name, u.first_name) as first_name, coalesce(c.last_name, u.last_name) as last_name, coalesce(c.phone_number, u.phone_number) as phone_number
	from "Booking" b
	join "Service" s on b.service_id = s.id
	join "Customer" c on b.customer_id = c.id
	left join "User" u on c.user_id = u.id
	where b.merchant_id = $1 and b.from_date >= $2 AND b.to_date <= $3 AND b.cancelled_by_user_on is null and b.cancelled_by_merchant_on is null
	order by b.group_id, b.id
	`

	rows, _ := s.db.Query(ctx, query, merchantId, start, end)
	bookings, err := pgx.CollectRows(rows, pgx.RowToStructByName[BookingDetails])
	if err != nil {
		return nil, err
	}

	return bookings, nil
}

type BookingTime struct {
	From_date time.Time
	To_date   time.Time
}

func (s *service) GetReservedTimes(ctx context.Context, merchant_id uuid.UUID, location_id int, day time.Time) ([]BookingTime, error) {
	query := `
    select b.from_date, b.to_date from "Booking" b
	inner join "ServicePhase" sp on b.service_phase_id = sp.id
    where b.merchant_id = $1 and b.location_id = $2 and DATE(b.from_date) = $3 and b.cancelled_by_user_on is null
		and b.cancelled_by_merchant_on is null and sp.phase_type = 'active'
    ORDER BY b.from_date`

	rows, _ := s.db.Query(ctx, query, merchant_id, location_id, day)
	reservedTimes, err := pgx.CollectRows(rows, pgx.RowToStructByName[BookingTime])
	if err != nil {
		return nil, err
	}

	return reservedTimes, nil
}

func (s *service) GetReservedTimesForPeriod(ctx context.Context, merchantId uuid.UUID, locationId int, startDate time.Time, endDate time.Time) ([]BookingTime, error) {
	query := `
	select b.from_date, b.to_date from "Booking" b
	inner join "ServicePhase" sp on b.service_phase_id = sp.id
	where b.merchant_id = $1 and b.location_id = $2 and DATE(b.from_date) >= $3 and DATE(b.to_date) <= $4
		and b.cancelled_by_merchant_on is null and b.cancelled_by_user_on is null and sp.phase_type = 'active'
	order by b.from_date`

	rows, _ := s.db.Query(ctx, query, merchantId, locationId, startDate, endDate)
	reservedTimes, err := pgx.CollectRows(rows, pgx.RowToStructByName[BookingTime])
	if err != nil {
		return nil, err
	}

	return reservedTimes, nil
}

func (s *service) TransferDummyBookings(ctx context.Context, merchantId uuid.UUID, fromCustomer uuid.UUID, toCustomer uuid.UUID) error {
	query := `
	update "Booking" b
	set transferred_to = $3
	from "Customer" c
	where b.customer_id = c.id and b.merchant_id = $1 and b.customer_id = $2 and c.user_id is null
	`

	_, err := s.db.Exec(ctx, query, merchantId, fromCustomer, toCustomer)
	if err != nil {
		return err
	}

	return nil
}

func (s *service) CancelBookingByMerchant(ctx context.Context, merchantId uuid.UUID, bookingId int, cancellationReason string) error {
	query := `
	update "Booking"
	set cancelled_by_merchant_on = $1, cancellation_reason = $2
	where merchant_id = $3 and group_id = $4 and cancelled_by_user_on is null and cancelled_by_merchant_on is null and from_date > $1
	`

	_, err := s.db.Exec(ctx, query, time.Now().UTC(), cancellationReason, merchantId, bookingId)
	if err != nil {
		return err
	}

	return nil
}

func (s *service) CancelBookingByUser(ctx context.Context, customerId uuid.UUID, bookingId int) (uuid.UUID, error) {
	query := `
	update "Booking"
	set cancelled_by_user_on = $1
	where customer_id = $2 and group_id = $3 and cancelled_by_merchant_on is null and cancelled_by_user_on is null and from_date > $1
	returning email_id`

	var emailId uuid.UUID
	err := s.db.QueryRow(ctx, query, time.Now().UTC(), customerId, bookingId).Scan(&emailId)
	if err != nil {
		return uuid.Nil, err
	}

	return emailId, nil
}

func (s *service) UpdateEmailIdForBooking(ctx context.Context, bookingId int, emailId string) error {
	emailUUID, err := uuid.Parse(emailId)
	if err != nil {
		return err
	}

	query := `
	update "Booking" set email_id = $1 where group_id = $2`

	_, err = s.db.Exec(ctx, query, emailUUID, bookingId)
	if err != nil {
		return err
	}

	return nil
}

type BookingEmailData struct {
	FromDate       time.Time `json:"from_date" db:"from_date"`
	ToDate         time.Time `json:"to_date" db:"to_date"`
	ServiceName    string    `json:"service_name" db:"service_name"`
	ShortLocation  string    `json:"short_location" db:"short_location"`
	CustomerEmail  string    `json:"customer_email" db:"customer_email"`
	EmailId        uuid.UUID `json:"email_id" db:"email_id"`
	MerchantName   string    `json:"merchant_name" db:"merchant_name"`
	CancelDeadline int       `json:"cancel_deadline" db:"cancel_deadline"`
}

func (s *service) GetBookingDataForEmail(ctx context.Context, bookingId int) (BookingEmailData, error) {
	query := `
	select distinct on (b.group_id)
		min(b.from_date) over (partition by b.group_id) as from_date,
		max(b.to_date) over (partition by b.group_id) as to_date,
		b.email_id, s.name as service_name, coalesce(u.email, c.email) as customer_email, m.name as merchant_name, m.cancel_deadline,
		l.address || ', ' || l.city || ', ' || l.postal_code || ', ' || l.country as short_location
	from "Booking" b
	join "Service" s on s.id = b.service_id
	join "Customer" c on c.id = b.customer_id
	left join "User" u on u.id = c.user_id
	join "Merchant" m on m.id = b.merchant_id
	join "Location" l on l.id = b.location_id
	where b.group_id = $1
	`

	var data BookingEmailData
	err := s.db.QueryRow(ctx, query, bookingId).Scan(&data.FromDate, &data.ToDate, &data.EmailId, &data.ServiceName,
		&data.CustomerEmail, &data.MerchantName, &data.CancelDeadline, &data.ShortLocation)
	if err != nil {
		return BookingEmailData{}, err
	}

	return data, nil
}

type PublicBookingInfo struct {
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

func (s *service) GetPublicBookingInfo(ctx context.Context, bookingId int) (PublicBookingInfo, error) {
	query := `
	select distinct on (b.group_id)
		min(b.from_date) over (partition by b.group_id) as from_date,
		max(b.to_date) over (partition by b.group_id) as to_date,
		b.price_then as price, m.name as merchant_name, s.name as service_name, m.cancel_deadline, s.price_note,
	b.cancelled_by_user_on is not null as cancelled_by_user,
	b.cancelled_by_merchant_on is not null as cancelled_by_merchant,
	l.address || ', ' || l.city || ' ' || l.postal_code || ', ' || l.country as short_location
	from "Booking" b
	join "Service" s on s.id = b.service_id
	join "Merchant" m on m.id = b.merchant_id
	join "Location" l on l.id = b.location_id
	where b.group_id = $1
	`

	var data PublicBookingInfo
	err := s.db.QueryRow(ctx, query, bookingId).Scan(&data.FromDate, &data.ToDate, &data.Price, &data.MerchantName,
		&data.ServiceName, &data.CancelDeadline, &data.PriceNote, &data.CancelledByUser, &data.CancelledByMerchant, &data.ShortLocation)
	if err != nil {
		return PublicBookingInfo{}, err
	}

	return data, nil
}
