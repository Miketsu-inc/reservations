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
	"golang.org/x/text/language"
)

type User struct {
	Id                uuid.UUID `json:"ID" db:"id"`
	FirstName         string    `json:"first_name" db:"first_name"`
	LastName          string    `json:"last_name" db:"last_name"`
	Email             string    `json:"email" db:"email"`
	PhoneNumber       string    `json:"phone_number" db:"phone_number"`
	PasswordHash      string    `json:"password_hash" db:"password_hash"`
	JwtRefreshVersion int       `json:"jwt_refresh_version" db:"jwt_refresh_version"`
	Subscription      int       `json:"subscription" db:"subscription"`
	PreferredLang     *string   `json:"preferred_lang" db:"preferred_lang"`
}

func (s *service) NewUser(ctx context.Context, user User) error {
	query := `
	insert into "User" (id, first_name, last_name, email, phone_number, password_hash, jwt_refresh_version, subscription)
	values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := s.db.Exec(ctx, query, user.Id, user.FirstName, user.LastName, user.Email, user.PhoneNumber, user.PasswordHash,
		user.JwtRefreshVersion, user.Subscription)
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
		&user.JwtRefreshVersion, &user.Subscription, &user.PreferredLang)
	if err != nil {
		return User{}, err
	}

	return user, nil
}

func (s *service) GetUserPasswordAndIDByUserEmail(ctx context.Context, email string) (uuid.UUID, string, error) {
	query := `
	select id, password_hash from "User"
	where email = $1
	`

	var userID uuid.UUID
	var password string
	err := s.db.QueryRow(ctx, query, email).Scan(&userID, &password)
	if err != nil {
		return uuid.Nil, "", err
	}

	return userID, password, nil
}

func (s *service) IsEmailUnique(ctx context.Context, email string) error {
	query := `
	select 1 from "User"
	where email = $1
	`

	var em string
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

	var pn string
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

func (s *service) DeleteCustomerById(ctx context.Context, customerId uuid.UUID, merchantId uuid.UUID) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	// nolint:errcheck
	defer tx.Rollback(ctx)

	deleteAppointmentsQuery := `
	delete from "Appointment" where customer_id = $1 and merchant_id = $2`

	_, err = tx.Exec(ctx, deleteAppointmentsQuery, customerId, merchantId)
	if err != nil {
		return err
	}

	deleteCcustomerQuery := `
	delete from "Customer"
	where user_id is null and id = $1 and merchant_id = $2
	`

	_, err = tx.Exec(ctx, deleteCcustomerQuery, customerId, merchantId)
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
	IsDummy         bool                    `json:"is_dummy"`
	IsBlacklisted   bool                    `json:"is_blacklisted"`
	BlacklistReason *string                 `json:"blacklist_reason"`
	TimesBooked     int                     `json:"times_booked"`
	TimesCancelled  int                     `json:"times_cancelled"`
	TimesUpcoming   int                     `json:"times_upcoming"`
	Appointments    []PublicAppointmentInfo `json:"appointments"`
}

func (s *service) GetCustomerStatsByMerchant(ctx context.Context, merchantId uuid.UUID, customerId uuid.UUID) (CustomerStatistics, error) {
	query := `
	with appointments as (
		select a.customer_id, 
			jsonb_agg(
				jsonb_build_object(
					'from_date', a.from_date,
					'to_date', a.to_date,
					'service_name', s.name,
					'price', a.price_then,
					'price_note', s.price_note,
					'merchant_name', m.name,
					'short_location', l.address || ', ' || l.city || ' ' || l.postal_code || ', ' || l.country,
					'cancelled_by_user', a.cancelled_by_user_on IS NOT NULL,
					'cancelled_by_merchant', a.cancelled_by_merchant_on IS NOT NULL
				) order by a.from_date desc
			) as appointments
		from (
			select distinct on (a.group_id) a.customer_id, a.group_id, min(a.from_date) over (partition by a.group_id) as from_date,
			max(a.to_date) over (partition by a.group_id) as to_date, a.merchant_id, a.location_id, a.service_id, a.price_then, a.cancelled_by_user_on, a.cancelled_by_merchant_on
			from "Appointment" a where a.merchant_id = $1 and (customer_id = $2 or a.transferred_to = $2)
		) a
		join "Service" s on s.id = a.service_id
		join "Merchant" m on m.id = a.merchant_id
		join "Location" l on l.id = a.location_id
		group by a.customer_id
	)
	select c.id, coalesce(c.first_name, u.first_name) as first_name, coalesce(c.last_name, u.last_name) as last_name, 
	coalesce(c.email, u.email) as email, coalesce(c.phone_number, u.phone_number) as phone_number,birthday, note, c.user_id is null as is_dummy, c.is_blacklisted, c.blacklist_reason,
	count(distinct a.group_id) as times_booked, count(distinct case when a.cancelled_by_user_on is not null then a.group_id end) as times_cancelled,
	count(distinct case when a.cancelled_by_user_on is null and a.cancelled_by_merchant_on is null and a.from_date >= now() then group_id end) as times_upcoming,
	coalesce(ca.appointments, '[]'::jsonb) as appointments
	from "Customer" c
	left join "User" u on u.id = c.user_id
	left join "Appointment" a on c.id = a.customer_id and a.merchant_id = $1
	left join appointments ca on c.id = ca.customer_id
	where c.id = $2 and c.merchant_id = $1
	GROUP BY c.id, u.first_name, u.last_name, u.email, u.phone_number, ca.appointments
	`

	var customer CustomerStatistics
	var appointmentsJSON []byte

	err := s.db.QueryRow(ctx, query, merchantId, customerId).Scan(&customer.Id, &customer.FirstName, &customer.LastName, &customer.Email, &customer.PhoneNumber, &customer.Birthday,
		&customer.Note, &customer.IsDummy, &customer.IsBlacklisted, &customer.BlacklistReason, &customer.TimesBooked, &customer.TimesCancelled, &customer.TimesUpcoming, &appointmentsJSON)
	if err != nil {
		return CustomerStatistics{}, err
	}

	if len(appointmentsJSON) > 0 {
		err = json.Unmarshal(appointmentsJSON, &customer.Appointments)
		if err != nil {
			return CustomerStatistics{}, err
		}
	} else {
		customer.Appointments = []PublicAppointmentInfo{}
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
