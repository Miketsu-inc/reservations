package db

import (
	"context"
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
		return fmt.Errorf("NewCustomer: %w", err)
	}

	return nil
}

func (r *customerRepository) NewCustomerFromUser(ctx context.Context, customerId, merchantId, userId uuid.UUID) (uuid.UUID, bool, bool, error) {
	query := `
	insert into "Customer" (id, merchant_id, user_id) values ($1, $2, $3)
	on conflict (merchant_id, user_id) do update
	set merchant_id = excluded.merchant_id
	returning id, is_blacklisted, (xmax = 0) as is_new`

	var IsBlacklisted bool
	var custId uuid.UUID
	var isNew bool

	err := r.db.QueryRow(ctx, query, customerId, merchantId, userId).Scan(&custId, &IsBlacklisted, &isNew)
	if err != nil {
		return uuid.UUID{}, false, false, fmt.Errorf("NewCustomerFromUser: %w", err)
	}

	return custId, IsBlacklisted, isNew, nil
}

// TODO: do we want patch or put here?
func (r *customerRepository) UpdateCustomer(ctx context.Context, merchantId uuid.UUID, customer domain.Customer) error {
	type field struct {
		name  string
		value any
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
	args := []any{merchantId, customer.Id}
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
		return fmt.Errorf("UpdateCustomer: %w", err)
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
		return fmt.Errorf("DeleteCustomer: %w", err)
	}

	return nil
}

func (r *customerRepository) GetCustomers(ctx context.Context, merchantId uuid.UUID, isBlacklisted bool) ([]domain.PublicCustomer, error) {
	query := `
	select c.id,
		   coalesce(c.first_name, u.first_name) as first_name, coalesce(c.last_name, u.last_name) as last_name,
		   coalesce(c.email, u.email) as email, coalesce(c.phone_number, u.phone_number) as phone_number, c.birthday, c.note,
		   c.user_id is null as is_dummy, c.is_blacklisted, c.blacklist_reason,
		count(b.id) as times_booked, count(b.id) filter (where bp.status = 'cancelled') as times_cancelled
	from "Customer" c
	left join "User" u on c.user_id = u.id
	left join "BookingParticipant" bp on c.id = bp.customer_id
	left join "Booking" b on bp.booking_id = b.id and b.merchant_id = $1
		and b.cancelled_by_merchant_on is null
	where c.merchant_id = $1 and c.is_blacklisted = $2
	group by c.id, u.first_name, u.last_name, u.email, u.phone_number
	`

	rows, _ := r.db.Query(ctx, query, merchantId, isBlacklisted)
	customers, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.PublicCustomer])
	if err != nil {
		return []domain.PublicCustomer{}, fmt.Errorf("GetCustomers: %w", err)
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
	coalesce(c.email, u.email) as email, coalesce(c.phone_number, u.phone_number) as phone_number, c.birthday, c.note,
	c.user_id is null as is_dummy, c.is_blacklisted, c.blacklist_reason
	from "Customer" c
	left join "User" u on u.id = c.user_id
	where c.id = $1 and c.merchant_id = $2`

	var customer domain.CustomerInfo
	err := r.db.QueryRow(ctx, query, customerId, merchantId).Scan(&customer.Id, &customer.FirstName, &customer.LastName,
		&customer.Email, &customer.PhoneNumber, &customer.Birthday, &customer.Note, &customer.IsDummy,
		&customer.IsBlacklisted, &customer.BlacklistReason)
	if err != nil {
		return domain.CustomerInfo{}, fmt.Errorf("GetCustomerInfo: %w", err)
	}

	return customer, nil
}

func (r *customerRepository) GetCustomerStats(ctx context.Context, merchantId uuid.UUID, customerId uuid.UUID) (domain.CustomerStatistics, error) {
	query := `
	with customer_bookings as (
		select b.id, b.from_date, b.to_date, b.service_name, b.price_per_person,
			case
				when b.booking_type = 'appointment' then b.status
				else bp.status
			end as status
		from "BookingParticipant" bp
		join "Booking" b on b.id = bp.booking_id and b.merchant_id = $1
		where bp.customer_id = $2 and b.cancelled_by_merchant_on is null
	)
	select count(cb.id) filter (where cb.status = 'booked') as times_booked,
		count(cb.id) filter (where cb.status = 'confirmed') as times_confirmed,
		count(cb.id) filter (where cb.status = 'completed') as times_completed,
		count(cb.id) filter (where cb.status = 'cancelled') as times_cancelled,
		count(cb.id) filter (where cb.status = 'no-show') as times_no_show,
		min(cb.from_date) as first_booking,
		case
			when count(cb.id) filter (where cb.status = 'completed') = 0 then
				array[row(0, m.currency_code)::price]
			else array(
				select row(
					sum((completed.price_per_person).number),
					(completed.price_per_person).currency
				)::price
				from customer_bookings completed
				where completed.status = 'completed'
				group by (completed.price_per_person).currency
				order by (completed.price_per_person).currency
			)
		end as completed_values,
		(
			select service_name
			from customer_bookings
			where status in ('booked', 'confirmed', 'completed')
			group by service_name
			order by count(*) desc, max(from_date) desc
			limit 1
		) as favorite_service,
		min(cb.from_date) filter (
			where cb.status in ('booked', 'confirmed') and cb.from_date >= now()
		) as next_booking
	from "Customer" c
	join "Merchant" m on m.id = c.merchant_id
	left join customer_bookings cb on true
	where c.id = $2 and c.merchant_id = $1
	group by c.id, m.currency_code
	`

	var customer domain.CustomerStatistics
	err := r.db.QueryRow(ctx, query, merchantId, customerId).Scan(
		&customer.TimesBooked, &customer.TimesConfirmed,
		&customer.TimesCompleted, &customer.TimesCancelled, &customer.TimesNoShow,
		&customer.FirstBooking, &customer.CompletedValues,
		&customer.FavoriteService, &customer.NextBooking,
	)
	if err != nil {
		return domain.CustomerStatistics{}, fmt.Errorf("GetCustomerStats: %w", err)
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
		return []domain.CustomerForCalendar{}, fmt.Errorf("GetCustomersForCalendar: %w", err)
	}

	return customers, nil
}

func (r *customerRepository) SetBlacklistStatusForCustomer(ctx context.Context, merchantId uuid.UUID, customerId uuid.UUID, isBlacklisted bool, blacklistReason *string) error {
	query := `
	update "Customer" set is_blacklisted = $3, blacklist_reason = $4
	where merchant_id = $1 and id = $2`

	_, err := r.db.Exec(ctx, query, merchantId, customerId, isBlacklisted, blacklistReason)
	if err != nil {
		return fmt.Errorf("SetBlacklistStatusForCustomer: %w", err)
	}

	return nil

}

func (r *customerRepository) GetCustomerEmailById(ctx context.Context, merchantId uuid.UUID, customerId uuid.UUID) (*string, error) {
	query := `
	select coalesce(c.email, u.email)
	from "Customer" c
	left join "User" u on u.id = c.user_id
	where c.id = $1 and c.merchant_id = $2
	`

	var email *string
	err := r.db.QueryRow(ctx, query, customerId, merchantId).Scan(&email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return email, nil
		}
		return nil, fmt.Errorf("GetCustomerEmailById: %w", err)
	}

	return email, nil
}
