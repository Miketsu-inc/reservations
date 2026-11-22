package favorites

import (
	"github.com/miketsu-inc/reservations/backend/internal/domain"
	mediaServ "github.com/miketsu-inc/reservations/backend/internal/service/media"
)

func mapToGetFavoritesResp(in []domain.FavoriteMerchant) []favoriteMerchantResp {
	resp := make([]favoriteMerchantResp, len(in))
	for i, f := range in {
		resp[i] = favoriteMerchantResp{
			MerchantId:        f.MerchantId,
			MerchantName:      f.MerchantName,
			MerchantUrlName:   f.MerchantUrlName,
			FormattedLocation: f.FormattedLocation,
			AverageRating:     f.AverageRating,
			ReviewCount:       f.ReviewCount,
			MainImageUrl:      mediaServ.NewURL(f.MainImageId, mediaServ.VariantSmall),
		}
	}
	return resp
}
