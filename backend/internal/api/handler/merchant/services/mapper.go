package services

import (
	"github.com/miketsu-inc/reservations/backend/internal/domain"
	catalogServ "github.com/miketsu-inc/reservations/backend/internal/service/catalog"
)

func mapToNewInput(in newReq) catalogServ.NewInput {
	phases := make([]catalogServ.NewPhasesInput, len(in.Phases))

	for i, p := range in.Phases {
		phases[i] = catalogServ.NewPhasesInput{
			Name:      p.Name,
			Sequence:  p.Sequence,
			Duration:  p.Duration,
			PhaseType: p.PhaseType,
		}
	}

	connProducts := make([]catalogServ.ConnectedProductsInput, len(in.UsedProducts))

	for i, p := range in.UsedProducts {
		connProducts[i] = catalogServ.ConnectedProductsInput{
			ProductId:  p.ProductId,
			AmountUsed: p.AmountUsed,
		}
	}

	return catalogServ.NewInput{
		Name:        in.Name,
		Description: in.Description,
		Color:       in.Color,
		Price:       in.Price,
		Cost:        in.Cost,
		PriceType:   in.PriceType,
		CategoryId:  in.CategoryId,
		IsActive:    in.IsActive,
		Settings: catalogServ.ServiceSettingsInput{
			CancelDeadline:   in.Settings.CancelDeadline,
			BookingWindowMin: in.Settings.BookingWindowMin,
			BookingWindowMax: in.Settings.BookingWindowMax,
			BufferTime:       in.Settings.BufferTime,
		},
		Phases:       phases,
		UsedProducts: connProducts,
	}
}

func mapToUpdateInput(in updateReq) catalogServ.UpdateInput {
	phases := make([]catalogServ.PhasesInput, len(in.Phases))

	for i, p := range in.Phases {
		phases[i] = catalogServ.PhasesInput{
			Id:        p.Id,
			ServiceId: p.ServiceId,
			Name:      p.Name,
			Sequence:  p.Sequence,
			Duration:  p.Duration,
			PhaseType: p.PhaseType,
		}
	}

	return catalogServ.UpdateInput{
		Id:          in.Id,
		Name:        in.Name,
		Description: in.Description,
		Color:       in.Color,
		Price:       in.Price,
		Cost:        in.Cost,
		PriceType:   in.PriceType,
		CategoryId:  in.CategoryId,
		IsActive:    in.IsActive,
		Settings: catalogServ.ServiceSettingsInput{
			CancelDeadline:   in.Settings.CancelDeadline,
			BookingWindowMin: in.Settings.BookingWindowMin,
			BookingWindowMax: in.Settings.BookingWindowMax,
			BufferTime:       in.Settings.BufferTime,
		},
		Phases: phases,
	}
}

func mapToGetResp(in domain.ServicePageData) getResp {
	phases := make([]phaseReq, len(in.Phases))

	for i, p := range in.Phases {
		phases[i] = phaseReq{
			Id:        p.Id,
			ServiceId: p.ServiceId,
			Name:      p.Name,
			Sequence:  p.Sequence,
			Duration:  p.Duration,
			PhaseType: p.PhaseType,
		}
	}

	products := make([]productResp, len(in.Products))

	for i, p := range in.Products {
		products[i] = productResp{
			Id:         p.Id,
			Name:       p.Name,
			Unit:       p.Unit,
			AmountUsed: p.AmountUsed,
		}
	}

	return getResp{
		Id:            in.Id,
		CategoryId:    in.CategoryId,
		Name:          in.Name,
		Description:   in.Description,
		Color:         in.Color,
		TotalDuration: in.TotalDuration,
		Price:         in.Price,
		Cost:          in.Cost,
		PriceType:     in.PriceType,
		IsActive:      in.IsActive,
		Sequence:      in.Sequence,
		Settings: serviceSettingsReq{
			CancelDeadline:   in.Settings.CancelDeadline,
			BookingWindowMin: in.Settings.BookingWindowMin,
			BookingWindowMax: in.Settings.BookingWindowMax,
			BufferTime:       in.Settings.BufferTime,
		},
		Phases:       phases,
		UsedProducts: products,
	}
}

func mapToUpdateServiceProductInput(in updateServiceProductReq) catalogServ.UpdateServiceProductInput {
	products := make([]catalogServ.ConnectedProductsInput, len(in.UsedProducts))

	for i, p := range in.UsedProducts {
		products[i] = catalogServ.ConnectedProductsInput{
			ProductId:  p.ProductId,
			AmountUsed: p.AmountUsed,
		}
	}

	return catalogServ.UpdateServiceProductInput{
		ServiceId:    in.ServiceId,
		UsedProducts: products,
	}
}

func mapToGetAllResp(in []domain.ServicesGroupedByCategory) []getAllResp {
	categories := make([]getAllResp, len(in))

	for i, c := range in {
		services := make([]serviceResp, len(c.Services))

		for j, s := range c.Services {
			phases := make([]phaseReq, len(s.Phases))

			for k, p := range s.Phases {
				phases[k] = phaseReq{
					Id:        p.Id,
					ServiceId: p.ServiceId,
					Name:      p.Name,
					Sequence:  p.Sequence,
					Duration:  p.Duration,
					PhaseType: p.PhaseType,
				}
			}

			services[j] = serviceResp{
				Id:              s.Id,
				MerchantId:      s.MerchantId,
				BookingType:     s.BookingType,
				CategoryId:      s.CategoryId,
				Name:            s.Name,
				Description:     s.Description,
				Color:           s.Color,
				TotalDuration:   s.TotalDuration,
				Price:           s.Price,
				Cost:            s.Cost,
				PriceType:       s.PriceType,
				IsActive:        s.IsActive,
				MinParticipants: s.MinParticipants,
				MaxParticipants: s.MaxParticipants,
				Sequence:        s.Sequence,
				Phases:          phases,
			}
		}

		categories[i] = getAllResp{
			Id:       c.Id,
			Name:     c.Name,
			Sequence: c.Sequence,
			Services: services,
		}
	}

	return categories
}

func mapToReorderInput(in reorderReq) catalogServ.ReorderInput {
	return catalogServ.ReorderInput{
		CategoryId: in.CategoryId,
		Services:   in.Services,
	}
}

func mapToGetFormOptionsResp(in domain.ServicePageFormOptions) getFormOptionsResp {
	products := make([]minimalProductResp, len(in.Products))

	for i, p := range in.Products {
		products[i] = minimalProductResp{
			Id:   p.Id,
			Name: p.Name,
			Unit: p.Unit,
		}
	}

	categories := make([]serviceCategoryResp, len(in.Categories))

	for i, c := range in.Categories {
		categories[i] = serviceCategoryResp{
			Id:         c.Id,
			MerchantId: c.MerchantId,
			LocationId: c.LocationId,
			Name:       c.Name,
			Sequence:   c.Sequence,
		}
	}

	return getFormOptionsResp{
		Products:   products,
		Categories: categories,
	}
}

func mapToNewGroupInput(in newGroupReq) catalogServ.NewGroupInput {
	products := make([]catalogServ.ConnectedProductsInput, len(in.UsedProducts))

	for i, p := range in.UsedProducts {
		products[i] = catalogServ.ConnectedProductsInput{
			ProductId:  p.ProductId,
			AmountUsed: p.AmountUsed,
		}
	}

	return catalogServ.NewGroupInput{
		Name:            in.Name,
		Description:     in.Description,
		Color:           in.Color,
		Price:           in.Price,
		Cost:            in.Cost,
		PriceType:       in.PriceType,
		Duration:        in.Duration,
		CategoryId:      in.CategoryId,
		MinParticipants: in.MinParticipants,
		MaxParticipants: in.MaxParticipants,
		IsActive:        in.IsActive,
		Settings: catalogServ.ServiceSettingsInput{
			CancelDeadline:   in.Settings.CancelDeadline,
			BookingWindowMin: in.Settings.BookingWindowMin,
			BookingWindowMax: in.Settings.BookingWindowMax,
			BufferTime:       in.Settings.BufferTime,
		},
		UsedProducts: products,
	}
}

func mapToUpdateGroupInput(in updateGroupReq) catalogServ.UpdateGroupInput {
	return catalogServ.UpdateGroupInput{
		Id:              in.Id,
		Name:            in.Name,
		Description:     in.Description,
		Color:           in.Color,
		Price:           in.Price,
		Cost:            in.Cost,
		PriceType:       in.PriceType,
		Duration:        in.Duration,
		CategoryId:      in.CategoryId,
		MinParticipants: in.MinParticipants,
		MaxParticipants: in.MaxParticipants,
		IsActive:        in.IsActive,
		Settings: catalogServ.ServiceSettingsInput{
			CancelDeadline:   in.Settings.CancelDeadline,
			BookingWindowMin: in.Settings.BookingWindowMin,
			BookingWindowMax: in.Settings.BookingWindowMax,
			BufferTime:       in.Settings.BufferTime,
		},
	}
}

func mapToGetGroupResp(in domain.GroupServicePageData) getGroupResp {
	products := make([]minimalProductResp, len(in.Products))

	for i, p := range in.Products {
		products[i] = minimalProductResp{
			Id:   p.Id,
			Name: p.Name,
			Unit: p.Unit,
		}
	}

	return getGroupResp{
		Id:              in.Id,
		CategoryId:      in.CategoryId,
		Name:            in.Name,
		Description:     in.Description,
		Color:           in.Color,
		Duration:        in.Duration,
		Price:           in.Price,
		Cost:            in.Cost,
		PriceType:       in.PriceType,
		IsActive:        in.IsActive,
		Sequence:        in.Sequence,
		MinParicipants:  in.MinParicipants,
		MaxParticipants: in.MaxParticipants,
		Settings: serviceSettingsReq{
			CancelDeadline:   in.Settings.CancelDeadline,
			BookingWindowMin: in.Settings.BookingWindowMin,
			BookingWindowMax: in.Settings.BookingWindowMax,
			BufferTime:       in.Settings.BufferTime,
		},
		Products: products,
	}
}
