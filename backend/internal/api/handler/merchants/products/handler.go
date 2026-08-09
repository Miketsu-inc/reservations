package products

import (
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

func (h *Handler) Routes() *httputil.Router {
	r := httputil.NewRouter()

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

func (h *Handler) New(w http.ResponseWriter, r *http.Request) error {
	var req newReq

	if err := validate.ParseStruct(r, &req); err != nil {
		return err
	}

	err := h.service.New(r.Context(), mapToNewInput(req))
	if err != nil {
		return productServ.ErrStatus.Resolve(err, "New")
	}

	w.WriteHeader(http.StatusCreated)

	return nil
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

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) error {
	var req updateReq

	if err := validate.ParseStruct(r, &req); err != nil {
		return err
	}

	urlProductId, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return validate.NewError("invalid product id")
	}

	err = h.service.Update(r.Context(), urlProductId, mapToUpdateInput(req))
	if err != nil {
		return productServ.ErrStatus.Resolve(err, "Update")
	}

	return nil
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) error {
	urlProductId, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return validate.NewError("invalid product id")
	}

	err = h.service.Delete(r.Context(), urlProductId)
	if err != nil {
		return productServ.ErrStatus.Resolve(err, "Delete")
	}

	return nil
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

func (h *Handler) GetAll(w http.ResponseWriter, r *http.Request) error {
	products, err := h.service.GetAll(r.Context())
	if err != nil {
		return productServ.ErrStatus.Resolve(err, "GetAll")
	}

	httputil.Success(w, http.StatusOK, mapToGetAllResp(products))

	return nil
}
