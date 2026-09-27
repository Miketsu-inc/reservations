package services

import (
	"github.com/miketsu-inc/reservations/backend/internal/domain"
	catalogServ "github.com/miketsu-inc/reservations/backend/internal/service/catalog"
	"github.com/miketsu-inc/reservations/backend/pkg/currencyx"
)

func mapToPhaseReqs(in []domain.ServicePhase) []phaseReq {
	phases := make([]phaseReq, len(in))
	for i, phase := range in {
		phases[i] = phaseReq{
			Id:        phase.Id,
			Name:      phase.Name,
			Sequence:  phase.Sequence,
			Duration:  phase.Duration,
			PhaseType: phase.PhaseType,
		}
	}
	return phases
}

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
		BookingType:     in.BookingType,
		Name:            in.Name,
		Description:     in.Description,
		Color:           in.Color,
		Price:           in.Price,
		PriceType:       in.PriceType,
		CategoryId:      in.CategoryId,
		MinParticipants: in.MinParticipants,
		MaxParticipants: in.MaxParticipants,
		IsActive:        in.IsActive,
		Settings: catalogServ.ServiceSettingsInput{
			CancelDeadline:   in.Settings.CancelDeadline,
			BookingWindowMin: in.Settings.BookingWindowMin,
			BookingWindowMax: in.Settings.BookingWindowMax,
			BufferTime:       in.Settings.BufferTime,
			ApprovalPolicy:   in.Settings.ApprovalPolicy,
		},
		EmployeeIds:  in.EmployeeIds,
		Phases:       phases,
		UsedProducts: connProducts,
	}
}

func mapToUpdateInput(in updateReq) catalogServ.UpdateInput {
	phases := make([]catalogServ.PhasesInput, len(in.Phases))

	for i, p := range in.Phases {
		phases[i] = catalogServ.PhasesInput{
			Id:        p.Id,
			Name:      p.Name,
			Sequence:  p.Sequence,
			Duration:  p.Duration,
			PhaseType: p.PhaseType,
		}
	}

	return catalogServ.UpdateInput{
		Id:              in.Id,
		BookingType:     in.BookingType,
		Name:            in.Name,
		Description:     in.Description,
		Color:           in.Color,
		Price:           in.Price,
		PriceType:       in.PriceType,
		CategoryId:      in.CategoryId,
		MinParticipants: in.MinParticipants,
		MaxParticipants: in.MaxParticipants,
		IsActive:        in.IsActive,
		Settings: catalogServ.ServiceSettingsInput{
			CancelDeadline:   in.Settings.CancelDeadline,
			BookingWindowMin: in.Settings.BookingWindowMin,
			BookingWindowMax: in.Settings.BookingWindowMax,
			BufferTime:       in.Settings.BufferTime,
			ApprovalPolicy:   in.Settings.ApprovalPolicy,
		},
		EmployeeIds: in.EmployeeIds,
		Phases:      phases,
	}
}

func mapToGetResp(in domain.ServicePageData) getResp {
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
		Id:              in.Id,
		BookingType:     in.BookingType,
		CategoryId:      in.CategoryId,
		Name:            in.Name,
		Description:     in.Description,
		Color:           in.Color,
		TotalDuration:   in.TotalDuration,
		Price:           in.Price,
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
			ApprovalPolicy:   in.Settings.ApprovalPolicy,
		},
		Phases:       mapToPhaseReqs(in.Phases),
		EmployeeIds:  in.EmployeeIds,
		UsedProducts: products,
	}
}

func mapToGetTeamMemberSettingsResp(in catalogServ.GetTeamMemberSettingsResult) getTeamMemberSettingsResp {
	employees := make([]employeeSettingsResp, len(in.TeamMemberSettings))
	for i, employee := range in.TeamMemberSettings {
		phaseOverrides := make([]employeePhaseOverrideResp, len(employee.PhaseOverrides))
		for j, phaseOverride := range employee.PhaseOverrides {
			phaseOverrides[j] = employeePhaseOverrideResp{
				ServicePhaseId: phaseOverride.ServicePhaseId,
				Duration:       phaseOverride.Duration,
			}
		}
		employees[i] = employeeSettingsResp{
			EmployeeId:      employee.EmployeeId,
			FirstName:       employee.FirstName,
			LastName:        employee.LastName,
			Role:            employee.Role,
			IsAssigned:      employee.IsAssigned,
			Price:           employee.PricePerPerson,
			PriceType:       employee.PriceType,
			MinParticipants: employee.MinParticipants,
			MaxParticipants: employee.MaxParticipants,
			BufferTime:      employee.BufferTime,
			PhaseOverrides:  phaseOverrides,
		}
	}

	return getTeamMemberSettingsResp{
		ServiceId:              in.ServiceId,
		ServiceName:            in.ServiceName,
		CurrencyCode:           in.CurrencyCode,
		BookingType:            in.BookingType,
		DefaultDuration:        in.TotalDuration,
		DefaultPrice:           in.PricePerPerson,
		DefaultPriceType:       in.PriceType,
		DefaultMinParticipants: in.MinParticipants,
		DefaultMaxParticipants: in.MaxParticipants,
		DefaultBufferTime:      in.BufferTime,
		Phases:                 mapToPhaseReqs(in.Phases),
		Employees:              employees,
	}
}

func mapToUpdateTeamMemberSettingsInput(in []employeeSettingsReq) []catalogServ.UpdateTeamMemberSettingsInput {
	employees := make([]catalogServ.UpdateTeamMemberSettingsInput, len(in))
	for i, employee := range in {
		phaseOverrides := make([]catalogServ.EmployeeServicePhaseInput, len(employee.PhaseOverrides))
		for j, phaseOverride := range employee.PhaseOverrides {
			phaseOverrides[j] = catalogServ.EmployeeServicePhaseInput{
				ServicePhaseId: phaseOverride.ServicePhaseId,
				Duration:       phaseOverride.Duration,
			}
		}
		employees[i] = catalogServ.UpdateTeamMemberSettingsInput{
			EmployeeId:      employee.EmployeeId,
			IsAssigned:      employee.IsAssigned,
			PricePerPerson:  employee.Price,
			PriceType:       employee.PriceType,
			MinParticipants: employee.MinParticipants,
			MaxParticipants: employee.MaxParticipants,
			BufferTime:      employee.BufferTime,
			PhaseOverrides:  phaseOverrides,
		}
	}

	return employees
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
		UsedProducts: products,
	}
}

func mapToGetAllResp(in []domain.ServicesGroupedByCategory) []getAllResp {
	categories := make([]getAllResp, len(in))

	for i, c := range in {
		services := make([]serviceResp, len(c.Services))

		for j, s := range c.Services {
			services[j] = serviceResp{
				Id:              s.Id,
				MerchantId:      s.MerchantId,
				BookingType:     s.BookingType,
				CategoryId:      s.CategoryId,
				Name:            s.Name,
				Description:     s.Description,
				Color:           s.Color,
				TotalDuration:   s.TotalDuration,
				Price:           currencyx.FormatPrice(s.Price),
				PriceType:       s.PriceType,
				IsActive:        s.IsActive,
				MinParticipants: s.MinParticipants,
				MaxParticipants: s.MaxParticipants,
				Sequence:        s.Sequence,
				Phases:          mapToPhaseReqs(s.Phases),
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
