package services

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	catalogServ "github.com/miketsu-inc/reservations/backend/internal/service/catalog"
	"github.com/miketsu-inc/reservations/backend/internal/types"
	"github.com/miketsu-inc/reservations/backend/pkg/currencyx"
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
	r.Get("/{id}", h.Get)

	r.Put("/{id}/products", h.UpdateServiceProduct)
	// TODO: maybe replace these by a unified status route?
	r.Patch("/{id}/activate", h.Activate)
	r.Patch("/{id}/deactivate", h.Deactivate)

	r.Get("/", h.GetAll)
	r.Put("/reorder", h.Reorder)
	r.Get("/form-options", h.GetFormOptions)

	r.Post("/group", h.NewGroup)
	r.Put("/group/{id}", h.UpdateGroup)
	r.Get("/group/{id}", h.GetGroup)

	return r
}

type newReq struct {
	Name         string                 `json:"name" validate:"required"`
	Description  *string                `json:"description"`
	Color        string                 `json:"color" validate:"required,hexcolor"`
	Price        *currencyx.Price       `json:"price"`
	Cost         *currencyx.Price       `json:"cost"`
	PriceType    types.PriceType        `json:"price_type"`
	CategoryId   *int                   `json:"category_id"`
	IsActive     bool                   `json:"is_active"`
	Settings     serviceSettingsReq     `json:"settings"`
	Phases       []newPhaseReq          `json:"phases" validate:"required"`
	UsedProducts []connectedProductsReq `json:"used_products" validate:"required"`
}

type serviceSettingsReq struct {
	CancelDeadline   *int                `json:"cancel_deadline"`
	BookingWindowMin *int                `json:"booking_window_min"`
	BookingWindowMax *int                `json:"booking_window_max"`
	BufferTime       *int                `json:"buffer_time"`
	ApprovalPolicy   *types.ApprovalType `json:"approval_policy"`
}

type newPhaseReq struct {
	Name      string                 `json:"name" validate:"required"`
	Sequence  int                    `json:"sequence" validate:"required"`
	Duration  int                    `json:"duration" validate:"required,min=1,max=1440"`
	PhaseType types.ServicePhaseType `json:"phase_type" validate:"required,eq=wait|eq=active"`
}

type connectedProductsReq struct {
	ProductId  int `json:"id" validate:"required"`
	AmountUsed int `json:"amount_used" validate:"min=0,max=1000000"`
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
	Id          int                `json:"id"`
	Name        string             `json:"name" validate:"required"`
	Description *string            `json:"description"`
	Color       string             `json:"color" validate:"required,hexcolor"`
	Price       *currencyx.Price   `json:"price"`
	Cost        *currencyx.Price   `json:"cost"`
	PriceType   types.PriceType    `json:"price_type"`
	CategoryId  *int               `json:"category_id"`
	IsActive    bool               `json:"is_active"`
	Settings    serviceSettingsReq `json:"settings"`
	Phases      []phaseReq         `json:"phases" validate:"required"`
}

type phaseReq struct {
	Id        int                    `json:"id"`
	ServiceId int                    `json:"service_id"`
	Name      string                 `json:"name" validate:"required"`
	Sequence  int                    `json:"sequence" validate:"required"`
	Duration  int                    `json:"duration" validate:"required,min=1,max=1440"`
	PhaseType types.ServicePhaseType `json:"phase_type" validate:"required,eq=wait|eq=active"`
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	var req updateReq

	if err := validate.ParseStruct(r, &req); err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	urlServiceId, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("invalid service id"))
		return
	}

	err = h.service.Update(r.Context(), urlServiceId, mapToUpdateInput(req))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	urlServiceId, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("invalid service id"))
		return
	}

	err = h.service.Delete(r.Context(), urlServiceId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}
}

type getResp struct {
	Id            int                `json:"id"`
	CategoryId    *int               `json:"category_id"`
	Name          string             `json:"name"`
	Description   *string            `json:"description"`
	Color         string             `json:"color"`
	TotalDuration int                `json:"total_duration"`
	Price         *currencyx.Price   `json:"price"`
	Cost          *currencyx.Price   `json:"cost"`
	PriceType     types.PriceType    `json:"price_type"`
	IsActive      bool               `json:"is_active"`
	Sequence      int                `json:"sequence"`
	Settings      serviceSettingsReq `json:"settings"`
	Phases        []phaseReq         `json:"phases"`
	UsedProducts  []productResp      `json:"used_products"`
}

type productResp struct {
	Id         int    `json:"id"`
	Name       string `json:"name"`
	Unit       string `json:"unit"`
	AmountUsed int    `json:"amount_used"`
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	urlServiceId, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("invalid service id"))
		return
	}

	service, err := h.service.Get(r.Context(), urlServiceId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	httputil.Success(w, http.StatusOK, mapToGetResp(service))
}

type updateServiceProductReq struct {
	ServiceId    int                    `json:"service_id" validate:"required"`
	UsedProducts []connectedProductsReq `json:"used_products" validate:"required"`
}

func (h *Handler) UpdateServiceProduct(w http.ResponseWriter, r *http.Request) {
	var req updateServiceProductReq

	if err := validate.ParseStruct(r, &req); err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	urlServiceId, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("invalid service id"))
		return
	}

	err = h.service.UpdateServiceProduct(r.Context(), urlServiceId, mapToUpdateServiceProductInput(req))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}
}

func (h *Handler) Activate(w http.ResponseWriter, r *http.Request) {
	urlServiceId, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("invalid service id"))
		return
	}

	err = h.service.Activate(r.Context(), urlServiceId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}
}

func (h *Handler) Deactivate(w http.ResponseWriter, r *http.Request) {
	urlServiceId, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("invalid service id"))
		return
	}

	err = h.service.Deactivate(r.Context(), urlServiceId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}
}

type getAllResp struct {
	Id       *int          `json:"id"`
	Name     *string       `json:"name"`
	Sequence *int          `json:"sequence"`
	Services []serviceResp `json:"services"`
}

type serviceResp struct {
	Id              int               `json:"id"`
	MerchantId      uuid.UUID         `json:"merchant_id"`
	BookingType     types.BookingType `json:"booking_type"`
	CategoryId      *int              `json:"category_id"`
	Name            string            `json:"name"`
	Description     *string           `json:"description"`
	Color           string            `json:"color"`
	TotalDuration   int               `json:"total_duration"`
	Price           *currencyx.Price  `json:"price"`
	Cost            *currencyx.Price  `json:"cost"`
	PriceType       types.PriceType   `json:"price_type"`
	IsActive        bool              `json:"is_active"`
	MinParticipants int               `json:"min_participants"`
	MaxParticipants int               `json:"max_participants"`
	Sequence        int               `json:"sequence"`
	Phases          []phaseReq        `json:"phases"`
}

func (h *Handler) GetAll(w http.ResponseWriter, r *http.Request) {
	services, err := h.service.GetAll(r.Context())
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	httputil.Success(w, http.StatusOK, mapToGetAllResp(services))
}

type reorderReq struct {
	CategoryId *int  `json:"category_id"`
	Services   []int `json:"services" validate:"required"`
}

func (h *Handler) Reorder(w http.ResponseWriter, r *http.Request) {
	var req reorderReq

	err := validate.ParseStruct(r, &req)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	err = h.service.Reorder(r.Context(), mapToReorderInput(req))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}
}

type getFormOptionsResp struct {
	Products   []minimalProductResp  `json:"products"`
	Categories []serviceCategoryResp `json:"categories"`
}

type minimalProductResp struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
	Unit string `json:"unit"`
}

type serviceCategoryResp struct {
	Id         int       `json:"id"`
	MerchantId uuid.UUID `json:"merchant_id"`
	LocationId int       `json:"location_id"`
	Name       string    `json:"name"`
	Sequence   int       `json:"sequence"`
}

func (h *Handler) GetFormOptions(w http.ResponseWriter, r *http.Request) {
	formOptions, err := h.service.GetFormOptions(r.Context())
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	httputil.Success(w, http.StatusOK, mapToGetFormOptionsResp(formOptions))
}

type newGroupReq struct {
	Name            string                 `json:"name" validate:"required"`
	Description     *string                `json:"description"`
	Color           string                 `json:"color" validate:"required,hexcolor"`
	Price           *currencyx.Price       `json:"price"`
	Cost            *currencyx.Price       `json:"cost"`
	PriceType       types.PriceType        `json:"price_type"`
	Duration        int                    `json:"duration" validate:"required"`
	CategoryId      *int                   `json:"category_id"`
	MinParticipants *int                   `json:"min_participants"`
	MaxParticipants int                    `json:"max_participants" validate:"required"`
	IsActive        bool                   `json:"is_active"`
	Settings        serviceSettingsReq     `json:"settings"`
	UsedProducts    []connectedProductsReq `json:"used_products" validate:"required"`
}

func (h *Handler) NewGroup(w http.ResponseWriter, r *http.Request) {
	var req newGroupReq

	if err := validate.ParseStruct(r, &req); err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	err := h.service.NewGroup(r.Context(), mapToNewGroupInput(req))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

type updateGroupReq struct {
	Id              int                `json:"id" validate:"required"`
	Name            string             `json:"name" validate:"required"`
	Description     *string            `json:"description"`
	Color           string             `json:"color" validate:"required,hexcolor"`
	Price           *currencyx.Price   `json:"price"`
	Cost            *currencyx.Price   `json:"cost"`
	PriceType       types.PriceType    `json:"price_type"`
	Duration        int                `json:"duration" validate:"required"`
	CategoryId      *int               `json:"category_id"`
	MinParticipants *int               `json:"min_participants"`
	MaxParticipants int                `json:"max_participants" validate:"required"`
	IsActive        bool               `json:"is_active"`
	Settings        serviceSettingsReq `json:"settings"`
}

func (h *Handler) UpdateGroup(w http.ResponseWriter, r *http.Request) {
	var req updateGroupReq

	if err := validate.ParseStruct(r, &req); err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	urlServiceId, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("invalid service id"))
		return
	}

	err = h.service.UpdateGroup(r.Context(), urlServiceId, mapToUpdateGroupInput(req))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}
}

type getGroupResp struct {
	Id              int                  `json:"id"`
	CategoryId      *int                 `json:"category_id"`
	Name            string               `json:"name"`
	Description     *string              `json:"description"`
	Color           string               `json:"color"`
	Duration        int                  `json:"duration"`
	Price           *currencyx.Price     `json:"price"`
	Cost            *currencyx.Price     `json:"cost"`
	PriceType       types.PriceType      `json:"price_type"`
	IsActive        bool                 `json:"is_active"`
	Sequence        int                  `json:"sequence"`
	MinParicipants  int                  `json:"min_participants"`
	MaxParticipants int                  `json:"max_participants"`
	Settings        serviceSettingsReq   `json:"settings"`
	Products        []minimalProductResp `json:"used_products"`
}

func (h *Handler) GetGroup(w http.ResponseWriter, r *http.Request) {
	urlServiceId, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("invalid service id"))
		return
	}

	service, err := h.service.GetGroup(r.Context(), urlServiceId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	httputil.Success(w, http.StatusOK, mapToGetGroupResp(service))
}
