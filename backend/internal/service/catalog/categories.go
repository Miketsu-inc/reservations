package catalog

import (
	"context"
	"fmt"

	"github.com/miketsu-inc/reservations/backend/internal/api/middleware/jwt"
	"github.com/miketsu-inc/reservations/backend/internal/domain"
)

type NewCategoryInput struct {
	Name string
}

func (s *Service) NewCategory(ctx context.Context, req NewCategoryInput) error {
	employee := jwt.MustGetEmployeeFromContext(ctx)

	err := s.catalogRepo.NewServiceCategory(ctx, employee.MerchantId, domain.ServiceCategory{
		Name:     req.Name,
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

func (s *Service) UpdateCategory(ctx context.Context, categoryId int, req UpdateCategoryInput) error {
	employee := jwt.MustGetEmployeeFromContext(ctx)

	err := s.catalogRepo.UpdateServiceCategory(ctx, employee.MerchantId, domain.ServiceCategory{
		Id:   categoryId,
		Name: req.Name,
	})
	if err != nil {
		return fmt.Errorf("error while updating service category: %s", err.Error())
	}

	return nil
}

func (s *Service) DeleteCategory(ctx context.Context, categoryId int) error {
	employee := jwt.MustGetEmployeeFromContext(ctx)

	err := s.catalogRepo.DeleteServiceCategory(ctx, employee.MerchantId, categoryId)
	if err != nil {
		return fmt.Errorf("error while deleting service category: %s", err.Error())
	}

	return nil
}

type ReorderCategoriesInput struct {
	Categories []int
}

func (s *Service) ReorderCategories(ctx context.Context, req ReorderCategoriesInput) error {
	employee := jwt.MustGetEmployeeFromContext(ctx)

	err := s.catalogRepo.ReorderServiceCategories(ctx, employee.MerchantId, req.Categories)
	if err != nil {
		return fmt.Errorf("error while ordering service categories: %s", err.Error())
	}

	return nil
}
