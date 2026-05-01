package blockedtime

import (
	"context"
	"fmt"

	"github.com/miketsu-inc/reservations/backend/internal/api/middleware/actor"
	"github.com/miketsu-inc/reservations/backend/internal/domain"
)

type NewTypeInput struct {
	Name     string
	Duration int
	Icon     string
}

func (s *Service) NewType(ctx context.Context, input NewTypeInput) error {
	actor := actor.MustGetFromContext(ctx)

	err := s.blockedTimeRepo.NewBlockedTimeType(ctx, actor.MerchantId, domain.BlockedTimeType{
		Id:       0,
		Name:     input.Name,
		Duration: input.Duration,
		Icon:     input.Icon,
	})
	if err != nil {
		return fmt.Errorf("could not create new blocked time type: %s", err.Error())
	}

	return nil
}

type UpdateTypeInput struct {
	Id       int
	Name     string
	Duration int
	Icon     string
}

func (s *Service) UpdateType(ctx context.Context, blockedTimeTypeId int, input UpdateTypeInput) error {
	if blockedTimeTypeId != input.Id {
		return fmt.Errorf("invalid blocked time type id")
	}

	actor := actor.MustGetFromContext(ctx)

	err := s.blockedTimeRepo.UpdateBlockedTimeType(ctx, actor.MerchantId, domain.BlockedTimeType{
		Id:       input.Id,
		Name:     input.Name,
		Duration: input.Duration,
		Icon:     input.Icon,
	})
	if err != nil {
		return fmt.Errorf("error updating blocked time type: %s", err.Error())
	}

	return nil
}

func (s *Service) DeleteType(ctx context.Context, blockedTimeTypeId int) error {
	actor := actor.MustGetFromContext(ctx)

	err := s.blockedTimeRepo.DeleteBlockedTimeType(ctx, actor.MerchantId, blockedTimeTypeId)
	if err != nil {
		return fmt.Errorf("error deleting blocked time type: %s", err.Error())
	}

	return nil
}

func (s *Service) GetTypes(ctx context.Context) ([]domain.BlockedTimeType, error) {
	actor := actor.MustGetFromContext(ctx)

	blockedTimetypes, err := s.blockedTimeRepo.GetAllBlockedTimeTypes(ctx, actor.MerchantId)
	if err != nil {
		return []domain.BlockedTimeType{}, fmt.Errorf("could not fetch blocked time types: %s", err.Error())
	}

	return blockedTimetypes, nil
}
