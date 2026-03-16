package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/miketsu-inc/reservations/backend/internal/domain"
	"github.com/miketsu-inc/reservations/backend/pkg/db"
)

type customerRepository struct {
	db db.DBTX
}

func NewCustomerRepository(db db.DBTX) domain.CustomerRepository {
	return &customerRepository{db: db}
}

func (r *customerRepository) WithTx(tx db.DBTX) domain.CustomerRepository {
	return &customerRepository{db: tx}
}

func (r *customerRepository) NewCustomer(ctx context.Context, merchantId uuid.UUID, customer domain.Customer) error {
	query := `
	insert into "Customer" (id, merchant_id, first_name, last_name, email, phone_number, birthday, note)
	values ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.Exec(ctx, query, customer.Id, merchantId, customer.FirstName, customer.LastName, customer.Email, customer.PhoneNumber, customer.Birthday, customer.Note)
	if err != nil {
		return err
	}

	return nil
}

func (r *customerRepository) NewCustomerFromUser(ctx context.Context, customerId, merchantId, userId uuid.UUID) (uuid.UUID, bool, error) {
	query := `
	insert into "Customer" (id, merchant_id, user_id) values ($1, $2, $3)
	on conflict (merchant_id, user_id) do update
	set merchant_id = excluded.merchant_id
	returning id, is_blacklisted`

	var IsBlacklisted bool
	var custId uuid.UUID

	err := r.db.QueryRow(ctx, query, customerId, merchantId, userId).Scan(&custId, &IsBlacklisted)
	if err != nil {
		return uuid.UUID{}, false, err
	}

	return custId, IsBlacklisted, nil
}

// TODO: this logic shouldn't live here and some of it is probably unnecessary
func (r *customerRepository) UpdateCustomer(ctx context.Context, merchantId uuid.UUID, customer domain.Customer) error {
	type field struct {
		name  string
		value interface{}
	}

	fields := []field{
		{"first_name", customer.FirstName},
		{"last_name", customer.LastName},
		{"email", customer.Email},
		{"phone_number", customer.PhoneNumber},
		{"birthday", customer.Birthday},
		{"note", customer.Note},
	}

	setClauses := []string{}
	args := []interface{}{merchantId, customer.Id}
	argPos := 3

	for _, f := range fields {
		if f.value != nil && !reflect.ValueOf(f.value).IsNil() {
			setClauses = append(setClauses, fmt.Sprintf(`%s = $%d`, f.name, argPos))
			args = append(args, f.value)
			argPos++
		}
	}

	if len(setClauses) == 0 {
		return nil
	}

	query := fmt.Sprintf(`
		UPDATE "Customer"
		SET %s
		WHERE merchant_id = $1 AND id = $2
	`, strings.Join(setClauses, ", "))

	_, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return err
	}

	return nil
}

func (r *customerRepository) DeleteCustomer(ctx context.Context, customerId uuid.UUID, merchantId uuid.UUID) error {
	query := `
	delete from "Customer"
	where user_id is null and id = $1 and merchant_id = $2
	`

	_, err := r.db.Exec(ctx, query, customerId, merchantId)
	if err != nil {
		return err
	}

	return nil
}

func (r *customerRepository) GetCustomers(ctx context.Context, merchantId uuid.UUID, isBlacklisted bool) ([]domain.PublicCustomer, error) {
	query := `
	select c.id,
		   coalesce(c.first_name, u.first_name) as first_name, coalesce(c.last_name, u.last_name) as last_name,
		   coalesce(c.email, u.email) as email, coalesce(c.phone_number, u.phone_number) as phone_number, c.birthday, c.note,
		   c.user_id is null as is_dummy, c.is_blacklisted, c.blacklist_reason,
		count(b.id) as times_booked, count(distinct bp.status = 'cancelled') as times_cancelled
	from "Customer" c
	left join "User" u on c.user_id = u.id
	left join "BookingParticipant" bp on c.id = coalesce(bp.transferred_to, bp.customer_id)
	left join "Booking" b on bp.booking_id = b.id and b.merchant_id = $1
	where c.merchant_id = $1 and c.is_blacklisted = $2
	group by c.id, u.first_name, u.last_name, u.email, u.phone_number
	`

	rows, _ := r.db.Query(ctx, query, merchantId, isBlacklisted)
	customers, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.PublicCustomer])
	if err != nil {
		return []domain.PublicCustomer{}, err
	}

	// if customers array is empty the encoded json field will be null
	// unless an empty slice is supplied to it
	if len(customers) == 0 {
		customers = []domain.PublicCustomer{}
	}

	return customers, nil
}

func (r *customerRepository) GetCustomerInfo(ctx context.Context, merchantId uuid.UUID, customerId uuid.UUID) (domain.CustomerInfo, error) {
	query := `
	select c.id, coalesce(c.first_name, u.first_name) as first_name, coalesce(c.last_name, u.last_name) as last_name,
	coalesce(c.email, u.email) as email, coalesce(c.phone_number, u.phone_number) as phone_number, c.birthday, c.note, c.user_id is null as is_dummy
	from "Customer" c
	left join "User" u on u.id = c.user_id
	where c.id = $1 and c.merchant_id = $2`

	var customer domain.CustomerInfo
	err := r.db.QueryRow(ctx, query, customerId, merchantId).Scan(&customer.Id, &customer.FirstName, &customer.LastName,
		&customer.Email, &customer.PhoneNumber, &customer.Birthday, &customer.Note, &customer.IsDummy)
	if err != nil {
		return domain.CustomerInfo{}, err
	}

	return customer, nil
}

func (r *customerRepository) GetCustomerStats(ctx context.Context, merchantId uuid.UUID, customerId uuid.UUID) (domain.CustomerStatistics, error) {
	query := `
	with bookings as (
		select b.customer_id,
			jsonb_agg(
				jsonb_build_object(
					'from_date', b.from_date,
					'to_date', b.to_date,
					'service_name', s.name,
					'price', b.price_per_person,
					'price_type', s.price_type,
					'merchant_name', m.name,
					'formatted_location', l.formatted_location,
					'is_cancelled', b.status in ('cancelled')
				) order by b.from_date desc
			) as bookings
		from (
			select bp.customer_id, b.id, b.from_date, b.to_date, b.merchant_id, b.location_id, b.service_id, bd.price_per_person, bp.status
			from "Booking" b
			join "BookingDetails" bd on bd.booking_id = b.id
			left join "BookingParticipant" bp on bp.booking_id = b.id and (bp.customer_id = $2 or bp.transferred_to = $2)
			where b.merchant_id = $1 and bd.cancelled_by_merchant_on is null
		) b
		join "Service" s on s.id = b.service_id
		join "Merchant" m on m.id = b.merchant_id
		join "Location" l on l.id = b.location_id
		group by b.customer_id
	)
	select c.id, coalesce(c.first_name, u.first_name) as first_name, coalesce(c.last_name, u.last_name) as last_name,
		coalesce(c.email, u.email) as email, coalesce(c.phone_number, u.phone_number) as phone_number,birthday, note, c.user_id is null as is_dummy, c.is_blacklisted, c.blacklist_reason,
		count(b.id) as times_booked, count(distinct case when bp.status in ('cancelled') then b.id end) as times_cancelled_by_user,
		count(distinct case when bp.status not in ('cancelled', 'completed') and b.status not in ('cancelled', 'completed') and b.from_date >= now() then b.id end) as times_upcoming,
		coalesce(ca.bookings, '[]'::jsonb) as bookings
	from "Customer" c
	left join "User" u on u.id = c.user_id
	left join "BookingParticipant" bp on bp.customer_id = c.id
	left join "Booking" b on bp.booking_id = b.id and b.merchant_id = $1
	left join bookings ca on c.id = ca.customer_id
	where c.id = $2 and c.merchant_id = $1
	GROUP BY c.id, u.first_name, u.last_name, u.email, u.phone_number, ca.bookings
	`

	var customer domain.CustomerStatistics
	var bookingsJSON []byte

	err := r.db.QueryRow(ctx, query, merchantId, customerId).Scan(&customer.Id, &customer.FirstName, &customer.LastName, &customer.Email, &customer.PhoneNumber, &customer.Birthday,
		&customer.Note, &customer.IsDummy, &customer.IsBlacklisted, &customer.BlacklistReason, &customer.TimesBooked, &customer.TimesCancelledByUser, &customer.TimesUpcoming, &bookingsJSON)
	if err != nil {
		return domain.CustomerStatistics{}, err
	}

	if len(bookingsJSON) > 0 {
		err = json.Unmarshal(bookingsJSON, &customer.Bookings)
		if err != nil {
			return domain.CustomerStatistics{}, err
		}
	} else {
		customer.Bookings = []domain.PublicBooking{}
	}

	return customer, nil
}

func (r *customerRepository) GetCustomersForCalendar(ctx context.Context, merchantId uuid.UUID) ([]domain.CustomerForCalendar, error) {
	query := `
	select c.id as customer_id, coalesce(c.first_name, u.first_name) as first_name, coalesce(c.last_name, u.last_name) as last_name, coalesce(c.email, u.email) as email,
		coalesce(c.phone_number, u.phone_number) as phone_number, c.birthday, c.user_id is null as is_dummy, max(b.from_date) as last_visited
	from "Customer" c
	left join "User" u on c.user_id = u.id
	left join "BookingParticipant" bp on bp.customer_id = c.id and bp.status = 'completed'
	left join "Booking" b on bp.booking_id = b.id and b.merchant_id = $1 and b.from_date < now()
	where c.merchant_id = $1
	group by c.id, u.first_name, u.last_name, u.email, u.phone_number
	`

	rows, _ := r.db.Query(ctx, query, merchantId)
	customers, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.CustomerForCalendar])
	if err != nil {
		return []domain.CustomerForCalendar{}, err
	}

	return customers, nil
}

func (r *customerRepository) SetBlacklistStatusForCustomer(ctx context.Context, merchantId uuid.UUID, customerId uuid.UUID, isBlacklisted bool, blacklistReason *string) error {
	query := `
	update "Customer" set is_blacklisted = $3, blacklist_reason = $4
	where merchant_id = $1 and id = $2`

	_, err := r.db.Exec(ctx, query, merchantId, customerId, isBlacklisted, blacklistReason)
	if err != nil {
		return err
	}

	return nil

}

func (r *customerRepository) GetCustomerIdByUserIdAndMerchantId(ctx context.Context, merchantId uuid.UUID, userId uuid.UUID) (uuid.UUID, error) {
	query := `
	select id from "Customer" where user_id = $1 and merchant_id = $2`

	var customerId uuid.UUID
	err := r.db.QueryRow(ctx, query, userId, merchantId).Scan(&customerId)
	if err != nil {
		return uuid.Nil, err
	}

	return customerId, nil
}

func (r *customerRepository) GetCustomerEmailById(ctx context.Context, merchantId uuid.UUID, customerId uuid.UUID) (*string, error) {
	query := `
	select coalesce(c.email, u.email)
	from "Customer" c
	join "User" u on u.id = c.user_id
	where c.id = $1 and c.merchant_id = $2
	`

	var email *string
	err := r.db.QueryRow(ctx, query, customerId, merchantId).Scan(&email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return email, nil
		}
		return nil, err
	}

	return email, nil
}
