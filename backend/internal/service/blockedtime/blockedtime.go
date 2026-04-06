package blockedtime

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/miketsu-inc/reservations/backend/internal/api/middleware/jwt"
	"github.com/miketsu-inc/reservations/backend/internal/domain"
	"github.com/miketsu-inc/reservations/backend/internal/jobs/args"
	"github.com/miketsu-inc/reservations/backend/pkg/db"
	"github.com/miketsu-inc/reservations/backend/pkg/queue"
)

type Service struct {
	blockedTimeRepo domain.BlockedTimeRepository
	enqueuer        queue.Enqueuer
	txManager       db.TransactionManager
}

func NewService(blockedTime domain.BlockedTimeRepository, enqueuer queue.Enqueuer, txManager db.TransactionManager) *Service {
	return &Service{
		blockedTimeRepo: blockedTime,
		enqueuer:        enqueuer,
		txManager:       txManager,
	}
}

func (s *Service) SetEnqueuer(client queue.Enqueuer) {
	s.enqueuer = client
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

	employeeIds := []int{employee.Id}

	return s.txManager.WithTransaction(ctx, func(tx pgx.Tx) error {
		ids, err := s.blockedTimeRepo.WithTx(tx).NewBlockedTime(ctx, employee.MerchantId, employeeIds, input.Name, input.FromDate, input.ToDate, input.AllDay, input.BlockedTypeId)
		// err := s.blockedTimeRepo.WithTx(tx).NewBlockedTime(ctx, employee.MerchantId, input.EmployeeIds, input.Name, input.FromDate, input.ToDate, input.AllDay)
		if err != nil {
			return fmt.Errorf("could not make new blocked time %s", err.Error())
		}

		if len(employeeIds) != 0 {
			_, err = s.enqueuer.InsertTx(ctx, tx, args.SyncNewBlockedTime{
				BlockedTimeId: ids[0],
			}, nil)
			if err != nil {
				return err
			}
		}

		return nil
	})
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

	return s.txManager.WithTransaction(ctx, func(tx pgx.Tx) error {
		err := s.blockedTimeRepo.WithTx(tx).UpdateBlockedTime(ctx, domain.BlockedTime{
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

		// TODO: only update if time was changed
		_, err = s.enqueuer.InsertTx(ctx, tx, args.SyncUpdateBlockedTime{
			BlockedTimeId: blockedTimeId,
		}, nil)
		if err != nil {
			return err
		}

		return nil
	})
}

// type DeleteInput struct {
// 	EmployeeId int
// }

// func (s *Service) Delete(ctx context.Context, blockedTimeId int, input DeleteInput) error {
func (s *Service) Delete(ctx context.Context, blockedTimeId int) error {
	employee := jwt.MustGetEmployeeFromContext(ctx)

	return s.txManager.WithTransaction(ctx, func(tx pgx.Tx) error {
		err := s.blockedTimeRepo.WithTx(tx).DeleteBlockedTime(ctx, blockedTimeId, employee.MerchantId, employee.Id)
		// err := s.blockedTimeRepo.WithTx(tx).DeleteBlockedTime(ctx, blockedTimeId, employee.MerchantId, input.EmployeeId)
		if err != nil {
			return fmt.Errorf("error while deleting blocked time for merchant: %s", err.Error())
		}

		_, err = s.enqueuer.InsertTx(ctx, tx, args.SyncDeleteBlockedTime{
			BlockedTimeId: blockedTimeId,
		}, nil)
		if err != nil {
			return err
		}

		return nil
	})
}
