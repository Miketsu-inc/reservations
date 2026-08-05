package team

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

func (s *Service) IsInActiveEmployees(ctx context.Context, merchantId uuid.UUID, employeeIds []int) error {
	if len(employeeIds) == 0 {
		return nil
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
			return fmt.Errorf("active employee with this id  does not exist")
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
