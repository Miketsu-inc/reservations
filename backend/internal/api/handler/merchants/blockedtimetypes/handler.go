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

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

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

func (h *Handler) New(w http.ResponseWriter, r *http.Request) {
	var req newReq

	if err := validate.ParseStruct(r, &req); err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	err := h.service.NewType(r.Context(), mapToNewTypeInput(req))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

type updateReq struct {
	Id       int    `json:"id" validate:"required"`
	Name     string `json:"name" validate:"required,max=50"`
	Duration int    `json:"duration" validate:"required,gte=1"`
	Icon     string `json:"icon" validate:"max=20"`
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	var req updateReq

	if err := validate.ParseStruct(r, &req); err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	urlBlockedTimeTypeId, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	err = h.service.UpdateType(r.Context(), urlBlockedTimeTypeId, mapToUpdateTypeInput(req))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	urlBlockedTimeTypeId, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	err = h.service.DeleteType(r.Context(), urlBlockedTimeTypeId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}
}

type getTypesResp struct {
	Id       int    `json:"id"`
	Name     string `json:"name"`
	Duration int    `json:"duration"`
	Icon     string `json:"icon"`
}

func (h *Handler) GetAll(w http.ResponseWriter, r *http.Request) {
	blockedTimeTypes, err := h.service.GetTypes(r.Context())
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	httputil.Success(w, http.StatusOK, mapToGetTypesResp(blockedTimeTypes))
}
