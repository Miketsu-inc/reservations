package locations

import (
	"github.com/miketsu-inc/reservations/backend/internal/domain"
	locationServ "github.com/miketsu-inc/reservations/backend/internal/service/location"
)

func mapToNewInput(in newReq) locationServ.NewInput {
	return locationServ.NewInput{
		Country:           in.Country,
		CountryCode:       in.CountryCode,
		Region:            in.Region,
		City:              in.City,
		PostalCode:        in.PostalCode,
		Address:           in.Address,
		GeoPoint:          *in.GeoPoint,
		GeocodingProvider: in.GeocodingProvider,
		ProviderPlaceId:   in.ProviderPlaceId,
		FormattedLocation: in.FormattedLocation,
	}
}

func mapToUpdateInput(in updateReq) locationServ.UpdateInput {
	return locationServ.UpdateInput{
		Name:              in.Name,
		ContactEmail:      in.ContactEmail,
		PhoneNumber:       in.PhoneNumber,
		Country:           in.Country,
		CountryCode:       in.CountryCode,
		Region:            in.Region,
		City:              in.City,
		PostalCode:        in.PostalCode,
		Address:           in.Address,
		GeoPoint:          *in.GeoPoint,
		GeocodingProvider: in.GeocodingProvider,
		ProviderPlaceId:   in.ProviderPlaceId,
		FormattedLocation: in.FormattedLocation,
	}
}

func mapToGetResp(in []domain.Location) []getResp {
	out := make([]getResp, len(in))

	for i, location := range in {
		out[i] = getResp{
			Id:                location.Id,
			Name:              location.Name,
			Sequence:          location.Sequence,
			Country:           location.Country,
			CountryCode:       location.CountryCode,
			Region:            location.Region,
			City:              location.City,
			PostalCode:        location.PostalCode,
			Address:           location.Address,
			GeoPoint:          location.GeoPoint,
			GeocodingProvider: location.GeocodingProvider,
			ProviderPlaceId:   location.ProviderPlaceId,
			FormattedLocation: location.FormattedLocation,
			ContactEmail:      location.ContactEmail,
			PhoneNumber:       location.PhoneNumber,
			ParkingInfo:       location.ParkingInfo,
			ArrivalInfo:       location.ArrivalInfo,
			IsActive:          location.IsActive,
		}
	}

	return out
}
