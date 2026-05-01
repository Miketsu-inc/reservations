package team

import (
	"context"
	"fmt"

	"github.com/miketsu-inc/reservations/backend/internal/api/middleware/actor"
	"github.com/miketsu-inc/reservations/backend/internal/api/middleware/jwt"
	"github.com/miketsu-inc/reservations/backend/internal/domain"
	"github.com/miketsu-inc/reservations/backend/internal/types"
)

type Service struct {
	teamRepo domain.TeamRepository
	userRepo domain.UserRepository
}

func NewService(team domain.TeamRepository, user domain.UserRepository) *Service {
	return &Service{
		teamRepo: team,
		userRepo: user,
	}
}

type MeResult struct {
	User        domain.User
	Memberships []domain.EmployeeAuthInfo
}

func (s *Service) Me(ctx context.Context) (MeResult, error) {
	userId := jwt.MustGetUserIDFromContext(ctx)

	user, err := s.userRepo.GetUser(ctx, userId)
	if err != nil {
		return MeResult{}, fmt.Errorf("error while retrieving user: %s", err.Error())
	}

	employeeInfo, err := s.userRepo.GetEmployeesByUser(ctx, userId)
	if err != nil {
		return MeResult{}, fmt.Errorf("error while getting employees for user: %s", err.Error())
	}

	return MeResult{
		User:        user,
		Memberships: employeeInfo,
	}, nil
}

type NewMemberInput struct {
	Role        types.EmployeeRole
	FirstName   string
	LastName    string
	Email       *string
	PhoneNumber *string
	IsActive    bool
}

func (s *Service) NewMember(ctx context.Context, input NewMemberInput) error {
	if input.Role == types.EmployeeRoleOwner {
		return fmt.Errorf("error there can only be 1 owner")
	}

	actor := actor.MustGetFromContext(ctx)

	err := s.teamRepo.NewEmployee(ctx, actor.MerchantId, domain.PublicEmployee{
		Role:        input.Role,
		FirstName:   &input.FirstName,
		LastName:    &input.LastName,
		Email:       input.Email,
		PhoneNumber: input.PhoneNumber,
		IsActive:    input.IsActive,
	})
	if err != nil {
		return fmt.Errorf("error while creating new employee by id: %s", err.Error())
	}

	return nil
}

type UpdateMemberInput struct {
	Role        types.EmployeeRole
	FirstName   string
	LastName    string
	Email       *string
	PhoneNumber *string
	IsActive    bool
}

func (s *Service) UpdateMember(ctx context.Context, memberId int, input UpdateMemberInput) error {
	if input.Role == types.EmployeeRoleOwner {
		return fmt.Errorf("error there can only be 1 owner")
	}

	actor := actor.MustGetFromContext(ctx)

	err := s.teamRepo.UpdateEmployee(ctx, actor.MerchantId, domain.PublicEmployee{
		Id:          memberId,
		Role:        input.Role,
		FirstName:   &input.FirstName,
		LastName:    &input.LastName,
		Email:       input.Email,
		PhoneNumber: input.PhoneNumber,
		IsActive:    input.IsActive,
	})
	if err != nil {
		return fmt.Errorf("error while updating employee by id: %s", err.Error())
	}

	return nil
}

func (s *Service) DeleteMember(ctx context.Context, memberId int) error {
	actor := actor.MustGetFromContext(ctx)

	err := s.teamRepo.DeleteEmployee(ctx, actor.MerchantId, memberId)
	if err != nil {
		return fmt.Errorf("error while deleting employee by id: %s", err.Error())
	}

	return nil
}

func (s *Service) GetMember(ctx context.Context, memberId int) (domain.PublicEmployee, error) {
	actor := actor.MustGetFromContext(ctx)

	teamMember, err := s.teamRepo.GetEmployee(ctx, actor.MerchantId, memberId)
	if err != nil {
		return domain.PublicEmployee{}, fmt.Errorf("error while retrieving employee by id: %s", err.Error())
	}

	return teamMember, nil
}

func (s *Service) GetTeam(ctx context.Context) ([]domain.PublicEmployee, error) {
	actor := actor.MustGetFromContext(ctx)

	teamMembers, err := s.teamRepo.GetEmployees(ctx, actor.MerchantId)
	if err != nil {
		return []domain.PublicEmployee{}, fmt.Errorf("error while retrieving employees for merchant: %s", err.Error())
	}

	return teamMembers, nil
}
