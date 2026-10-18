package media

import (
	"github.com/miketsu-inc/reservations/backend/internal/domain"
	mediaServ "github.com/miketsu-inc/reservations/backend/internal/service/media"
	"github.com/miketsu-inc/reservations/backend/internal/types"
)

func mapToUploadInput(in createUploadReq) mediaServ.UploadInput {
	return mediaServ.UploadInput{
		ContentType: in.ContentType,
		Size:        in.Size,
	}
}

func mapToUploadResp(in mediaServ.Upload) uploadResp {
	headers := make(map[string]string, len(in.Request.Headers))
	for key := range in.Request.Headers {
		headers[key] = in.Request.Headers.Get(key)
	}

	return uploadResp{
		ImageId:   in.ImageId,
		UploadUrl: in.Request.URL,
		Method:    in.Request.Method,
		Headers:   headers,
		ExpiresAt: in.Request.ExpiresAt,
	}
}

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
		resp.Url = mediaServ.URL(&in.Id, mediaServ.VariantLarge)
		resp.ThumbnailUrl = mediaServ.URL(&in.Id, mediaServ.VariantSmall)
	}

	return resp
}

func mapToLocationImagesResp(in mediaServ.LocationImages) locationImagesResp {
	images := make([]locationImageResp, len(in.Images))

	for i, img := range in.Images {
		images[i] = locationImageResp{
			ImageId:      img.ImageId,
			Sequence:     img.Sequence,
			AltText:      img.AltText,
			Width:        img.Width,
			Height:       img.Height,
			Url:          mediaServ.VariantURL(img.ImageId, mediaServ.VariantLarge),
			ThumbnailUrl: mediaServ.VariantURL(img.ImageId, mediaServ.VariantSmall),
		}
	}

	return locationImagesResp{
		Images: images,
		Limit:  in.Limit,
	}
}
