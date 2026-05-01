package bookings

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/miketsu-inc/reservations/backend/internal/api/middleware"
	bookingServ "github.com/miketsu-inc/reservations/backend/internal/service/booking"
	"github.com/miketsu-inc/reservations/backend/internal/types"
	"github.com/miketsu-inc/reservations/backend/pkg/httputil"
	"github.com/miketsu-inc/reservations/backend/pkg/validate"
)

type Handler struct {
	service    *bookingServ.Service
	middleware *middleware.Manager
}

func NewHandler(s *bookingServ.Service, m *middleware.Manager) *Handler {
	return &Handler{service: s, middleware: m}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Group(func(r chi.Router) {
		r.Use(h.middleware.JwtAuthentication)
		r.Use(h.middleware.EmployeeAuthentication)
		r.Use(h.middleware.Language)

		r.Post("/", h.CreateByMerchant)
		r.Patch("/{id}", h.UpdateByMerchant)
		r.Delete("/{id}", h.CancelByMerchant)

		r.Patch("/{b_id}/participant/{p_id}", h.UpdateParticipantStatus)
	})

	return r
}

type createByMerchantReq struct {
	Customers    []customerReq     `json:"customers"`
	ServiceId    int               `json:"service_id" validate:"required"`
	TimeStamp    string            `json:"timestamp" validate:"required"`
	MerchantNote *string           `json:"merchant_note"`
	IsRecurring  bool              `json:"is_recurring"`
	Rrule        *recurringRuleReq `json:"recurrence_rule"`
}

type customerReq struct {
	CustomerId  *uuid.UUID `json:"id"`
	FirstName   *string    `json:"first_name"`
	LastName    *string    `json:"last_name"`
	Email       *string    `json:"email"`
	PhoneNumber *string    `json:"phone_number"`
}

type recurringRuleReq struct {
	Frequency string   `json:"frequency"`
	Interval  int      `json:"interval"`
	Weekdays  []string `json:"weekdays"`
	Until     string   `json:"until"`
}

func (h *Handler) CreateByMerchant(w http.ResponseWriter, r *http.Request) {
	var req createByMerchantReq

	if err := validate.ParseStruct(r, &req); err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	err := h.service.CreateByMerchant(r.Context(), mapToCreateByMerchantInput(req))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// validate:"required" on MerchantNote would fail
// if an empty string arrives as a note
type updateByMerchantReq struct {
	Customers       []customerReq       `json:"customers"`
	ServiceId       int                 `json:"service_id" validate:"required"`
	TimeStamp       string              `json:"timestamp" validate:"required"`
	MerchantNote    *string             `json:"merchant_note"`
	BookingStatus   types.BookingStatus `json:"booking_status"`
	UpdateAllFuture bool                `json:"update_all_future"`
}

func (h *Handler) UpdateByMerchant(w http.ResponseWriter, r *http.Request) {
	var req updateByMerchantReq

	if err := validate.ParseStruct(r, &req); err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	urlId, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("invalid booking id: %w", err))
		return
	}

	err = h.service.UpdateByMerchant(r.Context(), urlId, mapToUpdateByMerchantInput(req))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}
}

type cancelByMerchantReq struct {
	CancellationReason string `json:"cancellation_reason"`
}

func (h *Handler) CancelByMerchant(w http.ResponseWriter, r *http.Request) {
	var req cancelByMerchantReq

	if err := validate.ParseStruct(r, &req); err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	urlId, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("invalid booking id: %w", err))
		return
	}

	err = h.service.CancelByMerchant(r.Context(), urlId, mapToCancelByMerchantInput(req))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}
}

type updatePaticipantStatusReq struct {
	Status types.BookingStatus `json:"status"`
}

func (h *Handler) UpdateParticipantStatus(w http.ResponseWriter, r *http.Request) {
	var req updatePaticipantStatusReq

	if err := validate.ParseStruct(r, &req); err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	urlBookingId, err := strconv.Atoi(chi.URLParam(r, "b_id"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("invalid booking id: %w", err))
		return
	}

	urlParticipantId, err := strconv.Atoi(chi.URLParam(r, "p_id"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("invalid participant id: %w", err))
		return
	}

	err = h.service.UpdateParticipantStatus(r.Context(), urlBookingId, urlParticipantId, mapToUpdateParticipantStatusInput(req))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

}
