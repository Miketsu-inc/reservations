package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/miketsu-inc/reservations/backend/internal/domain"
	"github.com/miketsu-inc/reservations/backend/internal/types"
	"github.com/miketsu-inc/reservations/backend/pkg/db"
	"golang.org/x/text/language"
)

type userRepository struct {
	db db.DBTX
}

func NewUserRepository(db db.DBTX) domain.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) WithTx(tx db.DBTX) domain.UserRepository {
	return &userRepository{db: tx}
}

func (r *userRepository) NewUser(ctx context.Context, user domain.User) error {
	query := `
	insert into "User" (id, first_name, last_name, email, phone_number, password_hash, jwt_refresh_version, preferred_lang,
		auth_provider, provider_id)
	values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := r.db.Exec(ctx, query, user.Id, user.FirstName, user.LastName, user.Email, user.PhoneNumber, user.PasswordHash,
		user.JwtRefreshVersion, user.PreferredLang, user.AuthProvider, user.ProviderId)
	if err != nil {
		return err
	}

	return nil
}

func (r *userRepository) GetUser(ctx context.Context, user_id uuid.UUID) (domain.User, error) {
	query := `
	select * from "User"
	where id = $1
	`

	var user domain.User
	err := r.db.QueryRow(ctx, query, user_id).Scan(&user.Id, &user.FirstName, &user.LastName, &user.Email, &user.PhoneNumber, &user.PasswordHash,
		&user.JwtRefreshVersion, &user.PreferredLang, &user.AuthProvider, &user.ProviderId)
	if err != nil {
		return domain.User{}, err
	}

	return user, nil
}

func (r *userRepository) GetUserPasswordAndIDByUserEmail(ctx context.Context, email string) (uuid.UUID, *string, error) {
	query := `
	select id, password_hash from "User"
	where email = $1
	`

	var userID uuid.UUID
	var password *string
	err := r.db.QueryRow(ctx, query, email).Scan(&userID, &password)
	if err != nil {
		return uuid.Nil, nil, err
	}

	return userID, password, nil
}

func (r *userRepository) GetUserJwtRefreshVersion(ctx context.Context, userID uuid.UUID) (int, error) {
	query := `
	select jwt_refresh_version from "User"
	where id = $1
	`

	var refreshVersion int
	err := r.db.QueryRow(ctx, query, userID).Scan(&refreshVersion)
	if err != nil {
		return 0, err
	}

	return refreshVersion, nil
}

func (r *userRepository) GetUserPreferredLanguage(ctx context.Context, userId uuid.UUID) (*language.Tag, error) {
	query := `
	select preferred_lang from "User"
	where id = $1
	`

	var pl *string
	err := r.db.QueryRow(ctx, query, userId).Scan(&pl)
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

func (r *userRepository) GetEmployeeByUser(ctx context.Context, merchantId uuid.UUID, userId uuid.UUID) (domain.EmployeeAuthInfo, error) {
	query := `
	select e.id, l.id as location_id, e.merchant_id, e.role
	from "Employee" e
	join "Location" l on l.merchant_id = e.merchant_id
	where e.merchant_id = $1 and user_id = $2
	`

	rows, _ := r.db.Query(ctx, query, merchantId, userId)
	employeeAuthInfo, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.EmployeeAuthInfo])
	if err != nil {
		return domain.EmployeeAuthInfo{}, err
	}

	return employeeAuthInfo, nil
}

func (r *userRepository) GetEmployeesByUser(ctx context.Context, userId uuid.UUID) ([]domain.EmployeeAuthInfo, error) {
	query := `
	select e.id, l.id as location_id, e.merchant_id, e.role
	from "Employee" e
	join "Location" l on l.merchant_id = e.merchant_id
	where user_id = $1
	`

	rows, _ := r.db.Query(ctx, query, userId)
	employeeAuthInfo, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.EmployeeAuthInfo])
	if err != nil {
		return []domain.EmployeeAuthInfo{}, err
	}

	return employeeAuthInfo, nil
}

func (r *userRepository) IsEmailUnique(ctx context.Context, email string) error {
	query := `
	select 1 from "User"
	where email = $1
	`

	var em *string
	err := r.db.QueryRow(ctx, query, email).Scan(&em)
	if !errors.Is(err, pgx.ErrNoRows) {
		if err != nil {
			return err
		}

		return fmt.Errorf("this email is already used: %s", email)
	}

	return nil
}

func (r *userRepository) IsPhoneNumberUnique(ctx context.Context, phoneNumber string) error {
	query := `
	select 1 from "User"
	where phone_number = $1
	`

	var pn *string
	err := r.db.QueryRow(ctx, query, phoneNumber).Scan(&pn)
	if !errors.Is(err, pgx.ErrNoRows) {
		if err != nil {
			return err
		}

		return fmt.Errorf("this phone number is already used: %s", phoneNumber)
	}

	return nil
}

func (r *userRepository) IncrementUserJwtRefreshVersion(ctx context.Context, userID uuid.UUID) error {
	query := `
	update "User"
	set jwt_refresh_version = jwt_refresh_version + 1
	where id = $1
	`

	_, err := r.db.Exec(ctx, query, userID)
	if err != nil {
		return err
	}

	return nil
}

func (r *userRepository) FindOauthUser(ctx context.Context, provider types.AuthProviderType, provider_id string) (uuid.UUID, error) {
	query := `
	select id from "User"
	where auth_provider = $1 and provider_id = $2
	`

	var id uuid.UUID
	err := r.db.QueryRow(ctx, query, provider, provider_id).Scan(&id)
	if err != nil {
		return uuid.UUID{}, err
	}

	return id, nil
}
