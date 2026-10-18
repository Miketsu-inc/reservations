package media

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/miketsu-inc/reservations/backend/internal/api/middleware"
	mediaServ "github.com/miketsu-inc/reservations/backend/internal/service/media"
	"github.com/miketsu-inc/reservations/backend/internal/types"
	"github.com/miketsu-inc/reservations/backend/pkg/httputil"
	"github.com/miketsu-inc/reservations/backend/pkg/validate"
)

type Handler struct {
	service    *mediaServ.Service
	middleware *middleware.Manager
}

func NewHandler(s *mediaServ.Service, m *middleware.Manager) *Handler {
	return &Handler{service: s, middleware: m}
}

// Routes of the authenticated user
func (h *Handler) Routes() *httputil.Router {
	r := httputil.NewRouter()

	r.Group(func(r *httputil.Router) {
		r.UseFunc(h.middleware.JwtAuthentication)
		r.UseFunc(h.middleware.Language)

		r.Post("/avatar", h.CreateUserAvatarUpload)
		r.Delete("/avatar", h.RemoveUserAvatar)

		r.Get("/uploads/{imageId}", h.GetUpload)
		r.Post("/uploads/{imageId}/complete", h.CompleteUpload)
	})

	return r
}

// Routes of a merchant, should be mounted after the employee authentication middleware
func (h *Handler) MerchantRoutes() *httputil.Router {
	r := httputil.NewRouter()

	// staff members can only change their own picture, checked in the service
	r.Post("/team/{employeeId}/avatar", h.CreateEmployeeAvatarUpload)
	r.Delete("/team/{employeeId}/avatar", h.RemoveEmployeeAvatar)

	r.Get("/locations/{locationId}/images", h.GetLocationImages)

	r.Group(func(r *httputil.Router) {
		r.UseFunc(h.middleware.RoleBasedAccessControl(types.EmployeeRoleOwner, types.EmployeeRoleAdmin))

		r.Post("/logo", h.CreateMerchantLogoUpload)
		r.Delete("/logo", h.RemoveMerchantLogo)

		r.Post("/locations/{locationId}/images", h.CreateLocationImageUpload)
		r.Put("/locations/{locationId}/images/order", h.ReorderLocationImages)
		r.Patch("/locations/{locationId}/images/{imageId}", h.UpdateLocationImage)
		r.Delete("/locations/{locationId}/images/{imageId}", h.DeleteLocationImage)
	})

	return r
}

type createUploadReq struct {
	ContentType types.ImageContentType `json:"content_type" validate:"required"`
	Size        int64                  `json:"size" validate:"required,min=1"`
}

type uploadResp struct {
	ImageId   uuid.UUID         `json:"image_id"`
	UploadUrl string            `json:"upload_url"`
	Method    string            `json:"method"`
	Headers   map[string]string `json:"headers"`
	ExpiresAt time.Time         `json:"expires_at"`
}

func (h *Handler) CreateUserAvatarUpload(w http.ResponseWriter, r *http.Request) error {
	var req createUploadReq

	if err := validate.ParseStruct(r, &req); err != nil {
		return err
	}

	upload, err := h.service.CreateUserAvatarUpload(r.Context(), mapToUploadInput(req))
	if err != nil {
		return mediaServ.ErrStatus.Resolve(err, "CreateUserAvatarUpload")
	}

	httputil.Success(w, http.StatusCreated, mapToUploadResp(upload))

	return nil
}

func (h *Handler) RemoveUserAvatar(w http.ResponseWriter, r *http.Request) error {
	err := h.service.RemoveUserAvatar(r.Context())
	if err != nil {
		return mediaServ.ErrStatus.Resolve(err, "RemoveUserAvatar")
	}

	w.WriteHeader(http.StatusNoContent)

	return nil
}

type uploadStatusResp struct {
	ImageId      uuid.UUID          `json:"image_id"`
	Purpose      types.ImagePurpose `json:"purpose"`
	Status       types.ImageStatus  `json:"status"`
	Url          *string            `json:"url"`
	ThumbnailUrl *string            `json:"thumbnail_url"`
}

func (h *Handler) GetUpload(w http.ResponseWriter, r *http.Request) error {
	imageId, err := uuid.Parse(chi.URLParam(r, "imageId"))
	if err != nil {
		return validate.NewError("invalid image id")
	}

	image, err := h.service.GetUpload(r.Context(), imageId)
	if err != nil {
		return mediaServ.ErrStatus.Resolve(err, "GetUpload")
	}

	httputil.Success(w, http.StatusOK, mapToUploadStatusResp(image))

	return nil
}

type cropReq struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

type completeUploadReq struct {
	// relative to the image as it is displayed, in pixels
	Crop *cropReq `json:"crop"`
}

func (h *Handler) CompleteUpload(w http.ResponseWriter, r *http.Request) error {
	var req completeUploadReq

	if err := validate.ParseStruct(r, &req); err != nil {
		return err
	}

	if c := req.Crop; c != nil && (c.X < 0 || c.Y < 0 || c.Width < 1 || c.Height < 1) {
		return validate.NewError("invalid crop")
	}

	imageId, err := uuid.Parse(chi.URLParam(r, "imageId"))
	if err != nil {
		return validate.NewError("invalid image id")
	}

	err = h.service.CompleteUpload(r.Context(), imageId, mapToImageCrop(req.Crop))
	if err != nil {
		return mediaServ.ErrStatus.Resolve(err, "CompleteUpload")
	}

	w.WriteHeader(http.StatusAccepted)

	return nil
}

func (h *Handler) CreateEmployeeAvatarUpload(w http.ResponseWriter, r *http.Request) error {
	var req createUploadReq

	if err := validate.ParseStruct(r, &req); err != nil {
		return err
	}

	employeeId, err := strconv.Atoi(chi.URLParam(r, "employeeId"))
	if err != nil {
		return validate.NewError("invalid team member id")
	}

	upload, err := h.service.CreateEmployeeAvatarUpload(r.Context(), employeeId, mapToUploadInput(req))
	if err != nil {
		return mediaServ.ErrStatus.Resolve(err, "CreateEmployeeAvatarUpload")
	}

	httputil.Success(w, http.StatusCreated, mapToUploadResp(upload))

	return nil
}

func (h *Handler) RemoveEmployeeAvatar(w http.ResponseWriter, r *http.Request) error {
	employeeId, err := strconv.Atoi(chi.URLParam(r, "employeeId"))
	if err != nil {
		return validate.NewError("invalid team member id")
	}

	err = h.service.RemoveEmployeeAvatar(r.Context(), employeeId)
	if err != nil {
		return mediaServ.ErrStatus.Resolve(err, "RemoveEmployeeAvatar")
	}

	w.WriteHeader(http.StatusNoContent)

	return nil
}

func (h *Handler) CreateMerchantLogoUpload(w http.ResponseWriter, r *http.Request) error {
	var req createUploadReq

	if err := validate.ParseStruct(r, &req); err != nil {
		return err
	}

	upload, err := h.service.CreateMerchantLogoUpload(r.Context(), mapToUploadInput(req))
	if err != nil {
		return mediaServ.ErrStatus.Resolve(err, "CreateMerchantLogoUpload")
	}

	httputil.Success(w, http.StatusCreated, mapToUploadResp(upload))

	return nil
}

func (h *Handler) RemoveMerchantLogo(w http.ResponseWriter, r *http.Request) error {
	err := h.service.RemoveMerchantLogo(r.Context())
	if err != nil {
		return mediaServ.ErrStatus.Resolve(err, "RemoveMerchantLogo")
	}

	w.WriteHeader(http.StatusNoContent)

	return nil
}

type locationImagesResp struct {
	Images []locationImageResp `json:"images"`
	Limit  int                 `json:"limit"`
}

type locationImageResp struct {
	ImageId      uuid.UUID `json:"image_id"`
	Sequence     int       `json:"sequence"`
	AltText      string    `json:"alt_text"`
	Width        int       `json:"width"`
	Height       int       `json:"height"`
	Url          string    `json:"url"`
	ThumbnailUrl string    `json:"thumbnail_url"`
}

func (h *Handler) GetLocationImages(w http.ResponseWriter, r *http.Request) error {
	locationId, err := strconv.Atoi(chi.URLParam(r, "locationId"))
	if err != nil {
		return validate.NewError("invalid location id")
	}

	images, err := h.service.GetLocationImages(r.Context(), locationId)
	if err != nil {
		return mediaServ.ErrStatus.Resolve(err, "GetLocationImages")
	}

	httputil.Success(w, http.StatusOK, mapToLocationImagesResp(images))

	return nil
}

func (h *Handler) CreateLocationImageUpload(w http.ResponseWriter, r *http.Request) error {
	var req createUploadReq

	if err := validate.ParseStruct(r, &req); err != nil {
		return err
	}

	locationId, err := strconv.Atoi(chi.URLParam(r, "locationId"))
	if err != nil {
		return validate.NewError("invalid location id")
	}

	upload, err := h.service.CreateLocationImageUpload(r.Context(), locationId, mapToUploadInput(req))
	if err != nil {
		return mediaServ.ErrStatus.Resolve(err, "CreateLocationImageUpload")
	}

	httputil.Success(w, http.StatusCreated, mapToUploadResp(upload))

	return nil
}

type updateLocationImageReq struct {
	AltText string `json:"alt_text" validate:"max=150"`
}

func (h *Handler) UpdateLocationImage(w http.ResponseWriter, r *http.Request) error {
	var req updateLocationImageReq

	if err := validate.ParseStruct(r, &req); err != nil {
		return err
	}

	locationId, err := strconv.Atoi(chi.URLParam(r, "locationId"))
	if err != nil {
		return validate.NewError("invalid location id")
	}

	imageId, err := uuid.Parse(chi.URLParam(r, "imageId"))
	if err != nil {
		return validate.NewError("invalid image id")
	}

	err = h.service.UpdateLocationImage(r.Context(), locationId, imageId, req.AltText)
	if err != nil {
		return mediaServ.ErrStatus.Resolve(err, "UpdateLocationImage")
	}

	return nil
}

func (h *Handler) DeleteLocationImage(w http.ResponseWriter, r *http.Request) error {
	locationId, err := strconv.Atoi(chi.URLParam(r, "locationId"))
	if err != nil {
		return validate.NewError("invalid location id")
	}

	imageId, err := uuid.Parse(chi.URLParam(r, "imageId"))
	if err != nil {
		return validate.NewError("invalid image id")
	}

	err = h.service.DeleteLocationImage(r.Context(), locationId, imageId)
	if err != nil {
		return mediaServ.ErrStatus.Resolve(err, "DeleteLocationImage")
	}

	w.WriteHeader(http.StatusNoContent)

	return nil
}

type reorderLocationImagesReq struct {
	ImageIds []uuid.UUID `json:"image_ids" validate:"required"`
}

func (h *Handler) ReorderLocationImages(w http.ResponseWriter, r *http.Request) error {
	var req reorderLocationImagesReq

	if err := validate.ParseStruct(r, &req); err != nil {
		return err
	}

	locationId, err := strconv.Atoi(chi.URLParam(r, "locationId"))
	if err != nil {
		return validate.NewError("invalid location id")
	}

	err = h.service.ReorderLocationImages(r.Context(), locationId, req.ImageIds)
	if err != nil {
		return mediaServ.ErrStatus.Resolve(err, "ReorderLocationImages")
	}

	return nil
}
