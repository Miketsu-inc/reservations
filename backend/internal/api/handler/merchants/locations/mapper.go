package locations

import (
	"github.com/miketsu-inc/reservations/backend/internal/domain"
	locationServ "github.com/miketsu-inc/reservations/backend/internal/service/location"
	mediaServ "github.com/miketsu-inc/reservations/backend/internal/service/media"
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

func mapToImagesResp(in mediaServ.GetLocationImagesResult) imagesResp {
	images := make([]imageResp, len(in.Images))

	for i, img := range in.Images {
		images[i] = imageResp{
			ImageId:      img.ImageId,
			Sequence:     img.Sequence,
			AltText:      img.AltText,
			Width:        img.Width,
			Height:       img.Height,
			Url:          mediaServ.VariantURL(img.ImageId, mediaServ.VariantLarge),
			ThumbnailUrl: mediaServ.VariantURL(img.ImageId, mediaServ.VariantSmall),
		}
	}

	return imagesResp{
		Images: images,
		Limit:  in.Limit,
	}
}

func mapToUploadInput(in createUploadReq) mediaServ.UploadInput {
	return mediaServ.UploadInput{
		ContentType: in.ContentType,
		Size:        in.Size,
	}
}

func mapToUploadResp(in mediaServ.UploadResult) uploadResp {
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
