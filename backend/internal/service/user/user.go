package user

import (
	"context"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/miketsu-inc/reservations/backend/internal/api/middleware/jwt"
	"github.com/miketsu-inc/reservations/backend/internal/domain"
	"github.com/miketsu-inc/reservations/backend/internal/repository"
	"github.com/miketsu-inc/reservations/backend/pkg/db"
)

type Service struct {
	userRepo     domain.UserRepository
	customerRepo domain.CustomerRepository
	teamRepo     domain.TeamRepository
	txManager    db.TransactionManager
}

func NewService(user domain.UserRepository, customer domain.CustomerRepository, team domain.TeamRepository, txManager db.TransactionManager) *Service {
	return &Service{
		userRepo:     user,
		customerRepo: customer,
		teamRepo:     team,
		txManager:    txManager,
	}
}

type EditInput struct {
	FirstName   string
	LastName    string
	PhoneNumber string
	Email       string
}

func (s *Service) Edit(ctx context.Context, input EditInput) error {
	userId := jwt.MustGetUserIDFromContext(ctx)
	user, err := s.userRepo.GetUser(ctx, userId)
	if err != nil {
		return err
	}

	if !strings.EqualFold(input.Email, user.Email) {
		unique, err := s.userRepo.IsEmailUnique(ctx, input.Email)
		if err != nil {
			return err
		}

		if !unique {
			return domain.ErrEmailNotUnique
		}
	}

	if user.PhoneNumber == nil || input.PhoneNumber != *user.PhoneNumber {
		unique, err := s.userRepo.IsPhoneNumberUnique(ctx, input.PhoneNumber)
		if err != nil {
			return err
		}

		if !unique {
			return domain.ErrPhoneNumberNotUnique
		}
	}

	err = s.userRepo.UpdateUser(ctx, domain.UserCore{
		Id:          userId,
		FirstName:   input.FirstName,
		LastName:    input.LastName,
		PhoneNumber: &input.PhoneNumber,
		Email:       input.Email,
	})
	if err != nil {
		if db.IsUniqueConstraintViolation(err, repository.UserEmailLowerUniqueConstraint) {
			return domain.ErrEmailNotUnique
		}

		return err
	}

	return nil
}

// TODO: this should be tested at some point.
func (s *Service) Delete(ctx context.Context) error {
	userId := jwt.MustGetUserIDFromContext(ctx)

	return s.txManager.WithTransaction(ctx, func(tx pgx.Tx) error {
		err := s.customerRepo.WithTx(tx).CopyUserNameToCustomer(ctx, userId)
		if err != nil {
			return err
		}

		err = s.teamRepo.WithTx(tx).CopyUserNameToEmployee(ctx, userId)
		if err != nil {
			return err
		}

		return s.userRepo.WithTx(tx).DeleteUser(ctx, userId)
	})
}
