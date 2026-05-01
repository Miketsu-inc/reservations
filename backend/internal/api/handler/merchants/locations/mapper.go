package locations

import merchantServ "github.com/miketsu-inc/reservations/backend/internal/service/merchant"

func mapToNewLocationInput(in newReq) merchantServ.NewLocationInput {
	return merchantServ.NewLocationInput{
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
