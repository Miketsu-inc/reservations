package auth

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/miketsu-inc/reservations/backend/internal/api/middleware/jwt"
	"github.com/miketsu-inc/reservations/backend/internal/api/middleware/lang"
	"github.com/miketsu-inc/reservations/backend/internal/domain"
	"github.com/miketsu-inc/reservations/backend/internal/jobs/args"
	"github.com/miketsu-inc/reservations/backend/internal/keys"
	merchantServ "github.com/miketsu-inc/reservations/backend/internal/service/merchant"
	"github.com/miketsu-inc/reservations/backend/internal/types"
	"github.com/miketsu-inc/reservations/backend/pkg/currencyx"
	"github.com/miketsu-inc/reservations/backend/pkg/db"
	"github.com/miketsu-inc/reservations/backend/pkg/oauthutil"
	"github.com/miketsu-inc/reservations/backend/pkg/queue"
	"github.com/miketsu-inc/reservations/backend/pkg/validate"
	"github.com/redis/go-redis/v9"
	"github.com/riverqueue/river"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	merchantRepo domain.MerchantRepository
	userRepo     domain.UserRepository
	teamRepo     domain.TeamRepository
	kv           *redis.Client
	enqueuer     queue.Enqueuer
	txManager    db.TransactionManager
}

func NewService(merchant domain.MerchantRepository, user domain.UserRepository, team domain.TeamRepository,
	kv *redis.Client, enqueuer queue.Enqueuer, txManager db.TransactionManager) *Service {
	return &Service{
		merchantRepo: merchant,
		userRepo:     user,
		teamRepo:     team,
		kv:           kv,
		enqueuer:     enqueuer,
		txManager:    txManager,
	}
}

func (s *Service) SetEnqueuer(client queue.Enqueuer) {
	s.enqueuer = client
}

func hashPassword(password string) (string, error) {
	passwordShaHash := fmt.Sprintf("%x", sha256.Sum256([]byte(password)))
	bytes, err := bcrypt.GenerateFromPassword([]byte(passwordShaHash), 14)
	return string(bytes), err
}

func hashCompare(password, hash string) error {
	passwordShaHash := fmt.Sprintf("%x", sha256.Sum256([]byte(password)))
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(passwordShaHash))

	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return fmt.Errorf("incorrect email or password")
		} else {
			// for debug purposes
			return err
		}
	}

	return nil
}

func newJwtTokens(userId uuid.UUID, refreshVersion int) (jwt.TokenPair, error) {
	accessToken, err := jwt.NewAccessToken(userId)
	if err != nil {
		return jwt.TokenPair{}, err
	}

	refreshToken, err := jwt.NewRefreshToken(userId, refreshVersion)
	if err != nil {
		return jwt.TokenPair{}, err
	}

	return jwt.TokenPair{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

type LoginInput struct {
	Email    string
	Password string
}

func (s *Service) Login(ctx context.Context, input LoginInput) (jwt.TokenPair, error) {
	user, err := s.userRepo.GetUserByEmail(ctx, input.Email)
	if err != nil {
		return jwt.TokenPair{}, err
	}

	err = hashCompare(input.Password, *user.PasswordHash)
	if err != nil {
		return jwt.TokenPair{}, err
	}

	tokens, err := newJwtTokens(user.Id, user.JwtRefreshVersion)
	if err != nil {
		return jwt.TokenPair{}, err
	}

	return tokens, nil
}

type UserSignupInput struct {
	FirstName   string
	LastName    string
	Email       string
	PhoneNumber string
	Password    string
}

func (s *Service) UserSignup(ctx context.Context, input UserSignupInput) (jwt.TokenPair, error) {
	err := s.userRepo.IsEmailUnique(ctx, input.Email)
	if err != nil {
		return jwt.TokenPair{}, err
	}

	err = s.userRepo.IsPhoneNumberUnique(ctx, input.PhoneNumber)
	if err != nil {
		return jwt.TokenPair{}, err
	}

	hashedPassword, err := hashPassword(input.Password)
	if err != nil {
		return jwt.TokenPair{}, err
	}

	userID, err := uuid.NewV7()
	if err != nil {
		return jwt.TokenPair{}, fmt.Errorf("unexpected error during creating user id: %s", err.Error())
	}

	err = s.userRepo.NewUser(ctx, domain.User{
		Id:                userID,
		FirstName:         input.FirstName,
		LastName:          input.LastName,
		Email:             input.Email,
		PhoneNumber:       &input.PhoneNumber,
		PasswordHash:      &hashedPassword,
		JwtRefreshVersion: 0,
		Language:          lang.LangFromContext(ctx).String(),
	})
	if err != nil {
		return jwt.TokenPair{}, err
	}

	tokens, err := newJwtTokens(userID, 0)
	if err != nil {
		return jwt.TokenPair{}, err
	}

	return tokens, nil
}

type MerchantSignupInput struct {
	Name         string
	ContactEmail string
	Timezone     string
}

func ctBH(timeStr string) time.Time {
	t, _ := time.Parse("15:04", timeStr)
	return time.Date(0, time.January, 1, t.Hour(), t.Minute(), 0, 0, time.UTC)
}

// TODO: create new jwts here... just don't know what to put as location
// I feel like most of this should be in merchant services
func (s *Service) MerchantSignup(ctx context.Context, input MerchantSignupInput) error {
	urlName, err := validate.MerchantNameToUrlName(input.Name)
	if err != nil {
		return fmt.Errorf("unexpected error during merchant url name conversion: %s", err.Error())
	}

	unique, err := s.merchantRepo.IsMerchantUrlUnique(ctx, urlName)
	if err != nil {
		return err
	}

	if !unique {
		return merchantServ.ErrMerchantUrlNotUnique{URL: urlName}
	}

	userID := jwt.MustGetUserIDFromContext(ctx)

	merchantID, err := uuid.NewV7()
	if err != nil {
		return fmt.Errorf("unexpected error during creating merchant id: %s", err.Error())
	}

	language := lang.LangFromContext(ctx)
	curr := currencyx.FindBest(language)

	err = s.txManager.WithTransaction(ctx, func(tx pgx.Tx) error {
		err := s.merchantRepo.WithTx(tx).NewMerchant(ctx, userID, domain.Merchant{
			Id:               merchantID,
			Name:             input.Name,
			UrlName:          urlName,
			ContactEmail:     input.ContactEmail,
			Introduction:     "",
			Announcement:     "",
			AboutUs:          "",
			ParkingInfo:      "",
			PaymentInfo:      "",
			Timezone:         input.Timezone,
			CurrencyCode:     curr,
			SubscriptionTier: types.SubTierFree,
		})
		if err != nil {
			return err
		}

		err = s.merchantRepo.WithTx(tx).NewPreferences(ctx, merchantID)
		if err != nil {
			return err
		}

		err = s.teamRepo.WithTx(tx).NewEmployee(ctx, merchantID, domain.PublicEmployee{
			UserId:   &userID,
			Role:     types.EmployeeRoleOwner,
			IsActive: true,
		})
		if err != nil {
			return err
		}

		businessHours := domain.BusinessHours{
			0: {{StartTime: ctBH("09:00"), EndTime: ctBH("17:00")}},
			1: {{StartTime: ctBH("09:00"), EndTime: ctBH("17:00")}},
			2: {{StartTime: ctBH("09:00"), EndTime: ctBH("17:00")}},
			3: {{StartTime: ctBH("09:00"), EndTime: ctBH("17:00")}},
			4: {{StartTime: ctBH("09:00"), EndTime: ctBH("17:00")}},
			5: {{StartTime: ctBH("09:00"), EndTime: ctBH("17:00")}},
			6: {{StartTime: ctBH("09:00"), EndTime: ctBH("17:00")}},
		}

		err = s.merchantRepo.WithTx(tx).NewBusinessHours(ctx, merchantID, businessHours)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("unexpected error creating a merchant: %s", err.Error())
	}

	return nil
}

func (s *Service) LogoutAllDevices(ctx context.Context) error {
	userId := jwt.MustGetUserIDFromContext(ctx)

	_, err := s.userRepo.IncrementUserJwtRefreshVersion(ctx, userId)
	if err != nil {
		return err
	}

	return nil
}

type UpdatePasswordInput struct {
	OldPassword string
	NewPassword string
}

func (s *Service) UpdatePassword(ctx context.Context, in UpdatePasswordInput) (jwt.TokenPair, error) {
	userId := jwt.MustGetUserIDFromContext(ctx)

	user, err := s.userRepo.GetUser(ctx, userId)
	if err != nil {
		return jwt.TokenPair{}, err
	}

	if user.IsOauthUser() {
		return jwt.TokenPair{}, fmt.Errorf("oauth users can't update password")
	}

	err = hashCompare(in.OldPassword, *user.PasswordHash)
	if err != nil {
		return jwt.TokenPair{}, err
	}

	newPasswordHash, err := hashPassword(in.NewPassword)
	if err != nil {
		return jwt.TokenPair{}, fmt.Errorf("error during password hashing: %s", err.Error())
	}

	err = s.userRepo.UpdatePassword(ctx, userId, newPasswordHash)
	if err != nil {
		return jwt.TokenPair{}, err
	}

	refreshVersion, err := s.userRepo.IncrementUserJwtRefreshVersion(ctx, userId)
	if err != nil {
		return jwt.TokenPair{}, err
	}

	tokens, err := newJwtTokens(userId, refreshVersion)
	if err != nil {
		return jwt.TokenPair{}, err
	}

	return tokens, nil
}

type ForgotPasswordInput struct {
	Email string
}

func (s *Service) ForgotPassword(ctx context.Context, in ForgotPasswordInput) error {
	user, err := s.userRepo.GetUserByEmail(ctx, in.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}

		return err
	}

	if user.IsOauthUser() {
		return fmt.Errorf("oauth users must use the auth provider for password reset")
	}

	token, err := oauthutil.RandomString(32)
	if err != nil {
		return fmt.Errorf("error generating token: %w", err)
	}

	key := keys.PasswordReset{Token: token}.String()

	err = s.kv.Set(ctx, key, user.Id.String(), time.Minute*10).Err()
	if err != nil {
		return fmt.Errorf("error setting password reset token: %w", err)
	}

	_, err = s.enqueuer.Insert(ctx, args.ForgotPasswordEmail{
		Language: lang.LangFromContext(ctx),
		UserId:   user.Id,
		Token:    token,
	}, &river.InsertOpts{})
	if err != nil {
		return fmt.Errorf("error scheduling forgot password email: %w", err)
	}

	return nil
}

type ResetPasswordInput struct {
	Token    string
	Password string
}

func (s *Service) ResetPassword(ctx context.Context, in ResetPasswordInput) (jwt.TokenPair, error) {
	key := keys.PasswordReset{Token: in.Token}.String()

	userIdStr, err := s.kv.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return jwt.TokenPair{}, fmt.Errorf("invalid or expired token")
		}

		return jwt.TokenPair{}, fmt.Errorf("error retrieving userId from token: %w", err)
	}

	userId, err := uuid.Parse(userIdStr)
	if err != nil {
		return jwt.TokenPair{}, fmt.Errorf("error parsing uuid string: %w", err)
	}

	passwordHash, err := hashPassword(in.Password)
	if err != nil {
		return jwt.TokenPair{}, fmt.Errorf("error hashing password: %w", err)
	}

	err = s.userRepo.UpdatePassword(ctx, userId, passwordHash)
	if err != nil {
		return jwt.TokenPair{}, err
	}

	err = s.kv.Del(ctx, key).Err()
	if err != nil {
		return jwt.TokenPair{}, fmt.Errorf("error deleting key: %w", err)
	}

	refreshVersion, err := s.userRepo.IncrementUserJwtRefreshVersion(ctx, userId)
	if err != nil {
		return jwt.TokenPair{}, err
	}

	tokens, err := newJwtTokens(userId, refreshVersion)
	if err != nil {
		return jwt.TokenPair{}, err
	}

	return tokens, nil
}
