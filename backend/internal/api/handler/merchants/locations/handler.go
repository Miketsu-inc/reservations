package locations

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/miketsu-inc/reservations/backend/internal/api/middleware"
	locationServ "github.com/miketsu-inc/reservations/backend/internal/service/location"
	mediaServ "github.com/miketsu-inc/reservations/backend/internal/service/media"
	"github.com/miketsu-inc/reservations/backend/internal/types"
	"github.com/miketsu-inc/reservations/backend/pkg/httputil"
	"github.com/miketsu-inc/reservations/backend/pkg/validate"
)

type Handler struct {
	service    *locationServ.Service
	mediaServ  *mediaServ.Service
	middleware *middleware.Manager
}

func NewHandler(s *locationServ.Service, ms *mediaServ.Service, m *middleware.Manager) *Handler {
	return &Handler{service: s, mediaServ: ms, middleware: m}
}

func (h *Handler) Routes() *httputil.Router {
	r := httputil.NewRouter()

	// TODO: temp until signup flow is figured out?
	r.Post("/", h.New)
	r.Get("/", h.GetAll)
	r.Patch("/{id}", h.Update)

	r.Get("/{id}/images", h.GetImages)

	r.Group(func(r *httputil.Router) {
		r.UseFunc(h.middleware.RoleBasedAccessControl(types.EmployeeRoleOwner, types.EmployeeRoleAdmin))

		r.WithFunc(h.middleware.RateLimitByUser(30, time.Hour)).Post("/{id}/images", h.CreateImageUpload)
		r.Put("/{id}/images/order", h.ReorderImages)
		r.Patch("/{id}/images/{imageId}", h.UpdateImage)
		r.Delete("/{id}/images/{imageId}", h.DeleteImage)
	})

	return r
}

type newReq struct {
	Country           *string         `json:"country"`
	CountryCode       *string         `json:"country_code"`
	Region            *string         `json:"region"`
	City              *string         `json:"city"`
	PostalCode        *string         `json:"postal_code"`
	Address           *string         `json:"address"`
	GeoPoint          *types.GeoPoint `json:"geo_point" validate:"required"`
	GeocodingProvider *string         `json:"geocoding_provider"`
	ProviderPlaceId   *string         `json:"provider_place_id"`
	FormattedLocation string          `json:"formatted_location" validate:"required"`
}

func (h *Handler) New(w http.ResponseWriter, r *http.Request) error {
	var req newReq

	if err := validate.ParseStruct(r, &req); err != nil {
		return err
	}
	if err := validateGeoPoint(*req.GeoPoint); err != nil {
		return err
	}

	err := h.service.New(r.Context(), mapToNewInput(req))
	if err != nil {
		return locationServ.ErrStatus.Resolve(err, "New")
	}

	w.WriteHeader(http.StatusCreated)

	return nil
}

type getResp struct {
	Id                int            `json:"id"`
	Name              string         `json:"name"`
	Sequence          int            `json:"sequence"`
	Country           *string        `json:"country"`
	CountryCode       *string        `json:"country_code"`
	Region            *string        `json:"region"`
	City              *string        `json:"city"`
	PostalCode        *string        `json:"postal_code"`
	Address           *string        `json:"address"`
	GeoPoint          types.GeoPoint `json:"geo_point"`
	GeocodingProvider *string        `json:"geocoding_provider"`
	ProviderPlaceId   *string        `json:"provider_place_id"`
	FormattedLocation string         `json:"formatted_location"`
	ContactEmail      *string        `json:"contact_email"`
	PhoneNumber       *string        `json:"phone_number"`
	ParkingInfo       string         `json:"parking_info"`
	ArrivalInfo       string         `json:"arrival_info"`
	IsActive          bool           `json:"is_active"`
}

func (h *Handler) GetAll(w http.ResponseWriter, r *http.Request) error {
	locations, err := h.service.GetAll(r.Context())
	if err != nil {
		return locationServ.ErrStatus.Resolve(err, "GetAll")
	}

	httputil.Success(w, http.StatusOK, mapToGetResp(locations))

	return nil
}

type updateReq struct {
	Name              string          `json:"name" validate:"required,max=100"`
	ContactEmail      *string         `json:"contact_email" validate:"omitempty,email,max=320"`
	PhoneNumber       *string         `json:"phone_number" validate:"omitempty,max=30"`
	Country           *string         `json:"country"`
	CountryCode       *string         `json:"country_code"`
	Region            *string         `json:"region"`
	City              *string         `json:"city"`
	PostalCode        *string         `json:"postal_code"`
	Address           *string         `json:"address"`
	GeoPoint          *types.GeoPoint `json:"geo_point" validate:"required"`
	GeocodingProvider *string         `json:"geocoding_provider"`
	ProviderPlaceId   *string         `json:"provider_place_id"`
	FormattedLocation string          `json:"formatted_location" validate:"required"`
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) error {
	var req updateReq

	if err := validate.ParseStruct(r, &req); err != nil {
		return err
	}

	if err := validateGeoPoint(*req.GeoPoint); err != nil {
		return err
	}

	locationId, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return validate.NewError("invalid location id")
	}

	err = h.service.Update(r.Context(), locationId, mapToUpdateInput(req))
	if err != nil {
		return locationServ.ErrStatus.Resolve(err, "Update")
	}

	w.WriteHeader(http.StatusNoContent)

	return nil
}

func validateGeoPoint(point types.GeoPoint) error {
	if point.Lat < -90 || point.Lat > 90 || point.Lon < -180 || point.Lon > 180 {
		return validate.NewError("geo_point is outside the valid coordinate range")
	}

	return nil
}

type imagesResp struct {
	Images []imageResp `json:"images"`
	Limit  int         `json:"limit"`
}

type imageResp struct {
	ImageId      uuid.UUID `json:"image_id"`
	Sequence     int       `json:"sequence"`
	AltText      string    `json:"alt_text"`
	Width        int       `json:"width"`
	Height       int       `json:"height"`
	Url          string    `json:"url"`
	ThumbnailUrl string    `json:"thumbnail_url"`
}

func (h *Handler) GetImages(w http.ResponseWriter, r *http.Request) error {
	locationId, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return validate.NewError("invalid location id")
	}

	images, err := h.mediaServ.GetLocationImages(r.Context(), locationId)
	if err != nil {
		return mediaServ.ErrStatus.Resolve(err, "GetImages")
	}

	httputil.Success(w, http.StatusOK, mapToImagesResp(images))

	return nil
}

type createUploadReq struct {
	ContentType string `json:"content_type" validate:"required,eq=image/jpeg|eq=image/png|eq=image/webp"`
	Size        int64  `json:"size" validate:"required,min=1"`
}

type uploadResp struct {
	ImageId   uuid.UUID         `json:"image_id"`
	UploadUrl string            `json:"upload_url"`
	Method    string            `json:"method"`
	Headers   map[string]string `json:"headers"`
	ExpiresAt time.Time         `json:"expires_at"`
}

func (h *Handler) CreateImageUpload(w http.ResponseWriter, r *http.Request) error {
	var req createUploadReq

	if err := validate.ParseStruct(r, &req); err != nil {
		return err
	}

	locationId, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return validate.NewError("invalid location id")
	}

	upload, err := h.mediaServ.CreateLocationImageUpload(r.Context(), locationId, mapToUploadInput(req))
	if err != nil {
		return mediaServ.ErrStatus.Resolve(err, "CreateImageUpload")
	}

	httputil.Success(w, http.StatusCreated, mapToUploadResp(upload))

	return nil
}

type updateImageReq struct {
	AltText string `json:"alt_text" validate:"max=150"`
}

func (h *Handler) UpdateImage(w http.ResponseWriter, r *http.Request) error {
	var req updateImageReq

	if err := validate.ParseStruct(r, &req); err != nil {
		return err
	}

	locationId, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return validate.NewError("invalid location id")
	}

	imageId, err := uuid.Parse(chi.URLParam(r, "imageId"))
	if err != nil {
		return validate.NewError("invalid image id")
	}

	err = h.mediaServ.UpdateLocationImage(r.Context(), locationId, imageId, req.AltText)
	if err != nil {
		return mediaServ.ErrStatus.Resolve(err, "UpdateImage")
	}

	w.WriteHeader(http.StatusNoContent)

	return nil
}

func (h *Handler) DeleteImage(w http.ResponseWriter, r *http.Request) error {
	locationId, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return validate.NewError("invalid location id")
	}

	imageId, err := uuid.Parse(chi.URLParam(r, "imageId"))
	if err != nil {
		return validate.NewError("invalid image id")
	}

	err = h.mediaServ.DeleteLocationImage(r.Context(), locationId, imageId)
	if err != nil {
		return mediaServ.ErrStatus.Resolve(err, "DeleteImage")
	}

	w.WriteHeader(http.StatusNoContent)

	return nil
}

type reorderImagesReq struct {
	ImageIds []uuid.UUID `json:"image_ids" validate:"required"`
}

func (h *Handler) ReorderImages(w http.ResponseWriter, r *http.Request) error {
	var req reorderImagesReq

	if err := validate.ParseStruct(r, &req); err != nil {
		return err
	}

	locationId, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return validate.NewError("invalid location id")
	}

	err = h.mediaServ.ReorderLocationImages(r.Context(), locationId, req.ImageIds)
	if err != nil {
		return mediaServ.ErrStatus.Resolve(err, "ReorderImages")
	}

	w.WriteHeader(http.StatusNoContent)

	return nil
}
