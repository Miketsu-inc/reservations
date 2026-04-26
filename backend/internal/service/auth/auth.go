package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/miketsu-inc/reservations/backend/internal/api/middleware/jwt"
	"github.com/miketsu-inc/reservations/backend/internal/api/middleware/lang"
	"github.com/miketsu-inc/reservations/backend/internal/domain"
	merchantServ "github.com/miketsu-inc/reservations/backend/internal/service/merchant"
	"github.com/miketsu-inc/reservations/backend/internal/types"
	"github.com/miketsu-inc/reservations/backend/pkg/currencyx"
	"github.com/miketsu-inc/reservations/backend/pkg/db"
	"github.com/miketsu-inc/reservations/backend/pkg/validate"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	merchantRepo domain.MerchantRepository
	userRepo     domain.UserRepository
	teamRepo     domain.TeamRepository
	txManager    db.TransactionManager
}

func NewService(merchant domain.MerchantRepository, user domain.UserRepository, team domain.TeamRepository,
	txManager db.TransactionManager) *Service {
	return &Service{
		merchantRepo: merchant,
		userRepo:     user,
		teamRepo:     team,
		txManager:    txManager,
	}
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func hashCompare(password, hash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))

	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return fmt.Errorf("incorrect email or password")

		} else if errors.Is(err, bcrypt.ErrPasswordTooLong) {
			return fmt.Errorf("password is too long")

		} else {
			// for debug purposes
			return err
		}
	}

	return nil
}

type LoginInput struct {
	Email    string
	Password string
}

func (s *Service) Login(ctx context.Context, input LoginInput) (jwt.TokenPair, error) {
	userID, password, err := s.userRepo.GetUserPasswordAndIDByUserEmail(ctx, input.Email)
	if err != nil {
		return jwt.TokenPair{}, fmt.Errorf("incorrect email or password %s", err.Error())
	}

	err = hashCompare(input.Password, *password)
	if err != nil {
		return jwt.TokenPair{}, err
	}

	employeeAuthInfo, err := s.userRepo.GetEmployeesByUser(ctx, userID)
	if err != nil {
		return jwt.TokenPair{}, fmt.Errorf("unexpected error when reading employees associated with user: %s", err.Error())
	}

	var merchantId *uuid.UUID
	var employeeId *int
	var locationId *int
	var role *types.EmployeeRole

	// TODO: later user should be able to select which merchant to log into
	if len(employeeAuthInfo) >= 1 {
		merchantId = &employeeAuthInfo[0].MerchantId
		employeeId = &employeeAuthInfo[0].Id
		locationId = &employeeAuthInfo[0].LocationId
		role = &employeeAuthInfo[0].Role
	}

	refreshVersion, err := s.userRepo.GetUserJwtRefreshVersion(ctx, userID)
	if err != nil {
		return jwt.TokenPair{}, fmt.Errorf("unexpected error when getting refresh version: %s", err.Error())
	}

	accessToken, err := jwt.NewAccessToken(userID, merchantId, employeeId, locationId, role)
	if err != nil {
		return jwt.TokenPair{}, err
	}

	refreshToken, err := jwt.NewRefreshToken(userID, merchantId, employeeId, locationId, role, refreshVersion)
	if err != nil {
		return jwt.TokenPair{}, err
	}

	return jwt.TokenPair{AccessToken: accessToken, RefreshToken: refreshToken}, nil
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
		return jwt.TokenPair{}, fmt.Errorf("the password is too long")
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
	})
	if err != nil {
		return jwt.TokenPair{}, fmt.Errorf("unexpected error when creating user: %s", err.Error())
	}

	accessToken, err := jwt.NewAccessToken(userID, nil, nil, nil, nil)
	if err != nil {
		return jwt.TokenPair{}, err
	}

	refreshToken, err := jwt.NewRefreshToken(userID, nil, nil, nil, nil, 0)
	if err != nil {
		return jwt.TokenPair{}, err
	}

	return jwt.TokenPair{AccessToken: accessToken, RefreshToken: refreshToken}, nil
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

	err := s.userRepo.IncrementUserJwtRefreshVersion(ctx, userId)
	if err != nil {
		return err
	}

	return nil
}
