package servicecategories

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	catalogServ "github.com/miketsu-inc/reservations/backend/internal/service/catalog"
	"github.com/miketsu-inc/reservations/backend/pkg/httputil"
	"github.com/miketsu-inc/reservations/backend/pkg/validate"
)

type Handler struct {
	service *catalogServ.Service
}

func NewHandler(s *catalogServ.Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/", h.New)
	r.Put("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)

	r.Put("/reorder", h.ReorderCategories)

	return r
}

type newReq struct {
	Name string `json:"name" validate:"required"`
}

func (h *Handler) New(w http.ResponseWriter, r *http.Request) {
	var req newReq

	err := validate.ParseStruct(r, &req)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	err = h.service.NewCategory(r.Context(), mapToNewCategoryInput(req))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}
}

type updateReq struct {
	Name string `json:"name" validate:"required"`
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	var req updateReq

	err := validate.ParseStruct(r, &req)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	urlCategoryId, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("invalid service category id"))
		return
	}

	err = h.service.UpdateCategory(r.Context(), urlCategoryId, mapToUpdateCategoryInput(req))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	urlCategoryId, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("invalid service category id"))
		return
	}

	err = h.service.DeleteCategory(r.Context(), urlCategoryId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}
}

type reorderCategoriesReq struct {
	Categories []int `json:"categories" validate:"required"`
}

func (h *Handler) ReorderCategories(w http.ResponseWriter, r *http.Request) {
	var req reorderCategoriesReq

	err := validate.ParseStruct(r, &req)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	err = h.service.ReorderCategories(r.Context(), mapToReorderCategoriesInput(req))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}
}
