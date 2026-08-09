package servicecategories

import (
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

func (h *Handler) Routes() *httputil.Router {
	r := httputil.NewRouter()

	r.Post("/", h.New)
	r.Put("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)

	r.Put("/reorder", h.ReorderCategories)

	return r
}

type newReq struct {
	Name string `json:"name" validate:"required"`
}

func (h *Handler) New(w http.ResponseWriter, r *http.Request) error {
	var req newReq

	err := validate.ParseStruct(r, &req)
	if err != nil {
		return err
	}

	err = h.service.NewCategory(r.Context(), mapToNewCategoryInput(req))
	if err != nil {
		return catalogServ.ErrStatus.Resolve(err, "NewCategory")
	}

	return nil
}

type updateReq struct {
	Name string `json:"name" validate:"required"`
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) error {
	var req updateReq

	err := validate.ParseStruct(r, &req)
	if err != nil {
		return err
	}

	urlCategoryId, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return validate.NewError("invalid service category id")
	}

	err = h.service.UpdateCategory(r.Context(), urlCategoryId, mapToUpdateCategoryInput(req))
	if err != nil {
		return catalogServ.ErrStatus.Resolve(err, "UpdateCategory")
	}

	return nil
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) error {
	urlCategoryId, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return validate.NewError("invalid service category id")
	}

	err = h.service.DeleteCategory(r.Context(), urlCategoryId)
	if err != nil {
		return catalogServ.ErrStatus.Resolve(err, "DeleteCategory")
	}

	return nil
}

type reorderCategoriesReq struct {
	Categories []int `json:"categories" validate:"required"`
}

func (h *Handler) ReorderCategories(w http.ResponseWriter, r *http.Request) error {
	var req reorderCategoriesReq

	err := validate.ParseStruct(r, &req)
	if err != nil {
		return err
	}

	err = h.service.ReorderCategories(r.Context(), mapToReorderCategoriesInput(req))
	if err != nil {
		return catalogServ.ErrStatus.Resolve(err, "ReorderCategories")
	}

	return nil
}
