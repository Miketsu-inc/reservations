package database

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/miketsu-inc/reservations/backend/cmd/types"
	"golang.org/x/text/language"
)

type User struct {
	Id                uuid.UUID               `json:"ID" db:"id"`
	FirstName         string                  `json:"first_name" db:"first_name"`
	LastName          string                  `json:"last_name" db:"last_name"`
	Email             string                  `json:"email" db:"email"`
	PhoneNumber       *string                 `json:"phone_number" db:"phone_number"`
	PasswordHash      *string                 `json:"password_hash" db:"password_hash"`
	JwtRefreshVersion int                     `json:"jwt_refresh_version" db:"jwt_refresh_version"`
	PreferredLang     *string                 `json:"preferred_lang" db:"preferred_lang"`
	AuthProvider      *types.AuthProviderType `json:"auth_provider" db:"auth_provider"`
	ProviderId        *string                 `json:"provider_id" db:"provider_id"`
}

func (s *service) NewUser(ctx context.Context, user User) error {
	query := `
	insert into "User" (id, first_name, last_name, email, phone_number, password_hash, jwt_refresh_version, preferred_lang,
		auth_provider, provider_id)
	values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := s.db.Exec(ctx, query, user.Id, user.FirstName, user.LastName, user.Email, user.PhoneNumber, user.PasswordHash,
		user.JwtRefreshVersion, user.PreferredLang, user.AuthProvider, user.ProviderId)
	if err != nil {
		return err
	}

	return nil
}

func (s *service) GetUserById(ctx context.Context, user_id uuid.UUID) (User, error) {
	query := `
	select * from "User"
	where id = $1
	`

	var user User
	err := s.db.QueryRow(ctx, query, user_id).Scan(&user.Id, &user.FirstName, &user.LastName, &user.Email, &user.PhoneNumber, &user.PasswordHash,
		&user.JwtRefreshVersion, &user.PreferredLang, &user.PreferredLang, &user.AuthProvider, &user.ProviderId)
	if err != nil {
		return User{}, err
	}

	return user, nil
}

func (s *service) GetUserPasswordAndIDByUserEmail(ctx context.Context, email string) (uuid.UUID, *string, error) {
	query := `
	select id, password_hash from "User"
	where email = $1
	`

	var userID uuid.UUID
	var password *string
	err := s.db.QueryRow(ctx, query, email).Scan(&userID, &password)
	if err != nil {
		return uuid.Nil, nil, err
	}

	return userID, password, nil
}

func (s *service) IsEmailUnique(ctx context.Context, email string) error {
	query := `
	select 1 from "User"
	where email = $1
	`

	var em *string
	err := s.db.QueryRow(ctx, query, email).Scan(&em)
	if !errors.Is(err, pgx.ErrNoRows) {
		if err != nil {
			return err
		}

		return fmt.Errorf("this email is already used: %s", email)
	}

	return nil
}

func (s *service) IsPhoneNumberUnique(ctx context.Context, phoneNumber string) error {
	query := `
	select 1 from "User"
	where phone_number = $1
	`

	var pn *string
	err := s.db.QueryRow(ctx, query, phoneNumber).Scan(&pn)
	if !errors.Is(err, pgx.ErrNoRows) {
		if err != nil {
			return err
		}

		return fmt.Errorf("this phone number is already used: %s", phoneNumber)
	}

	return nil
}

func (s *service) IncrementUserJwtRefreshVersion(ctx context.Context, userID uuid.UUID) error {
	query := `
	update "User"
	set jwt_refresh_version = jwt_refresh_version + 1
	where id = $1
	`

	_, err := s.db.Exec(ctx, query, userID)
	if err != nil {
		return err
	}

	return nil
}

func (s *service) GetUserJwtRefreshVersion(ctx context.Context, userID uuid.UUID) (int, error) {
	query := `
	select jwt_refresh_version from "User"
	where id = $1
	`

	var refreshVersion int
	err := s.db.QueryRow(ctx, query, userID).Scan(&refreshVersion)
	if err != nil {
		return 0, err
	}

	return refreshVersion, nil
}

type Customer struct {
	Id          uuid.UUID  `json:"id" db:"id"`
	FirstName   *string    `json:"first_name" db:"first_name"`
	LastName    *string    `json:"last_name" db:"last_name"`
	Email       *string    `json:"email" db:"email"`
	PhoneNumber *string    `json:"phone_number" db:"phone_number"`
	Birthday    *time.Time `json:"birthday" db:"birthday"`
	Note        *string    `json:"note" db:"note"`
}

func (s *service) NewCustomer(ctx context.Context, merchantId uuid.UUID, customer Customer) error {
	query := `
	insert into "Customer" (id, merchant_id, first_name, last_name, email, phone_number, birthday, note)
	values ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := s.db.Exec(ctx, query, customer.Id, merchantId, customer.FirstName, customer.LastName, customer.Email, customer.PhoneNumber, customer.Birthday, customer.Note)
	if err != nil {
		return err
	}

	return nil
}

// TODO: we should ask if they want to delete their booking history as well or not
// also letting them delete customers who are user's by just deleting their bookings
// we should also decide what to do with deleted class/event participants
func (s *service) DeleteCustomerById(ctx context.Context, customerId uuid.UUID, merchantId uuid.UUID) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	// nolint:errcheck
	defer tx.Rollback(ctx)

	deleteBookingsQuery := `
	delete from "Booking" b
	using "BookingParticipant" bp
	where bp.booking_id = b.id and bp.customer_id = $1 and b.merchant_id = $2 and b.booking_type = 'appointment'
	`

	_, err = tx.Exec(ctx, deleteBookingsQuery, customerId, merchantId)
	if err != nil {
		return err
	}

	updateParticipantCountQuery := `
	update "BookingDetails" bd
	set current_participants = current_participants - 1
	from "Booking" b
	left join "BookingParticipant" bp on b.id = bp.booking_id and bp.customer_id = $1
	where b.id = bd.booking_id and b.merchant_id = $2 and b.booking_type in ('event', 'class')
	`

	_, err = tx.Exec(ctx, updateParticipantCountQuery, customerId, merchantId)
	if err != nil {
		return err
	}

	deleteBookingParticipantQuery := `
	delete from "BookingParticipant" bp
	using "Booking" b
	where bp.booking_id = b.id and bp.customer_id = $1 and b.merchant_id = $2
	`

	_, err = tx.Exec(ctx, deleteBookingParticipantQuery, customerId, merchantId)
	if err != nil {
		return err
	}

	deleteCustomerQuery := `
	delete from "Customer"
	where user_id is null and id = $1 and merchant_id = $2
	`

	_, err = tx.Exec(ctx, deleteCustomerQuery, customerId, merchantId)
	if err != nil {
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (s *service) UpdateCustomerById(ctx context.Context, merchantId uuid.UUID, customer Customer) error {
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

	_, err := s.db.Exec(ctx, query, args...)
	if err != nil {
		return err
	}

	return nil
}

func (s *service) SetBlacklistStatusForCustomer(ctx context.Context, merchantId uuid.UUID, customerId uuid.UUID, isBlacklisted bool, blacklistReason *string) error {
	query := `
	update "Customer" set is_blacklisted = $3, blacklist_reason = $4
	where merchant_id = $1 and id = $2`

	_, err := s.db.Exec(ctx, query, merchantId, customerId, isBlacklisted, blacklistReason)
	if err != nil {
		return err
	}

	return nil

}

func (s *service) GetUserPreferredLanguage(ctx context.Context, userId uuid.UUID) (*language.Tag, error) {
	query := `
	select preferred_lang from "User"
	where id = $1
	`

	var pl *string
	err := s.db.QueryRow(ctx, query, userId).Scan(&pl)
	if err != nil {
		return nil, err
	}

	if pl == nil {
		return nil, err
	}

	tag, err := language.Parse(*pl)
	if err != nil {
		return nil, err
	}

	return &tag, nil
}

type CustomerStatistics struct {
	Customer
	IsDummy              bool                `json:"is_dummy"`
	IsBlacklisted        bool                `json:"is_blacklisted"`
	BlacklistReason      *string             `json:"blacklist_reason"`
	TimesBooked          int                 `json:"times_booked"`
	TimesCancelledByUser int                 `json:"times_cancelled_by_user"`
	TimesUpcoming        int                 `json:"times_upcoming"`
	Bookings             []PublicBookingInfo `json:"bookings"`
}

func (s *service) GetCustomerStatsByMerchant(ctx context.Context, merchantId uuid.UUID, customerId uuid.UUID) (CustomerStatistics, error) {
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
			join "BookingParticipant" bp on bp.booking_id = b.id
			join "BookingDetails" bd on bd.booking_id = b.id
			where b.merchant_id = $1 and (bp.customer_id = $2 or bp.transferred_to = $2) and bd.cancelled_by_merchant_on is null
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

	var customer CustomerStatistics
	var bookingsJSON []byte

	err := s.db.QueryRow(ctx, query, merchantId, customerId).Scan(&customer.Id, &customer.FirstName, &customer.LastName, &customer.Email, &customer.PhoneNumber, &customer.Birthday,
		&customer.Note, &customer.IsDummy, &customer.IsBlacklisted, &customer.BlacklistReason, &customer.TimesBooked, &customer.TimesCancelledByUser, &customer.TimesUpcoming, &bookingsJSON)
	if err != nil {
		return CustomerStatistics{}, err
	}

	if len(bookingsJSON) > 0 {
		err = json.Unmarshal(bookingsJSON, &customer.Bookings)
		if err != nil {
			return CustomerStatistics{}, err
		}
	} else {
		customer.Bookings = []PublicBookingInfo{}
	}

	return customer, nil
}

func (s *service) GetCustomerIdByUserIdAndMerchantId(ctx context.Context, merchantId uuid.UUID, userId uuid.UUID) (uuid.UUID, error) {
	query := `
	select id from "Customer" where user_id = $1 and merchant_id = $2`

	var customerId uuid.UUID
	err := s.db.QueryRow(ctx, query, userId, merchantId).Scan(&customerId)
	if err != nil {
		return uuid.Nil, err
	}

	return customerId, nil
}

type CustomerInfo struct {
	Customer
	IsDummy bool `json:"is_dummy"`
}

func (s *service) GetCustomerInfoByMerchant(ctx context.Context, merchantId uuid.UUID, customerId uuid.UUID) (CustomerInfo, error) {
	query := `
	select c.id, coalesce(c.first_name, u.first_name) as first_name, coalesce(c.last_name, u.last_name) as last_name,
	coalesce(c.email, u.email) as email, coalesce(c.phone_number, u.phone_number) as phone_number, c.birthday, c.note, c.user_id is null as is_dummy
	from "Customer" c
	left join "User" u on u.id = c.user_id
	where c.id = $1 and c.merchant_id = $2`

	var customer CustomerInfo
	err := s.db.QueryRow(ctx, query, customerId, merchantId).Scan(&customer.Id, &customer.FirstName, &customer.LastName,
		&customer.Email, &customer.PhoneNumber, &customer.Birthday, &customer.Note, &customer.IsDummy)
	if err != nil {
		return CustomerInfo{}, err
	}

	return customer, nil
}

type EmployeeAuthInfo struct {
	Id         int                `db:"id"`
	LocationId int                `db:"location_id"`
	MerchantId uuid.UUID          `db:"merchant_id"`
	Role       types.EmployeeRole `db:"role"`
}

func (s *service) GetEmployeesByUser(ctx context.Context, userId uuid.UUID) ([]EmployeeAuthInfo, error) {
	query := `
	select e.id, l.id as location_id, e.merchant_id, e.role
	from "Employee" e
	join "Location" l on l.merchant_id = e.merchant_id
	where user_id = $1
	`

	rows, _ := s.db.Query(ctx, query, userId)
	employeeAuthInfo, err := pgx.CollectRows(rows, pgx.RowToStructByName[EmployeeAuthInfo])
	if err != nil {
		return []EmployeeAuthInfo{}, err
	}

	return employeeAuthInfo, nil
}

func (s *service) GetCustomerEmailById(ctx context.Context, merchantId uuid.UUID, customerId uuid.UUID) (string, error) {
	query := `
	select coalesce(c.email, u.email)
	from "Customer" c
	join "User" u on u.id = c.user_id
	where c.id = $1 and c.merchant_id = $2
	`

	var email string
	err := s.db.QueryRow(ctx, query, customerId, merchantId).Scan(&email)
	if err != nil {
		return "", err
	}

	return email, nil
}

type CustomerForCalendar struct {
	Id        uuid.UUID `json:"id" db:"id"`
	FirstName string    `json:"first_name" db:"first_name"`
	LastName  string    `json:"last_name" db:"last_name"`
}

func (s *service) GetCustomersForCalendarByMerchant(ctx context.Context, merchantId uuid.UUID) ([]CustomerForCalendar, error) {
	query := `
	select c.id, coalesce(c.first_name, u.first_name) as first_name, coalesce(c.last_name, u.last_name) as last_name
	from "Customer" c
	join "User" u on c.user_id = u.id
	where c.merchant_id = $1
	`

	rows, _ := s.db.Query(ctx, query, merchantId)
	customers, err := pgx.CollectRows(rows, pgx.RowToStructByName[CustomerForCalendar])
	if err != nil {
		return []CustomerForCalendar{}, err
	}

	return customers, nil
}

func (s *service) FindOauthUser(ctx context.Context, provider types.AuthProviderType, provider_id string) (uuid.UUID, error) {
	query := `
	select id from "User"
	where auth_provider = $1 and provider_id = $2
	`

	var id uuid.UUID
	err := s.db.QueryRow(ctx, query, provider, provider_id).Scan(&id)
	if err != nil {
		return uuid.UUID{}, err
	}

	return id, nil
}
