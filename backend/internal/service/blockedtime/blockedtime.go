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
	FromDate      time.Time
	ToDate        time.Time
	AllDay        bool
}

func (s *Service) New(ctx context.Context, input NewInput) error {
	actor := actor.MustGetFromContext(ctx)

	if !input.ToDate.After(input.FromDate) {
		return fmt.Errorf("toDate must be after fromDate")
	}

	return s.txManager.WithTransaction(ctx, func(tx pgx.Tx) error {
		ids, err := s.blockedTimeRepo.WithTx(tx).BulkInsertBlockedTime(ctx, []domain.BlockedTime{{
			MerchantId:    actor.MerchantId,
			BlockedTypeId: input.BlockedTypeId,
			Name:          input.Name,
			FromDate:      input.FromDate,
			ToDate:        input.ToDate,
			AllDay:        input.AllDay,
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

			_, err = s.enqueuer.InsertTx(ctx, tx, args.SyncNewBlockedTimeDispatcher{
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
	BlockedTimeId int
	Name          string
	BlockedTypeId *int
	FromDate      time.Time
	ToDate        time.Time
	AllDay        bool
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

	if !input.ToDate.After(input.FromDate) {
		return fmt.Errorf("toDate must be after fromDate")
	}

	return s.txManager.WithTransaction(ctx, func(tx pgx.Tx) error {
		err := s.blockedTimeRepo.WithTx(tx).UpdateBlockedTime(ctx, domain.BlockedTime{
			Id:            input.BlockedTimeId,
			MerchantId:    actor.MerchantId,
			BlockedTypeId: input.BlockedTypeId,
			Name:          input.Name,
			FromDate:      input.FromDate,
			ToDate:        input.ToDate,
			AllDay:        input.AllDay,
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
			err = s.teamService.IsInActiveEmployees(ctx, actor.MerchantId, input.EmployeeIds)
			if err != nil {
				return err
			}

			btIds := utils.RepeatSlice([]int{input.BlockedTimeId}, len(employeeChanges.ToInsert))

			err = s.blockedTimeRepo.WithTx(tx).BulkInsertEmployeeBlockedTime(ctx, btIds, employeeChanges.ToInsert)
			if err != nil {
				return err
			}
		}

		if !blockedTime.FromDate.Equal(input.FromDate) || blockedTime.ToDate.Equal(input.ToDate) {
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
