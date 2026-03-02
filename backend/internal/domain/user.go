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
	GetUserPasswordAndIDByUserEmail(ctx context.Context, email string) (uuid.UUID, *string, error)
	GetUserJwtRefreshVersion(ctx context.Context, userId uuid.UUID) (int, error)
	GetUserPreferredLanguage(ctx context.Context, userId uuid.UUID) (*language.Tag, error)
	GetEmployeesByUser(ctx context.Context, userId uuid.UUID) ([]EmployeeAuthInfo, error)

	IsEmailUnique(ctx context.Context, email string) error
	IsPhoneNumberUnique(ctx context.Context, phoneNumber string) error

	// Increment User's refresh version, logging out the User.
	IncrementUserJwtRefreshVersion(ctx context.Context, userId uuid.UUID) error

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
	PreferredLang     *string                 `json:"preferred_lang" db:"preferred_lang"`
	AuthProvider      *types.AuthProviderType `json:"auth_provider" db:"auth_provider"`
	ProviderId        *string                 `json:"provider_id" db:"provider_id"`
}

type EmployeeAuthInfo struct {
	Id         int                `db:"id"`
	LocationId int                `db:"location_id"`
	MerchantId uuid.UUID          `db:"merchant_id"`
	Role       types.EmployeeRole `db:"role"`
}
