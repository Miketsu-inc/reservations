package favorites

import (
	"github.com/miketsu-inc/reservations/backend/internal/domain"
	mediaServ "github.com/miketsu-inc/reservations/backend/internal/service/media"
)

func mapToGetFavoritesResp(in []domain.FavoriteLocation) []favoriteLocationResp {
	resp := make([]favoriteLocationResp, len(in))
	for i, f := range in {
		resp[i] = favoriteLocationResp{
			MerchantId:        f.MerchantId,
			MerchantName:      f.MerchantName,
			MerchantUrlName:   f.MerchantUrlName,
			LocationId:        f.LocationId,
			FormattedLocation: f.FormattedLocation,
			AverageRating:     f.AverageRating,
			ReviewCount:       f.ReviewCount,
			MainImageUrl:      mediaServ.NewURL(f.MainImageId, mediaServ.VariantSmall),
		}
	}
	return resp
}
