package database

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

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
	IsDummy           bool      `json:"is_dummy" db:"is_dummy"`
	AddedBy           uuid.UUID `json:"added_by" db:"added_by"`
	PreferredLang     *string   `json:"preferred_lang" db:"preferred_lang"`
}

func (s *service) NewUser(ctx context.Context, user User) error {
	query := `
	insert into "User" (id, first_name, last_name, email, phone_number, password_hash, jwt_refresh_version, subscription, is_dummy, added_by)
	values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := s.db.Exec(ctx, query, user.Id, user.FirstName, user.LastName, user.Email, user.PhoneNumber, user.PasswordHash,
		user.JwtRefreshVersion, user.Subscription, user.IsDummy, user.AddedBy)
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
		&user.JwtRefreshVersion, &user.Subscription, &user.IsDummy, &user.AddedBy, &user.PreferredLang)
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
	Id          uuid.UUID `json:"id" db:"id"`
	FirstName   string    `json:"first_name" db:"first_name"`
	LastName    string    `json:"last_name" db:"last_name"`
	Email       string    `json:"email" db:"email"`
	PhoneNumber string    `json:"phone_number" db:"phone_number"`
	IsDummy     bool      `json:"is_dummy" db:"is_dummy"`
}

func (s *service) NewCustomer(ctx context.Context, merchantId uuid.UUID, customer Customer) error {
	query := `
	insert into "User" (id, first_name, last_name, email, phone_number, is_dummy, added_by)
	values ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := s.db.Exec(ctx, query, customer.Id, customer.FirstName, customer.LastName, customer.Email, customer.PhoneNumber, customer.IsDummy, merchantId)
	if err != nil {
		return err
	}

	return nil
}

func (s *service) DeleteCustomerById(ctx context.Context, customerId uuid.UUID, merchantId uuid.UUID) error {
	query := `
	delete from "User"
	where is_dummy = true and id = $1 and added_by = $2
	`

	_, err := s.db.Exec(ctx, query, customerId, merchantId)
	if err != nil {
		return err
	}

	return nil
}

func (s *service) UpdateCustomerById(ctx context.Context, merchantId uuid.UUID, customer Customer) error {
	query := `
	update "User"
	set first_name = $3, last_name = $4, email = $5, phone_number = $6
	where is_dummy = true and id = $2 and added_by = $1
	`

	_, err := s.db.Exec(ctx, query, merchantId, customer.Id, customer.FirstName, customer.LastName, customer.Email, customer.PhoneNumber)
	if err != nil {
		return err
	}

	return nil
}

func (s *service) AddCustomerToBlacklist(ctx context.Context, merchantId uuid.UUID, customerId uuid.UUID, reason string) error {
	query := `
	insert into "Blacklist" (merchant_id, user_id, reason)
	values ($1, $2, $3)
	`

	_, err := s.db.Exec(ctx, query, merchantId, customerId, reason)
	if err != nil {
		return err
	}

	return nil
}

func (s *service) RemoveCustomerFromBlacklist(ctx context.Context, merchantId uuid.UUID, customerId uuid.UUID) error {
	query := `
	delete from "Blacklist"
	where merchant_id = $1 and user_id = $2
	`

	_, err := s.db.Exec(ctx, query, merchantId, customerId)
	if err != nil {
		return err
	}

	return nil
}

func (s *service) IsUserBlacklisted(ctx context.Context, merchantId uuid.UUID, userId uuid.UUID) error {
	query := `
	select 1 from "Blacklist"
	where merchant_id = $1 and user_id = $2
	`

	var st string
	err := s.db.QueryRow(ctx, query, merchantId, userId).Scan(&st)
	if !errors.Is(err, pgx.ErrNoRows) {
		if err != nil {
			return err
		}

		return fmt.Errorf("please contact the merchant by email or phone to book an appointment")
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

type AllCustomerInfo struct {
	Customer
	IsBlacklisted   bool                    `json:"is_blacklisted"`
	BlacklistReason *string                 `json:"blacklist_reason"`
	TimesBooked     int                     `json:"times_booked"`
	TimesCancelled  int                     `json:"times_cancelled"`
	TimesUpcoming   int                     `json:"times_upcoming"`
	Appointments    []PublicAppointmentInfo `json:"appointments"`
}

func (s *service) GetCustomerInfoByMerchant(ctx context.Context, merchantId uuid.UUID, customerId uuid.UUID) (AllCustomerInfo, error) {
	query := `
	with appointments as (
		select a.user_id, 
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
			select distinct on (a.group_id) a.user_id, a.group_id, min(a.from_date) over (partition by a.group_id) as from_date,
			max(a.to_date) over (partition by a.group_id) as to_date, a.merchant_id, a.location_id, a.service_id, a.price_then, a.cancelled_by_user_on, a.cancelled_by_merchant_on
			from "Appointment" a where a.merchant_id = $1 and a.user_id = $2
		) a
		join "Service" s on s.id = a.service_id
		join "Merchant" m on m.id = a.merchant_id
		join "Location" l on l.id = a.location_id
		group by a.user_id
	)
	select u.id, u.first_name, u.last_name, u.email, u.phone_number, u.is_dummy, b.user_id is not null as is_blacklisted, b.reason as blacklist_reason,
	count(distinct a.group_id) as times_booked, count(distinct case when a.cancelled_by_user_on is not null then a.group_id end) as times_cancelled,
	count(distinct case when a.cancelled_by_user_on is null and a.cancelled_by_merchant_on is null and a.from_date >= now() then group_id end) as times_upcoming,
	coalesce(ca.appointments, '[]'::jsonb) as appointments
	from "User" u
	left join "Appointment" a on u.id = a.user_id and a.merchant_id = $1
	left join "Blacklist" b on u.id = b.user_id and b.merchant_id = $1
	left join appointments ca on u.id = ca.user_id
	where u.id = $2
	GROUP BY u.id, u.first_name, u.last_name, u.email, u.phone_number, u.is_dummy, b.user_id, b.reason, ca.appointments
	`

	var customer AllCustomerInfo
	var appointmentsJSON []byte

	err := s.db.QueryRow(ctx, query, merchantId, customerId).Scan(&customer.Id, &customer.FirstName, &customer.LastName, &customer.Email, &customer.PhoneNumber, &customer.IsDummy, &customer.IsBlacklisted,
		&customer.BlacklistReason, &customer.TimesBooked, &customer.TimesCancelled, &customer.TimesUpcoming, &appointmentsJSON)
	if err != nil {
		return AllCustomerInfo{}, err
	}

	if len(appointmentsJSON) > 0 {
		err = json.Unmarshal(appointmentsJSON, &customer.Appointments)
		if err != nil {
			return AllCustomerInfo{}, err
		}
	} else {
		customer.Appointments = []PublicAppointmentInfo{}
	}

	return customer, nil
}
