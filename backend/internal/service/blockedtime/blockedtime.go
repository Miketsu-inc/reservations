package blockedtime

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/miketsu-inc/reservations/backend/internal/api/middleware/actor"
	"github.com/miketsu-inc/reservations/backend/internal/domain"
	"github.com/miketsu-inc/reservations/backend/internal/jobs/args"
	"github.com/miketsu-inc/reservations/backend/internal/service/team"
	"github.com/miketsu-inc/reservations/backend/internal/utils"
	"github.com/miketsu-inc/reservations/backend/pkg/db"
	"github.com/miketsu-inc/reservations/backend/pkg/queue"
)

type Service struct {
	blockedTimeRepo domain.BlockedTimeRepository
	teamRepo        domain.TeamRepository
	teamService     *team.Service
	enqueuer        queue.Enqueuer
	txManager       db.TransactionManager
}

func NewService(blockedTime domain.BlockedTimeRepository, teamRepo domain.TeamRepository,
	teamService *team.Service, enqueuer queue.Enqueuer, txManager db.TransactionManager) *Service {
	return &Service{
		blockedTimeRepo: blockedTime,
		teamRepo:        teamRepo,
		teamService:     teamService,
		enqueuer:        enqueuer,
		txManager:       txManager,
	}
}

func (s *Service) SetEnqueuer(client queue.Enqueuer) {
	s.enqueuer = client
}

type NewInput struct {
	Name          string
	EmployeeIds   []int
	BlockedTypeId *int
	FromDate      *time.Time
	ToDate        *time.Time
	BlockedDay    *time.Time
	IsAllDay      bool
}

func validateDateShape(isAllDay bool, blockedDay, fromDate, toDate *time.Time) error {
	if isAllDay {
		if blockedDay == nil || fromDate != nil || toDate != nil {
			return fmt.Errorf("all-day blocked time requires date only")
		}

		return nil
	}

	if blockedDay != nil || fromDate == nil || toDate == nil {
		return fmt.Errorf("timed blocked time requires fromDate and toDate only")
	}

	if !toDate.After(*fromDate) {
		return fmt.Errorf("toDate must be after fromDate")
	}

	return nil
}

func optionalTimesEqual(a, b *time.Time) bool {
	if a == nil || b == nil {
		return a == b
	}

	return a.Equal(*b)
}

func (s *Service) New(ctx context.Context, input NewInput) error {
	actor := actor.MustGetFromContext(ctx)

	if err := validateDateShape(input.IsAllDay, input.BlockedDay, input.FromDate, input.ToDate); err != nil {
		return err
	}

	return s.txManager.WithTransaction(ctx, func(tx pgx.Tx) error {
		ids, err := s.blockedTimeRepo.WithTx(tx).BulkInsertBlockedTime(ctx, []domain.BlockedTime{{
			MerchantId:    actor.MerchantId,
			BlockedTypeId: input.BlockedTypeId,
			Name:          input.Name,
			FromDate:      input.FromDate,
			ToDate:        input.ToDate,
			BlockedDay:    input.BlockedDay,
			IsAllDay:      input.IsAllDay,
		}})
		if err != nil {
			return err
		}

		if len(input.EmployeeIds) > 0 {
			err = s.teamService.IsInActiveEmployees(ctx, actor.MerchantId, input.EmployeeIds)
			if err != nil {
				return err
			}

			err = s.blockedTimeRepo.WithTx(tx).BulkInsertEmployeeBlockedTime(ctx, utils.RepeatEach(ids, len(input.EmployeeIds)), input.EmployeeIds)
			if err != nil {
				return err
			}
		}

		_, err = s.enqueuer.InsertTx(ctx, tx, args.SyncNewBlockedTimeDispatcher{
			BlockedTimeId: ids[0],
		}, nil)
		if err != nil {
			return err
		}

		return nil
	})
}

type UpdateInput struct {
	BlockedTimeId int
	Name          string
	BlockedTypeId *int
	FromDate      *time.Time
	ToDate        *time.Time
	BlockedDay    *time.Time
	IsAllDay      bool
	EmployeeIds   []int
}

func (s *Service) Update(ctx context.Context, input UpdateInput) error {
	actor := actor.MustGetFromContext(ctx)

	blockedTime, err := s.blockedTimeRepo.GetBlockedTimeEmployees(ctx, input.BlockedTimeId)
	if err != nil {
		return err
	}

	if blockedTime.MerchantId != actor.MerchantId {
		return fmt.Errorf("blocked time with id %d not found for merchant", blockedTime.Id)
	}

	if err := validateDateShape(input.IsAllDay, input.BlockedDay, input.FromDate, input.ToDate); err != nil {
		return err
	}

	return s.txManager.WithTransaction(ctx, func(tx pgx.Tx) error {
		err := s.blockedTimeRepo.WithTx(tx).UpdateBlockedTime(ctx, domain.BlockedTime{
			Id:            input.BlockedTimeId,
			MerchantId:    actor.MerchantId,
			BlockedTypeId: input.BlockedTypeId,
			Name:          input.Name,
			FromDate:      input.FromDate,
			ToDate:        input.ToDate,
			BlockedDay:    input.BlockedDay,
			IsAllDay:      input.IsAllDay,
		})
		if err != nil {
			return err
		}

		employeeChanges, err := s.teamService.DetectEmployeeChanges(blockedTime.EmployeeIds, input.EmployeeIds)
		if err != nil {
			return err
		}

		if len(employeeChanges.ToDelete) > 0 {
			err := s.blockedTimeRepo.WithTx(tx).BulkDeleteEmployeeBlockedTime(ctx, []int{blockedTime.Id}, employeeChanges.ToDelete)
			if err != nil {
				return err
			}
		}

		if len(employeeChanges.ToInsert) > 0 {
			err = s.teamService.IsInActiveEmployees(ctx, actor.MerchantId, employeeChanges.ToInsert)
			if err != nil {
				return err
			}

			btIds := utils.RepeatSlice([]int{input.BlockedTimeId}, len(employeeChanges.ToInsert))

			err = s.blockedTimeRepo.WithTx(tx).BulkInsertEmployeeBlockedTime(ctx, btIds, employeeChanges.ToInsert)
			if err != nil {
				return err
			}
		}

		if blockedTime.IsAllDay != input.IsAllDay ||
			!optionalTimesEqual(blockedTime.BlockedDay, input.BlockedDay) ||
			!optionalTimesEqual(blockedTime.FromDate, input.FromDate) ||
			!optionalTimesEqual(blockedTime.ToDate, input.ToDate) {
			_, err = s.enqueuer.InsertTx(ctx, tx, args.SyncUpdateBlockedTimeDispatcher{
				BlockedTimeId: input.BlockedTimeId,
			}, nil)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

func (s *Service) Delete(ctx context.Context, blockedTimeId int) error {
	actor := actor.MustGetFromContext(ctx)

	// TODO: if the actor is not on the block time this will give an error
	blockedTime, err := s.blockedTimeRepo.GetBlockedTimeForEmployee(ctx, blockedTimeId, actor.EmployeeId)
	if err != nil {
		return err
	}

	if blockedTime.MerchantId != actor.MerchantId {
		return fmt.Errorf("blocked time with id %d not found for merchant", blockedTime.Id)
	}

	return s.txManager.WithTransaction(ctx, func(tx pgx.Tx) error {
		err := s.blockedTimeRepo.WithTx(tx).BulkDeleteBlockedTime(ctx, []int{blockedTime.Id})
		if err != nil {
			return err
		}

		_, err = s.enqueuer.InsertTx(ctx, tx, args.SyncDeleteBlockedTimeDispatcher{
			BlockedTimeId: blockedTime.Id,
		}, nil)
		if err != nil {
			return err
		}

		return nil
	})
}
