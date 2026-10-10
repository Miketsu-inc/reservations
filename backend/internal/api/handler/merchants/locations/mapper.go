package locations

import (
	"github.com/miketsu-inc/reservations/backend/internal/domain"
	locationServ "github.com/miketsu-inc/reservations/backend/internal/service/location"
)

func mapToNewInput(in newReq) locationServ.NewInput {
	return locationServ.NewInput{
		Country:           in.Country,
		City:              in.City,
		PostalCode:        in.PostalCode,
		Address:           in.Address,
		GeoPoint:          *in.GeoPoint,
		PlaceId:           in.PlaceId,
		FormattedLocation: in.FormattedLocation,
		IsPrimary:         in.IsPrimary,
		IsActive:          in.IsActive,
	}
}

func mapToUpdateInput(in updateReq) locationServ.UpdateInput {
	return locationServ.UpdateInput{
		Country:           in.Country,
		City:              in.City,
		PostalCode:        in.PostalCode,
		Address:           in.Address,
		GeoPoint:          *in.GeoPoint,
		PlaceId:           in.PlaceId,
		FormattedLocation: in.FormattedLocation,
	}
}

func mapToGetResp(in []domain.Location) []getResp {
	out := make([]getResp, len(in))

	for i, location := range in {
		out[i] = getResp{
			Id:                location.Id,
			Country:           location.Country,
			City:              location.City,
			PostalCode:        location.PostalCode,
			Address:           location.Address,
			GeoPoint:          location.GeoPoint,
			PlaceId:           location.PlaceId,
			FormattedLocation: location.FormattedLocation,
			IsPrimary:         location.IsPrimary,
			IsActive:          location.IsActive,
		}
	}

	return out
}
