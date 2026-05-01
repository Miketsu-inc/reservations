package product

import (
	"context"
	"fmt"

	"github.com/miketsu-inc/reservations/backend/internal/api/middleware/actor"
	"github.com/miketsu-inc/reservations/backend/internal/domain"
	"github.com/miketsu-inc/reservations/backend/pkg/currencyx"
)

type Service struct {
	productRepo  domain.ProductRepository
	merchantRepo domain.MerchantRepository
}

func NewService(product domain.ProductRepository, merchant domain.MerchantRepository) *Service {
	return &Service{
		productRepo:  product,
		merchantRepo: merchant,
	}
}

type NewInput struct {
	Name          string
	Description   string
	Price         *currencyx.Price
	Unit          string
	MaxAmount     int
	CurrentAmount int
}

func (s *Service) New(ctx context.Context, input NewInput) error {
	actor := actor.MustGetFromContext(ctx)

	curr, err := s.merchantRepo.GetMerchantCurrency(ctx, actor.MerchantId)
	if err != nil {
		return fmt.Errorf("error while getting merchant's currency: %s", err.Error())
	}

	if input.Price != nil {
		if input.Price.CurrencyCode() != curr {
			return fmt.Errorf("new product price's currency does not match merchant's currency")
		}
	}

	if err := s.productRepo.NewProduct(ctx, domain.Product{
		Id:            0,
		MerchantId:    actor.MerchantId,
		Name:          input.Name,
		Description:   input.Description,
		Price:         input.Price,
		Unit:          input.Unit,
		MaxAmount:     input.MaxAmount,
		CurrentAmount: input.CurrentAmount,
	}); err != nil {
		return fmt.Errorf("unexpected error inserting product for merchant: %s", err.Error())
	}

	return nil
}

type UpdateInput struct {
	Id            int
	Name          string
	Description   string
	Price         *currencyx.Price
	Unit          string
	MaxAmount     int
	CurrentAmount int
}

func (s *Service) Update(ctx context.Context, productId int, input UpdateInput) error {
	if productId != input.Id {
		return fmt.Errorf("invalid product id")
	}

	actor := actor.MustGetFromContext(ctx)

	err := s.productRepo.UpdateProduct(ctx, domain.Product{
		Id:            input.Id,
		MerchantId:    actor.MerchantId,
		Name:          input.Name,
		Description:   input.Description,
		Price:         input.Price,
		Unit:          input.Unit,
		MaxAmount:     input.MaxAmount,
		CurrentAmount: input.CurrentAmount,
	})
	if err != nil {
		return fmt.Errorf("error while updating product for merchant: %s", err.Error())
	}

	return nil
}

func (s *Service) Delete(ctx context.Context, productId int) error {
	actor := actor.MustGetFromContext(ctx)

	err := s.productRepo.DeleteProduct(ctx, actor.MerchantId, productId)
	if err != nil {
		return fmt.Errorf("error while deleting product for merchant: %s", err.Error())
	}

	return nil
}

func (s *Service) GetAll(ctx context.Context) ([]domain.ProductInfo, error) {
	actor := actor.MustGetFromContext(ctx)

	products, err := s.productRepo.GetProducts(ctx, actor.MerchantId)
	if err != nil {
		return []domain.ProductInfo{}, fmt.Errorf("error while retrieving products for merchant: %s", err.Error())
	}

	return products, nil
}
