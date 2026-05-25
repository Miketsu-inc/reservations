package catalog

import (
	"context"
	"fmt"

	"github.com/miketsu-inc/reservations/backend/internal/api/middleware/actor"
	"github.com/miketsu-inc/reservations/backend/internal/domain"
)

type NewCategoryInput struct {
	Name string
}

func (s *Service) NewCategory(ctx context.Context, input NewCategoryInput) error {
	actor := actor.MustGetFromContext(ctx)

	err := s.catalogRepo.NewServiceCategory(ctx, actor.MerchantId, domain.ServiceCategory{
		Name:     input.Name,
		Sequence: 0,
	})
	if err != nil {
		return fmt.Errorf("error while creating new service category %s", err.Error())
	}

	return nil
}

type UpdateCategoryInput struct {
	Name string
}

func (s *Service) UpdateCategory(ctx context.Context, categoryId int, input UpdateCategoryInput) error {
	actor := actor.MustGetFromContext(ctx)

	err := s.catalogRepo.UpdateServiceCategory(ctx, actor.MerchantId, domain.ServiceCategory{
		Id:   categoryId,
		Name: input.Name,
	})
	if err != nil {
		return fmt.Errorf("error while updating service category: %s", err.Error())
	}

	return nil
}

func (s *Service) DeleteCategory(ctx context.Context, categoryId int) error {
	actor := actor.MustGetFromContext(ctx)

	err := s.catalogRepo.DeleteServiceCategory(ctx, actor.MerchantId, categoryId)
	if err != nil {
		return fmt.Errorf("error while deleting service category: %s", err.Error())
	}

	return nil
}

type ReorderCategoriesInput struct {
	Categories []int
}

func (s *Service) ReorderCategories(ctx context.Context, input ReorderCategoriesInput) error {
	actor := actor.MustGetFromContext(ctx)

	idSet := make(map[int]struct{}, len(input.Categories))
	for _, id := range input.Categories {
		if _, ok := idSet[id]; ok {
			return fmt.Errorf("duplicate category id: %d", id)
		}

		idSet[id] = struct{}{}
	}

	err := s.catalogRepo.ReorderServiceCategories(ctx, actor.MerchantId, input.Categories)
	if err != nil {
		return fmt.Errorf("error while ordering service categories: %s", err.Error())
	}

	return nil
}
