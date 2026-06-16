package domain

import (
	"context"

	"github.com/google/uuid"
	"github.com/miketsu-inc/reservations/backend/internal/types"
	"github.com/miketsu-inc/reservations/backend/pkg/db"
	"golang.org/x/text/language"
)

type UserRepository interface {
	WithTx(tx db.DBTX) UserRepository

	NewUser(ctx context.Context, user User) error

	GetUser(ctx context.Context, userId uuid.UUID) (User, error)
	GetUserByEmail(ctx context.Context, email string) (User, error)
	GetUserJwtRefreshVersion(ctx context.Context, userId uuid.UUID) (int, error)
	GetUserLanguage(ctx context.Context, userId uuid.UUID) (language.Tag, error)
	GetEmployeeByUser(ctx context.Context, merchantId uuid.UUID, userId uuid.UUID) (EmployeeAuthInfo, error)
	GetEmployeesByUser(ctx context.Context, userId uuid.UUID) ([]EmployeeAuthInfo, error)

	UpdateUser(ctx context.Context, user UserCore) error
	UpdatePassword(ctx context.Context, userId uuid.UUID, passwordHash string) error
	DeleteUser(ctx context.Context, userId uuid.UUID) error

	IsEmailUnique(ctx context.Context, email string) error
	IsPhoneNumberUnique(ctx context.Context, phoneNumber string) error

	// Increment User's refresh version, logging out the User.
	IncrementUserJwtRefreshVersion(ctx context.Context, userId uuid.UUID) (int, error)

	FindOauthUser(ctx context.Context, authProviderType types.AuthProviderType, providerId string) (uuid.UUID, error)
}

type User struct {
	Id                uuid.UUID               `json:"ID" db:"id"`
	FirstName         string                  `json:"first_name" db:"first_name"`
	LastName          string                  `json:"last_name" db:"last_name"`
	Email             string                  `json:"email" db:"email"`
	PhoneNumber       *string                 `json:"phone_number" db:"phone_number"`
	PasswordHash      *string                 `json:"password_hash" db:"password_hash"`
	JwtRefreshVersion int                     `json:"jwt_refresh_version" db:"jwt_refresh_version"`
	Language          string                  `json:"language" db:"language"`
	AuthProvider      *types.AuthProviderType `json:"auth_provider" db:"auth_provider"`
	ProviderId        *string                 `json:"provider_id" db:"provider_id"`
}

func (u User) IsOauthUser() bool {
	return u.AuthProvider != nil || u.ProviderId != nil
}

type UserCore struct {
	Id          uuid.UUID `db:"id"`
	FirstName   string    `db:"first_name"`
	LastName    string    `db:"last_name"`
	Email       string    `db:"email"`
	PhoneNumber *string   `db:"phone_number"`
}

type EmployeeAuthInfo struct {
	Id         int                `db:"id"`
	LocationId int                `db:"location_id"`
	MerchantId uuid.UUID          `db:"merchant_id"`
	Role       types.EmployeeRole `db:"role"`
}
