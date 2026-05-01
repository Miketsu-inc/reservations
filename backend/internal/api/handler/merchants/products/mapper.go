package products

import (
	"github.com/miketsu-inc/reservations/backend/internal/domain"
	productServ "github.com/miketsu-inc/reservations/backend/internal/service/product"
)

func mapToNewInput(in newReq) productServ.NewInput {
	return productServ.NewInput{
		Name:          in.Name,
		Description:   in.Description,
		Price:         in.Price,
		Unit:          in.Unit,
		MaxAmount:     in.MaxAmount,
		CurrentAmount: in.CurrentAmount,
	}
}

func mapToUpdateInput(in updateReq) productServ.UpdateInput {
	return productServ.UpdateInput{
		Id:            in.Id,
		Name:          in.Name,
		Description:   in.Description,
		Price:         in.Price,
		Unit:          in.Unit,
		MaxAmount:     in.MaxAmount,
		CurrentAmount: in.CurrentAmount,
	}
}

func mapToGetAllResp(in []domain.ProductInfo) []getAllResp {
	out := make([]getAllResp, len(in))

	for i, product := range in {
		s := make([]servicesForProdcutResp, len(product.Services))

		for j, service := range product.Services {
			s[j] = servicesForProdcutResp{
				Id:    service.Id,
				Name:  service.Name,
				Color: service.Color,
			}
		}

		out[i] = getAllResp{
			Id:            product.Id,
			Name:          product.Name,
			Description:   product.Description,
			Price:         product.Price,
			Unit:          product.Unit,
			MaxAmount:     product.MaxAmount,
			CurrentAmount: product.CurrentAmount,
			Services:      s,
		}
	}

	return out
}
