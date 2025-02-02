package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

type User struct {
	Id                uuid.UUID `json:"ID"`
	FirstName         string    `json:"first_name"`
	LastName          string    `json:"last_name"`
	Email             string    `json:"email"`
	PhoneNumber       string    `json:"phone_number"`
	PasswordHash      string    `json:"password_hash"`
	JwtRefreshVersion int       `json:"jwt_refresh_version"`
	Subscription      int       `json:"subscription"`
	IsDummy           bool      `json:"is_dummy"`
	AddedBy           uuid.UUID `json:"added_by"`
}

func (s *service) NewUser(ctx context.Context, user User) error {
	query := `
	insert into "User" (id, first_name, last_name, email, phone_number, password_hash, jwt_refresh_version, subscription, is_dummy, added_by)
	values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := s.db.ExecContext(ctx, query, user.Id, user.FirstName, user.LastName, user.Email, user.PhoneNumber, user.PasswordHash,
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
	err := s.db.QueryRowContext(ctx, query, user_id).Scan(&user.Id, &user.FirstName, &user.LastName, &user.Email, &user.PhoneNumber, &user.PasswordHash,
		&user.JwtRefreshVersion, &user.Subscription, &user.IsDummy, &user.AddedBy)
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
	err := s.db.QueryRowContext(ctx, query, email).Scan(&userID, &password)
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
	err := s.db.QueryRowContext(ctx, query, email).Scan(&em)
	if !errors.Is(err, sql.ErrNoRows) {
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
	err := s.db.QueryRowContext(ctx, query, phoneNumber).Scan(&pn)
	if !errors.Is(err, sql.ErrNoRows) {
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

	_, err := s.db.ExecContext(ctx, query, userID)
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
	err := s.db.QueryRowContext(ctx, query, userID).Scan(&refreshVersion)
	if err != nil {
		return 0, err
	}

	return refreshVersion, nil
}

type Customer struct {
	Id        uuid.UUID `json:"id"`
	FirstName string    `json:"fist_name"`
	LastName  string    `json:"last_name"`
	IsDummy   bool      `json:"is_dummy"`
}

func (s *service) NewCustomer(ctx context.Context, merchantId uuid.UUID, customer Customer) error {
	query := `
	insert into "User" (id, first_name, last_name, is_dummy, added_by)
	values ($1, $2, $3, $4, $5)
	`

	_, err := s.db.ExecContext(ctx, query, customer.Id, customer.FirstName, customer.LastName, customer.IsDummy, merchantId)
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

	_, err := s.db.ExecContext(ctx, query, customerId, merchantId)
	if err != nil {
		return err
	}

	return nil
}

func (s *service) UpdateCustomerById(ctx context.Context, merchantId uuid.UUID, customer Customer) error {
	query := `
	update "User"
	set first_name = $3, last_name = $4
	where is_dummy = true and id = $2 and added_by = $1
	`

	_, err := s.db.ExecContext(ctx, query, merchantId, customer.Id, customer.FirstName, customer.LastName)
	if err != nil {
		return err
	}

	return nil
}
