package services

import (
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

func (h *Handler) Routes() *httputil.Router {
	r := httputil.NewRouter()

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

	return r
}

type newReq struct {
	BookingType     types.BookingType      `json:"booking_type" validate:"required"`
	Name            string                 `json:"name" validate:"required"`
	Description     *string                `json:"description"`
	Color           string                 `json:"color" validate:"required,hexcolor"`
	Price           *currencyx.Price       `json:"price"`
	PriceType       types.PriceType        `json:"price_type" validate:"required"`
	CategoryId      *int                   `json:"category_id"`
	MinParticipants *int                   `json:"min_participants"`
	MaxParticipants *int                   `json:"max_participants"`
	IsActive        bool                   `json:"is_active"`
	Settings        serviceSettingsReq     `json:"settings"`
	EmployeeIds     []int                  `json:"employee_ids"`
	Phases          []newPhaseReq          `json:"phases" validate:"required"`
	UsedProducts    []connectedProductsReq `json:"used_products" validate:"required"`
}

type serviceSettingsReq struct {
	CancelDeadline   *int                `json:"cancel_deadline"`
	BookingWindowMin *int                `json:"booking_window_min"`
	BookingWindowMax *int                `json:"booking_window_max"`
	BufferTime       *int                `json:"buffer_time"`
	ApprovalPolicy   *types.ApprovalType `json:"approval_policy"`
}

type newPhaseReq struct {
	Name      string                 `json:"name"`
	Sequence  int                    `json:"sequence" validate:"required,min=1"`
	Duration  int                    `json:"duration" validate:"required,min=1,max=1440"`
	PhaseType types.ServicePhaseType `json:"phase_type" validate:"required,eq=wait|eq=active"`
}

type connectedProductsReq struct {
	ProductId  int `json:"id" validate:"required"`
	AmountUsed int `json:"amount_used" validate:"min=0,max=1000000"`
}

func (h *Handler) New(w http.ResponseWriter, r *http.Request) error {
	var req newReq

	if err := validate.ParseStruct(r, &req); err != nil {
		return err
	}

	err := h.service.New(r.Context(), mapToNewInput(req))
	if err != nil {
		return catalogServ.ErrStatus.Resolve(err, "New")
	}

	w.WriteHeader(http.StatusCreated)

	return nil
}

type updateReq struct {
	Id              int                `json:"id"`
	BookingType     types.BookingType  `json:"booking_type" validate:"required"`
	Name            string             `json:"name" validate:"required"`
	Description     *string            `json:"description"`
	Color           string             `json:"color" validate:"required,hexcolor"`
	Price           *currencyx.Price   `json:"price"`
	PriceType       types.PriceType    `json:"price_type" validate:"required"`
	CategoryId      *int               `json:"category_id"`
	MinParticipants *int               `json:"min_participants"`
	MaxParticipants *int               `json:"max_participants"`
	IsActive        bool               `json:"is_active"`
	Settings        serviceSettingsReq `json:"settings"`
	EmployeeIds     []int              `json:"employee_ids"`
	Phases          []phaseReq         `json:"phases" validate:"required"`
}

type phaseReq struct {
	Id        int                    `json:"id"`
	Name      string                 `json:"name"`
	Sequence  int                    `json:"sequence" validate:"required,min=1"`
	Duration  int                    `json:"duration" validate:"required,min=1,max=1440"`
	PhaseType types.ServicePhaseType `json:"phase_type" validate:"required,eq=wait|eq=active"`
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) error {
	var req updateReq

	if err := validate.ParseStruct(r, &req); err != nil {
		return err
	}

	urlServiceId, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return validate.NewError("invalid service id")
	}

	if urlServiceId != req.Id {
		return validate.NewError("invalid service id")
	}

	err = h.service.Update(r.Context(), mapToUpdateInput(req))
	if err != nil {
		return catalogServ.ErrStatus.Resolve(err, "Update")
	}

	return nil
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) error {
	urlServiceId, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return validate.NewError("invalid service id")
	}

	err = h.service.Delete(r.Context(), urlServiceId)
	if err != nil {
		return catalogServ.ErrStatus.Resolve(err, "Delete")
	}

	return nil
}

type getResp struct {
	Id              int                `json:"id"`
	BookingType     types.BookingType  `json:"booking_type"`
	CategoryId      *int               `json:"category_id"`
	Name            string             `json:"name"`
	Description     *string            `json:"description"`
	Color           string             `json:"color"`
	TotalDuration   int                `json:"total_duration"`
	Price           *currencyx.Price   `json:"price"`
	PriceType       types.PriceType    `json:"price_type"`
	IsActive        bool               `json:"is_active"`
	Sequence        int                `json:"sequence"`
	MinParicipants  int                `json:"min_participants"`
	MaxParticipants int                `json:"max_participants"`
	Settings        serviceSettingsReq `json:"settings"`
	Phases          []phaseReq         `json:"phases"`
	UsedProducts    []productResp      `json:"used_products"`
}

type productResp struct {
	Id         int    `json:"id"`
	Name       string `json:"name"`
	Unit       string `json:"unit"`
	AmountUsed int    `json:"amount_used"`
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) error {
	urlServiceId, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return validate.NewError("invalid service id")
	}

	service, err := h.service.Get(r.Context(), urlServiceId)
	if err != nil {
		return catalogServ.ErrStatus.Resolve(err, "Get")
	}

	httputil.Success(w, http.StatusOK, mapToGetResp(service))

	return nil
}

type updateServiceProductReq struct {
	ServiceId    int                    `json:"service_id" validate:"required"`
	UsedProducts []connectedProductsReq `json:"used_products" validate:"required"`
}

func (h *Handler) UpdateServiceProduct(w http.ResponseWriter, r *http.Request) error {
	var req updateServiceProductReq

	if err := validate.ParseStruct(r, &req); err != nil {
		return err
	}

	urlServiceId, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return validate.NewError("invalid service id")
	}

	err = h.service.UpdateServiceProduct(r.Context(), urlServiceId, mapToUpdateServiceProductInput(req))
	if err != nil {
		return catalogServ.ErrStatus.Resolve(err, "UpdateServiceProduct")
	}

	return nil
}

func (h *Handler) Activate(w http.ResponseWriter, r *http.Request) error {
	urlServiceId, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return validate.NewError("invalid service id")
	}

	err = h.service.Activate(r.Context(), urlServiceId)
	if err != nil {
		return catalogServ.ErrStatus.Resolve(err, "Activate")
	}

	return nil
}

func (h *Handler) Deactivate(w http.ResponseWriter, r *http.Request) error {
	urlServiceId, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return validate.NewError("invalid service id")
	}

	err = h.service.Deactivate(r.Context(), urlServiceId)
	if err != nil {
		return catalogServ.ErrStatus.Resolve(err, "Deactivate")
	}

	return nil
}

type getAllResp struct {
	Id       *int          `json:"id"`
	Name     *string       `json:"name"`
	Sequence *int          `json:"sequence"`
	Services []serviceResp `json:"services"`
}

type serviceResp struct {
	Id              int                       `json:"id"`
	MerchantId      uuid.UUID                 `json:"merchant_id"`
	BookingType     types.BookingType         `json:"booking_type"`
	CategoryId      *int                      `json:"category_id"`
	Name            string                    `json:"name"`
	Description     *string                   `json:"description"`
	Color           string                    `json:"color"`
	TotalDuration   int                       `json:"total_duration"`
	Price           *currencyx.FormattedPrice `json:"price"`
	PriceType       types.PriceType           `json:"price_type"`
	IsActive        bool                      `json:"is_active"`
	MinParticipants int                       `json:"min_participants"`
	MaxParticipants int                       `json:"max_participants"`
	Sequence        int                       `json:"sequence"`
	Phases          []phaseReq                `json:"phases"`
}

func (h *Handler) GetAll(w http.ResponseWriter, r *http.Request) error {
	services, err := h.service.GetAll(r.Context())
	if err != nil {
		return catalogServ.ErrStatus.Resolve(err, "GetAll")
	}

	httputil.Success(w, http.StatusOK, mapToGetAllResp(services))

	return nil
}

type reorderReq struct {
	CategoryId *int  `json:"category_id"`
	Services   []int `json:"services" validate:"required"`
}

func (h *Handler) Reorder(w http.ResponseWriter, r *http.Request) error {
	var req reorderReq

	err := validate.ParseStruct(r, &req)
	if err != nil {
		return err
	}

	err = h.service.Reorder(r.Context(), mapToReorderInput(req))
	if err != nil {
		return catalogServ.ErrStatus.Resolve(err, "Reorder")
	}

	return nil
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

func (h *Handler) GetFormOptions(w http.ResponseWriter, r *http.Request) error {
	formOptions, err := h.service.GetFormOptions(r.Context())
	if err != nil {
		return catalogServ.ErrStatus.Resolve(err, "GetFormOptions")
	}

	httputil.Success(w, http.StatusOK, mapToGetFormOptionsResp(formOptions))

	return nil
}
