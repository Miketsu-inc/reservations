package blockedtime

import (
	"context"
	"fmt"
	"time"

	"github.com/miketsu-inc/reservations/backend/internal/api/middleware/jwt"
	"github.com/miketsu-inc/reservations/backend/internal/domain"
)

type Service struct {
	blockedTimeRepo domain.BlockedTimeRepository
}

func NewService(blockedTime domain.BlockedTimeRepository) *Service {
	return &Service{
		blockedTimeRepo: blockedTime,
	}
}

type NewInput struct {
	Name string
	// EmployeeIds []int
	BlockedTypeId *int
	FromDate      time.Time
	ToDate        time.Time
	AllDay        bool
}

func (s *Service) New(ctx context.Context, input NewInput) error {
	employee := jwt.MustGetEmployeeFromContext(ctx)

	if !input.ToDate.After(input.FromDate) {
		return fmt.Errorf("toDate must be after fromDate")
	}

	_, err := s.blockedTimeRepo.NewBlockedTime(ctx, employee.MerchantId, []int{employee.Id}, input.Name, input.FromDate, input.ToDate, input.AllDay, input.BlockedTypeId)
	// err := s.blockedTimeRepo.NewBlockedTime(ctx, employee.MerchantId, input.EmployeeIds, input.Name, input.FromDate, input.ToDate, input.AllDay)
	if err != nil {
		return fmt.Errorf("could not make new blocked time %s", err.Error())
	}

	return nil
}

type UpdateInput struct {
	Id   int
	Name string
	// EmployeeId int
	BlockedTypeId *int
	FromDate      time.Time
	ToDate        time.Time
	AllDay        bool
}

func (s *Service) Update(ctx context.Context, blockedTimeId int, input UpdateInput) error {
	if blockedTimeId != input.Id {
		return fmt.Errorf("invalid blocked time id")
	}

	employee := jwt.MustGetEmployeeFromContext(ctx)

	if !input.ToDate.After(input.FromDate) {
		return fmt.Errorf("toDate must be after fromDate")
	}

	err := s.blockedTimeRepo.UpdateBlockedTime(ctx, domain.BlockedTime{
		Id:         blockedTimeId,
		MerchantId: employee.MerchantId,
		// EmployeeId: input.EmployeeId,
		EmployeeId:    employee.Id,
		BlockedTypeId: input.BlockedTypeId,
		Name:          input.Name,
		FromDate:      input.FromDate,
		ToDate:        input.ToDate,
		AllDay:        input.AllDay,
	})
	if err != nil {
		return fmt.Errorf("error while updating blocked time for merchant: %s", err.Error())
	}

	return nil
}

// type DeleteInput struct {
// 	EmployeeId int
// }

// func (s *Service) Delete(ctx context.Context, blockedTimeId int, input DeleteInput) error {
func (s *Service) Delete(ctx context.Context, blockedTimeId int) error {
	employee := jwt.MustGetEmployeeFromContext(ctx)

	err := s.blockedTimeRepo.DeleteBlockedTime(ctx, blockedTimeId, employee.MerchantId, employee.Id)
	// err := s.blockedTimeRepo.DeleteBlockedTime(ctx, blockedTimeId, employee.MerchantId, input.EmployeeId)
	if err != nil {
		return fmt.Errorf("error while deleting blocked time for merchant: %s", err.Error())
	}

	return nil
}
