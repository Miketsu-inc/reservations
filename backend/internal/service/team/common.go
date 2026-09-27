package team

import (
	"context"

	"github.com/google/uuid"
	"github.com/miketsu-inc/reservations/backend/internal/domain"
)

func (s *Service) IsInActiveEmployees(ctx context.Context, merchantId uuid.UUID, employeeIds []int) error {
	if len(employeeIds) == 0 {
		return domain.ErrEmployeesRequired
	}

	seen := make(map[int]struct{}, len(employeeIds))
	for _, employeeId := range employeeIds {
		if _, exists := seen[employeeId]; exists {
			return domain.ErrDuplicateEmployee
		}
		seen[employeeId] = struct{}{}
	}

	activeEmployees, err := s.teamRepo.GetActiveEmployees(ctx, merchantId)
	if err != nil {
		return err
	}

	activeIdsMap := make(map[int]struct{}, len(activeEmployees))
	for _, e := range activeEmployees {
		activeIdsMap[e.Id] = struct{}{}
	}

	for _, id := range employeeIds {
		if _, ok := activeIdsMap[id]; !ok {
			return domain.ErrEmployeeNotActive
		}
	}

	return nil
}

type EmployeeChanges struct {
	ToInsert []int
	ToDelete []int
}

func (s *Service) DetectEmployeeChanges(existing, incoming []int) (EmployeeChanges, error) {
	var ec EmployeeChanges

	existingMap := make(map[int]struct{}, len(existing))
	for _, id := range existing {
		existingMap[id] = struct{}{}
	}

	incomingMap := make(map[int]struct{}, len(incoming))
	for _, id := range incoming {
		incomingMap[id] = struct{}{}
	}

	for _, id := range existing {
		if _, ok := incomingMap[id]; !ok {
			ec.ToDelete = append(ec.ToDelete, id)
		}
	}

	for _, id := range incoming {
		if _, ok := existingMap[id]; !ok {
			ec.ToInsert = append(ec.ToInsert, id)
		}
	}

	return ec, nil
}
