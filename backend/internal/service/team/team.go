package team

import (
	"context"
	"fmt"

	"github.com/miketsu-inc/reservations/backend/internal/api/middleware/jwt"
	"github.com/miketsu-inc/reservations/backend/internal/domain"
	"github.com/miketsu-inc/reservations/backend/internal/types"
)

type Service struct {
	teamRepo domain.TeamRepository
}

func NewService(team domain.TeamRepository) *Service {
	return &Service{
		teamRepo: team,
	}
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

	employeeAuth := jwt.MustGetEmployeeFromContext(ctx)

	err := s.teamRepo.NewEmployee(ctx, employeeAuth.MerchantId, domain.PublicEmployee{
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

	employeeAuth := jwt.MustGetEmployeeFromContext(ctx)

	err := s.teamRepo.UpdateEmployee(ctx, employeeAuth.MerchantId, domain.PublicEmployee{
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
	employeeAuth := jwt.MustGetEmployeeFromContext(ctx)

	err := s.teamRepo.DeleteEmployee(ctx, employeeAuth.MerchantId, memberId)
	if err != nil {
		return fmt.Errorf("error while deleting employee by id: %s", err.Error())
	}

	return nil
}

func (s *Service) GetMember(ctx context.Context, memberId int) (domain.PublicEmployee, error) {
	employeeAuth := jwt.MustGetEmployeeFromContext(ctx)

	teamMember, err := s.teamRepo.GetEmployee(ctx, employeeAuth.MerchantId, memberId)
	if err != nil {
		return domain.PublicEmployee{}, fmt.Errorf("error while retrieving employee by id: %s", err.Error())
	}

	return teamMember, nil
}

func (s *Service) GetTeam(ctx context.Context) ([]domain.PublicEmployee, error) {
	employeeAuth := jwt.MustGetEmployeeFromContext(ctx)

	teamMembers, err := s.teamRepo.GetEmployees(ctx, employeeAuth.MerchantId)
	if err != nil {
		return []domain.PublicEmployee{}, fmt.Errorf("error while retrieving employees for merchant: %s", err.Error())
	}

	return teamMembers, nil
}
