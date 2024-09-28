package database

import (
	"context"

	"github.com/google/uuid"
)

type User struct {
	Id           uuid.UUID       `json:"id"`
	FirstName    string          `json:"first_name"`
	LastName     string          `json:"last_name"`
	Email        string          `json:"email"`
	Password     string          `json:"password"`
	Subscription int             `json:"subscription"`
	Settings     map[string]bool `json:"settings"`
}

func (s *service) NewUser(ctx context.Context, user User) error {
	query := `
	insert into "user"(id, first_name, last_name, email, password, subscription, settings)
	values ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := s.db.ExecContext(ctx, query, user.Id, user.FirstName, user.LastName, user.Email, user.Password, user.Subscription, user.Settings)
	if err != nil {
		return err
	}

	return nil
}

func (s *service) GetUserById(ctx context.Context, id uuid.UUID) (User, error) {
	query := `
	select * from "user" where id = $1
	`

	var user User
	err := s.db.QueryRowContext(ctx, query, id).Scan(&user.Id, &user.FirstName, &user.LastName, &user.Email, &user.Password, &user.Subscription, &user.Settings)
	if err != nil {
		return User{}, err
	}

	return user, nil
}

func (s *service) GetUserPasswordByUserEmail(ctx context.Context, email string) (string, error) {
	query := `
	select password from "user" where email = $1
	`

	var password string
	err := s.db.QueryRowContext(ctx, query, email).Scan(&password)
	if err != nil {
		return "", err
	}

	return password, nil
}
