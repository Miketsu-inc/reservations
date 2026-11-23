package team

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/miketsu-inc/reservations/backend/internal/api/middleware/actor"
	"github.com/miketsu-inc/reservations/backend/internal/api/middleware/jwt"
	"github.com/miketsu-inc/reservations/backend/internal/domain"
	"github.com/miketsu-inc/reservations/backend/internal/permission"
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
	reviewRepo   domain.ReviewRepository
	enqueuer     queue.Enqueuer
	txManager    db.TransactionManager
}

func NewService(team domain.TeamRepository, user domain.UserRepository, merchant domain.MerchantRepository, review domain.ReviewRepository,
	enqueuer queue.Enqueuer, txManager db.TransactionManager) *Service {
	return &Service{
		teamRepo:     team,
		userRepo:     user,
		merchantRepo: merchant,
		reviewRepo:   review,
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

	if !permission.For(actor).Manage(input.Role) {
		return ErrRoleAssignForbidden
	}

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

	if employee.Role == types.EmployeeRoleOwner && !input.IsActive {
		return ErrOwnerCannotBeDeactivated
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

	can := permission.For(actor)

	if !can.Edit(permission.Team, employee.Id) && !can.Manage(employee.Role) {
		return ErrEmployeeManageForbidden
	}

	if err = validateUpdateEmployee(employee, input); err != nil {
		return err
	}

	if input.Role != employee.Role && !can.Manage(input.Role) {
		return ErrRoleAssignForbidden
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

	employee, err := s.teamRepo.GetEmployee(ctx, actor.MerchantId, memberId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrEmployeeNotFound
		}

		return err
	}

	if !permission.For(actor).Manage(employee.Role) {
		return ErrEmployeeManageForbidden
	}

	err = s.teamRepo.DeleteEmployee(ctx, actor.MerchantId, memberId)
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

func (s *Service) GetActiveMembers(ctx context.Context) ([]domain.Employee, error) {
	actor := actor.MustGetFromContext(ctx)

	if !permission.For(actor).ViewAll(permission.Team) {
		member, err := s.teamRepo.GetEmployee(ctx, actor.MerchantId, actor.EmployeeId)
		if err != nil {
			return []domain.Employee{}, err
		}

		return []domain.Employee{member}, nil
	}

	teamMembers, err := s.teamRepo.GetActiveEmployees(ctx, actor.MerchantId)
	if err != nil {
		return []domain.Employee{}, err
	}

	return teamMembers, nil
}

type RatedEmployee struct {
	domain.Employee
	Rating domain.EmployeeRating
}

func (s *Service) GetTeamByMerchantName(ctx context.Context, merchantName string, locationId int) ([]RatedEmployee, error) {
	merchantId, err := s.merchantRepo.GetMerchantIdByUrlName(ctx, strings.ToLower(merchantName))
	if err != nil {
		return []RatedEmployee{}, err
	}

	employees, err := s.teamRepo.GetActiveEmployees(ctx, merchantId)
	if err != nil {
		return []RatedEmployee{}, err
	}

	ratings, err := s.reviewRepo.GetEmployeeRatings(ctx, merchantId, &locationId)
	if err != nil {
		return []RatedEmployee{}, err
	}

	team := make([]RatedEmployee, len(employees))
	for i, emp := range employees {
		team[i] = RatedEmployee{Employee: emp, Rating: ratings[emp.Id]}
	}

	return team, nil
}

type RatedServiceEmployee struct {
	domain.EmployeeWithServiceOverrides
	Rating domain.EmployeeRating
}

func (s *Service) GetServiceTeamByMerchantName(ctx context.Context, merchantName string, locationId int, serviceId int) ([]RatedServiceEmployee, error) {
	merchantId, err := s.merchantRepo.GetMerchantIdByUrlName(ctx, strings.ToLower(merchantName))
	if err != nil {
		return []RatedServiceEmployee{}, err
	}

	employees, err := s.teamRepo.GetActiveEmployeesForService(ctx, merchantId, serviceId)
	if err != nil {
		return []RatedServiceEmployee{}, err
	}

	ratings, err := s.reviewRepo.GetEmployeeRatings(ctx, merchantId, &locationId)
	if err != nil {
		return []RatedServiceEmployee{}, err
	}

	team := make([]RatedServiceEmployee, len(employees))
	for i, emp := range employees {
		team[i] = RatedServiceEmployee{EmployeeWithServiceOverrides: emp, Rating: ratings[emp.Id]}
	}

	return team, nil
}
