package media

import (
	"net/http"

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

func (h *Handler) Routes() *httputil.Router {
	r := httputil.NewRouter()

	r.Group(func(r *httputil.Router) {
		r.UseFunc(h.middleware.JwtAuthentication)

		r.Get("/uploads/{imageId}", h.GetUpload)
		r.Post("/uploads/{imageId}/complete", h.CompleteUpload)
	})

	return r
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
