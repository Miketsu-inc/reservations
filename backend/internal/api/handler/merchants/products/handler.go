package products

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	productServ "github.com/miketsu-inc/reservations/backend/internal/service/product"
	"github.com/miketsu-inc/reservations/backend/pkg/currencyx"
	"github.com/miketsu-inc/reservations/backend/pkg/httputil"
	"github.com/miketsu-inc/reservations/backend/pkg/validate"
)

type Handler struct {
	service *productServ.Service
}

func NewHandler(s *productServ.Service) *Handler {
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
	Name          string           `json:"name" validate:"required"`
	Description   string           `json:"description"`
	Price         *currencyx.Price `json:"price"`
	Unit          string           `json:"unit" validate:"required"`
	MaxAmount     int              `json:"max_amount" validate:"min=0,max=10000000000"`
	CurrentAmount int              `json:"current_amount" validate:"min=0,max=10000000000"`
}

func (h *Handler) New(w http.ResponseWriter, r *http.Request) {
	var req newReq

	if err := validate.ParseStruct(r, &req); err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	err := h.service.New(r.Context(), mapToNewInput(req))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

type updateReq struct {
	Id            int              `json:"id"`
	Name          string           `json:"name" validate:"required"`
	Description   string           `json:"description"`
	Price         *currencyx.Price `json:"price"`
	Unit          string           `json:"unit" validate:"required"`
	MaxAmount     int              `json:"max_amount" validate:"min=0,max=10000000000"`
	CurrentAmount int              `json:"current_amount" validate:"min=0,max=10000000000"`
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	var req updateReq

	if err := validate.ParseStruct(r, &req); err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	urlProductId, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("invalid product id"))
		return
	}

	err = h.service.Update(r.Context(), urlProductId, mapToUpdateInput(req))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	urlProductId, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("invalid product id"))
		return
	}

	err = h.service.Delete(r.Context(), urlProductId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}
}

type getAllResp struct {
	Id            int                      `json:"id"`
	Name          string                   `json:"name"`
	Description   string                   `json:"description"`
	Price         *currencyx.Price         `json:"price"`
	Unit          string                   `json:"unit"`
	MaxAmount     int                      `json:"max_amount"`
	CurrentAmount int                      `json:"current_amount"`
	Services      []servicesForProdcutResp `json:"services"`
}

type servicesForProdcutResp struct {
	Id    int    `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

func (h *Handler) GetAll(w http.ResponseWriter, r *http.Request) {
	products, err := h.service.GetAll(r.Context())
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	result := mapToGetAllResp(products)

	httputil.Success(w, http.StatusOK, result)
}
