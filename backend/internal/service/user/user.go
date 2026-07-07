package user

import (
	"context"

	"github.com/miketsu-inc/reservations/backend/internal/api/middleware/jwt"
	"github.com/miketsu-inc/reservations/backend/internal/domain"
)

type Service struct {
	userRepo domain.UserRepository
}

func NewService(user domain.UserRepository) *Service {
	return &Service{
		userRepo: user,
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

	err := s.userRepo.UpdateUser(ctx, domain.UserCore{
		Id:          userId,
		FirstName:   input.FirstName,
		LastName:    input.LastName,
		PhoneNumber: &input.PhoneNumber,
		Email:       input.Email,
	})
	if err != nil {
		return err
	}

	return nil
}

// TODO: this should be tested at some point.
func (s *Service) Delete(ctx context.Context) error {
	userId := jwt.MustGetUserIDFromContext(ctx)

	err := s.userRepo.DeleteUser(ctx, userId)
	if err != nil {
		return err
	}

	return nil
}
