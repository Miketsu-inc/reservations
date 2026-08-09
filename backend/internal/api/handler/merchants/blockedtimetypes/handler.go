package blockedtimetypes

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	blockedtimeServ "github.com/miketsu-inc/reservations/backend/internal/service/blockedtime"
	"github.com/miketsu-inc/reservations/backend/pkg/httputil"
	"github.com/miketsu-inc/reservations/backend/pkg/validate"
)

type Handler struct {
	service *blockedtimeServ.Service
}

func NewHandler(s *blockedtimeServ.Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) Routes() *httputil.Router {
	r := httputil.NewRouter()

	r.Post("/", h.New)
	r.Put("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)

	r.Get("/", h.GetAll)

	return r
}

type newReq struct {
	Name     string `json:"name" validate:"required,max=50"`
	Duration int    `json:"duration" validate:"required,gte=1"`
	Icon     string `json:"icon" validate:"max=20"`
}

func (h *Handler) New(w http.ResponseWriter, r *http.Request) error {
	var req newReq

	if err := validate.ParseStruct(r, &req); err != nil {
		return err
	}

	err := h.service.NewType(r.Context(), mapToNewTypeInput(req))
	if err != nil {
		return blockedtimeServ.ErrStatus.Resolve(err, "NewType")
	}

	w.WriteHeader(http.StatusCreated)

	return nil
}

type updateReq struct {
	Id       int    `json:"id" validate:"required"`
	Name     string `json:"name" validate:"required,max=50"`
	Duration int    `json:"duration" validate:"required,gte=1"`
	Icon     string `json:"icon" validate:"max=20"`
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) error {
	var req updateReq

	if err := validate.ParseStruct(r, &req); err != nil {
		return err
	}

	urlBlockedTimeTypeId, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return validate.NewError("invalid blocked time type id")
	}

	err = h.service.UpdateType(r.Context(), urlBlockedTimeTypeId, mapToUpdateTypeInput(req))
	if err != nil {
		return blockedtimeServ.ErrStatus.Resolve(err, "UpdateType")
	}

	return nil
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) error {
	urlBlockedTimeTypeId, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return validate.NewError("invalid blocked time type id")
	}

	err = h.service.DeleteType(r.Context(), urlBlockedTimeTypeId)
	if err != nil {
		return blockedtimeServ.ErrStatus.Resolve(err, "DeleteType")
	}

	return nil
}

type getTypesResp struct {
	Id       int    `json:"id"`
	Name     string `json:"name"`
	Duration int    `json:"duration"`
	Icon     string `json:"icon"`
}

func (h *Handler) GetAll(w http.ResponseWriter, r *http.Request) error {
	blockedTimeTypes, err := h.service.GetTypes(r.Context())
	if err != nil {
		return blockedtimeServ.ErrStatus.Resolve(err, "GetTypes")
	}

	httputil.Success(w, http.StatusOK, mapToGetTypesResp(blockedTimeTypes))

	return nil
}
