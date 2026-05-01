package merchants

import (
	"github.com/miketsu-inc/reservations/backend/internal/domain"
	merchantServ "github.com/miketsu-inc/reservations/backend/internal/service/merchant"
	"github.com/miketsu-inc/reservations/backend/pkg/currencyx"
)

func mapToGetInfo(in domain.MerchantInfo) getInfoResp {
	servicesGroupedByCategory := make([]servicesGroupedByCategoryResp, len(in.Services))

	for i, serv := range in.Services {
		services := make([]serviceResp, len(serv.Services))

		for j, s := range serv.Services {
			services[j] = serviceResp{
				Id:            s.Id,
				CategoryId:    s.CategoryId,
				Name:          s.Name,
				Description:   s.Description,
				TotalDuration: s.TotalDuration,
				Price:         currencyx.FormatPrice(s.Price),
				PriceType:     s.PriceType,
				Sequence:      s.Sequence,
			}
		}

		servicesGroupedByCategory[i] = servicesGroupedByCategoryResp{
			Id:       serv.Id,
			Name:     serv.Name,
			Sequence: serv.Sequence,
			Services: services,
		}
	}

	businessHours := make(map[int][]timeSlotResp, len(in.BusinessHours))

	for day, slots := range in.BusinessHours {
		timeSlots := make([]timeSlotResp, len(slots))

		for i, s := range slots {
			timeSlots[i] = timeSlotResp{
				StartTime: s.StartTime.Format("15:04"),
				EndTime:   s.EndTime.Format("15:04"),
			}
		}

		businessHours[day] = timeSlots
	}

	return getInfoResp{
		Name:              in.Name,
		UrlName:           in.UrlName,
		ContactEmail:      in.ContactEmail,
		Introduction:      in.Introduction,
		Announcement:      in.Announcement,
		AboutUs:           in.AboutUs,
		ParkingInfo:       in.ParkingInfo,
		PaymentInfo:       in.PaymentInfo,
		Timezone:          in.Timezone,
		LocationId:        in.LocationId,
		Country:           in.Country,
		City:              in.City,
		PostalCode:        in.PostalCode,
		Address:           in.Address,
		FormattedLocation: in.FormattedLocation,
		GeoPoint:          in.GeoPoint,
		Services:          servicesGroupedByCategory,
		BusinessHours:     businessHours,
	}
}

func mapToGetNormalizedBusinessHoursResp(in domain.BusinessHours) map[int]timeSlotResp {
	businessHours := make(map[int]timeSlotResp, len(in))

	for day, ts := range in {
		businessHours[day] = timeSlotResp{
			StartTime: ts[0].StartTime.Format("15:04"),
			EndTime:   ts[0].EndTime.Format("15:04"),
		}
	}

	return businessHours
}

func mapToGetServiceDetailsResp(in domain.PublicServiceDetails) getServiceDetailsResp {
	phases := make([]phaseResp, len(in.Phases))

	for i, p := range in.Phases {
		phases[i] = phaseResp{
			Id:        p.Id,
			ServiceId: p.ServiceId,
			Sequence:  p.Sequence,
			Duration:  p.Duration,
			PhaseType: p.PhaseType,
		}
	}

	return getServiceDetailsResp{
		Id:                in.Id,
		Name:              in.Name,
		Description:       in.Description,
		TotalDuration:     in.TotalDuration,
		Price:             currencyx.FormatPrice(in.Price),
		PriceType:         in.PriceType,
		FormattedLocation: in.FormattedLocation,
		GeoPoint:          in.GeoPoint,
		Phases:            phases,
	}
}

func mapToGetSummaryResp(in domain.MinimalServiceInfo) getSummaryResp {
	return getSummaryResp{
		Name:              in.Name,
		TotalDuration:     in.TotalDuration,
		Price:             currencyx.FormatPrice(in.Price),
		PriceType:         in.PriceType,
		FormattedLocation: in.FormattedLocation,
	}
}

func mapToGetAvailabilityResp(in []merchantServ.MultiDayAvailableTimes) []getAvailabilityResp {
	availability := make([]getAvailabilityResp, len(in))

	for i, a := range in {
		availability[i] = getAvailabilityResp{
			Date:        a.Date,
			IsAvailable: a.IsAvailable,
			Morning:     a.Morning,
			Afternoon:   a.Afternoon,
		}
	}

	return availability
}

func mapToGetNextAvailabilityResp(in merchantServ.NextAvailable) getNextAvailabilityResp {
	return getNextAvailabilityResp{
		Date: in.Date,
		Time: in.Time,
	}
}

func mapToGetDisabledDaysResp(in merchantServ.DisabledDays) getDisabledDaysResp {
	return getDisabledDaysResp{
		ClosedDays: in.ClosedDays,
		MinDate:    in.MinDate,
		MaxDate:    in.MaxDate,
	}
}
