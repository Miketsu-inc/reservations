package domain

import (
	"context"

	"github.com/google/uuid"
	"github.com/miketsu-inc/reservations/backend/internal/types"
	"golang.org/x/text/language"
)

type UserRepository interface {
	NewUser(context.Context, User) error

	GetUserById(context.Context, uuid.UUID) (User, error)
	GetUserPasswordAndIDByUserEmail(context.Context, string) (uuid.UUID, *string, error)
	FindOauthUser(context.Context, types.AuthProviderType, string) (uuid.UUID, error)
	GetEmployeesByUser(context.Context, uuid.UUID) ([]EmployeeAuthInfo, error)

	IsEmailUnique(context.Context, string) error
	IsPhoneNumberUnique(context.Context, string) error

	// Increment User's refresh version, logging out the User.
	IncrementUserJwtRefreshVersion(context.Context, uuid.UUID) error
	GetUserJwtRefreshVersion(context.Context, uuid.UUID) (int, error)

	GetUserPreferredLanguage(context.Context, uuid.UUID) (*language.Tag, error)
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
