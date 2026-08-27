package team

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/miketsu-inc/reservations/backend/internal/api/middleware/actor"
	"github.com/miketsu-inc/reservations/backend/internal/api/middleware/jwt"
	"github.com/miketsu-inc/reservations/backend/internal/domain"
	"github.com/miketsu-inc/reservations/backend/internal/types"
	"github.com/miketsu-inc/reservations/backend/internal/utils"
	"github.com/miketsu-inc/reservations/backend/pkg/db"
	"github.com/miketsu-inc/reservations/backend/pkg/queue"
	"github.com/miketsu-inc/reservations/backend/pkg/validate"
)

type Service struct {
	teamRepo     domain.TeamRepository
	userRepo     domain.UserRepository
	merchantRepo domain.MerchantRepository
	enqueuer     queue.Enqueuer
	txManager    db.TransactionManager
}

func NewService(team domain.TeamRepository, user domain.UserRepository, merchant domain.MerchantRepository, enqueuer queue.Enqueuer,
	txManager db.TransactionManager) *Service {
	return &Service{
		teamRepo:     team,
		userRepo:     user,
		merchantRepo: merchant,
		enqueuer:     enqueuer,
		txManager:    txManager,
	}
}

func (s *Service) SetEnqueuer(client queue.Enqueuer) {
	s.enqueuer = client
}

type MeResult struct {
	User        domain.User
	Memberships []domain.EmployeeAuthInfo
}

func (s *Service) Me(ctx context.Context) (MeResult, error) {
	userId := jwt.MustGetUserIDFromContext(ctx)

	user, err := s.userRepo.GetUser(ctx, userId)
	if err != nil {
		return MeResult{}, err
	}

	employeeInfo, err := s.userRepo.GetEmployeesByUser(ctx, userId)
	if err != nil {
		return MeResult{}, err
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

	_, err := s.teamRepo.NewEmployee(ctx, domain.Employee{
		MerchantId:  actor.MerchantId,
		Role:        input.Role,
		FirstName:   &input.FirstName,
		LastName:    &input.LastName,
		Email:       input.Email,
		PhoneNumber: input.PhoneNumber,
		IsActive:    input.IsActive,
	})
	if err != nil {
		return err
	}

	return nil
}

func validateUpdateEmployee(employee domain.Employee, input UpdateMemberInput) error {
	if employee.Role != types.EmployeeRoleOwner && input.Role == types.EmployeeRoleOwner {
		return ErrMultipleOwnersNotAllowed
	}

	if employee.Role == types.EmployeeRoleOwner && input.Role != types.EmployeeRoleOwner {
		return ErrOwnerRoleChangeRestrictedToMerchantSettings
	}

	if employee.UserId != nil {
		if !utils.PtrEqual(employee.FirstName, input.FirstName) || !utils.PtrEqual(employee.LastName, input.LastName) ||
			!utils.PtrEqual(employee.Email, input.Email) || !utils.PtrEqual(employee.PhoneNumber, input.PhoneNumber) {

			return ErrUserEmployeeCannotBeEdited
		}
	} else {
		if input.FirstName == nil {
			return validate.NewError("invalid first name")
		}

		if input.LastName == nil {
			return validate.NewError("invalid last name")
		}
	}

	return nil
}

type UpdateMemberInput struct {
	Role        types.EmployeeRole
	FirstName   *string
	LastName    *string
	Email       *string
	PhoneNumber *string
	IsActive    bool
}

func (s *Service) UpdateMember(ctx context.Context, memberId int, input UpdateMemberInput) error {
	actor := actor.MustGetFromContext(ctx)

	employee, err := s.teamRepo.GetEmployee(ctx, actor.MerchantId, memberId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrEmployeeNotFound
		}

		return err
	}

	if err = validateUpdateEmployee(employee, input); err != nil {
		return err
	}

	err = s.teamRepo.UpdateEmployee(ctx, domain.Employee{
		Id:          memberId,
		MerchantId:  actor.MerchantId,
		Role:        input.Role,
		FirstName:   input.FirstName,
		LastName:    input.LastName,
		Email:       input.Email,
		PhoneNumber: input.PhoneNumber,
		IsActive:    input.IsActive,
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) DeleteMember(ctx context.Context, memberId int) error {
	actor := actor.MustGetFromContext(ctx)

	err := s.teamRepo.DeleteEmployee(ctx, actor.MerchantId, memberId)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) GetMember(ctx context.Context, memberId int) (domain.Employee, error) {
	actor := actor.MustGetFromContext(ctx)

	teamMember, err := s.teamRepo.GetEmployee(ctx, actor.MerchantId, memberId)
	if err != nil {
		return domain.Employee{}, err
	}

	return teamMember, nil
}

func (s *Service) GetTeam(ctx context.Context) ([]domain.Employee, error) {
	actor := actor.MustGetFromContext(ctx)

	teamMembers, err := s.teamRepo.GetEmployees(ctx, actor.MerchantId)
	if err != nil {
		return []domain.Employee{}, err
	}

	return teamMembers, nil
}
