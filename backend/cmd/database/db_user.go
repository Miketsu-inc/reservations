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
}

func (s *service) NewUser(ctx context.Context, user User) error {
	query := `
	insert into "User" (id, first_name, last_name, email, phone_number, password_hash, jwt_refresh_version, subscription)
	values ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := s.db.ExecContext(ctx, query, user.Id, user.FirstName, user.LastName, user.Email, user.PhoneNumber, user.PasswordHash, user.JwtRefreshVersion, user.Subscription)
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
	err := s.db.QueryRowContext(ctx, query, user_id).Scan(&user.Id, &user.FirstName, &user.LastName, &user.Email, &user.PhoneNumber, &user.PasswordHash, &user.JwtRefreshVersion, &user.Subscription)
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

func (s *service) IncrementUserJwtRefreshVersion(ctx context.Context, userID uuid.UUID) (int, error) {
	query := `
	update "User"
	set jwt_refresh_version = jwt_refresh_version + 1
	where id = $1
	returning jwt_refresh_version
	`

	var refreshVersion int
	err := s.db.QueryRowContext(ctx, query, userID).Scan(&refreshVersion)
	if err != nil {
		return 0, err
	}

	return refreshVersion, nil
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
