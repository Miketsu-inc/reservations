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
		GeoPoint:          in.GeoPoint,
		PlaceId:           in.PlaceId,
		FormattedLocation: in.FormattedLocation,
		IsPrimary:         in.IsPrimary,
		IsActive:          in.IsActive,
	}
}

func mapToGetResp(in domain.Location) getResp {
	return getResp{
		Id:                in.Id,
		Country:           in.Country,
		City:              in.City,
		PostalCode:        in.PostalCode,
		Address:           in.Address,
		FormattedLocation: in.FormattedLocation,
	}
}
