package bookings

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/miketsu-inc/reservations/backend/internal/api/middleware"
	bookingServ "github.com/miketsu-inc/reservations/backend/internal/service/booking"
	"github.com/miketsu-inc/reservations/backend/internal/types"
	"github.com/miketsu-inc/reservations/backend/pkg/currencyx"
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
		r.Use(h.middleware.Authentication)
		r.Use(h.middleware.Language)

		r.Post("/customer", h.CreateByCustomer)
		r.Delete("/customer/{id}", h.CancelByCustomer)
		r.Get("/customer/{id}", h.GetByCustomer)

		r.Post("/merchant", h.CreateByMerchant)
		r.Patch("/merchant/{id}", h.UpdateByMerchant)
		r.Delete("/merchant/{id}", h.CancelByMerchant)

		r.Patch("/merchant/{b_id}/participant/{p_id}", h.UpdateParticipantStatus)

		r.Get("/calendar/events", h.GetCalendarEvents)
	})

	return r
}

type createBookingByCustomerReq struct {
	MerchantName string `json:"merchant_name" validate:"required"`
	ServiceId    int    `json:"service_id" validate:"required"`
	LocationId   int    `json:"location_id" validate:"required"`
	TimeStamp    string `json:"timeStamp" validate:"required"`
	CustomerNote string `json:"customer_note"`
	// only present on group bookings
	BookingId *int `json:"booking_id"`
}

func (h *Handler) CreateByCustomer(w http.ResponseWriter, r *http.Request) {
	var req createBookingByCustomerReq

	if err := validate.ParseStruct(r, &req); err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	err := h.service.CreateByCustomer(r.Context(), mapToCreateByCustomerInput(req))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

type cancelByCustomerReq struct {
	BookingId    int    `json:"booking_id" validate:"required"`
	MerchantName string `json:"merchant_name" validate:"required"`
}

func (h *Handler) CancelByCustomer(w http.ResponseWriter, r *http.Request) {
	var req cancelByCustomerReq

	if err := validate.ParseStruct(r, &req); err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	urlId, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("invalid booking id: %w", err))
		return
	}

	err = h.service.CancelByCustomer(r.Context(), urlId, mapToCancelByCustomerInput(req))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}
}

type getByCustomerResp struct {
	FromDate          time.Time                `json:"from_date"`
	ToDate            time.Time                `json:"to_date"`
	ServiceName       string                   `json:"service_name"`
	CancelDeadline    int                      `json:"cancel_deadline"`
	FormattedLocation string                   `json:"formatted_location"`
	Price             currencyx.FormattedPrice `json:"price"`
	PriceType         types.PriceType          `json:"price_type"`
	MerchantName      string                   `json:"merchant_name"`
	IsCancelled       bool                     `json:"is_cancelled"`
}

func (h *Handler) GetByCustomer(w http.ResponseWriter, r *http.Request) {
	urlId, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("invalid booking id: %w", err))
		return
	}

	publicBooking, err := h.service.GetByCustomer(r.Context(), urlId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	httputil.Success(w, http.StatusOK, mapToGetByCustomerResp(publicBooking))
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

type getCalendarEventsResp struct {
	Bookings     []bookingForCalendar `json:"bookings"`
	BlockedTimes []blockedTime        `json:"blocked_times"`
}

type bookingForCalendar struct {
	ID              int                             `json:"id"`
	BookingType     types.BookingType               `json:"booking_type"`
	BookingStatus   types.BookingStatus             `json:"booking_status"`
	FromDate        time.Time                       `json:"from_date"`
	ToDate          time.Time                       `json:"to_date"`
	IsRecurring     bool                            `json:"is_recurring"`
	MerchantNote    *string                         `json:"merchant_note"`
	ServiceId       int                             `json:"service_id"`
	ServiceName     string                          `json:"service_name"`
	ServiceColor    string                          `json:"service_color" `
	MaxParticipants int                             `json:"max_participants"`
	Price           currencyx.FormattedPrice        `json:"price"`
	Cost            currencyx.FormattedPrice        `json:"cost"`
	Participants    []bookingParticipantForCalendar `json:"participants"`
}

type bookingParticipantForCalendar struct {
	Id           int                 `json:"id"`
	CustomerId   uuid.UUID           `json:"customer_id"`
	FirstName    *string             `json:"first_name"`
	LastName     *string             `json:"last_name"`
	CustomerNote *string             `json:"customer_note"`
	Status       types.BookingStatus `json:"status"`
}

type blockedTime struct {
	ID            int       `json:"id"`
	EmployeeId    int       `json:"employee_id"`
	Name          string    `json:"name"`
	FromDate      time.Time `json:"from_date"`
	ToDate        time.Time `json:"to_date"`
	AllDay        bool      `json:"all_day"`
	Icon          *string   `json:"icon"`
	BlockedTypeId *int      `json:"blocked_type_id"`
}

func (h *Handler) GetCalendarEvents(w http.ResponseWriter, r *http.Request) {
	start := r.URL.Query().Get("start")
	end := r.URL.Query().Get("end")

	bookings, err := h.service.GetCalendarEvents(r.Context(), start, end)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	httputil.Success(w, http.StatusOK, mapToGetCalendarEventsResp(bookings))
}
