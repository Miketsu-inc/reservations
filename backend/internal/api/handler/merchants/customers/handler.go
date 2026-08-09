package customers

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	customerServ "github.com/miketsu-inc/reservations/backend/internal/service/customer"
	"github.com/miketsu-inc/reservations/backend/internal/types"
	"github.com/miketsu-inc/reservations/backend/pkg/currencyx"
	"github.com/miketsu-inc/reservations/backend/pkg/httputil"
	"github.com/miketsu-inc/reservations/backend/pkg/validate"
)

type Handler struct {
	service *customerServ.Service
}

func NewHandler(s *customerServ.Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) Routes() *httputil.Router {
	r := httputil.NewRouter()

	r.Post("/", h.New)
	r.Put("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)
	r.Get("/{id}", h.Get)

	r.Get("/{id}/stats", h.GetStats)
	r.Put("/{id}/blacklist", h.Blacklist)
	r.Delete("/{id}/blacklist", h.UnBlacklist)

	r.Get("/", h.GetAll)
	r.Put("/transfer", h.TransferBookings)
	r.Get("/blacklist", h.GetAllBlacklisted)

	return r
}

type newReq struct {
	FirstName   *string    `json:"first_name" validate:"required"`
	LastName    *string    `json:"last_name" validate:"required"`
	Email       *string    `json:"email"`
	PhoneNumber *string    `json:"phone_number"`
	Birthday    *time.Time `json:"birthday"`
	Note        *string    `json:"note"`
}

func (h *Handler) New(w http.ResponseWriter, r *http.Request) error {
	var req newReq

	if err := validate.ParseStruct(r, &req); err != nil {
		return err
	}

	err := h.service.New(r.Context(), mapToNewInput(req))
	if err != nil {
		return customerServ.ErrStatus.Resolve(err, "New")
	}

	w.WriteHeader(http.StatusCreated)

	return nil
}

type updateReq struct {
	Id          uuid.UUID  `json:"id" validate:"required,uuid"`
	FirstName   *string    `json:"first_name"`
	LastName    *string    `json:"last_name"`
	Email       *string    `json:"email"`
	PhoneNumber *string    `json:"phone_number"`
	Birthday    *time.Time `json:"birthday"`
	Note        *string    `json:"note"`
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) error {
	var req updateReq

	if err := validate.ParseStruct(r, &req); err != nil {
		return err
	}

	urlCustomerId, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		return validate.NewError("invalid customer id")
	}

	err = h.service.Update(r.Context(), urlCustomerId, mapToUpdateInput(req))
	if err != nil {
		return customerServ.ErrStatus.Resolve(err, "Update")
	}

	return nil
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) error {
	urlCustomerId, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		return validate.NewError("invalid customer id")
	}

	err = h.service.Delete(r.Context(), urlCustomerId)
	if err != nil {
		return customerServ.ErrStatus.Resolve(err, "Delete")
	}

	return nil
}

type getResp struct {
	Id          uuid.UUID  `json:"id"`
	FirstName   *string    `json:"first_name"`
	LastName    *string    `json:"last_name"`
	Email       *string    `json:"email"`
	PhoneNumber *string    `json:"phone_number"`
	Birthday    *time.Time `json:"birthday"`
	Note        *string    `json:"note"`
	IsDummy     bool       `json:"is_dummy"`
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) error {
	urlCustomerId, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		return validate.NewError("invalid customer id")
	}

	customer, err := h.service.Get(r.Context(), urlCustomerId)
	if err != nil {
		return customerServ.ErrStatus.Resolve(err, "Get")
	}

	httputil.Success(w, http.StatusOK, mapToGetResp(customer))

	return nil
}

type getStatsResp struct {
	Id                   uuid.UUID              `json:"id"`
	FirstName            *string                `json:"first_name"`
	LastName             *string                `json:"last_name"`
	Email                *string                `json:"email"`
	PhoneNumber          *string                `json:"phone_number"`
	Birthday             *time.Time             `json:"birthday"`
	Note                 *string                `json:"note"`
	IsDummy              bool                   `json:"is_dummy"`
	IsBlacklisted        bool                   `json:"is_blacklisted"`
	BlacklistReason      *string                `json:"blacklist_reason"`
	TimesBooked          int                    `json:"times_booked"`
	TimesCancelledByUser int                    `json:"times_cancelled_by_user"`
	TimesUpcoming        int                    `json:"times_upcoming"`
	TimesCompleted       int                    `json:"times_completed"`
	Bookings             []customerBookingsResp `json:"bookings"`
}

type customerBookingsResp struct {
	FromDate          time.Time                `json:"from_date"`
	ToDate            time.Time                `json:"to_date"`
	ServiceName       string                   `json:"service_name"`
	CancelDeadline    int                      `json:"cancel_deadline"`
	FormattedLocation string                   `json:"formatted_location"`
	Price             currencyx.FormattedPrice `json:"price"`
	PriceType         types.PriceType          `json:"price_type"`
	MerchantName      string                   `json:"merchant_name"`
	Status            types.BookingStatus      `json:"status"`
}

func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) error {
	urlCustomerId, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		return validate.NewError("invalid customer id")
	}

	customerStats, err := h.service.GetStats(r.Context(), urlCustomerId)
	if err != nil {
		return customerServ.ErrStatus.Resolve(err, "GetStats")
	}

	httputil.Success(w, http.StatusOK, mapToGetStatsResp(customerStats))

	return nil
}

type blacklistReq struct {
	CustomerId      uuid.UUID `json:"id" validate:"required,uuid"`
	BlacklistReason *string   `json:"blacklist_reason"`
}

func (h *Handler) Blacklist(w http.ResponseWriter, r *http.Request) error {
	var req blacklistReq

	if err := validate.ParseStruct(r, &req); err != nil {
		return err
	}

	urlCustomerId, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		return validate.NewError("invalid customer id")
	}

	err = h.service.Blacklist(r.Context(), urlCustomerId, mapToBlacklistInput(req))
	if err != nil {
		return customerServ.ErrStatus.Resolve(err, "Blacklist")
	}

	return nil
}

func (h *Handler) UnBlacklist(w http.ResponseWriter, r *http.Request) error {
	urlCustomerId, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		return validate.NewError("invalid customer id")
	}

	err = h.service.UnBlacklist(r.Context(), urlCustomerId)
	if err != nil {
		return customerServ.ErrStatus.Resolve(err, "UnBlacklist")
	}

	return nil
}

type getAllResp struct {
	Id              uuid.UUID  `json:"id"`
	FirstName       *string    `json:"first_name"`
	LastName        *string    `json:"last_name"`
	Email           *string    `json:"email"`
	PhoneNumber     *string    `json:"phone_number"`
	Birthday        *time.Time `json:"birthday"`
	Note            *string    `json:"note"`
	IsDummy         bool       `json:"is_dummy"`
	IsBlacklisted   bool       `json:"is_blacklisted"`
	BlacklistReason *string    `json:"blacklist_reason"`
	TimesBooked     int        `json:"times_booked"`
	TimesCancelled  int        `json:"times_cancelled"`
}

func (h *Handler) GetAll(w http.ResponseWriter, r *http.Request) error {
	customers, err := h.service.GetAll(r.Context())
	if err != nil {
		return customerServ.ErrStatus.Resolve(err, "GetAll")
	}

	httputil.Success(w, http.StatusOK, mapToGetAllResp(customers))

	return nil
}

type transferBookingsReq struct {
	FromCustomerId uuid.UUID `json:"from_customer_id"`
	ToCustomerId   uuid.UUID `json:"to_customer_id"`
}

func (h *Handler) TransferBookings(w http.ResponseWriter, r *http.Request) error {
	var req transferBookingsReq

	if err := validate.ParseStruct(r, &req); err != nil {
		return err
	}

	err := h.service.TransferBookings(r.Context(), mapToTransferBookingsInput(req))
	if err != nil {
		return customerServ.ErrStatus.Resolve(err, "TransferBookings")
	}

	return nil
}

func (h *Handler) GetAllBlacklisted(w http.ResponseWriter, r *http.Request) error {
	blacklistedCustomers, err := h.service.GetAllBlacklisted(r.Context())
	if err != nil {
		return customerServ.ErrStatus.Resolve(err, "GetAllBlacklisted")
	}

	httputil.Success(w, http.StatusOK, mapToGetAllResp(blacklistedCustomers))

	return nil
}
