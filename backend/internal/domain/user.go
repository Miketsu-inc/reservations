package domain

import (
	"context"
	"time"

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
	GetEmployeeActorInfo(ctx context.Context, merchantId uuid.UUID, userId uuid.UUID) (EmployeeActorInfo, error)
	GetEmployeesByUser(ctx context.Context, userId uuid.UUID) ([]EmployeeAuthInfo, error)

	UpdateUser(ctx context.Context, user UserCore) error
	UpdatePassword(ctx context.Context, userId uuid.UUID, passwordHash string) error
	SetUserAvatar(ctx context.Context, userId uuid.UUID, imageId *uuid.UUID) (*uuid.UUID, error)
	DeleteUser(ctx context.Context, userId uuid.UUID) error

	IsEmailUnique(ctx context.Context, email string) (bool, error)
	IsPhoneNumberUnique(ctx context.Context, phoneNumber string) (bool, error)

	// Increment User's refresh version, logging out the User.
	IncrementUserJwtRefreshVersion(ctx context.Context, userId uuid.UUID) (int, error)

	FindOauthUser(ctx context.Context, authProviderType types.AuthProviderType, providerId string) (uuid.UUID, error)
}

type User struct {
	Id                uuid.UUID               `db:"id"`
	FirstName         string                  `db:"first_name"`
	LastName          string                  `db:"last_name"`
	Email             string                  `db:"email"`
	PhoneNumber       *string                 `db:"phone_number"`
	PasswordHash      *string                 `db:"password_hash"`
	JwtRefreshVersion int                     `db:"jwt_refresh_version"`
	Language          string                  `db:"language"`
	AuthProvider      *types.AuthProviderType `db:"auth_provider"`
	ProviderId        *string                 `db:"provider_id"`
	AvatarImageId     *uuid.UUID              `db:"avatar_image_id"`
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
	Id              int                `db:"id"`
	MerchantId      uuid.UUID          `db:"merchant_id"`
	MerchantName    string             `db:"merchant_name"`
	MerchantUrlName string             `db:"merchant_url_name"`
	Role            types.EmployeeRole `db:"role"`
	AvatarImageId   *uuid.UUID         `db:"avatar_image_id"`
}

type EmployeeActorInfo struct {
	Id           int
	Role         types.EmployeeRole
	SubTier      types.SubTier
	Timezone     *time.Location
	CurrencyCode string
}
