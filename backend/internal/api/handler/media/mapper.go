package media

import (
	"github.com/miketsu-inc/reservations/backend/internal/domain"
	mediaServ "github.com/miketsu-inc/reservations/backend/internal/service/media"
	"github.com/miketsu-inc/reservations/backend/internal/types"
)

func mapToImageCrop(in *cropReq) *domain.ImageCrop {
	if in == nil {
		return nil
	}

	return &domain.ImageCrop{
		X:      in.X,
		Y:      in.Y,
		Width:  in.Width,
		Height: in.Height,
	}
}

func mapToUploadStatusResp(in domain.Image) uploadStatusResp {
	resp := uploadStatusResp{
		ImageId: in.Id,
		Purpose: in.Purpose,
		Status:  in.Status,
	}

	if in.Status == types.ImageStatusReady {
		resp.Url = mediaServ.NewURL(&in.Id, mediaServ.VariantLarge)
		resp.ThumbnailUrl = mediaServ.NewURL(&in.Id, mediaServ.VariantSmall)
	}

	return resp
}
